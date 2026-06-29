# Bundle — component regeneration silently empties dependent pages

This is the `cmd/bundle` invocation for the cross-site platform bug where a shared
component's regeneration renames its fields and blanks every page that uses it. Run it
locally against your checkout; it produces the context bundle the next chat should start from.

---

## 1. The bug (single-paragraph statement — also used as `-task`)

Shared content components are referenced by many pages across multiple sites (one `content_components`
row, N `page_components` instances); at render time each page's stored `content_data` is bound into the
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
Routes worth investigating, offered as directions rather than a prescribed choice: preserving a
component's existing field names across a regeneration so established bindings stay valid; migrating each
dependent page's `content_data` keys from the old names to the new ones as part of (or immediately before)
the re-render; and adding a guard that treats a re-render which empties a previously-populated section as a
failure to surface rather than a silent blank to ship — with recovery of the already-affected pages then
being a key alignment/migration followed by a no-LLM re-render (`rerender_page_sections`) rather than a
full LLM rebuild.

---

## 2. The bundle command

```bash
go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root ~/projects/agentchassis \
  -constitution thin_slice_constitution.md -step debug \
  -task "Shared content components are referenced by many pages across multiple sites; at render time each page's content_data is bound into the component's html_template by exact field name via RenderTemplate. When component-creator regenerates a component through update_component_html it can rename the template/schema fields (2026-06-24 15:06 on system-stats fdd92ad4: stat_1_number->stat1_value, eyebrow->eyebrow_label, heading->section_headline, footnote->footnote_text, ...) and then re-renders every dependent page from that page's existing, un-migrated content_data (old keys). The new placeholders no longer match, every field renders empty, the section is blank, and the assembler silently drops the band with no error; it fans out to all 5 fdd92ad4 instances on 4 other sites, rewritten together at 15:06:12.956 (~16ms after the component update). content_data is intact but mis-keyed, so it is recoverable. Investigate: where in update_component_html the dependent re-render happens, and whether it should preserve field names, migrate dependents' content_data keys old->new before re-render, and/or fail loud when a re-render empties a previously-populated section; recovery = key alignment/migration + a no-LLM rerender_page_sections, not a full rebuild." \
  -scope platform/orchestration/actions/update_component_html_action.go:UpdateComponentHTMLAction \
  -scope platform/orchestration/actions/rerender_page_sections_action.go:RerenderPageSectionsAction \
  -include platform/orchestration/actions/component_library.go \
  -include platform/orchestration/actions/component_selector.go \
  -include platform/orchestration/actions/registry.go \
  -doc docs/.../016_debugging_guide_v2_56.md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables content_components,component_versions,page_components,pages,sites \
  -runtime-site '<AFFECTED-SITE-DOMAIN>' -runtime-page '<AFFECTED-PAGE-NAME>' \
  -capabilities -df-filter snapshot \
  -out /tmp/bundle_component_regen_clobber.md
```

---

## 3. Confirm before running (paths I inferred, not verified against the live tree)

These four are derived from the neighbourhood-signatures in `bundle_gamesdesign.md` plus the `update_component_html`
registry entry, not from reading the files directly — verify them in your checkout first:

- **`update_component_html_action.go:UpdateComponentHTMLAction`** — `update_component_html` maps to
  `Handler: UpdateComponentHTMLAction` in `registry.go`, but the handler's file is a guess. Confirm:
  `grep -rn "func UpdateComponentHTMLAction" ~/projects/agentchassis` — set `-scope` to the real path.
  This file is the prime suspect for the inline dependent re-render (the ~16ms gap points at the same process).
- **`rerender_page_sections_action.go:RerenderPageSectionsAction`** — the no-LLM rerender from the earlier
  Part 2 work (recovery path). Confirm: `grep -rn "RerenderPageSections\|rerender_page_sections" ~/projects/agentchassis`.
- **`-doc` path/version** — use the current latest debugging guide; `016_debugging_guide_v2_56.md` was the latest
  known. Swap for whatever the current file is, and consider adding the tool-lifecycle doc (`020_tool_lifecycle`).
- **`-runtime-site` / `-runtime-page`** — pick one of the five still-broken instances (RUNBOOK R2 lists how to
  resolve a page_id to its site domain). `gripper-detail` and `case-study-kafka-consumer-group-remediation` are
  distinctive page names among the five; `index` covers the other three (each on a different site).

Confident as-is: `component_library.go` (holds `RenderTemplate` + `RenderContext.ContentData` + `GetComponentByID`),
`component_selector.go` (holds the "empty `input_schema` → needs regeneration" trigger logic, `IncrementUsageCount`,
`SelectComponentByType`), and `registry.go` (wires `update_component_html` and the rerender actions). `-analysis`
and `-root` mirror the example; regenerate `analysis_repo.json` the same way you do for other bundles.

---

## 4. What the resulting bundle will contain

Same shape as `bundle_gamesdesign.md`: the thin-slice constitution (always-on rules), the task, the in-scope
code (the two focal symbols with their call-neighbourhood) plus the three included files verbatim, a
signatures-only neighbourhood of related functions, the live schema for the five tables, the DB capabilities
(`\dx`, `\df` filtered to snapshot/version helpers), and the runtime evidence for the chosen affected page.
That is enough for a cold-start chat to reason about the regen → bind → store chain and the recovery route
without re-deriving any of it.
