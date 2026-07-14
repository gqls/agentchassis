
<!-- SOURCE: U26_misc_dirs.md -->
### Agent spawning (agents as DB records claimed by generic pods)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001: "The spawn_group action creates new rows in the agent_instances table"; 004_debugging references `kubectl delete jobs -l spawned-by=orchestrator` — orchestrator-spawned jobs existed in the running cluster.
- **what:** A spawn_agent action creates (or reuses) an agent_instances row with type, workflow, capabilities and llm_config; a generic chassis pod (static env assignment, dynamic pool claim, or K8s Job spawned by the orchestrator) loads that config and becomes the agent. Includes existence-check-and-reuse logic and pod-type selection (CPU/GPU/memory) for specialised workloads.
- **sources:** docs/architecture/023-spawning-agents.md; docs/architecture/025-reusable-evolvable-agent-teams#step-1; docs/basic_usage/001basic_usage.txt#part-2
- **relations:** agent chassis; agent groups; agent discovery
- **verify-later:** spawn_actions.go; assigned_pod/status/last_heartbeat columns; Job-spawning code

<!-- SOURCE: U26_misc_dirs.md -->
### Agent groups (reusable multi-agent teams)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001 documents live spawn_group runs against the agent_groups table; hitl_agent_group_definition.sql (Nov 2025) inserts into a matured `agent_group_definitions` table with versioning and ON CONFLICT upsert.
- **what:** A named, versioned team definition: group_type, agent_configs (roles → agent types), an orchestration_workflow describing how they cooperate, capabilities/tags for search, and usage/performance metadata. `spawn_group` instantiates the team and starts the group workflow. Groups were intended to be saved from successful configurations, forked, and improved over generations.
- **sources:** docs/architecture/025-reusable-evolvable-agent-teams#phase-2; docs/humanintheloop/hitl_agent_group_definition.sql; docs/basic_usage/001basic_usage.txt
- **relations:** website-builder group; controlled group evolution; group discovery
- **verify-later:** agent_groups vs agent_group_definitions tables; SpawnGroupAction code

<!-- SOURCE: U26_misc_dirs.md -->
### Agent and group discovery by capability and performance
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 024-agent-discovery-framework.md reads as pure proposal ("Yes! Let's design this") with an unconfirmed agent_metrics/heartbeat table, BUT a live repo check (not just docs) found `platform/discovery/` actually exists — `agent_discovery.go` plus `README.001.agentdefinitions.md` and `README.002.dbtables.md` — confirming a real, if unknown-scope, implementation landed. Upgraded from aspirational on that direct code evidence; the sophistication of the shipped version (whether performance-ranking/heartbeats made it in) is still stage-2 work.
- **what:** A registry service that finds the best existing agent (or group) for a task by required capabilities (JSONB containment), success rate, response time, availability (heartbeat) and fuel cost — spawning a new one only when nothing matches. The "self-organizing system" goal: agents discover each other, learn which perform best, optimise over time.
- **sources:** docs/architecture/024-agent-discovery-framework.md; docs/architecture/027-create-website-creation-system#phase-3; docs/architecture/025-reusable-evolvable-agent-teams#group-discovery-service
- **relations:** agent spawning; agent groups; template classification
- **verify-later:** platform/discovery/agent_discovery.go, platform/discovery/README.001.agentdefinitions.md, platform/discovery/README.002.dbtables.md — confirm scope of what actually shipped vs. this proposal

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow template library, lineage and marketplace
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 005/006 propose workflow_templates + template_versions tables, fork lineage graphs and a monetised marketplace ("marketplace of evolving workflow templates"); roadmap slots it at Phase 6 (weeks 10-12); no later doc in the repo era references these tables.
- **what:** Successful workflow executions are saved as reusable templates with lineage (parent_template_id, source_correlation_id), performance metrics, ratings and usage counts; users fork and improve templates ("collective intelligence — natural selection of best-performing templates"), with a WorkflowOptimizer suggesting parallelisation/reordering from execution history. A precursor idea whose spirit partially survives in agent group versioning.
- **sources:** docs/architecture/005-template-classification-and-evolution.md; docs/architecture/007-roadmap.md#phase-6; docs/architecture/016-competitive-advantge.md#workflow-marketplace-2.0
- **relations:** agent groups; template classification and search; controlled group evolution
- **verify-later:** confirm workflow_templates tables never created

