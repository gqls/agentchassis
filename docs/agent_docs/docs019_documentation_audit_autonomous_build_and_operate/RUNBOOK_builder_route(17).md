# RUNBOOK — builder route (Option B): domain name → multipage website

> **THREAD HANDED OFF 2026-07-06 → HANDOFF_builder_thread.md** (fresh-context
> restart; the handoff carries the queue — Q5 classifier consolidation with
> the user's brief, Q8, CI indexing — the file manifest, and the gotchas.
> At handoff: makefile pins IMAGE_TAG v1.0.1096 (`latest` commented out);
> redeploy-agents = rollout restarts ONLY — the row-bump site is elsewhere in
> `release`, one grep left for the new thread.)

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
CORRECTION 2026-07-04 (user): the status tag is REDUNDANT — "experimental"
agents are still live. Liveness is determined by the pump + handler references
(verified in §B2/§B3), not by this column. The ✔/~ legend below overstates the
distinction; §B3's spine decision never used status.

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

## §B3 — SPINE DECIDED: the work-item relay (2026-07-04)

Dump read against the pre-registered criteria (files:
buildbriefingagentandbuildsiteplanner.txt + apply_adoption_plan_action.go +
load_work_item_actions.go):
- C1 relay-native ✔ — build-briefing-agent: read_specs →
  fetch_agent_questionnaire → LLM briefing_answers → write_site_spec
  aspect=briefing → create_work_item needs_site_plan (handler
  build-site-planner). No direct builder call.
- C2 ✔ — build-site-planner writes the plan in the relay: read_specs →
  ensure_site → load_existing_pages → load_components → load_styles →
  plan_site (LLM) → validate_plan → write_site_plan → sync_pages →
  populate_nav → reconcile_site_plan → emit_design → emit_imagery → complete.
- C3 ✔ (in Go, as triangulated) — load_work_item_actions.go's routing table:
  content/index/landing/blog-index/blog-post →
  {handler: page-build-handler, itemType: needs_content_page}, itemKey
  needs_page:<name> (dedupe); imagery items (needs_logo, severity high) from
  plan flags. COMMENTED-OUT FUTURE ROUTES present: entity-directory,
  entity-page, and **"tool" → tool-build-handler (needs_tool_page)**.
- C4 — both v1, experimental.
- C5 overlap CONFIRMED — build-site-planner duplicates site-planner's
  component-constrained planning pattern (live components + styles +
  validate_site_plan) and adds the relay-side sync/reconcile/emit.
  Consolidation candidate (pageflow could call build-site-planner, or the
  planning prompt gets shared); recorded, not acted on.

**DECISION (pre-stated rule fires): the relay reaches page-build-handler
natively ⇒ THE SPINE = THE WORK-ITEM RELAY.** pageflow-builder = intake's
initial-build convenience; decomposition optional, later.

**The spine, end-to-end, no unknowns:**
domain-submitter (blank domain) OR site-adoption-agent (existing site: specs
incl. site_archetype [Q7 ANSWERED — apply_adoption_plan reads
site_archetype_analysis from collected data regardless of declared
input_fields and writes it] + pages + needs_tool_recreation/needs_content_page
items + HANDOFF create needs_domain_research → the relay) → pump
(build-pipeline-trigger 30s → build-dispatch-loop) →
domain-research-classifier (adoption-fidelity-aware) → needs_strategy →
domain-strategist → needs_briefing → build-briefing-agent → needs_site_plan →
build-site-planner (⇒ needs_content_page items + design + imagery items) →
page-build-handler per page (plan_sections → write → validate →
save_page_sections → rerender deploy) → immune system; improvement-sweep
currently DISABLED.

**Gap (h) attachment found IN CODE:** the commented "tool" route means
infographics/tools = implement tool-build-handler + uncomment the route,
wiring the ACTIVE tool pipeline (suggester/generator/deployer) into the relay.

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
- Q7 ANSWERED (§B3): apply_adoption_plan reads site_archetype_analysis from
  CollectedData (declared input_fields incomplete but harmless) and writes the
  spec; adoption then hands off needs_domain_research into the relay.
- Q8 (new, hygiene): site_type taxonomy drift — classifier vs strategist use
  different canonical sets for the same spec chain.

## NEXT MOVE (§B4 — the exemplar-research hop, first build of the route)

Design against the settled spine; REUSE-ONLY ambition (zero new Go expected):

