# Taxonomy seed — starting categories for concept classification

The taxonomy is OPEN. These slugs are a starting reference, not a constraint.
Extraction agents tag concepts with the best-fitting slug below or propose
`NEW:<slug>` freely; consolidation settles the final set. Aim: each final
category could back one expert council agent.

## Spine (from docs024 000_documentation_index)

| slug | anchor docs | scope |
|---|---|---|
| development-guide | 001 | agent creation, patterns, work-item lifecycle, loops, data paths |
| system-architecture | 002 | architecture, data flow, topics, idle timeout, cleanup, orchestration |
| contracts-and-standards | 003 | component contracts, slot specs, CSS variables, rendering rules |
| improvement-loop | 004 | discovery → audit → triage → fix → rerender cycle |
| tool-pipeline | 005 | tool suggestion, generation, deployment, cross-linking |
| news-feed-pipeline | 006 | content sources, ingestion, triage, rendering, diversity |
| adoption-pipeline | 007 | site crawling, classification, content recreation, archetype |
| vet-med-pricing | 008 | price scraping, extraction, evidence, JSON export |
| model-infrastructure | 009 | endpoints, health, swap/revert, Ollama, GPU, training data |
| scheduler-and-tasks | 010 | Kafka scheduler, scheduled_tasks, concurrency groups |
| database-and-infrastructure | 011 | connections, pgbouncer, MySQL, client schemas, credentials |
| admin-dashboard-and-api | 012 | React SPA, nginx gateway, endpoints, VPN access |
| content-governance | 013 | locks, inline editing, briefs, regeneration, growth budget |
| site-snapshots-and-revert | 014 | point-in-time capture, SQL functions, revert workflow |
| batch-processing | 015 | Batch API, queue table, submitter/retriever pattern |
| debugging | 016/016b | pod health, work items, orchestration, timeouts, heuristics |
| companies-house-enrichment | 017 | collection, matching cascade, accounts, financials |
| canine-biology | 018 | canine biology feature plans |
| tool-library | 019 | component library, tool registry, matching |
| tool-lifecycle | 020 | tool creation, versioning, health checks, updates |
| site-spec-and-classifier | 021 | classification architecture, spec aspects, archetype |
| dynamic-applications | 022 | interactive app generation |
| llm-quality-testing | 023 | model evaluation, prompt testing, quality metrics |
| link-management | 024 | internal/external link management |
| design-composition | 025/026/027 | palette/layout/typography resolution, site-design-planner, webdesign-agent |
| site-plan-and-reconciler | 029/030 | site_plans domain, reconciler, phase 1 plan |
| locks | 031 | lock semantics, expiry question, coherence |
| storage-architecture | 032 | S3/B2 storage, credentials |
| adapters | 033/035 | adapter guide, thunder adapter, adapter patterns |
| deployment-github | 034 | GitHub actions, git-adapter deploy surface |
| styling-render-pipeline | 036 | CSS/render pipeline reference |
| documentation-system | 037, 000 | doc conventions, indexing, consolidation, travelling docs |

## Provisional additions (evident from directory names — confirm/refine during extraction)

| slug | evidence |
|---|---|
| imagery | docs024/imagery/*, site_plan_imagery pipeline |
| diagnosis-loop | docs019: bundles, contextkit, code retrieval, verdicts |
| fix-loop | fixloop_eg_dartsonline: council, veto, write step, build gate |
| content-quality | content_quality_and_internal_linking, content_quality2_* |
| finetuning-flywheel | docs024/finetuning, FOCUS_finetuning_flywheel |
| navigation | FOCUS_navigation* |
| vonc | docs root: provocations, sparks, lobby grid (vonc site/product) |
| traffic-analytics | docs024/traffic_probe |
| multicluster | docs024/multicluster, docs021_multiclustering |
| payments | docs024/stripe |
| business-strategy | pitch, docs019_business, plainjanedomain, domain strategy |
| onboarding-config | docs019 PLAN_onboarding_*, config derivation |
| hitl | docs002_hitl_parallel, docs011_api_hitl, humanintheloop |
| social-media | social001_vonc_tiktok_social |
| reasoning | docs024/reasoning, reasoning-agent |
| research-agents | vertical exemplar researcher, domain research |
| site-case-studies | idea.uk, dartsonline, gamesdesign, leopardess, robot-hands (only for genuinely site-specific concepts — platform concepts observed on a site belong to their platform category with the site as evidence) |

## Tagging rules

1. Prefer an existing slug when the fit is genuine; propose `NEW:<slug>` when it
   is not — do not force-fit.
2. One concept, one category (the best home); use `relations` for cross-cutting
   links rather than duplicating the concept.
3. Site-specific documents usually evidence platform concepts — extract the
   platform concept, cite the site doc as provenance.
