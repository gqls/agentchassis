# Site Planner Depth & Freshness — Two Structural Concerns

Surfaced 2026-05-20 while diagnosing the gaswholesalers.com FAQ
empty-items bug. The empty FAQ was the visible symptom; underneath it
were two planner-level problems that will keep producing similar bugs on
other sites until addressed at the planner, not per-site.

Companion to the debugging-guide addendum
(`016_debugging_guide_addendum_faq_diagnosis.md`), which documents how to
*diagnose* the symptom. This doc is about *preventing the class*.

## Concern 1: section plans are bare names, with no briefs

### Current state

`site_specs.site_plan.pages[].sections` is an array of bare strings:

```json
{
  "name": "faq",
  "sections": ["hero", "generic-text-block", "faq", "call_to_action"]
}
```

Each entry names a component function. Nothing says what the section is
*for*, what content it should hold, what data it binds, or how it differs
from a sibling section on the same page.

### How this caused the FAQ bug

The faq page plan carried BOTH `generic-text-block` (a freeform prose
section) and `faq` (a structured Q&A accordion). With no brief on either,
the content writer wrote all the Q&A content into the generic-text-block
as bolded-question prose, and left the structured `faq` component with
empty placeholder questions. The result: a fully-populated prose section
followed by an empty accordion — the same content conceptually targeted
at the wrong slot.

A brief on each section would have routed the content correctly:
- `generic-text-block`: "narrative intro to the page topic, no Q&A"
- `faq`: "5-7 question/answer pairs addressing common buyer objections"

### Why richer plans prevent a class of bugs

1. **Disambiguation between similar sections.** When two sections could
   plausibly hold the same content, the brief decides. Without it the
   writer guesses.
2. **A validation surface.** With declared intent ("faq needs ≥3
   questions"), a post-build check can assert the structured component
   was actually populated, and flag it when empty — catching this bug
   automatically instead of by eyeball.
3. **Plan-time duplicate detection.** A planner that records section
   intent can notice "this page has two sections that both want the FAQ
   content" before the build runs.
4. **Better content targeting generally.** Briefs let the writer produce
   section-appropriate content (audience, tone, length) rather than
   generic filler.

### Proposed shape (backward-compatible)

Allow each section to be EITHER a bare string (current) or an object with
a brief:

```json
"sections": [
  "hero",
  {
    "component": "faq",
    "intent": "Structured Q&A accordion for procurement/supply questions",
    "data": {
      "questions": "5-7 items; each a real buyer objection + a direct answer"
    },
    "audience": "procurement managers evaluating bulk fuel supply",
    "not": "narrative prose — intro narrative belongs in a separate intro section"
  },
  "call_to_action"
]
```

The section loader (`load_page_sections_from_spec`) accepts both forms:
a string is treated as "component only, no brief"; an object carries the
brief. Existing plans keep working unchanged; new or re-planned pages can
carry briefs. The content writer's prompt consumes the brief for the
section it builds.

### Where the change lives

The site planner / chief-strategist step that emits `site_plan`. Its
prompt is enriched to produce per-section briefs. This is one prompt
change plus a loader that tolerates the richer shape. No schema migration
— `site_plan` is jsonb.

### Token-budget caveat

Per the debugging guide's assumption #7 (token budgets scale with
structured output): adding a brief to every section on a multi-page site
materially increases the planner's output size. Estimate the token count
for a large site (e.g. 15 pages × 5 sections × a 40-token brief = ~3000
extra tokens) and confirm it fits the planner's `max_tokens` before
shipping, or the planner's `validate_*` step will fail with
`unexpected end of JSON input`.

## Concern 2: gap-planned pages aren't written back to the plan

### Current state

The faq page exists (live faq.html, in nav, `pages.sections` populated)
but is **absent from `site_plan` entirely**. The plan lists 8 pages; faq
is not one of them.

Pages added after the initial build — by the content-gap-planner or the
improvement loop — get a `pages` row and nav entries, but the new page is
never appended to `site_specs.site_plan`. The plan reflects only the
original build and drifts further from reality with every gap-added page.

### Why it matters

- Anything reading `site_plan` as the authoritative page list (audits,
  sitemap planning, regeneration, plan-based validation) silently misses
  gap-planned pages.
- `load_page_sections_from_spec` reads the plan first, falls back to
  `pages.sections`. Pages absent from the plan work via fallback — but
  the plan can never enrich their section briefs (Concern 1) because it
  doesn't know they exist.
- Debugging "what should this page contain" via the plan gives a false
  negative: the page looks unplanned/rogue when it was a legitimate
  gap addition.

### Fix

`apply_gap_plan` (which already creates the page record, nav items, and
build work item) should also append the new page to
`site_specs.site_plan` via a deep-merge update — mirroring how
`enrich_news_feed` deep-merges into the classification aspect. The
appended entry should carry a brief (Concern 1) from the gap planner's
own reasoning about why the page is being added, which it already has.

As a safety net, a periodic plan-reconciliation discovery check can diff
`pages` against `site_plan` and back-fill missing entries (with a
generated placeholder brief for any page that predates brief support).

### Diagnosis query (reusable)

```sql
-- Pages that exist but are missing from site_plan
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

## How the two concerns compound

A plan that is both *missing pages* and *too thin on the pages it has*
gives the build pipeline very little to work with. The FAQ page hit both:
it wasn't in the plan at all, and even its `pages.sections` fallback was
bare strings that couldn't route the FAQ content to the right component.

Both fixes belong in the planner and gap-planner, not in per-site data
patches. Fixing gaswholesalers' faq page by hand (prune the duplicate
section, populate the questions) resolves the symptom; only the planner
changes stop the next site from reproducing it.

## Immediate vs structural

| Action | Type | When |
|---|---|---|
| Prune duplicate `generic-text-block`, populate `faq` questions on gaswholesalers | Per-site data fix | Now (unblocks the live page) |
| Back-fill faq into gaswholesalers `site_plan` | Per-site data fix | Now (or via reconciliation check) |
| `apply_gap_plan` writes new pages back to `site_plan` | Structural (gap planner) | Scheduled |
| Section briefs in `site_plan` + brief-aware loader + writer | Structural (planner) | Scheduled |
| Plan-reconciliation discovery check | Structural (safety net) | Scheduled |
| Post-build validation: structured component populated per brief | Structural (validation) | After briefs exist |
