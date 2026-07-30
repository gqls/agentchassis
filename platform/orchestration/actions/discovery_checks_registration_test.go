// FILE: platform/orchestration/actions/discovery_checks_registration_test.go
//
// Contract for bugs_open/149 B4: a discovery run that could not do what it was
// asked must SAY so, and an unregistered check name must fail the step rather
// than silently shrink the run.
//
// The fixture below is the part that makes the hard failure safe to ship. It
// pins the check names configured on the three live agents that call
// run_discovery_checks, so that if anyone renames or deletes a check in Go, this
// test fails HERE rather than the fleet's discovery runs failing in production.
// Before this change that same rename produced a WARN and a smaller run; now it
// produces an error, and this is the trade the fixture pays for.
//
// A source grep for literal `Name()` returns CANNOT build this list: six of the
// names are registered dynamically per profile in check_directory.go's init(),
// and a previous session nearly filed them as dead config on exactly that
// mistake. The list is therefore enumerated against the REGISTRY, which is what
// the action itself consults.

package actions

import (
	"testing"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// liveConfiguredChecks is every distinct check name configured across the three
// live agents that call run_discovery_checks, measured against
// agent_definitions on 2026-07-30:
//
//	completeness-discovery-agent  30
//	design-discovery-agent        22
//	quality-discovery-agent        5
//
// maintenance-triage is deliberately NOT in this list. It has a `checks` array
// too, but no run_discovery_checks step — the array belongs to
// scan_sites_for_maintenance, whose vocabulary (stale_pages, missing_content,
// orphan_nav) is a different action's and resolves in a different registry. A
// jsonb query for `$.**.checks` sweeps it up; treating those three as
// unregistered discovery checks would be a fabricated defect.
var liveConfiguredChecks = []string{
	// completeness-discovery-agent
	"all_sources_erroring", "empty_blog", "empty_sections", "missing_news_page",
	"missing_news_section", "missing_news_sources", "missing_structure", "orphan_pages",
	"stale_news_section", "unlinked_page_components", "unresolved_sections",
	"page_component_status_drift", "sectionless_pages", "component_template_corrupted",
	"phantom_internal_links", "misdirected_cta", "incomplete_page_group",
	"required_fields_missing", "section_source_drift", "dead_controls",
	"orphan_element_refs", "contact_form_undeliverable", "missing_model_directory_section",
	"missing_model_directory_page", "backend_entry_orphaned", "truncated_component",
	"missing_adoption_tracker_section", "missing_adoption_tracker_page",
	"missing_protocol_tracker_section", "missing_protocol_tracker_page",

	// design-discovery-agent
	"stale_site_components", "missing_style_collection", "shared_css_theme",
	"forced_text_colors", "missing_tools", "duplicate_palette", "tool_health",
	"hardcoded_section_colors", "deactivated_site_components", "missing_css",
	"undeployed_assets", "unfulfilled_image_prompt", "placeholder_image_in_use",
	"image_url_404", "unfulfilled_imagery_plan", "image_source_unsatisfiable",
	"tool_acceptance", "tool_acceptance_due", "sprite_css_missing",
	"content_image_missing", "palette_contrast",

	// quality-discovery-agent
	"broken_nav_links", "placeholder_contact", "generic_theme", "unverified_claims",
	"voice_tells",
}

// TestEveryLiveConfiguredCheckResolves is the safety proof for making an
// unregistered name fatal. If this fails, the fleet's discovery runs are about
// to start failing too — fix the rename or the config before shipping, do not
// weaken this test.
func TestEveryLiveConfiguredCheckResolves(t *testing.T) {
	var missing []string
	for _, name := range liveConfiguredChecks {
		if checks.Get(name) == nil {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("check name(s) configured on a LIVE discovery agent resolve to nothing: %v.\n"+
			"Since bugs_open/149 B4 an unregistered name FAILS the step, so this is a "+
			"production outage waiting for the next discovery run — not a warning. "+
			"Either restore the check or remove the name from the agent's config.\n"+
			"Registered: %v", missing, checks.Names())
	}
}

// TestDynamicallyRegisteredDirectoryChecksResolve pins the near-miss recorded in
// bugs_open/149's Group B specifically, because it is the one class this file's
// whole approach could get wrong. These six are invisible to a grep for literal
// Name() returns; they exist only because check_directory.go's init() registers
// two checks per profile.
func TestDynamicallyRegisteredDirectoryChecksResolve(t *testing.T) {
	dynamic := []string{
		"missing_model_directory_section", "missing_model_directory_page",
		"missing_adoption_tracker_section", "missing_adoption_tracker_page",
		"missing_protocol_tracker_section", "missing_protocol_tracker_page",
	}
	for _, name := range dynamic {
		if checks.Get(name) == nil {
			t.Errorf("%q is registered dynamically per profile (check_directory.go init) "+
				"and must resolve via the registry — a source grep cannot see it, which is "+
				"why it is pinned here", name)
		}
	}
}

// TestRegistryNamesAreDiscoverable guards the error message rather than the
// behaviour: the failure path prints checks.Names() so an operator can see what
// they should have typed. An empty list would make the error useless at exactly
// the moment it matters.
func TestRegistryNamesAreDiscoverable(t *testing.T) {
	names := checks.Names()
	if len(names) < len(liveConfiguredChecks) {
		t.Fatalf("registry reports %d names but %d are configured live — the registry is not "+
			"fully populated at test time, which would also mean the action's error message "+
			"cannot name the alternatives", len(names), len(liveConfiguredChecks))
	}
}
