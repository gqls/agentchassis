// FILE: cmd/config-key-audit/defaultshadow_test.go
//
// Two layers, deliberately separate. The behaviour tests drive
// findDefaultShadowedKeys with SYNTHETIC specs, one per class, so each
// resolver arm the checker mirrors is pinned independently of what any real
// action declares. The calibration tests then ask the REAL registry (the
// actions package is linked in via main.go's blank import, so its init
// registrations run here too) to reproduce both live faces of bugs_open/231 —
// a detector that never fired on its motivating case is the failure shape
// recorded in MEMORY ("re-run it on the motivating case"), so those cases are
// pinned as tests rather than as a one-off census run. If deploy_image_asset's
// Defaults ever legitimately change, the calibration tests fail and must be
// updated deliberately, citing the change — same contract as the three pinned
// tests in the actions package.
package main

import (
	"encoding/json"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// shadowSpecs is the synthetic registry for the behaviour tests: one action
// exercising every class boundary.
func shadowSpecs() map[string]datahelpers.ActionInputSpec {
	return map[string]datahelpers.ActionInputSpec{
		"shadowed_action": {
			Required: []string{"site_id"},
			Optional: []string{"purpose", "max_items", "options"},
			Defaults: map[string]interface{}{
				"purpose":   "hero",
				"max_items": 10,
				"options":   map[string]interface{}{"quality": 90},
				"mode":      "fast", // NOT in Required or Optional → unextractable
			},
			Deprecated: map[string]string{
				"purpose_field": "purpose",
				"site_id_field": "site_id", // target has NO default → bridge is live, not a finding
			},
		},
		"no_defaults_action": {
			Optional: []string{"purpose"},
		},
	}
}

func decodeTestAgents(t *testing.T, input string) []liveAgent {
	t.Helper()
	agents, failed, err := decodeLiveAgents([]byte(input), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}
	if failed != 0 {
		t.Fatalf("expected 0 undecodable rows, got %d", failed)
	}
	return agents
}

// findOne asserts exactly one finding for the given config key and returns it.
func findOne(t *testing.T, findings []defaultShadowFinding, key string) defaultShadowFinding {
	t.Helper()
	var got []defaultShadowFinding
	for _, f := range findings {
		if f.Key == key {
			got = append(got, f)
		}
	}
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 finding for key %q, got %d: %+v", key, len(got), got)
	}
	return got[0]
}

func TestFindDefaultShadowedKeys_Classes(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "shadowed_action", "config": {
				"purpose":       "logo",
				"purpose_field": "stored.purpose",
				"max_items":     25,
				"options":       {"quality": 50},
				"mode":          "input_data.mode",
				"site_id":       "site_record.site_id"
			}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, shadowSpecs())

	static := findOne(t, findings, "purpose")
	if static.Class != "static_string" || static.MatchesDefault {
		t.Errorf("static 'logo' vs default 'hero': want static_string/mismatch, got %+v", static)
	}
	if !static.dead() {
		t.Errorf("static_string must be dead: %+v", static)
	}

	bridge := findOne(t, findings, "purpose_field")
	if bridge.Class != "deprecated_bridge" || bridge.Field != "purpose" {
		t.Errorf("purpose_field bridge onto defaulted purpose: want deprecated_bridge, got %+v", bridge)
	}

	literal := findOne(t, findings, "max_items")
	if literal.Class != "non_string_literal" || literal.MatchesDefault {
		t.Errorf("literal 25 vs default 10: want non_string_literal/mismatch, got %+v", literal)
	}

	composite := findOne(t, findings, "options")
	if composite.Class != "composite_literal" || composite.MatchesDefault {
		t.Errorf("object vs default object: want composite_literal/mismatch, got %+v", composite)
	}

	// "mode" is defaulted but in neither Required nor Optional: no strategy
	// iterates it, so even its DOTTED value is dead — unextractable_field must
	// dominate the shape classes.
	unextractable := findOne(t, findings, "mode")
	if unextractable.Class != "unextractable_field" || !unextractable.dead() {
		t.Errorf("defaulted field outside Required+Optional: want unextractable_field/dead, got %+v", unextractable)
	}

	// site_id is dotted but NOT defaulted — Strategy 0 or 4 reads it live, so
	// it must not appear at all; site_id_field's bridge target has no default
	// either, so the bridge is live too.
	for _, f := range findings {
		if f.Key == "site_id" || f.Key == "site_id_field" {
			t.Errorf("non-defaulted field flagged: %+v", f)
		}
	}
	if len(findings) != 5 {
		t.Errorf("expected exactly 5 findings, got %d: %+v", len(findings), findings)
	}
}

