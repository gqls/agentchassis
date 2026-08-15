-- CENSUS (read-only) — predicted router routes for every open required_fields_missing item.
-- v3 (round-2 council refinements, both measured): asset_sourced — still-empty fields whose
-- template schema declares source site_assets.* must NOT go to the prose writer (the first
-- canary conversion proved it: validate_content refused minted URLs); no_plan_unbuildable —
-- sectionless pages whose page_type has no generic archetype park instead of converting (the
-- blog-index canary's recreate no-opped at mark_no_ready_sections). Matches seed 410 v3.
--
-- This is the dry run for required-fields-missing-handler (bugs_open/277): the CASE below is
-- the same classification the router's classify step runs per item, applied over the whole
-- open population. Run BEFORE any assignment; save the output beside this file; re-run after
-- assignment — every predicted_route should have become the recorded route (or terminal).
--
-- Resolution key is (page_name, slot_name) — the revalidator's key — NEVER spec.component_id.
-- ⚠ When the seed's classifier changes, change THIS file in the same edit, and re-prove the
-- seed's own embedded string against known rows (RUNBOOK) — the census passing proves the
-- census, not the seed.
--
-- Run:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -f - < CENSUS_2026-08-15_predicted_routes.sql

WITH items AS (
  SELECT wi.id, wi.site_id, wi.status, wi.spec, s.domain
  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
  WHERE wi.item_type = 'required_fields_missing'
    AND wi.status IN ('needs_human_review','unresolved')
),
enriched AS (
  SELECT i.id, i.domain, i.status,
    i.spec->>'page_name' AS page_name,
    COALESCE(i.spec->>'slot_name','') AS slot_name,
    jsonb_typeof(i.spec->'missing_fields') AS mf_type,
    p.id AS pg_id,
    COALESCE(p.page_type,'') AS page_type,
    COALESCE(p.rebuild_policy,'generic') AS rebuild_policy,
    (p.sections IS NULL OR p.sections = '[]'::jsonb) AS sections_empty,
    EXISTS (SELECT 1 FROM site_plan_sections sps
              JOIN site_plans sp ON sp.id = sps.plan_id
             WHERE sp.site_id = i.site_id AND sp.is_current = true
               AND sps.page_name = i.spec->>'page_name') AS has_plan_sections,
    pc.id AS comp_id,
    pc.content_data,
    length(COALESCE(pc.rendered_html,'')) AS html_len,
    (pc.locked_at IS NOT NULL) AS locked,
    fs.n_named, fs.n_still_empty, fs.n_asset_empty
  FROM items i
  LEFT JOIN pages p
    ON p.site_id = i.site_id AND p.name = i.spec->>'page_name'
   AND COALESCE(p.status,'') <> 'deleted'
  LEFT JOIN LATERAL (
    SELECT pc2.id, pc2.content_data, pc2.rendered_html, pc2.locked_at, cc.input_schema AS sch
    FROM page_components pc2
    LEFT JOIN content_components cc ON cc.id = pc2.component_id
    WHERE pc2.page_id = p.id AND pc2.build_status = 'deployed'
      AND COALESCE(pc2.slot_name,'') = COALESCE(i.spec->>'slot_name','')
    ORDER BY pc2.updated_at DESC NULLS LAST LIMIT 1
  ) pc ON true
  LEFT JOIN LATERAL (
    SELECT count(*) AS n_named,
           count(*) FILTER (WHERE x.is_empty) AS n_still_empty,
           count(*) FILTER (WHERE x.is_empty AND x.src LIKE 'site_assets.%') AS n_asset_empty
    FROM (
      SELECT f.name,
             (pc.id IS NULL OR pc.content_data IS NULL
              OR pc.content_data->f.name IS NULL
              OR jsonb_typeof(pc.content_data->f.name) = 'null'
              OR (jsonb_typeof(pc.content_data->f.name) = 'string'
                  AND btrim(pc.content_data->>f.name) = '')
              OR (jsonb_typeof(pc.content_data->f.name) = 'array'
                  AND jsonb_array_length(pc.content_data->f.name) = 0)
              OR (jsonb_typeof(pc.content_data->f.name) = 'object'
                  AND pc.content_data->f.name = '{}'::jsonb)) AS is_empty,
             COALESCE(pc.sch->'fields'->f.name->>'source','llm') AS src
      FROM jsonb_array_elements_text(COALESCE(i.spec->'missing_fields','[]'::jsonb)) f(name)
    ) x
  ) fs ON true
),
routed AS (
  SELECT *,
    CASE
      WHEN page_name IS NULL OR mf_type IS DISTINCT FROM 'array' OR n_named = 0 THEN 'malformed'
      WHEN pg_id IS NULL OR comp_id IS NULL OR locked THEN 'stale'
      WHEN n_still_empty = 0 THEN 'resolved'
      WHEN sections_empty AND NOT has_plan_sections
           AND (page_type IN ('tool','game') OR rebuild_policy = 'owned') THEN 'no_plan_owned'
      WHEN content_data IS NULL OR content_data = '{}'::jsonb THEN 'no_content_data'
      WHEN n_asset_empty > 0 THEN 'asset_sourced'
      WHEN sections_empty AND NOT has_plan_sections
           AND page_type IN ('','content','landing') THEN 'no_plan_generic'
      WHEN sections_empty AND NOT has_plan_sections THEN 'no_plan_unbuildable'
      ELSE 'partial'
    END AS predicted_route
  FROM enriched
)
SELECT predicted_route, count(*) FROM routed GROUP BY 1 ORDER BY 2 DESC;

WITH items AS (
  SELECT wi.id, wi.site_id, wi.status, wi.spec, s.domain
  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
  WHERE wi.item_type = 'required_fields_missing'
    AND wi.status IN ('needs_human_review','unresolved')
),
enriched AS (
  SELECT i.id, i.domain, i.status,
    i.spec->>'page_name' AS page_name,
    COALESCE(i.spec->>'slot_name','') AS slot_name,
    jsonb_typeof(i.spec->'missing_fields') AS mf_type,
    p.id AS pg_id,
    COALESCE(p.page_type,'') AS page_type,
    COALESCE(p.rebuild_policy,'generic') AS rebuild_policy,
    (p.sections IS NULL OR p.sections = '[]'::jsonb) AS sections_empty,
    EXISTS (SELECT 1 FROM site_plan_sections sps
              JOIN site_plans sp ON sp.id = sps.plan_id
             WHERE sp.site_id = i.site_id AND sp.is_current = true
               AND sps.page_name = i.spec->>'page_name') AS has_plan_sections,
    pc.id AS comp_id,
    pc.content_data,
    length(COALESCE(pc.rendered_html,'')) AS html_len,
    (pc.locked_at IS NOT NULL) AS locked,
    fs.n_named, fs.n_still_empty, fs.n_asset_empty
  FROM items i
  LEFT JOIN pages p
    ON p.site_id = i.site_id AND p.name = i.spec->>'page_name'
   AND COALESCE(p.status,'') <> 'deleted'
  LEFT JOIN LATERAL (
    SELECT pc2.id, pc2.content_data, pc2.rendered_html, pc2.locked_at, cc.input_schema AS sch
    FROM page_components pc2
    LEFT JOIN content_components cc ON cc.id = pc2.component_id
    WHERE pc2.page_id = p.id AND pc2.build_status = 'deployed'
      AND COALESCE(pc2.slot_name,'') = COALESCE(i.spec->>'slot_name','')
    ORDER BY pc2.updated_at DESC NULLS LAST LIMIT 1
  ) pc ON true
  LEFT JOIN LATERAL (
    SELECT count(*) AS n_named,
           count(*) FILTER (WHERE x.is_empty) AS n_still_empty,
           count(*) FILTER (WHERE x.is_empty AND x.src LIKE 'site_assets.%') AS n_asset_empty
    FROM (
      SELECT f.name,
             (pc.id IS NULL OR pc.content_data IS NULL
              OR pc.content_data->f.name IS NULL
              OR jsonb_typeof(pc.content_data->f.name) = 'null'
              OR (jsonb_typeof(pc.content_data->f.name) = 'string'
                  AND btrim(pc.content_data->>f.name) = '')
              OR (jsonb_typeof(pc.content_data->f.name) = 'array'
                  AND jsonb_array_length(pc.content_data->f.name) = 0)
              OR (jsonb_typeof(pc.content_data->f.name) = 'object'
                  AND pc.content_data->f.name = '{}'::jsonb)) AS is_empty,
             COALESCE(pc.sch->'fields'->f.name->>'source','llm') AS src
      FROM jsonb_array_elements_text(COALESCE(i.spec->'missing_fields','[]'::jsonb)) f(name)
    ) x
  ) fs ON true
)
SELECT left(id::text,8) AS item, domain, page_name, slot_name,
  CASE
    WHEN page_name IS NULL OR mf_type IS DISTINCT FROM 'array' OR n_named = 0 THEN 'malformed'
    WHEN pg_id IS NULL OR comp_id IS NULL OR locked THEN 'stale'
    WHEN n_still_empty = 0 THEN 'resolved'
    WHEN sections_empty AND NOT has_plan_sections
         AND (page_type IN ('tool','game') OR rebuild_policy = 'owned') THEN 'no_plan_owned'
    WHEN content_data IS NULL OR content_data = '{}'::jsonb THEN 'no_content_data'
    WHEN n_asset_empty > 0 THEN 'asset_sourced'
    WHEN sections_empty AND NOT has_plan_sections
         AND page_type IN ('','content','landing') THEN 'no_plan_generic'
    WHEN sections_empty AND NOT has_plan_sections THEN 'no_plan_unbuildable'
    ELSE 'partial'
  END AS predicted_route,
  n_named, n_still_empty, n_asset_empty, html_len, page_type, rebuild_policy,
  sections_empty, has_plan_sections
FROM enriched
ORDER BY 5, 2, 3;
