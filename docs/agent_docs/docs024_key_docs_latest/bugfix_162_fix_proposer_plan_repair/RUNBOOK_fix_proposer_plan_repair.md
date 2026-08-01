# RUNBOOK — bugs_open/162 `fix-proposer` plan repair loop

Every command here was run on 2026-07-31 unless marked otherwise. Gotchas are attached to
the command that needs them, not collected at the bottom.

## 0. Is the Go half live? (do this BEFORE applying any config)

The repair loop is config + binary. Config is live immediately; the binary is not. A
config that names a `repair_step` against a binary that does not read it is inert.

```bash
for p in $(kubectl -n ai-persona-system get pods -o name | grep chassis | sed 's|pod/||'); do
  echo "=== $p ==="
  kubectl -n ai-persona-system exec $p -- sh -c \
    'strings /app/agent-chassis | grep -c "plan_validation_refusal";
     echo "-- positive control:"; strings /app/agent-chassis | grep -c "diagnose_persist_fix_plan";
     echo "-- negative control:"; strings /app/agent-chassis | grep -c "zzz_not_a_real_symbol_162"'
done
# 2026-07-31: 2 / 11 / 0 on BOTH replicas.
```

**Gotcha:** grep the symbol your change ADDED *plus a positive control in the same exec*
(`MEMORY.md`: a roll is not evidence your fix shipped — the image may predate your commit
and carries no provenance). A bare `grep -c` returning 0 exits 1, so a compound exec
"fails" while printing perfectly good numbers — read the output, not the exit code.

**Gotcha:** do BOTH replicas. `kubectl logs deploy/X` reads one pod of N.

## 1. Who actually consumes the shared action?

`orchestration_states` has **no `agent_type` column** and `agent_definitions` hides the
consumer inside a JSON step, so the obvious queries either error or lie. This is the one
that answers it:

```sql
SELECT a.type, s.key AS step_name,
       s.value->>'next_step'                     AS next_step,
       s.value->'config'->>'repair_step'         AS repair_step,
       s.value->'config'->>'max_repair_attempts' AS max_attempts,
       (a.default_config->'workflow'->'steps') ? 'check_plan_valid' AS has_router
  FROM agent_definitions a,
       LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
 WHERE a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
   AND s.value->>'action' = 'diagnose_persist_fix_plan'
 ORDER BY 1,2;
```

2026-07-31, **four** consumers — one more than `bugs_open/162` names:
`feature-designer` (opted in), `fix-proposer` (opted in by this work),
`council-gate` (deliberate opt-out), `council-gate-036scratch`
(**`is_active=false`, 0 runs in 30 days — inert**, which is why 162 counting three was
not wrong in substance).

## 2. Dry run, then apply

```bash
# DRY RUN — turns the file's COMMIT into a ROLLBACK
sed 's/^COMMIT;/ROLLBACK;/' docs/agent_docs/sql_for_agents/273_fix_proposer_plan_repair_loop.sql \
  | kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
      psql -U clients_user -d clients_db -v ON_ERROR_STOP=1

# APPLY
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  < docs/agent_docs/sql_for_agents/273_fix_proposer_plan_repair_loop.sql
```

**Gotcha:** the file's closing `DO` block raises a **WARNING** (not an exception) if
`repair_plan.max_tokens` differs from `propose`'s. A warning that does not appear is
indistinguishable from a check that did not run, and `IS DISTINCT FROM` is false when
**both sides are NULL** — so a missing `ai_service` on both steps would also print
nothing. Confirm the values are real rather than reading silence as agreement:

```sql
SELECT s.key, s.value->'config'->'ai_service'->>'max_tokens'
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
 WHERE a.type='fix-proposer' AND a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
   AND s.value->'config' ? 'ai_service' ORDER BY 1;
-- propose = 8000, and after 273 repair_plan = 8000. Both real, not both NULL.
```

**Gotcha (rollback):** `snapshot_agent(text,text)` writes to **`agent_definitions_backup`**,
NOT an `is_snapshot` row, and that table keeps the SOURCE row's `created_at` — so a
restore MUST order by **`snapshot_taken_at`**. Rollback SQL is in 273's own header.

## 3. Verify the applied graph AT THE ROW

Do not trust the migration's own verification block alone — it is the same code path that
just wrote the change.

```sql
SELECT s.key, s.value->>'output_field' AS output_field, s.value->>'next_step' AS next_step,
       s.value->'config'->>'repair_step' AS repair_step,
       s.value->'config'->>'condition'   AS condition,
       s.value->'config'->>'then_step'   AS then_step,
       s.value->'config'->>'else_step'   AS else_step
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
 WHERE a.type='fix-proposer' AND a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false
   AND s.key IN ('persist_plan','check_plan_valid','repair_plan','select_panel','propose');
```

