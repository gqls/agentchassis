# Step 3 — targeted prompt, section_plan loop, merge_with

## Scope

Step 3 closes the fabrication gap surfaced in the Step 2 observations.
The page-content-writer no longer dumps the full input_schema and asks
the LLM to invent the shape; it sends only the fields the schema declares
as `source: llm`, and the resolved_data (queryresolve results + static
fallbacks) lands directly in both rendered HTML and persisted content_data.

Three sub-changes, ordered for incremental deploy/verification:

| Sub-change | What it does | Risk |
|---|---|---|
| 3a | plan_sections produces `llm_field_specs` alongside `llm_fields`. Additive — nothing consumes it yet. | Negligible |
| 3b | page-content-writer workflow rewrite: drop `load_page_components`, iterate over `section_plan.sections_ready`, targeted prompt driven by `llm_field_specs`, `merge_with` on the render steps. | Moderate — change is substantive, snapshot required |
| 3c | RenderComponentAction honours `merge_with` config. Additive — config key is optional; existing callers without `merge_with` unaffected. | Low |

3a and 3c are additive. 3b is the substantive change.

## Sub-change 3a: llm_field_specs

File: `platform/orchestration/actions/plan_sections_action.go`
Diff: `step3a_plan_sections_action.diff` (89 lines)

Adds an `llmFieldSpec` struct and a `LLMFieldSpecs []llmFieldSpec` field
on `sectionPlanItem`. Populated in the same loop that builds `LLMFields`,
reading from the parsed input_schema field's `llm_guidance` key (verified
against production schemas — that's the actual field name, not `description`).

Each spec carries:
- `name` — field name
- `type` — text | url | image | rich_text etc.
- `required` — bool
- `description` — from `llm_guidance` in the schema
- `on_missing` — skip_field | use_fallback | error
- `fallback` — value when on_missing=use_fallback

Verification: gofmt clean.

## Sub-change 3b: workflow rewrite

File: `agent_definitions` row, type='page-content-writer', `default_config->'workflow'`
SQL: `step3b_apply.sql` (18,970 bytes — wraps the new workflow with snapshot-check guard and post-update verification)
Standalone workflow: `step3b_workflow.json` (for review)
Standalone prompt template: `step3b_prompt_template.txt` (for review)

### Structural changes

1. **Removed `load_page_components` step.** Its job (loading rich component
   data per section) is now done upstream by `plan_sections` and arrives
   on `section_plan.sections_ready[*].component`. The page-content-writer
   loop reads it directly instead of redundantly re-loading.

2. **`prepare_link_context.next_step`** changed from `load_page_components`
   to `build_render_context` (closing the gap left by step 1).

3. **`process_sections_loop.iterate_over`** changed from
   `section_components.components` to `input_data.section_plan.sections_ready`.

4. **`render_section.component_from` and `render_from_template.component_from`**
   changed from `current_section` to `current_section.component`. The
   nested component map carries html_template, render_mode etc. that
   render_component needs.

5. **`render_section.merge_with`** added: `current_section.resolved_data`.
   Same on `render_from_template` so the no-LLM path also gets merged data.

6. **Conditionals updated**: `check_render_mode` and `check_needs_research`
   now read `current_section.component.{render_mode,needs_llm,needs_research}`.

### Prompt template rewrite

The previous prompt dumped `{{.current_section.input_schema}}` as a wall
of text plus ~100 lines of fallback JSON examples (hero/feature/CTA/text/
contact/testimonial/case-study shapes). The LLM had to pick the right
shape from the examples and fill it in — leaving wide room for invented
items, urls, and labels.

The new prompt iterates `current_section.llm_field_specs` and shows only
the fields the schema declares as `source: llm`, each with its type,
required flag, and `llm_guidance` from the component definition. The
output format renders an example JSON object from the same field list,
so the LLM sees `{ "section_heading": "<text>", "section_intro": "<text>", ... }`
and the absence of `items`, `cta_url`, `eyebrow_label`, etc. from the
example IS the boundary.

