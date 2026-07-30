-- p4_34_gtm_container.sql — idea.uk: Google Tag Manager (GTM-PQ3WCTBD) on every page
--
-- WHAT THE OWNER ASKED FOR
--   1. The GTM <script> as high in <head> as possible, on every page of idea.uk.
--   2. The GTM <noscript> iframe immediately after the opening <body> tag.
--
-- WHERE THOSE TWO PLACES ACTUALLY ARE (this is the whole design)
--   assemblePage (rerender_single_page_action.go:574-593) writes, in order:
--       <!DOCTYPE html>\n<html lang="en">\n | head | \n<body>\n | header | <main>… | footer
--   So the <head> is the `head` site_components slot, and the FIRST thing after
--   <body> is the `header` slot. There is no other seam: the <body> tag itself is a
--   Go string literal, so "immediately after <body>" can ONLY be the top of the
--   header slot without a chassis change and an image roll.
--
-- WHY IT IS DONE WITH A GATED TEMPLATE + A PER-SITE KEY, NOT A HARDCODED SNIPPET
--   The `head` component "Document Head" (116c5f91-…) is SHARED BY NINE SITES —
--   measured, not assumed:
--     SELECT s.domain FROM site_components sc JOIN sites s ON s.id=sc.site_id
--     WHERE sc.component_id='116c5f91-bc0d-439d-9e13-a3ba2d145571';
--     → dartsonline.com, fundamentallyai.com, gamesdesign.co.uk, idea.uk, oufe.com,
--       relojistas.com, robot-hands.com, vetcomparison.uk, vonc.com  (9 rows)
--   Hardcoding idea.uk's container id there would put idea.uk's tag on eight other
--   people's sites. So the container id is a PER-SITE value resolved from site_specs,
--   and the template block is gated on it: a site with no id renders nothing, and its
--   output stays byte-identical.
--
--   The resolution path already exists (no Go change): render_site_components_action.go
--   :585-645 does a schema-driven gap-fill over the component's own input_schema via
--   newSourceResolver; `config.*` → resolveConfigPath (plan_sections_action.go:516-527)
--   searches site_specs aspects site_config / identity / design_intent for the dotted
--   path. Hence: input_schema field `gtm_container_id` with source
--   `config.analytics.gtm_container_id`, and a site_specs `site_config` row holding it.
--
--   ⚠ The head component's existing input_schema is FLAT and holds SCALARS
--   ({"title": "...", "description": "..."}), which the gap-fill loop skips as
--   "not a field descriptor" (:612-615). Adding a MAP-valued key alongside them is
--   what makes the loop see it. The flat vs wrapped {"fields":{…}} distinction is
--   handled at :604-607, so a flat add is correct here.
--
-- SCOPE / BLAST RADIUS — measured, not argued
--   * head template: 9 sites render it; 8 have no gtm_container_id ⇒ the {{if}} block
--     is empty for them ⇒ byte-identical output. Only idea.uk changes.
--   * header template f420f3fa-… : ONE site uses it (idea.uk). Verified:
--     SELECT count(*) FROM site_components WHERE component_id='f420f3fa-…';  → 1
--   * No other site gets a site_specs row here.
--
-- WHY BOTH THE TEMPLATE **AND** THE RENDERED ARTEFACT ARE WRITTEN
--   Pages are assembled from site_components.rendered_html, NOT from the template
--   (bugs_open/117: chrome is a stored artefact and no page re-render rebuilds it).
--   Writing only the template would change nothing on any page until some future
--   chrome rebuild; writing only the artefact would be silently reverted BY that
--   rebuild. Both, or the change is either inert or temporary.
--
-- STILL REQUIRED AFTER THIS FILE — this is DB only, so it is live at the next
--   assembly, but the 21 deployed pages have chrome BAKED IN. They must be
--   re-assembled and re-deployed:  scripts/fire_reassemble_idea_uk.sh
--
-- ROLLBACK: backup tables created below; see p4_34_ROLLBACK.sql.

BEGIN;

