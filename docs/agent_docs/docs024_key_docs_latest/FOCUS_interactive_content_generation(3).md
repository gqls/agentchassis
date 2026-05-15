# FOCUS — Interactive Content Generation

How the platform builds tools, games, news, and other interactive content
types — what works today, what's missing, and where the pieces fit.

This is a reference for the bigger goal: a site creator that produces
sites whose every content type has something real behind it, modelled on
adopted examples where useful, generated from capability where exact
reproduction isn't reachable. Not a roadmap — a map of the territory.

Anchored in `028_platform_mission_and_pipeline_direction` ("produce the
best possible website for each domain, end to end") and extending it to
the interactive-content side that's currently underdeveloped.

## The four-stage pattern

When the platform encounters interactive content on an adopted site
(a calculator, a game, a news feed), there's a pattern for handling it
that we'd like to apply consistently across content types:

1. **Parse** — understand the source. For a JS calculator: what inputs,
   what outputs, what the formula is. For a game: what mechanics, what
   the win/loss condition is, what the simulation loop does. For a news
   feed: what the editorial voice is, how stories are sourced.

2. **Assess** — decide what we can produce. "We can build exactly this."
   "We can build something in the same genre/league with our current
   capabilities." "We can only render a stub — and we mark the rest as
   `blocked` per the 028 spec model for the feasibility-recheck task to
   pick up later."

3. **Generate** — produce the artefact. Working HTML/CSS/JS for tools
   and games. Real curated articles for news. The output is whatever
   the assessment said we could build.

4. **Integrate** — wire the artefact into a page component, add it to
   the site plan, deploy it. Standard build-pipeline work from here.

This is the shape the tools pipeline already implements (mostly).
Extending it to games and news, with a parse stage feeding the
generators richer briefs, is the broader programme.

## What's working today

### Tools — most mature

Five agents, two discovery checks, full lifecycle documented in
`005_tool_pipeline`, `019_tool_library`, `020_tool_lifecycle`.

**Assess** (`tool-suggester`): reads site spec aspects, loads existing
pages and library tools, LLM evaluates what tools would help. Returns 0-5
suggestions, each with name, function, description, target page,
`library_source` (if forkable) or null (if novel), and `related_pages`
for cross-linking. Creates `add_tool` work items routed to either
deployer (library fork) or generator (novel) based on whether a library
source exists.

**Generate**:
- `tool-deployer` (`deploy_tool_to_site` action): library fork. Creates
  a per-site copy of the library template in `content_components` with
  `forked_from` set, plus a tool page and companion guide page.
- `tool-generator` (`create_tool_component` action): LLM-generated novel
  tool. Same downstream wiring (page, page_components, nav entry,
  content work items, companion guide) but the HTML/CSS/JS comes fresh
  from the LLM.

**Improve**:
- `tool-improver`: incremental fixes from issue descriptions. Loads the
  current HTML, LLM rewrites to fix the specified issue, updates the
  template, triggers rerender.

**Quality**:
- `tool_health` discovery check: Tier 1 structural (no script, no
  responsive, hardcoded colours, external CDN). Blockers create
  `improve_tool` items directly.
- `tool-auditor`: Tier 2 LLM code review (logic bugs, mobile, UX, CSS,
  accessibility, dependencies). Findings split by confidence —
  certain/likely auto-fix, possible go to HITL.
- Tier 3 (headless browser visual testing) is planned, not built.

**What tools have that games don't yet have:**
- A library of templates to fork from
- An LLM prompt schema for novel generation
- Quality checks that understand the artefact shape
- Cross-linking back into content pages
- A companion-guide-creation step

**What tools are MISSING relative to the four-stage pattern:**
- **No parse stage.** Adoption captures markdown content but not source
  JS or interactive logic. When adopting a site with calculators,
  tool-suggester evaluates the site's industry from specs and decides
  what tools would fit — it doesn't read the source code of the existing
  tools to model new ones on them. That works fine when the source's
  tools aren't unusual; falls down when they are.
- **Source-tool fidelity is loose.** The new site's tools are
  "appropriate for the industry" rather than "modelled on what the
  source site actually had." Acceptable today; might want tightening
  when fidelity dial is set high.

### Components more broadly

`component-creator` agent generates new component templates from
observed-pattern work items. The prompt includes the full component
contract (structure, naming, template variables, CSS rules, dark
sections, input_schema declaration). Documented in
`007_adoption_pipeline_v4` (component-creator handler section).

Used when a page build sees an unfamiliar section type the library
doesn't cover. Creates a new entry in `content_components`. Different
from tool-generator — produces section components (hero, feature-grid,
testimonial-strip etc.), not standalone tools.

### Adoption — captures content AND extracts structured design data

`007_adoption_pipeline_v4`. Firecrawl crawls source site, returning
both raw HTML and clean markdown per page.

- Per-page markdown content is stored in `research_results`
  (result_type: `adoption_page`).
- Per-page **raw HTML is parsed Go-side** by the
  `extract_design_fingerprint` step using goquery. It extracts hex
  colours, font families from `<style>` blocks and inline styles, CSS
  variable declarations, layout patterns, dark section detection.
- If `<link rel="stylesheet">` references external CSS files, the
  workflow fetches them via the `firecrawl_scrape` adapter and merges
  the parsed data into the fingerprint.
- An LLM step (`generate_design_intent`) reads the fingerprint to
  produce a semantic design brief.
- Stored as `design_reference` (concrete values) and `design_intent`
  (semantic brief) site_specs aspects.

The pattern in place for design — `rawHTML → Go extractor → external
asset fetch → fingerprint → LLM brief → site_specs aspect` — is the
proven template for the parse stage of any other interactive element
type. Extending to interactive logic isn't a new architectural layer;
it's a new extractor following the same pattern.

**What adoption doesn't capture yet for interactive content:**

- `<script>` block contents and `<script src="...">` external JS
  files. The crawl returns these (they're in rawHTML), but no
  extractor reads them, no scrape step fetches the external JS.
- Structural signals indicating interactive elements: `<canvas>`
  elements, animation loop calls, event-handler patterns, form-plus-
  script combinations that signal "calculator" vs "game" vs
  "dashboard".
- A spec aspect equivalent to `design_reference` but for interactive
  logic — call it `interactive_reference` or similar — containing
  per-element structured data extracted from the source.

Closing this gap is a known-shape change, not a wholesale new
capability. The work is a Go extractor (`extract_interactive_fingerprint`
or similar) plus a downstream LLM step (`generate_interactive_brief`),
following the design pipeline's pattern.

### News — pipeline exists, publishing doesn't

`006_news_feed_pipeline_v2`. News sources, ingestion, triage, content
diversity all built. Produces a `latest-news.json` artefact per site.
**Doesn't yet write blog posts.** The pipeline ends at curation; turning
curated items into deployed blog/article pages is the missing piece.
Tool-pipeline noted this gap: "When article writing is added, pass the
site's deployed tool list to the writer's prompt — the rewrite_guidance
pattern is already established."

### Games — nothing yet

No game-suggester, no game-generator, no game-improver, no game-auditor,
no game_health discovery check. The pipeline doesn't recognise games as
a content type:
- `page_type='game'` doesn't exist in the classifier vocabulary
- No work item types for game suggestion or generation
- No library of game templates (analogous to the tool library)
- No site_specs aspect describing games-related strategy
- Game-related components in the corpus (`game-list_pre_037`) declare
  numbered-flat schemas with `source: llm`, forcing fabrication

This is the largest gap relative to the four-stage pattern. Games need
the entire pattern built out.

## Where the parsing/capability work would slot in

Adoption is the natural home for parse-stage work because that's where
source-site content arrives. The pattern for design extraction is the
template. For interactive logic, the work is materially smaller than
"add a new pipeline stage" — closer to "add one goquery selector and
one workflow step that copies an existing one."

**What's already in our hands:**

- **Inline `<script>` content.** Firecrawl's `rawHtml` format returns
  unmodified HTML — every inline script block is preserved. The
  existing `extract_design_fingerprint` already reads `page["rawHtml"]`
  for every crawled page. The scripts are sitting there; the design
  extractor just doesn't look at them. Adding `doc.Find("script")`
  alongside the existing `doc.Find("style")` would capture them with
  no extra fetch.

- **JavaScript-rendered content.** Firecrawl renders JS by default,
  so even SPA games whose canvas is generated at runtime appear in the
  rendered rawHtml — `<canvas>` elements, dynamically-added DOM
  structure, etc. are all visible to goquery.

**What needs one extra fetch:**

- **External JS files** referenced by `<script src="...">`. The same
  pattern the design pipeline uses for external CSS:
  goquery finds the URL → workflow step calls `firecrawl_scrape` on
  it → response's `rawHtml`/`raw_html` field carries the JS source.
  Firecrawl is an HTTP fetcher with content extraction; a `.js` URL
  returns its source the same way a `.css` URL does. The
  `EnrichFingerprintWithCSSAction` already handles a response with
  JS-like content; the parsing differs but the fetch/storage shape
  is identical.

**The actual work for interactive parse-stage:**

- **Extend or add a Go extractor.** New goquery selectors on
  rawHtml: `<script>` (inline content), `<script src="...">` (external
  URLs), `<canvas>` (presence signal), inline event handlers (`onclick`,
  `onsubmit`), forms + script combinations. Token-count, function
  declarations, canvas-API references, requestAnimationFrame mentions
  — all reachable with regex on captured JS, same shape as the design
  extractor's regex bank.

- **Fetch external JS** via the existing `firecrawl_scrape` adapter.
  Same pattern as external CSS. New workflow step parallel to
  `enrich_fingerprint_with_css`.

- **Produce a fingerprint** alongside the design fingerprint, carrying
  per-interactive-element structured data: type signals, source
  pointer (where in the rawHtml the script lives), inputs/outputs
  detected, library/framework signals (jQuery, Three.js, Phaser etc.
  by import-string detection).

- **LLM step** that reads the fingerprint and produces a brief per
  interactive element: what it does, what category, what would be
  needed to recreate it. Stored as a new site_specs aspect (e.g.
  `interactive_reference` for the concrete extracted data,
  `interactive_intent` for the semantic brief — paralleling
  design_reference / design_intent).

The work isn't a new capability layer. It's an extractor (extending
goquery patterns we already use), one workflow step (repeating the
external-CSS fetch shape), and one LLM step. No new adapter, no new
dependency, no new infrastructure. Firecrawl already gives us what we
need; we just haven't asked for it on the JS side yet.

### Firecrawl features worth knowing about

If the basic rawHtml-plus-external-fetch approach hits limits on
heavily-obfuscated or build-bundled interactive content, Firecrawl
has more powerful options worth keeping in mind:

- **`executeJavascript` actions** let us run arbitrary JS in the page
  before scraping and capture return values. The specific call we want
  on the list of wants is `document.querySelectorAll('script')` mapped
  to each element's `src` and `textContent` — that gives us a complete
  inventory of every script element on the rendered page, including
  ones injected dynamically by other scripts that goquery on rawHtml
  would miss. Returns land in `actions.javascriptReturns`.
- **`waitFor`** delays scraping for slow-rendering pages.
- **Structured `json` extraction** with a schema can pull specific
  interactive-element signals out of a page in one call if we want to
  consolidate the goquery + LLM steps later.

These are upgrades, not prerequisites. The starting position is the
plain rawHtml + external-fetch pattern; the `executeJavascript`
script-inventory call is the natural escalation when basic extraction
isn't capturing what we need.

### Capability assessment

The 028 mission doc already lays the groundwork: "the classifier is not
constrained to what can be built today. If the best version of this
site requires a component the library doesn't yet contain or an agent
that hasn't been written, the classifier describes it anyway. Those
items are marked `blocked` in the spec. The `feasibility-recheck` task
promotes them to `planned` when the necessary capability comes online."

That's the model. The new `interactive_reference` / `interactive_intent`
aspects carry each detected element with a capability marker:

- `producible_now` — current generators (tool-generator) can produce
  something equivalent
- `producible_simpler` — current capability covers a reduced version
  in the same league
- `blocked` — no generator exists for this type yet (games today)

The planner reads these and queues the right work items. `producible_now`
becomes `add_tool` work items. `producible_simpler` becomes
`add_tool` with a "simpler variant" flag in the spec.  `blocked` stays
in the spec until the relevant generator exists, at which point
feasibility-recheck moves it forward.

Capability assessment isn't a separate agent — it's a property of the
spec lifecycle that runs at adoption time and re-runs whenever
generators change. The work is: extend the spec aspects to carry
interactive-element requirements, extend feasibility-recheck to know
about generators beyond tool-generator, extend the planner to read
these requirements.

## Cross-cutting concerns

### Classification vocabulary

`page_type` is a small fixed vocabulary today (`tool`, `blog_post`,
`content`, `index`, `entity_directory`, etc.). When we need to recognise
games as a page type, the current options are: hardcode `game` into the
classifier (Option 1 — incremental), let the classifier produce any
string and tolerate drift (Option 2 — loose), or move the vocabulary
into a `page_types` table (Option 4 — declarative).

The pragmatic next step when adding `game` is Option 1 — add the type,
add the recognition, add downstream consumers. The longer-term move is
Option 4. Either way: see `FOCUS_component_schema_patterns` for the
related concern that page_type values currently drift between
hyphenated and underscored forms; canonicalise before extending.

### Generator architecture

`tool-generator` is the prototype for "LLM produces interactive HTML/JS
from a brief." If we extend the pattern to games, news articles,
dashboards, etc., they share enough structure that some convergence is
likely. Each generator needs:

- A spec contract for the brief (what does it need to know?)
- An LLM prompt template (what does it produce?)
- A persistence action (where does the output go?)
- A page-creation step (how does it become a page on a site?)
- Quality checks (Tier 1 structural, Tier 2 LLM, Tier 3 visual)
- A companion-content step (does this artefact need a guide page?
  An explanation? Source citations?)

Each new generator could be written from scratch, or there could be a
shared `interactive-artefact-generator` base with content-type-specific
specialisation. Worth considering once two more generators exist; one
isn't enough to abstract from.

### Library model

Tools have a library — canonical templates in `content_components` with
`forked_from IS NULL`. Sites get forks. New tools can be added by
LLM generation when no library template fits.

This model translates cleanly to games: a library of game templates
(small playable mechanics), with sites getting forks customised to
their content/style. A library of news article templates would be
shallower (article structure is more uniform).

The cost of adding a new library is small (a `component_level` value
plus the discovery and quality checks specific to that type). Worth
copying the tool library shape when extending.

### Quality model

Tools have three tiers (structural, LLM, visual). Games would have the
same shape but different specifics — Tier 1 for games would check
playable surface (canvas element, animation loop, input handler), not
just script presence. Tier 2 would check game mechanics (does the
game terminate? does scoring work? does it have a fail state?). Tier 3
visual testing helps both.

This is parameter-tunable, not new architecture.

## Sequencing — agreed order

The agreed order of work, locked in 2026-05-14:

**1. Path C — Parse stage in adoption.** Extend the design-extraction
pattern to interactive logic. Inline `<script>` content is already in
`page["rawHtml"]` — no extra fetch needed. Add `<script>` to the
goquery selector bank, repeat the external-CSS fetch pattern for
`<script src="...">`, add an LLM step to produce
`interactive_reference` / `interactive_intent` site_specs aspects.
Test-bed: readopt gamesdesign.co.uk after this lands.

**2. Path D — Tool reliability.** Tools currently render and deploy
but the interactive behaviour isn't working. The tools pipeline is
the prototype for all subsequent generators — if it's flaky, that
flakiness propagates to games and any future content type. Worth
nailing down before extending.

**3. Path A — Games as the next content type.** Build the games
pipeline by copying the tool pattern (now reliable) and applying it to
playable artefacts. game-suggester, game-deployer (library forks),
game-generator (novel), game-improver, game-auditor. Seed a small
library of game templates. Adds `game` to the classifier vocabulary
and unblocks game-list resolution.

**Subsequent paths, no specific order yet:**

- **Path B — News publishing.** Turn `latest-news.json` into deployed
  blog posts. Smaller surface area; connects two existing pipelines
  (news ingestion + page deployment).
- **Path E — Component anti-pattern cleanup.** Migrate the 25
  numbered-pattern components from `FOCUS_component_schema_patterns`
  to items-array shape. Housekeeping that affects display of all
  content types.
- Other paths as they surface.

### Path detail kept for reference

The descriptions below remain for context on each path. The order
above is what we work to.

### Path A — Games as the next content type

Build the games pipeline by copying the tools pattern: game-suggester,
game-deployer (library forks), game-generator (novel), game-improver,
game-auditor. Add `game` to classifier vocab. Seed a small library of
game templates. Same shape as tools, applied to a different artefact
type. Largest pipeline gap closed; teaches us what generalises and what
doesn't between tools and games.

### Path B — News publishing

Close the news pipeline gap: turn `latest-news.json` into deployed blog
posts. Pattern is closer to content writing than tool generation — uses
the existing page-content-writer with a richer per-section prompt and
news-feed input. Smaller surface area than games. Connects an existing
pipeline (news ingestion) to an existing pipeline (page deployment).

### Path C — Parse stage in adoption

Extend the design-extraction pattern to interactive logic. Inline
`<script>` content is already in `page["rawHtml"]` from every crawled
page — no extra fetch needed. Add `<script>` to the goquery selector
bank, repeat the external-CSS fetch pattern for `<script src="...">`,
and add an LLM step to produce `interactive_reference` /
`interactive_intent` site_specs aspects. Capability verified against
the firecrawl docs (rawHtml format returns unmodified HTML; external
URLs fetched the same way as CSS). Smaller piece of work than I first
thought — closer to a couple of days than a couple of weeks.

### Path D — Tool quality and reliability

Tools render and deploy but reportedly "currently don't work" — exact
failure mode TBD. Investigating and fixing existing tool behaviour
might be more pressing than extending to new content types. Useful
because the tools pipeline is the prototype for all subsequent
generators; if it's flaky, that flakiness propagates.

### Path E — Component anti-pattern cleanup

The 25-component numbered-pattern audit (see
`FOCUS_component_schema_patterns`). Migrating these to items-array
shape unblocks any planned section that would use them. Not
content-type work — it's schema-shape work that affects display of
all content types. Probably belongs alongside the planner-level
guard against zero-resolved-items sections.

My honest read: Path D (fix tools first) followed by Path A (games
next) gives the cleanest progression — confirm the prototype works,
then extend. Path C is the bigger payoff but needs more groundwork.
Path B is the smallest piece. Path E is housekeeping.

But these are options, not recommendations. The decision depends on
which gap is currently most blocking.

## What this document is for

When work resumes on building real interactive content — games, better
tools, news publishing — start here. The four-stage pattern is the
reference shape. The "what's working today" section says what to copy
vs build fresh. The sequencing paths are starting points for the next
focused session.

This document grows. As the parse stage gets built or games come online,
update the "what's working today" section. As capability assessment
becomes a real agent, document its inputs and outputs here. The aim is
a single place to read for "where are we with interactive content
generation right now."

## Pointers to related docs

- `028_platform_mission_and_pipeline_direction` — the *why*
- `005_tool_pipeline` — tools end-to-end, deployed and working
- `019_tool_library` — tool storage, fork-on-deploy, library inventory
- `020_tool_lifecycle` — tool agents, discovery checks, work item flow
- `022_dynamic_applications` — three-tier ambition, framework targets,
  publishing adapters
- `006_news_feed_pipeline_v2` — news ingestion (publish step still open)
- `007_adoption_pipeline_v4` — source crawling and classification
- `021_site_spec_and_classifier` — spec aspects and what they carry
- `FOCUS_component_schema_patterns` — items-array contract, numbered
  anti-pattern, normalisation drift
- `FOCUS_language` — language-awareness reference (small adjacent
  concern)
