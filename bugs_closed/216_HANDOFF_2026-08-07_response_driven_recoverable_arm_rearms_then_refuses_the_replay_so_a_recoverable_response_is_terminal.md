# 216 — the response-driven recoverable arm re-arms the request, then refuses its own replay: an `error_recoverable` RESPONSE is terminal (at least cross-pod), and the retry_version bump it leaves looks like a retry happened

**Filed 2026-08-07** by the `bugfix_207_sender_convergence` lane, found live during 207's
post-roll induction on v1.0.1262. **Status: CLOSED 2026-08-15 — FIXED + LIVE on v1.0.1266 +
PROVEN BY INDUCTION 2026-08-08, all three close criteria met.** ~~stays in `bugs_open/` per the
owner's 08-06 ruling~~ — **superseded: the owner ruled on 2026-08-15 that this closes**, which
also settles the filing question the 213 lane raised at the foot of this file (see the closure
note there). Owned by the `bugfix_216_claimed_row_passthrough` lane
(docs024_key_docs_latest/bugfix_216_claimed_row_passthrough/). The seam 195/197/207 all
deliver INTO now delivers: a recoverable response produces a replay on the wire.

> **PROOF (2026-08-08, v1.0.1266, all ids in the lane NOTES).** (1) Pod-grep, both
> settled replicas, same exec: positive marker 1, removed-string negative 0, unchanged
> `RETRY_PAYLOAD_UNAVAILABLE` control 1. (2) Induction (207's recipe): parent
> `f5e167c5…` parked awaiting R=`59c49316…` (payload recorded, 600s timeout, sent
> ~16:07); a genuine deadline-exceeded `error_recoverable` envelope was captured and
> re-published to the parent's topic; **the void topic then carried the REPLAYED
> original request at offset 1 — `retry_version=1`, the child's own orchestration id
> `7b2b476c…` (self-address guard held), stamped 16:10:08 — seven minutes before the
> original timeout could first fire (~16:17), so only the response-driven arm can have
> produced it.** Offsets 2–3 (16:15:08, 16:20:08, +300s each) are the healthy timeout
> path's retries — which also proves the response arm RELEASED the row to `'waiting'`
> (a `'processing'` row is unclaimable by `ClaimAwaitedRequestForRetry`). The parent
> survived the response-driven retry and failed only at budget exhaustion (~16:25,
> "timed out after 3 retries" — the healthy-path message, cap wall at 3 intact). On
> v1.0.1262 the same drive died 4ms after the bump with nothing on the wire. Caveat,
> stated: the chassis log lines for the 16:10 event had rotated (<5-min window, 016b)
> before they were read — the wire artifact + DB timestamps carry the proof without
> them. Probe seed and void topic deleted; orchestration rows left as evidence (~24h).

> **FIX RECORD (2026-08-08).** Candidate 1 implemented as argued below:
> `handleRecoverableError` now takes the claimed row from its caller (mirroring
> `handleCompleteResponse`), and the doomed re-read + in-memory fallback are deleted;
> both call sites (response switch, timeout retry driver) pass the row their claim's
> `RETURNING` gave them. Regression tests in
> `platform/orchestration/response_retry_claimed_row_test.go` recreate the hostile world
> (unreachable DB, payload-less in-memory copy) and assert the replay reaches the wire;
> the discriminating test was mutation-verified (reintroducing the fallback fails it
> with this file's shape). Council gate: **APPROVED round 1, 2026-08-08 14:49 —
> verdict read** (`Council-Reviewed: fcf8794c-92df-4c8e-9677-5ca284a20cce`; the commit
> carries the `Council-Submitted:` trailer, which 098 resolves to this approval —
> forward-only forbids an amend). One advisory objection (editquality, low: the
> doc-comment edit is not mechanism coverage — correct, it was counted as
> doc-truthing). Guardian's containment checks, answered first-hand: callers verified
> by independent grep (exactly two, both in coordinator.go; the signature change makes
> a third a compile error); nothing writes the row between either claim and the decide
> (read line-by-line); the test fixtures are fresh but local to the one test file.
> Close criteria: pod-grep the positive marker
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

---

## RECONCILIATION 2026-08-15 from the `bugfix_213` lane — your unfiled sibling is now `bugs_open/274`, and your fix is doing more work than this file says

Two things, neither a criticism of the fix, which is correct and well proven.

### 1. The sibling symptom you flagged is filed, and it is large

This file's opening block recorded *"an unexplained sibling symptom — completed workflows whose
results fail 'message validation failed' on `complete_workflow` delivery to the parent … unfiled
at the time of writing, worth its own look"*. That look happened on 2026-08-14 and it is now
**`bugs_open/274`**: [MEASURED] **~15,000 events across 60 agent types, continuous since
2026-08-03 and still firing** (1,668 in the 20 hours to 2026-08-15). Root cause located —
`notifyParentOfSuccess` builds its reply headers without `sender_agent_type` or a step name, and
the validator requires both, so **that reply can never pass**. Your instinct to flag it was right
and it sat for a week.

### 2. YOUR FIX IS WHAT MAKES 274'S REPLAYS REAL — and that reframes both files

The chain, read at source (`bugs_open/274` §10):

1. A child **succeeds**; its success reply is refused by validation and dropped.
2. `notifyParentOfSuccess` falls through to `notifyParentOfFailure`, which builds the **same
   malformed envelope** but sets `IsError: true` — and `ProduceWithValidation` **exempts error
   messages** (*"those we always send"*). So the failure **is** delivered.
3. That failure's prose contains `failed_transient`, so `perrors.RetryDisposition` classifies it
   **`error_recoverable`** — and it lands in the arm this bug fixed.

**Before `v1.0.1266` your arm refused that replay, so the duplicate work never happened. After it,
the replay reaches the wire.** So this fix — correctly, exactly as designed — now faithfully
retries a large population of **fictional** failures produced by a workflow that actually
succeeded.

**Nothing here says the fix is wrong.** A recoverable response *should* replay; the defect is that
274 manufactures the recoverable response. But it means:

- **This fix is load-bearing far more often than this file suggests.** A large share of the
  recoverable responses it replays are 274's fictions, not genuine transient failures.
- **The duplicate-execution cost belongs to 274 and is [UNQUANTIFIED].** The honest instrument is
  the parent side — `awaited_requests.retry_version` and replayed offsets — not the child's error
  log. On `page-rerender` (4,794 events) a replay is a page rebuild.
- **If 274 is fixed at the header, this fix's traffic should drop sharply.** That is a usable
  before/after for whoever takes 274, and a reason to record this file's current replay volume
  *before* they land it.

### 3. Filing question for the owner, not an action taken here

This file reads **FIXED + LIVE on `v1.0.1266` + PROVEN BY INDUCTION**, and its header keeps it in
`bugs_open/` citing *"the owner's 08-06 ruling"*. CLAUDE.md's current text states the bar for
`bugs_closed/` as **fixed AND live**, which this meets. **[UNVERIFIED whether the 08-06 ruling was
later superseded — I believe it was, on 08-12, but I have not read the ruling itself.]** Flagged
rather than moved: it is your file, and a move against a cited ruling is the owner's call.
⚠ If it does move, name **both** paths on the commit — `git mv` plus a pathspec commit ships a
copy and leaves the original at HEAD.

— `bugfix_213_verifier_producer_join` lane

---

## CLOSED 2026-08-15 — owner ruling, and the one thing that does NOT travel with it

**The owner ruled on 2026-08-15 that this file closes.** That answers §3 above, which had
flagged the filing question and deliberately left it: the file met CLAUDE.md's stated bar
(**fixed AND live**) on 2026-08-08 and was being held open only by a cited 08-06 ruling that
the flagging lane suspected, but had not read, was superseded. It was.

**What is closed is the DEFECT, and it is closed on evidence rather than on age.** The
recoverable arm re-armed the request and then refused its own replay; it now passes the
claimed row through from the claim, both call sites were verified by independent grep, the
regression tests recreate the hostile world, and the fix was proven by live induction with
the replayed request consumed off the wire at `retry_version=1` seven minutes before the
timeout path could have produced it. Council APPROVED round 1
(`fcf8794c-92df-4c8e-9677-5ca284a20cce`).

**What does NOT close with it, and must not be read as closed:**

- **`bugs_open/274` is open, large and still firing** — ~15,000 events since 2026-08-03,
  1,668 in the 20 hours to 2026-08-15. It is the sibling symptom this file's own opening
  block flagged as unfiled on 08-08. **This fix is what makes 274's replays real**: a
  workflow that SUCCEEDS has its success reply refused by validation, falls through to a
  failure envelope that IS delivered (error messages are exempt), and that failure's prose
  classifies as `error_recoverable` — landing in the arm this bug fixed. So a large share of
  the recoverable responses now correctly replayed are 274's fictions rather than genuine
  transient failures. **Closing this file does not reduce that traffic by one event.**
- **The duplicate-execution cost is still [UNQUANTIFIED]**, and the honest instrument is
  parent-side (`awaited_requests.retry_version` and replayed offsets), not the child's error
  log. It belongs with 274. If whoever takes 274 wants a before/after, **record this arm's
  current replay volume before landing the header fix** — after it lands, the comparison is
  gone.
- **Independent verification of the mechanism remains open by this file's own statement.**
  The `090` run (`0e7e9640-7b22-4f10-8ea8-1994454993f3`) completed with no clean verdict on
  the filed symptom; it did not falsify the file and its final citations independently
  re-derive the central code fact, but that is not the same as a confirmation. The primary
  evidence is and remains the live induction plus the line-by-line code read.

Moved `bugs_open/` → `bugs_closed/` with **both paths named on the commit** — a `git mv` plus
a pathspec commit that names only the destination ships a COPY and leaves the original at
HEAD, and `ls` cannot tell you which happened. Verified at HEAD, not at the tree:
`git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 216`.

— closed by the `bugfix_213_verifier_producer_join` lane on the owner's ruling
