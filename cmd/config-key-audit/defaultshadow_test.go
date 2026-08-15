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
	"fmt"
	"strings"
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
		// Separate action for the two classes Strategy 6's guards create, kept out
		// of shadowed_action so the finding COUNTS asserted above stay stable.
		"guarded_action": {
			Required: []string{"label"},
			Optional: []string{"max_pages", "flag"},
			Defaults: map[string]interface{}{
				"label":     "fallback", // Required AND Defaulted: the only spec shape
				"max_pages": 25,         // that can produce required_empty_string
				"flag":      true,
			},
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

	// RE-SPECIFIED 2026-08-13 with candidate 2. A dotless scalar of the Default's
	// kind is now applied by Strategy 6, so these two are working config, not dead
	// keys — the assertions below were `static_string`/dead and
	// `non_string_literal`/dead until the resolver changed under them.
	static := findOne(t, findings, "purpose")
	if static.Class != classLiveOverride || static.MatchesDefault {
		t.Errorf("static 'logo' vs default 'hero': want live_override/mismatch, got %+v", static)
	}
	if static.dead() || static.Verdict != "live" {
		t.Errorf("a dotless string of the Default's kind is honoured by Strategy 6: %+v", static)
	}

	// RE-SPECIFIED 2026-08-15 (owner ruling decision 4). This fixture carries BOTH
	// spellings — `purpose: "logo"` and `purpose_field: "stored.purpose"` — on the
	// same defaulted field, which is exactly the collision the council's guardian
	// seat asked be made visible. It had gone unnoticed here, and the old assertion
	// ("the bridge beats a Default when its path resolves") was WRONG for this
	// fixture: `purpose` is a dotless string of the Default's kind, so Strategy 6
	// honours it and the alias never decides the value.
	//
	// The pure-bridge case the old assertion meant to cover has NOT been dropped —
	// it moved to TestAliasShadowedByCanonical/alias_only, along with the two cases
	// where the canonical key does NOT win.
	bridge := findOne(t, findings, "purpose_field")
	if bridge.Class != classAliasShadowed || bridge.Field != "purpose" {
		t.Errorf("purpose_field alongside a winning canonical purpose: want %s, got %+v",
			classAliasShadowed, bridge)
	}
	if !bridge.dead() || bridge.Verdict != "dead" {
		t.Errorf("an alias the canonical key beats cannot decide the value: %+v", bridge)
	}

	literal := findOne(t, findings, "max_items")
	if literal.Class != classLiveOverride || literal.MatchesDefault {
		t.Errorf("literal 25 vs default 10: want live_override/mismatch, got %+v", literal)
	}

	// Composites are still refused: LiteralKind returns "" for them, so Strategy 6
	// leaves the Default standing.
	composite := findOne(t, findings, "options")
	if composite.Class != classComposite || composite.MatchesDefault {
		t.Errorf("object vs default object: want composite_literal/mismatch, got %+v", composite)
	}
	if !composite.dead() {
		t.Errorf("composite_literal is still dead: %+v", composite)
	}

	// "mode" is defaulted but in neither Required nor Optional: no strategy
	// iterates it, so even its DOTTED value is dead — unextractable_field must
	// dominate the shape classes.
	unextractable := findOne(t, findings, "mode")
	if unextractable.Class != classUnextractable || !unextractable.dead() {
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
		// Since candidate 2 these are LIVE and redundant rather than dead. The
		// float-vs-int comparison is what keeps them out of type_mismatch: a jsonb
		// 10 arrives as float64 against a Go int Default, and LiteralKind calls
		// both "number" on purpose.
		if f.dead() || f.Class != classLiveOverride {
			t.Errorf("a matching dotless scalar is live and redundant, not dead: %+v", f)
		}
	}
}

