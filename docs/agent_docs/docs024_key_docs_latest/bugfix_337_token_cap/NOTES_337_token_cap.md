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
