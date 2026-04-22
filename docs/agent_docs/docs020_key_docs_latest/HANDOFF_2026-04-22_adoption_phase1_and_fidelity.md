# HANDOFF — Adoption Source/Destination Separation + Output Fidelity

**Session date:** 2026-04-22
**Status of Phase 1 plumbing:** Deployed, tested end-to-end, working.
**Status of output quality:** First real run produced a deployed site but with significant fidelity issues. Needs a dedicated next session.

---

## What was done in this session

### Phase 1 — source/destination separation (deployed)

The adoption flow previously conflated two concepts into one `site_id`. If you submitted `https://competitor.com` as adoption input, the system crawled it *and* built a new site called `competitor.com`. No way to say "crawl X as inspiration, build my site Y."

Phase 1 implements Option 1 from `FUTURE_adoption_source_destination_separation.md` — parameterise inputs. `target_url` is what to crawl, `destination_domain` is what to build. Legacy `url`/`domain` still accepted for backward compat.

**Deployed changes:**

1. **`001_adoption_source_destination_separation.sql`** — workflow JSON update for `site-adoption-agent`:
   - `crawl_site.url_field` → `input_data.target_url`
   - `ensure_site_record.domain_override_field` → `input_data.destination_domain`
   - `apply_plan.source_url_field` → `input_data.target_url`

2. **`site_db_actions.go` — `EnsureSiteRecordAction`**
   - Reads `domain_override_field` from step config; if present and path resolves, uses destination domain directly.
   - Falls through to existing `extractDomainFromInput` if absent (legacy path unchanged).
   - Added `isPlausibleDomain` validator to reject config placeholders (`site_record.domain`), paths (`input_data.*`), whitespace, and single-word non-domains. Prevents silent junk-write bug observed during initial testing.

3. **`apply_adoption_plan_action.go`**
   - New locals `sourceURL`/`sourceDomain` resolved from config `source_url_field` → `input_data.target_url` → `input_data.url` → fallback to destination `domain`.
   - Six `domain` → `sourceDomain` substitutions covering `identity.adopted_from`, `structure.adopted_from`, `site_specs.notes`, per-page `contentData.adopted_from`, `needs_composition` spec, `needs_design` spec.
   - Critical fix: `matchCrawlContent(crawlPages, pageURL, sourceDomain)` — was `domain` which silently dropped all page content when source ≠ destination.

### Wrapper orchestrator pattern (new)

User request: make `site-adoption-agent` run in its own pod for clean per-lifetime logs. Investigation showed the system's actual pattern is "every spawned agent is a K8s Job pod" — `site-adoption-agent` was the outlier, running in-chassis because it was invoked directly via the generic entry point.

**Deployed changes:**

1. **`002_create_site_adoption_orchestrator.sql`** — new agent definition `site-adoption-orchestrator`, a minimal wrapper with workflow `spawn_adopter → call_adopter → complete`. Modelled on `med-export-orchestrator` pattern.

2. **`trigger-adopt-separated.sh`** — updated to target `site-adoption-orchestrator` instead of `site-adoption-agent` directly. Payload shape uses `target_url` + `destination_domain`.

3. **`003_fix_site_adoption_orchestrator_input_mapping.sql`** — replaced the initial `{"input_data": "input_data"}` mapping (which double-wrapped the data, producing `input_data.input_data.target_url` in the child) with explicit per-field mapping.

4. **`004_fix_site_adoption_orchestrator_optional_legacy_fields.sql`** — added `?` suffixes to all four mapped fields so `ResolveInputMapping` doesn't fail when the caller passes only the separated shape (missing `url`/`domain`) or only the legacy shape (missing `target_url`/`destination_domain`).

### Documentation updates (deployed)

1. **`001_development_guide.md`** — two new subsections in the orchestration chapter:
   - "Every pod-running agent needs a parent that spawned it" — baseline rule, the outlier mistake, wrapper fix, `med-export-orchestrator` as canonical example, test for when a wrapper is needed. Distinguishes coordination work (conditionals, HITL, spawn/call — fine in-chassis) from substantive work (LLM, crawls, heavy DB — must be in a spawned pod).
   - "Topics: the generic entry point vs per-spawn dedicated topics" — reframed after pushback from "user-triggered" to "externally-initiated". Explicit note that the generic entry point is likely to evolve (user-proxy agents, UI triggers, cross-tree invocations) and shouldn't be treated as a permanent interface. Decision table for new agents. The rule is about pod lifecycle, not caller identity.
   - Fixed the wrapper example from the broken `{"input_data": "input_data"}` shape to per-field explicit mapping with `?` markers for optional fields.

2. **`002_system_architecture.md`** — "Active agents" table restructured from 20-row flat list to grouped-by-area (intake/build, design/assets, adoption, improvement, news feed, tools, vet-med BI, vet discovery, infrastructure), ~60 agents. Source-of-truth pointer to the SQL query for the live list added above the table. Note calling out the wrapper-orchestrator pattern with `med-*` and adoption as concrete examples.

