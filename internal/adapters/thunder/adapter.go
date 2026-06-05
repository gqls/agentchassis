// FILE: internal/adapters/thunder/adapter.go
//
// Thunder Compute Adapter.
//
// Status: Phase 3.4 — provision_instance and decommission_instance are
// implemented. The reaper (3.5) lands next. Currently:
//
//   - provision_instance: pre-check, keypair generation, Thunder API create,
//     k8s Secret persistence, WaitForRunning poll, thunder_instances INSERT
//     with retry, compensating cleanup on failure.
//   - decommission_instance: row lookup, atomic state transition to
//     'decommissioning', Thunder API delete (idempotent), Secret delete
//     (idempotent), cost computation, finalization to 'decommissioned'.
//   - all other actions: return error_unrecoverable / not_implemented.
//
// See FOCUS_adapter_design.md for the canonical pattern this follows.
// See 013/033_thunder_adapter_design.md for the design of the full adapter.

package thunder

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/internal/adapters/thunder/api"
	"github.com/gqls/agentchassis/internal/adapters/thunder/ssh"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/storage"
)

// Adapter is the long-running thunder-adapter service.
type Adapter struct {
	ctx       context.Context
	cancel    context.CancelFunc
	cfg       *config.ServiceConfig
	logger    *zap.Logger
	consumer  *kafka.Consumer
	producer  kafka.Producer
	db        *sql.DB
	adapterID uuid.UUID

	requestsTopic string

	// Phase 3 dependencies — wired in NewAdapter.
	thunderAPI         *api.Client
	secretMgr          *ssh.SecretManager
	provisionAction    *ProvisionAction
	decommissionAction *DecommissionAction

	// Phase 4 dependency — presigned-URL actions (B2 dataset/artefact).
	// nil if object storage isn't configured (TRAINING_BUCKET unset); the
	// prepare_*_url actions then return a clear error rather than panicking.
	dataURLAction *DataURLAction

	// Phase 4 dependency — SSH exec actions (ssh_exec, ssh_get_status).
	// Wired unconditionally (needs only db + secretMgr, both always present).
	sshExecAction *SSHExecAction

	healthServer *http.Server
	shutdownOnce sync.Once
	shutdownWg   sync.WaitGroup
}

