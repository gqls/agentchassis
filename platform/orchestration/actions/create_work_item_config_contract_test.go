// FILE: platform/orchestration/actions/create_work_item_config_contract_test.go
//
// bugs_open/136 item D — create_work_item is now opted into unknown-config-key
// detection (CheckConfig: true), so its declared contract has become
// load-bearing: a key the action reads but the spec omits will be reported as
// UNKNOWN on live definitions that legitimately carry it, and a key the spec
// declares but the action does not read is the bugs_closed/101 failure — a
// declaration used to SILENCE the detector rather than describe the code.
//
// These tests pin both directions. They are deliberately not a source scan:
// asserting over the file's text would make the comments load-bearing, and the
// first occurrence of a string would decide the result.
package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// TestCreateWorkItemDeclaresEveryLiteralKeyItReads pins the ConfigKeys list
// against the keys the action body reads straight from params.StepConfig.Config.
//
// MUTATION PROOF: remove any entry from CreateWorkItemInputSpec.ConfigKeys and
// this fails naming it. That matters because the failure it prevents is silent
// in production — the action keeps reading the key, and only the audit and the
// runtime validator start calling a working step's config unknown.
func TestCreateWorkItemDeclaresEveryLiteralKeyItReads(t *testing.T) {
	// Every literal-setting key create_work_item_action.go reads, with the line
	// it is read at. `priority` and `item_pipeline` are here because they are
	// read through GetIntField and ResolveConfigSetting — a grep for `config["`
	// misses both, which is the recorded WRONG_CALLS.md 2026-08-08 mistake.
	readByAction := map[string]string{
		"item_type":             "config[\"item_type\"] :124",
		"handler_agent":         "config[\"handler_agent\"] :128",
		"item_pipeline":         "ResolveConfigSetting :135",
		"severity":              "config[\"severity\"] :137",
		"source":                "config[\"source\"] :141",
		"status":                "config[\"status\"] :152",
		"priority":              "GetIntField :156",
		"item_key_prefix":       "config[\"item_key_prefix\"] :159",
		"item_key_suffix_field": "config[\"item_key_suffix_field\"] :184",
		"recurrence_expected":   "config[\"recurrence_expected\"] :198",
		"spec_paths":            "config[\"spec_paths\"] :219",
		"spec_literal":          "config[\"spec_literal\"] :236",
	}

	declared := make(map[string]bool, len(CreateWorkItemInputSpec.ConfigKeys))
	for _, k := range CreateWorkItemInputSpec.ConfigKeys {
		declared[k] = true
	}

	for key, where := range readByAction {
		if !declared[key] {
			t.Errorf("ConfigKeys omits %q, which the action reads at %s — "+
				"live steps setting it will be reported as an unknown key", key, where)
		}
	}
	for _, key := range CreateWorkItemInputSpec.ConfigKeys {
		if _, ok := readByAction[key]; !ok {
			t.Errorf("ConfigKeys declares %q, which this test does not record the "+
				"action as reading — either the action changed, or a key is being "+
				"declared to silence the detector (bugs_closed/101)", key)
		}
	}
}

// TestCreateWorkItemDoesNotRedeclareFrameworkKeys guards the other way a
// contract goes wrong: claiming a key the ORCHESTRATOR owns. input_fields is
// read by ExtractActionInputs for every action, error_step by the step router.
// Repeating one here would not break behaviour, which is precisely why it would
// survive — it would just quietly misstate whose contract the key belongs to.
func TestCreateWorkItemDoesNotRedeclareFrameworkKeys(t *testing.T) {
	for _, k := range CreateWorkItemInputSpec.ConfigKeys {
		if datahelpers.IsFrameworkStepConfigKey(k) {
			t.Errorf("ConfigKeys declares framework key %q — it is recognised "+
				"centrally by IsFrameworkStepConfigKey and belongs to the orchestrator", k)
		}
	}
	// Positive control: without this, the loop above passes vacuously if
	// IsFrameworkStepConfigKey ever starts returning false for everything.
	if !datahelpers.IsFrameworkStepConfigKey("input_fields") {
		t.Fatal("IsFrameworkStepConfigKey(\"input_fields\") is false — the check " +
			"above cannot detect anything")
	}
}

