# FOCUS — FAQ empty-items, page content-creation flow, and prevention

The complete arc of the gaswholesalers.com empty-FAQ bug: the build-path trace
that isolated the cause, the finding that it was a planner problem (not the
writer), the two defects, the four fixes with their implementations, the
planner-level structural prevention, and the deployed status.

> Consolidated from `faq_empty_items_prevention_findings.md` (the spine),
> `page_content_creation_flow.md` (the trace + isolated-test method),
> `planner_prompt_fixes_defect1.md` (Fix A implementation),
> `validate_components_implementation.md` (Fix B implementation),
> `site_planner_depth_and_freshness_concerns.md` (Fix C + stale-plan), and the
> FAQ portion of `HANDOFF_2026-05-21`. The *diagnostic method* for empty-shell
> bugs lives in `016_debugging_guide_addenda.md`; this doc is the cause, the
> fixes, and the prevention. Working-session scratch
> (`015_content_data_persisted.md`, `empty_faqs.md`) reached the same
> conclusions and is absorbed here.

---

# Deployed status (2026-05-21)

The FAQ empty-items bug is resolved and prevention is deployed. Root cause was
a duplicate-content-surface plan (`generic-text-block` + `faq` on one page)
made by the planners — NOT a content-writer fault, proven by an isolated build
test that produced a populated 9-item accordion. The live gaswholesalers faq
page was repaired through the pipeline (q_count=10, deployed). Prevention
shipped on three fronts, all live:

- **content-gap-planner prompt** (`fix1_content_gap_planner_prompt.sql`):
  removed the hardcoded `generic-text-block` from the new_page example, fixed
  the add_to_page example to `faq`, added the no-pairing + use-function-name
  rules. Confirmed live (`gap_has_rule = t`).
- **site-planner prompt** (`fix2_site_planner_prompt.sql`): component list now
  shows `[function]` first with a use-only-that instruction (kills the
  `"FAQ Section"` display-name leak); added faq/pricing to standard mappings;
  no-pairing rule; `call_to_action` → `call-to-action`. Confirmed live
  (`site_has_faq_mapping = t`).
- **Go (chassis v1.0.1029):** `validate_components` flag implemented in
  `ValidateSitePlanAction` (resolve each section name to a real
  `content_components.function`); the same resolver wired into `applyNewPage`
  and `applyAddToPage` (the gap-planner path doesn't route through
  validate_site_plan); archetype-aware `defaultSectionsForPage`. Files
  `v3_site_actions.go`, `apply_gap_plan_action.go`.

Prompt fixes take effect immediately (`loadAgentDefinition` reads per-spawn, no
cache); the Go fixes are in the deployed image. Remaining defence-in-depth
(per-section briefs, post-build structured-field validation) is tracked in
`TODO_remaining_work.md`. Source: `HANDOFF_2026-05-21_faq_prevention_and_news.md`.

---

# Findings and fix targets

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

---

# The page content-creation flow (the trace that isolated the cause)

How a page goes from a `pages` row (sections list) to populated, deployed
HTML with `page_components` rows carrying `content_data` and
`component_id`. Traced from chassis source 2026-05-20 while diagnosing the
FAQ empty-items bug. Marks clearly what is verified in code vs what the
isolated build test is designed to pin down.

## Entry: how a build is initiated

`page-build-handler` is the agent that builds a page. It is reached two
ways:

1. **Via work item** (normal path). A `site_work_items` row with
   `handler_agent = 'page-build-handler'` is claimed by the
   build-dispatch-loop, which spawns the handler with the item's `spec`
   as `input_data.spec`. Observed spec shape (from live data):
   ```json
   {
     "reason": "not_built",
     "plan_id": "<uuid>",
     "page_name": "contact",
     "page_role": "content"
   }
   ```
   Note what the spec does NOT contain: the sections list. The handler
   reads sections from the `pages` row, not the work item.

2. **Direct orchestrate** (manual/test). A Kafka message to
   `system.agent.generic.requests` with `action=orchestrate`,
   `config.agent_type=page-build-handler`, and `input_data.spec.page_name`
   (or `page_id`). Same contract — the handler still loads the page row.

## The page-build-handler workflow (step by step)

`start_step = ensure_site_record`. The relevant chain:

