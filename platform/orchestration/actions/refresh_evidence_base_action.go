// FILE: platform/orchestration/actions/refresh_evidence_base_action.go
//
// RefreshEvidenceBaseAction — V4 of the claims-verification layer
// (SPEC_claims_verification §5 V4): freshness for live-verifiable facts.
//
// A fact sourced from a SQL query goes stale the moment the underlying data
// moves. Published copy saying "2,767 records verified" stays true only while
// the database agrees. This action re-runs every `source.sql` fact in a site's
// evidence base, records the live value, and:
//
//   - updates `value` and `verified_at` on the fact (mechanical — the query IS
//     the source of truth for a sql-sourced fact);
//   - raises a `stale_evidence` work item when a live value moves outside the
//     fact's tolerance, INCLUDING when the site is now UNDER-claiming (the copy
//     says 2,767 and the database says 3,104 — not a lie, but worth knowing);
//   - regenerates the writer whitelist (`writer_block`, consumed by the
//     page-content-writer prompt — V2) so the numbers the writer is permitted
//     to assert can never quietly rot.
//
// Truth decisions stay human. This action updates a fact's NUMBER from its own
// declared source; it never invents a fact, never removes one, never edits
// published copy, and never rewrites the human-authored WORDS of the whitelist
// (see "Who owns what" below). Drift terminates at human review.
//
// ── Who owns what in the whitelist ──────────────────────────────────────────
//
// Humans own the words; the machine owns the numbers. Each fact may carry a
// `writer_line` — human-authored phrasing, with `{value}` where the number
// goes, and any caveat the audit demands ("This is a CATALOGUE count — never
// present it as a running fleet"). Regeneration substitutes fresh values into
// those lines; it never composes prose of its own beyond the three structural
// section headers. A fact with no `writer_line` is simply omitted from the
// whitelist — silence is the safe default for a fact nobody has phrased.
//
// Regeneration is opt-in per site via `writer_block_managed: true`. Where it is
// absent or false the hand-written `writer_block` is left exactly as it is, and
// any drift is reported in the work item so a human can update it.
//
// ── SQL safety ──────────────────────────────────────────────────────────────
//
// These queries live in a data column, so they are treated as untrusted input
// even though the evidence base is human-owned and pinned:
//   - single statement only (no `;` other than a trailing one);
//   - must begin with SELECT — no CTEs, no DDL/DML verbs anywhere;
//   - executed inside a READ ONLY transaction with a statement timeout;
//   - must return exactly one row, one column, numeric.
// A query failing any of these is skipped and reported, never executed blind.
//
// Config:
//
//	"refresh_evidence_base": {
//	    "action": "refresh_evidence_base",
//	    "config": {
//	        "site_id": "site_record.site_id",   // optional — omit to sweep
//	                                            // every site with an evidence base
//	        "dry_run": false                    // optional — report, write nothing
//	    }
//	}
//
// Registration: "refresh_evidence_base" in registry.go (Category "site").

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RefreshEvidenceBaseInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"site_id"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("refresh_evidence_base", RefreshEvidenceBaseInputSpec)
}

// evidenceFactRefresh is one fact's outcome in a refresh pass.
type evidenceFactRefresh struct {
	FactID     string   `json:"fact_id"`
	Claim      string   `json:"claim"`
	StoredVal  *float64 `json:"stored_value,omitempty"`
	LiveVal    *float64 `json:"live_value,omitempty"`
	Tolerance  string   `json:"tolerance,omitempty"`
	Direction  string   `json:"direction,omitempty"` // "up" | "down" | "unchanged"
	Outcome    string   `json:"outcome"`             // fresh | updated | drifted | error
	Detail     string   `json:"detail,omitempty"`
	VerifiedAt string   `json:"verified_at,omitempty"`
}

