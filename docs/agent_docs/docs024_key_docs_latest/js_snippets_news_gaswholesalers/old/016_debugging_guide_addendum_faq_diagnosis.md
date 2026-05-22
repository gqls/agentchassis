# Addendum to 016 — Debugging Guide

Three new entries for **Section 9. Specific Failure Patterns**, plus one
note for **Section 0. Assumption Checklist**. Drawn from the
gaswholesalers.com FAQ empty-items diagnosis, 2026-05-20.

The first entry is the headline one: it documents the *diagnostic method*
for "a component renders empty shells" — the layered narrowing from
rendered HTML down to the data source, which generalises to any
component-not-populating bug.

---

### A component renders empty structural shells (FAQ accordion, news grid, any `{{range}}` template)

**Symptom:** A page renders the *structure* of a component but none of
its repeated content. For the FAQ case: four `<details class="faq-item">`
blocks with empty `<summary>` and `<p>` — the accordion shell is there,
every question/answer is blank. The heading ("Wholesale Fuel FAQ")
renders fine because it's static template text; only the
`{{range .questions}}` body is empty.

**Why this specific shape matters:** empty *shells* (not a missing
section, not a raw `{{range}}`) tells you the template engine ran and
iterated — it just iterated over items whose fields were empty, or over
the right count of empty objects. That's different from "the component
didn't render at all." The structure-present/content-absent split points
at the data binding, not the template or the section list.

**Diagnostic method — narrow from the outside in.** Each step eliminates
a layer. Do them in order; stop when one returns the smoking gun.

```sql
-- STEP 1: List the page's components in order. Look for orphans
-- (component_id IS NULL) and check which slot the content should be in.
SELECT
  pc.id, pc.position,
  cc.function, cc.name AS component_name,
  pc.build_status,
  pc.component_id,                          -- NULL = orphaned, can't resolve template
  pc.content_item_id,                       -- where structured data may live
  jsonb_typeof(pc.content_data) AS cd_type, -- NULL = no inline data
  LEFT(pc.content_data::text, 120) AS cd_preview
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
LEFT JOIN content_components cc ON pc.component_id = cc.id
WHERE p.site_id = '<site-id>' AND p.name = '<page-name>'
ORDER BY pc.position;
```

Two things to read off STEP 1:
- **Is the target component orphaned?** `component_id IS NULL` means the
  renderer can't map this page_component back to a `content_components`
  template at all. It renders whatever stale `rendered_html` is stored,
  not a fresh template binding. (For the FAQ page, positions 2 and 3 were
  both orphaned.)
- **Where is the data?** `content_data` empty AND `content_item_id` NULL
  means there is no structured data for the template to bind. That alone
  explains empty shells.

```sql
-- STEP 2: Confirm what the component template expects to bind.
-- The input_schema names the data keys; the template shows the {{range}}.
SELECT id, name, function,
       LEFT(html_template, 1200) AS template_preview,
       input_schema
FROM content_components
WHERE function = '<function>'    -- e.g. 'faq'
ORDER BY updated_at DESC LIMIT 3;
```

This tells you the exact variable the data must supply. For `faq` the
template is `{{range .questions}}<summary>{{.question}}</summary>...`,
and `input_schema.fields.questions` is a required LLM-sourced array of
`{question, answer}` with `min_items: 3`. So the data key is `questions`
— if the page_component has no `questions` data, the range is empty.

```sql
-- STEP 3: Check whether the structured data exists anywhere for this
-- component instance — inline content_data OR a linked content_item.
SELECT pc.id, pc.position, pc.content_item_id,
       pc.content_data,
       ci.content_data AS content_item_data
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
LEFT JOIN content_items ci ON pc.content_item_id = ci.id
WHERE p.site_id = '<site-id>' AND p.name = '<page-name>'
  AND pc.position = <target-position>;
```

If both are empty/NULL, the data was never written to this component.
The content exists *somewhere* (the page clearly has FAQ prose) but not
bound to the structured component.

