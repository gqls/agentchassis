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
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gqls/agentchassis/internal/analysis"
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
		"hypothesis_field", "seed_hypothesis_field", "symptom_field",
		"analysis_field", "repo_root_field", "runtime_field", "schema_field", "max_body_chars",
		"persist_bundle", "iteration_field", "site_id_field", "bundle_retention_days",
	},
	Defaults: map[string]interface{}{
		"loop_scope_field":      "route.scope",              // set by diagnose_route on loop-back
		"scope_field":           "input_data.seed_scope",    // first iteration, if caller gave one
		"code_results_field":    "code_lookup.code_results", // first iteration fallback
		"hypothesis_field":      "route.hypothesis",         // revised hypothesis on loop-back
		"seed_hypothesis_field": "input_data.symptom",       // first iteration hypothesis
		"symptom_field":         "input_data.symptom",       // F0.4a: the ANCHOR — always in the bundle
		"analysis_field":        "repo_analysis",
		"repo_root_field":       "repo_analysis.root",
		"runtime_field":         "runtime.runtime_evidence",
		"schema_field":          "runtime.schema",
		"max_body_chars":        60000,
		// Durable egress (F0.1). Bundles are ~60KB × ≤5 iterations, so they
		// expire sooner than the prose notes they sit beside; 0 keeps forever.
		"persist_bundle":        true,
		"iteration_field":       "route.diagnose_state.iteration",
		"site_id_field":         "input_data.site_id",
		"bundle_retention_days": 30,
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
//
//	scope_field          path to the loop's current scope []string  (default "scope")
//	code_results_field   path to lookup_code_symbols' code_results   (default "code_lookup.code_results")
//	analysis_field       path to the analyser Output                 (default "repo_analysis")
//	repo_root_field      path to the checked-out repo root on disk    (default "repo_analysis.root")
//	runtime_field        path to runtime evidence text                (default "runtime_evidence")
//	max_body_chars       cap on total body text                       (default 60000)
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
	schemaField := datahelpers.GetStringField(config, "schema_field", "runtime.schema")
	maxBodyChars := datahelpers.GetIntField(config, "max_body_chars", 60000)

	// Scope FALLBACK CHAIN (doc-003 action-level defense): (1) loop scope from
	// diagnose_route on a loop-back iteration; (2) the caller's seed scope on the
	// first iteration; (3) lookup_code_symbols' code_results on the first
	// iteration when no seed was given. First non-empty wins.
	loopScopeField := datahelpers.GetStringField(config, "loop_scope_field", "route.scope")
	scope := datahelpers.ExtractStringListHelper(datahelpers.ExtractNestedField(params.CollectedData, loopScopeField))
	if len(scope) == 0 {
		scope = datahelpers.ExtractStringListHelper(datahelpers.ExtractNestedField(params.CollectedData, scopeField))
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

	// The analyser Output (already in collected_data under analysisField) carries
	// the start_line/end_line spans the body reader slices from. Decode it ONCE
	// into the typed analysis.Output. collected_data normally holds the adapter's
	// decoded Output map (repo_root_field "repo_analysis.root" only resolves when
	// it IS a map), so decodeAnalysisOutput handles the map case and, defensively,
	// a raw JSON string.
	anaOut, err := decodeAnalysisOutput(datahelpers.ExtractNestedField(params.CollectedData, analysisField))
	if err != nil || len(anaOut.Files) == 0 {
		return nil, fmt.Errorf("diagnose_assemble_bundle: analyser Output empty/undecodable at %q (need its line spans to slice bodies): %v", analysisField, err)
	}

	// Compose: the ORIGINAL symptom first, then the hypothesis under test, then
	// the in-scope code bodies, then the runtime/DB evidence.
	//
	// F0.4a (benchmark run 4d43d002): once diagnose_route revises the hypothesis,
	// the verdict used to see ONLY the revision — from iteration 3 the original
	// symptom was absent from every bundle, so the loop could confirm a true but
	// symptom-irrelevant hypothesis. The anchor keeps the question in view without
	// clamping the drift (following evidence away from the symptom's wording is
	// the loop's whole re-scope mechanism).
	var b strings.Builder
	symptom := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "symptom_field", "input_data.symptom"))
	if symptom != "" && symptom != hypothesis {
		fmt.Fprintf(&b, "## Original symptom (the question this diagnosis must answer)\n\n%s\n\n", symptom)
	}
	if hypothesis != "" {
		fmt.Fprintf(&b, "## Hypothesis under test\n\n%s\n\n", hypothesis)
	}
	b.WriteString("## In-scope code\n\n")
	total, truncated := 0, false
	included := 0
	for _, sym := range scope {
		body, err := analysis.ReadSymbolBody(repoRoot, anaOut, sym) // slice the symbol's lines from the analyser spans
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

	// F0.4c: signatures of the OTHER symbols in each in-scope file. Retrieval is
	// symbol-granular and can land on the right file's wrong function (benchmark
	// run 4d43d002: isLegalPage was scoped while the cause sat in loadPagesForNav,
	// same file, never seen). The signature list gives the verdict the map to
	// name a sibling in next_scope. Parity with contextkit's Neighbourhood
	// section, which this action's port dropped.
	if sigs := siblingSignatures(anaOut, scope, siblingSigCap); sigs != "" {
		b.WriteString("## Same-file signatures (siblings of the in-scope symbols — name these in next_scope to read their bodies)\n\n")
		b.WriteString(sigs)
		b.WriteString("\n")
	}

	// Schema (live tables) — real table/column names so the verdict writes a
	// correct data_request instead of guessing a relation that does not exist
	// (the gamesdesign loop wasted iterations on a non-existent "page_sections").
	if sch := datahelpers.ExtractNestedFieldString(params.CollectedData, schemaField); sch != "" {
		b.WriteString("## Schema (live tables)\n\n```\n")
		b.WriteString(sch)
		b.WriteString("```\n\n")
	}

	rt := datahelpers.ExtractNestedFieldString(params.CollectedData, runtimeField)

	// F0.4b: workflow logic lives in agent_definitions JSON, not Go, so the
	// static tier cannot reach it (code_symbols indexes .go only). When the
	// runtime evidence names `agent/step (action)` — the agent_error_log line
	// format — inline those steps' definitions so the verdict can cite the
	// control flow the error came from. Benchmark run 4d43d002: the causal step
	// (page-build-handler/complete_error, a success-labelled complete_workflow)
	// was named verbatim in all five bundles and was unreachable as evidence.
	// Enrichment failure degrades to a log line; it must never fail assembly.
	if rt != "" && params.DB != nil {
		if steps := renderWorkflowSteps(ctx, params.DB, logger, workflowRefsFromRuntime(rt, maxWorkflowRefs)); steps != "" {
			b.WriteString("## Workflow step definitions named in the runtime evidence (from agent_definitions — citable as static tier)\n\n")
			b.WriteString(steps)
			b.WriteString("\n")
		}
	}

	// Runtime/DB evidence already gathered upstream — include verbatim so the
	// verdict sees the logs/rows that name the layer (the §1a re-scope driver).
	if rt != "" {
		b.WriteString("## Runtime / DB evidence\n\n")
		b.WriteString(rt)
		b.WriteString("\n")
	}

	// The analyser Output also stays in collected_data for the engine's call-graph
	// re-scope (it reads `calls`); we don't duplicate it into the bundle text.

	logger.Info("diagnose_assemble_bundle: composed",
		zap.Int("symbols_in_scope", len(scope)),
		zap.Int("symbols_included", included),
		zap.Int("body_chars", total),
		zap.Bool("truncated", truncated))

	bundle := b.String()

	// ── Durable egress (F0.1) ───────────────────────────────────────────────
	// Until now each iteration's bundle lived only in memory: the loop could be
	// neither audited nor replayed after the fact. Persist it here, as a
	// write-through INSIDE this action, rather than as a new workflow step —
	// the diagnose-agent workflow's emit → persist_note → complete shape is an
	// active surface owned by another workstream, and a Go-side write touches
	// none of it.
	//
	// The loop's READ-ONLY contract concerns the system under diagnosis. This
	// writes only to our own artifacts table. Even so, persistence is
	// observability and observability must never cost us a diagnosis, so every
	// failure below degrades to a warning and the action returns normally.
	if datahelpers.GetBoolField(config, "persist_bundle", true) {
		persistBundleArtifact(ctx, params, bundleArtifact{
			Body:            bundle,
			Iteration:       assembleIteration(params.CollectedData, config),
			SiteID:          datahelpers.ExtractNestedFieldString(params.CollectedData, datahelpers.GetStringField(config, "site_id_field", "input_data.site_id")),
			SymbolsInScope:  len(scope),
			SymbolsIncluded: included,
			BodyChars:       total,
			Truncated:       truncated,
			RetentionDays:   datahelpers.GetIntField(config, "bundle_retention_days", 30),
		})
	}

	// Return a map, matching the codebase convention (lookup_code_symbols,
	// index_code_symbols, the maintenance actions all return map[string]interface{}
	// rather than a bespoke exported result struct — dev guide over-abstraction rule).
	return map[string]interface{}{
		"bundle":       bundle,
		"symbol_count": included,
		"truncated":    truncated,
	}, nil
}

