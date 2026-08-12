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
//   - Producer side: ONE producer, UnverifiedClaimsCheck in
//     check_unverified_claims.go. item_key is `claims:<page_id>` for a page and
//     the single grouped `claims:site_components` for site chrome.
//
//     > **CORRECTED 2026-08-09, and it was my error.** The first version of this
//     > header (and the council submission built on it) said "TWO checks converge
//     > on this one item_type", naming check_unverified_claims_stats.go as a
//     > second producer, and invoked the owner's 2026-08-02 converging-producers
//     > ruling to argue no RFC was needed. **There is no convergence, so that
//     > ruling never applied.** check_unverified_claims_stats.go registers no
//     > check, has no init(), and emits no WorkItemSpec — it is a HELPER FILE
//     > whose scanStoredStatClaims() is called from inside ScanDeployedClaims
//     > (check_unverified_claims.go:385 for pages, :427 for site chrome) and
//     > nowhere else in production. Its own header line "reuses the existing
//     > claims_unverified item type" describes reusing the type it contributes
//     > findings to, not filing items itself. The council's editquality seat
//     > raised this as a gating objection; its feared consequence — the
//     > revalidator judging by the wrong producer's predicate — is refuted by
//     > exactly that call graph: there is ONE scan, and both halves live in it.
//
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
// unregistered number verifiable, so an item would retract although the copy was
// never touched.
//
// OWNER RULING 2026-08-09 — that hole is CLOSED IN CODE, not documented as a
// caveat. The copy-changed gate at the foot of the ladder requires positive
// evidence that the PAGE moved (an examined component with
// page_components.updated_at later than the item's created_at) before anything
// may close. A register edit alone can no longer retract a finding.
//
// The owner chose this over shipping as-is, over downgrading instead of closing,
// and over abandoning the change, after the council's `compliance` seat gated on
// it in TWO successive rounds. Its formulation is the one worth keeping: THE
// REGISTER PROVES PROVENANCE, NOT CORRECTNESS. The machine can confirm a number
// is registered; it cannot confirm the number is true. So a careless register
// entry must not be able to close a human-review row about a factual claim —
// and now it cannot.
//
// ⚠ A `resolved` stamp on this type therefore means "the copy changed AND the
// page is now clean". It still does not mean a human agreed the new copy is
// TRUE — that was never in scope, and the gate does not claim it.
package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
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
	return unverifiedClaimsVerdict(pageID, item.CreatedAt, flaggedClaimsFromSpec(item.Spec), scan)
}

// flaggedClaim is ONE entry from the work item's own spec.findings[] — the text
// this finding actually cited, and the slot it cited it from. It is the item's
// record of what was wrong, written by the producer at filing time, and it is
// the only thing that makes a CLAIM-GRANULAR judgement possible: the scan can
// say the page is clean now, but only the item can say what "unclean" meant.
type flaggedClaim struct {
	Check   string
	Slot    string
	Matched string
}

// flaggedClaimsFromSpec reads spec.findings[]. Every live claims_unverified item
// carries it — 86 of 86 findings across the population on 2026-08-11 had a
// non-empty `matched` AND `snippet` — but this returns whatever is there and
// lets the ladder refuse, rather than assuming the shape.
func flaggedClaimsFromSpec(spec map[string]interface{}) []flaggedClaim {
	raw, _ := spec["findings"].([]interface{})
	out := make([]flaggedClaim, 0, len(raw))
	for _, r := range raw {
		m, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		str := func(k string) string {
			s, _ := m[k].(string)
			return s
		}
		out = append(out, flaggedClaim{Check: str("check"), Slot: str("slot_name"), Matched: str("matched")})
	}
	return out
}

// claimStillOnPage asks whether the exact text this finding cited is still
// sitting in the component it was cited from.
//
// The second return is whether the question could be ANSWERED at all, and it is
// separate from the answer on purpose: "the text is gone" and "I could not look"
// are the same empty result, and collapsing them is the no-op-that-reads-as-
// success shape this whole file exists to avoid.
//
// Unjudgeable when the finding carries no `matched` text, or when its slot is
// absent from the examined set — a slot goes absent because the component was
// deleted, renamed, or human-locked, and none of those is evidence the claim was
// removed.
//
// Case-insensitive substring, scoped to the slot. Deliberately crude in the SAFE
// direction: a token that matches something unrelated in its own slot produces a
// refusal (non-terminal, a human still sees the item), never a closure.
//
// PREDICATE PARITY (council round 6, `reuse_agent`): the search runs over
// datahelpers.ExtractAssertionText, which is the SAME extraction the emit side
// scans (check_unverified_claims.go's scanComponentClaims calls it before handing
// blocks to ScanBannedClaims / ScanUnregisteredNumbers). Searching raw markup here
// would mean the two ends of an item's life judge different text — the exact
// divergence this lane's shared-scanner design exists to prevent — and it would
// let a token match a class name, an id or a style value that no reader sees.
//
// ⚠ HONEST NOTE, so nobody re-litigates this as a fix: on the live population it
// changes NOTHING. Measured 2026-08-12 against both sides — false refusals on the
// 8 closed items 2 with raw markup and 2 with prose only; sensitivity on the
// still-holding population 40/41 either way. The two survivors ("5", "97") sit in
// the prose, not the markup. This is a correctness-and-reuse change, not a
// measured improvement, and it is made because parity is the design rather than
// because it moved a number.
func claimStillOnPage(scan *checks.ClaimsPageScan, c flaggedClaim) (present, judgeable bool) {
	if strings.TrimSpace(c.Matched) == "" {
		return false, false
	}
	raw, ok := scan.ExaminedTextBySlot[c.Slot]
	if !ok {
		return false, false
	}
	needle := strings.ToLower(strings.TrimSpace(c.Matched))
	for _, block := range datahelpers.ExtractAssertionText(raw) {
		if strings.Contains(strings.ToLower(block), needle) {
			return true, true
		}
	}
	return false, true
}

