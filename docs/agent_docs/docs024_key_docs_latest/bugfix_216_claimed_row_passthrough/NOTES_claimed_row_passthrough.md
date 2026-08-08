# NOTES — bugfix 216 claimed-row pass-through (append-only, newest at the bottom)

## 2026-08-08 — lane pickup, code read, fix + tests

- Picked up from `../bugfix_207_sender_convergence/HANDOFF_2026-08-08_continue_here.md`.
  Ownership checked: `who-owns.py 216` names only the ended 207 lane; live-transcript
  grep (`bugs_open/216|RETRY_PAYLOAD_UNAVAILABLE|ClaimAwaitedRequest` over
  `~/.claude/projects/.../\*.jsonl`) shows every other active session touches the terms
  once, incidentally; only `e68e25cd` (the ended lane, 66 hits) worked it. Queue clear:
  the one `site_work_items` hit on these terms is an unrelated tool-recreation-handler
  filing (status `failed`).
- Mechanism re-verified line-by-line at HEAD (`b320b2a34`), all four steps as filed:
  claim RETURNING carries `request_payload` (state.go:1669-1672 `awaitedRequestColumns`);
  the switch discards the claimed row on the `error_recoverable` arm only
  (coordinator.go:322 — `handleCompleteResponse` on line 320 already receives it);
  `GetAwaitedRequest` predicate `('waiting','retrying')` (state.go:1611) misses the
  `'processing'` row the claim just wrote; fallback row cannot carry a payload
  (state.go:95 `json:"-"`).
- **Sharpening beyond the filing:** the `[INFERRED]` same-pod-survival caveat is
  refuted — `ProcessResponse` always `repo.GetState`s fresh (coordinator.go:261), and
  the payload never round-trips `orchestration_states` (`json:"-"`; the contract is
  asserted by `retry_payload_capture_test.go`). Every response-driven recoverable died,
  same-pod included. Correction recorded in `bugs_open/216`.
- Fix applied (candidate 1): `handleRecoverableError` gains `awaited *AwaitedRequest`
  (last param, mirroring `handleCompleteResponse`); re-read + in-memory fallback
  deleted; both callers pass their claimed row (response path: `awaitedReq` from
  `processResponseClaimWithRetry`; timeout path: `awaited` from
  `ClaimAwaitedRequestForRetry`). Non-nil is structural: ProcessResponse returns at
  :252 on a nil claim. Timeout-path equivalence: nothing writes the row between its
  claim and the decide, so RETURNING == what the old re-read returned there.
- Log line renamed: `Loaded awaited request for retry` →
  `Retry decided on the claimed awaited row passed through from the claim`
  (+ `payload_present` field) — the old text claimed a DB load that no longer happens,
  and the new literal is the deploy's positive pod-grep marker. Negative control: the
  deleted literal `Using in-memory awaited request (DB lookup failed)`.
- Tests: `response_retry_claimed_row_test.go`, two cases. (1) DB unreachable
  (port-1 DSN — the one DB write on the path, `UpdateAwaitedRequestRetry`, is
  log-and-continue) + payload-less in-memory copy (production shape) + payload-bearing
  claimed row ⇒ exactly one replay on the original requests topic, child's
  orchestration_id, retry_version 1. (2) Payload-less claimed row still refuses
  (bugs_closed/129 guard intact). **Mutation run:** reintroduced the in-memory
  fallback ⇒ test 1 FAILS with `workflow failed: Request req-216 failed` — the exact
  216 shape — then reverted; the test could have come out otherwise.
- `go build ./...` green, `go vet` green, full package `go test` green. Diff:
  coordinator.go 17+/18-, state.go 5+/3- (comment truth-up: GetAwaitedRequest's doc
  cited the re-read this fix removes), + the new test file. Verified the two files
  carry no other session's edits before committing (same-file passenger check).

## 2026-08-08 (later) — submitted, committed, tag raced away from under me (fine)

- Council submitted: `SUBMISSION_CORR = fcf8794c-92df-4c8e-9677-5ca284a20cce`
  (`SUBMISSION_2026-08-08_council_216.json` in this dir). Run picked up fast — at
  `review_guardian` within minutes of publish, no 29-minute queue this time.
- Committed `22899b809` (fix + tests + bug-file update + lane docs, pathspec,
  `Council-Submitted:` trailer), then `9036d19d0` (ratchet line). HEAD re-verified from
  a clean `git archive`: `go build ./...` + orchestration tests green at HEAD, not just
  in the (other-sessions'-WIP-bearing) working tree.
- **IMAGE_TAG moved v1.0.1264 → v1.0.1265 (uncommitted) between two reads minutes
  apart** — another session is preparing the next build. Deliberately did NOT bump
  again: the deploy procedure is the owner's whole-fleet
  `make release redeploy-agents` (owner, 2026-08-03), and any roll at v1.0.1265 built
  from post-`22899b809` HEAD ships this fix. Verification (markers in the RUNBOOK)
  is owed at whichever roll lands next, against the running pods, per replica.
- Bug-file sharpening recorded in `bugs_open/216` as a visible correction: the
  same-pod `[INFERRED]` survival caveat is refuted (fresh `GetState` at
  coordinator.go:261 + `json:"-"` ⇒ the fallback was payload-less on every
  response-driven recoverable).

## 2026-08-08 (later still) — council APPROVED round 1; verdict read in full

- `complete_approved` at 14:49:53 — ~5 minutes after dispatch pickup, no queue this
  time. Decision: "approved with 1 advisory objection(s) — none high-severity",
  8 abstained, 9 seats reviewed. Verdict body read from `diagnosis_artifacts`
  (`kind='council_report'`, corr `fcf8794c`). ⚠ Gotcha re-learned: the runbook's
  "latest doc_notes row" query returned ANOTHER submission's verdict (concurrent
  council traffic) — filter `body LIKE '%<corr>%'`, never trust LIMIT 1 bare.
