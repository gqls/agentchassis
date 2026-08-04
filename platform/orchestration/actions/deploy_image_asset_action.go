// FILE: platform/orchestration/actions/deploy_image_asset_action.go
// Deploys an image asset: downloads from storage, optimizes, commits to git

package actions

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// DeployImageAssetInputSpec defines the standard input contract.
//
// New callers (asset-deployer) use input_fields: ["s3_uri", "purpose", "domain", "asset_key"]
// Existing callers (pageflow-builder) use config keys: "uri_field", "domain_field", "purpose"
// The Deprecated map bridges old config keys to new field names.
//
// `deploy_path` and its deprecated alias `deploy_path_field` were REMOVED
// 2026-08-04 (bugs_open/179 finding A): the deploy path is derived from
// (asset_key, purpose), not chosen by the caller. An explicit value now draws a
// refusal — see the guard in DeployImageAssetAction. They are absent from the
// spec deliberately, so that with CheckConfig: true a config still carrying
// either key is reported by the offline config-key audit as unrecognised: a
// second sensor pointing at the same mistake as the refusal.
var DeployImageAssetInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"domain"},
	Optional:    []string{"s3_uri", "purpose", "asset_id", "asset_key"},
	Defaults: map[string]interface{}{
		"purpose": "hero",
	},
	Deprecated: map[string]string{
		"uri_field":     "s3_uri",
		"domain_field":  "domain",
		"purpose_field": "purpose",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("deploy_image_asset", DeployImageAssetInputSpec)
}

