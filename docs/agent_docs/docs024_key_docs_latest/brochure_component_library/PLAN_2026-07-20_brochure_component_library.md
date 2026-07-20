# PLAN — brochure component library (started 2026-07-20)

## The ask

Owner wants more visually interesting components for our "consultancy-type" brochure
sites, modelled on best-in-class firms: bain.com, bcg.com, mckinsey.com. Named patterns
he called out explicitly:

- A hero made of several cards in a **self-refreshing (auto-advancing) carousel**.
- Each card has a **fancy image that slightly enlarges on view/hover** — a deliberate
  cheap substitute for video/gif: motion without the download weight.
- Each card: a well-placed **title + "read more" link**, sometimes one or two short
  extra strings of text, **never much copy**.
- Scrolling down: **several different component types**, many of them **carousels
  swipeable left/right on mobile**.
- Link targets ("information pages") also have **design variety**, not one template.
- McKinsey-style **people-focused imagery** — but we have a **veracity/claims
  checker** that forbids inventing people or facts, so people-imagery needs a design
  answer that stays honest (illustration/stock/anonymised, not "photo of a named
  person who doesn't exist" implying a real client or employee).
- Candidate new brand to experiment on: **fundamentallyai.com** — "my new consultancy
  branding."
- **Everything must render through the framework** — the spec/mission → page pipeline
  — not hand-authored HTML. New components are new entries in the existing component
  system, reachable from site design planning, so the plumbing is reusable across every
  site, present and future.

## Correction to the brief (2026-07-20) — fundamentallyai.com is not owned

Checked directly: `https://fundamentallyai.com` 307-redirects to an Afternic
"for sale" marketplace listing (`afternic.com/forsale/fundamentallyai.com`). It is
**not currently registered to the owner** — it's a third-party domain-for-sale parked
page. Nothing in the repo (`grep -ri fundamentallyai` across code/docs/seeds/deployments)
references it either, confirming it has never been onboarded as a site here.
**This blocks "onboard fundamentallyai.com" as a literal next step** until the owner
either buys it or names an already-owned domain to use instead. Flagged to the owner in
this session's reply; not treated as a green light to buy anything.

## Existing consultancy site as prior art — leopardessconsulting

The owner already runs a consultancy-branded site through this platform:
`docs/leopardessconsulting/`. Load-bearing facts from its docs (README_where_we_are.md,
PLAN_imagery_and_design_2026-07-18.md), relevant to this workstream:

- House style is **flat illustration** (Google `gemini-3-pro-image-preview`, the
  "Banana" generator, `kind:"illustration"`), explicitly **not photography** — chosen
  because photography-style prompts don't fit the brand and because `kind:"hero"`
  historically routed to SDXL, which cannot render legible text (a model-class
  limit, not a prompt problem — diffusion models synthesise glyph-shaped texture,
  not text).
- **Standing rule established there (worth inheriting):** "An image on this site is
  allowed to be a picture, or a real diagram. It is never allowed to be a picture
  pretending to be a diagram. Anything carrying words or numbers is code-rendered
  [SVG], driven by evidence-base values." This is the honest answer to "infographic /
  stat-band" style components — do NOT let an LLM/diffusion model render numbers.
- The **claims-verification / voice-tells checkers** already exist and are separate
  gates: claims (no fabricated facts/case studies/clients) and voice (no AI-prose
  tells). A components workstream doesn't need to build these, just not bypass them —
  new card copy still goes through the normal content-writer + validate_page_content
  path, never hand-written into content_data.
- **A structural trap already diagnosed elsewhere applies directly here**: `bugs_open/001`
  (re-plan clobbers built pages) — a full re-plan of a site can overwrite hand-tuned
  component wiring/copy. Per-image/per-component triggers (scope-less, content
  untouched) are the safe route; a full site re-plan is not, until 001's protections
  cover the target site's `build_status`.

Implication for fundamentallyai (or whatever brand is actually used): if it is meant to
look different from leopardessconsulting (which is intentionally illustration-only), that
is a **new house style decision**, not a default — see Open Decisions below.

## What the render pipeline actually is (grounded in
`docs024_key_docs_latest/036_REFERENCE_styling_render_pipeline.md`, verified-live findings
not theory)

