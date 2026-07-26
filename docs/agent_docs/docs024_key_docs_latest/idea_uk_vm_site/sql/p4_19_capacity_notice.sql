-- p4_19_capacity_notice.sql — say it on the site when the report queue is full.
--
-- Owner 2026-07-26: "if the queue of real orders is full then we want to state that on the site."
--
-- WHAT IT DOES. Adds a live capacity banner to the report request form: on page load it fetches
-- the tool's own /capacity endpoint and, when the queue is full, shows an honest notice above the
-- form. No new endpoint, no deploy — /capacity already exists (service.go:285) and is already
-- public on this origin (nginx reserves `location = /capacity` → the tool; verified 200 returning
-- {"active":5,"max":5,"open":false}).
--
-- THE COPY IS DELIBERATELY "THERE WILL BE A WAIT", NOT "YOU CANNOT ORDER" — and that is a
-- correctness point, not a tone preference. The capacity gate lives on the OPERATOR's /confirm
-- (service.go:411), NOT on /request: a visitor can always submit, and being full only means we
-- cannot START work yet. Blocking or disabling the form would therefore misdescribe the system
-- AND throw away real demand. The form stays fully usable; the banner sets expectations.
--
-- FAIL-OPEN, three ways, because a banner that wrongly says "full" costs real orders:
--   * fetch error / timeout / non-200  → show nothing, form behaves exactly as today
--   * malformed or unexpected JSON     → show nothing
--   * open:true                        → show nothing (we deliberately do NOT advertise
--                                        "slots available", which ages badly and reads as
--                                        pressure selling)
-- The banner is injected by script into an empty container, so a visitor with JS disabled sees
-- precisely today's page.
--
-- Component is single-instance (idea.uk only — verified), so this touches no other site.
-- The inline <script> sits alongside the existing external one, following p4_06's reasoning:
-- inline cannot be published-but-never-loaded (bugs_open/041 class).

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE cc.name = 'report-request-form';
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: expected report-request-form to be single-instance, found % — this edit would reach another site.', n;
  END IF;

  SELECT count(*) INTO n FROM content_components
  WHERE name = 'report-request-form' AND html_template LIKE '%rr-capacity%';
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: capacity banner already present — p4_19 already ran.';
  END IF;
END
$guard$;

DROP TABLE IF EXISTS bak_rrform_tpl_20260726;
CREATE TABLE bak_rrform_tpl_20260726 AS
SELECT id, name, html_template, now() AS snapshotted_at
FROM content_components WHERE name = 'report-request-form';

-- 1. Empty container immediately inside the form container, before the heading.
UPDATE content_components
SET html_template = replace(html_template,
      '  <div class="rr-form-container">',
      '  <div class="rr-form-container">' || E'\n' ||
      '    <div class="rr-capacity" id="rr-capacity" hidden></div>'),
    updated_at = now()
WHERE name = 'report-request-form';

-- 2. Styles + the fetch. Appended after the existing external script tag.
UPDATE content_components
SET html_template = html_template || E'\n' || $js$
<style>
.rr-form-section .rr-capacity {
  border: 1px solid var(--color-border);
  border-left: 4px solid var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 7%, transparent);
  border-radius: var(--border-radius, .4rem);
  padding: 1rem 1.15rem;
  margin: 0 0 1.75rem;
  line-height: 1.6;
}
.rr-form-section .rr-capacity strong { color: var(--color-heading); }
.rr-form-section .rr-capacity p { margin: 0; }
.rr-form-section .rr-capacity p + p { margin-top: .5rem; font-size: .93rem; color: var(--color-text-muted); }
</style>
<script>
(function () {
  var box = document.getElementById('rr-capacity');
  if (!box || !window.fetch) return;
  // Fail-open everywhere: any error, any unexpected shape, or open===true and
  // the visitor simply sees today's page. Never block, never claim capacity we
  // have not confirmed, and never advertise availability.
  var done = false;
  function show(active, max) {
    if (done) return;
    done = true;
    box.innerHTML =
      '<p><strong>We are at capacity right now.</strong> All ' + max + ' report slots are in ' +
      'progress, so there will be a wait before we can start a new one.</p>' +
      '<p>You can still send your request below — we work through them in order and will be in ' +
      'touch when a slot frees. Nothing is charged unless we look at your idea and decide we can ' +
      'do a useful job with it. In the meantime the free tools and guides on this site are the ' +
      'best place to start.</p>';
    box.hidden = false;
  }
  var ctl = null;
  try { ctl = new AbortController(); setTimeout(function () { ctl.abort(); }, 4000); } catch (e) {}
  fetch('/capacity', { cache: 'no-store', signal: ctl ? ctl.signal : undefined })
    .then(function (r) { return r.ok ? r.json() : null; })
    .then(function (d) {
      if (!d || d.open !== false) return;          // open, unknown, or malformed → say nothing
      var max = parseInt(d.max, 10);
      var active = parseInt(d.active, 10);
      if (!(max > 0)) return;
      show(active, max);
    })
    .catch(function () { /* fail open — deliberately silent */ });
})();
</script>
$js$,
    updated_at = now()
WHERE name = 'report-request-form';

DO $guard2$
DECLARE t text;
BEGIN
  SELECT html_template INTO t FROM content_components WHERE name = 'report-request-form';
  IF position('id="rr-capacity"' in t) = 0 OR position('/capacity' in t) = 0 THEN
    RAISE EXCEPTION 'ABORT: banner edit did not land.';
  END IF;
  IF position('action="/request"' in t) = 0 THEN
    RAISE EXCEPTION 'ABORT: the form itself was damaged — restore from bak_rrform_tpl_20260726.';
  END IF;
END
$guard2$;

COMMIT;

SELECT name, length(html_template) AS tpl_bytes,
       (html_template LIKE '%rr-capacity%') AS has_banner,
       (html_template LIKE '%action="/request"%') AS form_intact
FROM content_components WHERE name = 'report-request-form';