// DeployImageAssetAction downloads an image from storage, optimizes it, and commits via git
//
// Inputs (via ActionInputSpec):
//   - domain (required): site domain for git commit
//   - s3_uri (optional): storage URI — if not provided, falls back to findStorageURI lookup
//   - purpose (optional, default "hero"): image purpose — controls resize dimensions via ImagePurposes
//   - deploy_path: NO LONGER HONOURED. An explicit value draws a refusal result
//     (bugs_open/179 finding A) — see the guard below the brand-head one. The
//     output path is derived from (asset_key, purpose) by the one function in
//     platform/storage that every reader resolves through.
//
// The name of that function is deliberately not spelled with its call syntax
// anywhere above the refusal guard: the guard's ordering test anchors on the
// FIRST occurrence of the call token, and a mention in this comment would sit
// before the guard and fail it. Nothing in the guard's own comment names it either.
func DeployImageAssetAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "deploy_image_asset"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Extract inputs using standard pattern (checklist: ActionInputSpec)
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		config,
		DeployImageAssetInputSpec,
		logger,
	)
	if err != nil {
		// Don't hard-fail — fall back to legacy extraction for backward compat
		logger.Warn("ExtractActionInputs returned error, falling back to legacy extraction",
			zap.Error(err))
		inputs = &datahelpers.ActionInputs{Values: map[string]interface{}{}}
	}

	// Resolve purpose
	purpose := inputs.Get("purpose")
	if purpose == "" {
		if p, ok := config["purpose"].(string); ok && p != "" {
			// Legacy: static purpose in config (pageflow-builder pattern)
			purpose = p
		} else {
			purpose = "hero"
		}
	}

	// Phase 2E: resolve asset_key (optional). When present and != purpose,
	// it drives a per-variant deploy_path below so multi-image purposes
	// don't overwrite each other (e.g. hero_about.jpg alongside hero.jpg).
	// Pre-Phase-2E callers don't pass asset_key; the Phase 2E branch is
	// skipped and existing logo/hero deploys keep their default paths.
	//
	// Resolution priority:
	//   1. inputs.Get("asset_key")   — via input_fields config
	//   2. config["asset_key"]       — literal string
	//   3. config["asset_key_field"] — JSONPath into collected_data
	assetKey := inputs.Get("asset_key")
	if assetKey == "" {
		if k, ok := config["asset_key"].(string); ok && k != "" {
			assetKey = k
		}
	}
	if assetKey == "" {
		if kf, ok := config["asset_key_field"].(string); ok && kf != "" {
			assetKey = datahelpers.ExtractNestedFieldString(params.CollectedData, kf)
		}
	}

	// REFUSE a brand-head purpose. This action is not the writer for favicon /
	// og_card: derive_brand_head_assets_action composes them from the site logo
	// and commits them under fixed names. Deploying an arbitrary image to those
	// paths would replace a site's real favicon or social card.
	//
	// WHY THIS IS A REFUSAL AND NOT A COMMENT (bugs_open/179 finding B). Once
	// this action resolves paths through storage.DeployedAssetPath, a brand-head
	// purpose lands on the SAME path the deriver publishes — so what used to be
	// an inert orphan (`og_card.png`, referenced by nothing) becomes an
	// overwrite of the live artefact, committed to git before any provenance or
	// lock guard runs. The path unification is right; shipping it without this
	// refusal would not be.
	//
	// IT IS REACHABLE, which is why it ships in the same image as the change
	// that makes it dangerous. Measured 2026-08-02: 11 open `undeployed_asset`
	// items carry purpose favicon/og_card with NO `mode`, two of them at status
	// `detected`. asset-deployer's `check_mode` only diverts `mode=brand_head`,
	// so those fall through to this step. They predate the bugs_open/142 fix
	// that stopped check_undeployed_assets raising them, and nothing sweeps a
	// queue for items whose defining predicate has since changed.
	//
	// A refusal, not an error — "refusals-as-results per house style": a refused
	// action completes the workflow and reports why, so the work item resolves
	// instead of erroring the orchestration or retrying for ever against a guard
	// that will never pass it.
	//
	// The RESULT SHAPE is the sibling's, not a new one. ingest_staged_asset (the
	// same asset-deployer agent, fourth mode) refuses with the action's own
	// success flag false plus a `reason` — `{"ingested": false, ..., "reason": …}`
	// — and this action already declines the no-storage-URI case as
	// `{"deployed": false, "skipped": true, "reason": …}`. An earlier draft added
	// a `"refused": true` key on top; the council's reuse seat asked whether that
	// matched a convention or invented one, and it invented one — it was the only
	// occurrence of the key in the platform. Dropped. The reason string is what
	// distinguishes the two declines, exactly as it does in the sibling.
	if storage.IsBrandHeadPurpose(purpose) {
		logger.Warn("refusing to deploy a brand-head purpose — this action is not its writer",
			zap.String("purpose", purpose),
			zap.String("asset_key", assetKey),
			zap.String("published_path", storage.BrandHeadAssetPaths[purpose]),
			zap.String("writer", "derive_brand_head_assets"))
		return map[string]interface{}{
			"deployed": false,
			"skipped":  true,
			"reason": fmt.Sprintf("refused: purpose %q is a brand-head artefact published at %s by "+
				"derive_brand_head_assets, not by this action; re-derive it (mode=brand_head) "+
				"instead of deploying over it", purpose, storage.BrandHeadAssetPaths[purpose]),
		}, nil
	}

	// REFUSE an explicit deploy_path. Until 2026-08-04 this input overrode the
	// shared path derivation outright: the file was committed wherever the caller
	// said, while every reader still resolved (asset_key, purpose) through the one
	// derivation in platform/storage — the exact writer/reader drift
	// bugs_closed/168 removed, reintroduced through a supported input
	// (bugs_open/179 finding A).
	//
	// Recording the override on the asset row further down does NOT mend it: all
	// six readers DERIVE the path and none reads that row, so a recorded override
	// is still a path no reader resolves. Teaching a reader to prefer a recorded
	// path is bugs_open/152 and /155's seam, not this one.
	//
	// DELETED rather than gated behind an opt-in field. The owner ruling of
	// 2026-08-02 (RFC_010) makes new authority on a shared seam an opt-in field
	// when the unsafe branch is safe *iff callers behave*; this branch is unsafe
	// however disciplined the caller is, because the readers can never see the
	// caller's choice — so no caller-side field can license it. Measured before
	// removal: zero deploy_path VALUES in open work-item specs, in active agent
	// definitions and in all orchestration history. A caller needing a path the
	// derivation cannot express should widen the derivation in platform/storage,
	// so every reader moves with the writer.
	//
	// Only EXPLICIT intent draws the refusal: the step's own config keys and the
	// orchestration's input_data. inputs.Get("deploy_path") is deliberately NOT
	// consulted — ExtractActionInputs hunts declared fields through the whole of
	// collected_data with a depth-20 recursive search, so a stray key in a nested
	// step result would turn a legitimate deploy into a refusal. An over-strict
	// guard is a worse bug than the inert key it chases. A value only that search
	// can find was never an instruction: it is ignored, and the derived path wins.
	//
	// A refusal, not an error, for the sibling's reason above: the work item
	// resolves with a reason instead of retrying against a guard that will never
	// pass it. Same result shape too, and no new key.
	deployPath, deployPathSource := "", ""
	if dp, ok := config["deploy_path"].(string); ok && dp != "" {
		deployPath, deployPathSource = dp, "config.deploy_path"
	} else if dpf, ok := config["deploy_path_field"].(string); ok && dpf != "" {
		deployPath, deployPathSource = dpf, "config.deploy_path_field"
	} else if dp := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.deploy_path"); dp != "" {
		deployPath, deployPathSource = dp, "input_data.deploy_path"
	}
	if deployPath != "" {
		logger.Warn("refusing an explicit deploy_path — the deploy path is derived, not chosen",
			zap.String("source", deployPathSource),
			zap.String("deploy_path", deployPath),
			zap.String("derived_instead", storage.DeployedWebPath(assetKey, purpose)))
		return map[string]interface{}{
			"deployed": false,
			"skipped":  true,
			"reason": fmt.Sprintf("refused: deploy_path %q (%s) would publish a file no reader can "+
				"resolve — the deploy path is derived from (asset_key, purpose), identically for the "+
				"writer and every reader (bugs_open/179); drop the input and set asset_key, or widen "+
				"the shared derivation in platform/storage", deployPath, deployPathSource),
		}, nil
	}

	// Resolve storage URI
	// Priority: inputs.Get("s3_uri") (from input_fields) → findStorageURI (legacy lookup)
	storageURI := inputs.Get("s3_uri")
	if storageURI == "" {
		storageURI = findStorageURI(params.CollectedData, config, purpose, logger)
	}
	// DB fallback: resolve from asset_id → site content_data → {purpose}_uri
	if storageURI == "" && params.DB != nil {
		assetID := inputs.Get("asset_id")
		if assetID != "" {
			storageURI = resolveStorageURIFromAsset(ctx, params.DB, assetID, purpose, logger)
		}
	}
	if storageURI == "" {
		logger.Warn("No storage URI found for image asset",
			zap.String("purpose", purpose),
			zap.Int("collected_data_key_count", len(params.CollectedData)))
		return map[string]interface{}{
			"deployed": false,
			"skipped":  true,
			"reason":   "no storage URI found for " + purpose,
		}, nil
	}

	// Resolve domain
	// Priority: inputs.Get("domain") (from input_fields) → findDomain (legacy lookup)
	domain := inputs.Get("domain")
	if domain == "" {
		domain = findDomain(params.CollectedData, config)
	}
	if domain == "" {
		return nil, fmt.Errorf("domain not found")
	}

	// Get storage client
	if params.StorageClient == nil {
		return nil, fmt.Errorf("storage client not available")
	}

	s3Client, ok := params.StorageClient.(*storage.S3Client)
	if !ok {
		return nil, fmt.Errorf("storage client is not S3Client type")
	}

	logger.Info("Deploying image asset",
		zap.String("purpose", purpose),
		zap.String("storage_uri", storageURI),
		zap.String("domain", domain))

	// Download, optimize, and prepare paths using storage package
	processed, err := s3Client.DownloadOptimizeAndPrepare(ctx, storageURI, purpose, logger)
	if err != nil {
		logger.Error("Failed to process image", zap.Error(err))
		return map[string]interface{}{
			"deployed": false,
			"error":    err.Error(),
		}, nil
	}

	// Where the file goes. storage.DeployedAssetPath is THE derivation — the
	// same function the render-time resolvers and the discovery checks call —
	// so the path this action commits to and the path a page references cannot
	// drift (bugs_open/168). It subsumes what used to be spelled out here as
	// the "Phase 2E" branch:
	//
	//   asset_key=hero,       purpose=hero    → assets/images/hero.jpg      (unchanged)
	//   asset_key=hero_about, purpose=hero    → assets/images/hero-about.jpg
	//   asset_key=logo,       purpose=logo    → assets/images/logo.png      (unchanged)
	//   asset_key=og_card,    purpose=og_card → assets/images/og-card.png
	//
	// The last line is the one behaviour change: brand-head artefacts are
	// published by derive_brand_head_assets_action under fixed names, and this
	// action used to derive `og_card.png` for them — a path the site head never
	// references. Nothing routes brand-head work here today (check_undeployed
	// _assets excludes those purposes from the generic half and files
	// needs_brand_head_assets instead), so the change corrects a latent
	// disagreement rather than moving live traffic.
	//
	// DownloadOptimizeAndPrepare has already set Paths to the purpose-derived
	// default; for every non-brand-head purpose this reproduces it exactly.
	//
	// Nothing overrides it any more. The `deploy_path` input that used to replace
	// this result outright was removed 2026-08-04 (bugs_open/179 finding A) and an
	// explicit value is refused above, so this assignment is the ONLY thing that
	// decides where the file lands — which is what makes the derivation total for
	// this writer rather than merely preferred.
	derived := storage.DeployedAssetPath(assetKey, purpose)
	if derived.FilePath != processed.Paths.FilePath {
		logger.Info("derived asset deploy path",
			zap.String("asset_key", assetKey),
			zap.String("purpose", purpose),
			zap.String("default_path", processed.Paths.FilePath),
			zap.String("derived_path", derived.FilePath))
	}
	processed.Paths = derived

	logger.Info("Image processed",
		zap.Int("size_bytes", len(processed.Data)),
		zap.String("content_type", processed.ContentType),
		zap.String("output_path", processed.Paths.FilePath))

	// Build files map for git commit (base64 encoded for binary)
	filesMap := map[string]interface{}{
		processed.Paths.FilePath: map[string]interface{}{
			"content":  base64.StdEncoding.EncodeToString(processed.Data),
			"encoding": "base64",
		},
	}

	// Send to git adapter
	result, err := sendGitCommitRequest(ctx, params, domain, filesMap, purpose, logger)
	if err != nil {
		return nil, err
	}

	// Add image URL to result
	result["image_url"] = processed.Paths.RelativeURL
	result["output_path"] = processed.Paths.FilePath
	result["size_bytes"] = len(processed.Data)

	// Record the deployed local URL on the asset row so resolvers (plan_sections'
	// site_assets source) serve the site-local path instead of the expiring
	// presigned storage URL. The pre-update url's unsigned object path is
	// preserved into storage_path (COALESCE = only when empty), mirroring the
	// w9_04 backfill. Best-effort: a failure here must not fail the deploy.
	if params.DB != nil {
		if assetIDStr := inputs.Get("asset_id"); assetIDStr != "" {
			if assetUUID, parseErr := uuid.Parse(assetIDStr); parseErr == nil {
				if _, execErr := params.DB.ExecContext(ctx, `
					UPDATE assets
					SET storage_path     = COALESCE(storage_path, split_part(url, '?', 1)),
					    storage_provider = COALESCE(storage_provider, 'backblaze'),
					    filename         = $2,
					    url              = $3
					WHERE id = $1
				`, assetUUID, processed.Paths.Filename, processed.Paths.RelativeURL); execErr != nil {
					logger.Warn("deploy_image_asset: failed to record local url on asset",
						zap.String("asset_id", assetIDStr),
						zap.Error(execErr))
				} else {
					logger.Info("deploy_image_asset: recorded local url on asset",
						zap.String("asset_id", assetIDStr),
						zap.String("url", processed.Paths.RelativeURL))
				}
			} else {
				logger.Warn("deploy_image_asset: invalid asset_id for local url record",
					zap.String("asset_id", assetIDStr))
			}
		}
	}

	// Write URL directly to collected_data so it survives the git adapter
	// response overwriting this step's output_field (hero_deployed/logo_deployed).
	// The build loop's input_mapping reads hero_url/logo_url from collected_data
	// and passes them to page-content-writer → BuildRenderContextAction.
	urlKey := purpose + "_url"
	params.CollectedData[urlKey] = processed.Paths.RelativeURL
	if purpose == "logo" {
		params.CollectedData["brand_logo_url"] = processed.Paths.RelativeURL
	}
	logger.Info("Wrote image URL to collected_data",
		zap.String("key", urlKey),
		zap.String("url", processed.Paths.RelativeURL))

	return result, nil
}