```
ensure_site_record
  → load_page_record       (load the pages row: sections, title, page_type)
  → check_page_found
  → plan_sections          (resolve each section's data readiness)
  → check_has_ready_sections
  → spawn_content_writer    (generate content for ready sections)
  → [content writer runs: per-section generation + render]
  → compile / assemble      (CompilePageSectionsAction → sections_metadata)
  → deploy_page             (git commit)
  → save_page_sections      (persist page_components: html, content_data, component_id)
  → ...
```

### load_page_record — establishes the section list

`LoadPageRecordAction` looks up the page by `site_id` +
`page_name` (preferred) or `page_id`, and returns the full record
including parsed `sections`. Config:

```json
"config": {
  "page_id":   "input_data.spec.page_id",
  "site_id":   "site_record.site_id",
  "page_name": "input_data.spec.page_name"
}
```

It also has contract-compliant fallbacks for `page_name`
(`input_data.spec.page_name` → `input_data.spec.page.name` →
`current_page.name` → `page_record.name`).

**This is the authority for what sections the page has.** Whatever is in
`pages.sections` drives the rest of the build. For the FAQ page that was
`["hero", "generic-text-block", "faq", "call_to_action"]`.

### plan_sections — resolves data readiness per section

`PlanSectionsAction` reads each section's component `input_schema` and
triages each field by its `source`:

| Field `source` | Handling |
|---|---|
| `llm` | Added to `llm_fields` — the content writer must generate it. Section stays "ready". |
| `query.*` | Resolved now via the `queryresolve` package (SQL). Empty result → empty slice. |
| `renderer` / `static` | Deferred to render time — always "available". |
| `site_specs.*`, `pages.*`, `site_assets.*`, `config.*` | Resolved against current specs/pages/assets. Missing + required → `on_missing` rule (defer / skip / fallback). |

Section status becomes `ready` / `deferred` / `skipped`. Only `ready`
sections are passed to the content writer.

For the `faq` component, `input_schema.fields.questions` is
`type: array, items: {question, answer}, min_items: 3, source: llm,
required: true`. Because `source = llm`, plan_sections marks `questions`
as an LLM field and the section as **ready**. So the FAQ section is
correctly identified as "writer must generate a `questions` array." The
triage is not the bug.

`plan_sections` also has a `sectionDescription` resolver that already
looks for per-section briefs in the plan under three shapes:
`page.section_descriptions[sectionType]`,
`page.section_types[].description`, and a fall-back to `page.purpose`.
**The consumer for richer plans exists; the planner just doesn't emit any
of these shapes yet** (it emits bare section-name strings). See
`site_planner_depth_and_freshness_concerns.md`.

### content writer — generates content for ready sections

The writer is spawned per the workflow and processes ready sections (a
`process_sections_loop`). For each section it builds a prompt and calls
the LLM, then the result is consumed by `extractResponseContent`.

**Verified:** `extractResponseContent` returns the LLM result as a flat
**string** — it walks the response for `*_content` / `*_result` / `result`
/ `content` keys and returns the first string it finds. There is no
observed branch in this helper that parses the result into a structured
array of objects.

**The open question (what the test pins down):** for a component whose
schema needs a structured field (FAQ's `questions` = array of
`{question, answer}`), something must turn the LLM output into that array
and place it in the section's `content_data` under the key `questions`.
Whether the content-writer agent_definition's workflow (its prompt + a
parse step) does this is not visible in the Go action source — it lives
in the agent_definition config. The render and save layers (below) both
faithfully carry whatever `content_data` they're handed, so if
`content_data.questions` is absent here, it is absent everywhere
downstream and the accordion renders empty. The isolated faq-only build
test is designed to determine whether the writer produces
`content_data.questions` at all.

### render — binds content_data into the template

`RenderComponentAction` builds a `RenderContext` whose `ContentData` map
starts with site-level defaults (company_name, nav, cta, etc.) and then
**merges the section's `content_data` on top** (section values win):

```go
// Now merge the actual section content_data on top — these take priority
for key, value := range contentData {
    renderCtx.ContentData[key] = value
}
```

The template then reads `{{.headline}}`, `{{range .questions}}`, etc.
from this merged map. **Verified correct:** if `content_data` contains a
populated `questions` array, the FAQ template renders populated
accordion items. If `content_data.questions` is empty/absent, the
`{{range}}` produces empty shells. The render layer is faithful — it is
not the bug.

