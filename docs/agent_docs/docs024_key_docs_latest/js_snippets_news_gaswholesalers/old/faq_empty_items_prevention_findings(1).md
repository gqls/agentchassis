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

## Fix targets (prioritised) — verified against the live agent_definitions

Both planner prompts were pulled from `agent_definitions` and confirm the
cause is in prompt content + an unimplemented validation flag. No content
writer change is needed (test-proven correct). Fixes are prompt edits +
making one existing flag actually do its job — complexity stays in Go
action code and prompt config, not in workflows.

### A. Stop pairing a generic block with a structured component (prompts)

**content-gap-planner** — its prompt's `new_page` example JSON literally
hardcodes the bad shape:
```json
"new_page": { "sections": ["hero", "generic-text-block", "call-to-action"] }
```
The LLM anchors on this example, then adds `faq` when it recognises a FAQ
page → `["hero","generic-text-block","faq","call_to_action"]` (exactly the
gaswholesalers shape). Its `add_to_page` example also shows
`["faq-section","call-to-action"]` — `faq-section` is not the real
function (`faq`), so the example also teaches the wrong name.

Fix: change the `new_page` example to not pre-seed a generic block (e.g.
use `["hero","faq","call-to-action"]` or a neutral
`["hero","<content-section>","call-to-action"]` with a note), fix the
`add_to_page` example to `["faq","call-to-action"]`, and add a selection
rule (shared text for both planners):
> Structured components such as `faq`, `pricing`, and data-driven
> `features` are complete content surfaces. Do not pair them with a
> `generic-text-block` covering the same material on the same page.

**site-planner** — its prompt has good "ONLY use listed components" rules
and standard mappings, but the mappings list omits `faq` entirely, so the
LLM has no canonical instruction for FAQ pages.

Fix: add `faq` (and `pricing`) to the standard-mappings list, plus the
same no-pairing rule.

**Default fallback** (`applyNewPage`, `apply_gap_plan_action.go`): the
hardcoded default `["hero","generic-text-block","call-to-action"]` is a
reasonable generic default, but for recognised archetypes (faq, contact,
pricing) it should match the archetype. A small archetype→sections map
keyed on page_name/page_type covers common cases; unknown types keep the
generic default. Lower priority than the prompt edits since the LLM
usually overrides the default — but it's the backstop when it doesn't.

### B. Make `validate_components` actually validate + normalise (Go)

**Key finding:** `validate_site_plan`'s config already sets
`validate_components: true`, but `ValidateSitePlanAction` does NOT
implement it. The action strips site-chrome components, ensures required
pages, caps page count, and sets style/image defaults — but there is no
code path that checks each section name is a real
`content_components.function`, and no display-name→function resolution.
The flag is effectively dead. This is why `"FAQ Section"` (a display
name) passed straight through into a sections array and orphaned its
component.

This action is the right enforcement point: it already runs on every
site-plan, already loads from `content_components` (for chrome stripping),
and has the flag wired in config.

Implement it, reusing what exists:
1. Load the valid component set once: `SELECT function, name, display_name
   FROM content_components WHERE component_level IN ('section','element')
   AND is_active = true`. Build two maps: `function→true` and
   `display_name→function` (and `name→function`).
2. For each section name in each page:
   - If it's already a valid `function`, keep it.
   - Else run `NormalizeComponentFunction(name)` (EXISTS at
     ~line 31067; handles `call_to_action`→`call-to-action`, camelCase).
     If the normalised value is a valid function, use it.
   - Else look it up in the `display_name→function` / `name→function`
     maps (this is the NEW bit — handles `"FAQ Section"`→`faq`, which
     normalisation alone can't, since it'd only produce `faq-section`).
   - Else drop the section and log it (don't emit an unresolvable name
     that will orphan downstream).

Note: `NormalizeComponentFunction` and `NormalizeSectionNames` already
exist and are used in save_page_sections; reuse them. The only genuinely
new logic is the display_name/name → function lookup against the loaded
set. This also fixes the underscore/hyphen inconsistency seen across
sites (`call_to_action` vs `call-to-action`) as a side benefit.

The content-gap-planner path doesn't currently route through
`validate_site_plan` (it applies via `apply_gap_plan`). Either route its
`new_page.sections` through the same normalisation helper, or have
`applyNewPage` normalise+validate section names before writing the page
record, using the same component-set lookup.

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

**Result (verified):** the rebuild completed `complete` with the `faq`
component linked, `build_status=deployed`, `q_count=10`. Same clean
outcome as the isolated test, confirming the corrected-plan path produces
a populated faq end to end. This doubles as the regression proof for the
whole diagnosis: remove the duplicate surface, and the existing pipeline
builds the FAQ correctly with no writer change.

Note: the writer build for faq-test also emitted a `needs_section_data`
work item (`status=needs_human_review`, no handler_agent) as a child of
the build — even though the faq itself populated correctly. Worth a
separate look at why a successful structured build still raises a
section-data review item; relates to the debugging guide's
"needs_section_data items going straight to wont_fix" pattern.