<!-- SOURCE: U26_misc_dirs.md -->
### Multi-dimensional template classification and semantic search
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 006 is pure design (behavioral profiles, performance vectors, embedding fingerprints, vector DB + graph DB search service, recommendation engine); none of the infrastructure (vectorDB for templates, template_usage_metrics) appears in later docs.
- **what:** Replace flat tags with a rich classification for discovering workflows/templates: behavioral capabilities, execution style, resource usage, normalised performance vectors with trade-off scores, embedding-based semantic fingerprints, outcome-based deliverable descriptions and evolutionary metadata (lineage depth, fork count). A parallel multi-strategy search (semantic + behavioral + performance + lineage) with collaborative-filtering recommendations.
- **sources:** docs/architecture/006-categorisation-template-search-evolution.md
- **relations:** workflow template library; agent/group discovery
- **verify-later:** n/a (idea only) — check nothing similar exists under another name

<!-- SOURCE: U26_misc_dirs.md -->
### Controlled group evolution (observed mutation with rules)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 026 reads as pure future-tense recommendation ("Start with curated base templates that can dynamically evolve", MutationRules, sandbox testing, 10%-improvement gates) with nothing downstream in this unit confirming it. BUT a live repo check found `platform/evolution/` actually exists — `evolution.go`, `performance.go`, and a README stating in the present tense: "Evolution Service: Evaluates groups for potential improvements / Applies mutations (parallel agents, specialists, validators) / Tracks evolution history / Version management" and "Performance Analysis: Records and analyzes execution metrics / Identifies bottlenecks and failures / Generates improvement suggestions". Upgraded from abandoned on that direct code evidence — this idea did get built, just not confirmed anywhere in the docs sampled for this unit.
- **what:** Hybrid manual/dynamic strategy: hand-curated base agent and group templates act as "genetic seeds"; a metrics observer detects bottlenecks/missing capabilities after ≥5 uses and proposes constrained mutations (add parallel agent, add specialist, replace, adjust workflow, fork) which must beat a performance baseline in sandbox before becoming a new version with parent lineage. Human-in-the-loop approval tiers for major changes.
- **sources:** docs/architecture/026-manual-vs-dynamic-templates; docs/architecture/025-reusable-evolvable-agent-teams#usage-example
- **relations:** agent groups; dynamic prompt improvement loop; workflow template library
- **verify-later:** platform/evolution/evolution.go, platform/evolution/performance.go, platform/evolution/README.md — confirm how closely the shipped mutation rules match this design

<!-- SOURCE: U26_misc_dirs.md -->
### Dynamic prompt improvement loop (Prompt Improvement Agent)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** aspirational
- **status-evidence:** basic_usage/003 is a plan ("Phase 2: The Evolution Loop") — flag-for-improvement UI action, `system.prompt.improvement.request` topic, save_new_agent_definition action creating versioned definitions (html-developer-v2); the agent_definitions versioning columns (version, previous_version_id) DID ship (visible in hitl_agent_definition.sql), the loop itself is not evidenced.
- **what:** End-of-workflow human review offers "Approve" or "Flag for Improvement"; flagged runs dispatch the failing agent's prompt + failure context to a prompt-engineering specialist agent which generates an improved prompt, gets human approval, and saves a NEW versioned agent definition (never mutating the old). Includes bootstrap_prompt for generating a first prompt for brand-new agent types from a description.
- **sources:** docs/basic_usage/003_dynamic_prompt_improvement; docs/humanintheloop/hitl_agent_definition.sql (version/previous_version_id columns)
- **relations:** execute_llm_prompt; HITL approval mechanism; controlled group evolution; (spiritual ancestor of the current improvement-loop / finetuning flywheel)
- **verify-later:** prompt-improvement-agent definition; save_new_agent_definition action; version columns usage