- The one advisory objection (editquality, low): the state.go doc-comment edit is
  comment-only and should not count toward mechanism coverage. Correct — it was
  submitted as doc-truthing, no action.
- Guardian objected containment-only (medium: verify the "exactly two callers" claim
  independently; low: confirm nothing mutates the row between the timeout path's claim
  and its call; low: confirm the test fixtures aren't fresh shared infra). All three
  were verified first-hand before submission and the answers are now in the bug file's
  fix record. debug_historian's low objection (name the deploy-verification step) was
  already answered by the RUNBOOK's pod-grep markers — recorded there before the
  submission went in.
- Commit stays `Council-Submitted:`-trailed (forward-only forbids an amend); 098
  resolves it to this approval automatically. The verdict IS read, so citing
  `Council-Reviewed: fcf8794c-92df-4c8e-9677-5ca284a20cce` in prose is honest — it just
  cannot retroactively enter the commit message.
- Fleet baseline at verdict time: both agent-chassis replicas on **v1.0.1264**
  (started 13:08Z) — pre-fix. IMAGE_TAG sits at v1.0.1265 uncommitted (another
  session's bump). The fix ships with whichever whole-fleet release next builds from
  post-`22899b809` HEAD.

## 2026-08-08 evening — ROLLED (v1.0.1266), VERIFIED AT THE POD, PROVEN BY INDUCTION

- Roll detected ~16:00Z (monitor): chassis went v1.0.1264 → **v1.0.1266** (tag raced
  again, 1265→1266 — irrelevant, provenance is proven at the binary). Rollout churned
  through three replicasets; after `rollout status` settled on `856dff6b46` (2
  replicas), pod-grep same-exec on BOTH: `POS=1` (new marker) `NEG=0` (removed
  fallback string) `CTRL=1` (`RETRY_PAYLOAD_UNAVAILABLE`, unchanged — pipeline+guard
  control). Two mid-churn execs returned empty (terminating pods) — re-ran on the
  settled set rather than trusting silence.
- **Induction, per the RUNBOOK (all four criteria):**
  - Seeded `test-207-parent`, created void topic, held the 300s post-restart dispatch
    freeze. Dispatch 1 `CORR=32a4c28e… ORCH=f5e167c5…` parked at `call_child`
    awaiting **R=`59c49316-bdc6-41e4-85ca-cff3fda59ce5`** (`retry_version=0`,
    `payload_present=t`, topic=void, `timeout_seconds=600`, sent ~16:07:0x).
  - Dispatch 2 (`CHILD_ORCH=b417c210…`, inline workflow, `pg_sleep(5)` under
    `local_action_timeout_seconds:0.001`) FAILED deadline-exceeded as designed.
  - **BOTH failure senders answered R on legacy `system.generic.responses`** (196's
    header-drop wrinkle): `error_unrecoverable` `CHILD_ORCHESTRATION_FAILED`
    (`notifyParentOfFailure` — **bugs_open/217 observed live again on v1.0.1266**,
    16:07:45) then 207's converged `error_recoverable` (16:07:46). Unconsumed there,
    so the probe chose which drives the parent — 217 isolated away by construction.
  - Re-published the `error_recoverable` envelope byte-identical (headers included) to
    `system.agent.generic.responses` (PUBLISH_OK ~16:10).
  - **THE PROOF — void topic offset 1: the replayed original request,
    `retry_version=1`, orchestration_id `7b2b476c…` (child's own; parent id separate —
    RETRY_SELF_ADDRESSED held), `timestamp=2026-08-08T16:10:08.36Z`.** The original
    timeout could not fire before ~16:17 — a 16:10:08 replay is only reachable through
    the response-driven arm. Offsets 2–3: 16:15:08 / 16:20:08 (+300s each), sender
    `…f86mr` — the healthy timeout path, which could only claim the row because the
    response arm had released it to `'waiting'`.
  - Post-acceptance trajectory exactly as the RUNBOOK predicted: nothing answers the
    void topic, retries exhausted at the wall of 3, row `status='error'`, parent
    FAILED ~16:25 with "timed out after 3 retries" (the healthy-path message — NOT the
    216 signature "Request … failed: workflow failed", which on v1.0.1262 arrived 4ms
    after the bump).
  - **Caveat, honestly:** criterion 4 (the chassis log lines at 16:10) was NOT
    captured — a turn gap put the event outside the <5-min log window (016b landmine).
    Criteria 1–3 (wire artifact + DB timestamps + pod-grep) carry the proof; the log
    line would have been redundant corroboration.
- Cleanup done: probe seed `DELETE 1`, void topic deleted (list shows 0). Orchestration
  rows (`f5e167c5…`, `b417c210…`) left as evidence — reap ~24h, quoted here same-day.
- **216 is FIXED + LIVE + PROVEN. Next: `bugs_open/217`** — its mechanism re-confirmed
  live today (envelope above); with 216 fixed, converging `notifyParentOfFailure` no
  longer converts hardcoded-terminal into re-arm-then-terminal.
