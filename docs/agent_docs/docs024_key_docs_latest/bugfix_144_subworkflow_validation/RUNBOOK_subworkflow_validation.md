# RUNBOOK — sub-workflow validation (bugs_open/144)

Every command here had to be got right once. The gotcha is attached to each.

---

## Export the live definitions (the input to everything else)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -tAc "
  SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow'))
  FROM agent_definitions
  WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active
    AND default_config ? 'workflow';" > /tmp/live_workflows.json

python3 -c "import json;d=json.load(open('/tmp/live_workflows.json'));print(len(d),'agents')"
```

**Gotcha: a truncated `kubectl exec` exits 0.** The `python3` line is not decoration —
it is the only thing standing between you and a "clean, small fleet" that is actually
a short read. `json.load` fails loudly on a truncated array.

## Replay the validator over the fleet (the measurement that made this safe)

```bash
SUBWF_LIVE_EXPORT=/tmp/live_workflows.json \
  go test ./platform/orchestration/actions/ \
  -run TestLiveDefinitionsPassSubWorkflowValidation -v
```

Reads: `agents decoded`, `top-level steps / sub-workflows / nested steps`, warnings by
message, then two distinct lists —

* `NEWLY REJECTED BY SUB-WORKFLOW VALIDATION` → **fails the test**. Attributable to
  your change: the same plan validates clean with its sub-workflows stripped.
* `ALREADY REJECTED BEFORE THIS CHANGE` → logged only. Rejected by the validator as it
  stands, with or without the recursion.

**Gotcha: the two-verdict control is the whole point.** The first run of this reported
"3 live definitions would be rejected" and every one of them was pre-existing. Reading
a verdict off the wording of an error message is how someone else's defect gets
charged to your patch — or worse, yours to theirs.

**Gotcha: it fails deliberately if it finds zero nested steps.** A broken export or a
broken traversal would otherwise pass loudest exactly when it had stopped looking.

## The config-key audit, now including nested steps

```bash
./scripts/audit-config-keys.sh            # human
./scripts/audit-config-keys.sh --json     # machine; carries pairs_top_level / pairs_nested / pairs_nested_only
```

**Gotcha: `go run` needs the repo root.** The script `cd`s to `REPO_ROOT` for the
binary; if you run the binary by hand from elsewhere you get
`go.mod file not found in current directory or any parent directory` and no output.

```bash
# by hand:
cd /home/ant/projects/agentchassis
go run ./cmd/config-key-audit --live-pairs < /tmp/live_workflows.json > /tmp/pairs.tsv
awk -F'\t' '{print $3}' /tmp/pairs.tsv | sort | uniq -c    # top vs nested
```

**Gotcha: the exit code is not a health signal.** The script exits 1 when there are
unknown keys, which there are (3, all `bugs_open/136`'s half-landed rename). That is
its designed behaviour, not a failure of the audit.

## Which live definitions carry action X — the query that started all this

```sql
-- WRONG (silently under-reports; misses every nested carrier)
SELECT type FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps') e(k,v)
WHERE v->>'action' = 'plan_sections' AND is_active;

-- CRUDE BUT COMPLETE — use as the cross-check
SELECT type FROM agent_definitions
WHERE default_config::text LIKE '%plan_sections%'
  AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active;
```

The first form is what produced the wrong claim that led to this bug being filed. The
`LIKE` is crude and over-matches — it is the one that catches a nested carrier.

## Verify it is LIVE on the pod (after a chassis roll)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "uses fan_out, which cannot work inside a sub-workflow"'   # want 1
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "Checking disconnected step"'                              # want 0
```

**Gotcha: the second grep is the load-bearing one.** It is a string this change
**deleted**, and a delete-marker cannot be satisfied by a stale binary that happens to
contain a similar phrase. The first grep alone would pass on any image built after
some *other* session added a similar string.

**Gotcha: `logs deploy/X` reads one pod of N.** Grep the binary in each replica, or
name the pod.

## Census queries used in the PLAN

```sql
-- step-config type census (the old audit's string-config guard: 0 such today)
SELECT jsonb_typeof(v->'config'), count(*)
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') e(k,v)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
GROUP BY 1;

-- is a nested action local? (the topic rule's whole cost)
--   grep the registry rather than guessing:
--   grep -n '"<action>": {' -A 12 platform/orchestration/actions/registry.go | grep IsLocal
```
