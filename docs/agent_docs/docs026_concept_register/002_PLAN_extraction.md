# Extraction plan — work-unit ledger

26 units covering all 4,111 files under `docs/` (2,530 md; the rest sql/sh/txt/
code/binaries per README treatment rules). Status updated as units complete.
All paths relative to `/home/ant/projects/agentchassis`.

| unit | scope | ~text files | status |
|---|---|---|---|
| U01 | docs024 root: numbered docs 0xx/1xx, P1–P3, *.sql, *.patch | ~50 | done (154 concepts) |
| U02 | docs024 root: FOCUS_*, HANDOFF_*, ANALYSIS_*, ASSESSMENT_*, PLAN_*, ARCHITECTURAL_* | ~52 | done (105 concepts) |
| U03 | docs024/idea_uk_section_data_missing/ | 190 | done (48 concepts) |
| U04 | docs024/idea.uk/ | 172 | done (62 concepts) |
| U05 | docs024/content_quality_and_internal_linking/ | 133 | done (63 concepts) |
| U06 | docs024/finetuning/ (text only; ignore model binaries) | 117 | done (58 concepts) |
| U07 | docs024/content_quality2_silent_overwrite_of_dependent_pages/ | 114 | done (54 concepts) |
| U08 | docs024/travelling_docs/ | 114 | done (52 concepts) |
| U09 | docs024/adoption/ | 84 | done (61 concepts) |
| U10 | docs024/imagery/ | 80 | done (69 concepts) |
| U11 | docs024/traffic_probe/ | 78 | done (39 concepts) |
| U12 | docs024 archives: old/, old_design_and_styling/, debugging_old/, archive_april_26/ | 155 | done (89 concepts) |
| U13 | docs024 small dirs: js_snippets_news_gaswholesalers/, fixloop_eg_dartsonline/, layouts/, tools/, multicluster/, pitch/, stripe/, vetcomparison/, plainjanedomain/, temperature/, dartsonline.com_site_quality/, reasoning/ | ~110 | done (188 concepts) |
| U14 | docs019: RUNBOOK* families | ~150 | done (84 concepts) |
| U15 | docs019: NOTES_running_synthesis* + NOTES_running_fixloop* families | ~140 | done (51 concepts) |
| U16 | docs019: DESIGN_*, PLAN_*, FOCUS_*, HANDOFF_*, PROMPT_*, PATCH_*, MAPPING_*, README_*, NNN_*.sql, TRIGGER_*, 001_*, engines_tree_proposal, directory_tree, tasks/ | ~120 | done (103 concepts) |
| U17 | docs019/_archive/ + go_files/ (md/README files fully; go bodies list-only) | ~80 | done via U17a+U17b (105 concepts total) |
| U18 | agent_docs/sql_for_agents/ (each agent def = concept) | 170 | done (62 concepts) |
| U19 | agent_docs/sql_for_tables/, sql_for_components/, sql_for_tools/, sql_for_hitl/, sql_for_content/, tables_sql/ | ~70 | done (63 concepts) |
| U20 | agent_docs legacy A: docs001_flow_general/, docs001a/, docs002_hitl_parallel/, docs003_firecrawl/, docs004_website_capture_project/ | ~131 | done (79 concepts) |
| U21 | agent_docs legacy B: docs005–docs018 (briefing, workflow_builder, brochure, site_interrogation, multitrack, api_hitl, site_maps, research_agent, data_flow_verification, dogs_medicine, legacy_agent_rules, rerendering) | ~105 | done (58 concepts) |
| U22 | agent_docs recent small: docs019_business/, docs020_llm_training_rag/, docs021_multiclustering/, docs022_domain_authority/, docs023_canine_biology/, docs025_ai_chatbot_idea_uk/ | ~46 | done (64 concepts) |
| U23 | docs root files (vonc running notes/runbooks/plans, debugging guides, manifests) | 168 | done (66 concepts) |
| U24 | docs/_archive/ (delta/superseded focus) | 372 | done via U24a-f (160 concepts total) |
| U25 | docs/leopardessconsulting/ + docs/social001_vonc_tiktok_social/ | ~90 | done (66 concepts) |
| U26 | docs/architecture/, docs/humanintheloop/, docs/basic_usage/, docs/plans/, docs/operations/, docs/api/ | ~46 | done (50 concepts) |

## Consolidation (after extraction)

| step | scope | status |
|---|---|---|
| C1 | Merge extraction concepts into register/<category>.md files | done (19 consolidator agents, 107 files, 1627 concepts) |
| C2 | Assign concept ids, build register/000_concept_index.md | done |
| C3 | Final taxonomy note: settled categories vs seed, ready for stage 2 | done (005_TAXONOMY_final.md) |

## Gap-fill units (session-limit recovery, added after wave-2 failures)

