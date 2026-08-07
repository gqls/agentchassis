# PLAN — bugs_open/207: converge the processor's two failure senders onto the shared classifier

**Lane started 2026-08-06.** Owner thread: this one (session `e68e25cd`). Bug:
`bugs_open/207_HANDOFF_2026-08-06_the_processor_failure_senders_classify_typed_only_so_untyped_transient_failures_stay_terminal.md`

## The decision taken: CONVERGE (option 1 of the bug file's two)

The bug file offers converge-or-decline. We converge, because:

- The typed-only purity argument (option 2, "make producers set `AsRetryable`") requires
  auditing every error construction site fleet-wide — exactly the blast radius 034's lane
  deferred, twice, for good reason. Meanwhile the 1,001 deadline-exceeded rows (re-measured
  2026-08-06, up from 885 at filing) keep losing work a retry would save.
- The shared classifier is not hypothetical: `MatchedTransientFailure` is live and proven at
  the agentbase seam (measured today: 122 `connection` + 56 `deadline exceeded` dispositions
  recorded `error_recoverable`, 18 unmatched terminal).
- The framework already states the contract this fix enforces: `retryable_transient.go`'s
  header — "Every processing failure is classified twice, in sequence: first
  MatchedPermanentFailure … anything not permanent then reaches this file's question."
  The senders are the only failure paths not honouring it.

## Design — a sequenced disposition helper, not a one-line patch

The bug file's sketch (`if messaging.MatchedTransientFailure(err) != "" { status = "error_recoverable" }`)
is **not safe as written**, and this is the load-bearing finding of this lane's read:

**`sendErrorResponse` is called from BOTH branches of `handleError`** — including the
permanent/validation branch (`processor.go:603-605`), which drops the message and returns
nil precisely so it is never retried. A naive transient-only convergence would ask the
transient question about an error already judged permanent. The documented over-matches make
this concrete: `"invalid connection"` (a real DB-driver fault, quoted in
`validation_drop.go`'s header) contains both the permanent needle `invalid` and the
transient needle `connection`. Under the naive patch, handleError drops it as validation
while the wire tells the coordinator `error_recoverable` — the two layers issue
contradictory dispositions for one failure, and the coordinator re-dispatches a message the
processor just recorded as dropped-never-retry.

So the change encodes the sequence where the wire status is decided:

```go
// platform/messaging/retryable_transient.go
// RetryDisposition applies the two-stage classification in its documented
// order: permanent first (drop, never retry), then transient (a retry can
// cure), else terminal. Returns the wire disposition and the audit token.
func RetryDisposition(err error) (recoverable bool, matched string) {
    if match := MatchedPermanentFailure(err); match != "" {
        return false, "permanent:" + match
    }
    if match := MatchedTransientFailure(err); match != "" {
        return true, match
    }
    return false, ""
}
```

Both senders (`sendWorkflowFailureResponse` `processor.go:547`, `sendErrorResponse`
`processor.go:1960`) adopt it:

- `status` derives from `recoverable`.
- **`ErrorInfo.Recoverable` moves in lockstep** — it is read downstream:
  `datahelpers.determineStatus` (`data_helpers.go:813`) re-derives the status from it, so
  leaving it on `errors.IsRetryable` would make the header and the body disagree about the
  same failure.
- Each sender logs the disposition + audit token (mirroring agentbase's
  `handleProcessingError`), so the orchestrated path's decisions are as auditable as the
  agentbase path's.

**agentbase is untouched.** Its call-order guard (processMessage classifies permanent and
drops before `handleProcessingError` ever runs) already enforces the same sequence; its code
is the 197 lane's, proven live. The helper exists for any FUTURE caller that lacks such a
guard — that is what makes this a framework fix rather than a site patch.

For the handleError callers the permanent-first check is behaviour-preserving (the branch
already decided); for the other two callers — workflow-start failure (`processor.go:297`)
and config-load failure (`processor.go:1734`) — it introduces the documented sequencing they
should always have had. A config load that failed on a DB `deadline exceeded` becomes
retryable (correct); a genuinely missing config stays terminal.

## Guarantee change, named (owner ruling 2026-07-29)

- **What changes:** the orchestrated path's failure status stops being typed-only. An
  untyped transient failure (`deadline exceeded`, `connection refused`, …) now routes to the
  coordinator's `recoverable` arm instead of terminal.
- **Consumers, told not merely measured:** (1) the **SagaCoordinator** —
  `handleRecoverableError` now receives the wider population; its `retry_version >= 3` cap
  and the independent adapter cap (`bugs_open/075`) are the bounds, and RSH-006 landmine 3
  already names them load-bearing. No backoff exists on that arm (known gap, RSH-006, not
  fixed here). Live storm-watch after 197: retry_version histogram mass at 0-1, hard wall at
  3, zero rows above. (2) **`datahelpers.determineStatus`** — reads
  `ErrorInfo.Recoverable`; updated in lockstep by design. (3) the **196 lane's pinned
  tests** — flipped deliberately, below; the 196 lane is closed, no thread competes
  (checked: who-owns + live transcript grep, 2026-08-06).

## Test changes (the deliberate flips, plus new pins)

In `processor_response_status_test.go`:
- `TestErrorResponseStatusIsTypedFromTheError` → renamed to state the new contract
  (sequenced). The two untyped deadline-exceeded cases flip to `error_recoverable` /
  `Recoverable: true` — these are the 1,001 rows, and the flip is the fix.
- The `WORKFLOW_INVALID` case stays `error_unrecoverable` (typed permanent).
- **New pin — the sequencing itself:** an error whose prose contains BOTH a permanent and a
  transient needle (`"Invalid connection string"`) must stay `error_unrecoverable`. This is
  the case that distinguishes the sequenced helper from the naive patch; it must fail if
  someone reorders the checks.
- **New pin — unclassifiable stays terminal:** untyped prose with no needle.
- `TestWorkflowFailureResponseStatusIsTyped` → keeps its WORKFLOW_INVALID terminal
  assertion; gains a transient case (untyped `dial tcp … connection refused` →
  `error_recoverable`).
- Legacy body-blob dialect test untouched — the blob (`status: "failed"`) is deliberately
  not changed; several Go readers parse it.

## Verification

1. `go test ./platform/messaging/ ./platform/agentbase/` green, against `git archive HEAD`
   (shared-tree rule).
2. Mutation check: revert the sender line to typed-only → the flipped tests fail; reorder
   permanent/transient in the helper → the sequencing pin fails.
3. Council gate: submit before/alongside commit, `Council-Submitted:` trailer.
4. Post-roll (verify-later): pod-grep a LONG literal from the new sender log line on every
   replica with a negative control (`scripts/pick-pod-marker.py`), then the behavioural
   induction: a deadline-exceeded-shaped failure on the orchestrated path must show
   `awaited_requests.retry_version >= 1` where the pre-fix baseline is terminal-on-first.
   Storm watch stays the RSH-006 one: `retry_version` histogram, wall at 3.

## Register / docs updates in the same commit

- RSH-006: status corrected (stale "inert until roll" — 197 closed LIVE), landmine 2's
  "until the 196 lane converges its senders" resolved to this lane, `RetryDisposition`
  added to sources.
- `bugs_open/207`: decision recorded (option 1), moves to bugs_closed only when LIVE.
- 016b §10 index row when closed.