// NewAdapter wires up dependencies and returns an Adapter ready to Run.
// Returns an error if any dependency can't be initialised.
// Cleanup convention: every error path closes everything successfully
// opened before it (no defer magic; explicit).
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
	adapterCtx, cancel := context.WithCancel(ctx)

	// ── 1. Topics + consumer group (env override or default) ──
	requestsTopic := os.Getenv("REQUESTS_TOPIC")
	if requestsTopic == "" {
		requestsTopic = "system.adapter.thunder.requests"
	}
	consumerGroup := os.Getenv("CONSUMER_GROUP")
	if consumerGroup == "" {
		consumerGroup = "thunder.adapter.group"
	}

	adapterID, _ := uuid.NewUUID()

	// ── 2. Kafka consumer ──
	consumer, err := kafka.NewConsumer(
		cfg.Infrastructure.KafkaBrokers, requestsTopic, consumerGroup, logger,
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	// ── 3. Kafka producer ──
	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// ── 4. Postgres connection (owns thunder_instances, thunder_config) ──
	dbDSN, err := buildClientsDBDSN(cfg)
	if err != nil {
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to build db DSN: %w", err)
	}
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	if err := db.PingContext(adapterCtx); err != nil {
		db.Close()
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Sanity check — confirm thunder_config singleton exists (migration 025)
	var capCheck float64
	err = db.QueryRowContext(adapterCtx,
		"SELECT daily_cap_usd FROM thunder_config WHERE singleton = 'X'",
	).Scan(&capCheck)
	if err != nil {
		db.Close()
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("thunder_config row not found — apply migration 025 first: %w", err)
	}
	logger.Info("Thunder adapter db connected", zap.Float64("daily_cap_usd", capCheck))

	// ── 5. Thunder Compute API client ──
	// Env vars take precedence (deployment can override per environment);
	// fall back to the default URL if THUNDER_API_URL is unset.
	thunderAPIBaseURL := os.Getenv("THUNDER_API_URL")
	if thunderAPIBaseURL == "" {
		thunderAPIBaseURL = "https://api.thundercompute.com:8443/v1"
	}
	thunderAPIKey := os.Getenv("THUNDER_COMPUTE_API_KEY")
	if thunderAPIKey == "" {
		db.Close()
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("THUNDER_COMPUTE_API_KEY env var is empty — Thunder API auth will fail")
	}
	thunderAPIClient := api.NewClient(thunderAPIBaseURL, thunderAPIKey, logger)

	// ── 6. k8s Secret manager (in-cluster) ──
	// Namespace defaults to ai-persona-system to match the pod's own
	// namespace; THUNDER_SSH_NAMESPACE overrides if needed for testing.
	sshNamespace := os.Getenv("THUNDER_SSH_NAMESPACE")
	if sshNamespace == "" {
		sshNamespace = "ai-persona-system"
	}
	secretMgr, err := ssh.NewInClusterSecretManager(sshNamespace, logger)
	if err != nil {
		db.Close()
		producer.Close()
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to create ssh secret manager: %w", err)
	}

	// ── 7. Action handlers — wire API + Secrets + DB ──
	provisionAction := NewProvisionAction(thunderAPIClient, secretMgr, db, logger)
	decommissionAction := NewDecommissionAction(thunderAPIClient, secretMgr, db, logger)

	// ── 7b. Phase 4: object-storage client for presigned dataset/artefact URLs ──
	// Optional — mirrors the agent.go pattern: build only if a training bucket
	// is configured; don't fail startup if it isn't (provision/decommission
	// still work; only prepare_*_url need it). The training dataset/artefact
	// bucket is separate from the image bucket.
	var dataURLAction *DataURLAction
	trainingBucket := os.Getenv("TRAINING_BUCKET")
	if trainingBucket == "" {
		// "personae-model-training" is the REAL bucket holding training data
		// (confirmed against the live B2 bucket list 2026-05-23). NOTE: the
		// training-data-preparer agent_def lists s3_bucket="finetuning", but no
		// such bucket exists — that value is stale/logical. The data is in
		// personae-model-training (last written 2026-05-10, matches iter_0 prep).
		trainingBucket = "personae-model-training"
	}
	storageCfg := config.ObjectStorageConfig{
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Bucket:          trainingBucket,
		AccessKeyEnvVar: "B2_APPLICATION_KEY_ID",
		SecretKeyEnvVar: "B2_APPLICATION_KEY",
	}
	if s3Client, sErr := storage.NewS3Client(adapterCtx, storageCfg, *logger); sErr != nil {
		logger.Warn("Object storage client not initialised — prepare_dataset_url/prepare_artefact_url will be unavailable",
			zap.Error(sErr))
	} else {
		dataURLAction = NewDataURLAction(s3Client, logger)
		logger.Info("Object storage client initialised for presigned URLs",
			zap.String("training_bucket", trainingBucket),
			zap.String("endpoint", storageCfg.Endpoint))
	}

	// ── 7c. Phase 4: SSH exec actions (ssh_exec, ssh_get_status) ──
	// Needs only db + secretMgr (both always present), so always wired.
	sshExecAction := NewSSHExecAction(db, secretMgr, logger)

	a := &Adapter{
		ctx:                adapterCtx,
		cancel:             cancel,
		cfg:                cfg,
		logger:             logger.With(zap.String("component", "thunder-adapter")),
		consumer:           consumer,
		producer:           producer,
		db:                 db,
		adapterID:          adapterID,
		requestsTopic:      requestsTopic,
		thunderAPI:         thunderAPIClient,
		secretMgr:          secretMgr,
		provisionAction:    provisionAction,
		decommissionAction: decommissionAction,
		dataURLAction:      dataURLAction,
		sshExecAction:      sshExecAction,
	}

	logger.Info("Thunder adapter initialized",
		zap.Strings("kafka_brokers", cfg.Infrastructure.KafkaBrokers),
		zap.String("requests_topic", requestsTopic),
		zap.String("consumer_group", consumerGroup),
		zap.String("adapter_id", adapterID.String()),
		zap.String("thunder_api_url", thunderAPIBaseURL),
		zap.String("ssh_namespace", sshNamespace),
	)

	return a, nil
}

// Run is the main message processing loop. Returns nil on graceful shutdown.
// Sequential by design — no `go a.handleMessage(msg)`.
func (a *Adapter) Run() error {
	a.logger.Info("Thunder adapter starting message processing",
		zap.String("topic", a.requestsTopic))

	a.shutdownWg.Add(1)
	defer a.shutdownWg.Done()

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("Shutdown signal received, stopping message processing")
			return nil
		default:
			consumeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			msg, err := a.consumer.FetchMessage(consumeCtx)
			cancel()

			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					continue
				}
				select {
				case <-a.ctx.Done():
					return nil
				default:
				}
				a.logger.Error("Failed to fetch message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
			a.handleMessage(msg)
		}
	}
}

