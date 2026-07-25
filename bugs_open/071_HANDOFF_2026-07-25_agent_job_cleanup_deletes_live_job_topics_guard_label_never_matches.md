# BUG 071 — agent-job-cleanup deletes LIVE job.* topics: its "any spawned pods running?" guard has never matched a pod

**Filed:** 2026-07-25 · gauntlet_dead_cta / feature-builder B4 shakeout · **OPEN (guard fixed & live; residuals + fleet questions remain)**
**Severity:** critical — every 10 minutes, the cleanup cronjob deleted every
`job.*` Kafka topic in the cluster, including request/response topics under
agents that were mid-run. Long-running spawned agents (feature-implementer,
fix-proposer class) were near-guaranteed to die; short ones died whenever a
response landed in the deletion window.

**Numbering note:** the fix commit `9dc99c61c` says `bugs_open/070` in its
message — 070 was taken by another session's file between my commit and this
filing. This case is **071**; resolve by slug.

## Symptom

Two consecutive feature-implementer runs (the first-ever fires on an approved
staged plan, corr `c379f7b7`) died at a `stage_commit` await:

- orch `2b1a154e` (07-24 16:27–16:36): s1–s3 committed, s4 await expired.
- orch `fbac5548` (07-24 19:58–20:03): s1–s2 committed, s2's response await
  expired — **after the commit itself succeeded on the branch**.

In both, git-adapter's log shows the success response **produced ~4s after the
request** to the correct `job.<corr8>-….responses` topic with correct headers
(run 2: request `81456c97` received 20:03:42.529, response produced
20:03:46.391). The awaited_requests row nevertheless expired
(`status='expired'`, `processing_pod` empty): produced, never consumed.

## Root cause (structurally certain, verified in code + live)

`deployments/kustomize/services/agent-job-cleanup/agent-job-cleanup-cronjob.yaml`
(schedule `*/10`) step 4 guarded its delete-all-topics branch with:

```sh
RUNNING_COUNT=$(kubectl get pods -n ai-persona-system \
  -l spawned-by=orchestrator --field-selector=status.phase=Running ... | wc -l)
```

But **no pod has ever carried `spawned-by`**. Both spawn paths put the label on
the *Job* object only, never the pod template:

- `platform/orchestration/actions/spawn_actions.go` (~2714): Job labels include
  `spawned-by: orchestrator`; pod template labels are only
  `app/agent-type/agent-id/client-id`.
- `cmd/remote-job-spawner/main.go:377-400`: Job labels include
  `spawned-by: remote-job-spawner` (a *different value* — a second, masked
  defect); pod template again has no `spawned-by`.

Verified live 2026-07-25: `kubectl get pods -l spawned-by` → zero rows while
~40 `agent-*` pods were running. No manifest in `deployments/` sets the label
on a pod either.

So `RUNNING_COUNT` was always 0, the "conservative cleanup" branch was **dead
code**, and every tick with `job.*` topics present deleted them all —
`bin/kafka-topics.sh --delete` per topic, sequentially (~80 topics ≈ 2–5 min of
loop time). A live agent's topic dies mid-loop; a response produced after the
deletion auto-recreates the topic, but the group's offsets died with the old
topic and no reader was on the new one — the message sits unread, the await
expires (F1 reaper marks it), and the orchestration hangs until an external
sweep fails it (fbac5548: FAILED at 00:05:01 by `app - 10.20.39.61`).

Captured instances:
- 07-25 08:20:03 tick (old code): `"No running spawned pods — deleting all 82
  job topics"` — log captured before the pod TTL'd; its "Remaining" tail is
  UNCAPTURED. `successfulJobsHistoryLimit: 1` + `ttlSecondsAfterFinished: 300`
  mean per-tick logs from 07-24 are unrecoverable.
- Run 2 timeline against the 20:00 tick: s1's response was consumed at
  20:00:02.089 — the tick started 20:00:03; the run's `job.3c645ff4-…` topics
  (digit prefix, early in the alphabetical delete order) died seconds later
  [INFERRED: exact per-topic delete instant; the deletion itself is certain —
  the topic was in the listing and the loop deletes every listed topic].
