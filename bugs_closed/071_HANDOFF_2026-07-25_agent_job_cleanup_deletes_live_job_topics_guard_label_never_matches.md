# BUG 071 — agent-job-cleanup deletes LIVE job.* topics: its "any spawned pods running?" guard has never matched a pod

**Filed:** 2026-07-25 · gauntlet_dead_cta / feature-builder B4 shakeout ·
**CLOSED 2026-07-26 ~13:36Z — every shipped fix verified live.** Guard
live-proven 07-25; tombstone live-proven BOTH branches 07-26 (idle tick
deleted 2-of-2 on the second consecutive idle observation); pod-template
label live-proven 07-26 13:35Z: Job `agent-code-indexer-9cc5be96` (spawned by
chassis v1.0.1165) carries `spawned-by: orchestrator` in its pod template and
its running pod is the FIRST in the platform's history to match
`kubectl get pods -l spawned-by`. Council trail `d0fcf7ef` APPROVED (round 3).
Non-blocking notes carried below: remote-job-spawner path label is
[INFERRED, same v1.0.1165 build] until a remote spawn is observed
(check: that Job's `.spec.template.metadata.labels.spawned-by` =
`remote-job-spawner`); dedicated-SA least-privilege refactor is recorded
hardening, not a reproduction path; 003-family re-attribution stays with
bugs_open/003.
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
  **PROVEN 2026-07-25 09:13:** run `af286d2c` (fired 08:37, PR opened 09:13)
  crossed the 08:50, 09:00 and 09:10 ticks — each tick's log took the keep
  branch ("Live spawned workload (jobs: 25/32/…) — keeping all job topics") —
  and finished with 8/8 awaits `processed`, 0 `expired`, all six stage_commits
  consumed. Same plan, same code path that died twice the day before →
  machine PR #3. The core mechanism is dead; the case stays OPEN only for the
  residuals (pod-template labels, topic-age gating).