Output of the render for each section: `rendered_html`, `content_data`,
`component_id`, `component_function`/`component_name`.

### compile — gathers sections into metadata

`CompilePageSectionsAction` assembles the rendered sections into the page
HTML and produces a `sections_metadata` array, each entry carrying
`rendered_html` + the component metadata (`component_id`,
`component_function`, `content_data`). This metadata array is the
structured hand-off to the save step.

### save_page_sections — persists page_components

`SavePageSectionsAction` has two paths:

1. **Structured metadata path** (preferred): reads
   `sections_metadata`, and for each entry builds a `SectionData`
   `{ComponentName, ComponentID, HTML, Position, ContentData}` via
   `extractSectionsFromMetadata`. `content_data` and `component_id` are
   taken straight from the metadata entry.
2. **HTML-parsing fallback**: regex-extracts `<section>` blocks from the
   assembled HTML when no metadata is present (adopted sites / older
   pipelines). This path has no `content_data` and looks up
   `component_id` from the `data-component` attribute.

Both paths then run `enrichSectionsWithPlannedNames` and
`enrichSectionsWithComponentIDs` to fill `component_id` from
`content_components.function` matching, before the INSERT into
`page_components` (sets `rendered_html`, `slot_name`, `component_id`,
`content_data`, `content_brief`, `build_status`).

**Verified failure mode (matches the FAQ page):** when a section's
metadata is incomplete — no recovered `component_id`/`component_function`
and no `content_data` — the section saves with `ComponentName="section"`,
`component_id` NULL (orphaned), and empty `content_data`. The
`extractSectionFromMap` comment in the source documents exactly this:
unrecovered metadata → `enrichSectionsWithComponentIDs` skips the section
→ `page_components.component_id` NULL. The FAQ page's positions 2 and 3
are orphaned in precisely this way.

## The data contract at each hop

| Hop | Carries page identity as | Carries content as |
|---|---|---|
| Work item / trigger | `spec.page_name` / `spec.page_id` | (none — sections come from DB) |
| load_page_record | `page_record` (with `sections`) | — |
| plan_sections | section names + component info | `llm_fields` (to generate), `resolved_data` (already available) |
| content writer | per-section | **section `content_data`** ← the critical construction |
| RenderComponentAction | per-section | merges `content_data` → template |
| CompilePageSections | `sections_metadata[]` | `content_data` + `component_id` per entry |
| save_page_sections | `page_id` + position | INSERT `content_data` + `component_id` |

The single point where structured content (`questions` array) must be
*constructed* is the content writer. Every layer after it merely carries
or binds what it was given. Every layer before it (triage, page load) is
verified to set the FAQ section up correctly as "needs an LLM-generated
`questions` array." That narrows the cause to the writer's
structured-field generation, which the isolated test confirms.

## Creating a build target (test pages)

`pages.url` is `NOT NULL` — the INSERT must set it. Corrected form:

```sql
INSERT INTO pages (site_id, name, url, page_type, sections,
                   status, build_status, title, in_header, in_footer)
VALUES (
  '<site-id>',
  'faq-test',
  '/faq-test.html',
  'content',
  '["hero", "faq", "call_to_action"]'::jsonb,
  'active', 'pending', 'FAQ Build Test',
  false, false          -- keep test page out of nav/footer
)
RETURNING id, name, url, sections;
```

`page_type` must satisfy `chk_page_type_kebab_case`. `in_header` /
`in_footer` default true — set false for a throwaway so it doesn't enter
navigation (the `trg_invalidate_nav_on_page_change` trigger fires on
insert regardless, which is harmless).

Cleanup after the test: `DELETE FROM pages WHERE site_id = '<site-id>'
AND name = 'faq-test'` (cascades to its `page_components`).

## Trigger (direct orchestrate)

```json
{
  "action": "orchestrate",
  "config": {"agent_type": "page-build-handler"},
  "input_data": {
    "spec": {
      "page_name": "faq-test",
      "page_role": "content",
      "reason": "not_built"
    }
  }
}
```

Sent to `system.agent.generic.requests` with the standard header set
(correlation_id, orchestration_id, request_id, message_id,
message_type=request, client_id, action=orchestrate, sender_agent_type,
responses_topic, timestamp), matching the single-page rerender trigger
already in use this session.