// handleMessage parses one inbound request and routes it by action.
func (a *Adapter) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)
	l := a.logger.With(
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("request_id", headers["request_id"]),
	)

	// Parse the standard envelope
	var envelope map[string]interface{}
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		l.Error("Failed to unmarshal message", zap.Error(err))
		a.commit(msg)
		return
	}

	body, _ := envelope["body"].(map[string]interface{})
	if body == nil {
		body = make(map[string]interface{})
	}
	action, _ := body["action"].(string)
	replyToTopic, _ := body["reply_to_topic"].(string)

	l.Info("Received request",
		zap.String("action", action),
		zap.String("reply_to_topic", replyToTopic),
	)

	// Phase 3 dispatch: provision_instance is implemented; others still
	// return not_implemented until later phases.
	switch action {
	case "provision_instance":
		a.handleProvisionInstance(body, headers, replyToTopic, l)
	case "decommission_instance":
		a.handleDecommissionInstance(body, headers, replyToTopic, l)
	case "prepare_dataset_url":
		a.handlePrepareDatasetURL(body, headers, replyToTopic, l)
	case "prepare_artefact_url":
		a.handlePrepareArtefactURL(body, headers, replyToTopic, l)
	case "prepare_object_url":
		a.handlePrepareObjectURL(body, headers, replyToTopic, l)
	case "prepare_resume_url":
		a.handlePrepareResumeURL(body, headers, replyToTopic, l)
	case "ssh_exec":
		a.handleSSHExec(body, headers, replyToTopic, l)
	case "ssh_get_status":
		a.handleSSHGetStatus(body, headers, replyToTopic, l)
	default:
		a.sendErrorResponse(headers, replyToTopic, action,
			"not_implemented",
			"thunder-adapter does not yet implement this action",
			"error_unrecoverable", l,
		)
	}
	a.commit(msg)
}

// sendErrorResponse sends a chassis-shaped error response on the reply topic.
// Includes all the correlation headers the chassis needs to match the response
// to its awaited_requests row.
func (a *Adapter) sendErrorResponse(
	reqHeaders map[string]string,
	replyToTopic string,
	action string,
	errCode string, // short code, e.g. "not_implemented", "thunder_api_unreachable"
	errMsg string,  // human-readable detail
	status string,  // "error_recoverable" or "error_unrecoverable"
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic in request; cannot send response",
			zap.String("action", action))
		return
	}

	respHeaders := buildResponseHeaders(reqHeaders, status, a.cfg.ServiceInfo.Name, a.adapterID)

	body := map[string]interface{}{
		"success": false,
		"action":  action,
		"error":   errCode,
		"detail":  errMsg,
	}
	envelope := map[string]interface{}{
		"headers": respHeaders, // typed struct → is_complete/is_error marshal as JSON bools
		"body":    body,
	}
	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		l.Error("Failed to marshal response envelope", zap.Error(err))
		return
	}

	produceCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.producer.ProduceWithValidation(
		produceCtx,
		replyToTopic,
		respHeaders.toKafkaHeaders(),
		[]byte(reqHeaders["correlation_id"]), // key
		envelopeBytes,                        // value
	); err != nil {
		l.Error("Failed to produce response", zap.Error(err))
		return
	}
	l.Info("Sent error response",
		zap.String("action", action),
		zap.String("status", status),
		zap.String("error_code", errCode),
	)
}

