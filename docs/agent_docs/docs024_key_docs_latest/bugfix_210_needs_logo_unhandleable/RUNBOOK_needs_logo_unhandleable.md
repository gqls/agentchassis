# RUNBOOK — bugfix 210 (needs_logo slug)

Every query here was needed to get something right, with its gotcha attached. DB access:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## R1. Read the handler's LIVE step config (not a seed)

`seed-sql-is-history-live-row-is-fact`. **Gotcha:** `default_config->'workflow'->'steps'` is a
JSON **object**, not an array — `jsonb_array_elements` errors with *"cannot extract elements
from an object"*. Use `->` by name or `jsonb_each`.

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'call_logo_gen')
  FROM agent_definitions
 WHERE type='image-build-handler' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

**Second gotcha:** chaining two `jsonb_pretty(...)` calls with `||` in one `SELECT` returns
**nothing at all** if either side is NULL (`NULL || text` is NULL) — it looks like an empty
table rather than an error. Run them as separate statements.

## R2. Find every step in the fleet that runs a given action

Needed to establish that `generate_image` has exactly one call site. `jsonb_each` over the
steps object, because of R1's gotcha:

```sql
SELECT DISTINCT a.type, s.key AS step,
       s.value->'config'->>'prompt_template' IS NOT NULL AS step_has_prompt_template
  FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
 WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
   AND s.value->>'action' = 'generate_image';
```

**Do not** use `default_config::text LIKE '%generate_image%'` — the existing LANDMINE
(`LANDMINES.md:486`) records that prompt TEXT matches action names, so that predicate reports
agents with no such step.

## R3. Has the generic fallback ever produced a stored asset?

The measurement that licenses Fix C. **The point is that it could have come out otherwise** —
`a-count-you-kept-is-not-a-census` and the `[MEASURED]`-must-be-disconfirmable rule.

```sql
SELECT count(*) AS total_generated,
       count(*) FILTER (WHERE origin_prompt = 'Generate content based on the provided context.') AS exact_generic,
       count(*) FILTER (WHERE origin_prompt ILIKE '%provided context%') AS fuzzy_generic,
       count(*) FILTER (WHERE origin_prompt IS NULL OR origin_prompt='') AS null_or_empty
  FROM assets WHERE origin_type <> 'uploaded' OR origin_prompt IS NOT NULL;

-- POSITIVE CONTROL — run it, or the zero above proves nothing:
SELECT count(*) AS any_origin_prompt,
       count(*) FILTER (WHERE origin_prompt ILIKE '%logo%') AS mentions_logo,
       min(length(origin_prompt)) AS min_len, max(length(origin_prompt)) AS max_len
  FROM assets WHERE origin_prompt IS NOT NULL AND origin_prompt <> '';
```

2026-08-09: `399 / 0 / 0 / 55` and `344 / 25 / 147 / 3882`. The generic string is 46 chars —
**shorter than the minimum observed prompt**, so the two results corroborate each other.
**State the blind spot**: the 55 rows with no recorded prompt are invisible to this check.

## R4. Reproduce the cross-origin false positive

The single query that turned §6's `[UNVERIFIED]` guess into a mechanism.

```sql
SELECT p.name, substring(pc.rendered_html from '.{0,70}/assets/images/logo\.png') AS context
  FROM page_components pc JOIN pages p ON pc.page_id=p.id
 WHERE p.site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
   AND pc.build_status='deployed' AND pc.locked_at IS NULL
   AND pc.rendered_html LIKE '%/assets/images/logo.png%';
```

**Gotcha:** `page_components` has **no `component_name` column** — the obvious `SELECT` fails.
Use `p.name` (the page) or check `\d page_components` first.
**Use `substring(... from '<regex>')`, not `position`/`left`** — you need the surrounding
characters to see the `src="https://…` prefix, and that prefix *is* the finding.

## R5. Fleet split: same-origin vs cross-origin fallback references

⚠ **The mistake that cost a re-run** (logged in WRONG_CALLS): SQL `AND` binds tighter than
`OR`, so

```sql
WHERE build_status='deployed' AND locked_at IS NULL
  AND html LIKE '%logo.png%' OR html LIKE '%hero.jpg%'    -- WRONG
```

parses as `(A AND B AND C) OR D` and silently includes undeployed and locked components. It
returns a plausible table with no error. **Parenthesise the OR**:

```sql
WITH m AS (
  SELECT s.domain, pc.rendered_html AS h,
         CASE WHEN pc.rendered_html LIKE '%/assets/images/logo.png%' THEN 'logo.png' ELSE 'hero.jpg' END AS path
    FROM page_components pc JOIN pages p ON pc.page_id=p.id JOIN sites s ON s.id=p.site_id
   WHERE pc.build_status='deployed' AND pc.locked_at IS NULL
     AND (pc.rendered_html LIKE '%/assets/images/logo.png%'
       OR pc.rendered_html LIKE '%/assets/images/hero.jpg%')
)
SELECT domain, path, count(*) AS components,
       count(*) FILTER (WHERE h ~ 'https?://[^"'')> ]*/assets/images/(logo\.png|hero\.jpg)') AS crossorigin,
       count(*) FILTER (WHERE h ~ '["''(=]/assets/images/(logo\.png|hero\.jpg)') AS same_origin
  FROM m GROUP BY 1,2 ORDER BY 1,2;
```