## What the test reads out

After the build completes:

```sql
SELECT pc.position, cc.function, pc.component_id IS NOT NULL AS linked,
       jsonb_typeof(pc.content_data) AS cd_type,
       CASE WHEN pc.content_data ? 'questions'
            THEN jsonb_array_length(pc.content_data->'questions') END AS q_count,
       LEFT(pc.rendered_html, 120) AS html_start
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
LEFT JOIN content_components cc ON pc.component_id = cc.id
WHERE p.site_id = '<site-id>' AND p.name = 'faq-test'
ORDER BY pc.position;
```

- **faq populated** (`q_count >= 3`, accordion has text) → the writer CAN
  build structured fields when faq stands alone; the live faq page failed
  because of the competing `generic-text-block` (planner duplicate-surface
  cause).
- **faq empty** (`q_count` null/0, empty shells) even with faq alone →
  the writer's structured-field generation is the cause; fix is
  schema-aware generation (prompt the LLM for the declared structure,
  parse into `content_data[field]`), which then benefits every structured
  component, not just FAQ.

Either way the fix is provable by re-running the identical trigger, and
the live faq page is then repaired by the corrected pipeline rather than
by hand.

---

# Fix A implementation — planner prompt edits (Defect 1: duplicate surfaces)

Before/after edits against the two planner `agent_definitions` pulled from
the DB, plus the `applyNewPage` default-sections change. Pairs with the
`validate_components` Go implementation (separate doc) which handles
Defect 2 (display-name leak).

Goal: stop the planners emitting a `generic-text-block` paired with a
structured component (`faq`, `pricing`) covering the same content — the
plan shape that emptied the gaswholesalers FAQ.

A note that applies throughout: `content_components.function` is
hyphen-only (`chk_function_kebab_case` rejects underscores). So
`call_to_action` is NOT a valid function — canonical is `call-to-action`.
The current prompts use `call_to_action` in examples, which only works
because downstream normalisation is lenient. These fixes also correct the
examples to hyphens so the prompts teach the DB-valid form.

---

## Fix 1 — content-gap-planner prompt

Agent: `content-gap-planner` (id `637b750c-...`). The `new_page` example
JSON in its `prompt_template` hardcodes the bad shape and the
`add_to_page` example uses a non-canonical name.

### 1a. Fix the `new_page` example (removes the hardcoded generic-text-block)

**Before:**
```json
  "new_page": {
    "name": "kebab-case-name",
    "title": "Page Title | Company Name",
    "page_type": "content",
    "purpose": "What this page covers and why",
    "sections": ["hero", "generic-text-block", "call-to-action"],
    "nav_label": "Nav Label",
    "in_header": true,
    "in_footer": true
  },
```

**After:**
```json
  "new_page": {
    "name": "kebab-case-name",
    "title": "Page Title | Company Name",
    "page_type": "content",
    "purpose": "What this page covers and why",
    "sections": ["hero", "<choose content sections by purpose>", "call-to-action"],
    "nav_label": "Nav Label",
    "in_header": true,
    "in_footer": true
  },
```

The placeholder stops the LLM anchoring on `generic-text-block` as a
default. For a FAQ page it will then choose `faq`; for a pricing page,
`pricing`; etc.

### 1b. Fix the `add_to_page` example (canonical name)

**Before:**
```json
  "add_to_page": {
    "page_name": "existing page name",
    "add_sections": ["faq-section", "call-to-action"],
    "content_guidance": "What the new sections should cover"
  },
```

**After:**
```json
  "add_to_page": {
    "page_name": "existing page name",
    "add_sections": ["faq", "call-to-action"],
    "content_guidance": "What the new sections should cover"
  },
```

`faq-section` is not the component function (`faq`). Using the wrong name
in the example teaches the LLM the wrong token.

### 1c. Add the no-pairing rule

Insert into the `## Your Task` section, after the approach descriptions
(A/B/C/D) and before the `Return ONLY valid JSON` line:

```
### Section selection rules
- Structured components such as `faq` and `pricing` are COMPLETE content
  surfaces — they hold their own content. Do NOT also add a
  `generic-text-block` covering the same material on the same page. Use a
  `generic-text-block` only for narrative content that a structured
  component does not already cover (e.g. a short intro that is clearly
  distinct from the FAQ items themselves).
- Use ONLY component function names from the Available Section Components
  list. Use the function name (the value in parentheses), never a display
  name.
```

