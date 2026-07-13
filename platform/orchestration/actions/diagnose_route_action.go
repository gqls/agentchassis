// FILE: platform/orchestration/actions/diagnose_route_action.go
//
// DRAFT for the agent-chassis repo. Does NOT compile in the contextkit
// container — built in your env. Grounded in REAL contracts:
//   - ExecuteLLMPromptAction (ai_actions.go) returns {"result": <parsed JSON>},
//     so the verdict step's output is read at <verdict_field>.result.* — which is
//     exactly the VerdictWire shape (outcome/citations/next_scope/...).
//   - The coordinator (coordinator.go getNextStepFromResult) honours a "next_step"
//     key in an action's result map to OVERRIDE the workflow's next step — the
//     same mechanism conditional_route uses. That is how this action drives the
//     loop: it returns next_step = the gather step (iterate) or the emit step (stop).
//   - The loop guards + call-graph re-scope are the ALREADY-TESTED pkg/diagnose
//     engine (15 tests). This action is a thin chassis adapter over them.
//
// DESIGN DECISION (this turn): the VERDICT is its OWN execute_llm_prompt workflow
// step (observable, single responsibility), NOT called from inside a monolithic
// diagnose_run. This action — diagnose_route — is the per-iteration CONTROLLER
// that runs AFTER the verdict step: it applies the guards, decides continue/stop,
// computes the next scope by following the call graph, and routes the workflow.
// So the loop is expressed in the WORKFLOW (gather → verdict → route → back to
// gather | emit), with the guard/re-scope COMPLEXITY kept in Go (the guideline).
//
// READ-ONLY decision logic over collected_data, PLUS (§7D, 2026-07-02) the
// evidence-fed scope resolver: a read-only vector search over code_symbols and
// one embedding HTTP call per FUZZY next_scope entry. Still no writes, no LLM
// verdicts, no spawn — but the original "no DB" claim no longer holds.

package actions

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/diagnose"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnoseRouteInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"verdict_field", "state_field", "analysis_field",
		"gather_step", "emit_step", "max_iterations",
		"seed_hypothesis_field", "seed_scope_field", "code_results_field",
		"resolver_top_k", "resolver_min_similarity", "max_expanded_scope",
	},
	Defaults: map[string]interface{}{
		"verdict_field":  "verdict.result",       // ExecuteLLMPromptAction wraps JSON under .result
		"state_field":    "route.diagnose_state", // this action's result lands under output_field "route", so it reads its PRIOR LoopState back at route.diagnose_state — a bare "diagnose_state" never exists at top level, so the loop would re-seed every iteration (cap unenforced, trail truncated, guard memory reset). Must track the route step's output_field.
		"analysis_field": "repo_analysis",        // analyser Output, for call-graph re-scope
		"gather_step":    "assemble_bundle",      // the step to loop BACK to for the next iteration
		"emit_step":      "emit",                 // the step to proceed to when the loop stops
		"max_iterations": 5,
		// First-iteration SEED (no diagnose_state yet). These mirror the fields
		// diagnose_assemble_bundle seeds from, so the loop's initial hypothesis +
		// scope match the bundle actually assembled on iteration 1.
		"seed_hypothesis_field": "input_data.symptom",
		"seed_scope_field":      "input_data.seed_scope",
		"code_results_field":    "code_lookup.code_results",
		// §7D scope resolver. top_k default 2, NOT 3: substitution nets +K-1
		// entries per fuzzy item, and the engine's scope-narrowing guard stops at
		// prevSize+2 — K=2 keeps two fuzzy entries per verdict inside the guard.
		// 0 disables the resolver. min_similarity 0.55 is a permissive floor just
		// above the measured stale-corpus garbage band (0.547–0.574); §7E
		// calibrates it — all similarities are logged for that purpose.
		"resolver_top_k":          2,
		"resolver_min_similarity": 0.55,
		// Cap on the engine's call-graph enrichment of next_scope (named entries
		// always kept). 0 = engine default (18); <0 = unlimited.
		"max_expanded_scope": 18,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_route", DiagnoseRouteInputSpec)
}