// Strategy 6's kind guard: a scalar of the WRONG kind is refused and the Default
// stands, so a config typo cannot hand an action a type its spec ruled out.
func TestFindDefaultShadowedKeys_TypeMismatchIsDead(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "guarded_action", "config": {
				"max_pages": "60",
				"flag":      "true",
				"label":     "real"
			}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, shadowSpecs())

	for _, key := range []string{"max_pages", "flag"} {
		f := findOne(t, findings, key)
		if f.Class != classTypeMismatch || !f.dead() || f.Verdict != "dead" {
			t.Errorf("%q is a string against a non-string Default: want type_mismatch/dead, got %+v", key, f)
		}
		if f.MatchesDefault {
			t.Errorf("%q: a wrong-kind value cannot match its default: %+v", key, f)
		}
	}

	// Same kind, so the Required field's override is live — being Required is not
	// itself a reason to refuse a real value.
	label := findOne(t, findings, "label")
	if label.Class != classLiveOverride {
		t.Errorf("a same-kind override of a Required+Defaulted field is live: %+v", label)
	}
}

// The one shape Strategy 6 refuses on the Required side: an explicit "" would
// turn a satisfiable field into a hard validation failure, and "" is not a
// meaning a required field can carry.
func TestFindDefaultShadowedKeys_EmptyStringOnRequiredIsDead(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "guarded_action", "config": {"label": ""}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, shadowSpecs())
	f := findOne(t, findings, "label")
	if f.Class != classRequiredEmpty || !f.dead() {
		t.Errorf("explicit \"\" on a Required+Defaulted field: want required_empty_string/dead, got %+v", f)
	}
}

// The exit rule, the summary and the shell wrapper all read Verdict. A class the
// classifier can emit but this map has never heard of would be silently bucketed,
// which is the drift the Verdict field was added to prevent.
func TestEveryClassHasAVerdict(t *testing.T) {
	emitted := []string{
		classLiveOverride, classUnextractable, classTypeMismatch,
		classRequiredEmpty, classComposite, classDeprecatedBridge,
		classDottedConditional, classAliasShadowed,
	}
	for _, c := range emitted {
		if _, ok := shadowClassVerdict[c]; !ok {
			t.Errorf("class %q is emitted by the classifier but has no verdict — the exit rule "+
				"would silently treat it as dead", c)
		}
	}
	if len(shadowClassVerdict) != len(emitted) {
		t.Errorf("shadowClassVerdict has %d entries for %d emitted classes — a stale entry is a "+
			"class that no longer exists, or one the classifier stopped emitting",
			len(shadowClassVerdict), len(emitted))
	}
	for c, v := range shadowClassVerdict {
		switch v {
		case "dead", "conditional", "live":
		default:
			t.Errorf("class %q has verdict %q, which no consumer handles", c, v)
		}
	}
}

// The wrapper script groups by verdict instead of re-deriving it from the class
// name, so the field has to survive marshalling.
func TestVerdictIsEmittedInJSON(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "shadowed_action", "config": {"options": {"quality": 50}}}
		}}}
	]`)

	out, err := json.Marshal(findDefaultShadowedKeys(agents, shadowSpecs()))
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var decoded []map[string]interface{}
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(decoded))
	}
	if decoded[0]["verdict"] != "dead" {
		t.Errorf("verdict absent or wrong in emitted JSON: %v", decoded[0])
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
//
// FLIPPED 2026-08-13 with candidate 2, and this is the load-bearing calibration:
// the motivating instance is now HONOURED rather than flagged. Until today this
// asserted static_string/dead. Both readings are the detector agreeing with the
// resolver — which is the only property it has — but they are opposite verdicts
// on the same config, so the flip is recorded here rather than in a commit
// message nobody reads next to the code.
//
// It still earns its place: it is the one test that would catch the resolver
// change being reverted or the kind guard rejecting a plain string override,
// either of which would silently restore the original bug.
func TestCalibration_Pre348StaticLogoIsNowHonoured(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "pageflow-builder", "workflow": {"start_step": "deploy_logo_image", "steps": {
			"deploy_logo_image": {"action": "deploy_image_asset", "config": {"purpose": "logo"}}
		}}}
	]`)

	findings := findDefaultShadowedKeys(agents, registeredSpecs())
	f := findOne(t, findings, "purpose")
	if f.Class != classLiveOverride || f.MatchesDefault {
		t.Errorf("the motivating instance must now report live_override/mismatch — the config says "+
			"logo and the resolver delivers logo. Got %+v", f)
	}
	if f.DefaultValue != "hero" {
		t.Errorf("deploy_image_asset's live Default changed (%v) — update the 231 record before updating this test", f.DefaultValue)
	}
	// The behavioural half of this claim is pinned in the actions package by
	// TestLegacyLogoStep_StaticPurposeBeatsTheDefault, which drives the real
	// ExtractActionInputs. This one only asserts the CHECKER agrees with it; a
	// detector that says "live" while the resolver says "hero" is the failure
	// this pair exists to make impossible.
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

