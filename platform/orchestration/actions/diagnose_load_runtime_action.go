// FILE: platform/orchestration/actions/diagnose_load_runtime_action.go
//
// DRAFT for the agent-chassis repo. Does NOT compile in the contextkit
// container — built in your env. SQL written against the REAL schema (\d output
// for agent_error_log, site_work_items, orchestration_states — verified, not
// assumed). Modelled on the params.DB + QueryContext pattern in
// maintenance_actions.go (LoadSiteForRebuildAction / ScanSitesForMaintenanceAction).
//
// diagnose_load_runtime reads the runtime tier of the diagnosis bundle — the
// evidence that NAMES the failing layer the symptom words cannot reach (DESIGN
// §1a). It is the source of the "runtime_evidence" field that
// diagnose_assemble_bundle later folds into the bundle. READ-ONLY: three SELECTs,
// no writes, no triggered runs.
//
// Resolves the target from collected_data (set from the trigger's input_data):
//   site_id        (uuid)   — optional; filters all three tables
//   correlation_id (uuid)   — optional; filters orchestration_states precisely
//   domain         (text)   — optional fallback for agent_error_log
// At least one of site_id / correlation_id / domain is required.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/diagnose"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

var DiagnoseLoadRuntimeInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"site_id_field", "correlation_id_field", "domain_field",
		"error_limit", "work_item_limit", "data_requests_field",
		"code_requests_field", "max_code_checks", "code_row_cap", "code_excerpt_chars",
		"code_requests_dropped_field", "data_requests_dropped_field",
		"schema_exclude_patterns", "schema_include_patterns", "schema_full", "schema_table_cap",
		"explain_max_rows", "explain_max_cost", "row_cap", "cell_chars",
	},
	Defaults: map[string]interface{}{
		"site_id_field":        "input_data.site_id",
		"correlation_id_field": "input_data.correlation_id",
		"domain_field":         "input_data.runtime_site",
		"error_limit":          20,
		"work_item_limit":      20,
		// The prior verdict's data_requests are forwarded by diagnose_route into
		// route.data_requests, and the loop RETURNS to this step (the workflow's
		// gather_step = load_runtime), so each iteration runs them. Empty on the
		// first iteration (route has not run yet).
		"data_requests_field": "route.data_requests",
		// The code-search analogue, forwarded by diagnose_route the same way and
		// likewise empty on the first iteration (route has not run yet).
		"code_requests_field": "route.code_requests",
		"max_code_checks":     8,
		// Deliberately NOT the SQL row_cap/cell_chars above: those bound rows of a
		// model-written SELECT (200/600), while these bound index hits rendered as
		// source lines. A broad `content` pattern can match hundreds of symbols, so
		// reusing 200 here would bury the bundle's signal in near-duplicate code
		// lines (B4a). These match the council's code tier — one convention, and a
		// cap that has already been exercised on real runs.
		"code_row_cap":       40,
		"code_excerpt_chars": 400,
		// Where diagnose_route reports what its FORWARDING cap dropped. Kept in
		// step with the route's writer keys by
		// TestRouteDropFieldsStayInSyncWithLoadRuntimeDefaults — that test caught
		// these two missing from Defaults on its first run (declared Optional but
		// never defaulted, so the action's contract did not carry them even though
		// the inline fallback made it work).
		"code_requests_dropped_field": routeOutputPrefix + codeRequestsDroppedKey,
		"data_requests_dropped_field": routeOutputPrefix + dataRequestsDroppedKey,
		// Schema section: denylist so new tables appear automatically; relevance
		// include (used unless schema_full) keeps it to the build/content domain.
		"schema_exclude_patterns": []interface{}{"%backup%", "%bak%", "%archive%", "%supersede%"},
		"schema_include_patterns": []interface{}{"site%", "page%", "content%", "flow%"},
		"schema_full":             false,
		"schema_table_cap":        120,
		// data_request size guards (EXPLAIN-estimate caps + rendered-output caps).
		"explain_max_rows": 50000,
		"explain_max_cost": 0,
		"row_cap":          200,
		"cell_chars":       600,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_load_runtime", DiagnoseLoadRuntimeInputSpec)
}

// validateRouteWiring guards the name-based coupling between this reader and the
// diagnose_route step. This action reads route's forwarded requests and its
// upstream drop counts back under a prefix (data_requests_field / code_requests_field
// / *_dropped_field, all defaulting to the "route." namespace); diagnose_route writes
// them under ITS output_field. The two ends are joined only by a string, with no
// schema or compile-time tie -- a per-workflow override of the route step's
// output_field (or of this reader's *_field config) that moves one end but not the
// other makes every route.* read resolve to nothing, so a forwarded request or an
// upstream drop is silently counted as zero. That is exactly the silence the drop
// reporting exists to close (council-gate eba040a9, round 5): the spin guard credits
// a code/data request as progress on the promise its answer arrives next gather, and
// a silent drop breaks that promise with nothing in the trail.
//
// TestRouteDropFieldsStayInSyncWithLoadRuntimeDefaults guards the DEFAULT wiring;
// only the live workflow can reveal an OVERRIDE, so the check runs here, at the
// gather, against this orchestration's own persisted plan (loadOwnWorkflowSteps).
// A CONSISTENT override (both ends moved together) still names one namespace and
// passes -- only a divergence fails.
//
// Fails OPEN, never closed: a nil/empty step map or a workflow with no
// diagnose_route step means there is no coupling to verify -- return nil rather
// than invent a constraint. The caller logs every skip (council 6cdbc374,
// editquality: a silently-absent guard is the failure shape being guarded).
// loadOwnWorkflowSteps reads THIS orchestration's step map from its own persisted
// row. Deliberately a pipeline-local DB read, not a widening of ActionParams: the
// council's guardian vetoed threading the whole plan through the shared action
// contract for one pipeline's benefit (6cdbc374 — "the universal ActionParams
// contract and the coordinator's core dispatch function" are not where a
// two-sibling-step check belongs). The action already holds a DB handle and its
// own orchestration id; one indexed SELECT per gather is the entire cost.
func loadOwnWorkflowSteps(ctx context.Context, db *sql.DB, execCtx *types.ExecutionContext) (map[string]models.Step, error) {
	if execCtx == nil || execCtx.OrchestrationID == "" {
		return nil, fmt.Errorf("no orchestration id in execution context")
	}
	var raw []byte
	err := db.QueryRowContext(ctx,
		`SELECT workflow_plan->'steps' FROM orchestration_states WHERE orchestration_id = $1`,
		execCtx.OrchestrationID).Scan(&raw)
	if err != nil {
		return nil, fmt.Errorf("read own workflow_plan: %w", err)
	}
	var steps map[string]models.Step
	if err := json.Unmarshal(raw, &steps); err != nil {
		return nil, fmt.Errorf("parse workflow_plan steps: %w", err)
	}
	return steps, nil
}