- NEW AGENT `vertical-exemplar-researcher` (v1, experimental; category per
  guidelines 001): workflow from EXISTING actions —
  read_site_spec (identity + classification, incl. identity.competitors_found)
  → web_search (vertical's best sites / competitors)
  → scrape_web or firecrawl (top N exemplars)
  → format_crawl_for_analysis
  → execute_llm_prompt (success-factor synthesis: positioning, content
    patterns, design patterns, tools observed, trust signals — REASONS not
    copies)
  → write_site_spec aspect=vertical_landscape
  → create_work_item needs_strategy (handler domain-strategist)
  → complete (result contract per guidelines 003 preferred names —
    result_from).
- CHAIN CHANGE (one migration, snapshot first): domain-research-classifier's
  create_next_item: needs_strategy → needs_vertical_research (handler
  vertical-exemplar-researcher). The researcher chains needs_strategy onward,
  so the relay lengthens by one hop.
- CONSUMER CHANGE: domain-strategist's prompt reads the new
  vertical_landscape aspect (one prompt edit in its definition; snapshot).
- Adoption path: unchanged — adoption enters the relay at the classifier, so
  adopted sites get exemplar research too (fidelity rules still outrank).
- VERIFY plan: one blank-domain run; watch the relay: classifier item →
  research item claimed → spec aspect=vertical_landscape present → strategy
  spec references the landscape.

PLAIN-LANGUAGE EXPLAINER (recorded 2026-07-04 at user request):
- A "hop": work moves like a relay race — the baton is a site_work_items row
  naming a handler_agent; the 30s scheduler picks up unclaimed batons and runs
  the named agent; the agent does its one job, writes findings to site_specs
  (the site's shared notebook), creates the NEXT baton, stops. One hop = one
  baton + one agent + one spec written + one new baton.
- "Vertical-exemplar research synthesis": today the system researches only
  THE customer's domain. This hop finds the NICHE's best existing sites (the
  vertical's exemplars — e.g. for a locksmith: the best locksmith sites),
  scrapes a few, and one LLM step distils WHY they work (positioning, content
  patterns, design, tools, trust signals) into a written conclusion — reasons,
  not copies — saved as spec aspect=vertical_landscape for the strategist and
  planner to read. Look at the best of the competition, understand why it
  works, write the lessons where the planner can see them.
- Why each choice: NEW agent (distinct responsibilities: classifier =
  understand this domain; researcher = understand the niche; separate agents
  keep logs/maintenance separate — same reason it is a hop, not a
  subworkflow). AFTER the classifier (needs the classification to know what to
  search). BEFORE the strategist (strategy is the first consumer).
  REUSE-ONLY (every needed action exists ⇒ the agent is one DB row, no Go, no
  image). SPEC not message (specs are cross-hop shared memory; the 1.27MB
  message lesson). STRATEGIST PROMPT EDIT (research nobody reads is wasted).

DESIGN CALLS LOCKED (2026-07-04/05): (1) budget = 3 exemplars, shallow —
firecrawl limit 6 / markdown only / only_main_content / depth 1 (vs
adoption's 30/rawHtml/depth 4: one site deep there, three sites light here);
(2) adopted sites RUN the hop (fidelity keeps outranking); (3) item
vocabulary from live rows: item_domain "build", priority 7 (below strategy's
8 in the ascending 8/10/15 ladder; order within a site is enforced by
creation anyway), full create_work_item key set mirrored from
build-briefing-agent (source/summary/severity/item_key_prefix).

DOUBLE-CHECK (user-requested, 2026-07-05): does competitor research already
exist? IN THE DUMPED FLOW — NO, decisively: the only research-capable actions
are adoption's four firecrawl_* steps, all pointed at the site BEING ADOPTED;
the classifier researches THE domain and captures identity.competitors_found
(names only, never researched); the strategist's §4 "Competitive Positioning"
REASONS about "who currently ranks" with read_site_spec as its only input —
reasoning present, research absent, which is exactly the seam §B4 fills.
OUTSIDE THE FLOW — two pre-apply queries (in the apply order below) sweep all
147 definitions for competitor/exemplar mentions and read the research-named
agents' descriptions (research-agent, researcher, content-researcher pair);
expectation: content-researchers do per-SECTION topical research (different
responsibility, coexists by design). If a hit IS vertical-competitor
research, EXTEND it instead of seeding.

CHANGE-SET WRITTEN (reuse-only; zero Go; all key shapes mirrored from on-disk
siblings):
- NNN_seed_vertical_exemplar_researcher.sql — 12-step workflow: read_specs →
  select_exemplars (LLM, FLAT keys exemplar_1/2/3 so every path is a proven
  dotted lookup; forbids the customer's own domain) → 3×(firecrawl_crawl
  shallow + format_crawl_for_analysis 400 chars/page) → synthesise (LLM:
  per-exemplar success factors/tools/trust; cross-exemplar patterns; lessons
  adopt/adapt/avoid; differentiation_opportunity — REASONS NOT COPIES) →
  write_site_spec aspect=vertical_landscape → create_work_item needs_strategy
  @8 → complete (result_from, valid post-v1.0.1092). Donor for category/
  status: domain-research-classifier. Known v1 limitation accepted: one bad
  exemplar URL fails the item → immune-system retry.
- NNN_reroute_classifier_to_vertical_research.sql — SELF-GUARDED: updates
  create_next_item config only where it currently creates needs_strategy
  (the step name all three on-disk relay siblings use); UPDATE 0 = assumption
  wrong, nothing changed, inspection SELECT embedded. needs_vertical_research
  @7 → vertical-exemplar-researcher. Snapshot = byte-exact revert path.
- NNN_strategist_vertical_landscape_nudge.sql — OPTIONAL: the strategist
  injects {{.site_specs}} WHOLESALE, so the landscape reaches it with zero
  changes; this idempotent append just tells §4/§6 to weigh it.

APPLIED 2026-07-05 — ALL FOUR STEPS, verify run IN FLIGHT:
(0) Sweep CLEAN: competitor mentions = contextual prompt language only
    (classifier, strategist, content-quality-auditor, webdesign-agent); the
    four research-named agents are TOPIC/CONTENT researchers (descriptions
    read) — seed-not-extend confirmed correct.
(1) Seed applied: INSERT 1; all 12 steps verified present.
(2) Reroute applied: snapshot e6ca8cca; UPDATE 1; chain_config verified
    (needs_vertical_research @7 → vertical-exemplar-researcher).
(3) Nudge applied: snapshot ffa6b2da; UPDATE 1; nudged=t.
(4) VERIFY RUN: dartsonline.com via 082_submit_domain_unified.sh —
    correlation 2fc9eef6, orch c5b6d676, site 5fe8785b. Submitter COMPLETED;
    classifier COMPLETED (design_intent spec present with full palette/
    typography reference_values); ITEM CHAIN LIVE:
    needs_domain_research=complete → **needs_vertical_research=CLAIMED by
    vertical-exemplar-researcher** — the new hop is executing. This run
    doubles as Option A's live verification (expect NO deprecation warns;
    flattened completion).

INCIDENT 2026-07-06 — first claim STALLED; ROOT CAUSE CORRECTED after the
evidence pack. Pod spec proved: CMD EMPTY (normal — page-build-handler and
diagnose-agent also have empty command; nil-CMD is how ALL spawned agents
run) and the full env WAS injected (AGENT_TYPE, per-instance KAFKA_TOPICS
job.190d8369-…, KAFKA_CONSUMER_GROUP vertical-exemplar-researcher-group-…).
The column comparison showed res/health/env/topics NOT NULL on our row (DB
defaults filled them) and is_active=t. THE ONE REAL DIFFERENCE:
image_tag='latest' (column default) vs the siblings' pinned v1.0.1094 — and
the registry's `latest` is an ANCIENT chassis build from the
system.agent.generic.process era (guidelines 001 line 472: generic pods use
generic.REQUESTS; generic.process is not in the architecture — user
confirmed "we don't use it"). The old binary boots generic regardless of
env. Same staleness class as §7A's HEAD pin. My earlier command→entrypoint
narrative was WRONG on the differentiator and is superseded by this.
FIX APPLIED 2026-07-06: NNN_fix_researcher_spawn_columns.sql (snapshot
139362d5) — the donor copy set image_tag=v1.0.1094 (the operative change;
command stayed empty, correctly). Seed re-run duplicate-key was EXPECTED
(agent already existed); seed now ON CONFLICT DO NOTHING.
GUIDELINE SCORECARD (user-requested, vs 001's two checklists): Step-0
pre-flight ✔ (the sweep); single responsibility ✔; ActionInputSpec steps
N/A (zero new Go); workflow shape ✔ (3× crawl unroll noted — keeps loops
out of workflow JSON); standalone-mode parity with build-briefing-agent;
handler checklist ✔ incl. pipeline="build" (the item_domain→pipeline
naming trap, guideline 1083); end-with-deploy rule N/A mid-relay.
NEW PARKED TRAP (systemic): agent_definitions.image_tag DEFAULTS to
'latest' = the stale build — every future seeded agent inherits it. Options:
repoint/retire `latest` in the registry; ALTER the column default to a
pinned tag; add "set image_tag explicitly" to 001's New Agent checklist
(user's doc edit). Until one lands, seeds must copy image columns from a
live donor (the amended seed does).
CLEANUP + RE-RUN (pending): delete the stuck generic job
(kubectl -n ai-persona-system get jobs | grep vertical-exemplar → delete
job <name>); reset the item (UPDATE site_work_items SET status='triaged'
WHERE item_type='needs_vertical_research' AND status='claimed' AND
site_id='5fe8785b-223d-41a3-88ee-c07187622381';) or let
claimed-item-timeout's evidence check do it; pump re-dispatches ~30s; the
new pod log should open Agent starting type=vertical-exemplar-researcher on
the job.… topic with the per-instance group.

RE-RUN VERIFIED 2026-07-06 — §B4 FUNCTIONALLY GREEN. After the image-tag
fix + item reset, the relay ran unaided:
  needs_domain_research=complete → needs_vertical_research=COMPLETE
  (vertical-exemplar-researcher) → needs_strategy=complete (summary text
  "Strategy needed after vertical exemplar research" = the RESEARCHER's
  create_next_item wording — authorship of the re-wired chain proven) →
  needs_briefing=complete → needs_site_plan=CLAIMED (build-site-planner;
  normal downstream territory per §B3).
Spec ledger: eight current aspects incl. **vertical_landscape from
vertical-exemplar-researcher**, written BEFORE strategy — positioned for the
strategist's wholesale {{.site_specs}} injection. No blocked items (valid
zero — query pre-corrected for the ambiguous status column; note site_specs
and site_work_items joins to sites need wi./ss. prefixes on status/created_at).
§B4 CLOSED — QUALITY VERIFIED 2026-07-06 (vertical-exemplar.txt, 20,548B +
the strategy paste):
✔ Selection: dartscorner.com / darts.com / missiondarts.com — the vertical's
  real leaders; live; distinct; not the customer's domain (all prompt rules
  honoured).
✔ Synthesis = REASONS NOT COPIES: causal success factors ("spec completeness
  as trust — tungsten % in every title lets players scan without clicking"),
  actionable adopt-lessons (spec-in-titles; free-shipping threshold in the
  header), positioning avoid-lessons; confidence 0.82.
✔ CONSUMPTION PROVEN: the strategy's gap_opportunity QUOTES the hop ("The
  vertical landscape research identifies a clear structural gap"), names the
  three competitors with specifics, and builds the moat on the synthesis.
✔ TRANSMISSION THREE HOPS DEEP: the landscape's differentiator (setup-builder
  as first-class purchase journey) appears in the PLAN as the
  tool-setup-builder page.

MAP CORRECTIONS from the live 37-row chain (live truth > my §B3 notes):
- Page items are item_type **needs_page** (not needs_content_page as the
  guideline's historical text suggested) → page-build-handler; 16 planned
  pages, 2 complete, 1 claimed, rest triaged — churning normally.
- NEW HOPS OBSERVED, not in the §B2 map: needs_composition →
  site-design-planner (complete; doc 026's composition pipeline is wired into
  the relay — emitter not yet identified, not chased); needs_design →
  webdesign-agent (triaged, from emit_design_items); needs_imagery →
  **image-build-handler** (13 items — a handler the B0 census read missed);
  needs_rerender → rerender-pages (the assemble+deploy tail — the guideline's
  end-with-deployment rule is satisfied by the plan).
- KNOWN CLASS RECURRING: unresolved_cta | needs_human_review on the about
  page ("no real-page destination") — the same deferred-CTA class the §7
  diagnosis loop confirmed on gamesdesign. PARKED: a structural
  CTA-resolution candidate (resolve CTAs from the site plan/nav at
  plan_sections time).

§B5 CANDIDATE, NOW CONCRETE — the commented tool route is load-bearing:
tool-setup-builder went out as needs_page → page-build-handler because
reconcile's "tool" → tool-build-handler (needs_tool_page) route is COMMENTED
OUT. The vertical's headline differentiator will therefore ship as a PROSE
page unless the route lands. §B5 = implement tool-build-handler (wire the
ACTIVE tool pipeline: tool-suggester/generator/deployer) + uncomment the
route. Sequencing choice (user's): let this build finish as prose first
(rerender ships something; upgrade the page later) vs land the route before
the setup-builder page builds.

## §B5 — ON HOLD: coordination with the parallel tools chat (2026-07-06)

A second session is actively working the tool pipeline (travelling-docs:
per-tool PLAN/NOTES keyed by function; doc_plans/doc_notes; the tool-doc
header system whose sentinels are visible in tool-generator/tool-improver
prompts; persist_diagnosis_note after diagnose_emit; KB tool_docs;
deploy_tool_to_site source_* stamps). Evidence of active change: ALL FIVE
tool agent rows updated 2026-07-06 13:22 at image v1.0.1096.

BOUNDARY (adopted): the other chat owns everything INSIDE the tool pipeline +
tool docs + its drafted diagnose change (0NN_diagnose_load_runtime_error_step
.sql + a planned no-anchor softening of diagnose_load_runtime — noted here so
THIS chat does not touch that action without checking their draft). This chat
owns the RELAY (classifier → researcher → strategist → briefing → planner →
reconcile routing) and the builder map. The §B5 interface — how a PLANNED
tool page reaches the pipeline — is a JOINT decision, not taken unilaterally.

DESIGN FORK recorded for that joint decision (from today's definitions):
(i) thin tool-build-handler: reconcile's "tool" route → handler that drives
    tool-generator-style generation INTO the already-synced page (conflict to
    resolve: create_tool_component CREATES a page; the planner's page already
    exists — two page-creation paths);
(ii) tool-generator gains an existing-page mode (their territory);
(iii) most reuse-shaped: no handler — a relay hop runs tool-suggester after
    site_plan; suggester already reads ALL specs (so it sees the strategy +
    vertical_landscape that motivated tool-setup-builder) and its pipeline
    owns page creation end-to-end (deployer/generator create page + nav +
    component). The planner would then stop emitting tool page_types, or
    reconcile skips them.
Until decided: tool page_types fall to the prose path (dartsonline's
tool-setup-builder ships as prose and upgrades later — accepted sequencing).

USEFUL FROM THEIR NOTES for this map: (a) orchestrator can show COMPLETED
while the child FAILED — header status complete, body.status failed;
consumers of child results must check body.status (behaviour, recorded);
(b) agent-type (hyphen) label independently confirmed; (c) identical
updated_at across all five tool rows suggests a BULK image bump on deploy —
if such a mechanism exists it shrinks the pinned-tag drift worry from the
latest-tag incident (verify sometime: did vertical-exemplar-researcher get
bumped 1094→1096?).

## MILESTONE 2026-07-06 — first end-to-end domain→deployed site through the relay
dartsonline.com rendered and committed (about.html, contact.html, index.html +
JS snippets bundle; "Rerender:" commits) — bare domain in, deployed HTML out,
via submitter → classifier → §B4 researcher → strategist → briefing → planner
→ page-build-handler → rerender → git → Actions → Backblaze. The macro
objective functions; QUALITY is the next axis.

## §B6 — First-build quality read — HANDED OFF 2026-07-06
→ **RUNBOOK_site_quality.md** now owns this as a full programme (the four
rendered pages were measured: zero nav/img/svg/script everywhere; 14–50 CSS
vars referenced against a likely-undelivered stylesheet; a 3-href retail
page). This thread keeps the relay/spine; quality legs 1–7 move to the new
runbook; LEG 5 scope changes that alter relay hops come back through here.
The original §B6 query set below transplants verbatim into the new thread.

## §B6 (original proposal, kept for the query set)

SCOPING vs the parallel chat: content, design, imagery, and the unresolved-CTA
class are THIS thread's legs; tool interactivity (setup-builder as prose) and
tool-auditor findings wait on the joint seam / other chat. improvement-sweep
remains DISABLED platform-wide (§B2) — enabling it is a later lever, after the
baseline read.

READ FIRST (adjust to \d if any column differs; site 5fe8785b-…):
```sql
-- 1) what never landed: undone/failed build items
SELECT item_type, wi.status, handler_agent, attempts, LEFT(summary,60)
FROM site_work_items wi
WHERE wi.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND pipeline='build'
  AND wi.status <> 'complete'
ORDER BY wi.status, priority;

-- 2) did the design leg land?
SELECT status, build_status, style_collection_id
FROM sites WHERE id='5fe8785b-223d-41a3-88ee-c07187622381';
SELECT aspect, source_agent, ss.created_at
FROM site_specs ss
WHERE ss.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect IN ('composition','webdesign','design_intent') AND is_current=true
ORDER BY ss.created_at;

-- 3) content shape per page: components, rendered bytes, thin components
SELECT p.name, p.build_status, count(pc.*) AS comps,
       sum(length(COALESCE(pc.rendered_html,''))) AS html_bytes,
       count(*) FILTER (WHERE length(COALESCE(pc.rendered_html,'')) < 80) AS thin
FROM pages p LEFT JOIN page_components pc ON pc.page_id = p.id
WHERE p.site_id='5fe8785b-223d-41a3-88ee-c07187622381'
GROUP BY p.name, p.build_status ORDER BY p.name;
```
DECISION RULE (pre-stated): items undone (design/imagery triaged or failed) ⇒
work the DISPATCH/handler question first — pages can't look right without a
stylesheet or images, whatever the copy says. Items complete but pages thin ⇒
work the CONTENT leg (page-content-writer prompt / section family /
validate). Both ⇒ dispatch first, content second. If the direct read is
ambiguous on a specific page, the §7 diagnosis loop is available as the
deeper instrument (symptom-driven, cites code+data+runtime).

## THIS THREAD'S QUEUE — 2026-07-06 (non-overlapping; other threads' territory excluded)

Excluded by boundary: quality legs 1–7 incl. CTA resolution (→
RUNBOOK_site_quality.md); tool internals/docs/auditor + diagnose_load_runtime
(→ tools chat); the planned-tool-page seam (joint, on hold).

1. [small, FIRST] Bulk-bump verification — settles the image_tag trap's
   severity + a §B4 loose end:
   ```sql
   SELECT type, image_tag, updated_at FROM agent_definitions
   WHERE type IN ('vertical-exemplar-researcher','page-build-handler',
                  'domain-research-classifier')
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
   ```
   Researcher at v1.0.1096 ⇒ deploys bulk-bump pinned tags ⇒ the 'latest'
   default only bites between seed and first deploy ⇒ fix = one guideline
   line (001 checklist: set image columns from a live donor — the amended
   seed already does). Researcher still at 1094 ⇒ no bump mechanism ⇒ the
   ALTER-default / repoint-latest decision becomes real.
2. [MAIN] Q5 front-door consolidation — two classifiers, one responsibility:
   READ FIRST: site-classifier's workflow +
   evidence of intake-door usage:
   ```sql
   SELECT type, version, status, jsonb_pretty(default_config->'workflow')
   FROM agent_definitions WHERE type='site-classifier'
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
   -- usage evidence (\d orchestration_states first; adjust if no name col):
   SELECT count(*), max(created_at) FROM orchestration_states
   WHERE orchestration_name ILIKE '%intake%';
   ```
   Then decide: deprecate the intake door (queue door + a HITL gate item?)
   vs align site-classifier onto domain-research-classifier's contract. Also
   sweeps up intake-orchestrator's orphaned rerender steps.
3. [queued] Q8 taxonomy alignment — one canonical site_type set across
   classifier + strategist prompts (decision + two snapshot migrations).
4. [queued, self-contained] CI-triggered indexing: GitHub Actions step firing
   the index-orchestrator envelope with ${GITHUB_SHA} on push (runners
   in-cluster) — prevents the §7A staleness class recurring in the diagnosis
   corpus.
5. [background] Q2 content-researcher vs content_researcher twins (which do
   live workflows reference); stale agent_definitions column cosmetics.

WATCH (paste when the item flips):
```sql
SELECT item_type, wi.status, handler_agent, LEFT(summary,60)
FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE s.domain='dartsonline.com' AND pipeline='build' ORDER BY priority;

SELECT jsonb_pretty(data) FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE s.domain='dartsonline.com' AND aspect='vertical_landscape' AND is_current=true;

-- after needs_strategy completes: did the strategy USE the landscape?
SELECT jsonb_pretty(data->'competitive_position') FROM site_specs ss
JOIN sites s ON s.id=ss.site_id
WHERE s.domain='dartsonline.com' AND aspect='strategy' AND is_current=true;
```
FAILURE PATH (if the item goes failed — likely a bad exemplar URL, the
accepted v1 limitation): pod logs before cleanup reaps —
`kubectl -n ai-persona-system logs -l agent_type=vertical-exemplar-researcher --tail=3000`
plus orchestration_states error by the researcher's orchestration (find via
agent_error_log by site_id 5fe8785b or the item's claim).