// handleProvisionInstance parses the body, calls ProvisionAction.Execute,
// and sends either a success or error response. Error responses use
// error_recoverable / error_unrecoverable based on the cause (caps and
// auth → unrecoverable; 5xx / timeout → recoverable, chassis can retry).
func (a *Adapter) handleProvisionInstance(
	body map[string]interface{},
	reqHeaders map[string]string,
	replyToTopic string,
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic on provision_instance request — cannot send response")
		return
	}

	// Re-marshal the body subtree into a strongly-typed request.
	// We could parse map[string]interface{} field-by-field, but
	// round-tripping through JSON is simpler and the body is small.
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "provision_instance",
			"invalid_request_body",
			fmt.Sprintf("could not re-marshal body: %v", err),
			"error_unrecoverable", l,
		)
		return
	}
	var req ProvisionInstanceRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "provision_instance",
			"invalid_request_body",
			fmt.Sprintf("could not unmarshal into ProvisionInstanceRequest: %v", err),
			"error_unrecoverable", l,
		)
		return
	}
	// Populate the non-JSON RequestedBy from the message header.
	req.RequestedBy = reqHeaders["sender_agent_type"]

	// Use a generous ctx — provision can take up to ~5 min for WaitForRunning.
	// The action itself wraps WaitForRunning with its own configured timeout;
	// this outer ctx is the upper bound.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := a.provisionAction.Execute(ctx, req)
	if err != nil {
		status := "error_unrecoverable"
		errCode := "provision_failed"
		switch {
		case isProvisionDenial(err):
			errCode = "provision_denied"
			// status stays error_unrecoverable — cap denial isn't retryable
		case isInfrastructureError(err):
			status = "error_recoverable"
			errCode = "infrastructure_error"
		}
		a.sendErrorResponse(reqHeaders, replyToTopic, "provision_instance",
			errCode, err.Error(), status, l,
		)
		return
	}

	// Success path — send shaped response. Fields match what model-trainer's
	// call_launcher input_mapping reads from provisioning_result.* via
	// the .response auto-unwrap in input_mapping.go.
	a.sendSuccessResponse(reqHeaders, replyToTopic, "provision_instance",
		map[string]interface{}{
			"instance_ip":         result.InstanceIP,
			"ssh_user":            result.SSHUser,
			"ssh_key_secret_name": result.SSHKeySecretName,
			"provisioning_id":     result.ProvisioningID,
			"thunder_identifier":  result.ThunderIdentifier,
			"provisioned_at":      result.ProvisionedAt.Format(time.RFC3339),
		}, l)
}

// handleDecommissionInstance parses the body, calls DecommissionAction.Execute,
// and sends a success or error response. The action is idempotent end-to-end
// (rows already decommissioned return success without re-calling the Thunder API).
func (a *Adapter) handleDecommissionInstance(
	body map[string]interface{},
	reqHeaders map[string]string,
	replyToTopic string,
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic on decommission_instance request — cannot send response")
		return
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "decommission_instance",
			"invalid_request_body",
			fmt.Sprintf("could not re-marshal body: %v", err),
			"error_unrecoverable", l,
		)
		return
	}
	var req DecommissionInstanceRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "decommission_instance",
			"invalid_request_body",
			fmt.Sprintf("could not unmarshal into DecommissionInstanceRequest: %v", err),
			"error_unrecoverable", l,
		)
		return
	}

	// Decommission shouldn't take more than ~30s — Thunder API delete is
	// fast (no polling) and the DB UPDATE is trivial. The whole flow
	// stays well inside the chassis-level 120s budget.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := a.decommissionAction.Execute(ctx, req)
	if err != nil {
		status := "error_unrecoverable"
		errCode := "decommission_failed"
		if isInfrastructureError(err) {
			status = "error_recoverable"
			errCode = "infrastructure_error"
		}
		a.sendErrorResponse(reqHeaders, replyToTopic, "decommission_instance",
			errCode, err.Error(), status, l,
		)
		return
	}

	a.sendSuccessResponse(reqHeaders, replyToTopic, "decommission_instance",
		map[string]interface{}{
			"provisioning_id":    result.ProvisioningID,
			"thunder_identifier": result.ThunderIdentifier,
			"cost_usd":           result.CostUSD,
			"decommissioned_at":  result.DecommissionedAt.Format(time.RFC3339),
			"was_already_done":   result.WasAlreadyDone,
		}, l)
}

