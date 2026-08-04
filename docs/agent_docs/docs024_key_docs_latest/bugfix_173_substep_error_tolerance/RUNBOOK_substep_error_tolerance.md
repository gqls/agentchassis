# RUNBOOK — `bugs_open/173` per-substep `continue_on_error`

Commands that were hard to get right, with the gotcha attached. Change them **here**, not in
your scrollback.

---

## R1 — Is anyone else working this bug? (the check that actually discriminates)

**Do NOT count bare mentions of the bug number.** Every session lists `bugs_open/`, so all 54
numbers appear in most transcripts and every candidate scores "mentioned in 4–16 sessions".
That measurement cannot return a low number for any open bug — see `WRONG_CALLS.md`
2026-08-04.

Count sessions that actually **opened** the file:

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
find . -maxdepth 1 -name '*.jsonl' -mmin -300 | while read f; do
  grep -oE '"file_path":"[^"]*bugs_open/[0-9]{3}[^"]*"' "$f" | grep -oE 'bugs_open/[0-9]{3}'
done | sort | uniq -c | sort -rn
```

**Gotchas.** `-mmin -300` is the live-session window; widen it and you pick up finished
threads. Cross-check the top entries against `git log --since='36 hours ago'` — if the busy
numbers there match the busy numbers here, the signal is corroborated.

Then, and only as a second opinion, `./scripts/who-owns.py <n>`. **Its verdict line alone is
not an answer**: it reads commits, so the lane that *filed* a bug for someone else looks
exactly like the lane fixing it. Read the named lane's handoff to see whether it is closed.

## R2 — Census every loop step on the live fleet, with its substeps

**The gotcha that costs a wrong answer** (recorded in `173` itself and re-confirmed here): a
loop's body lives in `config.sub_workflow.steps` for most loops and `config.substeps` for a
few. Counting off one key alone reports "no substeps" for the other set and reads exactly
like "these loops are empty, nothing to worry about". `COALESCE` both, in
`loop_actions.go`'s precedence order (`substeps` wins).

```sql
WITH loops AS (
  SELECT a.type, s.key AS loop_step,
         COALESCE(s.value->'config'->>'continue_on_error','(unset)') AS loop_coe,
         COALESCE(s.value->'config'->'substeps', s.value->'config'->'sub_workflow'->'steps') AS body
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.value->>'action' = 'loop'
)
SELECT loop_coe, count(*) AS loops, count(DISTINCT type) AS agents,
       sum((SELECT count(*) FROM jsonb_object_keys(body))) AS substeps
FROM loops GROUP BY 1 ORDER BY 1;
```

2026-08-04: `false` 1/1/7 · `true` 9/9/26 · `(unset)` 8/7/46. **Total 18 loops, 79 substeps.**

## R3 — Does any live substep already declare the new key? (with its positive control)

The answer wanted is 0, which is exactly why the control is mandatory — a `0` from a
predicate that matches nothing looks identical to a `0` from a predicate that works.

```sql
-- the measurement
WITH loops AS ( /* as R2 */ )
SELECT type, loop_step, b.key AS substep
FROM loops, LATERAL jsonb_each(body) b
WHERE b.value->'config' ? 'continue_on_error';
```

**Run R2 in the same session as the control.** R2 returning 18 loops / 79 substeps proves the
CTE finds loops and reaches substep configs, so a non-zero answer was reachable. Report the
pair, never the 0 alone.

**Gotcha:** `?` is jsonb key-existence and psql/`kubectl exec` pass it through fine, but it is
a placeholder in most client libraries — use `jsonb_exists(b.value->'config','continue_on_error')`
if you port this query into Go or Python.

## R4 — Which live loops carry a substep that can now refuse?

The consumers to notify. Keep the action list in step with whatever has a completeness floor.

```sql
WITH loops AS ( /* as R2 */ )
SELECT type, loop_step, loop_coe, b.key AS substep, b.value->>'action' AS substep_action
FROM loops, LATERAL jsonb_each(body) b
WHERE b.value->>'action' IN ('save_page_sections','extract_and_sync_links',
                             'populate_nav_tables','index_code_symbols')
ORDER BY 1,2,4;
```

2026-08-04: three rows — `pageflow-builder`, `page-rebuild`, `site-work-orchestrator`, all
`build_*_loop` → `save_sections` → `save_page_sections`, all loop-level `(unset)`.

**Gotcha, and it is the reason to run this rather than copy the table out of the bug file:**
`173` names a fourth (`multipage-website-builder.generate_pages_loop`). That agent is
**deleted and inactive** — confirm an agent is live before naming it as a consumer:

```sql
SELECT type, is_active, COALESCE(is_snapshot,false) AS snap, deleted_at IS NOT NULL AS deleted
FROM agent_definitions WHERE type='<agent>' ORDER BY is_active DESC, deleted_at NULLS FIRST;
```

## R5 — Test the change, and prove the tests can fail

```bash
cd /home/ant/projects/agentchassis
gofmt -l platform/orchestration/          # must print nothing
go build ./...
go test ./platform/orchestration/... ./pkg/models/...
```

**A green suite proves nothing on its own here** — the guard is inert until induced. Mutate
the production line back to the unconditional
`injectedStep.Config["continue_on_error"] = continueOnError`, re-run, and confirm the
override tests **fail**; then restore. A test that passes against both the fixed and the
broken code is not a test (`mutate-the-code-to-prove-the-guard`).

**Gotcha:** if a mutation *passes*, suspect a guard in series before concluding the test is
weak — the loop-level value may coincide with the substep's in that fixture. Make the two
differ.

## R6 — Verify at the pod after a roll (NOT before)

`make build-*` builds from committed HEAD; `make release` is whole-fleet and owner-run, so a
session does not roll this itself. After a roll:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# ✓ discriminating — this change only
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'resolveSubstepContinueOnError'"
# ✓ NEGATIVE control — a string the change REMOVED; expect 0
# ✓ POSITIVE control — a string live since before this change; expect >=1, proves the probe works
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'continue_on_error is true for this loop iteration step'"
```