func validateRouteWiring(cfg map[string]interface{}, steps map[string]models.Step) error {
	if len(steps) == 0 {
		return nil
	}
	routeOutputFields := map[string]bool{}
	for _, st := range steps {
		if st.Action == "diagnose_route" {
			routeOutputFields[st.OutputField] = true
		}
	}
	if len(routeOutputFields) == 0 {
		return nil // no route step -> nothing forwards into this reader
	}
	// The four config fields that read INTO the route step's output namespace, with
	// the defaults the InputSpec declares. Each field's leading segment must name a
	// namespace some diagnose_route step actually writes under.
	coupled := [][2]string{
		{"data_requests_field", "route.data_requests"},
		{"code_requests_field", "route.code_requests"},
		{"data_requests_dropped_field", routeOutputPrefix + dataRequestsDroppedKey},
		{"code_requests_dropped_field", routeOutputPrefix + codeRequestsDroppedKey},
	}
	for _, c := range coupled {
		field := datahelpers.GetStringField(cfg, c[0], c[1])
		prefix := field
		if i := strings.IndexByte(field, '.'); i >= 0 {
			prefix = field[:i]
		}
		if !routeOutputFields[prefix] {
			return fmt.Errorf(
				"diagnose_load_runtime: route wiring mismatch -- %s=%q reads namespace %q, but no "+
					"diagnose_route step writes under that output_field (present: %s). A forwarded "+
					"request or upstream drop would read as absent and be silently counted as zero; "+
					"align the diagnose_route step's output_field with this reader's %s.",
				c[0], field, prefix, sortedQuotedKeys(routeOutputFields), c[0])
		}
	}
	return nil
}

func sortedQuotedKeys(m map[string]bool) string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, fmt.Sprintf("%q", k))
	}
	sort.Strings(ks)
	return strings.Join(ks, ", ")
}

func DiagnoseLoadRuntimeAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("diagnose_load_runtime: no DB handle")
	}
	// Guard the name-based coupling to diagnose_route BEFORE trusting any route.*
	// read below. On a mismatch every forwarded request and upstream drop reads as
	// zero; fail loudly on the first gather rather than run the whole loop blind.
	//
	// The plan comes from THIS orchestration's own persisted row — a diagnosis-
	// pipeline-local read (council 6cdbc374, guardian veto: the first delivery
	// threaded the whole plan through ActionParams + buildActionParams, widening
	// shared load-bearing infrastructure for one pipeline's two-step contract;
	// walked back in favour of this). Skip paths are LOGGED, never silent — a
	// guard that vanishes quietly is the failure shape it exists to catch
	// (council 6cdbc374, editquality).
	if steps, err := loadOwnWorkflowSteps(ctx, params.DB, params.ExecutionContext); err != nil {
		logger.Warn("diagnose_load_runtime: route-wiring guard SKIPPED — could not load own workflow plan",
			zap.Error(err))
	} else if err := validateRouteWiring(config, steps); err != nil {
		return nil, err
	}

	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "site_id_field", "input_data.site_id"))
	correlationID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "correlation_id_field", "input_data.correlation_id"))
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "domain_field", "input_data.runtime_site"))
	if siteID == "" && correlationID == "" && domain == "" {
		return nil, fmt.Errorf("diagnose_load_runtime: need at least one of site_id / correlation_id / domain in collected_data")
	}
	errLimit := datahelpers.GetIntField(config, "error_limit", 20)
	wiLimit := datahelpers.GetIntField(config, "work_item_limit", 20)

	// Size guards for the model-written data_requests (see runDataRequests): an
	// EXPLAIN estimate rejects a query BEFORE running it; row/cell caps bound the
	// rendered output. All tunable from step config without a rebuild.
	maxRows := datahelpers.GetIntField(config, "explain_max_rows", 50000)
	maxCost := datahelpers.GetIntField(config, "explain_max_cost", 0) // 0 = cost guard off
	rowCap := datahelpers.GetIntField(config, "row_cap", 200)
	cellChars := datahelpers.GetIntField(config, "cell_chars", 600)

	var b strings.Builder

	// Diagnosis target up top so a model-written data_request can scope itself
	// (e.g. WHERE site_id = …) instead of scanning every site in the database.
	fmt.Fprintf(&b, "### diagnosis target\n\nsite_id=%s  domain=%s  correlation_id=%s\n\n",
		dashIfEmpty(siteID), dashIfEmpty(domain), dashIfEmpty(correlationID))

	// ── agent_error_log ──────────────────────────────────────────────────────
	// Real columns: occurred_at, agent_type, step_name, action, error_message,
	// error_code, severity, site_id, domain, orchestration_id, work_item_id.
	// Filter by site_id OR domain; newest first (idx_error_log_site / _time).
	b.WriteString("### agent_error_log (most recent)\n\n")
	errRows, err := params.DB.QueryContext(ctx, `
		SELECT occurred_at, agent_type, COALESCE(step_name,''), COALESCE(action,''),
		       error_message, COALESCE(error_code,''), severity
		FROM agent_error_log
		WHERE ($1::uuid IS NULL OR site_id = $1::uuid)
		  AND ($2::text IS NULL OR domain = $2::text)
		ORDER BY occurred_at DESC
		LIMIT $3`,
		nullUUID(siteID), nullText(domain), errLimit)
	if err != nil {
		return nil, fmt.Errorf("diagnose_load_runtime: query agent_error_log: %w", err)
	}
	errCount := 0
	for errRows.Next() {
		var occ, agent, step, action, msg, code, sev string
		if err := errRows.Scan(&occ, &agent, &step, &action, &msg, &code, &sev); err != nil {
			errRows.Close()
			return nil, fmt.Errorf("diagnose_load_runtime: scan agent_error_log: %w", err)
		}
		fmt.Fprintf(&b, "- [%s] %s/%s (%s) %s: %s\n", occ, agent, step, action, sev, msg)
		errCount++
	}
	errRows.Close()
	if errCount == 0 {
		b.WriteString("(no error rows for this site/domain)\n")
	}

	// ── site_work_items ──────────────────────────────────────────────────────
	// Real columns: item_type, status, handler_agent, summary, attempt_count,
	// error, result, completed_at, updated_at, pipeline. The "completed but no-op"
	// signal (the gamesdesign symptom) shows as status='complete' with a thin
	// result / advanced timestamps — surface status+result+error so the verdict
	// can judge it. Filter by site_id; newest first (idx_swi_site_status).
	if siteID != "" {
		b.WriteString("\n### site_work_items (most recent)\n\n")
		wiRows, err := params.DB.QueryContext(ctx, `
			SELECT item_type, status, COALESCE(handler_agent,''), COALESCE(summary,''),
			       attempt_count, COALESCE(error,''), updated_at, pipeline
			FROM site_work_items
			WHERE site_id = $1::uuid
			ORDER BY updated_at DESC
			LIMIT $2`,
			siteID, wiLimit)
		if err != nil {
			return nil, fmt.Errorf("diagnose_load_runtime: query site_work_items: %w", err)
		}
		wiCount := 0
		for wiRows.Next() {
			var itype, status, handler, summary, errtxt, pipeline string
			var attempts int
			var updated string
			if err := wiRows.Scan(&itype, &status, &handler, &summary, &attempts, &errtxt, &updated, &pipeline); err != nil {
				wiRows.Close()
				return nil, fmt.Errorf("diagnose_load_runtime: scan site_work_items: %w", err)
			}
			fmt.Fprintf(&b, "- [%s] %s/%s status=%s attempts=%d%s — %s\n",
				updated, pipeline, itype, status, attempts, errSuffix(errtxt), summary)
			wiCount++
		}
		wiRows.Close()
		if wiCount == 0 {
			b.WriteString("(no work items for this site)\n")
		}
	}

	// ── orchestration_states ─────────────────────────────────────────────────
	// Real columns: status, current_step, error, currently_executing,
	// updated_at, correlation_id, site_id. Filter by correlation_id (precise) or
	// site_id; surface status + current_step + error (where a run stalled/failed).
	b.WriteString("\n### orchestration_states\n\n")
	osRows, err := params.DB.QueryContext(ctx, `
		SELECT orchestration_name, status, current_step,
		       COALESCE(error,''), COALESCE(currently_executing,''), updated_at
		FROM orchestration_states
		WHERE ($1::uuid IS NULL OR correlation_id = $1::uuid)
		  AND ($2::uuid IS NULL OR site_id = $2::uuid)
		ORDER BY updated_at DESC
		LIMIT 20`,
		nullUUID(correlationID), nullUUID(siteID))
	if err != nil {
		return nil, fmt.Errorf("diagnose_load_runtime: query orchestration_states: %w", err)
	}
	osCount := 0
	for osRows.Next() {
		var name, status, step, errtxt, executing, updated string
		if err := osRows.Scan(&name, &status, &step, &errtxt, &executing, &updated); err != nil {
			osRows.Close()
			return nil, fmt.Errorf("diagnose_load_runtime: scan orchestration_states: %w", err)
		}
		fmt.Fprintf(&b, "- [%s] %s status=%s step=%s exec=%s%s\n",
			updated, name, status, step, executing, errSuffix(errtxt))
		osCount++
	}
	osRows.Close()
	if osCount == 0 {
		b.WriteString("(no orchestration rows for this correlation/site)\n")
	}

	// ── model-written data requests (read-only, the §1a DB-following channel) ──
	// On loop-back the previous verdict's data_requests are forwarded by
	// diagnose_route into route.data_requests, and the loop returns HERE (the
	// workflow's gather_step = load_runtime). Each is run under a READ ONLY
	// transaction with a statement_timeout and appended.
	//
	// SELECT-only is enforced at THREE layers (defence in depth):
	//   1. the verdict prompt instructs a single read-only SELECT/WITH … SELECT only;
	//   2. the model's text is FILTERED twice through diagnose.IsReadOnlySQL — first at
	//      the route layer (diagnose_route reads data_requests from the verdict wire and
	//      drops non-read-only ones) and again here in runDataRequests before execution;
	//   3. the read-only transaction (BeginTx ReadOnly) is the REAL guarantee — it
	//      rejects any write (incl. data-modifying CTEs) regardless of the lint.
	// Empty on the first iteration.
	dataReqField := datahelpers.GetStringField(config, "data_requests_field", "route.data_requests")
	dataReqs := dataRequestsFromCollected(params.CollectedData, dataReqField)
	if len(dataReqs) > 0 {
		b.WriteString("\n### data_requests (model-written, read-only)\n")
		runDataRequests(ctx, params.DB, dataReqs, &b, maxRows, maxCost, rowCap, cellChars)
	}
	// Requests the route's forwarding cap dropped were never run and never will
	// be — same class of silence the code-request path reports (audited on the
	// council's prompting, council-gate eba040a9).
	dataDropField := datahelpers.GetStringField(config, "data_requests_dropped_field", routeOutputPrefix+dataRequestsDroppedKey)
	if n := datahelpers.GetIntField(params.CollectedData, dataDropField, 0); n > 0 {
		if len(dataReqs) == 0 {
			b.WriteString("\n### data_requests (model-written, read-only)\n")
		}
		b.WriteString(upstreamDropNotice("data_request", n))
	}

	// ── agent state (auto-gathered when the hypothesis names agent types) ─────
	// The two-evidence-family guard (pkg/diagnose/step.go coerceVerdict) demands a
	// state/runtime citation alongside static code for any CONFIRM. Config-shaped
	// bugs (dead ai_service blocks, wrong max_tokens, shadowed step config) ARE
	// observable in state — agent_definitions and llm_call_log — but a code-only
	// intake previously put none of that in the bundle, so a correct static
	// diagnosis could never pass the guard (run 960b554d, 2026-07-17: five
	// static-only confirms coerced to UNVERIFIABLE, iteration-cap). This section
	// closes that gap: any agent type NAMED in the symptom/hypothesis text gets
	// its effective LLM config (root + per-step ai_service blocks) and its recent
	// llm_call_log rows rendered into the runtime evidence — citable at tier
	// "state" without the verdicter having to invent a data_request first.
	if agentStateOn, ok := config["agent_state"].(bool); !ok || agentStateOn {
		symptomText := datahelpers.ExtractNestedFieldString(params.CollectedData,
			datahelpers.GetStringField(config, "symptom_field", "input_data.symptom")) +
			" " + datahelpers.ExtractNestedFieldString(params.CollectedData, "route.conclusion")
		agentStateCap := datahelpers.GetIntField(config, "agent_state_cap", 5)
		callLogLimit := datahelpers.GetIntField(config, "agent_call_log_limit", 10)
		gatherAgentState(ctx, params.DB, symptomText, &b, agentStateCap, callLogLimit, logger)
	}

	// ── schema (live tables) ──────────────────────────────────────────────────
	// So the verdict names REAL tables/columns instead of guessing (the gamesdesign
	// loop burned iterations on a non-existent "page_sections"). DENYLIST-driven, so
	// tables added later appear automatically; an optional relevance include keeps
	// the listing focused unless schema_full is set. Its OWN field; the assembler
	// renders it as a "## Schema" section.
	schemaExclude := configStringSlice(config, "schema_exclude_patterns", defaultSchemaExclude)
	schemaInclude := configStringSlice(config, "schema_include_patterns", defaultSchemaInclude)
	schemaFull, _ := config["schema_full"].(bool)
	schemaTableCap := datahelpers.GetIntField(config, "schema_table_cap", 120)
	schemaText, schErr := gatherSchema(ctx, params.DB, schemaExclude, schemaInclude, schemaFull, schemaTableCap)
	if schErr != nil {
		// Non-fatal: a missing schema section must not abort the diagnosis. Surface
		// it in-band so the trail shows the section was attempted.
		schemaText = fmt.Sprintf("(schema introspection failed: %v)\n", schErr)
		logger.Warn("diagnose_load_runtime: schema introspection failed", zap.Error(schErr))
	}

	// ── model-written code requests (the code-SEARCH tier) ───────────────────
	// The breadth counterpart to the call-graph re-scope: forwarded by
	// diagnose_route into route.code_requests and answered from the code_symbols
	// index by the SAME helpers the council's verify tier uses
	// (diagnose_code_lookup_action.go — same package, so this is reuse, not a
	// second implementation).
	//
	// Deliberately built into its OWN string, NOT appended to `b` (the runtime
	// evidence). Code search returns CODE, which is static-tier evidence; the
	// two-evidence-family guard (pkg/diagnose coerceVerdict) requires a CONFIRM to
	// carry both a static citation showing the mechanism AND a state/runtime
	// citation showing it occurring. Folding index results into a section headed
	// "Runtime / DB evidence" would invite the verdicter to cite them as the
	// observed half and satisfy the guard with code alone — defeating the one
	// check that stops a plausible code-only story being confirmed. The assembler
	// renders this under its own static-tier heading.
	var codeEvidence string
	codeReqField := datahelpers.GetStringField(config, "code_requests_field", "route.code_requests")
	if codeChecks := codeChecksFromCollected(params.CollectedData, codeReqField); len(codeChecks) > 0 {
		codeChecks = dedupCodeChecks(codeChecks)
		maxCodeChecks := datahelpers.GetIntField(config, "max_code_checks", 8)
		codeDropped := 0
		if maxCodeChecks > 0 && len(codeChecks) > maxCodeChecks {
			codeDropped = len(codeChecks) - maxCodeChecks
			codeChecks = codeChecks[:maxCodeChecks]
		}
		codeRowCap := datahelpers.GetIntField(config, "code_row_cap", 40)
		codeExcerpt := datahelpers.GetIntField(config, "code_excerpt_chars", 400)
		var cb strings.Builder
		cb.WriteString("Code questions this diagnosis asked, answered from the code_symbols index\n")
		cb.WriteString("(an INDEXED snapshot — each answer names its commit_sha; treat a stale or\nempty answer as 'unknown', NOT as 'absent'):\n")
		// The read-time freshness guard (bugs_open/059): the line above states the
		// rule; this line gives the verdicter the FACT needed to apply it. A stale
		// index answers "absent" identically to a genuine absence, and the verdict
		// prompt's cite-or-abstain acts on absence — so the answer must carry its
		// own freshness, loudly when stale.
		cb.WriteString(codeIndexFreshness(ctx, params.DB))
		for i, c := range codeChecks {
			fmt.Fprintf(&cb, "\n[code_request %d] kind=%s query=%q — %s\n", i+1, c.Kind, c.Query, c.Why)
			if err := answerCodeCheck(ctx, params.DB, c, "", codeRowCap, codeExcerpt, &cb); err != nil {
				// Never fatal: a failed lookup is one unanswered question, not a
				// failed gather. Surfaced in-band so the verdicter sees it was
				// attempted rather than silently reading absence as evidence.
				fmt.Fprintf(&cb, "  (lookup failed: %v)\n", err)
			}
		}
		if codeDropped > 0 {
			fmt.Fprintf(&cb, "\n> %d further code_request(s) dropped (max_code_checks=%d) — coverage was capped, not complete.\n", codeDropped, maxCodeChecks)
		}
		// Drops that happened UPSTREAM, at the route's forwarding cap, never reach
		// this step at all — so without this they would be invisible here and in the
		// bundle. Reported separately from the local cap because the two mean
		// different things: this one says a question was never even forwarded to be
		// answered, though the spin guard has already credited it as progress.
		codeDropField := datahelpers.GetStringField(config, "code_requests_dropped_field", routeOutputPrefix+codeRequestsDroppedKey)
		cb.WriteString(upstreamDropNotice("code_request", datahelpers.GetIntField(params.CollectedData, codeDropField, 0)))
		codeEvidence = cb.String()
		logger.Info("diagnose_load_runtime: answered code requests",
			zap.Int("code_requests", len(codeChecks)),
			zap.Int("code_requests_dropped", codeDropped))
	}

	logger.Info("diagnose_load_runtime: gathered runtime evidence",
		zap.String("site_id", siteID),
		zap.String("correlation_id", correlationID),
		zap.String("domain", domain),
		zap.Int("error_rows", errCount),
		zap.Int("data_requests", len(dataReqs)),
		zap.Bool("schema_full", schemaFull))

	// Returned under "runtime_evidence" + "schema" + "code_evidence" — all read by
	// diagnose_assemble_bundle, which renders each under its own tier heading.
	return map[string]interface{}{
		"runtime_evidence": b.String(),
		"schema":           schemaText,
		"code_evidence":    codeEvidence,
		"error_rows":       errCount,
	}, nil
}