---

## Fix 2 — site-planner prompt

Agent: `site-planner` (id `f7c8bee1-...`). Two gaps: `faq` is missing from
the standard mappings, and the component list interpolation shows three
names (`name`, `display_name`, `function`) without saying which to use in
`sections` — the source of the `"FAQ Section"` display-name leak.

### 2a. Add faq/pricing to the standard mappings

In the `Use these standard mappings:` list, add:

**Before** (end of the mappings list):
```
   - For about content: use "about-content"
   - For differentiators/why-us: use "differentiators-section"
```

**After:**
```
   - For about content: use "about-content"
   - For differentiators/why-us: use "differentiators-section"
   - For FAQ / question-and-answer content: use "faq"
   - For pricing tables/tiers: use "pricing"
```

### 2b. Clarify which name to use, in the component list instruction

**Before:**
```
## Available Section Components
The following components are available in our component library. You MUST use ONLY these exact component names in the "sections" arrays:

{{range .available_components}}
- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}
{{end}}
```

**After:**
```
## Available Section Components
The following components are available. In the "sections" arrays you MUST
use ONLY the function name — shown in [brackets] below — never the display
name or title.

{{range .available_components}}
- [{{.function}}] {{.display_name}}: {{.description}}
{{end}}
```

Putting the `function` first and in brackets, and dropping the bare
`{{.name}}` lead-in, removes the ambiguity that let the LLM emit
`"FAQ Section"`. The display name and description remain for the LLM to
understand what the component is, but the token it must use is
unmistakable.

### 2c. Add the no-pairing rule + fix the underscore example

In `STRICT RULES`, the home example uses `call_to_action` (underscore).
Change rule 6's neighbours and append a rule:

**Before** (rule 5–6 area and the example array):
```
      "sections": ["hero", "features", "testimonials", "call_to_action"]
...
5. Keep header navigation to 5-8 items maximum
6. Always include: index (home) and contact pages
```

**After:**
```
      "sections": ["hero", "features", "testimonials", "call-to-action"]
...
5. Keep header navigation to 5-8 items maximum
6. Always include: index (home) and contact pages
7. Structured components (faq, pricing, data-driven features) are complete
   content surfaces. Do not pair them with a generic-text-block covering
   the same content on the same page.
```

(Also update the standard-mapping line `For calls to action: use
"call_to_action"` → `use "call-to-action"`, and any other `call_to_action`
occurrences in the prompt, to the hyphen form.)

---

## Fix 3 — applyNewPage default (Go, backstop)

`apply_gap_plan_action.go`, `applyNewPage`. The hardcoded default fires
only when the LLM omits sections, but it is the last line of defence.

**Before:**
```go
	sections := []string{"hero", "generic-text-block", "call-to-action"}
	if sectionsRaw, ok := newPlan["sections"].([]interface{}); ok && len(sectionsRaw) > 0 {
		// ... use LLM-provided sections ...
	}
```

**After:**
```go
	// Archetype-aware default: a recognised page type gets its archetype
	// shape rather than a generic text block that competes with structured
	// content. Unknown types keep the generic default.
	sections := defaultSectionsForPage(pageName, pageType)
	if sectionsRaw, ok := newPlan["sections"].([]interface{}); ok && len(sectionsRaw) > 0 {
		// ... use LLM-provided sections (now also run through the resolver,
		//     per the validate_components implementation doc) ...
	}
```

```go
// defaultSectionsForPage returns archetype-appropriate default sections.
// Falls back to a generic content shape for unrecognised pages.
func defaultSectionsForPage(pageName, pageType string) []string {
	key := strings.ToLower(strings.TrimSpace(pageName))
	switch {
	case key == "faq" || strings.Contains(key, "faq"):
		return []string{"hero", "faq", "call-to-action"}
	case key == "contact":
		return []string{"contact-hero", "contact-form", "contact-info"}
	case key == "pricing" || strings.Contains(key, "pricing"):
		return []string{"hero", "pricing", "faq", "call-to-action"}
	case key == "about":
		return []string{"hero-about", "about-content", "call-to-action"}
	default:
		return []string{"hero", "generic-text-block", "call-to-action"}
	}
}
```

