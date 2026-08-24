-- 604_checklist_component.sql
-- A generic CHECKLIST section: a list of things checked, done, or required, each
-- with a short heading and an optional line of detail. Built for bugs_open/381,
-- deliberately generic — no vertical, no site, no product assumed.
--
-- WHY THIS EXISTS. bugs_open/381 found that the planner composes pages from
-- components that cannot express what the page promises, and fixed the half it
-- could: the planner now SEES what each component can produce (migration 591).
-- That fix was explicitly recorded as NECESSARY BUT NOT SUFFICIENT, because
-- `[MEASURED 2026-08-24]` the library's 44 structural components are all
-- special-purpose — directories, trackers, calculators, quizzes, spec sheets, one
-- pricing table, two carousels, and site-footer, which is chrome. **A planner told
-- what everything can express still had nothing generic to choose.** This is one of
-- the three components that closes that gap.
--
-- THE MOTIVATING CASE IS THE OWNER'S OWN COMPLAINT. garden-tools.uk's
-- `how-we-assess` page carried 1,486 words in 14 paragraphs with no subheads, no
-- list and no emphasis, and its centrepiece was a 300-word paragraph beginning
-- "What we check before a tool earns a recommendation…". That paragraph IS a
-- checklist, written as prose because nothing in the library was a checklist.
--
-- REUSE CHECKED FIRST, and it changed the plan. `mechanism-flow` (247) already
-- draws an ORDERED process with decision branches, so a "steps" component was NOT
-- built — that would have been a near-duplicate. A checklist is a different shape:
-- unordered, each item independently true or false, no flow between them. The
-- other near-neighbours are `features`/`differentiators` (marketing claims about
-- us, rendered as cards) and `faq` (question/answer pairs). None of them is a list
-- of criteria.
--
-- DESIGN RULES, each from a recorded failure rather than taste:
--
-- 1. NO NUMERIC OR SCORE FIELD, STRUCTURALLY (247's rule 1). A checklist invites
--    "9 out of 10 tools pass this" or a star rating, and a number-shaped slot is an
--    invitation for a writer to fill it — the 043 spec-poisoning precedent, and the
--    reason bugs_open/380 exists. There is no such field here. The absence of the
--    slot is the control; the writer prompt's rule 14 is the backstop, not the
--    other way round.
-- 2. THE TICK IS DRAWN IN CSS AND aria-hidden, NOT AN EMOJI OR AN <svg> WITH TEXT.
--    Text inside <svg> is INVISIBLE to the claims gate (recorded landmine), so no
--    label ever goes in one. A CSS tick is decorative, carries no words, and a
--    screen reader reads the item's real heading instead of "check mark".
-- 3. HEADINGS START AT <h3>. The section supplies its own <h2>; an <h2> here would
--    produce two page-level headings in one section and break the outline.
-- 4. NO LIGHT LITERAL FALLBACKS. Every colour is a theme variable with a
--    neutral/translucent fallback — a light literal is how a white card lands on a
--    dark page (the evidence-chart precedent). Structure does NOT rely on border
--    contrast alone: the tick, the heading weight and the spacing carry it, because
--    `--color-border` measured 1.66 against background on a real site palette and
--    FAILS the 3.0 non-text threshold (247's rule 3). This component ships
--    fleet-wide, so a single contrast measurement is not available the way it was
--    for a one-site component — the mitigation is to not make comprehension depend
--    on a hairline in the first place.
-- 5. `{{.InstanceID}}` ON THE SECTION ELEMENT (RFC_032 / bugs_open/283). New
--    components are born per-instance-scoped rather than converted later.
--
-- ⚠ content_shape IS SET AND IS CURRENTLY READ BY NOTHING. `[MEASURED 2026-08-24]`
-- that column has zero Go readers, is omitted from the birth INSERT in
-- store_generated_component_action.go, and is wrong on 12 existing rows. It is set
-- here so this row is not part of the drift if it is ever revived; the LIVE answer
-- to "what can this express" is `component_expresses(html_template, input_schema)`
-- (migration 591, register PLAN-053). See LANDMINES.md.
--
-- ROLLBACK: 604_checklist_component_ROLLBACK.sql

BEGIN;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM content_components WHERE function = 'checklist' AND is_active) THEN
    RAISE EXCEPTION '604: an active component with function=checklist already exists — refusing to double-apply or shadow it';
  END IF;
END $$;

INSERT INTO content_components (
  name, function, display_name, component_level, render_mode, section_type,
  html_template, is_active, category, content_shape, visual_density,
  suitable_site_types, suitable_page_types, semantic_tags,
  input_schema, description, created_from, created_at, updated_at
)
VALUES (
  'checklist',
  'checklist',
  'Checklist',
  'section',
  'template',
  'checklist',
  $tmpl$
<style>
  .checklist-section {
    padding: var(--spacing-section, 4.5rem 2rem);
    background: var(--color-background);
    color: var(--color-text);
  }
  .checklist__inner { max-width: var(--container-max-width, 64rem); margin: 0 auto; }
  .checklist__header { max-width: 46rem; margin: 0 0 2.5rem; }
  .checklist__eyebrow {
    display: block; font-size: 0.8125rem; font-weight: 600;
    letter-spacing: 0.1em; text-transform: uppercase;
    color: var(--color-primary-ink, var(--color-accent, currentColor));
    margin: 0 0 0.5rem;
  }
  .checklist__title {
    font-size: clamp(1.5rem, 2.6vw, 2.1rem); font-weight: 700;
    line-height: 1.25; margin: 0 0 0.75rem; color: var(--color-heading, inherit);
  }
  .checklist__intro {
    margin: 0; line-height: 1.7; color: var(--color-text-muted, inherit);
  }
  .checklist__list { list-style: none; margin: 0; padding: 0; display: grid; gap: 1.25rem; }
  .checklist__item {
    display: grid; grid-template-columns: 1.75rem 1fr; gap: 0.9rem;
    align-items: start; padding: 0 0 1.25rem;
    border-bottom: 1px solid var(--color-border, rgba(128,128,128,0.25));
  }
  .checklist__item:last-child { border-bottom: 0; padding-bottom: 0; }
  /* Decorative tick, drawn in CSS. No text, no svg, no emoji — see header rule 2. */
  .checklist__mark {
    width: 1.5rem; height: 1.5rem; border-radius: 50%; margin-top: 0.15rem;
    background: var(--color-primary-soft, rgba(128,128,128,0.16));
    position: relative; flex: none;
  }
  .checklist__mark::after {
    content: ""; position: absolute; left: 0.5rem; top: 0.34rem;
    width: 0.36rem; height: 0.66rem;
    border: solid var(--color-primary-ink, var(--color-accent, currentColor));
    border-width: 0 2px 2px 0; transform: rotate(45deg);
  }
  .checklist__item-title {
    font-size: 1.0625rem; font-weight: 650; line-height: 1.35;
    margin: 0; color: var(--color-heading, inherit);
  }
  .checklist__item-detail {
    margin: 0.35rem 0 0; line-height: 1.65; color: var(--color-text-muted, inherit);
  }
  .checklist__footnote {
    margin: 2rem 0 0; font-size: 0.9375rem; line-height: 1.6;
    color: var(--color-text-muted, inherit);
  }
  @media (max-width: 32rem) {
    .checklist__item { grid-template-columns: 1.4rem 1fr; gap: 0.7rem; }
  }
</style>
<section id="{{.InstanceID}}" class="checklist-section" data-component="checklist">
  <div class="checklist__inner">
    <header class="checklist__header">
      {{if .eyebrow}}<span class="checklist__eyebrow">{{.eyebrow}}</span>{{end}}
      <h2 class="checklist__title">{{if .section_title}}{{.section_title}}{{end}}</h2>
      {{if .intro}}<p class="checklist__intro">{{.intro}}</p>{{end}}
    </header>
    <ul class="checklist__list">
      {{range .items}}
      <li class="checklist__item">
        <span class="checklist__mark" aria-hidden="true"></span>
        <div class="checklist__body">
          <h3 class="checklist__item-title">{{if .title}}{{.title}}{{end}}</h3>
          {{if .detail}}<p class="checklist__item-detail">{{.detail}}</p>{{end}}
        </div>
      </li>
      {{end}}
    </ul>
    {{if .footnote}}<p class="checklist__footnote">{{.footnote}}</p>{{end}}
  </div>
</section>
$tmpl$,
  true,
  'content',
  'structured_list',
  'medium',
  '["brochure","saas","landing-page","portfolio","consultancy","professional-services","b2b"]'::jsonb,
  '["content","index","landing","blog-post"]'::jsonb,
  '["checklist","criteria","list","structured","generic"]'::jsonb,
  $schema$
{
  "fields": {
    "eyebrow": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "Short uppercase label above the title, under 5 words, e.g. 'How we assess' or 'Before you buy'. Omit if the title already says it."
    },
    "section_title": {
      "type": "text", "source": "llm", "required": true,
      "llm_guidance": "The heading for this checklist, under 10 words. Say what the list IS a list of — 'What we check before recommending a tool' beats 'Our standards'."
    },
    "intro": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "One or two sentences introducing the list. Plain string, no HTML. Omit rather than padding — the list is the content."
    },
    "items": {
      "type": "array", "source": "llm", "required": true, "min_items": 3,
      "items": {
        "type": "object",
        "required": ["title"],
        "properties": {
          "title": {"type": "string", "description": "the check itself, as a short phrase of 3-9 words"},
          "detail": {"type": "string", "description": "one sentence saying what it means in practice, or how it is judged. Optional."}
        }
      },
      "llm_guidance": "Between 3 and 8 checks. Each item: a short title naming the check, and optionally ONE sentence of detail. Write real, specific criteria — a checklist of vague virtues ('quality', 'value') is the wall of text this component exists to replace, in bullet form. State NO counts, scores, percentages, ratings or prices anywhere in these items: this component has no numeric field on purpose, and an invented figure publishes a false claim on a live site. If a check genuinely depends on a figure, name the thing measured without the number."
    },
    "footnote": {
      "type": "text", "source": "llm", "required": false, "on_missing": "skip_field",
      "llm_guidance": "Optional closing line — a limit of the list, or what a reader should do next. Say plainly where the list does not apply if that is true. Omit unless it adds something."
    }
  }
}
$schema$::jsonb,
  'A list of things checked, required, or done — each a short heading with an optional line of detail, rendered as a real <ul>. For criteria, standards, what-we-check, what-you-need and pre-purchase checks. Unordered by design: for an ordered process with stages use mechanism-flow instead. Has no numeric or score field.',
  'manual',
  now(), now()
);

