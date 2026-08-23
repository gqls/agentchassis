// FILE: platform/orchestration/actions/revalidate_unbuilt_link.go
//
// The review queue's drain for `unbuilt_internal_link` (registered in
// reviewRevalidators) — bugs_open/328.
//
// WHY THIS TYPE. bugs_open/220 gave the type correct routing and a completion
// verifier, and both work. What it does not have is any way to LEAVE
// needs_human_review. The verifier fires only inside complete_work_item, and
// CompleteWorkItemAction's status guard refuses to touch a needs_human_review
// row ("do NOT overwrite a status a handler deliberately set"), so an item the
// handler parked can never reach the verifier that would clear it. Meanwhile the
// handler's one remedy — build the target page — fails on this population by
// construction: it is parked precisely because the target could not be built.
//
// COUNT THE CLOSERS, NOT JUST THE PRODUCERS (the bar the earlier adopters set,
// after one lane shipped a duplicate closer — see LANDMINES on
// reviewRevalidators). Measured 2026-08-23 on live clients_db:
//
//   - Producer side: ONE producer, discovery_checks/check_phantom_internal_links.go
//     (grep of the whole tree for the literal, not filtered to ItemType:).
//   - Closer side: 58 sit at needs_human_review, every one with triaged_at set,
//     handler_agent = 'page-build-handler' and attempt_count >= 1 — dispatched,
//     attempted, FAILED, parked.
//
// > **CORRECTED 2026-08-23 (council round 3, prior_art_librarian).** This block
// > first read "72 rows in the type's whole history, ZERO ever reached complete
// > or verified". That is FALSE, and it was false because the census read a
// > ROLLING WINDOW and called it history: `work-item-archiver` moves terminal
// > rows to `site_work_items_archive`, which the query never touched. The true
// > history is **99 rows (72 live + 27 archived), of which 26 COMPLETED** —
// > 2026-08-02 to 08-14, **18 of them carrying a `_verification` stamp**, i.e.
// > bugs_open/220's verifier working exactly as designed.
// >
// > **The drain still stands, but its justification is NARROWER and this is the
// > accurate one.** The type is NOT born parked — the producing check files it at
// > `detected`. It dispatches, and on handler SUCCESS it closes normally (26
// > times). Only on handler FAILURE does it land at `needs_human_review`, where
// > CompleteWorkItemAction's status guard refuses to touch it and the registered
// > completion verifier can therefore never run. THAT population — 58 rows, every
// > one a build the handler could not perform — is what this revalidator drains.
// > So the "nothing has ever closed one" bar the earlier adopters set is NOT met
// > here, and claiming it would have been a false claim about a working mechanism.
//
// WHY IT DELEGATES INSTEAD OF RESTATING THE PREDICATE. It calls
// checks.VerifyUnbuiltInternalLinkResolved — the SAME function complete_work_item
// runs — so the queue drain and the completion gate cannot answer the same
// question differently. That function already encodes the decisions that matter:
// both disjuncts (the href is no longer rendered, OR the target has shipped),
// the container/target page_id split that was 220's whole mechanism, and
// position() rather than a LIKE built from a raw href.
//
// ⚠ THE SUPPRESSION FIX DOES NOT CLOSE THESE ITEMS, and that is deliberate.
// refused_link_targets.go removes the anchor from the OUTBOUND string only;
// stored rendered_html — which this verifier reads — keeps the authored anchor.
// So an item whose target is still unbuilt stays `still_holds` even once the
// wire is clean. That is the honest answer: the page no longer serves a 404, and
// the site still contains an authored link to a page nobody has built. Closing
// it on the strength of a suppression would erase the only record of the second
// fact.
//
// WHAT IT WILL ACTUALLY CLOSE, and what it will honestly refuse. Of the 63 open
// items on 2026-08-23, ~42 name targets that serve HTTP 200 today — the lendzy
// tools, gaswholesalers' fuel-pricing-framework, mortgagecalculator's
// contact/index. Those pages were built and served and their deployed_at was
// never stamped (the bugs_open/315 family), so NeverDeployedPagePredicate still
// reads them as unbuilt and this revalidator will report `still_holds` on every
// one. It is reporting a real stamping defect, not failing: the alternative —
// teaching this drain a second, looser definition of "has shipped" — is exactly
// the drift the shared predicate exists to prevent.
//
// SAFE DIRECTION, inherited from the other revalidators: `resolved` demands
// positive evidence; a verifier error returns `unknown`, which is non-terminal
// and leaves the item exactly where it was. A wrong close releases the dedup key
// (terminal statuses are outside idx_swi_dedup) and the producing check
// re-raises on its next discovery pass over the site.
package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func revalidateUnbuiltInternalLink(ctx context.Context, db *sql.DB, item parkedReviewItem, logger *zap.Logger) revalidationVerdict {
	// itemID is carried for the verifier's logging only.
	itemID, _ := uuid.Parse(item.ID)

	// THE TARGET PAGE COMES FROM THE SPEC, NOT FROM parkedReviewItem.
	// The verifier's second disjunct needs the page the href POINTS AT, which
	// the work item carries in its page_id COLUMN — a column this sweep's loader
	// does not select. The producing check writes the same id into
	// spec.target_page_id (check_phantom_internal_links.go, the unbuilt arm), so
	// reading it here needs no change to the shared loader and no second
	// definition of which page is actionable. Confusing this with spec.page_id —
	// the page CONTAINING the link — is bugs_open/220's entire mechanism, so the
	// two are named apart everywhere they appear.
	targetIDStr := specString(item.Spec, "target_page_id")
	targetID, perr := uuid.Parse(targetIDStr)
	if perr != nil {
		return revalidationVerdict{
			Verdict:  revalidationUnknown,
			Arm:      "no_target_page_id",
			Reason:   fmt.Sprintf("the item carries no usable spec.target_page_id (%q), so the target cannot be re-judged and it stays parked", targetIDStr),
			Evidence: map[string]interface{}{"href": specString(item.Spec, "href")},
		}
	}

	result, err := checks.VerifyUnbuiltInternalLinkResolved(ctx, db, checks.VerifyTarget{
		ItemID:   itemID,
		SiteID:   item.SiteID,
		PageID:   &targetID,
		ItemType: item.ItemType,
		Spec:     item.Spec,
	}, logger)
	return unbuiltInternalLinkVerdict(specString(item.Spec, "href"), targetIDStr, result, err)
}

