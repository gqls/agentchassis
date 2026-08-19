# RUNBOOK — bugfix 301 (owned-guard ordering)

Every command here was needed and has a gotcha attached. DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## Read the live workflow (NOT the seed file — seed is history, the row is fact)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'load_page_record')
FROM agent_definitions
WHERE type='page-build-handler' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
Gotcha: `->'workflow'->'steps'` is an OBJECT keyed by step name, not an array —
`jsonb_array_elements` fails on it; use `jsonb_object_keys` / `#>` paths.

## Is the wasted chain still burning? (validity check, re-runnable)

```sql
SELECT os.owner_agent_type, os.current_step, os.status, os.created_at
FROM orchestration_states os
WHERE os.owner_agent_type IN ('page-build-handler','page-content-writer','internal-link-resolver')
  AND os.created_at > now() - interval '3 hours' ORDER BY os.created_at DESC LIMIT 30;
```
Gotcha: ~24 h retention — this can only ever show a fresh burst, never history. A
`complete_error` handler flanked by COMPLETED writer+resolver children (created AFTER the
parent — created_at is spawn time) is the bug's signature.

## The queue that guarantees future waste

```sql
SELECT count(*) FROM site_work_items wi JOIN pages p ON p.id=wi.page_id
WHERE wi.handler_agent='page-build-handler'
  AND wi.status IN ('detected','needs_human_review','unresolved','failed')
  AND COALESCE(p.rebuild_policy,'generic')='owned';
```
Gotchas: (1) live table = ~7-day window for TERMINAL rows (archiver) — fine here because
open statuses are never archived, but any terminal count needs the
`UNION ALL site_work_items_archive` form. (2) `rebuild_policy` is MUTABLE — for
retrospective classification use the guard's error text, never this join (277 lane §8).

## Post-roll verification (BOTH controls, or it proves nothing)

```sql
-- positive: early refusals, stamped by the load step
SELECT wi.id, wi.item_type, wi.status, wi.updated_at
FROM site_work_items wi
WHERE wi.status='wont_fix' AND wi.result ? 'owned_page_refusal'
  AND wi.updated_at > '<roll time>' ORDER BY wi.updated_at DESC;

SELECT spec->>'refused_by', count(*) FROM site_work_items
WHERE item_type='owned_page_review' AND created_at > '<roll time>' GROUP BY 1;
-- expect refused_by='load_page_record' rows appearing; 'save_page_sections' rows CEASING

-- no writer spawned for a refused item: the refused item's orchestration must have
-- NO page-content-writer child
SELECT count(*) FROM orchestration_states
WHERE owner_agent_type='page-content-writer' AND created_at > '<roll time>';
-- read against the count of generic-page builds in the window (the demand control)
```
Negative control: a generic-page build in the same window must still spawn the writer and
complete. Zero refusals + zero builds = demand artefact, not success (277 lane's rule).

Binary probe (per-service, both replicas; provenance line scrolls, so probe, don't grep logs):
```bash
kubectl -n ai-persona-system exec <chassis-pod> -- sh -c \
 "grep -aoE 'refuse_owned_page|OWNED_PAGE_GUARD|ZZQQ_ABSENT_NEEDLE' /proc/1/exe | sort -u"
# expect: refuse_owned_page PRESENT, OWNED_PAGE_GUARD PRESENT (long-lived control),
# needle ABSENT (probe discriminates)
```

## Apply migration 488 (single file, NEVER a bare --apply)

The runner takes EVERY pending file (LANDMINES); apply this one file directly:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/488_page_build_handler_refuses_owned_pages_before_the_writer.sql
```
Then record it in the ledger the way the runner does — check `schema_migrations`' shape
first (`\d schema_migrations`) and insert the exact FILENAME (numbers collide on this tree;
the filename is the key).

## 090 diagnosis run for this lane

Intake corr `7281193f-59c2-489a-a9f2-fd4d58408cf5`; run corr
`dd61df1b-0d93-46e6-9065-1e0b9623379a`. Find it by payload, not printed id:
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'correlation_id' = 'dd61df1b-0d93-46e6-9065-1e0b9623379a'
   OR correlation_id = 'dd61df1b-0d93-46e6-9065-1e0b9623379a';
-- terminal verdict:
SELECT body FROM doc_notes WHERE body LIKE '%dd61df1b%' ORDER BY created_at DESC LIMIT 3;
```
Gotcha: publish→run start measured at 29 min under load — a missing row is latency, not a
dropped dispatch; do NOT retrigger.
