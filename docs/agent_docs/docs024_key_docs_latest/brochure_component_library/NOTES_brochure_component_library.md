# NOTES — brochure component library (append-only, newest at the bottom)

## 2026-07-20 — kickoff, initial checks

- Grepped the whole repo for `fundamentallyai` (code/docs/seeds/deployments,
  case-insensitive): **zero hits**. Confirms it has never been onboarded as a site
  here — nothing to reuse, nothing to conflict with.
- `WebFetch https://fundamentallyai.com` → 307 redirect to
  `https://www.afternic.com/forsale/fundamentallyai.com?...` — a domain-marketplace
  "for sale" parked page, not a live site. **[VERIFIED]** the domain is not currently
  the owner's. Recorded as a correction in the PLAN file, not silently absorbed into
  the brief.
- Checked `docs/leopardessconsulting/` for prior consultancy-site design lessons
  (README_where_we_are.md, PLAN_imagery_and_design_2026-07-18.md) since it's the
  closest existing precedent for "consultancy brochure site through this framework."
  Key transferable facts pulled into the PLAN file: house style is flat illustration
  via the Banana generator (`kind:"illustration"`), `kind:"hero"` historically fell
  through to SDXL which cannot render legible text (model-class limitation, not a
  prompt-quality problem), and the standing rule "anything carrying words or numbers
  is code-rendered [SVG] from evidence-base values, never generated" — directly
  relevant to any "stat band" / infographic-style component this workstream might
  propose.
- Read `docs024_key_docs_latest/036_REFERENCE_styling_render_pipeline.md` (existing,
  authoritative reference doc, not written by this workstream) for how components are
  actually assembled: `content_components.html_template` (Go template + inline
  `<style>`), resolved to a page section by function name via `plan_sections`
  (one active component per function), CSS assembled from the layout template +
  `css_snippets` (tagged `applies_to` — the existing extension point for new
  component CSS, e.g. carousel/hover-zoom rules) + a renderer-owned `--section-*`
  luminance block. Pages are static HTML artifacts (git → GH Actions → Backblaze B2),
  not server-rendered per request. **[FINDING, per that doc, not independently
  re-verified this session]** — treat as ground truth unless the Explore agent's
  independent pass contradicts it.
- Direct `WebFetch` of the three named reference sites: bain.com succeeded (full
  homepage component catalogue captured — hero triple-carousel, partnership
  spotlight card, an interactive "Mad Lib" industry/need selector, a 4-slide client
  case-study carousel with 2-3 stat callouts per card, a featured-insights strip,
  dual CTA promo blocks, email subscribe band, footer). bcg.com returned **HTTP 403**
  to WebFetch; mckinsey.com **timed out twice** (60s). Both are large JS-heavy sites
  that plausibly block or throttle a plain fetch — left to the `deep-research`
  workflow (run `wf_51d0513a-4d5`), which fans out via search rather than one direct
  fetch per site, and to secondary design-teardown sources it should surface. Do not
  re-try direct WebFetch on bcg.com/mckinsey.com repeatedly — two failures each is
  enough signal that the fetch path itself is blocked, not a fluke.
- Dispatched an Explore agent (background) to map: component-type registry, the full
  spec/mission → rendered-page chain, CSS/JS delivery mechanism (confirm whether any
  shipped component already carries real JS interactivity — none confirmed yet in
  this session's direct reading, the 036 reference doc's plumbing suggests
  CSS-template + inline `<style>` per component but says nothing about `<script>`
  usage, so **this is an open question, not yet answered either way**), imagery
  eligibility/kind logic, a worked recent component-addition example, and how a new
  site actually gets onboarded. Results not yet returned as of this note.

## 2026-07-20 — Explore agent returned; full pipeline + registry confirmed

**[FINDING, independently verified this session, matches 036 reference doc]**

- **Registry**: `content_components` table — `function`, `section_type`,
  `component_level` (`site|page|section|element|head|header|footer|tool`),
  `render_mode` (`template|agent|composite`), `html_template` (Go template, inline
  `<style>`/optional `<script>`), `input_schema`, `is_dark_section`,
  `suitable_site_types`/`suitable_page_types`. Loader for planners:
  `load_component_library_actions.go`. Scoring/resolution:
  `component_selector.go` (queries by `section_type` + `component_level='section'`;
  if no match, queues a `needs_new_component` work item for **component-creator**,
  `component_selector.go:324`).
- **No carousel/slider/hover-zoom/autoplay component exists anywhere in the
  framework today** (`grep -riE "carousel|slider|swiper|hover-zoom|autoplay"` across
  the whole repo hits nothing but unrelated social-media docs and an archived
  vendor JS dump). **This is a genuine, from-scratch build**, not an
  under-used existing capability.
- **Important divergent signal — leopardessconsulting's OWN `design_intent.json`
  explicitly bans this exact family of thing**, in its `avoid[]` list:
  *"Charts produced by an image generator, under any circumstances"*,
  *"Decorative animation, parallax, and anything that delays the first paragraph
  being readable"*, *"Testimonial carousels, client logo walls, and social proof we
  have not earned."* That is a **deliberate brand-voice choice for that one site**
  (serious, undecorated, no-hype register — see its `PLAN_leopardess_rebuild.md`
  §3 "standing rules"), not a platform-wide judgement that carousels are bad. It's
  strong confirmation that **new component types must be opt-in per site's
  `design_intent`/`suitable_site_types`, not force-added to every brochure site** —
  the existing `avoid[]` + `suitable_site_types` mechanism already models exactly
  this kind of per-brand inclusion/exclusion, so no new mechanism is needed for
  that part.
