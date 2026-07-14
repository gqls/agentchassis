# EXTRACTION U12 — docs024_key_docs_latest archives (old/, old_design_and_styling/, debugging_old/, archive_april_26/)
Extracted 2026-07-13. Files in scope: 158. Concepts found: 89.

Method note: this unit was processed by 7 parallel sub-extractions split by directory/family
(older1-A: 001/002/003 families; older1-B1: 005-012 families; older1-B2: 014-105 families;
old/ root + old_design_and_styling/; debugging_old/ full directory; archive_april_26 debugging-guide
family; archive_april_26 non-debugging files), then merged and deduplicated here. 8 duplicate
findings that surfaced independently across two or more sub-extractions (the same superseded/
abandoned concept evidenced by different archive drafts) were consolidated into 6 single entries
with combined provenance — flagged inline as "(merged from N independent findings)". One concept
(`Quality improvement flywheel`) was recategorized from a sub-agent's proposed `NEW:quality-improvement-flywheel`
into the existing seed slug `finetuning-flywheel`; one proposed category
(`NEW:quality-assurance-pipeline`) was dropped because its concept was merged into the
`system-architecture`-categorized "QA Architecture folded in" entry below (that's where the content
now actually lives, per the live doc).

## Coverage
| file | treatment |
|---|---|
| docs/agent_docs/docs024_key_docs_latest/old/older1/001_development_guide.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/001_development_guide_april26.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/001_development_guide_new_agents.md | family-delta (subset of v3, no unique concepts) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/001b_development_guide_new_agents_v3.md | family-delta (incremental, all content persists into v4+) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/001d_development_guide_new_agents_v4.md | family-delta (incremental, purely additive over v3) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/001e_development_guide_new_agents_v5.md | family-delta (incremental, adds LLM Infrastructure section) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/001f_development_guide_new_agents_v6.md | family-delta (incremental, adds bugs #11-15) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/001g_development_guide_new_agents_v7.md | family-delta (adds Domain Submission section; 1-line diff from v8) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/001h_development_guide_new_agents_v8.md | family-latest (highest archive dev-guide version, read in full) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/002_system_architecture.md | family-delta (early pre-QA-merge subset, ~same maturity as v3) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/002_system_architecture_april26.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/002b_system_architecture_v2.md | family-delta (incremental over baseline) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/002c_system_architecture_v3.md | family-latest (highest non-QA archive version, read in full) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/002d_quality_assurance_architecture.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/003_contracts_and_standards.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/003b_contracts_and_standards_v2.md | family-delta (incremental, adds CSS Colour Inheritance Model) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/003c_contracts_and_standards_v3.md | family-delta (incremental, adds Query DB Parameterisation Contract) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/003d_contracts_and_standards_v4.md | family-delta (incremental, adds Component Input Schema v2 + Content Validation) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/003_contracts_and_standards_v5.md | family-delta (incremental, introduces Component Quality Contract) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/003_contracts_and_standards_v6.md | full (contains Component Quality Contract in full — later dropped) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/003_contracts_and_standards_v7.md | family-latest (highest archive contracts version) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/005_build_expand_plan.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/006_news_feed_pipeline.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/007_adoption_pipeline.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/old/older1/007_adoption_pipeline_v2.md | family-latest |
| docs/agent_docs/docs024_key_docs_latest/old/older1/007_adoption_pipeline_v2.patch.diff | header-scan (mechanical patch, transforms v2_april26 → v2, dated 2026-04-20/21) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/007_adoption_pipeline_v2_april26.md | family-delta (superseded by the patch) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/007a_public_api_plan_v1.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/008a_admin_api_plan_v1.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/009_improvement_loop.md | family-delta (tiny transcript-style fragment, earliest) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/009b_improvement_loop_v2.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/old/older1/009c_improvement_loop_v3.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/old/older1/009d_improvement_loop_v4.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/old/older1/009e_improvement_loop_v5.md | family-latest |
| docs/agent_docs/docs024_key_docs_latest/old/older1/010_tool_library_guide.md | family-delta (v1) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/010b_tool_library_guide_v2.md | family-latest |
| docs/agent_docs/docs024_key_docs_latest/old/older1/012_tool_lifecycle_guide.md | family-delta (v1) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/012b_tool_lifecycle_guide_v2.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/old/older1/012c_tool_lifecycle_guide_v3.md | family-latest |
| docs/agent_docs/docs024_key_docs_latest/old/older1/014_site_snapshots_and_revert.md | family-delta (pure subset of live doc) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/015_batch_processing_architecture.md | family-delta (real deltas found vs v2) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/016_debugging_guide.md | full (early draft, terse extraction) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/016_debugging_guide_v2_april26.md | full (second early draft) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/017_creating_new_client_schemas.md | family-delta (live counterpart is a section inside 011_database_and_infrastructure.md) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/020_gpu_and_model_infrastructure.md | full (v1 — earliest, undecided-options version) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/020_tool_lifecycle.md | family-delta (pure subset of live doc) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/020b_gpu_and_model_infrastructure_v2.md | family-delta (diff-confirmed purely incremental) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/020d_gpu_and_model_infrastructure_v4.md | full (highest archive version) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/024_link_management.md | family-delta (mostly refinement, no strongly abandoned ideas) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/025_palette_layout_typography_migration(1).md | family-delta (diff-confirmed incremental) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/025_palette_layout_typography_migration(2).md | family-delta (diff-confirmed incremental) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/025_palette_layout_typography_migration.md | full (earliest draft — dropped "pole" taxonomy) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/026_component_regeneration_flow.md | family-delta (pure subset of live doc) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/026_design_and_site_planner_v1.md | full (real delta vs v2 — superseded removal plan) |
| docs/agent_docs/docs024_key_docs_latest/old/older1/105_dispatch-pipeline-failures-report-v2.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/105_dispatch-pipeline-failures-report.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/105_dispatch-pipeline-failures-report_v3.md | full |
| docs/agent_docs/docs024_key_docs_latest/old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md | full (standalone, no live counterpart found) |
| docs/agent_docs/docs024_key_docs_latest/old/001_development_guide.md | family-delta (diffed against 001_development_guide(5).md) |
| docs/agent_docs/docs024_key_docs_latest/old/009_model_infrastructure.md | family-delta (diffed against same-name live root file) |
| docs/agent_docs/docs024_key_docs_latest/old/029_site_plan_and_reconciler.md | family-delta (diffed against 029_site_plan_and_reconciler(2).md) |
| docs/agent_docs/docs024_key_docs_latest/old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md | family-delta (diffed against 030_phase1_plan_and_reconciler(5).md) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/016_debugging_guide_v2.md | family-delta (subsumed into 016_debugging_guide_v2_58_consolidated.md) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/FOCUS_design_and_styling.md | full (no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-14_v2.md | full (no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md | full (no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md | full (no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md | full (no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/FOCUS_design_and_styling_computed_styles_extraction_phase2.md | full (no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/FOCUS_design_and_styling_fp_extract_css_vars_integration.md | full (no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update2.md | family-delta (superseded within-batch by update4) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update3.md | family-delta (superseded within-batch by update4) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md | full (latest of the update2/3/4 sequence, no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/old_design_and_styling/PHASE_4_4_cleanup_summary.md | full (no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/001_development_guide(1).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/001_development_guide.md | family-latest of this mini-family |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/001_development_guide.patch.diff | header-scan |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/003_contracts_and_standards_v10.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/003_contracts_and_standards_v11.md | family-latest |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/003_contracts_and_standards_v8.md | family-delta (earliest of the four) |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/003_contracts_and_standards_v9.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide(3).md | family-delta (duplicate snapshot between v2(6) and v2(7)) |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(1).md | family-delta (contains the adapter-deployment section absent elsewhere) |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(10).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(11).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(12).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(13).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(14).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(15).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(17).md | family-latest (1576 lines, full-diffed vs live) |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(2).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(3).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(4).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(5).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(6).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(7).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(8).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2(9).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2.md | family-delta (earliest snapshot) |
| docs/agent_docs/docs024_key_docs_latest/debugging_old/016_debugging_guide_v2.patch.diff | header-scan |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_47(1).md | full (source of 2 recovered §9 entries) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_48.md | family-delta (branch point, drops 3 entries) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_49.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_49(1).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_49(2).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_50.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_50(3).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_51.md | header-scan |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_52.md | header-scan |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_53.md | header-scan |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_54.md | header-scan |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_55.md | header-scan |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_56.md | header-scan |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_debugging_guide_v2_57.md | family-delta (full diff vs live confirms only 2 recovered entries) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide.md | family-delta (v1 only, subsumed) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide(1).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide(2).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide(3).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide(4).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide(5).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide(6).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide(8).md | family-delta (promoted-to-live copy, additive only) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_5_.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_6_.md | family-delta (full-diffed vs live, purely additive) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_6_(2).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_6_(3).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_6_(4).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_7.md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_7(1).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_7(2).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_7(4).md | full (divergent fork — genuinely absent from canonical live doc) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016b_debugging_guide_7_3_.md | full (near-duplicate of 7(4), confirms divergent fork) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/001i_development_guide_new_agents_v9.md | full (fully absorbed verbatim into live) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/002d_system_architecture_v4.md | full (fully absorbed into live) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/002de_quality_assurance_architecture_v2.md | full (verbatim in live 002_system_architecture(4).md) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/003e_contracts_and_standards_v5.md | full (headers present live; painting-section delta extracted) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/004_site_work_orchestrator.md | full (abstract pattern absorbed, Go-level detail dropped) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/005b_build_expand_plan_v2.md | full (byte-identical to live plus one disclaimer line) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/006_useful_notes_for_llm.md | full (ephemeral notes, no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/006b_useful_notes_handoff_summary.md | full (ephemeral handoff, no live counterpart) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/009f_improvement_loop_v5.md | full (fully absorbed verbatim into live) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/010c_tool_library_guide_v3.md | full (fully absorbed into live) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/011_kafka_scheduler_guide.md | full (byte-identical to live 010_scheduler_and_tasks.md) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/011b_scheduler_and_tasks_guide.md | full (richer than live — main finding of this file) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/012d_tool_lifecycle_guide_v4.md | full (fully absorbed into live, live is a superset) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/014_loop_mechanisms_guide.md | full (fully absorbed verbatim as dev-guide Appendix C) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/015_consolidated_site_spec_classifier_architecture.md | full (identical to live except title) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/016_workflow_data_path_validation.md | full (fully absorbed into live dev guide appendix) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/017b_creating_new_client_schemas_v2.md | full (fully absorbed into 011_database_and_infrastructure.md) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/018_dynamic_application_guidelines.md | full (identical to live except title) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/019_admin_access_infrastructure.md | full (condensed into live, runnable configs dropped) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/019_tool_library(1).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/019_tool_library.md | family-delta (oldest of the 3-file family) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/019_tool_library_2_.md | full (family-latest, byte-identical to live) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/020_tool_lifecycle(1).md | family-delta |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/020_tool_lifecycle.md | family-delta (oldest of the 3-file family) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/020_tool_lifecycle_2_.md | full (family-latest, byte-identical to live) |
| docs/agent_docs/docs024_key_docs_latest/archive_april_26/020e_gpu_and_model_infrastructure_v5.md | full (byte-identical to first 229 lines of live) |

## Concepts

### Quality Assurance Agent Architecture — folded into system-architecture, not abandoned
*(merged from 2 independent findings)*
- **category:** system-architecture
- **status-signal:** superseded (as a standalone numbered doc; the architecture itself is deployed/partial)
- **status-evidence:** The standalone `002d_quality_assurance_architecture.md` (older1) and its later revision `002de_quality_assurance_architecture_v2.md` (archive_april_26) both appear verbatim as a "# 002d — Quality Assurance Agent Architecture" section inside live `002_system_architecture(4).md` (starting at line 897), continuing the main doc's Resolved-Decisions numbering (18-25) and extending it with a new "Layer 0: Pre-Generation Data Triage (plan_sections)" section, a "Content Validation as a Third Mode" table, and two further resolved decisions (24 "Quality gates before generation, not just after", 25 "needs_human_review is a first-class status") absent from any archived 002d draft. Its "Responsibility Boundaries" table was also updated to match the later composition/design-planner split.
- **what:** A three-layer QA model: Layer 1 structural/algorithmic checks (free, no LLM), Layer 2 LLM-assisted design/content audit (grouped agents sharing context, one LLM call per group), Layer 3 LLM-required strategic review (dream-spec gap analysis); plus a later-added Layer 0 pre-generation data triage (`plan_sections`). Includes the "promotion pattern" (a check starts as a `query_database` action step and is promoted to a spawned sub-agent only once it needs multi-step workflows or external calls) and the rule that audit agents "enforce, not override" the classifier/planner's stated intent. This was never a genuinely dropped concept area — it was consolidated into the numbered `002_system_architecture` doc rather than kept standalone, and then actively extended.
- **sources:** old/older1/002d_quality_assurance_architecture.md (whole file); archive_april_26/002de_quality_assurance_architecture_v2.md (whole file); docs024_key_docs_latest/002_system_architecture(4).md#"002d — Quality Assurance Agent Architecture" (line 897+)
- **relations:** design agent responsibility split (site-design-planner/webdesign-agent); improvement-loop (004); site-spec-and-classifier (021); triage drain loop
- **verify-later:** confirm `design-audit-agent`, `visual-design-auditor`, `content-quality-auditor`, `site-review-agent` agent_definitions still implement the three-layer split; confirm `plan_sections` and `needs_human_review` status are implemented as described.

### ExtractActionInputs nested-source collision affects required fields too (corrected scope)
*(merged from 2 independent findings)*
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Live `001_development_guide(5).md`'s opening consolidation note states it "supersedes the prior copy (which still had the older 'Field Name Collisions' wording — the nested-source collision affects **required and optional** fields, corrected here)"; live text: "This loop iterates the full field list — both Required and Optional. It does not distinguish between them."
- **what:** Two independent archived drafts (`old/older1/001h_development_guide_new_agents_v8.md` and `old/001_development_guide.md`) both claimed the nested-source lookup collision in `ExtractActionInputs` (an unmapped field silently matching `site_record.<field>`/`input_data.<field>`) applies only to optional fields. The live doc corrects this: the nested-source loop iterates the full field list regardless of Required/Optional status; required fields (e.g. `site_id`) carry the same latent risk, it's just usually masked because earlier resolution strategies (0-2) resolve them first. The live doc adds a "latent risk (required field)" example and recommends collision-free names (`target_site_id`) for new required fields, while leaving existing code alone unless it actually misbehaves.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md#"Field Name Collisions"; old/001_development_guide.md#"Field Name Collisions"; docs024_key_docs_latest/001_development_guide(5).md#"Field name collisions"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Note on the target_site_id input field name"
- **relations:** ExtractActionInputs / datahelpers resolution cascade (Strategy 0-4); whole-blob input_data anti-pattern; target_site_id naming convention
- **verify-later:** confirm `platform/orchestration/datahelpers/action_inputs.go`'s nested-source loop still doesn't distinguish Required/Optional.

### Anthropic client temperature parameter removed unconditionally
*(merged from 2 independent findings)*
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** Live dev guide, dated inline "(2026-05-27)": "The Anthropic client no longer sends a temperature parameter on any call... Opus 4.7+ returns a 400 for any non-default temperature."
- **what:** Archived drafts (`old/older1/001h_development_guide_new_agents_v8.md`'s "Extended Thinking Configuration" section and `old/001_development_guide.md`'s "Extended thinking" section) state temperature is stripped only when `budget_tokens` (extended thinking) is set — implying ordinary non-thinking calls still send temperature. The live doc broadens this: because newer Claude Opus models reject any non-default temperature outright, the Anthropic client now omits temperature unconditionally on every call, thinking or not. Temperature remains honoured for other providers (e.g. Ollama) — only the Anthropic client special-cases it.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md; old/001_development_guide.md#"Extended thinking"; docs024_key_docs_latest/001_development_guide(5).md#"Temperature (2026-05-27)"
- **relations:** model-infrastructure (endpoints, provider clients); LLM call logging (`__sent_temperature`)
- **verify-later:** grep the Anthropic client source for unconditional temperature stripping.

### site_work_items work-routing column renamed domain → pipeline
*(merged from 2 independent findings)*
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Live bug-log entry #18 in `001_development_guide(5).md`: "The `domain` column on site_work_items was renamed to `pipeline` in a migration."
- **what:** Two archived dev-guide drafts each devote a "Lessons Learned" section to `site_work_items.domain` being an internal work-routing namespace ("build"/"maintenance"/"marketing") that collides confusingly with the website's actual domain (e.g. "gaswholesalers.com") — citing real bugs this caused (a dispatch-loop filter mismatch, and a CSS-generation item never dispatching because it was written with `domain:"design"` instead of `domain:"build"`). Rather than keep relying on documentation warnings, the column was renamed to `pipeline` at the schema level, eliminating the ambiguity outright; the live doc drops the explanatory section entirely in favour of a terse bug-log line.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md#"Work item domain is NOT the site domain"; old/001_development_guide.md#"Work item domain is NOT the site domain"; docs024_key_docs_latest/001_development_guide(5).md#18; old_design_and_styling/016_debugging_guide_v2.md#"Schema reminder"
- **relations:** dispatch-loop input_mapping; site_work_items table
- **verify-later:** confirm `site_work_items.pipeline` column exists in the current schema and no code still reads/writes `domain` for this purpose.

### CSS section-colour model: inheritance → hardcoded dark-section variables → renderer-computed defaults → token-referencing painting sections
*(merged from 4 independent findings)*
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** Four independent archive lines converge on the same evolution. Earliest baseline (`old/older1/003_contracts_and_standards.md`, `old_design_and_styling/FOCUS_design_and_styling.md`): plain CSS inheritance, dark-background components just set literal `color:#fff`/`inherit`, no `--section-*` variables exist. Middle era (`old/older1/003_contracts_and_standards_v2..v7.md`, `debugging_old/003_contracts_and_standards_v8..v11.md`, `archive_april_26/003e_contracts_and_standards_v5.md`): a `--section-*` custom-property contract keyed off a boolean `is_dark_section` column, with LITERAL hardcoded values (`--section-heading:#ffffff`, `rgba(255,255,255,0.9)`) enforced by `ValidateDarkSectionContract()`. An intermediate renderer change (`old_design_and_styling/PHASE_4_4_cleanup_summary.md`): the renderer's `buildSectionDefaults` began computing these values automatically from palette luminance (WCAG-based), removing the manual per-component declaration burden. Live (`003_contracts_and_standards(8).md`): the "Section painting contract" — `is_dark_section` is demoted to inert catalogue metadata ("MUST NOT key styling"), and any section that paints its own background must instead RE-EXPORT `--section-*` as references to theme tokens via one of four models (pair band, palette band, image/ink-derived, ambient/no-background) using `color-mix()`, so colours flip automatically with the site's scheme; literal colours are forbidden and mechanically enforced by `fix_forced_text_colors`.
- **what:** Documents the multi-year hardening of how section backgrounds get correctly-coloured text: from ad hoc inline colours, to a hardcoded-value contract gated on a boolean flag (which locked every dark section into literal white-on-dark), through a renderer-side automation step, to the current token-referencing "painting" model that treats `is_dark_section` as inert metadata and derives colours mechanically from the active palette.
- **sources:** old/older1/003_contracts_and_standards.md; old/older1/003_contracts_and_standards_v7.md#"Section Context Variable Contract (Dark Sections)"; debugging_old/003_contracts_and_standards_v11.md#"Section Context Variable Contract (Dark Sections)"; archive_april_26/003e_contracts_and_standards_v5.md#"Section Context Variable Contract"; old_design_and_styling/FOCUS_design_and_styling.md#"CSS Colour Inheritance Model"; old_design_and_styling/PHASE_4_4_cleanup_summary.md#"Phase 4.5"; docs024_key_docs_latest/003_contracts_and_standards(8).md#"Section Context Variable Contract (Painting Sections)"
- **relations:** styling-render-pipeline (036); design-composition (site-design-planner palette resolution feeds these tokens); fix_forced_text_colors action
- **verify-later:** grep deployed component templates for literal `#ffffff`/`rgba(255,255,255,` inside `--section-*` declarations to confirm the old hardcoded pattern is gone; inspect `fix_forced_text_colors` and `buildSectionDefaults` Go source.

### design_reference / design_intent spec-aspect split
*(merged from 2 independent findings)*
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** `FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md`: "Replaces the old vague `design` spec"; `027_design_and_site_planner_v2.md`'s "Related Specs" table (an independent, later live doc) confirms both `design_reference` and `design_intent` as live, read spec aspects with defined priority cascades.
- **what:** `design_reference` holds concrete values (hex colours, font families, CSS variables, spacing) extracted mechanically (no LLM) from an adopted site's crawled HTML/CSS — a historical, immutable record. `design_intent` holds semantic creative direction (e.g. "dark IDE aesthetic... start here"), deliberately non-prescriptive so the improvement loop and webdesign-agent retain creative room; it may be auto-generated at adoption time or written later by a strategist/human. Together they replace a single, vague, LLM-guessed `design` spec aspect that conflated historical fact with creative direction (see the separate "Unified design spec aspect for adopted sites" concept below for that earlier, superseded state). Guiding principle: "design reference is history, design intent is direction" / "every build is conceptually an adoption."
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md#"Key Decisions Made"; old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"Principles Restated"; old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Decisions Made & Rationale"; docs024_key_docs_latest/027_design_and_site_planner_v2.md#"Related Specs"
- **relations:** Unified design spec aspect for adopted sites (the superseded precursor, below); webdesign-agent three-way design priority; palette-locked-until-design_intent policy; design agent write-back resolution
- **verify-later:** confirm `design` spec aspect is no longer written anywhere; check `site_specs` for population rate of `design_reference`/`design_intent` across adopted sites.

---
*(remaining concepts below are as extracted by the 7 sub-slices, grouped by their originating file cluster)*

### Design agent responsibility split — site-design-planner (composition) vs webdesign-agent (execution)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Live 002_system_architecture(4).md line 596: "The earlier 'one agent generates brand + CSS' shape is **superseded** by the composition/execution split above."
- **what:** All archived versions of the system-architecture doc (baseline through the april26 near-final draft) model design as owned by a single agent, `webdesign-agent`, which does brand analysis, colour/typography/spacing decisions, AND CSS generation — with `brand-designer`, `layout-architect`, and `style-generator` listed as "Future split"/"Planned" agents that never materialized under those names. The live doc replaces this with a Composition/Execution/Maintenance split: `site-design-planner` (new agent) deterministically resolves layout (weighted, scheme-aware match against a shared `layouts` library), typography (match-or-new against `typography_sets`), and a site-specific palette via signal cascades, then installs `css_themes` + `style_collections` + a `resolved_composition` decision-record; `webdesign-agent` is narrowed to rendering/committing `/assets/css/styles.css` from that installed composition — "the only writer of styles.css." The `Design | webdesign-agent | Colour palette, typography, spacing, CSS` row in the Responsibility Boundaries table is likewise split into separate "Composition" (site-design-planner) and "Render" (webdesign-agent) rows, with a `needs_new_layout_candidate` HITL escalation replacing the old simple "search → maybe reuse → maybe generate" theme-growth description.
- **sources:** old/older1/002c_system_architecture_v3.md#"Design Agent Family", #"Theme library growth"; old/older1/002d_quality_assurance_architecture.md#"Classifier → Planner → Design Agent → Audit Agent"; docs024_key_docs_latest/002_system_architecture(4).md#"Composition: how a site's design is resolved and installed", #"Classifier → Planner → Design Agent → Audit Agent"
- **relations:** superseded planned agents brand-designer / layout-architect / style-generator (never built under those names); fork_theme_composition.go resolvers; QA architecture's "Responsibility Boundaries" chain
- **verify-later:** confirm `site-design-planner` agent_definitions row and `resolve_composition_layout/typography/palette` actions exist and are active.

### Component Quality Contract (scoring formula, quality columns)
- **category:** contracts-and-standards
- **status-signal:** abandoned
- **status-evidence:** Introduced in `old/older1/003_contracts_and_standards_v6.md` ("## Component Quality Contract"); absent from v7 onward and absent from live `003_contracts_and_standards(8).md` (no "Component Quality Contract" heading), though `quality_score`/`quality_issues` fields still appear inline in the live doc's "Component Creation & Regeneration Contract" JSON examples.
- **what:** v6 fully specified a quality-tracking contract for `content_components`: eight quality columns (`template_variable_count`, `schema_field_count`, `template_closed`, `schema_template_synced`, `has_data_component`, `quality_score` 0-100, `quality_checked_at`, `quality_issues`), a scoring formula starting at 100 with fixed deductions per violation, three computation triggers (on-insert, periodic audit by `component-quality-auditor`, targeted rescan), an automatic `needs_component_regeneration` work item below score 50, and planner preference for higher-scored components. This entire standalone contract section vanished between v6 and v7 and was never restored, even though the live system-architecture doc still lists a `component-quality-auditor` agent and the live contracts doc still surfaces `quality_score`/`quality_issues` as return-payload fields — suggesting the mechanism may partly persist in code/DB while its dedicated documentation disappeared.
- **sources:** old/older1/003_contracts_and_standards_v6.md#"Component Quality Contract"; docs024_key_docs_latest/003_contracts_and_standards(8).md (residual field mentions); docs024_key_docs_latest/002_system_architecture(4).md (component-quality-auditor agent row)
- **relations:** StoreGeneratedComponentAction / component regeneration contract; component-quality-auditor agent
- **verify-later:** check whether `compute_component_quality`/`ScoreAndPersistComponent` and the `content_components` quality columns still exist and are actively populated.

### `query.{name}` field-source resolution timing (render-time → plan-time)
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** Archive v7 "Source prefixes" table: `query.{name}` resolved "At render time." Live table: resolved "At `plan_sections` time."
- **what:** In the Component Input Schema v2 Contract's source-prefix table, the `query.{name}` prefix (used for blog posts/categories lists) moved from being resolved at page-render time to being resolved earlier, during `plan_sections`, with the result projected into the field's declared shape. Consistent with the broader shift toward front-loading data-availability checks (the "Layer 0" pre-generation triage) rather than discovering missing/stale query data only at render.
- **sources:** old/older1/003_contracts_and_standards_v7.md#"Source prefixes"; docs024_key_docs_latest/003_contracts_and_standards(8).md#"Source prefixes"
- **relations:** Component Input Schema v2 Contract; plan_sections / Layer 0 pre-generation data triage; page_rerender item_type
- **verify-later:** check plan_sections Go action for query-prefix handling, confirm it projects results to field shape at plan time.

### Milestone-tagged site-spec history with inline git-snapshot function
- **category:** site-snapshots-and-revert
- **status-signal:** superseded
- **status-evidence:** Archive `site_specs` schema carries `milestone`, `superseded_by` columns and a `CommitSpecSnapshot` Go function called inline; live doc drops `milestone`/`superseded_by` entirely, replaces inline snapshotting with a work-item-triggered `snapshot-agent`, adds a bounded "last 5 rows" pruning policy, and drops the legacy `page_components.content_snapshot`/`schema_snapshot` columns.
- **what:** The original design kept unbounded site-spec history in the DB, labelled key rows with a `milestone` string (`initial_research`, `post_build`, `rebrand_q2`...), and relied on a bare Go function invoked directly by completing actions to write a `.site-spec.json` git checkpoint. Content-level rollback used `content_snapshot` on `page_components`. This whole history/rollback substrate was replaced by a decoupled model: `site_specs` prunes to last-5-per-aspect, `page_component_history` is a dedicated append-only table for component rollback, and snapshotting became an ordinary dispatched work item (`needs_snapshot` → `snapshot-agent`) rather than an inline side-effect call.
- **sources:** old/older1/005_build_expand_plan.md#"Table: site_specs", #"Git Spec Snapshots"; docs024_key_docs_latest/P1_build_expand_plan.md#"Removing legacy columns", #"Snapshots"
- **relations:** superseded by snapshot-agent + page_component_history; content-governance locking model
- **verify-later:** confirm in DB whether `page_components.content_snapshot`/`schema_snapshot` columns still exist.

### "Insights section" as the Tier-2 news-feed expansion target
- **category:** news-feed-pipeline
- **status-signal:** superseded
- **status-evidence:** Archive Tier 2 = "Insights section... Future"; live Tier 2 = "News listing page... ✅ Working," curated/rewritten-article idea folded into Tier 3.
- **what:** The original three-tier roadmap treated a dedicated `/insights/` section of rewritten, curated articles as the second expansion tier after homepage snippets. When the archive-first news-index/listing page was actually built, it took the Tier-2 slot instead, and the "curated rewritten articles" idea was pushed down into Tier 3, where `article-rewriter` and `feed-publisher` remain listed as not-yet-built in both versions.
- **sources:** old/older1/006_news_feed_pipeline.md#"Expansion Roadmap"; docs024_key_docs_latest/006_news_feed_pipeline_v2.md#"Expansion Roadmap"
- **relations:** article-rewriter/feed-publisher agents (still unbuilt)
- **verify-later:** check whether a `/insights/` route or `article-rewriter` agent definition exists anywhere.

### Single-agent adoption trigger (positional domain, no orchestrator wrapper)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** `007_adoption_pipeline_v2_april26.md`: a dedicated `site-adoption-agent` triggered directly via `./trigger-adopt-site.sh gamedesign.uk`; the patch rewrites this into "Two agents, one thin wrapper," documented fully in live `007_adoption_pipeline_v4.md`.
- **what:** Adoption originally ran as one agent invoked directly by a shell script with a positional domain argument, mixing "site being crawled" and "site being built" into a single identifier. Replaced by a thin `site-adoption-orchestrator` (spawn → call → complete) that spawns `site-adoption-agent` as its own K8s Job, and a JSON trigger payload separating `target_url` (crawl source) from `destination_domain` (site being built) — while keeping the old `url`/`domain` shape as legacy-compatible input.
- **sources:** old/older1/007_adoption_pipeline_v2_april26.md#"The adoption agent", #"Adoption modes"; old/older1/007_adoption_pipeline_v2.patch.diff; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"The adoption agent", #"Source vs destination"
- **relations:** "every pod-running agent needs a parent that spawned it" (development-guide)
- **verify-later:** confirm `site-adoption-orchestrator` agent_definitions row exists and `trigger-adopt-site.sh` uses the JSON payload shape today.

### Unified `design` spec aspect for adopted sites (superseded precursor)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** `007_adoption_pipeline.md` (v1): single `design` spec aspect; live `007_adoption_pipeline_v4.md` has a dated addendum, "Design Fingerprint Pipeline (added 2026-04-12)," documenting the two-aspect replacement.
- **what:** The earliest adoption design captured only one `design` spec aspect, generated by the LLM alongside identity/structure classification — a single blended palette-and-typography guess with no separation between what the source site actually used and what the new site should aim for. Replaced by the `design_reference`/`design_intent` split (see merged entry above).
- **sources:** old/older1/007_adoption_pipeline.md#"What gets stored where"; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"Design Fingerprint Pipeline (added 2026-04-12)"
- **relations:** design_reference/design_intent spec-aspect split (its replacement)
- **verify-later:** check `site_specs` for any legacy rows with `aspect='design'` from pre-2026-04-12 adoptions never migrated.

### Two-stage adoption processing (LLM classifies, Go extracts) → three-stage
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Archive header: "Two-stage processing: LLM classifies, Go extracts." Live header: "Three-stage processing: Go extracts design, LLM classifies, Go extracts content."
- **what:** Early adoption split work into just two stages — lightweight LLM classification from page summaries, then Go-only content extraction. The later design inserts a Go-only design-fingerprint extraction stage (colours/fonts/CSS vars/layout via goquery) ahead of LLM classification, on the principle "don't ask an LLM to read hex values when a regex can do it."
- **sources:** old/older1/007_adoption_pipeline.md#"Two-stage processing"; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"Three-stage processing"
- **relations:** unified design spec aspect (above), design_reference/design_intent split
- **verify-later:** `extract_design_fingerprint_action.go`/`enrich_fingerprint_with_css_action.go` existence and wiring.

### Work-item HITL model: approve/reject endpoints on pending_review status
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** `007a_public_api_plan_v1.md`: `POST /work-items/:item_id/approve|reject`; live `P2_public_api_plan.md`/`P3_admin_api_plan.md` have no approve/reject endpoints or `pending_review`/`rejected` statuses anywhere.
- **what:** The original API plan modelled human review as a binary approval gate on work items, with specs read-only initially. Replaced end-to-end by `needs_human_review` items with three resolution paths (provide missing spec data + retry, retry unchanged, or dismiss with a resolution note), and `PATCH /specs/:aspect` as a first-class, versioned write path feeding that retry flow.
- **sources:** old/older1/007a_public_api_plan_v1.md#"Work Items (build progress + HITL)"; docs024_key_docs_latest/P2_public_api_plan.md#"HITL Review Flow"
- **relations:** content-governance (locks, HITL)
- **verify-later:** grep core-manager handlers for any surviving `pending_review`/`HandleApproveWorkItem`/`HandleRejectWorkItem`.

### Admin work-item reassign + force-complete override endpoints
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** `008a_admin_api_plan_v1.md` E3 table has `reassign`/`force-complete`; live `P3_admin_api_plan.md`'s equivalent table has neither, only generic `PATCH`, `retry`, `resolve` (all Implemented).
- **what:** The original admin plan gave two narrow, single-purpose override endpoints for stuck work items: reassign the handler agent, or force-mark-complete with an arbitrary result. Generalised instead into one `PATCH` endpoint plus the shared `retry`/`resolve` pair — reassign and force-complete as distinct named actions never shipped.
- **sources:** old/older1/008a_admin_api_plan_v1.md#"E3: Work item administration"; docs024_key_docs_latest/P3_admin_api_plan.md#"E3: Work item administration + HITL review — IMPLEMENTED"
- **relations:** work-item HITL model (above)
- **verify-later:** confirm `site_admin_handlers.go` has no `HandleReassignWorkItem`/`HandleForceComplete`.

### Colour-fix algorithmic detail (countHardcodedColorComponents / findForcedTextColors)
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** Full detail present in `009b_improvement_loop_v2.md`; deleted outright from `009c_improvement_loop_v3.md` onward and absent from live `004_improvement_loop.md` (table-row summary only).
- **what:** Documented exact algorithmic mechanics for two `design-discovery-agent` colour checks: `hardcoded_section_colors` and `forced_text_colors` (parses `<style>` blocks, flags only child text elements, skips container/link rules), with a WCAG AA 4.5:1 contrast safety check and `--section-*` contract injection. Pruned from the docs from v3 onward in favour of a one-line table entry.
- **sources:** old/older1/009b_improvement_loop_v2.md#"Colour Fix Detail"; docs024_key_docs_latest/004_improvement_loop.md
- **relations:** color-variable-fixer handler; contracts-and-standards CSS variable contract
- **verify-later:** `fix_hardcoded_colors`/`findForcedTextColors` Go source accuracy.

### Per-site, per-audit-type cadence configuration (maintenance_profile.audit.{type})
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** Appears identically in v2-v4 as a "## Configuration" section, each time caveated "future enhancement." Absent from v5 and live, which document a simpler global 60-day auto-reset with no per-audit-type knobs.
- **what:** Three consecutive doc versions carried a designed-but-never-built configuration surface: per-site JSON config letting each audit type be individually enabled/disabled with its own re-run interval. Quietly dropped rather than implemented.
- **sources:** old/older1/009b_improvement_loop_v2.md#"Configuration"; old/older1/009d_improvement_loop_v4.md#"Configuration"; docs024_key_docs_latest/004_improvement_loop.md
- **relations:** Audit Pass Cap / Auto-reset mechanism (its replacement)
- **verify-later:** check `sites.settings.maintenance_profile` rows for leftover `audit.{type}` keys.

### Acceptance-test cheap-LLM verification call gating lock + retry
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** Documented in `009c_improvement_loop_v3.md`/`009d_improvement_loop_v4.md`, incl. literal verification prompt. Live `004_improvement_loop.md` retains `acceptance_test` as a required field but documents no corresponding verification-call step.
- **what:** Each finding carried an `acceptance_test` enabling a cheap follow-up LLM call after a fix: feed fixed HTML back, get YES/NO, gating section lock (pass) or retry up to `max_fix_attempts` before escalating to `needs_human_review`. The field survived but the explicit verify-then-lock mechanism dropped out of documentation by v5/live.
- **sources:** old/older1/009c_improvement_loop_v3.md#"Structured Findings Format"; docs024_key_docs_latest/004_improvement_loop.md#"1. Finding Cap"
- **relations:** Section Locking; Finding Cap
- **verify-later:** search for a dedicated verification-call step (`verify_fix`/`check_acceptance_test`) in fixer code.

### Content-writer chrome double-injection bug and chrome-ownership rule
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** `009d` v4 "## Content Writer Chrome Fix (v4)": full bug narrative + cleanup; only a one-line changelog bullet survives into v5/live.
- **what:** A production bug rule: site chrome (header/footer/head) must be injected exactly once, only at the rerender/assembly step — never by the content writer. Fix set all three inject flags false on `page-content-writer`'s `compile_page` step plus a cleanup pass removing baked-in chrome components.
- **sources:** old/older1/009d_improvement_loop_v4.md#"Content Writer Chrome Fix (v4)"; docs024_key_docs_latest/004_improvement_loop.md (changelog line only)
- **relations:** contracts-and-standards (component/slot contract); site-component-linker
- **verify-later:** confirm `page-content-writer` inject flags remain false; check for reappearance of baked-in header/footer components.

### Audit finding dedup + blocked-item filtering algorithm (write_audit_findings)
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** Full three-step algorithm documented in v3/v4; not present in v5/live (mentioned only in passing, then dropped even from the summary line).
- **what:** `write_audit_findings` was documented as implementing three dedup/safety layers: bulk-preloading blocked item keys, a broader item_type+page match against existing blocked items, and item-key-based dedup against pending items. This mechanism-level detail disappears from the documentation surface after v4.
- **sources:** old/older1/009d_improvement_loop_v4.md#"Finding Dedup and Blocked Item Filtering"; docs024_key_docs_latest/004_improvement_loop.md
- **relations:** Finding Cap; Triage Drain Controls
- **verify-later:** confirm `write_audit_findings` still implements bulk-preload + item_key pattern.

### Tag-based deterministic tool-to-site matching (matchToolToSite)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive `findMissingTools` matches via site-type/industry affinity plus a "universal tools" carve-out; live `020_tool_lifecycle(2).md` Bug History: "the matchToolToSite function classified security/password/privacy as universal, deploying a password checker to every site (including gas wholesalers). Fixed by removing tag-based matching entirely."
- **what:** The original tool-suggestion mechanism was a deterministic Go function comparing a library tool's `semantic_tags` against a site's type/industry. This produced the documented failure mode and was replaced entirely by `tool-suggester`, an LLM-judgment agent that can suggest zero tools.
- **sources:** old/older1/010_tool_library_guide.md#"Deploying automatically via discovery"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Bug history"
- **relations:** tool-suggester agent; mandatory minimum tool-suggestion count (below)
- **verify-later:** confirm `matchToolToSite` function/code path has actually been removed.

### Planned assets-table template/JS split for large tools (superseded plan)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive: for tools >200KB, split JS into a separate `assets`-table file — "This isn't built yet." Live: "A template/JS split IS built — but for the component-creator pipeline... not via the assets table this section once envisioned."
- **what:** The original plan routed oversized tool templates through the `assets` table/S3 pipeline. What was actually built instead — only for component-creator (games/feeds/explorers), not tools — is a `js_content` column on `content_components` populated by `separateInlineJS()`. Live docs warn against applying this to tools without first fixing two known gaps.
- **sources:** old/older1/010_tool_library_guide.md#"When to split template from component"; docs024_key_docs_latest/019_tool_library(2).md#"When to split template from component"
- **relations:** JS Content Separation Contract (003); component-creator pipeline
- **verify-later:** `SELECT count(*) FROM content_components WHERE component_level='tool' AND js_content IS NOT NULL`.

### Mandatory minimum tool-suggestion count (2–5, no "suggest zero" option)
- **category:** tool-lifecycle
- **status-signal:** superseded
- **status-evidence:** Archive: "It returns 2–5 suggestions." Live: "It can return 0-5 suggestions. Returning zero is correct when no tools are appropriate."
- **what:** The earliest `tool-suggester` design forced the LLM to always propose at least two tools per site. Replaced by an explicit zero-is-valid design, directly tied to the same failure class as `matchToolToSite` (irrelevant tools forced onto sites).
- **sources:** old/older1/012_tool_lifecycle_guide.md#"Agent: tool-suggester"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Agent: tool-suggester"
- **relations:** tag-based deterministic tool-to-site matching (above)
- **verify-later:** check tool-suggester's current prompt for the zero-suggestions instruction.

### GPU/AI-endpoint scheduling mechanism selection
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Live `009_model_infrastructure.md` "What's Deployed": `ai_endpoint_health` table + view "Applied, verified"; `endpoint-health-checker` agent + scheduled task "Applied".
- **what:** v1 posed four undecided GPU-scheduling options (priority-deprioritisation, boolean flag, health-check auto-discovery, back-to-triage only). v4 resolved this: a single `ai_endpoint_health` table (active vs reactive check modes) *is* the scheduler — dispatch skips claims against unhealthy endpoints; back-to-triage is the reactive safety net beneath it.
- **sources:** old/older1/020_gpu_and_model_infrastructure.md#"GPU Scheduling: Options Under Discussion"; old/older1/020d_gpu_and_model_infrastructure_v4.md#"Architecture: Three Layers"; docs024_key_docs_latest/009_model_infrastructure.md#"What's Deployed"
- **relations:** back-to-triage error handling (AIUnavailableError); model swap/revert functions
- **verify-later:** `ai_endpoint_health` table contents; `endpoint-health-checker` agent definition.

### Triage drain loop — structured audit findings, capped passes, section locking
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Live `009_model_infrastructure.md` "Done": structured findings, audit pass cap, section locking exclusion all checked off.
- **what:** Fix for unbounded audit/fix/re-audit token spend. Findings must carry `acceptance_test`/`acceptance_levels`/`minimum_required`. Audits capped at 3 numbered batches per site. Passing sections get `locked_at`; subsequent audits skip them; unlock is manual. Per-page sequential processing via `depends_on` prevents overlapping fixes.
- **sources:** old/older1/020d_gpu_and_model_infrastructure_v4.md#"Triage Drain Loop Fix"; docs024_key_docs_latest/009_model_infrastructure.md#"Decisions Made"
- **relations:** three-way audit-finding classification; GPU/AI-endpoint scheduling
- **verify-later:** `write_audit_findings_action.go`; section-lock column on `page_components`.

### Quality improvement flywheel (RAG + LoRA + prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** Live `009_model_infrastructure.md` "Future": RAG actions "registered, not workflow-tested"; LoRA training pipeline and training-data export from `llm_call_log` both still open.
- **what:** Three independently-valuable, compounding improvement channels: RAG (inject retrieved good examples at call time), LoRA (retrain periodically on filtered successful outputs), and deliberate prompt A/B testing (80/20 traffic split, promote on audit-success-rate). A `training-orchestrator` workflow packages LoRA training as an adapter-driven workflow (export → start_gpu_instance → train → evaluate → deploy_or_reject → stop_gpu_instance → log). A scraped-data "AI slop" quality gate filters what may enter the training set.
- **sources:** old/older1/020d_gpu_and_model_infrastructure_v4.md#"Quality Improvement Flywheel", #"Scraped Data Quality Gate (AI Slop Prevention)"; docs024_key_docs_latest/009_model_infrastructure.md#"Future"
- **relations:** GPU/AI-endpoint scheduling; llm_call_log flywheel columns
- **verify-later:** whether any `training_runs` completed beyond the one noted in 009; whether RAG actions are workflow-exercised.

### Debugging playbook (early runbook)
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** Both archive files are early drafts of the operational runbook; the current authoritative version is `016_debugging_guide_v2_58_consolidated.md`.
- **what:** A ten-section operational runbook: pod health check, work-item status queries, scheduled-task flight-status, orchestration-state staleness, agent error log, handler-agent-definition existence checks, timeout ordering chain, a failed-item cleanup transaction, named failure patterns, and a single "quick health dashboard" query. The second draft adds a systematic dispatch-loop `input_mapping` path-mismatch diagnosis, missing-handler-agent detection, and a log-hunting technique.
- **sources:** old/older1/016_debugging_guide.md; old/older1/016_debugging_guide_v2_april26.md
- **relations:** timeout chain ordering; dispatch-loop input_mapping mismatch; wont_fix/needs_section_data patterns
- **verify-later:** whether the consolidated live debugging guide still carries these same queries/patterns.

### Timeout chain ordering contract
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** Stated as a hard ordering requirement in both drafts (claim_timeout > call_handler timeout > workflow timeout), with the call_handler timeout bumped from 900s to 1200s between drafts; not verified against the current consolidated guide.
- **what:** Three timeouts must nest correctly or two failure modes occur: reset-claim double-handling, or dispatch marking an item failed while the handler is still working with nothing listening for its response.
- **sources:** old/older1/016_debugging_guide.md#"7. Timeout Chain"; old/older1/016_debugging_guide_v2_april26.md#"7. Timeout Chain"
- **relations:** debugging playbook
- **verify-later:** current values of `claimed-item-timeout`, `build-dispatch-loop` call_handler timeout, per-handler workflow timeouts.

### Dispatch-loop input_mapping path mismatch (spec-nested vs flat)
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** Documented as "most common systematic failure" with three named affected agents (`tool-improver`, `tool-auditor`, `rerender-pages`); not confirmed whether the flatten-in-dispatch-loop fix or per-handler fix was adopted.
- **what:** `build-dispatch-loop` maps a work item's `spec` JSONB as nested (`input_data.spec.component_id`), but handlers read flat (`input_data.component_id`), producing path-resolution errors. Preferred fix: flatten in the dispatch loop's `input_mapping`, following the existing `page_name?`/`reviewed_brief?` pattern.
- **sources:** old/older1/016_debugging_guide_v2_april26.md#"9. Specific Failure Patterns"
- **relations:** debugging playbook; ExtractActionInputs cross-link
- **verify-later:** current `build-dispatch-loop` `input_mapping` config.

### wont_fix/superseded dedup and needs_section_data data-honesty pattern
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Described as "correct behaviour... the dedup system working" in the second draft.
- **what:** When a recurring issue is detected while an older item is stuck, the loop creates a new item and marks the old one `wont_fix` ("superseded by active duplicate") — expected noise, not a bug. `needs_section_data` items requiring unfabricatable data (bios, pricing, case studies) correctly route to `wont_fix`/HITL rather than inventing content.
- **sources:** old/older1/016_debugging_guide_v2_april26.md#"9. Specific Failure Patterns"
- **relations:** debugging playbook; needs_section_data triage
- **verify-later:** current dedup logic; HITL routing for `needs_section_data`.

### Client schema manual-creation column drift (agent_instances)
- **category:** database-and-infrastructure
- **status-signal:** superseded
- **status-evidence:** Live `011_database_and_infrastructure.md` §"Method 3" explicitly warns "Do not invent column names" and includes a troubleshooting entry for `column "template_id" of relation "agent_instances" does not exist` matching the archive's error-prone DDL.
- **what:** The archive's Method-3 fallback DDL for `agent_instances` used columns that don't match what `spawn_actions.go` actually inserts. Live doc corrects the column list, adds an FK to `projects`, and instructs checking `create_client_schema()`'s source before hand-writing DDL.
- **sources:** old/older1/017_creating_new_client_schemas.md#"Method 3: Manual table creation"; docs024_key_docs_latest/011_database_and_infrastructure.md#"Method 3: Manual table creation"
- **relations:** `create_client_schema()` function; spawn_agent action contract
- **verify-later:** current `agent_instances` schema in a live `client_*` schema.

### Batch-processing control model evolution (two-gate → three-gate, manual → function escalation)
- **category:** batch-processing
- **status-signal:** deployed
- **status-evidence:** Live `015_batch_processing_architecture_v2.md` (dated v4: 2026-04-12) shows the three-gate model and SQL escalation functions in place; archive only had earlier forms (dated 2026-04-06).
- **what:** Two mechanisms evolved: (1) batch on/off control moved from a two-gate check to a three-gate resolution (global → `llm_batch_agent_config` per-agent-type opt-in with `batch_group` → provider); (2) urgent-item escalation moved from raw UPDATE statements to dedicated SQL functions `escalate_batch_item(id)`/`escalate_site_batch(site_id)`. A new `sync_executed` status was added for auditable batch-off proving runs.
- **sources:** old/older1/015_batch_processing_architecture.md#"The Table", #"3. Priority Override"; docs024_key_docs_latest/015_batch_processing_architecture_v2.md#"Per-Agent-Type Control Table"
- **relations:** llm_call_log flywheel columns; QueueLLMBatchAction three-gate check
- **verify-later:** `llm_batch_agent_config` table contents; escalation function definitions.

### Early "visual identity poles" layout taxonomy (dropped)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Diff-confirmed only in the earliest of four palette/layout/typography-migration drafts; final 15-layout table uses different (hyphenated) names though keeps several "pole" nicknames.
- **what:** The very first migration draft described layout diversity as nine named "poles" tied to specific sites (Brochure/corporate, Magazine/editorial, Portfolio/kinetic "vonc", Commerce/grid, Utility/tool "thunder compute", Media/streaming "youtube", Documentation/reference, High-energy/bold "boxing", Soft/editorial). Dropped in favour of vaguer prose, then crystallised differently as the final 15-layout table (adding six layouts absent from the original nine-pole list).
- **sources:** old/older1/025_palette_layout_typography_migration.md#"2. Scope Decisions"; docs024_key_docs_latest/025_palette_layout_typography_migration(3).md#"7. The Layouts to Build"
- **relations:** composable theme migration; site-design-planner
- **verify-later:** final `layouts` table row count/names vs. the 15-layout plan.

### Phased belt-and-braces removal plan for webdesign-agent install_theme (abandoned same-day)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** v1 doc's changelog (2026-04-19) states the belt-and-braces step "remains... pending the two-phase removal plan"; live v2 doc's changelog the same day: "Merge applied" (direct removal).
- **what:** `026_design_and_site_planner_v1.md` proposed a cautious two-phase removal of webdesign-agent's defensive `install_theme`/`check_should_install` steps (diagnostic no-op first, delete only after a week of zero firings). Abandoned within hours: live v2 shows the two steps deleted outright the same day, routing rewired directly to `generate_css`, relying instead on the renderer's emergency-fallback logging as the sole safety net.
- **sources:** old/older1/026_design_and_site_planner_v1.md#"6. Removing install_theme From Webdesign-Agent (Planned)"; docs024_key_docs_latest/027_design_and_site_planner_v2.md#"6... (Applied)", #"12. Change Log"
- **relations:** site-design-planner composition-install path; renderer emergency-fallback logging
- **verify-later:** confirm `install_theme`/`check_should_install` are absent from webdesign-agent's agent_definitions.

### Early pipeline-failure triage priorities dropped by root-cause diagnosis
- **category:** debugging
- **status-signal:** abandoned
- **status-evidence:** The 2026-04-14 report's P3 (vonc.com raw CSS), P4 (stale-item process gap), P5 (timeout tuning) don't appear in the 2026-04-15 v3 report's P1-P10 list at all.
- **what:** First-pass triage of 57 stuck work items framed three priorities at the symptom level. Within a day, deeper diagnosis replaced these with concretely-fixed root causes not originally identified: rate-limit errors misclassified as non-transient (1,869 occurrences), `load_page_record` lacking a `page_id` fallback, and later audit-finding routing/classification bugs.
- **sources:** old/older1/105_dispatch-pipeline-failures-report.md#"Priority Fixes"; old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes"
- **relations:** plan_sections pre-check evolution; three-way audit-finding classification
- **verify-later:** current state of vonc.com's about page (raw-CSS-serving bug).

### plan_sections pre-check → plan-then-reconcile evolution
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** v3 report P5 "DEPLOYED" (pre-check); live v4 shows the same P5 row "UPDATED" with a materially different mechanism.
- **what:** The original fix for wasteful LLM re-sends on sections with pending `needs_section_data` was a pre-check that simply skipped them. Revised to "plan-then-reconcile": ready sections auto-close stale data requests, deferred sections create new requests while skipping duplicates.
- **sources:** old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes (P5)"; docs024_key_docs_latest/105_dispatch-pipeline-failures-report_v4.md#"Priority Fixes (P5)"
- **relations:** early pipeline-failure triage priorities; needs_section_data triage
- **verify-later:** current `plan_sections_action.go` logic for auto-closing `needs_section_data` items.

### Three-way audit-finding classification (bug / recommendation / gap)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** v3 report's P10 marked "DEFERRED... ~1 week project"; referenced independently, still not built, in `027_design_and_site_planner_v2.md` months later.
- **what:** Auditors currently produce findings the pipeline auto-fixes uniformly as if bugs, but many are opinions/recommendations — producing false-positive fix attempts. Proposed fix: three-way classification with dedicated specialist agents per category and per-site approval mode for recommendations.
- **sources:** old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes (P10)"; docs024_key_docs_latest/027_design_and_site_planner_v2.md#"10. Open Design Areas"
- **relations:** audit gap-finding routing fix; triage drain loop
- **verify-later:** existence/status of `design-note-recommendation-specialists.md` or any implementing specialist agent.

### Audit gap-finding routing fix (existing-page gaps → needs_content_page)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** v3 report P9: "FIX WRITTEN — write_audit_findings_action.go: Rule 4 routes gap findings on existing pages to needs_content_page, not content_rewrite."
- **what:** Gap findings on existing pages were being routed to `content_rewrite` (edits, not rebuilds), causing validation-failed rewrites. Rule 4 redirects them to `needs_content_page` (full rebuild path).
- **sources:** old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes (P9)"
- **relations:** three-way audit-finding classification; needs_content_page work-item type
- **verify-later:** current `write_audit_findings_action.go` Rule 4 logic.

### Section recipes for adoption (purpose + structure + reference implementation)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4: Requirement-Driven Components (longer term)" in the 2026-04-11 plan; no confirmation of shipping in any later doc reviewed.
- **what:** When adopting a site, each section would be captured as a "recipe": purpose, structure, reference implementation (guide not spec), and component match. Recipes without a good match would generate `needs_new_component` work items where the recipe becomes the build brief.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Decisions Made & Rationale (4)", #"Phase 4"
- **relations:** component selector by functional requirement; needs_new_component work items
- **verify-later:** whether any adoption workflow step produces structured "recipes" today.

### Visual identity library and effects library (composable design assets)
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)" in the 2026-04-11 plan; not confirmed built.
- **what:** Longer-term plan for two accumulating libraries: a visual identity library of palettes/typography/effects searchable by purpose/audience, and an effects library treating elevation/corner radius/animation/density as composable modifiers independent of layout. Likely precursor idea to the `palettes`/`typography_sets`/`layouts` table split actually implemented.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4: Requirement-Driven Components (longer term)"
- **relations:** composable theme migration; component selector by functional requirement
- **verify-later:** whether structure_tokens/effects concepts in the live composable-theme schema fulfil this idea.

### Component selector by functional requirement
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)"; no later confirmation found.
- **what:** Proposed capability-based search over `content_components` — finding a component by what it does rather than by name/category — paired with section recipes.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4"
- **relations:** section recipes for adoption; tool-suggester's LLM-judgment matching
- **verify-later:** any capability/tag-based component search implementation in `component_library.go`.

### Three-layer design system (content_components / css_themes / style_collections)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Live: "Today css_themes.css_template mixes three distinct concerns into one row: palette, typography, and layout" — the migration splits this monolith.
- **what:** Early architecture with three independently-varying layers: HTML components, a monolithic CSS theme (one row = whole stylesheet), and a `style_collections` bridge table. Superseded internally when `css_themes` was split into three composable entities; `style_collections` survives as the outer bundle.
- **sources:** old_design_and_styling/FOCUS_design_and_styling.md#"1. The Design System: Three Independent Layers"; docs024_key_docs_latest/025_palette_layout_typography_migration(3).md#"Splitting css_themes"
- **relations:** composable theme system; style_collections bundle
- **verify-later:** confirm `css_themes` legacy columns actually dropped (Phase 7).

### Design fingerprint extraction pipeline
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Not started" → "✅ Deployed, works" (2026-04-14) → "Victory: Design Fingerprint Now Correct" (2026-04-16).
- **what:** Pipeline step parsing a crawled site's CSS into a colour/font/layout "fingerprint" so adoption rebuilds match the original. Went from unstarted idea to working end-to-end (gamedesign.uk) across several debugging sessions.
- **sources:** old_design_and_styling/FOCUS_design_and_styling.md#"4. The Adoption Design Gap"; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Victory"
- **relations:** design_reference/design_intent split; computed styles extraction; fpExtractCSSVars fix
- **verify-later:** `site_specs` rows with aspect='design_reference' for adopted sites.

### Webdesign-agent three-way design priority
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "✅ Applied" (2026-04-14 handoff).
- **what:** `analyze_design` step branches on which specs exist: design_intent present → creative freedom around described character; only design_reference → faithful reproduction, no invented palette; neither → generate from industry/audience/identity.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-14_v2.md#"Webdesign-Agent Prompt (deployed)"
- **relations:** design_reference/design_intent spec-aspect split
- **verify-later:** current webdesign-agent agent_definitions prompt text.

### Palette-locked-until-design_intent policy
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Palette is locked until design_intent exists."
- **what:** First adoption build reproduces the original palette exactly (locked); once design_intent is written, webdesign-agent gains creative freedom within the described character, letting the improvement loop evolve the palette over time.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md#"Key Decisions Made"
- **relations:** design_reference/design_intent split; audit loop "propose not apply"
- **verify-later:** improvement-loop audit code for propose-vs-enforce mode switch.

### Whole-blob input_data passthrough mapping (anti-pattern)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Archive presents `{"input_data": "input_data"}` as a working pattern used by three orchestrators; live: "It does not do what it looks like it does... map each expected field by name."
- **what:** A wrapper-orchestrator shorthand documented as valid in the archive. Live guide identifies it as broken (double-nests the caller's data) and replaces it with explicit per-field mapping using `?`-suffixed optional keys.
- **sources:** old/001_development_guide.md#"Standardized Input Extraction"; docs024_key_docs_latest/001_development_guide(5).md#"Map fields individually, not the whole input_data blob"
- **relations:** ExtractActionInputs nested-source collision; input_mapping `?` suffix convention
- **verify-later:** grep current agent_definitions for `"input_data": "input_data"` mapping still in use.

### agent_definitions backup naming convention (unversioned → _preNNN)
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** Live adds "Naming convention: agent_definitions_backup_YYYYMMDD_pre<NNN>... DO NOT use DROP TABLE IF EXISTS before CREATE TABLE."
- **what:** Archive's convention was a plain `agent_definitions_backup_YYYYMMDD` name with no migration tie and no never-drop rule. Live requires a `_pre<NNN>` suffix tying the backup to the migration it guards and forbids dropping/overwriting an existing backup.
- **sources:** old/009_model_infrastructure.md#"Migration Safety"; docs024_key_docs_latest/009_model_infrastructure.md#"Migration Safety"
- **relations:** model swap/rollback procedure
- **verify-later:** recent migration backup table names for `_preNNN` adoption.

### site_plan page-role enum naming (underscore → hyphen; index → landing)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** Archive: `"section_index" | ... | "blog_post"`; live: `"section-index" | ... | "blog-post" | "landing"`.
- **what:** `site_plan_pages.role` vocabulary was originally underscore-separated with a bare `index` role for the homepage. Renamed to hyphenated form and the homepage role renamed to `landing`, matching kebab-case conventions elsewhere.
- **sources:** old/029_site_plan_and_reconciler.md#"role table"; docs024_key_docs_latest/029_site_plan_and_reconciler(2).md#"role table"
- **relations:** page_type vocabulary and kebab constraint (016 §6.5)
- **verify-later:** DB check constraint on `site_plan_pages.role`/`pages.page_type` for hyphenated values.

### site_plan_partials — single JSONB-blob partial storage (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Live: "JSONB blobs were considered and rejected because at anticipated scale... loading whole blobs to read one slice is wasteful, surgical HITL edits become hard, and lock transfer at meaningful granularity is impossible."
- **what:** Archived Phase 1 plan proposed one table, `site_plan_partials`, storing each partial as a single versioned JSONB blob per plan. Abandoned for two normalized row-per-thing tables — `site_plan_sections` and `site_plan_directives` — enabling per-row HITL locking at 1000+ page scale.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"schema section"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"schema section"
- **relations:** lock transfer across plan rebuilds; lazy per-page brief generation (also abandoned)
- **verify-later:** confirm `site_plan_directives`/`site_plan_sections` tables exist, `site_plan_partials` does not.

### Three sequential per-partial plan-builder LLM calls (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Live: "Earlier draft of this doc proposed three sequential LLM calls. Looking at the existing build-site-planner agent, that lean was wrong."
- **what:** Archived plan proposed splitting the plan-builder into three sequential LLM calls for independent retry granularity. Abandoned once it was noticed the production build-site-planner agent already produces all three coherently in one call with no evidence of retry-granularity problems.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"Q2. Plan-builder LLM tier"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Q2. Plan-builder LLM call shape"
- **relations:** site_plan_partials (abandoned)
- **verify-later:** build-site-planner agent_definitions workflow — confirm single LLM call shape.

### Separate BuildPageURL path-resolver helper (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Live: "The earlier draft of this doc proposed a separate BuildPageURL helper... That argument was overly cautious... Consolidated."
- **what:** Archived plan proposed a brand-new ~50-line Go helper sibling to `page_canonical.go`. Abandoned as overly cautious: Phase 1 instead extends `CanonicalisePage` additively with an optional `ParentSection` field.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"Q3. URL paths"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Q3. URL paths"
- **relations:** site_plan page-role enum naming
- **verify-later:** `datahelpers/page_canonical.go` — confirm `CanonicalisePage` has `ParentSection`, no separate `BuildPageURL`.

### Lazy per-page brief generation via build_page_brief step (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Archive rollout step 8: "build_page_brief step in page-build-handler... generates site_plan_partials/page_brief:<name> if missing." Live replaces with a pure-Go brief renderer.
- **what:** Archived plan generated each page's brief lazily via an LLM step during page build. Abandoned for a deterministic, non-LLM Go helper that assembles a brief at read time by walking the directive cascade and applying cardinality rules.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"rollout table, step 7-8"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Directive cascade and brief assembly"
- **relations:** site_plan_partials (abandoned)
- **verify-later:** confirm `datahelpers/page_brief.go` exists; page-build-handler has no `build_page_brief` LLM step.

### Legacy monolithic CSS renderer internals (removed)
- **category:** design-composition
- **status-signal:** abandoned
- **status-evidence:** "Phase 4.3 already removed... cssTemplateData struct (and its 16 hardcoded fields)... Compile-clean."
- **what:** The original renderer held a flat struct populated by `extractDesignColors`/`designColorMaps`, loading one Go template per theme. Deleted wholesale in Phase 4.3 when the renderer switched to composable palette/layout/typography_set rows via FK.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"What Phase 4.3 already removed"; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Template variable system"
- **relations:** three-layer design system (superseded); layout archetype library
- **verify-later:** grep codebase for `loadCSSGoTemplate`/`extractDesignColors`/`designColorMaps` — should be absent.

### fpExtractCSSVars regex-based CSS variable extraction (superseded)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** BEM selectors like `.btn--primary:hover` captured as fake variables; replacement uses `:root` block targeting with semicolon-splitting.
- **what:** Original extractor used one whole-stylesheet regex, producing false positives on BEM class names. Replaced with a multi-strategy extractor isolating `:root`/body/`[data-theme]` blocks, with fallback frequency analysis for utility-CSS sites.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"P6 — fpExtractCSSVars BEM False Positives"; old_design_and_styling/FOCUS_design_and_styling_fp_extract_css_vars_integration.md
- **relations:** design fingerprint extraction pipeline; computed styles extraction
- **verify-later:** `extract_design_fingerprint_action.go` — confirm regex-based extractor removed.

### css_templating.go theme-forking bridge (known-broken, scheduled rewrite)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "fork_theme_from_site produces rows with NULL palette_id, NULL layout_id, NULL typography_set_id... Adoption-forked themes are unusable by the render path."
- **what:** `TemplateCSSFromSpec` converts a rendered CSS snapshot into old flat-field-name placeholders and writes it to the legacy `css_themes.css_template` column, which the post-Phase-4.3 renderer never reads — silently producing unusable NULL-FK theme rows. Flagged for a Phase 5 rewrite.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"1. css_templating.go"; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Ready for deployment"
- **relations:** fork_theme_from_site rewrite (Phase 5); parallel legacy HTML-assembly render path
- **verify-later:** confirm `fork_theme_from_site_action.go` now produces palette/typography_set rows.

### Parallel legacy HTML-assembly render path (getThemeByID/GetThemeByName)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "css_content is populated for 13 of the 14 themes. standard-brochure has empty css_content... falls through to GetThemeByName('default')."
- **what:** A second, older render path reads `css_themes.css_content` directly into assembled HTML, independent of the spec-driven render path. Left untouched by Phase 4, own known gap flagged for resolution when Phase 7 drops legacy columns.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"2. getThemeByID / GetThemeByName"
- **relations:** css_templating.go bridge; legacy css_themes columns drop (Phase 7)
- **verify-later:** grep for `getThemeByID`/`GetThemeByName` call sites.

### Component-creation via HITL work-item triage (superseded)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** "migration_025_component_triage.sql — an earlier work-item-based approach that was superseded by the direct insert approach... Do not run this file."
- **what:** Earlier plan for seeding new library components via work items routed through HITL triage. Superseded by a direct SQL insert once components were designed and reviewed.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"2. What's Been Completed"
- **relations:** layout archetype library
- **verify-later:** none — historical, file explicitly marked do-not-run.

### Computed-styles extraction via browser JS injection
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Computed styles (Phase 2) deferred... Spec written but not implemented" vs. a complete Go action + workflow SQL in the Phase 2 doc.
- **what:** Supplementary fingerprint step scraping a homepage with injected JS calling `getComputedStyle()`, writing resolved values for a Go action to parse and merge — "ground truth" overriding source-CSS guesses. Fully spec'd but recorded elsewhere as deferred/not implemented.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_computed_styles_extraction_phase2.md; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Fixes Ready But Not Deployed"
- **relations:** design fingerprint extraction pipeline; fpExtractCSSVars fix
- **verify-later:** registry.go for `extract_computed_styles`; site-adoption-agent workflow steps.

### Layout archetype library (15 named layouts)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Phase 1 is next: designing and writing the 15 layout CSS templates" → "Phase 1 — Layouts seeded (15 rows in layouts table)... deployed."
- **what:** Taxonomy of 15 named structural/visual archetypes (brochure-formal, portfolio-kinetic, utility-tool, media-grid, etc.), each with character/structural-trait descriptions, default header/footer/typography, and legacy-theme mappings — the target library for the composable-theme migration's `layouts` table.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"4. The 15 Layouts — Detailed Descriptions"; docs024_key_docs_latest/025_palette_layout_typography_migration(3).md
- **relations:** composable theme system; site-design-planner layout resolver
- **verify-later:** `layouts` table rows in DB — confirm 15 rows.

### Palette merge rule (core slots vs specialised slots)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Core slots (spec wins where present)... Specialised slots (theme wins)."
- **what:** When a site composes a theme, core palette slots let the site's own spec win when present; specialised slots (primary_hover, hero_title, cta_bg, etc.) always take the theme's value.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Palette merge rule"
- **relations:** layout archetype library; site-design-planner palette resolver
- **verify-later:** `resolve_composition_palette_action.go` merge logic.

### site-design-planner "Choice B" scope (composition-only)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Choice B adopted. The agent's exclusive responsibility is composition resolution... It does NOT write navigation or layout specs."
- **what:** Decision narrowing site-design-planner to write exactly one spec, `resolved_composition` (palette_id/layout_id/typography_set_id + lineage + reasoning), deferring `navigation`/`layout` spec ownership to future specialist agents — justified by "slim strict responsibilities."
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"3. Scope Refinement"
- **relations:** composition resolution architecture; design pipeline guiding principles
- **verify-later:** agent_definitions row for site-design-planner — confirm workflow only writes `resolved_composition`.

### Composition resolution architecture (3 resolvers + install action)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "validate_composition_inputs_action.go — DONE... install_site_composition_action.go — DONE (~562 lines)."
- **what:** site-design-planner pipeline: `validate_composition_inputs` → three resolvers (`resolve_composition_layout` tag-overlap match, `resolve_composition_typography`, `resolve_composition_palette` fingerprint→mission→design_intent→layout-inherit→default cascade) → `install_site_composition` (one transaction: css_themes+style_collections insert, sites update, resolved_composition spec write).
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"4. Work Plan — Deliverable 4"
- **relations:** site-design-planner Choice B scope; composition resolver orphan-rows policy; fork_theme_from_site
- **verify-later:** confirm `resolve_composition_*.go`/`install_site_composition_action.go` in current codebase.

### webdesign-agent install/render ordering bug ("first render wrong layout")
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "9.1 Known ordering issue in webdesign-agent... This is the exact 'first render wrong layout' bug site-design-planner was built to eliminate."
- **what:** webdesign-agent ran `generate_css → deploy_css → ... → install_theme`, so any site without a pre-installed composition hit the emergency fallback and committed it to git before the correct composition was installed a step later. Documented, deferred fix: reorder `install_theme` before `generate_css`.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"9.1 Known ordering issue"
- **relations:** composition resolution architecture; render_css_from_spec_action emergency fallback
- **verify-later:** webdesign-agent workflow step order.

### Fork_theme step double-creation guard
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** "Guard: require both should_fork_theme AND should_promote_to_library flags. Implementation deferred to Deliverable 6."
- **what:** Once site-design-planner runs, the pre-existing `fork_theme` step in webdesign-agent risks creating duplicate theme/collection rows. Documented mitigation requires two flags both true before forking proceeds.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"6. Risks Still Live"
- **relations:** composition resolution architecture; fork_theme_from_site rewrite
- **verify-later:** `fork_theme` step config in webdesign-agent — confirm both flags gate execution.

### Adopt-from vs deploy-to separation (unbuilt)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** "discussed but not implemented. Options: snapshot to S3, stage to subdomain, or store crawl artifacts."
- **what:** Unbuilt idea for a staging area distinct from the live deploy target, so a freshly-adopted rebuild could be reviewed before overwriting production. Workaround at time of writing was manual: pause work items, verify specs, unpause.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Architecture Decisions Made (item 6)"
- **relations:** site snapshots and revert (014); design fingerprint extraction pipeline
- **verify-later:** whether any staging/subdomain mechanism exists for adoption today.

### Design pipeline guiding principles (mottos)
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "Principles Restated" section repeated verbatim across 2026-04-19 handoffs, sourced from `007_adoption_pipeline_v2.md` and a FOCUS work-plan doc.
- **what:** A shared decision-shorthand invoked to settle scope questions: "Every build conceptually an adoption," "Design reference is history, design intent is direction," "Adoption is a starting point, not a ceiling," "LLM for reasoning, Go for extraction," "Handlers are self-contained," "Slim strict responsibilities."
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"7. Principles Restated"
- **relations:** site-design-planner Choice B scope; design_reference/design_intent split
- **verify-later:** none — a documentation/culture artifact, not directly code-verifiable.

### Adapter deployment troubleshooting (ImagePullBackOff / command-vs-args / Kafka topic provisioning)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** Appears only as "## 10. Adapter & Service Deployment Issues" in `debugging_old/016_debugging_guide_v2(1).md`; absent from every subsequent snapshot and from the live consolidated debugging guide (zero grep hits); the content lives instead in `035_adapter_guide.md`.
- **what:** Covers real thunder-adapter-era deployment failures: diagnosing Docker Hub `ImagePullBackOff`/`insufficient_scope`, the Kubernetes `command:` vs `args:` trap (args silently replaces the entire Dockerfile CMD), the Strimzi `auto.create.topics.enable=false` gotcha requiring an explicit KafkaTopic CRD, and a "deployment essentials checklist" required for every new adapter.
- **sources:** debugging_old/016_debugging_guide_v2(1).md#"10"; docs024_key_docs_latest/035_adapter_guide.md#"2.12-2.13"
- **relations:** adapters (035_adapter_guide.md), deployment-github, single-source relocation convention (below)
- **verify-later:** confirm `035_adapter_guide.md` §2.12/§2.13 still matches the checklist.

### Adapter Response Envelope Contract relocated from 003 to the adapter guide
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** `debugging_old/003_contracts_and_standards_v10/v11.md` contain the full section; live `003_contracts_and_standards(8).md` replaces it with one line: "Moved to 035_adapter_guide.md §1... now the single source for it."
- **what:** Defines how a long-lived adapter must shape its Kafka reply so the chassis recognises it as an awaited response: reuse the incoming `request_id`, fresh `message_id`, `ProduceWithValidation` (never plain `Produce`), and a typed Go struct for response `headers` (not `map[string]string`) so `is_complete`/`is_error` marshal as real JSON booleans. Motivated by a real production incident (thunder-adapter matcher failure, 2026-05-22).
- **sources:** debugging_old/003_contracts_and_standards_v11.md#"Adapter Response Envelope Contract"; docs024_key_docs_latest/035_adapter_guide.md#"1"
- **relations:** adapters, tool-pipeline, single-source relocation convention (below)
- **verify-later:** check adapter Go source for typed `ResponseHeaders` struct vs any remaining `map[string]string` header builders.

### Single-source relocation with pointer (doc consolidation convention)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Two independent content areas (adapter deployment mechanics, adapter response envelope shape) were both removed from their original host docs and consolidated into `035_adapter_guide.md`, with the live contracts doc leaving a one-line "Moved to X, now the single source for it" pointer.
- **what:** A recurring documentation practice: when a topic is found duplicated across a debugging guide and a contracts doc, maintainers consolidate it into one canonical doc and replace the other locations with a short pointer sentence, rather than letting copies drift out of sync.
- **sources:** docs024_key_docs_latest/003_contracts_and_standards(8).md; docs024_key_docs_latest/035_adapter_guide.md; debugging_old/016_debugging_guide_v2(1).md
- **relations:** adapters, documentation-system, 000_documentation_index
- **verify-later:** check `000_documentation_index.md`/travelling_docs conventions for whether this is a formal rule.

### CrashLoop `exec: "./X"` image/binary-content mismatch diagnosis
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Live v2_58 header: "Recovered two §9 entries that were present in the earliest file v2_47(1) but had been dropped from the v2_48-onward branch."
- **what:** A three-command image-inspection technique (`docker run --entrypoint ls`, `docker inspect .Config.Entrypoint`, `.RepoDigests`) for diagnosing `CrashLoopBackOff` with `exec: "./X": no such file or directory` — proves the running image lacks the named binary (wrong build context / tag-sharing), not a config problem.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 CrashLoop exec"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** temporarily abandoned in the v2_48→v2_57 main branch (this fork diverged at v2_45), recovered wholesale into live v2_58
- **verify-later:** whether a CI guard ("fail build if binary absent") was ever implemented for thunder-adapter/analyser-adapter Dockerfiles.

### Hand-applied agent/launcher-def migrations are not commutative
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** The other v2_47(1)-only recovered entry; incident resolved by re-applying 110 then 111, RUNBOOK "2d state check" added as a live procedural safeguard.
- **what:** Re-applying migration 109 (per a runbook's "safe to re-run" claim) silently reverted later migrations 110/111 because 109 rebuilt DB-object nodes that 110 had replaced. A migration is idempotent only against its own prior application, never against later migrations touching the same path.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 Re-running an idempotent migration"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** NEW:migration-governance proposal (below)
- **verify-later:** confirm the `training-launcher` agent_definitions row currently reflects migrations 109-111 in correct order.

### NEW:migration-governance — proposed migration runner/ledger for hand-applied agent-def changes (never built)
- **category:** NEW:migration-governance
- **status-signal:** aspirational
- **status-evidence:** "If this ever graduates beyond hand-application, a migration runner (or even a tiny applied_migrations log table) would enforce order and make re-applying an earlier one structurally impossible" — explicitly proposed, never implemented, in any version of the family or the live doc.
- **what:** An idea for formalizing ad-hoc `jsonb_set` migrations applied by hand to `agent_definitions`/launcher defs: a lightweight ledger table or runner tracking which numbered migrations had been applied, preventing accidental reversion.
- **sources:** archive_april_26/016_debugging_guide_v2_47(1).md#"§9 Re-running an idempotent migration"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** procedural stand-in currently in place is the RUNBOOK "2d state check" (manual, not automated)
- **verify-later:** grep codebase/DB for any `applied_migrations`/`schema_migrations`-style table scoped to `agent_definitions` — none expected to exist.

### Full heading+content-line diff across all forked copies before consolidating a travelling doc
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** Used explicitly twice: v2_58 ("A full heading-level AND content-line diff across all 14 files confirmed these were the ONLY entries missing") and the 016b consolidation ("Verified against ALL forked copies... a full heading-level AND content-line diff proved this copy already contains every one of the 9 distinct §9 entries").
- **what:** A consolidation methodology: before promoting one copy of a travelling/forked doc to canonical, diff it against every other known fork at both heading and content-line granularity, explicitly asserting "no content was removed," and recover anything found missing.
- **sources:** docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md; docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md
- **relations:** the method's completeness claim does not always hold in practice — see the diagnosis-loop fork below, which the 016b audit's own "verified against ALL forks" claim did not actually catch
- **verify-later:** none code-related — a documentation-process note for docs026 itself.

### gamesdesign `index` silent-staleness investigation — superseded hypothesis chain
- **category:** debugging
- **status-signal:** superseded
- **status-evidence:** Three successive, explicitly-labelled hypotheses in the same changelog: "silent-completion from a pod dying mid-flight" → "NOT a timeout/deploy issue at all... content-regression guard errors masked as success" → "SUPERSEDED-PENDING-CONFIRMATION" opening a metadata-path-mismatch thread.
- **what:** A multi-week live diagnosis of why gamesdesign.co.uk's `index` page stayed stale despite repeatedly "completing" rebuilds. Each hypothesis explicitly superseded the previous as new evidence arrived. Eventually-confirmed root cause is a more general mechanism — "Child workflow result silently replaced by a stub" (`output_field` vs `output_fields`), shipped 2026-06-18.
- **sources:** archive_april_26/016_debugging_guide_v2_49.md, v2_49(1).md, v2_49(2).md#"§9"; docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md
- **relations:** own recursive application of "don't trust a complete status" heuristic
- **verify-later:** confirm `platform/orchestration/result_spec.go` (`resolveResultSpec` fix) is present in the current codebase.

### `error_step`: config-level placement requirement + derive-from-next_step fix pattern
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "the routing FIRES (error_routed in ProcessingHistory)... Live-validated ×5 in one run."
- **what:** The chassis workflow coordinator only consults `step.Config["error_step"]` (config-level); a step-level `error_step` is parsed but never read, so placing it outside `config` is silently inert. Fix pattern: derive `error_step` from the step's own `next_step`. This entry and its three siblings below are genuinely absent from the canonical live `016b_debugging_guide_8_consolidated.md`/`merged(1).md` — they continue only in a parallel `travelling_docs/016b_debugging_guide_7_3_(2..7).md` fork the canonical consolidation's "verified against ALL forks" claim did not actually reconcile.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"error_step: config-level placement..."
- **relations:** dormant instances of the buggy shape found still live in `tool-recreation-handler` and `tool-auditor` agent definitions
- **verify-later:** grep `agent_definitions` for step-level `error_step` occurrences in those two agents.

### Anchorless (code-only) diagnosis dies at load_runtime
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** "Fix. Config-level error_step on load_runtime... Live-validated ×5" (deployed) but "Pending softening (next chassis build)" (aspirational remainder).
- **what:** A diagnosis run with no anchor was treated as optional by bundle-assembly but hard-errored the whole child workflow at `load_runtime`. Interim fix routes the error back to its own `next_step`; a proper code-level softening (treat no-anchor as a skip) was identified but not yet shipped.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"Anchorless (code-only) diagnosis..."
- **relations:** sibling of the error_step concept above; also absent from canonical live 016b
- **verify-later:** check `diagnose_load_runtime` action source for the `skipped:true` softening.

### Pod label key is `agent-type` (hyphen) vs log field `agent_type` (underscore)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Stated as a settled operational rule with a named failure mode already observed.
- **what:** Kubernetes pod labels use `agent-type` while structured log JSON fields use `agent_type`; using the underscore form in a `kubectl logs -l` selector silently matches zero pods. Separately, a correct selector spans ALL live pods of that type, so a tail can mix in a previous run's failure dump.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"Pod label key is agent-type..."
- **relations:** older trigger scripts (082/083c) still carry the underscore form; absent from canonical live 016b
- **verify-later:** grep trigger scripts 082/083c for the underscore `agent_type=` selector.

### Two failure envelopes — a COMPLETED parent orchestration does not mean the child succeeded
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Presented as a settled mechanism with two named, confirmed code paths (`sendWorkflowResponse` vs `notifyParentOfFailure`).
- **what:** A mid-run step failure is reported via `sendWorkflowResponse` with header `status:"complete"` but the real failure in the body, which the parent forwards and then itself shows COMPLETED with a non-empty `error` column; a START-time failure instead uses `notifyParentOfFailure` with `status:"error_unrecoverable"`. Consumers must check the body, never the header status alone.
- **sources:** archive_april_26/016b_debugging_guide_7(4).md#"Two failure envelopes"
- **relations:** same "trust the artefact, not the status" family as the guide's core silent-completion heuristics; absent from canonical live 016b
- **verify-later:** read the current `sendWorkflowResponse`/`notifyParentOfFailure` implementations to confirm the two-envelope shape.

### Site-work-orchestrator dispatch loop — asset self-resolving storage URI
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Archive documents specific Go additions (`PresignedURLToS3URI`, `resolveStorageURIFromAsset`) and a full finetuning.uk worked dispatch trace; none of this Go-function-level detail or the worked example appears in live `002_system_architecture(4).md`'s "Dispatch Loop... (from 004_site_work_orchestrator)" section, which keeps only the abstract principles.
- **what:** When the dispatch loop's discovery-written work items carry presigned HTTPS asset URLs but `deploy_image_asset` needs `s3://` URIs, the fix was to have `asset-deployer` resolve its own storage URI from `asset_id` via a DB lookup rather than have the orchestrator pre-resolve it — keeping handler self-containment.
- **sources:** archive_april_26/004_site_work_orchestrator.md (whole file); docs024_key_docs_latest/002_system_architecture(4).md#"Dispatch Loop: Dynamic Work Item Routing"
- **relations:** dispatch-pattern spawn→call; asset-deployer agent; handler self-containment principle
- **verify-later:** `grep -n "PresignedURLToS3URI\|resolveStorageURIFromAsset" platform/` to confirm these functions still exist.

### Self-spawning flat dispatch-loop (pre-scheduler design, superseded)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Archive (dated 2026-02-24) states as a "Key Design Decision": "No sub_workflows — they've been problematic," "No loops in dispatch — one item per invocation, self-spawns for next item"; the eventual system uses a genuine `"action":"loop"` construct driven by a scheduled 30s/120s kafka-scheduler tick.
- **what:** An early design decision to avoid the framework's loop/sub_workflow mechanism entirely, having `build-dispatch-loop` process exactly one item then spawn a fresh copy of itself. Later abandoned in favour of the scheduler-driven periodic trigger combined with the fully-developed in-workflow loop mechanism.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Key Design Decisions"; docs024_key_docs_latest/010_scheduler_and_tasks.md#"build-pipeline-trigger"; docs024_key_docs_latest/001_development_guide(5).md#"Appendix C — Loop Mechanisms"
- **relations:** loop-mechanisms (dev-guide appendix), scheduler-and-tasks, build-dispatch-loop agent
- **verify-later:** confirm `build-dispatch-loop`'s current agent_definition workflow uses the loop action, not self-spawning.

### claim_work_item atomic claim action + load_work_items first_item patch
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Archive lists these as "created, not yet committed" (Feb 2026); live loop-mechanisms appendix shows `claim_work_item` as a fully standard, already-existing action used in production.
- **what:** `claim_work_item` performs an atomic `UPDATE ... WHERE status IN ('triaged','approved') RETURNING id` so concurrent dispatch loops can't double-process the same item. The companion `load_work_items` patch added a `first_item` convenience field since the framework's path resolver doesn't support array indexing.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Completed Artifacts"; docs024_key_docs_latest/001_development_guide(5).md#"Appendix C"
- **relations:** dispatch-loop pattern, loop mechanisms, scheduler-and-tasks
- **verify-later:** none needed — graduated cleanly from draft to shipped mechanism.

### Loop array-iteration internals (early investigation notes)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Archive's "Open Investigation" records partial early findings about `LoopAction`/`handleLoopExpansion`/`setLoopVariable`; the fully worked-out version is documented completely in `archive_april_26/014_loop_mechanisms_guide.md`, itself fully absorbed into the live dev guide as Appendix C.
- **what:** Records still-open Feb-2026 questions about how loop expansion and substep naming work internally, later fully resolved and documented.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Open Investigation"; archive_april_26/014_loop_mechanisms_guide.md
- **relations:** claim_work_item / self-spawning dispatch loop (above)
- **verify-later:** none — superseded by a complete, later document.

### CTE-only scheduled tasks pattern ("Always Return a Row" rule)
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive `011b_scheduler_and_tasks_guide.md` (a later revision than `011_kafka_scheduler_guide.md`, which is byte-identical to live) has a full section on this; none of it appears in live `010_scheduler_and_tasks.md`.
- **what:** Some scheduled tasks do their real work directly inside the pre_query's CTEs rather than triggering an agent — but the scheduler still requires the SELECT to return at least one row, or `last_triggered_at`/`last_completed_at` never advance, silently breaking firing cadence and concurrency-group accounting. This is a documented, previously-hit production bug pattern completely absent from the current live scheduler doc.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"Pre-Queries", #"The fire_message Column"; docs024_key_docs_latest/010_scheduler_and_tasks.md (confirmed absent)
- **relations:** concurrency-group starvation; last_completed_at ownership
- **verify-later:** `SELECT name, pre_query FROM scheduled_tasks WHERE fire_message = false` for current CTE-only tasks.

### Concurrency group starvation problem and prevention rules
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive documents a real incident ("the original maintenance group had both claimed-item-timeout and database-cleanup. When database-cleanup stalled, it blocked claim resets, which blocked the entire pipeline") and gives four prevention rules; entirely absent from live doc.
- **what:** Tasks sharing a `concurrency_group` can starve each other if one never updates `last_completed_at`, permanently occupying the group's `max_concurrent` slot. Prevention: set `timeout_seconds < interval_seconds`, never group unrelated tasks together, ensure every completion path updates `last_completed_at`.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"The Group Starvation Problem", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; last_completed_at ownership
- **verify-later:** query current `scheduled_tasks` group assignments against the archive's "Recommended Group Assignments" table.

### last_completed_at ownership contract and fire_message known-gap
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive explicitly documents: "The scheduler Go code does not currently read this column [fire_message]. It always sends a Kafka message"; none of these operational caveats appear in live doc.
- **what:** Agent-triggered scheduled tasks must include an explicit `notify_scheduler` step on every completion path to set `last_completed_at`; the scheduler itself never sets this column and never reads `fire_message`, flagged as a known low-priority gap.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"last_completed_at — Who Updates It?", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; concurrency group starvation
- **verify-later:** `grep -rn "fire_message" cmd/scheduler/` to check if the Go scheduler now reads this column.

### WireGuard VPN admin-access implementation detail
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** Archive contains full runnable K8s manifests and nginx configs; live `012_admin_dashboard.md`'s condensed section keeps only one-line summaries, drops every YAML/config block.
- **what:** Three documented approaches to securely expose the admin dashboard without public ingress: (A) WireGuard-in-cluster with full K8s manifests, (B) external VM bastion with WireGuard + nginx + TLS + rate limiting, (C) plain `kubectl port-forward`. The live doc retains only the decision framework, not the deployable configuration.
- **sources:** archive_april_26/019_admin_access_infrastructure.md (whole file); docs024_key_docs_latest/012_admin_dashboard.md#"Network Access Options"
- **relations:** admin-dashboard-and-api; auth-service JWT/RequireRole security layer
- **verify-later:** check whether WireGuard was ever actually deployed or whether the system is still on Option C.

### Dual-signal self-heal on missing spec dependency
- **category:** NEW:resilience-self-heal
- **status-signal:** deployed
- **status-evidence:** "validate_composition_inputs both loud-logs AND queues a recovery work item on miss... the two-strike rule marks the item unresolved."
- **what:** General resilience pattern for a Go action depending on a spec aspect that may not yet exist: emit a loud error log AND queue a recovery work item that is both a durable dashboard signal and a self-heal mechanism — if it runs successfully, the dependent item auto-redispatches. Repeated failures accumulate via the two-strike rule into `unresolved`.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"1. Status Summary", #"5. Decisions Made This Session"
- **relations:** site-design-planner Choice B scope; work item two-strike/wont_fix pattern
- **verify-later:** `validate_composition_inputs_action.go` implementation; two-strike rule location in dispatch loop.

### Composition resolver orphan-rows policy
- **category:** NEW:resilience-self-heal
- **status-signal:** aspirational
- **status-evidence:** "If install fails, those rows become orphans... we extend the existing database-cleanup scheduled task to sweep them. Draft SQL in draft_composition_orphan_cleanup.sql."
- **what:** Because palette/typography_set resolvers each commit in their own transaction before `install_site_composition` runs, a failed install leaves orphaned rows. Accepted design: let low-cost orphans occur and sweep them periodically via an extension to the existing `database-cleanup` scheduled task, rather than cross-resolver rollback.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"4. Work Plan — Orphan policy"
- **relations:** composition resolution architecture
- **verify-later:** `database-cleanup` scheduled task pre_query — confirm orphan-sweep CTE merged in.

## Proposed new categories
- **NEW:migration-governance** — governance for hand-applied, DB-object-level migrations to `agent_definitions`/launcher defs (idempotency verification, ordering, ledger). Currently only a manual runbook step ("2d state check"); no automated ledger exists.
- **NEW:resilience-self-heal** — cross-cutting pattern for actions with missing upstream dependencies (spec aspects, composition rows) that combine loud logging with a self-healing recovery work item, tolerating cheap orphaned state rather than enforcing cross-action rollback.