func TestFindDefaultShadowedKeys_MatchingDefaultIsReportedNotFatal(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "shadowed_action", "config": {"purpose": "hero", "max_items": 10}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, shadowSpecs())
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d: %+v", len(findings), findings)
	}
	for _, f := range findings {
		if !f.MatchesDefault {
			t.Errorf("value equals its default (JSON float vs Go int for max_items) but matches_default is false: %+v", f)
		}
		if !f.dead() {
			t.Errorf("a matching static is still dead — only the SYMPTOM is absent: %+v", f)
		}
	}
}

func TestFindDefaultShadowedKeys_DottedOnDefaultedIsConditional(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "shadowed_action", "config": {"purpose": "input_data.purpose"}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, shadowSpecs())
	f := findOne(t, findings, "purpose")
	if f.Class != "dotted_conditional" {
		t.Errorf("dotted path on defaulted field: want dotted_conditional, got %+v", f)
	}
	if f.dead() {
		t.Errorf("dotted_conditional must not count as dead — resolvability is a runtime fact: %+v", f)
	}
}

// The bugs_open/144 traversal lesson: a finding inside a loop sub-workflow must
// be walked, and marked nested.
func TestFindDefaultShadowedKeys_WalksNestedSubWorkflows(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "loop-agent", "workflow": {"start_step": "build", "steps": {
			"build": {"action": "loop", "config": {"substeps": {
				"deploy": {"action": "shadowed_action", "config": {"purpose": "logo"}}
			}}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, shadowSpecs())
	f := findOne(t, findings, "purpose")
	if !f.Nested || f.Path != "steps.build.substeps.deploy" {
		t.Errorf("nested finding missed or mis-pathed: %+v", f)
	}
}

func TestFindDefaultShadowedKeys_NoSpecOrNoDefaultsIsSilent(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "no_defaults_action", "config": {"purpose": "logo"}},
			"s2": {"action": "unknown_action", "config": {"purpose": "logo"}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, shadowSpecs())
	if len(findings) != 0 {
		t.Errorf("actions without Defaults (or without a spec) must not be flagged: %+v", findings)
	}

	out, err := json.Marshal(findings)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(out) != "[]" {
		t.Errorf("zero findings must encode as '[]', got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Calibration against the REAL registry: both live faces of bugs_open/231.
// ---------------------------------------------------------------------------

// The pre-migration-348 pageflow-builder shape: static "logo" on
// deploy_image_asset, whose real spec defaults purpose to "hero". This is the
// exact config that shipped months of hero-shaped logos.
func TestCalibration_Pre348StaticLogoIsFlagged(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "pageflow-builder", "workflow": {"start_step": "deploy_logo_image", "steps": {
			"deploy_logo_image": {"action": "deploy_image_asset", "config": {"purpose": "logo"}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, registeredSpecs())
	f := findOne(t, findings, "purpose")
	if f.Class != "static_string" || f.MatchesDefault {
		t.Errorf("the motivating instance must fire as static_string/mismatch, got %+v", f)
	}
	if f.DefaultValue != "hero" {
		t.Errorf("deploy_image_asset's live Default changed (%v) — update the 231 record before updating this test", f.DefaultValue)
	}
}

// The second face (bugs_open/231, 2026-08-10 specimen): asset-deployer's
// dotted "input_data.purpose", which resolved nothing on the undeployed_asset
// dispatch shape and silently fell to the "hero" Default until migration 380
// gave it something to resolve. Offline this is exactly dotted_conditional —
// the checker cannot decide resolvability, and saying so is the honest report.
func TestCalibration_AssetDeployerDottedPurposeIsConditional(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "asset-deployer", "workflow": {"start_step": "deploy_asset", "steps": {
			"deploy_asset": {"action": "deploy_image_asset", "config": {"purpose": "input_data.purpose"}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, registeredSpecs())
	f := findOne(t, findings, "purpose")
	if f.Class != "dotted_conditional" || f.dead() {
		t.Errorf("the 380 face must report as dotted_conditional (not dead), got %+v", f)
	}
}