// siteRefreshResult is one site's outcome.
type siteRefreshResult struct {
	SiteID          string                `json:"site_id"`
	Domain          string                `json:"domain"`
	FactsChecked    int                   `json:"facts_checked"`
	FactsUpdated    int                   `json:"facts_updated"`
	Drifted         int                   `json:"drifted"`
	Errors          int                   `json:"errors"`
	WriterBlock     string                `json:"writer_block_action"` // regenerated | unmanaged | unchanged
	WorkItemCreated bool                  `json:"work_item_created"`
	Facts           []evidenceFactRefresh `json:"facts"`
}

func RefreshEvidenceBaseAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "refresh_evidence_base"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, RefreshEvidenceBaseInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	dryRun := configBoolOrDefault(params.StepConfig.Config, "dry_run", false)

	// Target sites: the one named, or every site with a current evidence base.
	siteIDs, err := resolveEvidenceSites(ctx, params.DB, inputs.Get("site_id"), logger)
	if err != nil {
		return nil, err
	}
	if len(siteIDs) == 0 {
		logger.Info("refresh_evidence_base: no sites with an evidence base — nothing to do")
		return map[string]interface{}{"sites_checked": 0, "results": []interface{}{}}, nil
	}

	results := make([]siteRefreshResult, 0, len(siteIDs))
	totalDrift := 0
	for _, siteID := range siteIDs {
		res, err := refreshOneSiteEvidence(ctx, params, siteID, dryRun, logger)
		if err != nil {
			// One site's failure must not abort the sweep.
			logger.Warn("refresh_evidence_base: site refresh failed",
				zap.String("site_id", siteID.String()), zap.Error(err))
			continue
		}
		totalDrift += res.Drifted
		results = append(results, *res)
	}

	logger.Info("refresh_evidence_base: complete",
		zap.Int("sites_checked", len(results)),
		zap.Int("total_drifted", totalDrift),
		zap.Bool("dry_run", dryRun))

	return map[string]interface{}{
		"sites_checked": len(results),
		"total_drifted": totalDrift,
		"dry_run":       dryRun,
		"results":       results,
	}, nil
}

