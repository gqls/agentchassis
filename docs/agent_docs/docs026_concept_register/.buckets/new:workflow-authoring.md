
<!-- SOURCE: U21_legacy_docs_b.md -->
### Workflow Builder & Validator (YAML DSL)
- **category:** NEW:workflow-authoring
- **status-signal:** abandoned
- **status-evidence:** docs006/001 full design with roadmap claiming "[x] Phase 1: Core parser & validator, [x] Phase 2: Path resolution, [x] Phase 3: JSON generation"; no later doc references the tool; workflows continued to be hand-written SQL.
- **what:** A validation-first system for authoring orchestration workflows in human-readable YAML instead of raw JSON: parses a DSL, validates agent types exist in agent_definitions, detects circular dependencies and invalid input references, auto-computes CollectedData paths (agent call vs local action nesting), generates the orchestration_workflow JSON, test cases, and docs, then inserts into the DB. CLI (`workflow-builder build/validate/test/list/show/docs`), planned HTTP API, web UI, and git-based CI/CD workflow deployment.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#Architecture; docs006_workflow_builder/001_workflow_builder.md#Path-Resolution; docs006_workflow_builder/001_workflow_builder.md#Roadmap
- **relations:** data-path resolution problem; workflow validator tool (docs017/002_standardising); superseded in spirit by input_mapping/ActionInputSpec conventions.
- **verify-later:** platform/workflowbuilder/ directory existence in repo history; any workflow YAML files.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Workflow Builder & Validator (YAML DSL)
- **category:** NEW:workflow-authoring
- **status-signal:** abandoned
- **status-evidence:** docs006/001 full design with roadmap claiming "[x] Phase 1: Core parser & validator, [x] Phase 2: Path resolution, [x] Phase 3: JSON generation"; no later doc references the tool; workflows continued to be hand-written SQL.
- **what:** A validation-first system for authoring orchestration workflows in human-readable YAML instead of raw JSON: parses a DSL, validates agent types exist in agent_definitions, detects circular dependencies and invalid input references, auto-computes CollectedData paths (agent call vs local action nesting), generates the orchestration_workflow JSON, test cases, and docs, then inserts into the DB. CLI (`workflow-builder build/validate/test/list/show/docs`), planned HTTP API, web UI, and git-based CI/CD workflow deployment.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#Architecture; docs006_workflow_builder/001_workflow_builder.md#Path-Resolution; docs006_workflow_builder/001_workflow_builder.md#Roadmap
- **relations:** data-path resolution problem; workflow validator tool (docs017/002_standardising); superseded in spirit by input_mapping/ActionInputSpec conventions.
- **verify-later:** platform/workflowbuilder/ directory existence in repo history; any workflow YAML files.
