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

## 2026-08-07 — v1.0.1262 CARRIES THE FIX; induction run; criteria met; TWO new defects found downstream

- Pod-grep (criterion 1): both replicas — added marker 1, negative 0, positive control 1.
- No real orchestrated failures since the roll (fleet quiet, 33 pipelines parked
  yesterday), so induced: adapted 196's two-dispatch recipe (their seeds + missteps
  pre-applied; SEED_test_207_probe.sql in this dir). Parent `812f770d` parked awaiting
  R=`cef0a691`; flat child with INLINE workflow (query_database +
  local_action_timeout_seconds 0.001 — RSH-004's machinery) failed exactly as designed.
- **The flip, live**: pod log `disposition=error_recoverable, matched='deadline exceeded'`
  (corr `b155c554`); wire envelope `status=error_recoverable, body.error.recoverable=true`
  (partition 1 offset 45163 on legacy `system.generic.responses` — 196's header-drop
  wrinkle holds, the reply headers do not survive this probe intake path even though the
  child execCtx logged reply_to_request_id).
- Criterion 2: re-published the captured envelope byte-identical to the parent's real
  topic → coordinator accepted it, `retry_version` 0→1 (sent_at 08:25:21.5635). **Then
  the parent FAILED 4ms later** — chased it instead of stopping at the green number:
  - `ClaimAwaitedRequest` sets `status='processing'` and RETURNS the full row;
    `handleRecoverableError` discards it, re-reads via `GetAwaitedRequest`
    (`status IN ('waiting','retrying')`) — misses by construction — falls back to
    in-memory whose `RequestPayload` is `json:"-"`-lost → `RETRY_PAYLOAD_UNAVAILABLE` →
    `handleUnrecoverableError`. DB row HAS the payload (queried). **Filed
    `bugs_open/216`**; 090 submitted, run corr `0e7e9640-7b22-4f10-8ea8-1994454993f3`.
  - The same failure also drew a SECOND answer to R, one second EARLIER, from
    `notifyParentOfFailure` — hardcoded `error_unrecoverable` (coordinator.go:3924/3937).
    First response claims. **Filed `bugs_open/217`** (census correction included:
    `TimeoutMonitor.sendTimeoutResponse` helpers.go:409 is a sibling suspect, liveness
    unmeasured; adapters have their own stamps, out of scope).
- Criterion 3: histogram 25025/2126/1413/126, wall at 3 holds, zero above.
- WRONG_CALLS row appended: my criterion 2 measured the re-arm, not the retry — the
  bump is written by the same function that then refuses.
- Cleanup: probe seed deleted, void topic deleted, parent terminal (FAILED) — no reaping
  needed. Orchestration rows left as evidence (~24h retention; ids above).
- 207 status: FIXED + LIVE + PROVEN at its own seam; stays in bugs_open per owner 08-06
  ruling; the ~30% prize is gated on 216 (and 217 for child-orchestration failures).

## 2026-08-08 — 090 verdict read for 216: no clean verdict, filing stands; verdict recovery route recorded

- The run COMPLETED (5 iterations) but wrote NO verdict artifact and its orchestration
  rows reaped within ~24h — the verdicts were recovered from `llm_call_log`
  (`step_name='verdict'`, correlation `0e7e9640…`). RUNBOOK note: that is the durable
  route to a 090 conclusion once the state rows are gone.
- Trail: 3× UNVERIFIABLE, then iter-4 REFUTED a misreading of its own (log rationale text
  taken for the mechanism), then iter-5 refuted iter-4 while citing the code in a way
  that RESTATES the filed mechanism (guard fires in the DecodeRetryPayload err branch).
  Outcome labels grade each round's REVISED hypothesis — recorded prominently in 216 so
  nobody quotes "REFUTED" against the filing without reading what was refuted.
- First-hand verification remains the primary evidence for 216; independent verification
  remains OPEN.
- By-product surfaced by the loop, unfiled: completed workflows failing
  "message validation failed" on complete_workflow delivery to the parent (corr
  `aee5853d`, several */complete steps). Left for the next thread — grep bugs first.

## 2026-08-08 (evening) — 217 taken up (new session)

- Ownership: `who-owns 217` → only this lane's finished dirs; live transcript tails all
  wrapped (216 done, 207 handed off); `site_work_items` queue clean on
  notifyParent/unrecoverable/failure-sender matches. Taken.
- Validity: `coordinator.go:3924/3937` literals unchanged at HEAD — bug live.
- **Import cycle confirmed** (the bug file's open question): `platform/messaging` →
  `platform/orchestration` via processor.go:21 + validation_drop.go:27, so the
  coordinator cannot import `messaging.RetryDisposition`. Chosen route: extract the
  classification core to `platform/errors` (leaf; stdlib-only imports, checked), keep
  messaging re-exports so the pinning tests and agentbase's source-scan test
  (`agent_test.go:109/138` asserts `messaging.Matched*` calls in agent.go) stay green
  and agentbase stays untouched.
- **Population measured** (`severity='fatal'` is written only at coordinator.go:3917,
  after the parent-exists check — the fatal rows ARE this sender's sends):
  ```sql
  -- 14d, needle logic replayed permanent-first (perm needles case-sensitive, transient folded)
  SELECT count(*), count(*) FILTER (WHERE perm), count(*) FILTER (WHERE NOT perm AND trans) ...
  -- → 11,970 total | 4,756 permanent-terminal | 6,239 RECOVERABLE (52%) | 975 unclassified
  -- needle split of the flip: connection 4,163 (ERR_TUNNEL_CONNECTION_FAILED dominant),
  -- deadline exceeded 1,989 (firecrawl POSTs), timeout 87
  -- chain depth via count of 'workflow failed:' per message: depth 0 = 8,557, depth 1 = 3,417, NO depth >= 2
  ```
- **TimeoutMonitor.sendTimeoutResponse is DEAD** — `NewTimeoutMonitor`,
  `MonitorChildOrchestration`, `MonitorRequest(`, and `TimeoutMonitor{` have zero hits
  outside helpers.go and tests (two spellings tried per the grep-absence rule). The bug
  file's sibling suspect resolves to "no convergence needed"; recording in 217.
- Plan: `PLAN_2026-08-08_217_notify_parent_disposition.md` (this dir). Amplification
  risk sized there: measured depth ≤ 1 ⇒ worst case (1+3)² = 16 innermost runs, monitors
  named, terminal-exhaustion marker recorded as the follow-up if storm-watch fires.

## 2026-08-08 (evening, cont.) — 217 council verdict READ: APPROVED r1; objections answered

Verdict landed 18:21Z, ~25 min after submission (`471a969e…`): **approved with 2
advisory objection(s), none high-severity** — 9 reviewers, 8 abstained, no truncation.
Commit `b19ef6930` went in pre-verdict with `Council-Submitted:` (correct; 098 credits
it at report time). Seat-by-seat answers, on the record:

- **guardian obj 1** ("no evidence the RSH-006 storm-watch / fatal-rate monitors exist
  before this merges"): fair — they are OPERATOR-RUN QUERIES, not deployed automation.
  Answered by writing the exact SQL into the bug file's close criteria (appended
  today), so the "monitor" is a runnable check, not a promise. The close criteria gate
  the CLOSE, not the merge — the roll is what arms the behaviour.
- **guardian obj 2** ("load-bearing behaviour change to foundational plumbing"):
  acknowledged, not disputed — it is why the close criteria demand the induction AND
  the storm watch before the file moves anywhere.
- **prior_art obj 1** (TimeoutMonitor deadness rested on symbol-name greps —
  asserted-absence pattern): re-checked wider, post-verdict: bare-type grep
  `TimeoutMonitor` across platform/internal/pkg/cmd → ONE hit outside helpers.go, a
  comment in reply_delivery_adoption_test.go:125; deployments/ → zero;
  `agent_definitions.default_config` LIKE both spellings → 0 rows. A Go method needs a
  receiver, no Go code constructs one, and config wiring can only name Go-registered
  actions. Dead stands, now on three legs.
- **prior_art obj 2** (cycle + platform/errors-shape claims "trusted from a session
  grep"): they were read first-hand this session (processor.go:21,
  validation_drop.go:27), and the COMPILER now proves both: the committed tree builds
  with orchestration→platform/errors beside messaging→orchestration; were the leaf
  claim wrong the build would have failed.
- **prior_art obj 3** (landmine-indexed symbol; guard text must be read): it WAS read
  before editing — the 090-verdict-recovery LANDMINES entry is why Code/Message stay
  verbatim; the code comment at the sender says so.

No WRONG_CALLS entry from this session: no recorded claim was refuted. The near-miss
worth naming: the first TimeoutMonitor deadness check was two spellings; the council's
prior-art seat is what pushed it to three legs. That is the gate working, not a wrong call.

## 2026-08-08 (night) — 217 PROVEN LIVE on v1.0.1269: all three close criteria met

Roll landed twice while verifying (v1.0.1268 pods 18:57Z were replaced by v1.0.1269
pods 22:01/22:02Z mid-check — re-ran the pod-grep on the survivors, per the
snapshot-goes-stale rule).

- **Criterion 1 — pod-grep, both replicas** (`8skkf`, `rd27g`): POS
  `retry disposition decided at the child-orchestration failure` = 1,
  POS `disposition_matched` = 1, NEG synthetic = 0. Marker pair from
  `scripts/pick-pod-marker.py b19ef6930` (which also listed 8 comment-only strings
  as traps — the moved headers are NOT in the binary; do not probe with them).
- **Criterion 2 — induction** (no natural firing in the pods' first 15 min; used the
  216 runbook recipe verbatim). Ids: CORR `1b4b43f2-bf13-425c-adec-a3d5f30265fd`,
  parent ORCH `b9b7f126-6363-40cb-b190-e3c3cc50a661`, R
  `a64d935a-6ff8-43ef-87c1-142aece85561`, chassis-minted child orch `1cfb581e…`.
  - **The flip, live at 22:14:04.503Z** (pod log, captured before rotation ate it):
    `disposition=error_recoverable, disposition_matched="deadline exceeded"` from THIS
    sender — the envelope that was hardcoded terminal at 16:07:45 yesterday.
  - **Wire**: legacy `system.generic.responses` part 2 offset 36692, key=R:
    `status=error_recoverable`, `body.error.code=CHILD_ORCHESTRATION_FAILED`,
    `recoverable:true` at 22:14:04.503 — and the processor's converged envelope
    (part 1 offset 45931) 1.0s later, now AGREEING instead of losing the claim.
    This sender still answers first; both verdicts now match.
  - Re-published offset 36692 byte-identical (single kcat container, consume→file→
    produce, headers reconstructed) to `system.agent.generic.responses`.
  - **Acceptance**: R `retry_version=1` + `status='waiting'` (22:17:21.673); parent
    still `AWAITING_RESPONSES`, `error` NULL; **THE PROOF: void topic offset 1 = the
    replay** — `retry_version:1`, fresh `message_id`/`timestamp` (22:17:21.675),
    otherwise byte-matching offset 0. (Replay carries `timeout_seconds:300` where the
    original had 600 — same as 216's run shape, noted, not a falsifier.)
  - Log corroboration is PARTIAL and honestly so: the claimed-row marker line rotated
    away before the grep — **oldest retained line on both pods was ~20 SECONDS old**
    (post-roll surge; the 216-era "minutes" figure is optimistic under load). The
    decisive falsifier (parent FAILED ms after re-arm) is refuted by the DB directly.
- **Criterion 3 — storm watch** (24h): retry_version 0→2,392 · 1→58 · 2→4 · 3→9 ·
  **nothing above 3**. Fatal-rate 41–146/hr across 6h, 22:00 hour mid-range (83) —
  no post-roll amplification signal. Re-check week-over-week per the bug file.
- Probe expected to exhaust at the wall (nothing answers the void topic) — that is
  RSH-006 binding, not a failure. Cleanup after exhaustion: seed row + void topic.
