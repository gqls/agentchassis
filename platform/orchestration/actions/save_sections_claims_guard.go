// FILE: platform/orchestration/actions/save_sections_claims_guard.go
//
// The CLAIMS FLOOR: claims checking at the PERSISTENCE point (bugs_open/149 C1).
//
// WHY THIS FILE EXISTS. The owner's requirement was that copy written off a
// discovery check "must write copy that follows the claims checking like
// everything else". The measurement behind it turned out to be about a different
// seam than the bug file assumed, so the fix is at that seam instead.
//
// Measured live 2026-07-30 (each figure re-run, none carried forward):
//
//   - SIX live agents persist page body sections through save_page_sections —
//     page-build-handler, pageflow-builder, page-rebuild, page-rerender,
//     site-work-orchestrator, tool-recreation-handler. TWO of them run
//     validate_page_content (page-build-handler, tool-recreation-handler). The
//     other four have never had a claims gate of any kind.
//   - `page-content-writer`, which bugs_open/149 named as the place the gate was
//     missing, PERSISTS NOTHING. Its workflow ends
//     compile_page → complete_workflow{output_field: page_content}, and it is
//     CALLED by those same four parents. Adding a gate there would have gated a
//     value on its way to a caller, not a write. check_empty_sections.go:7
//     records the same route being corrected once already:
//     `HandlerAgent: "page-build-handler" (was "page-content-writer")`.
//
// So the honest denominator is the persistence seam, and this is deliberately
// the same argument — and the same shape — as its neighbour
// save_sections_link_repair.go:20-29:
//
//	"Persistence is the one chokepoint every body-section writer passes
//	 through … Putting the repair here makes an unrepaired section unsaveable
//	 whatever the workflow config says, which is the property the config-shaped
//	 defect above actually needs."
//
// A config-shaped fix (add a validate step to four more agents) would have to be
// remembered by every future workflow author and would still leave the four
// agents unprotected until each was edited. A floor cannot be forgotten.
//
// THE SEVERITY RULE IS DRIVEN BY SEVERITY, NOT BY CHECK NAME. That is what keeps
// it correct as checks are added to the engine:
//
//	blocker (banned claim — a known falsehood, fleet-wide + per-site)
//	    → REFUSE the save, exactly as the ownership, content-regression and
//	      interactivity guards in the same function already do.
//	error   (unregistered number — opt-in on evidence_base, editorial page types
//	         exempt per bugs_open/102)
//	    → RECORD durably, allow the save.
//
// The numeric half records rather than refuses because its false-positive rate
// is the stated reason it is never a blocker in the engine either
// (validate_page_content.go:1079-1082): "number extraction has false positives by
// design, and error already routes the page to a human". Refusing on it at a
// seam four agents never opted into would convert a content-quality signal into
// a fleet-wide build outage. Recording it is also what gives the platform the
// write-path claims measurement bugs_open/149 C3 wanted, without seating a new
// check anywhere.
//
// BLAST RADIUS, MEASURED BEFORE THIS SHIPPED — not left for a reviewer to ask
// about (the ground bugs_open/124 drew a REJECTED verdict on). cmd/claimscan
// (CLM-014, the same engine) over the COMPLETE live corpus, 949 components across
// 14 sites, each site against its own register, 2026-07-30:
//
//	banned claims (would refuse a save):  3 components on 2 sites — 0.32%
//	    webdesign.co.uk  tool-blueprint-compiler / ported-page      "never invents"
//	    robot-hands.com  how-it-works / generic-text-block   "independently verified"
//	    robot-hands.com  gripper-catalog / info-card-grid    "independently verified"
//	unregistered numbers (record only): 59 findings on the 4 armed sites
//
// All three are genuine known falsehoods; the robot-hands pair IS bugs_open/147,
// already filed. So the pages this floor can strand are three pages that are
// asserting something untrue today. That is the whole exposure, and it is why
// refusing is affordable here. RE-MEASURE rather than quoting this: the corpus
// was 908 components on 07-28 and 949 on 07-30.
//
// SCOPE BOUNDARY, stated rather than implied. This covers save_page_sections. It
// is NOT fleet-wide coverage of page_components: ApplySectionEditAction,
// create_report_page and rebuild_blog_listing also persist LLM-authored prose,
// exactly as the link-repair file already admits at
// save_sections_link_repair.go:366-372. create_report_page belongs to
// bugs_open/149 C2's open question (report-builder is the only agent setting
// check_claims:false — find out whether that is deliberate before switching it
// on). bugs_open/123's content-creator path has no site and no page row, so this
// seam cannot reach it at all.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// claimsFloorErrorCode is DELIBERATELY distinct from validationDetailErrorCode
// and linkRepairErrorCode, for the reason recorded at
// validate_page_content.go:619-624: the three rows answer different questions,
// and every query already written against the gate's code must keep returning
// exactly what it returned before this change. A shared code would also make
// "which path caught this" unanswerable, which is the drift bugs_open/097
// exists to keep answerable.
const claimsFloorErrorCode = "CONTENT_CLAIMS_FLOOR_DETAIL"

