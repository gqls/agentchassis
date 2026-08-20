// FILE: platform/orchestration/actions/section_editor_regulated_guard.go
//
// Refuses a section edit that inserts a REGULATED-IDENTITY claim (register
// CGV-033) into a site with no recorded attestation.
//
// WHY THIS FILE EXISTS — a council objection that was correct. Round 2 of
// correlation `aac38d5b` (edit-quality, high severity) challenged the claim that
// `ScanAllBannedClaimsWithSuppressed` is "the single function every enforcement
// surface already calls", citing landmines that say `validate_content` protects
// the page-BUILD path and nothing else. Verified 2026-08-19 and the objection
// holds: `section_editor_actions.go` contains ZERO references to
// `checkBannedClaims`, `ScanAllBannedClaims` or `scanSectionClaims`, and it
// writes `page_components.rendered_html` directly. So a regulated claim inserted
// by a section edit — the copy-editor's own write path — bypassed the entire
// family, which is exactly the "bypassed, re-run, or edited" scenario the guard
// was built for.
//
// It is also the SAME CLASS the council caught once before on this very file.
// `section_editor_actions.go`'s own comment records it: "both floors were wired
// only into SavePageSectionsAction, so this path — the one decomposition exists
// to enable — bypassed them". A guard wired into the whole-page save and not into
// the per-slot editor is a guard with a documented hole in it.
//
// WHY IT SCANS THE REGULATED FAMILY ALONE, and not the full claims set. Turning
// the whole banned-claim gate on in the editor would change refusal behaviour for
// every copy edit the estate makes, on a seam another lane owns, for reasons that
// have nothing to do with this change. `ScanRegulatedClaims` is the minimum that
// closes the named hole: claims nobody should be inserting by any route, on a site
// that has not proved it may make them.
//
// WHAT IT DOES NOT COVER, stated rather than left to be discovered:
// **chrome is still unguarded.** `site_components.rendered_html` is written by
// `render_site_components_action.go` and `chrome_link_policy.go`, and neither
// scans claims (verified: the only file touching `site_components` that scans is
// `discovery_checks/check_unverified_claims.go`, which is a post-deploy AUDIT, not
// a gate). The footer is the traditional home of an "authorised and regulated by
// the FCA, FRN …" line, so this is a real residual — narrowed only by chrome
// content being template-plus-config rather than free LLM prose, with the
// compliance line coming from operator-set `site_config.chrome.compliance_lines`.
// An operator writing that line is a deliberate human act; an agent inventing it
// is what this guard stops. Recorded in CGV-033's verify-later.

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// refuseRegulatedIdentityEdit returns a non-nil error when newHTML asserts a
// regulated identity for a site that has no complete attestation.
//
// It fails CLOSED on the claim and OPEN on infrastructure: an unreadable evidence
// base means the site is unattested, which is the safe reading — a site whose
// register cannot be read has certainly not proved it is authorised.
func refuseRegulatedIdentityEdit(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	pageName, slotName, newHTML string,
	logger *zap.Logger,
) error {
	if newHTML == "" {
		return nil
	}
	blocks := datahelpers.ExtractAssertionText(newHTML)
	if len(blocks) == 0 {
		return nil
	}

	eb := loadEvidenceBase(ctx, params.DB, siteID, logger) // nil = no register
	findings := datahelpers.ScanRegulatedClaims(blocks, eb)
	if len(findings) == 0 {
		return nil
	}

	// Log every match, not just the first: an edit that trips two patterns is
	// more likely to be a deliberate rewrite than a stray phrase, and the
	// operator reading this needs to see the whole shape.
	for _, f := range findings {
		logger.Warn("section edit REFUSED: regulated-identity claim on an unattested site",
			zap.String("site_id", siteID.String()),
			zap.String("page", pageName),
			zap.String("slot", slotName),
			zap.String("pattern", f.Pattern),
			zap.String("matched", f.Matched),
			zap.String("snippet", f.Snippet))
	}

	return fmt.Errorf(
		"section edit refused: this edit asserts a regulated identity (%q) and %s has no "+
			"recorded regulated attestation. A regulated-firm posture is a legal position, not a "+
			"content choice. If the site IS authorised, record an attestation (firm name, FRN, who "+
			"checked it, what they saw) with scripts/regulated/record_attestation.py and re-run this "+
			"edit; see register CGV-033",
		findings[0].Matched, pageName)
}
