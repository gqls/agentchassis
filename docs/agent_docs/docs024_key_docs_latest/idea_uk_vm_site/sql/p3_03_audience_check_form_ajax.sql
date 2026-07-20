-- p3_03_audience_check_form_ajax.sql — idea.uk: stop the free taster throwing the
-- visitor onto a chrome-less fragment, and stop the tool cards linking to a POST-only URL.
--
-- ── DEFECT 1: the taster result page has no chrome ────────────────────────────────────
-- /audience-check is an AJAX FRAGMENT ENDPOINT BY DESIGN. audience_check.go:135-138 is
-- POST-only, and renderAudienceHTML's own comment says it "produces the HTML fragment the
-- taster widget drops into its result div". There is no widget. p2_01 seeded the form as a
-- plain native POST:
--     <form class="ac-form" action="/audience-check" method="POST">
-- with no JS partner, so the browser NAVIGATES to the endpoint and renders the bare
-- fragment — no header, no footer, no way back. That is exactly what the owner saw.
--
-- Why the report form does not do this: p2_02 gave report-request-form a js_content, and
-- collectJSAssets (rerender_single_page_action.go:156-176) publishes
-- /tools/assets/{function}.js ONLY when js_content is non-empty. Measured today:
--     report-request-form  js_content = 266 bytes  → /tools/assets/report-request-form.js  HTTP 200
--     audience-check-form  js_content =   0 bytes  → /tools/assets/audience-check-form.js  HTTP 404
-- So this is the same defect p2_02 fixed for the other form, never applied to this one.
-- The fix is that same established mechanism, not a new one.
--
-- ── DEFECT 2: "POST only" ─────────────────────────────────────────────────────────────
-- The tool-list cards on index + tools link to /audience-check, which answers GET with
-- HTTP 405 "POST only". Verified live. The card URLs are NOT hand-written: tool-list's
-- items field is source "query.pages_where_type:tool", so it is derived from the pages
-- table — the pointer page tool-audience-check carries url='/audience-check'. That row is
-- the source of truth and is what this fixes. The already-rendered content_data is patched
-- too, because a query source is resolved at plan time and a rerender alone would keep
-- serving the stored value; that second update is transient by nature (bugs_open/001 —
-- anything written to page_components has an undefined shelf life), which is precisely why
-- the pages row is corrected as well.
--
-- Target: /tools.html#audience-check — the id the seeded form section already carries.
--
-- ── NOT FIXED HERE (deliberately, and worth knowing) ──────────────────────────────────
-- (a) The tool's fragment ends with <a href="#request" class="taster-cta">, and NO page on
--     this site has id="request" — the real form is /report.html#request-a-report. That
--     href is emitted by the box binary, so it cannot be fixed from the database. The
--     injector below retargets that one known-dead anchor after injection. It is a stopgap
--     compensating for a tool-side defect: delete it once the tool emits the right href.
-- (b) /tools/assets/site-header.js 404s — the chrome template references it but
--     collectJSAssets only reads page_components (:157), never site_components, so a site
--     component can reference a JS asset that is never published. The hamburger/mobile
--     menu is therefore dead. Structural, fleet-wide, out of scope for this file.

\set ON_ERROR_STOP on

BEGIN;

-- 1. The submit interceptor. Without preventDefault the browser navigates to the bare
--    fragment; that single line is the whole bug.
UPDATE content_components
SET js_content = $JS$(function () {
  var form = document.querySelector('.ac-form');
  if (!form) return;
  var out = document.getElementById('ac-result');
  if (!out) return;
  var btn = form.querySelector('.form-submit');
  var busy = false;

  form.addEventListener('submit', function (ev) {
    ev.preventDefault();
    if (busy) return;
    busy = true;
    var label = btn ? btn.textContent : '';
    if (btn) { btn.disabled = true; btn.textContent = 'Checking…'; }
    out.hidden = false;
    out.innerHTML = '<p class="ac-result-status">Running your free audience check — this takes a few seconds.</p>';

    fetch(form.getAttribute('action'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams(new FormData(form)).toString()
    })
      .then(function (r) { return r.text(); })
      .then(function (html) {
        out.innerHTML = html;
        var cta = out.querySelector('a.taster-cta[href="#request"]');
        if (cta) { cta.setAttribute('href', '/report.html#request-a-report'); }
        out.scrollIntoView({ behavior: 'smooth', block: 'start' });
      })
      .catch(function () {
        out.innerHTML = '<p class="ac-result-status">Sorry — the check could not run just now. Please try again in a moment, or request the full report.</p>';
      })
      .then(function () {
        busy = false;
        if (btn) { btn.disabled = false; btn.textContent = label; }
      });
  });
})();$JS$,
    html_template = $TMPL$<section class="ac-form-section" id="audience-check" data-component="audience-check-form">
  <div class="ac-form-container">
    <h2>{{if .heading}}{{.heading}}{{else}}Free audience check{{end}}</h2>
    {{if .subtitle}}<p class="ac-form-subtitle">{{.subtitle}}</p>{{end}}
    <form class="ac-form" action="/audience-check" method="POST">
      <div class="form-group">
        <label for="ac-business">What is the business or idea?</label>
        <input type="text" id="ac-business" name="business" required maxlength="200"
               placeholder="e.g. a booking tool for small veterinary practices">
      </div>
      <div class="form-group">
        <label for="ac-audience">Who do you think it is for?</label>
        <input type="text" id="ac-audience" name="audience" required maxlength="200"
               placeholder="e.g. independent vets in the UK">
      </div>
      <button type="submit" class="form-submit">{{if .button_text}}{{.button_text}}{{else}}Run the free check{{end}}</button>
      {{if .footnote}}<p class="ac-form-note">{{.footnote}}</p>{{end}}
    </form>
    <div id="ac-result" class="ac-result" role="status" aria-live="polite" hidden></div>
  </div>
