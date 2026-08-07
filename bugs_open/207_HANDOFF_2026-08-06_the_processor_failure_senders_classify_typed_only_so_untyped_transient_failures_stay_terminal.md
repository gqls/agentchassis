# 207 — the processor's two failure senders classify typed-only, so an untyped transient failure is terminal on the orchestrated path

**Filed 2026-08-06** by the `bugfix_197_transient_classifier` lane, **at the council's
direction** (corr `7fbf4356`, `bug_historian`, medium severity): the convergence of these
senders onto the shared classifier must be *"tracked and not just asserted in prose … given
how often 'someone else owns the other call site' has been the exact gap that recurred."*

**Status: FIXED, LIVE on v1.0.1262, PROVEN BY INDUCTION 2026-08-07 — all three close
criteria met at this bug's own seam.** Stays in `bugs_open/` per the owner's 08-06 ruling
(finished bugs stay). **BUT the ~30% prize is NOT yet realised** — the induction that
proved this fix also found the next two defects downstream, and the retried work this fix
promises is gated on them: `bugs_open/216` (the coordinator's response-driven recoverable
arm re-arms then refuses its own replay — terminal anyway, cross-pod) and `bugs_open/217`
(`notifyParentOfFailure` hardcodes `error_unrecoverable` and answers the parent FIRST on
child-orchestration failures). Owned by the `bugfix_207_sender_convergence` lane (docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_207_sender_convergence/`).

## CLOSE CRITERIA RESULTS (2026-08-07, v1.0.1262, corr `b155c554-0753-4f57-97a0-fcaec5d229d8`)

1. **PASSED — pod-grep, both replicas:** added marker (the long sender log literal) = 1,
   negative control (removed comment text) = 0, positive control = 1, same exec.
2. **PASSED at the letter, and the spirit exposed 216:** the induced deadline-exceeded
   child produced `disposition=error_recoverable, matched='deadline exceeded'` in the pod
   log and `status=error_recoverable / body.error.recoverable=true` on the wire (the
   pre-fix binary pinned this exact shape `error_unrecoverable`); delivered to the
   parent, the coordinator accepted it and re-armed `retry_version` 0→1. **Then the
   replay refused** (`bugs_open/216`) — so `retry_version >= 1` was satisfied by a
   bookkeeping write on the way to terminal. The criterion as written measured the
   re-arm, not the retry; recorded in `WRONG_CALLS.md`.
3. **PASSED — storm-watch:** retry_version histogram 25025/2126/1413/126, hard wall at 3,
   zero above, all history.

Also live in the same induction: a second answer to the same awaited request from
`notifyParentOfFailure`, hardcoded `error_unrecoverable`, one second EARLIER than the
converged sender's — first response claims (`bugs_open/217`).
**Severity: medium — a retry-QUALITY defect, not a correctness
one.** Nothing is corrupted and nothing is silent: failures are correctly shaped and
correctly routed. Work that a retry would have saved is simply not retried.

## DECISION TAKEN 2026-08-06 — option 1, CONVERGE. Council APPROVED round 1

`Council-Reviewed: 155f4730-4526-4523-83d0-3ce4c4fc9f1c` (approved, 2 advisory objections,
none high-severity). Registered **RSH-007**. The convergence is NOT the bug file's one-line
sketch: `sendErrorResponse` also serves `handleError`'s validation branch, and an
`"invalid connection"`-shaped failure carries both a permanent and a transient needle — the
one-liner would have the processor drop it as validation while the wire tells the
coordinator to retry it. Both senders instead decide through **`messaging.RetryDisposition`**
(permanent first, then transient, else terminal — the sequence the classifier's own header
documents and agentbase enforces by call order), for the status header AND
`ErrorInfo.Recoverable` (which `datahelpers.determineStatus` re-derives a status from), and
log the disposition + audit token. The 196 lane's pinned tests were flipped deliberately,
with new pins: permanent-outranks-transient, and unclassifiable-stays-terminal.
Re-measured at fix time: 1,001 of 3,399 census rows (~29%) deadline-exceeded; classifier
proven live at the agentbase seam (122 `connection` + 56 `deadline exceeded` recoverable).

**CLOSE CRITERIA (whoever verifies after the next roll):**
1. Pod-grep EVERY replica for the long sender literal
   `retry disposition decided at the processor sender by the sequenced shared classifier`
   (expect ≥1) plus a negative control — use `scripts/pick-pod-marker.py <commit>`.
2. Behavioural induction: a deadline-exceeded-shaped failure on the ORCHESTRATED path shows
   `awaited_requests.retry_version >= 1` (pre-fix baseline: terminal on first attempt).
3. **Storm-watch (the guardian's and architecture seat's advisory, made an obligation):**
   `SELECT retry_version, count(*) FROM awaited_requests GROUP BY 1 ORDER BY 1;` — mass at
   0-1, hard wall at 3, ZERO rows above. The recoverable arm has NO backoff (RSH-006 gap,
   unchanged), so anything at 4+ falsifies the bound claim and is a stop-the-line finding.

> **VERIFICATION STATEMENT (owner ruling 2026-07-31).** No `090` run; this is **not** a
> new diagnosis. It is the *named, measured residual* of two closed bugs, filed as its own
> file because the council required the decision have a home that cannot expire. Every
> claim below is either read from source or measured live, and cited as such.

## The mechanism

`platform/messaging/processor.go` has two failure senders — `sendWorkflowFailureResponse`
(workflow-start failure) and `sendErrorResponse` — both fixed by `bugs_closed/196` to carry
a **failure status** instead of riding the success envelope. That fix is correct and is not
in question. What they decide **recoverable-vs-not** by is:

```go
status := "error_unrecoverable"
if errors.IsRetryable(err) {          // typed flag ONLY
    status = "error_recoverable"
}
```

`errors.IsRetryable` reads `DomainError.Retryable`. **Nothing in production sets that flag**
— `ErrorBuilder.AsRetryable` is called only from tests (measured 2026-08-05, unchanged
2026-08-06). So in practice **every** failure these senders emit is
`error_unrecoverable`, including genuinely transient ones.

Meanwhile `bugs_closed/197` gave agentbase a shared classifier,
`messaging.MatchedTransientFailure`, which is typed-first **with a census-derived prose
fallback** — and it is live and working: measured post-roll, 19 rows classified
`error_recoverable` via the `deadline exceeded` needle and 48 via `connection`, where all
of them were terminal before.

**Why the agentbase fix does not cover this path:** on the orchestrated path the
processor's response is sent **first**, and the coordinator's `ClaimAwaitedRequest` makes
the first response the deciding one (`DUPLICATE_SKIPPED` for the loser). So for those
failures the typed-only verdict wins and agentbase's better verdict is discarded.

## The measured population

From `bugs_closed/197`'s census (2,996 `agent_error_log` rows that reached the sibling
seam): **885 rows (~30%) contain `deadline exceeded`** — transient by definition, untyped,
and therefore terminal at these senders. That is the size of the prize, and it is why this
is filed rather than shrugged at.

## The decision — either answer closes this file; silence does not

1. **Converge.** One line per sender:
   `if messaging.MatchedTransientFailure(err) != "" { status = "error_recoverable" }`
   (or replace the typed check entirely, since the shared classifier already consults the
   typed flag first and reports `retryable:<code>`). **Cost:** the 196 lane's pinned tests
   (`processor_response_status_test.go`) assert untyped `"context deadline exceeded"` →
   `error_unrecoverable` here; converging flips those, which is a deliberate guarantee
   change and the reason this is a decision rather than a chore.
2. **Decline, on the record, with the reason.** There is a real argument for typed-only
   purity at the wire: prose classification at a protocol boundary is exactly what
   `bugs_closed/195` and `197` were about, and someone may prefer to fix this by making
   producers set `AsRetryable` properly instead. If so, say so here and **file that** as
   the follow-on — the 885 rows still need an answer.

**Bounds if you converge** (so the risk is not re-derived): the coordinator caps at
`retry_version >= 3` (`handleRecoverableError`), adapter re-executions are capped
independently (`bugs_open/075`), and **no backoff exists** on the recoverable arm — a known
gap named in RSH-006. Live storm-watch after 197 rolled: retry_version histogram
204 / 15 / 83 / 2 with **zero rows above 3 in all history**, so the wall holds.

## Related

- `bugs_closed/196` — fixed the envelope; carries this as a named residual after its close.
- `bugs_closed/197` — fixed the sibling seam and built the classifier this would adopt;
  RSH-006 registers it, with the pre-emption as landmine 2.
- `bugs_closed/195`, `bugs_closed/034` — the same family, permanent side.
