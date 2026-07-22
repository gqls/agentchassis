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
- `js_content` auto-publishes to `/tools/assets/{function}.js` and is `<script>`-
  included on any page using the component (the section path that works).
- Registration is enough to reach the planner: `load_component_library` returns all
  active `section` components in `AvailableFunctions` (no `suitable_site_types`
  gate). Whether the planner *chooses* it is a separate step — verify empirically.

## Acceptance checklist (per component)
- [ ] Go template parses + renders with sample data (see each component's dir).
- [ ] Registered in `content_components` (is_active, section, correct section_type).
- [ ] Appears in `load_component_library` AvailableFunctions for the site type.
- [ ] `curl` 200 on the published `/tools/assets/{function}.js` after a page uses it.
- [ ] Renders correctly on a live page; accessibility behaviours verified.
- [ ] Planner actually selects it (or the page is planned to use it explicitly).

## Components
| dir | function | status |
|---|---|---|
| `hero-card-carousel/` | `hero-card-carousel` | REGISTERED + template-validated 2026-07-22; live render on a page = next (needs the build queue, currently backlogged) |

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
