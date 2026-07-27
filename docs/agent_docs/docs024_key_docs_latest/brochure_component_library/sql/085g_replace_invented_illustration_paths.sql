-- 085g_replace_invented_illustration_paths.sql
--
-- model-fine-tuning's image-hover-card-grid carries six card images under
-- /images/illustrations/<name>.svg — a path convention that exists nowhere on
-- this site or any other (the site's images live at /assets/images/). All six
-- return 404. They were written into content_data as plausible-looking
-- filenames; nothing generated or deployed them, and nothing checked.
--
-- Found by scripts/render_audit.py, which renders the page and asks the browser
-- which images failed, then re-checks each one over HTTP before reporting it.
-- The DB-side image_url_404 check cannot see these: it matches
-- /assets/images/... references against the assets table, and these are neither.
--
-- Repointed at generated assets that exist and are on-brief (line illustration,
-- consistent tint). Verified 200 each before writing.

BEGIN;

UPDATE page_components pc
   SET content_data = jsonb_set(pc.content_data, '{cards}', (
         SELECT jsonb_agg(
                  CASE
                    WHEN card->>'image' LIKE '/images/illustrations/%'
                    THEN jsonb_set(card, '{image}', to_jsonb(
                           '/assets/images/' || CASE
                             WHEN card->>'image' LIKE '%fine-tuning%'   THEN 'hero-fine-tuning'
                             WHEN card->>'image' LIKE '%review-council%' THEN 'hero-review-council'
                             WHEN card->>'image' LIKE '%verification%'   THEN 'hero-about'
                             WHEN card->>'image' LIKE '%vector-search%'  THEN 'brand-illustration'
                             WHEN card->>'image' LIKE '%decision-record%' THEN 'hero-capabilities'
                             ELSE 'hero-home'
                           END || '.jpg'))
                    ELSE card
                  END ORDER BY ord)
         FROM jsonb_array_elements(pc.content_data->'cards') WITH ORDINALITY AS t(card, ord)),
        true),
       updated_at = NOW()
  FROM content_components cc, pages p
 WHERE cc.id = pc.component_id
   AND p.id = pc.page_id
   AND p.site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
   AND cc.function = 'image-hover-card-grid'
   AND pc.content_data::text LIKE '%/images/illustrations/%';

-- Re-render the page so the change reaches the served HTML.
INSERT INTO site_work_items (
  site_id, item_type, item_key, status, pipeline, summary, spec,
  handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
SELECT p.site_id, 'page_rerender',
       'page_rerender_model-fine-tuning_199733a8_illustration_paths_20260727',
       'triaged', 'build',
       'Republish model-fine-tuning: six invented /images/illustrations/*.svg paths repointed at real assets',
       jsonb_build_object('domain','fundamentallyai.com','page_id',p.id::text,
                          'page_name','model-fine-tuning','filename','model-fine-tuning.html',
                          'reason','section_data_resolved'),
       'page-rerender','operator:brochure_component_library','operator:brochure_component_library',
       0, 3, NOW(), NOW()
  FROM pages p
 WHERE p.site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND p.name = 'model-fine-tuning';

COMMIT;
