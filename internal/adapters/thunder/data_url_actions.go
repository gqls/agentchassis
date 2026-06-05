// FILE: internal/adapters/thunder/data_url_actions.go
//
// Phase 4 data-flow actions: presigned-URL generation for the training VM.
//
// The adapter is the credential boundary. The training VM never holds B2
// credentials; instead the adapter hands it time-limited presigned URLs:
//   - prepare_dataset_url  → GET URL: the VM downloads the training JSONL from B2.
//   - prepare_artefact_url → PUT URL: the VM uploads the trained adapter to B2.
//
// Key conventions MUST match training-data-preparer, which writes the dataset
// to bucket personae-model-training at key:
//     finetuning/datasets/{export_id}/training.jsonl
// The adapter therefore derives the dataset key from export_id, and the
// artefact key from training_run_id, so callers pass IDs — never raw paths.
//
// Both actions are read-only against Thunder (no instance lifecycle); they
// only sign URLs. They require the optional object-storage client; if it
// wasn't configured at startup (TRAINING_BUCKET / B2 creds absent), the
// handlers return a clear error_unrecoverable rather than panicking.

package thunder

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/storage"
)

// ─────────────────────────────────────────────────────────────────────────
// Key templates — keep in sync with training-data-preparer.
// ─────────────────────────────────────────────────────────────────────────

const (
	// datasetKeyTemplate matches the live training-data-preparer agent_def
	// (verified 2026-05-22): bucket=finetuning,
	// s3_key_template=finetuning/datasets/{export_id}/training.jsonl.
	// The adapter targets the same bucket (TRAINING_BUCKET default "finetuning"),
	// so the presigned GET resolves against the data the preparer actually wrote.
	// Format arg: export_id.
	datasetKeyTemplate = "finetuning/datasets/%s/training.jsonl"

	// artefactKeyTemplate is where a finished QLoRA adapter is uploaded.
	// Same bucket as the dataset; format arg: training_run_id.
	artefactKeyTemplate = "finetuning/artefacts/%s/adapter.tar.gz"

	// Default presigned-URL lifetimes (minutes). Downloads are quick; the
	// artefact upload window must outlast a full training run, hence longer.
	defaultDatasetURLExpiryMin  = 60      // 1h to start the download
	defaultArtefactURLExpiryMin = 24 * 60 // 24h: must outlive the training run
)

// ─────────────────────────────────────────────────────────────────────────
// Request / Result shapes
// ─────────────────────────────────────────────────────────────────────────

// PrepareDatasetURLRequest. The caller passes export_id; the adapter builds
// the canonical key. expiry_minutes is optional (defaults applied).
type PrepareDatasetURLRequest struct {
	ExportID      string `json:"export_id"`
	ExpiryMinutes int    `json:"expiry_minutes,omitempty"`
}

// PrepareArtefactURLRequest. The caller passes training_run_id; the adapter
// builds the canonical artefact key.
type PrepareArtefactURLRequest struct {
	TrainingRunID string `json:"training_run_id"`
	ExpiryMinutes int    `json:"expiry_minutes,omitempty"`
}

// ObjectURLRequest is the general presign primitive: the caller passes an
// explicit, already-bucket-relative key (the chassis dispatch action strips
// any s3://bucket/ prefix before sending), plus the HTTP method. Used by the
// training-launcher (Phase 5) to presign the dataset (GET, key derived from
// the preparer's dataset_uri) and the scripts bundle (GET, literal key).
// Method defaults to GET; PUT is supported for symmetry with artefact uploads.
type ObjectURLRequest struct {
	Key           string `json:"key"`
	Method        string `json:"method,omitempty"`
	ExpiryMinutes int    `json:"expiry_minutes,omitempty"`
}

// PreparedURLResult is the shared result shape for both actions.
type PreparedURLResult struct {
	PresignedURL string `json:"presigned_url"`
	Key          string `json:"key"`        // the B2 object key (for logging/audit)
	ExpiresAt    string `json:"expires_at"` // RFC3339
	Method       string `json:"method"`     // "GET" or "PUT"
}

// ─────────────────────────────────────────────────────────────────────────
// DataURLAction
// ─────────────────────────────────────────────────────────────────────────

// DataURLAction generates presigned B2 URLs. It holds only the storage
// client; key derivation is pure. Constructed in NewAdapter when object
// storage is configured.
type DataURLAction struct {
	storage storage.Client
	logger  *zap.Logger
}

// NewDataURLAction builds a DataURLAction.
func NewDataURLAction(s storage.Client, logger *zap.Logger) *DataURLAction {
	return &DataURLAction{
		storage: s,
		logger:  logger.Named("data_url"),
	}
}