// scanSectionClaims is the pure seam: no database, no logging, no config. It is
// where the behaviour is tested (save_sections_claims_guard_test.go), the same
// way repairSectionLinks owns the repair semantics for its own file.
//
// The SCAN semantics are not reimplemented here and must never be — this calls
// the identical engine the deploy gate calls (validate_page_content.go:326-337)
// and that cmd/claimscan predicts with. A second implementation of a claims scan
// is bugs_open/093's shape exactly: "one call site of a shared judgement gets the
// rigorous fix; the sibling stays heuristic".
//
// Sections are scanned INDIVIDUALLY rather than as one joined document. Two
// reasons, both load-bearing: ExtractAssertionText parses assertion text nodes,
// and joining raw section HTML can fuse a trailing fragment of one section to
// the opening of the next into a sentence neither section contains; and a
// finding must name the section that carries it, or the operator record cannot
// say which component to fix.
func scanSectionClaims(
	sections []SectionData,
	eb *datahelpers.EvidenceBase,
	fleetWide bool,
	surface datahelpers.ClaimSurface,
	logger *zap.Logger,
) (blockers, errors []claimsFloorFinding) {
	for i := range sections {
		if sections[i].HTML == "" {
			continue
		}
		blocks := datahelpers.ExtractAssertionText(sections[i].HTML)
		if len(blocks) == 0 {
			continue
		}

		for _, issue := range checkBannedClaims(blocks, eb, fleetWide, "", logger) {
			blockers = append(blockers, claimsFloorFinding{
				Section: sections[i].ComponentName,
				Issue:   issue,
			})
		}

		// The numeric half stays strictly opt-in on the register's presence,
		// mirroring the gate (validate_page_content.go:328). A site with no
		// evidence_base is scanned by the fleet-wide banned set above and by
		// nothing else — that asymmetry is deliberate and documented in
		// bugs_open/104; do not simplify it away.
		if eb != nil {
			for _, issue := range checkUnregisteredNumbers(blocks, eb, surface) {
				errors = append(errors, claimsFloorFinding{
					Section: sections[i].ComponentName,
					Issue:   issue,
				})
			}
		}
	}
	return blockers, errors
}

// claimsFloorFinding pairs an engine issue with the section that carries it.
// The section name is the whole point of the record: "this page has a banned
// claim" is not actionable, "this component has one" is.
type claimsFloorFinding struct {
	Section string
	Issue   ValidationIssue
}

// claimsGuardBeforePersist runs the floor over the sections about to be written.
//
// Returns a non-nil error when a blocker was found, which SavePageSectionsAction
// returns unchanged — refusing the save. Errors (the numeric half) never fail the
// save; they are persisted and reported.
//
// Called from SavePageSectionsAction AFTER repairSectionsBeforePersist, so the
// scan sees the bytes that will actually be persisted rather than a string the
// repair pass is about to change, and BEFORE the regression guards so a page
// refused for a falsehood is refused for that reason rather than for whatever
// the guards notice next.
func claimsGuardBeforePersist(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	pageID uuid.UUID,
	domain, pageName, pageURL string,
	sections []SectionData,
	logger *zap.Logger,
) error {
	if params.DB == nil || len(sections) == 0 {
		return nil
	}

	// The reversal levers, same names and same ON defaults as the gate's
	// (validate_page_content.go:220-247). They default ON because an
	// off-by-default floor would be inert and 149 C1 would still be live; DB
	// config is live-immediately, so the behaviour can be withdrawn fleet-wide
	// in seconds without waiting for an image roll. That is the same argument
	// repair_internal_links and check_claims_fleet_wide already settle.
	if !configBoolOrDefault(params.StepConfig.Config, "check_claims", true) {
		logger.Info("SavePageSectionsAction: claims floor disabled by step config",
			zap.String("page_name", pageName))
		return nil
	}
	fleetWide := configBoolOrDefault(params.StepConfig.Config, "check_claims_fleet_wide", true)

	eb := loadEvidenceBase(ctx, params.DB, siteID, logger)

	// A DISARMED FLOOR MUST NOT BE SILENT — the same reasoning the council's
	// compliance seat applied to the gate (validate_page_content.go:312-325). A
	// register-less site loses its ONLY banned-claim protection this way, and
	// nothing else would say so.
	if !fleetWide {
		logger.Warn("claims floor: fleet-wide banned-claim set is DISABLED for this save "+
			"(check_claims_fleet_wide=false) — per-site patterns only, and a site with no "+
			"evidence_base is scanned by nothing (bugs_open/104)",
			zap.String("site_id", siteID.String()),
			zap.Bool("site_has_register", eb != nil))
	}

	surface := datahelpers.ClaimSurface{PageType: loadPageTypeForClaims(ctx, params.DB, pageID, logger)}

	blockers, errs := scanSectionClaims(sections, eb, fleetWide, surface, logger)
	if len(blockers) == 0 && len(errs) == 0 {
		return nil
	}

	// Distinctive compiled string: this is the pod-grep marker for the deploy.
	logger.Info("SavePageSectionsAction: claims floor findings before persist",
		zap.String("page_name", pageName),
		zap.String("domain", domain),
		zap.Int("blockers", len(blockers)),
		zap.Int("errors", len(errs)),
		zap.Int("sections", len(sections)))

	writeClaimsFloorLog(ctx, params, siteID, domain, pageName, pageURL,
		blockers, errs, logger)

	if len(blockers) == 0 {
		return nil
	}

	for _, f := range blockers {
		logger.Warn("SavePageSectionsAction: CLAIMS FLOOR BLOCKED — banned claim in section content",
			zap.String("page_name", pageName),
			zap.String("section", f.Section),
			zap.String("matched", f.Issue.Value),
			zap.String("detail", f.Issue.Description))
	}

	return fmt.Errorf(
		"claims floor blocked: %d banned claim(s) in the content for page %s "+
			"(first: section %q asserts %q). A banned claim is a statement no site may make about "+
			"itself, so this is refused rather than published. Fix the copy, or withdraw the set for "+
			"this step with check_claims_fleet_wide:false. See bugs_open/149 C1.",
		len(blockers), pageName, blockers[0].Section, blockers[0].Issue.Value)
}