**Gotchas.** Run it on **every replica**, not one (`logs deploy/X` and a single-pod exec both
read one of N). A roll is not evidence your fix shipped — the image may predate your commit
(`bugs_open/153`). And a `strings | grep` returns 0 for a marker spanning a non-ASCII byte, so
keep probe strings ASCII — both markers above are.

## R7 — Docs obligations that are easy to forget

```bash
./scripts/landmines-sync.py --apply     # after appending to LANDMINES.md; --check exits 1 on drift
python3 scripts/pattern-check.py        # gofmt/pattern gate before committing
```

Concept register: new entry in `register/workflow-authoring.md`, **plus** the index row in
`register/000_concept_index.md` — the index is a separate file and a new entry without its row
is invisible to the count. Both in the **same commit as the code** (ordering-exemption
condition 2).

## R8 — Inducing a loop-substep failure on the live fleet (the whole procedure, with both traps)

Used to close `173`. Reusable for any loop error-routing question.

**Seed** `SEED_2026-08-04_induction_agents.sql` (in this directory). Two throwaway agents whose
loop-level flag is the **opposite** of the substep's — that opposition is the design: a
regression makes each run produce the *other* run's outcome, so the pass cannot be incidental.

**Induced fault:** `SELECT 1/0` via `query_database`. Deterministic, read-only, cannot touch a
production row. **Check the fault can actually fire before dispatching** — a fault that no-ops
makes the tolerant branch read COMPLETED for the wrong reason:
`QueryDatabaseAction` returns `return nil, fmt.Errorf("query failed: %w", err)` (a real Go
error that fails the step), and `SELECT 1/0` really does raise `ERROR: division by zero`.

**Dispatch** — envelope mirrors `097`, topic `system.agent.generic.requests`:

```bash
CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid); echo "ORCH=$ORCH"
jq -nc --arg a "test-173-tolerant-substep" \
   '{action:"orchestrate",config:{agent_type:$a},input_data:{note:"induction"}}' \
| kubectl -n kafka run -i --rm "kcat$(date +%s%N | tail -c 4)" \
    --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H "correlation_id=$CORR" -H "request_id=$(cat /proc/sys/kernel/random/uuid)" \
    -H "message_id=$(cat /proc/sys/kernel/random/uuid)" \
    -H "orchestration_id=$ORCH" -H "orchestration_name=ind-$(date +%H%M%S)" \
    -H "step_name=start" -H "client_id=demo_client" -H "message_type=request" \
    -H "action=orchestrate" -H "from_agent_type=user" -H "from_agent_id=cli" \
    -H "responses_topic=system.agent.generic.responses"
```

### ⚠ TRAP 1 — a terminal step is `complete_workflow`, NOT `complete`

`{"action":"complete"}` is rejected with `step 'done' with action 'complete' requires a topic`
— `complete` is a step **name** (172 live agents use it) whose **action** is
`complete_workflow`. The orchestration dies at validation and **no `orchestration_states` row
is ever created**, which looks exactly like queue latency.

### ⚠ TRAP 2 — grep the chassis log BEFORE theorising about the queue

An absent row has two very different causes and one of them is invisible from the DB. **This
is the first thing to run, not the last:**

```bash
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
  kubectl -n ai-persona-system logs "$p" --since=25m | grep "<your orchestration_name>" | head -3
done
```

- **not in the log** → genuinely queued or dropped; CLAUDE.md's "it is latency, do not retry"
  applies (publish→start measured at 29 min).
- **in the log** → latency is refuted outright. Now get the reason:

```bash
kubectl -n ai-persona-system logs <pod> --since=25m \
 | grep '"correlation_id":"<corr>"' | grep '"level":"error"' \
 | sed -E 's/.*"msg":"([^"]*)".*/\1/'
# then pretty-print the failing line's "error" field with python3 -c 'import sys,json;...'
```

**And never redirect kcat's output** (`>/dev/null`) — `kcat -P` can send nothing at exit 0, so
suppressing it makes that landmine unfalsifiable for you.

### Reading the result — status is NOT the proof

```sql
SELECT orchestration_name, status, current_step,
       collected_data ? 'run_loop_iter_0_error'             AS it0_err,
       collected_data #>> '{run_loop_iter_0_error,skipped}' AS it0_skipped,
       collected_data #>> '{run_loop_error_count}'          AS err_count
FROM orchestration_states WHERE orchestration_name LIKE 'ind%' ORDER BY updated_at;
```

The tolerant run reads COMPLETED whether the fault fired or not — **the discriminating
evidence is the `iter_N_error` records with `skipped=true`**, plus the log line
`Skipping failed loop iteration, advancing to next  step=<loop>_iter_N_<substep>`.

### ⚠ CLEANUP — owed whatever the outcome

```sql
DELETE FROM agent_definitions WHERE type IN ('test-173-tolerant-substep','test-173-strict-substep');
SELECT count(*) FROM agent_definitions WHERE type LIKE 'test-173-%';   -- must be 0
```
