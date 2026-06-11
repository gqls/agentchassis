# 035 — Adapter Guide

Canonical guide for building chassis adapters (long-lived cluster services that
wrap one external system or shared piece of infrastructure: GitHub, Firecrawl,
Stability, Thunder, Backblaze, Ollama, …).

**Scope.** Section 1 — the message envelope contract — is **normative** and
applies to *any* component that replies to a chassis request: adapters, spawned
agents, and inline actions alike. It was extracted from `003_contracts_and_standards.md`
§"Adapter Response Envelope Contract"; 003 now points here rather than restating
it. Section 2 is the adapter-specific construction recipe, consolidated from
`FOCUS_adapter_design`, which this doc now fully supersedes and replaces.

**Provenance.** Last verified against code: **2026-06-11** — `coordinator.go`
(`ProcessResponse`), the `git` / `thunder` / `image-generator` adapters (favoured
pattern), and the `web-search` adapter (deprecated pattern). Where this doc and
the code disagree, **the code wins**: fix the doc and update this date.

---

## 1. The message envelope contract (normative)

### 1.1 Why it matters — the silent-hang failure mode

When an adapter replies to a request it received on its fixed topic, the reply
must be shaped so the chassis recognises it as an **awaited response** and routes
it to the coordinator's claim path (`HandleResponse → ProcessResponse →
ClaimAwaitedRequest`). If the envelope is wrong, the chassis silently falls
through to the generic *process-as-work* path (`processMessage → ProcessMessage
→ BuildCollectedData`); the `awaited_requests` row stays `waiting` until it times
out, and the orchestration retries roughly every ten minutes with no error
logged. The work itself can have fully succeeded — only the reply is lost. Every
rule below exists to prevent that specific failure.

### 1.2 Request envelope — how an adapter parses an incoming request

Grounded in the working adapters (`thunder`, `git`, `image-generator`,
`web-search`):

- **`action` comes from the body** (`body.action`), never the headers. thunder
  reads `body["action"]`, git reads `req.Body.Action`, the image-generator reads
  `req.Body.Data`. An adapter that reads `headers.action` will see an empty
  action for a correctly-shaped request and reply "not implemented".
- **The payload is at `body.data`** — unmarshal that into the action's typed
  request, not the whole body.
- **The reply topic source is mixed across the codebase**: `git` and
  `web-search` read it from the **headers** (`responses_topic`, then
  `reply_to_topic`); `thunder` and the image-generator read it from the **body**
  (`reply_to_topic`). Accept all three keys rather than assume one. A *sender*
  should set `responses_topic` in the headers **and** `reply_to_topic` in the
  body, so any adapter convention finds it.
- **Correlation and request ids** come from the message headers
  (`correlation_id`, `request_id`); they are echoed into the response (below).

### 1.3 Response headers — three tiers

Not every header is equal. Three tiers, by who reads them and what breaks when
they are missing:

**Tier 1 — enforced by `platform/validation` (`ValidateOutgoingMessage`).** A
message missing any of these is rejected at send time *unless* `is_error=true`
(the escape hatch, so a partly-broken adapter can still signal failure):
`client_id`, `correlation_id`, `orchestration_id`, `sender_agent_type`, and
`step_name` **or** `in_response_to_step_name`. All adapters get these right by
passing them through from the incoming request.

**Tier 2 — the coordinator needs them to route, but the validator does NOT check
them.** Omit one and validation passes, the reply lands, gets logged, and the
orchestration sits in `AWAITING_RESPONSES` until timeout. This is the most
common silent-hang cause:
`in_response_to_request_id`, `message_type` (`"response"`), `status`,
`request_id`, `message_id`, `in_response_to_step_id`, `is_complete`/`is_error`.

**Tier 3 — observability only, safe to omit.** `sender_agent_id`,
`sender_pod_name`, `sender_agent_version`, `time_sent`, `correlation_name`,
`in_response_to_action`. Cheap to add; free insurance against future chassis
additions that *do* read them.

The required Tier-1/Tier-2 values in full:

| Header | Value | Why |
|---|---|---|
| `message_type` | `"response"` | Selects the response consumer path. |
| `in_response_to_request_id` | the **incoming** request's `request_id` | The coordinator claims the awaited row by this id (see 1.4). The single load-bearing field. |
| `request_id` | the incoming `request_id` (favoured: reused) **or** a fresh UUID | Message identity / fallback claim key (see 1.4). |
| `message_id` | a fresh UUID | Without it the chassis may synthesise one and treat the reply as unsolicited inbound. |
| `orchestration_id` | the incoming `orchestration_id` | Identifies the orchestration that owns the awaited row. |
| `correlation_id` | the incoming `correlation_id` | Kafka partition key and tracing. |
| `client_id` | the incoming `client_id` | Tier-1 validator field. |
| `status` | `"complete"` \| `"error_recoverable"` \| `"error_unrecoverable"` | Drives the coordinator's status switch (see 1.6). |
| `is_complete` / `is_error` | real bools in the body (see 1.5) | Multi-part / error routing. |
| `sender_agent_type` | the adapter's type (e.g. `"thunder-adapter"`) | Tier-1 validator field; tracing. |
| `in_response_to_step_id`, `in_response_to_step_name` | echoed from the request | Lets the coordinator resume the correct workflow step. |

The reply is sent to the request's reply topic (see 1.2) via
`ProduceWithValidation` (see 1.6).

**The body envelope.** The reply's JSON value carries both the headers (again,
in the body) and a result body:

```json
{
  "headers": { "in_response_to_request_id": "...", "request_id": "...",
               "status": "complete", "is_complete": true, "is_error": false,
               "orchestration_id": "...", "correlation_id": "...",
               "sender_agent_type": "..." },
  "body": { "success": true, "data": { /* action result */ } }
}
```

On the error path the body is `{ "success": false, "error": "<message>" }`. Both
the Kafka *message* headers and these body-embedded headers are set; the
body-embedded headers are the ones the chassis primarily reads, which is why they
are the type-sensitive consumer (1.5). The result field is `data` in the
reference adapters (git, thunder); if your `types.ResponseBody` tags the payload
field differently, confirm the consuming action reads the same name at first run.

### 1.4 The matcher — `in_response_to_request_id` is the load-bearing field

`coordinator.go ProcessResponse` claims the awaited request like this:

```go
var requestID string
if execCtx.InResponseTo != nil {
    requestID = execCtx.InResponseTo.RequestID   // in_response_to_request_id — tried first
}
if requestID == "" {
    requestID = execCtx.RequestID                // request_id — fallback only
}
// ... ClaimAwaitedRequest(ctx, requestID, ...)
```

So the **claim key is `in_response_to_request_id`**, and `request_id` is only a
fallback used when `in_response_to_request_id` is empty. The rules that follow
from the code:

- **Required:** `in_response_to_request_id` MUST equal the incoming request's
  `request_id`. This is the field that makes the match succeed.
- **`request_id` MAY be a fresh UUID.** Because the claim is on
  `in_response_to_request_id`, a new `request_id` does not miss the lookup.
- **Favoured pattern (use for new adapters):** the `git` adapter *reuses* the
  incoming `request_id` (`RequestID: requestHeaders.RequestID`) **and** adds a
  fresh `message_id`. Reusing it satisfies both the primary and the fallback
  match path, so it is the safest choice.

> Correction to earlier guidance: a previous version of this contract said "the
> router matches `request_id`; a freshly-generated id misses the lookup." That is
> only true if `in_response_to_request_id` is *also* absent or new. With
> `in_response_to_request_id` set to the incoming id (as required), a fresh
> `request_id` is fine.

### 1.5 The bool trap — body headers MUST be a typed struct

Build the response envelope's `headers` object (in the JSON **body**) from a
**typed Go struct**, not a `map[string]string`. The chassis unmarshals the
reply's body `headers` into `types.ResponseHeaders`, where `is_complete` and
`is_error` are Go `bool`. A `map[string]string` can only emit them as JSON
*strings* (`"true"`), and the unmarshal then fails:

```
json: cannot unmarshal string into Go struct field
ResponseHeaders.headers.is_complete of type bool
```

On that error the response-routing branch returns early and the reply is
**dropped before `ClaimAwaitedRequest`** — the awaited row sits in `waiting`
until timeout, even though the work fully succeeded. This was the root cause of a
multi-day `thunder-adapter` matcher failure (verified 2026-05-22): provision
completed, the response was consumed and identified as `message_type: response`,
then silently discarded on the bool unmarshal. The `thunder-adapter` source now
carries a comment recording this (verified 2026-06-11).

The split that resolves it:

- **JSON body `headers`** → a typed struct with real `bool` fields, so
  `is_complete`/`is_error` marshal as JSON `true`/`false`. **This is the
  type-sensitive consumer.**
- **Kafka *message* headers** → a `map[string]string` rendered by a
  `toKafkaHeaders()` method, where string bools are correct (Kafka headers are
  byte strings).

The `git` adapter's `ResponseHeaders` struct + `ToKafkaHeaders()` is the model.
Note the chassis is itself inconsistent here — some chassis paths emit a string
`is_complete` into the same bool-typed struct. That is a latent chassis bug;
adapters must not rely on it and must send real bools in the body.

### 1.6 `ProduceWithValidation`, never plain `Produce`

Send the reply with `ProduceWithValidation`, not `Produce`. It runs the
outgoing-message validator (when one is injected) and blocks a malformed
*non-error* response at send time instead of emitting one that fails silently
downstream; error responses (`is_error=true`) are always sent even if validation
fails. `git`, `thunder`, and the image-generator all use it.

Status vocabulary (the coordinator branches on `status`):
- `complete` — success, advance the workflow.
- `error_recoverable` — failed, retry may help.
- `error_unrecoverable` — failed, do not retry; propagate the failure.

The coordinator also accepts a few aliases (`success` as complete; `failed` /
`error` as unrecoverable; `awaiting` / `processing` for progress updates), but
author new adapters with the three canonical values.

### 1.7 The canonical-type gap (why adapters carry a local header struct)

`types.ResponseHeaders` (`platform/orchestration/types/context.go`) has **no
`request_id` and no `message_id` field**, and no `ToKafkaHeaders()` method. That
is precisely why the `git` adapter carries its own `ResponseHeaders` mirror with
those fields and a `ToKafkaHeaders()`. An adapter must therefore either mirror
that local struct (git's approach) or set `request_id`/`message_id` directly in
the Kafka message headers while reusing the canonical type for the body's
real-bool guarantee. **Fix path:** add `request_id`/`message_id` to the canonical
`types.ResponseHeaders`, after which adapters can drop the mirror. Until then the
duplication is expected, not a smell.

### 1.8 Favoured pattern (summary, for new adapters)

The `git` and `thunder` adapters are the verified reference implementations:

- request parse: `action` from `body.action`, payload from `body.data`, reply
  topic from headers or body (1.2);
- response: a typed `responseHeaders` struct with real bool `is_complete` /
  `is_error` + a `toKafkaHeaders()` method (1.5);
- `in_response_to_request_id` = incoming `request_id`; `request_id` reused;
  `message_id` fresh (1.4);
- sent via `ProduceWithValidation` (1.6).

### 1.9 Deprecated anti-pattern (do not copy)

A `map[string]string` body-`headers` builder with **string** `is_complete` /
`is_error`, a **fresh** `request_id` via `uuid.New()` with no
`in_response_to_request_id`, **no** `message_id`, and plain **`Produce`** is the
anti-pattern. The **`web-search` adapter** is a current example of it (distinct
from the `web-scrape` adapter). Any adapter doing any of these should be migrated.
Audit checklist for `sendSuccessResponse` / `sendErrorResponse`:

1. Are the body headers a **typed struct** with `bool` `is_complete`/`is_error`,
   or a `map[string]string` emitting string bools? *(must be a typed struct)*
2. Is `in_response_to_request_id` set to the **incoming** `request_id`? *(must be)*
3. Is `message_id` set to a fresh UUID? *(must be)*
4. Is it `ProduceWithValidation`? *(must be)*

### 1.10 TODO — promote this contract from prose to validator enforcement

The contract above lives as prose and has drifted before (003 vs the deprecated
adapter skeletons). The durable fix is to make it executable:

- Extend `platform/validation` `ValidateOutgoingMessage` (and
  `ValidateIncomingMessage` where applicable) to **require the Tier-2 routing
  fields when `message_type == "response"`** — above all `in_response_to_request_id`,
  plus `message_type`, `status`, `request_id`. Keep the `is_error=true` escape
  hatch so a partly-broken adapter can still signal failure.
- Promote `in_response_to_request_id` specifically: it is the single most common
  silent-hang cause, and a missing-field check would catch it at the producer
  rather than after a ~10-minute timeout.
- Optionally add a Tier-2.5 layer that logs a Warn (not Error) for missing
  *recommended* fields during a transition period.

A validator that rejects the deprecated shape is something a doc skeleton cannot
drift away from. **Status: not yet filed — raise as a tracked task.**

---

## 2. Construction (adapter-specific recipe)

Consolidated from the former `FOCUS_adapter_design`. The envelope material that
doc duplicated is not repeated here — it lives in Section 1. In particular, do
**not** copy a `map[string]string` / string-bool / plain-`Produce` send skeleton
(the shape §1.9 marks deprecated); use the favoured pattern in §1.8.

### 2.1 Decision: adapter vs agent vs inline

| Scenario | Pattern |
|---|---|
| Talks to one external API, multiple internal callers want it | **Adapter** (this doc) |
| Long-running per-orchestration work (training, site build) | Spawned agent (Job pod) |
| Short call inside one agent's workflow | Inline action |
| Shared infrastructure with no external API (db, kafka) | No wrapper needed |

When in doubt: if more than one agent type will call the same external API, build an adapter.

### 2.2 Responsibilities

**Adapter owns:** credentials for the external system (held in pod env, never
propagated outward); request validation and shape coercion before hitting the
external API; retry policy with backoff; rate limiting / throttle; circuit
breaker; per-call logging and metrics; mapping external responses back to
chassis-shape responses; database tracking when the work creates persistent state
(e.g. `thunder_instances`).

**Adapter does NOT own:** orchestration state (the chassis owns that);
per-orchestration data lifecycle (the calling workflow owns it); the decision of
when to call (the orchestrator decides); knowledge of which agent type called it
(it is just request/response).

### 2.3 File and package layout

```
cmd/<name>-adapter/main.go              ← entry point: load config, init, signal handling
internal/adapters/<name>/adapter.go     ← struct, NewAdapter, Run, handleMessage, Shutdown
internal/adapters/<name>/<api>_client.go  ← optional: external API client
internal/adapters/<name>/<feature>.go   ← optional: per-action handlers
configs/<name>-adapter.yaml             ← config file mounted as ConfigMap
build/docker/backend/<name>-adapter.dockerfile   ← optional, often shared multi-stage build
deployments/kustomize/services/<name>-adapter/   ← kustomize base + overlays
```

Use a hyphenated name for files and topics; one canonical name throughout (e.g.
`web-scrape`, not `webscrape` in some places and `web-scrape` in others). The
`internal/adapters/<name>/` package may use a single-word identifier
(`webscrape`) since Go packages do not allow hyphens.

### 2.4 The adapter struct

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

Drop fields the adapter does not need. Most adapters have producer + consumer +
apiClient + logger as the core; storage and DB are optional.

### 2.5 NewAdapter (cleanup ordering)

Order matters: each step can fail, and earlier resources must close cleanly on
later failure.

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

**Cleanup convention:** every failure path closes everything successfully opened
before it. No `defer` magic; explicit and manual. Matches `git-adapter` and
`image-generator-adapter`.

### 2.6 Run loop

Fixed shape: fetch one message, handle it, loop.

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
            a.handleMessage(msg)   // sequential by default
            if a.throttle != nil {
                a.throttle.Wait()
            }
        }
    }
}
```