// nullUUID / nullText pass a typed NULL when the value is empty, so the
// "($1 IS NULL OR col = $1)" filters degrade cleanly to "no filter".
//
// PRE-MERGE (dev guide: grep before adding helpers): nullUUID / nullText /
// errSuffix are generic names that may ALREADY exist in package actions or in
// datahelpers. Before merging, run:
//
//	grep -rn "func nullUUID\|func nullText\|func errSuffix" platform/orchestration/actions/
//	grep -rn "NullUUID\|NullText" platform/orchestration/actions/datahelpers/
//
// If an equivalent exists, delete these and use it (or move these to helpers.go).
// Do NOT introduce a second copy.
func nullUUID(s string) interface{} {
	if s == "" {
		return sql.NullString{}
	}
	return s
}
func nullText(s string) interface{} {
	if s == "" {
		return sql.NullString{}
	}
	return s
}
func errSuffix(e string) string {
	if e == "" {
		return ""
	}
	return " error=" + e
}

// dataReq is a single model-written read query and the reason the verdict wants
// it. The verdict step already linted these (Guard 2 at toVerdict) and dropped
// the unsafe ones; we re-lint here as defence in depth and run them read-only.
type dataReq struct{ SQL, Why string }

// dataRequestsFromCollected pulls the verdict's data_requests out of
// collected_data at `field` ([]{sql,why}). PRE-MERGE: confirm the wire keys
// against docs/PROMPT_diagnosis_verdict.md and VerdictWire's json tags — they are
// "sql"/"why" here; adjust if the prompt schema differs.
func dataRequestsFromCollected(collected map[string]interface{}, field string) []dataReq {
	raw := datahelpers.ExtractNestedField(collected, field)
	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var out []dataReq
	for _, it := range arr {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		q, _ := m["sql"].(string)
		why, _ := m["why"].(string)
		if strings.TrimSpace(q) != "" {
			out = append(out, dataReq{SQL: q, Why: why})
		}
	}
	return out
}

