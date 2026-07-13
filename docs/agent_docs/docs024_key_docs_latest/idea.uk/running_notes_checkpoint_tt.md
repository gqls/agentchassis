# Running notes — checkpoint (tt): trigger path, rerender finding, plan-freshness, hardening

Date 2026-06-21. Continues (ss). Append to `running_notes.md`. Concerns how to actually land
the differentiators / item-fields fix (artefacts from (ss): `plan_sections_action.go`,
`v3_site_actions.go`, `019_pcw_prompt_item_fields.sql` + `_down.sql`, and
`RUNBOOK_pcw_item_fields_fix.md`).

## Deploy topology (clarified)

Two separate deploy paths, previously conflated:
- **Code** (the Go change: `ItemFields` population + the reconciler) ships inside the
  `agent-chassis` Docker image — `docker.io/aqls/agent-chassis`, currently `image_tag
  v1.0.1066`. Specialist agents are spawned as job pods using the `image_tag` on their
  `agent_definitions` row. To deploy, build a new chassis tag and point `page-content-writer`
  (and any affected types) at it; confirm via running-pod image SHAs.
- **Generated site HTML** ships separately: sites git repo → GitHub Actions → Backblaze B2.
  This is what a rebuild/rerender produces downstream — not the code path.

**Prompt migration is already applied.** `page-content-writer.default_config` already renders
`item_fields` in both What To Write and Output Format (`updated_at 2026-06-20 19:00:54`). But
until a chassis image carrying the Go change is live, `{{if .item_fields}}` is always false
(nothing populates it) and the reconciler is absent — so the applied prompt is inert on its
own. Order: deploy the chassis image first, then trigger.

## Trigger mechanism (real, from the 081 scripts)

Triggers are Kafka `orchestrate` messages produced with `kcat` as a one-off pod in the `kafka`
namespace, to topic `system.agent.generic.requests`, bootstrap
`personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092`, headers
(`correlation_id`, `orchestration_id`, `request_id`, `message_id`, `message_type=request`,
`client_id`, `action=orchestrate`, `sender_agent_type=cli`, `sender_agent_id=cli-user`,
`responses_topic=system.agent.generic.responses`, `timestamp`), body
`{"action":"orchestrate","config":{"agent_type":"<T>"},"input_data":{...}}`. Monitor via
`orchestration_states` and `site_work_items`; logs via `-l app=agent-chassis` (single shared
chassis), scoped by correlation id.

Three relevant agent types:
- `page-rerender` — single page, re-assembles from existing components. **Does not rewrite
  content.**
- `rerender-pages` — whole site; inserts one `page_rerender` work item per deployed page;
  build-dispatch-loop processes them. Does not rewrite content.
- `page-rebuild` — **rewrites content**; regenerates pages flagged `needs_rebuild` by re-running
  page-content-writer (per the script's cost note). Input `{site_id, domain}`.

## Finding: page-rerender cannot fix the differentiators cards

Confirmed from `rerender_single_page_action.go`. `assemblePage` → `getPageSections` runs:

```
SELECT COALESCE(rendered_html,'') FROM page_components
WHERE page_id=$1 AND rendered_html IS NOT NULL AND rendered_html != '' ORDER BY position ASC
```

and concatenates the **stored `rendered_html`** of each section — file header: "Simple
concatenation - no template re-rendering". It never calls `render_component`, so the
reconciler is not on this path, and the differentiators row's stored `rendered_html` is the
broken empty-cards artefact (a rerender re-ships it). The reconciler cannot help at this stage
regardless: the items are already baked into HTML as `<h3></h3>` — there is no structured data
left to repair. Also note `sectionHasVisibleContent` keeps the differentiators section (its
heading exceeds the 10-char visible-text threshold), so it ships broken rather than being
filtered out.

→ **`page-rerender` is the wrong tool for this fix.** Only `page-rebuild` regenerates via
page-content-writer (`generate_content` → `render_section` → `RenderComponentAction`) and so
exercises both halves of the fix.

## Finding: plan-freshness dependency (gates whether the fix fires)

Both halves of the (ss) fix read `item_fields` off `current_section.llm_field_specs`: the
prompt's `{{if .item_fields}}` and the reconciler's expected keys. `llm_field_specs` is built
by `plan_sections` (`buildSectionPlanItem`). But **page-content-writer receives the plan as
input** (`select_sections` ← `input_data.section_plan.sections_ready`) — it does not build it.
The `pages` table carries `page_spec` and `built_from_plan_version`, which suggests section
plans may be **stored and replayed**. If `page-rebuild` replays a plan built before this
change, `llm_field_specs` lacks `item_fields` → the prompt renders as before *and* the
reconciler has no expected keys → the rebuild regenerates `title`/`body` again and the cards
stay empty.

Outstanding: we don't yet have the `page-rebuild` workflow / `load_rebuild_context`, so whether
it rebuilds the plan (fix fires) or replays a stored one (stale) is unconfirmed. Backstop is
empirical — after a rebuild, check `content_data` keys; `title`/`body` still present ⇒ stale
plan. Runbook §3 leads with a query to dump the `page-rebuild` definition and look for a
`plan_sections` step vs a stored-plan replay.

## Finding: Component.InputSchema type confirmed → hardening is feasible

`component_library.go`: `Component.InputSchema` is `map[string]interface{}` (line 37),
populated by `json.Unmarshal(schemaJSON, &comp.InputSchema)` (line 250) in the loaders used by
`GetComponentWithFallback`. Its shape is `{"fields": {...}}` — identical to what `plan_sections`
already walks (`comp.InputSchema["fields"].(map[string]interface{})`). And
`RenderComponentAction` reloads the component fresh every render
(`GetComponentWithFallback`, v3 line 1473), so `comp.InputSchema` is always current and
independent of plan state.

## Choices

- **Route for this fix: `page-rebuild`, index only.** Flag only the index `needs_rebuild`
  (guard-check that no other page is already flagged so the site-wide rebuild doesn't sweep
  others), then trigger `page-rebuild` with `{site_id, domain}`. `page-rerender` rejected for
  this purpose (see above). Cost: regenerates *all* index sections (hero/FAQ/method copy
  rewritten too).
- **Recommended hardening (now unblocked): make the reconciler plan-independent.** Source the
  reconciler's expected keys from `comp.InputSchema` (walk `["fields"][*].items` /
  `item_schema`, reusing `extractArrayItemFields`) instead of `current_section.llm_field_specs`.
  This removes the plan-freshness dependency entirely — the safety net then fires on any
  regeneration regardless of plan or prompt state, and the prompt change becomes an
  optimisation (fewer remaps) rather than a correctness dependency. Type-safe per the finding
  above. **Not yet implemented — awaiting go-ahead.**
- **Still open from (ss): reconciler fatal vs non-fatal.** Currently ERROR-and-continue; the
  alternative is to hard-fail the section. Unchanged; pending a call.

## Outstanding to confirm / obtain

- `page-rebuild` workflow + `load_rebuild_context`: does it rebuild the section plan or replay
  a stored one? Determines whether the prompt/reconciler fire under the current (plan-coupled)
  design; the hardening above makes this moot for correctness.
- Decision to implement the schema-sourced reconciler (and the fatal/non-fatal choice).
