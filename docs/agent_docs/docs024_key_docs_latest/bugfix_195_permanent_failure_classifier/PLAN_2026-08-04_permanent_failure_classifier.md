# PLAN 2026-08-04 — `bugs_open/195`: the classifier that guessed from prose

**Lane claimed 2026-08-04.** Filed by the `bugs_open/173` lane, explicitly **OPEN, UNOWNED**,
with an exemplary verification statement (induced the fault twice, read the whole path,
every negative claim paired with a positive control). Its diagnosis is confirmed; two of
its subsidiary claims are corrected — see NOTES and the appendix in the bug file itself.

## The mechanism

`ValidateWorkflow` rejection → `errors.New(errors.ErrWorkflowInvalid, "Invalid workflow
configuration").WithCause(err)` → renders as
`WORKFLOW_INVALID: Invalid workflow configuration (caused by: … requires a topic)`.

The permanent-vs-transient decision is a **case-sensitive substring match over that prose**
(`ValidationErrorNeedles = {"is required","validation","invalid","missing"}`). None match:
*"is required"* loses to the wording (*"requires a topic"*), *"invalid"* loses to the
**capital I**. So the fleet's commonest permanent configuration error is classified
transient — and `recordDroppedValidationError`, the durable record `bugs_closed/034` exists
to provide, is never called, **because the branch that calls it is never entered**.

034 is not regressed and this is not a duplicate of it. 034 made the DROP durable,
*conditional on the classifier recognising the error*. **A guarantee conditional on a
classifier inherits every gap in that classifier.** 034's own file header names the typed
fix and defers it — this is that deferral coming due.

## Fix, ordered by what closes the door

| # | change | why |
|---|---|---|
| 1 | `MatchedPermanentFailure` — typed `DomainError.Code` via `errors.As`, closed list; needles demoted to a fallback for untyped errors | exact where prose is not; undefeatable by rewording, capitalisation or `%w` |
| 2 | `recordFailedProcessing` — a row on **every** non-dropped failure | makes the trace independent of the classification being right at all |
| 3 | both call sites in **one** commit; bare `err.(*DomainError)` → `errors.AsDomainError` | splitting them is the drift 034 closed |

Rejected: adding needles / case-folding (more of the same fragile mechanism, and folding
would widen every documented hazard — `invalid connection`, `invalid memory address`,
x509 — to its capitalised variants). Rejected: classifying on `DomainError.Retryable`
alone (far too wide — every `InternalError` sets it false).

**Deliberately NOT done:** suppressing the fallback for non-listed codes. More principled,
but it would move existing drops back to retry fleet-wide. **This change only ever ADDS
permanent classifications; it removes none.**

## The hole my own test found

An error built `AsRetryable` skipped the typed branch and was then classified **permanent**
by the fallback matching the word *"validation"* in its own message — structure overridden
by prose, which is this bug in miniature. Hence `if de.Retryable { return "" }`, which
bypasses the fallback too.

## Verification that can come out FALSE

- **Mutation** (done): reverting to the prose-only body fails `ReproducesTheBug`,
  `SurvivesPercentWWrapping` and `ExplicitlyRetryableIsNeverPermanent`.
- **Post-roll**: induce the fault (throwaway agent whose only step is `{"action":"complete"}`
  — invalid, `complete` is a step NAME whose action is `complete_workflow`) and assert
  exactly one `agent_error_log` row, `VALIDATION_ERROR_DROPPED`, `matched_needle =
  'code:WORKFLOW_INVALID'`. Baseline to beat: **0 rows**, needle-HIT **0**.
- **Do NOT** count `Processing failed` — it is 2 per pass by construction, so the filer's
  proposed check would fail a working fix. Count `"process() in processor.go starting"`.

## Blast radius

Measured: `ErrWorkflowInvalid` has exactly **one** non-test construction site
(`processor.go:280`); `ErrValidation` one (`contentcreator/agent.go:226`, sent as a
response, never returned through the classifier). Told, not merely measured: the **173
lane** (filer, owed both corrections), the **029 lane** (hung-parent reading falsified for
this path), **diagnosis-loop owners** (`agent_error_log` gains a row class that appears in
every diagnosis bundle), and the **contentcreator** owner.

Registered **RSH-005** in the shipping commit. Council `9b1254f0`.
