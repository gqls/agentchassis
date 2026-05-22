# FOCUS — Adoption Fidelity and the Clone Variant

Date: 2026-04-22 (origin); standing reference for the adoption-fidelity gap.
Status: source/destination separation deployed and working. Output fidelity is
the open problem — the pipeline produces a planner *brief*, not a faithful copy.

> Distilled from `HANDOFF_2026-04-22_adoption_phase1_and_fidelity.md`. The
> session log (artefact lists, next-session prompt) is dropped; the decisions,
> the variant framing, and the ranked problems are kept because they define what
> "faithful adoption" needs and remain unbuilt.

## The core gap

Adoption as currently built produces **a site-planner brief plus specs, not a
deterministic copy.** It:

1. Extracts a design fingerprint (CSS variables, typography, palette, layout hints).
2. Writes a `design_intent` spec — an LLM *interpretation* of the design, phrased
   for a designer.
3. Writes `identity`, `content_direction`, `archetype` specs — LLM-summarised character.
4. Creates `needs_composition` / `needs_design` work items with `mode: "recreate"`.
5. Creates pages (where the crawl provided them) with `existing_content.raw_markdown` attached.

Then the **build pipeline** (site-design-planner, webdesign-agent,
page-content-writer) takes those specs and builds a site. It doesn't render the
crawled HTML directly; it re-plans and re-generates from specs. The mismatch is
the gap between *specs + LLM interpretation* and *the actual source site*. The
pipeline works as designed; the design is too loose for "I own this, keep it
essentially the same."

## The adoption variants (from `FUTURE_adoption_source_destination_separation.md`)

- **Variant A** — reference-only (design inspiration, fresh content)
- **Variant B** — design + structure (same pages, your content)
- **Variant C** — full clone (copy everything, rename)
- **Variant D** — multi-source analysis

Phase 1 plumbing (`target_url` / `destination_domain` separation) is deployed,
but the variant selector was never wired — everything defaults to the current
behaviour, which sits roughly between A and C and commits to neither. For the
user's case (own the source domain, want a near-replica), **Variant C is what's
needed and does not yet exist in a meaningful sense.**

## What the first real run showed (gamesdesign.co.uk from gamedesign.uk)

The deployed site read as a generic consultancy/brochure site wearing the
source's colours. Concrete failures observed in the rendered HTML:

- **Header is a generic brochure template** — `/services.html`, `/about.html`,
  `/case-studies.html`, `/start-a-project.html` nav items that never existed on
  the source (which was guides, tools, games).
- **Identity drifted to "consultancy."** Hero became "Game Design Consultancy
  and Co-Development." The source is an educational resource for game designers,
  not a consultancy. The `derive_content_direction` / `classify_archetype`
  prompts nudged toward a business model that doesn't match.
- **Tool pages have prose, not tools.** The crawled pages had interactive JS
  calculators; the rebuild has paragraphs *describing* what the calculator would
  do. Adoption pulled the markdown but not the interactive machinery.
- **Footer lists invented pages** (`/start-a-project.html`, `/case-studies.html`).
- **Empty `<main>` on the tools page** — page record created but the build
  pipeline lacked spec data to populate it.

## Problems, ranked by impact

Each is substantial; "fix adoption to produce a near-copy" is a week-or-more
project, not a one-SQL patch.

1. **Identity drifted (root cause).** The LLM summarised the source as
   "consultancy." This is the root spec the rest of the pipeline plans against —
   fixing it is upstream of problems below. `derive_content_direction` /
   `classify_archetype` output is the lever.
2. **Nav doesn't match source.** `populate_nav_tables` builds from an
   auto-generated brochure archetype instead of the adopted pages — the planner
   defaulting to its standard brochure archetype. (See also the planner-ignores-
   adopted-state diagnosis in `FOCUS_planner_ignores_adopted_state.md` /
   doc 029 Phase 1.)
3. **Tool pages have content but not tools.** Adoption saved markdown but not the
   `<script>` / `<canvas>` interactive elements. `tool-recreation-handler` exists
   (doc 002) but isn't triggered, or doesn't yet work for these tools. This is the
   subject of the design-fingerprint / interactive-fingerprint work (Path C —
   see `FOCUS_component_schema_patterns.md` and the baseline-queries note).