// runDataRequests runs each request in a READ ONLY transaction with a
// statement_timeout, appending the rows to `into`. params.DB is *sql.DB;
// pgbouncer pool_mode = transaction. Reads only; defer Rollback; never commits.
// (Guard 3 is the real guarantee; the IsReadOnlySQL lint is Guard 2 in depth.)
func runDataRequests(ctx context.Context, db *sql.DB, reqs []dataReq, into *strings.Builder, maxRows, maxCost, rowCap, cellChars int) {
	for _, r := range reqs {
		if err := diagnose.IsReadOnlySQL(r.SQL); err != nil {
			fmt.Fprintf(into, "\n> data_request skipped (lint): %v\n> %s\n", err, r.Why)
			continue
		}
		tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
		if err != nil {
			fmt.Fprintf(into, "\n> data_request tx error: %v\n", err)
			continue
		}
		func() {
			defer tx.Rollback() // this path never commits
			if _, err := tx.ExecContext(ctx, "SET LOCAL statement_timeout = '15s'"); err != nil {
				fmt.Fprintf(into, "\n> statement_timeout error: %v\n", err)
				return
			}
			// Pre-flight size guard: EXPLAIN (no ANALYZE) PLANS but does not run the
			// query. Reject one the planner estimates will be huge or need a heavy
			// (often unindexed) scan BEFORE executing it. A skip is feedback — the
			// model narrows the query next iteration (a NEW data_request = progress).
			estRows, cost, perr := explainEstimate(ctx, tx, r.SQL)
			if perr != nil {
				fmt.Fprintf(into, "\n> data_request EXPLAIN error: %v\n> %s\n", perr, r.Why)
				return
			}
			if (maxRows > 0 && estRows > float64(maxRows)) || (maxCost > 0 && cost > float64(maxCost)) {
				fmt.Fprintf(into, "\n> data_request skipped (planner estimate ~%.0f rows, cost %.0f; budget rows=%d cost=%d). Narrow it with a tighter WHERE or add a LIMIT.\n> %s\n",
					estRows, cost, maxRows, maxCost, r.Why)
				return
			}
			rows, err := tx.QueryContext(ctx, r.SQL)
			if err != nil {
				fmt.Fprintf(into, "\n> data_request error: %v\n> %s\n", err, r.Why)
				return
			}
			defer rows.Close()
			text, capped, err := formatRowsText(rows, rowCap, cellChars)
			if err != nil {
				fmt.Fprintf(into, "\n> data_request scan error: %v\n", err)
				return
			}
			capNote := ""
			if capped {
				capNote = fmt.Sprintf(" (output capped: <=%d rows, <=%d chars/cell)", rowCap, cellChars)
			}
			fmt.Fprintf(into, "\n#### %s%s\n\n```\n%s```\n", r.Why, capNote, text)
		}()
	}
}