// bundleArtifact is the row this action writes for one iteration.
type bundleArtifact struct {
	Body            string
	Iteration       int
	SiteID          string // "" for an anchorless (code-only) diagnosis → NULL
	SymbolsInScope  int
	SymbolsIncluded int
	BodyChars       int
	Truncated       bool
	RetentionDays   int // ≤0 keeps the bundle indefinitely
}

// assembleIteration derives the 1-based iteration number of the pass currently
// being assembled.
//
// diagnose_route writes its LoopState to collected_data AFTER this step runs,
// so on pass N we observe pass N-1's state, and on the genuine first pass there
// is no state at all (0 + 1 = 1). Note the field path: diagnose_route documents
// that a bare "diagnose_state" never exists at top level — the state lands under
// that step's output_field, hence "route.diagnose_state.iteration". Reading the
// bare name would silently yield 1 on every pass and stamp every bundle as
// iteration 1, which the partial unique index would then collapse into one row.
func assembleIteration(collected map[string]interface{}, config map[string]interface{}) int {
	field := datahelpers.GetStringField(config, "iteration_field", "route.diagnose_state.iteration")
	return datahelpers.ExtractNestedFieldInt(collected, field) + 1
}

// persistBundleArtifact write-throughs one iteration's bundle to
// diagnosis_artifacts. It never returns an error: see the caller's note on why
// a failed write must not fail the diagnosis.
//
// The ON CONFLICT target names the partial unique index
// (correlation_id, iteration) WHERE kind='bundle', so a workflow step RETRY
// replaces that iteration's bundle in place instead of duplicating it. Notes
// are deliberately outside that index — per-step notes share an iteration.
func persistBundleArtifact(ctx context.Context, params ActionParams, a bundleArtifact) {
	logger := params.Logger

	if params.DB == nil {
		logger.Warn("diagnose_assemble_bundle: no DB handle, bundle not persisted (diagnosis continues)")
		return
	}
	if params.ExecutionContext == nil || params.ExecutionContext.CorrelationID == "" {
		logger.Warn("diagnose_assemble_bundle: no correlation_id, bundle not persisted (diagnosis continues)")
		return
	}

	metadata, err := json.Marshal(map[string]interface{}{
		"symbols_in_scope": a.SymbolsInScope,
		"symbol_count":     a.SymbolsIncluded,
		"body_chars":       a.BodyChars,
		"truncated":        a.Truncated,
	})
	if err != nil {
		logger.Warn("diagnose_assemble_bundle: metadata marshal failed, bundle not persisted (diagnosis continues)",
			zap.Error(err))
		return
	}

	// site_id is a uuid column; an empty string is not a valid uuid, so send a
	// true NULL. Anchorless (code-only) diagnoses legitimately have no site.
	var siteID interface{}
	if a.SiteID != "" {
		siteID = a.SiteID
	}

	const q = `
		INSERT INTO diagnosis_artifacts (
			correlation_id, orchestration_id, iteration, kind, body,
			site_id, metadata, source_agent, created_by, expires_at
		) VALUES (
			$1, $2, $3, 'bundle', $4,
			$5, $6::jsonb, $7, 'diagnose_assemble_bundle',
			CASE WHEN $8::int > 0 THEN now() + make_interval(days => $8::int) END
		)
		ON CONFLICT (correlation_id, iteration) WHERE kind = 'bundle'
		DO UPDATE SET body             = EXCLUDED.body,
		              metadata         = EXCLUDED.metadata,
		              expires_at       = EXCLUDED.expires_at,
		              orchestration_id = EXCLUDED.orchestration_id,
		              site_id          = EXCLUDED.site_id,
		              source_agent     = EXCLUDED.source_agent,
		              created_at       = now()`

	if _, err := params.DB.ExecContext(ctx, q,
		params.ExecutionContext.CorrelationID,
		nullIfEmpty(params.ExecutionContext.OrchestrationID),
		a.Iteration,
		a.Body,
		siteID,
		string(metadata),
		nullIfEmpty(params.AgentType),
		a.RetentionDays,
	); err != nil {
		logger.Warn("diagnose_assemble_bundle: bundle not persisted (diagnosis continues)",
			zap.String("correlation_id", params.ExecutionContext.CorrelationID),
			zap.Int("iteration", a.Iteration),
			zap.Error(err))
		return
	}

	logger.Info("diagnose_assemble_bundle: bundle persisted",
		zap.String("correlation_id", params.ExecutionContext.CorrelationID),
		zap.Int("iteration", a.Iteration),
		zap.Int("body_chars", a.BodyChars))
}

