# Automatic news-vertical detection & sourcing — design (careful)

Status: **DESIGN for review** (2026-07-16). Emerged from the relojistas rebuild —
`verticalNewsMap` had no watch/horology entry, so `evaluate_news_feed` would have
defaulted the site to **no news feed** and we had to force the recommendation by hand.
This documents how to make the decision (and, harder, the *sourcing*) automatic for any
new domain/vertical without a code change per vertical — and where the sharp edges are.

## Current mechanism (what exists today)

- `evaluate_news_feed` (feed_news_recommendation_action.go) runs after classification,
  before the planner. It looks the site's `industry / site_type / category / domain`
  up in a **hardcoded Go map** `verticalNewsMap` → `{recommended, reason,
  vertical_keywords, source_types, separate_page}`, deep-merged into the classification
  spec's `content_features.news_feed`.
- The planner keys on `recommended=true` (+ `separate_page` → a `/noticias`-style page);
  `content-feed-orchestrator` → `seed_content_sources` reads `source_types` +
  `vertical_keywords`.

### Two gaps this design must close
1. **Recommendation gap.** The map is keyword/substring-brittle and English-centric. Any
   vertical not enumerated (watches, cycling, wine, aviation, tabletop games…) silently
   falls through to `recommended:false` — **no news feed, no signal that one was skipped.**
   Adding a vertical = code edit + chassis rebuild + redeploy.
2. **Sourcing gap (the harder one).** `SeedContentSourcesAction` **deliberately skips
   `rss` and `scrape`** ("requires manual URL config") — it can't invent feed URLs. So even
   a *recommended* vertical only gets `api_news` + `news_search` automatically; real RSS
   feeds (the highest-quality, lowest-cost source, and the thing that reactivated
   relojistas' subscribers) still need a human to find and verify them. Automation that
   stops at "recommend" is only half the job.

## Design principles ("carefully")

- **Never fabricate a feed URL.** An LLM may propose *publication names* and *keywords*;
  it must not emit `rss` `feed_url`s that then get seeded blind. Feeds are **discovered and
  verified** (fetched, parsed, recency-checked) before seeding — exactly the manual pass we
  did for relojistas. Triage rejects fabricated URLs downstream, but we don't want to rely
  on that.
- **Deterministic curated path wins.** Known verticals stay in a fast, reviewable,
  reproducible table; the LLM only fills the long tail. Curated entries override LLM.
- **Keep the build hot-path cheap.** LLM judgement runs **only on a map/table miss**, once
  per site, temperature 0, schema-constrained. Feed *discovery* runs **async** (a
  discovery-check / scheduled step), not inline in the build — news is two-pass anyway, so
  feeds arriving minutes later is fine.
- **Propagate language + region** end-to-end (classification.language / geo) so keywords,
  provider prompts, and discovered sources match the audience (relojistas = es; ES/CL/MX).
- **Safe defaults.** On low confidence, prefer a homepage snippet (`separate_page:false`)
  over a full section; better to under- than over-commit a page that may sit thin.

## Phased plan

### Phase 1 — DONE (2026-07-16): add watch/horology to the map
`watchHorologyNews` shared config, aliased `watch/watches/horology/watchmaking/reloj/relojes`
(the Spanish substrings catch relojistas.com etc.). Deterministic, immediate for that class.
**Takes effect only after a chassis rebuild+redeploy** (images build from local FS via the
Makefile) — it did NOT affect the in-flight relojistas build (already past classification;
that one used the forced spec).

### Phase 2 — LLM fallback in `evaluate_news_feed` (biggest leverage, contained)
On a map/table miss, call the LLM with the classification (industry, site_type, category,
description, audience, **language**) → strict JSON `{recommended, reason, vertical_keywords
(in site language), source_types ∈ {rss,news_search,api_news,scrape}, separate_page}`.
- Guardrails: only on miss; temp 0 + schema-validated; **no feed URLs in output**; one call
  per site (cache/persist — Phase 3). Reuses the existing LLM step infra the classifier uses.
- Result: any vertical gets a *reasoned* recommendation + language-correct keywords, so
  `news_search`/`api_news` seed automatically. Turns "silently no feed" into "always a
  considered decision."

### Phase 3 — DB-backed `news_verticals` table (self-learning, no redeploy)
Relocate the map into a table (bootstrap-seeded from today's map). Curated rows = fast
deterministic path; Phase 2's LLM decisions are **written back** keyed by normalized
vertical (status `suggested` → `active` on review, with an audit trail). The second site in
a new vertical hits the curated row — LLM cost amortized to once *per vertical*, not per
site. Verticals become editable/tunable without a code change or redeploy.

### Phase 4 — verified RSS feed-discovery agent (turnkey sourcing; highest value, most work)
The piece that makes feeds automatic **safely**. A `discovery_checks`-style async step:
given `vertical_keywords` + language + region →
1. **Search** for candidate publications (web-search adapter, language/region-scoped).
2. **Probe** feed URLs per candidate: `/feed/`, `/rss`, `/atom.xml`, and the HTML
   `<link rel="alternate" type="application/rss+xml">` in `<head>`.
3. **Validate**: FETCH each candidate → well-formed RSS/Atom XML, ≥N items, most-recent
   within M days, language matches. (This is exactly the manual pass that verified 5
   relojistas feeds and rejected 3 stale/404 ones.)
4. **Seed only verified** feeds as `content_sources rss` rows; log rejects with reason.
Re-validate on a schedule; existing `content_sources.error_count`/`last_error` already
supports dead-feed pruning. Natural home: alongside `check_news_feed.go` in
`discovery_checks/`, and it retires the standing "Discovery agents… should add these later"
TODO in `seed_content_sources_action.go`.

## Recommended build order & caveats
Phase 2 → Phase 4 → Phase 3. Phase 2 is the contained automation win; Phase 4 is the
turnkey-sourcing win (and the safety-critical one); Phase 3 is the durability/relocation
win. Each needs review + a chassis rebuild/redeploy. All three keep the same gate the
planner already reads (`content_features.news_feed.recommended`), so they're additive and
per-site reversible via data.

## Cross-references
- `feed_news_recommendation_action.go` (map + Phase 1 edit), `seed_content_sources_action.go`
  (rss/scrape skip + the discovery TODO), `discovery_checks/check_news_feed.go` (existing
  detector; Phase 4 sibling), concept register `news-feed-pipeline.md`.
- The manual analogue this generalizes: relojistas RSS verification (running notes 2026-07-15).
