# Bundle — component regeneration silently empties dependent pages

This is the `cmd/bundle` invocation for the cross-site platform bug where a shared
component's regeneration renames its fields and blanks every page that uses it. Run it
locally against your checkout; it produces the context bundle the next chat should start from.

---

## 1. The bug (single-paragraph statement — also used as `-task`)

Shared content components are referenced by many pages across multiple sites (one `content_components`
row, N `page_components` instances); at render time `RenderTemplate` (component_library.go) binds a page's
stored `content_data` into the component's `html_template` by exact field name using Go's `text/template`,
and any placeholder with no matching key renders to nothing — the function strips the resulting `<no value>`
tokens to empty string without raising an error. When `component-creator` regenerates a component it can
rewrite the template/schema field names (observed 2026-06-24 15:06 on the `system-stats` component
`fdd92ad4`: `stat_1_number`→`stat1_value`, `eyebrow`→`eyebrow_label`, `heading`→`section_headline`,
`footnote`→`footnote_text`, and so on) without migrating the dependent pages' `content_data`, which still
carries the old key names. The write itself is clean — `update_component_html` only snapshots the old
`html_template`, swaps in the new one, and marks every dependent page `build_status='pending'` (that single
`UPDATE … SET updated_at=NOW()` is what stamped all five instances at the same 15:06:12.956 instant; it
deliberately does not touch `rendered_html`). The damage lands when a render step subsequently rebuilds a
pending page from its old-key `content_data` against the renamed template: every renamed placeholder misses,
the section renders blank, and the assembler's visible-content filter silently drops the band — fanning out
to every page that shares the component. The exact render step that wrote the blanks is not yet pinned (the
no-LLM `rerender_page_sections` carries the stored HTML when a section isn't "ready" and is gated to
image/section-data reasons, so the prime suspects are `component-creator`'s own dependent re-render and
whatever consumes `build_status='pending'`); confirming it is the first job. The `content_data` is intact,
only mis-keyed, so recovery is cheap. Routes worth investigating, offered as directions rather than a
prescribed choice: preserving a component's field names across a regeneration; migrating each dependent's
`content_data` keys old→new as part of the regeneration before anything re-renders; and a guard in the
render path that treats a re-render which empties a previously-populated section as carry-the-stored-HTML or
fail-loud rather than ship-a-blank (the rerender already has a carry-forward path for not-ready sections that
this could extend). Recovery for the five affected pages is then a `content_data` key alignment followed by a
no-LLM re-render, not a full rebuild.

---

## 2. The bundle command

```bash
go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root ~/projects/agentchassis \
  -constitution thin_slice_constitution.md -step debug \
  -task "Shared content components are reused by many pages across multiple sites; RenderTemplate (component_library.go) binds a page's stored content_data into the component's html_template by exact field name using Go text/template and silently strips unmatched {{.field}} placeholders to empty (it replaces the <no value> tokens, no error). On 2026-06-24 15:06 component-creator regenerated system-stats fdd92ad4 with renamed fields (stat_1_number->stat1_value, eyebrow->eyebrow_label, heading->section_headline, footnote->footnote_text, ...) WITHOUT migrating dependents' content_data, which still has the old keys. The write was clean: update_component_html only snapshotted old html, swapped the template, and marked dependents build_status=pending (that UPDATE ... SET updated_at=NOW() is the synchronized 15:06:12.956 timestamp; it does NOT touch rendered_html). The blanks were written later when a render rebuilt the pending pages from old-key content_data against the renamed template: every renamed placeholder missed, sections rendered empty, the assembler dropped them silently across all 5 fdd92ad4 instances on 4 other sites (gamesdesign not affected — its instance was dropped). FIRST JOB: pin which render step wrote the blanks — update_component_html does not render, and rerender_page_sections carries stored HTML when a section is not ready and is gated to image/section-data reasons, so suspect component-creator's own dependent re-render (likely via RenderComponentAction) and whatever consumes build_status=pending. content_data is intact but mis-keyed, so recovery is cheap. Routes (directions, not a fixed choice): preserve field names across a regen; migrate dependents' content_data keys old->new during the regen before any re-render; a guard that carries stored HTML / fails loud when a re-render empties a previously-populated section (extend rerender's existing not-ready carry-forward). Recovery = re-key content_data + no-LLM rerender_page_sections, not a full rebuild." \
  -scope platform/orchestration/actions/update_component_html_action.go:UpdateComponentHTMLAction \
  -scope platform/orchestration/actions/rerender_page_sections_action.go:RerenderPageSectionsAction \
  -scope platform/orchestration/actions/v3_site_actions.go:RenderComponentAction \
  -scope platform/orchestration/actions/store_generated_component_action.go:StoreGeneratedComponentAction \
  -include platform/orchestration/actions/component_library.go \
  -include platform/orchestration/actions/component_selector.go \
  -include platform/orchestration/actions/registry.go \
  -doc docs/.../016_debugging_guide_v2_57.md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables content_components,component_versions,page_components,pages,sites \
  -runtime-site 'robot-hands.com' -runtime-page 'gripper-detail' \
  -capabilities -df-filter snapshot \
  -out /tmp/bundle_component_regen_clobber.md
```