- A page **section is a `content_components` row**: an `html_template` (Go
  `text/template`) with an inline `<style>` block, keyed by `component_level`
  (`section / header / footer / element / tool`) and resolved to a page section by
  **function name**, one active component per function (`plan_sections`).
- **Styling** is layered, not per-component free-form: the site's `css_themes` →
  `layouts.css_template` (Go template, `{{palette}}`/`{{typo}}`/`{{token}}` helpers)
  + `css_snippets` (rows tagged `applies_to`, e.g. animation/button/card utilities —
  **this is the existing extension point for new component CSS**, not a new
  mechanism) + a renderer-owned `--section-*` luminance block for dark sections.
  New components must **consume** `var(--section-*, var(--color-*))` rather than
  hardcode colours — `hero` and `call-to-action` currently violate this (self-declare
  dark backgrounds unconditionally), which is a known, separately-tracked defect, not
  a pattern to copy.
- **Pages are static HTML artifacts**, assembled by `CompilePageSectionsAction`
  (concatenate section HTML → inject head/header/footer) and deployed git → GitHub
  Actions → Backblaze B2. There is no per-request server render.
- Full end-to-end trace of spec/mission → agent chain → component registration →
  imagery eligibility → a worked recent example of adding a new component type:
  **returned 2026-07-20, folded into NOTES.** Headline facts: **no
  carousel/hover-zoom/slider component exists anywhere in the framework today** —
  this is a genuine from-scratch build, not an under-used capability. The mission
  text already flows all the way from `082_submit_domain_unified.sh --mission` down
  to the classifier that weights the site plan — the "reachable from the mission
  downwards" hook the owner asked for already exists; we're adding component
  *types* to the registry it selects from, not new plumbing.
- **Landmine confirmed, and confirmed NOT to block us**: `bugs_open/041` (filed
  today, unrelated thread) — a **chrome** (header/footer) component's declared JS
  is silently never published, because the asset collector only reads
  `page_components`. Our new carousel/hover-zoom components are ordinary
  `component_level='section'` components reached via `page_components`, i.e. the
  path that already works correctly. Still worth a real `curl` 200-check on the
  published JS asset after building, since nothing else checks this automatically.
- **The step most likely to be silently skipped**: registering a new component row
  is not enough — `component_selector`/`plan_sections` only ever select a type the
  **build-site-planner / site-architect prompt** actually names. A correctly-built,
  correctly-styled component that isn't mentioned in that prompt will simply never
  be chosen. Treat "confirmed present in the planner prompt" as a required
  acceptance item for every new component type this workstream ships, not an
  afterthought — see NOTES for the precedent this failure class has elsewhere in
  the repo.
- **A shared-build opportunity, not a new one**: leopardessconsulting's own rebuild
  brief (owner-authored, `docs/leopardessconsulting/PLAN_leopardess_rebuild.md`
  §2/A5) already asked for a **reusable code-rendered chart/infographic component
  (Go + JS renderer)** — phase L7, still not started. That is materially the same
  ask as this workstream's "stat band" component, under the same standing rule
  (numbers/words are code-rendered, never diffusion-generated). **Proposing this as
  ONE shared component built once and registered generically**, rather than two
  separate builds for two workstreams — see Open Decisions.

## Deep external research (dispatched 2026-07-20, in flight)

`deep-research` workflow run `wf_51d0513a-4d5` — catalogue of Bain/BCG/McKinsey/peer
brochure-site component patterns (hero card carousels, hover-zoom cards, swipeable
mobile carousels, stat bands, people-focused blocks), their cheap CSS/JS
implementation techniques (transform:scale + transition instead of video,
scroll-snap, lazy responsive images, `prefers-reduced-motion`), accessibility/perf
considerations, and how these sites use people photography editorially. To be folded
in below once returned.

