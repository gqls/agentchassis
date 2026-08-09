// FILE: platform/orchestration/actions/revalidate_unverified_claims.go
//
// The review queue's drain for `claims_unverified` (registered in
// reviewRevalidators).
//
// WHY THIS TYPE, AND WHY RETRACTION IS NOT THE AUTO-REWRITE THE CHECK FORBIDS.
// check_unverified_claims.go files HITL-terminal items whose stored `fix` text
// ends "Truth decisions are human — do not auto-rewrite", and the routing note
// at the head of that file states findings terminate at human review because
// auditors raise work items and never rewrite content. That governs the FIX
// path. Retraction is a different act: it never edits copy and never dispatches
// a rewrite. It asks only whether the page STILL asserts something the site's
// register does not support, and stops asserting a finding the current page no
// longer supports. The human's decision about an unsupported claim is
// untouched; what is withdrawn is a claim that has become false.
//
// COUNT THE CLOSERS, NOT JUST THE PRODUCERS. Measured 2026-08-08/09 on live
// clients_db:
//
//   - Producer side: TWO checks converge on this one item_type —
//     check_unverified_claims.go (the HTML banned-claim and unregistered-number
//     scans) and check_unverified_claims_stats.go (the stored-stat scans, which
//     reuse the type deliberately rather than minting a second one). Both file
//     through the same emission block, under item_key `claims:<page_id>` for a
//     page and the single grouped `claims:site_components` for site chrome.
//   - Closer side: `SELECT status, count(*) ... WHERE item_type='claims_unverified'
//     AND status IN ('complete','verified')` returns ZERO ROWS. Nothing has ever
//     closed one — no revalidation block, no `deploy_result` payload, no handler,
//     no dispatch path. HandlerAgent is "" by design.
//
// LIVE POPULATION, same date: 23 selectable items across 7 sites, every one
// `needs_human_review`, every one page-level with a real UUID in spec.page_id.
// No `claims:site_components` item is open today, so the missing-page_id arm
// below is UNEXERCISED rather than unreachable — it is written because that
// item shape exists in the emitter and carries `surface`, not `page_id`.
//
// SAFE DIRECTION, inherited from revalidateNeedsPage and revalidateVoiceTells:
// a wrong `still_holds` costs a human glance, a wrong `resolved` closes a live
// finding. Every arm that cannot answer returns `unknown`, which is
// deliberately non-terminal and leaves the item exactly where it was.
//
// ⚠ THE STANDARD IS EDITABLE, AND HERE THAT BITES HARDER THAN THE VOICE GATE.
// A site's evidence_base register is data. Adding a fact row makes a previously
// unregistered number verifiable, so an item retracts although the copy was
// never touched. That is arguably correct — the claim is now substantiated —
// but a `resolved` stamp on this type is NOT proof the copy was rewritten. To
// tell the two apart, compare page_components.updated_at against the item's
// created_at.
package actions

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func revalidateUnverifiedClaims(ctx context.Context, db *sql.DB, item parkedReviewItem, logger *zap.Logger) revalidationVerdict {
	// spec.page_id only. The item_key is `claims:<page_id>` and would parse, but
	// reading it would make the verdict depend on a prefix convention rather than
	// on the field the producer actually writes — and the site-chrome item's key
	// is the literal `claims:site_components`, which would parse into a page id
	// that cannot exist.
	pageID := specString(item.Spec, "page_id")
	if pageID == "" {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  "spec names no page_id, so the page this finding describes cannot be located; the item_key prefix is deliberately not parsed for it, and on the grouped site-chrome item that key is the literal claims:site_components",
		}
	}

	eb, rowExists, err := checks.LoadEvidenceBase(ctx, db, item.SiteID)
	if err != nil {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("could not load this site's evidence base, so the finding cannot be re-judged: %v", err),
		}
	}
	if !rowExists {
		// The register this finding was measured against is gone, or never
		// existed. The banned-claim half of the audit would still run (that set
		// is fleet-wide, not per-site), but the unregistered-number and
		// stored-stat halves key on this row — so a clean result here means half
		// the question was never asked. Refusing on the whole item rather than
		// per-finding is deliberate: the precision is not worth an arm that can
		// close a live finding when it is wrong.
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  "site has no current evidence_base spec, so the register these claims were measured against no longer exists; withdrawing the register is not evidence the claims were substantiated",
		}
	}

	scans, _, err := checks.ScanDeployedClaims(ctx, db, item.SiteID, pageID, eb, rowExists, logger)
	if err != nil {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("re-scan of the page failed, so the finding cannot be re-judged: %v", err),
		}
	}

	var scan *checks.ClaimsPageScan
	for _, s := range scans {
		if s.PageID == pageID {
			scan = s
			break
		}
	}
	return unverifiedClaimsVerdict(pageID, scan)
}