Keep the map small and obvious; it is only a fallback for when the LLM
gives nothing. The prompt fixes (1, 2) are the primary control.

---

## SQL to apply the prompt edits

Prompt edits live in `agent_definitions.default_config -> workflow ->
steps -> <step> -> config -> prompt_template`. They are large strings;
editing in place with `jsonb_set` is error-prone for multi-point edits.
Safer pattern: read the current template, edit the full string in an
editor, write it back as a parameter.

```sql
-- 1. Dump the current template to inspect/edit (content-gap-planner)
SELECT default_config #>> '{workflow,steps,plan_gaps,config,prompt_template}'
FROM agent_definitions
WHERE type = 'content-gap-planner';
```

Edit that text (apply 1a/1b/1c), then write it back. Using psql, set it
from a file to avoid quoting issues:

```sql
-- 2. Write the edited template back (content-gap-planner)
--    Load the edited prompt from a file into :newprompt first, e.g.:
--      \set newprompt `cat gap_planner_prompt_edited.txt`
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,plan_gaps,config,prompt_template}',
      to_jsonb(:'newprompt'::text),
      false
    ),
    updated_at = NOW()
WHERE type = 'content-gap-planner'
RETURNING type, version, updated_at;
```

```sql
-- 3. Same pattern for site-planner (step is plan_site)
SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}'
FROM agent_definitions
WHERE type = 'site-planner';

-- ... edit (2a/2b/2c) into site_planner_prompt_edited.txt ...
-- \set newprompt `cat site_planner_prompt_edited.txt`
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,plan_site,config,prompt_template}',
      to_jsonb(:'newprompt'::text),
      false
    ),
    updated_at = NOW()
WHERE type = 'site-planner'
RETURNING type, version, updated_at;
```

`loadAgentDefinition` reads per-spawn with no cache, so edits take effect
on the next planner run. No chassis rebuild needed for the prompt fixes
(Fix 1, 2). Fix 3 and the `validate_components` implementation are Go and
DO need a chassis rebuild + redeploy.

## Verification

After the prompt edits, run a planner against a test brief that would
previously produce a FAQ page, and confirm:
- the faq page's sections are `["hero","faq","call-to-action"]` (no
  `generic-text-block`),
- no section name is a display name (`"FAQ Section"`),
- `call-to-action` uses the hyphen form.

```sql
-- Inspect a freshly planned site's faq page sections
SELECT jsonb_pretty(section)
FROM site_specs ss, jsonb_array_elements(ss.data #> '{pages}') section
WHERE ss.site_id = '<new-test-site>'
  AND ss.aspect = 'site_plan'
  AND section->>'name' = 'faq';
```

The same isolated-build harness used for the writer test confirms the
end-to-end result: a planned-then-built faq page comes out with a
populated, linked accordion.

## Summary of the full prevention set

| Fix | Type | Defect | Rebuild? |
|---|---|---|---|
| 1. content-gap-planner prompt | prompt (SQL) | duplicate surface | no |
| 2. site-planner prompt | prompt (SQL) | duplicate surface + display-name | no |
| 3. applyNewPage default | Go | duplicate surface (backstop) | yes |
| validate_components impl | Go | display-name leak / orphan | yes |
| (later) per-section briefs | prompt + loader | disambiguate legit pairings | no |
| (later) post-build structured-field check | Go | empty structured component | yes |

---

# Fix B implementation — validate_components resolver (Defect 2: display-name leak)

Implements the dead `validate_components: true` flag in
`ValidateSitePlanAction`. Scope is deliberately narrow: **resolve each
section name to a real `content_components.function`, or drop it.** It does
NOT deduplicate or make content-intent decisions (see "Scope" below).

Reuses the existing `NormalizeComponentFunction` (~line 31067) and adds
only a display-name/name → function lookup, which normalisation alone
can't do (`"FAQ Section"` normalises to `faq-section`, still not the real
`faq`).

## Scope — what it does and does NOT do

**Does (deterministic, safe):**
- Resolves `call_to_action` → `call-to-action`, `FAQSection` →
  `faq-section` via existing normalisation.
- Resolves `"FAQ Section"` (display_name) → `faq` and any `name` → its
  `function` via DB lookup.
- Drops + logs a section name that resolves to nothing (it would orphan
  the page_component downstream).

