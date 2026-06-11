# 035 — Adapter Guide

Canonical guide for building chassis adapters (long-lived cluster services that
wrap one external system or shared piece of infrastructure: GitHub, Firecrawl,
Stability, Thunder, Backblaze, Ollama, …).

**Scope.** Section 1 — the message envelope contract — is **normative** and
applies to *any* component that replies to a chassis request: adapters, spawned
agents, and inline actions alike. It was extracted from `003_contracts_and_standards.md`
§"Adapter Response Envelope Contract"; 003 now points here rather than restating
it. Sections 2+ are the adapter-specific construction recipe and are being
migrated from `FOCUS_adapter_design` (which this doc supersedes).

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

> **To be migrated from `FOCUS_adapter_design`.** The sections below are the
> intended structure; until each is filled, `FOCUS_adapter_design` remains the
> detailed source. The envelope material that FOCUS duplicated is **not** copied
> here — it lives in Section 1, and FOCUS's "Sending responses" skeleton (which
> used the deprecated map/string-bool/`Produce` shape) should not be carried over;
> use the favoured pattern in 1.8 instead.

- **2.1 Decision: adapter vs agent vs inline** — when an adapter is the right
  pattern (one external credential boundary, request/response shaped, multiple
  internal callers) versus a spawned agent or an inline action.
- **2.2 Responsibilities** — what the adapter owns (credentials, validation,
  retry, throttle, circuit breaker, response mapping) and does not (orchestration
  state, the decision of when to call).
- **2.3 File and package layout** — `cmd/<name>-adapter/main.go`,
  `internal/adapters/<name>/adapter.go`, `configs/<name>-adapter.yaml`,
  `deployments/kustomize/services/<name>-adapter/`; one hyphenated name
  throughout.
- **2.4 The adapter struct** — ctx/cancel, cfg, logger, consumer, producer,
  adapterID, the API client, optional db/storage/throttle, requestsTopic,
  healthServer, shutdownOnce, shutdownWg.
- **2.5 `NewAdapter` (cleanup ordering)** — each failure path closes everything
  opened before it; manual, no `defer` magic.
- **2.6 Run loop** — fetch one, handle, loop; **sequential by default**; backoff
  on real fetch errors; return nil on shutdown.
- **2.7 `handleMessage`** — parse envelope (per 1.2), dispatch by action, send a
  response (per 1.3–1.6), commit.
- **2.8 Graceful shutdown** — `Shutdown()` guarded by `sync.Once`; cancel the
  context, stop the health server, wait on `shutdownWg`, close Kafka/DB.
- **2.9 Health endpoints** — `/health` unconditional 200; `/ready` checks the
  critical dependency (503 on failure). Keep both cheap.
- **2.10 Topic naming** — convention A: `system.adapter.<name>.requests`,
  consumer group `<name>.adapter.group`; `REQUESTS_TOPIC` / `CONSUMER_GROUP` env
  overrides.
- **2.11 Config YAML field names** — `service_info.name` (becomes
  `sender_agent_type`), `server.port` (string), `logging.level` (not `logger:`),
  `infrastructure.kafka_brokers`, `sslmode` (not `ssl_mode`),
  `custom.adapter_settings`.
- **2.12 Deployment** — kustomize base + overlays; `imagePullSecrets`, the
  service account, `envFrom: secretRef:` for credentials, health probes.
- **2.13 Credentials and secrets** — read from env, never config; config holds
  the env var *name*, never the value; least-privilege per adapter.
- **2.14 End-to-end testing** — smoke a request with `kcat`, confirm a response
  on the reply topic with `in_response_to_request_id` set and `status=complete`.
- **2.15 Pre-merge checklist** — main.go signal handler, struct + NewAdapter +
  Run + handleMessage + Shutdown, config field names, `ProduceWithValidation`
  with five args, response correlation headers, health endpoints, `sync.Once`
  shutdown, NewAdapter cleanup, no credentials in config, kustomize with probes +
  secrets, smoke test.
