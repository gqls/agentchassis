-- 250_evidence_timeseries_component.sql
-- `evidence-timeseries`: plots a DATED series from the evidence register.
--
-- The companion to `evidence-chart`. That one compares magnitudes between things;
-- this one shows one measurement moving over time. Both resolve every plotted
-- value through a fact id, so neither can carry a number of its own.
--
-- WHY IT COULD NOT EXIST BEFORE. An EvidenceFact held one Value and three dates,
-- all provenance (accessed / published / verified_at). None was the date the value
-- APPLIES TO, so there was no honest way to place a point on a time axis. The
-- substrate landed first, deliberately: a line chart with no legitimate series to
-- plot is an invitation for a writer to fill it from the model, which is the exact
-- failure the evidence layer exists to prevent. See
-- platform/orchestration/datahelpers/claims_series.go.
--
-- THREE CONSTRAINTS THAT SHAPED THIS, each measured rather than assumed:
--
-- 1. NO ARITHMETIC IS AVAILABLE IN THE TEMPLATE. `inc`/`add` are not in the render
--    funcmap (the only FuncMaps are render_css_from_spec_action.go:238 and
--    compute_component_quality.go:354), and a template referencing a missing
--    function fails to PARSE rather than degrading. That rules out computing SVG
--    polyline coordinates. So this is a COLUMN chart in CSS: each observation is a
--    column whose height is set by two custom properties, and the browser does the
--    division. Exactly the trick evidence-chart already uses for its bars.
--
-- 2. TEXT MUST BE HTML, NEVER SVG. Text inside <svg> is invisible to the claims
--    gate, so a chart could assert anything and scan clean. Every label, value and
--    date below is real HTML in the normal flow. Verified on the sibling
--    mechanism-flow component by decoding the payload claimscan actually received.
--
-- 3. CONTRAST MEASURED AGAINST THE ELEMENT'S OWN BACKGROUND, before writing. On
--    oufe's live stylesheet --color-border (#2E3F52) scores 1.66 on the page
--    background and FAILS the 3.0 non-text threshold, so the axis and columns use
--    --color-accent (6.86). Chart furniture is a graphical object required to
--    understand the content, so 3.0 genuinely applies to it. cmd/contrastscan is
--    the tool; bugs_open/122 is why.
--
-- THE SCALE DENOMINATOR follows evidence-chart's established rule: prefer
-- `max_fact_id` (another registered fact) so a business figure is never restated
-- inside a chart definition. A bare `max` is permitted only as a scale constant
-- (100 for a percentage), never as a quantity about the world.

BEGIN;

