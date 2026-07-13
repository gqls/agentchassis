# HANDOFF — component regeneration silently empties dependent pages

For a fresh chat picking this up cold. It is a durable, cross-site platform bug, flagged but not yet fixed.
Read this, then generate the context bundle with `BUNDLE_component_regen_clobber.md`, then work the
`PLAN_component_regen_clobber.md` using `RUNBOOK_component_regen_clobber.md` for the SQL/ops.

---

## 1. Platform model (the minimum to orient)

Agents are rows in Postgres (`clients_db`, table `agent_definitions`) run as Kubernetes pods (namespace
`ai-persona-system`), talking over Kafka. A page is assembled from sections; each section on a page is a row
in `page_components` that points at a shared template in `content_components`. The same `content_components`
row is reused by many pages across many sites. Rendering binds a page's `content_data` into the component's
`html_template`. DB access (a human runs all SQL/kubectl):

```
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db
```

Two change types: editing an agent's `default_config` jsonb is a live workflow change (no redeploy); a Go
change goes through GitHub Actions → Backblaze → a new chassis image → bump `image_tag` on the agent rows.

---

## 2. The bug

Shared content components are referenced by many pages across multiple sites (one `content_components` row,
N `page_components` instances); at render time each page's stored `content_data` is bound into the
component's `html_template` by exact field name via `RenderTemplate`, so a `{{.field}}` placeholder only
fills when `content_data` holds a key of that identical name. When `component-creator` regenerates a
component through the `update_component_html` action it can rewrite the template/schema field names
(observed 2026-06-24 15:06 on the `system-stats` component `fdd92ad4`: `stat_1_number`→`stat1_value`,
`eyebrow`→`eyebrow_label`, `heading`→`section_headline`, `footnote`→`footnote_text`, and so on), and the
regeneration then re-renders every dependent page from that page's existing, un-migrated `content_data`,
which still carries the old key names. Because the new placeholders no longer match the old keys, every
field resolves empty, the section's `rendered_html` comes out blank, and the page assembler's
visible-content filter silently drops the whole band with no error raised; the failure fans out to every
page that shares the component in a single batch (all five affected `fdd92ad4` instances, on four other
sites, were rewritten together at 15:06:12.956, roughly 16ms after the component's own update). The
underlying `content_data` is not lost — it is intact but mis-keyed — so the breakage is recoverable.

**Why it is being handed off rather than fixed in the originating thread:** the flaw is in shared
component-regeneration machinery, the 2026-06-24 regen was a manual `component-creator:regen` (likely driven
by another chat that co-manages components), and no `gamesdesign.co.uk` page is affected — its `system-stats`
instance was dropped during an unrelated rebuild, so the originating thread had nothing to fix on its own site.

---

## 3. Evidence and IDs

- **Component:** `content_components` id `fdd92ad4-521a-4602-89cf-7ee1a06...` (`system-stats`; use the full UUID
  from a query — see RUNBOOK R1). Relevant columns: `html_template`, `input_schema`, `template_variable_count`,
  `schema_field_count`, `schema_template_synced`, `usage_count` (**stale — do not trust it as a live count**).
- **Version history:** `component_versions` has exactly one row for this component (`version_number` 1,
  `changed_by` = `component-creator:regen`, `change_source` = `manual`, created 2026-06-24 15:06:12.923611).
  There is **no pre-rename version to revert to** — the only stored version is already the renamed one.
- **Five affected instances** (all `component_id` = `fdd92ad4`, all `build_status` = `pending`,
  `component_version_id` = NULL, all `rendered_html` empty, all rewritten at 15:06:12.956). Page-id prefixes:
  `b2edae00`, `0747e2fc`, `9baea9f9` (each an `index` page on a **different** site),
  `77f10b13` (`case-study-kafka-consumer-group-remediation`), `11364960` (`gripper-detail`). Drive enumeration
  off `component_id`, not the prefixes (RUNBOOK R2).
- **Old keys (in the pages' `content_data`):** `eyebrow`, `heading`, `subheading`, `footnote`,
  `stat_N_number`, `stat_N_label`, `stat_N_suffix`, `stat_N_description`, `stat_N_icon`.
- **New keys (in the regenerated template/schema):** `eyebrow_label`, `section_headline`, `section_intro`,
  `footnote_text`, `statN_value`, `statN_suffix`, `statN_label`, `statN_description`, `cta_url`, `cta_label`.
  Note the mapping is mostly 1:1 but not entirely — `stat_N_icon` has no new target, and `cta_url`/`cta_label`
  are new with no old source. **Derive the real mapping from the live `input_schema`, do not trust this list
  from memory** (RUNBOOK R1/R3).

---

## 4. Where it lives in the code

Confirmed from the neighbourhood-signatures and registry in the prior bundle (`bundle_gamesdesign.md`):

- `platform/orchestration/actions/component_library.go` — `RenderTemplate(templateStr string, ctx *RenderContext, logger) string`
  does the bind; `RenderContext` carries `ContentData map[string]interface{}` (the page's values); `GetComponentByID`
  loads the component. This is where "bind by exact field name" actually happens.
- `update_component_html` action → `Handler: UpdateComponentHTMLAction` (in `registry.go`), description "Update
  html_template of a content_component with optional version snapshot". This is the regen write. The ~16ms gap
  between the component update and the dependents' re-render strongly suggests the dependent re-render is **inline
  in this same action/process** (not a Kafka round-trip) — confirm by reading it (`grep -rn "func UpdateComponentHTMLAction"`).
- `platform/orchestration/actions/component_selector.go` — holds the "component has empty `input_schema` → needs
  regeneration with content fields" logic that likely *triggers* a regen, plus `IncrementUsageCount`, `SelectComponentByType`.
- Recovery render: the no-LLM `rerender_page_sections` action (from earlier Part 2 work) re-renders a page's
  sections from stored `content_data` without an LLM.
- Related (not in the focal bundle): `save_page_sections_action.go` writes per-page `content_data`/`rendered_html`
  during builds and carries the earlier Part 4 interactive-section guard.

---

## 5. Suggested routes (directions to weigh, not a fixed choice)

- Preserve a component's existing field names across a regeneration, so bindings that already work keep working.
- When a regeneration does rename fields, migrate each dependent page's `content_data` keys old→new as part of
  (or immediately before) the re-render, so the bind still matches.
- Add a guard that treats a re-render which empties a previously-populated section as a failure to surface
  (log loud / mark the item failed) rather than a silent blank to ship — so any future regression is visible.

These are not mutually exclusive; the plan treats them as candidate directions to assess against the code, not
an either/or.

---

## 6. Recovery for the five already-broken pages

`content_data` is intact, just mis-keyed, so recovery is cheap and needs no LLM: align the keys to the current
`input_schema` (per-page `content_data` migration old→new), then trigger the no-LLM `rerender_page_sections`
path for those pages (a `page_rerender` work item per page; the `rerender-pages` agent / `build-dispatch-loop`
/ `page-rerender` chain already exists — reuse it, do not build a new one). Verify the bands reappear (RUNBOOK R5/R6).

---

## 7. Standing rules and cautions

- Reuse and adapt existing functions/structs/agents before creating anything new. Keep workflows thin and put
  logic in Go. No SQL sub-workflows — spawn sub-agents with their own workflows if orchestration is needed.
  Every agent is an orchestrator owning a workflow of steps that call actions.
- Check the live schema before writing SQL. Do not trust a 0-row / empty result until the query itself is
  verified. Do not jump to conclusions. Keep `logger.Debug` out (it does not show in the logs).
- **Concurrency caution:** another chat co-manages components. Before any write to shared component state,
  re-check freshness (RUNBOOK R7) — no blind writes. The 2026-06-24 regen was `change_source = manual`,
  consistent with that other chat.
- Tone: pragmatic, no banned words; no summary documents unless asked.

---

## 8. Pointers

- `BUNDLE_component_regen_clobber.md` — the `cmd/bundle` command + paths to confirm.
- `PLAN_component_regen_clobber.md` — phased plan.
- `RUNBOOK_component_regen_clobber.md` — the SQL and recovery steps.
- Background only (do not re-derive from these): the prior gamesdesign threads are in the originating chat's
  `HANDOFF_page_pipeline.md` / `NOTES_gamesdesign_silent_norebuild.md`; the prior bundle outputs
  `bundle_gamesdesign.md` and `bundle_gamesdesign_generation.md` hold the relevant code signatures and schema.