**Do not `go a.handleMessage(msg)` by default.** Sequential is safer: Kafka
offsets commit cleanly only after the handler finishes; concurrent handlers can
starve the external API's rate limit; `throttle.Wait()` between handlers is
meaningless if they are concurrent. Some adapters use goroutines deliberately —
make it a conscious choice, not the default.

### 2.7 handleMessage

Three steps per message: parse the envelope (per §1.2), dispatch by action, send
a response (per §1.3–1.6) and commit.

```go
func (a *Adapter) handleMessage(msg kafka.Message) {
    headers := kafka.HeadersToMap(msg.Headers)
    l := a.logger.With(
        zap.String("correlation_id", headers["correlation_id"]),
        zap.String("request_id", headers["request_id"]),
    )

    var envelope map[string]interface{}
    if err := json.Unmarshal(msg.Value, &envelope); err != nil {
        l.Error("Failed to unmarshal message", zap.Error(err))
        a.commit(msg)   // commit anyway — re-processing won't help a malformed message
        return
    }

    body, _ := envelope["body"].(map[string]interface{})
    if body == nil { body = make(map[string]interface{}) }
    action, _ := body["action"].(string)            // action from body (§1.2)
    replyToTopic, _ := body["reply_to_topic"].(string)

    var result interface{}
    var err error
    switch action {
    case "provision_instance":
        result, err = a.handleProvisionInstance(body)
    default:
        err = fmt.Errorf("unknown action: %s", action)
    }

    if err != nil {
        a.sendErrorResponse(headers, replyToTopic, action, err, l)
    } else {
        a.sendSuccessResponse(headers, replyToTopic, action, result, l)
    }
    a.commit(msg)
}
```

