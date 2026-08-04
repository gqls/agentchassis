# 196 — a failure response reaches the parent stamped `complete`, and the parent records the error blob as the step's own output

**Filed 2026-08-04** by the `bugfix_195_permanent_failure_classifier` lane, at the direction
of the council's `bug_historian` seat (correlation `9b1254f0-2686-4a52-b736-1e212634ace6`,
verdict REVISE): *"the plan discovers and correctly defers fixing the success-shaped failure
envelope … registered only as a landmine text note rather than a work item with severity …
nothing in the plan creates a tracked follow-up that will survive if the landmine doc note is
missed by a future reader."* That is right, so here is the work item.

**Status: OPEN, UNOWNED.** **Severity: medium-high, pending the one measurement below.**
**Class:** silent success-shaped failure — the "no error, no warning" family.

> **VERIFICATION STATEMENT (owner ruling 2026-07-31).** I have **not** run `090`, and I am
> **not** asserting a completed root cause. What follows is a **code-read finding with an
> explicit unmeasured half**, stated as such:
> - **Read, not inferred:** `platform/messaging/processor.go:596-600` (`handleError`'s
>   non-permanent branch calling `sendErrorResponse`), `processor.go:1917-1929`
>   (`sendErrorResponse` → `CreateResponseContext()` → `sendWorkflowResponse`),
>   `platform/messaging/context.go:77-80` (`CreateResponseContext("complete", 100)`),
>   `platform/orchestration/coordinator.go:316-331` (the status switch).
> - **NOT established:** what a real awaiting parent actually records end to end. I have not
>   induced it with a live `call_agent`/`spawn_agent` parent. **The whole severity rests on
>   that, and it is the first thing the next session should do** — see "How to settle it".
> - I did not file this while fixing `195` precisely because of that gap; I am filing it now
>   because a reviewer correctly pointed out that an unfiled finding is one missed doc-read
>   away from being lost, and "unmeasured" is a property the file can carry honestly.

## The mechanism, read from source

When a message fails and is **not** classified as a permanent/validation failure,
`handleError` sends an error response to the parent:

```go
// processor.go:596-600
if domainErr, ok := errors.AsDomainError(err); ok {
    p.sendErrorResponse(ctx, msgCtx, domainErr)
}
```

`sendErrorResponse` builds its envelope from `msgCtx.CreateResponseContext()`, which is:

```go
// platform/messaging/context.go:77-80
func (mc *MessageContext) CreateResponseContext() *types.ExecutionContext {
    if mc.ExecutionContext != nil {
        return mc.ExecutionContext.CreateResponseContext("complete", 100)
    }
```

So the **status header says `complete`** and the body carries `Success: true`, with the
failure buried inside the body map as `{"error": …, "status": "failed"}`.

The coordinator dispatches on the **header**, not the body:

```go
// coordinator.go:316-331
switch execCtx.Status {
case "awaiting", "processing":      return s.handleProgressUpdate(...)
case "complete", "success":         return s.handleCompleteResponse(...)   // <- goes here
case "error_recoverable":           return s.handleRecoverableError(...)
case "error_unrecoverable", "failed", "error":
                                    return s.handleUnrecoverableError(...)
```

**So a failure is routed to `handleCompleteResponse`.** On the face of it the parent marks
the awaited step **complete** and stores the error blob as that step's data — then continues
the workflow with an error message where its results should be.

## Why this matters more than the bug it was found under

`bugs_open/195` was about a failure leaving **no** record. This is worse in kind: the failure
leaves a **positive** record. Every status says complete, the orchestration proceeds, and the
only trace is an error object sitting in a step's `collected_data` where downstream steps
expect content. Nothing errors, so nothing is investigated.

Same family as `bugs_open/132` (an error object rendered to a visitor as though it were
content) and as the "detected-then-discarded" class (`071`, `083`, `091`).

## What this DOES settle, for another lane

`bugs_open/029` (hung spawns) can stop looking at this path for its hang mechanism: on this
route the parent **is** answered, promptly. If 029's population is "parent waited for ever",
this is not it. If any of it is "parent proceeded with junk step data", this is a candidate.
Both statements are already appended to `029`.

## How to settle it — the measurement I did not run

Dispatch a parent whose only step is a `call_agent` at an agent guaranteed to fail
non-permanently, then read the parent, not the child:

```sql
SELECT status, current_step,
       collected_data->'<the_call_agent_step>' AS step_data
FROM orchestration_states WHERE orchestration_id = '<parent>';
```

- **Prediction from the code read:** the parent reaches `complete` (or advances to the next
  step) with `{"error": …, "status": "failed"}` recorded as that step's data.
- **Falsified if:** the parent sits in `AWAITING_RESPONSES`, or lands on
  `handleUnrecoverableError`. Either outcome means the code read above is wrong somewhere and
  this file should be corrected in place, not quietly closed.

A second, cheaper probe worth running first — is it already happening in production?

```sql
SELECT count(*) FROM orchestration_states
WHERE collected_data::text LIKE '%"status": "failed"%' AND status = 'COMPLETED';
```
Pair any non-zero with a positive control (a total row count for the same window), and note
the ~24h retention on terminal rows before reading a zero as "never happens".

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Send failures with a failure status.** `sendErrorResponse` should build its envelope
   with `error_unrecoverable` (or `error_recoverable` where a retry is intended) rather than
   inheriting the generic `CreateResponseContext("complete", 100)`. Closes the door: the
   parent's own switch already handles those correctly, so no coordinator change is needed.
   **Blast radius must be measured first** — `CreateResponseContext` is shared, and any
   caller relying on the current "always complete" behaviour would change.
2. **Make the coordinator distrust the header when the body says otherwise** — treat
   `Success: false` or a `status: failed` body as an error regardless of the header. Defence
   in depth, but it entrenches two sources of truth.
3. **Assert it in the response contract** so a mismatch between header status and body
   success is a hard error at the boundary. Strongest, largest change.

Prefer **1**, gated on the census its own risk names.

## Related

- `bugs_open/195` — where this was found; its fix does **not** touch this. RSH-005 carries
  it as the primary landmine.
- `bugs_closed/034` — this is its fix-candidate-3 residue.
- `bugs_open/029` — see above.