```sql
-- STEP 4: Count the rendered shells to distinguish "iterated over empty
-- objects" from "iterated over nothing". A non-zero shell count with
-- empty fields means the binding ran with placeholder/empty objects.
SELECT pc.position,
       LENGTH(pc.rendered_html) AS html_len,
       (LENGTH(pc.rendered_html)
        - LENGTH(REPLACE(pc.rendered_html, 'faq-item__question', '')))
        / LENGTH('faq-item__question') AS shell_count
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
WHERE p.site_id = '<site-id>' AND p.name = '<page-name>'
  AND pc.position = <target-position>;
```

4 shells with empty fields (the FAQ case) means a `questions` array of 4
empty objects was bound at some past render, then frozen into
`rendered_html`. Zero shells would mean the range got an empty array.

```sql
-- STEP 5: Compare the page's ACTUAL sections against the site_plan.
-- This catches duplicate/competing sections and stale plans.
SELECT name, jsonb_pretty(sections) AS actual_sections
FROM pages
WHERE site_id = '<site-id>' AND name = '<page-name>';

SELECT jsonb_pretty(section) AS planned_sections
FROM site_specs ss,
     jsonb_array_elements(ss.data #> '{pages}') AS section
WHERE ss.site_id = '<site-id>' AND ss.aspect = 'site_plan'
  AND section->>'name' = '<page-name>';
```

For the FAQ page this exposed the real root cause (below): the page's
`pages.sections` is `["hero","generic-text-block","faq","call_to_action"]`
— it has BOTH a freeform `generic-text-block` and a structured `faq`
component. And the `site_plan` query returned **zero rows** — the faq
page isn't in the plan at all.

**Root cause (FAQ case):** the page carries two competing FAQ surfaces.
The content writer, given both a `generic-text-block` and a `faq`
component on one page, wrote all the Q&A content into the
generic-text-block as `<strong>`-question prose (fully populated) and
left the structured `faq` component with empty placeholder questions. The
empty accordion is a *symptom*; the cause is a page plan that put two
overlapping content slots on the same page with no guidance on which
should hold the FAQ.

**Fix:** This is a content/plan problem, NOT a rerender. Rerendering
re-emits the same empty shells from the same empty `content_data`. The
correct sequence:
1. Decide the canonical surface (the structured `faq` component is the
   intended UX — accordion, not prose).
2. Remove the redundant `generic-text-block` from the page (and from the
   plan once the plan is fixed — see stale-plan entry).
3. Populate the `faq` component's `questions` data — either migrate the
   Q&A already written in the generic block into a `questions` array on
   the faq page_component, or rebuild the faq section via the content
   writer with the generic block removed so the writer targets the faq
   component.
4. Relink the orphaned `component_id` (see orphaned-without-data-component
   entry below) so future rerenders bind the template instead of
   re-emitting frozen HTML.

**General lesson:** for any "renders empty" bug, walk
rendered HTML → page_components (orphan? data?) → content_components
(what key?) → content_data/content_item (is the data there?) →
sections vs plan (is there a competing slot?). The layer that first shows
"nothing here" is your cause. Don't trigger a rebuild before STEP 5 —
a rebuild against an unchanged plan reproduces the same split.

---

### Page exists with sections but is absent from `site_plan` (stale plan)

**Symptom:** A page renders and is in nav, but a `site_plan` lookup for
it returns zero rows. For gaswholesalers the `faq` page has
`pages.sections` populated and a live faq.html, but the `site_plan`
aspect lists only 8 pages and faq is not among them.

**Root cause:** Pages added *after* initial planning — by the
content-gap-planner or the improvement loop — are written directly to the
`pages` table (with `pages.sections`) but the new page is **not written
back to `site_specs.site_plan`**. The plan reflects only the original
build. Over time the plan drifts from reality: the more gap-planned pages
a site accumulates, the staler its plan.

**Why it matters:**
- Any process that reads `site_plan` as the source of truth for "what
  pages should exist" (audits, regeneration, sitemap planning) will miss
  gap-planned pages.
- `load_page_sections_from_spec` reads `site_plan` first and falls back to
  `pages.sections`. For pages absent from the plan, it silently uses the
  fallback — which works, but means the plan can never improve those
  pages' section briefs because it doesn't know they exist.
