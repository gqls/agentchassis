# Phase 4.4 — Cleanup summary and outstanding work

**Scope:** Remove dead code from the pre-4.3 renderer and document what's still coupled to the legacy `css_themes` columns so Phase 5 and beyond can unwind it cleanly.

## What Phase 4.3 already removed

The replacement `render_css_from_spec_action.go` dropped in Phase 4.3 no longer contains:

- `cssTemplateData` struct (and its 16 hardcoded fields)
- `extractDesignColors` (the legacy struct-field populator)
- `designColorMaps` wrapper type and its `color`/`typo`/`space` methods
- `getMapString` helper (used only by the wrapper)
- `loadCSSGoTemplate` (direct read of `css_themes.css_template` with "standard-brochure" fallback)

`sectionStyleEntry` lived here before and continues to live here (still used).

These were all internal to the action file. Deleting them is accomplished by replacing the file — no other action needed.

## Codebase callers of legacy renderer internals: audit result

Grepped the entire project context for legacy type/function names. Results:

| Identifier | References outside the old action file? |
|---|---|
| `loadCSSGoTemplate` | None. Confined to the old action. Gone after 4.3. |
| `cssTemplateData` | None. Gone after 4.3. |
| `extractDesignColors` | None. Gone after 4.3. |
| `designColorMaps` | None. Gone after 4.3. |

All four live and die inside `render_css_from_spec_action.go`. The replacement file doesn't reintroduce them. Compile-clean.

## What's NOT gone — flagged for Phase 5 and beyond

### 1. `css_templating.go` — inverse of the old renderer, used by fork path

This file (`platform/orchestration/actions/css_templating.go`) contains:

- `rootBlockRE` — regex to find `:root { ... }` in a stylesheet
- `paletteTemplateField` — map of `"primary" → "{{.Primary}}"` etc.
- `typographyTemplateField` — same for font fields
- `spacingTemplateField` — same for spacing fields
- `TemplateCSSFromSpec(renderedCSS, spec) string` — takes rendered CSS and produces templated CSS by substituting hex values with `{{.Primary}}`-style placeholders

**Used by:** `fork_theme_from_site` (line 55053 in the context dump) — converts a rendered CSS snapshot of an adopted site into the value it writes to `css_themes.css_template`.

**Coupled to:** the legacy `cssTemplateData` struct. Its placeholder map uses the exact field names (`.Primary`, `.FontFamily`, etc.) that the old renderer expected.

**Status after Phase 4.3:** This file still compiles and still runs. But the output it produces — templated CSS with `{{.Primary}}` — is written to `css_themes.css_template` and then NEVER READ, because the new renderer reads from `layouts.css_template` (via `layout_id`), not `css_themes.css_template`.

**Result:** `fork_theme_from_site` produces rows with NULL `palette_id`, NULL `layout_id`, NULL `typography_set_id`. When the renderer tries to resolve one of these themes (`loadThemeComposition`), it hard-errors with the "NULL palette_id — migration gap; run Phase 3 mapping" message I set up in Phase 4.2. Adoption-forked themes are unusable by the render path.

**What must change:** Phase 5 rewrites `fork_theme_from_site` for granular forks. After Phase 5, forking an adopted site produces three new rows: a `palettes` row (from the site's colours), a `typography_sets` row (from the site's fonts, or a reference to an existing one if the stack matches), optionally a new `layouts` row (only if the site's structural grammar doesn't match an existing one; usually reuses). The new `css_themes` row links to those three via FKs.

**Not a Phase 4.4 concern because:** Phase 4.4 is renderer-side cleanup. The fork path was broken for adopted sites either way; Phase 4.4 just makes the failure mode explicit instead of silent.

**Do not delete `css_templating.go` in Phase 4.4.** Deleting it breaks `fork_theme_from_site` compilation, which Phase 5 rewrites. Leave it in place until then.

### 2. `getThemeByID` / `GetThemeByName` — parallel HTML-assembly render path

These functions read `css_themes.css_content` (the rendered CSS, not a template) and feed it to a different pipeline that builds HTML with the theme's CSS baked in. Not the `render_css_from_spec_action` path. They remain unchanged by Phase 4.

`css_content` is populated for 13 of the 14 themes. `standard-brochure` has empty `css_content` because its only populated column was `css_template` (the one the new renderer now ignores).

**Impact on the HTML-assembly path:** if something asks for `standard-brochure` theme via that legacy path, it gets empty CSS and falls through to `GetThemeByName("default")`. Not broken, but also not ideal.

**Resolution path:** this isn't on Phase 4's critical path. Flagging so it doesn't get forgotten. Phase 7's column-drop cleanup will eventually need a migration step that re-renders `css_content` from the new composition data (or rewrites this reader to use the new composition too). Could be called Phase 6.5 or folded into Phase 7.

### 3. Legacy `css_themes` columns still being written

After Phase 4.3, these columns are written by `fork_theme_from_site` but not read by the render path:

- `css_themes.css_template` — written by fork, read by legacy renderer (now removed)
- `css_themes.color_palette` JSONB — written by fork, read by `getThemeByID` and the HTML-assembly path
- `css_themes.typography` JSONB — written by fork, NOT read anywhere after Phase 4.3
- `css_themes.css_content` — written by fork, read by `getThemeByID` (see #2 above)

None of this is urgent. Phase 7 drops these columns per the plan.

## Phase 4 status

- Phase 4.1 — pure helpers ✓
- Phase 4.2 — DB loader ✓
- Phase 4.3 — renderer cutover ✓
- Phase 4.4 — cleanup + audit (this document)

Phase 4 is complete. The renderer now sources palette/layout/typography from the independently-versioned tables via FKs, applies the correct merge rules, and hard-errors on migration gaps.

## Ready to move to

**Phase 4.5** — decouple surface-painting from layout templates + renderer class list, as flagged during Phase 1. Adds `data-section-bg="surface"` as the coupling mechanism between components and the renderer's section-defaults emission, rather than the hardcoded class list `(.features-section, .services-section, .differentiators-section, .about-section, .faq-section)`.

**Phase 5** — rewrite `fork_theme_from_site` for granular forks. Without this, adoption is broken. This is arguably higher-priority than 4.5 because it affects the production adoption pipeline.

Recommend Phase 5 next. 4.5 is a refactor; 5 is a fix.