// sendSuccessResponse builds and produces a success-shaped response on the
// reply topic. Mirrors sendErrorResponse but with success=true, status=complete,
// is_error=false. Action handler is responsible for constructing the body map.
func (a *Adapter) sendSuccessResponse(
	reqHeaders map[string]string,
	replyToTopic string,
	action string,
	body map[string]interface{},
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic on success response — dropping",
			zap.String("action", action))
		return
	}

	respHeaders := buildResponseHeaders(reqHeaders, "complete", a.cfg.ServiceInfo.Name, a.adapterID)

	// Inject action and success flag into the body envelope for callers
	// that branch on body.success (mirroring sendErrorResponse's shape).
	envelopeBody := map[string]interface{}{
		"success": true,
		"action":  action,
	}
	for k, v := range body {
		envelopeBody[k] = v
	}

	envelope := map[string]interface{}{
		"headers": respHeaders, // typed struct → is_complete/is_error marshal as JSON bools
		"body":    envelopeBody,
	}
	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		l.Error("Failed to marshal success envelope", zap.Error(err))
		return
	}

	produceCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.producer.ProduceWithValidation(
		produceCtx,
		replyToTopic,
		respHeaders.toKafkaHeaders(),
		[]byte(reqHeaders["correlation_id"]),
		envelopeBytes,
	); err != nil {
		l.Error("Failed to produce success response", zap.Error(err))
		return
	}
	l.Info("Sent success response", zap.String("action", action))
}

// isProvisionDenial returns true if err originated from a thunder_provision_check
// denial (cap breach, manual pause) rather than an infrastructure failure.
// Used to decide between error_unrecoverable (intended denial) and
// error_recoverable (transient infra). Substring match is fragile but
// adequate while these are the only two paths that produce these strings.
func isProvisionDenial(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "provision denied") ||
		strings.Contains(msg, "provisioning paused")
}

// isInfrastructureError detects transient failures — Thunder 5xx, rate limit,
// or ctx deadline/cancel during WaitForRunning. Caller returns
// error_recoverable so the chassis retry policy applies.
func isInfrastructureError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsServer() || apiErr.IsRateLimit()
	}
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled)
}

// responseHeaders is the TYPED response-header envelope. It mirrors the JSON
// shape of the chassis-side types.ResponseHeaders that the chassis unmarshals
// the reply into (platform/types). The fields the chassis treats as booleans
// — is_complete, is_error — are real Go bools here, so json.Marshal emits them
// as JSON `true`/`false`, not the string `"true"`.
//
// WHY THIS EXISTS (the bug it fixes): the previous implementation built the
// headers as a map[string]string, so is_complete went out as the STRING
// "true". The chassis unmarshals the reply's `headers` object into a typed
// struct with `IsComplete bool`, and string→bool fails:
//
//	json: cannot unmarshal string into Go struct field
//	ResponseHeaders.headers.is_complete of type bool
//
// On that error the chassis response-routing branch returned early and the
// reply was dropped BEFORE ClaimAwaitedRequest — the awaited_requests row sat
// in 'waiting' until timeout, even though the work had fully succeeded.
//
// This is the favoured adapter response pattern (see contracts §"Adapter
// Response Envelope Contract"): a typed struct marshalled into the body, plus
// toKafkaHeaders() for the Kafka message-header arg. It matches the git
// adapter's ResponseHeaders + ToKafkaHeaders() approach.
type responseHeaders struct {
	CorrelationID   string `json:"correlation_id"`
	OrchestrationID string `json:"orchestration_id"`

	// request_id and in_response_to_request_id BOTH carry the ORIGINAL
	// request's request_id (matching git/webscrape). The chassis routes a
	// reply to the coordinator's claim path by matching request_id against an
	// awaited entry; a freshly-generated id would miss.
	RequestID             string `json:"request_id"`
	InResponseToRequestID string `json:"in_response_to_request_id"`

	// message_id: the working git adapter always sets one; absence makes the
	// chassis synthesise it and can mis-route the reply as unsolicited inbound.
	MessageID string `json:"message_id"`

	ClientID    string `json:"client_id"`
	MessageType string `json:"message_type"` // always "response"
	Status      string `json:"status"`       // "complete" | "error_recoverable" | "error_unrecoverable"

	// Real bools — serialise as JSON true/false. This is the fix.
	IsComplete bool `json:"is_complete"`
	IsError    bool `json:"is_error"`

	SenderAgentType      string `json:"sender_agent_type"`
	SenderAgentID        string `json:"sender_agent_id"`
	InResponseToStepName string `json:"in_response_to_step_name,omitempty"`
	InResponseToStepID   string `json:"in_response_to_step_id,omitempty"`
	TimeSent             string `json:"time_sent"`
}

