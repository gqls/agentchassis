# NOTES — bugfix 337 (generate_template blows its 16,000-token cap)

Append-only, newest at the bottom. Technical log: what was tried, what the system said,
and every misstep.

## 2026-08-22 — session start: ownership, validity, first measurements

**Ownership check.** `scripts/who-owns.py 337` names `bugfix_311_component_keys` as the
likely owner (9 mentions), but their `HANDOFF_2026-08-20_continue_here.md:155` lists 337 as
"filed from this lane, 08-20" and the bug file itself says **"Status: OPEN, unowned"** —
the lane's live work is 311/345/351, not 337. Queue check: no open `site_work_items` row
targets 337 or a fix for it (query in RUNBOOK). Taking the bug; contributing findings into
`bugs_open/337_…md`, not a parallel account.

**Bug still valid** [MEASURED 2026-08-22 ~09:55Z]:
- Live `agent_definitions` row: `component-creator.generate_template` still
  `{"model":"claude-sonnet-4-6","provider":"anthropic","max_tokens":16000}`.
- The failed items still parked `failed`, `attempt_count=3`:
  loanzy.uk `needs_new_component:loans-credit-health-check` (08-19),
  loanzy.uk `…_run1` (08-18 — a FOURTH item the bug file predates),
  loancalculator.co.uk `needs_new_component:loans-credit-health-check` (08-15).
- `llm_call_log`: **9 truncation failures** for `generate_template`, 08-15 → 08-19, every
  one `output_tokens=16000`, recovered chars 46,441–48,817. Matches the bug file's
  three-items × three-attempts account.

**New measurement — the cap is tight even on SUCCESSES** [MEASURED 2026-08-22, 14-day
window, `llm_call_log`, successful calls with output_tokens & max_tokens present]:
- `component-creator/generate_template`: 154 calls, p50 8,894, **p95 13,633 of a 16,000
  cap (85%)**, max 15,374 (96%), 6 calls ≥90% of cap. The step runs hot for everything,
  not just `loans-credit-health-check`.
- Same census fleet-wide: `council-gate/review_editquality` max 15,777/16,000 (98.6%);
  `diagnose-agent/verdict` max 31,033/32,000 (97%); `page-content-writer` rewrite_negations
  2 at-cap at 2,000. The near-cap class is fleet-wide, which is what makes "threshold
  management" (the owner's steer) a framework question, not a one-step tweak.
- All-history truncation-failure census per step (error `%reached the configured cap%`):
  top rows `vet-practice-verifier/extract_and_reconcile` 94 (08-05→07, cap 2048),
  `generic+council-gate/review_editquality` 49, `component-creator/generate_template` 9,
  `tool-auditor/llm_audit` 9, `diagnose-agent/verdict` 4, `tool-improver/improve_tool` 3.