INSERT INTO content_components (
  name, function, display_name, component_level, render_mode, html_template,
  is_active, category, content_shape, visual_density, input_schema,
  description, created_at, updated_at
)
VALUES (
  'evidence-timeseries',
  'evidence-timeseries',
  'Evidence time series',
  'section',
  'template',
  $tmpl$
<style>
  .ev-ts { padding: var(--spacing-xl, 4.5rem) var(--spacing-lg, 2rem);
           background: var(--color-background, #101820); color: var(--color-text, #e8e2d9); }
  .ev-ts__inner { max-width: 62rem; margin: 0 auto; }
  .ev-ts__eyebrow { display: block; font-size: 0.8125rem; font-weight: 600;
    letter-spacing: 0.1em; text-transform: uppercase;
    color: var(--color-accent, #c49a3c); margin: 0 0 0.5rem; }
  .ev-ts__title { font-size: clamp(1.5rem, 2.6vw, 2.1rem); font-weight: 700;
    line-height: 1.25; margin: 0 0 0.75rem; }
  .ev-ts__intro { max-width: 62ch; margin: 0 0 2.5rem; line-height: 1.7;
    color: var(--color-text-muted, #8a9bae); }

  .ev-ts__figure { margin: 0 0 2.5rem; }
  .ev-ts__caption { display: block; margin: 0 0 1rem; }
  .ev-ts__label { display: block; font-size: 1.02rem; font-weight: 650; }
  .ev-ts__note { display: block; font-size: 0.9rem; line-height: 1.6;
    color: var(--color-text-muted, #8a9bae); margin-top: 0.2rem; }

  /* The plot. Heights come from --v (value) and --m (scale max); the browser
     does the division, because the template cannot do arithmetic. */
  .ev-ts__plot { display: flex; align-items: flex-end; gap: 0.5rem;
    min-height: 11rem; padding: 0 0 0.5rem;
    border-bottom: 2px solid var(--color-accent, #c49a3c);
    overflow-x: auto; }
  .ev-ts__col { flex: 1 1 0; min-width: 2.75rem; display: flex;
    flex-direction: column; justify-content: flex-end; align-items: center; gap: 0.35rem; }
  .ev-ts__bar { width: 100%;
    height: calc(10rem * var(--v) / var(--m));
    min-height: 2px;
    background: var(--color-accent, #c49a3c); opacity: 0.85; border-radius: 2px 2px 0 0; }
  .ev-ts__val { font-size: 0.8rem; font-variant-numeric: tabular-nums;
    color: var(--color-text, #e8e2d9); white-space: nowrap; }

  .ev-ts__axis { display: flex; gap: 0.5rem; margin: 0.4rem 0 0; }
  .ev-ts__tick { flex: 1 1 0; min-width: 2.75rem; text-align: center;
    font-size: 0.72rem; letter-spacing: 0.02em;
    color: var(--color-text-muted, #8a9bae); white-space: nowrap; }

  /* Provenance travels with the chart. A plotted point with no visible source is
     the thing this component exists to make impossible. */
  .ev-ts__sources { margin: 0.9rem 0 0; padding: 0.7rem 0.95rem;
    background: rgba(127,127,127,0.12);
    border-left: 3px solid var(--color-accent, #c49a3c);
    font-size: 0.85rem; line-height: 1.65; color: var(--color-text-muted, #8a9bae); }
  .ev-ts__sources ul { margin: 0.4rem 0 0; padding-left: 1.1rem; }
  .ev-ts__sources a { color: var(--color-accent, #c49a3c); }
  .ev-ts__footnote { margin: 2rem 0 0; padding-top: 1rem;
    border-top: 1px solid var(--color-accent, #c49a3c);
    font-size: 0.9rem; line-height: 1.65; max-width: 62ch;
    color: var(--color-text-muted, #8a9bae); }

  @media (max-width: 40rem) {
    .ev-ts { padding: 3rem 1.15rem; }
    .ev-ts__plot { min-height: 9rem; }
    .ev-ts__bar { height: calc(8rem * var(--v) / var(--m)); }
  }
</style>

{{- /* $facts is NOT a built-in: Go templates treat an undeclared variable as a
       PARSE error, not a runtime one, so the whole component would fail to render
       rather than degrade. evidence-chart declares it the same way; this is that
       contract, not a new one. Caught by executing the template before shipping. */ -}}
{{- $facts := .facts -}}
<section id="{{.ComponentID}}" class="ev-ts" data-component="evidence-timeseries">
  <div class="ev-ts__inner">
    {{if .eyebrow}}<span class="ev-ts__eyebrow">{{.eyebrow}}</span>{{end}}
    {{if .section_title}}<h2 class="ev-ts__title">{{.section_title}}</h2>{{end}}
    {{if .intro}}<p class="ev-ts__intro">{{.intro}}</p>{{end}}

    {{range $s := .series}}
      {{- /* Resolve the scale. Prefer another registered fact; a bare max is a
             scale constant only. */ -}}
      {{- $max := $s.max -}}
      {{- if $s.max_fact_id -}}
        {{- range $f := $facts}}{{if eq $f.id $s.max_fact_id}}{{$max = $f.value}}{{end}}{{end -}}
      {{- end -}}
      {{- range $f := $facts}}{{if eq $f.id $s.fact_id}}
      <figure class="ev-ts__figure" data-series="{{$s.fact_id}}">
        <figcaption class="ev-ts__caption">
          <span class="ev-ts__label">{{if $s.label}}{{$s.label}}{{else}}{{$f.claim}}{{end}}</span>
          {{if $s.note}}<span class="ev-ts__note">{{$s.note}}</span>{{end}}
        </figcaption>

        <div class="ev-ts__plot">
          {{range $o := $f.observations}}
          <div class="ev-ts__col">
            <span class="ev-ts__val">{{if $o.display}}{{$o.display}}{{else}}{{printf "%.10g" $o.value}}{{end}}{{$s.unit}}</span>
            <span class="ev-ts__bar" style="--v:{{printf "%.4f" $o.value}};--m:{{printf "%.4f" $max}}" aria-hidden="true"></span>
          </div>
          {{end}}
        </div>
        <div class="ev-ts__axis">
          {{range $o := $f.observations}}<span class="ev-ts__tick">{{$o.as_of}}</span>{{end}}
        </div>

        <div class="ev-ts__sources">
          Every point above is a separately sourced observation. Each carries the date the
          figure applies to, and where we read it:
          <ul>
            {{range $o := $f.observations}}
            <li>{{$o.as_of}} —
              {{if $o.source}}{{with $o.source.citation}}<a href="{{.url}}" rel="noopener noreferrer">{{if .title}}{{.title}}{{else}}{{.publisher}}{{end}}</a>{{if .accessed}}, read {{.accessed}}{{end}}{{end}}{{end}}
              {{if $o.verified_at}}(last checked {{$o.verified_at}}){{end}}
            </li>
            {{end}}
          </ul>
        </div>
      </figure>
      {{end}}{{end}}
    {{end}}

    {{if .footnote}}<p class="ev-ts__footnote">{{.footnote}}</p>{{end}}
  </div>
</section>
$tmpl$,
  true,
  'evidence',
  'series',
  'medium',
  $schema${
    "type": "object",
    "required": ["series"],
    "properties": {
      "eyebrow":       {"type": "string"},
      "section_title": {"type": "string"},
      "intro":         {"type": "string"},
      "footnote":      {"type": "string"},
      "series": {
        "type": "array", "minItems": 1,
        "items": {
          "type": "object",
          "required": ["fact_id"],
          "properties": {
            "fact_id":     {"type": "string", "description": "id of a kind=series fact; its observations are the points"},
            "label":       {"type": "string", "description": "defaults to the fact's own claim"},
            "note":        {"type": "string"},
            "unit":        {"type": "string", "description": "suffix shown after each value, e.g. %"},
            "max_fact_id": {"type": "string", "description": "PREFERRED scale denominator: another registered fact"},
            "max":         {"type": "number", "description": "scale constant ONLY (e.g. 100 for a percentage), never a quantity about the world"}
          }
        }
      }
    }
  }$schema$,
  'Plots a dated series from the evidence register: one measurement over time, one column per observation. Companion to evidence-chart, which compares magnitudes instead. Every point resolves through a fact id and every point renders its own source and date beneath the plot, because a series whose points are individually sourced is the only kind this platform may publish. No arithmetic in the template (none is available) and no SVG text (it is invisible to the claims gate).',
  now(), now()
)
ON CONFLICT (name) DO UPDATE
  SET html_template = EXCLUDED.html_template,
      render_mode   = EXCLUDED.render_mode,
      input_schema  = EXCLUDED.input_schema,
      description   = EXCLUDED.description,
      is_active     = true,
      updated_at    = now();

COMMIT;

SELECT name, function, render_mode, component_level, length(html_template) AS bytes
FROM content_components WHERE name = 'evidence-timeseries';
