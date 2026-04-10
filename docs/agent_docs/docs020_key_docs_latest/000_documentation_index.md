# Documentation Index

Master index for the AI Agent Orchestration System documentation. Consolidated from 57 source files into the structure below.

---

## Core Reference (load for any development work)

| # | Document | Scope |
|---|----------|-------|
| 001 | Development Guide | Agent creation, patterns, bugs, loops, data paths, checklist |
| 002 | System Architecture | Architecture, data flow, topics, idle timeout, cleanup, orchestration |
| 003 | Contracts & Standards | Component contracts, slot specs, CSS variables, rendering rules |

## Pipeline & Loop Reference

| # | Document | Scope |
|---|----------|-------|
| 004 | Improvement Loop | Discovery → audit → triage → fix → rerender cycle |
| 005 | Tool Pipeline | Tool suggestion, generation, deployment, cross-linking |
| 006 | News Feed Pipeline | Content sources, ingestion, triage, rendering, diversity |
| 007 | Adoption Pipeline | Site crawling, classification, content recreation, archetype |
| 008 | Vet Med Pricing Pipeline | Price scraping, extraction, evidence, JSON export |

## Infrastructure Reference

| # | Document | Scope |
|---|----------|-------|
| 009 | Model Infrastructure & Routing | Endpoints, health, swap/revert, Ollama, GPU, training data |
| 010 | Scheduler & Tasks | Kafka scheduler, scheduled_tasks table, concurrency groups |
| 011 | Database & Infrastructure | Connections, pgbouncer, MySQL, client schemas, credentials |
| 012 | Admin Dashboard & API | React SPA, nginx gateway, endpoints, VPN access |
| 013 | Content Governance | Locks, inline editing, briefs, regeneration, growth budget |
| 014 | Site Snapshots & Revert | Point-in-time capture, SQL functions, admin API, workflow actions |
| 015 | Batch Processing Architecture | Batch API integration, queue table, submitter/retriever pattern |

## Operational

| # | Document | Scope |
|---|----------|-------|
| 016 | Debugging Guide | Pod health, work items, scheduled tasks, orchestration, timeouts |

## Domain-Specific

| # | Document | Scope |
|---|----------|-------|
| 017 | Companies House Enrichment | Collection, matching cascade, detail/accounts fetch, financials |
| 018 | Canine Biology | Implementation plan for canine biology features |
| 019 | Tool Library | Component library, tool registry, matching |
| 020 | Tool Lifecycle | Tool creation, versioning, health checks, updates |
| 021 | Site Spec & Classifier | Classification architecture, spec aspects, archetype |
| 022 | Dynamic Applications | Interactive app generation guidelines |
| 023 | LLM Quality Testing | Model evaluation, prompt testing, quality metrics |

## Plans (review for currency)

| # | Document | Scope |
|---|----------|-------|
| P1 | Build Expand Plan | Build pipeline expansion roadmap |
| P2 | Public API Plan | Public API design |
| P3 | Admin API Plan | Admin API design |

---

## Source Mapping

Each consolidated doc traces back to its sources. When a source is marked "archived", its live content has been absorbed into the target doc and the original can be removed from active context.

### 001 — Development Guide
- **Base:** 001i_development_guide_new_agents_v9.md
- **Absorbed:** 004_unblocking_items.md (work item lifecycle), 005_extended_thinking.md (LLM config note), 001_config_driven_dont_remove_old_pages (nav sync issue), 016_workflow_data_path_validation.md (Appendix B), 014_loop_mechanisms_guide.md (Appendix C), 031_guidelines_model_update.md (model aliases update), 037_unresolved_what_it_does.md (work item lifecycle)

### 002 — System Architecture
- **Base:** 002d_system_architecture_v4.md
- **Absorbed:** 002de_quality_assurance_architecture_v2.md, 004_site_work_orchestrator.md, 010_idle_timeout_guide.md, 011_shared_topics.md, 014_cleanup.md, 003_random_notes (rollback pattern only)