// formatRowsText renders arbitrary result rows (unknown columns) as a header +
// pipe-separated rows. PRE-MERGE: grep the actions/maintenance code for an
// existing generic row->text helper before keeping this:
//
//	grep -rn "func formatRows\|rows.Columns()" platform/orchestration/actions/
//
// If one exists, use it and delete this; do NOT add a second copy.
func formatRowsText(rows *sql.Rows, rowCap, cellChars int) (text string, capped bool, err error) {
	cols, err := rows.Columns()
	if err != nil {
		return "", false, err
	}
	var sb strings.Builder
	sb.WriteString(strings.Join(cols, " | "))
	sb.WriteString("\n")
	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	n := 0
	for rows.Next() {
		if rowCap > 0 && n >= rowCap {
			capped = true
			break
		}
		if err := rows.Scan(ptrs...); err != nil {
			return "", capped, err
		}
		cells := make([]string, len(cols))
		for i, v := range vals {
			cells[i] = truncateCell(cellToString(v), cellChars)
		}
		sb.WriteString(strings.Join(cells, " | "))
		sb.WriteString("\n")
		n++
	}
	if n == 0 {
		sb.WriteString("(0 rows)\n")
	}
	if capped {
		fmt.Fprintf(&sb, "… (capped at %d rows)\n", rowCap)
	}
	return sb.String(), capped, rows.Err()
}

func cellToString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// ── schema-section + data_request size-guard helpers ─────────────────────────
// PRE-MERGE (dev guide: grep before adding helpers): dashIfEmpty, compactType,
// truncateCell, configStringSlice, gatherSchema, explainEstimate are generic
// names — grep package actions + datahelpers and delete any that already exist:
//   grep -rn "func dashIfEmpty\|func compactType\|func truncateCell\|func configStringSlice" platform/orchestration/actions/

// dashIfEmpty renders "-" for an empty identifier in the target header.
func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// defaultSchema{Exclude,Include} back the schema_{exclude,include}_patterns config
// keys. Denylist so tables added later appear automatically; the include keeps the
// listing to the build/content domain unless schema_full is set.
var defaultSchemaExclude = []string{"%backup%", "%bak%", "%archive%", "%supersede%"}
var defaultSchemaInclude = []string{"site%", "page%", "content%", "flow%"}

