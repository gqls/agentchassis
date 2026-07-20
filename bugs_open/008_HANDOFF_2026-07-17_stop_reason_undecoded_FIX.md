# HANDOFF — FIX: `GenerateText` never decodes `stop_reason` (silent max_tokens truncation)

> ## STATUS 2026-07-20 — ALL ITEMS NOW FIXED IN CODE; OPEN ONLY UNTIL THE IMAGE ROLLS
>
> **Read this before §3–§8, which describe work that has since shipped.** The
> sections below are preserved as the diagnosis and decision record, not as a
> to-do list.
>
> | item | state | where |
> |---|---|---|
> | 1–2. decode `stop_reason`, fail loud on `max_tokens` | **DONE, in image** | `f32b208e5`, `anthropic.go:163,195` |
> | 3. retry semantics | **DONE** — typed error; the retry loop (`ai_actions.go:428`) retries 5xx only, so a cap is never re-spent | `f32b208e5` |
> | 4. siblings | **DONE** — `ollama.go:158` decodes `done_reason=="length"` | `f32b208e5` |
> | 5. `stop_reason == "refusal"` | **DONE, NOT YET IN AN IMAGE** | `45e90acbb`, `anthropic.go` + new `refusal.go` |
> | §8 CI guard (owner D1/D2) | **DONE, NOT YET IN AN IMAGE** | `stop_signal_test.go` |
>
> Beyond the original scope, `f32b208e5` also made `TruncatedError` carry the
> **partial text** (`truncation.go`), which is the transport-layer half of
> `bugs_open/019` — the case file there had assumed the partial was unrecoverable.
>
> **Why this file is still in `/bugs_open/`:** the bar is *fixed AND live*. Item 5
> and the CI guard are committed but inert until a chassis image is rebuilt and
> rolled. Verify against the running pod, then move to `/bugs_closed/`:
> ```
> kubectl exec -n ai-persona-system <pod> -- sh -c \
>   'strings /app/agent-chassis | grep -c "model declined to answer"'
> ```
>
> **A trap this file caused, worth carrying elsewhere:** a triage pass on
> 2026-07-20 read the working tree, found the refusal fix and the test file, and
> reported item 5 as already shipped in `f32b208e5` — which contains neither. It
> had read another session's uncommitted work. In this repo "verify against
> current code" must mean `git show HEAD:<path>`, not the tree. Pattern filed in
> `016b §9`.

**Filed:** 2026-07-17, from the "diagnosis fixloop 3" thread. Cold-start for a fixing
thread. **Diagnosis is DONE — the loop CONFIRMED it** (see §2); this handoff is about
the FIX only.

**Severity:** High, platform-wide. Any LLM call that hits its `max_tokens` cap returns
a mid-generation-truncated string as a *successful* result. 17 proven occurrences
across 5 agent types; this mechanism is upstream of the article-body blanking
incident (truncated JSON stored, then rendered empty).