DO $$
DECLARE tpl text; sch jsonb; exp text[];
BEGIN
  SELECT html_template, input_schema, component_expresses(html_template, input_schema)
    INTO tpl, sch, exp
    FROM content_components WHERE function = 'checklist' AND is_active;

  IF tpl IS NULL THEN
    RAISE EXCEPTION '604 VERIFY: the row was not inserted';
  END IF;
  -- The whole point of the component: it must actually be able to render a list.
  IF NOT ('list' = ANY(exp)) THEN
    RAISE EXCEPTION '604 VERIFY: component_expresses = % — a checklist that cannot express a list is the defect this closes', exp;
  END IF;
  IF NOT ('items' = ANY(exp)) THEN
    RAISE EXCEPTION '604 VERIFY: component_expresses = % — expected items (a range over an llm array field)', exp;
  END IF;
  -- Rule 3: the section owns the page-level heading.
  IF tpl ~* '<h1[\s>]' THEN
    RAISE EXCEPTION '604 VERIFY: template contains an <h1>';
  END IF;
  IF (length(tpl) - length(replace(tpl, '<h2', ''))) / length('<h2') <> 1 THEN
    RAISE EXCEPTION '604 VERIFY: expected exactly one <h2> (the section title)';
  END IF;
  -- Rule 2: no svg anywhere, so no label can hide from the claims gate.
  IF tpl ~* '<svg' THEN
    RAISE EXCEPTION '604 VERIFY: template contains an <svg> — text inside one is invisible to the claims gate';
  END IF;
  -- Rule 1: no numeric slot.
  IF sch::text ~* '"(score|rating|price|count|percent|stat|number)' THEN
    RAISE EXCEPTION '604 VERIFY: the schema declares a numeric-shaped field — see design rule 1';
  END IF;
  -- Rule 5: per-instance scope from birth.
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
    RAISE EXCEPTION '604 VERIFY: {{.section_title}} is interpolated with NO {{if .section_title}} guard anywhere — an absent key renders the literal <no value> onto a live page';
  END IF;
  IF position($i${{.title}}$i$ in tpl) > 0 AND position($g${{if .title}$g$ in tpl) = 0 THEN
    RAISE EXCEPTION '604 VERIFY: {{.title}} is interpolated with NO {{if .title}} guard anywhere — an absent key renders the literal <no value> onto a live page';
  END IF;
  IF position('{{.InstanceID}}' in tpl) = 0 THEN
    RAISE EXCEPTION '604 VERIFY: template does not carry {{.InstanceID}} (RFC_032)';
  END IF;
  -- The birth gate (581) must have been satisfied, not bypassed.
  IF (SELECT section_type FROM content_components WHERE function = 'checklist' AND is_active) IS NULL THEN
    RAISE EXCEPTION '604 VERIFY: section_type is NULL — the row would be invisible to the selector';
  END IF;
  RAISE NOTICE '604 OK: checklist component created, expresses %', exp;
END $$;

COMMIT;