// ObjectURL is the general presign primitive: sign a GET or PUT for an
// explicit, already-bucket-relative key. DatasetURL/ArtefactURL delegate to it
// so there is a single presign code path (no parallel signing logic). Method
// defaults to GET; expiry defaults by method (GET short, PUT long enough to
// outlast a training run) unless the caller overrides.
func (d *DataURLAction) ObjectURL(ctx context.Context, req ObjectURLRequest) (*PreparedURLResult, error) {
	if req.Key == "" {
		return nil, fmt.Errorf("key is required")
	}
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "GET"
	}
	expiry := req.ExpiryMinutes

	var (
		url string
		err error
	)
	switch method {
	case "GET":
		if expiry <= 0 {
			expiry = defaultDatasetURLExpiryMin
		}
		url, err = d.storage.GetPresignedURL(ctx, req.Key, expiry)
	case "PUT":
		if expiry <= 0 {
			expiry = defaultArtefactURLExpiryMin
		}
		url, err = d.storage.GetPresignedPutURL(ctx, req.Key, expiry)
	default:
		return nil, fmt.Errorf("unsupported method %q (want GET or PUT)", method)
	}
	if err != nil {
		return nil, fmt.Errorf("presign %s %s: %w", method, req.Key, err)
	}
	return &PreparedURLResult{
		PresignedURL: url,
		Key:          req.Key,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(expiry) * time.Minute).Format(time.RFC3339),
		Method:       method,
	}, nil
}

// DatasetURL returns a presigned GET URL for the dataset belonging to export_id.
// Delegates to ObjectURL after deriving the canonical key.
func (d *DataURLAction) DatasetURL(ctx context.Context, req PrepareDatasetURLRequest) (*PreparedURLResult, error) {
	if req.ExportID == "" {
		return nil, fmt.Errorf("export_id is required")
	}
	return d.ObjectURL(ctx, ObjectURLRequest{
		Key:           fmt.Sprintf(datasetKeyTemplate, req.ExportID),
		Method:        "GET",
		ExpiryMinutes: req.ExpiryMinutes,
	})
}

// ArtefactURL returns a presigned PUT URL for the artefact of training_run_id.
// Delegates to ObjectURL after deriving the canonical key.
func (d *DataURLAction) ArtefactURL(ctx context.Context, req PrepareArtefactURLRequest) (*PreparedURLResult, error) {
	if req.TrainingRunID == "" {
		return nil, fmt.Errorf("training_run_id is required")
	}
	return d.ObjectURL(ctx, ObjectURLRequest{
		Key:           fmt.Sprintf(artefactKeyTemplate, req.TrainingRunID),
		Method:        "PUT",
		ExpiryMinutes: req.ExpiryMinutes,
	})
}

// ─────────────────────────────────────────────────────────────────────────
// Adapter handler methods (dispatch targets)
// ─────────────────────────────────────────────────────────────────────────

// handlePrepareDatasetURL parses the body and returns a presigned GET URL.
// Mirrors handleProvisionInstance's parse + response pattern.
func (a *Adapter) handlePrepareDatasetURL(
	body map[string]interface{},
	reqHeaders map[string]string,
	replyToTopic string,
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic on prepare_dataset_url request — cannot send response")
		return
	}
	if a.dataURLAction == nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_dataset_url",
			"storage_unavailable",
			"object storage not configured (TRAINING_BUCKET / B2 credentials missing at startup)",
			"error_unrecoverable", l)
		return
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_dataset_url",
			"invalid_request_body", fmt.Sprintf("could not re-marshal body: %v", err),
			"error_unrecoverable", l)
		return
	}
	var req PrepareDatasetURLRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_dataset_url",
			"invalid_request_body", fmt.Sprintf("could not unmarshal into PrepareDatasetURLRequest: %v", err),
			"error_unrecoverable", l)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := a.dataURLAction.DatasetURL(ctx, req)
	if err != nil {
		// Presign failures are typically config/credential issues — unrecoverable.
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_dataset_url",
			"presign_failed", err.Error(), "error_unrecoverable", l)
		return
	}

	a.sendSuccessResponse(reqHeaders, replyToTopic, "prepare_dataset_url",
		map[string]interface{}{
			"presigned_url": result.PresignedURL,
			"key":           result.Key,
			"expires_at":    result.ExpiresAt,
			"method":        result.Method,
		}, l)
}

