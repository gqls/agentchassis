// FILE: platform/orchestration/actions/select_review_panel_action.go
//
// Stage-3 relevance filter for the council reviewer panel (concept register
// docs026, DESIGN_relevance_filter.md). A DETERMINISTIC (no-LLM) step that runs
// once, right after the fix plan is persisted, and decides which OPTIONAL
// reviewer seats are relevant to THIS fix — so a specialist seat only pays its
// LLM cost when the fix actually touches its domain, instead of every seat
// firing on every council decision.
//
// Deliberately GENERIC and CONFIG-DRIVEN: the seat->footprint map lives in the
// workflow SQL config, not in this code. So adding, removing, or re-scoping a
// seat is a config edit, no rebuild — and the SAME action serves both councils
// (fix-proposer and the council-gate) with different configs. This file only
// does the matching.
//
// Config:
//
//	"plan_field":   "plan_persisted"        (default) — the persisted plan object;
//	                its .files ([]string of edited paths) is the primary signal,
//	                its .plan_json (string) a secondary one.
//	"extra_text_fields": ["diagnosis_row.conclusion"]  (optional) — collected-data
//	                paths whose text is also scanned (e.g. the diagnosis names a
//	                table/action the plan's file paths don't).
//	"footprints":  { "<seat>": ["substr1","substr2", ...], ... }  — a seat fires
//	                if ANY of its patterns is a case-insensitive substring of any
//	                edited file path OR of the scanned text. Patterns are plain
//	                substrings (not regex) — simple, predictable, no escaping.
//
// Output (output_field, e.g. "panel"): { "run_<seat>": true|false, ... } plus
// "_matched": ["seat", ...] for the digest/audit. A conditional step then gates
// each optional seat on panel.run_<seat> == true; the always-on seats
// (edit-quality, guardian) get no footprint and no gate.
//
// BIAS: a seat with no configured footprint, or an empty one, defaults to
// RUN (fail-open) — a filter should never silently drop a seat because its
// config was forgotten; over-running a seat is cheap, under-running it is a
// blind spot. Only a seat with a non-empty footprint that matches NOTHING is
// skipped.
//
// Pairs with diagnose_council_decide's absent-reviewer tolerance: a skipped
// seat writes no review field, and council_decide treats an absent field as an
// abstention rather than an error. Neither half works without the other.
package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// SelectReviewPanelAction computes which optional reviewer seats are relevant.
func SelectReviewPanelAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "select_review_panel"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	planField := "plan_persisted"
	if s, ok := config["plan_field"].(string); ok && strings.TrimSpace(s) != "" {
		planField = strings.TrimSpace(s)
	}

	// The edited file paths — the primary signal. plan_persisted.files is a
	// []string the persist step already extracted from the plan's edits[].
	files := extractStringList(params.CollectedData, planField+".files")

	// Text corpus scanned as a secondary signal: the plan JSON plus any
	// configured extra fields (e.g. the diagnosis conclusion). Lower-cased once.
	var corpus strings.Builder
	if pj := datahelpers.ExtractNestedFieldString(params.CollectedData, planField+".plan_json"); pj != "" {
		corpus.WriteString(pj)
		corpus.WriteString("\n")
	}
	for _, f := range configStringSlice(config, "extra_text_fields", nil) {
		if t := datahelpers.ExtractNestedFieldString(params.CollectedData, f); t != "" {
			corpus.WriteString(t)
			corpus.WriteString("\n")
		}
	}
	corpusLower := strings.ToLower(corpus.String())

	lowerFiles := make([]string, 0, len(files))
	for _, f := range files {
		lowerFiles = append(lowerFiles, strings.ToLower(f))
	}

	footprints, _ := config["footprints"].(map[string]interface{})
	if len(footprints) == 0 {
		// No footprints configured — nothing to filter; run everything (the
		// conditionals will all see true). Log loudly: this is almost certainly
		// a wiring mistake, but fail-open is the safe direction.
		logger.Warn("select_review_panel: no footprints configured — all seats will run (fail-open)")
	}

	out := map[string]interface{}{}
	matched := []string{}
	for seat, raw := range footprints {
		patterns := toStringSlice(raw)
		var fires bool
		if len(patterns) == 0 {
			fires = true // fail-open: an empty footprint never silently drops a seat
		} else {
			fires = anyPatternMatches(patterns, lowerFiles, corpusLower)
		}
		out["run_"+seat] = fires
		if fires {
			matched = append(matched, seat)
		}
	}
	out["_matched"] = matched

	logger.Info("select_review_panel: decided",
		zap.Int("edited_files", len(files)),
		zap.Int("seats_configured", len(footprints)),
		zap.Strings("seats_firing", matched))

	return out, nil
}

// anyPatternMatches is true if any pattern is a substring of any edited file
// path, or of the scanned text corpus. All inputs are already lower-cased.
func anyPatternMatches(patterns, lowerFiles []string, corpusLower string) bool {
	for _, p := range patterns {
		pl := strings.ToLower(strings.TrimSpace(p))
		if pl == "" {
			continue
		}
		for _, f := range lowerFiles {
			if strings.Contains(f, pl) {
				return true
			}
		}
		if corpusLower != "" && strings.Contains(corpusLower, pl) {
			return true
		}
	}
	return false
}

// extractStringList pulls a []string from collected data at a dotted path,
// tolerating the []interface{} shape JSON round-trips produce.
func extractStringList(data map[string]interface{}, path string) []string {
	raw := datahelpers.ExtractNestedField(data, path)
	return toStringSlice(raw)
}

// toStringSlice coerces an interface (typically []interface{} of strings, or a
// single string) into []string.
func toStringSlice(raw interface{}) []string {
	switch v := raw.(type) {
	case nil:
		return nil
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				if s = strings.TrimSpace(s); s != "" {
					out = append(out, s)
				}
			} else if item != nil {
				out = append(out, strings.TrimSpace(fmt.Sprintf("%v", item)))
			}
		}
		return out
	case string:
		if s := strings.TrimSpace(v); s != "" {
			return []string{s}
		}
		return nil
	default:
		return nil
	}
}