Direct fetch of bain.com succeeded (see NOTES for the raw catalogue); bcg.com
returned HTTP 403 to WebFetch and mckinsey.com timed out twice — both are left to the
deep-research workflow's search-based fetch, which doesn't hit the same 403/timeout
as a single direct `WebFetch`.

## Open decisions the owner needs to make (do not assume)

1. **Domain.** Buy fundamentallyai.com, use a domain already owned, or park this as a
   design exercise with no live domain yet?
2. **House style for the new brand.** Photography-led (like Bain/BCG/McKinsey) needs
   a real imagery source — stock library, commissioned photography, or an
   illustration/duotone treatment that reads as "people" without depicting a specific
   invented individual. Illustration-only (leopardessconsulting's current approach)
   sidesteps the fabrication risk entirely but won't visually match the target
   references. This is a brand decision, not an engineering one — flag options,
   don't pick for him.
3. **Scope of "the framework."** Confirmed ask: new components must be registered in
   the same component system every site already uses (content_components +
   css_snippets + design_intent), reachable via site design planning — not a
   one-off hand-built page. Plan should propose component **types** (e.g.
   `hero-card-carousel`, `image-hover-card-grid`, `swipeable-insight-carousel`,
   `stat-band` [code-rendered, per the leopardess rule], `people-feature-block`) as
   additions to that registry, sized to fit the existing `component_level` /
   `applies_to` / `is_dark_section` conventions — not a parallel system.
4. **Shared chart/stat component with leopardessconsulting's outstanding L7?**
   Recommend yes — one code-rendered (Go + JS, never diffusion-generated) chart/
   stat-band component, registered generically enough that both leopardess's
   infographics need and this workstream's Bain-style stat band draw from it. Needs
   the owner's confirmation it's fine to size the component for two workstreams'
   use rather than leopardess-specific fields only.

## Phasing (draft — to be firmed up once Explore + deep-research land)

- **P0 (this session):** research (external patterns + internal pipeline), owner
  decisions above, a written component-by-component spec proposal.
- **P1:** build 1–2 new component types end-to-end through the real pipeline (likely
  `hero-card-carousel` first — it's the most-requested, most reusable across every
  brochure-style site, and exercises the carousel-JS + hover-zoom-CSS + imagery-kind
  questions in one go), council-reviewed per CLAUDE.md before commit (touches
  `platform/`/`internal/`).
  Prove it on ONE site first — not leopardessconsulting (already in a delicate,
  actively-audited state; see the clobber risk above) and not a live customer site.
  A throwaway or the new brand's site (once a domain decision is made) is the
  correct first target.
- **P2:** roll additional component types (swipeable card grid, stat band, people
  block) once P1's pattern is proven. Stat band = candidate shared build with
  leopardess L7 (Open Decisions #4).
- **P3:** apply the set to real domains once the owner has reviewed a design sample.

**Acceptance checklist for every new component type shipped in P1/P2** (added after
the Explore agent's findings — these are the specific ways a "done" component
silently fails on this platform, not generic best practice):
1. Row exists in `content_components` with correct `component_level`.
2. **Confirmed present in the build-site-planner / site-architect prompt** — the
   planner will otherwise never select it even though it renders fine standalone.
3. CSS consumes `var(--section-*, var(--color-*))`, never hardcodes colour —
   verified against a dark-section AND a light-section site, not just the one it
   was designed on.
4. Any JS asset it publishes returns a real `200`, checked by `curl` against the
   live deployed URL — not assumed from a "complete" work-item status.
5. If it carries a `site_plan_imagery` kind: confirmed against a real generated
   image before wiring (leopardess's own rule — don't let a bad generation ship
   unseen).
6. Copy path is content-writer + `validate_page_content`, never hand-authored into
   `content_data` — so claims/voice gates still apply.

## Corrections log (this file)

- 2026-07-20: Original ask named fundamentallyai.com as available to use; corrected
  above — it's a parked domain for sale, not owned.
