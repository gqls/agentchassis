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
