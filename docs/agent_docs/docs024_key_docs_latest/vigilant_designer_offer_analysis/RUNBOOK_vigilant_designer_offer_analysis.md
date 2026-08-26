# RUNBOOK — vigilant designer + offer analyser

Commands that were hard to get right, with their gotchas. Update HERE when one changes.

## DB access

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## The improvement loop, hand-fired (the manual-trigger mode this programme runs in)

The per-site sweep script (read its blast-radius header first):

```
./run_improvement_sweep_once.sh <site-domain-or-id>     # repo root; check the header before first use
```

Gotchas already known:
- `scheduled_tasks.last_triggered_at` / `last_completed_at` keep advancing while nothing runs
  (fire-and-forget stamp). Measure liveness at `orchestration_states` (newest run for the
  agent), never at the task row.
- A wedged head orchestration freezes a dispatch lane until a pod roll (dispatch-queue lane).
- No orchestration dispatch within ~300s of a chassis pod (re)start — silently dropped.

## Watching a finding travel (the drain proof)

```sql
-- born
SELECT id, item_type, status, handler_agent, item_key, created_at
FROM site_work_items WHERE site_id='<id>' AND item_type='<type>' ORDER BY created_at DESC LIMIT 5;
-- promoted (only improvement-loop.triage_findings may do this — migration 286, single owner)
-- claimed/complete: watch status + claimed_by; a FAILED step can show COMPLETED with error NULL — read __step_error
```

## Verifying a deploy (image roll)

```
kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<symbol you ADDED>"'
# every replica; plus a NEGATIVE control (a string the change REMOVED, expect 0); grep -ic for case traps
```

Image before config, always: a check name in a checks array before the binary carries it is a
FATAL run (149 B4). An unregistered action name in a workflow is inert at best.

## Migrations (sql_for_agents)

- Dry-run per session and after every roll; `--apply` takes EVERY pending file — scope the dir.
- Take a snapshot before UPDATE-ing an agent definition (`bak_ad_<agent>_<date>` pattern —
  see `bak_ad_designdiscovery_20260727` precedent).
- Seed key is `start_step`, never `initial_step` (VIZ-012 lesson); VERIFY blocks read `start_step`.
- Next free migration number: check `ls docs/agent_docs/sql_for_agents/ | sort -V | tail` at
  write time — concurrent sessions take numbers hourly; 289 was in use on 2026-08-02.

## Council submissions (platform/ internal/ pkg/ changes)

```
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
# budget ~30 min (dispatch queues behind the fleet); find your run by payload, not printed id;
# commit with Council-Submitted: <corr> if committing before the verdict
```

## Single-owner audit (must stay clean after any loop/workflow edit)

```
./scripts/audit-single-owner-actions.sh
```

---

## Imagery / `component_expresses` (added 2026-08-26, migration 644)

### What a component is advertising to the planner

Three live planner menus embed this call, so this IS what the model is shown:

```sql
SELECT name, array_to_string(component_expresses(html_template, input_schema), ', ') AS expresses
  FROM content_components
 WHERE is_active AND component_level IN ('section','element')
 ORDER BY name;
```

⚠ **Gotcha:** `component_expresses` is `IMMUTABLE` and lives in the DB, not in Go. `grep` will not
find its consumers — they are embedded in `agent_definitions` config. Find them with:

```sql
SELECT a.type FROM agent_definitions a,
  LATERAL jsonb_path_query(a.default_config, 'strict $.**.query ? (@ like_regex "component_expresses")') x
 WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;
```

### Before/after control for ANY change to a derived vocabulary — in ONE snapshot

⚠ **Do not capture BEFORE and AFTER as two queries.** Other lanes create components continuously;
five appeared mid-run on 2026-08-26 and the two-snapshot control broke outright (381 rows vs 386).
Compute both sides in one query so no concurrent writer can skew the pairing — see
`scratchpad/imagery/control_atomic.sql` pattern, reproduced in migration 644's own temp table:

```sql
CREATE TEMP TABLE _before ON COMMIT DROP AS
SELECT id, name, component_expresses(html_template, input_schema) AS tok FROM content_components;
-- ... apply the change ...
-- 1. nothing may LOSE a token:
SELECT count(*) FROM _before b JOIN content_components c ON c.id=b.id
 WHERE EXISTS (SELECT 1 FROM unnest(b.tok) t WHERE t <> ALL (component_expresses(c.html_template,c.input_schema)));
-- 2. nothing may change by anything OTHER than gaining the new token:
SELECT count(*) FROM _before b JOIN content_components c ON c.id=b.id
 WHERE array_remove(component_expresses(c.html_template,c.input_schema),'image')
       IS DISTINCT FROM array_remove(b.tok,'image');
```

⚠ **The row COUNT is not a control** — a variant that also suppressed `list` changed the same number
of rows while three components silently lost a capability. Assert the two above, and **induce them**
(run the broken variant and watch them fire) or you have only proven they can pass.

### What a `site_assets.<path>` field will ACTUALLY render

```bash
grep -n -A14 'imageRoleAliases = map' platform/orchestration/imageryplan/imageryplan.go
```

If your `<path>` is a key in that map it resolves to the **ROLE**, not to your field — `image`,
`background`, `banner`, `header_image`, `product_screenshot`, `screenshot` and others all mean
`hero`. Confirm at the artefact, across the whole estate, never on one component:

```sql
WITH f AS (SELECT c.id, c.name, k.key AS field
             FROM content_components c, jsonb_each(COALESCE(c.input_schema->'fields','{}'::jsonb)) k
            WHERE k.value->>'source' = 'site_assets.image' AND k.value->>'type' IN ('url','image','image_url'))
SELECT s.domain, p.url, f.name, pc.content_data->>f.field
  FROM f JOIN page_components pc ON pc.component_id=f.id
  JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id ORDER BY 1,2;
```

⚠ A per-section-looking value is **not** proof the source works — it may be hand-seeded and surviving
by `carryStored`. The disconfirming shape is a value matching the page's HERO asset.

### Dry-running a migration safely against the live DB

```bash
sed 's/^COMMIT;$/ROLLBACK;/' <migration>.sql > /tmp-scratch/dryrun.sql
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < /tmp-scratch/dryrun.sql
```

Then mutate the dry-run file to make each guard fail, and confirm each aborts. Recording it here
because `CREATE OR REPLACE FUNCTION` **is** transactional in Postgres, so this genuinely leaves the
live function untouched — verify with a must-be-absent probe afterwards.
