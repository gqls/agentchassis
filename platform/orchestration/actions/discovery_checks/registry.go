// FILE: platform/orchestration/actions/discovery_checks/registry.go
//
// Discovery check modularity pattern.
//
// Current state: RunDiscoveryChecksAction is a ~480-line function with
// if-chains for each check. Every new check adds ~30-50 lines to the
// main function plus a finder function + struct. The file grows linearly.
//
// Proposed: Extract each check into its own file implementing a common
// interface. The main action becomes a simple loop over enabled checks.
//
// Directory layout:
//   actions/
//     run_discovery_checks_action.go       ← slimmed down, just the loop
//     discovery_checks/
//       registry.go                        ← this file: interface + registry
//       check_empty_sections.go
//       check_undeployed_assets.go
//       check_missing_css.go
//       check_duplicate_palette.go
//       check_hardcoded_section_colors.go
//       check_forced_text_colors.go
//       check_broken_nav_links.go
//       check_placeholder_contact.go
//       check_generic_theme.go
//       check_missing_tools.go
//       ... future checks just add a file

package discovery_checks

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DiscoveryCheckContext holds the shared state that every check needs.
// Passed by the main action — checks don't reach into ActionParams directly.
type DiscoveryCheckContext struct {
	Ctx       context.Context
	DB        *sql.DB
	TX        *sql.Tx // for work item inserts (shared transaction)
	SiteID    uuid.UUID
	Pipeline  string    // check_pipeline from config, e.g. "design"
	AgentType string    // sender agent type for created_by
	BatchID   uuid.UUID // groups work items from one run
	Logger    *zap.Logger
}

// WorkItemSpec describes a work item to be inserted.
// Mirrors the existing workItem struct in run_discovery_checks_action.go —
// this avoids creating a separate struct. During migration, the existing
// workItem struct can be aliased or replaced.
type WorkItemSpec struct {
	SiteID       uuid.UUID
	PageID       *uuid.UUID // optional
	Source       string     // "discovery"
	Pipeline     string     // "design", "build", "content"
	ItemType     string     // e.g. "undeployed_asset", "add_tool"
	Severity     string     // "high", "medium", "low"
	Summary      string
	SpecJSON     string // JSON-encoded spec
	Priority     int
	HandlerAgent string
	Status       string // "detected"
	CreatedBy    string
	ItemKey      string // dedup key
	BatchID      uuid.UUID

	// RecurrenceExpected marks an item whose RE-REQUEST is normal rather than a
	// sign the previous attempt failed, so the anti-churn brake must not brand or
	// drop it. Mirrors workItem.recurrenceExpected, which the runner sets from
	// this field; false keeps exactly today's behaviour for every other check.
	//
	// Set it when each finding is a NEW EVENT rather than a repeat of the same
	// one — the page_divergence_overwritten precedent: a second destroyed
	// hand-patch is a second loss, and without this the brake silently drops the
	// third same-key finding.
	RecurrenceExpected bool
}

// CheckResult is what a check returns.
type CheckResult struct {
	// Findings are appended to allFindings in the action return value.
	// Each entry is a free-form map — checks decide their own shape.
	Findings []map[string]interface{}

	// WorkItems are inserted into site_work_items by the main loop.
	// The check builds them; the main action inserts them (so the
	// check doesn't need to know about insertWorkItem).
	WorkItems []WorkItemSpec

	// Resolved names work items this check has POSITIVELY OBSERVED to be
	// fixed. The runner closes them; the check does not touch the table.
	//
	// WHY THIS EXISTS (RFC_010, owner ruling 2026-08-02 "Decision 1: option 1").
	// Until this field, 49 of 50 checks were monotonic: each computed the
	// current truth set on every run, filed what it found, and discarded the
	// complement — the free information that would let it close items it had
	// previously raised that no longer reproduce. Items therefore outlived the
	// predicate that raised them, stayed dispatchable, and could be acted on
	// months later against code that had moved underneath them. That is not
	// hypothetical: eleven such items were three days stale and would have
	// overwritten a live social card once bugs_closed/168 shipped.
	//
	// THE SAFETY PROPERTY, and it is why this is a field and not an inference.
	// A retraction fires ONLY on a positive observation. Nothing anywhere
	// derives "resolved" from an EMPTY Findings slice, because a check that
	// errored, or that was silently blinded by a bug, returns exactly that —
	// an empty result indistinguishable from a healthy site (016b §9: a gate's
	// 0 findings has two causes with opposite fixes). Deriving retraction from
	// absence would quietly close real defects fleet-wide, which is a far worse
	// failure than the one this fixes. The runner additionally skips Resolved
	// entirely when Run returned an error.
	//
	// OPT-IN, with the wide branch behind an explicit flag — per the owner
	// ruling of 2026-08-02 that new authority on a shared seam ships as a field
	// with the unsafe default OFF, not as a rule in a doc comment. Populating
	// nothing retracts nothing, so the other 49 checks are unaffected until
	// each is edited deliberately.
	Resolved []ResolvedFinding
}