// resolveStorageURIFromAsset looks up the s3:// URI for an asset.
//
// Priority 1: site content_data → {purpose}_uri (written by StoreAssetAction)
// Priority 2: asset URL → convert presigned HTTPS to s3:// URI
//
// This makes the asset-deployer self-contained: callers pass asset_id,
// the agent resolves the storage URI itself.
func resolveStorageURIFromAsset(ctx context.Context, db *sql.DB, assetIDStr string, purpose string, logger *zap.Logger) string {
	assetUUID, err := uuid.Parse(assetIDStr)
	if err != nil {
		logger.Warn("resolveStorageURIFromAsset: invalid asset_id",
			zap.String("asset_id", assetIDStr),
			zap.Error(err))
		return ""
	}

	// Priority 1: site content_data has the s3:// URI keyed by purpose
	// StoreAssetAction writes: content_data->>'{purpose}_uri' = 's3://bucket/path'
	uriKey := purpose + "_uri"
	query1 := `
		SELECT s.content_data->>$2
		FROM assets a
		JOIN sites s ON a.site_id = s.id
		WHERE a.id = $1
	`
	var s3URI sql.NullString
	if err := db.QueryRowContext(ctx, query1, assetUUID, uriKey).Scan(&s3URI); err == nil {
		if s3URI.Valid && storage.IsS3URI(s3URI.String) {
			logger.Info("Resolved s3_uri from site content_data via asset_id",
				zap.String("asset_id", assetIDStr),
				zap.String("key", uriKey),
				zap.String("s3_uri", s3URI.String))
			return s3URI.String
		}
	}

	// Priority 2: convert asset's presigned URL to s3:// URI
	query2 := `SELECT url FROM assets WHERE id = $1`
	var assetURL sql.NullString
	if err := db.QueryRowContext(ctx, query2, assetUUID).Scan(&assetURL); err == nil {
		if assetURL.Valid && assetURL.String != "" {
			if converted := storage.PresignedURLToS3URI(assetURL.String); converted != "" {
				logger.Info("Converted asset presigned URL to s3_uri",
					zap.String("asset_id", assetIDStr),
					zap.String("s3_uri", converted))
				return converted
			}
		}
	}

	logger.Warn("resolveStorageURIFromAsset: could not resolve",
		zap.String("asset_id", assetIDStr),
		zap.String("purpose", purpose))
	return ""
}