// DiagnoseRouteAction is the loop controller. It runs once per iteration, after
// the verdict step. It returns a result map carrying:
//   - "next_step": the gather step (iterate) or the emit step (stop) — the
//     coordinator obeys this to route the workflow;
//   - "scope": the next iteration's scope (symbols/tables/runtime), which the
//     gather step reads to assemble the next bundle;
//   - "diagnose_state": the accumulated loop state (iteration, evidence trail,
//     hypothesis history) threaded into the next iteration;
//   - "data_requests": the verdict's read-only SELECTs (when iterating), which the
//     next load_runtime executes under a read-only transaction;
//   - "status"/"conclusion": when stopping, the engine's terminal result.
func DiagnoseRouteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	verdictField := datahelpers.GetStringField(config, "verdict_field", "verdict.result")
	stateField := datahelpers.GetStringField(config, "state_field", "route.diagnose_state")
	analysisField := datahelpers.GetStringField(config, "analysis_field", "repo_analysis")
	gatherStep := datahelpers.GetStringField(config, "gather_step", "assemble_bundle")
	emitStep := datahelpers.GetStringField(config, "emit_step", "emit")
	maxIter := datahelpers.GetIntField(config, "max_iterations", 5)

	// 1) Parse the verdict the LLM step produced (its JSON, under .result).
	verdictRaw := datahelpers.ExtractNestedField(params.CollectedData, verdictField)
	if verdictRaw == nil {
		return nil, fmt.Errorf("diagnose_route: no verdict at %q (expected ExecuteLLMPromptAction output under .result)", verdictField)
	}
	verdict, err := diagnose.ParseVerdictValue(verdictRaw) // wire-shape map -> domain Verdict
	if err != nil {
		return nil, fmt.Errorf("diagnose_route: parse verdict: %w", err)
	}

	// 2) Rehydrate loop state from the prior iteration, or SEED it on the FIRST
	//    iteration (when no diagnose_state exists yet). Seeding mirrors the standalone
	//    Run(): hypothesis = the symptom, scope = the SAME seed diagnose_assemble_bundle
	//    used (seed_scope → lookup code_results), via diagnose.InitLoopState — which sets
	//    PrevScopeSize = seed.size()+1 (the "first iteration always narrows" buffer) and
	//    initialises SeenCitations + Follow. WITHOUT this, st is the zero LoopState:
	//    PrevScopeSize = 0, so the scope-must-narrow guard trips on the very first
	//    re-scope (next.size() > 0+2) and the loop stops at iteration 1 with
	//    "scope-not-narrowing"; the hypothesis is also empty in the trail.
	var st diagnose.LoopState
	if prev := datahelpers.ExtractNestedField(params.CollectedData, stateField); prev != nil {
		if err := diagnose.DecodeLoopState(prev, &st); err != nil {
			return nil, fmt.Errorf("diagnose_route: decode loop state: %w", err)
		}
	} else {
		// FAILSAFE (defence in depth): we are about to SEED because no prior state was
		// found at state_field. On a genuine first iteration that is correct. But if the
		// route step has ALREADY produced output (route.diagnose_state exists) and we
		// still found nothing, the state did not THREAD — state_field is misconfigured
		// relative to the route step's output_field (the exact bug that silently re-seeds
		// every iteration, disarming the iteration cap and the guards, leaving only the
		// engine timeout to catch a runaway). Fail LOUD and stop to emit rather than spin.
		// (Best-effort: assumes the route step's output_field is "route"; a false negative
		// just falls back to the engine timeout, and it never false-positives — it only
		// fires when route.diagnose_state actually exists.)
		if datahelpers.ExtractNestedField(params.CollectedData, "route.diagnose_state") != nil {
			logger.Error("diagnose_route: loop state did not thread — route has run before but no state at state_field; state_field must match the route step's output_field (expected route.diagnose_state). Stopping to avoid a runaway re-seed loop.",
				zap.String("state_field", stateField))
			return map[string]interface{}{
				"next_step":  emitStep,
				"status":     diagnose.Unverifiable.String(),
				"stopped_by": "state-threading-error",
				"conclusion": "Diagnosis aborted: loop state failed to thread between iterations (state_field misconfigured relative to the route step's output_field). Stopped to avoid a runaway; no diagnosis. Fix: set the route step's state_field to '<route output_field>.diagnose_state'.",
			}, nil
		}

		seedHyp := datahelpers.ExtractNestedFieldString(params.CollectedData,
			datahelpers.GetStringField(config, "seed_hypothesis_field", "input_data.symptom"))
		seedScope := seedScopeForRoute(params.CollectedData, config)
		// follow the call graph on re-scope, matching DefaultConfig().FollowCallGraph.
		st = diagnose.InitLoopState(seedHyp, seedScope, maxIter, true)
		logger.Info("diagnose_route: seeded loop state (first iteration)",
			zap.String("seed_hypothesis", seedHyp),
			zap.Int("seed_scope_size", len(seedScope.Symbols)),
			zap.Int("prev_scope_size", st.PrevScopeSize))
	}
	if st.MaxIterations == 0 {
		st.MaxIterations = maxIter
	}

	// 3) Call-graph for re-scope (follow the evidence, not the symptom — §1a).
	//    analysisRaw is hoisted (same name, wider scope — no rename) because the
	//    §7D resolver below reads the same value.
	analysisRaw := datahelpers.ExtractNestedField(params.CollectedData, analysisField)
	var cg diagnose.CallGraph
	if analysisRaw != nil {
		if g, err := diagnose.NewCallGraphFromValue(analysisRaw); err == nil {
			cg = g
		} else {
			logger.Warn("diagnose_route: could not build call graph; re-scope will use verdict next_scope verbatim", zap.Error(err))
		}
	}

	// 3.5) §7D: evidence-fed scope resolver — translate the verdict's FUZZY
	// next_scope entries (natural-language descriptions of code) into real
	// path:Symbol handles by searching code_symbols with the ENTRY TEXT as the
	// query. Runs BEFORE Advance so the call-graph expansion and the narrowing
	// guard operate on real symbols. Exact entries pass through; entries that
	// resolve to nothing survive as labels (previous behaviour, fail-open).
	// DELIBERATE CHANGE (RUNBOOK §7D): mutates verdict.NextScope, so the
	// evidence trail records the RESOLVED scope — the more auditable record.
	if analysisRaw != nil {
		knownFiles, knownSyms := knownScopeIdentities(analysisRaw)
		resolveFuzzyNextScope(ctx, params, config, &verdict, knownFiles, knownSyms, logger)
	}

	// 4) ONE engine step: apply guards, decide continue/stop, compute next scope.
	//    This is the tested pkg/diagnose logic, exposed as a single advance.
	st.MaxExpandedScope = datahelpers.GetIntField(config, "max_expanded_scope", 18)
	decision := diagnose.Advance(&st, verdict, cg)

	// 5) Translate the engine decision into a workflow route + carry state/scope.
	result := map[string]interface{}{
		"diagnose_state": diagnose.EncodeLoopState(&st),
		"iteration":      st.Iteration,
	}

	if decision.Stop {
		// Loop ends — go to emit; carry the terminal result for emit/complete.
		result["next_step"] = emitStep
		result["status"] = decision.Status.String()
		result["conclusion"] = decision.Conclusion
		result["stopped_by"] = decision.StoppedBy
		result["evidence_trail"] = diagnose.EncodeTrail(st.Trail)
		logger.Info("diagnose_route: loop stopping",
			zap.String("status", decision.Status.String()),
			zap.String("stopped_by", decision.StoppedBy),
			zap.Int("iterations", st.Iteration))
		return result, nil
	}

	// Iterate — loop back to the gather step with the narrowed scope.
	result["next_step"] = gatherStep
	result["scope"] = diagnose.EncodeScope(decision.NextScope)
	result["hypothesis"] = decision.Hypothesis // the (possibly revised) hypothesis for the next bundle

	// Forward the verdict's read-only data_requests so the next gather (load_runtime)
	// runs them under a read-only transaction and folds the rows into the next bundle.
	// Read them from the RAW verdict wire value (verdict.result) rather than a typed
	// Verdict field — the data_requests are model-supplied wire data, and reading the
	// map directly keeps this independent of the engine's Verdict shape (the chassis
	// pkg/diagnose Verdict does not carry them). Keep ONLY the read-only ones
	// (diagnose.IsReadOnlySQL) — the route-layer filter; load_runtime re-lints (defence
	// in depth) and the read-only transaction is the real backstop. Code re-scope (the
	// call graph, carried in Scope) and data re-gather (these) are separate channels.
	dataReqs, droppedDR := readOnlyDataRequestsFromWire(verdictRaw)
	// F0.5: forward the CUMULATIVE request set, not just this verdict's.
	// Benchmark run 5120c0dc (2026-07-10): answers were one-shot — they rode only
	// the bundle immediately after the requesting verdict, so a guard-refused
	// confirm LOST the fetched evidence; the loop re-requested near-identical SQL
	// and tripped scope-not-narrowing. The engine already accumulates every issued
	// request in LoopState.SeenRequests (keyed by trimmed SQL, for the spin
	// guard); re-forwarding those keys makes load_runtime re-run them every
	// iteration, so answered evidence PERSISTS for the price of a few capped
	// SELECTs. The spin guard is unaffected — it judges what the MODEL issues,
	// never what the route forwards.
	dataReqs = withPriorRequests(dataReqs, st.SeenRequests, maxForwardedDataRequests)
	if len(dataReqs) > 0 {
		result["data_requests"] = dataReqs
	}
	if droppedDR > 0 {
		logger.Warn("diagnose_route: dropped non-read-only data_requests at route", zap.Int("dropped", droppedDR))
	}

	logger.Info("diagnose_route: iterating",
		zap.Int("next_iteration", st.Iteration+1),
		zap.Int("scope_size", len(decision.NextScope.Symbols)),
		zap.Int("data_requests", len(dataReqs)),
		zap.String("next_step", gatherStep))
	return result, nil
}

