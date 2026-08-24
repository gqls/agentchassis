# RUNBOOK — bugs_open/382 empty-kind image routing

Every query/command that was hard to get right, with its gotcha attached.
DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## 1. The 382 census (does the bug still bite?)

```sql
SELECT s.domain, a.purpose, a.asset_key, a.created_at
FROM assets a JOIN sites s ON s.id=a.site_id
WHERE a.origin_model='stability/stable-diffusion-xl-1024-v1-0' AND a.created_at > '2026-07-18'
ORDER BY a.created_at DESC;
```
**Gotcha:** this is an OUTPUT census. It cannot date the MECHANISM — the same lane already
learned that the hard way on this site (`vetcomparison/NOTES` correction, 2026-08-24). Assets
only accrue when generation RUNS, so a quiet fortnight reads identical to a fix. Pair it with §2
(the live config, which is where the defect lives) and §5 (demand).

## 2. Which live `call_agent` steps forward `kind` to the image generator

```sql
SELECT type, step.key, step.value->'config'->'input_mapping' AS mapping,
       step.value->'config'->>'default_kind' AS dead_default_kind
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') step
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND step.value->>'action'='call_agent'
  AND step.value->'config'->>'agent_type'='image-generator';
```
**Gotcha:** `default_kind` is at the CONFIG level and is read by nothing (see PLAN §2) — so a
non-null value in the last column is the trap, not the reassurance. Only a `kind` or `kind?` key
*inside* `input_mapping` actually travels.

## 3. Dead-key census across ALL live `call_agent` steps

```sql
WITH steps AS (
  SELECT ad.type AS agent, s.key AS step, s.value->'config' AS cfg
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s.value->>'action'='call_agent'
)
SELECT k, count(*) FROM steps, jsonb_object_keys(cfg) k GROUP BY 1 ORDER BY 2 DESC;
```
Result 2026-08-24: target_role 59, timeout_seconds 58, input_mapping 57, agent_type 42,
output_mapping 8, error_step 7, **default_kind 3**, prompt 1, input_data 1.
**Gotcha (count staleness, CLAUDE.md rule):** these are counts as of 2026-08-24. Re-run before
quoting; the population grows by addition.

## 4. Joining an SDXL asset to the work item that produced it

`orchestration_states` retains roughly **one day** (min/max on 2026-08-24 spanned 2026-08-23
12:15 → 2026-08-24 13:16, 4,956 rows), so for anything older the orchestration row is GONE.
Use the work-item archive and match on time:

```sql
SELECT w.item_type, w.status, w.summary, w.updated_at
FROM site_work_items_archive w JOIN sites s ON s.id=w.site_id
WHERE s.domain='ai-agent-orchestration.com'
  AND w.updated_at BETWEEN '2026-08-11 15:00' AND '2026-08-11 18:00'
ORDER BY w.updated_at;
```
**Gotcha (this cost a wrong turn — see NOTES):** querying `site_work_items` alone returns ZERO
`unfulfilled_hero_variant` rows, which reads as "that producer never ran". It is a rolling
window; closing a row moves it to `site_work_items_archive`. Query BOTH, always.

## 5. Demand control — has the producer even run since?

```sql
SELECT item_type, status, count(*), max(updated_at) FROM site_work_items_archive
WHERE item_type IN ('unfulfilled_hero_variant','needs_hero_image','needs_logo') GROUP BY 1,2
UNION ALL
SELECT item_type, status, count(*), max(updated_at) FROM site_work_items
WHERE item_type IN ('unfulfilled_hero_variant','needs_hero_image','needs_logo') GROUP BY 1,2;
```
**Why it is in the runbook:** a post-fix zero in §1 is meaningless without this. If no
`unfulfilled_hero_variant` has run since the fix, §1's empty result measures the absence of
traffic, not the presence of a fix.

## 6. Is a migration actually applied?

```sql
SELECT filename, applied_at FROM schema_migrations WHERE filename LIKE '390%';
```
**Gotcha:** returns 0 rows for `390`, and 390 IS live (verified in the agent row). It was applied
by hand — its commit message says "APPLIED + verified live" — without a ledger row. So
`schema_migrations` is necessary but NOT sufficient evidence; read the live `agent_definitions`
row for the change itself.

## 7. Provenance of the running adapter (did a code fix ship?)

```bash
kubectl -n ai-persona-system logs -l app=image-generator --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <your-commit> <the stamp>   # exit 0 = your commit is in it
```
**Gotcha:** the provenance line is a STARTUP line and scrolls out of range on a busy service. An
empty grep means "not in range", not "unstamped". Fall back to the binary probe with BOTH
controls (a sha that must be absent and one that must be present) — never `strings`, never a
discovery grep for "some 40-hex string".
