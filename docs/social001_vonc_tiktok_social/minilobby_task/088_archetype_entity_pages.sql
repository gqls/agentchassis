-- 088_archetype_entity_pages.sql — vonc.com Approach A: build the archetype hub properly
-- Created 2026-07-12.
--
-- Context (design-discovery survey, 2026-07-12): archetype-grid's items field
-- sources query.pages_where_type:entity_page — a page_type that is (a) kebab-
-- forbidden by chk_page_type_kebab_case and (b) attached to zero pages, so the
-- grid has never rendered a card and the 8 archetype icon assets are orphaned.
--
-- This migration makes the design buildable with EXISTING machinery only:
--   1. 8 site_plan_pages rows       role entity-page under parent archetypes
--   2. 24 site_plan_sections rows   hero / content-block-about / call-to-action
--   3. 8 site_plan_imagery rows     page-scope hero keyed to each icon_<slug>
--      (content-block-about.image_src -> site_assets.image -> hero alias ->
--       per-page icon; consumes all 8 orphaned icons, zero Go changes)
--   4. 8 pages rows                 page-build-handler does NOT create missing
--                                   pages (check_page_found -> complete_error)
--   5. archetype-grid schema fix    source entity_page -> entity-page, limit 6->8
--   6. archetype-grid template fix  card link text {{.name}} (slug) -> {{.title}}
--
-- Then: reconcile_site_plan emits needs_page:<slug> x8 (triaged) + a terminal
-- reconcile_rerender. Park needs_page:provocation (Spark pipeline's page) back
-- to 'detected' before dispatching. After the 8 pages deploy, light-rerender
-- the archetypes page (reason section_data_resolved) so the grid query refills.
--
-- Reversal: restore from the _vonc_archetype_backup_20260712_* tables; DELETE
-- the inserted rows by the WHERE clauses below (plan_id + names/slugs).

BEGIN;

-- ── 0. Backups (fresh dated names — never reuse: CREATE IF NOT EXISTS no-ops) ──
CREATE TABLE _vonc_archetype_backup_20260712_plan_pages AS
  SELECT * FROM site_plan_pages WHERE plan_id='77493277-f510-47ea-aa27-8fca415743d6';
CREATE TABLE _vonc_archetype_backup_20260712_plan_sections AS
  SELECT * FROM site_plan_sections WHERE plan_id='77493277-f510-47ea-aa27-8fca415743d6';
CREATE TABLE _vonc_archetype_backup_20260712_plan_imagery AS
  SELECT * FROM site_plan_imagery WHERE plan_id='77493277-f510-47ea-aa27-8fca415743d6';
CREATE TABLE _vonc_archetype_backup_20260712_cc_aggrid AS
  SELECT * FROM content_components WHERE function='archetype-grid' AND is_active
   AND id IN (SELECT component_id FROM page_components pc JOIN pages p ON p.id=pc.page_id
              WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74');

-- ── 1. Plan pages (role entity-page, nested under archetypes) ──────────────
INSERT INTO site_plan_pages (plan_id, name, role, slug, url, parent_section,
                             in_header, in_footer, nav_order, title, nav_label, meta_description)
VALUES
 ('77493277-f510-47ea-aa27-8fca415743d6','surgeon','entity-page','surgeon','/archetypes/surgeon.html','archetypes',
  false,false,NULL,'The Surgeon','The Surgeon',
  'Precise, analytical, cuts to the core. Strips arguments to essential structure. Fast on logical frameworks. Strengths: analytical speed, structural clarity, finding hidden flaws.'),
 ('77493277-f510-47ea-aa27-8fca415743d6','wildcard','entity-page','wildcard','/archetypes/wildcard.html','archetypes',
  false,false,NULL,'The Wildcard','The Wildcard',
  'Unpredictable, highest remix rate. Takes ideas in unexpected directions. Crosses domains freely. Strengths: creative leaps, remix ability, cross-domain thinking.'),
 ('77493277-f510-47ea-aa27-8fca415743d6','oracle','entity-page','oracle','/archetypes/oracle.html','archetypes',
  false,false,NULL,'The Oracle','The Oracle',
  'Quiet but accurate, rare but impactful. Does not respond often but when they do it lands. Best prediction accuracy. Strengths: prediction accuracy, timing, selective engagement.'),
 ('77493277-f510-47ea-aa27-8fca415743d6','catalyst','entity-page','catalyst','/archetypes/catalyst.html','archetypes',
  false,false,NULL,'The Catalyst','The Catalyst',
  'Responses generate longest chains. Individual responses might not be best, but they spark others. Strengths: provocation, opening angles, generating momentum.'),
 ('77493277-f510-47ea-aa27-8fca415743d6','judge','entity-page','judge','/archetypes/judge.html','archetypes',
  false,false,NULL,'The Judge','The Judge',
  'Spectates, but reactions predict outcomes. Lives in the Gallery. Reaction pattern most predictive of final outcomes. Strengths: pattern recognition, taste, predictive judgement.'),
 ('77493277-f510-47ea-aa27-8fca415743d6','maker','entity-page','maker','/archetypes/maker.html','archetypes',
  false,false,NULL,'The Maker','The Maker',
  'Stage-dominant, consistent showcaser. Shows up with creations not opinions. Demonstrates skill. Strengths: creative output, consistency, skill demonstration.'),
 ('77493277-f510-47ea-aa27-8fca415743d6','scout','entity-page','scout','/archetypes/scout.html','archetypes',
  false,false,NULL,'The Scout','The Scout',
  'Identifies breakout creators early. The curator. Finds things others miss. Strengths: early identification, taste, discovery.'),
 ('77493277-f510-47ea-aa27-8fca415743d6','mentor','entity-page','mentor','/archetypes/mentor.html','archetypes',
  false,false,NULL,'The Mentor','The Mentor',
  'Breakdowns among most saved content. Explanations are teaching material. Turns admiration into learning. Strengths: teaching, explanation, knowledge sharing.');

-- ── 2. Plan sections: hero(0) / content-block-about(1) / call-to-action(2) ─
INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
SELECT '77493277-f510-47ea-aa27-8fca415743d6', s.slug, o.ord, o.comp
FROM (VALUES ('surgeon'),('wildcard'),('oracle'),('catalyst'),('judge'),('maker'),('scout'),('mentor')) AS s(slug)
CROSS JOIN (VALUES (0,'hero'),(1,'content-block-about'),(2,'call-to-action')) AS o(ord,comp);

-- ── 3. Page-scope hero imagery keyed to the existing icon assets ───────────
-- ensureAssets: scope='page' + scope_ref=<page name> + kind='hero' joined to
-- assets by key -> DeployedWebPath('icon_<slug>','icon') = /assets/images/icon-<slug>.jpg
-- (all 8 verified live HTTP 200 on 2026-07-12).
INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, ordering, source)
SELECT '77493277-f510-47ea-aa27-8fca415743d6', 'page', s.slug, 'icon_'||s.slug, 'hero',
       'Existing archetype icon asset, reused as the entity page''s image', 0, 'manual'
FROM (VALUES ('surgeon'),('wildcard'),('oracle'),('catalyst'),('judge'),('maker'),('scout'),('mentor')) AS s(slug);

-- ── 4. Pages rows (handler loads, never creates; reconcile rule 3 emits) ───
INSERT INTO pages (site_id, name, url, title, page_type, nav_label, in_header, in_footer,
                   meta_description, sections, build_status, status)
SELECT '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', spp.name, spp.url, spp.title, 'entity-page',
       spp.nav_label, false, false, spp.meta_description,
       '["hero","content-block-about","call-to-action"]'::jsonb, 'planned', 'active'
FROM site_plan_pages spp
WHERE spp.plan_id='77493277-f510-47ea-aa27-8fca415743d6' AND spp.role='entity-page'
ON CONFLICT (site_id, name) DO NOTHING;

-- ── 5+6. archetype-grid: schema re-point + limit + card-link polish ────────
UPDATE content_components
SET input_schema = jsonb_set(
      jsonb_set(input_schema, '{fields,items,source}', '"query.pages_where_type:entity-page"'),
      '{fields,items,limit}', '8'),
    html_template = replace(html_template,
      '<a class="ag-card-link" href="{{.url}}">{{.name}} <span aria-hidden="true">&rarr;</span></a>',
      '<a class="ag-card-link" href="{{.url}}">{{.title}} <span aria-hidden="true">&rarr;</span></a>'),
    updated_at = NOW()
WHERE function='archetype-grid' AND is_active
  AND id IN (SELECT component_id FROM page_components pc JOIN pages p ON p.id=pc.page_id
             WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74');

-- ── Verify before COMMIT ────────────────────────────────────────────────────
DO $$
DECLARE pp INT; ps INT; pi INT; pg INT; src TEXT; lim INT; tmpl_ok BOOLEAN;
BEGIN
  SELECT COUNT(*) INTO pp FROM site_plan_pages
   WHERE plan_id='77493277-f510-47ea-aa27-8fca415743d6' AND role='entity-page';
  SELECT COUNT(*) INTO ps FROM site_plan_sections
   WHERE plan_id='77493277-f510-47ea-aa27-8fca415743d6'
     AND page_name IN ('surgeon','wildcard','oracle','catalyst','judge','maker','scout','mentor');
  SELECT COUNT(*) INTO pi FROM site_plan_imagery
   WHERE plan_id='77493277-f510-47ea-aa27-8fca415743d6' AND scope='page' AND key LIKE 'icon\_%';
  SELECT COUNT(*) INTO pg FROM pages
   WHERE site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND page_type='entity-page';
  SELECT input_schema#>>'{fields,items,source}', (input_schema#>>'{fields,items,limit}')::int,
         html_template LIKE '%{{.title}} <span aria-hidden="true">&rarr;</span>%'
    INTO src, lim, tmpl_ok
  FROM content_components WHERE function='archetype-grid' AND is_active
   AND id IN (SELECT component_id FROM page_components pc JOIN pages p ON p.id=pc.page_id
              WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74');
  IF pp<>8 OR ps<>24 OR pi<>8 OR pg<>8
     OR src<>'query.pages_where_type:entity-page' OR lim<>8 OR NOT tmpl_ok THEN
    RAISE EXCEPTION 'verify failed: pp=% ps=% pi=% pg=% src=% lim=% tmpl=%', pp,ps,pi,pg,src,lim,tmpl_ok;
  END IF;
  RAISE NOTICE 'verified: 8 plan pages, 24 sections, 8 imagery, 8 pages, schema+template updated';
END $$;

COMMIT;
