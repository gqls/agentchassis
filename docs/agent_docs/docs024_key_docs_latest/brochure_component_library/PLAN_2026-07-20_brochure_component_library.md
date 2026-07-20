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

## Correction to the brief (2026-07-20) — fundamentallyai.com, resolved

Checked directly: `https://fundamentallyai.com` 307-redirected to an Afternic
"for sale" marketplace listing at the time of checking. Flagged to the owner
rather than assumed either way. **Owner has since confirmed he owns the domain
and will point it at static hosting shortly** — the parked-page redirect was
presumably the registrar's default landing page for a domain not yet pointed
anywhere, not evidence of third-party ownership. DNS/hosting is **not yet live**
as of this note, so `082_submit_domain_unified.sh` can't complete an end-to-end
onboard today, but content/design/research work proceeds now in parallel —
fundamentallyai.com is the confirmed target site for this workstream.

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

## Part 2 (added 2026-07-20) — content & positioning: marketing our own real capabilities

Owner has broadened the brief substantially. fundamentallyai.com isn't just a
design exercise in nicer components — it should **market this platform's own
real capabilities** as service lines and case studies, the same way Bain
markets "M&A" or McKinsey markets "AI transformation." Candidate service
lines/case studies the owner named, to be grounded in true facts before any
copy is written (see NOTES for the dispatched research and the pgvector/
rag_actions.go grep hit that suggests the embeddings idea may be real, not
just aspirational — confirmation pending):

- **Private in-house search via embeddings** — "implement the whole framework
  in-house to some corporate or partnership and use embeddings to safely let
  them search their in-house databases without leaking info to outside
  organisations." Needs the capability-inventory research to confirm how much
  of this exists today vs. would need building.
- **Instant marketing/product-test/presentation sites** — spinning up a site
  fast for a launch, a test, or a pitch. This platform's own spec→live-site
  pipeline (the subject of Part 1 above) IS the proof-point here, if we can
  cite a real turnaround time.
- **Fine-tuning capability** as a service line.
- **Council/multi-agent review decision-making** as a service line/trust story
  — genuinely distinctive and, per CLAUDE.md itself, live and commissioned with
  a real track record.