- Debugging "what should this page contain" by reading the plan gives a
  false negative — you conclude the page is unplanned/rogue when it was
  legitimately gap-added.

**Diagnosis:**

```sql
-- Pages that exist but are not in the site_plan
SELECT p.name, p.page_type, p.build_status
FROM pages p
WHERE p.site_id = '<site-id>'
  AND p.status IN ('active','deployed')
  AND NOT EXISTS (
    SELECT 1
    FROM site_specs ss,
         jsonb_array_elements(ss.data #> '{pages}') AS pl
    WHERE ss.site_id = p.site_id
      AND ss.aspect = 'site_plan'
      AND pl->>'name' = p.name
  )
ORDER BY p.name;
```

Any rows are pages the plan doesn't know about.

**Fix:** The gap planner's `apply_gap_plan` step (which creates the page
record + nav items + build work item) should also append the new page to
`site_specs.site_plan` via a deep-merge update, mirroring how
`enrich_news_feed` updates the classification aspect. Until that's added,
the plan must be reconciled manually or by a periodic
"plan-reconciliation" check that diffs `pages` against `site_plan` and
back-fills missing entries.

This is filed as a structural improvement, not just a data fix — patching
this one site's plan won't stop the next gap-planned page from drifting.

---

### Page plan sections are bare component names with no briefs (planner depth gap)

**Symptom:** Two content surfaces on one page compete for the same
content, and the writer fills the wrong one (see the FAQ empty-shells
entry). Underlying cause is that the plan gives the writer nothing to
disambiguate them.

**Root cause:** `site_plan.pages[].sections` is an array of bare strings:

```json
"sections": ["hero", "generic-text-block", "faq", "call_to_action"]
```

There is no per-section brief — no statement of what content each section
should hold, what data it binds, what audience/intent it serves, or how
it differs from a sibling section. When a page has both a freeform
`generic-text-block` and a structured `faq`, nothing in the plan tells
the content writer that the Q&A belongs in the `faq` component. The writer
makes a reasonable-but-wrong choice and the structured component is left
empty.

**Why a richer plan would prevent a class of bugs:**
- **Disambiguation:** a brief on each section ("faq: 5-7 question/answer
  pairs addressing procurement objections; generic-text-block: narrative
  intro, no Q&A") routes content to the right slot.
- **Validation surface:** with declared section intent, a post-build check
  can flag "faq component empty but plan says it needs ≥3 questions."
- **Duplicate detection at plan time:** a planner that records section
  intent can catch "this page has two sections that both want the FAQ
  content" before the build, rather than after.

**Proposed shape** (illustrative — for discussion, not yet implemented):

```json
"sections": [
  {
    "component": "faq",
    "intent": "Structured Q&A accordion answering common procurement and supply questions",
    "data": { "questions": "5-7 items, each a real buyer objection + answer" },
    "audience": "procurement managers evaluating bulk supply",
    "not": "narrative prose — that belongs in an intro section if present"
  }
]
```

Backward-compatible: the loader can accept either a bare string
(`"faq"`) or an object (`{"component": "faq", ...}`), treating the string
form as "component only, no brief." Existing plans keep working; new
plans can carry briefs.

**Where this would be produced:** the site planner / chief-strategist step
that emits `site_plan`. Enriching its prompt to produce per-section briefs
is the change. The content writer's prompt then consumes the brief for
the section it's building.

**Relationship to the stale-plan entry:** these compound. A plan that is
both *missing pages* and *too thin on the pages it has* gives the build
pipeline very little to work with. Both want fixing in the planner, not
per-site.

---

### Section 0 — new assumption-checklist note

**16. "Renders empty" is a data-binding diagnosis, not a template one.**
When a component shows structural shells with no content, resist editing
the template or re-triggering a render first. The template ran — that's
why the shells exist. Walk the data path: is the page_component orphaned
(`component_id IS NULL`)? Is `content_data`/`content_item_id` empty? Is
there a competing section that got the content instead? Only after the
data path is cleared should you suspect the template. A rerender before
this diagnosis just re-emits the same empty shells and costs a cycle.
