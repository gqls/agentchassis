-- 208_webdesign_ported_page_component.sql
--
-- One shared library component that every ported page on webdesign.co.uk
-- instantiates. There are ~97 such pages; they differ only in the HTML the
-- importer writes into page_components.rendered_html, so they share one row
-- rather than getting a component each.
--
-- Design notes, each of which is load-bearing:
--
--   component_level = 'section'. NOT chrome. Chrome components had their JS
--   dropped entirely before bugs 018/041 were fixed, and a section is what the
--   page assembler concatenates anyway.
--
--   js_content is EMPTY, deliberately. Each ported page carries its own inline
--   <script> inside rendered_html — that is how these tools work and the
--   assembly path preserves it verbatim. A non-empty js_content here would make
--   collectJSAssets publish a /tools/assets/ported-page.js that every page would
--   then be expected to load, for no reason.
--
--   render_mode = 'template' so no content-writer agent is ever asked to fill
--   it. The content is authored, not generated.
--
--   The template is a bare passthrough. The importer writes finished HTML into
--   rendered_html; this template exists only so the row is well-formed for any
--   code path that reads it.

\set ON_ERROR_STOP on

BEGIN;

INSERT INTO content_components (
    name, function, section_type, component_level, render_mode,
    html_template, input_schema, js_content, is_active, created_at, updated_at
)
SELECT
    'Ported Page (webdesign.co.uk)',
    'ported-page',
    'ported-page',
    'section',
    'template',
    $tmpl$<section class="ported-page" data-component="ported-page">{{.body}}</section>$tmpl$,
    $schema${
      "fields": {
        "body": {
          "type": "html",
          "source": "authored",
          "required": true,
          "llm_guidance": "Never generated. Written by cmd/webdesignport from the hand-built source pages."
        }
      }
    }$schema$::jsonb,
    '',
    true,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM content_components
     WHERE function = 'ported-page' AND is_active
);

DO $verify$
DECLARE n int; js text;
BEGIN
    SELECT count(*), MAX(COALESCE(js_content, '')) INTO n, js
      FROM content_components WHERE function = 'ported-page' AND is_active;

    IF n <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 active ported-page component, found %', n;
    END IF;
    IF COALESCE(js, '') <> '' THEN
        RAISE EXCEPTION 'ported-page.js_content must stay empty (see header)';
    END IF;
    RAISE NOTICE 'ported-page component ready';
END
$verify$;

COMMIT;
