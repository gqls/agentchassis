# Step 2 — plan_sections enrichment, shared section loader

## Scope

Step 2 of the A1c sequence (resolve queryresolve before the LLM call, route
pre-resolved items into the prompt). This change is **additive only** — no
existing call site changes behaviour, and no workflow definition is touched.
Step 3 will consume the new `sectionPlanItem.Component` field by switching
`page-content-writer` to iterate over `section_plan.sections_ready` and
removing the redundant `load_page_components` step.

## Files

- `v3_site_actions.go` — adds shared `loadSectionComponents` helper, refactors
  `LoadPageSectionComponentsAction` to a thin wrapper.
- `plan_sections_action.go` — adds `sectionPlanItem.Component` and
  `componentInfo.Raw`, rewrites `loadComponentSchemas` and
  `loadSingleComponentSchema` over the shared helper, adds
  `sectionTemplateValid`, attaches `comp.Raw` onto plan items in `planSection`.

## New shared helper

`loadSectionComponents(ctx, db, sectionNames, pageID, activeOnly, logger) []map[string]interface{}`

Single SQL path for the by-name → by-function fallback pattern. Behaviour
exactly matches what `LoadPageSectionComponentsAction` had before, plus an
`activeOnly` flag for callers that want `is_active = true` filtering.

Support functions extracted:
- `scanSectionComponentRow` — turns a `*sql.Rows` into the component map shape
- `buildStubSectionComponents` — stub fallbacks for the no-DB code path
- `enrichSectionComponentsWithBriefs` — `content_brief` attachment from
  `page_components`

## sectionPlanItem.Component

Carries the full per-section component map produced by the shared helper.
Populated for Paths 1 (direct lookup) and 2 (selector). Nil for Path 3
(not-found / selector unavailable) — `omitempty` on the JSON tag keeps the
serialised plan tidy.

Step 3 reads `Component.input_schema`, `Component.html_template`,
`Component.render_mode`, `Component.description`, `Component.category`,
`Component.content_brief`, etc. directly from this field instead of
re-loading via `LoadPageSectionComponentsAction`.

## activeOnly: explicit behaviour preservation

The two callers historically had different filter behaviour:

- `plan_sections.loadComponentSchemas` had `WHERE is_active = true` inline.
- `LoadPageSectionComponentsAction` had no `is_active` filter.

Rather than choose for them, the shared helper takes `activeOnly bool`:
- `plan_sections` passes `true` — preserves the original filter
- `LoadPageSectionComponentsAction` passes `false` — preserves its behaviour

Whether `LoadPageSectionComponentsAction` *should* filter `is_active = true`
is a separate question for a later step — inactive components are deactivated
for a reason and page-content-writer using one is probably a regression
waiting to happen, but changing that here would bundle behaviour change into
a refactor and would breach the additive-only property of Step 2.

## Behaviour preserved exactly

- by-name lookup, function fallback (DISTINCT ON ... ORDER BY function, created_at DESC)
- order preservation against input sectionNames
- stub generation for unfound sections
- content_brief enrichment when pageID provided
- template-truncation guard (mirrored in `sectionTemplateValid`):
  null/empty/short templates pass; long-with-no-closing-section fails

## Behaviour changes (intentional, non-breaking)

- `LoadPageSectionComponentsAction` no longer returns `db_error` on the
  name-query failure path. Searched the codebase — no consumer reads it. On
  query failure the helper now logs and falls through to the function-lookup
  pass and then to stubs, which is a strictly better degradation path.
- `LoadPageSectionComponentsAction`'s old query-failure early-return put
  raw `sectionNames []string` into the `components` field, which downstream
  consumers expecting `[]map[string]interface{}` couldn't have handled
  correctly. New behaviour: stub maps, consistent shape.

## Verification done

- `gofmt -e` clean on both modified files.
- Brace count balanced on both files.
- All referenced identifiers (`detectNeedsLLMContent`, `containsString`,
  `NormalizeSectionNames`, `extractWithInputDataFallback`,
  `SelectComponentByType`, `IncrementUsageCount`) exist in the codebase.

## Not done in Step 2

- No workflow definition changes.
- Page-content-writer still uses the old `load_page_components` step.
- Validator (`compute_component_quality.go`, `recordValidationRejection`) is
  not touched. Step 3 may surface orphan-fields false positives when the LLM
  produces only `llm_fields` but validation runs against the full schema —
  flagged at the time, not pre-emptively.

## What Step 3 will do

1. Update `page-content-writer` agent definition:
   - Remove `load_page_components` step (its successors chain straight through)
   - Change `process_sections_loop.iterate_over` from
     `section_components.components` to `input_data.section_plan.sections_ready`
   - Replace the schema-dump prompt with a targeted "write these N fields" prompt
     driven by `current_section.llm_fields`
   - Add `merge_with` config on the `render_section` step's `render_component`
     action so resolved_data merges into the LLM output before render & persist
2. Extend `render_component` action to honour `merge_with`.
3. Re-adopt gamesdesign.co.uk and confirm tool-list and guide-list items
   come from real `pages` rows, not LLM fabrication.
