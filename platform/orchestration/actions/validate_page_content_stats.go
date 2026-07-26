package actions

// Check 9 of validate_page_content: the stat-field audit over content_data.
// bugs_open/043 fix candidates 2 (post-generation numeric audit) and 4 (the
// placeholder-unit lint).
//
// WHY A SEPARATE SURFACE FROM CHECK 8. Check 8 scans the rendered HTML, which
// is structurally blind to a stat card — the value and its label are separate
// block elements, so the number's claim window is the bare figure and the
// prose context gate can never match it. The full argument, with the two
// independent reasons, is in datahelpers/claims_stats.go's header.
//
// WHY IT LIVES IN THE BUILD GATE RATHER THAN A DISCOVERY CHECK. The routing is
// already correct here and costs nothing: an `error` severity makes
// validate_page_content return invalid, which fires the step's
// error_step: mark_needs_review -> fail_work_item with
// status_override: needs_human_review (065_page_build_handler_wrapper.sql).
// No new action, no registry entry, and — the load-bearing part — no new
// item_type, so discovery_checks/verifier_coverage_test.go has nothing to say.
//
// OPT-IN, AND WHY THE PREDICATE IS NOT ParseEvidenceBase's. The numeric audit
// compares a figure against a register of that site's verified facts; a site
// with no register has nothing to compare against, so running it there would
// flag every number including the true ones and route every build on every
// site to human review. It stays opt-in. But ParseEvidenceBase returns nil
// when a row holds no facts[] and no banned_claims[], and the rows seeded for
// bug 043's own four sites are writer_block-only — which is precisely how the
// existing claims gate came to be a silent no-op on the three sites that had
// actually fabricated. So this check keys on whether the ROW EXISTS, and
// grades by whether it has facts: facts present -> error (a machine checked it
// and it failed); row present but no facts -> warning that names the gap.
// Severity `error` must never mean "we could not check this".
//
// The unit lint is different in kind and therefore scoped differently: it
// compares a component against ITSELF, needs no register, and runs fleet-wide.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// defaultSectionsMetadataField is where CompilePageSectionsAction's per-section
// content_data arrives in collected data.
//
// THE CALLERS OF THIS GATE, measured 2026-07-26 rather than assumed — an earlier
// version of this comment named "tool-recreation, report-builder and
// content-reviewer" as bare-HTML callers, and a council objection asked for the
// premise to be checked. It was wrong twice:
//
//		SELECT ad.type, k, v->'config'->>'html_field'
//		FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') e(k,v)
//		WHERE COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL AND ad.is_active
//		  AND v->>'action' = 'validate_page_content';
//
//		page-build-handler      validate_content  page_content.response.page_html
//		tool-recreation-handler validate_tool     completeness_check.clean_html
//		content-reviewer        validate_content  (default)
//
//	  - page-build-handler is the only caller that has ever actually run this gate
//	    over section-built pages, and it declares require_sections_metadata.
//	  - tool-recreation-handler validates a tool HTML blob, not sections. Correct
//	    to leave undeclared.
//	  - content-reviewer is configured but DORMANT — zero runs in the retained
//	    orchestration window. It uses the DEFAULT html field, i.e. the same
//	    page_content path, so if it is ever enabled over section-built pages it
//	    will need the declaration too, or the silent-skip ambiguity returns there.
//	  - "report-builder" does not exist as an agent type at all. It was in the
//	    old comment because it was repeated from a summary without being checked.
const defaultSectionsMetadataField = "page_content.response.sections_metadata"

// collectStatClaims walks sections_metadata and lifts the stat claims out of
// every section's content_data.
//
// The bool reports whether the metadata was USABLE, which is not the same as
// whether any claims came back. Council objection (2026-07-26, bug_historian,
// medium): returning a bare nil made "checked, no figures to audit"
// indistinguishable from "could not check" — the exact shape this whole bug is
// about, and a violation of this file's own severity rule. The caller turns an
// unusable-but-expected source into a warning; see requireSectionsMetadata.
func collectStatClaims(collected map[string]interface{}, metaField string, logger *zap.Logger) ([]datahelpers.StatClaim, bool) {
	raw := datahelpers.ExtractNestedField(collected, metaField)
	if raw == nil {
		return nil, false
	}
	sections, ok := raw.([]interface{})
	if !ok {
		logger.Debug("validate_page_content: sections_metadata is not a list — skipping the stat audit",
			zap.String("field", metaField),
			zap.String("type", fmt.Sprintf("%T", raw)))
		return nil, false
	}

	var claims []datahelpers.StatClaim
	for _, s := range sections {
		sec, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		cd, ok := sec["content_data"].(map[string]interface{})
		if !ok || len(cd) == 0 {
			continue
		}
		component, _ := sec["component_function"].(string)
		if component == "" {
			component, _ = sec["component_name"].(string)
		}
		if component == "" {
			component = "unknown-component"
		}
		claims = append(claims, datahelpers.ExtractStatClaims(component, cd)...)
	}
	return claims, true
}

