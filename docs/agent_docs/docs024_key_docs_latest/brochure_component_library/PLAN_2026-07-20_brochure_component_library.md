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
  imagery eligibility → a worked recent example of adding a new component type: **see
  the Explore agent findings, to be folded in below once returned** (dispatched
  2026-07-20, in flight as this file was created).

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
  block) once P1's pattern is proven.
- **P3:** apply the set to real domains once the owner has reviewed a design sample.

## Corrections log (this file)

- 2026-07-20: Original ask named fundamentallyai.com as available to use; corrected
  above — it's a parked domain for sale, not owned.
