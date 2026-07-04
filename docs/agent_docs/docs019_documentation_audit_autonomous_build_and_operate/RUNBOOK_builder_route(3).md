# RUNBOOK — builder route (Option B): domain name → multipage website

**What this document is.** This is the working runbook for the "builder
route": making an agent-orchestration platform take a bare domain name and
plan, design, write, build, and deploy a sophisticated multipage website for
it. The platform is a Go chassis (repo `agentchassis`) running on Kubernetes,
where every agent is an orchestrator owning a JSON-defined workflow of steps
that call Go actions; agents communicate over Kafka, spawn sub-agents for
sub-tasks, persist state in Postgres (`orchestration_states`, `sites`,
`site_specs`, `site_work_items`, `pages`, `page_components`…), and deploy
sites via git commits that GitHub Actions push to Backblaze S3. Roughly 147
agent types already exist at varying maturity; the point of this route is NOT
to build a site-builder from scratch but to map what exists, pick one spine
(one owner per hop: intake → research → strategy → plan → design → content →
tools/feeds → build → deploy → improve), retire or consolidate the overlaps,
and attach the genuinely missing pieces (vertical-exemplar research synthesis;
interactive infographics). It proceeds in numbered slices (§B0 inventory, §B1
spine decision, …), records evidence and open questions as it goes, and
changes nothing until a slice's read-out justifies it. A companion route —
the read-only diagnosis loop that investigates build faults by citing code,
data, and runtime evidence — is closed and documented separately in
RUNBOOK_code_retrieval_route.md.

> ## ▶ CURRENT POSITION — 2026-07-04 (§B0: inventory mapping, no code yet)
> Rule honoured: map what EXISTS against what the problem statement wants
> BEFORE creating anything. Sources: the 147-row agent_definitions census
> (2026-07-03), the documentation index (000), the problem statement.
> NOTHING in this file creates or changes agents yet.

## §B0 — Inventory: problem-statement sections × existing agents

Legend: ✔ = active agents exist; ~ = experimental only; ✗ = gap.
(Census caveat: several types have TWO live rows — see Open Questions.)

**(a) Domain intake → strategy/positioning — ✔ (fragmented)**
Active: domain-submitter, chief-strategist (×2 rows), site-strategist.
Experimental: intake-orchestrator, site-classifier, briefing-agent,
build-briefing-agent, domain-analyst, domain-strategist,
domain-research-classifier. Docs: 021 (Site Spec & Classifier).

**(b) Vertical research / best-example capture — ✔ capture, ~ synthesis**
Active: research-agent. Experimental: researcher, content-researcher AND
content_researcher (hyphen/underscore pair), web-search (adapter),
website-capture-firecrawl, website-extract-structured, site-scraper.
Capture of competitor sites exists (firecrawl/extract); the "understand WHY
they succeed" synthesis step has no clear owner. Docs: 007 (Adoption
pipeline — related: crawling+classification of EXISTING sites).

**(c) Site planning (multipage structure) — ✔**
Active: site-planner, landing-page-architect, site-component-architect (×2
rows). Experimental: build-site-planner, site-architect,
content-site-architect, portfolio-architect, multipage-wrapper.

**(d) Design — ✔ (documented pipeline)**
Active: webdesign-agent, site-design-planner, design-discovery-agent,
design-audit-agent, visual-design-auditor, css-patch-agent,
color-variable-fixer(exp), component-template-fixer. Experimental:
brand-designer, visual-designer (adapter). Docs: 026 (Design Composition &
Site Design Planner — palette/layout/typography resolution).

**(e) Content — ✔ core, ~ per-section family**
Active: content-writer, page-content-writer, content-creator (one of two
rows), content-reviewer, content-quality-auditor, content-gap-planner.
Experimental: the SECTION FAMILY — content-creator-{hero, about, contact ×2,
cta, features, testimonials}, content-creator-hero-without-research,
copywriter, simple-content-writer-with-approval. The problem statement's
"separate agents per section" is ALREADY PROTOTYPED here, all experimental.