Two cross-references that matter: the reply topic may arrive in the headers
rather than the body depending on the caller (§1.2 — accept both); and the
`sendSuccessResponse` / `sendErrorResponse` helpers must build a **typed**
response per §1.8, not the deprecated map/string-bool skeleton (§1.9). The
correlation rules (§1.4), the bool trap (§1.5), and the body envelope (§1.3) all
apply to those helpers.

### 2.8 Graceful shutdown

```go
func (a *Adapter) Shutdown() {
    a.shutdownOnce.Do(func() {
        a.logger.Info("Adapter shutting down")
        a.cancel()  // signals Run() to exit at next iteration

        if a.healthServer != nil {
            sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            _ = a.healthServer.Shutdown(sctx)
        }

        done := make(chan struct{})
        go func() { a.shutdownWg.Wait(); close(done) }()
        select {
        case <-done:
        case <-time.After(10 * time.Second):
            a.logger.Warn("Shutdown wait exceeded")
        }

        a.consumer.Close()
        a.producer.Close()
        if a.db != nil { _ = a.db.Close() }
        a.logger.Info("Shutdown complete")
    })
}
```

`sync.Once` makes this safe to call more than once — the signal handler in
`main()` and a test harness can both call it.

### 2.9 Health endpoints

Two endpoints, on the port from `cfg.Server.Port` (default 8080):