// seedScopeForRoute builds the FIRST iteration's loop scope from the SAME chain
// diagnose_assemble_bundle seeds from — (1) the caller's seed_scope, else (2)
// lookup_code_symbols' code_results — so the loop's PrevScopeSize matches the scope
// actually bundled on iteration 1. REUSES scopeFromCodeResults (the assemble action's
// code_results→"path:Symbol" helper, same package) rather than re-implementing it. An
// empty seed is fine: InitLoopState still sets PrevScopeSize = 0+1 = 1, which keeps the
// narrowing guard from tripping on the first re-scope.
func seedScopeForRoute(collected map[string]interface{}, config map[string]interface{}) diagnose.Scope {
	seedScopeField := datahelpers.GetStringField(config, "seed_scope_field", "input_data.seed_scope")
	syms := datahelpers.ExtractStringListHelper(datahelpers.ExtractNestedField(collected, seedScopeField))
	if len(syms) == 0 {
		crField := datahelpers.GetStringField(config, "code_results_field", "code_lookup.code_results")
		syms = scopeFromCodeResults(collected, crField)
	}
	return diagnose.Scope{Symbols: syms}
}

// maxForwardedDataRequests caps route.data_requests (the current verdict's plus
// re-forwarded prior ones). The strings are small SQL (~hundreds of bytes) and
// each re-run is bounded by load_runtime's per-request row/cost caps, so this
// bound is mostly about keeping bundle noise down, not safety.
const maxForwardedDataRequests = 12