**(f) Tools — ✔ complete active pipeline**
tool-suggester, tool-generator, tool-deployer, tool-auditor, tool-improver,
tool-recreation-handler — ALL active. Docs: 005 (Tool Pipeline), 019 (Tool
Library), 020 (Tool Lifecycle). Little to build; mostly wire-in.

**(g) Blog / news / features — ✔ planner, ~ feed pipeline**
Active: blog-content-planner. Experimental: content-feed-orchestrator,
feed-ingester, feed-triage, content-feed-trigger. Docs: 006 (News Feed
Pipeline — sources, ingestion, triage, rendering, diversity).

**(h) Interactive graphs / infographics — ✗ GAP**
No agent owns this. Nearest: tool-generator (interactive tools),
site-asset-renderer (active), image-generator (×2 rows, adapter). Candidate
resolution: a tool-pipeline VARIANT rather than a new family — decide after
the spine question.

**(i) Build / render / deploy — ✔ specialists, ~ FRAGMENTED top level**
Top-level builders (the overlap problem): multipage-website-builder (×2 rows,
v1+v2), content-site-builder, landing-page-builder, pageflow-builder
(ACTIVE), website-builder, build-dispatch-loop, build-pipeline-trigger,
site-work-orchestrator — EIGHT types claim "build a site/pages".
Active specialists beneath: page-build-handler, page-rebuild, html-assembler,
site-asset-renderer, spec-updater, internal-linker, internal-link-resolver,
site-component-linker, nav-updater, component-quality-auditor; deployment:
site-deployer AND deployer-agent (TWO active deployers) + site-publisher
(adapter) + asset-deployer(exp). Deploy target per constitution: github →
Actions → backblaze s3.

**(j) Improvement / audit loop — ✔ (documented)**
improvement-loop (exp orchestrator) + design/quality/completeness-discovery,
site-review-agent, maintenance-triage (active). Docs: 004.

**(k) Observability ("closely log and track … messages") — partial, platform-level**
Exists: MESSAGE_TRACE + trace files (/var/log/agent-traces, seen throughout
§6/§7), orchestration_states.processing_history, correlation-id discipline.
Wanted-vs-have review is a LATER slice; no new agents implied yet.

## §B0 findings

1. COVERAGE: every section except (h) has agents; (d)(f)(j) have documented,
   active pipelines. The build problem is NOT missing pieces.
2. THE REAL DEFECT vs the distinct-responsibility rule sits at the TOP
   ORCHESTRATOR TIER: ~8 overlapping "build the site" types, mostly
   experimental, incl. a v1+v2 pair of multipage-website-builder both live.
   The spine — intake → strategy → research → plan → design → content →
   tools/feeds → build → deploy → improve — needs ONE owner per hop;
   candidates exist for every hop.
3. The per-section content family (e) matches the problem statement's shape
   and is already prototyped — extend, don't recreate.
4. Gaps to actually build: (h) infographics owner (likely a tools-pipeline
   variant); (b) success-factor SYNTHESIS step (capture exists).

## §B1 — Spine decision: the nine workflow dumps read (2026-07-04)

**Three generations coexist:**
- GEN-1 (template era, experimental): website-builder;
  landing-page-builder / content-site-builder — strategist → architect →
  content-writer → html-assembler → multipage-wrapper → **site-deployer**.
- GEN-2 (in-memory multipage, experimental): multipage-website-builder v1+v2
  (near-identical; v2 adds input_fields to assemble_site) — chief-strategist →
  loop(content-creator → html-developer → extract_links) →
  assemble_multipage_site → **deployer-agent**. No components/specs/review.
