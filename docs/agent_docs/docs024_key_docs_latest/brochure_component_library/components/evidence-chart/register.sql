\set ON_ERROR_STOP on
-- evidence-chart — shared, evidence-sourced chart component.
-- GENERATED from components/evidence-chart/{template.html,input_schema.json}
-- by scripts/gen_component_register_sql.py. Edit those files and regenerate;
-- do not hand-edit this file.
--
-- The guarantee this component exists to make: the ONLY place a figure lives is
-- site_specs.evidence_base.facts. A chart definition names fact ids and never
-- restates a value, and both fields are system-resolved — resolved data beats
-- LLM content at render time, so the writer cannot supply a number even if it
-- tries. No evidence_base charts => the section is skipped, not invented.
BEGIN;
INSERT INTO content_components
  (id, name, function, display_name, description, category, semantic_tags,
   section_type, component_level, render_mode, is_dark_section, is_active,
   suitable_site_types, suitable_page_types, html_template, input_schema)
VALUES (
  gen_random_uuid(),
  'evidence-chart','evidence-chart','Evidence Chart',
  'Code-rendered bar charts whose values come from the site''s evidence_base register, never from the model. Chart definitions name fact ids; the figures, units and verified dates are read from the audited fact rows. Bars are drawn in CSS from the real value; the label and figure are real selectable text, so screen readers and the claims gate both see the number. Skipped entirely on a site with no audited series.',
  'data','["chart","data","evidence","code-rendered","brochure","infographic"]'::jsonb,
  'evidence-chart','section','agent',false,true,
  '["brochure","consultancy","professional-services","b2b"]'::jsonb,
  '["index","home","about","capabilities","landing","content"]'::jsonb,
  $HTML$<style>
  .evidence-chart {
    padding: var(--spacing-section, 4.5rem 2rem);
    background: var(--color-background, #ffffff);
    color: var(--color-text, #1a1a1a);
  }
  .evidence-chart__inner {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
  }
  .evidence-chart__header {
    max-width: 62ch;
    margin: 0 0 2.75rem;
  }
  .evidence-chart__eyebrow {
    display: block;
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--color-primary, #1e40af);
  }
  .evidence-chart__title {
    font-size: clamp(1.5rem, 2.6vw, 2.1rem);
    font-weight: 700;
    line-height: 1.25;
    margin: 0.5rem 0 0;
  }
  .evidence-chart__intro {
    margin: 0.75rem 0 0;
    line-height: 1.6;
    color: var(--color-text-muted, #555);
  }
  .evidence-chart__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 2.5rem;
  }
  .evidence-chart__figure {
    margin: 0;
    padding: 1.5rem 1.5rem 1.25rem;
    background: var(--color-surface, var(--color-background-alt, #f6f7f9));
    border: 1px solid color-mix(in srgb, var(--color-text, #1a1a1a) 12%, transparent);
    border-radius: var(--border-radius, 10px);
  }
  .evidence-chart__figcaption {
    display: block;
    margin: 0 0 1.25rem;
  }
  .evidence-chart__chart-title {
    display: block;
    font-size: 1.0625rem;
    font-weight: 700;
    line-height: 1.35;
  }
  .evidence-chart__chart-note {
    display: block;
    margin-top: 0.35rem;
    font-size: 0.875rem;
    line-height: 1.5;
    color: var(--color-text-muted, #555);
  }
  .evidence-chart__row {
    display: grid;
    grid-template-columns: minmax(9ch, 30%) 1fr auto;
    align-items: center;
    gap: 0.75rem;
    padding: 0.55rem 0;
  }
  .evidence-chart__row + .evidence-chart__row {
    border-top: 1px solid color-mix(in srgb, var(--color-text, #1a1a1a) 8%, transparent);
  }
  .evidence-chart__label {
    font-size: 0.9375rem;
    line-height: 1.35;
  }
  .evidence-chart__track {
    display: block;
    height: 1.35rem;
    border-radius: 3px;
    background: color-mix(in srgb, var(--color-text, #1a1a1a) 8%, transparent);
    overflow: hidden;
  }
  /* Geometry is computed by the browser from the real value: --v is the
     figure itself and --m the chart's declared maximum. Nothing rounds it
     into the template, and there is no width to get wrong by hand. */
  .evidence-chart__bar {
    display: block;
    height: 100%;
    width: calc(100% * var(--v, 0) / var(--m, 1));
    min-width: 2px;
    border-radius: 3px;
    background: var(--color-primary, #1e40af);
  }
  .evidence-chart__bar--muted {
    background: color-mix(in srgb, var(--color-text, #1a1a1a) 32%, transparent);
  }
  .evidence-chart__bar--accent {
    background: var(--color-accent, var(--color-secondary, #0f766e));
  }
  .evidence-chart__value {
    font-size: 1.0625rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }
  .evidence-chart__verified {
    grid-column: 1 / -1;
    margin: 0;
    font-size: 0.75rem;
    color: var(--color-text-muted, #666);
  }
  .evidence-chart__source {
    margin: 1.1rem 0 0;
    padding-top: 0.85rem;
    border-top: 1px solid color-mix(in srgb, var(--color-text, #1a1a1a) 12%, transparent);
    font-size: 0.8125rem;
    line-height: 1.5;
    color: var(--color-text-muted, #555);
  }
  @media (max-width: 620px) {
    .evidence-chart { padding: 3.25rem 1.25rem; }
    .evidence-chart__row { grid-template-columns: 1fr auto; }
    .evidence-chart__track { grid-column: 1 / -1; order: 3; }
  }
</style>

{{- $facts := .facts -}}
{{- $page := .current_page -}}
<section class="evidence-chart" data-component="evidence-chart">
  <div class="evidence-chart__inner">
    {{if or .section_title .section_eyebrow .section_intro}}
    <header class="evidence-chart__header">
      {{if .section_eyebrow}}<span class="evidence-chart__eyebrow">{{.section_eyebrow}}</span>{{end}}
      {{if .section_title}}<h2 class="evidence-chart__title">{{.section_title}}</h2>{{end}}
      {{if .section_intro}}<p class="evidence-chart__intro">{{.section_intro}}</p>{{end}}
    </header>
    {{end}}
    <div class="evidence-chart__grid">
      {{range $c := .charts}}
      {{- $show := true -}}
      {{- if and $page $c.pages -}}
        {{- $show = false -}}
        {{- range $c.pages -}}
          {{- if or (eq . $page) (eq (printf "%s.html" .) $page) -}}{{- $show = true -}}{{- end -}}
        {{- end -}}
      {{- end -}}
      {{if $show}}
      {{- /* The denominator: a scale constant (max, e.g. 100 for a percentage)
             or, preferred, another fact (max_fact_id) so a business figure is
             never restated inside a chart definition. */ -}}
      {{- $max := $c.max -}}
      {{- if $c.max_fact_id -}}
        {{- range $f := $facts}}{{if eq $f.id $c.max_fact_id}}{{$max = $f.value}}{{end}}{{end -}}
      {{- end -}}
      <figure class="evidence-chart__figure" data-chart="{{$c.id}}">
        <figcaption class="evidence-chart__figcaption">
          <span class="evidence-chart__chart-title">{{$c.title}}</span>
          {{if $c.caption}}<span class="evidence-chart__chart-note">{{$c.caption}}</span>{{end}}
        </figcaption>
        {{range $p := $c.points}}
          {{- $pid := $p.fact_id -}}
          {{- range $f := $facts -}}
            {{- if eq $f.id $pid}}
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">{{$p.label}}</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar{{if eq $p.tone "muted"}} evidence-chart__bar--muted{{else if eq $p.tone "accent"}} evidence-chart__bar--accent{{end}}" style="--v:{{printf "%.4f" $f.value}};--m:{{printf "%.4f" $max}}"></span></span>
          <span class="evidence-chart__value">{{if $f.display}}{{$f.display}}{{else}}{{printf "%.10g" $f.value}}{{end}}{{$c.unit}}</span>
          <span class="evidence-chart__verified">verified {{$f.verified_at}}</span>
        </div>
            {{- end -}}
          {{- end -}}
        {{end}}
        {{if $c.source_note}}<p class="evidence-chart__source">{{$c.source_note}}</p>{{end}}
      </figure>
      {{end}}
      {{end}}
    </div>
  </div>
</section>
$HTML$,
  $SCHEMA${
  "fields": {
    "charts": {
      "type": "array",
      "source": "site_specs.evidence_base.charts",
      "required": true,
      "on_missing": "skip_section",
      "missing_reason": "no chart definitions in this site's evidence_base \u2014 a chart with no audited series must not be drawn",
      "items": {
        "id": {
          "type": "text"
        },
        "title": {
          "type": "text"
        },
        "caption": {
          "type": "text"
        },
        "unit": {
          "type": "text"
        },
        "max": {
          "type": "number"
        },
        "pages": {
          "type": "array"
        },
        "source_note": {
          "type": "text"
        },
        "points": {
          "type": "array"
        }
      }
    },
    "facts": {
      "type": "array",
      "source": "site_specs.evidence_base.facts",
      "required": true,
      "on_missing": "skip_section",
      "missing_reason": "no evidence_base facts \u2014 the register is the only place a charted figure may come from"
    },
    "section_eyebrow": {
      "type": "text",
      "source": "llm",
      "required": false,
      "llm_guidance": "Short uppercase eyebrow, 2-4 words, e.g. 'Evidenced, not asserted'. Optional."
    },
    "section_title": {
      "type": "text",
      "source": "llm",
      "required": false,
      "llm_guidance": "Short heading for the chart section, under 10 words. Describe what the charts show WITHOUT stating any figure \u2014 the figures are rendered by the system from the site's evidence register and must never be written here."
    },
    "section_intro": {
      "type": "text",
      "source": "llm",
      "required": false,
      "llm_guidance": "One or two plain sentences introducing the charts. NEVER state, round, summarise or preview a number: every figure on this section is drawn by the system from an audited fact row, and a number written here is by definition unverified. Write about what is being measured and why it matters, not how big it is."
    }
  }
}$SCHEMA$::jsonb
);
COMMIT;

SELECT function, section_type, component_level, is_active,
       length(html_template) AS template_bytes,
       jsonb_object_keys(input_schema->'fields') AS field
  FROM content_components WHERE function = 'evidence-chart';
