# Register — llm-call-observability

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

_Concept count retired 2026-08-09 — derived, not stored; run the drift pair in `000_concept_index.md`, or read `concept-register-drift-check`'s daily row (DOC-074)._ (4 consolidated from 5 raw extractions across unit U13, one exact
duplicate pair collapsed; LCO-005/006/007 added post-freeze).

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

### LCO-007 — `fleet-step-token-pressure`: a standing headroom check over every non-review LLM step
- **status:** deployed (live `scheduled_tasks` row, seeded and proven 2026-08-06; first note `3186dcfa`)
- **what:** A CTE-only scheduled task (`fire_message=false`, 6-hourly) that measures,
  per `(step_name, cap)` over a 90-day window, how close each non-review LLM step's
  output distribution runs to its token cap — p95, peak, and truncation count — and
  writes ONE `doc_notes` row when the flagged set CHANGES (md5 digest in
  `subject_key`, 30-day dedup look-back: an event, not a heartbeat). Thresholds:
  **C died on the CLOCK** · T any truncation · N near-miss peak ≥ 95% of cap ·
  P pressure p95 ≥ 85%, at n ≥ 5 (no n floor for C).
  It is **FIX-058 (`council-seat-token-pressure`) generalised past `review_%`**; the
  two tasks partition the fleet exactly, and the `review_%` naming convention IS the
  contract between them. `fire_message=false` means the `pre_query` is the whole
  work — no Kafka, no orchestration, no LLM, no credits, and it cannot wake a chassis
  pod. Exists because a cap is live DB config: `bugs_open/183`'s step sat at p95 92.5%
  of the fleet's only 6000 cap for months, then burned every attempt on several sites
  in one afternoon, and the failure presented as SITE-specific because nothing was
  watching the distance to the ceiling.
- **sources:** `docs024_key_docs_latest/bugfix_183_step_token_pressure/SQL_2026-08-06_fleet_step_token_pressure_task.sql`
  (the seed, with its reasoning); `.../RUNBOOK_step_token_pressure.md` (how to run,
  read and re-verify it); `bugs_open/183` candidate 4; found `bugs_open/205`
- **relations:** FIX-058 (the council-seat original, and its still-open question of
  whether the near-miss threshold should scale with cap size — inherited unanswered);
  LCO-002 (the cap fallback chain this measures the OUTPUT of); MDL-039 /
  `bugs_open/009` (root-block shadowing — sidestepped by reading observed caps from
  `llm_call_log` rather than a definitions jsonb path); RFC_006 (live config is
  guarded by a scheduled check, never a commit-time hook); RFC_012 second sitting
  ("online within the framework")
- **C AND T WANT OPPOSITE FIXES — this is the reason C is a separate kind (added
  2026-08-06).** T means the cap was REACHED: raise it, or shrink the unit. C means the
  cap could NOT be reached, because the call died on wall-clock first — the chassis does
  not stream (`aiservice/anthropic.go:72` and `gemini.go:185` are 600s blocking HTTP
  clients; `ollama.go:55` is 120s). **Raising a cap in response to a C is actively wrong**
  — the bigger number is unreachable too. Scoring a clock kill as `frac = 1.0` would have
  merged the two signals and produced exactly that recommendation, which is why it is
  counted separately and outranks T. The clock converts to a token ceiling: ~98 tok/s on
  Sonnet 5 and 47–82 on Sonnet 4.6 ⇒ ~58,000 and ~28,000–42,000 tokens respectively.
  Two sub-decisions: the C arm requires **no known cap** (the only real case in the
  history, `scrape_prices` Apr 2026, recorded none), and `context canceled` needs a
  **480,000 ms floor** or it fires on every pod restart.
- **LANDMINE — the check reports a NUMBER; the SHAPE is a different question.** A `T`
  flag means "this step truncated", not "this step's cap is too small for its
  workload". On its first run the top line (64 truncations) was **two byte-identical
  prompts re-dispatched for 34 hours** — a poison-pill loop, where a bigger cap is
  candidate 2 at best (`bugs_open/205`). Before acting on a flag, group the step's
  calls by `md5(prompt_rendered)`: many distinct prompts near the cap is cap drift;
  one or two prompts repeating is a stuck item, and raising the cap treats the
  symptom.