**Code read (transport → step → item):**
- `platform/aiservice/truncation.go` — typed `TruncatedError` CARRIES the partial;
  `IsTruncated` is the opt-in salvage hook. Its header states the design position:
  raising caps is NOT a class fix ("whatever the number, the step that writes most
  approaches it on the work most worth doing").
- `platform/aiservice/max_tokens.go` — budget resolution precedence: per-call
  `options["max_tokens"]` wins → `ai_service.max_tokens` at client construction →
  `DefaultMaxOutputTokens` 2048. So a per-call escalated value needs NO client change.
- `platform/orchestration/actions/ai_actions.go:415-553` — `execute_llm_prompt`'s
  truncation path: `tolerate_truncation` opt-in (with the 076 guarded-consumer check),
  5xx transient retries, `AIUnavailableError` back-to-triage. `:743-747`: the 119 JSON
  re-ask DELIBERATELY does not raise max_tokens, citing truncation.go. Distinction that
  matters for this fix: that re-ask is for a *judgement* that can be asked shorter; a
  47k-char *document* cannot be — its length is the work product.
- `platform/errors/permanent_failure.go` — `MatchedPermanentFailure` is deliberately
  closed/conservative; the truncation message matches no needle, so item-level retries
  burn all 3 attempts (candidate 4 territory — noted, not taken).
- `anthropic.go` sends whatever `options["max_tokens"]` says — no clamp.
- `./scripts/audit-optional-key-budget.sh`: `execute_llm_prompt` has **no
  ActionInputSpec** (67 carriers, listed under "NOT COUNTED — surface UNKNOWABLE, not
  zero"). A new opt-in key does not move the RFC_022 counter; the missing spec is a
  pre-existing gap to name, not to fix here.

**Missteps so far:** none recorded yet.

## 2026-08-22 — design decision: escalate-on-truncation + resize + headroom check

Decision and reasons in `PLAN_2026-08-22_token_cap_management.md`. In one line: the
routine cap stays a cost control, a per-step opt-in `max_tokens_ceiling` gives
`execute_llm_prompt` ONE in-call retry at the ceiling when the provider says
`stop_reason=max_tokens`, the step's cap is resized from measurement (16,000 → 24,000),
and a daily headroom check over `llm_call_log` makes every step's cap a monitored
threshold instead of folklore.

> **CORRECTED 2026-08-22 (same session, before any code):** the third leg — "a daily
> headroom check" — was about to duplicate a mechanism that already exists and already
> fired: LCO-007 `fleet-step-token-pressure` (6-hourly) has flagged this exact step in
> `doc_notes` since 08-18. Caught by the prior-art sweep. No new monitor is built; the
> gap is flag→action dispatch, named as a residual in the bug file. See next section.

## 2026-08-22 — implementation + council submission

- Prior-art sweep (subagent, full report absorbed into PLAN) **corrected the draft plan
  before any code**: the "threshold management" monitor I was about to build already
  exists — LCO-007 `fleet-step-token-pressure` (6-hourly, C/T/N/P) — and it has flagged
  `T generate_template@16000 … truncated 9` in doc_notes since 08-18 with no consumer.
  Not a WRONG_CALL (caught pre-assertion, pre-code, by the research pass working as
  intended); recorded here so the near-miss is visible.
- Throughput measured to size the ceiling: the nine cut calls took 165.6–170.2s for
  16,000 tokens (~94–97 tok/s); worst successful throughput 91.7 tok/s. 32,000 is
  clock-safe (~349s worst) against the 600s non-streaming timeout; 40,000 at a
  conservative 60 tok/s is not (667s).
- At-cap SUCCESS rows (output_tokens=16000, success=t) checked before they could
  poison sizing: all April 2026 — pre-TruncatedError history, ignored.
- Shipped: `truncation_escalation.go` (+17-case test, green), wiring in
  `ai_actions.go` (between GenerateText and the error block — the existing block only
  ever sees the FINAL attempt's error), migration 549 (+ROLLBACK), register MDL-042,
  bug-file append. `go vet` clean for my files (one pre-existing warning in
  `load_component_library_actions.go:207`, not mine).
- Council: DRY_RUN passed, submitted for real.
  **SUBMISSION_CORR = 3d531c9a-4351-42bc-806c-17ed25636a8c** — budget ~30 min; find
  the run by payload, not by the printed id.

## 2026-08-22 — before-state pinned; council run located

- Council run found by payload: orchestration `4f031702-0e36-4a4e-a1b1-9eb5d8b00f06`,
  dispatched 09:21:36Z (fast — no queue delay this time), `gate_bug_historian` at
  09:25Z. Watching for the verdict; migration apply waits for it per the PLAN phasing.
- Before-state pinned, TWO reads each, stable:
  - `https://loanzy.uk/tools/credit-health-check/index.html` — 24,323 bytes,
    md5 `bdc997300740612af0625a53d61416d3`, `grep -c '<input'` = **0**.
  - `https://loancalculator.co.uk/tools/credit-roadmap/index.html` — 1,201 bytes,
    md5 `8561e9f73e44ec554f8864deea13833a`, `<input` = **0** (a stub — this page has
    never been built, consistent with `incomplete_page_group` open item).
- Migration runner dry-run running in background (slow — per-file ledger checks).

## 2026-08-22 — same-file passenger disclosure on commit 9e23fb852

My WRONG_CALLS append (the LCO-007 duplicate-monitor near-miss) committed at 68
insertions where my entry is ~35 lines: **the bugfix_342 lane's uncommitted entry ("I
read a VACUOUS zero as a clean census") rode along** — the same-file-passenger case a
pathspec commit cannot exclude. Nothing lost, forward-only holds; their entry is intact
under my commit message. 342 lane: your WRONG_CALLS entry is already committed
(9e23fb852) — do not re-append it.

## 2026-08-22 — council round 1: REVISE, and the gating objection found a REAL defect

Verdict at 09:37:52Z, `decided_by: gating objection from editquality`, 4 abstained,
`gated_by_truncation: false`. Objection on edit 3, verbatim: *"The wiring reads
`options["__sent_max_tokens"]` to get sentCap for the refusal check
(`ceiling <= sentMaxTokens`), but no edit anywhere sets this key. If it isn't already an
established key elsewhere in ai_actions.go, sentCap silently defaults to 0 on every
call, which makes the stated safety refusal … "*

**Checked rather than argued, and the objection is RIGHT — narrowly, and worse than its
own framing.** The key IS established (`grep -rn '__sent_max_tokens' --include=*.go`):
- `anthropic.go:313-318` — set from `requestBody["max_tokens"]`, guarded only by
  `options != nil`, and `options` is never nil on this path (`ai_actions.go:350`
  `make(map…)`). Always set.
- `gemini.go:376` — unconditional inside the same `options != nil`. Always set.
- **`ollama.go:121` — set ONLY when `optionsBlock["num_predict"]` is present as an int.
  And ollama deliberately OMITS `num_predict` when no budget is configured
  (`aiservice/max_tokens.go`: "ollama does NOT use it — it omits the field entirely
  when unconfigured").** So an unconfigured ollama step reaches the escalation with no
  wire number at all.

So the reviewer's failure mode is reachable on one provider, and its consequence is the
inverse of safety: `sentCap` = 0 makes `ceiling <= 0` false for EVERY positive ceiling,
so the refusal I documented as the safety property would have been **vacuous** — it
would escalate against a baseline nobody established. A comment asserting a guard that
the code cannot enforce is exactly the estate's "a doc comment is not an enforcement
mechanism" shape.

**Fix (round 2):** carry the type assertion's `ok` instead of discarding it —
`sentCap, sentCapKnown := options["__sent_max_tokens"].(int)` — and make UNKNOWN a
third state that **refuses to escalate** (fail-closed). Same three-state discipline
`rewrite_negations_action.go` applies to unreported usage. Two new test rows pin it, and
**the guard is mutation-proven, not asserted**: deleting `if !sentKnown { return 0,
false }` fails exactly those two cases and nothing else (run 2026-08-22; the rest of the
table stayed green, so the mutant is not killed by collateral). No behaviour change for
the motivating anthropic case.

**Cost of the round: one REVISE, ~16 minutes, and it bought a defect that no test I had
written could see** — my table set `sent` as a plain int, so absence was unrepresentable
in the fixture. The estate's line holds again: a REVISE round is cheaper than the defect
it finds.

## 2026-08-22 — council round 2 APPROVED; advisories answered; migration 549 APPLIED

**APPROVED** (`decided_by`: "approved with 3 advisory objection(s) — none high-severity",
4 abstained). 8 seats clean; 4 carried objections, all advisory, all answered:

- **prior_art_librarian [low] — RIGHT, and it caught an unmeasured claim of mine.** The
  "067-sweep: generate_template is component-creator's only `execute_llm_prompt` step"
  line carried no `[MEASURED]` tag because I had asserted it from a step-NAME list, not
  a query. Now queried and recorded in the migration header: six steps, exactly one LLM
  action, none of the other five carries a cap. The query could have returned a second
  LLM step and did not.
- **reuse_agent [medium] — checked, declined, reason recorded IN THE CODE.** Unlike
  `aiservice/max_tokens.go` (where the same objection was answered by an import cycle),
  `package actions` can import `datahelpers` freely, so reuse genuinely was available.
  Declined anyway: `GetIntField` handles float64+int only, while the SIBLING key in the
  same config block (`max_tokens`) is read by `aiservice.configMaxTokens`, which also
  takes int64/json.Number. Two coercion rules for two keys of one block is the
  two-readers-of-one-concept drift class. ~10 lines to match the sibling. Noted that if
  257 candidate 2 unifies them, this should follow rather than keep its copy.
- **guardian [medium+low]** — blast radius on a 67-carrier shared action, and the
  in-place `options` mutation. Both already answered by construction: opt-in (no key →
  byte-identical path, pinned by test) and the provider clients read `options` per-call
  inside `generate()`, with `LogLLMCall` reading the map synchronously before its
  goroutine (`llm_call_logger.go:35-39`).
- **bug_historian [medium]** — no remediation step for the two hollow pages. Correct:
  the forward path and the repair are separate, and the repair is PLAN phases 5-7. Doing
  it next; it is the close-out bar, not an optional extra.
- **editquality [low]** — MDL-042 not among the plan's edits. It IS committed (register
  entry + `000_concept_index.md` row, commits `9e89e8ca1`/`c7b2c708e`); the plan's 4-edit
  cap meant docs were described in the rationale rather than listed. No action.

**Migration 549 APPLIED 09:56:36Z**, scoped via `MIGRATIONS_DIR=<scratch dir holding only
549>` — a bare `--apply` would have swept ~12 other threads' pending files. Snapshot
taken (`23720180-…`), `UPDATE 1`, ledger recorded. Verified at the live row by the
RESOLVED value, not the written key: **resolved_cap 24000, ceiling 32000, dead
`config.max_tokens` key NULL, version 2.**

## 2026-08-22 — VERIFICATION BY DEMAND OVERTURNED THE BUG'S DIAGNOSIS (the session's main finding)

Re-drive of both live pages, items filed 09:59Z, both `complete` at `attempt_count=0`.
Three generations at cap 24000: **14,244 / 12,709 / 14,816 tokens, zero truncations.**

**And all three are BELOW the old 16,000 cap** — which cannot happen if the brief
"reliably exceeds" it. That forced the census I should have run at the start, joining
`llm_call_log` to `site_work_items` ON `work_item_id` and filtering on the item's own
`spec->>'section_type'`:

- **82 generations of `loans-credit-health-check`, 73 SUCCESSFUL at cap 16000**
  (8,641–15,374 tokens), **9 cut.** The cap fits ~89% of the time.
- Per item, `attempt_count` 3 vs **13 / 55 / 11 actual LLM calls** — an in-workflow
  regeneration loop that never touches the attempt counter. (The 55-call item is
  `8c8f5de5`, which `bugs_open/345`/migration 533 measured from the other end as "52
  rejections in 3h34m". Two lanes measured the same item and neither saw the other.)
- The real blocker, proven the same hour AT A TALLER CAP: loanzy's fresh generation was
  refused by `store_component` pre-store validation — `field "cta_primary_url" declares
  source "site_specs.ctas.primary_url" but no site carries a site_specs aspect named
  "ctas" … (bugs_open/309)`. No component stored, despite a clean 12.7k-token generation.

**My misstep, logged in WRONG_CALLS:** I inherited "nine cap-hits, zero successes, 100%
reproducible" from the bug file, repeated it in PLAN/NOTES, put it in a council
submission's `grounded_in`, and was APPROVED on it. Thirteen seats could not catch it —
every figure I gave them was true of the population I had selected. The bug file's census
(`site_work_items WHERE error ILIKE '%reached the configured cap%'`) selects on a
LAST-WRITE-WINS column, so it could only ever return items whose final error was the thing
being tested, and the 73 successes were structurally invisible (different table, one row
per CALL not per ITEM). Pattern written up in `016b` §9.

**Disposition — re-scoped, NOT reverted.** 549 + MDL-042 stand on independent evidence
(successful-call p95 13,633/16,000 = 85%, max 96%; LCO-007 flagging since 08-18; 9 real
truncations in 82 calls is a genuine ~11% loss). They must NOT be credited with healing the
pages, and 337 must not close on them. Bug re-scoped in its own file to the validation-driven
regeneration loop, routed at the 309/345 territory.

**Repair state:** loancalculator's component STORED (22,236 chars, closes properly); page
re-render filed to attach it (page had 5 planned / 4 slots). loanzy: no component — blocked
by 309. Both pages still serve 0 `<input`.
