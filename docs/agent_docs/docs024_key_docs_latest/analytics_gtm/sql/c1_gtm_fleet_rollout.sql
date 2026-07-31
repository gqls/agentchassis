-- c1_gtm_fleet_rollout.sql — Phase C: GTM-PQ3WCTBD on the remaining 13 deployed domains,
-- and retirement of the competing gtag.js seam.
--
-- OWNER DECISION 2026-07-31: one GTM container for the whole estate (per-domain reporting
-- comes off the hostname dimension); retire `analytics_id` rather than run two seams.
--
-- WHAT THIS TOUCHES
--   2 head components  : head-seo-standard (4 domains), webdesign.co.uk Document Head (1)
--                        — `Document Head` (9 domains) was already done by p4_34
--   5 header components: header-bold-gradient (7), header-professional-dark (3),
--                        header-leopardess (1), header-theme-chrome (1),
--                        webdesign.co.uk Site Header (1)  — `site-header` done by p4_34
--   13 site_specs rows, 26 rendered artefacts (13 sites x head+header)
--
-- FOUR TRAPS THIS FILE EXISTS TO NOT FALL INTO
--   1. input_schema has TWO live shapes. `Document Head` is FLAT; every component here is
--      WRAPPED ({"fields":{...}}) except `header-theme-chrome`, whose input_schema is
--      NULL. render_site_components_action.go:604-607 handles both, but the field must go
--      in the right place or the gap-fill never sees it.
--   2. `webdesign.co.uk Document Head` uses a LOWERCASE <meta charset="utf-8">. replace()
--      is case-sensitive, so a single uppercase-anchored UPDATE would touch 0 rows there
--      and still report success. Both anchors are handled, and asserted per site.
--   3. Template-only is inert and artefact-only is temporary (bugs_open/117) — both are
--      written, as in p4_34.
--   4. Retiring analytics_id is safe ONLY because it is dormant: 0 sites set the key
--      (verified below, and the DO block aborts if that has changed).
--
-- NOT FIXED HERE, DELIBERATELY: `webdesign.co.uk Document Head` emits no <head> open tag
-- at all (it starts at <meta charset>), so that site's pages have an implicit head. GTM
-- lands correctly inside it, but the missing tag is a pre-existing defect on 99 pages and
-- fixing it is a separate change with its own blast radius.
--
-- STILL REQUIRED AFTER THIS FILE: every deployed page has chrome baked in and must be
-- re-assembled — scripts/fire_reassemble_site.sh <domain>.

BEGIN;

-- ── 0. Snippets, defined once ──────────────────────────────────────────────────
CREATE TEMP TABLE _gtm(k text PRIMARY KEY, v text) ON COMMIT DROP;
INSERT INTO _gtm VALUES
('tpl_head', $q$
{{if .gtm_container_id}}<!-- Google Tag Manager -->
<script>(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
})(window,document,'script','dataLayer','{{.gtm_container_id}}');</script>
<!-- End Google Tag Manager -->{{end}}$q$),
('tpl_body', $q$'{{if .gtm_container_id}}<!-- Google Tag Manager (noscript) -->
<noscript><iframe src="https://www.googletagmanager.com/ns.html?id={{.gtm_container_id}}"
height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>
<!-- End Google Tag Manager (noscript) -->{{end}}$q$),
('art_head', $q$
<!-- Google Tag Manager -->
<script>(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
})(window,document,'script','dataLayer','GTM-PQ3WCTBD');</script>
<!-- End Google Tag Manager -->$q$),
('art_body', $q$<!-- Google Tag Manager (noscript) -->
<noscript><iframe src="https://www.googletagmanager.com/ns.html?id=GTM-PQ3WCTBD"
height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>
<!-- End Google Tag Manager (noscript) -->
$q$);