2026-08-09: `logo.png` matches **one** component fleet-wide and it is **cross-origin**;
`hero.jpg` matches 141 components across 17 sites, **all same-origin**.

## R6. Current and latent exposure (the blast-radius query)

Who would file a promptless, unconsumable item on the next discovery run:

```sql
WITH refs AS (
  SELECT DISTINCT s.id AS site_id, s.domain,
         CASE WHEN pc.rendered_html LIKE '%/assets/images/logo.png%' THEN 'logo' ELSE 'hero' END AS purpose
    FROM page_components pc JOIN pages p ON pc.page_id=p.id JOIN sites s ON s.id=p.site_id
   WHERE pc.build_status='deployed' AND pc.locked_at IS NULL
     AND (pc.rendered_html LIKE '%/assets/images/logo.png%'
       OR pc.rendered_html LIKE '%/assets/images/hero.jpg%')
)
SELECT r.domain, r.purpose,
       EXISTS (SELECT 1 FROM assets a
                WHERE a.site_id=r.site_id AND a.purpose=r.purpose AND a.status='active') AS has_asset,
       (SELECT sp.data->'image_prompts' ? CASE WHEN r.purpose='logo' THEN 'logo' ELSE 'hero_home' END
          FROM site_specs sp
         WHERE sp.site_id=r.site_id AND sp.aspect='site_plan' AND sp.is_current LIMIT 1) AS plan_has_prompt
  FROM refs r ORDER BY 3, 4 NULLS FIRST, 1;
```

**Read `plan_has_prompt` NULL carefully** — it means *no usable planned prompt*, and it conflates
"no current `site_plan` row" with "row exists, `image_prompts` is NULL". Disambiguate:

```sql
SELECT count(*) FILTER (WHERE sp.site_id IS NULL) AS no_current_site_plan_row,
       count(*) FILTER (WHERE sp.site_id IS NOT NULL AND sp.data->'image_prompts' IS NULL) AS row_but_no_image_prompts,
       count(*) FILTER (WHERE sp.data->'image_prompts' IS NOT NULL) AS has_image_prompts_obj,
       count(*) AS sites_total
  FROM sites s LEFT JOIN site_specs sp
    ON sp.site_id=s.id AND sp.aspect='site_plan' AND sp.is_current;
```

2026-08-09: **33 / 1 / 5 / 39** — the producer's recovery branch is unavailable on 87% of sites.

## R7. Producer census — who files these items and do they carry the key

```sql
SELECT item_type, created_by, status, count(*) AS n,
       count(*) FILTER (WHERE spec->'image_prompts' IS NULL) AS no_key,
       count(*) FILTER (WHERE spec->'image_prompts'->>'logo' IS NOT NULL) AS has_logo,
       count(*) FILTER (WHERE spec->'image_prompts'->>'hero_home' IS NOT NULL) AS has_hero_home,
       count(*) FILTER (WHERE spec->>'prompt' IS NOT NULL) AS has_flat_prompt
  FROM site_work_items
 WHERE item_type IN ('needs_logo','needs_hero_image','unfulfilled_hero_variant')
 GROUP BY 1,2,3 ORDER BY 1,2,3;
```

**Do not** reach for `spec::text LIKE '%"image_prompts":%'` — the jsonb LANDMINE
(`jsonb-text-like-cannot-match-a-key-value-pair`): jsonb renders a **space** after the colon,
so that predicate matches nothing and reads as "no rows have it".

## R8. Verifying a fix (from the bug file §7, sharpened)

A stored logo alone proves nothing — assert the prompt's **source**:

```sql
SELECT status, error FROM site_work_items
 WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
   AND item_key='placeholder_image_in_use:logo';

-- the asset's recorded prompt is the durable evidence; chassis logs rotate in <1s
SELECT purpose, left(origin_prompt,120), origin_model, created_at
  FROM assets WHERE site_id='<site>' AND purpose='logo' ORDER BY created_at DESC LIMIT 3;
```

## R9. Ownership check before touching a shared bug number

```sql
-- N/A: shell, not SQL
python3 scripts/who-owns.py needs_logo_items_cannot_be_handled
```

**Gotcha:** `who-owns.py 210` is **ambiguous** — two unrelated bugs share the number and the
tool says so. Always pass the slug. And it reads **commits**, so a session mid-fix is invisible;
grep the live transcripts too:

```bash
for f in $(find ~/.claude/projects/-home-ant-projects-agentchassis -name '*.jsonl' -mmin -720); do
  n=$(grep -c 'check_placeholder_image_in_use\.go' "$f" 2>/dev/null)
  [ "$n" -gt 3 ] 2>/dev/null && echo "$n $f"
done | sort -rn
```

**Gotcha:** grepping for the bug's *filename* is useless — every session that lists `bugs_open/`
matches it. Grep for the **source file** the fix would touch.