// F0.4b/c caps — bounded additions to a bundle whose bodies are already capped
// by max_body_chars; these sections must never crowd the evidence out.
const (
	siblingSigCap   = 6000 // chars of same-file signature listing
	workflowStepCap = 8000 // chars of inlined workflow-step JSON
	maxWorkflowRefs = 4    // distinct agent/step pairs to look up per bundle
)

// workflowRefRe matches the agent_error_log line shape
// `page-build-handler/complete_error (complete_workflow)`.
var workflowRefRe = regexp.MustCompile(`([a-z][a-z0-9-]*)/([a-z][a-z0-9_]*) \(([a-z_]+)\)`)

// workflowRefsFromRuntime extracts up to max distinct (agent_type, step) pairs
// named by the runtime evidence, in order of first appearance.
func workflowRefsFromRuntime(runtime string, max int) [][2]string {
	seen := map[string]bool{}
	var out [][2]string
	for _, m := range workflowRefRe.FindAllStringSubmatch(runtime, -1) {
		key := m[1] + "/" + m[2]
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, [2]string{m[1], m[2]})
		if len(out) >= max {
			break
		}
	}
	return out
}

// renderWorkflowSteps inlines the named steps' JSON from agent_definitions.
// Read-only; every failure degrades to a log line and an absent subsection.
func renderWorkflowSteps(ctx context.Context, db *sql.DB, logger *zap.Logger, refs [][2]string) string {
	var b strings.Builder
	total := 0
	for _, ref := range refs {
		var stepJSON sql.NullString
		err := db.QueryRowContext(ctx, `
			SELECT jsonb_pretty(default_config->'workflow'->'steps'->$2)
			FROM agent_definitions
			WHERE type = $1 AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
			ORDER BY version DESC LIMIT 1
		`, ref[0], ref[1]).Scan(&stepJSON)
		if err != nil || !stepJSON.Valid || stepJSON.String == "" || stepJSON.String == "null" {
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				logger.Warn("diagnose_assemble_bundle: workflow-step lookup failed (bundle continues without it)",
					zap.String("agent", ref[0]), zap.String("step", ref[1]), zap.Error(err))
			}
			continue
		}
		if total+len(stepJSON.String) > workflowStepCap {
			b.WriteString("_(further named steps omitted — cap reached)_\n")
			break
		}
		fmt.Fprintf(&b, "### %s / %s\n```json\n%s\n```\n\n", ref[0], ref[1], stepJSON.String)
		total += len(stepJSON.String)
	}
	return b.String()
}

