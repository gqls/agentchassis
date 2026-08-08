# 216 — the response-driven recoverable arm re-arms the request, then refuses its own replay: an `error_recoverable` RESPONSE is terminal (at least cross-pod), and the retry_version bump it leaves looks like a retry happened

**Filed 2026-08-07** by the `bugfix_207_sender_convergence` lane, found live during 207's
post-roll induction on v1.0.1262. **Status: OPEN — FIXED IN CODE 2026-08-08, awaiting
roll + live proof.** Owned by the `bugfix_216_claimed_row_passthrough` lane
(docs024_key_docs_latest/bugfix_216_claimed_row_passthrough/). **Severity: high for the
retry-quality programme** — it is the seam 195/197/207 all deliver INTO, and while it
stands, every classifier improvement upstream buys a bookkeeping write instead of a retry.

> **FIX RECORD (2026-08-08).** Candidate 1 implemented as argued below:
> `handleRecoverableError` now takes the claimed row from its caller (mirroring
> `handleCompleteResponse`), and the doomed re-read + in-memory fallback are deleted;
> both call sites (response switch, timeout retry driver) pass the row their claim's
> `RETURNING` gave them. Regression tests in
> `platform/orchestration/response_retry_claimed_row_test.go` recreate the hostile world
> (unreachable DB, payload-less in-memory copy) and assert the replay reaches the wire;
> the discriminating test was mutation-verified (reintroducing the fallback fails it
> with this file's shape). Council gate: `Council-Submitted:
> fcf8794c-92df-4c8e-9677-5ca284a20cce`. Close criteria: pod-grep the positive marker
> (`Retry decided on the claimed awaited row passed through from the claim`) + negative
> control (`Using in-memory awaited request`, expect 0) on every replica, then re-run
> the 207 induction and CONSUME the replayed request from the target topic — a
> retry_version bump is NOT a retry.
>
> **CORRECTED 2026-08-08 (by the fixing lane's code read):** the `[INFERRED]` caveat in
> step 3 below — that a pod still holding the creating process's state object might
> retain the payload, so same-pod responses may survive — is **refuted**.
> `ProcessResponse` loads state fresh from the DB on every response
> (`coordinator.go:261`, `repo.GetState`), and the payload never enters that JSONB
> (`json:"-"`, the contract `retry_payload_capture_test.go` asserts), so the fallback
> was payload-less on EVERY response-driven recoverable. The arm was 100% dead on this
> path, same-pod included — the scope is wider than filed, not narrower.

> **VERIFICATION STATEMENT (owner ruling 2026-07-31).** First-hand verification, declared:
> the failure was **induced live** (one observed case, cross-pod), the deciding code path
> is **read and cited line-by-line**, and the DB rows are quoted. The fleet-wide scope
> claim (every cross-pod response-driven recoverable dies this way) is the code read
> generalising from one induced case — marked `[INFERRED]` where it appears. A `090` run
> has been submitted alongside this filing for independent verification: run correlation
> `0e7e9640-7b22-4f10-8ea8-1994454993f3` (find it by
> `spec.dispatch_correlation_id`, not the intake id). Read its verdict before building
> the fix on this file's mechanism claim.
>
> **090 OUTCOME READ 2026-08-08 — the run COMPLETED after 5 iterations with NO clean
> verdict on the filed symptom; it did not falsify this file, and its final citations
> independently re-derive the central code fact.** The trail (verdicts in `llm_call_log`,
> `step_name='verdict'`, this correlation — the orchestration rows are reaped and no
> verdict artifact was written): iterations 1–3 UNVERIFIABLE ("no information", per 016b —
> not "confirmed hard"); iteration 4 REFUTED **a misreading of its own** (it took the
> `RETRY_PAYLOAD_UNAVAILABLE` log message's rationale text — "would carry this
> orchestration's own id" — for the firing mechanism); iteration 5 REFUTED **iteration 4's
> revision**, citing the code to show the guard "is logged inside the `if err != nil`
> branch immediately after `types.DecodeRetryPayload(awaited.RequestPayload)` fails —
> i.e. exactly the no/undecodable-stored-payload case", with `RETRY_SELF_ADDRESSED` as the
> separate later check — **which is this file's filed mechanism, restated**. The loop's
> outcome labels grade each round's REVISED hypothesis, not the filing; do not quote
> "REFUTED" against this file without reading what was refuted. First-hand verification
> (live induction + code read + DB row) stands as the primary evidence; treat independent
> verification as still open. Two useful by-products: the loop re-read the induced row
> live (`cef0a691…: status=expired, retry_version=1, payload_present=true`), and it
> surfaced an unexplained sibling symptom — completed workflows whose results fail
> **"message validation failed"** on `complete_workflow` delivery to the parent
> (correlation `aee5853d`, several `*/complete` steps) — unfiled at the time of writing,
> worth its own look.

## The mechanism (read from source, confirmed by induction)

When a child answers a parent's awaited request with status `error_recoverable`:

1. `SagaCoordinator` claims the request: `ClaimAwaitedRequest` (`coordinator.go:442-488`)
   — `UPDATE awaited_requests SET status = 'processing' … WHERE status = 'waiting'
   RETURNING <all columns>`. **The claim returns the full row, `request_payload`
   included.** That returned row is then **discarded** by the recoverable arm.
2. The status switch routes to `handleRecoverableError` (`coordinator.go:2951`), which
   **re-reads** the row: `GetAwaitedRequest` (`state.go:1606-1623`) — predicate
   `status IN ('waiting', 'retrying')`. The row is now `'processing'`, because step 1 just
   made it so. **The re-read misses by construction.**
3. The miss falls back to the in-memory copy (`coordinator.go:2961-2969`,
   "Using in-memory awaited request (DB lookup failed)"). The in-memory copy **cannot
   carry the payload**: `AwaitedRequest.RequestPayload` is `json:"-"` (deliberate,
   RSH-003 — it must never enter the hot `orchestration_states` JSONB), so any state
   loaded from the DB has it empty. `[INFERRED]`: a pod that still holds the creating
   process's state object in memory may retain the payload, so same-pod responses may
   survive — unmeasured; the induced case was cross-pod and died.
4. `awaited.RetryVersion++` is **persisted** (`UpdateAwaitedRequestRetry`,
   `coordinator.go:3103`) — *then* `DecodeRetryPayload("")` fails →
   `RETRY_PAYLOAD_UNAVAILABLE` → `handleUnrecoverableError` (`coordinator.go:3125-3138`)
   → `failWorkflow("Request %s failed: %s")` (`coordinator.go:3271`).

So the row shows `retry_version = 1` and the workflow is FAILED, milliseconds apart. **The
counter bump is written by the same function that then refuses — a retry_version histogram
cannot distinguish this dead arm from real retries.** The timeout-driven path is healthy:
its claim sets `'retrying'`, which the re-read predicate includes (the `state.go:1602`
comment says exactly this — the two claim paths write different statuses and only one was
kept in the predicate).

## The induced case (v1.0.1262, 2026-08-07, all ids live in the DB)

Probe corr `b155c554-0753-4f57-97a0-fcaec5d229d8` (recipe:
`docs024_key_docs_latest/bugfix_207_sender_convergence/`, adapted from 196's two-dispatch
induction): parent `812f770d-53fb-437d-aa29-1d715b4af312` parked awaiting
R=`cef0a691-88bd-48f0-beff-10d51a8da4df` (created on pod `…-dfk4b`); a deadline-exceeded
child failure produced a genuine `error_recoverable` envelope (207's converged sender,
wire-captured, `body.error.recoverable=true`); the envelope was delivered to the parent's
responses topic and processed on pod `…-5ghft` (cross-pod). Result, from
`orchestration_states.processing_history` and `awaited_requests`:

- `08:25:21.5635` — awaited row re-armed: `retry_version` 0→1, new `sent_at`.
- `08:25:21.5677` — `workflow_failed`: `Request cef0a691… failed: workflow failed: step
  probe_deadline failed: … context deadline exceeded` — `handleUnrecoverableError`'s
  message shape, 4ms after the re-arm.
- The DB row **has** a valid payload (`request_payload IS NOT NULL`, correct topic) — so
  the copy the refusal read was not the DB row. Only the in-memory fallback fits.

## Why this matters more than one probe

- `agentbase` has emitted `error_recoverable` on real traffic since 197 rolled
  (v1.0.1259): those responses reach this arm. 197's storm-watch histogram counted
  `retry_version` bumps — which this defect also writes on its way to terminal — so that
  histogram **overstates** how much real retrying the response path has ever done.
  `[INFERRED]` scope pending the 090 verdict; the timeout-driven population is healthy and
  also in those histograms.
- 207 converged the processor senders onto the shared classifier — proven live at the
  wire — but the ~30% deadline-exceeded prize is **not realised** while this arm refuses
  the replay it just booked.

## Fix candidates (ordered by what closes the door)

1. **Pass the claimed row through.** `ClaimAwaitedRequest` already returns the full,
   authoritative row (payload included) at `coordinator.go:340`; hand it to
   `handleRecoverableError` instead of re-reading. One read, one row — the predicate
   mismatch becomes unrepresentable, and the in-memory fallback can be deleted from this
   path rather than patched.
2. Widen `GetAwaitedRequest`'s predicate to include `'processing'`. Smaller diff, but
   keeps two reads of one row in one decision — the drift class this platform keeps
   re-buying (034's twin lists, one level up).
3. Make the fallback load `request_payload` explicitly before deciding. Narrowest; leaves
   both the double-read and the fallback in place.

**Bounds if fixed:** the recoverable arm still caps at `retry_version >= 3` and has no
backoff (RSH-006 landmine 3) — fixing this arm widens the live retry population to what
197/207 already promised, no further.

## Relations

- `bugs_open/207` (fixed, live) — the sender-side convergence whose induction found this;
  its close criteria are met at its own seam and its prize is gated here.
- `bugs_open/217` — filed together: `notifyParentOfFailure` hardcodes
  `error_unrecoverable`, the third unclassified failure sender; fixing 217 without 216
  converts hardcoded-terminal into re-arm-then-terminal.
- `bugs_closed/003` (F2 retry driver — the healthy timeout path), `bugs_closed/129` /
  RSH-003 (replay-not-rebuild; the `json:"-"` design this arm's fallback collides with),
  `bugs_open/075` lineage (the adapter cap branch, unaffected).