-- ── 0. Fail loudly rather than silently updating 0 rows ────────────────────────
DO $guard$
BEGIN
  IF (SELECT count(*) FROM content_components
       WHERE id IN ('116c5f91-bc0d-439d-9e13-a3ba2d145571',   -- Document Head (shared ×9)
                    'f420f3fa-43a2-4a2f-b2e1-39770d45b494')) <> 2 THEN
    RAISE EXCEPTION 'expected the head + idea.uk header components to exist';
  END IF;

  IF (SELECT count(*) FROM site_components
       WHERE component_id = 'f420f3fa-43a2-4a2f-b2e1-39770d45b494') <> 1 THEN
    RAISE EXCEPTION 'idea.uk site-header is no longer single-site — re-check blast radius';
  END IF;

  -- Exactly one charset meta to anchor the head insertion on, and no GTM already.
  IF (SELECT (length(rendered_html) - length(replace(rendered_html,'<meta charset="UTF-8">','')))
             / length('<meta charset="UTF-8">')
      FROM site_components
      WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND slot_name='head') <> 1 THEN
    RAISE EXCEPTION 'head slot does not have exactly one <meta charset="UTF-8"> anchor';
  END IF;

  IF EXISTS (SELECT 1 FROM site_components
              WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a'
                AND rendered_html LIKE '%googletagmanager%') THEN
    RAISE EXCEPTION 'GTM already present in idea.uk chrome — not applying twice';
  END IF;
END
$guard$;

-- ── 1. Backups ─────────────────────────────────────────────────────────────────
DROP TABLE IF EXISTS bak_ideauk_gtm_20260730_site_components;
CREATE TABLE bak_ideauk_gtm_20260730_site_components AS
  SELECT * FROM site_components
   WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a';

DROP TABLE IF EXISTS bak_ideauk_gtm_20260730_content_components;
CREATE TABLE bak_ideauk_gtm_20260730_content_components AS
  SELECT * FROM content_components
   WHERE id IN ('116c5f91-bc0d-439d-9e13-a3ba2d145571',
                'f420f3fa-43a2-4a2f-b2e1-39770d45b494');

-- ── 2. The per-site container id (this is the ONLY per-site knob) ──────────────
-- idx_site_specs_current is UNIQUE on (site_id, aspect) WHERE is_current, so an
-- existing current row must be superseded rather than duplicated. idea.uk has no
-- site_config aspect today (verified: 10 aspects, none of them site_config).
UPDATE site_specs
   SET is_current = false, superseded_at = now()
 WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a'
   AND aspect='site_config' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes)
VALUES (
  '1244516d-014d-421c-88c6-090bb1e9552a',
  'site_config',
  '{"analytics": {"gtm_container_id": "GTM-PQ3WCTBD"}}'::jsonb,
  'operator',
  'claude-session-gtm-2026-07-30',
  'Google Tag Manager container for idea.uk. Read by the head/header chrome templates '
  'via input_schema source config.analytics.gtm_container_id. Setting this key on any '
  'other site is all that is needed to give that site its own GTM container.'
);

-- ── 3. Shared head template: GTM <script>, gated, as high as possible ──────────
-- Placement: immediately AFTER <meta charset="UTF-8">, not before it. The encoding
-- declaration must appear entirely within the first 1024 bytes of the document
-- (HTML spec); GTM is ~370 bytes, so putting GTM first would still fit today, but it
-- makes a spec requirement depend on the length of a third-party snippet. Charset
-- first, GTM immediately next, is the highest position that cannot break parsing.
UPDATE content_components
   SET html_template = replace(
         html_template,
         '<meta charset="UTF-8">',
         '<meta charset="UTF-8">' || E'\n' ||
         '{{if .gtm_container_id}}<!-- Google Tag Manager -->' || E'\n' ||
         '<script>(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({''gtm.start'':' || E'\n' ||
         'new Date().getTime(),event:''gtm.js''});var f=d.getElementsByTagName(s)[0],' || E'\n' ||
         'j=d.createElement(s),dl=l!=''dataLayer''?''&l=''+l:'''';j.async=true;j.src=' || E'\n' ||
         '''https://www.googletagmanager.com/gtm.js?id=''+i+dl;f.parentNode.insertBefore(j,f);' || E'\n' ||
         '})(window,document,''script'',''dataLayer'',''{{.gtm_container_id}}'');</script>' || E'\n' ||
         '<!-- End Google Tag Manager -->{{end}}'
       ),
       input_schema = COALESCE(input_schema, '{}'::jsonb) || jsonb_build_object(
         'gtm_container_id', jsonb_build_object(
           'type',        'string',
           'source',      'config.analytics.gtm_container_id',
           'required',    false,
           'description', 'Google Tag Manager container id (e.g. GTM-XXXXXXX). '
                          'Unset ⇒ the gated GTM block renders nothing.'
         )
       ),
       updated_at = now()
 WHERE id = '116c5f91-bc0d-439d-9e13-a3ba2d145571';