**Does NOT (intent decisions it can't make safely):**
- Does NOT deduplicate or strip a `generic-text-block` sitting next to a
  `faq`. Validation cannot know whether that block is a legitimate intro
  or a redundant duplicate. Guessing risks deleting wanted content or
  keeping the empty-FAQ pairing. The duplicate-surface problem is solved
  at the planner prompt (don't emit the bad pairing) and by per-section
  briefs (disambiguate legitimate pairings) — NOT here.

This keeps the action's behaviour predictable: a section name either
resolves to a valid component or is removed for being unresolvable. No
silent intent-based deletions.

## Code

### 1. Helper: load the resolution maps

Add near `loadSiteChromeNames` (same file). One query, three maps.

```go
// componentNameResolver holds lookup maps for resolving section names in a
// site plan to canonical content_components.function values.
type componentNameResolver struct {
	validFunctions map[string]bool   // function -> true
	displayToFunc  map[string]string // lower(display_name) -> function
	nameToFunc     map[string]string // lower(name) -> function
}

// loadComponentNameResolver loads section/element component identity from
// the DB so plan section names can be resolved to a canonical function.
// Returns an empty resolver (not nil) on error so callers can no-op safely.
func loadComponentNameResolver(ctx context.Context, db *sql.DB, logger *zap.Logger) *componentNameResolver {
	r := &componentNameResolver{
		validFunctions: make(map[string]bool),
		displayToFunc:  make(map[string]string),
		nameToFunc:     make(map[string]string),
	}
	if db == nil {
		return r
	}
	rows, err := db.QueryContext(ctx,
		`SELECT "function", name, COALESCE(display_name, '')
		   FROM content_components
		  WHERE component_level IN ('section','element')
		    AND is_active = true
		    AND "function" <> ''`)
	if err != nil {
		logger.Warn("loadComponentNameResolver: query failed", zap.Error(err))
		return r
	}
	defer rows.Close()
	for rows.Next() {
		var fn, name, display string
		if err := rows.Scan(&fn, &name, &display); err != nil {
			continue
		}
		r.validFunctions[fn] = true
		if name != "" {
			r.nameToFunc[strings.ToLower(name)] = fn
		}
		if display != "" {
			r.displayToFunc[strings.ToLower(display)] = fn
		}
	}
	return r
}

// resolve attempts to map a raw section name to a canonical component
// function. Returns (function, true) if resolved, ("", false) if not.
func (r *componentNameResolver) resolve(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	// 1. Already a valid function.
	if r.validFunctions[raw] {
		return raw, true
	}
	// 2. Normalise (underscore->hyphen, camelCase->kebab) and re-check.
	norm := NormalizeComponentFunction(raw)
	if norm != raw && r.validFunctions[norm] {
		return norm, true
	}
	// 3. Display-name lookup (handles "FAQ Section" -> "faq").
	if fn, ok := r.displayToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 4. Component name lookup (handles a row's `name` differing from `function`).
	if fn, ok := r.nameToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 5. Last try: display lookup on the normalised form.
	if fn, ok := r.displayToFunc[strings.ToLower(norm)]; ok {
		return fn, true
	}
	return "", false
}
```

### 2. The validation block in `ValidateSitePlanAction`

Insert AFTER the existing site-chrome-strip `if params.DB != nil { ... }`
block and BEFORE the final
`params.Logger.Info("ValidateSitePlanAction: Complete", ...)` /
`return plan, nil`.

Gated on the existing `validate_components` config flag (currently set
true for site-planner but never read).

```go
	// ── Resolve section names to canonical component functions ───────────
	// Implements config flag `validate_components`. Each section name must
	// map to a real content_components.function. Names that are display
	// names ("FAQ Section"), wrong-case, or underscore variants are
	// normalised/resolved; unresolvable names are dropped and logged.
	// This does NOT deduplicate or make content-intent decisions — it only
	// guarantees every surviving section name is a valid component function.
	validateComponents := false
	if vc, ok := config["validate_components"].(bool); ok {
		validateComponents = vc
	}
	if validateComponents && params.DB != nil {
		resolver := loadComponentNameResolver(ctx, params.DB, params.Logger)
		if len(resolver.validFunctions) > 0 { // only act if we actually loaded components
			for _, p := range pages {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				sectionsRaw, ok := pm["sections"].([]interface{})
				if !ok {
					continue
				}
				resolved := make([]interface{}, 0, len(sectionsRaw))
				for _, s := range sectionsRaw {
					name, ok := s.(string)
					if !ok {
						// Non-string (e.g. a section-brief object) — leave as-is.
						resolved = append(resolved, s)
						continue
					}
					fn, ok := resolver.resolve(name)
					if !ok {
						params.Logger.Warn("ValidateSitePlanAction: dropped unresolvable section name",
							zap.Any("page", pm["name"]),
							zap.String("section", name))
						continue
					}
					if fn != name {
						params.Logger.Info("ValidateSitePlanAction: resolved section name to function",
							zap.Any("page", pm["name"]),
							zap.String("from", name),
							zap.String("to", fn))
					}
					resolved = append(resolved, fn)
				}
				pm["sections"] = resolved
			}
		} else {
			params.Logger.Warn("ValidateSitePlanAction: validate_components set but no components loaded — skipping name resolution")
		}
	}
```

Notes on the design:
- **Non-string entries pass through untouched.** When per-section briefs
  land (objects, not strings), this loop won't mangle them — it only
  resolves string entries. Forward-compatible with the briefs work.
- **Empty-resolver guard.** If the DB load returns nothing (query error,
  empty table), it skips rather than dropping every section. Fail-safe:
  better to pass an unvalidated plan than to empty every page.
- **Drop, don't substitute.** An unresolvable name is removed, not
  replaced with a guess. A wrong substitution would be worse than a
  missing section (which downstream gap-detection can re-surface).

## 3. The gap-planner path

`content-gap-planner` applies via `apply_gap_plan` → `applyNewPage`, which
does NOT route through `validate_site_plan`. So the resolver must also run
there, or the `"faq-section"`-style names from that planner's prompt slip
through.

In `applyNewPage` (`apply_gap_plan_action.go`), where it reads
`newPlan["sections"]` and falls back to the default, resolve each name
before writing the page record:

```go
	sections := []string{"hero", "generic-text-block", "call-to-action"}
	if sectionsRaw, ok := newPlan["sections"].([]interface{}); ok && len(sectionsRaw) > 0 {
		raw := make([]string, 0, len(sectionsRaw))
		for _, s := range sectionsRaw {
			if str, ok := s.(string); ok {
				raw = append(raw, str)
			}
		}
		// NEW: resolve names to canonical functions, drop unresolvable.
		resolver := loadComponentNameResolver(ctx, params.DB, logger)
		if len(resolver.validFunctions) > 0 {
			resolved := make([]string, 0, len(raw))
			for _, name := range raw {
				if fn, ok := resolver.resolve(name); ok {
					resolved = append(resolved, fn)
				} else {
					logger.Warn("applyNewPage: dropped unresolvable section name",
						zap.String("page", pageName), zap.String("section", name))
				}
			}
			if len(resolved) > 0 {
				sections = resolved
			}
		} else if len(raw) > 0 {
			sections = raw // resolver unavailable — use as-is rather than lose them
		}
	}
```

(If `applyNewPage` doesn't currently have `ctx`/`params.DB`/`logger` in
scope, thread them in — they're available on the action's `params`.)

## What this fixes and what it doesn't

Fixes (deterministically):
- `"FAQ Section"` → `faq` (the display-name leak that orphaned the
  component on `containment-first-architecture`).
- `call_to_action` ↔ `call-to-action` inconsistency across sites.
- Any future typo'd/display-named section that would otherwise orphan.

Does NOT fix (by design — handled elsewhere):
- The `generic-text-block` + `faq` duplicate pairing → planner prompt edit
  (Defect 1) + per-section briefs.
- Empty structured components when a pairing does occur → post-build
  validation (Fix D) catches it before deploy.

## Testing it

After deploying the chassis with this change, the existing isolated-build
harness covers it: trigger a build whose plan contains a deliberately
display-named section (e.g. inject `"FAQ Section"` into a test page's
sections) and confirm the resulting `page_components.component_id` is
linked (resolved to `faq`) rather than NULL (orphaned). The same
faq-test page pattern used for the writer test works here — only the
input section name changes.

---

# Fix C + stale-plan — planner depth and freshness (preventing the class)

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
