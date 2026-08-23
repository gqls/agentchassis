# RUNBOOK — bugs_open/367 (the commands, with their gotchas attached)

## Read what the router ACTUALLY does — always from the live row, never the seed file

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -tAq -c \
  "SELECT default_config->'workflow'->'steps'->'classify'->'config'->>'query'
     FROM agent_definitions
    WHERE type='required-fields-missing-handler' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL" > classify_live.sql
```

> **Gotcha:** seed `410` was edited in place three times and its comments say "v3", but the
> live row is **`version = 1`**. A subagent reported "currently v3" from the file and it was
> wrong about the row. Write migration guards against the row.

## Run the router's own classification by hand (this is the §6 verification bar)

`$1`/`$2` are placeholders, so psql needs `PREPARE`/`EXECUTE`:

```bash
{ printf 'PREPARE cls(uuid,uuid) AS\n'; cat classify_live.sql; printf ';\n';
  echo "EXECUTE cls('<site_id>'::uuid,'<work_item_id>'::uuid) \gx"; } \
| kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

To classify an arbitrary (page, slot) with **no work item**, swap the `item` CTE for a
literal — read-only, writes nothing:

```
WITH item AS (SELECT spec FROM site_work_items WHERE id = $2::uuid)
  ->
WITH item AS (SELECT '{"page_name":"X","slot_name":"Y","missing_fields":["headline"]}'::jsonb AS spec)
```

## Re-classify the WHOLE population under two queries and diff them

The blast-radius check. `PREPARE` cannot be used in a lateral join, so use a plpgsql loop:

```sql
DO $$ DECLARE r record; a text; b text; BEGIN
  CREATE TEMP TABLE c(item uuid, cur text, new text);
  FOR r IN SELECT id, site_id FROM site_work_items WHERE item_type='required_fields_missing' LOOP
    EXECUTE $q$ SELECT route FROM ( <OLD QUERY> ) s $q$ INTO a USING r.site_id, r.id;
    EXECUTE $q$ SELECT route FROM ( <NEW QUERY> ) s $q$ INTO b USING r.site_id, r.id;
    INSERT INTO c VALUES (r.id,a,b);
  END LOOP; END $$;
SELECT cur, new, (cur IS DISTINCT FROM new) AS changed, count(*) FROM c GROUP BY 1,2,3;
```

> **Gotcha:** run it with `--single-transaction` so the TEMP table survives to the SELECT.

## Test a migration WITHOUT touching production

Three levels, each worth doing:

```bash
# 1. does it parse and do its own verify assertions pass?
sed 's/^COMMIT;$/ROLLBACK;/' <migration>.sql | kubectl ... psql -v ON_ERROR_STOP=1 -f -

# 2. does the PATCHED row behave? (apply, run controls, roll back — all one transaction)
#    see 574_behav.sql in the scratchpad pattern: read the query back from the row
#    INSIDE the transaction, EXECUTE it, assert, then ROLLBACK.

# 3. does the rollback round-trip? apply + rollback in one transaction and require
#    default_config to come back BYTE-IDENTICAL:
#      CREATE TEMP TABLE before_cfg AS SELECT default_config FROM agent_definitions WHERE ...;
#      <migration without BEGIN/COMMIT>  <rollback without BEGIN/COMMIT>
#      IF a IS DISTINCT FROM b THEN RAISE EXCEPTION 'ROUND-TRIP FAILED'; END IF;
#      ROLLBACK;
```

> **Gotcha:** `snapshot_agent` is overloaded (`(text)` and `(text,text)`) — a bare literal
> gives `function snapshot_agent(unknown) is not unique`. Use the two-arg form with casts.
> Same class: `to_jsonb('a' 'b')` on adjacent literals is `unknown`; cast with `::text`.

## Where a CLOSED item's route actually lives

**Not on the row.** The dispatch loop's `mark_complete` overwrites `result` with spawn
bookkeeping. Parked rows keep `result->>'route'` and the message in `error`; completed ones
do not.

```sql
SELECT collected_data->'triage'->>'route' AS route,
       collected_data->'triage'->>'target_state' AS target_state
FROM orchestration_states WHERE workflow_plan->>'start_step'='classify' ORDER BY created_at;
```

> **Gotcha, and it cost me a wrong claim:** `orchestration_states` retains about **two
> days**. `min(created_at)` reads 2026-07-19 only because ~24 stuck `CANCELLED` rows survive
> the cleanup. Plot rows per day before quoting a minimum as a retention window.

## Council

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```
> **Gotcha:** an edit whose `sketch` is entirely comments is refused client-side — *"a fix
> plan proposes changes, not observations"*. Drop the edit and put the observation in
> `rationale`; a comment-only edit gives reviewers nothing to judge anyway.