| Path | Purpose | Implementation |
|---|---|---|
| `/health` | Liveness — am I alive? | Return 200 with `{"status":"ok"}` unconditionally |
| `/ready` | Readiness — can I serve traffic? | Ping the DB and any critical dependency; 200 on success, 503 on failure |

K8s probes use these: `/health` triggers pod restart on failure; `/ready` removes
the pod from Service endpoints. Keep them cheap (<100ms).

### 2.10 Topic naming

New adapters use **convention A**:

- **Convention A — `system.adapter.<name>.requests`** (used by `git-adapter`,
  `image-generator-adapter`, `thunder-adapter`), e.g. `system.adapter.thunder.requests`.
- **Convention B — `system.agent.<name>-adapter.process`** (legacy, e.g.
  `system.agent.webscrape-adapter.process`). Predates A; old adapters keep their
  topic for compatibility but should migrate when convenient.

**Consumer group:** `<name>.adapter.group` (e.g. `thunder.adapter.group`); stable
across pod restarts for offset continuity. **Env overrides:** `REQUESTS_TOPIC`
and `CONSUMER_GROUP` on the Deployment override the YAML defaults (useful for
staging environments that share a cluster).

### 2.11 Config YAML

**Use these exact field names** — the Go struct tags expect them:

```yaml
service_info:
  name: "thunder-adapter"          # used as sender_agent_type in responses
  version: "0.1.0"
  environment: "production"        # production | development | staging

server:
  port: "8080"                     # STRING in quotes; reads as cfg.Server.Port
                                   # (some adapters use http_port/grpc_port — a different schema, avoid)

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

**Codebase inconsistency:** `image-adapter.yaml` uses `http_port` + `grpc_port` +
`shutdown_timeout` under `server:` — a different schema with extra fields. For new
adapters use the single-`port` shape from `agent-chassis.yaml` and
`web-scrape-adapter.yaml`.

### 2.12 Deployment

Real lessons from the thunder-adapter Phase 2 deploy; none optional.

**The full manifest pattern** — the first three (`serviceAccountName`,
`imagePullSecrets`, and `command:` with full invocation) are the ones most often
missed and produce confusing failures (see §10 of `016_debugging_guide` for the
symptom-to-cause table):

```yaml
spec:
  template:
    spec:
      # ── Pod-spec fields (NOT container-spec) ──

      # Required: SA with the permissions the adapter actually needs.
      serviceAccountName: ai-persona-app

      # Required: pulling private aqls/* images from Docker Hub.
      # Without this: "insufficient_scope: authorization failed" on image pull.
      imagePullSecrets:
        - name: docker-hub-creds

      containers:
        - name: <name>-adapter
          image: docker.io/aqls/<name>-adapter:vX.Y.Z

          # CRITICAL: use `command:` with full path, NOT `args:` alone.
          # Our Dockerfiles use CMD (not ENTRYPOINT); `args:` would REPLACE the
          # entire CMD including the binary path, and kubelet then tries to exec
          # the first arg as the binary ("exec: '--config': ... not found").
          command: ["./<name>-adapter", "-config", "/app/configs/<name>-adapter.yaml"]

          ports:
            - containerPort: 8080
              name: http

          env:
            - name: POD_NAME            # Required for Tier-3 sender_pod_name
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: AGENT_VERSION
              value: "0.1.0"

          envFrom:
            - secretRef:
                name: personae-default-secrets    # API keys (ANTHROPIC_, THUNDER_, …)
            - secretRef:
                name: personae-platform-secrets   # DB passwords, B2 credentials

          volumeMounts:
            - name: config
              mountPath: /app/configs/<name>-adapter.yaml
              subPath: <name>-adapter.yaml
              readOnly: true

          readinessProbe:
            httpGet: {path: /ready, port: 8080}
            initialDelaySeconds: 10
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          livenessProbe:
            httpGet: {path: /health, port: 8080}
            initialDelaySeconds: 30
            periodSeconds: 30
            timeoutSeconds: 5
            failureThreshold: 5

          resources:
            requests: {cpu: 100m, memory: 128Mi}
            limits:   {cpu: 500m, memory: 512Mi}

      volumes:
        - name: config
          configMap:
            name: <name>-adapter-config

  replicas: 1
  strategy:
    type: Recreate     # avoids two adapters processing the same partition during rollout
