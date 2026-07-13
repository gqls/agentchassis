# FAQ Empty-Items — Prevention Findings & Fix Targets

Closes the diagnostic arc on the gaswholesalers.com empty-FAQ bug.
Establishes the cause with an isolated build test, scopes it across sites,
traces the origin in code, and lists the structural fixes with the exact
step each belongs to. 2026-05-20.

## The test that settled the cause

A throwaway `faq-test` page was created with sections
`["hero", "faq", "call_to_action"]` — the structured `faq` component
standing alone, no `generic-text-block` — and built through the full
production path: `site_work_items` row (triaged) → build-dispatch-loop
claimed it → spawned page-build-handler → content writer → render →
save_page_sections.

Result: the `faq` component came out **linked** (component_id set),
`build_status=deployed`, with a populated `questions` array
(`q_count = 9`) and a fully-rendered accordion. The hero populated too.

**Conclusion:** the content writer's structured-field generation works.
Given a clean plan where `faq` is the content surface, it generates the
`questions` array, binds it, and saves it linked. The writer is not the
bug.

## Cross-site scope

faq pages across active sites:

| Site | Created | faq sections | Clean? |
|---|---|---|---|
| leopardessconsulting.co.uk | 2026-02-06 | `["hero","faq","call_to_action"]` | Yes |
| gaswholesalers.com | 2026-03-15 | `["hero","generic-text-block","faq","call_to_action"]` | **No** |
| finetuning.uk | 2026-03-21 | `["hero","faq","call-to-action"]` | Yes |

Two of three faq pages are clean. So the duplicate-surface is NOT applied
uniformly to every faq page. But it is also not purely historical — see
below.

## The risk is current, not just legacy

A scan of pages created in the last 30 days that pair `generic-text-block`
with a structured component (faq / pricing / features) returned 17 rows.
Examples:

- `wholesale-pricing-explained` (2026-05-01):
  `["hero","generic-text-block","features","pricing","generic-text-block","faq","call-to-action"]`
- `pricing-transparency` (2026-04-24):
  `["hero","generic-text-block","features","generic-text-block","faq","call-to-action"]`
- `fuel-pricing-framework` (2026-04-24):
  `["hero","generic-text-block","pricing","features","faq","call-to-action"]`
- `containment-first-architecture` (2026-05-01):
  `["hero","generic-text-block","features","generic-text-block","differentiators","FAQ Section","call-to-action"]`

So the current pipeline regularly produces pages that pair a freeform
`generic-text-block` with a structured component. Each such page is a
candidate for the same content-diversion bug (writer fills the prose
block, structured component left empty). Whether each one actually
emptied depends on the writer's routing per page, but the **structural
setup for the bug is being produced now**.

## Two distinct defects found

### Defect 1 — duplicate content surfaces on one page

Pages carry both `generic-text-block` (freeform prose) and a structured
component (`faq`, `pricing`) intended to hold the same content. The
writer, with no brief distinguishing them, can route content to the
freeform block and leave the structured one empty.

Origin (two contributing sources):

1. **`applyNewPage` default** (gap planner apply step,
   `apply_gap_plan_action.go`): when a new content page is created and the
   plan doesn't specify sections, the default is
   `["hero", "generic-text-block", "call-to-action"]`. If the LLM then
   adds a `faq` (because it recognises a FAQ page) without removing the
   generic block, the page gets both.

2. **Planner LLM section selection** (content-gap-planner /
   site-planner agent_definition prompt): the LLM is handed a flat
   component list (`formatComponentsForPrompt` — grouped by category,
   `function` + description) with no signal that `generic-text-block` and
   `faq`/`pricing` are competing content surfaces. Nothing discourages
   pairing them.

### Defect 2 — section name uses display name, not function

`containment-first-architecture` has a section literally named
`"FAQ Section"` — that is the component's `display_name`, not its
`function` (`faq`). Section lists must carry the kebab `function` so the
renderer/linker can resolve the component. A display-name entry can't be
matched to a `content_components.function`, which orphans the
page_component (`component_id` NULL) — exactly the orphaning seen on the
gaswholesalers faq page. Whatever planner path emitted `"FAQ Section"`
is leaking display names into the sections array.

## Fix targets (prioritised)

### A. Stop pairing a generic block with a structured component (planner)

Two places:

- **Default fallback** (`applyNewPage`): when the page is a recognised
  archetype (faq, contact, pricing), the default sections should be the
  archetype's shape (`["hero","faq","call_to_action"]` for faq), not the
  generic `["hero","generic-text-block","call-to-action"]`. A small
  archetype→sections map keyed on page_name/page_type covers the common
  cases; unknown types keep the generic default.

- **Planner prompt** (content-gap-planner / site-planner
  agent_definition): tell the LLM that structured components
  (faq, pricing, features-with-data) are complete content surfaces and
  should not be combined with a `generic-text-block` covering the same
  content. Best done by tagging components with a "role" in
  `formatComponentsForPrompt` (e.g. `[structured content surface]` vs
  `[freeform prose]`) and adding a selection rule.

### B. Enforce function-name (not display-name) in section lists (planner + validation)

- Wherever the planner emits sections, normalise each entry to the
  component `function` (kebab) before writing — never the display name.
  `NormalizeComponentFunction` already exists and is used in
  save_page_sections; the planner output should pass through the same
  normalisation, or validate against `content_components.function`.

- A cheap validation backstop: when a plan/section list is applied,
  reject or normalise any section name that isn't a known
  `content_components.function`. A `"FAQ Section"` entry would be caught.

### C. Per-section briefs (planner + already-built consumer)

The reinforcing fix from `site_planner_depth_and_freshness_concerns.md`:
emit a brief per section so the writer can disambiguate even when two
surfaces legitimately coexist. The consumer already exists —
`plan_sections.sectionDescription` reads `section_descriptions` /
`section_types[].description`. The planner just needs to emit one of
those shapes.

### D. Post-build validation: structured component populated (validation)

After a build, assert that any component whose `input_schema` declares a
required structured field (e.g. `faq.questions min_items 3`) actually has
that field populated in `content_data`. If empty, flag the page rather
than deploying a silently-empty accordion. This catches the bug class
regardless of which planner produced the plan.

## Sequencing note

The writer is correct (test-proven), so none of these fixes touch it. A
/ B are the highest leverage — they stop the malformed plans at source.
C and D are defence in depth. Fixing A/B prevents the *production* of
duplicate-surface and display-name plans; D catches anything that slips
through before it reaches users.

## Live gaswholesalers faq repair (in progress)

The live faq page was reset to `["hero","faq","call_to_action"]`, its 4
stale page_components deleted, and a rebuild work item queued
(`manual_faq_repair_5fe15466`) through the same path the test validated.
This repairs the page via the proven pipeline rather than by hand. The
old generic-block prose content was deleted with the components; the
rebuild regenerates fresh FAQ content (writer test-proven to populate it).

Note: the writer build for faq-test also emitted a `needs_section_data`
work item (`status=needs_human_review`, no handler_agent) as a child of
the build — even though the faq itself populated correctly. Worth a
separate look at why a successful structured build still raises a
section-data review item; relates to the debugging guide's
"needs_section_data items going straight to wont_fix" pattern.
