# NOTES — `bugs_open/197`, the retryable-side prose classifier

Append-only, newest at the bottom.

---

## 2026-08-06 — how the lane started, and the census that decided everything

Picked up from the 195 handoff ("plan it"). `196` turned out to be OWNED and fixed by
another lane (claimed 08-05, fix `d16e6d23c` committed) — competing was off the table, so
`197` (still "OPEN, UNOWNED", untouched since filed) was the continuation.

**The census the bug file demanded before any fix** was suddenly cheap, because `195`'s
`recordFailedProcessing` created exactly the needed population: `agent_error_log` rows with
`error_code='PROCESSING_FAILED'` or `classification='transient'` are precisely the errors
that reached `isRecoverableError`. Results (2,996 rows): **885 (~30%) contain
"deadline exceeded" and were made TERMINAL** — the headline; **882 missed on capitalisation
alone**; rate/5xx 37 rows, 20 missed; and the "recoverable" side includes SSL/TLS
**certificate** errors matching `connection` inside "establish a secure connection".

Two admission decisions the census made for us:
- `unreachable` / `network` (from the dead twin's list): **0 rows each** → **neither
  admitted**. A needle with no observed input is policy nobody asked for.
- bare `502/503/504`: rejected — the dominant message shape nests hex UUIDs, so `503`
  matches inside an arbitrary correlation id. Pinned OUT by a test.

## 2026-08-06 — the pre-emption caveat (the plan agent caught this; verified in code)

"Decides fleet-wide" needed sharpening: post-196 the **processor's** `sendErrorResponse` is
typed-only, sends FIRST on the orchestrated path, and wins `ClaimAwaitedRequest` — the
agentbase verdict decides only where that sender does not fire. So the live flip is smaller
than the raw 885, and so is the storm risk. Convergence of the processor's senders is the
**196 lane's** guarantee change — scheduled in RSH-006 and told in their notes, NOT done
here as a passenger edit on a file with their pinned tests.

## 2026-08-06 — implementation, and the five mutations

`MatchedTransientFailure` in `platform/messaging/retryable_transient.go` beside its
permanent twin (same package → the AST lockstep guard extends for free; `platform/errors`
is the wrong altitude — the dead `IsRecoverable` rotting there unexercised is the exhibit).
Both old classifiers deleted. `wireErrorCode` ends the DB-row/wire code disagreement.
`recordFailedProcessing` writes `retry_disposition` + `transient_matched`, so the
after-census reads a column the classifier does not consume.

**Mutations, run under the fail-closed backup recipe** (`mktemp` + `test -s` before
mutating — the corrected recipe after the 195-lane misstep; restore diffed clean each time):

| mutation | failures |
|---|---|
| remove the `ToLower` fold | 5 |
| delete `deadline exceeded` | 7 |
| re-add bare `503` | 6 |
| kill the `Retryable=false` fall-through | 5 |
| revert agent.go to a local substring helper | AST guard, both prongs |

## 2026-08-06 — state

Council submitted: `Council-Submitted: 7fbf4356-da03-4a48-832f-fd06fec5a3d7`. Registered
**RSH-006** (index re-grepped 1,764 → 1,771; ~6 rows arrived from other lanes since the
last recount). Go-only → **inert until the next chassis roll**; `197` stays OPEN until the
post-roll induction passes (recipe in RSH-006's verify-later and the bug file's appendix).

## 2026-08-06 — the first submission died at INTAKE, and the reason is worth keeping

The council run completed at `complete_invalid` with no verdict: **no reviewer ever ran.**
`__step_error`: *"edit 6: sketch is comment-only — a fix plan proposes changes, not
observations; drop the edit or make it real."* The server-side plan validator
(`diagnose_persist_fix_plan`) refuses comment-only sketches as edits.

Two lessons, the second more general than the first:
- **A comments-only change cannot be an EDIT in a council plan** — carry it in the
  rationale, where reviewers judging the touched file's blast radius still see it. Done;
  resubmitted with 7 edits on the same correlation (`RESUBMIT_CORR`), run
  `adca43f0-c749-49f9-a645-9c15aaff9bed`.
- **`RUN=COMPLETED` + `VERDICT=pending` is not "still deliberating"** — my monitor treated
  only verdict values as terminal, so it read an intake death as a pending review. The
  terminal condition for a council watch must include the invalid arm
  (`@complete_invalid`), or an intake rejection looks like queue latency for ever — the
  same shape as `192`'s "a missing row is not latency" lesson, one layer up. Monitor
  corrected.