3. **`007_adoption_pipeline_v2.md`** — "The adoption agent" section updated to describe the two-agent arrangement (orchestrator + worker) and the `target_url`/`destination_domain` separation. "Adoption modes" table updated with full input_data shapes. New subsection "Running an adoption: what to expect and what to watch" with the six-step trigger flow, expected timings, three log streams to tail, and success/trouble signals.

4. **`016_debugging_guide_v2.md`** — one-line fix clarifying `target_topic` rules for scheduled tasks (entry point vs fixed adapter topic; `job.*` topics are never valid).

---

## What happened on the first real run (2026-04-21)

Trigger:
- `SOURCE_URL='https://gamedesign.uk'`
- `DEST_DOMAIN='gamesdesign.co.uk'`
- correlation_id `2de43823-a497-409f-b14c-5a36bd412ad8`
- orchestration_id `bc1ef7b9-67fe-474d-8025-1962843b6edd`

Adoption completed. A site was deployed to `gamesdesign.co.uk`. The 16-step workflow ran through: crawl → fingerprint → analyze → classify → derive_content_direction → apply_adoption_plan → design_intent → complete. Source domain `gamedesign.uk` was untouched. The palette colours came through (dark sections; brand colours visible in hero/CTA).

**But the output isn't a copy.** It reads as a generic consultancy/brochure site wearing the source's colours.

### Adoption as designed vs what's wanted

Adoption as currently built produces **a site-planner brief, not a deterministic copy**. It:

1. Extracts a design fingerprint — CSS variables, typography, palette, layout hints.
2. Writes a `design_intent` spec — an LLM-generated *interpretation* of what the design is like, phrased for a designer.
3. Writes an `identity` spec, `content_direction`, `archetype` — LLM-summarised character.
4. Creates `needs_composition` and `needs_design` work items with `mode: "recreate"` — flags for the build pipeline that "this site should look like something that was adopted."
5. Creates pages (if the crawl provided them) with `existing_content.raw_markdown` attached.

Then the **build pipeline** — site-design-planner, webdesign-agent, page-content-writer — takes those specs and builds a site. It doesn't render the crawled HTML directly; it re-plans and re-generates from specs.

The mismatch you're seeing is the gap between **specs + LLM interpretation** and **the actual source site**. The pipeline is working as designed. The design is just too loose.

### What the output tells us about where it's going wrong

Concrete observations from the three HTML files reviewed (index, tools/index, jump-physics-architect):

**Header says `gamesdesign.co.uk` with ◆ icon, teal gradient, `/services.html`, `/about.html`, `/case-studies.html`, `/blog.html`, `/start-a-project.html`, `/games/index.html` nav items.** None of this matches gamedesign.uk. The original had no services/about/case-studies/start-a-project — it was guides, tools, games. Header is a **generic brochure template** slotted in because the build pipeline picked its closest component-library match for "a site" and filled in boilerplate nav.

**Hero says "Game Design Consultancy and Co-Development. Built in the UK."** The site became a consultancy positioning. That's the LLM's `content_direction` interpretation — probably went "gamedesign.uk + professional + UK" → "this is a consultancy site." gamedesign.uk is an educational resource for game designers, not a consultancy.

**"Jump Physics Architect" page has prose describing a tool but no actual tool.** The crawled page had interactive JavaScript calculators; the rebuild has paragraphs describing what the calculator would do. The adoption pipeline pulled the markdown but not the interactive machinery, and the build pipeline wrote explanatory content around the concept.

**Footer lists `/start-a-project.html`, `/case-studies.html`** — pages that never existed on the source. Boilerplate again.

**Tools page (`tools/index.html`) has an empty `<main>`.** Page was created but had no content added, probably because adoption created the page record but the build pipeline didn't have enough spec data to populate it.

### The design question

The user wants adoption for "I own this, keep it essentially the same" — and the current pipeline is designed for "take inspiration, build something new." Those are genuinely different workflows and the current flow conflates them.

The doc `FUTURE_adoption_source_destination_separation.md` called these out as variants:

- **Variant A** — reference-only (design inspiration, fresh content)
- **Variant B** — design + structure (same pages, your content)
- **Variant C** — full clone (copy everything, rename)
- **Variant D** — multi-source analysis

Phase 1 plumbing (`target_url` / `destination_domain`) exists, but the variant selector wasn't wired — everything defaults to the current behaviour, which is roughly between A and C and doesn't commit to either. For the user's case (own site, want a copy), **Variant C** is what's needed — and it doesn't exist yet in a meaningful sense.

### Problems identified, ranked by impact

Each is substantial. "Fix adoption to produce a near-copy" is a week-or-more project, not a one-SQL-patch change.

1. **Nav doesn't match source.** Adoption-created pages exist in the DB, but the nav population (`populate_nav_tables` step) is using an auto-generated brochure nav template instead of building from the adopted pages. This looks like the planner defaulting to its standard brochure archetype.

2. **Identity drifted.** The LLM summarised gamedesign.uk as "consultancy" when it's "educational resource with tools." The `derive_content_direction` and `classify_archetype` prompts are nudging toward a business model that doesn't match. Specs got written with hallucinated positioning.