- **verify-later:** the pinned known-case test in the RUNBOOK (as-of 2026-08-01 it
  must flag `classify_and_extract@6000` as P, BEFORE that step's first truncation) —
  re-run it if the pre_query is ever edited; a check that cannot flag its own
  motivating case is inert.

### LCO-008 — `CacheBreakpointMarker`: opt-in Anthropic prompt caching on the shared AI client
- **status:** deployed and PROVEN IN PRODUCTION (chassis v1.0.1283, 2026-08-10 22:03 UTC)
- **status-evidence:** proven at the artefact, not inferred from a roll. Pod-grepped on **both** replicas (marker present, `cache_read_input_tokens` present, `anthropic-version` as a pipeline positive control, and a never-existed string returning 0 as the negative control). Then the first council run on the new binary showed the mechanism working end to end: `review_editquality` (seat 1) `cache_creation=102,088 / cache_read=0`; `review_constitution` and `review_mission` (seats 2–3) `cache_creation=0 / cache_read=102,088`. **`input_tokens` collapsed from ~100,000 to ~1,800 per seat** — the uncached remainder is exactly the seat-specific persona, which is the arithmetic signature of the split working as designed. `input + creation + read` holds steady at ~103,870, confirming the prompt did not shrink; only the billing changed.
- **MEASURED SAVING: 68%**, over 46 seats (38 reading cache) — 4,900,476 true prompt tokens billed as 1,566,555 equivalent. **This is lower than the ~80% predicted, and the gap is instructive rather than a defect:** the estimate assumed one cache write per run, but several council submissions run concurrently and each carries its own `plan_json`, so each pays its own write. **The saving is per-run and dilutes with concurrency** — quote 68% (measured) rather than 80% (modelled), and expect it to vary with how many councils overlap.
- **earlier status, kept for the trail:** "built, unrolled" — code + migration committed 2026-08-10, inert until both the image rolled AND a caller added the marker. Both happened the same day.
- **what:** A sentinel string `<!--CACHE_BREAKPOINT-->` that a prompt template may contain. The Anthropic client splits on it and sends `[shared prefix (cache_control ephemeral, ttl 1h)][per-call suffix]` instead of one flat user-message string. **⚠ THIS LINE WAS FALSE FROM 2026-08-10 TO 2026-08-15 AND NOTHING CAUGHT IT.** The council removed `ttl:"1h"` before the code shipped (see the verdict bullet below) and the entry's `what` was never updated to match, so it asserted an hour while production sent no `ttl` field at all and ran the 5-minute default. It is true again as of the owner ruling of 2026-08-15 — but the failure mode is the point: **a register entry can be corrected in one bullet and left wrong in another, and council seats read this file as ground truth.** When a verdict changes what shipped, fix the `what`, not only the verdict trail. Absent the marker the request is byte-identical to before, so the ~40 other agents on this seam are structurally unaffected — the unsafe default is OFF, per the RFC_010 owner ruling that new authority on a shared seam ships as an opt-in field rather than a documented contract. Reads back `cache_creation_input_tokens` / `cache_read_input_tokens` (previously discarded) into `options.__usage_*`, and `llm_call_logger.cacheTelemetry` persists them to the two new columns.
- **the economics it exists for:** council-gate sends ~100k input tokens to each of 17 SEQUENTIALLY-chained seats per submission — measured 2026-08-10: 119 calls / 11.6M input tokens in 24h, ~85% of all fleet LLM spend, against 0.37M output. The seats' prompts share their whole evidence body (every 10–90% slice of one appears verbatim in another) but had a **zero-character common prefix**, because each seat's persona sat at the top. Sequential execution is what makes this cacheable at all: the skill-documented fan-out hazard (parallel requests cannot read a cache still being written) does not apply to a chain.
- **⚠ LANDMINE — a byte above the marker costs money and returns NOTHING, and it looks exactly like success.** Anthropic caching is a PREFIX match. A timestamp, a UUID, a run id, a per-seat name, or a reordered JSON key anywhere before the marker makes every call write a fresh entry and read none — strictly worse than no caching (you pay the 1.25×/2× write premium for zero reads), with no error, no warning, and correct-looking answers throughout. **The tell is in the data, not the logs:** `cache_read_input_tokens` stays 0 across calls that should be hitting. Before believing caching works, assert a NON-ZERO read on the 2nd+ call of a run — a zero is the failure mode, not the absence of one.
- **⚠ LANDMINE — `input_tokens` silently changes meaning the moment a caller opts in.** The API reports it as the UNCACHED REMAINDER, so a cached seat logs ~5k where the prompt was ~100k. Every existing cost query in this estate reads `input_tokens` and will understate by ~95% — **in the flattering direction**, reporting a saving far larger than the real one. True prompt size is `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`; migration 376 states this on the columns themselves. NULL in those columns means "binary predates cache support", 0 means "no cache used" — do not collapse them.
- **open review question:** the marker is a magic string in template text, which means a template author can silently disable caching by editing above it and nothing will complain. A lint (does any council template interpolate a per-seat value above its marker?) would close that, and does not exist. Raised deliberately rather than left for discovery.
- **council verdict `b54f173e-ebd4-45c4-954a-dfc70005e62c`: APPROVED, 5 advisory objections, 2 acted on before the code shipped.** Both are worth knowing because both were invisible from the tests as first written:
  - **`ttl:"1h"` was REMOVED (edit-quality seat, medium).** The extended TTL is gated behind a beta header this client does not send, so the first caller to opt in would have got a **400, not a cache hit** — and since council-gate reviews every platform change, that 400 takes out the review path for the whole estate rather than merely losing a saving. Now sends no `ttl` field at all (the 5-minute default needs no header on any model). **The asymmetry is the argument: too-short TTL = a worse saving, visible instantly in `cache_read_input_tokens`; an unsupported ttl = an outage.** If measurement later shows late-chain seats missing, confirm the current beta header FIRST — do not reintroduce `ttl` speculatively.
    - **↑ SUPERSEDED 2026-08-15 (owner ruling, decision 6 of six) — the confirmation was run and THERE IS NO HEADER TO SEND.** `cacheTTL` is now `"1h"` (commit `c5010ac26`, council corr `176d921e`). Probed the live account from inside a chassis pod, no beta header, on both models the fleet uses: `claude-sonnet-5` returned HTTP 200 **and credited the write to the 1-hour bucket** — `cache_creation: {ephemeral_5m: 0, ephemeral_1h: 6003}` — which is what proves the TTL is honoured rather than merely tolerated; `claude-sonnet-4-6` returned HTTP 200 with no 400 (stated limit: that call returned a cache READ with 0 in both creation buckets, so it proves acceptance, not the bucket). **The seat's reasoning was right and only its factual premise expired** — the 400-asymmetry argument still governs any future revert. **WHY IT MATTERED ENOUGH TO REVISIT:** at the 5-minute TTL, `content-gap-planner` hits cache on **1.0%** of repeat calls (99.8% at an hour), so enabling the marker on it would have cost **~24% MORE** than not caching at all — the saving was not being left on the table, it was unreachable. **⚠ THE WRITE MULTIPLIER RISES 1.25x → 2x, so break-even moves from a ~22% hit rate to ~53%:** an agent adopting the marker below that is now worse off than it would have been at 5m. Measure the gap between same-prefix calls before adopting. **INERT UNTIL A ROLL** — Go change; the marker is DB config and live immediately, so do not add a marker to a low-hit-rate agent before this ships. **COUNCIL APPROVED 2026-08-15, corr `176d921e`, 'all reviewers approve', 12 clean approvals — after TWO revise rounds, both gated by `prior_art_librarian` for the same reason: the blast-radius facts were narrated in the rationale instead of attached in `grounded_in` (`WRONG_CALLS` 2026-08-15). The code never changed across the three rounds; only the evidence did.** **POST-ROLL GATE, still owed and part of the approved plan:** enumerate the pods running this binary by IMAGE (never by `-l app=agent-chassis`, which returns 2 of many), confirm one tag across the set, re-probe from a rolled pod AND from a non-`agent-chassis` service, confirm council-gate's next real round still logs non-zero `cache_read_input_tokens` (a zero is the documented failure signature and looks exactly like success), and confirm no fleet-wide 400 spike. Revert the constant to `""` if the probe or the 400 check fails — the field handling is correct for both values.
  - **A marker at position 0 leaked the literal marker text to the model (edit-quality seat, low).** The fallback path returned the prompt unchanged instead of stripping it. **The original test passed throughout**, because it asserted the *type* of the content (still a string — true) and never its *content*. A type-only assertion will keep passing through this whole class of defect; the test now asserts the marker's absence and was mutation-proven against the exact bug.
- **blast-radius audit the `bug_historian` seat asked for (measured 2026-08-10, answers its objection):** the `input_tokens` meaning-change affects **zero live automated consumers**. Only two `scheduled_tasks` pre_queries reference `llm_call_log` at all — `council-seat-token-pressure` and `fleet-step-token-pressure` — and **both read `output_tokens` only**, neither `input_tokens` nor `total_tokens`. Note `total_tokens` *would* be affected if anything read it (migration 246 records that it includes `input_tokens`), and nothing does. Re-run that audit before assuming this stays true: `SELECT name, pre_query ILIKE '%input_tokens%' FROM scheduled_tasks WHERE pre_query ILIKE '%llm_call_log%';`
- **SECOND ADOPTER, 2026-08-15: `content-gap-planner` (`plan_gaps`), migrations `413`+`415`+`414`. Adoption is now 2 of ~191 live agents — the "council-gate only" figure quoted in handoffs up to 2026-08-15 is superseded.** First live call: `cache_creation=4,991 / cache_read=0 / input_tokens=1,400` — the write fired and the cacheable prefix is **78%** of the prompt. Marker sits immediately before `## Content Gap to Address`; placement was **measured, not reasoned** — that anchor falls at 10,103–11,875 rendered chars across the six dominant groups with `count(DISTINCT prefix-at-boundary) = 1` in **every** group (393 of 404 calls in six groups), which is the exact byte-identity precondition prefix caching requires. The ~11 singleton sites each pay a write they never read back; that cost is inside the ~82%-saving arithmetic rather than excluded from it.
  - **⚠ ADOPTING THE MARKER FORCED A MODEL MOVE, AND THIS IS THE PART TO CARRY FORWARD.** The 1h bucket is proven on `claude-sonnet-5` **only** (see the superseding bullet above — on `claude-sonnet-4-6` the probe proves *acceptance*, not the bucket). All 17 council-gate seats run `claude-sonnet-5`, so **100% of this mechanism's production evidence is sonnet-5 evidence**. `content-gap-planner` ran `claude-sonnet-4-6`, where its hit rate is 1.0% at 5m against a ~22% break-even — so marking it there would have bet the whole saving on an unverified bucket, with the documented silent failure signature (writes, no reads, ~24% worse than not caching). **Owner ruling 2026-08-15: move the agent to the proven model rather than hedge.** Generalises: **before adding a marker, check the agent's model against the models the TTL has actually been proven on — not merely against the models that accept the field.**
  - **⚠ AND THAT SWAP HAS A SECOND-ORDER COST NOTHING WARNS ABOUT.** `claude-sonnet-4-6` omitting `thinking` runs thinking OFF; `claude-sonnet-5` omitting it runs **adaptive thinking**, and `max_tokens` caps thinking PLUS response together — so a budget sized for the answer alone can truncate structured output. **This client cannot disable it**: its only thinking path emits `{"type":"enabled","budget_tokens":N}`, which `sonnet-5` rejects with a 400; no path emits `{"type":"disabled"}`. Headroom is the sole lever (raised to 16000, the non-streaming ceiling). **Corollary: any agent already setting `budget_tokens` will 400 on EVERY call the moment its model moves to sonnet-5/opus-5 — check before switching.** Also: sonnet-5 bills ~30% more tokens for the same text at identical sticker price (measured here: 6,391 vs 4,765 for the same work, **+34%**), and its introductory pricing ran only to **2026-08-31**, so any August figure understates the steady state.
  - **⚠ `max_tokens` written to `...steps.<step>.config.max_tokens` is INERT** — `ai_actions.go` resolves top-level `default_config.max_tokens`, then the merged `ai_service` block, then the client's `2048`; `agentConfig` is `agentDef.DefaultConfig`, **not** the step's config despite the name. Migration `413` wrote the inert key and its own post-condition asserted that same key, so it could not fail; `415` corrected it and asserts the **resolved** value. Full trap in `LANDMINES.md` (2026-08-15) and `WRONG_CALLS.md`.
- **sources:** `platform/aiservice/anthropic.go`; `platform/aiservice/prompt_caching_test.go`; `platform/orchestration/actions/llm_call_logger.go` (`cacheTelemetry`); `docs/agent_docs/sql_for_agents/376_llm_call_log_cache_token_columns.sql`; `docs/agent_docs/sql_for_agents/413_content_gap_planner_to_sonnet_5.sql`; `docs/agent_docs/sql_for_agents/415_content_gap_planner_max_tokens_at_the_key_the_code_reads.sql`; `docs/agent_docs/sql_for_agents/414_content_gap_planner_cache_breakpoint.sql`
- **relations:** `llm_call_log` schema (LCO-001, LCO-007 token-pressure checks read `input_tokens` and are affected by the meaning change above); council-gate roster (CLAUDE.md § council review); RFC_010 opt-in-field ruling
- **verify-later:** ~~whether any caller actually carries the marker (until one does this is dead code); first observed `cache_read_input_tokens > 0` in production~~ **both SETTLED 2026-08-10 (council-gate) and re-confirmed 2026-08-15 (content-gap-planner's first write)**; whether LCO-007's pre_query needs updating for the `input_tokens` meaning change. ~~**STILL OPEN:** a `cache_read_input_tokens > 0` on a `content-gap-planner` 2nd+ call … and the fleet TTL gate~~ — **BOTH CLOSED 2026-08-15 14:04Z, in one observation.** Second call read back `cache_read=4,991 / cache_creation=0`, and the gap to the write was **8m46.566s**. That single figure is the artefact-level proof of the 1-hour TTL: at the old 5-minute TTL the entry would have expired at ~14:00:06 and the call would have forced a rewrite, which is exactly what the **pre-roll control shows happening 29 times out of 29 (0 reads, 28 writes, 2026-08-12 → 08-15)**. Nothing could have refreshed it mid-gap — only two calls exist in the window and the prefix is site- and template-specific. **⚠ Note WHICH agent closed it: `council-gate` cannot.** Its traffic is too dense to open a >5min gap (widest with a read in the same window: 1m28s), so the gate query returns 0 for it indefinitely — **`content-gap-planner` is now the fleet's TTL sentinel and is where this check should be run.** `llm_call_log` still stores totals only, with no 5m/1h bucket breakdown, so a behavioural gap-read remains the only log-based proof available.

### LCO-009 — Silent row-cap detection at `query_database` (a result set the size of its own LIMIT)
- **status:** built, committed, **council APPROVED**, **NOT LIVE** (Go — inert until the next chassis roll). Both migration halves (445 + 446) of the same lane ARE live.
- **status-evidence:** commit `eb137faed` 2026-08-16; **council APPROVED round 2** (corr `b684a399-bb4d-4b1f-82f0-fe1429ebdceb`, *"3 advisory objections, none high-severity"*, `gated_by_truncation: false`). Round 1 was REVISE and changed the shipped work — it caught that migration 445 had traded a silent ROW cap for a silent COLUMN cap, fixed by 446. A round-2 advisory also found a real detector gap: a trailing SQL comment (`LIMIT 30 -- note`) was a FALSE NEGATIVE, silent under the very mechanism built to end silent caps; the regex now tolerates `--` and `/* */`, mutation-proven both ways. Nine mutations, all killed; two changed the code (see below). Package suite green on a clean `git archive HEAD` tree.
- **what:** `QueryDatabaseAction` now parses its query's **trailing literal `LIMIT n`** and, when the returned row count reaches it, logs a WARN naming the step, agent type, limit and row count. `platform/orchestration/actions/query_row_cap.go` holds the two helpers (`queryRowCap`, `resultHitItsRowCap`). **`LIMIT 1` is deliberately excluded** — 19 of 26 live hits are the fetch-one/claim-one idiom, which returns one row by design on every execution and would make the channel pure noise. A **parameterised** `LIMIT $2` and a LIMIT inside a subquery or CTE deliberately do not match (no literal to compare; a non-trailing limit bounds a different set).
- **why another workstream needs to know it exists:** **a row-count `LIMIT` feeding an LLM prompt is invisible by construction** — the model returns plausible output whether it saw 30 rows or 74, so there is no wrong output to notice. Census 2026-08-16: **26** literal LIMITs in query-shaped step configs across live agents, of which **seven are multi-row caps** that can bite this way — `tool-suggester.load_library_tools` 30, `internal-linker.load_candidate_pages` 15, `model-directory-trigger.find_directory_sites` 12, `tool-recreation-handler.load_related_context` 10, `content-feed-trigger.find_news_sites` 5, `visual-design-auditor.load_design_context` 5, `fix-proposer.load_last_bundle` 2. **If you add a `query_database` step with a multi-row LIMIT, this will warn when it fills** — that is not a bug report against you, it is the check asking whether the cap is a judgement or an accident.
- **deliberately OBSERVATIONAL — it takes no authority on the seam.** It cannot change a query, a result or a row; `QueryDatabaseAction`'s guarantee is byte-identical afterwards. That is the 2026-08-02 §2 / RFC_010 shape applied to a shared action serving 26+ live steps: there is nothing to default OFF because there is nothing switched on.
- **what it is NOT:** it does **not** tell you "30 of 74" — only that the result reached its ceiling and *may* be truncated. A definitive count needs either a second `COUNT(*)` per step (doubling DB work fleet-wide) or a `LIMIT n+1` probe that trims the extra row — the latter changes what a shared action RETURNS and is the recorded follow-up, not part of this. **Known false positive, stated:** a population exactly equal to its cap warns while nothing is hidden. That is the right trade — a false "go and check" costs one query; the false silence cost 44 of 74 tools.
- **sources:** `platform/orchestration/actions/query_row_cap.go`, `query_row_cap_test.go`, `database_actions.go`; `bugs_open/275`; lane `docs024_key_docs_latest/bugfix_275_silent_row_caps/`; migration `445_tool_suggester_whole_library.sql` (the instance fix)
- **relations:** LCO-007 (`fleet-step-token-pressure` — the sibling "is a budget being hit" monitor, but for LLM output tokens, and it reads `llm_call_log`); MDL-041 (`bugs_open/257`, budget resolved at the provider client — same doctrine, the other end of the prompt); OPP-003 (why this ships a BEHAVIOURAL call-site test, not a source scan); `bugs_open/242` (*"a capped render audit is indistinguishable from a complete one"* — the same class in a different subsystem, still open)
- **⚠ READING THE WARNING — not every multi-row cap is a defect, and the SQL cannot tell you which.** Measured 2026-08-16, the census corrected twice: of the 26 literal LIMITs, 19 are `LIMIT 1` (fetch-one), **2 are INSIDE a subquery** (`fix-proposer.load_last_bundle`, `visual-design-auditor.load_design_context` — their outer result is ONE row, so the end-anchored regex correctly ignores them and flagging them would have been wrong), and **5 are whole-result caps**. Of those five the distinguishing question is what the rows ARE: a **WORK QUEUE** takes N per run and the rest arrive next run (`find_news_sites` 5 of 9; `find_directory_sites` 12 of 3, `ORDER BY random()`) — coverage is eventual, not a defect; a **CORPUS shown to a model** takes N and the rest are never seen, and the model answers confidently either way — that is the defect. **Expect the WARN to fire on work-queue steps.** That is the check asking a question only a human can answer, not a false positive.
- **TWO UNFIXED INSTANCES, measured 2026-08-16 and filed nowhere** (grepped `bugs_open/` + `bugs_closed/`): **`tool-recreation-handler.load_related_context` LIMIT 10 against up to 107 pages** (worse in ratio than `bugs_open/275` itself — 9% vs 41%) and **`internal-linker.load_candidate_pages` LIMIT 15 against up to 68**, so a page's internal links are chosen from at most 15 of 68 candidates, ordered alphabetically. Both feed an LLM a truncated corpus. **NOW FILED as `bugs_open/297` and `bugs_open/298`** (owner directed 2026-08-17). 297 is LIVE (290 `llm_call_log` rows) and bites at 19 of 24 sites; 298 bites at 8 of 24 but its reachability is explicitly UNMEASURED — zero `llm_call_log` rows for `internal-linker` ever, which **this detector will answer by itself once the chassis rolls**.
- **verify-later:** post-roll, whether the WARN actually fires in production and on which steps — that is a live census nobody has run; and the `LIMIT n+1` probe that would turn the suspicion into a definitive "30 of 74"
