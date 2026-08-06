# 207 — the processor's two failure senders classify typed-only, so an untyped transient failure is terminal on the orchestrated path

**Filed 2026-08-06** by the `bugfix_197_transient_classifier` lane, **at the council's
direction** (corr `7fbf4356`, `bug_historian`, medium severity): the convergence of these
senders onto the shared classifier must be *"tracked and not just asserted in prose … given
how often 'someone else owns the other call site' has been the exact gap that recurred."*

**Status: OPEN, UNOWNED.** **Severity: medium — a retry-QUALITY defect, not a correctness
one.** Nothing is corrupted and nothing is silent: failures are correctly shaped and
correctly routed. Work that a retry would have saved is simply not retried.

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