3. **Tool pages have content but not tools.** The adoption saved markdown of each page but not the `<script>` or interactive `<canvas>` elements. The `tool-recreation-handler` agent exists (per doc 002) but is apparently not being triggered, or is triggered but doesn't yet work for these kinds of tools.

4. **Some pages empty.** Pages created from the crawl's URL list but content not attached — probably because the crawl index lookup in `apply_adoption_plan` (the `matchCrawlContent` call) missed them. Could be the shape of `crawl_result` we noticed earlier — the crawled pages are nested under `crawl_result.response.body.data.pages`, not directly under `crawl_result.pages`. The matchCrawlContent fix in Phase 1 used `sourceDomain` to key the lookup but the pages map itself may still be misshapen.

5. **Design inherited colours but not the structure.** Dark section colours came through; the overall layout (header style, section shapes) didn't. The design-planner filled in its preferred components around the adopted colours.

---

## What's out of scope / deferred

### Resume logic (explicitly deferred)

Investigation found that `orchestration_states.collected_data` already persists per-step output (378KB of state survived the first failed run). `ResumeWorkflowTopic` is defined as a constant but has no subscriber — resume machinery was anticipated but never built. User explicitly said: "I don't think we should worry about the resume logic, a new crawl is fine I now think."

### med-* wrappers have the same double-wrap bug

Query confirmed all four med-* wrappers (`med-price-scrape-orchestrator`, `med-url-discover-orchestrator`, `med-url-map-orchestrator`, `med-export-orchestrator`) use the broken `{"input_data": "input_data"}` mapping. Either their children tolerate the extra nesting, or these wrappers aren't invoked in production. Not blocking adoption. Worth a spot-check when next in that area.

### Stale `crawl_result` shape

The exact shape of `crawl_result.response.body.data.pages` was observed but not mapped fully. If we ever build resume, or if problem #4 above turns out to be a pages-lookup issue, we'll need to verify the shape used in `matchCrawlContent` matches what firecrawl actually returns.

### fetch_primary_css brittleness

User noted: "I think if it fails to get the css that it should be a hard fail, it is important." So no workflow-level skip-on-failure for this step. A firecrawl timeout anywhere downstream still hard-fails. The 378KB of state that survives the failure isn't currently reusable (see resume logic above). Accepted trade-off for now.

---

## Artefacts produced this session

All under `/mnt/user-data/outputs/`:

**SQL patches (apply in order):**
- `001_adoption_source_destination_separation.sql`
- `002_create_site_adoption_orchestrator.sql`
- `003_fix_site_adoption_orchestrator_input_mapping.sql`
- `004_fix_site_adoption_orchestrator_optional_legacy_fields.sql`

**Go code (patched):**
- `site_db_actions.go` + `site_db_actions.patch.diff`
- `apply_adoption_plan_action.go` + `apply_adoption_plan_action.patch.diff`

**Shell:**
- `trigger-adopt-separated.sh`

**Docs (patched):**
- `001_development_guide.md` + `.patch.diff`
- `002_system_architecture.md` + `.patch.diff`
- `007_adoption_pipeline_v2.md` + `.patch.diff`
- `016_debugging_guide_v2.md` + `.patch.diff`

---

## Starting prompt for the next session

Copy the following as the first message of a new conversation:

> Adoption Phase 1 (source/destination separation + wrapper orchestrator) is deployed and working. A real run produced a deployed site at `gamesdesign.co.uk` adopted from `gamedesign.uk`, but the fidelity is poor — it's a generic brochure/consultancy site wearing the source's colours, not a replica. Nav doesn't match, identity drifted to "consultancy", tool pages have prose where interactive JS was, some pages empty, structure doesn't match.
>
> The current adoption pipeline produces **a site-planner brief plus specs, not a deterministic copy**. For the user's case ("I own the source domain, I want a near-replica"), **Variant C (full clone with substitution)** from `FUTURE_adoption_source_destination_separation.md` is what's needed and doesn't exist yet.
>
> Please read the handoff doc `HANDOFF_2026-04-22_adoption_phase1_and_fidelity.md` and the five ranked problems. Start with **problem 2 (identity drifted to "consultancy")** because that's what's driving problems 1 and 5 downstream — the `classify_archetype` / `derive_content_direction` output is the root spec that the rest of the build pipeline plans against.
>
> Key guidelines (repeat of the project's):
> - Please always work small steps at a time and think hard.
> - Please always check the schema of databases before writing sql.
> - Prioritise fixing structural issues above quick fixes.
> - Don't jump to conclusions.
> - Every agent is an orchestrator.
> - Reuse existing functions/architecture before creating new ones.
> - `kubectl -n ai-persona-system get pods` / `kubectl -n kafka get pods` namespaces.
> - Don't use logger.Debug (won't show in logs).
>
> Orchestration IDs from the failed+successful runs if diagnostic lookup needed:
> - Successful: correlation `2de43823-a497-409f-b14c-5a36bd412ad8`, orch `bc1ef7b9-67fe-474d-8025-1962843b6edd`
> - gamesdesign.co.uk site_id is visible via `SELECT id FROM sites WHERE domain='gamesdesign.co.uk'`
