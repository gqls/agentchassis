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
- Q4 (new): adoption→classifier relationship — direct call or specs+work-items
  only? (query in §B1).
- Q5 (new): TWO classifiers — site-classifier (intake HITL path) vs
  domain-research-classifier (queue path) — same responsibility, two owners;
  consolidation candidate once the spine is fixed. Plus intake-orchestrator's
  unreferenced rerender steps (hygiene).
- Q6 (new): what TRIGGERS pageflow-builder / site-work-orchestrator today —
  the intake spawn_builder path, a scheduled task, or manual? Check
  scheduled_tasks (schema first) + where domain-strategist's chain ends.

## NEXT MOVE (§B2 — close the spine's middle)

§B1's dumps covered the front (submitter→classifier) and the back
(pageflow/site-work). The MIDDLE and the adoption flow remain. Dump:

```sql
SELECT type, version, status,
       jsonb_pretty(default_config->'workflow') AS workflow
FROM agent_definitions
WHERE COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND type IN ('domain-strategist','site-planner','page-build-handler',
               'site-adoption-orchestrator','site-adoption-agent',
               'build-pipeline-trigger')
ORDER BY type, version;

-- and the pump schedule (schema first):
\d scheduled_tasks
SELECT name, agent_type, enabled, schedule, last_completed_at
FROM scheduled_tasks ORDER BY name;   -- adjust columns to \d output
```

Read-out closes: how strategy becomes build items (Q6), the adoption→
classifier relationship (Q4), and where site-planner's plan meets the
work-item pipeline — after which the spine is namable end-to-end and the two
gap agents (exemplar synthesis; infographics) get their attachment points.