// ---------------------------------------------------------------------------
// The --report surface (council REVISE round 1, bug_historian seat): the doc_notes
// body is the only durable record of a rejected override, so it has to say what
// was LOOKED AT as well as what was found, and it must be disconfirmable.
// ---------------------------------------------------------------------------

func TestDefaultShadowRunSummary_CleanRunStatesWhatWouldHaveShown(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "shadowed_action", "config": {"purpose": "logo"}}
		}}}
	]`)
	findings := findDefaultShadowedKeys(agents, shadowSpecs())
	summary := defaultShadowRunSummary(len(agents), 0, 1, findings)

	// A clean run must record the population, not just the absence of findings —
	// otherwise it is indistinguishable from a job that never ran (bugs_open/140).
	for _, want := range []string{
		"1 live agents scanned",
		"1 registered specs carry Defaults",
		"DEAD (the spec Default wins",
		"LIVE (honoured by Strategy 6): 1",
		"live in code since 2026-08-14",
		"disconfirmable",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("clean summary is missing %q:\n%s", want, summary)
		}
	}
	if strings.Contains(summary, "Dead AND mismatched") {
		t.Errorf("a clean run must not print the findings header:\n%s", summary)
	}
}

func TestDefaultShadowRunSummary_NamesEveryDeadMismatchedEntry(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "guarded_action", "config": {"max_pages": "60"}}
		}}}
	]`)
	findings := findDefaultShadowedKeys(agents, shadowSpecs())
	summary := defaultShadowRunSummary(len(agents), 0, 1, findings)

	if !strings.Contains(summary, "DEAD (the spec Default wins whatever the config says): 1 mismatched") {
		t.Errorf("summary undercounts the dead entry:\n%s", summary)
	}
	for _, want := range []string{"Dead AND mismatched", "max_pages", "type_mismatch", "CAVEAT"} {
		if !strings.Contains(summary, want) {
			t.Errorf("findings summary is missing %q:\n%s", want, summary)
		}
	}
}