- **Backend engineering proof-points**: idea.uk's Stripe integration,
  relojistas.com's expired-domain traffic-revival work — proposed as internal
  case studies (real projects, our own platform's output, not fabricated
  clients — matches leopardess's own "our sites demonstrate the platform, never
  implied to be a client roster" rule).

**Standing rule for this half of the workstream**: get the facts before the
copy. This is a claims-verification-constrained brand exactly like leopardess —
every capability claim needs an evidence citation (file, doc, DB figure) before
it becomes a card title or a case-study paragraph. Dispatched an Explore agent
2026-07-20 to build a 7-part LIVE/BUILT-BUT-INERT/ASPIRATIONAL capability
inventory — see NOTES for the brief and (once returned) the findings.

**This workstream's own output is a fact base and content brief, not final
copy.** Final page text is written by the framework's content-writer agent from
`site_specs`/mission/briefing, same as every other site — matching the owner's
explicit preference stated on the leopardess voice rewrite ("I don't want it
written here manually"). Don't hand-author finished marketing prose here; hand
the content-writer a well-evidenced brief instead.

## Part 2 (continued) — grounded content brief (capability inventory returned 2026-07-20)

Full evidence in NOTES. Proposed positioning pillars, each tagged with its
honest status — this is the content brief for whenever the content-writer
agent actually fires, not copy to ship as-is:

1. **Rapid site delivery** — LIVE/VERIFIED. Real dated example: a 33-page site
   rebuilt overnight, largely unattended, with a live 9-source news feed and 5
   interactive calculators (verified against production, 2026-07-10). Directly
   answers "instant marketing/product-test/presentation sites" — cite the real
   example, don't hypothesise a capability we haven't shown.
2. **Multi-agent council review / AI governance** — LIVE/VERIFIED, the
   strongest and most differentiated pillar. 13 independent reviewer seats,
   live since 2026-07-17, real growing decision record, a genuine
   production-risk bug caught from an external submission, and a
   self-correcting culture on record (a commit retracting its own falsely
   claimed review). Recommend this as the FLAGSHIP case study.
3. **Fine-tuning on real usage data** — LIVE/VERIFIED for one completed,
   honestly-evaluated cycle (real production data, blind A/B against a
   frontier model, non-flattering result disclosed). Do NOT claim the fully
   automated unattended flywheel — that part is built but has never completed
   a live cycle.
4. **Real backend engineering (payments)** — idea.uk's Stripe integration,
   LIVE/VERIFIED as a real hand-rolled checkout+webhook flow; do not cite a
   transaction volume (unconfirmed in the docs).
5. **Expired-domain revival / traffic engineering** — relojistas.com,
   LIVE/VERIFIED, the cleanest evidence-grade case study found: legacy feed
   404→~97% success within 24h of launch, with the crawler-vs-human caveat
   disclosed rather than hidden. Strong, ready to use largely as-is.
6. **Private in-house search via embeddings — NEEDS OWNER FRAMING DECISION,
   see Open Decisions #5.** Real, production-proven vector search
   infrastructure exists; tenant isolation does not exist yet on our own
   shared testbed (the platform's own prior audit rules this out explicitly).
   Must be framed as "buildable, because the hard technical part is proven,"
   not "already delivered safely for outside parties" — the latter would be
   the exact class of claim the platform's own verification system exists to
   catch.
7. **Claims-verification / anti-hallucination discipline — NEEDS OWNER CALL,
   see Open Decisions #6.** A genuinely compelling meta-narrative (an AI
   platform that checks its own AI-generated content against evidence and
   catches fabrications, including from its own prior site) — but using it
   means referencing that a past mistake happened, which is a brand-tone
   decision, not an engineering one.
8. **"We run our own fleet"** — 11 live production sites is a true, checkable
   headline stat. Per the exclusion list, present it in rounded/categorical
   form (content sites, a game platform, a paid tool with real payments, a
   revived expired domain) rather than naming every site — and do NOT name
   leopardessconsulting specifically without sign-off (Open Decisions #7).

**A hard constraint carried into every pillar above**: a specific, named list
of past leopardess fabrications (an invented founder, an invented "70+ agents/8
departments" org structure, invented case-study clients, a fake uptime stat, a
garbled "awards won" figure) must **never** resurface on fundamentallyai.com —
full list in NOTES. This isn't a style preference; it's the exact failure mode
the claims-verification system exists to prevent, on a brand-new site that
won't yet have that system's benefit of hindsight.

## Open decisions the owner needs to make (do not assume)

1. **Domain — RESOLVED by owner 2026-07-20.** fundamentallyai.com is owned;
   hosting to be pointed shortly (not yet live as of this note).
2. **House style for the new brand — RESOLVED by owner 2026-07-20: line
   illustration for people.** Not photography, not stock, not full-colour
   illustration necessarily — line-illustration figures, which sidesteps the
   fabrication risk by construction (a line drawing never reads as "a photo of
   someone who doesn't exist"). McKinsey's own technique (a single uniform
   duotone/tint treatment applied across all imagery, per the deep-research
   findings in NOTES) is a directly transferable idea for giving our line
   illustrations the same fleet-wide visual cohesion, cheaply.
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
5. **Embeddings/private-search framing — must be resolved before that pillar's
   copy is written.** Recommend: pitch it as "buildable, because the hard
   technical part (production-proven vector search) is already solved" — NOT
   as an already-delivered guarantee of safe multi-tenant isolation, which the
   platform's own audit says doesn't exist yet. This is close to a hard
   constraint (see NOTES), flagged here as a decision point only because the
   owner may want different wording, not because the underlying fact is
   negotiable.
6. **Use the claims-verification/anti-hallucination story as a case study?**
   Recommend yes, strategically — it's genuinely differentiated ("we verify
   our own AI's output and catch its mistakes, including our own past ones")
   — but it means referencing that leopardess previously shipped fabricated
   content, which is a brand-tone call for the owner, not something to include
   or exclude by default.
7. **Name leopardessconsulting.co.uk specifically as a proof-point site?**
   Recommend no, or only with explicit sign-off — it's a sibling brand with its
   own identity and a documented (if now-fixed) fabrication history. The "11
   live sites" stat works fine without naming every one.

## Phasing (updated 2026-07-20 — domain + house style resolved, scope broadened)

- **P0 (this session):** research — external design patterns (done, see NOTES),
  internal render pipeline (done, see NOTES), internal capability/case-study
  inventory (dispatched, in flight) — converging into a written
  component-by-component spec proposal AND a grounded content brief, both for
  owner review before any build/content-writer firing.
- **P1:** build 1–2 new component types end-to-end through the real pipeline
  (likely `hero-card-carousel` first — most-requested, most reusable, exercises
  carousel-JS + hover-zoom-CSS + imagery-kind in one go), council-reviewed per
  CLAUDE.md before commit (touches `platform/`/`internal/`). Target site is now
  confirmed: **fundamentallyai.com**, once its hosting is pointed and it's
  onboarded via `082_submit_domain_unified.sh` — not leopardessconsulting
  (delicate, actively-audited state; see clobber risk above) and not any other
  live customer site.
- **P2:** roll additional component types (swipeable card grid, stat band,
  line-illustration people block) once P1's pattern is proven. Stat band =
  candidate shared build with leopardess L7 (Open Decisions #4).
- **P3:** apply the content brief (Part 2) — real capability case studies,
  written by the content-writer agent from a grounded brief, never hand-authored
  here — once the owner has reviewed both the component design sample and the
  fact base.

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

- 2026-07-20: This file initially recorded fundamentallyai.com as not owned
  (Afternic parked-page redirect at time of check) and flagged it to the owner
  rather than assuming. Owner confirmed same-session that he does own it —
  recorded above as resolved, not silently overwritten.