// ResolvedFinding names work the check has confirmed is done.
//
// Exactly one of ItemKey or AllOfType must be set. Both unset, or both set, is
// a programming error: the runner refuses the entry and logs loudly rather than
// guessing, because the two mean very different things and the wide one is not
// something to arrive at by leaving a field blank.
type ResolvedFinding struct {
	// ItemType is the work item type to close. Required.
	ItemType string

	// ItemKey closes exactly the item carrying this key, for this site.
	// This is the narrow, ordinary case: the check looked at one specific
	// thing and found it healthy.
	ItemKey string

	// AllOfType closes EVERY open item of ItemType for this site. This is the
	// wide branch, and it is deliberately a separate boolean rather than "leave
	// ItemKey empty" so that a reviewer reading the CALL SITE sees the breadth
	// of the claim being made. Only correct when the check's observation covers
	// the whole item type for the site — e.g. a health probe that succeeded, so
	// every open "unreachable" item for that site is answered at once.
	AllOfType bool

	// Reason is recorded on the row in result.reason. Required — an item that
	// closes itself with no stated cause is indistinguishable later from one a
	// human closed by hand.
	Reason string

	// Receipt, when set, is a work item that MUST be durably present before this
	// retraction is applied. resolveWorkItems writes it FIRST, in the SAME
	// transaction, and REFUSES the retraction if it can neither insert it nor
	// confirm an open one already holds its key.
	//
	// WHY THIS EXISTS (bugs_open/469). Resolved's safety property is that a
	// retraction fires only on a positive observation. That is necessary and it
	// is not sufficient, because for one class of check the observation that the
	// finding no longer reproduces IS the observation that the damage completed.
	// check_section_source_drift files "the authoritative section list disagrees
	// with pages.sections; the next build will overwrite the cache". When the
	// build duly does so, the two stores agree again — the finding stops
	// reproducing precisely BECAUSE a human's composition was destroyed. A
	// retraction on agreement alone would therefore close, automatically and
	// fleet-wide, exactly the cases that most need a human. It ran once by hand
	// (migration 753, three of six pages) and one of the three had lost a section
	// a human deliberately added five weeks earlier.
	//
	// So the rule this field enforces is: A CHECK MAY RETRACT A FINDING WHOSE
	// RESOLUTION DESTROYED SOMETHING ONLY BY RECORDING WHAT WAS DESTROYED, IN THE
	// SAME TRANSACTION. Not by convention — the owner's 2026-08-02 ruling is that
	// a comment is not a control on a tree this many sessions share.
	//
	// It lives on the SEAM rather than in any one check because resolveWorkItems
	// has two callers today (the discovery runner and work_item_retraction) and a
	// control in either one protects only that one.
	//
	// OPT-IN, unsafe default OFF, per the same ruling: nil is exactly today's
	// behaviour, so no existing check is affected until it is edited deliberately.
	Receipt *WorkItemSpec

	// Evidence is merged into the closed row's `result` jsonb alongside
	// resolved_by/reason. It exists so a machine close and a hand-written
	// migration close record the SAME shape: migration 753 wrote
	// result->>'direction' by hand, and a reader must not have to know which
	// closed a row in order to ask which side won.
	//
	// Nil omits the key entirely, leaving the statement byte-identical to before
	// this field — which is what keeps every existing caller's test unchanged.
	Evidence map[string]interface{}
}

// DiscoveryCheck is the interface every check implements.
type DiscoveryCheck interface {
	// Name returns the check identifier used in workflow config,
	// e.g. "missing_tools", "undeployed_assets".
	Name() string

	// Run executes the check and returns findings + work items.
	// Returning an error means the check failed (logged, not fatal).
	// Returning empty results means nothing was found (normal).
	Run(dctx DiscoveryCheckContext) (*CheckResult, error)
}

// --- Registry ---

var registry = map[string]DiscoveryCheck{}

// Register adds a check to the registry. Called from init() in each check file.
func Register(check DiscoveryCheck) {
	registry[check.Name()] = check
}

// Get returns a check by name, or nil if not registered.
func Get(name string) DiscoveryCheck {
	return registry[name]
}

// Names returns all registered check names (for logging/debugging).
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
