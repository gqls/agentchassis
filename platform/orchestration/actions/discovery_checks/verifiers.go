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
//
// Returning an error means the verification could not run at all. Since the
// OWNER RULING of 2026-08-08 (architecture_review/RFC_017) that is FAIL-CLOSED
// by default: CompleteWorkItemAction does NOT stamp 'complete', and routes the
// item into the attempt machinery instead. An error is therefore a real refusal,
// not an annotation — if your ambiguous case should still complete, you must say
// so explicitly at registration (see VerifierPolicy).
type ItemVerifier func(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error)

// VerifierPolicy carries per-item_type registration options.
//
// It exists because the previous behaviour — every verifier failing OPEN on
// error — was licensed by a doc comment rather than by anything a reviewer of
// the CALLING check could see, and the owner's 2026-08-02 ruling on shared seams
// is that authority like that "ships as an OPT-IN FIELD, not a documented
// contract", with the unsafe default OFF. The unsafe direction here is
// completing an item nobody verified, so that is what now costs a line of code.
type VerifierPolicy struct {
	// FailOpenOnError completes the item when the verifier returns an error.
	//
	// UNSAFE, and OFF by default. Measured before the flip (RFC_017, 2026-08-08):
	// verifiers had errored twice in the registry's whole life, both on the same
	// "target row is absent" branch, and on BOTH the page still declared the slot
	// — so the honest verdict was Resolved:false and fail-open completed a defect
	// that is still live. Zero infrastructural errors were ever observed, which is
	// the failure mode fail-open exists to protect against.
	//
	// Set it only when completing-without-verification is genuinely the lesser
	// harm for THIS item type, and say why at the registration site.
	FailOpenOnError bool
}

var (
	verifiers = map[string]ItemVerifier{}
	policies  = map[string]VerifierPolicy{}
)

// RegisterVerifier adds a verifier for an item_type, FAIL-CLOSED on error.
// Called from init() in check files, alongside Register(check).
func RegisterVerifier(itemType string, v ItemVerifier) {
	RegisterVerifierWithPolicy(itemType, v, VerifierPolicy{})
}

// RegisterVerifierWithPolicy registers a verifier with an explicit policy.
// Use this only to opt INTO an unsafe behaviour; the zero policy is the safe one.
func RegisterVerifierWithPolicy(itemType string, v ItemVerifier, p VerifierPolicy) {
	verifiers[itemType] = v
	policies[itemType] = p
}

// GetVerifier returns the verifier for an item_type (nil if none) and the policy
// governing it. The policy is returned alongside rather than fetched separately
// so a caller cannot consult the verifier while forgetting the terms it runs
// under — the zero VerifierPolicy is fail-closed, so an unregistered type and a
// caller that ignores this value both land on the safe branch.
func GetVerifier(itemType string) (ItemVerifier, VerifierPolicy) {
	return verifiers[itemType], policies[itemType]
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
