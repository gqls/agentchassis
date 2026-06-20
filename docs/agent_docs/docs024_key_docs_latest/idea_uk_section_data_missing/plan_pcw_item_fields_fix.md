# Plan — apply and verify the differentiators / item-fields fix

Goal: differentiators (and any array component) renders populated cards, with a render-time
safety net catching future LLM key drift. Artefacts: `plan_sections_action.go`,
`v3_site_actions.go`, `019_pcw_prompt_item_fields.sql` (all in outputs).

The Go deploy and the migration are independent — neither creates a broken intermediate
state — but content generated *between* them uses the old behaviour, so keep the gap short.

## 1. Repo / deploy

1. Drop in the two patched files.
2. `gofmt -w` both, then `go build` the chassis. Resolve anything gofmt/build raises
   (notably the new struct-tag alignment, left compilable but not column-aligned).
3. Commit; deploy GitHub → GH Actions → Backblaze B2 (new chassis image tag).
4. Confirm the running image is the new one: `kubectl -n ai-persona-system get pods` and
   check the deployed `image_tag`.

## 2. Database

1. Run `019_pcw_prompt_item_fields.sql` (renumber if 019 is taken). Expect NOTICE
   "prompt patched".
2. Verify both markers are present:
   ```sql
   SELECT
     position('{{if .item_fields}}'   IN (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}')) AS wtw_marker,
     position('{{if $f.item_fields}}' IN (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}')) AS out_marker
   FROM agent_definitions WHERE type = 'page-content-writer';
   ```
   Both `> 0`.
3. Optional: read the `prompt_template` back and eyeball the What To Write + Output Format
   blocks.

## 3. Regenerate idea.uk index

1. Re-run page-content-writer for site `97ed2f64-65ca-4b67-8a98-dfd8195a0d3a`, `/index.html`.
   The stored `content_data` still holds `title`/`body`; regeneration produces
   `name`/`description` under the new prompt (the reconciler is belt-and-braces). A plain
   re-render would also be repaired by the reconciler, but regenerating exercises the prompt
   fix too.
2. Confirm in DB: the differentiators `page_components.content_data` now has
   `features[].name` / `features[].description` (not `title`/`body`).
3. Confirm in the deployed HTML on B2: the seven `.differentiator-item` cards have non-empty
   `<h3>`/`<p>`.

## 4. Watch the logs (tells us whether the prompt change alone sufficed)

- `reconcileGeneratedItemKeys` WARN ("remapped"/"normalised") — the model still drifted but
  we caught it → prompt steering imperfect, safety net earned its place.
- `reconcileGeneratedItemKeys` ERROR ("unrecoverable") — an item field neither matched nor
  had a synonym → investigate that component/field.
- No reconcile logs for differentiators — the prompt change alone produced the right keys.

(No `logger.Debug` used anywhere in the change, so these surface.)

## 5. Decide non-fatal vs fatal

Keep ERROR-and-continue, or switch the `!remapped` branch to a returned error + caller
propagation so incomplete sections fail the build. Pending a call.

## Follow-on threads (not part of this fix)

- **services-grid** — byte-identical schema; benefits automatically. Verify whenever it is
  first used on a built page.
- **info-card-grid** — stored `html_template` contains literal `<no value>`; repair the
  template (separate task). The `item_schema` path is already handled by
  `extractArrayItemFields` once the template is fixed.
- **idea.uk other parked gaps** (from the handoff, main-chat items): empty hero + CTA buttons
  (no destination pages), contact form posting to `#contact`, thin nav/footer. Revisit after
  the differentiators fix lands.
