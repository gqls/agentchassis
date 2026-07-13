# Bundle 2 — chassis: the idea generator/validator as a workflow + actions

**What this chat is for:** re-implement the idea.uk engine (the idea **generator + validator**
method) natively in the agent chassis as a **separate workflow** whose steps call actions —
to be **merged into the site-build workflow later**. Build it the chassis way: every agent is
an orchestrator that owns a workflow of steps calling actions; keep the workflow simple and
put the logic in Go action code; reuse existing actions before writing new ones.

This bundle has two halves: **(A)** the engine to port (download from the idea.uk outputs) and
**(B)** the chassis framework to build it in (already in your project).

---

## A. The thing to port — download from the idea.uk outputs

The method, stage by stage (this is what becomes workflow steps + actions):
*challenge the audience → generate ideas across four lenses → cut them against the free
alternative → verify survivors with web search → score (incl. the **Risk** column — 6th
factor, gate Definition≥3 AND Willingness≥3) → rank.*

- `idea-go/engine.go` — the orchestration of those stages (the source of truth for the flow)
- `idea-go/prompts.go` — the per-stage prompts → become prompt templates / `execute_llm_prompt`
  inputs
- `idea-go/audience_check.go` — the free 30-second taster (a natural sub-agent / sub-workflow)
- `idea_uk_method_v0.md`, `idea_method_prompt.md` — the method written up (the spec)
- `idea_uk_testrun_v2.md` — a worked example of the output; **validate the chassis port
  against this**
- `PARALLEL_engine_deployment_and_layer5.md`, `CONSOLIDATION_where_it_all_fits.md` — how the
  engine is meant to fit the chassis (the chassis-native engine = Phase D)

---

## B. The chassis framework — already in your project

### Start + the guidelines (the rules to follow)
- `000_documentation_index.md` — master index
- `001_development_guide_3_.md`, `002_system_architecture.md` — the **agent-creation
  guidelines** (how agents/workflows/actions are built)
- `003_contracts_and_standards.md` — the workflow/action contracts + variable conventions

### Current platform state (so you match live patterns)
- The most recent `HANDOFF_2026-06-09_*.md` + `running_notes_16_content_quality_and_internal_linking.md`
- `016_debugging_guide_v2.md` — the chassis debug guide (note: the idea.uk one is the v2_32
  copy in Bundle 1)

### Tools / actions — reuse before building
- `019_tool_library.md`, `020_tool_lifecycle.md` — how actions/tools are catalogued and built
- `production_agent-chassis-actions-current_context.txt` — the current actions; reuse
  `execute_llm_prompt` and the existing search/extract actions rather than writing new ones

### Core orchestration (Go) — reuse, don't recreate
- `registry.go` (+ `registry_go.txt`), `coordinator.go`, `queryresolve.go`,
  `action_inputs.go`, `spawn_actions.go`, `safe_unmarshal.go`, `timeout_helpers.go`,
  `sql_helpers.go`

### Search / extract (Go) — for the "verify with web search" stage
- `content_search.go`, `deep_search.go`, `unified_extractor.go` (the chassis likely already
  has the search/extract the verify step needs)

### Model selection (best models + vendor swap-ability)
- `009_model_infrastructure.md`, `019_model_lifecycle_schema.sql`,
  `021_model_swap_and_rollback.sql`, `101_switch_to_haiku.sql`

### Quality / validation
- `023_llm_quality_testing.md` — relevant to the validator/scoring half

### Schemas — check before writing any SQL
- `002_intake_orchestrator.sql` (orchestrator/workflow), `agent_definition_types.sql`
  (agent/workflow/action definitions), `018_briefing_questionnaire.sql`,
  `schemas_all`, `schemas_some` (full DB schema dumps — the quickest reference)

### Seed / table content — the message-envelope pattern for logging
- `initial_messages__without_current_ids_`, `initial_vet_practice_check_message`
  (show how agents/messages are seeded and the header/body shape you want logged)

---

## C. For the LATER merge into site-build (attach when you get there, not now)
- `021_site_spec_and_classifier.sql` + `021_site_spec_and_classifier.md`
- `029_site_plan_and_reconciler.md`, `030_phase1_plan_and_reconciler.md`
- `v3_site_actions.go`, `write_site_plan_action.go`, `site_spec_actions.go`, `site_db_actions.go`

---

## Rules of the road (restate to the new chat if the project instructions don't carry over)
- **Every agent is an orchestrator.** An agent owns a workflow of one-or-more steps that call
  actions.
- Keep **workflows simple**; put complexity in **Go action code**.
- Keep **workflow variable names in sync** with what the actions expect.
- **No subworkflows in SQL** — spawn **sub-agents** with their own workflows (keeps logs
  clear, maintenance easier, responsibilities separate).
- Agents respond to the **caller's (parent's) responses topic**, not their own.
- Distinct, minimally-overlapping responsibilities; detailed per-stage prompts; sub-agents for
  e.g. research-before-write.
- **Reuse/alter existing functions and structs** before creating new.
- **Check the DB schema before writing SQL.** A 0-row result isn't decisive until you've ruled
  out the query itself.
- **Don't use `logger.Debug`** (won't show in the logs).
- Log **agent creation** and **inter-agent messages (headers + body)**.
- Deploy: **github → GitHub Actions → Backblaze B2**.
- Kafka cluster `personae-kafka-cluster-...`; namespaces `ai-persona-system` and `kafka`.

## First task framing
Stand up a **new agent** whose workflow steps are the engine's stages, each step calling an
action (reuse `execute_llm_prompt` + the existing search/extract actions). Keep it a
**standalone workflow** (its own agent), validate its output against `idea_uk_testrun_v2.md`,
and design the stage boundaries so it can later be **merged into the site-build workflow** as
a sub-agent. Check the relevant schema before any SQL; reuse existing structs/functions before
writing new ones.
