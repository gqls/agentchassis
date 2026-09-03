# RUNBOOK — bugfix 329

Commands that were hard to get right, with the gotcha attached. Change them HERE, not in scrollback.

## Is the race physically possible right now?

```bash
kubectl -n ai-persona-system get deploy agent-chassis -o wide
```
⚠ **Replica count is the precondition, and it is not in the bug file.** 2/2 as of 2026-09-03. A
single-replica reading would NOT make the bug invalid — the in-pod worker pool and every other
agent deployment still apply — so do not close it on that evidence alone.

## Which services actually carry the guard that masks this?

```bash
kubectl -n ai-persona-system get deploy -o json | python3 -c "
import json,sys
d=json.load(sys.stdin)
for it in d['items']:
    n=it['metadata']['name']; c=it['spec']['template']['spec']['containers'][0]
    env={e.get('name'):e.get('value') for e in c.get('env',[])}
    print(f\"{n:35s} replicas={it['spec'].get('replicas')} intake_mode={env.get('CHASSIS_INTAKE_MODE')}\")
"
```
⚠ **`CHASSIS_INTAKE_MODE` is set on `agent-chassis` ONLY.** Every other agent runs the same
`SagaCoordinator` (constructed in `platform/agentbase/agent.go`) with **no** intake claim. Reading
only the chassis and generalising is the mistake this command exists to prevent.

## The three constants whose ORDERING is the whole argument

```bash
grep -n "StuckOrchestrationTimeout = " platform/orchestration/coordinator.go   # 300s
grep -n "intakeLeaseDefault\s*=" platform/agentbase/intake_workers.go          # 180s
grep -n "intakeWorkersDefault\s*=" platform/agentbase/intake_workers.go        # 4 per pod
```
⚠ `180 < 300` is the point: a serialisation key can change hands **before** the orchestration row is
old enough to look stuck. Quote the ordering, never one number alone.

## Is anyone else mid-edit in the files I am about to touch?

```bash
git status --porcelain platform/orchestration/coordinator.go platform/orchestration/state.go
```
⚠ Run this **again** immediately before committing, not just at session start — the tree is shared
and the snapshot goes stale within minutes. A pathspec commit still takes a **same-file** passenger.

## Do NOT verify this fix by reading the function

The bug file says so explicitly, and the reason is on the record: that lane got the same class of
thing wrong three times in two days by reading a body instead of its callers.

⚠ **And do not verify it with a test that the intake claim could have passed for you.** On
`agent-chassis` the claim serialises per orchestration fleet-wide, so a "two callers, one proceeds"
test can pass with the fix reverted. Isolate the arm under test — see the PLAN's test section.

## Diagnosis loop: do not bother on this one

```bash
ls -l platform/orchestration/coordinator.go   # 199,136 bytes
```
⚠ `LANDMINES.md`: a `090` run on a symbol in a file over ~60KB returns bundles and **no verdict**,
and that looks exactly like a run still in progress. Both files here are over it. State the
substitute instead (owner ruling 2026-07-31) — see NOTES (e).