- Negative control: with zero spawned jobs/pods, a tick still deletes orphan
  topics (the cleanup's actual job) — verified by the delete-all branch running
  on the 08:20 tick and topics re-listed at 88 by 08:34 (churn + zombies).

## Residuals progress (2026-07-25 pm, dedicated bugfix-071 session)

**Residual 1 — pod-template labels: COMMITTED `5540d203e`, INERT UNTIL IMAGE ROLL.**
`spawned-by: orchestrator` added to the pod-template labels map in
`spawn_actions.go` (~2739) and `spawned-by: remote-job-spawner` in
`cmd/remote-job-spawner/main.go` (~400); both build. The two divergent values
are KEPT deliberately and documented in-code: a full-repo audit shows every
live `spawned-by` selector targets **Jobs** (cleanup steps 2/3/5,
`deploy-agents.sh:21`, `test-agent-spawn.sh:52`, `test-workflow.sh:82`), so
pods gaining the label changes no existing consumer — it only makes pod-level
selection possible, which is this residual's whole point. No third spawn path
exists (`diagnose_build_gate_action.go:143` creates Jobs with disjoint labels,
deliberately outside the guard). **No image roll this session** (bug-003's
parked F2/F3 ride any chassis build; that roll is the owner's call). Post-roll
check: `kubectl get pods -n ai-persona-system -l spawned-by` returns spawned
pods.

**Residual 2 — two-pass tombstone: SHIPPED & LIVE (commit `bc1f12718`,
applied, cronjob generation 6).** An idle tick no longer deletes all `job.*`
topics: deletion now requires a topic to be observed orphaned on two
CONSECUTIVE idle ticks (~20 min), protecting queued work with no Job yet.
State in new ConfigMap `agent-job-cleanup-state` (pre-created by manifest;
RBAC `create` can't be resourceName-scoped); busy ticks, broker-unobservable
ticks and zero-topic ticks wipe the state so "consecutive" stays strict; a
failed state read deletes nothing (fail-safe) but still stores. Verified:
- 6-scenario stubbed-`kubectl` sh harness against the script extracted from
  the applied yaml: idle tick 1 deletes 0 + tombstones all; idle tick 2
  deletes exactly the intersection + stores survivors; busy tick prints the
  keep line **byte-identical** to the guard fix's (this file's check above
  still holds) and wipes state; read-failure deletes nothing; empty listing
  yields a single-line 0; jobs-query failure takes the keep branch.
- Live busy-tick run (`cleanup-manual-071r2`, 10:02): 334 topics, keep branch,
  no state-write error. `can-i patch configmaps/agent-job-cleanup-state`
  flipped no→yes; `create configmaps` still no.
- ~~**Idle branch [UNVERIFIED live]** until an idle window occurs.~~
  > **VERIFIED LIVE 2026-07-26 ~13:10Z.** The fleet went idle (0 dynamic-agent
  > pods) and the scheduled tick `agent-job-cleanup-29751190` logged the full
  > protocol in production: `Found 2 job topics` → `No live spawned workload —
  > 2 of 2 job topics were already orphaned on the previous tick; deleting
  > those only` → `Deleted 2 of 2 tombstoned job topics` → `Tombstoned 0 job
  > topics for deletion on the next idle tick`. So the PREVIOUS idle tick
  > tombstoned them and this tick deleted them — two consecutive idle
  > observations, exactly as designed. Negative control holds: orphan cleanup
  > still happens, one tick later; `job.*` topic count now 0; state CM empty;
  > no ERROR lines. Both branches of the tombstone are now live-proven.

**New findings fixed en route (same commit):**
- **Cronjob step 1 has been silently Forbidden since inception**: the SA
  `ai-persona-app` had no `pods delete` verb, so `kubectl delete pods
  --field-selector=status.phase=Failed` failed every tick. Granted
  (namespace-wide — unscopable by resourceName) and **proven by induced
  fault**: pod `test-071-step1-failed` (busybox `false`, phase=Failed) was
  actually deleted by the next run — the first time that step ever worked.
- Latent `TOPIC_COUNT` bug: `grep -c ... || echo 0` yields the two-line string
  `0\n0` on no match (grep -c prints 0 AND exits 1); `[ -gt 0 ]` then errors
  and happened to fall through false. Fallback dropped.
- The old `DELETED` counter was incremented inside a piped-`while` subshell
  and never counted; replaced with word-split `for` loops (topic charset is
  alphanumeric/dot/hyphen, so splitting is safe).

**Council trail (advisory), corr `d0fcf7ef-f43b-4bbc-933f-ededfa4963a4`:**
round 1 REVISE — gating objection (prior_art_librarian): the absence claim
"kafka-topics.sh exposes no creation time" was asserted without evidence.
Fair, and worth recording the answer here because it licenses the whole
design: `--describe` output carries **no timestamp field**; the cluster is
**KRaft** (zero zookeeper pods → no znode-ctime route); 3 brokers at RF=1
with partitions spread across them (sampled topic: brokers 2 and 0) → a
single-broker `exec` cannot stat all topics' data dirs; offset timestamps are
message-age, not topic-age (empty topic → no signal). Round 2 resubmitted
with all checks attached — **but its run (orch `92d87be0`) never verdicted: it
wedged at `review_guardian` EXECUTING_STEP when the single-replica chassis was
restarted at 10:36:45Z (another session's v1.0.1159 roll), 22 min into the
step.** Council seats run in-process (zero `awaited_requests` rows), so the
step died with the pod and nothing resumes it — a first-hand live instance of
the 003-class fragility (F3 durable timers), observed from the submitter's
side this time. Round 3 (orch `8da7ea0e`, content identical to round 2)
resubmitted 13:53Z: **APPROVED** ("approved with 2 advisory objection(s) —
none high-severity"; 7 approve, 8 abstained via relevance gating). The two
advisory objections — namespace-wide `pods delete` on the shared SA, and the
idle branch being harness-only-verified — are both already carried as open
items in this file (dedicated-SA residual; idle-window observation). The
shipping commits (`bc1f12718`, `5540d203e`) predate the verdict
(advisory-alongside path), so they carry the corr in their messages but no
`Council-Reviewed:` trailer — the 098 report will list them as unreviewed;
this paragraph is the join.

**Build-timing note for residual 1:** v1.0.1159 (the 10:36:45Z roll) was
built ONE MINUTE before `5540d203e` landed (10:37:31Z — the commit stamp
reads 11:37+01:00, the BST/UTC trap), so the label change is NOT in it:
verified behaviourally — 18 dynamic-agent pods running, 0 match
`-l spawned-by`. It rides the next chassis/spawner build after `5540d203e`.
(A `strings`-on-binary pod-grep cannot verify this change at all: it adds no
new unique literal — comments don't compile and both label strings pre-exist
in the Job-labels map. The behavioural check above is THE post-roll check.)

> **UPDATE 2026-07-26 ~13:12Z:** BOTH images rolled to **v1.0.1165** at
> 12:06Z (chassis AND remote-job-spawner). The label change is in them
> [INFERRED from build ordering: every build since v1.0.1163 (07-25 pm)
> postdates `5540d203e`] but not yet behaviourally proven — the fleet is
> idle, so no spawned Job exists to inspect. Proof closes on the next spawn:
> `kubectl get jobs -n ai-persona-system -l spawned-by -o
> jsonpath='{range .items[*]}{.metadata.name}
> {.spec.template.metadata.labels.spawned-by}{"\n"}{end}'` must show the
> template label (works even after the pod is gone; Jobs live O(1h) via TTL).

**Remaining open here:**
1. Image roll carrying residual 1 (owner-gated via bug-003's parked roll).
2. Idle-window observation of the tombstone delete branch.
3. Dedicated ServiceAccount for the cleanup cronjob (the `pods delete` grant
   is namespace-wide on the shared `ai-persona-app` SA — bounded, since its
   existing jobs-delete cascades to pods anyway, but a scoped SA is cleaner).
4. The 003-family re-attribution hypothesis (unchanged, owned by 003).
