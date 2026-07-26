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

## 2026-07-20 (continued) — owner corrections + deep-research workflow returned partial

**Two open decisions resolved by the owner directly, not inferred:**
- **Domain**: owner confirmed **fundamentallyai.com is owned** and will be pointed
  at static hosting "shortly." Supersedes the earlier finding that it's a parked
  domain for sale — that finding was correct *at the time it was checked*
  (redirects to an Afternic listing), the owner has since clarified he holds the
  registration. DNS/hosting is **not yet pointed** as of this note, so full
  onboarding via `082_submit_domain_unified.sh` still can't complete end-to-end
  today — but content/design work can proceed in parallel.
- **People-imagery house style**: owner specified **line illustration**, not
  photography. This resolves Open Decision #2 outright and sidesteps the
  fabricated-person risk by construction — a line-illustration figure never reads
  as "a photo of someone who doesn't exist."
- **leopardessconsulting content is confirmed factually correct** by the owner
  directly ("the content we have in leopardessconsulting.co.uk (in the framework)
  is all factually correct"). Consistent with the claims-verification workstream
  memory (V0-V5 live, 2,767 verified records) — this means leopardess's own
  case-study/capability content can be reused/cross-referenced as grounded source
  material for fundamentallyai's positioning, not just as prior-art precedent.

**Workflow `wf_51d0513a-4d5` hit an account-wide session limit mid-run** (81/101
sub-agents completed; the final synthesis agent and most 3-vote verification
agents failed with "session limit, resets 6:20pm Europe/London"). **The
aggregated `findings`/`confirmed` fields the tool returned are therefore
incomplete** — read `journal.jsonl` directly (per the tool's own guidance) and
pulled the raw per-agent claims that never made it into a synthesised report.
Treat these as single-source research claims (search-agent findings, not full
3-vote adversarial verification) except where marked otherwise — good enough to
design from, not to publish as fact without a second check if it ever became a
site claim itself (it won't — this is competitor research, not a claim about
our own platform).

**Named-site specifics (supplements the earlier direct-fetch Bain catalogue
already in this file):**
- **Bain**: hero slider/carousel **with embedded video** (not just static images
  — a data point against a video-free assumption; Bain evidently accepts the
  heavier cost there), sticky header with mega menu, hamburger nav on mobile.
- **BCG**: leads with **heavy data-visualisation/interactive charts** as the
  primary homepage device — signalling analytical capability through
  motion/interactivity, not just imagery; its insight/blog carousel is
  secondary. Strong signal that our proposed **code-rendered stat/chart
  component** (shared build with leopardess's outstanding L7, see PLAN Open
  Decisions #4) should be a first-class *homepage* element for an AI/strategy
  consultancy brand, not an afterthought.
- **McKinsey**: two-column hero using **original artwork/illustration, not stock
  photography**, in the hero specifically; clean editorial grid, blue-and-white
  palette; insights/reports foregrounded over a plain services list. Brand
  system (Wolff Olins, rolled out Feb 2019): commissioned custom typefaces
  (Bower for headings, McKinsey Sans for body), **one uniform photographic
  treatment applied across all imagery — greyscale plus a blue/purple tint at
  the edges** (duotone-style, not bespoke per-image direction), a purpose-built
  data-viz system, a recurring blue line-pattern graphic motif substituting for
  photography in places, and a custom minimalist monochrome icon set.
  **This duotone/tint approach is directly transferable to our own
  line-illustration decision** — one consistent tint/treatment applied across
  every line illustration gives the same fleet-wide visual cohesion McKinsey
  gets from photography, cheaply, and it was already illustration before the
  tint, so it doubly avoids the fabricated-person problem.
- **Accenture**: hover-reveal cards (image+headline box, extra info slides in
  on hover — a sibling of our hover-zoom ask, revealing text rather than
  zooming the image) and an "ecosystem" grid layering industries/capabilities/
  case-studies/thought-leadership in one scroll.
- **KPMG**: brand-colour (purple/lavender) tint on photography — another
  duotone example.
- **EY**: subtle animation (not video), large hero imagery with compelling
  titles, content personalised to the visitor's presumed journey.
- **Stat bands with animated count-up figures** are a named recurring pattern
  (example cited: RHR International). Per the standing rule inherited from
  leopardess: an animated count-up of a REAL number read from the evidence base
  is fine; a rounded-up or invented number is not — same discipline as the
  chart component.

**Award-gallery correction (saves re-searching this later):** neither the
current CSS Design Awards Website-of-the-Day roster nor the Awwwards Business &
Services gallery (both checked as of July 2026) features Bain, BCG, McKinsey,
Deloitte, Accenture, or KPMG at all — those galleries reward studio portfolios,
brand microsites, and campaign sites. **Comparable award-gallery-grade
inspiration, if wanted beyond the three owner-named references, sits with
agency/tech sites**: Obys (obys.agency), Zentry, Igloo Inc, Wembi, Microsoft AI
— not the consultancies themselves. Also, neither gallery documents
component-level implementation (they tag technique categories only —
Scrolling/Animation/Microinteractions/Transitions) — a source of exemplar names
to look at directly, not a ready-made teardown.

**Hover-zoom card — concrete implementation recipe** (the owner's "fancy image
that slightly enlarges" ask, now with an exact recipe):
- Wrapping container: `overflow: hidden`, fixed `aspect-ratio`; image at 100%
  width/height, `object-fit: cover` (the clip is required — the scaled image
  would otherwise spill past the container).
- Transition the **standalone `scale` property** (not the legacy `transform`
  shorthand) roughly `1 → 1.1` over ~250ms on hover.
- Required accessibility pairing: `prefers-reduced-motion: reduce` zeroes the
  transition, and `@media (hover: none)` disables the effect on touch devices
  (a phone tap can't "hover" — it should simply not apply there, not get stuck
  mid-zoom).

**Swipeable mobile carousel — concrete implementation recipe:**
- **CSS `scroll-snap` alone** (`scroll-snap-type: x mandatory` on the
  container, `scroll-snap-align: start` on each card) gives the swipe-and-settle
  behaviour with **zero JavaScript** — no `touchmove` listeners, no
  `requestAnimationFrame` loop, the browser does the work. Right default for
  "swipeable left/right on mobile."
- Semantic markup: `<ul>/<li>`, native `<button>` controls for any prev/next
  arrows, so a screen reader announces it as a list, not `<div>` soup.

**Auto-advancing hero carousel — this one genuinely needs JS, with real
accessibility obligations (WCAG 2.2.2 + ARIA APG), not optional polish:**
- Visible pause/stop control, label changes to reflect the next action.
- **Must** stop on keyboard focus entering the carousel and on mouse hover;
  must not silently resume.
- Correct ARIA: container `role="region"`/`role="group"` + accessible name +
  `aria-roledescription="carousel"`; each slide `role="group"` +
  `aria-roledescription="slide"` + name; off-screen slides `aria-hidden`, not
  just visually hidden.
- Controls precede the slides in DOM order for tab order; respect
  `prefers-reduced-motion`.
- **Real usability caveat, not just an accessibility footnote**: cited research
  says only ~1% of users interact with a carousel at all, and 89% of those only
  ever see the first slide; WebAIM's position is to avoid the pattern
  altogether for exactly this reason. **Design implication**: the first card
  must carry the complete message on its own — a hero carousel is "a rotating
  hero that happens to have more cards," never a place to hide anything
  essential on slide 2+.

**Lazy-loading / perf, for card-grid and carousel images:** never lazy-load the
hero/LCP image (eager-load it); set explicit width/height or `aspect-ratio` on
every image to reserve layout space; a ~500px `rootMargin` buffer before
viewport entry avoids visible pop-in; Chrome's own measured numbers show 97.5%
of lazy images fully load within 10ms of becoming visible on 4G — the
lazy-loading perf cost is genuinely negligible when done correctly.

**People-imagery licensing** (lower priority now line-illustration is decided,
kept on record for if photography is ever used elsewhere in the fleet): stock
licensing splits Royalty-Free / Rights-Managed / Editorial-Use-Only (prohibits
commercial use) / Creative Commons; commercial use of an identifiable real
person's photo needs a signed model release. Doesn't apply to line
illustration, which sidesteps licensing and fabrication both by construction.

## 2026-07-20 (continued) — broadened ask: internal capability/case-study research

Owner has broadened the brief significantly: fundamentallyai.com shouldn't just
have nice-looking generic consultancy components — it should **market this
platform's own real, true capabilities** as service lines/case studies (embeddings-
based private in-house search for a partner without leaking data outside their
org, instant marketing/product-test/presentation sites, our finetuning
capability, our council/multi-agent review decision-making, backend engineering
proof-points like idea.uk's Stripe integration and the relojistas.com expired-
domain traffic work). Explicit instruction: **get a decent amount of TRUE facts
before writing any content** — this is a claims-verification-constrained brand,
same as leopardess, so nothing ships that isn't grounded.

Dispatched an Explore agent (background, very-thorough breadth) to build a
7-part capability inventory: embeddings/RAG (real or aspirational — a quick
grep already found `platform/database/pgvector.go` and
`platform/orchestration/actions/rag_actions.go`, so there is at least SOMETHING
real here, scope TBC from the agent's findings), fine-tuning flywheel, the
council/review system's track record, idea.uk's Stripe integration, the
relojistas.com traffic-testing result, other genuinely demonstrable capabilities
(claims-verification itself as a meta case study, the site-generation pipeline
itself, imagery pipeline, multi-session coordination engineering), and a census
of how many real sites are live today. Each finding tagged
[LIVE/VERIFIED] / [BUILT BUT INERT] / [ASPIRATIONAL] — only the first two
categories (clearly labelled) are usable as honest marketing content. Results
not yet returned as of this note.

**Scope note for this workstream, stated explicitly rather than assumed**: the
deliverable from this research is a grounded FACT BASE / content brief per
proposed page or section (bullet facts, evidence citations) — not hand-written
final marketing copy. Final page copy still goes through the framework's
content-writer agent (from `site_specs`/mission/briefing), same as every other
site — matching the owner's own stated preference on the leopardess voice
rewrite ("I don't want it written here manually"). This session's job is to make
sure that when the content-writer runs, it has true, well-evidenced material to
draw from and a clear brief of which components/case-studies to write for.

## 2026-07-20 (continued) — capability inventory returned; one load-bearing correction

The Explore agent's 7-part capability inventory (methodology: code + all of
`docs024_key_docs_latest/` + `bugs_open/`/`bugs_closed/`/`features_open/` +
`git log`, every claim tagged LIVE/VERIFIED, BUILT BUT INERT, or ASPIRATIONAL
with file:line/doc citations) returned. Condensed here; full per-claim evidence
is in the agent's report (not re-transcribed in full — this file records the
findings that change what we can honestly say, not a duplicate of the source
material).

### The one finding that changes the brief itself

**[LIVE/VERIFIED] Real, production pgvector-based RAG infrastructure exists**:
`knowledge_base` table (`vector(768)`, IVFFlat cosine index + trigram
fallback), `code_symbols` (source-code semantic search, HNSW), `agent_memory`
(`platform/database/pgvector.go`, per-client-schema). `rag_actions.go`
implements index/lookup with Nomic task-prefixing, proven with a real
discriminating-content test (French Bulldog BOAS ranked correctly above
Labrador/piano/EV-battery content) and a live chassis smoke test. It's used for
real work today (tool-generation PLANs indexed into a `tool_docs` collection).
**Genuinely strong, verifiable infrastructure — this part of the owner's
embeddings idea is real, not speculative.**

**But [NOT TRUE TODAY, per the platform's OWN prior audit]: there is no
tenant/client data isolation on this shared store.** `docs/leopardessconsulting/
AUDIT_verified_facts.md` §4 finding P5 (code-verified, not inferred): *"Single
shared Postgres (no row-level security anywhere in the schema), single shared
Kafka cluster, single shared ollama-adapter pod — separated only by a `site_id`
column in shared tables."* `knowledge_base` is explicitly documented as a
**shared** resource, not per-tenant. A fine-tuning planning doc even lists
"add `tenant_id` to `knowledge_base` + enforce in `rag_lookup`/`rag_index`" as
an **unstarted** Week-1 TODO.

**This directly constrains the owner's own pitch** ("use embeddings to safely
let them search their in-house databases without leaking info to outside
organisations"): **that specific claim — as a standing, already-shipped
guarantee — would be an overclaim of exactly the kind the claims-verification
system exists to catch.** The honest, still-compelling version: *we have real,
production-proven vector search/retrieval infrastructure (running today across
our own 11 sites and our own code search) — and building a client a properly
isolated, private instance of it (adding the tenant boundary that our own
shared testbed doesn't have) is a scoped, buildable engagement, not a
speculative one.* "Buildable because we've already solved the hard technical
part" is a true and still strong claim; "we already do this safely for
multiple outside parties" is not. **Flagging this to the owner as a required
framing decision before any copy is written — see PLAN.**

### Fine-tuning

**[LIVE/VERIFIED]** one real completed LoRA fine-tune: Llama-3.3-70B via
Unsloth QLoRA on 1,958 real rows exported from the platform's own
`llm_call_log` (every real LLM call the platform makes is logged — the
data-capture "flywheel" mechanism is itself real and live), ~9h/~$20 on a
rented A100. **Genuinely evaluated, not just claimed**: held-out real briefs,
automated + Claude-judged blind A/B, honestly reported non-flattering result
(Claude won 16/20 vs the fine-tune's 4/20), verdict recorded as *"shippable for
low-stakes use, not client-facing."* Total real cost for the cycle: ~$22.
**[BUILT BUT INERT]**: the fully-automated unattended version (auto-provision →
train → evaluate → redeploy without a human watching) has real engineering
behind it (a working Thunder Compute GPU-provisioning adapter) but has **never
completed one full unattended cycle** — the decommission/monitor branch "has
never fired live." Framing: cite the real fine-tune + honest evaluation
methodology as the proof point; do not claim the automated flywheel is running
unattended today.

### Council-gate / multi-agent review — the strongest, cleanest pillar

**[LIVE/VERIFIED]**, and the single most differentiated, quotable capability
found. **13 seats** (code-verified as of 2026-07-19 — not 16; that figure
belongs to a *different* council, the concept-register one, per this session's
own memory — don't conflate the two when citing a seat count), live since
2026-07-17, growing on a documented, independently-verified trajectory
(2→6→7→9→13). Real decision record: 18 commits carry a `Council-Reviewed:`
trailer; a real external submission (imagery D14) took 3 rounds (6/3→8/1→7/2)
and caught 3 real defects including an unguarded `jsonb_array_length()` call
that could **abort an entire discovery sweep in production** — a genuine
production-risk catch with commit hashes and a correlation ID as the audit
trail. **Self-correcting culture on record**: a commit literally titled
"CORRECTION — the Council-Reviewed trailer... was not earned." Baseline
adoption metric tracked honestly (28 in-scope commits/3 days, 0 reviewed, on
first measurement). **This is strong enough to be the flagship pillar** — an
AI platform whose changes are independently reviewed by other AI agents before
they ship, with a real, growing, self-correcting decision record, is a
genuinely rare and verifiable claim.

### idea.uk (Stripe) and relojistas.com — both usable, one with a caveat

**idea.uk [LIVE/VERIFIED]**: real hand-rolled Stripe integration (direct REST
calls to the Checkout Sessions API, not the SDK; HMAC webhook signature
verification implemented by hand; idempotent webhook processing; a
human-review gate before auto-delivery) selling a real £29 report product, live
on production nginx routing tuned specifically for Stripe's retry behaviour.
**Caveat**: the docs themselves flag that the final "a real Stripe test event
reaches a paid order" verification step was still outstanding as of the most
recent dated note — cite the integration as real and live, do **not** cite a
specific transaction volume (unconfirmed).

**relojistas.com [LIVE/VERIFIED] — the cleanest, most quotable case study
found.** A dead Spanish watch-forum domain, measured (not guessed) to still be
receiving real subscriber traffic to one specific legacy RSS feed URL (~136
hits/day, 100% failing). Rebuilt as a live Spanish watch-news portal; the
legacy feed URL flipped from 100% failure to **~97% success within 24 hours of
launch** (122/125 the first full day). Honestly caveated in the same doc: most
traffic is crawlers, ~55 non-crawler fetches identified, Cloudflare-fronted IPs
mean genuine subscriber counting isn't possible — **the honesty of the caveat
is itself part of what makes this evidence-grade**, not a weakness to hide.

### Other real, demonstrable capabilities (condensed; full table + citations in
the agent's report)

- **Claims-verification system itself** — a real anti-hallucination pipeline
  that checks generated site copy against a per-site verified-evidence base,
  live, catching real fabrications within hours of sites going live. Strong
  meta-narrative potential (see PLAN — this one needs an explicit owner call,
  not an assumption, because it involves referencing that a past mistake
  happened).
- **Voice-tells checker** — deterministic AI-prose-tell detector, live,
  calibration verified against real pages.
- **Site-generation speed** — a real, dated (2026-07-10), verified-against-
  production example: a 33-page site rebuilt from scratch overnight, largely
  unattended, including a live 9-source news feed and 5 interactive
  calculators. **Directly answers the owner's "instant marketing/product-test/
  presentation sites" idea** — this is the proof point for that pillar, not a
  hypothetical.
- **Imagery pipeline** — 14 bespoke hero images generated overnight, ~90
  seconds each, prompt→model→optimise→git-commit, fully automated.
- **Self-healing discovery loops** — 14 corrupted components found fleet-wide,
  10 healed within a day, at least one with zero human involvement.
- **Multi-session coordination engineering** — real infrastructure built so
  many autonomous AI sessions can safely share one production codebase
  (commit-per-task discipline, ref-pinned builds from committed HEAD). Found a
  real production bug in the process (Kafka at-most-once consumption wedging
  orchestrations up to 1,224 hours) — cite the coordination discipline as live;
  the underlying bug it found (`bugs_open/003`) is filed with a fix specified
  but **not yet shipped**, so don't claim that specific bug as "fixed."

### The load-bearing exclusion list — never reuse these for fundamentallyai

**[CRITICAL — read before any copy is written for this site]** Leopardess's own
`AUDIT_verified_facts.md` documents specific fabrications a past thread found
and stripped from leopardessconsulting.co.uk on 2026-07-09. These exact figures
must **never** resurface anywhere, including on fundamentallyai.com, even
though an LLM asked to write "about us" copy might reproduce them from
training-adjacent patterns or stale cached context:
- "70+ agents across 8 functional departments" (a fabricated org taxonomy —
  `information_schema.columns WHERE column_name ILIKE '%department%'` returns 0
  rows; the true, verified figure is over 150 agent definitions, no
  departments).
- Any invented founder/leadership bio (a "Peter Grenfell" headshot/bio was
  invented and deleted; the real background is the owner's own, first-person).
- Invented case-study titles/clients (leopardess's own subsystems were
  relabelled as third-party client engagements — the platform's own standing
  rule, worth repeating for fundamentallyai: **our own sites demonstrate the
  platform; they are never to be implied to be a client roster.** Any of our
  11 real sites cited on fundamentallyai must be labelled as our own build.)
- Fabricated stats: "99.9% uptime" (invented), "2,767 Awards Won" (a garbled,
  meaningless figure).
- **Owner has since (2026-07-20) explicitly approved naming
  leopardessconsulting.co.uk directly**, specifically as the worked example
  for the self-correction/claims-verification case study — the point being
  transparency: name the real site, tell the real (audit-documented) story of
  what was fabricated and how it was caught and fixed. The exclusion list
  above still applies in full: the *specific fabricated figures themselves*
  ("70+ agents/8 departments", the invented founder, invented case studies,
  the fake uptime/awards figures) are what must never resurface as if true —
  naming the site is now fine; restating the fabrication as fact is not. Any
  copy telling this story must be built strictly from
  `AUDIT_verified_facts.md`'s actual findings/dates/fixes, not embellished.

### 2026-07-20 — fundamentallyai.com onboarding triggered; queued behind a known platform bottleneck

Owner confirmed hosting is being pointed (DNS still showed Afternic's parking
nameservers at time of check — not yet a problem, propagation can take up to
48h and doesn't block onboarding). Fired
`bash 082_submit_domain_unified.sh fundamentallyai.com --email
fundamentallyai@contactforsales.com --mission-file
MISSION_BRIEF_fundamentallyai_2026-07-20.md` after owner review of the mission
brief (same file, committed separately).

**Correlation:** `099ca178-92fc-41ac-bf6c-bc17c0aa6ec6` · **Orchestration:**
`c6a53a35-f479-48cc-9845-067cd4a729d2` · orchestration_name
`submit-fundamentallyai.com-20260720-202254`.

Verified by artifact, not by the script's own printed success message
(`sites`/`orchestration_states` both showed 0 rows immediately after firing,
which is exactly the "looks dropped" symptom `bugs_open/030` warns about — did
NOT conclude drop or retry): consumed the raw Kafka topic directly and
confirmed the message landed at offset 96081. Checked consumer lag for
`generic-requests-group` on `system.agent.generic.requests`: current offset
96024 vs log-end 96151 — **127 messages behind**, our message not yet reached.
This is precisely `bugs_open/030` (single-partition, single-consumer,
`replicas:1`, no HPA dispatch topology — a platform-wide, already-filed,
OPEN issue, not specific to this trigger). Per that bug's own explicit
warning: **do not retry** (duplicates the work and spends credits twice) —
the message WILL be processed, on a timeline that bug's own measurements put
anywhere from ~25 minutes to several hours depending on concurrent session
load, not something to predict precisely.

**Next-session/continuation note**: if `sites`/`site_specs`/`orchestration_states`
still show nothing for `fundamentallyai.com` when this is picked up again, check
`bugs_open/030`'s "one command" (consumer-group lag) before assuming anything
failed — this is a known, expected wait, not a new problem.

### Real site count

**11 live sites** (two independent code-checked counts converge, 2026-07-19/20):
dartsonline.com, finetuning.uk, gamesdesign.co.uk, idea.uk,
leopardessconsulting.co.uk, robot-hands.com, vonc.com,
ai-agent-orchestration.com, gaswholesalers.com, vetcomparison.uk,
relojistas.com. (`wayfaringlondoner.com` is an experimental lander, not a full
site — exclude it from any "11 sites" count; `worldsoccernews.com` is an
owner-personal reference, not a platform-built site — never cite it as ours.)
"We operate 11 live production sites on our own platform" is a true, checkable
headline stat — but per the exclusion list above, don't enumerate
leopardessconsulting by name without sign-off; a rounded framing ("content
sites, an interactive game platform, a paid tool with real payments, a revived
expired domain") works without naming every one.
> **UPDATE 2026-07-20 (later): owner approved naming leopardessconsulting
> directly** (see the exclusion-list update above and PLAN Open Decisions #8).
> The "name it" approval is specifically for the self-correction case study.

## 2026-07-21 — the onboarding RAN; site built overnight; two blockers to live

All [VERIFIED] against the live DB / live HTTP on 2026-07-21, on chassis
**v1.0.1144** (image tag confirmed on running pod `agent-chassis-59c675c4f-pxr9f`
and on the deployment — the fresh build the owner flagged is live). Full
readable write-up: `HANDOFF_2026-07-21_start_here.md` (the new cold-start entry
point). Condensed evidence here:

- **The queued trigger survived and completed.** Domain-submitter orchestration
  `c6a53a35-...` (corr `099ca178-...`) COMPLETED 2026-07-20 20:36. The
  `bugs_open/030` queue wait resolved on its own, exactly as that bug says it
  would — **the decision NOT to retry was correct**; a retry would have been a
  wasted duplicate.
- **The whole spec cascade completed overnight** (verified via `site_specs`):
  submission → mission_brief → identity → classification → content_direction →
  design_intent → vertical_landscape → strategy → briefing →
  resolved_composition, all `is_current=t`. Note a step not in the CLAUDE.md-era
  mental model: a `vertical-exemplar-researcher` agent ran between classifier
  and strategist (aspect `vertical_landscape`) — the pipeline has a
  best-in-class-exemplar research step now.
- **The mission brief propagated faithfully** — this is the headline. Verified
  `design_intent.imagery_direction` reads, verbatim: *"Line illustration only
  for any human or figurative element — never photography of real or generated
  individuals. All illustrations treated with a single consistent tint (navy or
  amber overlay...) ... Charts are real, code-generated from verified data ...
  No decorative data visualisation."* Palette resolved to a dark consultancy
  navy/amber (`primary #0E1B2E`, `accent #C8902A`, `background #090F1A`). Pages
  created by name include `multi-agent-review-council`, `model-fine-tuning`,
  `self-correction-leopardessconsulting`, `capabilities`, `platform-log-index`,
  `tool-decision-record` — i.e. our chosen pillars became real pages. Every
  owner decision from 2026-07-20 shows up in the machine's output.
- **Page build states**: `contact` + `model-fine-tuning` = `deployed` (DB);
  `index` + `about` + `capabilities` + `multi-agent-review-council` =
  `needs_rebuild`; `platform-log-index` + `self-correction-leopardessconsulting`
  + `tool-decision-record` = `planned`.
- **Blocker 1 — content-validation gate.** 5 content pages sit at
  `needs_human_review` with `validate_content failed: ... content validation
  failed: 1 blockers, 0 errors` — one blocker each, consistent, but `contact`/
  `model-fine-tuning` passed, so it's content-specific not universal. **The
  blocker reason is NOT in the DB** (`site_work_items.result` jsonb is empty)
  and **the overnight logs rotated on the v1.0.1144 restart** — so it is
  currently [UNRECOVERABLE without a live re-fire]. Next thread: re-fire one
  blocked page and capture the blocker from the chassis log during
  `validate_page_content`. **Do NOT assume the cause** — candidate hypotheses
  (leopardess-class `contact-block` placeholder-email false positive; a
  claims/banned-language blocker legitimately firing on honesty-heavy copy) are
  UNVERIFIED. Also 2 pages failed differently: `page-build-handler no-op: no
  sections ready to build (empty spec sections)` — a planner-level empty-section
  issue, separate from the validation gate.
- **Blocker 2 — nothing serves.** DNS has propagated to Cloudflare (NS now
  `alexis`/`leah.ns.cloudflare.com`, matching the live fleet). But
  `https://fundamentallyai.com` returns a Cloudflare **404** at root, `/contact`
  times out (000), `/model-fine-tuning` 404s — i.e. even the DB-"deployed" pages
  don't serve. `sites.github_repo`/bucket empty, but **that's normal** —
  robot-hands.com and leopardessconsulting.co.uk are both live with those empty
  (the B2 fleet deploys per-domain dirs via a *separate portfolio repo* the
  git-adapter pushes to + `deploy-to-b2.yml`, NOT via `sites.github_repo`;
  neither `fundamentallyai.com/` nor `robot-hands.com/` exists in THIS repo,
  confirming the deploy dirs live elsewhere). So the serving gap is [INFERRED,
  NOT YET DIAGNOSED]: either the git-adapter hasn't pushed rendered pages to the
  portfolio deploy repo, or the new Cloudflare zone's origin→B2 wiring isn't set
  up (likely a per-domain infra/owner step, as idea.uk's cutover was). Flagged
  as next-thread action #2, not solved this session.
- **2 `needs_section_data` items** (correctly asking for real data, not
  inventing it): `portfolio-showcase` on index needs real project data;
  `contact-info` needs a business email. Feed our own 11 real sites (honestly
  labelled as ours) + a real address.
- **The fancy components still don't exist** — the pipeline built this site from
  the *existing* standard section components. The carousel/hover-zoom/swipeable
  components are still a from-scratch build (Thread B in the handoff). Today's
  site proves the *content/positioning* half; the *visual-components* half is
  the remaining original ask.

## 2026-07-21 — Blocker 1 DIAGNOSED (no re-fire needed); filed bugs_open/055

**Root cause CONFIRMED, self-evidencing** — every one of the 5 blocked content
pages fails on the **same single blocker**: the content-validation gate's
cross-site contamination check (`checkDomainContamination`,
`platform/orchestration/actions/validate_page_content.go:481-534`) flags
`leopardessconsulting.co.uk` (in its hardcoded `knownSites` list, line 491) as a
`blocker` (line 508) when it appears in fundamentallyai's copy. But naming
leopardessconsulting.co.uk is the **entire, owner-approved point** of the
self-correction story (MISSION_BRIEF lines 46-49). The anti-leakage check has
**no per-site allowlist** — it assumes no site ever legitimately references
another of ours. False for a portfolio/meta site, which fundamentallyai is by
design. Full write-up + fix candidates: `bugs_open/055`.

> **CORRECTED 2026-07-21 — the handoff's premise was WRONG, and it cost nothing
> because I checked the code first.** `HANDOFF_2026-07-21_start_here.md` (and the
> 07-21 SUMMARY/README) stated the blocker reason was "NOT recoverable from the
> DB" (the `site_work_items.result` jsonb is empty) and prescribed re-firing a
> page live on v1.0.1144 and watching the chassis log during
> `validate_page_content` to capture it — a ~30-min single-queue wait
> (`bugs_open/030`) plus a live log-tail. **That was unnecessary.** Reading
> `validate_page_content.go` first showed it *deliberately persists* the full
> structured issue list to `agent_error_log` (`writeValidationFailureLog`,
> lines 344-420, `error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`) exactly so a
> post-mortem never needs pod logs. **9 such rows** were already in the DB, each
> naming the exact blocker. What caught it: reading the action's code end-to-end
> before acting on the handoff's "unrecoverable" claim — a good instance of
> "read the function before you assume." Logged in WRONG_CALLS.

**Evidence [VERIFIED live DB, chassis v1.0.1144, 2026-07-21]:**
- 9 × `agent_error_log` rows, `error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'`,
  every one a single `cross_site_domain` / `leopardessconsulting.co.uk` issue,
  `location` snippet = the self-correction copy verbatim ("invented details on
  leopardessconsulting.co.uk. Our verification system flagged it; we corrected
  it."). Query in `bugs_open/055`.
- **Second-order damage:** `count(*) FILTER (WHERE rendered_html ILIKE
  '%leopardess%')` = **0** across all 9 pages' saved `page_components`. The 2
  pages that reached `deployed` (`contact`, `model-fine-tuning`) did so by
  *dropping* the leopardess story on retry; the pages that kept it stayed
  blocked. So the flagship narrative is **absent from the site**, not merely
  held back — a degraded page passed looking "done". This is the "trust the
  artefact, not the status" rule biting exactly as CLAUDE.md warns.

**Blocker 2 (serving) quick-scope [VERIFIED, with a self-correction]:**
fundamentallyai's clients_db config is **identical to robot-hands.com** (a
known-live B2 site): both have empty `github_repo`, `status='deployed'`,
`network_id=…002`. So **nothing in the app DB explains why one serves and the
other doesn't** — the gap is entirely in the external serving path (the
per-domain rendered dir in the separate *portfolio deploy repo* the git-adapter
pushes to, and/or the Cloudflare zone → B2 origin wiring for the new domain).
Per-domain infra/owner step, as idea.uk's cutover was.
> **CORRECTED 2026-07-21:** I first inferred "empty
> `page_components.deploy_commit` on the deployed pages ⇒ git-adapter never
> pushed". **Wrong** — robot-hands.com serves live with `deploy_commit` empty
> on all 87 of its components, so it is not the signal. Caught by comparing
> against a known-live site rather than reading one site's column in isolation
> (the recurring lesson: a value only means something relative to a working
> control). Deferred behind Blocker 1 regardless — nothing worth serving until
> the pages rebuild *with* the story; re-check serving after the fix, then
> whatever remains is the pure Cloudflare/B2 owner step.

**Fix direction (see 055 for candidates):** per-site allowlist of referenced
domains, consulted by `checkDomainContamination`, opt-in (absent allowlist →
unchanged behaviour, zero regression for the other 10 sites). Touches
`platform/` → council-review before commit; Go change → inert until image roll.
Next: design the fix, put it through the council gate, build+roll, seed
fundamentallyai's allowlist, re-fire ALL content pages (incl. re-building the 2
generic "deployed" ones so the story is restored), verify the leopardess
reference is present in saved `rendered_html`.

## 2026-07-21 — fix written + tested; council round 1 = REVISE; resubmitted round 2

**Owner decisions (2026-07-21):** image roll = "roll it once approved" (durable
authorization to build+roll the fleet on an APPROVED verdict, quiet-check first);
allowlist scope = the full intended portfolio (leopardessconsulting.co.uk,
finetuning.uk, idea.uk, relojistas.com). NB only leopardess + finetuning.uk are
in the checker's hardcoded knownSites, so those two are the ones that actually
suppress anything; idea.uk + relojistas.com are harmless future-proofing.

**Fix (candidate 1, built):** `checkDomainContamination` gains a per-site
allowlist loaded from `sites.content_data->'allowed_reference_domains'`
(`loadAllowedReferenceDomains`, fail-closed to nil). Unit tests pass
(`TestContamination_*`): no allowlist → leopardess flagged blocker; allowlisted →
leopardess + company suppressed, un-listed dartsonline.com still flagged.
`go vet` clean (the one warning is pre-existing in another file). Single caller
confirmed (one: `ValidatePageContentAction`). **Code held UNCOMMITTED** pending
an APPROVED verdict so it lands with the `Council-Reviewed:` trailer — safe to
hold because the change is INERT until the allowlist is seeded (which I control),
so there's no prod risk and no urgency.

**Council round 1 (corr `03908b72-2471-474e-baaf-7952d1903460`) = REVISE**, back
in ~15 min (faster than the ~30 CLAUDE.md budgets). **5 approve / 3 object.** The
code itself was NOT vetoed — constitution/mission/guardian/render_guardian/
prior_art called it sound, minimal, opt-in, fail-closed, reuse-respecting
(`loadSiteContactEmail`/`loadEvidenceBase` pattern). The 3 objections
(editquality, bug_historian, debug_historian) were all **plan COMPLETENESS**, and
all fair:
1. The plan fixes the guard but never flips the switch — no edit performs the
   config seed (editquality, med).
2. The prod jsonb UPDATE needs needle-gate discipline — backup/idempotent
   guard/RETURNING/verify+rollback (debug_historian, med).
3. No deploy-verification (pod-grep the running binary for the new symbol)
   (debug_historian, med).
4. The 5 stuck + 2 degraded pages need re-queue + REGENERATION (not a
   rendered_html SQL patch) or the incident doesn't close (editquality +
   debug_historian).
5. **The deeper defect:** the content-writer/regeneration path *silently drops
   flagged content on retry* with no record — that's what actually erased the
   story, and it's generic; patching this one trigger leaves it exploitable
   (bug_historian, med). **→ filed as `bugs_open/056` (mechanism [INFERRED],
   needs-diagnosis) and explicitly scoped OUT of the 055 plan.**
6. Minor (both low, handled): company-suppression rationale slightly over-reaches
   the domain-only evidence (it's deliberate — same first-party site, forward-
   proofs a company-name rewrite); verify single caller (done).

**Round 2 resubmitted** (`RESUBMIT_CORR=03908b72`, new run
`6c385a7f-46ae-4f2f-9651-2cfa458ddb72`), addressing every objection: added the
needle-gated seed script (`sql/055_seed_allowlist.sql` + `_VERIFY`/`_ROLLBACK`
sidecars) as an explicit edit; spelled out the full rollout (build→roll→pod-grep
verify→seed→regenerate→verify-by-artefact); tightened the company-suppression
rationale + code comment; filed and scoped-out 056. Awaiting round-2 verdict
(monitor armed). **A REVISE is a success, not a waste** — the council caught that
"fix the mechanism" ≠ "close the incident", which is exactly the gap that leaves
sites half-live.

## 2026-07-22 — CORRECTION: the fix was already LIVE; only the seed was needed; seed done

> **CORRECTED 2026-07-22 — most of the round-1/round-2 fix effort was redundant.**
> The owner said "double check your findings, v1.0.1146 is on production." That
> made me pod-grep the running chassis instead of INFERRING from timestamps. The
> live v1.0.1146 pod (agent-chassis-687cdf6db5-fq2fd, running since 2026-07-21
> 18:50 UTC) **already contains the full allowlist implementation** —
> `loadAllowedReferenceDomains` (2), `checkDomainContamination` (2),
> `allowed_reference_domains` literal (3). `git log -S allowed_reference_domains`
> shows it was introduced by `fe2ba5e52` ("v1.0.1146 sweep"), byte-identical to
> what I wrote. My commit `f2780e1bd` changed only a COMMENT in that file (+ a
> test + the seed). **No build/roll was ever needed** — I nearly cut an
> unnecessary fleet image roll (caught only by the owner). Logged in WRONG_CALLS.
> The lesson: grep the current code + pod-grep the live binary for the mechanism
> BEFORE writing OR rolling a fix — "is it already done?" precedes "how do I do
> it?".

**What was actually missing — and is now done:** the DATA seed. The live code
reads `sites.content_data->'allowed_reference_domains'`; fundamentallyai's was
absent → loader returns nil → contamination still fired. Ran
`sql/055_seed_allowlist.sql` against prod 2026-07-22: PRE key_already_present=f →
UPDATE 1 → POST key_present=t, n=4, domains =
`["leopardessconsulting.co.uk","finetuning.uk","idea.uk","relojistas.com"]`,
COMMIT. [VERIFIED, live DB.] The live loader (byte-identical to fe2ba5e52's) reads
exactly this key/format, so the seed is now effective.

**Remaining:** regenerate the 5 stuck (needs_human_review) + 2 degraded
(deployed-but-storyless) content pages so they rebuild WITH the leopardess story
and pass validation (now that the allowlist skips it), then verify the reference
is present in saved `page_components.rendered_html` — trust the artefact, not the
status. Watch for the bugs_open/056 non-determinism (a regeneration may still omit
the reference by chance — verify, re-fire if so). The self-correction-
leopardessconsulting page is separately blocked (planned, 0 sections — empty spec
sections, a planner-level issue, not the contamination gate).

## 2026-07-22 — Missteps this session, the complete list (the record IS the point)

Per CLAUDE.md the missteps are not an appendix — they are the most valuable thing
the next thread cannot rederive. Collected here in full, each with what caught it
and the cheap check that would have prevented it. (Cross-filed in WRONG_CALLS
where the check-class is fleet-wide.)

1. **[BIG ONE] Re-derived and council-reviewed (×2) a fix whose code was ALREADY
   LIVE.** Claimed: the contamination check has no per-site allowlist, so I must
   WRITE one, council-review it, and roll a new chassis image. Actual: the full
   allowlist implementation (`loadAllowedReferenceDomains`,
   `checkDomainContamination(..., allowedRefs)`, the `allowed_reference_domains`
   read) was already committed in `fe2ba5e52` ("v1.0.1146 sweep") and running on
   production since 2026-07-21 18:50 UTC — byte-identical to what I wrote. My
   commit changed only a comment. **Caught by:** the owner ("double check your
   findings, v1.0.1146 is on production") → which finally made me pod-grep. **Cheap
   check:** `grep -rn 'allowed_reference_domains\|allowlist'
   platform/orchestration/actions/` + a pod-grep of the live binary, BEFORE
   writing a line. "Is it already built and live?" precedes "how do I build it?".
   I checked `/bugs_open` + `/bugs_closed` for the mechanism but never the CURRENT
   CODE or the LIVE POD.

2. **Inferred "the live 1146 pod lacks my fix" from timestamps instead of
   pod-grepping.** Claimed: pod built 18:50 07-21 < my commit 09:07 07-22, so the
   fix cannot be in it. Actual: the fix WAS in it (built by another session).
   **Caught by:** the owner's nudge → pod-grep found `loadAllowedReferenceDomains`
   in the binary. **Cheap check:** the pod-grep IS the authoritative test CLAUDE.md
   mandates ("verify against the running pod, never git, never the tag"). I knew
   the rule and inferred anyway — confidence/logic is not a substitute for the grep.

3. **Nearly cut an unnecessary fleet image roll off misstep #2.** I got as far as
   bumping the tag mentally to 1147 and building a clean HEAD archive to pre-flight
   a roll — for code that was already live. **Caught by:** the owner, before any
   push/deploy. **Cheap check:** same as #2. Momentum is not evidence; the roll was
   downstream of an unverified premise.

4. **"Blocker 2: empty `page_components.deploy_commit` ⇒ the git-adapter never
   pushed."** Actual: robot-hands.com (live and serving) also has deploy_commit
   empty on all 87 components — it is not the signal. **Caught by:** comparing to a
   known-live control. **Cheap check:** a column value means nothing without a
   working control; diff the new thing against a known-good one before reading a
   field in isolation.

5. **(Inherited from the handoff, caught early) "the blocker reason is
   unrecoverable from the DB."** The handoff checked only `site_work_items.result`
   (empty) and concluded the reason was lost to log rotation, prescribing a live
   re-fire. Actual: it was in `agent_error_log` (`CONTENT_VALIDATION_BLOCKER_DETAIL`)
   all along — a deliberate persistence path. **Caught by:** reading
   `validate_page_content.go` before acting on the handoff. **Cheap check:**
   `SELECT DISTINCT error_code FROM agent_error_log WHERE site_id=…` before
   declaring any reason "not in the DB"; a gate that returns a vague error usually
   logs the detail on purpose.

6. **Timeline/working-tree archaeology rabbit hole.** Spent several cycles trying
   to reconstruct how an 18:50-07-21 pod could contain code I "committed" at
   09:07-07-22, in a shared mutable working tree. **Lesson:** in a shared tree the
   LIVE POD and committed HEAD (`git log -S <symbol>`) are ground truth; my
   session's file-read and my own edits are NOT reliable indicators of what is
   real. Stop reconstructing the sequence; grep the authoritative state and move.

**The through-line:** every one of these is the same failure — trusting an
inference (a timestamp, an empty column, a vague error, my own edit) over a
cheap authoritative check (a pod-grep, a control comparison, one query, `git
log -S`). The single most expensive one (#1) would have cost one grep at the
very start. The owner caught the two that mattered; the loop/tools did not,
because I never ran the check that would have surfaced them.

## 2026-07-22 — seed PROVEN end-to-end on `about`; 3 remaining pages re-queued

**Re-fire mechanism [VERIFIED]:** the `build-pipeline-trigger` scheduled task
(every 120s) fires when a site has a `site_work_items` row with `status='triaged'`,
`pipeline='build'`, `attempt_count < max_attempts` (its `pre_query`). The stuck
pages sit at `needs_human_review` (the auto-loop won't re-claim that). Safe
re-queue = set the `needs_page:<page>` build item back to `triaged`,
`attempt_count=0`, clear `error`/`claimed_*` (same row → no dedup conflict). It
then rebuilds the page's content through the content-writer.

**`about` proof [VERIFIED by artefact, not status]:** re-queued item `7e03bcc8`
→ triaged. Outcome: work item `complete`, page `build_status=deployed`, and **2
`page_components` mention leopardess** — the actual rendered text is the
owner-approved narrative verbatim: *"We caught our own platform generating
invented details on leopardessconsulting.co.uk. Our verification system flagged
it; we corrected it. That is not a cautionary anecdote — it is what ha[ppens]…"*.
**Zero new `CONTENT_VALIDATION_BLOCKER_DETAIL` rows** since the seed. So: the
seeded allowlist makes the live guard skip leopardess for this site, the page
rebuilds, and the content-writer included the reference (no bug-056 drop this
time). The whole chain — seed → re-queue → rebuild → story on the page — works.

**Batch re-queued (same reset):** `needs_page:index`, `needs_page:capabilities`,
`needs_page:multi-agent-review-council` → triaged. Monitor watching all three to
terminal state + verifying leopardess in each. Watch for bug-056: a page could
rebuild WITHOUT the reference by chance (a non-deterministic generation) — that
would show as build=deployed but leopardess=0; re-queue again if so.

**Still separate / not fixed by the seed:**
- `self-correction-leopardessconsulting`, `platform-log-index`,
  `tool-decision-record` = `planned` with **0 sections** ("empty spec sections")
  — a planner-level issue, NOT the contamination gate. The self-correction page
  is the dedicated home for the story and needs its sections populated before it
  can build. Separate sub-task.
- 2 `needs_section_data`: `portfolio-showcase` on index (real project data),
  `contact-info` (business email).
- `contact` + `model-fine-tuning` are `deployed` but storyless (dropped the
  reference on the pre-seed retry) — optional rebuild to restore it.
- Blocker 2 (serving): Cloudflare→B2 origin for the new zone — owner/infra step.

## 2026-07-22 — the LAST MILE: site serving (owner), nav 404s removed, phone, portfolio

Owner deployed a fresh chassis image (v1.0.1149) and wired serving — **the site is
live and visible** (Blocker 2's infra half done by the owner). Then the last-mile
cleanup:

**Two 404s** — `platform-log-index` (page_type `section-index`, a listing with
nothing to list) and `tool-decision-record` (page_type `tool`) — linked from top +
bottom nav but never built (0 planned sections by type, can't be re-queued). Owner
chose **remove from nav**. Non-obvious mechanics discovered the hard way (full
misstep trail below):
- The rendered nav is NOT per-page (`pages.rendered_header/footer` are empty) and
  NOT re-derived from `v_navigation_pages` at render time. It lives in
  **`site_components`** — the shared `site-header`/`site-footer` chrome
  (`component_level='site'`), rendered once and reused by every page assemble.
- Fixing `pages.in_header/footer` + `site_nav_items.status` did nothing to the
  live pages, because the cached `site_components` chrome still held the links and
  an assemble-only rerender reuses it unchanged (→ deploy skipped, bytes
  identical).
- Fix that worked: **surgically strip the two links from `site_components` header
  + footer `rendered_html`** (regexp, backup, verify still_has_dead=f), THEN
  re-deploy each page so the changed chrome ships.
- The build-dispatch-loop was **stalled** (bugs_open/029/030 — 33 `page_rerender`
  items triaged, 1 complete in 30 min), so the proper `rerender-pages` agent
  route queued forever. Bypassed it with the **direct `049b_deploy_single_page.sh`**
  per page (assemble-only page-rerender orchestration) — these still queue but
  drained; result: **0 links to `/platform-log/` or `/tools/decision-record/`
  across all 6 live pages** [VERIFIED via cache-busted origin fetch].

**Phone:** `sites.phone` set to `+44 (0) 7934 524 911` (owner-provided).

**Portfolio showcase** (homepage `portfolio-showcase`, reads
`site_specs.portfolio.projects`): published an owner-approved 3-project spec
(relojistas / idea.uk / leopardess), every claim grounded, `build_time="rebuilt
same day"` (owner's wording). Resolved the `needs_section_data` item, re-queued
`needs_page:index` — will render when the stalled build queue reaches it.

**Residual (pre-existing, NOT the nav 404s):** the `multi-agent-review-council`
`info-card-grid` has a "Review a sample record" card linking to
`/multi-agent-review-council#decision-record` — a no-`.html` internal URL that
404s (verified: `/multi-agent-review-council` → 404). A content-quality bug from
when the page was built, unrelated to the nav removal; left for a content pass.

> **Missteps this last-mile stretch (the record is the point):**
> 1. Assumed the nav came from `v_navigation_pages` (pages.in_header/footer) —
>    changed those + `site_nav_items`, re-deployed `about`, nothing happened.
>    WRONG: the nav lives in cached `site_components` chrome. Caught by checking
>    `last-modified` didn't move (deploy skipped) then finding the chrome table.
> 2. Assumed the assemble-only 049b rerender would pick up nav changes — it
>    reuses cached chrome, so it skipped. Only changing the chrome itself made a
>    rerender not-skip.
> 3. My dead-link detector regexp `(platform-log|decision-record)` false-positived
>    on the same-page anchor `#decision-record`; the precise removed-page paths
>    (`/platform-log/`, `/tools/decision-record/`) are the right needle.
> 4. Read a live page's dead-link count without a cache-buster and saw a CDN-cached
>    old copy flap; cache-busted origin fetch is the true state.

## 2026-07-22 — LAST MILE COMPLETE; Stage-2 backlog (owner-confirmed)

The site is **live and functional**: serving, the two 404 nav links removed
(verified 0 real dead links across all pages), and the homepage "our work"
showcase live with the 3 grounded projects. `sites.phone` is set. Bug 055 closed.

**Stage-2 backlog (the remaining work, all owner-confirmed as stage-2):**
1. **The interactive components** — the ORIGINAL Thread-B ask, still unbuilt:
   auto-advancing hero card carousel, hover-zoom cards, swipeable mobile
   carousels, code-rendered stat bands. None exist in the framework (from-scratch
   build). Recipes + acceptance checklist already in PLAN/NOTES. The site today
   uses standard section components — making it *look like the brief* is this work.
2. **Contact-details block** (owner routed here 2026-07-22): the contact page is a
   form only; `hero-contact`/`contact-form` have no phone/`tel:` field in schema or
   template, so `sites.phone` (+ email) can't display. Needs a component with a
   phone/email slot. Data is already set and ready.
3. **The 3 empty `planned` pages** — each needs a type-specific build (NOT a
   generic re-queue; they have 0 planned sections by page-type):
   - `self-correction-leopardessconsulting` (`blog-post`) — the story's dedicated
     home; needs a blog-post content build. (Story is already live on about/
     capabilities/multi-agent, so this is a deep-dive, not load-bearing.)
   - `platform-log-index` (`section-index`) — a listing page; needs child entries.
   - `tool-decision-record` (`tool`) — needs the owner-aware tool build.
   All three are currently OUT of nav (unlinked in the last-mile fix), so no 404s.
   Building them = re-linking + a type-specific build.
4. **Council-page stray link** — `info-card-grid` "Review a sample record" →
   `/multi-agent-review-council#decision-record` (no `.html`) 404s. Content-pass fix
   (add `.html` or make it a same-page `#decision-record` anchor + ensure the
   anchor target exists).

**Not this workstream:** `bugs_open/056` (silent-drop-on-regeneration — another
session owns it), `features_open/010` (council decision adjudicator).

## 2026-07-22 — STAGE 2 kickoff: component 1 (hero-card-carousel) built + registered

Mechanism confirmed (see components/README.md): components are `content_components`
rows; `render_mode='agent'` → content-writer fills `input_schema`; `js_content`
auto-publishes to `/tools/assets/{function}.js`; `load_component_library` returns
all active `section` components in the planner's AvailableFunctions (NO
`suitable_site_types` gate — so registration IS enough to REACH the planner; the
NOTES landmine about the planner prompt is about whether the LLM *chooses* it, a
separate empirical check).

**hero-card-carousel [BUILT + REGISTERED + template-validated]:** function
`hero-card-carousel`, section_type `hero-carousel`, section level, agent mode,
is_active. Combines swipe (CSS scroll-snap) + accessible auto-advance (WCAG 2.2.2)
+ hover-zoom. Go template parses + renders with sample data (standalone
`html/template` validator, EXIT 0; empty-image card falls back cleanly, first card
`loading=eager`). Source version-controlled in `components/hero-card-carousel/`.

**NOT yet done for this component:** a **live render on a page** — the meaningful
proof (a registered-but-never-rendered component is the "inert until used" trap).
Blocked by (a) the build/dispatch queue being backlogged (bugs_open/029/030), and
(b) it needs to reach a page either by the planner selecting it or by explicit
placement. Also unverified: per-card imagery generation for the `type:image` card
field (the template falls back gracefully, so it renders regardless).

**Remaining Stage-2 components** (registry additions, same pattern):
`image-hover-card-grid`, `swipeable-insight-carousel`, `stat-band` (code-rendered
from verified data — shared with leopardess L7), `people-feature-block`
(line-illustration). Then: re-plan selected fundamentallyai pages to actually USE
the new components (that's what makes the site "look like the brief"), and verify
each renders live + its JS asset 200s.

## 2026-07-22 — component 1 PROVEN LIVE + a real JS-delivery finding

hero-card-carousel is **live and fully working** on
fundamentallyai.com/capabilities.html (hand-placed page_component at position 2,
grounded pillar content pointing at real pages + real hero images, deployed
assemble-only via direct 049b to bypass the stalled queue). Verified against the
live origin: carousel HTML renders, hover-zoom CSS present, and the auto-advance
JS runs.

**Finding (important for every Stage-2 component with JS):** the per-component
`content_components.js_content` lane **publishes `/tools/assets/{function}.js`
(curl 200) but the assemble injects NO `<script>` tag** — so the JS is
published-but-inert. This is the `bugs_open/041` class, and it applies to SECTION
components, not just chrome (the earlier NOTES/PLAN assumption that the section
`js_content` path "works correctly" is **wrong** — it publishes, but nothing
loads it). The working lane is `js_snippets` (`applies_to:["<function>"]`) →
`render_js_snippets_for_site` bundles it into the site-wide
`/assets/js/snippets.js` that every page already `<script src>`s. Fired the
`site-asset-renderer` agent for the site to rebundle; the carousel JS
(`data-hcc-track`/`initCarousel`) is now in the live bundle — no page re-deploy
needed. Component `js_content` cleared to avoid an orphan asset. Corrected in
components/README.md (JS-delivery convention + acceptance checklist) and
components/hero-card-carousel/snippet.sql.

**De-risked for components 2–5:** register the component + put any JS in
`js_snippets` (not `js_content`) + fire `site-asset-renderer`. The template/CSS
render path and the JS-delivery path are both now proven end-to-end.

## 2026-07-23 — hero-card-carousel design iteration (owner feedback)

Owner reviewed the live carousel: liked the images, the hover-enlarge, the style
and colours. Requested changes (now applied):
- **Overlaid prev/next arrows** on the carousel edges (left arrow on the left edge,
  right on the right), floated over the cards, instead of a top control bar. Left =
  previous card, right = next.
- **Default is PAUSED, no pause/play button.** Auto-advance is now an opt-in flag
  (`autoplay`, default false) — "I like the movement but not for all carousels and
  not for this one". When `autoplay:true`, the pause button appears and rotation
  stops on hover/focus (the previous behaviour); when false (default) the carousel
  sits still and the visitor uses the arrows. The component reads
  `data-hcc-autoplay`; the JS only starts a timer when it is "true".
Applied to the DB (content_components.html_template + input_schema, js_snippets),
re-rendered the live capabilities.html instance, rebundled snippets.js, redeployed.
Source synced in components/hero-card-carousel/. Design direction APPROVED by owner
— proceed to place the components across the pages.

## 2026-07-24 — carousel tweak-2 + FULL ROLLOUT placed & deploying

**Carousel tweak-2 (owner feedback):** arrows nudged up (top:34%, over the image
not the text); the WHOLE CARD is now the click target (card renders as an <a> when
link_url present — the small link below is now a visual cue inside the card, not
the click target); whitespace tightened (align-items:flex-start on the track so
cards size to content; removed flex-grow/margin-auto filler). JS unchanged.
Default-paused JS confirmed LIVE in snippets.js (the earlier stale bundle was an
earlier-queued rebundle racing the DB update — one snippet row, correct content;
re-fire after the update resolved it).

**ROLLOUT (owner: "go ahead and roll it out") — one new component per page so the
design varies per the brief, every instance grounded:**
| page | component | position | content |
|---|---|---|---|
| index | stat-band (dark) | 2, under hero | 3 verified figures: 97% feed restored (relojistas), 11 live sites, <24h to a working site |
| capabilities | hero-card-carousel | 2 (existing, tweaked) | 4 pillar cards → real pages |
| about | people-feature-block | 3 | "Review first, ship second, correct openly" + hero-about illustration → council page |
| multi-agent-review-council | swipeable-insight-carousel | 4 | 4 qualitative on-the-record insights (NO volatile counts — seat numbers change; attributions "our own decision records/commit history") |
| model-fine-tuning | image-hover-card-grid | 4 | "Explore more" — 3 cards → council/capabilities/about with hero images |

All placed via hand-rendered instances (page_components INSERT with position
shift) + direct 049b deploys (queue backlogged as usual); stat-band count-up JS
needs the site-asset-renderer rebundle (fired). Placement content_data JSONs
version-controlled in components/placements/. Monitor watching all 6 signals
(5 components + countup JS) to live.

**Trap hit (shell):** printf with rendered-HTML containing % / ) breaks — build
SQL files with echo/cat only (the shell-tool-traps memory strikes again).

## 2026-07-24 — human style prompt IMPLEMENTED in page-content-writer (owner priority 1)

Owner review session: site-quality automation specs filed (features_open/016
brief-fidelity audit, 017 component-adoption check, 018 specialist design critic
— Gemini for now, 019 sweep enrolment LAST since the improvement loop is off).
Then the prompt work, first per owner.

**Where the copy actually comes from (two paths, easily conflated):**
- `page-content-writer` (chassis agent def) — writes the site page sections via
  `process_sections_loop → generate_content` (`execute_llm_prompt`,
  `prompt_template` 7.8K chars). **Still `claude-sonnet-4-6`/anthropic.**
- `content-creator-agent` (separate service, kustomize configmap) — THIS is what
  the other thread swapped to Gemini (`7b27edfa9`, `gemini-2.5-pro`;
  `014e45ffa` added the provider). The page copy the owner read did NOT come
  through Gemini.

**Style prompt [IMPLEMENTED, LIVE — config is live immediately, no image roll]:**
distilled `REVERSE_ENGINEERED_STYLE_PROMPT_v3.md` (the 3-round owner-refined
de-AI-ify prompt) into a "Voice & Style" block in
`page-content-writer.prompt_template`: one idea per sentence; no em dashes;
start with the fact (no negative-frame/manufactured-reveal openings);
word-weight matched to claim in BOTH directions (no grand words, no dramatised
humility); cut self-flagging filler (crucially/seamless/robust/leverage/delve…);
contractions; active we/our; no cadence-templates; landing sentence ≤1/section
and only if genuinely surprising; one rough edge left standing; no exclamation
points. Document-structure rules (tables/fenced logs/headings) deliberately NOT
carried over — they're for documents, not section fields. STRICT RULE 6
("Professional but engaging tone") now points at the block instead.
Applied patch-style: backup table `bak_agent_definitions_pcw_20260724`,
optimistic guard on `updated_at` (the row was touched by another thread today —
guard held), verified stored (`has_style_block=t`, old rule 6 gone).

**Gemini flip for page-content-writer: READY but NOT DONE** — chassis supports
`provider:"gemini"` (`platform/aiservice/gemini.go`, factory), `GEMINI_API_KEY`
is set in the running pod. One-line ai_service change when the owner says go.
Verification plan either way: rebuild ONE page, read the copy against the
style rules (and voice-tells), before any site-wide rewrite.

## 2026-07-24 — page-content-writer FLIPPED to Gemini; one-page test in flight

Owner: "Flip to Gemini + test one page." `generate_content.ai_service` patched
(guarded on updated_at, same backup table): `anthropic/claude-sonnet-4-6` →
`gemini/gemini-2.5-pro` (`GEMINI_API_KEY`, verified present in the pod;
provider supported per platform/aiservice/gemini.go). Now matches the
content-creator service's model. Verified stored: `gemini/gemini-2.5-pro`.

**Test:** `about` re-queued for a full content rebuild — the FIRST page build
through Gemini + the new Voice & Style prompt. Baseline copy snapshotted
(`about_copy_before.txt`, 15KB) for before/after comparison. Monitor covers:
success (compare copy against the style rules + leopardess story still present —
bug-056 vigilance), failure (work-item error / new agent_error_log rows — a
provider misconfig would fail the generate_content step). Rollback if Gemini
misbehaves: restore ai_service from `bak_agent_definitions_pcw_20260724`
(or just re-point provider/model back to anthropic/claude-sonnet-4-6).

## 2026-07-24 — writer back on Claude (fleet switch-back); style prompt RETAINED; test now Claude+prompt

Owner: "we have switched back. please continue with the prompt change." The
switch-back was fleet-level (sweep `fb6d6ad44` "v1.0.1151 — prior to more
automated reliance on Claude" reverted the content-creator service). My
page-content-writer flip was still standing (DB untouched since my write) —
reverted it to `anthropic/claude-sonnet-4-6` to align, guarded on updated_at,
**keeping the Voice & Style block** (verified: provider/model reverted,
style_block=true). Net state: the model experiment is off; the PROMPT change
stands and is live.

The queued `about` rebuild never ran under Gemini (still triaged behind the
backlog) — so the one-page test is now exactly the clean experiment: **Claude +
new voice prompt vs the Claude + old prompt baseline** (about_copy_before.txt).
Monitor unchanged. On completion: before/after comparison against the style
rules + voice-tells + leopardess-story presence.

## 2026-07-24 — prompt test round 1: real improvement, two residual tells; v2 tightening applied

**The about rebuild ran (Claude + v1 style prompt, 16:11), content SAVED; only
`deploy_page` timed out** ("Request 59150fa3 timed out after 3 retries" — the
awaited-response/backlog class) → item auto-reset to triaged; the retry will
deploy. **Before/after vs the baseline snapshot:**
- Mechanical: filler words 1→0; em dashes 19→14 (down, NOT gone — rule partially
  obeyed); negative frames 0→0; leopardess mentions 2→3 (story survived).
- Qualitative: fact-first openings landed ("FundamentallyAI is an AI
  consultancy." vs the old "not a methodology deck, not a proof-of-concept…"
  stacking); sentences shorter and plainer. But "**That second part matters**"
  reproduced the exact "That X matters" tic, and em dashes persisted.
**v2 tightening applied from this evidence** (guarded update, 16:22): the
em-dash rule now demands a pre-return scan ("a draft containing an em dash is
wrong even if it reads well"); a dedicated rule names and bans the "That X
matters" family. **The pending retry of `needs_page:about` becomes the v2
test** — it will build with the current (v2) prompt.

**STRUCTURAL FINDING — hand-placed components do NOT survive full rebuilds.**
The about rebuild rebuilt sections from the SITE PLAN: my hand-placed
people-feature-block is GONE from about (page now hero-about/about-content/
differentiators/call-to-action). Full-page content rebuilds on the other pages
would likewise strip stat-band (index), swipeable-insight-carousel (council),
image-hover-card-grid (fine-tuning), hero-card-carousel (capabilities).
**Before rolling the new voice across pages: add the 5 placements to
`site_plan_sections` (plan-level data) so rebuilds preserve them** — the
rebuild-proof fix, and the data half of features_open/017. Then rebuild pages,
then restore about's block.

## 2026-07-24 — components written into the SITE PLAN; site-wide v2-voice rebuild wave fired

Chassis now v1.0.1155 (owner deployed). **Plan-level placement done** — the 5
components are now `site_plan_sections` rows (backup `bak_sps_fai_20260724`,
25 rows): index/stat-band@1, capabilities/hero-card-carousel@1,
about/people-feature-block@2, multi-agent-review-council/
swipeable-insight-carousel@3, model-fine-tuning/image-hover-card-grid@3.
Rebuilds now PRESERVE the components (the about clobber cannot recur), and the
content-writer will FILL them through their input_schemas (grounded llm_guidance)
rather than my hand-rendered instances.
**TRAP hit:** `idx_site_plan_sections_key` UNIQUE(plan_id,page_name,ordering) —
an in-place `ordering=ordering+1` shift collides mid-UPDATE; shift high (+100),
insert, then bring back down (−99).

**All 5 pages re-queued** (needs_page:* → triaged) for full rebuilds = v2 voice
prompt + plan components, one wave. Verification when terminal (monitor armed):
per page — plan components present in page_components; copy v2-compliant (em
dashes ≈ 0, no "That X matters" family); leopardess story intact; **stat-band
values grounded** (the writer now authors them per schema guidance "never invent
— leave empty if no verified figure"; any number NOT in the known-true set
{97%, 11, <24h} is a finding). Then: restore-check about, and onwards to
features_open/016 (brief-fidelity audit).

## 2026-07-24 — features_open/016 BUILT: brief-fidelity-auditor seeded + commissioning run fired

While the 5-page rebuild wave drains (all triaged, queued), built 016 v1 — the
"did the machine build what was asked" audit. **Config-only, live on insert**
(agent_definitions seed; every action already registered — no image roll):
`ensure_site_record → load_brief_context (SQL: mission_brief + design_intent +
content_direction + per-page component inventory + imagery counts) →
run_fidelity_audit (execute_llm_prompt, claude-sonnet-4-6 — matching the fleet's
switch-back; model is one config line) → write_audit_findings
(audit_source='brief-fidelity-audit') → complete`. Mirrors visual-design-auditor's
proven shape + the auditFinding output contract exactly (category/severity/
description/current_value/suggestion/acceptance_test/…, findings_field
'<llm_output>.result'). Audit rules in the prompt: extract CONCRETE promises
from the brief's own words; report ONLY broken/under-delivered ones (≤8), each
quoting the brief phrase; current_value must come from the given inventory (no
invented facts); kept promises are NOT findings — this grades fidelity, not
taste (taste = 018). v1 deliberately inventory-based, no screenshots (component
classes/imagery density/layout variety are inspectable from the DB; pixels come
with 018).
Source: `agents/brief-fidelity-auditor.{config.json,seed.sql}`.
Commissioning run: corr `ca900c2a` vs fundamentallyai (findings will reflect the
PRE-wave state; re-run post-wave for the real read). Findings land as
site_work_items, status 'detected'.

## 2026-07-24 — wave 1 partial (reconciler interaction); index verified; wave 2 re-queued properly

**FLEET-RELEVANT FINDING — re-queueing a DEPLOYED page's needs_page item no
longer works by itself.** The new superseded-review reconciler (bugfix-056
thread, live on v1.0.1155) sweeps non-terminal needs_page items whose page is
`build_status='deployed'` → my 5 triaged re-queues were marked `unresolved`
before the build loop claimed them; only `index` won the race and rebuilt.
**Correct re-queue for a deployed page: set `pages.build_status='needs_rebuild'`
FIRST, then the item to triaged.** (Done for the 4 swept pages — wave 2 running.)

**index rebuild VERIFIED (the one that ran; v2 prompt + plan components):**
- stat-band arrived from the PLAN and was filled by the pipeline. Guardrails
  MOSTLY held: "1 hallucination caught and corrected" (grounded), "0 fabricated
  client claims" (honest), and an EMPTY value for build-time — the writer obeyed
  "leave empty rather than invent". **But "5 — Agent roles in review council" is
  WRONG** (council is 13+ seats; "5" looks like a stale fragment) — the
  unverified-number class, unblocked because **fundamentallyai has NO
  evidence_base** so checkUnregisteredNumbers never runs here. → The systematic
  fix is to SEED an evidence_base for this site (the designed opt-in gate), not
  to hand-edit the stat. Proposed to owner.
- v2 copy: "That X matters" family = 0 ✓; **em dashes persist (11, mostly in
  stat captions)** despite the v2 hard rule — model keeps reaching for them.
  Options: a 3rd prompt tightening, accept some, or a mechanical post-pass.
- **Component template fix shipped:** stat-band now `{{if .value}}`-skips
  empty-value stats (no blank cells; validated; applied to content_components;
  synced to repo).

## 2026-07-24 — EVIDENCE BASE SEEDED for fundamentallyai (owner-directed) + fault-injection PROVEN

`site_specs` aspect `evidence_base` seeded: **9 facts** (all live-verified at
seeding or artifact-dated: 12 deployed sites → floor "more than 10" gte; 16
council-gate seats → floor "more than a dozen" gte, both with live-SQL sources
so the evidence-freshness task re-verifies them; relojistas 97/100/24; the
leopardess correction=1; 0 fabricated clients; idea.uk £29; private-search
"buildable not delivered"), **6 banned patterns** (the leopardess exclusion
list + "5 agent roles" — the wrong figure that reached the index stat-band,
banned by name with the verified 16 cited), allowed_entities = the four
approved reference domains, writer_block (managed) feeding the content-writer
prompt. Mirrors the leopardess shape exactly (incl. floor-form gte discipline —
"11 sites" had ALREADY gone stale to 12, proving the rule).

**Fault-injection VERIFIED via claimscan** (clean HEAD archive — shared tree
broken again by another session's WIP): all 3 induced banned claims fired incl.
"5 agent roles"; unregistered numbers (99.9, 70, invented "43 clients") flagged;
the legitimate registered claims produced ZERO false positives. Both branches
exercised. The number-gate that was missing when "5" shipped is now LIVE for
this site: validate_page_content check 8 runs on every future build.

index re-queued (needs_rebuild first) so the stat-band regenerates under the
gate with the writer_block facts. Source: `sql/evidence_base_fundamentallyai.json`.

## 2026-07-24 — wave-2 friction decoded (partially); SEQUENTIAL works; about PROVEN end-to-end

Two theories REFUTED by evidence (recorded so nobody re-walks them): (1) the
superseded-review reconciler only touches `status='needs_human_review'`
(reconcile_superseded_reviews_action.go:96) — my items were triaged; (2) the
two-strike rule counts terminal ROWS per item_key in 7 days
(load_work_item_actions.go:1041-79) — these keys each have ONE row (my re-queues
UPDATE it), 0 terminal rows. Working theory (fits both waves, unproven):
single-flight per site — claiming one needs_page item parks the sibling triaged
ones as unresolved. Fallback that WORKS: **sequential** — one page at a time,
re-queue on completion.

**about rebuilt & VERIFIED — the full plan-driven chain is proven:** 5
components, `people-feature-block` RESTORED at position 3 straight from
`site_plan_sections` and filled by the pipeline (no hand-render), leopardess
story present in 3 components. Plan-level placement → rebuild → component
survives: the clobber class is closed.

**Backlog note:** operator-driven bulk rebuilds have no paved road — the immune
mechanisms treat deliberate iteration as recurrence. A first-class "rebuild
these pages" operation (or an operator flag the rules respect) is the fix.
Sequential queue: capabilities (now) → multi-agent-review-council →
model-fine-tuning → index (stat-band under the new evidence-base gate).

## 2026-07-25 — the parking mechanism FOUND; my "single-flight" theory was wrong

> **CORRECTED 2026-07-25 — the working theory recorded above ("single-flight per
> site parks sibling triaged items") is WRONG as a description of what parks
> them.** The parker is `stale-work-item-reaper`, a `scheduled_tasks.pre_query`
> (hourly, live, `enabled=true`) that flips any `triaged` build item to
> `unresolved` when **`created_at`** is 48h+ old and `claimed_at IS NULL` — row
> age, not time spent in `triaged`. Our re-queued rows were created 2026-07-20
> 21:27, so every re-queue of them was eligible the instant it was set `triaged`.
> Filed as `bugs_open/070`; pattern in 016b §9; feature `features_open/021`.

**What caught it: the row's own `summary` column.** It reads
`[stale: triaged 48h+] [stale: triaged 48h+] Build index page (not_built)` — the
mechanism, its rule, and the fact that it fired twice, stored in the row I had
been querying all along. My `SELECT`s truncated `summary` to 50 chars for
readability and cut the prefix off. `grep -rn "stale: triaged"` then found the
writer in one hop — and it is **DB config, not Go**, which is why reading actions
code could never have found it. Logged in WRONG_CALLS.md.

**Why sequential works and batches don't** (the half the theory got right, now
with the real reason): the claimer `build-pipeline-trigger` runs every **120s**,
the reaper every **3600s**. One re-queued item is claimed in ~2 min and
`claimed_at IS NULL` then excludes it forever. A batch loses because siblings
wait behind an in-flight build (tens of minutes) — long enough for the reaper's
tick. Verified both directions on this site today: `needs_page:capabilities`
re-queued ~08:09 -> claimed 08:11 -> complete 08:17; the three batch rows parked,
twice. The site is NOT locked (`sites.locked_at IS NULL`) and every item has
`attempt_count=0 < max_attempts=3`, so neither of those is the gate.
[INFERRED, not traced] the serialisation that makes siblings wait is downstream
of the trigger — possibly the same mechanism as `bugs_open/029`/`030`.

**capabilities rebuilt & VERIFIED** — second page proving the plan chain: 5
components, `hero-card-carousel` restored at **position 2** from
`site_plan_sections` (11,714 chars rendered), pipeline-filled. `about` and
`capabilities` now both demonstrate plan-level placement surviving a rebuild.

**Practice change for the rest of this rollout:** keep queueing sequentially, and
prefer creating a NEW work-item row over re-queueing a historic one — a fresh
`created_at` is outside the reaper's rule, and the summary then describes the
rebuild actually being asked for rather than a 2026-07-20 first build. Caveat:
a hand-written `INSERT` bypasses the Go-side two-strike suppression in
`insertWorkItem`, so that guard is skipped deliberately, not accidentally.

## 2026-07-25 — the evidence-base gate's first organic catch: a real defect, flagged for the wrong reason

index's rebuild parked at `needs_human_review`: check 8 flagged
`unregistered_number "2192"` ×3. Classification wrong — `\2192` is the CSS
escape for `→`, not a business claim — but the page defect was REAL: the
`portfolio-showcase` template had `Visit Site \2192` in anchor TEXT (HTML
context), so visitors literally saw "Visit Site \2192". CSS escapes only mean
anything inside CSS `content:`; the older components (`content-sidebar`,
`guide-list`, `blog-listing`) all do it correctly. Fixed in the template with
`&#8594;` (entity, immune to the [[escape-sequence-emission-trap]]), snapshot in
`bak_cc_portfolio_showcase_20260725`; fleet sweep of rendered HTML = 1 instance,
index only, which the re-queued rebuild re-renders. DB-only, live immediately.

Grep gotcha re-hit: plain `grep -rn "2192"` on the repo went SILENT (NUL-byte
binary trap from another session's WIP); `git grep` found the check in one hop.

**Wave verification so far (4 of 5 pages):** every rebuilt page carries its
planned interactive component — about/`people-feature-block` pos 3,
capabilities/`hero-card-carousel` pos 2, multi-agent-review-council/
`swipeable-insight-carousel` pos 4, model-fine-tuning/`image-hover-card-grid`
pos 4. Copy: "That X matters" = 0 on all four; em dashes persist (8/6/5/2 per
page, aside-style) — the third-round decision (worked example vs mechanical
post-pass) remains with the owner. Also seen: the "isn't a log entry. It's a
decision record" negative-frame twice on one page (v3 prompt rule 3/13
residue). index pending under the fixed template.

## 2026-07-25 — index LANDED; bugs_open/070 confirmed live on our own queue; two brief-core pages worked

**index rebuilt & deployed** with all 7 planned components incl. `stat-band`
(pos 2) and the fixed `portfolio-showcase` (pos 6). Fleet sweep for the literal
`Visit Site \2192` in rendered HTML: **0**. Wave complete — all 5 primary pages
carry their planned interactive component:

| page | interactive component | pos |
|---|---|---|
| index | stat-band | 2 |
| about | people-feature-block | 3 |
| capabilities | hero-card-carousel | 2 |
| multi-agent-review-council | swipeable-insight-carousel | 4 |
| model-fine-tuning | image-hover-card-grid | 4 |

**`bugs_open/070` reproduced on our own queue, unprompted.** Re-queued
`needs_page:self-correction-leopardessconsulting` (created 2026-07-20) at ~15:20;
the reaper's tick parked it `unresolved` at **15:32** with a fresh
`[stale: triaged 48h+]` prefix — 12 minutes after a re-queue, the label claiming
48h. Best possible evidence for the bug file, from the exact failing branch.
Then applied 070's own workaround: **INSERT a fresh row** (`c2a3cb85`, source
`operator:brochure_component_library`, summary describing the REBUILD) rather
than resurrect the historic one. `unresolved` IS in `idx_swi_dedup`'s terminal
set, so the same `item_key` inserts cleanly alongside the parked row — the
workaround is legal, not a hack around a constraint. Landmine found doing it:
`site_work_items.created_by` is NOT NULL with no default, so a copy-INSERT
must name it.

**Two brief-core pages are missing from the site plan entirely** — that is why
they never built, and it is a REAL brief-fidelity gap, not queue friction:
`self-correction-leopardessconsulting` (the leopardess trust story the brief
names as the differentiator, "Name that site directly as the worked example")
and `platform-log-index` (the decision record). Both sat `planned` with 0
sections and 0 components since 2026-07-20; the current plan (30 rows) has
sections for only 6 pages. The build refused correctly and said so:
`plan_sections` -> `check_has_ready_sections` -> **`mark_no_ready_sections`** in
38 seconds, no LLM spend. Placed 5 plan sections for the self-correction page
(hero / generic-text-block / swipeable-insight-carousel / info-card-grid /
call-to-action) so the pipeline has something to fill.

**Also found, not yet fixed:** every same-page anchor on the two fresh
mid-pages is a dead target — `#decision-record`, `#reviewer-seats`,
`#role-design`, `#production-sites`, `#self-correction` on the council page and
6 more on model-fine-tuning resolve to **0** matching `id="..."` attributes. The
writer emits anchor hrefs; no component emits section ids. That is a
generic-detector gap (a phantom-link check that only tests PATHS passes all of
these), so it belongs with the 023/049 link-integrity family rather than as a
hand-fix here.

## 2026-07-25 — the owner's phone: a fleet-wide contract mismatch, not a missing edit

**The phone number was never missing — it was stored where nothing reads it.**
Owner gave `+44 (0) 7934 524 911` on 07-24; it went to `sites.phone`. The
contact-details component `contact-info` sources
`site_specs.identity.email/.phone` as **FLAT** keys, while
`domain-research-classifier` writes them **nested** under `contact.*`. Missing
`email` carries `on_missing: needs_human_review`, so the entire section was
withheld — silently, `sections_ready`=2 with **no** `sections_skipped` or
`sections_deferred` key at all (orchestration `de2da37d`). Three contact builds,
two components every time.

**Fleet discriminator is exact** (all 13 deployed sites, no exceptions either
direction): the 5 sites with a flat `email` key are precisely the 5 where
`contact-info` has ever rendered; the 8 without have no contact-details block.
idea.uk renders on flat `email` alone with no flat `phone` — consistent with
`email` being the withholding field. The classifier's own shape is the nested
one, **so a new site is broken by default**. Filed `bugs_open/072`.

Fixed at the DATA level for this site only (contract untouched): flat `email` +
`phone` added alongside the nested pair, prior row superseded FIRST —
`idx_site_specs_current` is UNIQUE `(site_id, aspect) WHERE is_current`, so
insert-before-supersede dies with 23505 (hit it). Backup
`bak_site_specs_fai_identity_20260725`, all 6 `services` verified preserved.
**Rebuild now yields 3 components and the phone is LIVE** on
`/contact.html` as `tel:+44 (0) 7934 524 911`. (Minor: that `tel:` URI keeps
spaces and parens — invalid per RFC 3966, tolerated by browsers. Template-level,
fleet-wide, left alone.)

Also recorded in 072: the work item reported `complete` while the page stayed
`needs_rebuild` and undeployed — the partial-build guard
(`v3_site_actions.go:684`) correctly refusing a short build, but the item status
doesn't reflect it. Reading the item alone says "rebuild succeeded".

## 2026-07-25 — link repair: my check and my fix shared a blind spot, twice

Census found **21 of 22 internal links broken** (`bugs_open/071`). Repaired in
`rendered_html` + `content_data`, quoted-exact replacement, backup
`bak_pc_fai_links_20260725`. Post-check: 0 unresolvable. **Then the live sweep
found 21 MORE.**

> **CORRECTED same day:** my census regex was `href="(/[^"#?]*)"` — it required
> the closing quote right after the path, so it **excluded every anchored href**
> (`/capabilities#approach`). My fix used the same quoted-exact form, so it
> skipped exactly the class my census could not see, and my post-check reported
> clean because it reused the blind regex. The live crawl is what caught it.
> Correct form: capture `href="(/[^"]*)"` then `split_part(...,'#',1)` for
> resolution. Second fix: 21 anchored hrefs → `.html#anchor`.

**Two blind-spot pairs in one day** (this, and the truncated `summary` that hid
the reaper). Both are now in the RUNBOOK. The pattern worth keeping: *a
verification that shares code, regex or assumption with the fix cannot falsify
it* — the live artefact is the only independent witness.

Verification-method trap also hit: hammering the origin with cache-busted
requests in a tight loop produced `000` and one spurious `404`, i.e.
**throttling read as broken links**. Re-probed serially: 200 in 0.5s. The final
sweep retries three times before condemning a link.

Publishing route: **`049b_deploy_single_page.sh` silently failed to ingest** ×4
(no orchestration row at all — the documented kubectl-run stdin race).
`scripts/republish_page_086.sh` (new, committed) DOES work — ~2 min to show a
row, so don't declare it dead at +45s (I nearly did). The route needing no Kafka
envelope is a fresh `page_rerender` work item; used it for 5 pages, all
`complete`. Live state now: 6 of 7 pages clean, capabilities republishing (it
was missed from the first batch — the anchor fix landed after those items were
created).

## 2026-07-25 18:10 — ONE ITEM OUTSTANDING, and it is blocked outside this workstream

**State: 43 of 44 internal link targets resolve live. The single failure is
`/capabilities`, referenced 10 times from `capabilities.html` itself.** The
database is already correct (census returns 0 unresolvable); only the *publish*
is outstanding. Nothing in this workstream can close it.

**Why it is stuck:** the build dispatch lane stalled fleet-wide at ~17:45.
Measured 18:00: 99 `triaged` build items (95 of them `webdesign.co.uk`'s bulk
enqueue), **0** items in `claimed`, **0** claims in the preceding 15 minutes,
`build-pipeline-trigger` firing on time every 120s, `agent-build-dispatch-loop`
pod `Completed` 23 min earlier with none since, no site locks, no errors. That is
`bugs_open/030`, which another thread OWNS and shipped a lane fix for **today**
(`f9bc7f45f`, IMAGE_TAG v1.0.1164) — and the cluster is still on **v1.0.1159**, so
what I saw is pre-fix behaviour. Ran `scripts/who-owns.py 030`, confirmed
ownership, contributed the measurements into their bug file and **backed off
rather than start a competing diagnosis**.

**To finish it** (any thread, once the lane moves — no re-analysis needed):
```sql
-- the queued item is already correct and waiting; just confirm it lands
SELECT status, claimed_at FROM site_work_items WHERE id::text LIKE '4a7f5520%';
```
then verify live, which is the only witness that counts:
```bash
curl -s "https://fundamentallyai.com/capabilities.html?cb=$RANDOM" \
  | grep -cE 'href="/capabilities#'     # 0 = done
```
If the item was reaped in the meantime, INSERT a fresh row (see RUNBOOK) — do not
re-queue the parked one (`bugs_open/070`).

**Two dead ends recorded so the next thread does not repeat them:** no content
edit helps, because the content is already right and *publishing* is the blocked
step; and every publish route (049b, the 086 script, `page_rerender` work items,
`apply_section_edit`) goes through the same dispatch lane, so switching route
cannot route around a stalled lane. I fired two routes before understanding this;
a third would only have queued more work.

**Latency baseline for this site, worth keeping:** 7–9 minutes from dispatch to an
`orchestration_states` row was NORMAL while the lane was healthy, and ~5 minutes
from a fresh `page_rerender` item to `claimed`. Anything shorter than that is not
evidence of failure.

## 2026-07-25 18:35 — CLOSED: every live link resolves. The queued work item did it.

**Final independent verification (live crawl, 3 retries per link, all 7 pages):
`43 unique targets checked, 0 broken`.** No database claim involved — this is the
served artefact.

The dispatch lane recovered on its own at ~18:30 and the **queued `page_rerender`
work item** (`4a7f5520`) went `triaged` → `claimed` → `complete`, publishing
capabilities. Worth recording precisely: **the route that finally worked was the
one I could not hurry**, and the two direct dispatches I fired in impatience
contributed nothing. Waiting was the correct action from ~17:50 onward; I only
reached it after measuring the stall and finding 030 already owned.

Brief-fidelity findings re-tested against the rebuilt site (016's earlier run, now
meaningful because the wave is complete):

| finding | verdict now |
|---|---|
| self-correction page absent (`958804fa`) | **RESOLVED** — deployed today, 5 sections |
| model-fine-tuning ≡ council template (`26ae50f9`) | **RESOLVED** — sets now differ (`image-hover-card-grid` vs `swipeable-insight-carousel`) |
| imagery thin (`722111c4`) | **IMPROVED, not resolved** — was 2 of 27 components with images, now **10 of 43** (23%) |
| no chart component anywhere (`5501b583`) | **NOT RESOLVED, and structural: zero chart components are registered in the entire fleet** ([UNCHANGED] since the audit) |
| `platform-log-index` absent (`0ec97996`) | **NOT RESOLVED** — never in the site plan; owner decision pending (publishing internal review records outward is their call, not mine) |

Statuses left as `detected` deliberately: another workstream owns work-item
completion semantics, and hand-closing an audit's own findings would corrupt its
accounting. The re-test lives here instead.

**The chart gap is the sharpest remaining brief miss** and it is not
site-specific: the brief requires "numbers rendered as real, code-generated charts
from true figures, never an AI-generated picture of a chart", the owner asked for
"charts and infographics", and **no such component exists to select**. `stat-band`
renders verified numbers but is not a chart. Prior art to reuse rather than
duplicate: leopardess **L7** ("the one genuinely-new build", scoped in
`docs/leopardessconsulting/PLAN_leopardess_rebuild.md`, listed as `[gap]` in
`REPLICATION_in_chassis.md`), plus `features_open/023` (infographic figures from
the evidence base). Recommend one shared component, values sourced from the
evidence base so a chart cannot show an unverified figure — not a new per-site
build.

## 2026-07-25/26 — Missteps this session, the complete list (the record IS the point)

Same discipline as the 2026-07-22 list above: every one, with what caught it and
the cheap check. Cross-filed in `WRONG_CALLS.md` where the check-class is
fleet-wide (tally updated there — 4 increments, 4 new rows; **not updating that
tally was itself a slip on the first three entries**, exactly what the
`missteps-need-a-check-not-a-paragraph` memory warns about).

**The single shape behind five of these: I concluded from a check that could not
see the evidence.** Different costume each time.

1. **[BIG ONE] Built and refuted two theories about parked work items while the
   row's own `summary` column named the mechanism in plain text.** Read the
   superseded-review reconciler (predicate is `needs_human_review`, mine were
   `triaged` — refuted), then the two-strike rule (counts terminal rows; these had
   none — refuted), then told the owner "single-flight per site" was the surviving
   theory. The answer was `[stale: triaged 48h+] [stale: triaged 48h+] Build index
   page (not_built)` — **stored in the row I had been SELECTing all along**, naming
   the mechanism and that it had fired twice. *Caught by:* finally reading the row
   untruncated. *Cheap check:* read every column of the row whose fate you are
   explaining — `summary`, `status`, `attempt_count`, `claimed_at` — before
   reasoning about which code moved it; and **my own `left(summary,50)` had cut off
   the answer**. Corollary I also got wrong: I never asked whether the mover was
   code at all — it was a `scheduled_tasks.pre_query`, so no amount of Go-reading
   would have found it. → `bugs_open/070`.

2. **[BIG ONE] Declared the links fixed using a check that shared the fix's blind
   spot.** Censused with `href="(/[^"#?]*)"`, repaired with the same quoted-exact
   form, re-ran the same census, got zero remaining, and **reported the site's links
   fixed to the owner**. The pattern cannot see anchored hrefs
   (`/capabilities#approach`), so 21 survived: the census missed them, the fix
   skipped them, the post-check confirmed success. *Caught by:* crawling the served
   pages. *Cheap check:* **a check sharing the fix's regex/query/assumption can only
   echo it** — verify against an independent witness, the served artefact. Correct
   form: capture `href="(/[^"]*)"` then `split_part(href,'#',1)`.

3. **[BIG ONE] Declared four dispatches "silently failed to ingest" from checks ~5
   minutes too early — then published it as a landmine in two new files.** No
   `orchestration_states` row at +2 and +5 min; there IS a famous kcat/stdin race
   (`016b` §9) that fits exactly; I matched symptom to famous cause, switched
   scripts, and wrote the claim into a **new RUNBOOK and a new script's header as
   fact**. All four had landed — rows at 17:12–13 for 17:05 dispatches. Normal
   latency here is **7–9 minutes**. *Caught by:* querying the whole window later,
   looking for something else. *Cheap checks:* establish the healthy baseline before
   calling a reading abnormal; **a famous failure mode that fits your symptom is a
   hypothesis, not a diagnosis** — most dangerous precisely because the docs made it
   famous; and prove a landmine before writing it into a shared doc, because a
   runbook entry asserts at higher confidence than a note and propagates.

4. **Compounded #3 with a monitor whose window could not contain the evidence.**
   Filtered `created_at > '2026-07-25 18:00'` while the clock read **17:54** — it
   can never match, and reports identically to a dead dispatch. Four more
   "no orchestration row" ticks that felt like confirmation. *Cheap check:* print
   the window and the clock together; an absent row is only evidence if your query
   could have contained it.

5. **Never checked `start_step`, so I credited the wrong route for a success.** The
   homepage republish was attributed to the 086 script when the queued work item may
   equally have done it. *Cheap check:*
   `initial_request_data->'config'->'workflow'->>'start_step'` = `spawn_rerender`
   for an 086 dispatch, NULL for 049b/work-item. Without it you cannot tell which of
   your own attempts did the work.

6. **Fired two direct dispatches out of impatience while a queued item was
   pending.** The route I could not hurry is the one that published the last page.
   Worse, no route could have helped: **every publish path shares the dispatch
   lane**, which was stalled. *Cheap check:* before switching route, ask whether the
   routes share the blocked component.

7. **Wrote an em dash into the rule forbidding em dashes**, and it reached the live
   agent config before I read it back. Reading it back then exposed the real cause
   of two failed refinements: **the prompt already contained 17**, 14 in its own
   instructions including the `## Voice & Style` heading. *Cheap check:* after
   editing an instruction, grep the whole instruction for the thing it prohibits.

8. **Filtered a "does this page exist" lookup on `build_status='deployed'`** and
   reported `/contact` as an invented target while `/contact.html` served 200 (its
   row read `needs_rebuild`, its artefact was live). I had recorded this exact trap
   in `016b` §9 three days earlier. *Cheap check:* existence is a row, not a build
   state — and one `curl` settles it.

9. **Over-engineered the link-repair SQL before trying the plain form.** Wrote a
   tangled `CROSS JOIN LATERAL` fold, discarded it, and did it in nested
   `replace()` calls that were correct first time. No false claim, pure waste —
   recorded because the reflex is the problem.

10. **`\copy … FROM PROGRAM` with JSON containing newlines** read them as row
    separators; the `NOT NULL` constraint caught it before any corruption. Fixed
    with a base64 round-trip (now in `sql/README_writer_prompt_v3.md`). *Cheap
    check:* a multi-line payload needs an encoding that has no line semantics.

11. **A relative `cd` in the shell tool silently wrote nothing.** `cd X && cat >>
    f << EOF` — the working directory had already changed from a previous call, so
    `cd` failed, the `&&` short-circuited, and the heredoc was consumed with no
    write. *Caught by:* `tail`ing the file. *Cheap check:* absolute paths for
    appends. (In the RUNBOOK's shell traps.)

12. **Hammered the origin with cache-busted requests and read the throttling as
    broken links** — `000`s and one spurious `404`. *Cheap check:* retry 3× with a
    pause before condemning a link; serial re-probe returned 200 in 0.5s.

**What the distribution says for this workstream:** none of these were missing
information. Every one was available in a field, a window, or a page I could have
read. The corrective that would have caught the most (1, 2, 4, 8) is a single
habit: **read the actual artefact, untruncated, with a check that does not share
its assumptions with what you are checking.**

---

## 2026-07-26 (later session) — building the chart component

Owner green-lit the chart (handoff §3a). Decisions taken with the owner at the
start of this session, before any code: **config now, Go later** for the route;
**all three** candidate charts rather than one; **both pages** (index and
capabilities).

### Why the config route was available at all

The handoff assumed a chart needed a Go build. It does not. `plan_sections_action.go:399`
(`resolve`) → `:478` (`resolveSpecPath`) → `:509` (`navigateMap`) already resolves
`source: "site_specs.<aspect>.<path>"` for any component field, and resolved data
**beats LLM content at render time**. So "the LLM may not supply figures" stops
being a prompt instruction and becomes a property of the pipeline: the values are
system-resolved and overwrite whatever the writer wrote.

The Go prior art is real but already inert — `platform/orchestration/actions/report_charts.go`
(`renderBarChartSVG`, committed 2026-07-24 for the gripper dossier) is waiting on
an image roll like everything else Go-side. Reuse it for the later lift; it was
right not to fork it now.

### The data contract, and the one rule that shapes everything

`charts` holds ids, labels, order, prose and a scale constant. `facts` holds the
values. A chart definition **names fact ids and never restates a value** — so
there is nowhere for an unverified number to live. Where a denominator is itself
a measured quantity it comes from `max_fact_id`, not a literal.

That rule immediately bit, and correctly: the register holds F3b, *"100% failure
before relaunch"*. A success-rate chart therefore had no baseline to draw —
deriving `0` in the chart definition would have been exactly the invented figure
the design exists to prevent. Fixed by recording the baseline as its own fact
(F3c, same source, same measurement, stated as a success rate).

### Traps found by testing rather than by reasoning

1. **A round million renders as `1e+06`.** JSONB numbers arrive as `float64`, and
   `{{.value}}` prints via `%v` (i.e. `%g`). In the bar geometry that is invalid
   CSS; as visible text it is nonsense. Geometry now uses `printf "%.4f"`, the
   visible fallback `printf "%.10g"`. **Found by putting a round million in the
   sample data on purpose**, not by reading the template.
2. **`html/template` neutralises a hostile value in a `style` attribute**
   (`ZgotmplZ`) — proven, so the data layer cannot inject CSS. The same filter
   rejects a *string*-typed value under `%.4f`, so a charted fact's `value` must
   be a JSON number. VERIFY check 3 enforces it.
3. **`<svg>` is in `nonAssertionElements`** (`datahelpers/claims.go:137`): text
   inside an SVG is **invisible to the claims gate**. Keeping the figures in real
   HTML text is what keeps them checkable. This is now the first thing the later
   Go-SVG lift has to solve, and it is written into the component README so it is
   not rediscovered.
4. **The claims gate reads a ±70-character window** and needs one of the fact's
   `context_terms` inside it (`claims.go:493`), with block elements delimiting
   the window. So a point label must carry its fact's own wording or the gate
   reports **our own charted figure** as an unregistered number. VERIFY check 6.
5. **`refresh_evidence_base` rewrites `value` and `verified_at` but never
   `display`.** A hand-written `display` on a SQL-sourced fact would therefore
   drift away from the bar beside it as the query result moved. Rule: SQL-sourced
   facts carry no `display`. VERIFY check 5.

### MISSTEP — and the check that caught it

Adding `max_fact_id` support, I introduced the `$max` variable and **left the use
site reading `$c.max`**. Every fact-sourced denominator rendered `--m:ZgotmplZ`,
i.e. a dead bar. Nothing about the diff looked wrong; the harness caught it on
the next run because the sample data exercises that path.

*Cheap check, and the general form of it:* **the sample data must contain the
case you are adding, not just the case you already had.** The three template
mechanics that could silently fail (exponent formatting, CSS filtering, the
fact join) are each represented by a deliberately awkward row in
`components/evidence-chart/sample_data.json`, and the harness asserts on them
rather than printing output for a human to skim.

### CORRECTION to an existing fact

F3/F3b carried `verified_at: 2026-07-16` and a source of "measured 2026-07-16".
**That date cannot be true**: the relojistas cutover was on the 17th, the first
full day of data was the 18th, and the measurement was written up on the 19th
(`traffic_probe/README_where_we_are.md`, "## 2026-07-19 — the reactivation number
is in"). Corrected to 2026-07-19 in the seed, with the source pointed at the
document that holds the measurement. It matters here more than usual because
**the chart renders the verified date on the page**.

### The em-dash baseline recorded in the handoff no longer matches

Handoff next-action (a) records the 2026-07-25 baseline as index 6, capabilities 6.
Measured immediately before this rebuild: **index 11**, capabilities 6, about 8,
fine-tuning 5, council 2. So index has moved since that baseline was written
(the index was rebuilt after it). The comparison below therefore uses **today's
pre-rebuild numbers**, not the handoff's, and I have not overwritten the earlier
figure — a baseline that turns out to be stale is itself worth seeing.

### The CSS variables everyone copies do not all exist

Acceptance item 3 says "CSS consumes `var(--section-*, var(--color-*))`, never
hardcodes colour — verified against a dark-section AND a light-section site".
Checking that properly meant asking what the themes actually define:

```sql
SELECT name, (SELECT string_agg(DISTINCT m[1], ', ' ORDER BY m[1])
              FROM regexp_matches(css_content, '(--[a-z0-9-]+)\s*:', 'g') AS m)
FROM css_themes WHERE is_active;
```

**`--color-surface`, `--spacing-section` and `--container-max-width` are defined
by NO active theme.** I had used the first of those for the chart card, so its
fallback — a light grey — would have rendered on *every* site including the dark
one. A light card on a dark page, from a rule that reads as theme-aware.

The vocabulary that does exist, on both `default` and `leopardess-dark-gold`:
`--color-background`, `--color-text`, `--color-text-muted`, `--color-primary`,
`--color-secondary`, `--color-accent`, `--color-card-bg`, `--color-border`,
`--border-radius`, `--shadow`, `--spacing-xs…xl`. Switched to those, and where
no variable exists the fallback is now a **neutral translucent grey** rather than
a light literal, so it reads correctly on either background.

Worth noting `stat-band` also references `--spacing-section` and
`--container-max-width`; it has always been rendering their fallbacks. Harmless
there (spacing and width, not colour), but the same reflex produced a real defect
here.

*Cheap check, and it is one query:* before using a CSS variable, confirm a theme
defines it. A `var()` fallback makes an undefined variable invisible — that is
the point of the fallback, and also why the mistake never announces itself.

### The voice fix, measured at last — mixed, and not a clean win

Handoff next-action (a): refinement 3 was live but unmeasured. These rebuilds are
the first page writes since, so the measurement came free.

Like-for-like, counting only the components that existed before and after:

| page | before | after | note |
|---|---|---|---|
| index (the 7 pre-existing sections) | 11 | **6** | portfolio-showcase alone accounts for 4 of the 6 |
| capabilities | 6 | **6** | unchanged; hero-card-carousel alone accounts for 4 |

Index as a whole reads 8 after the rebuild, but that includes the **new**
evidence-chart section contributing 2 — one of them in the LLM-written intro
("…where it stands today — not what we project it will do"). Comparing whole-page
totals would have flattered the result on one page and hidden the new section's
contribution; per component is the only honest cut.

**So: a real drop on one page, none on the other.** Per the handoff's own
instruction, that is not grounds for a fourth prompt round — the alternative
already offered to the owner is a mechanical post-pass, and this is the
measurement that should decide it. Note the concentration: two components
(`portfolio-showcase`, `hero-card-carousel`) hold 8 of the 12 remaining em dashes
across both pages, which is a strong hint that a **per-component** fix would beat
another site-wide prompt edit.

`That X matters`: **1 instance**, not 0 — on `self-correction-leopardessconsulting`,
which was NOT rebuilt in this session, so it is pre-existing rather than a new
regression. The handoff records "was 0", which was presumably scoped to the pages
it rebuilt; leaving both figures visible rather than overwriting either.
