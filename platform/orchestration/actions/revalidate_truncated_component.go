// FILE: platform/orchestration/actions/revalidate_truncated_component.go
//
// The review queue's drain for `truncated_component` (registered in
// reviewRevalidators).
//
// WHY THIS TYPE. bugs_closed/303 retired the substring tag counter whose false
// positives filed two of the three parked items of this type, and its close-out
// said "complete them normally and the verifier resolves them as balanced".
// That sentence pointed at a path that cannot fire: the registered completion
// verifier (VerifyTruncatedComponentResolved) runs only inside
// complete_work_item, and CompleteWorkItemAction's status guard refuses to
// touch a needs_human_review row (load_work_item_actions.go, "do NOT overwrite
// a status a handler deliberately set"). check_truncated_component files every
// item AT needs_human_review with no handler — so the type had a birth status
// its own completion path is forbidden to leave. This revalidator is the drain
// the close-out assumed existed.
//
// COUNT THE CLOSERS, NOT JUST THE PRODUCERS (the bar the earlier adopters set,
// after one lane shipped a duplicate closer — see LANDMINES on reviewRevalidators).
// Measured 2026-08-19 on live clients_db:
//
//   - Producer side: one producer, check_truncated_component.go. No other Go
//     file emits ItemType "truncated_component" (grep of the whole tree, not
//     filtered to ItemType:).
//   - Closer side: all 3 rows ever filed sit at needs_human_review; handler_agent
//     empty on every one; count(DISTINCT resolution_path) = 0. Nothing has ever
//     closed one, and nothing can.
//
// WHY IT DELEGATES INSTEAD OF RESTATING THE PREDICATE. It calls
// checks.VerifyTruncatedComponentResolved — the SAME function the completion
// gate runs — so the queue drain and complete_work_item cannot answer the same
// question differently. That function already encodes the decisions that
// matter: markup-context counting (bugs_closed/303), deactivated → resolved,
// and a missing component row is an ERROR, never Resolved, because absence is
// ambiguous (bugs_open/012, /032).
//
// RETRACTION IS NOT THE AUTO-REMEDY THE CHECK FORBIDS. The item's stored `fix`
// text warns against blindly regenerating a data-backed tool (bugs_open/020).
// This never restores, regenerates or removes anything: it only stops asserting
// a finding the component's current template no longer supports. Any item that
// still_holds stays parked for a human, remedy decision untouched.
//
// SAFE DIRECTION, inherited from the other revalidators: `resolved` demands
// positive evidence (the stored template balances every paired tag, or the
// component is deactivated); a verifier error returns `unknown`, which is
// non-terminal and leaves the item exactly where it was. A wrong close releases
// the dedup key (terminal statuses are outside idx_swi_dedup), and the
// producing check re-raises on its next discovery pass over any site whose
// pages use the component — with the honest caveat the sweep's file header
// records for every type: a component no longer on any page's site never
// re-raises, and for this type that is acceptable because such a component is
// no longer served.
package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func revalidateTruncatedComponent(ctx context.Context, db *sql.DB, item parkedReviewItem, logger *zap.Logger) revalidationVerdict {
	// itemID is carried for the verifier's logging only; the deciding key is
	// spec.component_id, which the verifier itself validates — an unparseable id
	// here must not pre-empt the verdict that validation would give.
	itemID, _ := uuid.Parse(item.ID)
	result, err := checks.VerifyTruncatedComponentResolved(ctx, db, checks.VerifyTarget{
		ItemID:   itemID,
		SiteID:   item.SiteID,
		ItemType: item.ItemType,
		Spec:     item.Spec,
	}, logger)
	return truncatedComponentVerdict(specString(item.Spec, "component_id"), result, err)
}

// truncatedComponentVerdict maps the completion verifier's answer onto the
// sweep's verdict vocabulary. Split from the database glue above so every arm
// is testable without a database — the same reason voiceTellsVerdict is.
//
// The mapping is total and direction-preserving: an error (could not judge)
// stays parked as `unknown`, exactly as RFC_017 made the same error fail closed
// at completion time; Resolved carries the verifier's own evidence sentence.
func truncatedComponentVerdict(componentID string, result checks.VerifyResult, err error) revalidationVerdict {
	evidence := map[string]interface{}{"component_id": componentID}
	if err != nil {
		return revalidationVerdict{
			Verdict:  revalidationUnknown,
			Arm:      "verifier_error",
			Reason:   fmt.Sprintf("the completion verifier could not re-judge this component, so the item stays parked: %v", err),
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
		Arm:      "verifier_still_truncated",
		Reason:   result.Detail,
		Evidence: evidence,
	}
}