// unverifiedClaimsVerdict is the decision ladder, split from the database glue
// above so every arm is testable without a database — the same reason
// voiceTellsVerdict and validateTypeFilter are functions rather than inline
// blocks. A nil scan means the re-scan returned no row for the page at all.
//
// Order matters: "still carries claims" must be answered before "some
// components were locked", or a page that still asserts an unsupported claim
// would be reported as unreadable rather than as still holding.
func unverifiedClaimsVerdict(pageID string, scan *checks.ClaimsPageScan) revalidationVerdict {
	if scan == nil {
		// No row came back for this page: it was deleted, or it has no component
		// carrying either rendered_html or content_data. Note this check has never
		// filtered on page status, so — unlike the voice scan — an archived page
		// still comes back and is still judged. A page with nothing built on it
		// cannot be read as copy that was corrected.
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  "page is absent, or has no component carrying rendered html or stored content, so there is no copy to re-judge; an unbuilt page is not evidence the claims were removed",
			Evidence: map[string]interface{}{
				"page_id": pageID,
			},
		}
	}

	if scan.ComponentsExamined == 0 {
		// Reached only when every component on the page is locked — the arm that
		// makes "no findings" mean "nothing was read".
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("every rendered component on this page is human-locked (%d), so no claim was re-read; an empty finding list here means the audit examined nothing", scan.ComponentsSkippedLocked),
			Evidence: map[string]interface{}{
				"page_id":                   pageID,
				"components_skipped_locked": scan.ComponentsSkippedLocked,
			},
		}
	}

	if len(scan.Findings) > 0 {
		byCheck := map[string]int{}
		for _, f := range scan.Findings {
			byCheck[f.Check]++
		}
		return revalidationVerdict{
			Verdict: revalidationStillHolds,
			Reason:  fmt.Sprintf("page %s still carries %d claim(s) the register does not support, across %d examined component(s)", scan.PageName, len(scan.Findings), scan.ComponentsExamined),
			Evidence: map[string]interface{}{
				"page_id":             pageID,
				"page_name":           scan.PageName,
				"findings":            len(scan.Findings),
				"by_check":            byCheck,
				"components_examined": scan.ComponentsExamined,
			},
		}
	}

	if scan.ComponentsSkippedLocked > 0 {
		// Some components were read and are clean, but others were pinned and never
		// read. The claims this item reported may live in the pinned ones, so a
		// close here would be asserting something the scan did not check.
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("the %d component(s) scanned are clean, but %d human-locked component(s) were not read, and the reported claims may sit in those", scan.ComponentsExamined, scan.ComponentsSkippedLocked),
			Evidence: map[string]interface{}{
				"page_id":                   pageID,
				"components_examined":       scan.ComponentsExamined,
				"components_skipped_locked": scan.ComponentsSkippedLocked,
			},
		}
	}

	return revalidationVerdict{
		Verdict: revalidationResolved,
		Reason:  fmt.Sprintf("re-scanned all %d component(s) on page %s against this site's current evidence base and found no unsupported claim; the copy this item flagged no longer asserts one", scan.ComponentsExamined, scan.PageName),
		Evidence: map[string]interface{}{
			"page_id":             pageID,
			"page_name":           scan.PageName,
			"components_examined": scan.ComponentsExamined,
		},
	}
}
