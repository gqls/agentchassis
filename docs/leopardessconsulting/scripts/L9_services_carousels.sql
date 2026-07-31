-- L9 — /services.html: both blocks become carousels (2026-07-31)
--
-- HOW TO LOAD A TEMPLATE OR SNIPPET, because the documented recipe is wrong in
-- one detail. Build the statement LOCALLY with a dollar-quoted literal (python),
-- pipe the finished SQL to psql, then assert on md5 — NOT on `length()`.
-- `length()` counts CHARACTERS and `wc -c` counts BYTES, so on any template with
-- an em dash in it the documented "byte counts must match exactly" check FAILS on
-- a byte-perfect load and looks exactly like truncation. Use octet_length(), or
-- better md5(), against md5sum of the local file. (LANDMINES 2026-07-31.)
--
--   psql -c "SELECT octet_length(html_template), md5(html_template) FROM ..."
--   wc -c file && md5sum file
--
-- Template body lives beside this file:
--   L9_info_card_grid_template_with_carousel.html
-- Byte-identity harness (run it BEFORE installing a shared-component template):
--   prove_component_template_identity.go   -- 18/18 identical + a mutant control
--
-- ORDER MATTERS: image first (assets must be `active` and serving 200 before the
-- content_data referencing them is rendered), then config, then ONE rerender.

\set ON_ERROR_STOP on

-- ---------------------------------------------------------------------------
-- 0. Backups. Owner rule for this site: bak_* before ANY change, incl. forks
--    and the site_plan aspect.
-- ---------------------------------------------------------------------------
-- CREATE TABLE bak_leo_services_pc_20260731        AS SELECT * FROM page_components WHERE id IN (...);
-- CREATE TABLE bak_cc_info_card_grid_20260731      AS SELECT * FROM content_components WHERE id='fc56f085-8e9a-4f6b-8e8d-600f9a1381e2';
-- CREATE TABLE bak_js_hero_card_carousel_20260731  AS SELECT * FROM js_snippets WHERE name='hero-card-carousel';
-- CREATE TABLE bak_leo_sitespec_siteplan_20260731  AS SELECT * FROM site_specs WHERE id='8439e6b2-e671-44c1-a63a-c90175ed59c6';
-- CREATE TABLE bak_leo_services_cta_20260731       AS SELECT * FROM page_components WHERE page_id='ebc2c413-...' AND slot_name='call-to-action';
-- NOTE page_components.id is NOT STABLE across rerenders — the ids above were
-- already dead by the end of the session. Back up and re-query by (page_id, slot_name).

