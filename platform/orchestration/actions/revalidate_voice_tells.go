// FILE: platform/orchestration/actions/revalidate_voice_tells.go
//
// The review queue's drain for `voice_tells` (registered in reviewRevalidators).
//
// WHY THIS TYPE, AND WHY IT IS NOT THE AUTO-REWRITE THE CHECK FORBIDS.
// check_voice_tells.go files HITL-terminal items whose stored `fix` text reads
// "never an unreviewed auto-rewrite", and bugs_open/033 quotes that line as
// evidence the type was filed correctly for human review. Retraction is a
// different act: it never edits copy and never dispatches a rewrite. It asks
// only whether the page STILL trips the site's own voice gate, and stops
// asserting a finding the current page no longer supports. The human's decision
// about how to fix machine-written prose is untouched; what is withdrawn is a
// claim that has become false.
//
// COUNT THE CLOSERS, NOT JUST THE PRODUCERS (the lesson this lane paid for).
// Measured 2026-08-08 on live clients_db:
//
//   - Producer side: one producer, check_voice_tells.go:142. No other Go file
//     emits ItemType "voice_tells".
//   - Closer side: `SELECT status, count(*) ... WHERE item_type='voice_tells'
//     AND status IN ('complete','verified')` returns ZERO ROWS. Nothing has ever
//     closed one — no revalidation block, no handler, no dispatch path.
//     HandlerAgent is "" by design. This is the check that a previous adopter in
//     this lane skipped, shipping a duplicate closer that 14 council seats
//     accepted; it is run first here, and it comes back empty.
//
// LIVE POPULATION, same date: 25 items, all `needs_human_review`, all filed
// 2026-07-17, all on leopardessconsulting.co.uk. Every one has a page that
// still exists and is active/deployed, with unlocked rendered components and no
// locked components at all — so the three refusal arms below are UNEXERCISED on
// today's data, not proven unreachable. They are written because the population
// will change, and because a scan that examined nothing returns the same empty
// findings list as a page that was genuinely cleaned up. 13 of the 25 pages have
// had a component updated since the item was filed, so this is expected to
// judge real change rather than run inert.
//
// SAFE DIRECTION, inherited from revalidateNeedsPage: a wrong `still_holds`
// costs a human glance, a wrong `resolved` closes a live finding. Every arm
// that cannot answer returns `unknown`, which is deliberately non-terminal and
// leaves the item exactly where it was.
package actions

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func revalidateVoiceTells(ctx context.Context, db *sql.DB, item parkedReviewItem, logger *zap.Logger) revalidationVerdict {
	// spec.page_id only. The item_key is `voice:<page_id>` and would parse, but
	// reading it would make the verdict depend on a prefix convention rather than
	// on the field the producer actually writes — the mistake §0b of this lane's
	// handoff spent a session refuting.
	pageID := specString(item.Spec, "page_id")
	if pageID == "" {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  "spec names no page_id, so the page this finding describes cannot be located; the item_key prefix is deliberately not parsed for it",
		}
	}

	gate, err := checks.LoadVoiceGate(ctx, db, item.SiteID)
	if err != nil {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("could not load this site's voice gate, so the finding cannot be re-judged: %v", err),
		}
	}
	if gate == nil {
		// The site turned its voice gate off, or never had one. The question this
		// item asks is defined BY that gate — without it there is no standard to
		// re-measure against, so we do not pretend to answer. Not `resolved`:
		// opting out of the audit is not evidence the copy improved.
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  "site has no enabled voice_gate, so the standard this finding was measured against no longer exists; opting out is not evidence the copy was fixed",
		}
	}

	scans, err := checks.ScanVoiceTells(ctx, db, item.SiteID, pageID, gate, logger)
	if err != nil {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("re-scan of the page failed, so the finding cannot be re-judged: %v", err),
		}
	}

	var scan *checks.VoicePageScan
	for _, s := range scans {
		if s.PageID == pageID {
			scan = s
			break
		}
	}
	return voiceTellsVerdict(pageID, scan)
}

// voiceTellsVerdict is the decision ladder, split from the database glue above
// so every arm is testable without a database — the same reason validateTypeFilter
// is a function rather than four lines inline. A nil scan means the re-scan
// returned no row for the page at all.
//
// Order matters: "found tells" must be answered before "some components were
// locked", or a page that still trips the gate would be reported as unreadable
// rather than as still holding.
func voiceTellsVerdict(pageID string, scan *checks.VoicePageScan) revalidationVerdict {
	if scan == nil {
		// No row came back for this page: it was deleted, left active/deployed, or
		// has no rendered components. A page that is not being served cannot be
		// read as prose that was rewritten, so this refuses rather than closes.
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  "page is absent, not active/deployed, or has no rendered components, so there is no copy to re-judge; an unserved page is not evidence the prose was fixed",
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
			Reason:  fmt.Sprintf("every rendered component on this page is human-locked (%d), so nothing was scanned; an empty finding list here means the scan read nothing, not that the copy is clean", scan.ComponentsSkippedLocked),
			Evidence: map[string]interface{}{
				"page_id":                   pageID,
				"components_skipped_locked": scan.ComponentsSkippedLocked,
			},
		}
	}

	if len(scan.Findings) > 0 {
		byCheck := map[string]int{}
		for _, f := range scan.Findings {
			if name, ok := f["check"].(string); ok {
				byCheck[name]++
			}
		}
		return revalidationVerdict{
			Verdict: revalidationStillHolds,
			Reason:  fmt.Sprintf("page %s still trips this site's voice gate: %d finding(s) across %d examined component(s)", scan.PageName, len(scan.Findings), scan.ComponentsExamined),
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
		// read. The tells this item reported may live in the pinned ones, so a
		// close here would be asserting something the scan did not check.
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason:  fmt.Sprintf("the %d component(s) scanned are clean, but %d human-locked component(s) were not read, and the reported tells may sit in those", scan.ComponentsExamined, scan.ComponentsSkippedLocked),
			Evidence: map[string]interface{}{
				"page_id":                   pageID,
				"components_examined":       scan.ComponentsExamined,
				"components_skipped_locked": scan.ComponentsSkippedLocked,
			},
		}
	}

	return revalidationVerdict{
		Verdict: revalidationResolved,
		Reason:  fmt.Sprintf("re-scanned all %d rendered component(s) on page %s against this site's current voice gate and found no tells; the copy this item flagged no longer reads machine-written", scan.ComponentsExamined, scan.PageName),
		Evidence: map[string]interface{}{
			"page_id":             pageID,
			"page_name":           scan.PageName,
			"components_examined": scan.ComponentsExamined,
		},
	}
}