// handlePrepareArtefactURL parses the body and returns a presigned PUT URL.
func (a *Adapter) handlePrepareArtefactURL(
	body map[string]interface{},
	reqHeaders map[string]string,
	replyToTopic string,
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic on prepare_artefact_url request — cannot send response")
		return
	}
	if a.dataURLAction == nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_artefact_url",
			"storage_unavailable",
			"object storage not configured (TRAINING_BUCKET / B2 credentials missing at startup)",
			"error_unrecoverable", l)
		return
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_artefact_url",
			"invalid_request_body", fmt.Sprintf("could not re-marshal body: %v", err),
			"error_unrecoverable", l)
		return
	}
	var req PrepareArtefactURLRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_artefact_url",
			"invalid_request_body", fmt.Sprintf("could not unmarshal into PrepareArtefactURLRequest: %v", err),
			"error_unrecoverable", l)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := a.dataURLAction.ArtefactURL(ctx, req)
	if err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_artefact_url",
			"presign_failed", err.Error(), "error_unrecoverable", l)
		return
	}

	a.sendSuccessResponse(reqHeaders, replyToTopic, "prepare_artefact_url",
		map[string]interface{}{
			"presigned_url": result.PresignedURL,
			"key":           result.Key,
			"expires_at":    result.ExpiresAt,
			"method":        result.Method,
		}, l)
}

// handlePrepareObjectURL parses the body and returns a presigned URL for an
// explicit key (GET or PUT). Mirrors handlePrepareDatasetURL's parse +
// response pattern; the reply key "presigned_url" is what the launcher's
// ssh_exec_launch tokens (scripts_url_result.presigned_url /
// dataset_url_result.presigned_url) resolve against.
func (a *Adapter) handlePrepareObjectURL(
	body map[string]interface{},
	reqHeaders map[string]string,
	replyToTopic string,
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic on prepare_object_url request — cannot send response")
		return
	}
	if a.dataURLAction == nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_object_url",
			"storage_unavailable",
			"object storage not configured (TRAINING_BUCKET / B2 credentials missing at startup)",
			"error_unrecoverable", l)
		return
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_object_url",
			"invalid_request_body", fmt.Sprintf("could not re-marshal body: %v", err),
			"error_unrecoverable", l)
		return
	}
	var req ObjectURLRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_object_url",
			"invalid_request_body", fmt.Sprintf("could not unmarshal into ObjectURLRequest: %v", err),
			"error_unrecoverable", l)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := a.dataURLAction.ObjectURL(ctx, req)
	if err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_object_url",
			"presign_failed", err.Error(), "error_unrecoverable", l)
		return
	}

	a.sendSuccessResponse(reqHeaders, replyToTopic, "prepare_object_url",
		map[string]interface{}{
			"presigned_url": result.PresignedURL,
			"key":           result.Key,
			"expires_at":    result.ExpiresAt,
			"method":        result.Method,
		}, l)
}

// ─────────────────────────────────────────────────────────────────────────
// Resume support (Phase D): pick the latest checkpoint for a run and presign
// a GET so a fresh box can continue an interrupted training_run_id.
// ─────────────────────────────────────────────────────────────────────────

const (
	// checkpointPrefixTemplate MUST match the prefix of compute_checkpoint_keys'
	// key template ("finetuning/checkpoints/%s/ckpt-%d.tar.gz"). Format arg:
	// training_run_id. If that template is customised in the launcher, the
	// parser in latestCheckpointKey must change to match.
	checkpointPrefixTemplate = "finetuning/checkpoints/%s/"

	// defaultResumeURLExpiryMin: the resume GET only needs to outlast the VM's
	// setup + download at startup; 6h is ample.
	defaultResumeURLExpiryMin = 6 * 60
)

// objectLister is the NARROW capability the resume path needs: list object keys
// under a prefix. It is deliberately defined HERE (not added to storage.Client)
// so that introducing listing does NOT widen the broadly-implemented
// storage.Client interface and force every importer (core-manager, all adapters)
// to change — see debugging guide item 18. *storage.S3Client satisfies this once
// it gains a ListKeys method; if the configured storage backend does not, ResumeURL
// degrades to found=false (a fresh run) rather than failing.
type objectLister interface {
	ListKeys(ctx context.Context, prefix string) ([]string, error)
}

// ResumeURLRequest. Caller passes training_run_id; the adapter derives the
// checkpoint prefix, lists it, picks the latest checkpoint, and presigns a GET.
type ResumeURLRequest struct {
	TrainingRunID string `json:"training_run_id"`
	ExpiryMinutes int    `json:"expiry_minutes,omitempty"`
}

// ResumeURLResult is the resume-presign reply. Found=false means the run has no
// checkpoints in B2 yet — the launcher treats that as a fresh start (no resume key).
type ResumeURLResult struct {
	Found        bool   `json:"found"`
	PresignedURL string `json:"presigned_url"`
	Key          string `json:"key"`
	Index        int    `json:"index"`
	ExpiresAt    string `json:"expires_at"`
}