- GEN-3 (component/spec/DB era — the LIVE architecture):
  **pageflow-builder v20 (ACTIVE)**: ensure_site_record → site-planner →
  briefs+plans persisted to site_specs → sync_pages_to_db → nav → conditional
  logo/hero (image-generator) → style collection → render_site_components →
  js snippets → per-page loop (page-content-writer → content-reviewer
  [HITL/auto gate] → assemble_page → git_commit → save_page_sections →
  update_page_status) → webdesign-agent → deployer-agent → complete.
  **site-work-orchestrator (experimental)** = pageflow's queue-native sibling:
  plan → write_build_items → build_items_loop over site_work_items → fix-items
  dispatch loop (DYNAMIC handler_agent per item) → maintenance mode
  (check_mode skips planning). This is the shape the improvement loop needs.

**The queue spine already exists at the front:** domain-submitter (ACTIVE) →
site record + mission/roadmap specs + work item needs_domain_research
(handler domain-research-classifier) → build-dispatch-loop (scheduler-pumped:
claim → spawn dynamic handler → call → complete/fail → touch scheduled_tasks
'build-pipeline-trigger') → domain-research-classifier: web_search + scrape
(the DOMAIN itself) → read_layout_taxonomy (LIVE library tags) → one LLM step
emitting identity + classification + content_direction + design_intent as four
site_specs writes → chains needs_strategy (handler domain-strategist). Its
prompt HARDCODES recommended_builder="pageflow-builder".

**Two front doors, two classifiers (overlap):** the queue path above, and
intake-orchestrator v3 (HITL): fetch available %-builder agents →
site-classifier → hitl_confirm_type → per-builder questionnaire →
briefing-agent → hitl_review_brief → spawn_builder (DYNAMIC type from
confirmed_type.recommended_builder) → call_builder (7200s). Hygiene note: its
spawn_rerender/call_rerender steps appear unreferenced by any next_step.

**Adoption orientation (user: recent attention; believed to lean on the
classifier):** CONFIRMED HALF — the classifier CONSUMES adoption output:
check_adoption_skip_scrape branches on site_specs.specs.site_archetype, and
the prompt's adoption-fidelity block treats adopted identity/archetype/
content_direction/design_intent as ground truth outranking search+scrape.
UNVERIFIED HALF — whether site-adoption-orchestrator CALLS the classifier
directly or connects only via those specs + the work-item chain. Settle with:
```sql
SELECT type, version, status, jsonb_pretty(default_config->'workflow')
FROM agent_definitions
WHERE COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND type IN ('site-adoption-orchestrator','site-adoption-agent');
```

**Where the problem statement's gaps attach (from this read):** the classifier
already asks the LLM for identity.competitors_found — but NOTHING researches
the vertical's best exemplars; the synthesis slot is the classifier's inputs
(alongside search_results/scraped_data) or a step between classifier and
strategist. Infographics remain a tools-pipeline variant (B0). Blog/news
(docs 006) attach post-build.

## §B2 — The pump, the adoption flow, and the relay (dumps read 2026-07-04)

**Q6 ANSWERED — what pumps the pipeline.** The scheduler fires
`build-pipeline-trigger` EVERY 30s (enabled; pre_query gates on sites with
triaged build items, unlocked, attempts < max; concurrency dispatch/8). Its
workflow: seed_build_queue (build_queue → site records + initial items) → find
ONE dispatchable site (none of its items claimed) → spawn+call
build-dispatch-loop → per item: atomic claim → spawn DYNAMIC handler
(current_item.handler_agent) → call → complete/fail → touch scheduled_tasks.
The queue's immune system is all ENABLED: claimed-item-timeout (evidence-based
auto-complete — its SQL documents the gamesdesign false-positive lesson),
feasibility-recheck (blocked→triaged when the handler exists),
stale-orchestration-reaper (30-min special rule for build-dispatch-loop),
stale-work-item-reaper (48h), work-item-archiver, database-cleanup.
**FLAG: improvement-sweep is DISABLED** — the improvement loop is not
currently running. content-feed-refresh IS enabled (6h) — the news pipeline
is pumped.

