# Step 2 verification — observations

Date: 2026-05-12
Site: gamesdesign.co.uk (`9a8baddf-77c2-4486-a56b-1b7dde9c1e9e`)
Build correlation_id: `344ef66f-646f-41c9-ae5b-beaa34f8617c` (build-dispatch run)
Adoption correlation_id: `f0ecba7e-df56-44eb-bb93-2ad2d846bbbc` (separate)
Verified against orchestration: `daeb3542-628e-4fed-8491-ea62f5d67178`
  (page-build-handler EXECUTING_STEP at save_sections — section_plan fully populated)

## Step 2 status: PASS

`sectionPlanItem.Component` is being populated end-to-end.

Counts on `section_plan.sections_ready` for the verified run:
- ready_count: 2
- with_component: 2/2
- with_html_template: 2/2
- with_input_schema: 2/2
- with_resolved_data: 1/2 (hero had fallback values; the other section is all-LLM)
- with_llm_fields: 2/2

Sample (hero section): `component` block carries `component_id`, `name`,
`function`, `display_name`, `category`, `description`, `render_mode`,
`component_level`, `needs_llm`, `html_template` (full template inline),
`input_schema` (as JSON-encoded string). Alongside: `resolved_data` populated
with the `use_fallback` values (`cta_url=/contact.html`,
`background_image=/assets/images/hero.jpg`, `secondary_cta_url=/services.html`)
and `llm_fields=["headline","subheadline","secondary_cta","cta_text"]`.

The shared `loadSectionComponents` helper is working; the new `Component`
field is reaching `orchestration_states.collected_data` correctly; both
existing callers (LoadPageSectionComponentsAction and plan_sections) are
producing matching shapes.

## Observations to carry into Step 3

### `input_schema` is a JSON-encoded string, not parsed JSON
Intentional in Step 2 to preserve LoadPageSectionComponentsAction's existing
behaviour (its prompt template dumps it as a string). For Step 3's targeted
prompt — where we want per-field type info ("headline is type=text,
required=true") — we'll need to parse it. Cleanest options:
  - Expose a parsed `field_specs` array alongside `llm_fields` from
    plan_sections (single source, no client-side parsing).
  - Or parse client-side in a small pre-LLM step.
  Decision deferred to Step 3 seam verification.

### llm_fields is ready to drive the targeted prompt as-is
For hero on tool-ttk-calculator (the verified run), llm_fields is
`["headline","subheadline","secondary_cta","cta_text"]`. The Step 3
targeted prompt can ask the LLM for exactly these four keys instead of
dumping the full schema with six fallback examples. This is the win.

### resolved_data here is fallback-driven, not query-driven
The hero section's resolved_data is filled from `use_fallback` defaults
(e.g. `cta_url=/contact.html`). To verify the Tier D queryresolve path
(which is the fabrication-fix concern), we should look at sections with
`source: query.*` in their schema — `guide-list` on guides-index or
`tool-list` on tools page. Those weren't in this particular orchestration's
section_plan; eyeball them in `926683b1` (guides-index) or one of the tools
builds before declaring Step 3's reasoning sound.

## Pre-existing failures (not caused by Step 2)

Two page-content-writer orchestrations failed at
`process_sections_loop_iter_1_generate_content` before Step 2's verified
run:
- `fe87d5f2-18ed-4ff5-bd31-b52f621a4541` (index, 19:46:13)
- `94ef94e6-4ca8-40c9-912a-908da9c4e162` (tools, 19:51:04)

Root cause: Anthropic API `status 529: overloaded_error` after 4 retries.
Not structural. Same code path succeeded from 19:55 onwards
(`926683b1`, `97c57f46`, `daeb3542`).

Implications:
- Pages `index` and `tools` are still `build_status='planned'` despite
  their work_items showing `complete` — the work_item completion does
  not reflect that page-content-writer actually succeeded. Worth a
  separate look at how work_item status is set when the writer's
  orchestration FAILs.
- The 4-retry budget on `execute_llm_prompt` doesn't ride through
  sustained API overload. Backoff/retry tuning is a separate concern.
- Both pages need a manual or scheduled rebuild before this site is
  fully deployed.

## Schema quirks noted

- `orchestration_states.workflow_states` doesn't exist — the table is
  `orchestration_states` keyed by `orchestration_id` (uuid).
- `agent_error_log.orchestration_id` is `text`, not `uuid`. Casts needed
  for joins against `orchestration_states.orchestration_id::text`.
- `build_queue` has no `site_id` column — it's keyed by domain with its
  own `id` uuid.
- `page_components` has no `page_name` column — join to `pages` for the
  human-readable name. Foreign key is `page_components.page_id → pages.id`.
- `pages.build_status` distinct from `pages.status` — `status='active'`
  means the page row exists; `build_status='deployed'` means content
  rendered and pushed. `build_status='planned'` means awaiting build.

## What Step 3 changes (recap)

Workflow definition: page-content-writer
- Remove `load_page_components` step (use `section_plan.sections_ready`
  directly).
- Change `process_sections_loop.iterate_over` from
  `section_components.components` to `input_data.section_plan.sections_ready`.
- Targeted prompt driven by `current_section.llm_fields` instead of
  dumping `current_section.input_schema`.
- Add `merge_with` config on `render_section` so `resolved_data` lands in
  the rendered HTML and the persisted `content_data`.

Go action: RenderComponentAction
- Honour new `merge_with` config: extract resolved_data from the named
  path, merge into `sectionContentData` before render context merge.
- Order: LLM output ∪ resolved_data (resolved_data wins on conflicts since
  it's authoritative).

Verification surface for Step 3 success:
- Tool-list and guide-list sections — items array populated from real
  `pages` rows, not LLM fabrication.
- LLM output JSON contains only the llm_fields, not the full schema.
- Persisted content_data carries the merged document.
