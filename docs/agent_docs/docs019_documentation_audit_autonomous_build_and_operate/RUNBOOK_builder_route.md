# RUNBOOK — builder route (Option B): domain name → multipage website

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

## Open questions (settle before touching anything)

- Q1 DUPLICATE ROWS: multipage-website-builder, chief-strategist,
  content-creator, site-component-architect, image-generator each have TWO
  live rows (v1/v2 display names suggest version pairs — census didn't select
  `version`). Which does spawn/adopt resolve? Needs one query + one code look
  (spawn resolution order) — NOT assumed a defect yet.
- Q2 content-researcher vs content_researcher: hyphen/underscore near-twins —
  which is referenced by live workflows?
- Q3 TWO ACTIVE DEPLOYERS (site-deployer, deployer-agent): who calls which?

## NEXT MOVE (§B1 — the spine decision, one query)

Dump the top-level builders' workflows + versions to decide extend-vs-
consolidate (schema already known):

```sql
SELECT type, version, status,
       jsonb_pretty(default_config->'workflow') AS workflow
FROM agent_definitions
WHERE COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND type IN ('multipage-website-builder','content-site-builder',
               'landing-page-builder','pageflow-builder','website-builder',
               'site-work-orchestrator','intake-orchestrator',
               'build-dispatch-loop')
ORDER BY type, version;
```

Read-out decides: which orchestrator becomes the spine's top (pageflow-builder
is the only ACTIVE candidate; multipage-website-builder v2 the named-intent
candidate), what each duplicate/overlap actually does, and where the seams for
(b) synthesis and (h) infographics attach.