// siblingSignatures renders the signatures of symbols that share a file with the
// in-scope symbols but are NOT themselves in scope. Pure over the analyser
// Output already in collected_data — no IO.
func siblingSignatures(out analysis.Output, scope []string, capChars int) string {
	inScope := map[string]map[string]bool{} // path -> symbol names in scope
	for _, s := range scope {
		path, name := s, ""
		if i := strings.LastIndex(s, ":"); i > 0 {
			path, name = s[:i], s[i+1:]
		}
		if inScope[path] == nil {
			inScope[path] = map[string]bool{}
		}
		if name != "" {
			inScope[path][name] = true
		} else {
			inScope[path]["*"] = true // whole file already included; no siblings to add
		}
	}

	var b strings.Builder
	total := 0
	for _, f := range out.Files {
		named, ok := inScope[f.Path]
		if !ok || named["*"] {
			continue
		}
		var lines []string
		for _, fn := range f.Functions {
			if !named[fn.Name] {
				lines = append(lines, fmt.Sprintf("- `%s:%s` — `%s`", f.Path, fn.Name, fn.Signature))
			}
		}
		if len(lines) == 0 {
			continue
		}
		section := fmt.Sprintf("**%s**\n%s\n\n", f.Path, strings.Join(lines, "\n"))
		if total+len(section) > capChars {
			b.WriteString("_(further files omitted — cap reached)_\n")
			break
		}
		b.WriteString(section)
		total += len(section)
	}
	return b.String()
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

// decodeAnalysisOutput coerces the analyser Output sitting in collected_data
// (normally a decoded map[string]interface{}; defensively also a raw JSON string)
// into the typed analysis.Output whose start_line/end_line spans
// analysis.ReadSymbolBody slices from. The body-reading itself lives in the
// shared analysis package, so cmd/assembler and this action slice IDENTICALLY
// (verified byte-for-byte) rather than re-implementing the convention here.
func decodeAnalysisOutput(raw interface{}) (analysis.Output, error) {
	var out analysis.Output
	if raw == nil {
		return out, fmt.Errorf("not present")
	}
	var jb []byte
	switch v := raw.(type) {
	case string:
		jb = []byte(v)
	case []byte:
		jb = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return out, err
		}
		jb = b
	}
	if err := json.Unmarshal(jb, &out); err != nil {
		return out, err
	}
	return out, nil
}
