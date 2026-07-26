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
