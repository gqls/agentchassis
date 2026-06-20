// FILE: platform/orchestration/actions/diagnose_assemble_bundle_action.go
//
// DRAFT for the agent-chassis repo. Does NOT compile in the contextkit
// container — built in your env. Sketched against the REAL signatures of
// request_repo_analysis (analyser_request_action.go), lookup_code_symbols and
// index_code_symbols (code_symbols_actions.go).
//
// ── What the STEP-ZERO search established ────────────────────────────────────
// The chassis "gather" is ALREADY a sequence of existing actions. Only ONE new
// action is needed — the code-context ASSEMBLER — because:
//   - request_repo_analysis  -> analyses the repo at ref, awaits the adapter,
//                               leaves the analyser Output under output_field
//                               (e.g. "repo_analysis"), incl. commit_sha.
//   - lookup_code_symbols    -> vector/trigram search over code_symbols, returns
//                               {code_results:[{path,symbol,signature,...}],
//                                code_context:"<path::symbol + signatures>",
//                                result_count, search_method}. Bodies are NOT in
//                               code_symbols — the comment is explicit: "the
//                               assembler reads them from the repo at commit_sha".
//   - load_site_for_rebuild  -> existing read-only DB context (brief, nav, pages,
//                               brand) when the bug is site-scoped.
//
// So the assembler's job: take lookup's code_results (the WHICH — path+symbol),
// read the bodies from the repo at the analysis commit_sha (the WHAT), and
// compose the bundle text the verdict prompt reads. This is the standalone
// cmd/assembler's role, as a chassis action — the one genuinely new piece.
//
// READ-ONLY: reads code_symbols (already done upstream), reads file bodies from
// the checked-out repo at a commit, reads the analyser Output already in
// collected_data. Writes nothing; triggers nothing.
//
// ── Where runtime_evidence comes from (the §1a re-scope driver) ──────────────
// This action READS runtime_evidence from collected_data; a SIBLING read-only
// action puts it there one step earlier. That sibling follows the exact pattern
// in maintenance_actions.go: take params.DB and run db.QueryContext against the
// runtime tables — for diagnosis that means agent_error_log (the failing step +
// error_message), site_work_items (status / completed-but-no-op), and
// orchestration_states (current_step / error) for the correlation/site in
// input_data. It is a NEW thin read action (e.g. diagnose_load_runtime), modelled
// on LoadSiteForRebuildAction / ScanSitesForMaintenanceAction — same params.DB +
// QueryContext shape, read-only, returns the rows as text under "runtime_evidence".
// (Not sketched here; it is a straight DB-read in the maintenance_actions mould.)

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// DiagnoseAssembleBundleInputSpec declares the action's contract (dev guide §3).
// Field names are PREFIXED/specific (scope, repo_root, runtime_evidence) to avoid
// the nested-lookup collisions the guide warns about — none of them are
// content_data/status/domain/site_id, which resolve via site_record.*/input_data.*
// even when unset. Registered in init() below.
var DiagnoseAssembleBundleInputSpec = datahelpers.ActionInputSpec{
	// Nothing is hard-required: the action resolves scope from a FALLBACK CHAIN
	// (loop scope written by diagnose_route → seed scope → lookup's code_results)
	// and validates at runtime with a clear error if none is present. The chain is
	// the doc-003 "action-level defense" pattern: the same action serves the FIRST
	// iteration (seed/code_results) and every LOOP-BACK iteration (route.scope),
	// because diagnose_route loops the workflow back to this step with a new scope.
	Optional: []string{
		"loop_scope_field", "scope_field", "code_results_field",
		"hypothesis_field", "seed_hypothesis_field",
		"analysis_field", "repo_root_field", "runtime_field", "max_body_chars",
	},
	Defaults: map[string]interface{}{
		"loop_scope_field":      "route.scope",          // set by diagnose_route on loop-back
		"scope_field":           "input_data.seed_scope", // first iteration, if caller gave one
		"code_results_field":    "code_lookup.code_results", // first iteration fallback
		"hypothesis_field":      "route.hypothesis",      // revised hypothesis on loop-back
		"seed_hypothesis_field": "input_data.symptom",    // first iteration hypothesis
		"analysis_field":        "repo_analysis",
		"repo_root_field":       "repo_analysis.root",
		"runtime_field":         "runtime.runtime_evidence",
		"max_body_chars":        60000,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_assemble_bundle", DiagnoseAssembleBundleInputSpec)
}

// DiagnoseAssembleBundleAction composes the diagnosis bundle from:
//   - the in-scope symbols (config.scope_field, default "scope" in collected_data,
//     OR lookup's code_results via config.code_results_field) — the WHICH;
//   - their bodies, read from the repo at the analyser commit_sha — the WHAT;
//   - the runtime/DB evidence already gathered into collected_data (passed through
//     by reference so the verdict sees it).
//
// Config keys (all optional, with defaults):
//   scope_field          path to the loop's current scope []string  (default "scope")
//   code_results_field   path to lookup_code_symbols' code_results   (default "code_lookup.code_results")
//   analysis_field       path to the analyser Output                 (default "repo_analysis")
//   repo_root_field      path to the checked-out repo root on disk    (default "repo_analysis.root")
//   runtime_field        path to runtime evidence text                (default "runtime_evidence")
//   max_body_chars       cap on total body text                       (default 60000)
func DiagnoseAssembleBundleAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	scopeField := datahelpers.GetStringField(config, "scope_field", "input_data.seed_scope")
	analysisField := datahelpers.GetStringField(config, "analysis_field", "repo_analysis")
	repoRootField := datahelpers.GetStringField(config, "repo_root_field", "repo_analysis.root")
	runtimeField := datahelpers.GetStringField(config, "runtime_field", "runtime.runtime_evidence")
	maxBodyChars := datahelpers.GetIntField(config, "max_body_chars", 60000)

	// Scope FALLBACK CHAIN (doc-003 action-level defense): (1) loop scope from
	// diagnose_route on a loop-back iteration; (2) the caller's seed scope on the
	// first iteration; (3) lookup_code_symbols' code_results on the first
	// iteration when no seed was given. First non-empty wins.
	loopScopeField := datahelpers.GetStringField(config, "loop_scope_field", "route.scope")
	scope := datahelpers.ExtractStringSlice(params.CollectedData, loopScopeField)
	if len(scope) == 0 {
		scope = datahelpers.ExtractStringSlice(params.CollectedData, scopeField)
	}
	if len(scope) == 0 {
		crField := datahelpers.GetStringField(config, "code_results_field", "code_lookup.code_results")
		scope = scopeFromCodeResults(params.CollectedData, crField)
	}
	if len(scope) == 0 {
		return nil, fmt.Errorf("diagnose_assemble_bundle: no scope (tried %q, %q, then code_results)",
			loopScopeField, scopeField)
	}

	// Current hypothesis: the revised one on a loop-back, else the seed symptom.
	// It goes at the TOP of the bundle so the verdict step reads a SELF-CONTAINED
	// bundle (hypothesis + code + runtime) and needs no other field threaded in.
	hypothesis := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "hypothesis_field", "route.hypothesis"))
	if hypothesis == "" {
		hypothesis = datahelpers.ExtractNestedFieldString(params.CollectedData,
			datahelpers.GetStringField(config, "seed_hypothesis_field", "input_data.symptom"))
	}

	repoRoot := datahelpers.ExtractNestedFieldString(params.CollectedData, repoRootField)
	if repoRoot == "" {
		return nil, fmt.Errorf("diagnose_assemble_bundle: repo root not found at %q (need the analysed checkout to read bodies)", repoRootField)
	}

	// Compose: hypothesis first (so the verdict judges THIS hypothesis), then the
	// in-scope code bodies, then the runtime/DB evidence.
	var b strings.Builder
	if hypothesis != "" {
		fmt.Fprintf(&b, "## Hypothesis under test\n\n%s\n\n", hypothesis)
	}
	b.WriteString("## In-scope code\n\n")
	total, truncated := 0, false
	included := 0
	for _, sym := range scope {
		body, err := readSymbolBody(repoRoot, sym) // reads file at path, slices the symbol's lines
		if err != nil {
			logger.Warn("diagnose_assemble_bundle: could not read body", zap.String("symbol", sym), zap.Error(err))
			continue
		}
		if total+len(body) > maxBodyChars {
			truncated = true
			break
		}
		fmt.Fprintf(&b, "### %s\n```go\n%s\n```\n\n", sym, body)
		total += len(body)
		included++
	}

	// Runtime/DB evidence already gathered upstream — include verbatim so the
	// verdict sees the logs/rows that name the layer (the §1a re-scope driver).
	if rt := datahelpers.ExtractNestedFieldString(params.CollectedData, runtimeField); rt != "" {
		b.WriteString("## Runtime / DB evidence\n\n")
		b.WriteString(rt)
		b.WriteString("\n")
	}

	// The analyser Output stays in collected_data for the engine's call-graph
	// re-scope (it reads `calls`); we don't duplicate it into the bundle text.
	_ = analysisField

	logger.Info("diagnose_assemble_bundle: composed",
		zap.Int("symbols_in_scope", len(scope)),
		zap.Int("symbols_included", included),
		zap.Int("body_chars", total),
		zap.Bool("truncated", truncated))

	// Return a map, matching the codebase convention (lookup_code_symbols,
	// index_code_symbols, the maintenance actions all return map[string]interface{}
	// rather than a bespoke exported result struct — dev guide over-abstraction rule).
	return map[string]interface{}{
		"bundle":       b.String(),
		"symbol_count": included,
		"truncated":    truncated,
	}, nil
}

// scopeFromCodeResults turns lookup_code_symbols' code_results ([]{path,symbol})
// into "path:Symbol" scope entries.
func scopeFromCodeResults(collected map[string]interface{}, field string) []string {
	raw := datahelpers.ExtractNestedField(collected, field)
	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var out []string
	for _, it := range arr {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		path, _ := m["path"].(string)
		symbol, _ := m["symbol"].(string)
		if path != "" && symbol != "" {
			out = append(out, path+":"+symbol)
		} else if path != "" {
			out = append(out, path)
		}
	}
	return out
}

// readSymbolBody reads the body of "path:Symbol" (or whole file for "path") from
// the repo root. Implementation note for your env: parse the file with go/ast (or
// reuse the analyser's funcDef line spans, already in repo_analysis) to slice the
// symbol's start_line..end_line — the analyser Output ALREADY has these spans, so
// prefer reading them from collected_data over re-parsing. Left as a stub here.
func readSymbolBody(repoRoot, symbol string) (string, error) {
	return "", fmt.Errorf("readSymbolBody not implemented in the sketch; wire to the analyser Output's start_line/end_line spans for %q under %s", symbol, repoRoot)
}