// unbuiltInternalLinkVerdict maps the completion verifier's answer onto the
// sweep's verdict vocabulary. Split from the database glue above so every arm is
// testable without a database — the same reason truncatedComponentVerdict is.
//
// The mapping is total and direction-preserving: an error (could not judge)
// stays parked as `unknown`, exactly as RFC_017 made the same error fail closed
// at completion time; Resolved carries the verifier's own evidence sentence,
// which names WHICH disjunct fired — the link was removed, or the target
// shipped. That distinction is the one a reader of the closed item needs, and it
// is why the reason is the verifier's Detail rather than a sentence of our own.
func unbuiltInternalLinkVerdict(href, targetPageID string, result checks.VerifyResult, err error) revalidationVerdict {
	evidence := map[string]interface{}{"href": href, "target_page_id": targetPageID}
	if err != nil {
		return revalidationVerdict{
			Verdict:  revalidationUnknown,
			Arm:      "verifier_error",
			Reason:   fmt.Sprintf("the completion verifier could not re-judge this link, so the item stays parked: %v", err),
			Evidence: evidence,
		}
	}
	if result.Resolved {
		return revalidationVerdict{
			Verdict:  revalidationResolved,
			Arm:      "verifier_resolved",
			Reason:   result.Detail,
			Evidence: evidence,
		}
	}
	return revalidationVerdict{
		Verdict:  revalidationStillHolds,
		Arm:      "verifier_target_still_unbuilt",
		Reason:   result.Detail,
		Evidence: evidence,
	}
}