// loadPageTypeForClaims reads pages.page_type for the page being written.
//
// An unresolved type is "" — UNKNOWN — and unknown SCANS (datahelpers.ClaimSurface,
// bugs_closed/102). That direction is deliberate: a page-type-blind scan is the
// status quo, whereas a page-type-blind SKIP would be new blindness. So a failed
// read costs false positives on the recording half, never a missed banned claim.
func loadPageTypeForClaims(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) string {
	var pageType sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT page_type FROM pages WHERE id = $1`, pageID).Scan(&pageType); err != nil {
		if err != sql.ErrNoRows {
			logger.Warn("claims floor: could not read page_type — scanning as UNKNOWN",
				zap.String("page_id", pageID.String()), zap.Error(err))
		}
		return ""
	}
	return pageType.String
}

// writeClaimsFloorLog persists what the floor found, on BOTH paths — the refusal
// and the record-and-allow.
//
// A pod log line is not a record (bugs_open/071 gap 3): a page that shipped with
// an unregistered number must be findable a day later, and pod logs roll. This is
// the durable half of the "record" severity rule, and without it the numeric
// scan would be a Warn nobody ever reads.
//
// This writes a work RECORD, not a work ITEM, and deliberately so — the same
// judgement writeLinkRepairLog made and for the same reason: bugs_open/083
// measured detections filed as items and fixed zero times, and bugs_open/077
// warns specifically against filing items whose handler has no remit. There is no
// handler that can rewrite a claim; a human must. bugs_open/149 C3's detector
// (unverified_claims → needs_human_review) is the path that raises items for
// exactly this, and it fired automatically on 2026-07-29 — so duplicating it here
// would double-file the same finding under two mechanisms.
//
// Best-effort: a logging failure must never mask the refusal it describes, and
// must never fail a save whose content is acceptable.
func writeClaimsFloorLog(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	domain, pageName, pageURL string,
	blockers, errs []claimsFloorFinding,
	logger *zap.Logger,
) {
	if params.DB == nil {
		return
	}

	toMaps := func(fs []claimsFloorFinding) []map[string]string {
		out := make([]map[string]string, 0, len(fs))
		for _, f := range fs {
			out = append(out, map[string]string{
				"section":     f.Section,
				"type":        f.Issue.Type,
				"category":    f.Issue.Category,
				"severity":    f.Issue.Severity,
				"value":       f.Issue.Value,
				"location":    f.Issue.Location,
				"description": f.Issue.Description,
			})
		}
		return out
	}

	// The outcome is recorded explicitly rather than inferred from the counts.
	// "blockers > 0 means it was refused" is true today and would silently stop
	// being true the moment anyone adds a lever; a stored outcome cannot drift
	// from what actually happened.
	outcome := "recorded"
	if len(blockers) > 0 {
		outcome = "refused"
	}

	contextData := map[string]interface{}{
		"outcome":       outcome,
		"page_name":     pageName,
		"page_url":      pageURL,
		"blocker_count": len(blockers),
		"error_count":   len(errs),
		"blockers":      toMaps(blockers),
		"errors":        toMaps(errs),
	}
	contextJSON, err := json.Marshal(contextData)
	if err != nil {
		logger.Warn("claims floor: failed to marshal finding context", zap.Error(err))
		return
	}

	agentType := params.AgentType
	if agentType == "" {
		agentType = "unknown"
	}

	severity := "warning"
	if len(blockers) > 0 {
		severity = "error"
	}

	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO agent_error_log
		    (site_id, domain, agent_type, step_name, action,
		     error_message, error_code, severity, context)
		VALUES ($1, NULLIF($2, ''), $3, $4, 'save_page_sections', $5, $6, $7, $8::jsonb)`,
		siteID, domain, agentType, saveSectionsStepName(params),
		fmt.Sprintf("Claims floor %s: %d banned claim(s), %d unregistered number(s) on page %s",
			outcome, len(blockers), len(errs), pageName),
		claimsFloorErrorCode, severity, string(contextJSON),
	); err != nil {
		logger.Warn("claims floor: failed to write finding record", zap.Error(err))
	}
}
