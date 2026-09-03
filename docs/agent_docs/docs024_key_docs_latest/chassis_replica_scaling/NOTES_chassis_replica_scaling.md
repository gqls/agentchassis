# NOTES — chassis replica scaling

Append-only, newest at the bottom. Technical log: what was tried, what the
system said, and every misstep.

---

**2026-07-20 (bugfix-003 thread).** Directory created with the problem
statement (`PLAN_2026-07-20_…`) and `README_where_we_are.md`. This NOTES file
was created later the same day by the fixing-throughput thread — a misstep in
itself: the standing five should exist from the start.

---

**2026-07-20 ~21:00–22:30Z (fixing-throughput thread) — working the problem
statement up into the plan (PLAN §§8–13).**

Owner steer that triggered this: "think hard about the best long term
solution, we will be dealing with thousands of domains." That answers §7 Q1/Q2
— throughput is in scope, not just deploy safety.

Evidence trail, in order:

- **§5A's open question (must a response land on its sender?) settled by code
  read.** `ProcessResponse` (`platform/orchestration/coordinator.go:166–301`):
  DB-loaded state, atomic `ClaimAwaitedRequest` — pod-agnostic. BUT
  `coordinator.go:271–277` then drops mismatched-pod responses AFTER the
  claim, without releasing it. Stamp: `SetExecutingStep`
  (`state.go:1109`, `HOSTNAME`), callers `coordinator.go:698`/`:820`; nothing
  clears or re-stamps it. Only reset path is CLAIM_RECOVERY inside
  `processResponseClaimWithRetry`, which needs a SECOND response for the same
  request_id.
- **Live check of the drop:** `kubectl logs deploy/agent-chassis --since=12h |
  grep -c "owned by different pod"` → 0. **Weak evidence, nearly misread:**
  the pod was only 115 min old, and the DB showed the in-flight population at
  that moment was ONE orchestration (EXECUTING_STEP, stamped by the current
  pod) and zero AWAITING_RESPONSES — so exposure was ~nil and the 0 proves
  nothing about a busy roll. RUNBOOK R1/R2 record the honest way to check.