**Q4 ANSWERED — adoption does NOT call the classifier; the lean is inverted.**
site-adoption-orchestrator is a thin spawn→call wrapper. site-adoption-agent
does the work: firecrawl (30 pages, markdown+rawHtml) → NO-LLM design +
interactive fingerprints (+ external-CSS fetch/enrich) → THREE LLM analyses
(32k site analysis: identity/design/pages/interactive_features; archetype
snapshot incl. improvement-loop constraints; content-direction style guide
from representative pages) → apply_adoption_plan ("write specs, create pages,
create work items") → nav → design_intent generation + spec write. No call to
domain-research-classifier anywhere. So: adoption writes the specs FIRST; the
classifier, when the relay later reaches it, CONSUMES them under its fidelity
rules. Minor unverified detail (Q7): which step writes aspect=site_archetype —
classify_archetype's output is not in apply_plan's declared input_fields;
presumably apply_adoption_plan writes it (code look would confirm).

**The spine as it exists today (end-to-end):**
- DOORS: (1) domain-submitter (blank domain), (2) the adoption pair (existing
  site), (3) intake-orchestrator (HITL; overlapping — Q5).
- RELAY (work items, dispatch-pumped): needs_domain_research →
  domain-research-classifier → needs_strategy → domain-strategist (domain-type
  analysis, revenue-model fits, canonical site_type, page-TYPE recommendations
  → strategy spec) → needs_briefing → build-briefing-agent → **[UNKNOWN
  MIDDLE — §B3]** → needs_content_page items → page-build-handler (ACTIVE):
  page record → adoption existing-content (mode=recreate) → sections from
  site_specs.site_plan → plan_sections readiness triage (the §7-diagnosed
  action) → page-content-writer → validate_page_content (placeholders,
  contamination) → save_page_sections → page-rerender deploy.
- site-planner (ACTIVE) is the component-constrained planning LLM (loads LIVE
  content_components + style_collections; validate_site_plan; index+contact
  guaranteed) — invoked by pageflow-builder / site-work-orchestrator; the
  relay's planning-hop owner is the §B3 unknown (build-site-planner exists in
  the census, experimental).
- MONOLITH path: pageflow-builder v20 does the whole build inline (intake's
  recommended builder per the classifier prompt).

**Gap attachments, now precise:**
- Exemplar synthesis: NEITHER door researches the vertical's other best sites
  (classifier: the domain itself; adoption: the source site). Attachment = ONE
  new relay hop — e.g. needs_vertical_research between classifier and
  strategist — whose handler's workflow can be built from EXISTING actions
  only (web_search, firecrawl_scrape/crawl, format_crawl_for_analysis,
  execute_llm_prompt, write_site_spec aspect=vertical_landscape,
  create_work_item): likely zero new Go; strategist/planner prompts then read
  the new aspect. FIRST BUILD CANDIDATE once the spine decision is made.
- Infographics: adoption already captures interactive_features; the tools
  pipeline is active; page_type "tool" exists in both vocabularies — a
  tool-generator variant, later.

**New hygiene observation (Q8):** vocabulary drift between hops — the
classifier's site_type set (brochure|landing|portfolio|content|ecommerce|
tools|interactive-platform|social) vs the strategist's canonical set
(brochure|authority-portal|local-directory|review-site|content-hub|
landing-page|portfolio). Two taxonomies for the same concept, one spec chain.

## Open questions (settle before touching anything)

- Q1 DUPLICATE ROWS: multipage-website-builder, chief-strategist,
  content-creator, site-component-architect, image-generator each have TWO
  live rows (v1/v2 display names suggest version pairs — census didn't select
  `version`). Which does spawn/adopt resolve? Needs one query + one code look
  (spawn resolution order) — NOT assumed a defect yet.
- Q2 content-researcher vs content_researcher: hyphen/underscore near-twins —
  which is referenced by live workflows?
- Q3 ANSWERED (§B1): generational — site-deployer serves GEN-1 chains;
  deployer-agent serves GEN-2/GEN-3 (pageflow, multipage, site-work).
