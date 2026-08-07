# NOTES — bugs_open/207 sender convergence (append-only, newest at the bottom)

## 2026-08-06 — lane opened, bug re-verified valid

- Ownership: `who-owns.py 207` says "OWNED or recently active" but the hits are the
  **finished** 197 lane (it filed 207 as its residual and closed the same morning —
  commit `710b8e40d`). Live-transcript grep (the lagging-check remedy): session
  `424d6591` = the 197 lane, done; `ad5665d0` = unrelated checker work. **207 unowned,
  taken.**
- Code re-read at HEAD: both senders still typed-only —
  `processor.go:563-566` (`sendWorkflowFailureResponse`) and `:1979-1982`
  (`sendErrorResponse`), each with the standing-decision comment naming this follow-up.
- Live DB (re-measured, could have come out otherwise — a converged sender would have
  shown recoverable dispositions on the orchestrated path):
  - classifier live at agentbase seam: dispositions in `agent_error_log.context`:
    `error_recoverable/connection` 122, `error_recoverable/deadline exceeded` 56,
    `error_unrecoverable/(no match)` 18.
  - population: 3,399 rows (`PROCESSING_FAILED` or classification=transient), 1,001
    (~29%) contain `deadline exceeded`. Filing figures were 2,996/885 — grown, not stale.
  - open work items touching the mechanism: **0 rows**.
- Load-bearing finding vs the bug file's sketch: the one-liner in the bug file is unsafe —
  `sendErrorResponse` is also the permanent branch's sender (`handleError:603-605`), and
  `"invalid connection"`-shaped errors carry both a permanent and a transient needle. The
  fix must sequence permanent-first at the seam. See PLAN §Design.
- `ErrorInfo.Recoverable` is not decorative: `datahelpers.determineStatus`
  (`data_helpers.go:813`) derives a status from it. It moves in lockstep or header/body
  disagree.
- Only two `IsRetryable` decision sites exist fleet-wide (grep, non-test): the two senders.
  The convergence closes the whole class.
- Register staleness caught (the standing landmine): RSH-006 `status:` still says "built
  (inert until roll)" while 197 closed LIVE on v1.0.1259 the same morning. Corrected in
  this lane's commit, visibly.

## 2026-08-06 — implemented, mutation-proven, submitted, APPROVED round 1

- Implementation: `RetryDisposition` in `retryable_transient.go`; both senders adopt it for
  status + `ErrorInfo.Recoverable` + audit log line. gofmt clean.
- Tests green against a clean `git archive HEAD` overlay (the working tree's
  `platform/orchestration/actions` is broken by ANOTHER session's WIP — not mine, not
  HEAD's; do not "fix" it from this lane). Two mutations each flipped named tests:
  typed-only revert → the two flipped deadline-exceeded cases fail; helper reorder →
  `PermanentNeedleOutranksTransientNeedle` fails at both classification and wire level.
- Council corr `155f4730-4526-4523-83d0-3ce4c4fc9f1c`: **APPROVED round 1**, 2 advisory
  objections, 8 abstained. Verdict READ from `diagnosis_artifacts` before writing the
  trailer.
- The advisories, answered:
  - `editquality` (medium): "MatchedPermanentFailure not attested to exist" — it exists
    (`validation_drop.go:120`), read first-hand before the submission; the code compiles
    and its tests pass. Lesson kept: cite the EXISTENCE of every composed symbol in
    `grounded_in`, not just the new one — same-package helpers never appear in a diff
    (the RUNBOOK's standing lesson, re-earned).
  - `guardian` (medium ×2): fleet-wide-at-once with no staged rollout — inherent to a
    wire-status fix (the coordinator claims the FIRST response, so the sender must be
    right; guardian's own notes concede the higher-layer alternative cannot work). Acted
    on the actionable part: the post-roll storm-watch is now a named CLOSE CRITERION in
    the bug file, not an assertion.
  - `prior_art_librarian` / `architecture` (asks): AsRetryable test-only claim and the
    two-sites grep were both first-hand this session (grep quoted in NOTES above);
    histogram to be WATCHED post-roll, not asserted — same close criterion.

## 2026-08-06 (later) — v1.0.1261 rolled WITHOUT this fix; commit lands after it

The fresh chassis build (`v1.0.1261`, another session's roll) predates this lane's commit —
the fix is NOT in it (uncommitted at build time, and `make build-*` builds from committed
HEAD). Expected, not a misstep. The commit carries `Council-Reviewed:` (verdict read).
Bug stays in `bugs_open/` per the owner's 08-06 ruling (finished bugs stay), with close
criteria written in the bug file for whoever verifies after the NEXT roll.

## 2026-08-07 — committed `9fa6f923b`; twins deliberately untouched

Commit `9fa6f923b` (8 files, pathspec, `Council-Reviewed:` trailer — verdict read first).
The pattern-check's two untouched-twin advisories are deliberate, not oversights:
- `sendErrorResponseOLD` — legacy, ZERO callers (grep, non-def). Dead code; deleting it is
  not this lane's task and would widen a reviewed commit.
- `sendWorkflowResponse` — the SUCCESS sender; it has no failure to classify, and
  `TestSuccessResponseStatusStillComplete` pins that its envelope is unchanged.
LANDMINES.md deliberately NOT appended from this lane: the file is dirty with another
session's edits (same-file passenger risk), and the trap ("call RetryDisposition, not
MatchedTransientFailure bare, unless you have agentbase's call-order guard") is carried by
RSH-007's register entry, the helper's own doc comment, and the memory topic file.
Remaining to close 207: the next chassis roll + the three close criteria in the bug file.