4. **Some pages empty.** Crawl URL list created page rows but content wasn't
   attached — likely the `matchCrawlContent` lookup missing them. The crawled
   pages are nested under `crawl_result.response.body.data.pages`, not directly
   under `crawl_result.pages`; the Phase 1 fix keyed the lookup on `sourceDomain`
   but the pages map shape may still be misaligned.
5. **Design inherited colours but not structure.** Dark-section colours came
   through; header style and section shapes didn't. The design-planner filled in
   its preferred components around the adopted colours.

## What Phase 1 deployed (source/destination separation) — for context

The adoption flow previously conflated source and destination into one `site_id`.
Phase 1 parameterised it: `target_url` = what to crawl, `destination_domain` =
what to build (legacy `url`/`domain` still accepted). Deployed via migrations
`001`–`004` (separation, the `site-adoption-orchestrator` wrapper, input-mapping
fixes, optional-legacy-field markers) plus `EnsureSiteRecordAction` /
`apply_adoption_plan_action.go` changes (the `sourceDomain` vs `domain` fix that
was silently dropping page content when source ≠ destination).

The **wrapper-orchestrator pattern** also landed here: `site-adoption-agent` was
the outlier running in-chassis; it now runs under a `site-adoption-orchestrator`
wrapper (`spawn_adopter → call_adopter → complete`), modelled on
`med-export-orchestrator`. Baseline rule recorded in `001_development_guide.md`:
every pod-running agent needs a parent that spawned it; coordination work
(conditionals, HITL, spawn/call) is fine in-chassis, substantive work (LLM,
crawls, heavy DB) must be in a spawned pod.

## Deferred / out of scope (with rationale)

- **Resume logic.** `orchestration_states.collected_data` already persists
  per-step output (378KB survived a failed run), but `ResumeWorkflowTopic` has no
  subscriber — resume was anticipated, never built. User: "a new crawl is fine."
- **med-\* wrappers have the same double-wrap bug.** All four
  (`med-price-scrape/url-discover/url-map/export-orchestrator`) use the broken
  `{"input_data": "input_data"}` mapping. Either their children tolerate the extra
  nesting or they aren't invoked in production. Not blocking; spot-check when next
  in that area.
- **Stale `crawl_result` shape.** `crawl_result.response.body.data.pages` observed
  but not fully mapped; relevant if resume is built or if problem 4 is a
  pages-lookup issue.
- **`fetch_primary_css` brittleness.** User wants CSS-fetch failure to be a hard
  fail ("it is important"), so no workflow-level skip-on-failure. A firecrawl
  timeout downstream still hard-fails; the surviving state isn't reusable without
  resume. Accepted trade-off. (Note: the separate `error_step` fix in
  `FOCUS_chassis_config_location_bugs.md` later made `fetch_primary_css` able to
  route to `analyze_site` with a partial fingerprint rather than dying — that
  changes this trade-off; cross-check the two.)

## References

- `FUTURE_adoption_source_destination_separation.md` — the variant definitions; `clone` (Variant C) = faithful-copy intent
- `FOCUS_planner_ignores_adopted_state.md` — why the planner overlays a generic skeleton (problem 2's mechanism)
- `FOCUS_adoption_faithfulness_via_locks.md` — the timed-lock mechanism that keeps a faithful first pass faithful once it exists
- `FOCUS_chrome_templates_and_page_shape.md` — the chrome/page-shape half of the nav-pollution problem
- `FOCUS_component_schema_patterns.md` — Path C interactive-fingerprint work for problem 3 (tools rebuilt as prose)
- `007_adoption_pipeline_v4` — adoption pipeline doc (two-agent arrangement, modes, what-to-watch)
- Code: `EnsureSiteRecordAction`, `apply_adoption_plan_action.go` (`sourceDomain` fix, URL divergence at lines ~52169-52176)
- Origin: `HANDOFF_2026-04-22_adoption_phase1_and_fidelity.md`, first real run correlation `2de43823-a497-409f-b14c-5a36bd412ad8`