```

**Required cluster resources before first deploy** — the manifest assumes these
exist; if any are missing the pod starts but fails at first message:

| Resource | Purpose | How to check |
|---|---|---|
| `docker-hub-creds` Secret in `ai-persona-system` | imagePullSecret for private repos | `kubectl -n ai-persona-system get secret docker-hub-creds` |
| `ai-persona-app` ServiceAccount | Pod identity, future k8s API access | `kubectl -n ai-persona-system get sa ai-persona-app` |
| `personae-default-secrets`, `personae-platform-secrets` | Env API keys / DB / B2 | `kubectl -n ai-persona-system get secrets \| grep personae-` |
| Docker Hub repo grant for the new image | Pull credential needs read on this repo | `docker pull docker.io/aqls/<name>-adapter:<tag>` |
| **Kafka topics the adapter produces to** | Strimzi `auto.create.topics.enable=false` | `kubectl -n kafka get kafkatopic` |

The Kafka-topic gotcha is easy to miss because it does not fail at startup: the
adapter consumes happily, then fails on first response with `Unknown Topic Or
Partition`; and a missing *requests* topic makes every fetch hang with `context
deadline exceeded`. Declare each topic the adapter reads or writes:

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: <topic-name>
  namespace: kafka
  labels:
    strimzi.io/cluster: personae-kafka-cluster
spec:
  partitions: 1
  replicas: 1
  config:
    retention.ms: 604800000   # 7 days; tune per use case
```