- Run 1 timeline against the 16:30 tick: awaits consumed 16:28:41 and
  **16:32:25**, then 16:36:05 expired — consistent with its topic dying midway
  through the 16:30 tick's multi-minute loop [INFERRED: loop position].

> **CORRECTED attribution:** run 1's death was previously recorded (bug 003
> sighting, NOTES, commit 442c4b48d) as caused by the 16:29:38Z chassis
> restart. Wrong — the restart preceded a *successful* consume at 16:32:25, so
> it did not kill the consumer. Temporal correlation, not cause. The 10-minute
> tick grid was the real clock in the room.

## Why the fleet mostly "worked" anyway

Most spawned work (page-rerender etc.) completes in well under 10 minutes and
whole runs often fit between ticks; kafka auto-create resurrects topics on the
next produce, so the system self-heals around the carnage. The losses
concentrate exactly where this workstream lives: multi-stage spawned agents
whose run MUST cross a tick, and any response produced into a deletion window.

**Hypothesis, not asserted [INFERRED]:** this mechanism plausibly accounts for
part of the bug-003 sighting family (silent spawn drops, lost responses, dead
awaits — e.g. the experience-planner drop corr `4d3d89fa`, council seat
`complete_invalid` infra deaths). Each prior sighting would need its own
timestamp-vs-tick check before being re-attributed. Contributed as a note to
`bugs_open/003`, whose workstream owns that question.

## Fix shipped (live) + residuals

**Shipped 2026-07-25, commit `9dc99c61c`, `kubectl apply`'d (generation 5),
config-only:** the guard now counts active spawned **Jobs** across both spawner
labels (`spawned-by in (orchestrator, remote-job-spawner)` — Jobs exist through
pod Pending, so this also covers the pre-start window) plus `app=dynamic-agent`
pods in Running/Pending/ContainerCreating, and **fails safe to keep-topics** if
the kubectl queries error. First run of the fixed script (manual job
`cleanup-manual-fix-test`, 08:34:37): `"Live spawned workload (jobs: 39; pods:
39) — keeping all job topics"` — the old guard would have deleted 88 topics
under those 39 agents at that moment.

No 090 diagnosis run was filed for the root cause: it was directly observed at
code level, refutation-checked live (zero pods carry the label), and the fix
was watched failing (08:20) then passing (08:34) — the self-evidencing
carve-out. The 003-family re-attribution above is exactly the part that is NOT
self-evidencing, hence marked hypothesis.

**Residuals / candidates:**
1. **Pod-template labels** (Go, both spawn paths): add `spawned-by` to the pod
   templates so pod-level selectors work at all; unify or document the two
   label values. Inert until image roll.
2. **Topic-age gating:** delete-all-when-idle still nukes topics of QUEUED work
   that has no Job yet (window is seconds normally, but dispatch backlogs
   exist). kafka-topics.sh exposes no creation time; a two-pass tombstone
   (delete only topics also listed in the *previous* idle tick, state in a
   configmap) would make deletion require ~20 min of continuous orphanhood.
3. **Structural:** bug-003 F2/F3 (durable awaits, DB-driven timers) make lost
   responses recoverable rather than fatal — this bug is one more producer of
   exactly that failure shape.

## How to verify

- Any tick's log while spawned work runs must say `Live spawned workload … —
  keeping all job topics` (old text: `No running spawned pods`).
- A feature-implementer run (8–25 min, guaranteed tick crossings) completes all
  stages without an expired `stage_commit` await: `SELECT status FROM
  awaited_requests WHERE orchestration_id='<run>'` shows no `expired` rows.
  First live trial fired 08:4x 07-25 on corr `c379f7b7`.
- Negative control: with zero spawned jobs/pods, a tick still deletes orphan
  topics (the cleanup's actual job) — verified by the delete-all branch running
  on the 08:20 tick and topics re-listed at 88 by 08:34 (churn + zombies).
