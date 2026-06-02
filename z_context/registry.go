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