// presignedURLToS3URI converts a presigned S3/B2 URL to an s3:// URI.
//
// Input:  https://s3.us-east-005.backblazeb2.com/bucket-name/path/to/file.png?X-Amz-...
// Output: s3://bucket-name/path/to/file.png
//
// Returns "" if the URL doesn't match the expected pattern.
func presignedURLToS3URI(presignedURL string) string {
	u, err := url.Parse(presignedURL)
	if err != nil || u.Path == "" {
		return ""
	}

	// Path is /bucket/key — strip leading slash and split
	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return ""
	}

	return fmt.Sprintf("s3://%s/%s", parts[0], parts[1])
}

// findStorageURI looks for the storage URI in multiple locations
func findStorageURI(collectedData map[string]interface{}, config map[string]interface{}, purpose string, logger *zap.Logger) string {
	// Priority 1: Config-specified field
	if uriField, ok := config["uri_field"].(string); ok && uriField != "" {
		if uri := datahelpers.ExtractNestedFieldString(collectedData, uriField); uri != "" {
			logger.Debug("Found URI via config uri_field", zap.String("path", uriField))
			return uri
		}
	}

	// Priority 2: {purpose}_uri (set by StoreAssetAction)
	if uri := datahelpers.ExtractNestedFieldString(collectedData, purpose+"_uri"); uri != "" {
		logger.Debug("Found URI at purpose_uri", zap.String("path", purpose+"_uri"))
		return uri
	}

	// Priority 3: {purpose}_result.image_uri pattern
	resultField := purpose + "_result.image_uri"
	if uri := datahelpers.ExtractNestedFieldString(collectedData, resultField); uri != "" {
		logger.Debug("Found URI at result.image_uri", zap.String("path", resultField))
		return uri
	}

	// Priority 4: Deep nested from image-generator agent response
	// Pattern: {purpose}_result.response.generate.response.image_uri
	deepField := purpose + "_result.response.generate.response.image_uri"
	if uri := datahelpers.ExtractNestedFieldString(collectedData, deepField); uri != "" {
		logger.Debug("Found URI at deep nested path", zap.String("path", deepField))
		return uri
	}

	// Priority 5: Slightly less nested pattern
	// Pattern: {purpose}_result.response.image_uri
	lessDeepField := purpose + "_result.response.image_uri"
	if uri := datahelpers.ExtractNestedFieldString(collectedData, lessDeepField); uri != "" {
		logger.Debug("Found URI at response.image_uri", zap.String("path", lessDeepField))
		return uri
	}

	// Priority 6: Check generate_hero_image step directly (step name pattern)
	generateStepField := "generate_" + purpose + "_image.response.generate.response.image_uri"
	if uri := datahelpers.ExtractNestedFieldString(collectedData, generateStepField); uri != "" {
		logger.Debug("Found URI at generate step", zap.String("path", generateStepField))
		return uri
	}

	// Priority 7: {purpose}_stored.asset_url if it's a storage URI
	storedField := purpose + "_stored.asset_url"
	if url := datahelpers.ExtractNestedFieldString(collectedData, storedField); storage.IsS3URI(url) {
		logger.Debug("Found URI at stored.asset_url", zap.String("path", storedField))
		return url
	}

	// Log what we searched for debugging
	logger.Debug("Storage URI not found, searched paths",
		zap.String("purpose", purpose),
		zap.Strings("searched", []string{
			purpose + "_uri",
			purpose + "_result.image_uri",
			purpose + "_result.response.generate.response.image_uri",
			purpose + "_result.response.image_uri",
			"generate_" + purpose + "_image.response.generate.response.image_uri",
			purpose + "_stored.asset_url",
		}))

	return ""
}

