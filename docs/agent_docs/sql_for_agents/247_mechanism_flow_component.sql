-- 247_mechanism_flow_component.sql
-- A section component that DRAWS A MECHANISM: a numbered flow of steps, with
-- optional decision branches, for explaining how a legal or financial process
-- actually works. Built for oufe.com, deliberately generic.
--
-- WHY THIS EXISTS. The owner's review of oufe.com asked for "infographics for
-- the more difficult concepts" and a layout that reads "less like a heavy text
-- book". The flagship Thames Water case page was ONE section — a single
-- generic-text-block — which is that complaint in concrete form. Nothing in the
-- estate draws a process: the only chart code is renderBarChartSVG in
-- platform/orchestration/actions/report_charts.go, unexported and bound to one
-- report page, and evidence-chart plots magnitudes, not sequence.
--
-- THREE DESIGN RULES, each from a recorded failure rather than taste:
--
-- 1. NO FIGURES, STRUCTURALLY. This component has no numeric field. A mechanism
--    diagram explains sequence and consequence, and needs no quantities to do
--    it. That matters on this site specifically: every figure must arrive
--    through the evidence register with a source (PLAN §C2, the 043
--    spec-poisoning precedent), and a component with a number-shaped slot is an
--    invitation for a writer to fill it. The absence of the slot is the control.
--
-- 2. TEXT IS HTML, NOT SVG. The obvious implementation is inline <svg> with
--    <text> labels. It is the wrong one here: text inside <svg> is INVISIBLE to
--    the claims gate (recorded landmine, brochure workstream), so a diagram
--    could assert anything and scan clean. Every label below is real HTML in the
--    normal flow, so claimscan, ScanBannedClaims and a screen reader all see the
--    same words. Connectors are drawn in CSS. There is no SVG and no dependency.
--
-- 3. CONTRAST MEASURED BEFORE WRITING, against the element's ACTUAL background.
--    Measured on oufe's live stylesheet (WCAG relative luminance):
--      --color-text   #E8E2D9 on --color-surface #1B2A3B = 11.32  (AA 4.5)
--      --color-text-muted #8A9BAE on surface            =  5.12  (AA 4.5)
--      --color-accent #C49A3C on surface                =  5.58  (AA 4.5)
--      --color-accent on --color-background #0F1820     =  6.86  (AA 4.5)
--      --color-border #2E3F52 on background             =  1.66  FAILS 3.0
--    So the rail and connectors use --color-accent, NOT --color-border. In a
--    mechanism diagram the connectors are graphical objects required to
--    understand the content, so the 3.0 non-text threshold applies to them and
--    --color-border cannot carry them. This is the same class of defect as
--    bugs_open/122, caught before shipping rather than after.
--
-- Variable names are checked against what the themes actually define. Fallbacks
-- are neutral/translucent, never light literals — a light fallback is how a
-- white card lands on a dark page (see the evidence-chart header comment).

BEGIN;