- **Synergy worth acting on**: leopardessconsulting's own rebuild plan (owner brief,
  `PLAN_leopardess_rebuild.md` §1/§2/A5) already asked for "infographics that
  explain the guides and news visually" and a **"reusable chart component: Go + a
  JS renderer"** — phase **L7, still `⬜ not started`** as of that doc. That is the
  same "stat band / infographic" need this workstream's Bain-style brief implies
  (code-rendered, never diffusion-generated, per the standing rule both
  independently arrived at). **Building the code-rendered stat/chart component once,
  registered generically, serves both workstreams** rather than being built twice.
  Flag to owner as a proposed shared P1/P2 item.
- **Full pipeline, agent by agent** (fresh-site path):
  `domain-submitter` (creates `sites` row + mission text) → `domain-research-classifier`
  → `domain-strategist` → `build-briefing-agent` → `build-site-planner` (writes
  `site_plans`/`site_plan_pages`/`site_plan_sections`/`site_plan_imagery` — the
  authoritative, normalized plan; a page's `sections` is an ordered list of
  component **type names**, not content) → `site-design-planner` (composition/CSS,
  the `RenderCSSFromSpecAction` path in the 036 reference doc) → `webdesign-agent`
  → `page-build-handler` (resolves each section to a component via
  `plan_sections_action.go`, content-writer fills it) → rerender. Entry point:
  `082_submit_domain_unified.sh` (`--mission`/`--mission-file` records the owner's
  brief into the spec that weights the classifier — this is the "reachable from the
  mission downwards" hook the owner asked for; it already exists).
- **JS delivery — two lanes, and a landmine that DOESN'T apply to us but is worth
  knowing about**:
  1. `js_snippets` table (`applies_to` JSONB of component functions) → bundled into
     site-wide `/assets/js/snippets.js` by `render_js_snippets_for_site_action.go`.
  2. `content_components.js_content` → published per-component as
     `/tools/assets/{function}.js` by `collectJSAssets`
     (`rerender_single_page_action.go:156-176`), which SQL-joins
     `page_components` → `pages` → `content_components`.
  - **`bugs_open/041` (filed today, 2026-07-20, OPEN)**: `collectJSAssets` only
    reads `page_components`, not `site_components` — so a **chrome** component
    (header/footer, `component_level='header'/'footer'`, reached only via
    `site_components`) can declare `js_content` that is silently never published
    (idea.uk's mobile-menu hamburger does nothing, right now, in production).
    **This does not block us**: a new hero-carousel/swipeable-carousel component is
    a `component_level='section'` component reached via ordinary `page_components`
    — exactly the path `collectJSAssets` already handles correctly. Confirmed by
    reading the bug file directly, not inferred. Still worth a real 200-check on
    the published asset after building (bug 041's own fix-candidate #3 — a
    post-deploy "does every `<script src>` resolve" check — doesn't exist yet, so
    nothing else will catch it for us).
  - Implementation choice between the two lanes not yet made — leaning `js_snippets`
    for the auto-advance/hover-zoom/swipe behaviour (it's the documented
    general-interactivity extension point per the Explore report), decide at build
    time.
- **Imagery kind routing, current state**: `internal/adapters/imagegenerator/routing.go:59-66`
  — **every** declared kind (`icon, logo, illustration, infographic, sprite_sheet,
  content_hero, hero`) now routes to Banana (Gemini image). This supersedes the
  leopardess docs' "kind:hero falls through to SDXL" note — that was true when
  written and has since been fixed platform-wide (matches the imagery-workstream
  memory of "A6 Banana routing... deployed"). **[VERIFIED against current routing.go,
  not stale]** — so a photography-style hero for a new brand is not blocked by that
  particular historical bug; the open question is prompt/style direction (photography
  vs illustration), not provider routing.
- **Canonical "add a new component type" path** (for when we actually build):
  `needs_new_component` work item → component-creator agent generates
  `html_template` + `input_schema` → `StoreGeneratedComponentAction` inserts the
  `content_components` row. To make a *visually-rich* type actually land and get
  *used*, four more things are needed beyond the row itself: a `js_snippets` (or
  `js_content`) entry for behaviour, layout CSS/`css_snippets` for styling, a
  `site_plan_imagery` kind if it needs generated images, and — **this is the step
  most likely to be silently skipped** — the type name has to be enumerated in the
  **build-site-planner / site-architect prompt**, or `plan_sections`/
  `component_selector` will simply never select it even though it exists and
  renders fine standalone. (This exact failure class — a real fix that the planner
  never reaches because a prompt wasn't updated — is a recorded landmine elsewhere
  in this repo; see `travelling-docs-workstream` memory: "prompt seams dropping
  spec intent.") Treat "is it in the planner prompt" as a required item on this
  workstream's own acceptance checklist, not an afterthought.
- **Consultancy-tagged layout already exists**: `layouts` table row
  `brochure-formal` is tagged for `consultancy, law, finance, b2b,
  professional-services` and is what leopardessconsulting itself uses. Any new
  consultancy brand would plausibly start from the same layout tag, or a sibling
  `brochure-bold`/new layout if the visual language needs to diverge further (the
  Bain/BCG references are notably more kinetic than `brochure-formal`'s current
  css_template — to confirm once the deep-research pattern catalogue lands).