// TestCreateWorkItemUnknownKeyDetectionIsLive proves the opt-in actually took
// effect, rather than asserting the CheckConfig field's value. A field read is
// not a behaviour: checksConfig() is unexported, and it is what decides whether
// UnknownConfigKeys inspects this action at all.
//
// MUTATION PROOF, stated accurately: this fails only if BOTH opt-in signals are
// removed, because checksConfig() is `CheckConfig || len(ConfigKeys) > 0` — a
// guard in SERIES. Flipping CheckConfig to false alone leaves the action opted
// in via its non-empty ConfigKeys and this test still passes; that mutation was
// run, and it did pass. TestCreateWorkItemCheckConfigCarriesTheOptInAlone below
// is what isolates the second signal.
func TestCreateWorkItemUnknownKeyDetectionIsLive(t *testing.T) {
	// A key no strategy resolves and the action never reads — the shape of every
	// key bugs_open/136 found (summary_template, spec_fields, domain, spec).
	unknown, checked := datahelpers.UnknownConfigKeys("create_work_item",
		map[string]interface{}{
			"item_type":     "x",
			"handler_agent": "y",
			"zzz_dead_key":  "never read by anything",
		})
	if !checked {
		t.Fatal("create_work_item is not opted into unknown-config-key detection")
	}
	if len(unknown) != 1 || unknown[0] != "zzz_dead_key" {
		t.Errorf("unknown = %v, want exactly [zzz_dead_key]", unknown)
	}

	// The negative control the section above cannot supply: a config of nothing
	// but recognised keys must come back clean, or the test above would pass for
	// an action that calls EVERYTHING unknown.
	clean, checked := datahelpers.UnknownConfigKeys("create_work_item",
		map[string]interface{}{
			"site_id":       "site_record.site_id", // Required
			"spec_data":     "audit",               // Optional
			"item_type":     "x",                   // ConfigKeys
			"priority":      20,                    // ConfigKeys, read via GetIntField
			"item_domain":   "content",             // DeprecatedConfigKeys — wired, not unknown
			"input_fields":  []interface{}{"x"},    // framework
			"error_step":    "complete_error",      // framework
			"handler_agent": "human-review",        // ConfigKeys
		})
	if !checked {
		t.Fatal("second call reported not-checked")
	}
	if len(clean) != 0 {
		t.Errorf("recognised-only config reported unknown keys: %v", clean)
	}
}

// TestCreateWorkItemCheckConfigCarriesTheOptInAlone isolates the signal the
// test above cannot see. CheckConfig is REDUNDANT on this spec today — a
// non-empty ConfigKeys already opts the action in — so nothing about live
// behaviour would change if it were dropped, and nothing would fail. That is
// exactly why it needs its own test: a field whose removal breaks no test and
// no behaviour is a field that gets removed as tidy-up, and the opt-in would
// then quietly become a side effect of the settings list rather than a
// decision. Should this action ever stop reading settings keys, ConfigKeys
// empties and CheckConfig is the only thing still holding detection on.
//
// It works by registering a COPY of the spec with ConfigKeys emptied, under a
// name nothing else uses, and asking the same public entry point.
//
// MUTATION PROOF: set CreateWorkItemInputSpec.CheckConfig to false and this
// fails — the copy is then opted in by nothing.
func TestCreateWorkItemCheckConfigCarriesTheOptInAlone(t *testing.T) {
	const probe = "test_only_bug136_create_work_item_without_configkeys"

	stripped := CreateWorkItemInputSpec
	stripped.ConfigKeys = nil
	datahelpers.RegisterActionInputSpec(probe, stripped)

	_, checked := datahelpers.UnknownConfigKeys(probe,
		map[string]interface{}{"zzz_dead_key": "never read by anything"})
	if !checked {
		t.Error("with ConfigKeys emptied the spec is no longer checked — CheckConfig " +
			"is not carrying the opt-in, so detection depends entirely on the " +
			"settings list being non-empty")
	}

	// Control: an otherwise identical spec with BOTH signals off must come back
	// unchecked. Without this, the assertion above would pass against a
	// checksConfig() that had started returning true unconditionally.
	off := stripped
	off.CheckConfig = false
	datahelpers.RegisterActionInputSpec(probe+"_control", off)
	if _, checkedControl := datahelpers.UnknownConfigKeys(probe+"_control",
		map[string]interface{}{"zzz_dead_key": "x"}); checkedControl {
		t.Error("a spec with neither ConfigKeys nor CheckConfig reported as checked — " +
			"this test cannot distinguish opted-in from not")
	}
}

