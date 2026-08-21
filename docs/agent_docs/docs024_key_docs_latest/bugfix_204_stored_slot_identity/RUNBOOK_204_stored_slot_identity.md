# RUNBOOK — 204 stored-slot-identity

Every query/command that was hard to get right, with its gotcha attached.
Change it HERE when it changes, not in scrollback.

---

## Is the defect still live? (three checks, run all three)

### 1. The census — how many section names no component can resolve

```sql
WITH s AS (
  SELECT si.domain, p.name AS page, jsonb_array_elements_text(p.sections) AS sec
  FROM pages p JOIN sites si ON si.id = p.site_id
  WHERE p.status = 'active'
    AND jsonb_array_length(COALESCE(p.sections, '[]'::jsonb)) > 0)
SELECT domain, count(*) AS names,
       count(*) FILTER (WHERE NOT EXISTS (
         SELECT 1 FROM content_components cc WHERE cc.function = s.sec OR cc.name = s.sec)) AS unresolvable,
       count(DISTINCT page) FILTER (WHERE NOT EXISTS (
         SELECT 1 FROM content_components cc WHERE cc.function = s.sec OR cc.name = s.sec)) AS pages_affected
FROM s GROUP BY domain ORDER BY 3 DESC;
```

⚠ **Gotcha:** this measures the *population at risk*, not the damage. A site drops OFF
this list when its sections are re-pointed to functions — which is what happened to
loancalculator.co.uk (57 → 0) after the 08-06 fix. A shrinking number here is not
evidence the defect is shrinking.

⚠ **`cc.name` is deliberately in the predicate.** Dropping it over-reports: the
resolver's arm 4 (`nameToFunc`) really does resolve by component name.

### 2. The drop record — what the resolver actually threw away

```sql
SELECT action, count(*) AS drops,
       count(*) FILTER (WHERE context->>'section' ~ '-[0-9]+$') AS positional_shaped,
       count(DISTINCT context->>'page') AS pages,
       min(occurred_at)::date AS first_seen, max(occurred_at)::date AS last_seen
FROM agent_error_log
WHERE error_code = 'PLAN_SECTION_NAME_DROPPED'
GROUP BY action ORDER BY 2 DESC;
```

⚠ **Gotchas, three of them:**
- The timestamp column is **`occurred_at`**, not `created_at`. `created_at` does not
  exist on this table and the query errors out (`\d agent_error_log` before writing SQL).
- The record only exists from the **2026-08-16 roll** (bugs_open/282's lane shipped it).
  `min(occurred_at)` is therefore the record's birthday, not the defect's — never quote
  it as the date the drops began.
- `action` is the *step* name (`validate_plan`), or `apply_gap_plan:<arm>` for the three
  gap-plan sites, which name their provenance explicitly rather than inheriting it.

### 3. Is the arm armed? — the live config, not the seed

```sql
SELECT type, step.key AS step_name,
       step.value->'config'->>'validate_components' AS validate_components,
       step.value->'config'->>'menu_field'          AS menu_field,
       step.value->'config'->>'existing_pages_field' AS existing_pages_field
FROM agent_definitions ad,
     LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS step
WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
  AND step.value->'config' ? 'validate_components'
ORDER BY 1, 2;
```

⚠ The `is_snapshot` / `deleted_at` filters are load-bearing — without them you read
archived rows and get a config that is not running. And read the LIVE row, never the
`SEED_*.sql`: the seed records what the agent WAS.

## Who else is on this bug

```bash
scripts/who-owns.py 204          # advisory, ~0.3s, no cluster calls
```
⚠ It reads COMMITS, so a session mid-fix with a dirty tree is invisible. Also check the
tree, and re-run at each phase boundary — the answer goes stale in minutes.

```sql
SELECT id, item_type, status, left(summary,110), created_at
FROM site_work_items
WHERE status NOT IN ('complete','cancelled','rejected')
  AND (summary ILIKE '%positional%' OR summary ILIKE '%plan_sections%'
    OR summary ILIKE '%validate_plan%' OR summary ILIKE '%validate_components%')
ORDER BY created_at DESC;
```

## Enumerate the resolver's call sites (do not trust a memory of "one")

```bash
grep -rn "loadComponentNameResolver(\|resolver.resolve(\|recordDroppedSectionNamesFor" \
  --include=*.go platform/ | grep -v "_test.go"
```
⚠ `resolver.resolve(` also matches an UNRELATED resolver in `plan_sections_action.go:2403`
and `render_site_components_action.go:902` — that one takes `(ctx, source)` and is the
source-field resolver, not the component-name one. Check the arity before counting.

## The optional-key budget (RFC_022 / WFA-013), before adding any config key

```bash
./scripts/audit-optional-key-budget.sh --json | python3 -c "
import json,sys
d=json.load(sys.stdin)
for a in d['actions']:
    if a['action'] in ('plan_sections','apply_gap_plan','validate_site_plan'):
        print(a['action'], a['optional_keys'], a['over_budget'], a['optional'])"
```
⚠ **`validate_site_plan` prints NOTHING, and that is a finding, not an empty result.**
It has no `RegisterActionInputSpec` at all (207 exist across the estate), so every config
key it reads — `validate_components`, `menu_field`, `existing_pages_field`, `plan_field`,
`honour_realised_identity` — is invisible to the budget check. An action counted as ZERO
is exactly the shape CLAUDE.md warns about under RFC_022.

## Before firing any planner canary at a decomposed site

⚠ **Take the snapshot FIRST and be ready to cancel the queue.** The 2026-08-20 incident
is the worked example: the run emptied `sections` on 41 pages AND filed 20 `needs_page`
+ 12 sibling items, all `triaged` and claimable within seconds. The DB damage is
reversible from a snapshot; a claimed `needs_page` that builds an empty page over a live
one is not.

```sql
CREATE TABLE pages_bak_<yyyymmdd>_<slug> AS SELECT * FROM pages WHERE site_id = '<uuid>';
```
Then, after the run, cancel before repairing — the queue is what turns a DB-only
regression into a deployed one.