// toKafkaHeaders renders the typed headers as the map[string]string the Kafka
// producer takes for the message-header arg. Booleans become canonical
// "true"/"false" strings here — that's correct for Kafka headers (which are
// always byte strings); the type-sensitive consumer is the JSON body unmarshal,
// which reads the typed struct above, not these. Mirrors git's ToKafkaHeaders().
func (h responseHeaders) toKafkaHeaders() map[string]string {
	return map[string]string{
		"correlation_id":            h.CorrelationID,
		"orchestration_id":          h.OrchestrationID,
		"request_id":                h.RequestID,
		"in_response_to_request_id": h.InResponseToRequestID,
		"message_id":                h.MessageID,
		"client_id":                 h.ClientID,
		"message_type":              h.MessageType,
		"status":                    h.Status,
		"is_complete":               strconv.FormatBool(h.IsComplete),
		"is_error":                  strconv.FormatBool(h.IsError),
		"sender_agent_type":         h.SenderAgentType,
		"sender_agent_id":           h.SenderAgentID,
		"in_response_to_step_name":  h.InResponseToStepName,
		"in_response_to_step_id":    h.InResponseToStepID,
		"time_sent":                 h.TimeSent,
	}
}

// buildResponseHeaders constructs the typed response headers. status drives the
// IsComplete/IsError flags: a "complete" status is a successful completion;
// any error_* status sets IsError.
func buildResponseHeaders(
	reqHeaders map[string]string,
	status string,
	senderType string,
	senderID uuid.UUID,
) responseHeaders {
	isError := status != "complete"
	return responseHeaders{
		CorrelationID:         reqHeaders["correlation_id"],
		OrchestrationID:       reqHeaders["orchestration_id"],
		RequestID:             reqHeaders["request_id"],
		InResponseToRequestID: reqHeaders["request_id"],
		MessageID:             uuid.New().String(),
		ClientID:              reqHeaders["client_id"],
		MessageType:           "response",
		Status:                status,
		IsComplete:            true, // the reply always completes the awaited request
		IsError:               isError,
		SenderAgentType:       senderType,
		SenderAgentID:         senderID.String(),
		InResponseToStepName:  reqHeaders["step_name"],
		InResponseToStepID:    reqHeaders["step_id"],
		TimeSent:              time.Now().UTC().Format(time.RFC3339),
	}
}

// commit acknowledges a message; logs but doesn't error if the commit fails.
func (a *Adapter) commit(msg kafka.Message) {
	if err := a.consumer.CommitMessages(context.Background(), msg); err != nil {
		a.logger.Warn("Failed to commit message", zap.Error(err))
	}
}

// StartHealthServer starts /health (liveness) and /ready (readiness) endpoints.
func (a *Adapter) StartHealthServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"db_unreachable"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	a.healthServer = &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := a.healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Health server error", zap.Error(err))
		}
	}()
	a.logger.Info("Health server started", zap.String("port", port))
}

// Shutdown stops the adapter gracefully. Safe to call multiple times.
func (a *Adapter) Shutdown() {
	a.shutdownOnce.Do(func() {
		a.logger.Info("Thunder adapter shutting down")
		a.cancel()

		if a.healthServer != nil {
			sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = a.healthServer.Shutdown(sctx)
		}

		done := make(chan struct{})
		go func() {
			a.shutdownWg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			a.logger.Warn("Shutdown wait exceeded; closing connections anyway")
		}

		a.consumer.Close()
		a.producer.Close()
		if a.db != nil {
			_ = a.db.Close()
		}
		a.logger.Info("Thunder adapter shutdown complete")
	})
}

// buildClientsDBDSN constructs a Postgres DSN from the ServiceConfig.
func buildClientsDBDSN(cfg *config.ServiceConfig) (string, error) {
	dbc := cfg.Infrastructure.ClientsDatabase
	password := os.Getenv(dbc.PasswordEnvVar)
	if password == "" {
		return "", fmt.Errorf("clients_db password env var %q not set", dbc.PasswordEnvVar)
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbc.Host, dbc.Port, dbc.User, password, dbc.DBName, dbc.SSLMode,
	)
	return dsn, nil
}