## Working rules (hold these)
Go, not Python. British English. Schema first (`\d <table>` before SQL); read the
function before changing it. Structural fixes over patches; reuse existing functions.
Go changes are inert until a chassis image ships. **Deploy from a committed ref with
`make build-agent-chassis-ref`** (exists as of 2026-07-16 — never build from the
working tree; it bundles other sessions' WIP). Bump IMAGE_TAG (makefile line 16),
never rebuild an existing tag; verify by grepping the RUNNING POD's binary for a log
string from your change. **Commit per task, `git add <explicit paths>` only, and read
`git diff --cached --name-only` before committing** — other sessions leave files
staged in the shared index (this exact trap fired on 2026-07-17).
DB: `PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`.

## 1. The mechanism (confirmed)

`platform/aiservice/anthropic.go` → `GenerateText`:

1. Builds `requestBody` with hardcoded `"max_tokens": 2048` (overridden only if
   `options["max_tokens"]` is set).
2. POSTs to `/v1/messages`, reads the HTTP 200 body.
3. Unmarshals into a struct containing ONLY `Content []{Type,Text}` and
   `Usage {InputTokens,OutputTokens}` — **`stop_reason` is not a field, so it is
   discarded at parse time**.
4. Returns the first `text` block as a successful result.

When the API stops at the cap it returns 200 + `"stop_reason":"max_tokens"` + a body
cut off mid-generation. Step 3 makes that byte-for-byte indistinguishable from a
complete response at this layer and every layer above. Downstream, the truncation
surfaces (if at all) as a JSON-parse failure attributed to a "malformed envelope",
never to the cut-off — or is stored/served as-is.

## 2. Diagnosis provenance (do not re-diagnose)

- Loop run **CONFIRMED**: correlation `e505f70f-b9e2-4654-9942-30fb13731ca9`
  (2026-07-16), intake `needs_diagnosis:stop-reason-undecoded` (closed, terminal in
  its `error` field). Citations: the response struct; the text-block return loop; and
  a state-tier citation of live `llm_call_log` rows the loop fetched via its own
  data_request. Graded PASS against the pre-registered rubric
  (`fixloop_eg_dartsonline/RUBRIC_2026-07-16_two_config_bugs.md`).
- Bundles/verdicts: `diagnosis_artifacts` by that correlation_id.

**Evidence query (re-run to confirm the premise still holds before fixing — cases
shift under handoffs; this one did three times on 2026-07-16):**
```sql
SELECT agent_type, step_name, max_tokens, output_tokens, success
FROM llm_call_log
WHERE output_tokens >= max_tokens AND output_tokens IS NOT NULL
  AND max_tokens IS NOT NULL
ORDER BY created_at DESC LIMIT 20;
-- 17 rows as of 2026-07-16, ALL success=true, error_message NULL
```

## 3. The fix (sketch — the fixing thread owns the final shape)

In `GenerateText` (`platform/aiservice/anthropic.go`):

1. Add `StopReason string \`json:"stop_reason"\`` to the response struct.
2. After unmarshal, if `StopReason == "max_tokens"`, return an **error** naming the
   cap and usage (e.g. `output truncated at max_tokens=%d (output_tokens=%d)`), NOT
   the partial text. Fail loud at the client boundary — this is the whole fix.
3. **Retry semantics matter:** callers wrap this in "AI call failed after 4 retries".
   Retrying an identical request against the same cap re-truncates identically — 4x
   wasted spend. Prefer a typed/sentinel error the retry loop does NOT retry (or
   detects and raises max_tokens on retry — owner's call; simplest correct behaviour
   is fail-fast, no retry).
4. Check the SIBLINGS: grep `platform/aiservice/` for any other function that
   unmarshals a Messages API response (streaming path, tools path, other providers'
   adapters with equivalent semantics). This handoff verified ONLY `GenerateText`.
5. Optional, same commit: handle `stop_reason == "refusal"` explicitly (Sonnet 5+
   returns it; currently it would surface as "no text content in response").

**Write a unit test** with a canned 200 body carrying `"stop_reason":"max_tokens"` —
assert error, not text. There is precedent for fixture-style tests in
`json_envelope_test.go`.

## 4. Two routes to ship it — pick one

**(a) Through the loop (proves F1 on a real case #2):** fire
`091_TRIGGER_fix_proposer` on correlation `e505f70f` → council (now 3 seats incl.
bug-historian, fix-proposer v6) → `092_TRIGGER_fix_implementer` (MUST go via
fix-implementer-orchestrator) → build gate → PR. **Gotchas:** implementer ref/base
were live-set to `084_site_improvements_local_ai` (stale — update to the current
branch or make ref a per-run input, the open F1.2 cleanup); delete any stale `fix/*`
branch before re-firing; nothing merges itself.

**(b) By hand:** small enough for a direct fix + PR. Still commit per task and
deploy from ref.

Either way the PR is the gate — no direct-to-main.

## 5. Verification after deploy

- Pod binary: `strings /app/agent-chassis | grep -c "<your new error string>"`.
- Behavioural: temporarily set a tiny max_tokens on a scratch agent def, fire once,
  expect a LOUD step failure + `llm_call_log` row with `success=false` and the new
  error text; then restore.
- The 17 historical rows stay `success=true` (they predate the fix) — new capped
  calls must not.

## 6. Related (do not conflate)

- **009 (root ai_service shadows step)** is the OTHER half of why callers sat at
  2048 — different bug, same neighbourhood. Fixing 009 changes which config wins;
  fixing THIS bug makes the consequence loud. They compose; ship separately.
- The render-side guard ("refusing to render an empty section", live since
  v1.0.1126) already catches the article-body consequence — it does NOT cover the
  other 4 capped agent types. This fix is the platform-level closure.

## 7. UPDATE 2026-07-17 — the loop ran F1: council ESCALATED, and the objection is right

Route (a) was fired from the tool thread (fix_correlation `e505f70f`, run
`ca064df2`). Three full rounds: editquality **approve**, guardian **approve**
(no hard veto), **bug-historian OBJECT every round** → revise cap → decision
`exhausted` → **escalation artifact on `e505f70f`** (the full hand-off package;
`decided_by: "objection from bug_historian — revise cap reached"`).

The historian's objection (its first live vote): the edit is correct but may
patch ONE call site of a generic mechanism — "does the codebase have other LLM
provider adapters?" **Answered from source: YES — `platform/aiservice/ollama.go`
has its own `GenerateText`.** So the fixing thread should treat §3's
"check the SIBLINGS" as CONFIRMED WORK, not a maybe: either cover ollama.go's
equivalent (its stop semantics differ — check its `done`/`done_reason` fields)
in the same PR, or file the follow-on item explicitly, per the historian's own
framing. The escalation artifact carries the 3 plan revisions + all council
reports — start from it rather than re-running 091.

Loop residual this exposed (tool thread's problem, not this thread's): the
historian's blocking question was CODE-shaped; the verify tier only runs SQL
(`run_checks`), so the loop could not self-resolve it and correctly escalated —
the F2.3b(c) code-lookup check tier is now demonstrated-needed on a real case.

## 8. UPDATE 2026-07-18 — council-APPROVED, and a CI-guard to bundle into the PR
- **BUG A's fix plan is now COUNCIL-APPROVED** (fix-proposer run on e505f70f,
  all seats incl. bug-historian, round 3). The approved plan already covers BOTH
  provider adapters (anthropic.go + ollama.go) — the code-lookup tier confirmed
  ollama.go's `(*OllamaClient).GenerateText` is the only other implementation.
  You can take the approved plan straight to the implementer (092 → build gate →
  PR); it does not need re-proposing.
- **BUNDLE THIS TEST INTO THE PR (owner decision D1/D2, 2026-07-18):** the
  bug-historian approved WITH one advisory residual it raised twice — nothing
  stops a FUTURE third provider adapter being added without the stop_reason
  guard. Add a Go test (or a `discovery_check`) in the same PR asserting every
  provider client's `GenerateText`-equivalent decodes its stop/finish signal and
  fails loud on truncation — so a new adapter that skips it fails CI, not
  production. Ship the guard with the mechanism it guards; do not defer it to a
  separate workstream.
- Validation baseline the historian handed you: **23 historical
  silently-truncated `llm_call_log` rows** (was 17; more accrued) should replay
  post-fix as `success=false` with the new error string, not `success=true`.
- Family note (per CLAUDE.md § Debugging): 005/008/009/012 are ONE
  truncation-and-config family. `bugs_open/012` is a concrete instance of the
  consequence (a 10,272-char component saved back as 1,253 chars, reported
  success) — the same silent-CUT the `output_tokens == max_tokens` rule now
  guards. This fix (fail loud at the client boundary) closes the ROOT of that
  family for the LLM-generation path.

---

## REAL-CASE EVIDENCE for item 5 (experience-loop thread, 2026-07-18)

Item 5 above — handle `stop_reason == "refusal"` explicitly, "currently it
would surface as 'no text content in response'" — was filed as **optional**.
It has now happened in production, exactly as predicted:

```
step review_feasibility failed: failed to execute action execute_llm_prompt:
AI call failed with unhandled error: no text content in response (had 1 blocks)
```

`experience-planner` / `review_feasibility`, 2026-07-18 15:52, on
`claude-sonnet-5`. It killed a council round outright — the workflow routed to
`error_step` and the whole run terminated at `complete_refused` after round 1,
discarding four critics' work and ~6 minutes of LLM spend. The only other
occurrence in 7 days is `diagnose-agent`/`verdict` on 07-16, i.e. this
document's own originating case.

Two things this adds:
1. **Item 5 is not optional.** Sonnet 5 is now the default model for several
   agents (migration 158 moved the travelling-docs family), so the shape that
   produces this is in the fleet's hot path, not a corner.
2. **"had 1 blocks" is diagnostic and should be preserved.** A non-text single
   block is the tell. Whatever the fix decodes (`refusal`, or a thinking-only
   response), the error text should name the `stop_reason` and the block type
   rather than just the absence of text — otherwise the next person gets the
   same uninformative message this one did.

Not fixed here; this belongs with the rest of 008 in the fixloop thread.