// TestCreateWorkItemSpecKeyIsRemoved (bugs_open/234): `spec` is RETIRED, not
// unknown and NOT recognised. Three live steps carried it for months while the
// action read spec_data/spec_paths/spec_literal, so every item they filed had
// an empty spec and improvement-loop's refresh_site_components flag never
// reached the rerender gate. Migration 364 translated the carriers; this pin is
// what stops the declaration quietly disappearing and the key coming back.
//
// MUTATION PROOF: delete the "spec" entry from RemovedConfigKeys and the first
// subtest fails; move `spec` into ConfigKeys instead (the bugs_closed/101
// mistake — declaring a dead key to silence the detector) and the second fails.
func TestCreateWorkItemSpecKeyIsRemoved(t *testing.T) {
	t.Run("spec is declared removed, with the replacement named", func(t *testing.T) {
		inUse, declared := datahelpers.RemovedConfigKeysInUse("create_work_item",
			map[string]interface{}{"spec": map[string]interface{}{"refresh_site_components": true}})
		if !declared {
			t.Fatal("create_work_item declares no RemovedConfigKeys — the retired `spec` " +
				"key can be written again and will silently file empty specs (bugs_open/234)")
		}
		msg, ok := inUse["spec"]
		if !ok {
			t.Fatal("`spec` is not in create_work_item's RemovedConfigKeys")
		}
		for _, replacement := range []string{"spec_data", "spec_paths", "spec_literal"} {
			if !strings.Contains(msg, replacement) {
				t.Errorf("the removal message must name %s so the error tells the author "+
					"the fix, got: %q", replacement, msg)
			}
		}
	})

	t.Run("spec is not recognised", func(t *testing.T) {
		for _, k := range CreateWorkItemInputSpec.ConfigKeys {
			if k == "spec" {
				t.Error("`spec` found in ConfigKeys — declaring a dead key recognised " +
					"silences the detector while the behaviour stays broken (bugs_closed/101)")
			}
		}
		for _, k := range append(append([]string{}, CreateWorkItemInputSpec.Required...),
			CreateWorkItemInputSpec.Optional...) {
			if k == "spec" {
				t.Error("`spec` found in Required/Optional — same failure, different list")
			}
		}
	})
}

// TestCreateWorkItemIsStrict (bugs_open/234, owner decision 2026-08-10): the
// recognised set was verified complete against every live step at all depths
// after migration 364, which is the precondition the ConfigKeys doc comment
// sets for strict. From here an unrecognised key on this action is a definition
// error at seed time, not an empty spec found by archaeology months later.
//
// Asserted through IsStrictConfigAction rather than the struct field, because
// the field alone is not the behaviour: strict only fires for an action that
// also checksConfig(), and that composition is what the validator consults.
func TestCreateWorkItemIsStrict(t *testing.T) {
	if !datahelpers.IsStrictConfigAction("create_work_item") {
		t.Error("create_work_item is not strict — an unrecognised config key is back " +
			"to a once-per-pod warning nobody reads (bugs_open/234)")
	}
}
