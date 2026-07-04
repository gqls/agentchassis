// result_contract_test.go — the shared complete-step contract (Option A).
package datahelpers

import (
	"testing"

	"go.uber.org/zap"
)

func TestResolvePrecedencePreferredWins(t *testing.T) {
	cfg := map[string]interface{}{
		"result_from":   "diagnosis",
		"output_fields": []interface{}{"a", "b"},
	}
	spec := ResolveResultSpec(cfg, zap.NewNop())
	if spec.Mode != ResultModeFlatten || spec.MatchedKey != "result_from" || spec.From != "diagnosis" {
		t.Fatalf("preferred result_from must win: %+v", spec)
	}
}

func TestResolveDeprecatedAliasStillResolves(t *testing.T) {
	spec := ResolveResultSpec(map[string]interface{}{
		"output_fields": []interface{}{"diagnosis"},
	}, zap.NewNop())
	if spec.Mode != ResultModeFields || spec.MatchedKey != "output_fields" || len(spec.Fields) != 1 {
		t.Fatalf("deprecated output_fields must resolve to Fields: %+v", spec)
	}
}

func TestApplyFlattenDottedPath(t *testing.T) {
	collected := map[string]interface{}{
		"call_diagnoser": map[string]interface{}{
			"diagnosis": map[string]interface{}{"status": "CONFIRMED"},
		},
	}
	v := ApplyResultSpec(ResultSpec{Mode: ResultModeFlatten, From: "call_diagnoser.diagnosis"}, collected, zap.NewNop())
	m, ok := v.(map[string]interface{})
	if !ok || m["status"] != "CONFIRMED" {
		t.Fatalf("flatten must return the named field's contents: %#v", v)
	}
}

func TestApplyFieldsSkipsMissing(t *testing.T) {
	collected := map[string]interface{}{"diagnosis": "x"}
	v := ApplyResultSpec(ResultSpec{Mode: ResultModeFields, Fields: []string{"diagnosis", "absent"}}, collected, zap.NewNop())
	m := v.(map[string]interface{})
	if m["diagnosis"] != "x" || len(m) != 1 {
		t.Fatalf("fields must nest found + skip missing: %#v", m)
	}
}

func TestApplyMapping(t *testing.T) {
	collected := map[string]interface{}{"a": map[string]interface{}{"b": 7}}
	v := ApplyResultSpec(ResultSpec{Mode: ResultModeMapping, Mapping: map[string]string{"out": "a.b"}}, collected, zap.NewNop())
	m := v.(map[string]interface{})
	if m["out"] != 7 {
		t.Fatalf("mapping must build target<-source: %#v", m)
	}
}

func TestApplyFallbackLegacyProbesAndFilteredDump(t *testing.T) {
	if v := ApplyResultSpec(ResultSpec{Mode: ResultModeFallback},
		map[string]interface{}{"process": "P", "other": 1}, zap.NewNop()); v != "P" {
		t.Fatalf("fallback must return legacy process result, got %#v", v)
	}
	v := ApplyResultSpec(ResultSpec{Mode: ResultModeFallback}, map[string]interface{}{
		"__raw_message__": 1, "agent_config": 2, "kept": 3,
	}, zap.NewNop())
	m := v.(map[string]interface{})
	if len(m) != 1 || m["kept"] != 3 {
		t.Fatalf("filtered dump must skip system keys: %#v", m)
	}
}

func TestApplyMissingContractSourceFallsBack(t *testing.T) {
	v := ApplyResultSpec(ResultSpec{Mode: ResultModeFlatten, From: "nope"},
		map[string]interface{}{"kept": true}, zap.NewNop())
	m := v.(map[string]interface{})
	if m["kept"] != true {
		t.Fatalf("missing flatten source must fall back to the dump: %#v", v)
	}
}
