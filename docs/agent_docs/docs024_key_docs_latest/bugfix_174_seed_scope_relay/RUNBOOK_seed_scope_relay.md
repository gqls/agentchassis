# RUNBOOK — bugs_open/174 seed_scope relay

Every command here had to be got right once. The gotcha is attached to each.

## Read the two allow-lists that decide whether a seed survives

They are **both** on `diagnose-dispatch-loop`, and reading only the second is how
the ticket reached an insufficient fix candidate.

```bash
# 2. the input_mapping (the one the ticket named)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT jsonb_pretty(default_config #> '{workflow,steps,call_handler,config,input_mapping}')
  FROM agent_definitions WHERE type='diagnose-dispatch-loop'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"

# 1. the RETURNING projection (the one it did not) — a key absent HERE cannot be
#    forwarded no matter what the mapping says
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT substring(default_config #>> '{workflow,steps,claim_item,config,query}'
       from position('RETURNING' in default_config #>> '{workflow,steps,claim_item,config,query}'))
  FROM agent_definitions WHERE type='diagnose-dispatch-loop'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```

**GOTCHA:** the authority both must agree with is a third place — the *callee's*
declared contract:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT jsonb_pretty(input_contract) FROM agent_definitions WHERE type='diagnose-orchestrator'
 AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```

## Run the relay-gap check

```bash
./scripts/audit-relay-gaps.sh          # human-readable; --json for the raw report
# exit 0 = clean · 1 = findings OR unmatched registry entries · 2 = could not determine
```

**GOTCHA (the important one):** `config-key-audit` dispatches on `os.Args[1]` and
**falls through to its DEFAULT report for an argument it does not recognise** —
valid JSON, exit 0. If `--relay-gaps` is ever unwired from `main()`, a naive
script would report "clean" having checked nothing. The script therefore asserts
the report's own keys (`findings`/`uncovered_relays`/`unmatched_registry_entries`)
and exits 2 if they are absent. **Do not remove that guard to make the script
shorter.**

**GOTCHA:** the export query must include `input_contract`. The other
`config-key-audit` modes do not need it, so copying their query gives you a
report where every relay is "unmatched — callee declares no input_contract".

## Prove the detector still detects (do this when you change it)

A detector that has never fired is unproven. Rebuild the **pre-fix** config from
migration 289's own snapshot and require exit 1:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
WITH prefix AS (
  SELECT default_config FROM agent_definitions_backup
   WHERE type='diagnose-dispatch-loop' ORDER BY snapshot_taken_at DESC LIMIT 1)
SELECT jsonb_agg(jsonb_build_object(
         'type', a.type,
         'workflow', CASE WHEN a.type='diagnose-dispatch-loop'
                          THEN (SELECT default_config->'workflow' FROM prefix)
                          ELSE a.default_config->'workflow' END,
         'input_contract', a.input_contract))
FROM agent_definitions a
WHERE a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false AND a.is_active
  AND a.default_config ? 'workflow';" > /tmp/prefix.json

go run ./cmd/config-key-audit --relay-gaps < /tmp/prefix.json; echo "exit=$?"   # MUST be 1
```

**GOTCHA:** `echo "EXIT=$?"` after a pipe into `head`/`tail` reports the **pager's**
exit code, not the tool's. Redirect to `/dev/null` or use `${PIPESTATUS[0]}`.

## Verify a seed actually arrived (after the chassis roll)

Field-present is a **weaker claim than scope-used**, and the fallback chain is
exactly what pulls them apart. Assert both.

```bash
# (a) the field reached the agent. Use the RUN correlation stamped onto the item
#     as spec.dispatch_correlation_id — NOT the intake correlation. They differ.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT collected_data->'input_data'->'seed_scope'
  FROM orchestration_states WHERE owner_agent_type='diagnose-agent'
   AND correlation_id='<run corr>';"

# (b) the scope was USED — the arm the fallback chain actually took
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT collected_data->'assembled'->>'scope_source'
  FROM orchestration_states WHERE correlation_id='<run corr>';"   -- expect 'seed'
```

`scope_source` values: `route` (loop-back revision) · `seed` (the caller's) ·
`code_results` (nobody chose it). Only the last is ambiguous, and only it renders
a warning into the bundle.

**GOTCHA:** `orchestration_states` retains barely a **day** here. Any count off it
is a snapshot, not a census; `site_work_items` retains longer.

## Find the run correlation for an intake

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT item_key, claimed_by, spec->>'dispatch_correlation_id' AS run_corr,
       spec->'seed_scope' AS seed
  FROM site_work_items
 WHERE item_type='needs_diagnosis' ORDER BY created_at DESC LIMIT 5;"
```

## Applying the migration

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/agent_docs/sql_for_agents/289_diagnose_dispatch_loop_forwards_seed_scope.sql
```

**GOTCHA:** there is **no `min(jsonb)` aggregate**. The first run of this file
failed on exactly that (`SELECT count(*), min(default_config #> '{...}')`), and —
correctly — the whole transaction rolled back, leaving config untouched. Count in
one statement, select the values in another.

**GOTCHA:** a `replace()` on a query string is a **silent no-op** if the needle
does not match byte-for-byte. Assert the projection is present afterwards; do not
trust the UPDATE's row count, which counts rows *touched*, not text *changed*.
