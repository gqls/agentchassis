# Register — workflow-authoring

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 1 raw extraction across unit U21. No duplicates
found within this category's raw material (this new category surfaced a single
raw block from the whole cluster). Note: material about workflow-authoring
mechanics that is more current/live (loop mechanisms, error_step placement,
input_mapping/ActionInputSpec, spawn/step naming conventions) was tagged
`development-guide` by the extracting units and is registered there
(development-guide.md) rather than here — read alongside this entry, since it
covers the same broad subject (how workflows are authored) from the currently
deployed angle, whereas this entry is a single abandoned historical proposal.

### WFA-001 — Workflow Builder & Validator (YAML DSL)
- **status:** abandoned
- **status-evidence:** docs006/001 full design with a roadmap claiming "[x] Phase 1: Core parser & validator, [x] Phase 2: Path resolution, [x] Phase 3: JSON generation"; no later doc references the tool, and workflows continued to be hand-written SQL/JSON thereafter.
- **what:** A validation-first system for authoring orchestration workflows in human-readable YAML instead of raw JSON: parses a DSL, validates agent types exist in agent_definitions, detects circular dependencies and invalid input references, auto-computes CollectedData paths (agent call vs local action nesting), generates the orchestration_workflow JSON, test cases, and docs, then inserts into the DB. Planned CLI (`workflow-builder build/validate/test/list/show/docs`), an HTTP API, a web UI, and a git-based CI/CD workflow-deployment pipeline — none of which are evidenced as ever having shipped or been used beyond the initial design phases.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#Architecture; docs006_workflow_builder/001_workflow_builder.md#Path-Resolution; docs006_workflow_builder/001_workflow_builder.md#Roadmap
- **relations:** data-path resolution problem (agent vs local action nesting) — development-guide register; a later, differently-scoped workflow validator tool (docs017/002_standardising); superseded in spirit by the input_mapping/ActionInputSpec conventions that development-guide documents as the live mechanism
- **verify-later:** platform/workflowbuilder/ directory existence in repo history; any workflow YAML files ever produced by it

### WFA-002 — `$ctx.` execution-context parameter namespace for `query_database`
- **status:** deployed
- **status-evidence:** live on chassis v1.0.1191 (digest `sha256:2f96b795a5c4…`), pod-verified `strings /app/agent-chassis | grep -c "unknown execution-context field"` → 1; first consumer live via migration `258_diagnose_loop_stamps_run_correlation.sql`, verified end-to-end on a real diagnosis 2026-07-28 17:05 (`spec.dispatch_correlation_id = 66a65287-…`).
- **what:** A reserved parameter namespace letting any workflow's `query_database` step bind **the identity of the run executing it** — `$ctx.correlation_id`, `$ctx.orchestration_id`, `$ctx.parent_orchestration_id`, `$ctx.orchestration_name`, `$ctx.client_id`, `$ctx.request_id`, `$ctx.step_name`, `$ctx.group_id` — as ordinary `$1`-style SQL parameters. Written as `'params', jsonb_build_array('$ctx.correlation_id')` beside the query. Before it, `query_database` could only bind values found in `collected_data`, and a run's own correlation is not there: it lives in the execution context. So a dispatch loop could claim a queued row but could not record **which run took it**, and every lane that wanted that had to grow a bespoke Go action. Additive by construction — only paths beginning `$ctx.` take the branch, and no `collected_data` key can start with `$`. An unknown field or an empty value is an **error**, never a silent empty string (a silently-empty bind fills a stamp column with `''` and looks populated in every subsequent query); there is deliberately no "optional" mode.
- **why it is registerable:** "which run picked this row up" is a question every queue-driven workflow on this platform has, not a diagnose-specific one. Any dispatch loop, claim step or audit trail can use it without new Go.
- **sources:** platform/orchestration/actions/execution_context_params.go; platform/orchestration/actions/database_actions.go#QueryDatabaseAction; docs024/bugfix_124_double_dispatch/PLAN_2026-07-28_double_dispatch.md §4 P3; docs/agent_docs/sql_for_agents/258_diagnose_loop_stamps_run_correlation.sql
- **relations:** SCH-012 diagnose-dispatch-loop (first consumer); work-item claim/dedup contract; diagnosis_artifacts correlation keying (`diagnose_assemble_bundle_action.go` writes on `params.ExecutionContext.CorrelationID`, which is what made the item↔run join impossible)
- **landmine:** **ordering is load-bearing and permanent.** A config binding `$ctx.…` run against a chassis that predates this action FAILS that step (`query param path '$ctx.correlation_id' resolved to nil`) — for a claim step that stops the whole lane. Ship the image, pod-grep, then apply the config; and a rollback below the image must revert the config too.
- **review status:** council **REJECTED** it (guardian veto on SCOPE — a platform seam inside a bug patch, corr `90361922`), then **OWNER RULING 2026-07-28: Option A — keep the code, fix the precedent.** The seam stays; the rule it produced is in `CLAUDE.md` §"Platform seams and the ordering exemption". Full review + costed options: `docs024/bugfix_124_double_dispatch/REVIEW_2026-07-28_ctx_namespace.md`. **Whether it should acquire a SECOND consumer is deliberately not ruled on** — nothing depends on it today, and a reviewer asked for it to be judged on its own merits.
- **measured, not asserted (the council forced these):** 63 `params` entries across every live workflow, exactly **1** `$`-prefixed (this one) ⇒ nothing can be shadowed. Of 34 files reading `ExecutionContext.CorrelationID`, only **2** bind it into SQL and 16 use it as a Kafka partition key — so this does **not** duplicate them and there is nothing to migrate.
- **verify-later:** `strings /app/agent-chassis | grep -c "unknown execution-context field"` on the running pod; `SELECT default_config->'workflow'->'steps'->'claim_item'->'config'->'params' FROM agent_definitions WHERE type='diagnose-dispatch-loop'`
