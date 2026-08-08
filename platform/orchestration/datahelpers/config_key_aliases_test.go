package datahelpers

import (
	"testing"

	"go.uber.org/zap"
)

// The spec every precedence test resolves against: one canonical setting with
// one declared old name.
func aliasSpec() ActionInputSpec {
	return ActionInputSpec{
		ConfigKeys:           []string{"check_pipeline"},
		DeprecatedConfigKeys: map[string]string{"check_domain": "check_pipeline"},
	}
}

// TestResolveConfigSettingPrecedence pins the whole truth table of the
// hand-rolled shim this helper replaces.
//
// MUTATION PROOF: delete the alias loop in ResolveConfigSetting and the
// "alias only" case fails — i.e. this test does NOT pass with the rule gone.
// Swap the two branches and "both set" fails.
func TestResolveConfigSettingPrecedence(t *testing.T) {
	spec := aliasSpec()
	log := zap.NewNop()

	cases := []struct {
		name    string
		config  map[string]interface{}
		want    string
		wantVia string
	}{
		{
			name:   "canonical only",
			config: map[string]interface{}{"check_pipeline": "content"},
			want:   "content",
		},
		{
			name:    "alias only — the case the whole mechanism exists for",
			config:  map[string]interface{}{"check_domain": "content"},
			want:    "content",
			wantVia: "check_domain",
		},
		{
			name:   "both set — the new name wins",
			config: map[string]interface{}{"check_pipeline": "build", "check_domain": "content"},
			want:   "build",
		},
		{
			name:   "neither set — the default",
			config: map[string]interface{}{},
			want:   "design",
		},
		{
			name:   "canonical present but empty falls through to the alias",
			config: map[string]interface{}{"check_pipeline": "", "check_domain": "content"},
			want:   "content", wantVia: "check_domain",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, via := ResolveConfigSetting(tc.config, spec, "check_pipeline", "design", log)
			if got != tc.want {
				t.Errorf("value = %q, want %q", got, tc.want)
			}
			if via != tc.wantVia {
				t.Errorf("viaDeprecatedKey = %q, want %q", via, tc.wantVia)
			}
		})
	}
}

// TestResolveConfigSettingNonStringIsAbsent pins the .(string) semantics every
// call site already had. Coercing a number to its text would be a behaviour
// change smuggled in under a refactor.
//
// MUTATION PROOF: replace the type assertions with fmt.Sprintf and both cases
// fail.
func TestResolveConfigSettingNonStringIsAbsent(t *testing.T) {
	spec := aliasSpec()

	got, via := ResolveConfigSetting(
		map[string]interface{}{"check_pipeline": 42, "check_domain": "content"},
		spec, "check_pipeline", "design", zap.NewNop())
	if got != "content" || via != "check_domain" {
		t.Errorf("numeric canonical should be absent: got %q via %q, want \"content\" via \"check_domain\"", got, via)
	}

	got, _ = ResolveConfigSetting(
		map[string]interface{}{"check_domain": 42},
		spec, "check_pipeline", "design", zap.NewNop())
	if got != "design" {
		t.Errorf("numeric alias should be absent: got %q, want the default", got)
	}
}

// TestResolveConfigSettingIgnoresOtherCanonicalKeys proves an alias is bound to
// ITS canonical key, not to any key the caller happens to ask for.
//
// MUTATION PROOF: drop the `canonical == key` guard and this fails — without it
// every declared alias would answer every question.
func TestResolveConfigSettingIgnoresOtherCanonicalKeys(t *testing.T) {
	spec := ActionInputSpec{
		DeprecatedConfigKeys: map[string]string{"check_domain": "check_pipeline"},
	}
	got, via := ResolveConfigSetting(
		map[string]interface{}{"check_domain": "content"},
		spec, "some_other_setting", "fallback", zap.NewNop())
	if got != "fallback" || via != "" {
		t.Errorf("alias leaked to an unrelated key: got %q via %q", got, via)
	}
}

// TestResolveConfigSettingNilLoggerDoesNotPanic — actions have been observed
// calling helpers with a nil logger on the initialize path.
func TestResolveConfigSettingNilLoggerDoesNotPanic(t *testing.T) {
	got, _ := ResolveConfigSetting(
		map[string]interface{}{"check_domain": "content"}, aliasSpec(), "check_pipeline", "design", nil)
	if got != "content" {
		t.Errorf("got %q, want \"content\"", got)
	}
}