// ResumeURL lists the run's checkpoints in B2, picks the latest, and returns a
// presigned GET for it. Returns Found=false (no error) when the run has no
// checkpoints yet — the launcher treats that as a fresh start. Listing reuses
// the existing storage.Client.ListObjects; the presign reuses the existing GET
// path (GetPresignedURL). A genuine list/presign failure is returned as an error
// (handled like the other prepare_* actions); an empty prefix is not an error —
// ListObjects returns no objects and we report Found=false.
func (d *DataURLAction) ResumeURL(ctx context.Context, req ResumeURLRequest) (*ResumeURLResult, error) {
	if req.TrainingRunID == "" {
		return nil, fmt.Errorf("training_run_id is required")
	}

	prefix := fmt.Sprintf(checkpointPrefixTemplate, req.TrainingRunID)
	objs, err := d.storage.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("list checkpoints under %s: %w", prefix, err)
	}

	// latestCheckpointKey works on plain keys (pure/testable); adapt ObjectInfo.
	keys := make([]string, 0, len(objs))
	for _, o := range objs {
		keys = append(keys, o.Key)
	}

	key, index, found := latestCheckpointKey(keys, prefix)
	if !found {
		d.logger.Info("no checkpoints found for run; fresh start",
			zap.String("training_run_id", req.TrainingRunID),
			zap.String("prefix", prefix),
			zap.Int("objects_seen", len(objs)))
		return &ResumeURLResult{Found: false}, nil
	}

	expiry := req.ExpiryMinutes
	if expiry <= 0 {
		expiry = defaultResumeURLExpiryMin
	}
	url, err := d.storage.GetPresignedURL(ctx, key, expiry)
	if err != nil {
		return nil, fmt.Errorf("presign GET %s: %w", key, err)
	}

	d.logger.Info("resume checkpoint selected",
		zap.String("training_run_id", req.TrainingRunID),
		zap.String("key", key),
		zap.Int("index", index))

	return &ResumeURLResult{
		Found:        true,
		PresignedURL: url,
		Key:          key,
		Index:        index,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(expiry) * time.Minute).Format(time.RFC3339),
	}, nil
}

// latestCheckpointKey scans keys directly under the run's checkpoint prefix,
// parses the trailing ckpt-<N>.tar.gz index from each, and returns the key with
// the highest index. found=false if none match. The ckpt-<N>.tar.gz shape MUST
// match compute_checkpoint_keys' default template. Pure (testable without B2).
func latestCheckpointKey(keys []string, prefix string) (key string, index int, found bool) {
	const namePrefix = "ckpt-"
	const nameSuffix = ".tar.gz"
	best := -1
	for _, k := range keys {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		name := k[len(prefix):]
		if strings.Contains(name, "/") {
			continue // nested deeper than the prefix — not a checkpoint object
		}
		if !strings.HasPrefix(name, namePrefix) || !strings.HasSuffix(name, nameSuffix) {
			continue
		}
		numStr := name[len(namePrefix) : len(name)-len(nameSuffix)]
		n, err := strconv.Atoi(numStr)
		if err != nil || n < 0 {
			continue
		}
		if n > best {
			best = n
			key = k
		}
	}
	if best < 0 {
		return "", 0, false
	}
	return key, best, true
}

// handlePrepareResumeURL lists the run's checkpoints, picks the latest, and
// returns a presigned GET (or found=false if none). Mirrors handlePrepareArtefactURL's
// parse + response pattern.
func (a *Adapter) handlePrepareResumeURL(
	body map[string]interface{},
	reqHeaders map[string]string,
	replyToTopic string,
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic on prepare_resume_url request — cannot send response")
		return
	}
	if a.dataURLAction == nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_resume_url",
			"storage_unavailable",
			"object storage not configured (TRAINING_BUCKET / B2 credentials missing at startup)",
			"error_unrecoverable", l)
		return
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_resume_url",
			"invalid_request_body", fmt.Sprintf("could not re-marshal body: %v", err),
			"error_unrecoverable", l)
		return
	}
	var req ResumeURLRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_resume_url",
			"invalid_request_body", fmt.Sprintf("could not unmarshal into ResumeURLRequest: %v", err),
			"error_unrecoverable", l)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := a.dataURLAction.ResumeURL(ctx, req)
	if err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "prepare_resume_url",
			"resume_presign_failed", err.Error(), "error_unrecoverable", l)
		return
	}

	a.sendSuccessResponse(reqHeaders, replyToTopic, "prepare_resume_url",
		map[string]interface{}{
			"found":         result.Found,
			"presigned_url": result.PresignedURL,
			"key":           result.Key,
			"index":         result.Index,
			"expires_at":    result.ExpiresAt,
		}, l)
}