<!-- SOURCE: U26_misc_dirs.md -->
### Agent spawning (agents as DB records claimed by generic pods)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001: "The spawn_group action creates new rows in the agent_instances table"; 004_debugging references `kubectl delete jobs -l spawned-by=orchestrator` — orchestrator-spawned jobs existed in the running cluster.
- **what:** A spawn_agent action creates (or reuses) an agent_instances row with type, workflow, capabilities and llm_config; a generic chassis pod (static env assignment, dynamic pool claim, or K8s Job spawned by the orchestrator) loads that config and becomes the agent. Includes existence-check-and-reuse logic and pod-type selection (CPU/GPU/memory) for specialised workloads.
- **sources:** docs/architecture/023-spawning-agents.md; docs/architecture/025-reusable-evolvable-agent-teams#step-1; docs/basic_usage/001basic_usage.txt#part-2
- **relations:** agent chassis; agent groups; agent discovery
- **verify-later:** spawn_actions.go; assigned_pod/status/last_heartbeat columns; Job-spawning code

<!-- SOURCE: U26_misc_dirs.md -->
### Agent groups (reusable multi-agent teams)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001 documents live spawn_group runs against the agent_groups table; hitl_agent_group_definition.sql (Nov 2025) inserts into a matured `agent_group_definitions` table with versioning and ON CONFLICT upsert.
- **what:** A named, versioned team definition: group_type, agent_configs (roles → agent types), an orchestration_workflow describing how they cooperate, capabilities/tags for search, and usage/performance metadata. `spawn_group` instantiates the team and starts the group workflow. Groups were intended to be saved from successful configurations, forked, and improved over generations.
- **sources:** docs/architecture/025-reusable-evolvable-agent-teams#phase-2; docs/humanintheloop/hitl_agent_group_definition.sql; docs/basic_usage/001basic_usage.txt
- **relations:** website-builder group; controlled group evolution; group discovery
- **verify-later:** agent_groups vs agent_group_definitions tables; SpawnGroupAction code

<!-- SOURCE: U26_misc_dirs.md -->
### Agent and group discovery by capability and performance
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 024-agent-discovery-framework.md reads as pure proposal ("Yes! Let's design this") with an unconfirmed agent_metrics/heartbeat table, BUT a live repo check (not just docs) found `platform/discovery/` actually exists — `agent_discovery.go` plus `README.001.agentdefinitions.md` and `README.002.dbtables.md` — confirming a real, if unknown-scope, implementation landed. Upgraded from aspirational on that direct code evidence; the sophistication of the shipped version (whether performance-ranking/heartbeats made it in) is still stage-2 work.
- **what:** A registry service that finds the best existing agent (or group) for a task by required capabilities (JSONB containment), success rate, response time, availability (heartbeat) and fuel cost — spawning a new one only when nothing matches. The "self-organizing system" goal: agents discover each other, learn which perform best, optimise over time.
- **sources:** docs/architecture/024-agent-discovery-framework.md; docs/architecture/027-create-website-creation-system#phase-3; docs/architecture/025-reusable-evolvable-agent-teams#group-discovery-service
- **relations:** agent spawning; agent groups; template classification
- **verify-later:** platform/discovery/agent_discovery.go, platform/discovery/README.001.agentdefinitions.md, platform/discovery/README.002.dbtables.md — confirm scope of what actually shipped vs. this proposal

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow template library, lineage and marketplace
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 005/006 propose workflow_templates + template_versions tables, fork lineage graphs and a monetised marketplace ("marketplace of evolving workflow templates"); roadmap slots it at Phase 6 (weeks 10-12); no later doc in the repo era references these tables.
- **what:** Successful workflow executions are saved as reusable templates with lineage (parent_template_id, source_correlation_id), performance metrics, ratings and usage counts; users fork and improve templates ("collective intelligence — natural selection of best-performing templates"), with a WorkflowOptimizer suggesting parallelisation/reordering from execution history. A precursor idea whose spirit partially survives in agent group versioning.
- **sources:** docs/architecture/005-template-classification-and-evolution.md; docs/architecture/007-roadmap.md#phase-6; docs/architecture/016-competitive-advantge.md#workflow-marketplace-2.0
- **relations:** agent groups; template classification and search; controlled group evolution
- **verify-later:** confirm workflow_templates tables never created