-- ── 4. idea.uk header template: GTM <noscript>, first thing after <body> ───────
UPDATE content_components
   SET html_template =
         '{{if .gtm_container_id}}<!-- Google Tag Manager (noscript) -->' || E'\n' ||
         '<noscript><iframe src="https://www.googletagmanager.com/ns.html?id={{.gtm_container_id}}"' || E'\n' ||
         'height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>' || E'\n' ||
         '<!-- End Google Tag Manager (noscript) -->{{end}}' || E'\n' ||
         html_template,
       input_schema = COALESCE(input_schema, '{}'::jsonb) || jsonb_build_object(
         'gtm_container_id', jsonb_build_object(
           'type',        'string',
           'source',      'config.analytics.gtm_container_id',
           'required',    false,
           'description', 'Google Tag Manager container id; renders the <noscript> '
                          'iframe that must sit immediately after <body>.'
         )
       ),
       updated_at = now()
 WHERE id = 'f420f3fa-43a2-4a2f-b2e1-39770d45b494';

-- ── 5. The STORED ARTEFACTS the pages are actually assembled from ──────────────
-- Same two insertions, already resolved to the literal container id. Without this
-- the change is inert until some future chrome rebuild (bugs_open/117).
UPDATE site_components
   SET rendered_html = replace(
         rendered_html,
         '<meta charset="UTF-8">',
         '<meta charset="UTF-8">' || E'\n' ||
         '<!-- Google Tag Manager -->' || E'\n' ||
         '<script>(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({''gtm.start'':' || E'\n' ||
         'new Date().getTime(),event:''gtm.js''});var f=d.getElementsByTagName(s)[0],' || E'\n' ||
         'j=d.createElement(s),dl=l!=''dataLayer''?''&l=''+l:'''';j.async=true;j.src=' || E'\n' ||
         '''https://www.googletagmanager.com/gtm.js?id=''+i+dl;f.parentNode.insertBefore(j,f);' || E'\n' ||
         '})(window,document,''script'',''dataLayer'',''GTM-PQ3WCTBD'');</script>' || E'\n' ||
         '<!-- End Google Tag Manager -->'
       ),
       updated_at = now()
 WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND slot_name='head';

UPDATE site_components
   SET rendered_html =
         '<!-- Google Tag Manager (noscript) -->' || E'\n' ||
         '<noscript><iframe src="https://www.googletagmanager.com/ns.html?id=GTM-PQ3WCTBD"' || E'\n' ||
         'height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>' || E'\n' ||
         '<!-- End Google Tag Manager (noscript) -->' || E'\n' ||
         rendered_html,
       updated_at = now()
 WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND slot_name='header';

-- ── 6. Post-conditions — assert, don't hope ────────────────────────────────────
DO $verify$
DECLARE
  head_hits int; hdr_hits int; tpl_hits int; spec_val text;
BEGIN
  SELECT count(*) INTO head_hits FROM site_components
   WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND slot_name='head'
     AND rendered_html LIKE '%googletagmanager.com/gtm.js%'
     AND rendered_html LIKE '%GTM-PQ3WCTBD%';
  IF head_hits <> 1 THEN RAISE EXCEPTION 'head artefact did not take the GTM script'; END IF;

  SELECT count(*) INTO hdr_hits FROM site_components
   WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND slot_name='header'
     AND rendered_html LIKE '<!-- Google Tag Manager (noscript) -->%';
  IF hdr_hits <> 1 THEN RAISE EXCEPTION 'header artefact did not take the noscript FIRST'; END IF;

  SELECT count(*) INTO tpl_hits FROM content_components
   WHERE id IN ('116c5f91-bc0d-439d-9e13-a3ba2d145571','f420f3fa-43a2-4a2f-b2e1-39770d45b494')
     AND html_template LIKE '%{{.gtm_container_id}}%'
     AND input_schema ? 'gtm_container_id';
  IF tpl_hits <> 2 THEN RAISE EXCEPTION 'templates/schemas did not both take the gated block'; END IF;

  SELECT data->'analytics'->>'gtm_container_id' INTO spec_val FROM site_specs
   WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND aspect='site_config' AND is_current;
  IF spec_val IS DISTINCT FROM 'GTM-PQ3WCTBD' THEN
    RAISE EXCEPTION 'site_config spec did not land: %', COALESCE(spec_val,'<null>');
  END IF;

  RAISE NOTICE 'p4_34 OK — head+header artefacts carry GTM, both templates gated, spec key set';
END
$verify$;

COMMIT;
