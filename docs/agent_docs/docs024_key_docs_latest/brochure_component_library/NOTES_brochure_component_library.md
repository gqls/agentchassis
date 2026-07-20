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