// resolveEvidenceSites returns the target site list: the named site, or every
// site holding a current evidence_base spec.
func resolveEvidenceSites(ctx context.Context, db *sql.DB, siteIDStr string, logger *zap.Logger) ([]uuid.UUID, error) {
	if siteIDStr != "" {
		id, err := uuid.Parse(siteIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
		}
		return []uuid.UUID{id}, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT site_id FROM site_specs
		WHERE aspect = 'evidence_base' AND is_current = true
	`)
	if err != nil {
		return nil, fmt.Errorf("list evidence-base sites: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			logger.Warn("refresh_evidence_base: site id scan failed", zap.Error(err))
			continue
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func refreshOneSiteEvidence(
	ctx context.Context, params ActionParams, siteID uuid.UUID, dryRun bool, logger *zap.Logger,
) (*siteRefreshResult, error) {
	db := params.DB

	// The row id read here is the compare-and-swap token for the write below.
	var specRowID uuid.UUID
	var rawJSON []byte
	var pinned sql.NullBool
	err := db.QueryRowContext(ctx, `
		SELECT id, data, pinned FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
	`, siteID).Scan(&specRowID, &rawJSON, &pinned)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("site has no current evidence_base")
	}
	if err != nil {
		return nil, fmt.Errorf("read evidence_base: %w", err)
	}

	// Work on the generic map so unknown keys (audit_doc, banned_claims,
	// allowed_entities, anything a future revision adds) survive the rewrite
	// untouched — this action must never drop a human's data.
	var eb map[string]interface{}
	if err := json.Unmarshal(rawJSON, &eb); err != nil {
		return nil, fmt.Errorf("unmarshal evidence_base: %w", err)
	}

	domain := loadSiteDomain(ctx, db, siteID, logger)
	res := &siteRefreshResult{SiteID: siteID.String(), Domain: domain, WriterBlock: "unchanged"}

	factsRaw, _ := eb["facts"].([]interface{})
	today := currentDateString(ctx, db)

	changed := false
	for _, fr := range factsRaw {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		// V5: citation facts are re-verified by re-fetching their source and
		// matching the stored verbatim quote (SPEC_V5_researched_citations §3c).
		if src, ok := fact["source"].(map[string]interface{}); ok {
			if _, has := src["citation"]; has {
				entry := refreshCitationFact(ctx, fact, today)
				if entry != nil {
					res.FactsChecked++
					switch entry.Outcome {
					case "drifted":
						res.Drifted++
					case "error":
						res.Errors++
					}
					if entry.Outcome == "updated" || (entry.Outcome == "fresh" && entry.VerifiedAt == today) {
						res.FactsUpdated++
						changed = true
					}
					res.Facts = append(res.Facts, *entry)
				}
				continue
			}
		}

		query := factSQLSource(fact)
		if query == "" {
			continue // artifact/attested facts are checked for presence, not re-proved
		}

		res.FactsChecked++
		entry := evidenceFactRefresh{
			FactID:    datahelpers.GetStringField(fact, "id", ""),
			Claim:     datahelpers.GetStringField(fact, "claim", ""),
			Tolerance: datahelpers.GetStringField(fact, "tolerance", "exact"),
		}
		if v, ok := numericField(fact["value"]); ok {
			stored := v
			entry.StoredVal = &stored
		}

		live, err := runEvidenceQuery(ctx, db, query)
		if err != nil {
			entry.Outcome = "error"
			entry.Detail = err.Error()
			res.Errors++
			res.Facts = append(res.Facts, entry)
			logger.Warn("refresh_evidence_base: fact query failed",
				zap.String("fact_id", entry.FactID), zap.Error(err))
			continue
		}
		liveVal := live
		entry.LiveVal = &liveVal

		switch {
		case entry.StoredVal == nil:
			entry.Outcome = "updated"
			entry.Detail = "fact had no stored value; recorded the live one"
		case math.Abs(*entry.StoredVal-live) < 1e-9:
			entry.Outcome = "fresh"
			entry.Direction = "unchanged"
		default:
			entry.Direction = "up"
			if live < *entry.StoredVal {
				entry.Direction = "down"
			}
			if evidenceValueWithinTolerance(*entry.StoredVal, live, entry.Tolerance) {
				entry.Outcome = "updated"
				entry.Detail = fmt.Sprintf("live value moved %s within tolerance %s",
					entry.Direction, entry.Tolerance)
			} else {
				entry.Outcome = "drifted"
				res.Drifted++
				entry.Detail = fmt.Sprintf(
					"live value %.0f is outside tolerance %q of the published %.0f (moved %s) — published copy may need a human ruling",
					live, entry.Tolerance, *entry.StoredVal, entry.Direction)
			}
		}

		// The query is the source of truth for a sql-sourced fact, so the
		// register's number tracks it either way. Drift is reported, not
		// suppressed — the work item is what a human acts on.
		if entry.Outcome != "fresh" {
			fact["value"] = live
			fact["verified_at"] = today
			entry.VerifiedAt = today
			res.FactsUpdated++
			changed = true
		} else if datahelpers.GetStringField(fact, "verified_at", "") != today {
			fact["verified_at"] = today
			entry.VerifiedAt = today
			changed = true
		}

		res.Facts = append(res.Facts, entry)
	}

	if res.FactsChecked == 0 {
		return res, nil // no live-verifiable facts — nothing to do, no noise
	}

	// ── Whitelist regeneration (V2's writer_block) ──
	managed := false
	if b, ok := eb["writer_block_managed"].(bool); ok {
		managed = b
	}
	if managed {
		if block := composeWriterBlock(eb); block != "" {
			if existing, _ := eb["writer_block"].(string); existing != block {
				eb["writer_block"] = block
				changed = true
				res.WriterBlock = "regenerated"
			}
		}
	} else {
		res.WriterBlock = "unmanaged"
	}

	if dryRun || !changed {
		return res, nil
	}

	if err := writeRefreshedEvidenceBase(ctx, db, siteID, specRowID, eb, pinned, res, logger); err != nil {
		return nil, err
	}

	if res.Drifted > 0 {
		// bugs_open/091 candidate 2. This used to read `else { res.WorkItemCreated
		// = true }`, so work_item_created meant "no error", not "a record exists".
		// The item is keyed per SITE, and insertWorkItem dedups on that key while
		// any non-terminal row is open — so when a SECOND, different fact drifts
		// while an earlier item is still open, nothing is written and the run
		// reported the write anyway. The only trace of the truth was a log line
		// carrying inserted=false, in a pod that was replaced four minutes later.
		//
		// This does NOT fix the dedup dropping the finding (candidate 1) — that
		// changes a helper every detector in the fleet calls and belongs to the
		// work_item_completion_integrity workstream. It makes the report honest
		// while that is decided, which is true under either answer.
		inserted, err := createStaleEvidenceItem(ctx, db, siteID, domain, res, params.AgentType, logger)
		if err != nil {
			logger.Warn("refresh_evidence_base: failed to create stale_evidence item", zap.Error(err))
		}
		res.WorkItemCreated = inserted
		if err == nil && !inserted {
			logger.Warn("refresh_evidence_base: drift found but NO work item written — an open "+
				"stale_evidence item already holds this site's key, and its spec still describes "+
				"the EARLIER drift (bugs_open/091)",
				zap.String("site_id", siteID.String()),
				zap.String("domain", domain),
				zap.Int("drifted", res.Drifted))
		}
	}
	return res, nil
}

// refreshCitationFact re-verifies one citation fact in place and returns its
// refresh entry. Outcomes and what they mean:
//
//	fresh    — quote still present at the source; accessed/verified_at bumped.
//	drifted  — the source no longer supports the claim (quote gone on a 200),
//	           OR the citation has aged past its staleness_days policy. Either
//	           way the published claim needs a human ruling.
//	error    — the source could not be checked (network, 403, PDF). Unknown is
//	           not loss: reported, never treated as drift.
//
// Facts marked "reverifiable": false are never fetched — they age by policy
// only, and re-attesting them is a human act.
func refreshCitationFact(ctx context.Context, fact map[string]interface{}, today string) *evidenceFactRefresh {
	entry := &evidenceFactRefresh{
		FactID:    datahelpers.GetStringField(fact, "id", ""),
		Claim:     datahelpers.GetStringField(fact, "claim", ""),
		Tolerance: "citation",
	}
	src, _ := fact["source"].(map[string]interface{})
	cit, err := datahelpers.ParseCitation(src)
	if err != nil {
		entry.Outcome = "error"
		entry.Detail = err.Error()
		return entry
	}

	stalenessDays := 0.0
	if v, ok := numericField(fact["staleness_days"]); ok {
		stalenessDays = v
	}
	verifiedAt := datahelpers.GetStringField(fact, "verified_at", "")
	now, _ := parseFlexibleDate(today)

	if datahelpers.GetBoolField(fact, "reverifiable", true) == false {
		if citationDateStale(cit.Published, verifiedAt, stalenessDays, now) {
			entry.Outcome = "drifted"
			entry.Detail = "citation is past its staleness_days policy and is marked reverifiable:false — re-attest it by hand or retire the claim"
		} else {
			entry.Outcome = "fresh"
			entry.Detail = "reverifiable:false — aged by policy only, not re-fetched"
		}
		return entry
	}

	outcome := verifyCitationLive(ctx, cit)
	switch {
	case outcome.FailClass == "fetch_error":
		entry.Outcome = "error"
		entry.Detail = outcome.FailDetail
	case !outcome.Found:
		entry.Outcome = "drifted"
		entry.Detail = "citation_lost: " + outcome.FailDetail
	case citationDateStale(cit.Published, verifiedAt, stalenessDays, now):
		entry.Outcome = "drifted"
		entry.Detail = "quote still present, but the citation is past its staleness_days policy — the source itself has aged; re-research the figure"
	default:
		entry.Outcome = "fresh"
		fact["verified_at"] = today
		if citRaw, ok := src["citation"].(map[string]interface{}); ok {
			citRaw["accessed"] = today
		}
		entry.VerifiedAt = today
	}
	return entry
}

// factSQLSource returns the fact's SQL source, if it has one.
func factSQLSource(fact map[string]interface{}) string {
	src, ok := fact["source"].(map[string]interface{})
	if !ok {
		return ""
	}
	return strings.TrimSpace(datahelpers.GetStringField(src, "sql", ""))
}

func numericField(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

var selectOnlyRe = regexp.MustCompile(`(?is)^\s*select\s`)

// forbiddenVerbRe catches write verbs anywhere in the statement, so a nested
// or trailing write cannot ride along inside an otherwise-SELECT query.
var forbiddenVerbRe = regexp.MustCompile(`(?is)\b(insert|update|delete|drop|alter|truncate|grant|revoke|create|copy|vacuum|call|do|merge)\b`)

// validateEvidenceQuery applies the safety rules in the file header and
// returns the statement to execute. Pure and side-effect free, so the guard is
// testable on its own — the part that must never be got wrong is the part that
// never touches a database.
func validateEvidenceQuery(query string) (string, error) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(query), ";")
	if strings.Contains(trimmed, ";") {
		return "", fmt.Errorf("refused: query contains multiple statements")
	}
	if !selectOnlyRe.MatchString(trimmed) {
		return "", fmt.Errorf("refused: query must begin with SELECT")
	}
	if forbiddenVerbRe.MatchString(trimmed) {
		return "", fmt.Errorf("refused: query contains a data-modifying keyword")
	}
	return trimmed, nil
}