// unverifiedClaimsVerdict is the decision ladder, split from the database glue
// above so every arm is testable without a database — the same reason
// voiceTellsVerdict and validateTypeFilter are functions rather than inline
// blocks. A nil scan means the re-scan returned no row for the page at all.
//
// Order matters: "still carries claims" must be answered before "some
// components were locked", or a page that still asserts an unsupported claim
// would be reported as unreadable rather than as still holding.
//
// filedAt is the work item's created_at, and it is what the copy-changed gate
// below compares against. A zero filedAt makes that gate refuse, which is the
// safe direction: an item whose filing date we cannot establish is one whose
// "has the page moved since?" question has no answer.
func unverifiedClaimsVerdict(pageID string, filedAt time.Time, flagged []flaggedClaim, scan *checks.ClaimsPageScan) revalidationVerdict {
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

	// THE CLAIM-GRANULAR GATE — council round 5, `compliance` HIGH, 2026-08-11.
	//
	// It NARROWS the owner's gate below; it does not replace it. Both must pass.
	// That is a deliberate governance choice as much as an engineering one: adding
	// a stricter condition in front of an owner-mandated control needs no new
	// ruling, whereas removing the timestamp comparison would. If the flagged text
	// being demonstrably gone should be allowed to STAND IN FOR the timestamp (it
	// is the stronger evidence of the two — a token that has left the slot proves
	// the copy moved, where a timestamp only asserts it), that is the owner's call
	// and is recorded as an open question rather than taken here.
	//
	// WHAT compliance ACTUALLY OBJECTED TO, three rounds running: the gate below
	// verifies THE PAGE MOVED, not THAT THE FLAGGED CLAIM WAS ADDRESSED. A typo
	// fix, a CSS tweak or an edit to an adjacent field bumps page_components
	// .updated_at, and the item then closes provided the current scan happens to
	// find nothing — on the platform's highest-stakes content-integrity type,
	// designed HITL-terminal because truth decisions are human.
	//
	// ⚠ THE SEAT'S OWN SUGGESTED FIX WAS MEASURED AND REJECTED, and this is the
	// part worth reading. It proposed comparing "the specific finding's cited
	// snippet". Measured against the live population on 2026-08-11: on items that
	// STILL HOLD — where the claim is by definition still on the page — a
	// slot-scoped SNIPPET comparison found the claim in only 18 of 41 cases, while
	// the slot-scoped MATCHED token found it in 40 of 41. The snippet is long
	// enough that any whitespace or markup churn breaks the match, and a broken
	// match reads as "the copy changed" — which grants `resolved`. Shipping the
	// seat's literal suggestion would have made the gate FAIL OPEN on roughly half
	// of all claims, i.e. strictly worse than the timestamp it was meant to
	// strengthen. The token is the version that discriminates, so the token is what
	// ships, and the objection is answered rather than transcribed.
	//
	// Order: this sits after "still carries findings" (a page that still trips the
	// check is reported as still holding, never as unreadable) and before the
	// timestamp gate, because "the text you flagged is still there" is the more
	// informative refusal of the two.
	if len(flagged) == 0 {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason: fmt.Sprintf(
				"page %s no longer trips the check, but this item records no finding text, so there is nothing to confirm was removed and a clean scan cannot be told from a moved standard",
				scan.PageName),
			Evidence: map[string]interface{}{
				"page_id": pageID, "page_name": scan.PageName,
				"components_examined": scan.ComponentsExamined,
			},
		}
	}
	var stillPresent, unjudgeable []string
	for _, c := range flagged {
		present, judgeable := claimStillOnPage(scan, c)
		switch {
		case !judgeable:
			unjudgeable = append(unjudgeable, fmt.Sprintf("%s:%s", c.Check, c.Slot))
		case present:
			stillPresent = append(stillPresent, fmt.Sprintf("%s:%s", c.Check, c.Slot))
		}
	}
	if len(stillPresent) > 0 {
		return revalidationVerdict{
			Verdict: revalidationStillHolds,
			Reason: fmt.Sprintf(
				"page %s no longer trips the check, but %d of the %d text(s) this finding cited are STILL in the component they were cited from — so the standard moved, not the copy; a claim that stopped being flagged while its words are untouched has not been addressed",
				scan.PageName, len(stillPresent), len(flagged)),
			Evidence: map[string]interface{}{
				"page_id": pageID, "page_name": scan.PageName,
				"components_examined":      scan.ComponentsExamined,
				"flagged_texts":            len(flagged),
				"flagged_texts_still_here": stillPresent,
			},
		}
	}
	if len(unjudgeable) > 0 {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason: fmt.Sprintf(
				"page %s no longer trips the check, but %d of the %d text(s) this finding cited could not be re-checked — the slot they were cited from was not among the components examined (deleted, renamed or human-locked), and an absent component is not evidence the claim was removed",
				scan.PageName, len(unjudgeable), len(flagged)),
			Evidence: map[string]interface{}{
				"page_id": pageID, "page_name": scan.PageName,
				"components_examined":       scan.ComponentsExamined,
				"flagged_texts":             len(flagged),
				"flagged_texts_unjudgeable": unjudgeable,
			},
		}
	}

	// THE COPY-CHANGED GATE — OWNER RULING 2026-08-09, and the last thing between
	// a clean scan and a closed item.
	//
	// Everything above establishes that the page no longer trips the check. That is
	// NOT the same as the page having been fixed, because the standard is editable
	// data: adding a fact to the site's evidence_base makes a previously
	// unregistered number verifiable, and the finding evaporates with the copy
	// untouched. The council's `compliance` seat gated on exactly this (twice), and
	// its formulation is the one to keep in mind — THE REGISTER PROVES PROVENANCE,
	// NOT CORRECTNESS. A machine cannot tell a substantiated claim from a
	// carelessly-registered one, so it must not be the thing that closes a
	// human-review row on the strength of a register edit alone.
	//
	// So: require positive evidence that the PAGE moved. If no examined component
	// has been touched since the finding was filed, something other than the page
	// changed, and this refuses — non-terminally, leaving the item exactly where a
	// human can still see it.
	//
	// The owner chose this over three alternatives (ship as-is; downgrade instead
	// of closing; abandon the change) precisely because it converts a policy
	// argument into a mechanical guarantee that a reviewer of the CALLER can see.
	//
	// The zero-filedAt arm is separate and deliberate. `x.After(time.Time{})` is
	// true for any real timestamp, so folding this into the comparison below would
	// make an item with NO known filing date close on ANY component edit, however
	// old — the exact opposite of the gate's purpose, and it would have shipped
	// silently behind a doc comment that claimed otherwise. (It nearly did: the
	// comment above said "a zero filedAt makes that gate refuse" before the code
	// did, and the test caught the difference.)
	if filedAt.IsZero() {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason: fmt.Sprintf(
				"page %s no longer trips the check, but this finding carries no filing date, so there is nothing to compare the page's last edit against and no way to tell a fixed page from a moved register",
				scan.PageName),
			Evidence: map[string]interface{}{
				"page_id":                 pageID,
				"page_name":               scan.PageName,
				"components_examined":     scan.ComponentsExamined,
				"newest_component_update": scan.NewestComponentUpdate,
			},
		}
	}
	if !scan.NewestComponentUpdate.After(filedAt) {
		return revalidationVerdict{
			Verdict: revalidationUnknown,
			Reason: fmt.Sprintf(
				"page %s no longer trips the check, but no component on it has changed since this finding was filed — so the site's evidence register moved, not the page; a register entry proves a claim was registered, never that it is true",
				scan.PageName),
			Evidence: map[string]interface{}{
				"page_id":                 pageID,
				"page_name":               scan.PageName,
				"components_examined":     scan.ComponentsExamined,
				"item_filed_at":           filedAt,
				"newest_component_update": scan.NewestComponentUpdate,
			},
		}
	}

	return revalidationVerdict{
		Verdict: revalidationResolved,
		Reason:  fmt.Sprintf("re-scanned all %d component(s) on page %s against this site's current evidence base and found no unsupported claim; all %d text(s) this finding cited have gone from the slots they were cited from, and the copy has been edited since the finding was filed", scan.ComponentsExamined, scan.PageName, len(flagged)),
		Evidence: map[string]interface{}{
			"page_id":                 pageID,
			"page_name":               scan.PageName,
			"components_examined":     scan.ComponentsExamined,
			"item_filed_at":           filedAt,
			"newest_component_update": scan.NewestComponentUpdate,
			// Both gates are recorded, so a closure states which evidence it rests
			// on rather than leaving a reader to infer it from the code version.
			"flagged_texts":          len(flagged),
			"flagged_texts_verified": "all absent from their own slot",
		},
	}
}
