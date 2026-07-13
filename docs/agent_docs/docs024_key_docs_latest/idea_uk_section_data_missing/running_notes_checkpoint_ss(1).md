# Running notes — checkpoint (ss): idea.uk differentiators empty cards — root cause and fix

Date 2026-06-20. Follows (qq)/(rr). Append to `running_notes.md`. Continues the
investigation from `HANDOFF_idea_uk_differentiators_section_data.md`. Artefacts produced
this session are in outputs: `plan_sections_action.go`, `v3_site_actions.go`,
`019_pcw_prompt_item_fields.sql`.

## Symptom (confirmed)

idea.uk index (`/index.html`, site_id `97ed2f64-65ca-4b67-8a98-dfd8195a0d3a`) renders the
differentiators heading plus seven empty `<div class="differentiator-item">` cards — every
`<h3>`/`<p>` blank. The method narrative and 13-item FAQ on the same page populate.

## What we established (not what the handoff assumed)

- **The content is generated and stored.** `page_components.content_data` for the
  differentiators row holds `headline` + a 7-element `features` array, each item keyed
  **`title`/`body`**, full editorial copy. So this was never a missing-data or
  reconciler-scope problem — `reconcile_section_data` is irrelevant to this fault.
- **The component:** `content_components(function='differentiators')` has
  `render_mode='template'`; `input_schema` fields `features` (array, `source: llm`, items
  `{name, description}`) and `headline` (text, `source: llm`). The `html_template` iterates
  `{{range .features}}` reading **`{{.name}}`/`{{.description}}`**.
- **The mismatch:** the LLM emitted `title`/`body`; the template reads `name`/`description`
  → every card renders empty. `headline` (a scalar) matched and rendered.
- **Routing:** the writer sub-workflow (`agent_definitions` page-content-writer,
  `process_sections_loop`) branches on
  `current_section.component.render_mode == 'agent' OR current_section.component.needs_llm == true`.
  `needs_llm` is computed by `detectNeedsLLMContent` (v3_site_actions.go ~4095), which
  returns true for any non-empty `input_schema`. So differentiators takes the LLM path via
  `needs_llm` regardless of `render_mode` — which is why it had generated content despite
  `render_mode='template'`, and why reverting render_mode (below) is safe.

## Why FAQ worked and differentiators didn't

The prompt's field list never told the model the array **item** field names. For a field
named `questions` the model's natural guess (`question`/`answer`) happens to match the faq
template; for `features`/cards the natural guess (`title`/`body`) does not match
`name`/`description`. FAQ worked by coincidence, not contract.

## Root cause in code

`plan_sections_action.go`, `buildSectionPlanItem`: it reads `fieldDef["items"]` (~line 1020)
but never carries it onto the `llmFieldSpec` it appends (~1029); the struct had no field for
it. So the per-section prompt — `## What To Write` (iterating
`current_section.llm_field_specs`) and the `## Output Format` skeleton — listed array fields
with **type only, never element shape**, and the model guessed the item keys.

- `services-grid` has a byte-identical schema → the same latent bug.
- `info-card-grid` uses `item_schema` (not `items`) **and** is separately broken: its stored
  `html_template` literally contains `<no value>` where tokens belong (looks like
  rendered-against-nil output written back into the template column). Its own thread — not
  fixed here.

Secondary finding: the `## Output Format` skeleton renders every field as a scalar
(`"features": "..."`) regardless of type. Misleading for arrays, but the model already
overrides it. The fix makes the skeleton type-aware too.

## Correction logged

`differentiators.render_mode` had been changed to `'agent'` during investigation; reverted
to its original `'template'`
(`UPDATE content_components SET render_mode='template' WHERE function='differentiators' AND is_active=true`).
Confirmed harmless by `detectNeedsLLMContent` (non-empty schema → `needs_llm` true keeps it
on the LLM path either way).

## Fix delivered (three artefacts)

- **plan_sections_action.go** — add `sort` import; `ItemFields []string` on `llmFieldSpec`;
  new `extractArrayItemFields` helper (reads `items` and `item_schema`, sorted for stable
  prompts/specs); populate `ItemFields` in the `source=="llm"` block. No existing names
  changed; `fieldItems`/`fieldMinItems` still serve the HITL branch.
- **v3_site_actions.go** — render-time reconciler appended (`itemKeySynonyms`,
  `synonymsFor`, `normaliseKeyForMatch`, `expectedItemFieldsFromSpecs`,
  `reconcileGeneratedItemKeys`) + 6-line wire-in in `RenderComponentAction` running it on
  `contentData` before the merge (so corrected keys land in both rendered HTML and persisted
  `content_data`). Case/separator-insensitive synonym remap (`title`/`body` →
  `name`/`description` etc.); WARN on remap, ERROR + continue on unrecoverable (non-fatal by
  default). No new imports.
- **019_pcw_prompt_item_fields.sql** — idempotent migration patching the page-content-writer
  prompt to render `item_fields` in What To Write and the Output Format skeleton. `replace()`
  on the two exact fragments (verified present at positions 2330 and 3402 via a `position()`
  check); guards abort if the path is missing or the fragments have moved, and skip if
  already applied.

## Verification done this session

`position()` fragment check (2330/3402); brace/paren balance on both patched files; regions
inspected. **Not compiled** — no Go toolchain in the working environment; `gofmt` + build
pending on the repo side (gofmt will also align the new struct tag).

## Open decision

Unrecoverable missing item keys currently ERROR-and-continue (non-fatal). Alternative: hard-
fail the section so nothing incomplete deploys — one-line change in
`reconcileGeneratedItemKeys` (return an error from the `!remapped` branch) plus caller
propagation. Pending a call.