- **Checked the diagnosis queue before filing** (CLAUDE.md rule): the
  synchronous-handler claim was already filed by the bugfix-030 thread at
  19:23Z (corr `78470372-7617-40e4-888c-66cac94006bf`, still
  `awaiting_diagnosis` — queued behind the very backlog it describes).
  Distinct mechanism from mine, so filed the ownership-drop claim separately:
  corr `2d02d62a-7d96-41f0-a82b-e1ebd7ef5d6b`, ~21:45Z, verdict pending.
  Trigger advisory: local HEAD is 580 commits ahead of origin, and the
  diagnosis reads origin/085 — the cited code paths pre-date the divergence
  (git log -S shows the ownership check is old), so they should be visible to
  it. [ASSUMED — if the verdict can't find the symbols, push and re-run.]
- **Grepped `/bugs_open/` + `/bugs_closed/` for the mechanism first**: no file
  mentions `ProcessingNode` / "owned by different pod". Unfiled before today.
- **`who-owns`**: 030 → `dispatch_queue_serialisation` (ACTIVE, same-day
  measurement work); 003 → `bugfix_003_spawn_loss` (ACTIVE, F1 live, F2/F3
  staged). The plan routes P0 INTO their fixes rather than forking.
- **Volumes** (live DB): orchestration_states 7-day group-by →
  520/2,237/2,167/2,843/3,872/3,480/2,046/1,918 per day (07-13→07-20). Sites:
  11 deployed, 17 pool, 1 system. scheduled_tasks: 14 enabled of 27.
- **`orchestration_requests` is a dead table**: full intake-shaped schema (\d
  verified), FK to orchestration_states, zero Go references (tree-wide grep).
  Recorded in PLAN §8.4 so nobody assumes it is live.
- Wrote PLAN §§8–13 (the design: thin ingest → Postgres-claimed workers;
  Kafka delivers, Postgres decides), appended the owner prose to
  `README_where_we_are.md`, wrote the first SUMMARY, created this file and the
  RUNBOOK.

Missteps this session, recorded per the rules:

- Log-grepped with `--since=12h` against a 115-minute-old pod and nearly read
  "0 hits" as strong absence. The DB exposure query is what showed the window
  was empty of at-risk orchestrations. A 0 from an idle window is not a
  refutation.
- First sites query guessed `status='active'`; the vocabulary is
  pool/deployed/system. `\d` (or a GROUP BY probe) first, as CLAUDE.md says.
- First DB probe used `correlation_id LIKE '78470372%'` — `correlation_id` is
  uuid-typed and `LIKE` fails with `operator does not exist: uuid ~~ unknown`.
  Cast `::text` first (RUNBOOK R5).

---

## 2026-07-26 — the ownership discard this PLAN is blocked on is GONE (from the 075 thread)

Contributed by the bugfix-003/075 thread, not by this workstream. Two things
here change the premises of PLAN §§ on the response path — please re-read them
before the next step rather than trusting the text as written.

1. **`coordinator.go`'s "owned by different pod, ignoring" discard no longer
   exists.** It is now a takeover: `StateRepository.TakeOverOrchestration` CASes
   `processing_node` to the consuming pod (guarded on the previous holder, does
   not touch `version`), logs `ORCHESTRATION_TAKEN_OVER`, and processes the
   response **either way**. Committed 2026-07-26, INERT until a chassis roll;
   case is `bugs_open/075`. The PLAN's "Option A must ship in the same change as
   the removal of this check" is therefore half-satisfied from the other side:
   the removal has shipped first, which is safe because the discard destroyed
   responses on its own (a dead owner's name never matches anything), whereas a
   shared group without the removal destroys ~(N−1)/N of them.

2. **The PLAN's description of `SetExecutingStep` stamping `ProcessingNode`
   needs a correction.** That assignment never reached the database:
   `UpdateStateWithVersion`'s UPDATE column list omits `processing_node`. The
   column is written at row creation and (now) by `TakeOverOrchestration`, and
   by nothing else. So the pre-existing wedge described in the PLAN was not
   "stamped by the executing pod" — it was "stamped by the creating pod, for
   ever".

**What this hands you, and what it does not.** With the gate gone, the remaining
double-apply risk under a shared group with replicas≥2 is the one the PLAN
already names: `processResponseClaimWithRetry`'s CLAIM_RECOVERY resets *any*
claimed-but-unprocessed request to `waiting`, including one claimed milliseconds
earlier by a live pod, so two pods can both claim one response. That is now the
**only** thing standing between this workstream and a shared group, and it is
deliberately NOT fixed by the 075 change (harmless at replicas=1; fixing it
blind would have been scope the council could not judge). Suggested shape when
you get there: guard the reset on staleness for `status='processing'` rows
(`processing_started_at < NOW() - INTERVAL 'N'`) while keeping the immediate
reset for `retrying`/`expired`, which F2 relies on for the late-response window.

Checked while doing the above (2026-07-26): **no multi-replica service owns any
orchestration row today** — all agent-chassis rows (replicas 1), single-pod
spawned Job agents, business-intel (replicas 1); core-manager (2) and
reasoning-agent (3) own none.

---

## 2026-07-28 — workstream ADOPTED; owner approved the whole P1→P2 programme

This session adopted the workstream (it was design-only and unowned; the restart
list's §A "do first: build P1" stands). The owner was asked directly and chose
**the whole programme** — P1 (thin ingest + claim workers), then responses
through the pool, then P2 (shared response group + replicas 2–3), then the
dispatch fan-out handoff — **with a written read-out at each phase boundary**
(appended to README_where_we_are + said in chat). Roll policy ruled: **gate on a
quiet lane** (no `review_*`/`gate_*`/`verdict`/`route` step in flight), not
owner-designated windows. Full build plan (phases, schema, claim design,
verification): `~/.claude/plans/the-processing-of-work-ancient-bunny.md`; the
substance will land in this directory's PLAN as annotations + this file as the
build proceeds.

**P0 items from the PLAN, discharged today:**

- **The two filed diagnosis claims, resolved as far as they resolve.** Corr
  `78470372` (single synchronous consumer): work item `complete`,
  `diagnosis_artifacts` holds the bundle (iteration 1) but no verdict body I
  could locate. Moot in substance — the mechanism has since been confirmed twice
  independently: 030's lane fix worked exactly as the mechanism predicts, and
  096's 07-27 reproduction measured 78–126 ms hand-offs. Corr `2d02d62a`
  (ownership discard): work item `failed` 2026-07-20, no artifacts — also moot,
  the mechanism was confirmed and FIXED from the 075 side
  (`TakeOverOrchestration`, live since v1.0.1174). Neither refutes the design.
  [Checked live: `site_work_items` + `diagnosis_artifacts`, this session.]
- **096's lane split is COMPLETE — done by the oufe thread, not us.** The
  overnight roll (v1.0.1180, ~22:06Z 07-27) carried the council-lane manifest;
  `f8947b9b8` (08:08 today) flipped `097` to `system.agent.council-gate.requests`
  after verifying the lane end-to-end with a cheap rerender. The dominant
  head-of-line source is off the generic lane while P1 is built. Nothing for
  Phase 0 to do there.
- **The spawn "race" under Candidate 4 was REFUTED by its own filer**
  (`47d700946`, 096's 07-28 correction): the wrapper failed at `call_council`
  with "timed out after 3 retries" — a **non-response**, not an
  early-response-lost race. `agent_error_log` shows 79 `spawn_dispatch` timeouts
  in 29 h for `build-pipeline-trigger` (orchestration_states undercounts: a
  timed-out awaited request does not reliably fail its orchestration). Owned by
  bugs_open/029 (active). **Consequence for our Phase 3:** the hoped-for side
  effect — durable response intake dissolving the ~1.5 s orphan window — is now
  UNLIKELY to fix Candidate 4, because the failure is the response never
  arriving, not arriving early. The re-test stays in Phase 3 (cheap, and the
  measurement is owed to 029 either way), with expectations stated here first.
- **The mutable-state audit the P1 pool needs is half done already**: this
  file's 07-26 entry records 318 package-level vars under
  `platform/orchestration/actions/` with exactly one benign startup-only write.
  CS-2 still owes the same pass over `platform/orchestration` (non-actions) +
  `platform/agentbase` before the council submission.

**One correction this plan makes to this file's 07-26 entry:** the
CLAIM_RECOVERY staleness guard is gated there as "when you get to a shared
group" (P2). It is actually a prerequisite for **any** concurrent response
processing, including a single-pod worker pool — two workers handling duplicate
deliveries of one response reach the same reset. So it ships FIRST (CS-1, own
council run), before the pool exists.

---

## 2026-07-28 — Phase 0 baseline: sequential 5/5 COMPLETED; concurrent 0/5. The layer below the lane is the next constraint, with evidence.

Method: cheap `page-rerender` dispatches on vonc.com (the designated test
site), the oufe TRIGGER envelope, five different pages (about, catalyst,
judge, maker, mentor — all 0 NULL-content sections). Same five pages, two
shapes, ~4 minutes apart, on pod v1.0.1180 (11 h old — outside every
post-roll window):

**Sequential (publisher slower than the drain — kcat pod spin-up spaced
arrivals ~15 s apart).** Published 09:42:12–09:43:14Z; offsets 105496→105501
(exactly +5, nothing else in the window); **5/5 COMPLETED**, per-orchestration
created→terminal 6.0–8.2 s. Closed mass balance: 5 published, 5 consumed, 5
terminal, 0 unaccounted by 09:43:55Z.

**Concurrent (all five inside 7.8 s — faster than one segment, so a genuine
burst).** Published 09:46:15.0–09:46:22.8Z. Lane consumed serially over ~37 s
(creations 09:46:09.6→09:46:46.4). Outcome: **0/5 COMPLETED — 4 FAILED at
`deploy_page` with "Request timed out (code: TIMEOUT)"** (error stamped
~09:58–09:59, i.e. ~12 min of F2 retries), **1 WEDGED** at `check_skipped`,
`EXECUTING_STEP`, `updated_at` frozen at 09:46:30 — the exact idle-orchestration
census class from the foot of bugs_closed/030. Checked before concluding: the
response lane is alive (`max(processed_at)` was 10 s old at 10:07), so the
wedge holds no lane; it is a parked row only the watchdog would find.

> **CORRECTED 2026-07-28, twenty minutes after writing the paragraph below —
> before it was committed, caught by reading bugs_open/029's same-morning
> entry.** I first wrote this up as "five concurrent spawns → 4/5 timeouts, the
> 029 family under concurrency". Then 029's 10:00 re-measure recorded **a
> chassis roll to v1.0.1182 at 09:55:02Z — mid-way through my burst's await
> window.** A roll kills in-flight awaits by design, and the ~300 s-to-20-min
> post-roll degraded window follows it. The F2 retry cadence means the first
> retries pre-date the roll and still got no response `[INFERRED from the ~3 min
> retry cycle; not confirmed per-request]`, so concurrency may still be the
> cause — but the burst's failures are CONFOUNDED and the attribution is
> withdrawn until the clean re-run below. The sequential 5/5 baseline (09:42–43)
> pre-dates the roll entirely and stands.

**What survives unconfounded:** the lane consumed the burst strictly serially
(~37 s for five, creations spaced 6–20 s); the wedged row froze at 09:46:30,
nine minutes BEFORE the roll, so it is a genuine parked specimen of the
idle-orchestration census class, not a roll casualty.

**Re-run required (and scheduled this session): same 5-page burst on a pod
>25 min past the 09:55 roll.** If it completes, the concurrency claim was a
roll artefact and P1's enablement risk shrinks; if it fails the same way, the
spawn/response layer genuinely cannot take 5 concurrent spawns and that
evidence goes to the 029 lane. Either way CS-2's enablement verification must
measure COMPLETIONS, not just lane LAG. `[ALSO UNSEPARATED: all five pages are
one site, so same-site deploy contention vs generic spawn-loss needs a
cross-site burst to distinguish.]`

## 2026-07-28 ~10:40 — the clean re-run: 0/5 AGAIN, and this time the mechanism is caught end to end. It is a QUEUE-DEPTH vs AWAIT-TIMEOUT treadmill, not a spawn race and not the roll.

Re-ran the identical 5-page burst at 10:25:26–28Z on the stable v1.0.1182 pod
(30 min post-roll, start time re-verified before firing). **All five FAILED at
`deploy_page` at 10:37:30–35 — exactly the 4×3-minute retry budget, same
signature as the confounded run.** The roll explanation is dismissed. What the
logs and rows show, hop by hop, for the `maker` run (corr `5dec2cdc…`):

1. `deploy_page` is not a spawn at all — it is `git_commit`, a **call to the
   git-adapter** (step config read from `agent_definitions`; the adapter is
   one of the eight explicitly sequential-by-design services).
2. Requests sent 10:25:30 / 10:28:32 / 10:31:33 / 10:34:35 — each timed out
   at +3 min (`awaited_requests.timeout_at`), each retry re-published and so
   **rejoined the BACK of the adapter's queue**. The adapter's queue was
   minutes deep with the fleet's own build traffic (seven
   `agent-build-dispatch-loop` pods in the window) — and note the adapter had
   also just restarted in the fleet-wide 09:55 roll.
3. The adapter DID answer the 4th request — pod `…-tdpcs` log:
   `Message processed, processing_time: 3.0s`, success response for request
   `85891169…` produced to `system.agent.generic.responses` at **10:34:45**,
   inside that await's window.
4. The chassis processed that response at **10:37:35** — ~2m50s later,
   because the RESPONSE lane is the same one-goroutine serial design and was
   itself queued — **five seconds after the await's fourth timeout had
   expired the row and exhausted the budget.** The row ends `status='error'`;
   the orchestration FAILED with a success response sitting processed beside
   it.

**The mechanism, stated once:** an await fails not when its callee is slow
but when `(callee queue delay) + (response lane delay) > await timeout
(3 min)`, and the F2 retry makes it a **treadmill** — each retry re-queues at
the back, so once the inequality holds it holds for every attempt until the
whole queue drains inside one window. Sequential dispatch never triggers it
(shallow queues); FIVE concurrent dispatches plus ambient fleet traffic did,
twice, reproducibly.

Consequences for the programme, in order:
- **Phase 3 (responses through the pool) is load-bearing, not optional** — the
  response lane's ~3-minute delay was half the inequality.
- **CS-2 request-side alone would NOT have saved this batch** (the adapter
  queue was the other half) — completions, not lane LAG, remain the
  enablement metric, now with a measured reason.
- **The adapter tier is the third serialisation layer** (sequential by
  design, one effective consumer per adapter). P1 does not touch it. Phase 5
  material, possibly its own workstream; at minimum the await timeout (3 min,
  step-config) vs adapter-queue-depth interaction needs an owner-visible
  decision — a longer timeout or a smarter retry (re-use the outstanding
  request rather than re-queue at the back) each has different failure
  economics.
- **This is very likely the 029 lane's ambient `spawn_dispatch`-timeout
  mechanism** (79 in 29 h): nothing here is specific to `git_commit` — any
  await whose callee+response path queues past 3 minutes enters the same
  treadmill. Contributed to bugs_open/029 (their bug, their queue).
- The wedged specimen from the first burst (`6c4a0bdf`, `check_skipped`,
  frozen 09:46:30) remains unexplained by this mechanism — it is a genuinely
  parked EXECUTING_STEP row, still the watchdog's case, still in place.

Cost note: two bursts + one sequential control ≈ 15 cheap re-renders on the
test site, no LLM steps. The vonc.com pages were left re-rendered; the four
FAILED-at-deploy runs from the first burst may have deployed anyway on late
responses (unchecked — test site).

---

## 2026-07-28 ~11:10 — CS-2 APPROVED (round 3), DEPLOYED, ENABLED — the pool WORKS in production. And the response-replay outage is measured: ~23 min of response deafness per pod restart.

**CS-2 verdict: APPROVED** on round 3 (corr `9f0499b9…`, "approved with 3
advisory objection(s) — none high-severity"). The three advisories, answered
here: (1) *why two new tables rather than extending
processed_messages/site_work_items* — `site_work_items` is owner-visible
domain work with its own fragile dedup-index↔Go-list contract and a different
lifecycle; `processed_messages` is a dedupe ledger keyed by message_id whose
two-phase contract 003 paid for; the intake table needs raw payload BYTEA,
kafka coordinates, and its own retention clock. Overloading either would be
the "one column, three jobs" failure (bugs 108) — the shape was reused, the
storage deliberately not. (2) *the core-edit signal must be logged, not waved
past* — logged: the commit hook's architecture signal fired on `28b1a0305`
and this line is the record. (3) *the seats' Schema section cannot see
`orchestration_states`/`doc_notes`, so those checks were still author-quoted*
— true and noted for future submissions: put checkable claims in
council-visible tables where possible; the artifact stays in `doc_notes`
regardless (id `701bce70…`) because operators CAN see it.

**Rollout, in order, all gated on a quiet lane (zero review/gate steps both
times):** image v1.0.1184 built from committed HEAD, pushed, binary-verified
(all six discriminating strings + positive control), rolled 10:54:56Z; dark
check passed (env unset → zero `INTAKE_` lines); flag flipped 10:56:37Z
(`CHASSIS_INTAKE_MODE=worker_pool`, `CHASSIS_DB_MAX_OPEN_CONNS=12`).
Pod-grep on the RUNNING pod (not git, not the tag): 1/1/1/1 on
`CLAIM_RECOVERY_STALENESS_HELD` / `INTAKE_MODE_ACTIVE` /
`INTAKE_BACKPRESSURE` / `CHASSIS_DB_MAX_OPEN_CONNS`. **CS-1 is therefore
LIVE** (unconditional code); its induced-branch tests are still owed and
queued behind the replay (below).

**The pool works.** Ambient scheduled traffic flowed immediately: intake rows
consumed→persisted→claimed→executed→done with claims released (3 events in
the first minutes, 14 by 11:08, 0 stuck). Sequential control (`judge`
re-render, corr `a6a66e1e…`): **received 11:00:37 → claimed 11:00:38 →
event done 11:00:40** — one-second pickup, two-second execution, exactly the
thin-ingest shape. `[The orchestration itself then parked at deploy_page —
NOT a pool defect; see below.]`

**The response-replay outage, measured.** The fresh pod's per-pod response
group (`StartOffset: FirstOffset`) replays the ENTIRE
`system.agent.generic.responses` history on every start: measured LAG 5,370
of 12,280 thirteen minutes after pod start, drain ~530/min ⇒ **~23 minutes
per restart during which the response lane processes history and is deaf to
fresh responses** — and the window grows with the topic. Consequences:
- The control (and any await) inside that window rides the treadmill: its
  response arrives while the lane is still in yesterday.
- **This is very likely most of what "the ~300s post-roll degraded window"
  (now observed ~20 min, bugs_open/029) actually is** — not just spawn
  drops: every await whose response lands in the replay window times out.
  Numbers match: 029's 07-27 losses "start 13 min after a roll, run 20 min".
  Contributed to 029.
- Phase 3 (responses through the pool) + the seed-to-latest option kill
  this class. Building the env-gated seed-to-latest now
  (`CHASSIS_RESPONSES_START_AT=latest`, default unchanged = today's
  FirstOffset), dark; **the flip waits for the owner's Phase 3→4 consent as
  the plan records** (blind window = pod-restart seconds, covered by F2
  re-send + at-least-once, per PLAN §13).

Decisive burst test deferred ~15 min until the replay catches up — running it
against a deaf response lane would measure the replay, not the pool.

> **CORRECTED ~11:20, by a closed-window measurement — my "~23 minutes" was
> 5× optimistic.** The early drain (~530/min) was the shallow part; a proper
> two-point window (124 s, position 7,145→7,248) gives **49 messages/min**,
> because ancient responses whose `awaited_requests` rows are long purged
> each burn the ~1.5 s not-found retry loop in
> `processResponseClaimWithRetry`. Remaining lag 5,037 at 11:16 ⇒ **~103
> further minutes**; total response deafness per pod restart at today's
> history ≈ **2–3 hours**, not 23 minutes. The sequential control FAILED at
> 11:12:45 (deploy_page, 4×3 min) exactly as this predicts — the pool's
> request side did its job in 3 s; the response side was in yesterday.
> Fleet impact while the window runs: every spawned handler's response rides
> `system.agent.generic.responses`, so build work fleet-wide treadmills
> until ~13:00 unless CS-3a flips. Operational call recorded below.

---

## 2026-07-28 ~11:30 — CS-3a APPROVED, live, VERIFIED (LAG 0); the decisive burst: 0/5 this morning → 4/5 in 47 seconds now; one wrong call logged on the way

**The wrong call first** (full entry in WRONG_CALLS.md): I flipped
`CHASSIS_RESPONSES_START_AT=latest` while the pod still ran v1.0.1184 — an
image built before the flag's code existed. Unknown env keys are silently
ignored; the "fix" was a no-op that cost one more restart and restarted the
replay from zero. The pre-flip pod-grep gate I had written into the audit
artifact that same hour would have caught it; I had applied it to every flag
except the one I'd just invented. Corrected by building v1.0.1186,
**verifying the symbol in the image BEFORE rolling**, then rolling on a quiet
lane.

**CS-3a verdict: APPROVED** (corr `f4e425dc…`, "approved with 3 advisory
objection(s) — none high-severity"). Live and verified on pod
`…-zfc2h` (started 11:23:47Z): `RESPONSES_START_AT_LATEST` logged, response
group born at **pos 12,289 = log-end, LAG 0** — zero replay, response lane
live from the first second. The 2–3-hour-per-restart response outage class is
closed for the chassis.

**Sequential control on the full stack: COMPLETED end-to-end in 24.9 s**
(corr `e506b005…`), including the deploy round trip.

**THE DECISIVE BURST** (same five pages that went 0/5 twice this morning),
corr prefixes e6a0cf7a/3f264f03/3e22ad91/e6223786/6aedced7, published
11:25:44–46Z:
- Intake: four events **started within 0.9 s of each other** — four workers
  running four orchestrations simultaneously; the fifth queued 7 s for a free
  worker. Serial-by-construction is over; the claims table did the ordering.
- Outcomes: **4/5 COMPLETED, whole batch terminal in 47 s** (vs 12 minutes of
  retries to 0/5 this morning).
- The 1 failure is a NEW, smaller, honest class: GitHub **422 "Update is not
  a fast forward"** — five concurrent commits to gqls/sites raced the ref
  update and the loser was refused, fail-fast with a real error in seconds
  (no treadmill, no wedge). Contributed to `bugs_open/120` (same repo, same
  missing serialisation, one account): the worker pool makes concurrent
  same-repo commits routine, so the git-adapter wants per-repo serialisation
  or 422-retry-with-rebase. Until then, expect occasional prompt
  non-fast-forward failures on same-site concurrent deploys.

## 2026-07-28 afternoon — git-adapter serialisation (owner-ruled) LIVE and proven both ways; the council gate is DOWN on the fleet's LLM spend cap until 08-01

**Owner ruled** ("the git adapter should serialise same-site deploys… go
ahead"): implemented as re-base-and-retry at GitHub's own ref CAS
(`CommitToRepo`, commit `7dc876795`) — on a non-fast-forward the loser
re-reads the winner's head and rebuilds tree+commit on it, max 4 attempts;
blobs hoisted (content-addressed); only "fast forward" 422s retry, everything
else still fails loudly. Deployed v1.0.1187 to both replicas (rolled on the
owner's go-ahead with one in-flight git op accepted as an F2-recoverable
casualty), verified against the RUNNING processes via `/proc/1/exe` (the
adapter's binary lives at `/root/git-adapter`, unreadable by the container
user — `/proc/1/exe` is the pod-grep route on this image).

**Proof, both branches:**
- Happy path at burst scale: the 5-burst went **5/5 COMPLETED in 20.3 s**
  (14:27), then a 10-concurrent double-burst — **16/16 COMPLETED in the
  window, 0 failed** (vs 0/5 morning, 4/5 midday).
- The failing branch: no natural race fired across all 16 (zero
  REF_RACE_RETRY lines), so it is proven by an induced test instead — a fake
  GitHub returns the verbatim 422 on the first ref PATCH while moving the
  head; the test asserts head re-read, tree+commit rebuilt on the winner's
  base, second PATCH succeeds, blobs created once (`1602dcd95`). Live firing
  will log `REF_RACE_RETRY` when it happens.

**Council: submission corr `bf2bef0a…` died `complete_invalid` at the FIRST
seat — not on the merits: the Anthropic usage cap is exhausted fleet-wide
("You will regain access on 2026-08-01 at 00:00 UTC"). The advisory gate is
therefore DOWN until 08-01; no resubmission is possible before then.** The
deploy stands on the owner's explicit direction; resubmit for the record
after 08-01 if anyone wants the trailer. (Corollary: no further council
submissions this session; the Phase 3 flip needs none — `worker_pool_all`
was reviewed inside CS-2's approved plan.)

**Two incidental discoveries recorded:**
- A stale-orchestration **reaper exists**: the morning's wedged specimen
  `6c4a0bdf` was failed at 13:46:56 with "reaper: stale EXECUTING_STEP for
  >4h". Wedge bookkeeping heals at 4 h — detection is still the watchdog's
  job (4 h is an outage, not an alert), but 029's fix-candidate 1 exists in
  some form.
- The shared tree's branch moved mid-afternoon: `086_experience_loop` →
  `087_towards_multiple_domains` (another session; 087 is a strict superset —
  verified all this workstream's commits are ancestors of the new HEAD).

## 2026-07-28 late afternoon — Phase 3 + Phase 4 LIVE; CS-1's induced pair fired; Candidate 4 completed; CS-2d found-and-fixed the orphaned-running hole; both new councils APPROVED

**Phase 3** (`worker_pool_all`): flipped after the vacuous-marker check (the
`worker_pool_all` const does NOT grep in any binary — Go prefix-merges it;
the zap literals are the real discriminators; control: my own 1186 showed 0
while running worker_pool). Verified: `responses_via_pool:true`,
response-kind intake events flowing, control COMPLETED.

**CS-1 induced pair: BOTH branches fired live** via awaited_requests fixtures
+ crafted kcat responses — 14:58:20 `CLAIM_RECOVERY_STALENESS_HELD` (fresh
claim held; end-state still owned by `induced-live-holder`), 14:58:33
`CLAIM_RECOVERY` (stale claim reset and re-claimed by the real pod).
Fixtures deleted after. Discharged AHEAD of its Phase-4 trigger.

**Candidate 4 wrapper: COMPLETED on its first post-fix attempt** (corr
`dd30cb5c…`) — spawn fired, child booted, `call_council` received the
response. Same path failed "timed out after 3 retries" on 07-27 and ~half of
history. Contributed to 096; the 097 default flip stays their owner's call
after more runs. (The spend cap made the test free — the child's seats
failed instantly; the cap lifted ~15:05 per the owner, and councils work
again.)

**Phase 4: replicas=2, owner-consented, LIVE.** The intake layer changed the
maths: claims CAS coordinates workers ACROSS pods, and the intake UNIQUE
(topic,partition,offset) dedupes the response broadcast between per-pod
groups — so scaling was config-only; CS-3's shared group is now an
efficiency refinement, not a correctness prerequisite. Two-pod burst: **5/5
in 20.6 s**. Observation, not a defect: at low volume pod 1 wins every claim
(its ingest→claim latency beats pod 2's 750 ms poll grid); distribution
appears under sustained load — and the CS-2d recovery below DID execute on
pod 2 while pod 1 worked, proving cross-pod claiming works when work is
visible.

**CS-2d — the pool's first operational defect, found live 90 min after
enablement, fixed same hour, council-APPROVED (round 4 under `9f0499b9…`).**
Two intake events (211/212, requests from 13:53) sat `running` under a dead
pod's expired claims forever: the pending-only CandidateKeys never surfaced
a key with no pending siblings, so the takeover reset was unreachable
precisely for the case it existed to serve. One-line widen to
`status IN ('pending','running')` (live holders stay excluded by the lease
NOT EXISTS); tripwire pins both predicates together. **Recovery verified on
v1.0.1190: both orphans re-ran at 16:35:01, attempts=2, done — 2.5 h late
instead of never — with exactly 2 `INTAKE_TAKEOVER_RESET` lines on pod
sqdzd.** The 10-burst that exposed it: 3 of 10 publishes were the kcat drop
trap (zero intake rows — instrument, not platform); all 7 that arrived
COMPLETED.

**Git-adapter re-review: APPROVED** (corr `bf2bef0a…`, 15:13, after the cap
lifted). Trailer recorded in this commit series.

**Owed and explicitly deferred, with triggers:**
- CS-1's induced live tests (duplicate response → `STALENESS_HELD`; stale
  claim → `CLAIM_RECOVERY` still fires). The guard is live and
  tripwire-tested; the induced pair MUST run **before Phase 4 sets
  replicas>1**, when the guard becomes load-bearing across pods.
- Phase 3 enablement (`worker_pool_all`) + the Candidate 4 re-test — next
  session's opener; the response-side code is already in the live binary.
- The two parked wedge specimens (`6c4a0bdf` check_skipped, and CS-1's
  first council run at review_editquality) stay in place for the
  dispatch_queue_serialisation watchdog work.

Instrument gotcha for whoever repeats this: `oufe/TRIGGER_rerender_page.sh`
names its kcat pod `kcat-rerender-$(date +%s)` — seconds granularity, so
PARALLEL invocations collide with "AlreadyExists" (3 of 5 lost that way on the
first attempt). The burst script with per-page+$RANDOM names is in this
session's scratchpad (burst5.sh); fold it into the RUNBOOK if the cross-site
check gets run.

Residue: 4 vonc.com pages may be re-rendered but undeployed (deploy_page timed
out; trust the artefact, not the status — not checked, test site). The wedged
row 6c4a0bdf sits in `EXECUTING_STEP` indefinitely; left in place deliberately
as a live specimen for the watchdog work (dispatch_queue_serialisation), noted
here so nobody diagnoses it as fresh.

---

## 2026-07-28 midday — CS-1/CS-2 built+submitted; the roll landmine fired on our own council; pre-enablement audit discharged

**Build state.** CS-1 (staleness guard) committed `afbd005f9`; CS-2 (intake +
pool, dark) committed `28b1a0305` with migration 249 applied+recorded; CS-2b
(`924df47d0`) makes the chassis DB pool cap env-configurable — found in the
audit below: `SetMaxOpenConns(4)` equals the default worker count before the
consume loops, heartbeats, response path and retry driver are counted, so
enablement must set `CHASSIS_DB_MAX_OPEN_CONNS` (≥12 suggested) alongside the
flag; every other agent keeps 4.

**The "a roll costs an in-flight council" landmine fired on CS-1's own
review run** — second specimen after 096's f849afaf. The run was consumed at
09:54:32Z, the roll (v1.0.1182, another thread's) landed at 09:55:02Z, and the
row froze at `review_editquality` / `EXECUTING_STEP`, `updated_at` 09:54:33,
parked for good. The council LANE is fine (CS-2's run, submitted post-roll,
progressed normally). Resubmitted under the same trail
(`RESUBMIT_CORR=a45f59af…`, new run `12a378af…`), queued behind CS-2's run.
Verdicts pending for both: CS-1 corr `a45f59af…`, CS-2 corr `9f0499b9…`.

**CS-1 verdict: APPROVED** (rerun `12a378af…` under corr `a45f59af…`,
"approved with 3 advisory objection(s) — none high-severity"). The three
advisories, answered: (1) land the test-compile repair separately from the fix
— it already was (`2da1c5f50` + `d0cda1e39` vs `afbd005f9`); (2) confirm no
out-of-repo caller of the deleted `ResetAwaitedRequestForRetry` — confirmed by
`go build ./cmd/...`: every binary builds except `cmd/reasoningset`, which is
another session's unrelated in-flight WIP; (3) pod-grep at roll time, not
commits-and-green-tests — agreed and already in the verification plan: grep
the running pod for `CLAIM_RECOVERY_STALENESS_HELD` (a string the change
CREATED) plus a positive control, then induce both branches. CS-2 round 2
resubmitted under corr `9f0499b9…` (run `e85ae92f…`) answering the guardian
point-by-point: CS-1 approved + structural co-shipping, audit complete (below),
backpressure cap (CS-2c `31f2fc1e2`), DB pool knob (CS-2b `924df47d0`).

**Pre-enablement mutable-state audit: DISCHARGED.** `platform/orchestration`
(root) + `platform/agentbase` package-level vars, every one read-only after
init: three error sentinels (coordinator.go:48), one compiled regexp
(loop_error_handler.go:26), one read-only map (agent_error_log.go:125, only
read at :132), one read-only slice (agent.go requiredIncomingHeaders, only
ranged), one uuid namespace (intake.go, new, never written). Combined with the
029 lane's earlier pass over `platform/orchestration/actions` (318
declarations, one benign startup-only write, setter has no callers), the
worker pool introduces no package-level shared-state hazard. Remaining
cross-worker safety = the DB claims (ClaimAwaitedRequest, the intake CAS),
`UpdateStateWithVersion`, component locks, and CS-1's guard.

---

## CONTRIB 2026-09-03 from the `bugfix_329_takeover_claim` lane — your cross-worker safety inventory has one more entry, and it was a GAP until today

Contributed rather than forked, per CLAUDE.md, and because the owner ruling of 2026-07-29 §3 says a
shared mechanism's other consumers must be **told**, not merely measured. **Nothing here asks anything
of this lane** — it is a fact about the seam you inventoried.

This file's pre-enablement audit ends with:

> *"Remaining cross-worker safety = the DB claims (ClaimAwaitedRequest, the intake CAS),
> `UpdateStateWithVersion`, component locks, and CS-1's guard."*

**There was a hole in that list.** `SagaCoordinator.handleOrchestrationStatus`'s two stale-takeover
arms (`coordinator.go` `case StatusExecutingStep` / `case StatusRunning`) decided a row was abandoned
on a **5-minute clock** and resumed it with **nothing claiming it** — `bugs_open/329`. Every write on
that path *was* version-CAS'd (⚠ contrary to what `bugs_open/329` and `bugs_closed/294` both say:
`UpdateState` **is** `UpdateStateWithVersion`, `state.go:883-885`), but the CAS answered *"has this
row changed since I read it?"* while the arm had decided on *"is this row stale?"* — a **check-then-act
across two reads**. Two takers seconds apart both won, each CAS-ing against the version it had just
read.

**Fixed and LIVE on `v1.0.1359`** (~2026-09-03 13:28Z, probed at both pod binaries with a control):
`StateRepository.ClaimStaleOrchestration` re-judges staleness **inside** the version CAS, so the
write is the claim. **Add to the inventory: "the coordinator's stale-takeover claim
(`ClaimStaleOrchestration`, WFA-025)".**

### Three things specific to this lane's concerns

1. **Your intake claim was masking it, and only on statically-deployed chassis pods.**
   `intakeSerialisationKey` keys requests on the orchestration_id and `ClaimSerialisationKey` is a
   shared-table CAS, so on `agent-chassis` two pods generally could not both be inside the arms for
   one orchestration. But `intake.go` refuses the mode when `a.spawned` — and `processing_node` over
   14 d `[MEASURED 2026-09-03]` shows **spawned Job pods are the MAJORITY of orchestration drivers**
   (`agent-chassis` 3,332 against 4,900+ from `agent-page-rerender` 2,215, `agent-build-dispatch-loop`
   660, `agent-page-build-handler` 412 and 11 more families, every one absent from `kubectl get deploy`).
   So the guard covered under 40% of the drivers.
2. **A finding on YOUR constants that this lane owns and is not attributing to you.**
   `intakeLeaseDefault = 180 s` (`intake_workers.go:43`) against `StuckOrchestrationTimeout = 300 s`.
   **180 < 300**, so a serialisation key can change hands *before* the row is old enough to look
   stuck — the handover and takeover windows are adjacent, not exclusive. And `drainKey` only tests
   `claimLost` **between events** while `processMessage` takes no context (your own header says so),
   so after a handover the old holder finishes the event it is inside while the new holder starts the
   next one on the same orchestration, bounded by the heartbeat period (`lease/3` = 60 s). The
   `dispatch_throughput` lane was offered this and declined it on the correct ground that they have
   never measured it, so it lives in `bugs_open/329` §5 attributed to **nobody**. **[UNVERIFIED]**
   whether the ordering was chosen or defaulted — if this lane knows, that would settle it.
3. **The residual is now `bugs_open/461`, OPEN and UNOWNED, and it is architecture-scope territory
   you may care about.** A live driver holds **nothing**, so the 300 s clock can judge a
   correctly-working orchestration dead (`defaultLocalActionTimeout` is **7200 s** and nothing
   refreshes `last_activity` during a local action — a 24× gap). Post-329 the bound is **2**
   concurrent actors, down from unbounded. The fix candidate is a **driver heartbeat**, which would
   be a *third* guarded mechanism on `orchestration_states` — flagged as needing an RFC, not a bug
   patch. ⚠ Your `intake_workers.go` already solved this shape one layer up, and **its limitation is
   the instructive part**: the heartbeat keeps a live-but-stuck holder's claim alive but can only
   check between events, so a coordinator heartbeat inherits that limit. Whoever takes 461 should
   read your implementation first.

Full account: `docs024_key_docs_latest/bugfix_329_takeover_claim/HANDOFF_2026-09-03_continue_here.md`.
