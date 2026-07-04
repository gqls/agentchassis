// FILE: platform/orchestration/result_spec.go
//
// Centralised resolution of a completed workflow's result contract, used by
// SagaCoordinator.extractWorkflowResult (coordinator.go). This replaces the old
// inline "if output_fields ... else skipPatterns dump" branch, which silently
// ignored the singular `output_field` and the `output` mapping and so dropped
// or stubbed the results of every agent not using the plural key.
//
// PURE CHASSIS CHANGE — no agent-config edits are required: the writer's
// `output_field`, the plural `output_fields`, and the `output` mappings are all
// READ as deprecated aliases. Deploy = rebuild the chassis image + roll agents
// to the new tag. The rename to result_from / multiple_output_fields /
// result_mapping is a separate, optional migration (old names keep resolving).
//
// Semantics (settled, verified against live consumers — see NOTES):
//   - singular  -> FLATTEN: the named field's CONTENTS become the body. The
//     writer feeds page-build-handler / page-rebuild / site-work-orchestrator,
//     which all read page_content.response.{page_html,sections_metadata,skipped}
//     flat; site-work-orchestrator reads the planner's flattened sub-fields;
//     model-trainer reads preparation_result.dataset_uri flat. Keys: result_from
//     (preferred), output_field (deprecated).
//   - plural    -> FIELDS (nest each under its name). Unchanged for the ~90
//     plural agents. Keys: multiple_output_fields (preferred), output_fields
//     (deprecated).
//   - mapping   -> apply target<-source.path pairs. Was silently dumped before.
//     Keys: result_mapping (preferred), output (deprecated).
//   - none      -> FALLBACK skipPatterns dump (discouraged; large/unbounded).
//
// REUSE: field-name lists use datahelpers.ToStringSlice (path
// platform/orchestration/datahelpers, same as coordinator.go). toStringMap and
// setIfAbsent have no datahelpers equivalent so they are local to this package.
// Pre-merge: grep the rest of package orchestration for any existing toStringMap
// / setIfAbsent to avoid redeclaration (none in coordinator.go / processor.go).

//
// 2026-07-03 (Option A, run 17933a83/73ed55c6 aftermath): the contract table —
// ResultMode/ResultSpec/resolveResultSpec/toStringMap — is LIFTED VERBATIM to
// platform/orchestration/datahelpers/result_contract.go so the response-building
// action (CompleteWorkflowAction.extractFinalResult) reads the SAME table. This
// file keeps the coordinator-facing names as thin aliases/delegates, so
// coordinator.go is untouched; fallbackDumpInto and setIfAbsent (coordinator
// apply path) stay here.

package orchestration

import (
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ResultMode / ResultSpec now live in datahelpers (shared table); the
// aliases keep every existing reference in this package compiling unchanged.
type ResultMode = datahelpers.ResultMode

const (
	ResultModeFallback = datahelpers.ResultModeFallback
	ResultModeFields   = datahelpers.ResultModeFields
	ResultModeFlatten  = datahelpers.ResultModeFlatten
	ResultModeMapping  = datahelpers.ResultModeMapping
)

type ResultSpec = datahelpers.ResultSpec

// resolveResultSpec delegates to the shared table (single source of truth for
// the complete-step key vocabulary).
func resolveResultSpec(completeConfig map[string]interface{}, logger *zap.Logger) ResultSpec {
	return datahelpers.ResolveResultSpec(completeConfig, logger)
}

// fallbackDumpInto is the previous default behaviour, extracted verbatim so
// extractWorkflowResult's switch stays readable. Unchanged skipPatterns.
func (s *SagaCoordinator) fallbackDumpInto(result map[string]interface{}, state *OrchestrationState) {
	skipPatterns := []string{
		"page_content_",     // Individual page content (can be large)
		"reviewed_content_", // Reviewed versions
		"build_pages_loop_", // Loop iteration data
		"assembled_page",    // Full HTML
		"page_deployed_",    // Deployment results per page
	}
	for key, value := range state.CollectedData {
		if strings.HasPrefix(key, "__") {
			continue
		}
		if key == "loop_metadata" {
			continue
		}
		skip := false
		for _, pattern := range skipPatterns {
			if strings.HasPrefix(key, pattern) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		result[key] = datahelpers.ExtractStepData(value)
	}
}

// --- small helpers ----------------------------------------------------------
// REUSE: field-name lists go through datahelpers.ToStringSlice (above).
// toStringMap and setIfAbsent have no datahelpers equivalent (the package has
// CleanDataMap / GetMapKeys / ExtractNestedField* / ExtractStepData / ToStringSlice
// only), so they stay here, local to package orchestration.

func setIfAbsent(m map[string]interface{}, key string, value interface{}) {
	if _, exists := m[key]; !exists {
		m[key] = value
	}
}
