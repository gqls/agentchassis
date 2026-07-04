// FILE: platform/orchestration/datahelpers/result_contract.go
//
// THE result contract — lifted from platform/orchestration/result_spec.go
// (2026-07-03) so that BOTH readers of a `complete` step's config share ONE
// table. History: the coordinator's resolveResultSpec and the action's
// extractFinalResult each had their own key vocabulary with partial overlap;
// a config speaking only the coordinator's preferred `result_from` fell to the
// action's ship-everything fallback and produced a 1.27MB Kafka rejection
// (run 17933a83), and the imagined-key variant produced silent dump fallbacks
// (run 73ed55c6 parent). One table, two callers, zero drift.
//
// Log message prefixes deliberately retain "resolveResultSpec:" so existing
// log greps/dashboards keep matching after the lift.
//
// PRE-MERGE (dev guide): grep this package for pre-existing toStringMap before
// merging — the inventory at lift time (CleanDataMap / GetMapKeys /
// ExtractNestedField* / ExtractStepData / ToStringSlice) had no equivalent:
//
//	grep -rn "func toStringMap" platform/orchestration/datahelpers/
package datahelpers

import (
	"strings"

	"go.uber.org/zap"
)

// ResultMode is how a completed workflow's payload is shaped before return.
// Exactly one mode is selected per workflow by ResolveResultSpec.
type ResultMode int

const (
	// ResultModeFallback: no contract declared -> legacy filtered dump.
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

// ResolveResultSpec inspects the `complete` step config and returns the
// normalised contract. Precedence is the candidates order below; when (against
// guidance) more than one key is present the highest-precedence one wins and
// the conflict is logged. Deprecated keys are accepted with a migration Warn.
// No logger.Debug — Debug does not surface in our logs.
func ResolveResultSpec(completeConfig map[string]interface{}, logger *zap.Logger) ResultSpec {
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
			f := ToStringSlice(raw)
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
			f := ToStringSlice(raw)
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

// ApplyResultSpec materialises a ResultSpec against collected data, for the
// RESPONSE-BUILDING caller (CompleteWorkflowAction). Behaviour preserves the
// old extractFinalResult exactly where contracts are absent or unmatched: the
// legacy process/aggregate_results probes, then the filtered dump — but the
// dump now WARNS, because a silent multi-megabyte dump is how run 17933a83
// died. (The coordinator's own apply, incl. its skipPatterns dump, is
// unchanged and remains in coordinator.go/result_spec.go.)
func ApplyResultSpec(spec ResultSpec, collectedData map[string]interface{}, logger *zap.Logger) interface{} {
	switch spec.Mode {
	case ResultModeFlatten:
		if v := ExtractNestedField(collectedData, spec.From); v != nil {
			logger.Info("ApplyResultSpec: flatten",
				zap.String("from", spec.From))
			return v
		}
		logger.Warn("ApplyResultSpec: flatten source missing; falling back",
			zap.String("from", spec.From))
	case ResultModeFields:
		result := make(map[string]interface{})
		for _, fn := range spec.Fields {
			if v := ExtractNestedField(collectedData, fn); v != nil {
				result[fn] = v
			}
		}
		if len(result) > 0 {
			logger.Info("ApplyResultSpec: fields",
				zap.Int("fields_found", len(result)),
				zap.Int("fields_named", len(spec.Fields)))
			return result
		}
		logger.Warn("ApplyResultSpec: none of the named fields present; falling back",
			zap.Strings("fields", spec.Fields))
	case ResultModeMapping:
		result := make(map[string]interface{})
		var missing []string
		for target, src := range spec.Mapping {
			if v := ExtractNestedField(collectedData, src); v != nil {
				result[target] = v
			} else {
				missing = append(missing, target+"<-"+src)
			}
		}
		if len(missing) > 0 {
			logger.Warn("ApplyResultSpec: mapping sources missing",
				zap.Strings("missing", missing))
		}
		if len(result) > 0 {
			logger.Info("ApplyResultSpec: mapping",
				zap.Int("mapped", len(result)))
			return result
		}
		logger.Warn("ApplyResultSpec: no mapping sources present; falling back")
	}

	// Fallback — legacy behaviour, preserved: common result locations first.
	if processResult, ok := collectedData["process"]; ok {
		return processResult
	}
	if aggResult, ok := collectedData["aggregate_results"]; ok {
		return aggResult
	}

	// Filtered dump — previous default, now LOUD.
	filteredData := make(map[string]interface{})
	for key, value := range collectedData {
		if !strings.HasPrefix(key, "__") && key != "agent_config" {
			filteredData[key] = value
		}
	}
	if len(filteredData) == 0 {
		logger.Warn("ApplyResultSpec: no result data found in CollectedData")
		return map[string]interface{}{"message": "workflow completed"}
	}
	logger.Warn("ApplyResultSpec: result fallback dump in use — declare a result contract (guidelines 003); shipping everything non-system",
		zap.Int("keys", len(filteredData)))
	return filteredData
}

// toStringMap converts a map[string]interface{} of string values (a
// result_mapping / output config block) into map[string]string.
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