-- tpl_body carried a stray leading quote from authoring; strip it rather than risk it.
UPDATE _gtm SET v = ltrim(v, '''') WHERE k='tpl_body';

-- ── 1. Pre-guards ──────────────────────────────────────────────────────────────
DO $guard$
DECLARE n int;
BEGIN
  -- the old seam must still be dormant, or retiring it is not a no-op
  SELECT count(*) INTO n FROM site_specs
   WHERE is_current AND data::text ILIKE '%analytics_id%';
  IF n <> 0 THEN
    RAISE EXCEPTION 'analytics_id is no longer dormant (% site(s) set it) — do not retire blind', n;
  END IF;

  -- all 7 components must exist
  SELECT count(*) INTO n FROM content_components WHERE id IN (
    'aec98dbe-76b7-4e13-9641-e5b6ba2502aa','14cf6193-c8f0-4640-9cf1-f8b5347e6885',
    '9644c86f-18b0-4f75-b086-5b79a74a48d7','990b7162-a8f1-4691-886f-9d018fdd86b7',
    'e99b0dfa-a3b3-4a29-ae83-73211cb3975e','58fde68f-9190-4e5e-b6a5-ea21cf27a9af',
    'ad6033ae-5c3c-46af-8a1a-f11020b83594');
  IF n <> 7 THEN RAISE EXCEPTION 'expected 7 chrome components, found %', n; END IF;

  -- every deployed non-idea.uk site must have exactly one charset anchor in its head
  -- artefact (either case), or the artefact UPDATE below silently misses it
  SELECT count(*) INTO n FROM site_components sc JOIN sites s ON s.id=sc.site_id
   WHERE s.status='deployed' AND s.domain<>'idea.uk' AND sc.slot_name='head'
     AND (sc.rendered_html LIKE '%<meta charset="UTF-8">%'
       OR sc.rendered_html LIKE '%<meta charset="utf-8">%');
  IF n <> 13 THEN RAISE EXCEPTION 'only %/13 head artefacts have a usable charset anchor', n; END IF;
END $guard$;

-- ── 2. Backups ─────────────────────────────────────────────────────────────────
DROP TABLE IF EXISTS bak_gtm_fleet_20260731_site_components;
CREATE TABLE bak_gtm_fleet_20260731_site_components AS
  SELECT sc.* FROM site_components sc JOIN sites s ON s.id=sc.site_id
   WHERE s.status='deployed' AND sc.slot_name IN ('head','header');

DROP TABLE IF EXISTS bak_gtm_fleet_20260731_content_components;
CREATE TABLE bak_gtm_fleet_20260731_content_components AS
  SELECT * FROM content_components WHERE id IN (
    'aec98dbe-76b7-4e13-9641-e5b6ba2502aa','14cf6193-c8f0-4640-9cf1-f8b5347e6885',
    '9644c86f-18b0-4f75-b086-5b79a74a48d7','990b7162-a8f1-4691-886f-9d018fdd86b7',
    'e99b0dfa-a3b3-4a29-ae83-73211cb3975e','58fde68f-9190-4e5e-b6a5-ea21cf27a9af',
    'ad6033ae-5c3c-46af-8a1a-f11020b83594');

-- ── 3. Retire the competing gtag.js seam (head-seo-standard, 4 domains) ────────
-- Removes the whole {{if .analytics_id}} … {{end}} block and the schema field. Dormant,
-- so this changes no rendered byte today; doing it now is what stops a future session
-- switching it on beside GTM and double-counting every pageview.
UPDATE content_components
   SET html_template = regexp_replace(html_template,
         '\s*\{\{if \.analytics_id\}\}.*?\{\{end\}\}', '', 'gs'),
       input_schema = jsonb_set(input_schema, '{fields}',
                                (input_schema->'fields') - 'analytics_id'),
       updated_at = now()
 WHERE id = 'aec98dbe-76b7-4e13-9641-e5b6ba2502aa';

-- ── 4. Head templates: gated GTM script after the charset meta ─────────────────
UPDATE content_components cc
   SET html_template = replace(cc.html_template, '<meta charset="UTF-8">',
                               '<meta charset="UTF-8">' || (SELECT v FROM _gtm WHERE k='tpl_head')),
       input_schema = jsonb_set(COALESCE(cc.input_schema,'{"fields":{}}'::jsonb), '{fields,gtm_container_id}',
         '{"type":"text","source":"config.analytics.gtm_container_id","required":false,"on_missing":"skip_field"}'::jsonb, true),
       updated_at = now()
 WHERE cc.id = 'aec98dbe-76b7-4e13-9641-e5b6ba2502aa';

UPDATE content_components cc
   SET html_template = replace(cc.html_template, '<meta charset="utf-8">',
                               '<meta charset="utf-8">' || (SELECT v FROM _gtm WHERE k='tpl_head')),
       input_schema = jsonb_set(COALESCE(cc.input_schema,'{"fields":{}}'::jsonb), '{fields,gtm_container_id}',
         '{"type":"text","source":"config.analytics.gtm_container_id","required":false,"on_missing":"skip_field"}'::jsonb, true),
       updated_at = now()
 WHERE cc.id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885';

-- ── 5. Header templates: gated noscript, prepended (renders right after <body>) ─
-- WRAPPED-schema headers.
UPDATE content_components cc
   SET html_template = (SELECT v FROM _gtm WHERE k='tpl_body') || E'\n' || cc.html_template,
       input_schema = jsonb_set(COALESCE(cc.input_schema,'{"fields":{}}'::jsonb), '{fields,gtm_container_id}',
         '{"type":"text","source":"config.analytics.gtm_container_id","required":false,"on_missing":"skip_field"}'::jsonb, true),
       updated_at = now()
 WHERE cc.id IN ('9644c86f-18b0-4f75-b086-5b79a74a48d7',
                 '990b7162-a8f1-4691-886f-9d018fdd86b7',
                 'e99b0dfa-a3b3-4a29-ae83-73211cb3975e',
                 'ad6033ae-5c3c-46af-8a1a-f11020b83594');

-- header-theme-chrome has a NULL input_schema. The gap-fill treats a schema with no
-- "fields" key as FLAT (:604-607), so the descriptor goes at top level here.
UPDATE content_components cc
   SET html_template = (SELECT v FROM _gtm WHERE k='tpl_body') || E'\n' || cc.html_template,
       input_schema = COALESCE(cc.input_schema,'{}'::jsonb) || jsonb_build_object(
         'gtm_container_id', '{"type":"text","source":"config.analytics.gtm_container_id","required":false,"on_missing":"skip_field"}'::jsonb),
       updated_at = now()
 WHERE cc.id = '58fde68f-9190-4e5e-b6a5-ea21cf27a9af';

-- ── 6. The per-site container id, for the 13 remaining deployed domains ────────
UPDATE site_specs ss SET is_current=false, superseded_at=now()
  FROM sites s
 WHERE s.id=ss.site_id AND s.status='deployed' AND s.domain<>'idea.uk'
   AND ss.aspect='site_config' AND ss.is_current;

INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes)
SELECT s.id, 'site_config',
       '{"analytics": {"gtm_container_id": "GTM-PQ3WCTBD"}}'::jsonb,
       'operator', 'claude-session-gtm-2026-07-31',
       'Estate-wide GTM container (owner decision 2026-07-31: one container, per-domain '
       'reporting off the hostname dimension). Read by the head/header chrome templates '
       'via input_schema source config.analytics.gtm_container_id.'
  FROM sites s
 WHERE s.status='deployed' AND s.domain<>'idea.uk';

-- ── 7. The stored artefacts (what pages are actually assembled from) ───────────
UPDATE site_components sc
   SET rendered_html = replace(sc.rendered_html, '<meta charset="UTF-8">',
                               '<meta charset="UTF-8">' || (SELECT v FROM _gtm WHERE k='art_head')),
       updated_at = now()
  FROM sites s
 WHERE s.id=sc.site_id AND s.status='deployed' AND s.domain<>'idea.uk'
   AND sc.slot_name='head' AND sc.rendered_html LIKE '%<meta charset="UTF-8">%';

UPDATE site_components sc
   SET rendered_html = replace(sc.rendered_html, '<meta charset="utf-8">',
                               '<meta charset="utf-8">' || (SELECT v FROM _gtm WHERE k='art_head')),
       updated_at = now()
  FROM sites s
 WHERE s.id=sc.site_id AND s.status='deployed' AND s.domain<>'idea.uk'
   AND sc.slot_name='head' AND sc.rendered_html LIKE '%<meta charset="utf-8">%';

UPDATE site_components sc
   SET rendered_html = (SELECT v FROM _gtm WHERE k='art_body') || sc.rendered_html,
       updated_at = now()
  FROM sites s
 WHERE s.id=sc.site_id AND s.status='deployed' AND s.domain<>'idea.uk'
   AND sc.slot_name='header';

-- ── 8. Post-conditions ─────────────────────────────────────────────────────────
DO $verify$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_components sc JOIN sites s ON s.id=sc.site_id
   WHERE s.status='deployed' AND sc.slot_name='head'
     AND sc.rendered_html LIKE '%googletagmanager.com/gtm.js%'
     AND sc.rendered_html LIKE '%GTM-PQ3WCTBD%';
  IF n <> 14 THEN RAISE EXCEPTION 'head artefacts tagged: %/14', n; END IF;

  SELECT count(*) INTO n FROM site_components sc JOIN sites s ON s.id=sc.site_id
   WHERE s.status='deployed' AND sc.slot_name='header'
     AND sc.rendered_html LIKE '<!-- Google Tag Manager (noscript) -->%';
  IF n <> 14 THEN RAISE EXCEPTION 'header artefacts with noscript FIRST: %/14', n; END IF;

  SELECT count(*) INTO n FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.status='deployed' AND ss.aspect='site_config' AND ss.is_current
     AND ss.data->'analytics'->>'gtm_container_id' = 'GTM-PQ3WCTBD';
  IF n <> 14 THEN RAISE EXCEPTION 'site_specs container set on %/14 sites', n; END IF;

  SELECT count(*) INTO n FROM content_components
   WHERE html_template LIKE '%{{.gtm_container_id}}%' AND id IN (
     '116c5f91-bc0d-439d-9e13-a3ba2d145571','aec98dbe-76b7-4e13-9641-e5b6ba2502aa',
     '14cf6193-c8f0-4640-9cf1-f8b5347e6885','f420f3fa-43a2-4a2f-b2e1-39770d45b494',
     '9644c86f-18b0-4f75-b086-5b79a74a48d7','990b7162-a8f1-4691-886f-9d018fdd86b7',
     'e99b0dfa-a3b3-4a29-ae83-73211cb3975e','58fde68f-9190-4e5e-b6a5-ea21cf27a9af',
     'ad6033ae-5c3c-46af-8a1a-f11020b83594');
  IF n <> 9 THEN RAISE EXCEPTION 'gated templates: %/9', n; END IF;

  -- the old seam is gone, template and schema
  SELECT count(*) INTO n FROM content_components
   WHERE (html_template LIKE '%analytics_id%' OR input_schema::text LIKE '%analytics_id%');
  IF n <> 0 THEN RAISE EXCEPTION 'analytics_id survives in % component(s)', n; END IF;

  RAISE NOTICE 'c1 OK — 14/14 sites tagged in artefact+spec, 9/9 templates gated, gtag seam retired';
END $verify$;

COMMIT;
