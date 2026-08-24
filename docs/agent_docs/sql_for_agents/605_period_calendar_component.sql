-- 605_period_calendar_component.sql
-- A generic PERIOD CALENDAR: an ordered list of named periods — months, seasons,
-- quarters, weeks, stages of a year — each with a focus line and optional detail.
-- Built for bugs_open/381, deliberately generic.
--
-- WHY THIS EXISTS, and it is the literal motivating case of the bug. garden-tools.uk's
-- `seasonal-planner` page carried the heading "What your shed needs, month by month"
-- and delivered four <h3> prose blocks (Spring/Summer/Autumn/Winter). `[MEASURED
-- 2026-08-24]` THREE month names appeared on the whole page, all incidental. The
-- writer had nowhere to put twelve months, so it wrote four seasons in prose — the
-- page degrading to fit its container. Migration 591 made the planner able to SEE
-- that its chosen components were prose-only; this is the component it can now
-- choose instead.
--
-- WHY "PERIOD" AND NOT "MONTH". A hard twelve-month component would be unreachable
-- for a quarterly roadmap, a four-season guide, a six-week programme or a
-- financial-year cycle — and the estate would grow a near-duplicate for each. The
-- writer names the periods, so twelve months is one instance of it and "Q1..Q4" is
-- another. The ORDER is the writer's, in the order emitted.
--
-- REUSE CHECKED FIRST. `mechanism-flow` (247) draws an ordered process, which is the
-- closest existing shape — but its steps are CAUSAL (each follows from the last, with
-- decision branches), and a calendar's periods are not: March does not cause April,
-- and a reader jumps straight to the month they are in. Rendering a calendar as a
-- mechanism would imply a dependency that is not there. `faq` and `features` are
-- unordered card sets. Nothing periodised exists.
--
-- ⚠ NOT A TIMELINE, AND THIS BOUNDARY IS AGREED WITH THE `editorial_design_uplift`
-- LANE. Their Phase E timeline is FACT-FED: dated real-world events, each carrying
-- its own citation from the evidence register, failing closed on a site with no
-- evidence base. This component is the opposite by design — authored guidance about
-- a RECURRING cycle ("what to do in March", not "what happened in March 2024"), with
-- no dated events and no citations. A site can legitimately have both. If you find
-- yourself putting a year in a period label, you want their component, not this one.
--
-- DESIGN RULES, each from a recorded failure rather than taste:
--
-- 1. NO NUMERIC FIELD, STRUCTURALLY (247's rule 1). A calendar invites "prune in the
--    first 2 weeks" or a temperature or a price, and a number-shaped slot is an
--    invitation to fill it (the 043 spec-poisoning precedent; bugs_open/380). There
--    is none. `label` is a period NAME, not a date or a quantity.
-- 2. AN ORDERED LIST, IN THE MARKUP. <ol> rather than a div grid: the order is
--    meaning here, and a screen reader should announce it. This is also what makes
--    `component_expresses` report `list`.
-- 3. COMPACT ROWS, NOT CARDS. Twelve cards is the "wall of cards" the owner
--    complained about on the same review (bugs_open/381 §7: the wall comes from the
--    NUMBER of card sections, not cards per section). A calendar is scanned to find
--    one period, so it renders as a compact table-like list that stays legible at
--    twelve rows and collapses to one column cleanly.
-- 4. HEADINGS START AT <h3>; the section owns its <h2>.
-- 5. NO <svg>, NO EMOJI MARKERS. Text inside <svg> is invisible to the claims gate
--    (recorded landmine). The period label is real HTML text.
-- 6. NO LIGHT LITERAL FALLBACKS, and comprehension does not depend on a hairline —
--    `--color-border` measured 1.66 on a real palette and FAILS the 3.0 non-text
--    threshold (247's rule 3). The period label's weight and placement carry the
--    structure; the rule is decoration.
-- 7. `{{.InstanceID}}` from birth (RFC_032 / bugs_open/283).
--
-- ⚠ content_shape is set and is currently READ BY NOTHING — see 604's header and
-- LANDMINES.md. The live answer is component_expresses() (591, PLAN-053).
--
-- ROLLBACK: 605_period_calendar_component_ROLLBACK.sql

BEGIN;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM content_components WHERE function = 'period-calendar' AND is_active) THEN
    RAISE EXCEPTION '605: an active component with function=period-calendar already exists — refusing to double-apply or shadow it';
  END IF;
END $$;

INSERT INTO content_components (
  name, function, display_name, component_level, render_mode, section_type,
  html_template, is_active, category, content_shape, visual_density,
  suitable_site_types, suitable_page_types, semantic_tags,
  input_schema, description, created_from, created_at, updated_at
)
VALUES (
  'period-calendar',
  'period-calendar',
  'Period calendar',
  'section',
  'template',
  'period-calendar',
  $tmpl$
<style>
  .period-cal {
    padding: var(--spacing-section, 4.5rem 2rem);
    background: var(--color-background);
    color: var(--color-text);
  }
  .period-cal__inner { max-width: var(--container-max-width, 64rem); margin: 0 auto; }
  .period-cal__header { max-width: 46rem; margin: 0 0 2.5rem; }
  .period-cal__eyebrow {
    display: block; font-size: 0.8125rem; font-weight: 600;
    letter-spacing: 0.1em; text-transform: uppercase;
    color: var(--color-primary-ink, var(--color-accent, currentColor));
    margin: 0 0 0.5rem;
  }
  .period-cal__title {
    font-size: clamp(1.5rem, 2.6vw, 2.1rem); font-weight: 700;
    line-height: 1.25; margin: 0 0 0.75rem; color: var(--color-heading, inherit);
  }
  .period-cal__intro { margin: 0; line-height: 1.7; color: var(--color-text-muted, inherit); }
  .period-cal__list { list-style: none; margin: 0; padding: 0; counter-reset: none; }
  .period-cal__row {
    display: grid; grid-template-columns: minmax(6rem, 9rem) 1fr; gap: 1.25rem;
    padding: 1.1rem 0;
    border-top: 1px solid var(--color-border, rgba(128,128,128,0.25));
  }
  .period-cal__row:first-child { border-top: 0; padding-top: 0; }
  .period-cal__label {
    font-size: 1rem; font-weight: 700; line-height: 1.3; margin: 0;
    color: var(--color-primary-ink, var(--color-heading, inherit));
    letter-spacing: 0.01em;
  }
  .period-cal__focus {
    margin: 0; font-weight: 600; line-height: 1.45; color: var(--color-heading, inherit);
  }
  .period-cal__detail {
    margin: 0.3rem 0 0; line-height: 1.65; color: var(--color-text-muted, inherit);
  }
  .period-cal__footnote {
    margin: 2rem 0 0; font-size: 0.9375rem; line-height: 1.6;
    color: var(--color-text-muted, inherit);
  }
  @media (max-width: 34rem) {
    .period-cal__row { grid-template-columns: 1fr; gap: 0.3rem; }
    .period-cal__label { color: var(--color-primary-ink, var(--color-accent, currentColor)); }
  }
</style>
<section id="{{.InstanceID}}" class="period-cal" data-component="period-calendar">
  <div class="period-cal__inner">
    <header class="period-cal__header">
      {{if .eyebrow}}<span class="period-cal__eyebrow">{{.eyebrow}}</span>{{end}}
      <h2 class="period-cal__title">{{if .section_title}}{{.section_title}}{{end}}</h2>
      {{if .intro}}<p class="period-cal__intro">{{.intro}}</p>{{end}}
    </header>
    <ol class="period-cal__list">
      {{range .periods}}
      <li class="period-cal__row">
        <h3 class="period-cal__label">{{if .label}}{{.label}}{{end}}</h3>
        <div class="period-cal__body">
          {{if .focus}}<p class="period-cal__focus">{{.focus}}</p>{{end}}
          {{if .detail}}<p class="period-cal__detail">{{.detail}}</p>{{end}}
        </div>
      </li>
      {{end}}
    </ol>
    {{if .footnote}}<p class="period-cal__footnote">{{.footnote}}</p>{{end}}
  </div>
</section>
$tmpl$,
  true,
  'content',
  'sequence',
  'medium',
  '["brochure","saas","landing-page","portfolio","consultancy","professional-services","b2b"]'::jsonb,
  '["content","index","landing","blog-post"]'::jsonb,
  '["calendar","seasonal","month-by-month","periodised","schedule","list","structured","generic"]'::jsonb,
  $schema$
{
  "fields": {
    "eyebrow": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "Short uppercase label above the title, under 5 words, e.g. 'Through the year'. Omit if the title already says it."
    },
    "section_title": {
      "type": "text", "source": "llm", "required": true,
      "llm_guidance": "The heading for this calendar, under 10 words. If the page promises a month-by-month guide, say so here and then DELIVER twelve periods below."
    },
    "intro": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "One or two sentences on how to use the calendar. Plain string, no HTML. Omit rather than padding."
    },
    "periods": {
      "type": "array", "source": "llm", "required": true, "min_items": 3,
      "items": {
        "type": "object",
        "required": ["label", "focus"],
        "properties": {
          "label": {"type": "string", "description": "the period's NAME, e.g. 'January' or 'Q1' or 'Early spring'. A name, never a date and never a quantity."},
          "focus": {"type": "string", "description": "the one thing this period is for, as a short phrase of 3-10 words"},
          "detail": {"type": "string", "description": "one or two sentences of practical detail. Optional."}
        }
      },
      "llm_guidance": "The periods, IN ORDER, as many as the subject genuinely has. If the page or its heading promises 'month by month', emit ALL TWELVE months, named, in calendar order — a four-season summary does not keep that promise and is the defect this component exists to fix. For a quarterly or staged cycle, use that cycle's own names. Each period needs a distinct focus: if two periods would say the same thing, the subject is not periodised and a checklist or prose section is the honest choice. State NO dates, years, counts, temperatures, prices or measurements — this component has no numeric field on purpose, and an invented figure publishes a false claim on a live site. Name what happens, not how much."
    },
    "footnote": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "Optional closing line — most usefully, what changes the timings (climate, region, model). Say plainly that the calendar is a guide where that is true. Omit unless it adds something."
    }
  }
}
$schema$::jsonb,
  'An ordered run of named periods — months, quarters, seasons, stages — each with a focus line and optional detail, rendered as a real <ol>. THE component for a month-by-month or season-by-season page. Not a timeline: it describes a recurring cycle, not dated real-world events with citations. Has no numeric or date field.',
  'manual',
  now(), now()
);