### 003 — Contracts & Standards
- **Base:** 003e_contracts_and_standards_v5.md (no changes)

### 004 — Improvement Loop
- **Base:** 009f_improvement_loop_v5.md
- **Absorbed:** 009_build_pipeline_vs_improvement_loop_vs_audits.md, 002_component_standards_validation_design.md, 012_audit_needs.md (remaining items only)

### 005 — Tool Pipeline
- **Base:** 034_tool_pipeline_handoff_consolidated.md (strip handoff framing)

### 006 — News Feed Pipeline
- **Base:** 027i_news_feed_pipeline_consolidated_v4.md
- **Absorbed:** 027g_news_expansion_architecture.md, 036b_news_content_diversity_plan_v2.md

### 007 — Adoption Pipeline
- **Base:** 028e_adoption_and_infrastructure_layers_v5.md
- **Absorbed:** 028j_adoption_and_component_handoff_consolidated.md (live content only)

### 008 — Vet Med Pricing Pipeline
- **Base:** 031_vet_med_pricing_pipeline.md (supersedes 029_med_pricing_collection_plan.md, 029b_med_pricing_implementation_status.md)

### 009 — Model Infrastructure & Routing
- **Base:** 020e_gpu_and_model_infrastructure_v5.md
- **Absorbed:** 024b_agent_backup_and_swap_reference_v2.md, 021_model_swap_and_rollback.sql, 022_ai_endpoint_health_and_flywheel_llm_call_log.sql

### 010 — Scheduler & Tasks
- **Base:** 011_kafka_scheduler_guide.md
- **Absorbed:** 011b_scheduler_and_tasks_guide.md (check overlap)

### 011 — Database & Infrastructure
- **Base:** 016_database_connections.md
- **Absorbed:** 017b_creating_new_client_schemas_v2.md, 017_create_user_in_mysql.md (runbook commands only)

### 012 — Admin Dashboard & API
- **Base:** 019d_admin_dashboard_and_gateway_v4.md
- **Absorbed:** 019_admin_access_infrastructure.md

### 013 — Content Governance
- **Base:** 021d_admin_content_governance_plan_v4.md (no changes)

### 014 — Site Snapshots & Revert
- **Base:** 030_site_snapshots.md (no changes)

### 015 — Batch Processing Architecture
- **Base:** 035c_batch_processing_architecture_v3.md (latest version)

### 016 — Debugging Guide
- **Base:** 020_debugging_guide.md (no changes)

### 017 — Companies House Enrichment
- **Base:** 015f_companies_house_enrichment_plan_v7.md
- **Absorbed:** 022b_companies_house_matching_cascade_plan_v2.md

### 018-023 — Domain-specific docs
- Each is its base file with renumbering only.

### Archived (content absorbed, remove from active context)
003_random_notes_business_case_rollback.md, 004_unblocking_items.md, 005_extended_thinking.md, 006_useful_notes_for_llm.md, 006b_useful_notes_handoff_summary.md, 008_session_summary_20260308.md, 009_build_pipeline_vs_improvement_loop_vs_audits.md, 011_shared_topics.md, 012_audit_needs.md, 014_cleanup.md, 017_create_user_in_mysql.md, 019_admin_access_infrastructure.md, 021_model_swap_and_rollback.sql, 022_ai_endpoint_health_and_flywheel_llm_call_log.sql, 024b_agent_backup_and_swap_reference_v2.md, 026_drain_and_fix_ai_models_session_handoff_20260325.md, 028j_adoption_and_component_handoff_consolidated.md, 029_b_ch_vet_med_session_handoff_2026_03_26.md, 029_med_pricing_collection_plan.md, 029b_med_pricing_implementation_status.md, 031_guidelines_model_update.md, 035_batch_processing_architecture.md, 035b_batch_processing_architecture_v2.md, 036_news_content_diversity_plan.md, 037_unresolved_what_it_does.md
