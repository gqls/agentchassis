# Register — llm-call-observability

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

4 concepts, consolidated from 5 raw extractions (one exact duplicate pair
collapsed) across unit U13.

### LCO-001 — Temperature/max_tokens logging gap in llm_call_log
- **status:** partial
- **status-evidence:** "the schema column exists but is never written, so llm_call_log.temperature remains NULL by construction" — verified against a 2026-05-26 chassis snapshot.
- **what:** Although `llm_call_log` already has `temperature real` and `max_tokens integer` columns, the Go writer (`llm_call_logger.go`) never populates them, and the two call sites in `execute_llm_prompt_action.go` don't pass the values through — even though the actual values sent to the LLM API are already computed a few lines earlier. This makes it impossible to observe from the log alone whether a configured temperature ever reached the API call.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Verification table, #Step 1
- **relations:** Per-field LLM config resolution fallback chain (LCO-002); Possibility-A-vs-B diagnostic method (LCO-003); Anthropic client temperature parameter removed unconditionally (model-infrastructure); LLM step config shadowing bug (model-infrastructure)
- **verify-later:** platform/orchestration/actions/llm_call_logger.go LLMCallLogParams struct; execute_llm_prompt_action.go call sites

### LCO-002 — Per-field LLM config resolution fallback chain (temperature parity with max_tokens)
- **status:** aspirational
- **status-evidence:** "Per-field fix not applied" — temperature read is "Still single read: agentConfig['temperature'].(float64)" versus max_tokens' existing two-level fallback.
- **what:** Proposes lifting temperature to the same multi-level fallback chain max_tokens already has (step config → agent top → step ai_service → top-level ai_service) via shared `readFloat`/`readNestedFloat` helpers, replacing the single inline float64 type assertion. Would activate 6 currently-dead step-level temperature settings configured in the DB but never actually taking effect.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 3, #Step 4
- **relations:** Temperature/max_tokens logging gap (LCO-001); Possibility-A-vs-B diagnostic method (LCO-003); LLM step config shadowing bug (model-infrastructure, same top-level-shadows-step-level family)
- **verify-later:** readFloat/readNestedFloat helpers (proposed, not yet added)

### LCO-003 — Possibility-A-vs-B diagnostic method for silent LLM config failures
- **status:** partial
- **status-evidence:** "Smallest change that makes possibility A (logging gap) vs B (temperature never set → API default ~1.0) distinguishable from the log" with an exact before/after SQL audit query.
- **what:** A diagnostic technique for a silently-broken config field: ship the cheapest possible observability fix first, before attempting any structural resolution fix, then re-run a COUNT(*)/COUNT(temperature)/COUNT(max_tokens) audit query pre- and post-deploy. If temperature becomes non-null post-deploy, the bug was pure logging gap; if still null, the upstream read itself is silently failing — distinguishing the two determines whether every historical LLM call ran at the intended temperature.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 2, #Sequencing summary
- **relations:** Temperature/max_tokens logging gap (LCO-001); Per-field LLM config resolution fallback chain (LCO-002)
- **verify-later:** SQL audit query in Step 2

### LCO-004 — Default temperature hardening (chassis-level fallback ~0.4)
- **status:** aspirational
- **status-evidence:** Explicitly sequenced last and conditional: "Hold this until the read path is proven, so we don't stack a default on top of an unverified read."
- **what:** Once the observability and per-field resolution fixes are proven, proposes a chassis-level default temperature (~0.4) applied only when none is configured at any level, overridable by an explicit value — reasoning that Anthropic's API default (~1.0) is likely too high for the extraction/classification-style prompts most affected.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 5
- **relations:** Per-field LLM config resolution fallback chain (LCO-002)
- **verify-later:** n/a (not yet implemented; gated on Steps 1-3)

### LCO-005 — `aiservice.Fingerprint`: log a model response's SHAPE, never its text
- **status:** deployed
- **what:** `Fingerprint(s string) string` returns one stable line —
  `chars=1834 first=L fence=yes objects=2 parses=false keys=[]` — describing an
  LLM response structurally: length, first non-space character, whether a
  markdown fence is present, how many **top-level** JSON objects it contains,
  whether it parses, and its keys if so. Exists because every question an
  unusable completion raises is structural (prose wrapper? fence? two objects?
  empty?) and **none of them needs the text.** Logging the text instead publishes
  whatever the model echoed back — on a debate/chat/support endpoint that is the
  visitor's own words — to anyone who can read the container's logs.
  It is also strictly MORE diagnostic than a capped excerpt: `bugs_closed/088`'s
  second JSON object begins ~1,500 chars in, so a 300-char excerpt could never
  reveal the case it existed for. `TopLevelJSONObjects` is exported separately and
  is string- and escape-aware, because a brace inside a quoted value otherwise
  miscounts and the count is the whole point.
- **sources:** `platform/aiservice/fingerprint.go`; `platform/aiservice/fingerprint_test.go`;
  consumer at `internal/tools-api/handlers/ailog.go`
- **relations:** LCO-006; CNV-003; `bugs_open/083`; council corr `e004fd81`
- **note:** the owner ruled on 2026-07-27 that no model text is logged, on an item
  the council explicitly recorded it could not close itself.

### LCO-006 — A 5xx with a discarded error is undiagnosable, and bursty faults cannot be reproduced
- **status:** deployed (tools-api on the island VM; **not** in the chassis)
- **what:** Every LLM-backed handler in `tools-api` discarded `err` before
  returning 503, so a 429, a 529, a network timeout with no client timeout, a
  truncated completion and a malformed response all reached the caller as one
  opaque message. Now every **5xx fault path** logs its cause (16 sites), with
  truncation labelled distinctly via `aiservice.IsTruncated` because that is our
  own cap and not an upstream fault. **4xx caller paths are deliberately NOT
  logged** — caller mistakes are not faults and `gin.Logger()` already records
  method/path/status. Two structural findings came with it: the service had **no
  request logging at all**, so there was no denominator and no honest failure
  *rate* could be quoted; and `/round` returned a literal **502**, whose body
  Cloudflare replaces with its own HTML, destroying the JSON error the front end
  needed — the one status code that eats its own evidence.
- **sources:** `internal/tools-api/handlers/ailog.go`, `.../defend.go`,
  `.../position.go`, `.../round.go`, `internal/tools-api/api/server.go`
- **relations:** LCO-005; `bugs_open/083`; `RUNBOOK_gauntlet_dead_cta.md` §5
- **note:** proven by an INDUCED fault, not a green path — the same invalid key
  against both images: the new one logs `status 401 … invalid x-api-key`, the old
  one logs nothing at all.