// loadEvidenceBaseForStats reports both the parsed evidence base and whether
// the site_specs row exists at all.
//
// ParseEvidenceBase's nil-means-nothing-scannable contract is deliberately NOT
// changed: it is shared with cmd/claimscan (which exits 2 on nil) and with the
// post-deploy audit, and widening "opted in" there would make both scan sites
// with nothing to scan against.
func loadEvidenceBaseForStats(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (*datahelpers.EvidenceBase, bool) {
	var data []byte
	err := db.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
		LIMIT 1`, siteID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		logger.Warn("validate_page_content: could not load evidence_base for the stat audit",
			zap.Error(err), zap.String("site_id", siteID.String()))
		return nil, false
	}

	eb, perr := datahelpers.ParseEvidenceBase(data)
	if perr != nil {
		logger.Warn("validate_page_content: evidence_base did not parse for the stat audit",
			zap.Error(perr), zap.String("site_id", siteID.String()))
		return nil, true
	}
	return eb, true
}

// checkStatClaims grades unregistered stat figures.
func checkStatClaims(claims []datahelpers.StatClaim, eb *datahelpers.EvidenceBase, hasFacts bool) []ValidationIssue {
	if eb == nil {
		return nil
	}
	var issues []ValidationIssue
	for _, f := range eb.ScanStatClaims(claims) {
		severity := "error"
		description := f.Reason
		switch {
		case !hasFacts:
			// The site opted in but registered no countables, so nothing could
			// have been checked. Saying `error` here would be a lie.
			severity = "warning"
			description = f.Reason +
				" — NOTE: this site has an evidence_base but no machine-readable facts[], so nothing could be verified. Register its real countables as facts to turn this into a gate."
		case f.Check == "unregistered_stat_unpaired":
			// The figure is real and unsupported, but we could not say what it
			// claims to count. Never block on a claim we cannot state.
			severity = "warning"
		}
		issues = append(issues, ValidationIssue{
			Type:        f.Check,
			Category:    "stat_claims",
			Severity:    severity,
			Location:    f.Pattern,
			Value:       f.Matched,
			Description: description,
		})
	}
	return issues
}

// checkStatUnits grades placeholder-unit findings. No evidence base involved.
func checkStatUnits(claims []datahelpers.StatClaim) []ValidationIssue {
	var issues []ValidationIssue
	for _, f := range datahelpers.LintStatUnits(claims) {
		severity := "warning"
		if f.Check == "stat_unit_impossible" {
			// "2,400+%" / "140+ms" / "1,000sms" cannot be a real unit under any
			// reading, and this is bugs_open/043's original live instance.
			severity = "error"
		}
		issues = append(issues, ValidationIssue{
			Type:        f.Check,
			Category:    "stat_units",
			Severity:    severity,
			Location:    f.Pattern,
			Value:       f.Matched,
			Description: f.Reason,
		})
	}
	return issues
}

// runStatChecks is check 9. Silent no-op when the workflow produced no
// sections_metadata, which is the normal case for this gate's other callers.
func runStatChecks(
	ctx context.Context,
	db *sql.DB,
	collected map[string]interface{},
	config map[string]interface{},
	siteIDStr string,
	doClaims, doUnits bool,
	logger *zap.Logger,
) []ValidationIssue {
	metaField := defaultSectionsMetadataField
	if mf, ok := config["sections_metadata_field"].(string); ok && mf != "" {
		metaField = mf
	}

	claims, usable := collectStatClaims(collected, metaField, logger)

	// "Could not check" must not look like "checked and found nothing".
	//
	// This gate has three configured callers and only one of them has ever run
	// over section-built pages (page-build-handler) — see the measured table at
	// defaultSectionsMetadataField. So the expectation is
	// DECLARED per step rather than guessed from the payload shape — a
	// heuristic here would either false-positive on the other three or go quiet
	// on the one that matters. page-build-handler's validate_content step sets
	// require_sections_metadata:true (migration 219).
	if !usable {
		if configBoolOrDefault(config, "require_sections_metadata", false) {
			logger.Warn("validate_page_content: stat audit could not run — sections_metadata absent on a step that declares it",
				zap.String("field", metaField))
			return []ValidationIssue{{
				Type:        "stat_audit_unavailable",
				Category:    "stat_claims",
				Severity:    "warning",
				Location:    metaField,
				Description: "the stat audit could not run: this step declares require_sections_metadata but no usable sections_metadata arrived, so the page's figures were NOT checked. This is not a clean pass — find out why the section metadata is missing.",
			}}
		}
		return nil
	}

	if len(claims) == 0 {
		return nil
	}

	var issues []ValidationIssue

	if doUnits {
		issues = append(issues, checkStatUnits(claims)...)
	}

	if doClaims && db != nil && siteIDStr != "" {
		if siteID, err := uuid.Parse(siteIDStr); err == nil {
			eb, rowExists := loadEvidenceBaseForStats(ctx, db, siteID, logger)
			if rowExists {
				if eb == nil {
					// writer_block-only row: opted in, nothing registered. Scan
					// against an empty register so every figure is reported —
					// at warning severity, per checkStatClaims.
					eb = &datahelpers.EvidenceBase{}
				}
				issues = append(issues, checkStatClaims(claims, eb, len(eb.Facts) > 0)...)
			}
		}
	}

	if len(issues) > 0 {
		logger.Info("validate_page_content: stat audit findings",
			zap.Int("stat_claims_examined", len(claims)),
			zap.Int("issues", len(issues)))
	}
	return issues
}
