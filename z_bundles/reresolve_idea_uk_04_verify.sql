-- =====================================================================
-- idea.uk composition RE-RESOLVE — STEP 4 of 4: VERIFY
-- site_id = 1244516d-014d-421c-88c6-090bb1e9552a
-- Run after the planner (and webdesign-agent) have finished.
-- =====================================================================

-- 1) the new chain: layout_name should be tool-portal-light, scheme light
\echo '--- new chain (expect layout_name = tool-portal-light, layout_scheme = light) ---'
SELECT s.style_collection_id, sc.css_theme_id, t.layout_id,
       l.name AS layout_name, l.scheme AS layout_scheme
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes t         ON t.id = sc.css_theme_id
JOIN layouts l            ON l.id = t.layout_id
WHERE s.id = '1244516d-014d-421c-88c6-090bb1e9552a';

-- 2) the fresh resolved_composition spec: layout + palette source/values
\echo '--- resolved_composition spec (expect layout tool-portal-light + parchment palette, palette_source design_intent_values) ---'
SELECT data->>'layout_name'                          AS layout_name,
       data->'lineage'->>'palette_source'            AS palette_source,
       data->'palette'->'reference_values'->>'background' AS bg,
       data->'palette'->'reference_values'->>'accent'     AS accent,
       data->'palette'->'reference_values'->>'text'       AS text
FROM site_specs
WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND aspect = 'resolved_composition' AND is_current;

-- 3) the scheme-gap item should NOT have been queued this run
\echo '--- needs_new_layout_candidate should be absent for this run (a same-scheme layout matched) ---'
SELECT id, status, item_type, created_at, spec->>'applied_layout' AS applied_layout
FROM site_work_items
WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND item_type = 'needs_new_layout_candidate'
ORDER BY created_at DESC LIMIT 3;

-- 4) (outside SQL) confirm webdesign-agent re-rendered styles.css for the build
--    output (the B2 deploy), and that the page now reads light/parchment.
