\set ON_ERROR_STOP on
BEGIN;
INSERT INTO content_components
  (id, name, function, display_name, description, category, semantic_tags,
   section_type, component_level, render_mode, is_dark_section, is_active,
   suitable_site_types, suitable_page_types, html_template, input_schema)
VALUES (
  gen_random_uuid(),
  'stat-band','stat-band','Stat Band',
  'A band of 3-4 code-rendered stat callouts (value, label, optional caption) with an accessible scroll-triggered count-up. Numbers are rendered from real, evidenced figures — never generated as a picture.',
  'stats','["stats","metrics","countup","code-rendered","brochure"]'::jsonb,
  'stat-band','section','agent',true,true,
  '["brochure","consultancy","professional-services","b2b"]'::jsonb,
  '["index","home","about","capabilities","landing"]'::jsonb,
  $HTML$
<style>
  .stat-band {
    padding: var(--spacing-section, 4.5rem 2rem);
    background: var(--color-primary);
    color: var(--color-primary-text, #fff);
  }
  .stat-band__inner {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
  }
  .stat-band__header {
    text-align: center;
    max-width: 60ch;
    margin: 0 auto 2.5rem;
  }
  .stat-band__eyebrow {
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    opacity: 0.85;
  }
  .stat-band__title {
    font-size: clamp(1.5rem, 2.6vw, 2.1rem);
    font-weight: 700;
    line-height: 1.25;
    margin: 0.5rem 0 0;
  }
  .stat-band__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 2rem 1.5rem;
    text-align: center;
  }
  .stat-band__stat {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    padding: 0 0.5rem;
  }
  .stat-band__stat + .stat-band__stat {
    border-inline-start: 1px solid color-mix(in srgb, var(--color-primary-text, #fff) 22%, transparent);
  }
  @media (max-width: 620px) {
    .stat-band__stat + .stat-band__stat { border-inline-start: 0; }
  }
  .stat-band__value {
    font-size: clamp(2.5rem, 5vw, 3.5rem);
    font-weight: 800;
    line-height: 1;
    letter-spacing: -0.02em;
    font-variant-numeric: tabular-nums;
  }
  .stat-band__label {
    font-size: 1rem;
    font-weight: 600;
  }
  .stat-band__caption {
    font-size: 0.8125rem;
    opacity: 0.8;
    line-height: 1.5;
  }
  @media (max-width: 768px) { .stat-band { padding: 3.25rem 1.25rem; } }
</style>

<section class="stat-band" data-component="stat-band">
  <div class="stat-band__inner">
    {{if or .section_title .section_eyebrow}}
    <header class="stat-band__header">
      {{if .section_eyebrow}}<span class="stat-band__eyebrow">{{.section_eyebrow}}</span>{{end}}
      {{if .section_title}}<h2 class="stat-band__title">{{.section_title}}</h2>{{end}}
    </header>
    {{end}}
    <div class="stat-band__grid">
      {{range .stats}}
      <div class="stat-band__stat">
        <span class="stat-band__value" data-countup aria-label="{{.value}}">{{.value}}</span>
        <span class="stat-band__label">{{.label}}</span>
        {{if .caption}}<span class="stat-band__caption">{{.caption}}</span>{{end}}
      </div>
      {{end}}
    </div>
  </div>
</section>
$HTML$,
  $SCHEMA${
  "fields": {
    "section_eyebrow": { "type": "text", "source": "llm", "required": false, "llm_guidance": "Short uppercase eyebrow, e.g. 'By the numbers'. Optional." },
    "section_title": { "type": "text", "source": "llm", "required": false, "llm_guidance": "Optional short heading for the stat band. Under 9 words." },
    "stats": {
      "type": "array", "source": "llm", "required": true,
      "items": { "value": {"type":"text"}, "label": {"type":"text"}, "caption": {"type":"text"} },
      "llm_guidance": "3 or 4 stats. Each: value (the figure exactly as it should read, e.g. '97%', '11', '<24h', '£29' — it is code-rendered and count-up animated, so it must be a REAL, EVIDENCED number; NEVER invent, round up, or estimate a figure here — if you do not have a verified number, do not add the stat), label (what it measures, 2-5 words), caption (a one-line source or context, optional). Prefer figures that are demonstrably true about the platform's own work."
    }
  }
}
$SCHEMA$::jsonb
);
INSERT INTO js_snippets (id,name,description,js_content,applies_to,semantic_tags,is_active) VALUES (gen_random_uuid(),'stat-band','Accessible count-up for stat-band values.',$JS$/* stat-band — accessible count-up for code-rendered stats.
   Animates the numeric part of a stat into view, then restores the exact
   authored value. The real value is always in the DOM + on aria-label, so
   screen readers read the true figure; the count-up is purely visual and is
   skipped under prefers-reduced-motion. Never invents digits — it only counts
   up to whatever real value was rendered. Idempotent, no dependencies. */
(function () {
  "use strict";
  if (window.__statBandInit) return;
  window.__statBandInit = true;
  var reduce = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  function animate(el) {
    var original = el.getAttribute("data-final");
    if (original === null) { original = el.textContent.trim(); el.setAttribute("data-final", original); }
    // Match: optional prefix, one number (with thousands/decimals), optional suffix.
    var m = original.match(/^(\D*)(\d[\d,]*(?:\.\d+)?)(\D*)$/);
    if (!m || reduce) { el.textContent = original; return; }
    var prefix = m[1], numStr = m[2].replace(/,/g, ""), suffix = m[3];
    var target = parseFloat(numStr);
    var decimals = (numStr.split(".")[1] || "").length;
    var dur = 1100, start = null;
    function fmt(v) {
      return prefix + v.toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, ",") + suffix;
    }
    function step(ts) {
      if (!start) start = ts;
      var p = Math.min((ts - start) / dur, 1);
      var eased = 1 - Math.pow(1 - p, 3);
      if (p < 1) { el.textContent = fmt(target * eased); requestAnimationFrame(step); }
      else { el.textContent = original; } // restore exact authored value
    }
    el.textContent = fmt(0);
    requestAnimationFrame(step);
  }

  function init() {
    var els = Array.prototype.slice.call(document.querySelectorAll(".stat-band [data-countup]"));
    if (!els.length) return;
    if (!("IntersectionObserver" in window)) { els.forEach(animate); return; }
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (en) {
        if (en.isIntersecting) { animate(en.target); io.unobserve(en.target); }
      });
    }, { threshold: 0.4 });
    els.forEach(function (el) { io.observe(el); });
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", init);
  else init();
})();
$JS$,'["stat-band"]'::jsonb,'["countup","stats","accessibility"]'::jsonb,true);
COMMIT;
SELECT function, is_active FROM content_components WHERE function='stat-band'; SELECT name FROM js_snippets WHERE name='stat-band';
