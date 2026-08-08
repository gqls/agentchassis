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
