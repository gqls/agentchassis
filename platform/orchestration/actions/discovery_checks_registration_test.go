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
//	design-discovery-agent        24  (23 until stylesheet_gutted landed 2026-08-26; 22 until asset_reference_404, 2026-08-06)
//	quality-discovery-agent        6  (re-measured 2026-08-06; literal_markdown was live and unlisted)
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
	"content_duplication",      // seed 296, applied 2026-08-03
	"page_canonical_collision", // seed 306, applied 2026-08-04
	// cta_nonpage_destination — bugs_open/299 (slug
	// home_page_cta_names_the_brief_starter_tool_and_dials_the_phone_instead),
	// migration 475, applied 2026-08-19 on chassis v1.0.1316. Added HERE in the
	// same commit as the agent_definitions UPDATE, for asset_reference_404's
	// reason four lines further down: this list is what the live agents are
	// configured with, so a config change without it leaves the fixture
	// asserting a roster that no longer exists. Pod-verified on BOTH pods with a
	// negative control BEFORE the config landed, because an unregistered name
	// fails the whole step (discovery_checks.go:198-216) and takes the run's
	// already-collected findings down with it (the return precedes tx.Commit()).
	"cta_nonpage_destination",
	// ── Added 2026-08-26 by the bugfix_359 lane, by UNION (the rule this file
	// states: a type can vanish from a live query while its producer is still
	// deployed, so never replace this list). Found while preparing to add a
	// twentieth name and checking that the guard which protects that change is
	// sound: it was not. [MEASURED 2026-08-26] the live agents configure **82**
	// distinct check names across FIVE agents and this fixture asserted **63** —
	// a 23% blind spot in the one test that stands between a config typo and a
	// fleet-wide discovery outage. All 82 resolve today (verified by dumping
	// discovery_checks.Names() and diffing against the live `checks` arrays), so
	// this was an UNDER-ASSERTION and not a live risk — but the file's own header
	// says it exists so a rename fails HERE rather than in the fleet, and it
	// records finding `literal_markdown` live-and-unlisted with the note that
	// leaving a known gap "would be the same defect one level up".
	//
	// The five live SEO/structural checks of check_site_structural_validity.go
	// (bugs_open/251's generalisation), plus page_list_stale, all on
	// completeness-discovery-agent:
	"dead_internal_link_live", "canonical_mismatch", "structured_data_invalid",
	"head_essentials_missing", "sitemap_entry_dead_live",
	"page_list_stale",
	// The three finance-directory profiles registered dynamically by
	// check_directory.go's init() — the header already warns that a source grep
	// for `Name()` cannot build this list, and these are exactly that class.
	"missing_mortgage_lender_directory_section", "missing_mortgage_lender_directory_page",
	"missing_savings_provider_directory_section", "missing_savings_provider_directory_page",
	"missing_health_insurer_directory_section", "missing_health_insurer_directory_page",

	// design-discovery-agent
	"stale_site_components", "missing_style_collection", "shared_css_theme",
	"forced_text_colors", "missing_tools", "duplicate_palette", "tool_health",
	"hardcoded_section_colors", "deactivated_site_components", "missing_css",
	"undeployed_assets", "unfulfilled_image_prompt", "placeholder_image_in_use",
	"image_url_404", "unfulfilled_imagery_plan", "image_source_unsatisfiable",
	"tool_acceptance", "tool_acceptance_due", "sprite_css_missing",
	"content_image_missing", "palette_contrast",
	// asset_reference_404 — bugs_open/084, enabled 2026-08-06 on chassis
	// v1.0.1257 (pod-verified with a negative control before the config landed,
	// because an unregistered name fails the whole step). Added HERE in the same
	// commit as the agent_definitions UPDATE: this list is what the live agents
	// are configured with, so a config change without it leaves the fixture
	// asserting a roster that no longer exists.
	"asset_reference_404",
	// stylesheet_gutted — bugs 198+211, migration 541 (HOLD discharged and
	// applied 2026-08-26 on chassis 2fb40a960). Fills asset_reference_404's
	// stated blind spot: that check fetches the stylesheet URL and scores HTTP
	// status only, so a 136-byte 200 reads as healthy. Capability verified on
	// 217/217 live pods via service_binary_capabilities with a negative control
	// BEFORE the config landed; real-predicate calibration over all 31 deployed
	// sites the same morning filed 0.
	"stylesheet_gutted",
	// unrendered_page_imagery — bugs_open/114's closing detector (IMG-077,
	// council APPROVED corr 3b568104 round 3), migration 708 (HOLD discharged
	// and applied 2026-09-03 on chassis v1.0.1356). Capability verified on
	// 79/79 reporting pods via service_binary_capabilities with a negative
	// control BEFORE the config landed. Added HERE in the same commit as the
	// discharge, per this list's standing rule.
	"unrendered_page_imagery",

	// availability-discovery-agent (bugs_open/236 522-half, migration 372,
	// applied 2026-08-10). A FOURTH agent now calls run_discovery_checks, so the
	// header's "three live agents" is one out of date from this line down. It
	// carries exactly one check and no LLM steps, which is why it is the only
	// discovery agent that still functions during the Anthropic cap (bugs_open/243).
	"site_unreachable",
	// page_content_divergence — bugs_open/315 / RFC_038, live on this same agent
	// and missing from this fixture until 2026-08-26. It is the SECOND half of
	// the agent the 359 lane is about to add a third check to, so the guard was
	// blind to the very agent being edited.
	"page_content_divergence",

	// quality-discovery-agent
	"broken_nav_links", "placeholder_contact", "generic_theme", "unverified_claims",
	"voice_tells",
	// literal_markdown — NOT mine (bugs_open/184's lane). Found live on
	// quality-discovery-agent while enabling asset_reference_404 and missing from
	// this list, so the fixture was under-asserting by one. It resolves
	// (check_literal_markdown.go:95), so there was no production risk — but a
	// roster that silently drifts is exactly what this file exists to prevent,
	// and leaving a known gap in it would be the same defect one level up.
	"literal_markdown",
	// Also live on quality-discovery-agent and unlisted until 2026-08-26.
	"decision_guards", "premise_incomplete", "revenue_shape",

	// acceptance-discovery-agent — a FIFTH agent calls run_discovery_checks
	// (RFC_056's acceptance seats), and NONE of its three names was in this
	// fixture. The header's "three live agents" was already one out of date at
	// the availability line above; it is now two.
	"build_prerequisites", "heading_promise", "structure_floor",
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
// whole approach could get wrong. These names are invisible to a grep for
// literal Name() returns; they exist only because check_directory.go's init()
// registers two checks per profile.
func TestDynamicallyRegisteredDirectoryChecksResolve(t *testing.T) {
	dynamic := []string{
		"missing_model_directory_section", "missing_model_directory_page",
		"missing_adoption_tracker_section", "missing_adoption_tracker_page",
		"missing_protocol_tracker_section", "missing_protocol_tracker_page",
		// Phase B finance kinds (2026-08-13) — same dynamic registration, three
		// more profiles. NOT added to liveConfiguredChecks above: that list
		// mirrors what the live agents are configured with, and enablement is
		// Phase B3f, after the image roll.
		"missing_mortgage_lender_directory_section", "missing_mortgage_lender_directory_page",
		"missing_savings_provider_directory_section", "missing_savings_provider_directory_page",
		"missing_health_insurer_directory_section", "missing_health_insurer_directory_page",
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
