// FILE: platform/orchestration/actions/discovery_checks/verifiers.go
//
// Completion-time verification for site_work_items.
//
// Dispatch loops (build-dispatch-loop, site-work-orchestrator's
// fix_items_loop) call complete_work_item on every successful handler
// saga — but a saga can "succeed" without touching the defect
// (page-build-handler's complete_error path is a success-labelled
// complete_workflow, so a page with no spec sections no-ops straight
// through it). First live catch: robot-hands' gripper-detail product
// sections, marked complete on 2026-07-10 while still serving empty
// markup on 2026-07-14.
//
// An ItemVerifier answers "is the defect this item describes actually
// gone?" at completion time, re-using the SAME predicate the discovery
// check used to create the item. CompleteWorkItemAction consults the
// registry before stamping 'complete'; a failed verification routes the
// item into the fail/attempt machinery instead.
//
// Verifiers are registered per item_type, NOT per check name — the check
// "empty_sections" creates items of item_type "empty_section", and it is
// the item_type that completion sees.

package discovery_checks

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

// VerifyResult is a verifier's verdict on one work item.
type VerifyResult struct {
	Resolved bool   // true → defect gone, safe to complete
	Detail   string // human-readable evidence either way
}

// ItemVerifier re-checks the defect described by a work item's spec.
// Returning an error means the verification could not run at all —
// the caller decides policy (CompleteWorkItemAction fails open and
// records the error in the result).
type ItemVerifier func(ctx context.Context, db *sql.DB, spec map[string]interface{}, logger *zap.Logger) (VerifyResult, error)

var verifiers = map[string]ItemVerifier{}

// RegisterVerifier adds a verifier for an item_type. Called from init()
// in check files, alongside Register(check).
func RegisterVerifier(itemType string, v ItemVerifier) {
	verifiers[itemType] = v
}

// GetVerifier returns the verifier for an item_type, or nil if none.
func GetVerifier(itemType string) ItemVerifier {
	return verifiers[itemType]
}
