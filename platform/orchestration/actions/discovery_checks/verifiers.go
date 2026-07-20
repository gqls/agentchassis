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

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VerifyResult is a verifier's verdict on one work item.
type VerifyResult struct {
	Resolved bool   // true → defect gone, safe to complete
	Detail   string // human-readable evidence either way
}

// VerifyTarget identifies the work item being verified.
//
// It exists because the spec alone is not enough to locate the defect for most
// item types, and that — not author forgetfulness — is why this registry sat at
// ONE verifier for so long (bugs_open/021 §INSTANCE 2 diagnosed the symptom but
// not this cause). Measured 2026-07-20 over all 5,514 live work items: 2,370
// specs carry page_id and 310 carry component_id, but only 9 carry site_id. So a
// site-scoped check like hardcoded_section_colors — which files ONE aggregate
// item per site and whose predicate needs the site_id — could not express a
// verifier at all under the old spec-only contract, no matter how willing its
// author. site_id is NOT NULL on site_work_items; PageID is nil for items that
// are not page-scoped.
type VerifyTarget struct {
	ItemID   uuid.UUID
	SiteID   uuid.UUID
	PageID   *uuid.UUID
	ItemType string
	Spec     map[string]interface{}
}

// ItemVerifier re-checks the defect described by a work item.
// Returning an error means the verification could not run at all —
// the caller decides policy (CompleteWorkItemAction fails open and
// records the error in the result).
type ItemVerifier func(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error)

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

// RegisteredVerifierItemTypes returns every item_type that has a verifier.
//
// Exists for the coverage guard (verifier_coverage_test.go), which asserts that
// no item_type silently lacks verification. Note what this deliberately does NOT
// do: derive the set of item types that OUGHT to be covered. That set cannot be
// computed here — the check registry keys on check NAME, and each check's item
// types are string literals inside its Run method, so they are not enumerable at
// runtime. Worse, many high-volume item types (cta_improvement, spacing_fix,
// needs_content_planning) come from paths that never register a discovery check
// at all, so a guard iterating the check registry would under-report the very
// gap it exists to expose — the failure mode a council bug_historian seat warned
// about when this was first proposed. The guard therefore holds a maintained
// list sourced from the live database; see its refresh query in
// RUNBOOK_work_item_completion_integrity.md.
func RegisteredVerifierItemTypes() []string {
	out := make([]string, 0, len(verifiers))
	for itemType := range verifiers {
		out = append(out, itemType)
	}
	return out
}