INSERT INTO content_components (
  name, function, display_name, component_level, render_mode, html_template,
  is_active, category, content_shape, visual_density, input_schema,
  description, created_at, updated_at
)
VALUES (
  'mechanism-flow',
  'mechanism-flow',
  'Mechanism flow',
  'section',
  'template',
  $tmpl$
<style>
  .mech-flow {
    padding: var(--spacing-xl, 4.5rem) var(--spacing-lg, 2rem);
    background: var(--color-background, #101820);
    color: var(--color-text, #e8e2d9);
  }
  .mech-flow__inner { max-width: 62rem; margin: 0 auto; }
  .mech-flow__eyebrow {
    display: block; font-size: 0.8125rem; font-weight: 600;
    letter-spacing: 0.1em; text-transform: uppercase;
    color: var(--color-accent, #c49a3c); margin: 0 0 0.5rem;
  }
  .mech-flow__title {
    font-size: clamp(1.5rem, 2.6vw, 2.1rem); font-weight: 700;
    line-height: 1.25; margin: 0 0 0.75rem;
  }
  .mech-flow__intro {
    max-width: 62ch; margin: 0 0 2.5rem;
    color: var(--color-text-muted, #8a9bae); line-height: 1.7;
  }

  /* The rail. Drawn in CSS so there is no SVG and nothing to load. */
  .mech-flow__steps {
    list-style: none; margin: 0; padding: 0; position: relative;
    counter-reset: mech;
  }
  .mech-flow__step {
    position: relative; padding: 0 0 2.25rem 3.25rem; counter-increment: mech;
  }
  .mech-flow__step:last-child { padding-bottom: 0; }
  /* the connector between markers */
  .mech-flow__step:not(:last-child)::before {
    content: ""; position: absolute; left: 1.09rem; top: 2.4rem; bottom: 0.35rem;
    width: 2px; background: var(--color-accent, #c49a3c); opacity: 0.45;
  }
  .mech-flow__marker {
    position: absolute; left: 0; top: 0.1rem;
    width: 2.25rem; height: 2.25rem; border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    font-size: 0.9rem; font-weight: 700; font-variant-numeric: tabular-nums;
    color: var(--color-accent, #c49a3c);
    border: 2px solid var(--color-accent, #c49a3c);
    background: var(--color-background, #101820);
  }
  /* Numbering is a CSS counter, NOT a template function: `inc`/`add` are not in
     the render funcmap, and a template that references a missing function fails
     to parse rather than degrading. An explicit "marker" in content_data
     overrides it for non-numeric sequences (A/B, 1a/1b). */
  .mech-flow__marker:not(.mech-flow__marker--custom)::before { content: counter(mech); }
  .mech-flow__step-title {
    font-size: 1.05rem; font-weight: 650; margin: 0.3rem 0 0.4rem;
    line-height: 1.35;
  }
  .mech-flow__step-body {
    margin: 0; line-height: 1.7; max-width: 60ch;
    color: var(--color-text, #e8e2d9);
  }
  .mech-flow__note {
    margin: 0.7rem 0 0; padding: 0.6rem 0.9rem;
    border-left: 3px solid var(--color-accent, #c49a3c);
    background: rgba(127, 127, 127, 0.12);
    font-size: 0.9375rem; line-height: 1.65; max-width: 60ch;
    color: var(--color-text-muted, #8a9bae);
  }

  /* A decision point: two or more outcomes side by side. */
  .mech-flow__branches {
    display: grid; gap: 0.85rem; margin: 0.9rem 0 0; max-width: 60ch;
    grid-template-columns: repeat(auto-fit, minmax(15rem, 1fr));
  }
  .mech-flow__branch {
    padding: 0.85rem 1rem; border-radius: var(--border-radius, 6px);
    background: var(--color-surface, rgba(127, 127, 127, 0.16));
    border: 1px solid var(--color-accent, #c49a3c);
  }
  .mech-flow__branch-label {
    display: block; font-size: 0.78rem; font-weight: 700;
    letter-spacing: 0.06em; text-transform: uppercase;
    color: var(--color-accent, #c49a3c); margin: 0 0 0.35rem;
  }
  .mech-flow__branch-body {
    margin: 0; font-size: 0.9375rem; line-height: 1.6;
    color: var(--color-text, #e8e2d9);
  }
  .mech-flow__footnote {
    margin: 2.25rem 0 0; padding-top: 1.1rem;
    border-top: 1px solid var(--color-accent, #c49a3c);
    font-size: 0.9rem; line-height: 1.65; max-width: 62ch;
    color: var(--color-text-muted, #8a9bae);
  }

  @media (max-width: 40rem) {
    .mech-flow { padding: 3rem 1.15rem; }
    .mech-flow__step { padding-left: 2.9rem; }
    .mech-flow__step:not(:last-child)::before { left: 0.97rem; }
    .mech-flow__marker { width: 2rem; height: 2rem; }
  }
</style>

<section id="{{.ComponentID}}" class="mech-flow" data-component="mechanism-flow">
  <div class="mech-flow__inner">
    {{if .eyebrow}}<span class="mech-flow__eyebrow">{{.eyebrow}}</span>{{end}}
    {{if .section_title}}<h2 class="mech-flow__title">{{.section_title}}</h2>{{end}}
    {{if .intro}}<p class="mech-flow__intro">{{.intro}}</p>{{end}}

    <ol class="mech-flow__steps">
      {{range $s := .steps}}
      <li class="mech-flow__step">
        <span class="mech-flow__marker{{if $s.marker}} mech-flow__marker--custom{{end}}" aria-hidden="true">{{if $s.marker}}{{$s.marker}}{{end}}</span>
        {{if $s.title}}<h3 class="mech-flow__step-title">{{$s.title}}</h3>{{end}}
        {{if $s.body}}<p class="mech-flow__step-body">{{$s.body}}</p>{{end}}
        {{if $s.note}}<p class="mech-flow__note">{{$s.note}}</p>{{end}}
        {{if $s.branches}}
        <div class="mech-flow__branches">
          {{range $s.branches}}
          <div class="mech-flow__branch">
            {{if .label}}<span class="mech-flow__branch-label">{{.label}}</span>{{end}}
            <p class="mech-flow__branch-body">{{.body}}</p>
          </div>
          {{end}}
        </div>
        {{end}}
      </li>
      {{end}}
    </ol>

    {{if .footnote}}<p class="mech-flow__footnote">{{.footnote}}</p>{{end}}
  </div>
</section>
$tmpl$,
  true,
  'explainer',
  'sequence',
  'medium',
  $schema${
    "type": "object",
    "required": ["steps"],
    "properties": {
      "eyebrow":       {"type": "string"},
      "section_title": {"type": "string"},
      "intro":         {"type": "string"},
      "footnote":      {"type": "string"},
      "steps": {
        "type": "array", "minItems": 2,
        "items": {
          "type": "object",
          "required": ["title"],
          "properties": {
            "marker": {"type": "string", "description": "optional override for the auto number"},
            "title":  {"type": "string"},
            "body":   {"type": "string"},
            "note":   {"type": "string", "description": "an aside, rendered as a callout"},
            "branches": {
              "type": "array",
              "description": "a decision point: two or more outcomes, rendered side by side",
              "items": {
                "type": "object",
                "required": ["body"],
                "properties": {
                  "label": {"type": "string"},
                  "body":  {"type": "string"}
                }
              }
            }
          }
        }
      }
    }
  }$schema$,
  'Draws a process as a numbered vertical flow with optional decision branches. For explaining a legal or financial MECHANISM (sequence and consequence), not magnitudes — use evidence-chart for those. Has NO numeric field by design: on an evidence-gated site a number-shaped slot invites a writer to fill it. All labels are real HTML, never SVG text, so the claims gate can read them. Connectors are CSS, so there is no SVG and nothing to load.',
  now(), now()
)
-- `name` is the unique key here. `function` has only a PARTIAL unique index
-- (component_level='tool' AND forked_from IS NULL AND is_active), so
-- ON CONFLICT (function) raises 42P10 rather than upserting. Checked with
-- \d content_components before writing this, after being bitten by exactly this
-- class of guess before (agent_definitions has display_name, not name).
ON CONFLICT (name) DO UPDATE
  SET html_template = EXCLUDED.html_template,
      render_mode   = EXCLUDED.render_mode,
      input_schema  = EXCLUDED.input_schema,
      description   = EXCLUDED.description,
      is_active     = true,
      updated_at    = now();

COMMIT;

-- VERIFY (run after applying):
--   SELECT function, render_mode, component_level, length(html_template)
--   FROM content_components WHERE function = 'mechanism-flow';
--
-- Numbering is a CSS counter. `inc`/`add` are NOT in the render funcmap (checked:
-- the only FuncMaps are render_css_from_spec_action.go:238 and
-- compute_component_quality.go:354), and a template referencing a missing
-- function fails to PARSE — it does not degrade. Verified before applying.
