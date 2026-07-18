-- p2_02 — correct two defects in p2_01's raw INSERTs (caught by 016b §9 before they shipped).
--
-- 1. `function` defaulted to 'generic-text-block'. The convention for these is
--    function = component name (contact-form carries function='contact-form'), and
--    `function` is what names the deployed JS asset: /tools/assets/{function}.js.
--    Left as-is, report-request-form's JS would have been published as
--    generic-text-block.js and collided with anything else using that default.
--
-- 2. report-request-form was inserted with a RAW inline <script>. A raw SQL INSERT
--    bypasses separateInlineJS (store_generated_component_action.go), which is what
--    normally moves inline JS into js_content and leaves a <script src=…> reference.
--    So the row landed in exactly the shape 016b §9 calls broken:
--        js_len=0, has_src_ref=f, has_raw_inline=t
--    and collectJSAssets (rerender_single_page_action.go) only deploys
--    /tools/assets/{function}.js when js_content is NON-empty — so no asset would
--    have been published. Move the JS by hand into the healthy shape.
--
-- Consequence if this had shipped: _elapsed would never be set, and the /request
-- timing gate would silently fail open on every submission. The form would still
-- work (that is the gate's designed no-JS behaviour) but the defence would be dead
-- while appearing present — the worst of both.

\set ON_ERROR_STOP on

BEGIN;

UPDATE content_components SET function = name
WHERE name IN ('audience-check-form','report-request-form') AND forked_from IS NULL;

UPDATE content_components
SET js_content = $JS$(function () {
  var loaded = Date.now();
  var f = document.getElementById('rr-form');
  if (!f) return;
  f.addEventListener('submit', function () {
    var el = document.getElementById('rr-elapsed');
    if (el) el.value = String(Date.now() - loaded);
  });
})();$JS$,
    html_template = regexp_replace(
      html_template,
      '<script>.*</script>',
      '<script src="/tools/assets/report-request-form.js" defer></script>',
      'ns'   -- n: . matches newline; s: single-line mode for the greedy block
    )
WHERE name = 'report-request-form' AND forked_from IS NULL;

\echo '=== shape check — want js_len>0, src_ref=t, raw_inline=f ==='
SELECT name, function,
       LENGTH(COALESCE(js_content,''))            AS js_len,
       (html_template LIKE '%<script src=%')      AS has_src_ref,
       (html_template LIKE '%<script>%')          AS has_raw_inline
FROM content_components
WHERE name IN ('audience-check-form','report-request-form') AND forked_from IS NULL;

COMMIT;