| unit | scope | status |
|---|---|---|
| U17a | docs019/_archive/excellent_discussions + working/main (recovered as text, saved to disk) | done (74 concepts) |
| U17b | docs019/_archive/go_files (contextkit tooling tree) | done (31 concepts) |
| U24a | docs/_archive classic tree + docs024 archived misc (recovered as text, saved to disk) | done (36 concepts) |
| U24b | docs/_archive docs024_key_docs_latest/finetuning (recovered as text, saved to disk) | done (26 concepts) |
| U24c | docs/_archive docs024_key_docs_latest/traffic_probe (recovered as text, saved to disk) | done (36 concepts) |
| U24d | docs/_archive docs024_key_docs_latest/{adoption,content_quality_and_internal_linking} | done (21 concepts) |
| U24e | docs/_archive docs024_key_docs_latest/idea.uk | done (20 concepts) |
| U24f | docs/_archive remaining small dirs (imagery,old,docs004/007/020/025,sql_for_agents,sql_for_tables,nested docs019 archive) | done (21 concepts) |

## Consolidation clusters (C1/C2, launched after full extraction completed)

2185 raw concept blocks split from 32 extraction files, bucketed into 115 category
files, grouped into 19 clusters (188-383KB each) for parallel consolidator agents.
Each writes register/<category>.md files + an index fragment.

| cluster | categories | status |
|---|---|---|
| diagnosis-and-context | diagnosis-loop(41), context-assembly(23), contextkit-toolchain(17), context-pack-tooling(3), context-engineering-principles(6) | done (90 concepts) |
| fixloop-and-governance | fix-loop, autonomy-governance, autonomous-build-operate, autonomy-trust-model, reasoning, investigation-discipline, operating-doctrine, operator-practice | done (50+12+6+2+12+2+2+2=88 concepts) |
| debugging | debugging(74), resilience-self-heal(2), migration-governance(1), sql-change-management(1) | done (78 concepts) |
| development-guide | development-guide(87), workflow-authoring(1), prompt-composition(1), language-i18n(1) | done (90 concepts) |
| contracts-standards-locks | contracts-and-standards, locks, rag-retrieval | done (57+6+1=64 concepts) |
| system-architecture-infra | system-architecture, database-and-infrastructure | done (89+24=113 concepts) |
| multicluster-adapters-deploy | multicluster(15), adapters(16), storage-architecture(9), deployment-github(5), site-snapshots-and-revert(4), vm-backend-sites(14, absorbed backend/persistent-service-deployment) | done (63 concepts) |
| design-composition | design-composition | done (80 concepts) |
| styling-nav-links | styling-render-pipeline, navigation, link-management | done (48+12+22=82 concepts) |
| imagery | imagery, data-charts, component-asset-pipeline | done (64+1+1=66 concepts) |
| documentation-system | documentation-system, site-quality, site-quality-audit-methodology | done (66+2+2=70 concepts) |
| site-plan-and-buildpipeline | site-plan-and-reconciler(42), page-build-pipeline(24), build-pipeline(15, absorbed site-build-pipeline+orchestration-generations), rebuild-cascade(7), work-dispatch(13, absorbed work-item-system+dispatch-pipeline), work-item-integrity(7), action-build-pipeline(1) | done (109 concepts) |
| improvement-loop-tool-lifecycle | improvement-loop(49), tool-lifecycle(30), tool-library(23, absorbed a tool-lifecycle dup), tool-pipeline(7), component-lifecycle(11), games-lifecycle(1) | done (121 concepts) |
| adoption-pipeline-and-cases | adoption-pipeline(36), site-spec-and-classifier(22), dynamic-applications(12), site-case-studies(18) | done (88 concepts) |
| finetuning-model-infra | finetuning-flywheel(41), model-infrastructure(37), rag-knowledge-base(6), llm-quality-testing(5), llm-call-observability(4) | done (93 concepts) |
| content-quality-news-traffic | content-governance, content-quality, news-feed-pipeline, traffic-analytics, topic-intelligence | done (28+17+19+21+3=88 concepts) |
| business-strategy-products | business-strategy(30), idea-product(15), business-intelligence-platform(10), conversion-playbooks(4), portfolio-evolution(3), vertical-knowledge-architecture(4), payments(8), legal-and-compliance(1, absorbed legal-liability), marketing(1), seo(1), affiliate-commerce(1, absorbed affiliate-and-products) | done (78 concepts) |
| hitl-onboarding-agentorg-scheduler | hitl(22), onboarding-config(22), agent-spawning-and-groups(7), agent-memory-and-evolution(3), persona-architecture(2), flows-and-narrative(3), org-framework(1), agent-tree-navigation(1), agent-swarm-simulations(1), agent-definition-registry(1), scheduler-and-tasks(22), batch-processing(4) | done (89 concepts) |
| vertical-social-misc | vet-med-pricing, companies-house-enrichment, canine-biology, business-intel-collection, social-media, vonc, site-chatbot, saas-isolation-architecture, email-infrastructure, admin-dashboard-and-api, research-agents, entity-data, deploy-mechanics-reference, public-api, adopting-and-scraping | done (77 concepts) |