</section>
<style>
/* layout only — colours inherit from the global CSS variables */
.ac-form-section { padding: var(--spacing-section, 5rem 2rem); }
.ac-form-container { max-width: 640px; margin: 0 auto; }
.ac-form-section h2 { text-align: center; margin-bottom: 1rem; }
.ac-form-subtitle { text-align: center; margin-bottom: 2rem; }
.ac-form { display: flex; flex-direction: column; gap: 1.5rem; }
.ac-form .form-group { display: flex; flex-direction: column; gap: .5rem; }
.ac-form input { padding: .75rem; border: 1px solid var(--color-border, #ccc); border-radius: 4px; font: inherit; }
.ac-form .form-submit { padding: .9rem 1.5rem; border: 0; border-radius: 4px; font: inherit; cursor: pointer;
  background: var(--color-primary, #333); color: var(--color-on-primary, #fff); }
.ac-form .form-submit[disabled] { opacity: .6; cursor: default; }
.ac-form-note { font-size: .9rem; text-align: center; }
/* Result panel. The injected markup comes from the tool (h4/p/ol/li plus
   .taster-upsell / .taster-cta), so these style ITS classes, not ours. */
.ac-result { margin-top: 2.5rem; }
.ac-result[hidden] { display: none; }
.ac-result-status { text-align: center; font-style: italic; }
.ac-result h4 { margin: 1.5rem 0 .35rem; }
.ac-result p, .ac-result li { line-height: 1.6; }
.ac-result ol { padding-left: 1.25rem; }
.ac-result li { margin-bottom: .75rem; }
.ac-result .taster-upsell { margin-top: 2.5rem; padding: 1.5rem;
  border: 1px solid var(--color-border, #ccc); border-radius: 6px; }
.ac-result .taster-cta { display: inline-block; margin-top: .5rem; padding: .9rem 1.5rem;
  border-radius: 4px; text-decoration: none; font-weight: 600;
  background: var(--color-primary, #333); color: var(--color-on-primary, #fff); }
</style>
<script src="/tools/assets/audience-check-form.js" defer></script>$TMPL$,
    updated_at = now()
WHERE name = 'audience-check-form' AND forked_from IS NULL;

-- 2. Source of truth for the tool cards: the pointer page must not advertise a POST-only
--    endpoint as a browsable URL.
UPDATE pages
SET url = '/tools.html#audience-check', updated_at = now()
WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND name = 'tool-audience-check'
  AND url = '/audience-check';

-- 3. The already-resolved copies, so the live cards are right before the next plan pass.
UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{items,0,url}', '"/tools.html#audience-check"'),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND pc.slot_name = 'tool-list'
  AND pc.content_data #>> '{items,0,url}' = '/audience-check';

COMMIT;

-- Post-apply checks.
\echo '=== form: want js_len>0, src_ref=t, result_div=t, raw_inline=f ==='
SELECT name, function,
       length(coalesce(js_content,''))                     AS js_len,
       html_template LIKE '%<script src=%'                 AS src_ref,
       html_template LIKE '%id="ac-result"%'               AS result_div,
       html_template ~ '<script>[^s]'                      AS raw_inline
FROM content_components WHERE name IN ('audience-check-form','report-request-form') AND forked_from IS NULL;

\echo '=== no route should still advertise the POST-only endpoint ==='
SELECT name, url FROM pages
WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND url LIKE '%audience-check%';

SELECT p.url AS page, pc.content_data #>> '{items,0,url}' AS card_target
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND pc.slot_name='tool-list';