// findDomain extracts domain from collected data
func findDomain(collectedData map[string]interface{}, config map[string]interface{}) string {
	// Try config-specified field
	domainField := "site_record.domain"
	if df, ok := config["domain_field"].(string); ok && df != "" {
		domainField = df
	}

	if domain := datahelpers.ExtractNestedFieldString(collectedData, domainField); domain != "" {
		return domain
	}

	// Fallback to input_data.domain
	return datahelpers.ExtractNestedFieldString(collectedData, "input_data.domain")
}

// sendGitCommitRequest sends the commit request to git-adapter
func sendGitCommitRequest(ctx context.Context, params ActionParams, domain string, filesMap map[string]interface{}, purpose string, logger *zap.Logger) (map[string]interface{}, error) {
	if params.Producer == nil {
		return nil, fmt.Errorf("kafka producer not available")
	}

	// Get execution context
	correlationID := params.ExecutionContext.CorrelationID
	orchestrationID := params.ExecutionContext.OrchestrationID
	clientID := params.ExecutionContext.ClientID
	responsesTopic := params.ExecutionContext.ResponsesTopic

	// Fallback for responses topic
	if responsesTopic == "" {
		if alt, ok := params.CollectedData["__my_responses_topic__"].(string); ok && alt != "" {
			responsesTopic = alt
		}
	}
	if responsesTopic == "" {
		if parent, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parent != "" {
			responsesTopic = parent
		}
	}

	newRequestID := uuid.New().String()

	// Same deploy-target resolution as GitCommitAction (incl. the sites-table fallback
	// for workflows without a loaded site record). Without this the images would
	// still go to "sites" while the pages went to the site's own repo — a split brain
	// where a VM-hosted site's pages land on the box but its hero/logo do not.
	repoName := resolveGitRepoNameDB(ctx, params.DB, params.StepConfig.Config, params.CollectedData, domain, logger)

	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":    correlationID,
			"orchestration_id":  orchestrationID,
			"client_id":         clientID,
			"step_name":         params.ExecutionContext.StepName,
			"request_id":        newRequestID,
			"message_type":      "request",
			"sender_agent_type": params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":   orchestrationID,
			"sender_pod_name":   params.ExecutionContext.Sender.PodName,
			"responses_topic":   responsesTopic,
		},
		"body": map[string]interface{}{
			"action": "commit",
			"data": map[string]interface{}{
				"repo_name":      repoName,
				"domain":         domain,
				"files":          filesMap,
				"commit_message": fmt.Sprintf("Deploy %s image for %s", purpose, domain),
			},
		},
	}

	// Serialize and send
	requestBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal git request: %w", err)
	}

	adapterTopic := "system.adapter.git.requests"
	headers := map[string]string{
		"correlation_id":   correlationID,
		"orchestration_id": orchestrationID,
		"request_id":       newRequestID,
		"message_type":     "request",
	}

	err = params.Producer.Produce(ctx, adapterTopic, headers, []byte(correlationID), requestBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send git commit request: %w", err)
	}

	logger.Info("Image deploy request sent to git adapter",
		zap.String("request_id", newRequestID),
		zap.String("domain", domain))

	return map[string]interface{}{
		"deployed":       true,
		"await_response": true,
		"request_id":     newRequestID,
		"purpose":        purpose,
		"domain":         domain,
	}, nil
}