<!-- SOURCE: U26_misc_dirs.md -->
### Multi-dimensional template classification and semantic search
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 006 is pure design (behavioral profiles, performance vectors, embedding fingerprints, vector DB + graph DB search service, recommendation engine); none of the infrastructure (vectorDB for templates, template_usage_metrics) appears in later docs.
- **what:** Replace flat tags with a rich classification for discovering workflows/templates: behavioral capabilities, execution style, resource usage, normalised performance vectors with trade-off scores, embedding-based semantic fingerprints, outcome-based deliverable descriptions and evolutionary metadata (lineage depth, fork count). A parallel multi-strategy search (semantic + behavioral + performance + lineage) with collaborative-filtering recommendations.
- **sources:** docs/architecture/006-categorisation-template-search-evolution.md
- **relations:** workflow template library; agent/group discovery
- **verify-later:** n/a (idea only) — check nothing similar exists under another name

<!-- SOURCE: U26_misc_dirs.md -->
### Controlled group evolution (observed mutation with rules)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 026 reads as pure future-tense recommendation ("Start with curated base templates that can dynamically evolve", MutationRules, sandbox testing, 10%-improvement gates) with nothing downstream in this unit confirming it. BUT a live repo check found `platform/evolution/` actually exists — `evolution.go`, `performance.go`, and a README stating in the present tense: "Evolution Service: Evaluates groups for potential improvements / Applies mutations (parallel agents, specialists, validators) / Tracks evolution history / Version management" and "Performance Analysis: Records and analyzes execution metrics / Identifies bottlenecks and failures / Generates improvement suggestions". Upgraded from abandoned on that direct code evidence — this idea did get built, just not confirmed anywhere in the docs sampled for this unit.
- **what:** Hybrid manual/dynamic strategy: hand-curated base agent and group templates act as "genetic seeds"; a metrics observer detects bottlenecks/missing capabilities after ≥5 uses and proposes constrained mutations (add parallel agent, add specialist, replace, adjust workflow, fork) which must beat a performance baseline in sandbox before becoming a new version with parent lineage. Human-in-the-loop approval tiers for major changes.
- **sources:** docs/architecture/026-manual-vs-dynamic-templates; docs/architecture/025-reusable-evolvable-agent-teams#usage-example
- **relations:** agent groups; dynamic prompt improvement loop; workflow template library
- **verify-later:** platform/evolution/evolution.go, platform/evolution/performance.go, platform/evolution/README.md — confirm how closely the shipped mutation rules match this design

<!-- SOURCE: U26_misc_dirs.md -->
### Dynamic prompt improvement loop (Prompt Improvement Agent)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** aspirational
- **status-evidence:** basic_usage/003 is a plan ("Phase 2: The Evolution Loop") — flag-for-improvement UI action, `system.prompt.improvement.request` topic, save_new_agent_definition action creating versioned definitions (html-developer-v2); the agent_definitions versioning columns (version, previous_version_id) DID ship (visible in hitl_agent_definition.sql), the loop itself is not evidenced.
- **what:** End-of-workflow human review offers "Approve" or "Flag for Improvement"; flagged runs dispatch the failing agent's prompt + failure context to a prompt-engineering specialist agent which generates an improved prompt, gets human approval, and saves a NEW versioned agent definition (never mutating the old). Includes bootstrap_prompt for generating a first prompt for brand-new agent types from a description.
- **sources:** docs/basic_usage/003_dynamic_prompt_improvement; docs/humanintheloop/hitl_agent_definition.sql (version/previous_version_id columns)
- **relations:** execute_llm_prompt; HITL approval mechanism; controlled group evolution; (spiritual ancestor of the current improvement-loop / finetuning flywheel)
- **verify-later:** prompt-improvement-agent definition; save_new_agent_definition action; version columns usage