// TestUnknownConfigKeysRecognisesDeprecatedConfigKeys — an honoured alias must
// not be reported as unknown, and the negative control must still be, so
// "recognise everything" fails this too.
//
// MUTATION PROOF: delete the DeprecatedConfigKeys loop in UnknownConfigKeys and
// the first assertion fails.
func TestUnknownConfigKeysRecognisesDeprecatedConfigKeys(t *testing.T) {
	const action = "test_alias_recognised_action"
	RegisterActionInputSpec(action, ActionInputSpec{
		CheckConfig:          true,
		ConfigKeys:           []string{"check_pipeline"},
		DeprecatedConfigKeys: map[string]string{"check_domain": "check_pipeline"},
	})

	unknown, checked := UnknownConfigKeys(action, map[string]interface{}{
		"check_domain": "content",
		"bogus_key":    "x",
	})
	if !checked {
		t.Fatal("checked = false; the action opted in")
	}
	for _, k := range unknown {
		if k == "check_domain" {
			t.Error("check_domain reported as unknown, but the action honours it")
		}
	}
	if len(unknown) != 1 || unknown[0] != "bogus_key" {
		t.Errorf("negative control: unknown = %v, want exactly [bogus_key]", unknown)
	}
}

// TestDeprecatedConfigKeysInUseIndependentOfOptIn is the load-bearing one for
// create_work_item, which carries an alias on nine live steps and has NOT opted
// into unknown-key detection.
//
// MUTATION PROOF: gate DeprecatedConfigKeysInUse on spec.checksConfig() and this
// fails.
func TestDeprecatedConfigKeysInUseIndependentOfOptIn(t *testing.T) {
	const action = "test_alias_no_optin_action"
	RegisterActionInputSpec(action, ActionInputSpec{
		// deliberately NOT opted in: no ConfigKeys, no CheckConfig
		DeprecatedConfigKeys: map[string]string{"item_domain": "item_pipeline"},
	})

	inUse, declared := DeprecatedConfigKeysInUse(action, map[string]interface{}{"item_domain": "build"})
	if !declared {
		t.Fatal("declared = false for an action that declares an alias")
	}
	if inUse["item_domain"] != "item_pipeline" {
		t.Errorf("inUse = %v, want item_domain -> item_pipeline", inUse)
	}

	// Absent alias: declared stays true, the map is empty. The two must not
	// collapse — that is the same distinction UnknownConfigKeys' `checked` draws.
	inUse, declared = DeprecatedConfigKeysInUse(action, map[string]interface{}{"item_pipeline": "build"})
	if !declared || len(inUse) != 0 {
		t.Errorf("absent alias: inUse = %v, declared = %v; want empty and true", inUse, declared)
	}

	// An action with no aliases at all reports declared = false.
	if _, declared := DeprecatedConfigKeysInUse("no_such_action_at_all", nil); declared {
		t.Error("declared = true for an unregistered action")
	}
}

// TestDeprecatedConfigKeyAliasSpecParity is the fleet-wide lint: it walks every
// registered spec, so a bad declaration cannot reach production by being added
// somewhere this file never mentions.
//
// MUTATION PROOF: add "check_domain" to RunDiscoveryChecksInputSpec.ConfigKeys
// alongside its alias and rule (1) fails; point an alias at a key the spec does
// not recognise and rule (2) fails.
func TestDeprecatedConfigKeyAliasSpecParity(t *testing.T) {
	for action, spec := range actionInputSpecs {
		if len(spec.DeprecatedConfigKeys) == 0 {
			continue
		}

		recognised := map[string]bool{}
		for _, k := range spec.Required {
			recognised[k] = true
		}
		for _, k := range spec.Optional {
			recognised[k] = true
		}
		for _, k := range spec.ConfigKeys {
			recognised[k] = true
		}

		for old, canonical := range spec.DeprecatedConfigKeys {
			// (1) A key cannot be both current and deprecated. Declaring the old
			// name as live is exactly how bugs_closed/101's fix hid its own
			// remaining case: recognised, therefore silent.
			if recognised[old] {
				t.Errorf("%s: %q is declared BOTH as a live key and as a deprecated alias", action, old)
			}
			if old == canonical {
				t.Errorf("%s: %q aliases itself", action, old)
			}
			// (2) For an opted-in spec the target must be a key the spec actually
			// declares, or the alias resolves into a contract nobody stated.
			if spec.checksConfig() && !recognised[canonical] {
				t.Errorf("%s: alias %q targets %q, which the spec does not declare", action, old, canonical)
			}
		}
	}
}

// TestListDeprecatedConfigKeysCopiesTheMap — the audit binary must not be able
// to mutate a registered spec through the map it is handed.
func TestListDeprecatedConfigKeysCopiesTheMap(t *testing.T) {
	const action = "test_alias_copy_action"
	RegisterActionInputSpec(action, ActionInputSpec{
		DeprecatedConfigKeys: map[string]string{"old_key": "new_key"},
	})

	got := ListDeprecatedConfigKeys()
	if got[action]["old_key"] != "new_key" {
		t.Fatalf("listing = %v, want old_key -> new_key", got[action])
	}
	got[action]["old_key"] = "clobbered"

	spec, _ := GetActionInputSpec(action)
	if spec.DeprecatedConfigKeys["old_key"] != "new_key" {
		t.Error("the registered spec was mutated through the returned map")
	}
}
