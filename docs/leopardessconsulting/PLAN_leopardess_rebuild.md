# PLAN — leopardessconsulting.co.uk rebuild

**Started:** 2026-07-09 · **Owner:** uk@websy.uk · **Site ID:** `4851f6fc-71cf-4160-a270-e03d6d3e0732`
**Evidence base:** `AUDIT_verified_facts.md` (read that first — nothing here is asserted without a row there)
**Companions:** `RUNNING_NOTES.md` (updated every turn) · `RUNBOOK.md` (human tasks)

Phases are numbered **L0–L9** to avoid collision with the imagery programme's I0–I8
and the fixloop's F0–F3.

---

## 1. The brief, as given

Redeploy the site to be the best it can be. No inventing projects or staff; we may
say what we *might* be able to do provided it is not pie in the sky. A great
leopardess logo and branding consistent across the site, graphics to match, mobile
friendly. Tools, news, useful guides, games, and example AI working frontends —
simulations are acceptable at this stage. Fewer panels repeating the same thing;
three cards per row and no orphan card on the next line. Illustrations or images
that explain what the cards link to; infographics that explain the guides and news
visually. Copy that does not sound like an LLM: no negative framing, no claims that
are too bold, and a willingness to use three or four more words when they make the
meaning clearer rather than always reaching for the one perfect word.

**Reader we are writing for.** An intelligent future buyer of AI for their business.
Commercially sharp. Not an AI specialist. They have heard the hype and are not
convinced. Our job is to keep our promises small and deliverable, and to stay out of
the hype fest.

## 2. Owner decisions taken (2026-07-09/10)

| # | Decision | Consequence |
|---|---|---|
| **A1** | "Peter Grenfell" is invented → **delete**. The `identity.team[0]` background (30 years, ex-Bumble, worldsoccernews.com) is the owner's own. | About page written in the first person from the real background. Name to be confirmed by owner (RUNBOOK H1). |
| **A2** | **Pivot the audience** to the sceptical business buyer. | The `identity` spec's "non-technical SMB buyers are explicitly out of scope" clause is revoked. Technical depth becomes evidence available on demand, not the register of every sentence. |
| **A3** | Logo is a **stylised leopardess head in profile** — minimal strokes, single weight, spots implied rather than drawn. | Overrides `design_intent.avoid` ("literal leopard or animal imagery"). That line must be edited before any generation, or the imagery-direction prepender will fight the prompt. |
| **A4** | **Dark chrome, light reading surfaces.** Dark hero/header/footer/demo panels; warm off-white for guides, news, long-form. | Requires a **site-specific palette fork** — the current collection is shared with three other sites. |
| **A5** | Build a **reusable chart component: Go + a JS renderer**, honouring prior decisions D1/D3 and Phase I4. | See §5. Real numbers only, from the database. |

## 3. Standing rules for this rebuild

1. **No claim ships without a row in `AUDIT_verified_facts.md`.** If a sentence makes
   a factual claim and there is no evidence row, either verify it and add the row, or
   cut the sentence.
2. **Verify by artifact, never by report.** A `complete` work item means nothing. Check
   the DB row, `curl` the asset, read the rendered HTML. (This platform has a long
   history of builds reporting success while building nothing.)
3. **Our own sites are not client work.** The eight deployed sites demonstrate the
   platform. They are never to be implied to be a client roster.
4. **Say the smaller true thing.** Where a claim is nearly true, state the smaller
   version that is exactly true. "We ran 2,767 business records through Companies
   House verification" beats "thousands of verified records" and is far better than
   "enterprise-scale data enrichment".
5. **Voice.** Positive framing. No swipes at competitors, no "not X, but Y" where X is
   a strawman. Prefer a clear extra clause over a clever compression. Never use a word
   like "leverage" or "unlock" when "use" or "open" is what is meant.

## 4. Phases

### L0 — Evidence and setup *(done)*
Audit complete. Docs directory established. Four owner decisions taken; one correction
issued (the live site is dark, not light — see RUNNING_NOTES turn 3).

### L1 — Truth pass on the specs *(highest leverage; do first)*
The specs are what the agents read. While they contain fabrications, every future
agent run reproduces them. Before any content work:
- `identity`: delete `leadership_team` (Peter Grenfell). Delete `departments` (the
  8-department taxonomy and its per-department agent counts are fabricated — U1).
  Correct `capabilities_summary` ("70+ agents in 8 departments" → the true figure:
  143 agent definitions, 56 active). Rewrite `tagline` (D4/X3). Revoke the
  "non-technical SMB buyers out of scope" clause (A2).
- `voice`: rewrite for A2. Add the positive-framing and plain-language rules.
- `design_intent`: remove the animal-imagery ban (A3); reconcile `color_scheme` to the
  real dark chrome + light reading surface (A4, D6).
- `portfolio`: reframe the four "case studies" from invented client engagements (U3)
  into honestly-labelled demonstrations of the platform, each anchored to an
  `AUDIT_verified_facts.md` row.
- `sites` row: fix the stale `tagline` (D4) and set `logo_url` once L2 lands.