-- ---------------------------------------------------------------------------
-- 1. Block B: turn the shared info-card-grid into a carousel FOR THIS INSTANCE
--    ONLY, via an opt-in flag. Do NOT fork the component: RUNBOOK landmine 14 —
--    a section-component fork does not survive rerender, save_page_sections
--    re-links to the canonical component by `function`.
-- ---------------------------------------------------------------------------
-- (a) html_template  <- L9_info_card_grid_template_with_carousel.html  (md5-asserted)
-- (b) input_schema.fields.carousel documented as {type:boolean,required:false,source:config}
-- (c) the instance opts in, and its six dead /services/*.html links are repointed:
UPDATE page_components SET content_data = content_data
    || '{"carousel": true}'::jsonb
 WHERE page_id = 'ebc2c413-61e2-465e-b22b-9aab0167abc9' AND slot_name = 'info-card-grid';
-- link_url values set to live pages (all verified 200 BEFORE writing them):
--   monitoring/entity-verification/tool-generation -> /case-studies.html
--   orchestration/audit-trail                     -> /technical-architecture.html
--   human-approval                                -> /how-it-works.html
-- WHY they were "broken": they pointed at six /services/*.html pages that never
-- existed, and RepairPageLinks unlinks a dead anchor while KEEPING its inner text
-- — which is the label — so the page showed "See how it works ->" as dead prose.

-- ---------------------------------------------------------------------------
-- 2. The arrows. Reuse the existing snippet; do not write a second carousel.
--    Widen ONE selector line and let the bundler ship it.
-- ---------------------------------------------------------------------------
-- js_snippets['hero-card-carousel']:
--   initAll selector += ", [data-hcc-carousel]"
UPDATE js_snippets SET applies_to = '["hero-card-carousel", "info-card-grid"]'::jsonb
 WHERE name = 'hero-card-carousel';
-- Then regenerate the bundle — WITHOUT this the arrows are inert and the track
-- only scrolls by swipe/trackpad. The site's snippets.js was 334 bytes /
-- "0 active snippet(s)" before this:
--   ./orchestrate_safe.sh site-asset-renderer \
--     '{"site_id":"4851f6fc-...","domain":"leopardessconsulting.co.uk"}'

-- ---------------------------------------------------------------------------
-- 3. Block A: services-grid -> teaser-reveal-panel. A PLACEMENT LIVES IN THREE
--    PLACES and a `complete` rerender drops a section missing from any of them.
--    On this site site_plan_sections holds ZERO rows, so the site_specs aspect
--    is the real authority — verify that before trusting either.
-- ---------------------------------------------------------------------------
UPDATE page_components
   SET component_id = '22c12251-73aa-4232-bd67-ef9edcfe8061', slot_name = 'teaser-reveal-panel'
 WHERE page_id = 'ebc2c413-61e2-465e-b22b-9aab0167abc9' AND slot_name = 'services-grid';
UPDATE pages SET sections =
   '["hero-services","teaser-reveal-panel","info-card-grid","call-to-action"]'::jsonb
 WHERE id = 'ebc2c413-61e2-465e-b22b-9aab0167abc9';
UPDATE site_specs SET data = jsonb_set(data, '{pages,1,sections}',
   '["hero-services","teaser-reveal-panel","info-card-grid","call-to-action"]'::jsonb)
 WHERE id = '8439e6b2-e671-44c1-a63a-c90175ed59c6';   -- is_current site_plan; index 1 = services

-- ---------------------------------------------------------------------------
-- 4. PRE-RERENDER GATE. Run BOTH branches or the whole page escalates to the
--    LLM writer and hand-authored copy is rewritten.
-- ---------------------------------------------------------------------------
-- (a) every slot must be a non-empty OBJECT:
SELECT slot_name, jsonb_typeof(content_data) AS t
  FROM page_components WHERE page_id = 'ebc2c413-61e2-465e-b22b-9aab0167abc9'
   AND (content_data IS NULL OR jsonb_typeof(content_data) <> 'object' OR content_data = '{}'::jsonb);
-- (b) every required source:"llm" schema field must be present:
WITH slots AS (SELECT pc.slot_name, pc.content_data, cc.input_schema
                 FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
                WHERE pc.page_id = 'ebc2c413-61e2-465e-b22b-9aab0167abc9'),
     req AS (SELECT s.slot_name, f.key AS field FROM slots s, jsonb_each(s.input_schema->'fields') f
              WHERE f.value->>'source' = 'llm' AND COALESCE((f.value->>'required')::boolean, false))
SELECT r.slot_name, r.field FROM req r JOIN slots s USING (slot_name) WHERE NOT (s.content_data ? r.field);
-- Both must return 0 rows. Then:
--   ./rerender_page_safe.sh 4851f6fc-... leopardessconsulting.co.uk services

-- ---------------------------------------------------------------------------
-- 5. While here: the CTA on this page had NO BUTTONS. Both labels were set and
--    neither URL was, and the template gates on {{if and .primary_cta
--    .primary_cta_url}} — so it drew nothing and read as a deliberate design.
--    Also carried the third prose instance of the phantom gap-finder URL.
-- ---------------------------------------------------------------------------
UPDATE page_components SET content_data = content_data || jsonb_build_object(
        'primary_cta_url',   '/contact.html',
        'secondary_cta_url', '/tools/ai-agent-roi-estimator.html')
 WHERE page_id = 'ebc2c413-61e2-465e-b22b-9aab0167abc9' AND slot_name = 'call-to-action';

-- ---------------------------------------------------------------------------
-- 6. VERIFY AT THE ARTEFACT. Poll on MARKUP, never on a class name — a class
--    matches the component's own inline <style> and a poll on it returns
--    instantly against a stale page (this cost a wasted verification round):
--   until curl -s "https://.../services.html?cb=$(date +%s%N)" \
--     | grep -q '<a href="/contact.html" class="cta-btn cta-btn-primary">'; do sleep 15; done
-- Then drive a REAL click with --force-prefers-reduced-motion (a smooth scroll
-- under --virtual-time-budget never advances, and reports a working carousel as
-- dead) and re-run against a mutant to prove the check can still fail.
