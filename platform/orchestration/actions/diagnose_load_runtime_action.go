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
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/pkg/diagnose"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnoseLoadRuntimeInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"site_id_field", "correlation_id_field", "domain_field",
		"error_limit", "work_item_limit", "data_requests_field",
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
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_load_runtime", DiagnoseLoadRuntimeInputSpec)
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

	var b strings.Builder

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
		runDataRequests(ctx, params.DB, dataReqs, &b)
	}

	logger.Info("diagnose_load_runtime: gathered runtime evidence",
		zap.String("site_id", siteID),
		zap.String("correlation_id", correlationID),
		zap.String("domain", domain),
		zap.Int("error_rows", errCount),
		zap.Int("data_requests", len(dataReqs)))

	// Returned under "runtime_evidence" — the field diagnose_assemble_bundle reads.
	return map[string]interface{}{
		"runtime_evidence": b.String(),
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
func runDataRequests(ctx context.Context, db *sql.DB, reqs []dataReq, into *strings.Builder) {
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
			rows, err := tx.QueryContext(ctx, r.SQL)
			if err != nil {
				fmt.Fprintf(into, "\n> data_request error: %v\n> %s\n", err, r.Why)
				return
			}
			defer rows.Close()
			text, err := formatRowsText(rows)
			if err != nil {
				fmt.Fprintf(into, "\n> data_request scan error: %v\n", err)
				return
			}
			fmt.Fprintf(into, "\n#### %s\n\n```\n%s```\n", r.Why, text)
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
func formatRowsText(rows *sql.Rows) (string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return "", err
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
		if err := rows.Scan(ptrs...); err != nil {
			return "", err
		}
		cells := make([]string, len(cols))
		for i, v := range vals {
			cells[i] = cellToString(v)
		}
		sb.WriteString(strings.Join(cells, " | "))
		sb.WriteString("\n")
		n++
	}
	if n == 0 {
		sb.WriteString("(0 rows)\n")
	}
	return sb.String(), rows.Err()
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