**Acceptance:** a fresh read of every current spec contains no statement that
contradicts `AUDIT_verified_facts.md`.

### L2 — Brand: one logo, used everywhere
- Fix the imagery routing so a `logo` reaches a model that can actually draw one.
  Today `kind=="icon"` → Banana (Gemini `gemini-3-pro-image-preview`), everything else
  → Stability SDXL. SDXL is the wrong tool for a flat vector-like mark, and it ignores
  reference images, which is precisely the mechanism brand consistency depends on.
  **Route `logo` (and `illustration`, `infographic`) to Banana** in
  `internal/adapters/imagegenerator/dynamic_adapter.go`.
- Commission **one** canonical logo as `asset_key='logo'` so it deploys to the fixed
  path `/assets/images/logo.png` via `storage.DeployedWebPath`. Set
  `sites.logo_url='/assets/images/logo.png'` (the convention finetuning.uk and
  gaswholesalers.com already prove).
- Delete the three dead asset rows first (D1). They are unrepairable: presigned,
  expired, `storage_path` empty, and they predate commit `84f07d38`.
- Favicon and OG card are **not** pipeline capabilities. They are manual derivations
  from the logo. Wire `<link rel="icon">` and `<meta property="og:image">` into the
  `head` site component.
- Use the approved logo as the **reference image** for all subsequent site imagery, so
  the graphic language is one family (this is why L2 must precede L6).

**Acceptance:** `curl` the logo path → HTTP 200, correct bytes. Favicon renders in a
browser tab. `assets.url` is a git path, `storage_path` non-empty.

### L3 — Palette fork and the leaking slots
The deployed `styles.css` is a build artifact reproducible from no current DB row.
Core slots were overridden to dark/gold; **specialised slots were not**, so the live
site serves white cards, a navy header and footer, and a blue gradient CTA on a
black-and-gold page. That is the single worst visual defect.
- Fork a **site-specific palette** (`fork_theme_composition.go` /
  `install_site_composition`). Do **not** edit the shared seed — it dresses four sites.
- Set `card_bg`, `header_bg`, `footer_bg`, `cta_bg`, `hero_title`, `primary_hover` to
  the leopardess charcoal/gold system, and introduce the light reading surface (A4).
- Re-render CSS from spec and deploy, so the artifact becomes reproducible from source.

**Acceptance:** no `#1e40af`, `#0f172a` or bare `#ffffff` card in the deployed CSS.
`design_intent.color_scheme` re-rendered produces byte-identical output.

### L4 — Layout: three per row, no orphans
Neither of the two rules currently in play achieves this. Global CSS says
`repeat(3, 1fr)` (leaves a short, left-aligned last row); most components override with
`repeat(auto-fit, minmax(280px,1fr))` (**stretches** an orphan card to full width).
- Choose card counts divisible by three wherever the content allows — this is a
  *content* fix as much as a CSS one, and it serves the brief's "not too many panels
  saying the same thing".
- Where a count of 3n is not natural, add explicit last-row rules
  (`:last-child:nth-child(3n-2)` centring) so a remainder is centred, not stretched.
- Consolidate the fragmented per-component breakpoints onto the global 1024/768 scale.
- Note `case-studies-grid` is **hard-wired to five cards** (`card1_…card5_…`). Five
  cards cannot be 3-up without an orphan. Use a different component or supply six.

### L5 — Content and copy
Rewrite against A2 and the §3 rules. Cut the repetition: the site currently says the
same "production-grade agents on Kubernetes" thing on the index, services, our-approach,
how-it-works, how-we-work, technical-architecture and for-engineering-teams pages.
- Merge the near-duplicate pages rather than restyling all of them.
- Rewrite the four portfolio entries as *what we have built and can build again*,
  each traceable to an audit row.
- Delete the invented leadership entry; write About in the first person from the real
  background (A1).
- Fix nav labels (D3): they are currently raw `<title>` strings with `| Leopardess
  Consulting` appended.

### L6 — Imagery that explains
Only after L2 (so everything is generated against the logo as reference image).
- A card image per link card that reflects the destination, and a sibling image on the
  destination page, so the two read as one family (imagery goal G3).
- Illustrated guides. Concept diagrams for the pipeline explainers.
- Performance budgets: WebP, responsive sizes, lazy loading (G7).
- Alt text on every meaningful image (G8).

### L7 — Charts (Phase I4, adapted)
See §5 below. The one place we will build genuinely new platform capability.

### L8 — Tools, guides, news, games, demos
The tool library already holds reusable, genuinely interactive components — verified
active: `tool-agent-complexity-estimator`, `tool-ai-agent-roi-estimator`,
`tool-ai-data-risk-checker`, `tool-ai-readiness-quiz`, `tool-ab-test-calculator`,
`tool-bayesian-ranking`, archetype quizzes, `tool-favicon-generator`, `tool-bg-remover`.
These are deterministic client-side widgets — **not live model inference**. They are
honest as "simulations" and must be labelled as such where a visitor might assume
otherwise.
- A "game" has no formal existence in the platform; it is simply a
  `component_level='tool'` component. That is fine — build one.