// withPriorRequests appends prior iterations' requests (LoopState.SeenRequests
// keys — trimmed SQL, maintained by the engine's spin guard) after the current
// verdict's, deduped, sorted for determinism, capped, and re-linted read-only.
// The state round-trips through collected_data, so treat its keys as data:
// anything failing the read-only lint is skipped, and load_runtime's read-only
// transaction remains the real guarantee.
func withPriorRequests(current []interface{}, seen map[string]bool, max int) []interface{} {
	have := map[string]bool{}
	out := make([]interface{}, 0, len(current))
	for _, it := range current {
		if m, ok := it.(map[string]interface{}); ok {
			if s, _ := m["sql"].(string); s != "" {
				have[strings.TrimSpace(s)] = true
			}
		}
		out = append(out, it)
		if len(out) >= max {
			return out
		}
	}
	prior := make([]string, 0, len(seen))
	for k := range seen {
		if strings.TrimSpace(k) == "" || have[k] || diagnose.IsReadOnlySQL(k) != nil {
			continue
		}
		prior = append(prior, k)
	}
	sort.Strings(prior)
	for _, k := range prior {
		if len(out) >= max {
			break
		}
		out = append(out, map[string]interface{}{
			"sql": k,
			"why": "re-run of a prior iteration's request so its answer persists across iterations (F0.5)",
		})
	}
	return out
}

