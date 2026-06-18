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

package orchestration

import (
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ResultMode is how a completed workflow's payload is shaped before return.
// Exactly one mode is selected per workflow by resolveResultSpec.
type ResultMode int

const (
	// ResultModeFallback: no contract declared -> legacy skipPatterns dump.
	ResultModeFallback ResultMode = iota
	// ResultModeFields: nest each named field under its own key.
	ResultModeFields
	// ResultModeFlatten: the single named field's CONTENTS become the body.
	ResultModeFlatten
	// ResultModeMapping: build result from explicit target<-source.path pairs.
	ResultModeMapping
)

func (m ResultMode) String() string {
	switch m {
	case ResultModeFields:
		return "fields"
	case ResultModeFlatten:
		return "flatten"
	case ResultModeMapping:
		return "mapping"
	default:
		return "fallback"
	}
}

// ResultSpec is the normalised result contract for a `complete` step.
type ResultSpec struct {
	Mode       ResultMode
	Fields     []string          // ResultModeFields
	From       string            // ResultModeFlatten
	Mapping    map[string]string // ResultModeMapping
	MatchedKey string            // which config key selected this (for logs)
}

// resolveResultSpec inspects the `complete` step config and returns the
// normalised contract. Precedence is the candidates order below; when (against
// guidance) more than one key is present the highest-precedence one wins and the
// conflict is logged. Deprecated keys are accepted with a migration Warn. No
// logger.Debug — Debug does not surface in our logs.
func resolveResultSpec(completeConfig map[string]interface{}, logger *zap.Logger) ResultSpec {
	if completeConfig == nil {
		return ResultSpec{Mode: ResultModeFallback}
	}

	type candidate struct {
		key        string
		deprecated string // preferred replacement name, "" if current
		build      func() (ResultSpec, bool)
	}

	candidates := []candidate{
		{key: "result_from", build: func() (ResultSpec, bool) {
			s, ok := completeConfig["result_from"].(string)
			if !ok || s == "" {
				return ResultSpec{}, false
			}
			return ResultSpec{Mode: ResultModeFlatten, From: s, MatchedKey: "result_from"}, true
		}},
		{key: "output_field", deprecated: "result_from", build: func() (ResultSpec, bool) {
			s, ok := completeConfig["output_field"].(string)
			if !ok || s == "" {
				return ResultSpec{}, false
			}
			return ResultSpec{Mode: ResultModeFlatten, From: s, MatchedKey: "output_field"}, true
		}},
		{key: "multiple_output_fields", build: func() (ResultSpec, bool) {
			raw, ok := completeConfig["multiple_output_fields"].([]interface{})
			if !ok {
				return ResultSpec{}, false
			}
			f := datahelpers.ToStringSlice(raw)
			if len(f) == 0 {
				return ResultSpec{}, false
			}
			return ResultSpec{Mode: ResultModeFields, Fields: f, MatchedKey: "multiple_output_fields"}, true
		}},
		{key: "output_fields", deprecated: "multiple_output_fields", build: func() (ResultSpec, bool) {
			raw, ok := completeConfig["output_fields"].([]interface{})
			if !ok {
				return ResultSpec{}, false
			}
			f := datahelpers.ToStringSlice(raw)
			if len(f) == 0 {
				return ResultSpec{}, false
			}
			return ResultSpec{Mode: ResultModeFields, Fields: f, MatchedKey: "output_fields"}, true
		}},
		{key: "result_mapping", build: func() (ResultSpec, bool) {
			m, ok := toStringMap(completeConfig["result_mapping"])
			if !ok {
				return ResultSpec{}, false
			}
			return ResultSpec{Mode: ResultModeMapping, Mapping: m, MatchedKey: "result_mapping"}, true
		}},
		{key: "output", deprecated: "result_mapping", build: func() (ResultSpec, bool) {
			m, ok := toStringMap(completeConfig["output"])
			if !ok {
				return ResultSpec{}, false
			}
			return ResultSpec{Mode: ResultModeMapping, Mapping: m, MatchedKey: "output"}, true
		}},
	}

	var chosen *ResultSpec
	var chosenDeprecated string
	var present []string

	for i := range candidates {
		c := candidates[i]
		spec, ok := c.build()
		if !ok {
			continue
		}
		present = append(present, c.key)
		if chosen == nil {
			s := spec
			chosen = &s
			chosenDeprecated = c.deprecated
		}
	}

	if chosen == nil {
		logger.Info("resolveResultSpec: no result contract on complete step; using fallback dump")
		return ResultSpec{Mode: ResultModeFallback}
	}

	if chosenDeprecated != "" {
		logger.Warn("resolveResultSpec: deprecated complete-step key in use; migrate to preferred name",
			zap.String("deprecated_key", chosen.MatchedKey),
			zap.String("preferred", chosenDeprecated))
	}

	if len(present) > 1 {
		logger.Warn("resolveResultSpec: multiple result-contract keys present; using highest precedence",
			zap.Strings("present_keys", present),
			zap.String("using", chosen.MatchedKey))
	}

	logger.Info("resolveResultSpec: resolved result contract",
		zap.String("mode", chosen.Mode.String()),
		zap.String("matched_key", chosen.MatchedKey),
		zap.Int("field_count", len(chosen.Fields)),
		zap.Int("mapping_count", len(chosen.Mapping)))

	return *chosen
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

func toStringMap(v interface{}) (map[string]string, bool) {
	raw, ok := v.(map[string]interface{})
	if !ok || len(raw) == 0 {
		return nil, false
	}
	out := make(map[string]string, len(raw))
	for k, val := range raw {
		if s, ok := val.(string); ok {
			out[k] = s
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func setIfAbsent(m map[string]interface{}, key string, value interface{}) {
	if _, exists := m[key]; !exists {
		m[key] = value
	}
}