- News: the feed pipeline is real and running (5,652 items, 4,672 scored). Surface it.
- Guides: illustrated, each paired with a tool where one exists (the tool-deployer
  already creates companion guides and cross-links).

### L9 — Deploy and verify
`rerender-pages` → git commit → GitHub Actions → Backblaze B2, fronted by a Cloudflare
Worker. **Not Cloudflare Pages** — the site copy must not say so.
Verify: every page 200s, logo/favicon/OG resolve, no orphan cards at 1440/1024/768/375,
Lighthouse mobile, and a link check.

---

## 5. L7 in detail — the chart component (honouring D1, D3, I4)

**Prior decisions to honour** (`RUNNING_NOTES_imagery_best_in_class.md` L26–36,
`PLAN_imagery_best_in_class.md` L126–128, L229–235, `old/FUTURE_data_graph_pipeline.md`):
- **D1** Data graphics are code-rendered from real data. Diffusion never plots data.
  The LLM proposes the story; **the code owns the numbers.**
- **D3** `chart` is *not* a `site_plan_imagery` kind. Charts are Lane-B artefacts,
  stored as assets with their own purpose. Keeps diffusion away from data.
- **Confirmed 2026-07-08:** chart runtime is **go-echarts, in-chassis**.
- **PLAN §6:** *"static SVG/PNG must always exist as fallback so charts survive in
  feeds and OG cards."*

**The conflict, stated plainly.** go-echarts renders an HTML page that loads
`echarts.min.js` and draws in the browser. It has no server-side SVG or PNG output;
producing a static image from it requires driving a headless Chrome
(`snapshot-chromedp`). Committing a headless browser to an agent pod is a heavy
dependency for a decorative fallback. So "go-echarts, in-chassis" and "static SVG must
always exist" cannot both be satisfied by go-echarts alone.

**Proposed resolution** (owner to confirm — RUNBOOK H2):
Split the two requirements across the two runtimes that are each naturally good at one
of them, which is also exactly what the owner asked for ("go plus a js renderer"):

- **Go emits the static SVG.** A small, dependency-free SVG emitter in the chassis takes
  a typed series (`dates|labels + values + units + source attribution`) and writes a
  clean, accessible `<svg>` — axes, ticks, labels, a caption, and a source line. This
  is the artifact stored as an asset and used in feeds, OG cards, and no-JS contexts.
  It satisfies PLAN §6 with no new dependency and no browser.
- **A JS renderer progressively enhances it.** Where a component supports interactivity,
  the same typed series is emitted as JSON and a small self-contained renderer upgrades
  the inline SVG in place (hover readouts, toggles). Self-contained and inline, matching
  how every existing tool component ships its JS — **no CDN**, because the deployed
  tools already pull `unpkg` for Lucide and that is a dependency I would rather not add
  to a second place.

This keeps D1 exactly (Go owns the numbers, in both paths), keeps D3 (Lane-B asset, no
new diffusion kind), and delivers the confirmed *intent* of the go-echarts decision —
in-chassis, real interactive charts — while meeting the static-fallback requirement the
same plan sets. If the owner prefers literal go-echarts, the cost is a headless-Chrome
sidecar; that is a real option, just a much bigger one.

**Data.** First charts use numbers this project can prove, from its own database:
2,767 verified business records; 937 Companies House-enriched; 5,652 feed items with
4,672 credibility-scored; 143 agent definitions, 56 active; 75,061 orchestration state
rows; 8 deployed sites. Every chart carries its source line and the query date. No
external data source is needed for the first cut, which sidesteps RUNBOOK B4
(API keys) entirely.

**Scope discipline.** L7 delivers: a Go SVG emitter, a chart component template with
caption + source line, a JS enhancement layer, and three real charts on the site. It
does **not** deliver the `data-chart-generator` agent, external data fetching, or the
LLM annotation layer. Those remain Phase I4 proper.

---

## 6. Explicitly not in this plan

- Live model inference in the browser. The interactive demos are simulations and will
  say so.
- Any client logo, testimonial, or named engagement. We have none.
- Claims about uptime, revenue impact, or headcount.
- `data-chart-generator` agent and external data APIs (deferred to I4 proper).
- Editing the shared seed style collection (it dresses four sites).

## 7. Risks

| Risk | Mitigation |
|---|---|
| An agent run silently reproduces a fabrication from a spec I have not yet fixed | L1 precedes all content work, by construction |
| `needs_rebuild` pages never rebuild | Verified root cause: `needs_rebuild` is **inert** without a `site_work_items` row. Nothing scans `pages`. Insert items explicitly. |
| A work item reports `complete` having built nothing | Verify by artifact (§3.2). Never trust status. |
| `apply_section_edit` sets `build_status='approved'`, making an edited section invisible to every discovery check | Known landmine. After any section edit, set `deployed` then clear `schema_mode`/`locked_at`/`locked_by` in a second statement. |
| Palette edit restyles three other sites | Fork first. Never touch collection `3196d966`. |
| Regenerated logo drifts between images | Logo generated first, then used as `reference_image_uris` for all later imagery — which only works on Banana, hence the L2 routing change. |