// runEvidenceQuery validates one evidence query, executes it read-only, and
// returns its single numeric scalar.
func runEvidenceQuery(ctx context.Context, db *sql.DB, query string) (float64, error) {
	trimmed, err := validateEvidenceQuery(query)
	if err != nil {
		return 0, err
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return 0, fmt.Errorf("begin read-only tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SET LOCAL statement_timeout = '15s'`); err != nil {
		return 0, fmt.Errorf("set statement timeout: %w", err)
	}

	var raw interface{}
	if err := tx.QueryRowContext(ctx, trimmed).Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("query returned no rows")
		}
		return 0, fmt.Errorf("query failed: %w", err)
	}

	switch v := raw.(type) {
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case []byte:
		var f float64
		if _, err := fmt.Sscanf(string(v), "%g", &f); err != nil {
			return 0, fmt.Errorf("non-numeric result %q", string(v))
		}
		return f, nil
	case string:
		var f float64
		if _, err := fmt.Sscanf(v, "%g", &f); err != nil {
			return 0, fmt.Errorf("non-numeric result %q", v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unsupported result type %T", raw)
	}
}

// evidenceValueWithinTolerance mirrors the scan-side tolerance semantics in
// datahelpers/claims.go, from the register's point of view: does the live value
// still sit within what the published claim tolerates?
//
//   - exact: any movement is drift (the copy names a specific number)
//   - gte:   the copy says "more than N"; growth is fine, a FALL below the
//     published figure is drift (the copy would be overclaiming)
//   - approx_pct:N: movement beyond N percent is drift
func evidenceValueWithinTolerance(stored, live float64, tolerance string) bool {
	switch {
	case tolerance == "gte":
		return live >= stored
	case strings.HasPrefix(tolerance, "approx_pct:"):
		var pct float64
		if _, err := fmt.Sscanf(strings.TrimPrefix(tolerance, "approx_pct:"), "%g", &pct); err != nil {
			return false
		}
		if stored == 0 {
			return live == 0
		}
		return math.Abs(live-stored)/math.Abs(stored)*100 <= pct
	default: // exact
		return math.Abs(live-stored) < 1e-9
	}
}

// composeWriterBlock rebuilds the V2 writer whitelist from the register.
// Structure (the three headers) is the machine's; every claim sentence is the
// human's `writer_line`, with `{value}` replaced by the current number. A fact
// without a writer_line is omitted — never auto-phrased.
func composeWriterBlock(eb map[string]interface{}) string {
	factsRaw, _ := eb["facts"].([]interface{})
	var numbers, capabilities []string

	for _, fr := range factsRaw {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		line := strings.TrimSpace(datahelpers.GetStringField(fact, "writer_line", ""))
		if line == "" {
			continue
		}
		if v, ok := numericField(fact["value"]); ok {
			line = strings.ReplaceAll(line, "{value}", formatEvidenceNumber(v))
			numbers = append(numbers, "- "+line)
		} else {
			capabilities = append(capabilities, "- "+line)
		}
	}

	if len(numbers) == 0 && len(capabilities) == 0 {
		return "" // nothing phrased — leave the existing block alone
	}

	var b strings.Builder
	if len(numbers) > 0 {
		b.WriteString("NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine):\n")
		b.WriteString(strings.Join(numbers, "\n"))
	}
	if len(capabilities) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("CAPABILITIES (assert without inventing numbers):\n")
		b.WriteString(strings.Join(capabilities, "\n"))
	}
	if ents, ok := eb["allowed_entities"].([]interface{}); ok && len(ents) > 0 {
		names := make([]string, 0, len(ents))
		for _, e := range ents {
			if s, ok := e.(string); ok && s != "" {
				names = append(names, s)
			}
		}
		if len(names) > 0 {
			b.WriteString("\n\nNAMED ENTITIES you may assert relationships with: ")
			b.WriteString(strings.Join(names, ", "))
			b.WriteString(".")
		}
	}
	return b.String()
}

