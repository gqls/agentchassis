# Register — workflow-authoring

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
