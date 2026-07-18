-- p2_01 — Restore idea.uk's funnel: author the tool's entry forms as chassis sections.
--
-- WHY (see /bugs_open/017): the tool served its OWN landing page at "/", and that page
-- carried the entry forms. The VM cutover gave "/" to the static site, so the forms
-- vanished. /audience-check and /request are POST-only HANDLERS, not pages — the two
-- surviving href="/audience-check" links are GET requests, hence "POST only". The tool
-- is healthy, reachable, and unusable: there is no way to buy.
--
-- FIX: author both forms as normal chassis components and place them on existing pages.
-- Same origin post-cutover, so a plain form POST works — no CORS, no JS required.
--
-- FIELD CONTRACTS taken from the tool source (do NOT guess these):
--   /audience-check  (audience_check.go:159-160)  business, audience
--   /request         (service.go:327-355)         name, email, business, audience, notes
--                                                 + company_url (honeypot)
--                                                 + _elapsed    (timing gate, ms)
-- The honeypot rejects only when FILLED and the timing gate FAILS OPEN when _elapsed is
-- absent, so these forms work against the currently-deployed binary AND the hardened one.
--
-- Placement (no new pages — avoids the re-plan landmine, /bugs_open/001):
--   report page → report-request-form   (the £29 buy path)
--   tools  page → audience-check-form   (the free taster)

\set ON_ERROR_STOP on
\set PLAN 'ff03bdef-3bb2-40eb-93ff-efa70f46b6b8'
\set SID  '1244516d-014d-421c-88c6-090bb1e9552a'

BEGIN;

-- ── 1. the free-taster form ────────────────────────────────────────────────────
INSERT INTO content_components
    (name, function, section_type, component_level, render_mode, is_active,
     template_closed, suitable_page_types, display_name, html_template)
VALUES (
 'audience-check-form', 'generic-text-block', 'audience-check-form', 'section',
 'template', true, true, '["content","tool"]'::jsonb, 'Free audience check form',
$TMPL$<section class="ac-form-section" id="audience-check" data-component="audience-check-form">
  <div class="ac-form-container">
    <h2>{{if .heading}}{{.heading}}{{else}}Free audience check{{end}}</h2>
    {{if .subtitle}}<p class="ac-form-subtitle">{{.subtitle}}</p>{{end}}
    <form class="ac-form" action="/audience-check" method="POST">
      <div class="form-group">
        <label for="ac-business">What is the business or idea?</label>
        <input type="text" id="ac-business" name="business" required
               placeholder="e.g. a booking tool for small veterinary practices">
      </div>
      <div class="form-group">
        <label for="ac-audience">Who do you think it is for?</label>
        <input type="text" id="ac-audience" name="audience" required
               placeholder="e.g. independent vets in the UK">
      </div>
      <button type="submit" class="form-submit">{{if .button_text}}{{.button_text}}{{else}}Run the free check{{end}}</button>
      {{if .footnote}}<p class="ac-form-note">{{.footnote}}</p>{{end}}
    </form>
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
.ac-form-note { font-size: .9rem; text-align: center; }
</style>$TMPL$
);

-- ── 2. the £29 report request form ─────────────────────────────────────────────
-- company_url is the honeypot: off-screen, not display:none (bots skip hidden inputs
-- less often than they skip 'display:none'), never focusable, never autofilled.
-- _elapsed is set by JS at submit time; absent (no-JS) fails open by design.
INSERT INTO content_components
    (name, function, section_type, component_level, render_mode, is_active,
     template_closed, suitable_page_types, display_name, html_template)
VALUES (
 'report-request-form', 'generic-text-block', 'report-request-form', 'section',
 'template', true, true, '["content","landing"]'::jsonb, 'Report request form',
$TMPL$<section class="rr-form-section" id="request-a-report" data-component="report-request-form">
  <div class="rr-form-container">
    <h2>{{if .heading}}{{.heading}}{{else}}Request your report{{end}}</h2>
    {{if .subtitle}}<p class="rr-form-subtitle">{{.subtitle}}</p>{{end}}
    <form class="rr-form" action="/request" method="POST" id="rr-form">
      <div class="form-group">
        <label for="rr-name">Your name</label>
        <input type="text" id="rr-name" name="name" required autocomplete="name">
      </div>
      <div class="form-group">
        <label for="rr-email">Email</label>
        <input type="email" id="rr-email" name="email" required autocomplete="email">
      </div>
      <div class="form-group">
        <label for="rr-business">The business or idea</label>
        <input type="text" id="rr-business" name="business" required
               placeholder="e.g. a booking tool for small veterinary practices">
      </div>
      <div class="form-group">
        <label for="rr-audience">Who you think it is for</label>
        <input type="text" id="rr-audience" name="audience" required
               placeholder="e.g. independent vets in the UK">
      </div>
      <div class="form-group">
        <label for="rr-notes">Anything else we should know <span class="rr-optional">(optional)</span></label>
        <textarea id="rr-notes" name="notes" rows="4"></textarea>
      </div>
      <div class="rr-hp" aria-hidden="true">
        <label for="rr-company-url">Do not fill this in</label>
        <input type="text" id="rr-company-url" name="company_url" tabindex="-1" autocomplete="off">
      </div>
      <input type="hidden" name="_elapsed" id="rr-elapsed" value="">
      <button type="submit" class="form-submit">{{if .button_text}}{{.button_text}}{{else}}Request the report{{end}}</button>
      {{if .footnote}}<p class="rr-form-note">{{.footnote}}</p>{{end}}
    </form>
  </div>
</section>
<style>
.rr-form-section { padding: var(--spacing-section, 5rem 2rem); }
.rr-form-container { max-width: 640px; margin: 0 auto; }
.rr-form-section h2 { text-align: center; margin-bottom: 1rem; }
.rr-form-subtitle { text-align: center; margin-bottom: 2rem; }
.rr-form { display: flex; flex-direction: column; gap: 1.5rem; }
.rr-form .form-group { display: flex; flex-direction: column; gap: .5rem; }
.rr-form input, .rr-form textarea { padding: .75rem; border: 1px solid var(--color-border, #ccc);
  border-radius: 4px; font: inherit; }
.rr-form .form-submit { padding: .9rem 1.5rem; border: 0; border-radius: 4px; font: inherit; cursor: pointer;
  background: var(--color-primary, #333); color: var(--color-on-primary, #fff); }
.rr-optional { font-weight: 400; opacity: .7; }
.rr-form-note { font-size: .9rem; text-align: center; }
/* honeypot: off-screen rather than display:none, and never focusable */
.rr-hp { position: absolute; left: -9999px; width: 1px; height: 1px; overflow: hidden; }
</style>
<script>
(function () {
  var loaded = Date.now();
  var f = document.getElementById('rr-form');
  if (!f) return;
  f.addEventListener('submit', function () {
    var el = document.getElementById('rr-elapsed');
    if (el) el.value = String(Date.now() - loaded);
  });
})();
</script>$TMPL$
);

-- ── 3. place them on existing pages (append; no new pages, no re-plan) ─────────
INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
VALUES (:'PLAN', 'report', 4, 'report-request-form'),
       (:'PLAN', 'tools',  3, 'audience-check-form');

\echo '=== components created ==='
SELECT name, section_type, render_mode, template_closed, length(html_template) AS len
FROM content_components WHERE name IN ('audience-check-form','report-request-form');

\echo '=== plan sections now ==='
SELECT page_name, ordering, component_name FROM site_plan_sections
WHERE plan_id = :'PLAN' AND page_name IN ('report','tools') ORDER BY page_name, ordering;

COMMIT;