DO $$
DECLARE tpl text; sch jsonb; exp text[];
BEGIN
  SELECT html_template, input_schema, component_expresses(html_template, input_schema)
    INTO tpl, sch, exp
    FROM content_components WHERE function = 'period-calendar' AND is_active;

  IF tpl IS NULL THEN
    RAISE EXCEPTION '605 VERIFY: the row was not inserted';
  END IF;
  IF NOT ('list' = ANY(exp)) THEN
    RAISE EXCEPTION '605 VERIFY: component_expresses = % — a calendar that cannot express a list is the defect this closes', exp;
  END IF;
  IF NOT ('items' = ANY(exp)) THEN
    RAISE EXCEPTION '605 VERIFY: component_expresses = % — expected items', exp;
  END IF;
  -- Rule 2: the order must be in the markup, not just in the data.
  IF tpl !~* '<ol[\s>]' THEN
    RAISE EXCEPTION '605 VERIFY: template has no <ol> — the order is meaning here and must be announced';
  END IF;
  IF tpl ~* '<h1[\s>]' THEN
    RAISE EXCEPTION '605 VERIFY: template contains an <h1>';
  END IF;
  IF (length(tpl) - length(replace(tpl, '<h2', ''))) / length('<h2') <> 1 THEN
    RAISE EXCEPTION '605 VERIFY: expected exactly one <h2> (the section title)';
  END IF;
  IF tpl ~* '<svg' THEN
    RAISE EXCEPTION '605 VERIFY: template contains an <svg>';
  END IF;
  IF sch::text ~* '"(score|rating|price|count|percent|stat|number|date|year)' THEN
    RAISE EXCEPTION '605 VERIFY: the schema declares a numeric- or date-shaped field — see design rule 1';
  END IF;
  -- ⚠ CORRECTED 2026-08-24, hours after this file was written: THE PLATFORM ALREADY
  -- STRIPS "<no value>". `RenderTemplate` does
  -- `strings.ReplaceAll(result, "<no value>", "")` (component_library.go:1258) on the
  -- live path (`RenderComponentAction` -> v3_site_actions.go:2459), and immediately
  -- above it `missingBareFields` REPORTS the fields that rendered empty, at Error
  -- level, by name (bugs_open/018), while `missingRequiredLLMFields` gates an absent
  -- required field (bugs_open/342). So an unguarded interpolation does NOT reach a
  -- visitor, and the guards below are HYGIENE, not a defect being prevented: they
  -- render a deliberate empty element instead of relying on a downstream string
  -- replace, and they keep the intent visible in the template. The earlier framing in
  -- this header — that an unguarded field publishes "<no value>" onto a live page —
  -- was WRONG. The measurement behind it (0 live occurrences, control 1,907) was
  -- right; the explanation was not: it is 0 because the platform strips, not because
  -- writers fill every key. Kept as a guard because it is still better, corrected
  -- because the reason was false. Full incident in WRONG_CALLS.md.
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
    RAISE EXCEPTION '605 VERIFY: {{.section_title}} is interpolated with NO {{if .section_title}} guard anywhere — an absent key renders <no value>, which the platform strips (component_library.go:1258) — this is hygiene, not a live-page defect';
  END IF;
  IF position($i${{.label}}$i$ in tpl) > 0 AND position($g${{if .label}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '605 VERIFY: {{.label}} is interpolated with NO {{if .label}} guard anywhere — an absent key renders <no value>, which the platform strips (component_library.go:1258) — this is hygiene, not a live-page defect';
  END IF;
  IF position($i${{.focus}}$i$ in tpl) > 0 AND position($g${{if .focus}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '605 VERIFY: {{.focus}} is interpolated with NO {{if .focus}} guard anywhere — an absent key renders <no value>, which the platform strips (component_library.go:1258) — this is hygiene, not a live-page defect';
  END IF;
  IF position('{{.InstanceID}}' in tpl) = 0 THEN
    RAISE EXCEPTION '605 VERIFY: template does not carry {{.InstanceID}} (RFC_032)';
  END IF;
  IF (SELECT section_type FROM content_components WHERE function = 'period-calendar' AND is_active) IS NULL THEN
    RAISE EXCEPTION '605 VERIFY: section_type is NULL';
  END IF;
  RAISE NOTICE '605 OK: period-calendar component created, expresses %', exp;
END $$;

COMMIT;
