-- 606_comparison_table_component.sql
-- A generic COMPARISON TABLE: several options compared across the same few
-- writer-named criteria, rendered as a real <table>. Built for bugs_open/381,
-- deliberately generic.
--
-- WHY THIS EXISTS. `[MEASURED 2026-08-24]` garden-tools.uk carried **zero** <table>
-- elements across all seven served pages — on a site whose own classification is a
-- BUYING-GUIDE COMPARISON HUB, a vertical whose entire value is structured
-- comparison. Fleet-wide the same census found 64 of 741 pages with any table at
-- all. The library's only table-capable generic component was `pricing` (pricing
-- tiers, brochure sites); everything else was a directory, a calculator or a
-- site-specific spec sheet. So the planner could not choose a comparison even when
-- comparison was the page's whole job.
--
-- ⚠ THIS IS THE RISKIEST OF THE THREE COMPONENTS AND THE SCHEMA REFLECTS THAT.
-- A comparison table is where a writer most wants to invent: prices, ratings,
-- weights, star scores, "best value" badges. bugs_open/380 (same build, same day)
-- found practice claims and ungated assertions on this very site, and 29 of 48 live
-- sites have NO evidence base at all, so the claims audit has never run on them.
-- Design rule 1 of 247 says the ABSENCE of a slot is the control — but a comparison
-- component with no comparable cells is not a comparison component, so the absence
-- rule cannot be applied wholesale here. What is done instead, deliberately:
--   * NO price, rating, score, rank or "winner" field exists. Not one. A table that
--     cannot render a star rating cannot publish an invented one.
--   * Cells are free text under the writer prompt's existing rule 14 (never state a
--     figure not given in THIS prompt) and, where a site has one, its evidence
--     register — the same regime as every other text field on the estate.
--   * Every cell's guidance says so explicitly, at the point of writing.
--   * A `source_note` field exists so an unsourced comparison is VISIBLE rather than
--     merely disallowed. It is not a control and is not claimed as one.
-- This is stated plainly rather than presented as solved: a prompt instruction is
-- not an enforcement mechanism (owner ruling), and the real control for figures on
-- a site with no evidence base is bugs_open/380's work, not this component's schema.
--
-- NAMED CELLS, NOT A POSITIONAL GRID — and this is a reliability decision, not a
-- style one. The obvious schema is `columns: [...]` plus `rows: [{cells: [...]}]`,
-- i.e. a 2D array. An LLM that emits four columns and one row of three cells
-- produces a table that is silently misaligned — every subsequent cell shifts, and
-- the DB row looks perfectly valid. So each row carries NAMED cells against
-- writer-named column labels: a missing cell renders as an empty cell in the right
-- place, never a shifted table. The cost is a fixed maximum of four columns, which
-- is also the most a table can carry on a phone.
--
-- RESPONSIVE BY STACKING, because the owner complained about exactly this. A wide
-- table on a phone either overflows or shrinks to unreadable. Below 40rem each row
-- becomes its own block and each cell is prefixed with its column label via
-- `data-label` + CSS ::before — the standard pattern, and the reason the column
-- labels are duplicated onto every cell in the markup.
--
-- OTHER DESIGN RULES: <h3> for row names is NOT used (rows are table cells, and a
-- heading inside a <td> breaks the table semantics); the section owns one <h2>; no
-- <svg> (text inside one is invisible to the claims gate); theme variables with
-- neutral/translucent fallbacks and no reliance on hairline contrast for
-- comprehension; `{{.InstanceID}}` from birth (RFC_032).
--
-- ⚠ content_shape is set and is currently READ BY NOTHING — see 604's header.
--
-- ROLLBACK: 606_comparison_table_component_ROLLBACK.sql

BEGIN;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM content_components WHERE function = 'comparison-table' AND is_active) THEN
    RAISE EXCEPTION '606: an active component with function=comparison-table already exists — refusing to double-apply or shadow it';
  END IF;
END $$;