// configStringSlice reads a []string from config[key] (a JSON array of strings in
// the step config), returning def if absent/empty/wrong-typed.
func configStringSlice(config map[string]interface{}, key string, def []string) []string {
	raw, ok := config[key]
	if !ok {
		return def
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return def
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if str, ok := v.(string); ok && strings.TrimSpace(str) != "" {
			out = append(out, str)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}

// gatherSchema returns a compact "table(col type, …)" listing of the LIVE public
// tables, so the verdict names real tables/columns instead of guessing. READ-ONLY
// (information_schema only). DENYLIST-driven via `exclude` (NOT ILIKE, patterns
// bound as parameters — injection-safe), so newly-added tables appear without
// editing a list. When !full an `include` relevance filter (ILIKE ANY) keeps the
// listing focused. Capped at tableCap tables.
func gatherSchema(ctx context.Context, db *sql.DB, exclude, include []string, full bool, tableCap int) (string, error) {
	conds := []string{"table_schema = 'public'"}
	args := []interface{}{}
	n := 1
	for _, p := range exclude {
		if strings.TrimSpace(p) == "" {
			continue
		}
		conds = append(conds, fmt.Sprintf("table_name NOT ILIKE $%d", n))
		args = append(args, p)
		n++
	}
	if !full && len(include) > 0 {
		var ors []string
		for _, p := range include {
			if strings.TrimSpace(p) == "" {
				continue
			}
			ors = append(ors, fmt.Sprintf("table_name ILIKE $%d", n))
			args = append(args, p)
			n++
		}
		if len(ors) > 0 {
			conds = append(conds, "("+strings.Join(ors, " OR ")+")")
		}
	}
	query := "SELECT table_name, column_name, data_type FROM information_schema.columns WHERE " +
		strings.Join(conds, " AND ") + " ORDER BY table_name, ordinal_position"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	cur := ""
	var cols []string
	tables := 0
	flush := func() {
		if cur == "" {
			return
		}
		fmt.Fprintf(&sb, "%s(%s)\n", cur, strings.Join(cols, ", "))
	}
	for rows.Next() {
		var t, c, dt string
		if err := rows.Scan(&t, &c, &dt); err != nil {
			return "", err
		}
		if t != cur {
			flush()
			if tableCap > 0 && tables >= tableCap {
				sb.WriteString("… (schema truncated; raise schema_table_cap or narrow the relevance include)\n")
				cur = ""
				break
			}
			cur = t
			cols = cols[:0]
			tables++
		}
		cols = append(cols, c+" "+compactType(dt))
	}
	flush()
	if err := rows.Err(); err != nil {
		return "", err
	}
	if tables == 0 {
		return "(no tables matched the schema filter)\n", nil
	}
	return sb.String(), nil
}

// compactType shortens the verbose information_schema data_type names.
func compactType(dt string) string {
	switch dt {
	case "character varying":
		return "varchar"
	case "timestamp with time zone":
		return "timestamptz"
	case "timestamp without time zone":
		return "timestamp"
	case "double precision":
		return "float8"
	default:
		return dt
	}
}

// explainEstimate returns the planner's estimated row count and total cost for a
// SELECT via EXPLAIN (FORMAT JSON) — which PLANS but does NOT execute the query.
// A high estimate (or cost, which a missing index inflates) lets runDataRequests
// reject a query before it runs, rather than waiting on statement_timeout.
func explainEstimate(ctx context.Context, tx *sql.Tx, query string) (estRows, cost float64, err error) {
	q := strings.TrimSuffix(strings.TrimSpace(query), ";")
	var planJSON string
	if err = tx.QueryRowContext(ctx, "EXPLAIN (FORMAT JSON) "+q).Scan(&planJSON); err != nil {
		return 0, 0, err
	}
	var plans []struct {
		Plan struct {
			PlanRows  float64 `json:"Plan Rows"`
			TotalCost float64 `json:"Total Cost"`
		} `json:"Plan"`
	}
	if err = json.Unmarshal([]byte(planJSON), &plans); err != nil {
		return 0, 0, fmt.Errorf("parse EXPLAIN json: %w", err)
	}
	if len(plans) == 0 {
		return 0, 0, fmt.Errorf("empty EXPLAIN plan")
	}
	return plans[0].Plan.PlanRows, plans[0].Plan.TotalCost, nil
}

// truncateCell caps s to max runes (rune-safe, so the bundle never carries a
// split UTF-8 sequence), appending an ellipsis when it trims.
func truncateCell(s string, max int) string {
	if max <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

// ── agent-state autogather (the config-shaped state-evidence section) ─────────

// matchAgentTypes returns the agent types whose name appears as a whole token in
// text (case-insensitive). Pure; unit-tested in diagnose_load_runtime_test.go.
// Whole-token matching stops "generic" matching inside "generically" while still
// matching "diagnose-agent" inside "the diagnose-agent definition had".
func matchAgentTypes(text string, types []string) []string {
	lower := strings.ToLower(text)
	var out []string
	for _, t := range types {
		lt := strings.ToLower(strings.TrimSpace(t))
		if lt == "" {
			continue
		}
		idx := 0
		for {
			i := strings.Index(lower[idx:], lt)
			if i < 0 {
				break
			}
			start := idx + i
			end := start + len(lt)
			beforeOK := start == 0 || !isTypeTokenChar(lower[start-1])
			afterOK := end == len(lower) || !isTypeTokenChar(lower[end])
			if beforeOK && afterOK {
				out = append(out, t)
				break
			}
			idx = end
		}
	}
	return out
}

// isTypeTokenChar: characters that CONTINUE an agent-type token. '-' and '_' are
// part of type names (page-content-writer, content_researcher), so a match
// flanked by them is a substring of a longer type, not a whole token.
func isTypeTokenChar(c byte) bool {
	return c == '-' || c == '_' ||
		(c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

// gatherAgentState renders, for every agent type named in the symptom/hypothesis
// text: the ROOT ai_service block, every per-step ai_service block (both cited
// often in config bugs — the step OVERLAYS the root per key since bugs_open/009
// was fixed; before that the root shadowed the step entirely), the
// top-level max_tokens key, and recent llm_call_log rows (max_tokens vs
// output_tokens is the truncation signal, see bugs_open/008). Non-fatal
// throughout: evidence gathering must never abort a diagnosis.
func gatherAgentState(ctx context.Context, db *sql.DB, symptomText string, b *strings.Builder, typeCap, callLogLimit int, logger *zap.Logger) {
	if strings.TrimSpace(symptomText) == "" {
		return
	}
	typeRows, err := db.QueryContext(ctx,
		`SELECT DISTINCT type FROM agent_definitions WHERE deleted_at IS NULL`)
	if err != nil {
		logger.Warn("diagnose_load_runtime: agent_state type listing failed", zap.Error(err))
		return
	}
	var allTypes []string
	for typeRows.Next() {
		var t string
		if err := typeRows.Scan(&t); err == nil {
			allTypes = append(allTypes, t)
		}
	}
	typeRows.Close()

	matched := matchAgentTypes(symptomText, allTypes)
	if len(matched) == 0 {
		return
	}
	if len(matched) > typeCap {
		matched = matched[:typeCap]
	}

	b.WriteString("\n### agent state (auto-gathered: agent types named in the symptom/hypothesis)\n\n")

	// Root ai_service + top-level max_tokens per matched type.
	cfgRows, err := db.QueryContext(ctx, `
		SELECT type,
		       COALESCE(default_config #>> '{ai_service,model}', ''),
		       COALESCE(default_config #>> '{ai_service,max_tokens}', ''),
		       COALESCE(default_config #>> '{max_tokens}', ''),
		       (default_config ? 'ai_service')
		FROM agent_definitions
		WHERE type = ANY($1::text[]) AND deleted_at IS NULL`,
		toPGTextArrayLiteral(matched))
	if err != nil {
		fmt.Fprintf(b, "(agent_definitions root-config query failed: %v)\n", err)
		return
	}
	for cfgRows.Next() {
		var t, rootModel, rootMax, topMax string
		var hasRoot bool
		if err := cfgRows.Scan(&t, &rootModel, &rootMax, &topMax, &hasRoot); err != nil {
			continue
		}
		fmt.Fprintf(b, "- agent_definitions[%s]: root ai_service present=%v model=%s max_tokens=%s; top-level max_tokens=%s\n",
			t, hasRoot, dashIfEmpty(rootModel), dashIfEmpty(rootMax), dashIfEmpty(topMax))
	}
	cfgRows.Close()

	// Per-step ai_service blocks. These OVERLAY the root block key-by-key
	// (resolveAIServiceConfig, ai_actions.go). Both are emitted unannotated and
	// the two lines can disagree — that is intentional: the bundle states what
	// the config HOLDS and lets the verdicter reason about precedence, rather
	// than baking in a precedence claim that a later fix would falsify. It did
	// falsify one: until bugs_open/009 shipped, the root shadowed the step
	// wholesale, and a bundle asserting the opposite would have been evidence
	// for a wrong diagnosis.
	stepRows, err := db.QueryContext(ctx, `
		SELECT ad.type, s.key,
		       COALESCE(s.value #>> '{config,ai_service,model}', ''),
		       COALESCE(s.value #>> '{config,ai_service,max_tokens}', '')
		FROM agent_definitions ad,
		     LATERAL jsonb_each(COALESCE(ad.default_config #> '{workflow,steps}', '{}'::jsonb)) s
		WHERE ad.type = ANY($1::text[]) AND ad.deleted_at IS NULL
		  AND s.value #> '{config,ai_service}' IS NOT NULL`,
		toPGTextArrayLiteral(matched))
	if err != nil {
		fmt.Fprintf(b, "(agent_definitions step-config query failed: %v)\n", err)
	} else {
		for stepRows.Next() {
			var t, step, model, max string
			if err := stepRows.Scan(&t, &step, &model, &max); err != nil {
				continue
			}
			fmt.Fprintf(b, "- agent_definitions[%s] step %q ai_service: model=%s max_tokens=%s\n",
				t, step, dashIfEmpty(model), dashIfEmpty(max))
		}
		stepRows.Close()
	}

	// Recent llm_call_log rows for the matched types — the observed tier for any
	// max_tokens / model / truncation claim (output_tokens == max_tokens is the
	// silent-truncation signature, 17 live rows as of 2026-07-16).
	logRows, err := db.QueryContext(ctx, `
		SELECT created_at, agent_type, COALESCE(step_name,''), model,
		       COALESCE(max_tokens, 0), COALESCE(output_tokens, 0), success
		FROM llm_call_log
		WHERE agent_type = ANY($1::text[])
		ORDER BY created_at DESC
		LIMIT $2`,
		toPGTextArrayLiteral(matched), callLogLimit)
	if err != nil {
		fmt.Fprintf(b, "(llm_call_log query failed: %v)\n", err)
		return
	}
	n := 0
	for logRows.Next() {
		var created, agent, step, model string
		var maxTok, outTok int
		var success bool
		if err := logRows.Scan(&created, &agent, &step, &model, &maxTok, &outTok, &success); err != nil {
			continue
		}
		fmt.Fprintf(b, "- llm_call_log [%s] %s/%s model=%s max_tokens=%d output_tokens=%d success=%v\n",
			created, agent, step, model, maxTok, outTok, success)
		n++
	}
	logRows.Close()
	if n == 0 {
		b.WriteString("(no llm_call_log rows for the named agent types)\n")
	}
	logger.Info("diagnose_load_runtime: agent state auto-gathered",
		zap.Strings("matched_types", matched), zap.Int("call_log_rows", n))
}

// upstreamDropNotice renders the bundle line for requests the ROUTE's forwarding
// cap dropped before this gather ran — or "" when none were dropped.
//
// Its own function so it can be unit-tested without a DB (council-gate
// eba040a9, editquality: the render branch was "the second half of the fix and
// currently unverified"). The wording is load-bearing, not decoration: the
// verdicter must not read a capped-away question as an answered one, which is
// the same empty-vs-absent trap the code tier guards against everywhere else.
func upstreamDropNotice(kind string, dropped int) string {
	if dropped <= 0 {
		return ""
	}
	return fmt.Sprintf(
		"\n> %d further %s(s) were dropped BEFORE this gather (route forwarding cap) — they were asked but never answered. Coverage was capped, not complete; do not read their absence as an answer.\n",
		dropped, kind)
}