**Service permissions (when applicable)** — if the adapter manages k8s resources
(e.g. SSH key Secrets), scope a Role/RoleBinding to just those resources; never
`cluster-admin`, never broad `*` verbs:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: thunder-adapter-secret-manager
  namespace: ai-persona-system
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["thunder-ssh-*"]   # specific pattern, not all secrets
    verbs: ["create", "get", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: thunder-adapter-secret-manager
  namespace: ai-persona-system
subjects:
  - kind: ServiceAccount
    name: ai-persona-app
    namespace: ai-persona-system
roleRef:
  kind: Role
  name: thunder-adapter-secret-manager
  apiGroup: rbac.authorization.k8s.io
```

**Makefile integration** — four insertion points: a `build-<name>-adapter`
target; add to the `build-adapters` aggregate; add `docker push
$(REGISTRY)/<name>-adapter:$(IMAGE_TAG)` to `push-backend`; add a deploy block to
`deploy-agents`. `deploy-agents` only seds `newTag`, so the overlay sets `newName`
once:

```yaml
images:
  - name: <name>-adapter
    newName: docker.io/aqls/<name>-adapter
    newTag: vX.Y.Z          # overwritten by sed on each release
```

**Pre-deploy checklist** (before `make release-backend IMAGE_TAG=vX.Y.Z`):
Dockerfile at `build/docker/backend/<name>-adapter.dockerfile` matching the
canonical two-stage pattern; the four Makefile insertions; `base/` has
deployment.yaml, service.yaml, kustomization.yaml, and the config YAML; base
deployment has `serviceAccountName`, `imagePullSecrets`, and `command:`; overlay
sets `newName` once; all KafkaTopic CRDs applied; Docker Hub read granted.

**Post-deploy verification:**

```bash
# 1. Pod running
kubectl -n ai-persona-system get pods -l app=<name>-adapter            # Expect: 1/1 Running
# 2. SA + image actually applied
kubectl -n ai-persona-system describe pod -l app=<name>-adapter | grep -E "Service Account|Image:"
# 3. Startup logs clean
kubectl -n ai-persona-system logs deploy/<name>-adapter --tail=30      # connection lines, then "starting message processing"
# 4. Smoke test (see 2.15)
```

### 2.13 Credentials and secrets

Adapters hold credentials so individual agents do not have to. Two rules:

1. **Read credentials from env, never from code.** Set them via `envFrom:
   secretRef:` in the Deployment manifest.
2. **Never put credentials in the config YAML.** Config holds the env var NAME
   (`access_key_env_var: "B2_APPLICATION_KEY_ID"`), not the value.

If the adapter forwards a temporary credential to an external system (e.g. a
presigned URL to a VM), generate it on the fly with short expiry and never log the
full URL. Each adapter gets the least privilege it needs — a read-only token for a
read-only adapter, not a shared broad credential.

### 2.14 Common mistakes (real examples)

- **`Producer.Produce` arguments** — five args:
  `(ctx, topic, headers map[string]string, key []byte, value []byte)`. Do not pass
  a `[]kafka.Header` slice where `map[string]string` is expected; do not forget
  `key` (conventionally the correlation_id as bytes); four args will not compile,
  and the error blames the wrong line.
- **Response correlation and bools** — covered in §1.3–1.6. The two separate
  failure modes (validator rejection on the five Tier-1 fields; silent
  `AWAITING_RESPONSES` hang on a missing `in_response_to_request_id`) and the bool
  trap all live there.
- **Status vocabulary** — only `complete` / `error_recoverable` /
  `error_unrecoverable` (§1.6). Do not invent values.
- **Topic naming inconsistency** — two conventions exist; new work uses A; check
  the similar adapter and do not mix.
- **YAML field names** — `logger:` vs `logging:`, `ssl_mode` vs `sslmode`,
  `http_port` vs `port`. Copy from a known-good config (`agent-chassis.yaml`,
  `git-adapter.yaml`) rather than typing fresh; cryptic config-load crashes are
  usually this.
- **Closing resources on `NewAdapter` failure** — manual cleanup at every step
  (§2.5); forgetting one leaks open Kafka/DB connections.
- **`go a.handleMessage(msg)` by default** — sequential is the default (§2.6).
- **Logging credentials** — strip bearer tokens / signed URLs from any
  request-body logging; never log `os.Environ()` raw in production.

### 2.15 Testing a new adapter end-to-end

```bash
# 1. Pod up and ready
kubectl -n ai-persona-system get pods -l app=<name>-adapter
kubectl -n ai-persona-system logs deploy/<name>-adapter | head -50
# Expect: connection logs for kafka, db (if used), and "Adapter initialized"

# 2. Health endpoints
kubectl -n ai-persona-system port-forward deploy/<name>-adapter 8081:8080 &
curl -s http://localhost:8081/health
curl -s http://localhost:8081/ready

# 3. Smoke message — produce a test request, listen for the response
CORRELATION_ID=$(uuidgen)
REQUEST_ID=$(uuidgen)
# Open the consumer first, in another terminal:
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

If the consumer in step 3 never receives anything, the adapter consumed the
request but did not produce — almost always a header/envelope shape bug (check
§1.2–1.6 and the adapter logs).

### 2.16 Pre-merge checklist

- [ ] `cmd/<name>-adapter/main.go` matches the standard signal-handler pattern
- [ ] `internal/adapters/<name>/adapter.go` has Adapter struct + NewAdapter + Run + handleMessage + Shutdown + StartHealthServer
- [ ] `configs/<name>-adapter.yaml` uses correct field names (logging, port, sslmode)
- [ ] `Producer.Produce` / `ProduceWithValidation` called with the right args, headers as `map[string]string`
- [ ] Response sets `in_response_to_request_id` and `status`, body headers a typed struct (§1.4, §1.5)
- [ ] `/health` and `/ready` endpoints functional
- [ ] Shutdown is `sync.Once`-protected and idempotent
- [ ] All resources closed on `NewAdapter` failure paths
- [ ] No credentials in config YAML (only env var NAMES)
- [ ] Kustomize manifest deployed with health probes + envFrom secrets
- [ ] All KafkaTopic CRDs the adapter reads/writes are applied
- [ ] Smoke test produces a response with correct correlation headers
- [ ] If touching a database, integration-tested against `clients_db` (or whichever schema)
- [ ] If touching B2/S3, integration-tested with the actual bucket
- 