INSERT INTO content_components (
  name, function, display_name, component_level, render_mode, section_type,
  html_template, is_active, category, content_shape, visual_density,
  suitable_site_types, suitable_page_types, semantic_tags,
  input_schema, description, created_from, created_at, updated_at
)
VALUES (
  'comparison-table',
  'comparison-table',
  'Comparison table',
  'section',
  'template',
  'comparison-table',
  $tmpl$
<style>
  .cmp-table { padding: var(--spacing-section, 4.5rem 2rem); background: var(--color-background); color: var(--color-text); }
  .cmp-table__inner { max-width: var(--container-max-width, 68rem); margin: 0 auto; }
  .cmp-table__header { max-width: 46rem; margin: 0 0 2.25rem; }
  .cmp-table__eyebrow {
    display: block; font-size: 0.8125rem; font-weight: 600;
    letter-spacing: 0.1em; text-transform: uppercase;
    color: var(--color-primary-ink, var(--color-accent, currentColor)); margin: 0 0 0.5rem;
  }
  .cmp-table__title {
    font-size: clamp(1.5rem, 2.6vw, 2.1rem); font-weight: 700;
    line-height: 1.25; margin: 0 0 0.75rem; color: var(--color-heading, inherit);
  }
  .cmp-table__intro { margin: 0; line-height: 1.7; color: var(--color-text-muted, inherit); }
  .cmp-table__scroll { overflow-x: auto; -webkit-overflow-scrolling: touch; }
  .cmp-table__table { width: 100%; border-collapse: collapse; text-align: left; }
  .cmp-table__table th, .cmp-table__table td {
    padding: 0.85rem 1rem; vertical-align: top; line-height: 1.55;
    border-bottom: 1px solid var(--color-border, rgba(128,128,128,0.25));
  }
  .cmp-table__table thead th {
    font-size: 0.8125rem; font-weight: 700; letter-spacing: 0.06em;
    text-transform: uppercase; color: var(--color-text-muted, inherit);
    border-bottom-width: 2px;
  }
  .cmp-table__table tbody th {
    font-weight: 700; color: var(--color-heading, inherit); width: 22%;
  }
  .cmp-table__table tbody tr:last-child th,
  .cmp-table__table tbody tr:last-child td { border-bottom: 0; }
  .cmp-table__note { display: block; margin-top: 0.3rem; font-size: 0.875rem; color: var(--color-text-muted, inherit); }
  .cmp-table__source {
    margin: 1.5rem 0 0; font-size: 0.875rem; line-height: 1.6;
    color: var(--color-text-muted, inherit);
  }
  /* Stack on narrow screens: each row becomes a block, each cell keeps its label. */
  @media (max-width: 40rem) {
    .cmp-table__table, .cmp-table__table tbody, .cmp-table__table tr,
    .cmp-table__table th, .cmp-table__table td { display: block; width: auto; }
    .cmp-table__table thead { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
    .cmp-table__table tbody tr {
      padding: 1rem 0; border-bottom: 1px solid var(--color-border, rgba(128,128,128,0.25));
    }
    .cmp-table__table tbody tr:last-child { border-bottom: 0; }
    .cmp-table__table tbody th { padding: 0 0 0.5rem; font-size: 1.0625rem; border: 0; }
    .cmp-table__table tbody td { padding: 0.3rem 0; border: 0; }
    .cmp-table__table tbody td::before {
      content: attr(data-label); display: block;
      font-size: 0.75rem; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase;
      color: var(--color-text-muted, inherit); margin-bottom: 0.1rem;
    }
    .cmp-table__table tbody td:empty { display: none; }
  }
</style>
<section id="{{.InstanceID}}" class="cmp-table" data-component="comparison-table">
  <div class="cmp-table__inner">
    <header class="cmp-table__header">
      {{if .eyebrow}}<span class="cmp-table__eyebrow">{{.eyebrow}}</span>{{end}}
      <h2 class="cmp-table__title">{{if .section_title}}{{.section_title}}{{end}}</h2>
      {{if .intro}}<p class="cmp-table__intro">{{.intro}}</p>{{end}}
    </header>
    <div class="cmp-table__scroll">
      <table class="cmp-table__table">
        <thead>
          <tr>
            <th scope="col">{{if .option_column_label}}{{.option_column_label}}{{end}}</th>
            <th scope="col">{{if .column_two_label}}{{.column_two_label}}{{end}}</th>
            {{if .column_three_label}}<th scope="col">{{.column_three_label}}</th>{{end}}
            {{if .column_four_label}}<th scope="col">{{.column_four_label}}</th>{{end}}
          </tr>
        </thead>
        <tbody>
          {{range .rows}}
          <tr>
            <th scope="row">{{if .name}}{{.name}}{{end}}{{if .note}}<span class="cmp-table__note">{{.note}}</span>{{end}}</th>
            <td data-label="{{$.column_two_label}}">{{if .cell_two}}{{.cell_two}}{{end}}</td>
            {{if $.column_three_label}}<td data-label="{{$.column_three_label}}">{{if .cell_three}}{{.cell_three}}{{end}}</td>{{end}}
            {{if $.column_four_label}}<td data-label="{{$.column_four_label}}">{{if .cell_four}}{{.cell_four}}{{end}}</td>{{end}}
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
    {{if .source_note}}<p class="cmp-table__source">{{.source_note}}</p>{{end}}
  </div>
</section>
$tmpl$,
  true,
  'content',
  'structured_card',
  'high',
  '["brochure","saas","landing-page","portfolio","consultancy","professional-services","b2b"]'::jsonb,
  '["content","index","landing","blog-post"]'::jsonb,
  '["comparison","table","buying-guide","options","structured","generic"]'::jsonb,
  $schema$
{
  "fields": {
    "eyebrow": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "Short uppercase label above the title, under 5 words, e.g. 'At a glance'. Omit if the title already says it."
    },
    "section_title": {
      "type": "text", "source": "llm", "required": true,
      "llm_guidance": "The heading for this comparison, under 10 words. Name what is being compared and on what basis."
    },
    "intro": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "One or two sentences saying how to read the table, or who each option suits. Plain string, no HTML. Omit rather than padding."
    },
    "option_column_label": {
      "type": "text", "source": "llm", "required": true,
      "llm_guidance": "Heading for the FIRST column, which names the things being compared, e.g. 'Tool type', 'Option', 'Plan'. Two or three words."
    },
    "column_two_label": {
      "type": "text", "source": "llm", "required": true,
      "llm_guidance": "Heading for the second column: the first criterion every option is judged on, e.g. 'Best for'. Two to four words. Choose criteria a reader actually decides on."
    },
    "column_three_label": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "Heading for the optional third column, e.g. 'Watch out for'. Omit the field entirely if two columns say everything — an empty column is worse than no column."
    },
    "column_four_label": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "Heading for the optional fourth column. Four columns is the maximum; beyond that a phone cannot read it. Omit unless it earns its place."
    },
    "rows": {
      "type": "array", "source": "llm", "required": true, "min_items": 2,
      "items": {
        "type": "object",
        "required": ["name", "cell_two"],
        "properties": {
          "name": {"type": "string", "description": "the option being compared, as a short name"},
          "note": {"type": "string", "description": "optional half-line qualifier under the name"},
          "cell_two": {"type": "string", "description": "this option's value for column_two_label. A short phrase, not a sentence."},
          "cell_three": {"type": "string", "description": "value for column_three_label. Required if that column exists; leave empty string if genuinely not applicable."},
          "cell_four": {"type": "string", "description": "value for column_four_label. Required if that column exists; leave empty string if genuinely not applicable."}
        }
      },
      "llm_guidance": "Two to six options. Fill a cell for EVERY column you declared — a cell you omit renders empty under its own label, which is honest, but a table half-full reads as unfinished. Keep cells to a short phrase so the table stays scannable. ⚠ STATE NO PRICES, STAR RATINGS, SCORES, RANKS, WEIGHTS, DIMENSIONS, PERCENTAGES OR ANY OTHER FIGURE unless that exact figure was given to you in THIS prompt (Verified Facts, Research Findings, Admin Content Brief or Existing Content). This component deliberately has no price, rating or score field, and a figure typed into a text cell publishes a false claim on a live site just as surely as one in a numeric field. Compare on things you can state without inventing: what each option suits, what it struggles with, what a buyer should check. Never declare a winner or a 'best value' — say who each option is for and let the reader choose."
    },
    "source_note": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "Optional line under the table saying where the comparison comes from and what it does not cover. If the comparison is general guidance rather than tested results, SAY SO here in plain words — that is more useful to a reader than an implied authority, and it is the honest description of an untested comparison. Omit only if the intro already makes the basis clear."
    }
  }
}
$schema$::jsonb,
  'Several options compared across the same two to four writer-named criteria, rendered as a real <table> that stacks into labelled blocks on a phone. For buying guides, option round-ups and plan comparisons. Cells are named rather than positional, so a missing value renders in place instead of shifting the table. Has no price, rating, score or rank field.',
  'manual',
  now(), now()
);