// formatEvidenceNumber renders a fact value the way published copy states it:
// whole numbers with thousands separators, fractions left alone.
func formatEvidenceNumber(v float64) string {
	if v != math.Trunc(v) {
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", v), "0"), ".")
	}
	s := fmt.Sprintf("%.0f", math.Abs(v))
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if v < 0 {
		return "-" + string(out)
	}
	return string(out)
}

// writeRefreshedEvidenceBase supersedes the current spec row and inserts the
// refreshed one, preserving `pinned` — same shape as write_site_spec's
// versioning, kept local so no other spec semantics (deep merge, formatting)
// apply to an evidence base.
//
// COMPARE-AND-SWAP, not blind overwrite. This action rewrites the whole `data`
// column from a copy read moments earlier, so a human editing the evidence base
// in between would otherwise have their edit silently dropped — the classic
// lost update, and the evidence base is exactly the kind of human-owned row
// where losing an edit is unacceptable. The supersede is therefore keyed on the
// row id we read; if it no longer matches (someone wrote first), we affect zero
// rows and abort WITHOUT writing. The refresh is safely re-runnable, so losing
// a pass costs one scheduled tick; losing a human's edit costs trust.
func writeRefreshedEvidenceBase(
	ctx context.Context, db *sql.DB, siteID, specRowID uuid.UUID,
	eb map[string]interface{}, pinned sql.NullBool,
	res *siteRefreshResult, logger *zap.Logger,
) error {
	newJSON, err := json.Marshal(eb)
	if err != nil {
		return fmt.Errorf("marshal refreshed evidence_base: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	sres, err := tx.ExecContext(ctx, `
		UPDATE site_specs SET is_current = false, superseded_at = now()
		WHERE id = $1 AND is_current = true
	`, specRowID)
	if err != nil {
		return fmt.Errorf("supersede evidence_base: %w", err)
	}
	affected, err := sres.RowsAffected()
	if err != nil {
		return fmt.Errorf("supersede rows affected: %w", err)
	}
	if affected != 1 {
		logger.Warn("refresh_evidence_base: evidence base changed under us — skipping write to avoid clobbering a concurrent edit",
			zap.String("site_id", siteID.String()),
			zap.String("read_row_id", specRowID.String()))
		res.WriterBlock = "skipped_concurrent_edit"
		return nil
	}

	notes := fmt.Sprintf(
		"V4 freshness pass: %d live-verifiable fact(s) checked, %d updated, %d drifted, whitelist %s",
		res.FactsChecked, res.FactsUpdated, res.Drifted, res.WriterBlock)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO site_specs
		    (site_id, aspect, data, source, notes, is_current, created_by, pinned)
		VALUES ($1, 'evidence_base', $2::jsonb, 'scheduled', $3, true, 'evidence-refresher', $4)
	`, siteID, string(newJSON), notes, pinned.Valid && pinned.Bool); err != nil {
		return fmt.Errorf("insert refreshed evidence_base: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit refreshed evidence_base: %w", err)
	}

	logger.Info("refresh_evidence_base: evidence base updated",
		zap.String("site_id", siteID.String()),
		zap.Int("facts_updated", res.FactsUpdated),
		zap.Int("drifted", res.Drifted),
		zap.String("writer_block", res.WriterBlock))
	return nil
}

// createStaleEvidenceItem raises drift for human ruling. HITL-terminal: the
// handler is the human-review pseudo-handler, matching claims_unverified —
// changing published copy because a number moved is a human decision (spec
// open question 3), even when the site is merely under-claiming.
// createStaleEvidenceItem returns whether a row was actually WRITTEN, not merely
// whether the attempt errored (bugs_open/091). The item is keyed per site, so a
// second drift arriving while an earlier item is open dedups to nothing — and
// the caller must be able to tell that apart from success.
func createStaleEvidenceItem(
	ctx context.Context, db *sql.DB, siteID uuid.UUID, domain string,
	res *siteRefreshResult, agentType string, logger *zap.Logger,
) (bool, error) {
	drifted := make([]evidenceFactRefresh, 0, res.Drifted)
	for _, f := range res.Facts {
		if f.Outcome == "drifted" {
			drifted = append(drifted, f)
		}
	}

	specJSON, err := json.Marshal(map[string]interface{}{
		"check":   "evidence_freshness",
		"domain":  domain,
		"drifted": drifted,
		"fix": "A live-verifiable fact has moved outside the tolerance its published wording allows. " +
			"Review the pages that state it: either update the copy to the new figure, re-word the claim " +
			"to a tolerance that survives movement (e.g. 'more than N'), or accept it. The register's " +
			"number has already been re-synced to the live query — the COPY is what needs a human ruling.",
	})
	if err != nil {
		return false, fmt.Errorf("marshal stale_evidence spec: %w", err)
	}

	summary := fmt.Sprintf("Evidence freshness (%s): %d fact(s) drifted outside tolerance", domain, len(drifted))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "scheduled",
		pipeline:     "content",
		itemType:     "stale_evidence",
		severity:     "medium",
		summary:      summary,
		spec:         string(specJSON),
		priority:     35,
		handlerAgent: "human-review",
		status:       "needs_human_review",
		createdBy:    agentType,
		itemKey:      "stale_evidence:" + siteID.String(),
	}, logger)
	if err != nil {
		return false, fmt.Errorf("insert stale_evidence item: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit stale_evidence item: %w", err)
	}

	logger.Warn("refresh_evidence_base: evidence drift raised for human review",
		zap.String("site_id", siteID.String()),
		zap.Int("drifted", len(drifted)),
		zap.Bool("inserted", inserted))
	return inserted, nil
}

func loadSiteDomain(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) string {
	var domain sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain); err != nil {
		logger.Warn("refresh_evidence_base: domain lookup failed", zap.Error(err))
		return ""
	}
	return domain.String
}

// currentDateString reads the date from the database so `verified_at` matches
// the clock that produced the values.
func currentDateString(ctx context.Context, db *sql.DB) string {
	var today string
	if err := db.QueryRowContext(ctx, `SELECT to_char(now(), 'YYYY-MM-DD')`).Scan(&today); err != nil {
		return ""
	}
	return today
}
