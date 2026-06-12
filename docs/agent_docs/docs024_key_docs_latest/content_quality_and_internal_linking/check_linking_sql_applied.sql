-- Which linking SQLs are already applied? Run all four blocks; expected values
-- in comments. Any 'f' / wrong value = that SQL (or one statement in it) has
-- not been applied — rerun just that file (all are snapshot-guarded/idempotent
-- in effect; replace() no-ops if already replaced, jsonb_set re-sets the same).

-- 0) Snapshot tables (each SQL creates its own before editing)
SELECT
  to_regclass('public.content_components_bak_cta0610')     IS NOT NULL AS step1_snapshot,
  to_regclass('public.content_components_bak_navfix_0610') IS NOT NULL AS layer1b_snapshot,
  to_regclass('public.content_components_bak_hubfix_0610') IS NOT NULL AS b4b5_snapshot;

-- 1) step1_hero_cta_phantom_fix.sql — expect per row:
--    primary_on_missing='skip_field', secondary_on_missing='skip_field',
--    fallbacks_gone=t, no_contact_literal=t, no_services_literal=t,
--    no_features_literal=t, has_and_gate=t
SELECT name,
  COALESCE(input_schema#>>'{fields,cta_url,on_missing}',
           input_schema#>>'{fields,primary_cta_url,on_missing}') AS primary_on_missing,
  input_schema#>>'{fields,secondary_cta_url,on_missing}'         AS secondary_on_missing,
  (input_schema#>'{fields,cta_url,fallback}'         IS NULL AND
   input_schema#>'{fields,primary_cta_url,fallback}' IS NULL AND
   input_schema#>'{fields,secondary_cta_url,fallback}' IS NULL)  AS fallbacks_gone,
  html_template NOT LIKE '%/contact.html%'                        AS no_contact_literal,
  html_template NOT LIKE '%/services.html%'                       AS no_services_literal,
  html_template NOT LIKE '%#features%'                            AS no_features_literal,
  html_template LIKE '%{{if and %'                                AS has_and_gate
FROM content_components
WHERE name IN ('hero', 'call-to-action')
ORDER BY name;

-- 2) layer1b_header_footer_phantom_fix.sql — expect:
--    header-bold-gradient: header_cta_gated=t
--    footer-4-column: footer_literals_gone=t, footer_legal_data_driven=t
SELECT name,
  CASE WHEN name = 'header-bold-gradient'
       THEN html_template LIKE '%{{if .cta_url}}%' END            AS header_cta_gated,
  CASE WHEN name = 'footer-4-column'
       THEN html_template NOT LIKE '%/privacy.html%'
        AND html_template NOT LIKE '%/terms.html%' END            AS footer_literals_gone,
  CASE WHEN name = 'footer-4-column'
       THEN html_template LIKE '%{{range .legal_links}}%' END     AS footer_legal_data_driven
FROM content_components
WHERE name IN ('header-bold-gradient', 'footer-4-column')
ORDER BY name;

-- 3+4) b4_b5_hub_links_schema.sql + b4_b5_hub_links_template_gate.sql — expect:
--    cta_source='query.section_index_for:<tool|game|guide>',
--    fallback_gone=t, browse_all_gated=t
SELECT name,
  input_schema#>>'{fields,cta_url,source}'                        AS cta_source,
  input_schema#>'{fields,cta_url,fallback}' IS NULL               AS fallback_gone,
  html_template LIKE '%{{if .cta_url}}<a%'                        AS browse_all_gated
FROM content_components
WHERE name IN ('tool-list', 'game-list_pre_037', 'guide-list_pre_037')
ORDER BY name;
