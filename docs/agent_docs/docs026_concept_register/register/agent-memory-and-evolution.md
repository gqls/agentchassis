# Register — agent-memory-and-evolution

3 concepts, consolidated from 6 raw extractions (3 unique blocks, each present
twice in the source cluster file due to mechanical duplication in the input),
all from unit U21.

### AME-001 — entity_state_log — append-only cross-orchestration memory
- **status:** abandoned
- **status-evidence:** Full schema + five Go actions (append_entity_state, read_latest_entity_state, read_entity_history, read_my_state, write_my_state) in docs006/002 and migration SQL in docs006/007; no later documents reference entity_state_log in the build pipeline.
- **what:** Persistent data that survives across orchestrations: an append-only log keyed by entity_id/namespace/path with accumulation patterns (additive, evolutionary, versioned, singleton), agent-namespaced storage ("read_my_state"/"write_my_state" use AGENT_TYPE as namespace), supersession pointers for future compaction, and LLM-based consolidation as a future enhancement. Intended for accumulating research, brand learnings, and build history per domain.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Part-5; docs006_workflow_builder/007_new_tables_entity_state_log.sql; docs006_workflow_builder/004_agent_groups_or_not.md#Where-Learnings-Live
- **relations:** four-level learnings model; relationships table; improvement_proposals; conceptual ancestor of per-site content_data accumulation
- **verify-later:** entity_state_log table existence in clients_db; entity_state_actions.go in repo

### AME-002 — Agent variants + snapshot versioning
- **status:** partial
- **status-evidence:** docs006/004: "Snapshot model (preferred): variants explicitly reference a snapshot version"; is_snapshot column added in docs006/007 migration; agent_variants table proposed but never seen again.
- **stage2-verified (2026-07-14):** abandoned → partial — agent_variants table itself: 0 CREATE TABLE hits (never built, discovery_actions.go:671 comment 'if exists'). But is_snapshot flag from the same design is live and wired: platform/messaging/processor.go:351,360; platform/discovery/agent_discovery.go:99,118,188; platform/orchestration/actions/spawn_actions.go:2124,21...
- **what:** A controlled-evolution model for agent definitions: base agents are versioned and can be frozen as snapshots (is_snapshot flag); task variants (agent_variants) reference a specific base version with config_overrides, metrics, and lineage, so the base can evolve without breaking variants. Three evolution types (bug fix / improvement / innovation) with escalating oversight; promotion of successful variants to new bases left as an open question.
- **sources:** docs006_workflow_builder/004_agent_groups_or_not.md#The-Fragility-Problem; docs006_workflow_builder/006b_evolution_design_discussion.md; docs006_workflow_builder/007_new_tables_entity_state_log.sql
- **relations:** improvement_proposals; four-level learnings model; agent_definitions versioning today; Agent definition snapshot/revert via backup table (agent-definition-registry register — a later, differently-shaped snapshot mechanism)
- **verify-later:** is_snapshot/usage_count columns on agent_definitions; whether agent_variants table was ever created

### AME-003 — improvement_proposals — HITL-gated agent evolution queue
- **status:** abandoned
- **status-evidence:** docs006/004: "The system proposes changes but requires HITL approval before applying"; docs006/006 lists ReviewPerformanceAction → improvement_proposals and ApproveImprovementAction; not referenced by later architectures.
- **what:** A review queue where proposed changes to agent definitions, variants, or entity knowledge — sourced from metrics regressions, agent observations, or humans — wait as pending proposals until a human approves, rejects, or applies them. Included review_performance action recording execution metrics to entity_state_log and generating proposals.
- **sources:** docs006_workflow_builder/004_agent_groups_or_not.md#What-Triggers-Evolution; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Discovery-Actions-Changes
- **relations:** conceptual ancestor of the improvement-loop's suggest/flag resolution paths and HITL approvals (hitl register)
- **verify-later:** improvement_proposals table; discovery_actions.go history