// The exit rule and the summary must be the SAME count. A report that reads clean
// while the process exits 1 (or the reverse) is the drift this file's Verdict
// field was added to prevent, one layer up.
func TestCountDeadMismatchedAgreesWithTheSummary(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "guarded_action", "config": {"max_pages": "60", "flag": "yes"}},
			"s2": {"action": "shadowed_action", "config": {"purpose": "logo", "options": {"q": 1}}}
		}}}
	]`)
	findings := findDefaultShadowedKeys(agents, shadowSpecs())

	n := countDeadMismatched(findings)
	if n != 3 {
		// 2 type_mismatch (max_pages, flag) + 1 composite_literal (options).
		t.Fatalf("countDeadMismatched = %d, want 3: %+v", n, findings)
	}
	summary := defaultShadowRunSummary(len(agents), 0, 2, findings)
	if !strings.Contains(summary, fmt.Sprintf("%d mismatched", n)) {
		t.Errorf("summary headline disagrees with the exit count (%d):\n%s", n, summary)
	}
	// And the live override in the same batch must NOT be counted as dead.
	if strings.Contains(summary, "purpose='logo'") {
		t.Errorf("a live override appeared in the dead list:\n%s", summary)
	}
}

// TestAliasShadowedByCanonical covers the collision guard added on the owner's ruling of
// 2026-08-15 (decision 4 of six), raised by the council's guardian seat.
//
// THE OBJECTION IT ANSWERS. Strategy 6 lets a canonical key beat a value the deprecated
// bridge supplied, so a step carrying both spellings resolves to the canonical one. That
// was justified by a census — "zero live definitions carry both today" — which is a fact
// about today, not a constraint on tomorrow. Nothing stopped a future agent_definitions
// row from carrying both and silently getting the new precedence.
//
// WHY THE THREE NEGATIVE CASES ARE THE IMPORTANT ONES. It would be easy, and wrong, to
// implement this as "both keys present ⇒ the alias is dead". Canonical does not always
// win: Strategy 6 refuses a dotted canonical value (a reference that failed to resolve)
// and refuses one whose kind differs from the Default's, and in both the bridge's value
// still stands. A guard without these cases would report a live bridge as dead and send
// someone to delete config that is doing work.
func TestAliasShadowedByCanonical(t *testing.T) {
	cases := []struct {
		name        string
		config      string
		wantClass   string
		wantVerdict string
		why         string
	}{
		{
			name:        "alias_only",
			config:      `{"purpose_field": "stored.purpose"}`,
			wantClass:   classDeprecatedBridge,
			wantVerdict: "conditional",
			why:         "no canonical key present, so the bridge decides the value if its path resolves",
		},
		{
			name:        "canonical_wins",
			config:      `{"purpose_field": "stored.purpose", "purpose": "logo"}`,
			wantClass:   classAliasShadowed,
			wantVerdict: "dead",
			why:         "dotless string of the Default's kind — Strategy 6 honours it, so the alias is inert",
		},
		{
			name:        "canonical_is_dotted_so_bridge_still_stands",
			config:      `{"purpose_field": "stored.purpose", "purpose": "elsewhere.value"}`,
			wantClass:   classDeprecatedBridge,
			wantVerdict: "conditional",
			why:         "Strategy 6 refuses a dotted canonical (an unresolved reference); the bridge's value stands",
		},
		{
			name:        "canonical_kind_mismatch_so_bridge_still_stands",
			config:      `{"purpose_field": "stored.purpose", "purpose": 7}`,
			wantClass:   classDeprecatedBridge,
			wantVerdict: "conditional",
			why:         "the kind guard refuses a number against a string Default; the bridge's value stands",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			agents := decodeTestAgents(t, `[
				{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
					"s1": {"action": "shadowed_action", "config": `+tc.config+`}
				}}}
			]`)

			alias := findOne(t, findDefaultShadowedKeys(agents, shadowSpecs()), "purpose_field")
			if alias.Class != tc.wantClass || alias.Verdict != tc.wantVerdict {
				t.Errorf("want %s/%s, got %s/%s — %s\nfinding: %+v",
					tc.wantClass, tc.wantVerdict, alias.Class, alias.Verdict, tc.why, alias)
			}
			if alias.Field != "purpose" {
				t.Errorf("the finding must be reported against the CANONICAL field, got %q", alias.Field)
			}
		})
	}
}

// TestAliasCollisionIsAbsentFleetWideToday pins the census the guard was built on, so a
// future reader can tell "the guard found nothing" from "the guard is not running".
//
// This is the demand control for a check whose live answer is currently ZERO: the cases
// above prove it FIRES on a constructed collision, so a zero here measures the fleet
// rather than measuring a broken detector. Without them, a permanent zero is
// indistinguishable from a classifier that never emits.
func TestAliasCollisionGuardCanFire(t *testing.T) {
	agents := decodeTestAgents(t, `[
		{"type": "fixture-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "shadowed_action", "config": {"purpose_field": "stored.purpose", "purpose": "logo"}}
		}}}
	]`)

	var fired bool
	for _, f := range findDefaultShadowedKeys(agents, shadowSpecs()) {
		if f.Class == classAliasShadowed {
			fired = true
		}
	}
	if !fired {
		t.Fatal("the collision class was never emitted on a fixture built to trigger it — " +
			"a zero from the live fleet would therefore prove nothing")
	}
}
