# PLAN 2026-08-08 — bugs_open/216: pass the claimed row through the recoverable arm

Successor lane to `bugfix_207_sender_convergence` (cold-start:
`../bugfix_207_sender_convergence/HANDOFF_2026-08-08_continue_here.md`). That lane
proved 207's sender fix live and found, by induction, that the retry it enables dies
downstream: the coordinator's response-driven recoverable arm re-arms the awaited
request (`retry_version`++) and then refuses its own replay
(`RETRY_PAYLOAD_UNAVAILABLE`), failing the parent terminal. Until this is fixed, every
upstream classifier win (195/197/207) buys a bookkeeping write instead of a retry.

## Decision: fix candidate 1 (pass the claimed row through), as the bug file argues

- `ClaimAwaitedRequest` (coordinator.go:442) already RETURNS the full row,
  `request_payload` included. `handleRecoverableError` now takes that row as a
  parameter (mirroring `handleCompleteResponse`'s existing signature) instead of
  re-reading with `GetAwaitedRequest`, whose predicate `status IN
  ('waiting','retrying')` misses by construction after the claim set `'processing'`.
- The in-memory fallback is DELETED from this path, not patched: it could never carry
  a payload (`RequestPayload` is `json:"-"`, RSH-003), so it was a refusal with extra
  steps.
- Both callers hold a freshly-claimed row from a `RETURNING`, so the parameter is
  total: the response switch (coordinator.go:322, claim → `'processing'`) and the
  timeout retry driver (coordinator.go:3448 via `retryExpiredAwaitedRequest`, claim →
  `'retrying'`).
- Candidates 2 (widen the predicate) and 3 (make the fallback load the payload)
  rejected for the reason the bug file gives: both keep two reads of one row in one
  decision — the drift class the platform keeps re-buying.

## Correction to the bug file, found while reading the code (2026-08-08)

The filing marked same-pod survival `[INFERRED]` ("a pod that still holds the creating
process's state object in memory may retain the payload"). The code refutes the
inference: `ProcessResponse` loads state fresh from the DB on every response
(coordinator.go:261 `repo.GetState`), and the payload never enters
`orchestration_states`' awaited_requests JSONB (`json:"-"`, asserted by
`retry_payload_capture_test.go`). So the fallback was payload-less on EVERY
response-driven recoverable — the arm was 100% dead on this path, same-pod included,
not merely cross-pod. Recorded in the bug file as a visible correction.

## Phasing

1. **DONE 2026-08-08** — code change (coordinator.go, state.go comment) + regression
   tests (`response_retry_claimed_row_test.go`): fixed-code pass, mutation run
   (fallback reintroduced) fails test 1 with the 216 shape, full package green.
2. Council gate submission (platform/orchestration/ → gate before/alongside commit);
   commit by pathspec with `Council-Submitted:` trailer; verify HEAD compiles from
   `git archive` (shared tree — local green is not HEAD green).
3. Bump IMAGE_TAG, `make build-agent-chassis` from HEAD, roll, pod-grep positive
   marker (`Retry decided on the claimed awaited row passed through from the claim`)
   + negative control (`Using in-memory awaited request`, expect 0) on every replica.
4. Induction re-run (207 lane's recipe): acceptance is the replayed request CONSUMED
   from the target topic — a retry_version bump is NOT a retry (WRONG_CALLS
   2026-08-07).
5. Then `bugs_open/217` (notifyParentOfFailure hardcodes `error_unrecoverable`) —
   with 216 fixed, converging 217 no longer converts hardcoded-terminal into
   re-arm-then-terminal.

## Bounds

The recoverable arm still caps at `retry_version >= 3` and has no backoff (RSH-006
landmine 3). This fix widens the live retry population to what 197/207 already
promised — no further.
