
<!-- SOURCE: U21_legacy_docs_b.md -->
### Copywriter persona roster
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/010 SQL seeds six personas (Elena Martinez B2B, James Chen technical, Marcus Williams conversion, Aisha Okonkwo thought-leadership, Raj Patel data, Sophie Dubois premium) with style agents; persona_assignments schema in docs010/009; no later builder references personas.
- **what:** A roster of copywriter personas — each a personality profile (biography, Big Five psychology, expertise weights, voice traits) with attached specialized style agents — assigned to flow stages or content types ("assign Marcus to all conversion pages") via personas / specialized_agents / persona_assignments tables and get_persona_for_page lookup (page → stage → default). Voice emerges from persona choice rather than parameter tuning; maps to real agency roles.
- **sources:** docs010_multitrack_flows_persona_architecture/008_example_personas.md; docs010_multitrack_flows_persona_architecture/009_persona_system_schema.sql; docs010_multitrack_flows_persona_architecture/007_personas_discussion.md
- **relations:** persona cognitive architecture; multi-track flows; page-content-writer.
- **verify-later:** personas/specialized_agents/persona_assignments tables in clients_db.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Persona cognitive architecture (swappable cognitive components)
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/015 "The architecture is ready for the full vision while starting simple today"; full schema + 8 cognitive actions + Dr Bimpton example SQL delivered; nothing downstream implements the Go actions.
- **what:** Personas as complete cognitive entities: immutable personality DNA plus swappable subsystems (perception, working/episodic/semantic memory, knowledge retrieval, reasoning engine, response generator, style applicator, learning system), each with pluggable implementations evolving Phase 1 all-LLM → vector-DB memory → fine-tuned persona models → multi-model per task → custom reasoning services, switchable via is_default without workflow changes. Running instances persist memory and emotional state; persona_knowledge holds facts/beliefs/opinions with confidence and future embeddings; task executions log full cognitive traces. Eight-step cognitive workflow per task (initialize→perceive→retrieve→reason→generate→style→learn→complete).
- **sources:** docs010_multitrack_flows_persona_architecture/015_persona_README_architecture.md; docs010_multitrack_flows_persona_architecture/011_persona_cognitive_architecture.sql; docs010_multitrack_flows_persona_architecture/014_drBimpton_setup_example.sql
- **relations:** finetuning-flywheel (fine-tuned persona models); reasoning; entity_state_log (parallel memory design); copywriter roster.
- **verify-later:** personas/persona_cognitive_components/persona_instances/persona_knowledge/persona_task_executions tables; load_cognitive_system etc. in action registry (expected absent).

<!-- SOURCE: U21_legacy_docs_b.md -->
### Copywriter persona roster
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/010 SQL seeds six personas (Elena Martinez B2B, James Chen technical, Marcus Williams conversion, Aisha Okonkwo thought-leadership, Raj Patel data, Sophie Dubois premium) with style agents; persona_assignments schema in docs010/009; no later builder references personas.
- **what:** A roster of copywriter personas — each a personality profile (biography, Big Five psychology, expertise weights, voice traits) with attached specialized style agents — assigned to flow stages or content types ("assign Marcus to all conversion pages") via personas / specialized_agents / persona_assignments tables and get_persona_for_page lookup (page → stage → default). Voice emerges from persona choice rather than parameter tuning; maps to real agency roles.
- **sources:** docs010_multitrack_flows_persona_architecture/008_example_personas.md; docs010_multitrack_flows_persona_architecture/009_persona_system_schema.sql; docs010_multitrack_flows_persona_architecture/007_personas_discussion.md
- **relations:** persona cognitive architecture; multi-track flows; page-content-writer.
- **verify-later:** personas/specialized_agents/persona_assignments tables in clients_db.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Persona cognitive architecture (swappable cognitive components)
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/015 "The architecture is ready for the full vision while starting simple today"; full schema + 8 cognitive actions + Dr Bimpton example SQL delivered; nothing downstream implements the Go actions.
- **what:** Personas as complete cognitive entities: immutable personality DNA plus swappable subsystems (perception, working/episodic/semantic memory, knowledge retrieval, reasoning engine, response generator, style applicator, learning system), each with pluggable implementations evolving Phase 1 all-LLM → vector-DB memory → fine-tuned persona models → multi-model per task → custom reasoning services, switchable via is_default without workflow changes. Running instances persist memory and emotional state; persona_knowledge holds facts/beliefs/opinions with confidence and future embeddings; task executions log full cognitive traces. Eight-step cognitive workflow per task (initialize→perceive→retrieve→reason→generate→style→learn→complete).
- **sources:** docs010_multitrack_flows_persona_architecture/015_persona_README_architecture.md; docs010_multitrack_flows_persona_architecture/011_persona_cognitive_architecture.sql; docs010_multitrack_flows_persona_architecture/014_drBimpton_setup_example.sql
- **relations:** finetuning-flywheel (fine-tuned persona models); reasoning; entity_state_log (parallel memory design); copywriter roster.
- **verify-later:** personas/persona_cognitive_components/persona_instances/persona_knowledge/persona_task_executions tables; load_cognitive_system etc. in action registry (expected absent).
