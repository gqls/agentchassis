# RUNBOOK — bug 020 data-integrity gate

Every command below was run (or is the exact command to run) for this workstream.
DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.

## Inspect the live tool-recreation-handler config

```sql
-- The live prompts (this row is keyed on type, and gets re-seeded fleet-wide —
-- always read the LIVE row, the 099 seed file is stale).
SELECT default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}'
FROM agent_definitions WHERE type='tool-recreation-handler' AND deleted_at IS NULL;

-- Live models (change via migrations, not the seed):
SELECT 'analyze:'  || (default_config #>> '{workflow,steps,analyze_tool,config,ai_service,model}'),
       'recreate:' || (default_config #>> '{workflow,steps,recreate_tool,config,ai_service,model}')
FROM agent_definitions WHERE type='tool-recreation-handler' AND deleted_at IS NULL;
-- 2026-07-21: analyze=claude-sonnet-5 @8000, recreate=claude-opus-4-8 @64000.

-- The step graph (step -> action -> next/error):
SELECT s.key, s.value->>'action', s.value->>'next_step', s.value->>'error_step'
FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
WHERE a.type='tool-recreation-handler' AND a.deleted_at IS NULL ORDER BY s.key;
```

## Verify Half A (prompt fix, migration 183) is live

```sql
WITH s AS (SELECT default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}' AS r,
                  default_config #>> '{workflow,steps,analyze_tool,config,prompt_template}'  AS a
           FROM agent_definitions WHERE type='tool-recreation-handler' AND deleted_at IS NULL)
SELECT position('## Data Integrity' in r)>0        AS integrity_section,
       position('9. No fake data or dummy' in r)=0 AS old_rule9_gone,
       position('"data_source"' in a)>0            AS data_source_captured
FROM s;   -- all three must be true
```

## Apply a patch-style agent-config migration (the 183 pattern)

Never re-INSERT the agent_definitions row (a concurrent re-seed clobbers it, and
you clobber theirs). Patch the one nested field with `jsonb_set(... replace(...))`,
anchored on a string confirmed UNIQUE against the live row first. Open with
`snapshot_agent`, close with a `DO` block that RAISEs on a no-op replace (so a bad
anchor rolls back instead of silently applying nothing).

```bash
# Confirm anchor uniqueness BEFORE writing the replace:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c "
SELECT (length(t)-length(replace(t,'<ANCHOR>','')))/length('<ANCHOR>')
FROM (SELECT default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}' t
      FROM agent_definitions WHERE type='tool-recreation-handler' AND deleted_at IS NULL) s;"  # must be 1

# Apply out of band (NOT run-migrations.sh --apply — that drags in other threads' pending):
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/183_tool_recreation_no_invented_data.sql

# Record the hand-applied migration in the ledger:
bash scripts/migration/run-migrations.sh --record-only \
  docs/agent_docs/sql_for_agents/183_tool_recreation_no_invented_data.sql --note "..."
```

Next free migration number = max(dir, ledger) + 1. Check BOTH:
```bash
ls -1 docs/agent_docs/sql_for_agents/ | grep -E '^[0-9]{3}_' | sort | tail
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
  "SELECT filename FROM schema_migrations ORDER BY filename DESC LIMIT 5;"   # NB column is 'filename', not 'version'
```

## Test Half B (detector) locally — no cluster, no image

```bash
go build ./platform/orchestration/actions/
go test  ./platform/orchestration/actions/ -run 'TestDetect_|TestDataSourceIsExternal' -v
```

## Ship Half B (the gate)

**DONE 2026-07-22: the gate is WIRED & LIVE on v1.0.1146 via migration 189.** The
detector action was pod-verified first; wiring applied + ledger-recorded. Routing:
`check_completeness → check_fabrication → route_fabrication` (fabricated → review →
complete, else → save_training_data). `WIRING_..._APPLY_AFTER_IMAGE.sql` is SUPERSEDED
by `sql_for_agents/189_wire_tool_fabrication_gate.sql`.

### What REMAINS to CLOSE 020 (do on the next image roll):

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')

# 1. Confirm the FAIL-SAFE hardening (commit 37d3bb119) is now in the image —
#    v1.0.1146 was built BEFORE it, so it runs the fail-OPEN detector:
kubectl -n ai-persona-system exec $POD -- sh -c 'strings /app/agent-chassis | grep -c uninspectable'  # must be >= 1

# 2. INDUCE THE FAULT (do NOT trust `complete`): drive a real fabrication through the
#    wired path and confirm it is HELD, not deployed. Recreate a data-backed tool, then:
#    - a needs_human_review item was raised (checkpoint_for_review), page NOT deployed:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
  "SELECT status, summary FROM site_work_items WHERE item_type='needs_human_review' AND summary ILIKE '%INVENT data%' ORDER BY created_at DESC LIMIT 5;"
#    - the rendered page carries NO generator symbols:
curl -s "https://<site>/<tool-page>?cb=$RANDOM" | grep -ciE 'Mulberry32|makePostcode|buildData|SUFFIXES'  # must be 0

# 3. Move 020 to bugs_closed.
```
> If a natural data-backed recreation won't fabricate (the prompt fix makes it
> unlikely), induce it with a SCRATCH agent (fixloop-036 pattern): clone
> tool-recreation-handler, replace recreate_tool with a stub emitting the known
> vetcomparison fabrication HTML, run it against a data-backed page, confirm the item
> lands at needs_human_review and never reaches save_sections/deploy_page.

## Council gate

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
# schema: .plan is an OBJECT {summary, edits[<=8]{file,operation,rationale,sketch}, grounded_in[], risks}
# SUBMISSION_CORR for this gate = 8eef369f-5d93-4ebc-8491-e4d96397a91a
```
Verdict (budget ~30 min — dispatch queues behind the fleet; a missing row is latency, not a drop):
```sql
-- Run state. If it sits at EXECUTING_STEP for hours it is WEDGED (003-class spawn
-- loss), NOT queued and NOT a verdict — resubmit with RESUBMIT_CORR (first run
-- 8eef369f wedged at review_tooling_provenance for 3h42m on 2026-07-21):
SELECT current_step, status, now()-updated_at AS since_update FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '8eef369f-5d93-4ebc-8491-e4d96397a91a';

-- THE verdict — keyed on YOUR correlation. This is the authoritative source:
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='8eef369f-5d93-4ebc-8491-e4d96397a91a' AND kind='council_report'
ORDER BY created_at;
```
> ⚠️ Do NOT read the verdict from `SELECT body FROM doc_notes WHERE categories ?
> 'council-gate' ORDER BY created_at DESC LIMIT 1` — that is the most recent council
> note **fleet-wide** and will hand you ANOTHER thread's verdict. If you must read the
> prose note, confirm the "submission correlation:" line in the body matches yours.
