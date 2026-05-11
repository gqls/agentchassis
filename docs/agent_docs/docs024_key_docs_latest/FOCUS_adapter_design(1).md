# FOCUS — Adapter Design and Creation

Canonical guide for building long-running cluster services ("adapters") that wrap external APIs or shared infrastructure. Read this before adding a new adapter or modifying an existing one. Examples drawn from the working adapters in the repository: `git-adapter`, `web-scrape-adapter`, `image-generator-adapter`, `web-search-adapter`, `ollama-adapter`, and `thunder-adapter`.

---

## TL;DR

An adapter is a Go binary deployed as a single-replica (or small-scale-out) `Deployment`. It consumes from a fixed Kafka topic and proxies one external system (Stability AI, Firecrawl, GitHub, Backblaze, Thunder Compute, etc.). It holds the credentials so individual agents don't have to. It's the right pattern when:

- The external system has a single credential boundary (one API key/secret per cluster)
- The work is request/response shaped (not long-running per-VM service)
- Multiple agents would otherwise duplicate the integration logic
- You want one place to put retry, circuit-breaker, rate-limit, and per-call logging

It is the wrong pattern when:
- The work is naturally per-orchestration and isolated (spawn an agent instead)
- The work is short, infrequent, and well-isolated to one agent type (do it inline)
- The "external system" is actually our database or Kafka (those are already shared)

---

## Decision: adapter vs agent vs inline

| Scenario | Pattern |
|---|---|
| Talks to one external API, multiple internal callers want it | **Adapter** (this doc) |
| Long-running per-orchestration work (training, site build) | Spawned agent (Job pod) |
| Short call inside one agent's workflow | Inline action |
| Shared infrastructure with no external API (db, kafka) | No wrapper needed |

When in doubt: if more than one agent type will call the same external API, build an adapter.

---

## Responsibilities

**Adapter owns:**
- Credentials for the external system (held in pod env, never propagated outward)
- Request validation and shape coercion before hitting the external API
- Retry policy with backoff
- Rate limiting / throttle
- Circuit breaker
- Per-call logging and metrics
- Mapping of external responses back to chassis-shape responses
- Database tracking when the work creates persistent state (e.g. `thunder_instances`)