Per the prompting principle discussed: the prompt does NOT list field
names that should NOT be produced (avoiding "don't think of a pink
elephant"). It states once that "the system separately handles lists,
URLs, images, labels, and any database-resolved data for this section"
— categorical, no specific names — and the JSON example shape is the
concrete constraint.

All upstream context blocks retained: Company Context, Contact Info,
Internal Linking, Content Direction, Rewrite Guidance, Admin Brief,
Research Findings, Existing Content for Recreate mode.

All 17 non-fabrication strict rules retained in full (LLMs sometimes
forget the rules, and we want them in the message every time).

### Apply procedure

```sql
-- 1. Snapshot (mandatory — the apply script verifies this)
SELECT snapshot_agent('page-content-writer');
SELECT * FROM agent_snapshots WHERE type = 'page-content-writer';

-- 2. Apply (the script aborts if no snapshot exists)
\i /path/to/step3b_apply.sql

-- 3. If the rolling chassis caches definitions
kubectl -n ai-persona-system rollout restart deployment/agent-chassis
```

### Revert procedure

```sql
SELECT revert_agent('page-content-writer');
```

The 3c Go change does NOT need reverting — the new `merge_with` config
is optional; without it, the old workflow's render_component calls (no
merge_with key) behave exactly as before.

## Sub-change 3c: RenderComponentAction honours merge_with

File: `platform/orchestration/actions/v3_site_actions.go`
Diff: `step3c_v3_site_actions.diff` (43 lines)

Adds a `merge_with` config handler immediately after the existing
`content_from` extraction block. The handler:

1. Extracts the merge map via `datahelpers.ExtractNestedField` at the
   configured path (e.g. `current_section.resolved_data`).
2. Overlays the merge map onto `sectionContentData` — the field that
   becomes the persisted `content_data` in the action's output.
3. Also merges into `renderCtx` so the resolved values appear in the
   rendered HTML template alongside LLM output.

Order is critical: `content_from` extraction first (LLM output captured),
then `merge_with` overlay (resolved_data wins on conflicts). This is the
authoritative-data semantics — queryresolve output and static fallbacks
should never be overridden by LLM invention.

Verification: gofmt clean.

## Verification plan (post-deploy)

### Confirm the wire is intact

```sql
-- Find the next page-build-handler run with a tool-list or guide-list section
WITH recent AS (
  SELECT orchestration_id, created_at
  FROM orchestration_states
  WHERE site_id = '9a8baddf-77c2-4486-a56b-1b7dde9c1e9e'
    AND owner_agent_type = 'page-build-handler'
    AND created_at > '<deploy-timestamp>'
  ORDER BY created_at DESC LIMIT 5
)
SELECT
  r.orchestration_id,
  s->>'name' AS section,
  s ? 'llm_field_specs' AS has_specs,
  jsonb_array_length(COALESCE(s->'llm_field_specs','[]'::jsonb)) AS spec_count
FROM recent r
JOIN orchestration_states o ON o.orchestration_id = r.orchestration_id,
     jsonb_array_elements(o.collected_data->'section_plan'->'sections_ready') AS s
WHERE s->>'name' IN ('tool-list','guide-list');
```

Pass: `has_specs=true`, `spec_count` matches the count of `source: llm`
fields in the component schema.

### Confirm content_data carries resolved items

```sql
SELECT
  p.name AS page_name,
  pc.slot_name,
  jsonb_array_length(COALESCE(pc.content_data->'items','[]'::jsonb)) AS items_count,
  pc.content_data->'items'->0 AS first_item
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '9a8baddf-77c2-4486-a56b-1b7dde9c1e9e'
  AND pc.slot_name IN ('tool-list','guide-list')
ORDER BY p.name, pc.position;
```

Pass: `items_count = 6` (or whatever the resolver returned), `first_item`
contains real `pages` row data (`title`, `url`, `meta_description`), NOT
LLM-fabricated content like "Calculator A" with `/tools/calc-a.html`.

### Confirm LLM output is small

In the agent_chassis logs for a page-content-writer run, look for the
LLM response from `generate_content`. It should contain ONLY the keys
listed in `llm_field_specs` for that section — typically 3-5 short keys
— not a sprawling JSON with items arrays.

## Known follow-ups (not addressed by Step 3)

1. **`cta_url` silently dropped** when site_specs path returns nothing and
   no fallback is declared. Template renders empty href. Schema authors
   should declare fallbacks for required URLs.

2. **Tier B fields with `llm_guidance`** (e.g. `cta_label` saying "Override
   if site tone differs") currently always use the static fallback. The
   schema author's intent suggests a "soft static" classification where
   the LLM could override when context warrants. Out of scope for Step 3.

3. **work_item completion lag**: when page-content-writer FAILs (e.g. API
   529), the work_item still gets marked complete eventually with
   "Auto-completed: work verified done despite lost response". This
   creates the index/tools `build_status='planned'` situation from the
   Step 2 observations. Separate bug from this change.

4. **Retry budget on LLM overload**: 4 retries don't ride through
   sustained 529s. Worth tuning separately.
