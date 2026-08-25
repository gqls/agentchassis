# RUNBOOK — `bugs_open/388`, component contract vs storage identity

Every command here had a gotcha attached when it was first got right. The gotcha is the point.

## Re-run the bug's own population query (is 388 still valid?)

```sql
WITH c AS (
  SELECT cc.section_type, cc.function, cc.updated_at,
    (SELECT count(DISTINCT p.site_id) FROM page_components pc JOIN pages p ON p.id=pc.page_id
      WHERE pc.component_id=cc.id AND pc.build_status<>'removed') AS sites
  FROM content_components cc
  WHERE cc.is_active AND cc.forked_from IS NULL AND cc.component_level='section'
    AND cc.section_type IS NOT NULL)
SELECT count(*) FILTER (WHERE adv <> section_type) AS disagree, count(*) AS total_section_types
FROM (SELECT section_type, (array_agg(function ORDER BY sites DESC, updated_at DESC))[1] AS adv
      FROM c GROUP BY 1) t;
```

`27 | 120` as of 2026-08-25 (was `27 | 117` on 2026-08-24, the day it was filed).

⚠ **The count is not the finding, and this query alone CANNOT establish the bug.** It measures
whether two resolvers *could* name different rows. Whether they *do* depends on a third thing —
the prompt's function pin — which this query cannot see. See NOTES, "the first wrong turn".

## Does the writer obey the function pin? (the measurement the bug file did not have)

```sql
WITH calls AS (
  SELECT created_at,
    substring(prompt_rendered from 'Also set the top-level "function" in your output JSON to exactly: ([a-z0-9-]+)') AS pinned,
    substring(response_text from '"function"\s*:\s*"([^"]+)"') AS emitted,
    prompt_rendered LIKE '%REGENERATION FIELD-NAME RULE%' AS has_regen_block
  FROM llm_call_log
  WHERE agent_type='component-creator' AND step_name='generate_template')
SELECT count(*) AS total, count(*) FILTER (WHERE has_regen_block) AS regen_block,
       count(*) FILTER (WHERE pinned IS NOT NULL) AS pinned,
       count(*) FILTER (WHERE pinned IS NOT NULL AND pinned = emitted) AS obeyed,
       count(*) FILTER (WHERE pinned IS NOT NULL AND pinned <> emitted) AS disobeyed
FROM calls;
```

`672 | 11 | 11 | 11 | 0` all-history as of 2026-08-25.

⚠ **THE DENOMINATOR IS THE TRAP, TWICE OVER.**
1. 672 is all-history back to 2026-03-31, but the pin only exists since **2026-08-22** (`e1951c24b`,
   the `bugs_open/337` fix). Quoting "11 of 672" as an adoption rate is meaningless; the honest
   window is 08-22 onward, where it is 7 of 37 calls carrying the regeneration block.
2. **0 failures out of 11 is not evidence the pin is reliable.** Name the failure rate the sample
   could have detected: with n=11 and zero failures the 95% upper bound on the disobedience rate is
   roughly 24%. This sample cannot distinguish "always obeyed" from "obeyed 4 times in 5".

⚠ **`psql -tAc` and the backslash.** The `\s` and `\d` in these regexes survive because the SQL is
inside single quotes in the heredoc. Typed at a `psql` prompt directly, a leading backslash on a
line is eaten as a psql meta-command.

## Read the live component-creator prompt (where the pin actually lives)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT default_config->>'prompt_template'
FROM agent_definitions WHERE type='component-creator'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```

⚠ **The prompt is NOT under the `generate_template` step's config**, which is where you will look
first. `generate_template.config` carries only `ai_service` and `input_fields`; the template is a
top-level key of `default_config`. Ten minutes lost to this.

## Which agents actually use these two actions (the blast radius)

```sql
SELECT type, s.key AS step, s.value->>'action' AS action
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' IN ('load_existing_component','store_generated_component');
```

`component-creator` only, both actions, as of 2026-08-25.

⚠ **A top-level `jsonb_each` misses steps nested inside a `sub_workflow`.** The
`bugfix_357_component_identity` RUNBOOK records exactly this: three of its six `save_page_sections`
steps were nested and a top-level scan did not see them. This query is honest for
`component-creator` (flat workflow, verified by listing its six step keys) and is NOT a general
census. Use the recursive version in that lane's runbook if you need one.

## Duplicate pairs born from the generated route (the damage this class produces)

```sql
SELECT section_type, count(*) AS rows_,
       string_agg(function || ' [' || COALESCE(created_from,'?') || ' ' || created_at::date || ']', ' | ' ORDER BY created_at)
FROM content_components
WHERE is_active AND forked_from IS NULL AND component_level='section' AND section_type IS NOT NULL
GROUP BY 1 HAVING count(*) > 1 ORDER BY 2 DESC, 1;
```

⚠ **`created_from` separates the finding from the noise.** The `hero` pool has 7 rows and looks
alarming; all 7 are `manual` and are deliberate vocabulary. Only two pairs are `generated`, and both
second rows carry `function == section_type` — the signature of the section_type fallback
derivation. Without the `created_from` column in the output this query reports 4 findings, of which
2 are real.