// readOnlyDataRequestsFromWire pulls the verdict's data_requests out of the RAW
// verdict wire value (the parsed LLM JSON at verdict.result) as a []{sql,why} list,
// keeping ONLY the read-only ones (diagnose.IsReadOnlySQL) and counting the rest as
// dropped. Reading the wire map directly (not a typed Verdict field) keeps this
// independent of the engine's Verdict shape; load_runtime re-lints and runs each
// survivor under a read-only transaction.
func readOnlyDataRequestsFromWire(verdictRaw interface{}) (kept []interface{}, dropped int) {
	m, ok := verdictRaw.(map[string]interface{})
	if !ok {
		return nil, 0
	}
	arr, ok := m["data_requests"].([]interface{})
	if !ok {
		return nil, 0
	}
	for _, it := range arr {
		dm, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		sqlText, _ := dm["sql"].(string)
		why, _ := dm["why"].(string)
		if strings.TrimSpace(sqlText) == "" {
			continue
		}
		if diagnose.IsReadOnlySQL(sqlText) != nil {
			dropped++
			continue
		}
		kept = append(kept, map[string]interface{}{"sql": sqlText, "why": why})
	}
	return kept, dropped
}

// ── §7D: evidence-fed next_scope resolver ────────────────────────────────────

// knownScopeIdentities builds the exactness sets the resolver checks against:
// bare file paths and "path:Name" handles (functions AND types) — the SAME
// identity scopeFromCodeResults emits and diagnose_assemble_bundle slices by.
// Built here from the repo_analysis value rather than extending the engine's
// AnalysisCallGraph with a presence method: the graph's callsBySym covers only
// functions (types define no calls), and keeping the engine untouched avoids a
// second engine-file copy/deploy for a read-only classification concern.
func knownScopeIdentities(analysisRaw interface{}) (files map[string]bool, syms map[string]bool) {
	files, syms = map[string]bool{}, map[string]bool{}
	m, ok := analysisRaw.(map[string]interface{})
	if !ok {
		return
	}
	arr, ok := m["files"].([]interface{})
	if !ok {
		return
	}
	for _, fi := range arr {
		fm, ok := fi.(map[string]interface{})
		if !ok {
			continue
		}
		path, _ := fm["path"].(string)
		if path == "" {
			continue
		}
		files[path] = true
		for _, listKey := range []string{"functions", "types"} {
			items, _ := fm[listKey].([]interface{})
			for _, it := range items {
				im, ok := it.(map[string]interface{})
				if !ok {
					continue
				}
				if name, _ := im["name"].(string); name != "" {
					syms[path+":"+name] = true
				}
			}
		}
	}
	return
}

// resolveFuzzyNextScope applies the resolver to the verdict IN PLACE (flagged
// deliberate mutation — the trail then records the resolved scope). A cited
// Confirmed stops the loop in DecideStep, so no embeddings are spent on it; a
// CITATION-LESS Confirmed is coerced to Unverifiable (item 24) and continues,
// so it IS resolved.
func resolveFuzzyNextScope(ctx context.Context, params ActionParams, config map[string]interface{}, v *diagnose.Verdict, knownFiles, knownSyms map[string]bool, logger *zap.Logger) {
	topK := datahelpers.GetIntField(config, "resolver_top_k", 2)
	if topK <= 0 || len(v.NextScope) == 0 {
		return
	}
	if v.Outcome == diagnose.Confirmed && len(v.Citations) > 0 {
		return
	}
	minSim := configFloatField(config, "resolver_min_similarity", 0.55)
	repo := resolveCodeRepoLabel(config, params.CollectedData)
	search := buildScopeResolverSearch(ctx, params, config, repo, topK, logger)
	if search == nil {
		return
	}
	before := len(v.NextScope)
	v.NextScope = resolveScopeEntries(v.NextScope, knownFiles, knownSyms, search, minSim, logger)
	logger.Info("diagnose_route: scope resolver applied",
		zap.Int("entries_in", before),
		zap.Int("entries_out", len(v.NextScope)),
		zap.Float64("min_similarity", minSim),
		zap.Int("resolver_top_k", topK))
}