- Q4 ANSWERED (§B2): no direct call; adoption writes specs first, the
  classifier consumes them later under fidelity rules (inverted lean).
- Q5 (new): TWO classifiers — site-classifier (intake HITL path) vs
  domain-research-classifier (queue path) — same responsibility, two owners;
  consolidation candidate once the spine is fixed. Plus intake-orchestrator's
  unreferenced rerender steps (hygiene).
- Q6 ANSWERED (§B2): the pump is scheduler → build-pipeline-trigger (30s,
  pre-query gated) → build-dispatch-loop → dynamic handlers. pageflow-builder
  is reached via intake's spawn_builder (and manual); the relay's own path to
  planning/build items is the §B3 unknown. improvement-sweep DISABLED.
- Q7 (new, minor): confirm which adoption step writes aspect=site_archetype
  (apply_adoption_plan presumed; code look at apply_adoption_plan_action).
- Q8 (new, hygiene): site_type taxonomy drift — classifier vs strategist use
  different canonical sets for the same spec chain.

## NEXT MOVE (§B3 — the relay middle, one dump; then the spine decision)

The only unread hop is briefing → planning → build items. Dump:

```sql
SELECT type, version, status,
       jsonb_pretty(default_config->'workflow') AS workflow
FROM agent_definitions
WHERE COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND type IN ('build-briefing-agent','build-site-planner');
```

PRE-REGISTERED READ CRITERIA (fixed before the dump is read, §7E-style):
1. Does build-briefing-agent chain onward via create_work_item (relay-native)
   or call/spawn a builder directly?
2. Who writes the site_plan spec in the relay — build-site-planner,
   site-planner, or nobody?
3. Who writes the needs_content_page items (write_build_items? which agent)?
4. Version/status of both agents (active vs experimental)?
5. Does build-site-planner REUSE site-planner's component-constrained
   planning (live content_components + style_collections + validate_site_plan)
   or duplicate it — an overlap in the sense of the distinct-responsibility
   rule?

DECISION RULE (pre-stated): if the relay reaches page-build-handler natively,
the spine = the work-item relay, with pageflow-builder positioned as intake's
initial-build convenience (decomposition optional, later). If the relay
dead-ends at briefing, the missing hop is the FIRST structural fix — spine
before features; the exemplar-research hop (§B4) waits behind it either way.
§B4 then designs that hop against the settled spine — reusing existing
actions, following guidelines 001 for the agent definition.

STATUS 2026-07-04 (later): the re-sent files contained the §B2 material
(scheduled_tasks full dump + adoption pair + six-agent dump — file-backed
confirmation of §B2, incl. domain-strategist chaining needs_briefing →
build-briefing-agent) and unrelated other-thread material (lobby-grid
component template checks). THE §B3 DUMP IS STILL OUTSTANDING — second
attempt (buildbriefingagentandbuildsiteplanner.txt) arrived 0 bytes, same
save fault as the first sql_result.txt; re-save fixed it then.
TRIANGULATION while blocked (from the §B2 material on disk): needs_content_page
is live (claimed-item-timeout's evidence SQL handles it: page_id +
page_components written by the claim's handler), needs_briefing →
build-briefing-agent confirmed (line 838), but NO dumped workflow references
build-site-planner / needs_planning / creates needs_content_page — the relay's
item CREATORS are Go actions (apply_adoption_plan "create work items from
analysis"; write_build_items) whose item-type vocabulary lives in code. So
criterion C3 may need ONE code look (those two actions) even after the dump. Option A meanwhile CLOSED: image v1.0.1092 deployed, rename
migration applied + verified (four agents on result_from).

Hygiene note from the full scheduled_tasks dump: feasibility-recheck's
pre_query is an UPDATE-returning CTE — the scheduler's GATE query itself
performs the blocked→triaged promotion as a side effect (no agent involved).
Works, but a pre_query with writes is an architectural wrinkle worth
remembering when reasoning about who mutates work items.