DO $$
DECLARE tpl text; sch jsonb; exp text[];
BEGIN
  SELECT html_template, input_schema, component_expresses(html_template, input_schema)
    INTO tpl, sch, exp
    FROM content_components WHERE function = 'comparison-table' AND is_active;

  IF tpl IS NULL THEN
    RAISE EXCEPTION '606 VERIFY: the row was not inserted';
  END IF;
  IF NOT ('table' = ANY(exp)) THEN
    RAISE EXCEPTION '606 VERIFY: component_expresses = % — a comparison table that cannot express a table is the defect this closes', exp;
  END IF;
  IF NOT ('items' = ANY(exp)) THEN
    RAISE EXCEPTION '606 VERIFY: component_expresses = % — expected items', exp;
  END IF;
  -- Real table semantics, not a div grid pretending: header cells must be scoped.
  IF tpl !~* '<th scope="col"' OR tpl !~* '<th scope="row"' THEN
    RAISE EXCEPTION '606 VERIFY: the table lacks scoped header cells — a screen reader cannot associate cells with headers';
  END IF;
  -- The stacking pattern depends on every body cell carrying its label.
  IF (length(tpl) - length(replace(tpl, 'data-label=', ''))) / length('data-label=') <> 3 THEN
    RAISE EXCEPTION '606 VERIFY: expected 3 data-label cells (one per optional/required data column) — the mobile stack shows a cell with no label without them';
  END IF;
  IF tpl ~* '<h1[\s>]' THEN
    RAISE EXCEPTION '606 VERIFY: template contains an <h1>';
  END IF;
  IF (length(tpl) - length(replace(tpl, '<h2', ''))) / length('<h2') <> 1 THEN
    RAISE EXCEPTION '606 VERIFY: expected exactly one <h2> (the section title)';
  END IF;
  -- A heading inside a table cell breaks the table's semantics.
  IF tpl ~* '<h3[\s>]' THEN
    RAISE EXCEPTION '606 VERIFY: template contains an <h3> — row names are table header cells, not headings';
  END IF;
  IF tpl ~* '<svg' THEN
    RAISE EXCEPTION '606 VERIFY: template contains an <svg>';
  END IF;
  -- The claims control that IS structural: no figure-shaped field may exist.
  IF sch::text ~* '"(price|rating|score|rank|stars|percent|count|number|weight)' THEN
    RAISE EXCEPTION '606 VERIFY: the schema declares a figure-shaped field — the one structural control this component has is that none exists';
  END IF;
  -- ⚠ EVERY INTERPOLATION MUST BE {{if}}-GUARDED — asserted per field, and the guard
  -- may be INLINE ({{if .x}}{{.x}}{{end}}) or wrap the whole element
  -- ({{if .x}}<p>{{.x}}</p>{{end}}); both are correct and the second is preferred where
  -- an empty element would otherwise render. An earlier version of this assertion
  -- demanded the inline spelling and FALSELY failed on the better one — the check was
  -- wrong, not the template, and it is written this way because of that.
  -- WHY IT MATTERS: `missingkey=zero` does NOT protect a map[string]interface{} — the
  -- zero value of interface{} is nil, and text/template prints nil as the literal
  -- "<no value>". An unguarded {{.field}} therefore publishes "<no value>" onto a live
  -- page whenever the writer omits that key: schema-valid row, no error, wrong page.
  -- MEASURED 2026-08-24: 0 live page_components carry that string (control: 1,907 carry
  -- "<section"), so this is a hazard not to introduce, not damage to repair. The render
  -- harness in the lane RUNBOOK is what actually proves it; this is the cheap sentry.
  IF position($i${{.section_title}}$i$ in tpl) > 0 AND position($g${{if .section_title}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '606 VERIFY: {{.section_title}} is interpolated with NO {{if .section_title}} guard anywhere — an absent key renders the literal <no value> onto a live page';
  END IF;
  IF position($i${{.option_column_label}}$i$ in tpl) > 0 AND position($g${{if .option_column_label}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '606 VERIFY: {{.option_column_label}} is interpolated with NO {{if .option_column_label}} guard anywhere — an absent key renders the literal <no value> onto a live page';
  END IF;
  IF position($i${{.column_two_label}}$i$ in tpl) > 0 AND position($g${{if .column_two_label}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '606 VERIFY: {{.column_two_label}} is interpolated with NO {{if .column_two_label}} guard anywhere — an absent key renders the literal <no value> onto a live page';
  END IF;
  IF position($i${{.name}}$i$ in tpl) > 0 AND position($g${{if .name}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '606 VERIFY: {{.name}} is interpolated with NO {{if .name}} guard anywhere — an absent key renders the literal <no value> onto a live page';
  END IF;
  IF position($i${{.cell_two}}$i$ in tpl) > 0 AND position($g${{if .cell_two}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '606 VERIFY: {{.cell_two}} is interpolated with NO {{if .cell_two}} guard anywhere — an absent key renders the literal <no value> onto a live page';
  END IF;
  IF position($i${{.cell_three}}$i$ in tpl) > 0 AND position($g${{if .cell_three}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '606 VERIFY: {{.cell_three}} is interpolated with NO {{if .cell_three}} guard anywhere — an absent key renders the literal <no value> onto a live page';
  END IF;
  IF position($i${{.cell_four}}$i$ in tpl) > 0 AND position($g${{if .cell_four}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '606 VERIFY: {{.cell_four}} is interpolated with NO {{if .cell_four}} guard anywhere — an absent key renders the literal <no value> onto a live page';
  END IF;
  IF position('{{.InstanceID}}' in tpl) = 0 THEN
    RAISE EXCEPTION '606 VERIFY: template does not carry {{.InstanceID}} (RFC_032)';
  END IF;
  IF (SELECT section_type FROM content_components WHERE function = 'comparison-table' AND is_active) IS NULL THEN
    RAISE EXCEPTION '606 VERIFY: section_type is NULL';
  END IF;
  RAISE NOTICE '606 OK: comparison-table component created, expresses %', exp;
END $$;

COMMIT;
