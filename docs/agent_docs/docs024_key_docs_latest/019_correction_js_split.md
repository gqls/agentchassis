# 019 correction — "When to split template from component" (replaces that subsection)

> Splice target: `019_tool_library.md` → `### When to split template from
> component` under "Storage and Query Patterns". The current text ends
> "This isn't built yet and shouldn't be needed for current tools" — verified
> stale on 2026-06-11 against store_generated_component_action.go,
> rerender_single_page_action.go, deploy_tool_action.go,
> create_tool_component_action.go, and the tool-generator/tool-improver
> agent definitions. Replacement text below.

### When to split template from component

For most tools (under 200KB), keep everything in `html_template`.

A template/JS split IS built — but for the component-creator pipeline
(games, feeds, explorers), not for tools, and not via the `assets` table this
section once envisioned. The implemented mechanism is the **JS Content
Separation Contract (003)**: `store_generated_component_action.go`'s
`separateInlineJS()` extracts attribute-less inline `<script>` blocks into
`content_components.js_content` (multiple blocks combined; `<script src>` /
`<script type=…>` left untouched), inserts
`<script src="/tools/assets/{function}.js"></script>` after the last
`</section>`, and `RerenderSinglePageAction.collectJSAssets()` emits the
asset into the same multi-file git commit as the page HTML. The `assets`
table is not involved.

The TOOL pipeline deliberately stays inline end-to-end: the tool-generator
prompt mandates a single inline `<script>` block, `create_tool_component`'s
INSERT writes `html_template` only (no `js_content`), the tool-improver
prompt mandates inline script, and `update_component_html` writes
`html_template`.

**Do not move tools onto the split without fixing two known gaps first:**

1. **Fork copy omits `js_content`** — `deploy_tool_to_site`'s
   `INSERT … SELECT` copies `html_template` (which would carry the
   `<script src>` reference) but NOT `js_content`. A library tool using the
   split would fork into a page whose script 404s. Currently unarmed (no
   library tool has `js_content`); armed the day one does.
2. **`component_versions` has no `js_content` column** — version snapshots
   of split components lose the JS
   (`store_generated_component_action.go` self-documents this at the
   `_ = jsContent` line).