// buildScopeResolverSearch wires the REUSED retrieval chain — the same repo
// label, embedding client, nomic search_query prefix, vector search, and
// trigram fallback lookup_code_symbols uses. Returns nil (resolver disabled)
// only when no DB handle is available; embedding failures fall back to trigram
// PER ENTRY, mirroring the lookup.
func buildScopeResolverSearch(ctx context.Context, params ActionParams, config map[string]interface{}, repo string, topK int, logger *zap.Logger) func(string) ([]map[string]interface{}, error) {
	if params.DB == nil {
		logger.Warn("diagnose_route: scope resolver disabled — no DB handle on this step")
		return nil
	}
	embClient, embErr := createRAGEmbeddingClient(ctx, config)
	if embErr != nil {
		logger.Warn("diagnose_route: resolver embedding client unavailable — trigram fallback for all entries", zap.Error(embErr))
	}
	return func(entry string) ([]map[string]interface{}, error) {
		if embErr == nil {
			promptText, _ := applyNomicPrefix(config, entry, "search_query")
			if emb, genErr := embClient.GenerateEmbedding(ctx, promptText); genErr == nil {
				return vectorSearchCodeSymbols(ctx, params.DB, repo, emb, topK)
			} else {
				logger.Warn("diagnose_route: resolver embedding failed — trigram fallback for this entry", zap.Error(genErr))
			}
		}
		return trigramSearchCodeSymbols(ctx, params.DB, repo, entry, topK)
	}
}

// resolveScopeEntries is the pure substitution core (tested directly).
// Exact entries (known file or known path:Name) pass through untouched. Fuzzy
// entries are replaced by their hits at/above the similarity floor; rows
// WITHOUT a similarity key (the trigram fallback) are accepted as lexical
// matches. A fuzzy entry with no surviving hits — or a search error — stays as
// a label: the previous behaviour, so the failure mode is "no worse".
func resolveScopeEntries(entries []string, knownFiles, knownSyms map[string]bool, search func(string) ([]map[string]interface{}, error), minSim float64, logger *zap.Logger) []string {
	seen := map[string]bool{}
	var out []string
	add := func(sym string) {
		if sym != "" && !seen[sym] {
			seen[sym] = true
			out = append(out, sym)
		}
	}
	for _, e := range entries {
		t := strings.TrimSpace(e)
		if t == "" {
			continue
		}
		if knownSyms[t] || knownFiles[t] {
			add(t) // exact — the engine/assembler identity
			continue
		}
		hits, err := search(t)
		if err != nil {
			logger.Warn("diagnose_route: resolver search failed — keeping entry as label",
				zap.String("entry", truncateForLog(t, 100)), zap.Error(err))
			add(t)
			continue
		}
		resolved := 0
		for _, h := range hits {
			path, _ := h["path"].(string)
			symbol, _ := h["symbol"].(string)
			if path == "" || symbol == "" {
				continue
			}
			if sim, hasSim := h["similarity"].(float64); hasSim {
				if sim < minSim {
					logger.Info("diagnose_route: resolver hit below similarity floor",
						zap.String("entry", truncateForLog(t, 100)),
						zap.String("symbol", path+":"+symbol),
						zap.Float64("similarity", sim))
					continue
				}
				logger.Info("diagnose_route: resolved fuzzy scope entry",
					zap.String("entry", truncateForLog(t, 100)),
					zap.String("symbol", path+":"+symbol),
					zap.Float64("similarity", sim))
			} else {
				logger.Info("diagnose_route: resolved fuzzy scope entry (trigram)",
					zap.String("entry", truncateForLog(t, 100)),
					zap.String("symbol", path+":"+symbol))
			}
			add(path + ":" + symbol)
			resolved++
		}
		if resolved == 0 {
			add(t)
		}
	}
	return out
}

// configFloatField reads a float config value (jsonb numbers arrive as
// float64; ints tolerated), returning def when absent or mistyped.
// PRE-MERGE (dev guide: grep before adding helpers): if datahelpers has grown a
// GetFloatField, use it instead:
//
//	grep -rn "func GetFloatField" platform/orchestration/datahelpers/
func configFloatField(config map[string]interface{}, key string, def float64) float64 {
	switch v := config[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return def
}
