# Stage-2 interactive components (source of truth for the DB rows)

Components live in the `content_components` DB table, which is NOT in git. These
files are the **version-controlled source** for the Stage-2 components this
workstream builds — edit here, then re-apply `register.sql` (or an UPDATE) to the
live DB. Keep them in sync.

## Conventions (learned from existing section components, e.g. `info-card-grid`)
- `html_template` = an inline `<style>` block (CSS using the theme vars
  `--color-*`, `--spacing-section`, `--container-max-width`, `--border-radius`,
  `--shadow`) + Go-template HTML (`{{.field}}`, `{{range $i,$c := .cards}}`,
  `{{if …}}`). Context-aware `html/template` — do not pre-escape.
- `render_mode='agent'` → the content-writer fills `input_schema.fields` (each with
  `type`, `source:"llm"`, `required`, `llm_guidance`). `type:"image"` fields get a
  generated image URL + fallback.
- `component_level='section'` (never chrome — `bugs_open/041` drops chrome JS).
- **JS delivery — use `js_snippets`, NOT `content_components.js_content`.**
  PROVEN LIVE 2026-07-22 (hero-card-carousel): `js_content` publishes the file to
  `/tools/assets/{function}.js` (curl 200) but the assemble injects **no
  `<script>` tag** for it — so it is published-but-inert (the `bugs_open/041`
  class, and it applies to SECTION components too, not just chrome). The working
  lane is a `js_snippets` row (`applies_to: ["<function>"]`, `is_active`) →
  `render_js_snippets_for_site` bundles it into the site-wide
  `/assets/js/snippets.js` that every page already loads. Trigger the rebundle by
  firing the `site-asset-renderer` agent for the site (`{site_id, domain}`); no
  page re-deploy needed — pages already `<script src>` the bundle. Keep the
  component's `js_content` NULL to avoid a published-but-unused orphan asset.
- Registration is enough to reach the planner: `load_component_library` returns all
  active `section` components in `AvailableFunctions` (no `suitable_site_types`
  gate). Whether the planner *chooses* it is a separate step — verify empirically.

## Acceptance checklist (per component)
- [ ] Go template parses + renders with sample data (see each component's dir).
- [ ] Registered in `content_components` (is_active, section, correct section_type).
- [ ] Appears in `load_component_library` AvailableFunctions for the site type.
- [ ] JS (if any) added as a `js_snippets` row; after `site-asset-renderer` runs,
      the marker string appears in the live `/assets/js/snippets.js` (NOT via
      `js_content`/`/tools/assets/` — that publishes but is never `<script>`-loaded).
- [ ] Renders correctly on a live page; accessibility behaviours verified.
- [ ] Planner actually selects it (or the page is planned to use it explicitly).

## Components
| dir | function | section_type | JS | status |
|---|---|---|---|---|
| `hero-card-carousel/` | `hero-card-carousel` | hero-carousel | js_snippets | **PROVEN LIVE 2026-07-22** on fundamentallyai.com/capabilities.html — render + hover-zoom + scroll-snap swipe + auto-advance JS all working. |
| `image-hover-card-grid/` | `image-hover-card-grid` | image-hover-cards | none (CSS) | REGISTERED + template-validated 2026-07-22; hover-reveal cards, dark. Not yet placed on a page. |
| `swipeable-insight-carousel/` | `swipeable-insight-carousel` | insight-carousel | none (CSS) | REGISTERED + template-validated 2026-07-22; scroll-snap text cards. Not yet placed. |
| `stat-band/` | `stat-band` | stat-band | js_snippets (count-up) | REGISTERED + template-validated 2026-07-22; code-rendered stats, dark. Not yet placed. |
| `people-feature-block/` | `people-feature-block` | people-feature | none (CSS) | REGISTERED + template-validated 2026-07-22; line-illustration + statement, reversible. Not yet placed. |

**All five are the from-scratch interactive brochure components the workstream was
created to build.** hero-card-carousel is proven end-to-end (render + JS lane); the
other four are registered + template-validated but not yet rendered on a live page.
The remaining work is to **place them on the fundamentallyai pages** (re-plan pages
to use them, or place explicitly) — that is what makes the site "look like the
brief" — plus verify each renders live and, for stat-band, its count-up JS bundles
(fire `site-asset-renderer` once a page uses it).

### hero-card-carousel
Auto-advancing, swipeable hero carousel. Combines three of the requested effects:
- **Swipeable on mobile** — native CSS `scroll-snap` track (zero JS for the swipe).
- **Auto-advance** — JS (`behaviour.js`), fully accessible per WCAG 2.2.2 + ARIA APG:
  visible pause/prev/next `<button>`s (before the track in DOM), pauses on hover
  and on keyboard focus, stops when scrolled out of view, respects
  `prefers-reduced-motion` (no rotation), announces via an `aria-live` region.
- **Hover-zoom** card images — clipped container + `scale` transition, disabled
  under `prefers-reduced-motion` and `@media (hover:none)`.
The schema's `llm_guidance` enforces the research finding: the FIRST card must
carry the complete message (≈89% of carousel viewers only ever see slide one).
