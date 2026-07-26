// FILE: platform/orchestration/actions/discovery_checks/check_unverified_claims_stats.go
//
// The stored-content_data half of the post-deploy claims audit —
// bugs_open/093, the second call site the stat audit never had.
//
// WHY THIS EXISTS. datahelpers.ExtractStatClaims / ScanStatClaims /
// LintStatUnits are generic, but they were invoked from exactly one place:
// validate_page_content check 9, which runs only on the BUILD path
// (page-build-handler -> validate_content -> save_sections). The re-render path
// (rerender_page_sections_action.go) renders STORED content_data with no LLM
// and no validation step at all — 034_page_rerender_agent.sql has no
// validate_content — so a figure that was already persisted, however it got
// there, republishes for ever without ever being compared to the site's
// register. bugs_closed/073 measured aao/index being re-rendered three times in
// one morning while the writer path was blocked: the exact route by which a
// fabrication reprints itself indefinitely.
//
// The council gate escalated this at `high` in rounds 1 and 5 of submission
// 569241fb, naming the family — a mechanism made generic, then guarded at one
// call site while every other call site keeps identical exposure.
//
// WHY HERE RATHER THAN ON THE RE-RENDER PATH ITSELF (093 candidate 1, not 2).
// The re-render path is deliberately cheap and LLM-free; bolting
// validate_page_content onto it would drag in seven checks about freshly
// GENERATED prose and would give a path chosen for having no failure surface a
// new way to fail — on content it did not author, which is precisely 073's
// trap (a page became unbuildable for telling the truth). This check already
// sweeps deployed surfaces, already reads page_components, and reuses the
// existing `claims_unverified` item type, so it incurs no
// verifier_coverage_test.go obligation. It DETECTS after deploy rather than
// preventing; that is the accepted trade, and it is also what covers hand-edits
// and every page that predates the gate.
//
// SCOPE — two scans, two different opt-in predicates, deliberately not one gate:
//
//   - LintStatUnits compares a component against ITSELF and needs no register,
//     so it runs FLEET-WIDE, exactly as it does in the build gate
//     (validate_page_content_stats.go checkStatUnits). Two sites carrying
//     persisted stat fields have no evidence_base row at all — finetuning.uk (3
//     fields) and idea.uk (2), measured 2026-07-26 — so gating the unit lint on
//     opt-in would leave them unaudited on BOTH paths.
//   - ScanStatClaims compares a figure against a register, so it keys on
//     whether the site_specs ROW EXISTS — NOT on ParseEvidenceBase returning
//     non-nil. That distinction is this lane's central landmine:
//     ParseEvidenceBase returns nil when a row carries no facts[] AND no
//     banned_claims[], so a writer_block-only row switches the checker off
//     SILENTLY while the writer block itself keeps working. Three sites were
//     "protected" that way for two days with the checkers blind. Row present
//     but no facts is graded `low` with the gap named in the reason, because a
//     severity must never be allowed to mean "we could not check this".
//
// MEASURED EXPOSURE AT THE TIME OF WRITING (2026-07-26), so a later reader can
// tell drift from a bad change: 64 stat value fields carrying digits across 18
// pages fleet-wide, and the unit lint finds NOTHING — the fleet-wide sweep of
// every persisted *_suffix/_unit/_units value returns five rows, all legitimate
// tool units (leopardess ROI estimator: "employees", "hrs / week per person",
// "per hour", "time saved"; robot-hands cycle-time: "seconds per cycle"), none
// of which is in statDimensionalSuffixes. Site-level chrome carries zero stat
// fields of any kind. So this ships silent on the unit half by measurement, not
// by hope.

package discovery_checks

import (
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// scanStoredStatClaims audits ONE component's stored content_data.
//
// rowExists is whether the site has an evidence_base site_specs row at all;
// eb may be nil even when it does (the writer_block-only case above).
func scanStoredStatClaims(
	component, slotName string,
	contentData map[string]interface{},
	eb *datahelpers.EvidenceBase,
	rowExists bool,
) []unverifiedClaimFinding {
	if len(contentData) == 0 {
		return nil
	}
	// ExtractStatClaims puts this in every finding's Pattern (component.field),
	// which is what a human uses to find the field again.
	if component == "" {
		component = slotName
	}
	if component == "" {
		component = "unknown-component"
	}

	claims := datahelpers.ExtractStatClaims(component, contentData)
	if len(claims) == 0 {
		return nil
	}

	var out []unverifiedClaimFinding

	// (a) The unit lint — fleet-wide, no register consulted.
	for _, f := range datahelpers.LintStatUnits(claims) {
		out = append(out, unverifiedClaimFinding{
			Check: f.Check, SlotName: slotName, Matched: f.Matched,
			Pattern: f.Pattern, Reason: f.Reason, Snippet: f.Snippet,
			Occurrences: f.Occurrences,
			Severity:    statFindingSeverity(f.Check, false),
			Source:      "content_data",
		})
	}

	// (b) The register comparison — opted in means the ROW exists.
	if !rowExists {
		return out
	}
	scanEB := eb
	if scanEB == nil {
		// writer_block-only row: opted in, nothing registered. Scanning against
		// an empty register reports every figure, at `low` — see the header.
		scanEB = &datahelpers.EvidenceBase{}
	}
	hasFacts := len(scanEB.Facts) > 0

	for _, f := range scanEB.ScanStatClaims(claims) {
		reason := f.Reason
		if !hasFacts {
			reason += " — NOTE: this site has an evidence_base row but no machine-readable facts[], " +
				"so nothing could actually be verified. Register its real countables as facts to turn " +
				"this into a check rather than a list."
		}
		out = append(out, unverifiedClaimFinding{
			Check: f.Check, SlotName: slotName, Matched: f.Matched,
			Pattern: f.Pattern, Reason: reason, Snippet: f.Snippet,
			Occurrences: f.Occurrences,
			Severity:    statFindingSeverity(f.Check, hasFacts),
			// A human who clears this in the rendered HTML alone has fixed
			// nothing: the next re-render reprints it from content_data.
			Source: "content_data",
		})
	}

	return out
}

// statFindingSeverity grades a stat finding for the work-item layer.
//
// The build gate's error/warning split (validate_page_content_stats.go) maps
// onto this check's high/medium/low with one deliberate difference: the gate
// BLOCKS a build, whereas this raises a human-review item about content that is
// already live and has been for some time. Nothing here is urgent enough to
// outrank a banned claim, which is a known falsehood the site was explicitly
// told not to repeat.
func statFindingSeverity(check string, hasFacts bool) string {
	switch check {
	case "stat_unit_impossible":
		// "2,400+%", "140+ms", "1,000sms" — a magnitude marker handed a
		// dimensional unit cannot mean anything under any reading, so this is
		// not a judgement call. It is also bug 043's original live instance,
		// and it needs no register, so hasFacts does not enter into it.
		return "high"
	case "unregistered_stat":
		if !hasFacts {
			// Nothing could be verified; see the header. Reporting `medium`
			// here would dress an absence of evidence as evidence of absence.
			return "low"
		}
		return "medium"
	default:
		// unregistered_stat_unpaired — the figure is unsupported but we cannot
		// state what it claims to count, so never grade it as if we could.
		// stat_unit_mismatch — a site owner may genuinely intend "36.6" + "%".
		return "low"
	}
}