**Adapter does NOT own:**
- Orchestration state (chassis owns that)
- Per-orchestration data lifecycle (the calling workflow owns it)
- The decision of when to call (the orchestrator decides)
- Knowledge of which agent type called it (it's just request/response)

---

## File and package layout

Every adapter follows the same physical layout:

```
cmd/<name>-adapter/main.go              ← entry point: load config, init, signal handling
internal/adapters/<name>/adapter.go     ← struct, NewAdapter, Run, handleMessage, Shutdown
internal/adapters/<name>/<api>_client.go  ← optional: external API client
internal/adapters/<name>/<feature>.go   ← optional: per-action handlers
configs/<name>-adapter.yaml             ← config file mounted as ConfigMap
build/docker/backend/<name>-adapter.dockerfile   ← optional, often shared multi-stage build
deployments/kustomize/services/<name>-adapter/   ← kustomize base + overlays
```

Use a hyphenated name for files and topics; one canonical name throughout (e.g. `web-scrape`, not `webscrape` in some places and `web-scrape` in others). The `internal/adapters/<name>/` package can use a single-word identifier (`webscrape`) since Go packages don't allow hyphens.

---

## The adapter struct

Standard shape, drawn from `git-adapter` and `image-generator-adapter`:

```go
type Adapter struct {
    ctx       context.Context
    cancel    context.CancelFunc
    cfg       *config.ServiceConfig
    logger    *zap.Logger
    consumer  *kafka.Consumer
    producer  kafka.Producer       // interface, not struct
    adapterID uuid.UUID

    // External-system client (per adapter)
    apiClient *MyAPIClient

    // Optional persistent state
    db *sql.DB

    // Optional cluster-side storage for outputs (e.g. images, adapters)
    storageClient storage.Client

    // Throttle / rate limit / circuit breaker
    throttle *throttle.Throttle

    // Topic config
    requestsTopic string

    // Health server + shutdown coordination
    healthServer *http.Server
    shutdownOnce sync.Once
    shutdownWg   sync.WaitGroup
}
```

Drop fields the adapter doesn't need. Most adapters have producer + consumer + apiClient + logger as the core; storage and DB are optional.

---

## NewAdapter pattern

Order matters: each step can fail, and earlier resources must close cleanly on later failure.

```go
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
    adapterCtx, cancel := context.WithCancel(ctx)

    // 1. Resolve topics + consumer group (env overrides config)
    requestsTopic := envOrDefault("REQUESTS_TOPIC", "system.adapter.<name>.requests")
    consumerGroup := envOrDefault("CONSUMER_GROUP", "<name>.adapter.group")

    // 2. Kafka consumer
    consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestsTopic, consumerGroup, logger)
    if err != nil { cancel(); return nil, err }

    // 3. Kafka producer
    producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
    if err != nil { consumer.Close(); cancel(); return nil, err }

    // 4. External API client (with credentials from env)
    apiClient, err := NewMyAPIClient(os.Getenv("MY_API_KEY"), logger)
    if err != nil { producer.Close(); consumer.Close(); cancel(); return nil, err }

    // 5. Optional: DB, storage, throttle, circuit breaker — each with the same
    //    pattern of cleaning up prior resources on failure
    
    return &Adapter{ /* fields */ }, nil
}
```

**Cleanup convention**: every failure path in `NewAdapter` closes everything that was successfully opened before it. There's no `defer` magic; it's manual and explicit. This matches `git-adapter` and `image-generator-adapter` exactly.

---

## Run loop

Adapters have a fixed shape: fetch one message, handle it, loop.

```go
func (a *Adapter) Run() error {
    a.logger.Info("Adapter starting message processing",
        zap.String("topic", a.requestsTopic))

    a.shutdownWg.Add(1)
    defer a.shutdownWg.Done()

    for {
        select {
        case <-a.ctx.Done():
            a.logger.Info("Shutdown signal received")
            return nil   // normal shutdown returns nil, not an error
        default:
            consumeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            msg, err := a.consumer.FetchMessage(consumeCtx)
            cancel()

            if err != nil {
                if err == context.Canceled || err == context.DeadlineExceeded {
                    continue   // normal — no messages or shutdown — re-poll
                }
                select {
                case <-a.ctx.Done():
                    return nil   // shutdown is the real reason for the error
                default:
                }
                a.logger.Error("Failed to fetch message", zap.Error(err))
                time.Sleep(time.Second)   // backoff before retry
                continue
            }
            a.handleMessage(msg)   // sequential by default; only goroutine if you've thought about concurrency
            if a.throttle != nil {
                a.throttle.Wait()
            }
        }
    }
}
```

**Do not `go a.handleMessage(msg)` by default.** Sequential processing is safer:
- Kafka offsets commit cleanly only after the handler finishes
- Concurrent handlers can starve the external API's rate limit
- Throttle.Wait() between handlers is meaningless if they're concurrent

Some adapters (the reasoning agent) do use goroutines deliberately — but the choice should be conscious, not default.

---

## handleMessage pattern

Three things every adapter does on each message:

1. Parse the envelope
2. Dispatch by action
3. Send a response (success or error), commit the message

```go
func (a *Adapter) handleMessage(msg kafka.Message) {
    headers := kafka.HeadersToMap(msg.Headers)
    l := a.logger.With(
        zap.String("correlation_id", headers["correlation_id"]),
        zap.String("request_id", headers["request_id"]),
    )

    // Parse envelope
    var envelope map[string]interface{}
    if err := json.Unmarshal(msg.Value, &envelope); err != nil {
        l.Error("Failed to unmarshal message", zap.Error(err))
        a.commit(msg)   // commit anyway — re-processing won't help a malformed message
        return
    }

    body, _ := envelope["body"].(map[string]interface{})
    if body == nil {
        body = make(map[string]interface{})
    }
    action, _ := body["action"].(string)
    replyToTopic, _ := body["reply_to_topic"].(string)

    l.Info("Received request", zap.String("action", action))

    // Dispatch
    var result interface{}
    var err error
    switch action {
    case "provision_instance":
        result, err = a.handleProvisionInstance(body)
    case "decommission_instance":
        result, err = a.handleDecommissionInstance(body)
    default:
        err = fmt.Errorf("unknown action: %s", action)
    }

    // Respond
    if err != nil {
        a.sendErrorResponse(headers, replyToTopic, action, err, l)
    } else {
        a.sendSuccessResponse(headers, replyToTopic, action, result, l)
    }
    a.commit(msg)
}
```

The chassis side sends messages with `body.action` and `body.reply_to_topic`. The adapter is expected to respond on `reply_to_topic`. Without those, the orchestrator can't route the response.

---

## Sending responses (the bit that's easy to get wrong)

**The Producer interface signature:**

```go
type Producer interface {
    Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error
    ProduceWithValidation(...) error
    Close() error
}
```

Note: `headers` is `map[string]string`, NOT `[]kafka.Header`. `key` is `[]byte`, conventionally the correlation_id. `value` is the marshalled JSON envelope.

**Response headers, split into two tiers:**

**Tier 1 — required by `platform/validation/Validator`** (`ValidateOutgoingMessage` and `ValidateIncomingMessage` enforce these; messages without them get rejected unless `is_error=true`):

| Header | Source | Required by |
|---|---|---|
| `client_id` | pass-through from request | outgoing + incoming |
| `correlation_id` | pass-through from request | outgoing + incoming |
| `orchestration_id` | pass-through from request | outgoing + incoming |
| `sender_agent_type` | adapter name (e.g. `"thunder-adapter"`) | outgoing |
| `step_name` OR `in_response_to_step_name` | latter for responses; pass-through from request's `step_name` | outgoing |

**Escape hatch**: any message with `is_error=true` bypasses both validators. Useful when an adapter can't construct a fully-valid response (e.g. malformed incoming request).

**Tier 2 — needed for the chassis to route the response back to the awaiting orchestration** (validator doesn't check these, but the chassis processor reads them; omit and the orchestration sits in `AWAITING_RESPONSES` forever):

| Header | Source | Why |
|---|---|---|
| `in_response_to_request_id` | original request's `request_id` | chassis uses this to find awaited_requests row |
| `message_type` | `"response"` | distinguishes from new request |
| `status` | `"complete"` / `"error_recoverable"` / `"error_unrecoverable"` | chassis branches on this |
| `is_complete` | `"true"` / `"false"` | for multi-part responses; usually true |
| `is_error` | mirrors status — `"true"` if status starts with `error_` | error-path routing |
| `request_id` | new UUID for this response | new message identity |
| `in_response_to_step_id` | pass-through from request | step matching for nested orchestrations |

**Tier 3 — useful for observability/tracing but not strictly required:**

| Header | Source | Used by |
|---|---|---|
| `sender_agent_id` | adapter's run UUID | trace logging |
| `sender_pod_name` | `os.Getenv("POD_NAME")` | pod-level tracing |
| `sender_agent_version` | `os.Getenv("AGENT_VERSION")` | version tracking |
| `sender_role` | usually `"adapter"` | role-based logging |
| `correlation_name` | pass-through | human-readable trace |
| `time_sent` | now (RFC3339) | latency measurement |
| `fuel_used` | pass-through (or incremented if adapter consumes fuel) | fuel governance |
| `in_response_to_action` | pass-through from request's `action` | trace logging |
| `in_response_to_parent_orchestration_id` | pass-through (for spawn-child responses) | nested orchestration tracking |

Including Tier 3 is recommended in `buildResponseHeaders` because adding them is one line each and they're free insurance against future chassis additions that *do* read them. But missing one won't break anything today.

---

### TODO — tighten validator coverage

The current `platform/validation/Validator.ValidateOutgoingMessage` only enforces the 5 Tier-1 fields. Tier-2 fields (`in_response_to_request_id`, `message_type`, `status`, `is_complete`, `is_error`, `request_id`, `in_response_to_step_id`) are **necessary for the orchestration to advance** but **not validated** — meaning a buggy adapter or action can silently emit a response that passes validation, lands in the chassis, gets logged, and leaves the awaiting orchestration permanently stuck.

Proposal:
- Extend `ValidateOutgoingMessage` (and `ValidateIncomingMessage` where applicable) to require Tier-2 fields when `message_type == "response"`. The `is_error=true` escape hatch should continue to bypass everything for error responses (a partly-broken adapter still needs to be able to signal failure).
- Promote `in_response_to_request_id` in particular: it's the single most common silent-hang cause, and a missing-field check would catch it at the producer rather than after a 30-minute timeout.
- Consider a Tier-2.5 layer where the validator logs a Warn (not Error) for missing recommended fields, giving us visibility without immediate rejection during the transition.

Tracking issue: not yet filed. Add when the next adapter is built.

**The body envelope** of the response message:

```json
{
  "headers": {
    "correlation_id": "...",
    "in_response_to_request_id": "...",
    "request_id": "...",
    "orchestration_id": "...",
    "status": "complete",
    "sender_agent_type": "thunder-adapter",
    "sender_agent_id": "..."
  },
  "body": {
    "success": true,
    "data": { /* action-specific result */ }
  }
}
```

Kafka headers AND body-embedded headers are both set. The body-embedded headers are what the chassis primarily reads; the Kafka headers are for cross-cluster routing and observability.

**Skeleton helpers** worth copying when starting a new adapter:

```go
func (a *Adapter) sendSuccessResponse(reqHeaders map[string]string, replyToTopic, action string, data interface{}, l *zap.Logger) {
    a.sendResponse(reqHeaders, replyToTopic, action, "complete", true, data, "", l)
}

func (a *Adapter) sendErrorResponse(reqHeaders map[string]string, replyToTopic, action string, err error, l *zap.Logger) {
    a.sendResponse(reqHeaders, replyToTopic, action, "error_recoverable", false, nil, err.Error(), l)
}

func (a *Adapter) sendResponse(reqHeaders map[string]string, replyToTopic, action, status string,
    success bool, data interface{}, errMsg string, l *zap.Logger,
) {
    if replyToTopic == "" {
        l.Warn("No reply_to_topic in request; cannot send response")
        return
    }

    respHeaders := map[string]string{
        "correlation_id":            reqHeaders["correlation_id"],
        "orchestration_id":          reqHeaders["orchestration_id"],
        "in_response_to_request_id": reqHeaders["request_id"],
        "request_id":                uuid.New().String(),
        "client_id":                 reqHeaders["client_id"],
        "message_type":              "response",
        "status":                    status,
        "is_complete":               "true",
        "is_error":                  boolStr(!success),
        "sender_agent_type":         a.cfg.ServiceInfo.Name,
        "sender_agent_id":           a.adapterID.String(),
        "in_response_to_step_name":  reqHeaders["step_name"],
        "in_response_to_step_id":    reqHeaders["step_id"],
        "time_sent":                 time.Now().UTC().Format(time.RFC3339),
    }

    bodyMap := map[string]interface{}{
        "success": success,
        "data":    data,
    }
    if !success {
        bodyMap["error"] = errMsg
    }

    envelope := map[string]interface{}{
        "headers": respHeaders,
        "body":    bodyMap,
    }
    envelopeBytes, _ := json.Marshal(envelope)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := a.producer.Produce(ctx, replyToTopic, respHeaders, []byte(reqHeaders["correlation_id"]), envelopeBytes)
    if err != nil {
        l.Error("Failed to produce response", zap.Error(err))
        return
    }
    l.Info("Sent response", zap.String("action", action), zap.String("status", status))
}

func boolStr(b bool) string { if b { return "true" }; return "false" }
```

---

## Health endpoints

Two endpoints, on the port from `cfg.Server.Port` (default 8080):

| Path | Purpose | Implementation |
|---|---|---|
| `/health` | Liveness — am I alive? | Return 200 with `{"status":"ok"}` unconditionally |
| `/ready` | Readiness — can I serve traffic? | Ping the DB and any critical dependency; return 200 on success, 503 on failure |

K8s Deployment configures these as probes. `/health` triggers pod restart on failure; `/ready` removes the pod from Service endpoints. Keep them cheap (<100ms).

---

## Graceful shutdown

```go
func (a *Adapter) Shutdown() {
    a.shutdownOnce.Do(func() {
        a.logger.Info("Adapter shutting down")
        a.cancel()  // signals Run() to exit at next iteration

        // Stop accepting new health requests
        if a.healthServer != nil {
            sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            _ = a.healthServer.Shutdown(sctx)
        }

        // Wait for Run() to exit (up to 10s)
        done := make(chan struct{})
        go func() { a.shutdownWg.Wait(); close(done) }()
        select {
        case <-done:
        case <-time.After(10 * time.Second):
            a.logger.Warn("Shutdown wait exceeded")
        }

        // Close Kafka, DB, etc.
        a.consumer.Close()
        a.producer.Close()
        if a.db != nil { _ = a.db.Close() }
        a.logger.Info("Shutdown complete")
    })
}
```

`sync.Once` makes this safe to call multiple times — the signal handler in `main()` and the test harness can both call it.

---

## Topic naming conventions

Two conventions exist in the codebase. New adapters should use **convention A**:

**Convention A — `system.adapter.<name>.requests`** (used by `git-adapter`, `image-generator-adapter`, `thunder-adapter`):
```
system.adapter.git.requests
system.adapter.image-generator.requests
system.adapter.thunder.requests
```

**Convention B — `system.agent.<name>-adapter.process`** (legacy, used by `web-scrape-adapter` deployment manifest per dev guide line 453):
```
system.agent.webscrape-adapter.process
```

Why two: convention B predates A. New adapters adopt A. Old adapters keep their topic for compatibility but should migrate when convenient.

**Consumer group naming**: `<name>.adapter.group` (e.g. `thunder.adapter.group`). The group must be stable across pod restarts for offset continuity.

**Topic override via env**: `REQUESTS_TOPIC` and `CONSUMER_GROUP` env vars on the Deployment override the YAML defaults. Useful for staging environments that share a cluster.

---

## Config YAML structure

**Use these exact field names** — the Go struct tags expect them:

```yaml
service_info:
  name: "thunder-adapter"          # used as sender_agent_type in responses
  version: "0.1.0"
  environment: "production"        # production | development | staging

server:
  port: "8080"                     # STRING in quotes; reads as cfg.Server.Port
                                   # (some adapters use http_port/grpc_port — that's a different schema, avoid)

logging:                           # NOT `logger:` — Go field is cfg.Logging.Level
  level: "info"                    # debug | info | warn | error
  format: "json"                   # optional

observability:
  tracing_endpoint: "${SERVICE_OBSERVABILITY_TRACING_ENDPOINT:http://jaeger-collector:14268/api/traces}"

infrastructure:
  kafka_brokers:
    - "personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

  # Only include database configs the adapter actually uses
  clients_database:
    host: "pgbouncer.ai-persona-system.svc.cluster.local"
    port: 6432
    user: "clients_user"
    password_env_var: "CLIENTS_DB_PASSWORD"
    db_name: "clients_db"
    sslmode: "disable"             # ONE WORD `sslmode`, not `ssl_mode`

  templates_database: {}           # empty map = not used
  auth_database: {}

  # object_storage only if the adapter writes to S3
  object_storage:
    endpoint: "https://s3.us-east-005.backblazeb2.com"
    region: "us-east-005"
    bucket: "personae-model-training"
    access_key_env_var: "B2_APPLICATION_KEY_ID"
    secret_key_env_var: "B2_APPLICATION_KEY"
    use_ssl: true
    use_path_style: false

custom:                            # adapter-specific settings — free-form map
  adapter_settings:
    consumer_group: "thunder.adapter.group"
    main_topic: "system.adapter.thunder.requests"

  thunder_api:                     # the external system this adapter wraps
    url: "${THUNDER_API_URL:https://api.thundercompute.com/v1}"
    key_env_var: "THUNDER_COMPUTE_API_KEY"

  circuit_breaker:                 # standard knobs (matches existing adapters)
    name: "thunder-api"
    max_requests: 3
    interval_seconds: 60
    timeout_seconds: 60
    consecutive_failures: 5
    failure_ratio: 0.6
```

**Codebase inconsistency to know about**: `image-adapter.yaml` uses `http_port: 8080` + `grpc_port: 9090` + `shutdown_timeout: 5s` under `server:`. That's a different schema with extra fields — works if your Go code parses them. For new adapters, stick with the single-`port` shape from `agent-chassis.yaml` and `web-scrape-adapter.yaml`.

---

## Deployment essentials

A k8s Deployment manifest for an adapter must include:

```yaml
# Image and command
image: docker.io/aqls/<name>-adapter:vX.Y.Z
args: ["--config", "/app/configs/<name>-adapter.yaml"]

# Env from secrets
envFrom:
  - secretRef:
      name: personae-default-secrets    # API keys (ANTHROPIC_, THUNDER_, etc.)
  - secretRef:
      name: personae-platform-secrets   # DB passwords, B2 credentials

# Config mount
volumeMounts:
  - name: config
    mountPath: /app/configs/<name>-adapter.yaml
    subPath: <name>-adapter.yaml
volumes:
  - name: config
    configMap:
      name: <name>-adapter-config

# Health probes
livenessProbe:
  httpGet: {path: /health, port: 8080}
  initialDelaySeconds: 30
  periodSeconds: 30
readinessProbe:
  httpGet: {path: /ready, port: 8080}
  initialDelaySeconds: 10
  periodSeconds: 10

# Resources — adapters are usually small
resources:
  requests: {cpu: 100m, memory: 128Mi}
  limits:   {cpu: 500m, memory: 512Mi}

# Replicas — usually 1 (Kafka consumer group handles failover)
replicas: 1
```

**Service permissions** (if the adapter needs to manage k8s resources, e.g. SSH key Secrets):

```yaml
# Separate ServiceAccount + Role + RoleBinding scoping creation/get/delete
# of just the resources the adapter manages. Avoid `cluster-admin`.
```

---

## Credentials and secrets

Adapters hold credentials so individual agents don't have to. Two rules:

1. **Read credentials from env, never from code.** Set them via `envFrom: secretRef:` in the Deployment manifest.
2. **Never put credentials in the config YAML.** Config holds the env var NAME (`access_key_env_var: "B2_APPLICATION_KEY_ID"`), not the value.

If the adapter needs to forward a temporary credential to an external system (e.g. presigned URL to a Thunder VM), generate it on the fly with short expiry and never log the full URL.

---

## Common mistakes (real examples)

These bit during the thunder-adapter build. Documented so they don't again.

### `Producer.Produce` arguments

The signature is **five arguments**: `(ctx, topic, headers map[string]string, key []byte, value []byte)`. Common mistakes:
- Passing `kafka.MapToHeaders(...)` (a `[]kafka.Header` slice) where `map[string]string` is expected
- Forgetting the `key` parameter (conventionally the correlation_id as bytes)
- Passing 4 args by accident — won't compile, but the error message blames the wrong line

### Response correlation headers

Two failure modes here, and they're separate:

**Validator rejection** — `validation.ValidateOutgoingMessage` only requires 5 fields: `client_id`, `correlation_id`, `orchestration_id`, `sender_agent_type`, and `step_name` (or `in_response_to_step_name`). Missing any of these gets the message rejected unless `is_error=true` (the escape hatch). All adapters get this right by passing through from incoming `reqHeaders`.

**Silent orchestration hang** — even when validation passes, the chassis matches the response back to its awaited_requests row using `in_response_to_request_id`. Missing this header means the chassis logs the response, but the awaiting orchestration sits in `AWAITING_RESPONSES` forever. The validator won't catch this; only your integration test will. Always set `in_response_to_request_id` from the request's `request_id`.

### Status vocabulary

The `status` header is read by the chassis to decide next steps:
- `complete` — success, advance the workflow
- `error_recoverable` — failed but retry might help
- `error_unrecoverable` — failed, don't retry; propagate the failure

Don't invent new values. The chassis has explicit handling only for these.

### Topic naming inconsistency

Two conventions in the codebase (`system.adapter.X.requests` vs `system.agent.X-adapter.process`). New work uses convention A. When in doubt, check what the existing similar adapter uses and follow it; don't mix.

### YAML field names

`logger:` vs `logging:`, `ssl_mode` vs `sslmode`, `http_port` vs `port` — the Go struct tags are picky. Copy from a known-good config (`agent-chassis.yaml`, `git-adapter.yaml`) rather than typing fresh. If the adapter crashes on startup with cryptic config-load errors, this is the first thing to check.

### Closing resources on `NewAdapter` failure

Manual cleanup at every step. Forgetting one means a partial init leaks (open Kafka connections, open DB pool). No `defer` magic — explicit `consumer.Close()` etc. on every error return.

### `go a.handleMessage(msg)` by default

Sequential is the default. Only go-routine if you've explicitly thought about concurrent rate limits, offset commit semantics, and shared state. Most adapters should be sequential.

### Logging credentials

If the external API client logs request bodies for debugging, those bodies may contain bearer tokens or signed URLs. Strip them before logging. Same applies to env var dumps at startup (which exist in some adapter init code) — never log `os.Environ()` raw in production.

---

## Testing a new adapter end-to-end

After deploy:

```bash
# 1. Pod up and ready
kubectl -n ai-persona-system get pods -l app=<name>-adapter
kubectl -n ai-persona-system logs deploy/<name>-adapter | head -50
# Expect: connection logs for kafka, db (if used), and "Adapter initialized" line

# 2. Health endpoints
kubectl -n ai-persona-system port-forward deploy/<name>-adapter 8081:8080 &
curl -s http://localhost:8081/health
curl -s http://localhost:8081/ready

# 3. Smoke message — produce a test request, listen for response
CORRELATION_ID=$(uuidgen)
REQUEST_ID=$(uuidgen)
# Open consumer first in another terminal:
kubectl -n kafka run -i --rm listen-$(date +%s) \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -C -c 1 -e \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.<name>.smoke.responses

# Then produce:
kubectl -n kafka run -i --rm send-$(date +%s) \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.adapter.<name>.requests \
    -H correlation_id=$CORRELATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_type=request <<JSON
{"headers":{"correlation_id":"$CORRELATION_ID","request_id":"$REQUEST_ID"},
 "body":{"action":"<some-action>","reply_to_topic":"system.<name>.smoke.responses"}}
JSON

# 4. Verify the response shape
# Expect headers including in_response_to_request_id=$REQUEST_ID, status=complete-or-error_*
```

If the consumer in step 3 never receives anything, the adapter consumed the request but didn't produce — almost always a header/envelope shape bug. Check the adapter logs.

---

## Adapter checklist before merging

- [ ] `cmd/<name>-adapter/main.go` matches the standard signal-handler pattern
- [ ] `internal/adapters/<name>/adapter.go` has Adapter struct + NewAdapter + Run + handleMessage + Shutdown + StartHealthServer
- [ ] `configs/<name>-adapter.yaml` uses correct field names (logging, port, sslmode)
- [ ] Producer.Produce called with 5 args, headers as `map[string]string`
- [ ] Response includes `in_response_to_request_id` and `status` headers
- [ ] `/health` and `/ready` endpoints functional
- [ ] Shutdown is `sync.Once`-protected and idempotent
- [ ] All resources closed on `NewAdapter` failure paths
- [ ] No credentials in config YAML (only env var NAMES)
- [ ] Kustomize manifest deployed with health probes + envFrom secrets
- [ ] Smoke test produces a response with correct correlation headers
- [ ] If touching a database, integration-tested against `clients_db` (or whichever schema)
- [ ] If touching B2/S3, integration-tested with the actual bucket