---

## 3. Before running (status of each input)

- **`update_component_html_action.go:UpdateComponentHTMLAction`** — CONFIRMED (you supplied the file). It is
  clean: snapshots the old `html_template`, swaps the new one in, and marks dependents `build_status='pending'`;
  it does **not** render or touch `rendered_html`. Kept in scope to document that and rule it out as the clobber.
- **`rerender_page_sections_action.go:RerenderPageSectionsAction`** — CONFIRMED (you supplied the file). The
  no-LLM re-render; it carries the stored HTML when a section isn't "ready" and is gated to
  `image_landed`/`section_data_resolved`, so it is partly protective and not the obvious writer of the blanks.
- **`v3_site_actions.go:RenderComponentAction`** — CONFIRMED location (it is in `v3_site_actions.go`, not a
  `render_component_action.go`). It renders `comp.HTMLTemplate` against a `RenderContext` built from the supplied
  `content_field`/`content_from` ⊕ `merge_with`, and **returns** `{rendered_html, content_data, …}` — it does
  **not** write to the DB and does **not** set `updated_at`. So it is the *writer's* render (fresh content), not a
  re-render from stored `content_data`; a persist step (`save_page_sections`) does the write.
- **`store_generated_component_action.go:StoreGeneratedComponentAction`** — CONFIRMED (you supplied the file) and
  now in scope. It is the **rename writer**: when a component with the same `function` exists it takes the
  regeneration branch — snapshots the old `html_template`/`input_schema` to `component_versions`
  (`changed_by='component-creator:regen'`), then `UPDATE content_components SET html_template=…, input_schema=…`
  **in place** (same `component_id`, so dependents now point at the renamed template), marks dependents
  `build_status='pending'`, and raises one `needs_rerender` work item per affected site. It does **not** migrate
  dependents' `content_data`, and the `needs_rerender` it raises has **no `reason`**, so the rerender it triggers
  is **assemble-only** (re-ships existing `rendered_html`) — it cannot repair a content-key mismatch. The fix
  likely belongs here (migrate dependents' `content_data` old→new using `existingSchema`/`inputSchemaJSON`, or
  preserve field names), and recovery must use `reason=section_data_resolved` to actually re-render sections.
- **`component-creator` agent** — CONFIRMED from its `default_config` (v1.0.1080): the workflow is
  `ensure_site_record → read_site_spec → generate_template → store_component (store_generated_component) → complete`.
  There is **no** `update_component_html` step and **no** re-render-of-dependents step — so component-creator does
  not re-render dependents, and the `component_versions` row tagged `changed_by='component-creator:regen'`,
  `change_source='manual'` points to an out-of-band/manual regen rather than this workflow (which would tag
  differently). Read `StoreGeneratedComponentAction` to see whether it overwrites an existing component and what it
  tags.
- **`-doc`** — `016_debugging_guide_v2_57.md` (the latest you uploaded; `016b_debugging_guide_6_.md` is the Vol.2
  companion). Swap if a newer one exists; consider adding the tool-lifecycle doc (`020_tool_lifecycle`).
- **`-runtime-site` / `-runtime-page`** — wired to `robot-hands.com` / `gripper-detail`, one of the five affected
  instances (single-instance, distinctive). The other picks: `ai-agent-orchestration.com` /
  `case-study-kafka-consumer-group-remediation`, or `index` on `leopardessconsulting.co.uk` / `robot-hands.com` /
  `ai-agent-orchestration.com`. (`gamesdesign.co.uk` is **not** affected — its instance was dropped — so it is not
  a useful runtime target here.)

Confident as-is: `component_library.go` (holds `RenderTemplate` — including the `<no value>` → empty cleanup that
makes the miss silent — plus `RenderContext.ContentData` and `GetComponentByID`), `component_selector.go` (the
"empty `input_schema` → needs regeneration" trigger logic, `IncrementUsageCount`, `SelectComponentByType`), and
`registry.go` (wires `update_component_html` and the rerender actions). `-analysis` and `-root` mirror the example;
regenerate `analysis_repo.json` the same way you do for other bundles.

---

## 4. What the resulting bundle will contain

Same shape as `bundle_gamesdesign.md`: the thin-slice constitution (always-on rules), the task, the in-scope
code (the three focal symbols with their call-neighbourhood) plus the three included files verbatim, a
signatures-only neighbourhood of related functions, the live schema for the five tables, the DB capabilities
(`\dx`, `\df` filtered to snapshot/version helpers), and the runtime evidence for the chosen page. That is
enough for a cold-start chat to reason about the regen → bind → store chain and the recovery route without
re-deriving any of it.
