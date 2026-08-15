-- CENSUS (read-only) — predicted router routes for every open required_fields_missing item.
-- This is the dry run for the required-fields-missing-handler (bugs_open/277, seed 410):
-- the CASE below is the same classification the router's classify step runs per item,
-- applied over the whole open population at once. Run it BEFORE any assignment, save the
-- output beside this file, and re-run after the fleet assignment — every predicted_route
-- should have become the recorded route (or the row should be terminal).
--
-- Resolution key is (page_name, slot_name) — the revalidator's key — NEVER spec.component_id
-- (016b: component ids are unstable across rerenders; 11/45 resolved to nothing when keyed so).
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
    fs.n_named, fs.n_still_empty
  FROM items i
  LEFT JOIN pages p
    ON p.site_id = i.site_id AND p.name = i.spec->>'page_name'
   AND COALESCE(p.status,'') <> 'deleted'
  LEFT JOIN LATERAL (
    SELECT pc2.id, pc2.content_data, pc2.rendered_html, pc2.locked_at
    FROM page_components pc2
    WHERE pc2.page_id = p.id AND pc2.build_status = 'deployed'
      AND COALESCE(pc2.slot_name,'') = COALESCE(i.spec->>'slot_name','')
    ORDER BY pc2.updated_at DESC NULLS LAST LIMIT 1
  ) pc ON true
  LEFT JOIN LATERAL (
    SELECT count(*) AS n_named,
           count(*) FILTER (WHERE pc.id IS NULL OR pc.content_data IS NULL
             OR pc.content_data->f.name IS NULL
             OR jsonb_typeof(pc.content_data->f.name) = 'null'
             OR (jsonb_typeof(pc.content_data->f.name) = 'string'
                 AND btrim(pc.content_data->>f.name) = '')
             OR (jsonb_typeof(pc.content_data->f.name) = 'array'
                 AND jsonb_array_length(pc.content_data->f.name) = 0)
             OR (jsonb_typeof(pc.content_data->f.name) = 'object'
                 AND pc.content_data->f.name = '{}'::jsonb)
           ) AS n_still_empty
    FROM jsonb_array_elements_text(COALESCE(i.spec->'missing_fields','[]'::jsonb)) f(name)
  ) fs ON true
),
routed AS (
  SELECT *,
    CASE
      WHEN page_name IS NULL OR mf_type IS DISTINCT FROM 'array' THEN 'malformed'
      WHEN pg_id IS NULL OR comp_id IS NULL OR locked THEN 'stale'
      WHEN n_still_empty = 0 THEN 'resolved'
      WHEN sections_empty AND NOT has_plan_sections
           AND (page_type IN ('tool','game') OR rebuild_policy = 'owned') THEN 'no_plan_owned'
      WHEN content_data IS NULL OR content_data = '{}'::jsonb THEN 'no_content_data'
      WHEN sections_empty AND NOT has_plan_sections THEN 'no_plan_generic'
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
    fs.n_named, fs.n_still_empty
  FROM items i
  LEFT JOIN pages p
    ON p.site_id = i.site_id AND p.name = i.spec->>'page_name'
   AND COALESCE(p.status,'') <> 'deleted'
  LEFT JOIN LATERAL (
    SELECT pc2.id, pc2.content_data, pc2.rendered_html, pc2.locked_at
    FROM page_components pc2
    WHERE pc2.page_id = p.id AND pc2.build_status = 'deployed'
      AND COALESCE(pc2.slot_name,'') = COALESCE(i.spec->>'slot_name','')
    ORDER BY pc2.updated_at DESC NULLS LAST LIMIT 1
  ) pc ON true
  LEFT JOIN LATERAL (
    SELECT count(*) AS n_named,
           count(*) FILTER (WHERE pc.id IS NULL OR pc.content_data IS NULL
             OR pc.content_data->f.name IS NULL
             OR jsonb_typeof(pc.content_data->f.name) = 'null'
             OR (jsonb_typeof(pc.content_data->f.name) = 'string'
                 AND btrim(pc.content_data->>f.name) = '')
             OR (jsonb_typeof(pc.content_data->f.name) = 'array'
                 AND jsonb_array_length(pc.content_data->f.name) = 0)
             OR (jsonb_typeof(pc.content_data->f.name) = 'object'
                 AND pc.content_data->f.name = '{}'::jsonb)
           ) AS n_still_empty
    FROM jsonb_array_elements_text(COALESCE(i.spec->'missing_fields','[]'::jsonb)) f(name)
  ) fs ON true
)
SELECT left(id::text,8) AS item, domain, page_name, slot_name,
  CASE
    WHEN page_name IS NULL OR mf_type IS DISTINCT FROM 'array' THEN 'malformed'
    WHEN pg_id IS NULL OR comp_id IS NULL OR locked THEN 'stale'
    WHEN n_still_empty = 0 THEN 'resolved'
    WHEN sections_empty AND NOT has_plan_sections
         AND (page_type IN ('tool','game') OR rebuild_policy = 'owned') THEN 'no_plan_owned'
    WHEN content_data IS NULL OR content_data = '{}'::jsonb THEN 'no_content_data'
    WHEN sections_empty AND NOT has_plan_sections THEN 'no_plan_generic'
    ELSE 'partial'
  END AS predicted_route,
  n_named, n_still_empty, html_len, page_type, rebuild_policy, sections_empty, has_plan_sections
FROM enriched
ORDER BY 5, 2, 3;