Expected: `persist_plan` → `check_plan_valid` (output_field `plan_persisted`,
repair_step `repair_plan`); `check_plan_valid` → then `select_panel` / else `repair_plan`;
`repair_plan` → `persist_plan` with output_field `proposal` (the SAME field `propose`
writes, so `persist_plan`'s default `plan_field: proposal.result` reads the repaired plan).

Re-running the migration must now fail:
`ERROR: 273: already applied — fix-proposer already has a check_plan_valid step`.

## 4. THE CHECK THAT MATTERS — does the router's field actually resolve?

This is the one that would have caught a silent fleet-wide regression, and it is a
**data** check, not a config check.

The router reads `plan_persisted.plan_valid`. If a step's result were *wrapped* — the way
`execute_llm_prompt` with `output_format=json` leaves its object under
`<output_field>.result` — the path would resolve to nothing, `compareValues(nil,"true")`
returns false, and **every valid plan would route to `repair_plan`**. That loop is *not*
bounded: the repair counter counts refusal artefacts, and a misrouted VALID plan writes
none, so it would spin to `fuel_budget`.

```sql
SELECT orchestration_id||' | '||current_step||' | keys=['||
       (SELECT string_agg(k, ', ' ORDER BY k) FROM jsonb_object_keys(collected_data->'plan_persisted') k)||']'||
       ' | plan_valid='||COALESCE(collected_data->'plan_persisted'->>'plan_valid','<ABSENT>')
  FROM orchestration_states WHERE collected_data ? 'plan_persisted'
 ORDER BY updated_at DESC LIMIT 4;
```

2026-07-31, four live rows: keys
`[edit_count, files, persisted, plan_json, plan_valid, summary]`, `plan_valid=true`.
**Unwrapped, at collected_data root.** Hazard excluded.

## 5. Inducing a refusal — the method, and why this lane did NOT run it

Method (from the 099 lane's runbook), set the per-stage edit cap to 1 so any natural
multi-edit plan trips it — repairable without losing scope, which is what the repair
prompt asks for:

```sql
-- ARM
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,persist_plan,config,max_edits}', '1'), updated_at = now()
 WHERE type='fix-proposer' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- ... fire ./docs/.../091_TRIGGER_fix_proposer_v1.sh <confirmed_diagnosis_correlation> ...
-- DISARM — restore to 8, NOT to absent: 8 is the default in both the action and the spec
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,persist_plan,config,max_edits}', '8'), updated_at = now()
 WHERE type='fix-proposer' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
```

**CHECK THE QUEUE BEFORE ARMING — the 099 lane recorded getting this order wrong.**
The arm is a fleet-wide edit to a shared live agent, and the dispatch window is ~30
minutes, so every other lane's run in that window trips too:

```sql
SELECT orchestration_id, current_step, status, created_at FROM orchestration_states
 WHERE workflow_plan::text LIKE '%select_panel%' AND workflow_plan::text LIKE '%repropose%'
   AND status NOT IN ('COMPLETED','FAILED') ORDER BY created_at DESC LIMIT 8;
```

**2026-07-31: three other lanes' runs were in flight, so this lane did not arm.** The
route `persist_plan → check_plan_valid → repair_plan → persist_plan` is therefore
**[UNVERIFIED] on `fix-proposer` specifically**. It is proven on `feature-designer`, which
runs the identical shared Go code and the same-shaped graph (3 refusals on 2026-07-31: 2
routed to repair, 1 exhausted terminal). Run the induction in a quiet window to close it.

## 6. Did the mechanism fire? (both tables — they answer different questions)

```sql
-- the operator-facing record
SELECT agent_type, severity, count(*), max(occurred_at) FROM agent_error_log
 WHERE error_code='FIX_PLAN_VALIDATION_REFUSED' GROUP BY 1,2 ORDER BY 1;
-- the rejected plan itself, recoverable verbatim
SELECT source_agent, count(*), max(created_at) FROM diagnosis_artifacts
 WHERE kind='iteration_note' AND metadata->>'note_kind'='plan_validation_refusal' GROUP BY 1;
```

**Gotcha — `agent_error_log` has no `created_at`; the column is `occurred_at`.** Read
`\d agent_error_log` before writing the filter: the wrong name errors out loudly here, but
the same guess against a table that *does* have both is how a window silently slips.

**LANDMINE — an empty result here does NOT mean "no plan was refused".** See LANDMINES.md;
the action's own comment claimed it did, and that claim is false for four of its five
terminal exits.
