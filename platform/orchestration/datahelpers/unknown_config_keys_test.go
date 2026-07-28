// FILE: platform/orchestration/datahelpers/unknown_config_keys_test.go
//
// bugs_open/101 — a config key that no code reads is indistinguishable by
// inspection from a live one. These tests pin the detector that makes the two
// distinguishable, and in particular pin the checked/unchecked distinction, which
// is the part most easily lost in a later refactor.

package datahelpers

import (
	"reflect"
	"testing"
)

func TestUnknownConfigKeysRequiresOptIn(t *testing.T) {
	RegisterActionInputSpec("test_action_no_config_keys", ActionInputSpec{
		Required: []string{"site_id"},
	})

	unknown, checked := UnknownConfigKeys("test_action_no_config_keys", map[string]interface{}{
		"site_id":       "x",
		"totally_bogus": 1,
	})

	// The important assertion is `checked == false`, not the empty slice. An
	// action that has not declared ConfigKeys has NOT been examined, and reporting
	// "no unknown keys" for it would be the bug wearing the fix's clothes.
	if checked {
		t.Error("action with no ConfigKeys should report checked=false (not opted in)")
	}
	if len(unknown) != 0 {
		t.Errorf("unopted-in action should report no keys, got %v", unknown)
	}
}

func TestUnknownConfigKeysUnregisteredAction(t *testing.T) {
	if _, checked := UnknownConfigKeys("no_such_action_registered", map[string]interface{}{"a": 1}); checked {
		t.Error("unregistered action must report checked=false, never a verdict")
	}
}

func TestUnknownConfigKeysDetectsAndIgnoresCorrectly(t *testing.T) {
	RegisterActionInputSpec("test_action_declared", ActionInputSpec{
		Required:   []string{"site_id"},
		Optional:   []string{"page_id"},
		ConfigKeys: []string{"max_pages", "scrape_config"},
		Deprecated: map[string]string{"site_id_field": "site_id"},
	})

	tests := []struct {
		name   string
		config map[string]interface{}
		want   []string
	}{
		{
			name:   "clean config",
			config: map[string]interface{}{"site_id": "x", "max_pages": 3},
			want:   nil,
		},
		{
			name: "required, optional, declared and deprecated keys are all recognised",
			config: map[string]interface{}{
				"site_id": "x", "page_id": "y",
				"max_pages": 3, "scrape_config": map[string]interface{}{},
				"site_id_field": "input_data.site_id",
			},
			want: nil,
		},
		{
			name: "framework keys are never unknown",
			config: map[string]interface{}{
				"input_fields": []interface{}{"site_id"},
				"agent_type":   "x", "timeout_seconds": 30, "error_step": "fail",
				"loop_iteration": 0, "output_mapping": map[string]interface{}{},
			},
			want: nil,
		},
		{
			name:   "unknown keys are reported, sorted",
			config: map[string]interface{}{"site_id": "x", "zebra": 1, "follow_links": []interface{}{"a"}},
			want:   []string{"follow_links", "zebra"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unknown, checked := UnknownConfigKeys("test_action_declared", tc.config)
			if !checked {
				t.Fatal("declared action must report checked=true")
			}
			if !reflect.DeepEqual(unknown, tc.want) {
				t.Errorf("unknown = %v, want %v", unknown, tc.want)
			}
		})
	}
}

// TestUnknownConfigKeysCatchesTheOriginalBug is the regression proper: the exact
// step config from vet-practice-verifier's scrape_website step, as read from the
// live agent_definitions row on 2026-07-28. Four of these keys were read by
// nothing. Two of them are now implemented (extract_mode, fallback_url_field) and
// two are honoured only on the crawl path (max_pages, follow_links), so all four
// are declared — the test asserts the detector's verdict on the SHAPE, using a
// spec that deliberately omits them.
func TestUnknownConfigKeysCatchesTheOriginalBug(t *testing.T) {
	// scrape_web as it was BEFORE this fix: five keys read, four ignored.
	RegisterActionInputSpec("scrape_web_as_it_was", ActionInputSpec{
		ConfigKeys: []string{"url_field", "url", "upload_results", "scrape_config"},
	})

	liveConfig := map[string]interface{}{
		"max_pages":          float64(3),
		"extract_mode":       "text",
		"url_field":          "business_record.business.website_url",
		"fallback_url_field": "search_results.results.0.url",
		"follow_links":       []interface{}{"fees", "prices", "about", "team", "contact", "services"},
	}

	unknown, checked := UnknownConfigKeys("scrape_web_as_it_was", liveConfig)
	if !checked {
		t.Fatal("checked=false — the detector did not run on an opted-in action")
	}
	want := []string{"extract_mode", "fallback_url_field", "follow_links", "max_pages"}
	if !reflect.DeepEqual(unknown, want) {
		t.Errorf("detector missed the original bug.\n got: %v\nwant: %v", unknown, want)
	}
}

func TestListDeclaredConfigKeysOnlyIncludesOptedInActions(t *testing.T) {
	RegisterActionInputSpec("test_action_listed", ActionInputSpec{
		Required:   []string{"a"},
		ConfigKeys: []string{"b"},
	})
	RegisterActionInputSpec("test_action_unlisted", ActionInputSpec{
		Required: []string{"a"},
	})

	declared := ListDeclaredConfigKeys()

	if _, ok := declared["test_action_unlisted"]; ok {
		t.Error("an action without ConfigKeys must not appear as declared — the audit would read it as covered")
	}
	keys, ok := declared["test_action_listed"]
	if !ok {
		t.Fatal("opted-in action missing from declared list")
	}
	if !reflect.DeepEqual(keys, []string{"a", "b"}) {
		t.Errorf("declared keys = %v, want [a b]", keys)
	}
}

// TestConditionalConfigKeys pins the third state. Round 1 of this fix had two —
// unknown and recognised — and the real world had three, so declaring a key was a
// way of hiding it (WRONG_CALLS.md 2026-07-28).
func TestConditionalConfigKeys(t *testing.T) {
	RegisterActionInputSpec("test_action_conditional", ActionInputSpec{
		ConfigKeys: []string{"url_field", "max_pages", "follow_links"},
		ConditionalKeys: map[string]string{
			"max_pages":    "only on a crawl",
			"follow_links": "only on a crawl",
		},
	})

	t.Run("present conditional keys are returned with their condition", func(t *testing.T) {
		got, checked := ConditionalConfigKeys("test_action_conditional", map[string]interface{}{
			"url_field": "input_data.url",
			"max_pages": 3,
		})
		if !checked {
			t.Fatal("checked=false for an action that declares ConditionalKeys")
		}
		if len(got) != 1 || got["max_pages"] != "only on a crawl" {
			t.Errorf("got %v, want just max_pages with its condition", got)
		}
	})

	t.Run("absent conditional keys are not invented", func(t *testing.T) {
		got, checked := ConditionalConfigKeys("test_action_conditional", map[string]interface{}{
			"url_field": "input_data.url",
		})
		if !checked {
			t.Fatal("checked=false")
		}
		if len(got) != 0 {
			t.Errorf("got %v, want none — only keys PRESENT in the config count", got)
		}
	})

	t.Run("a conditional key is still recognised, never unknown", func(t *testing.T) {
		// The two mechanisms must not double-report: a key cannot be both.
		unknown, _ := UnknownConfigKeys("test_action_conditional", map[string]interface{}{
			"max_pages": 3,
		})
		if len(unknown) != 0 {
			t.Errorf("conditional key reported as unknown: %v", unknown)
		}
	})

	t.Run("action with no ConditionalKeys reports checked=false", func(t *testing.T) {
		RegisterActionInputSpec("test_action_no_conditional", ActionInputSpec{
			ConfigKeys: []string{"a"},
		})
		if _, checked := ConditionalConfigKeys("test_action_no_conditional", map[string]interface{}{"a": 1}); checked {
			t.Error("an action declaring no ConditionalKeys must report checked=false")
		}
	})
}

// TestUndeclaredConditionalKeysCatchesSpecError: a conditional key MUST also be in
// ConfigKeys, or it reports as both unknown and conditional — two mechanisms
// disagreeing about the same key.
func TestUndeclaredConditionalKeysCatchesSpecError(t *testing.T) {
	RegisterActionInputSpec("test_action_broken_spec", ActionInputSpec{
		ConfigKeys:      []string{"url_field"},
		ConditionalKeys: map[string]string{"max_pages": "only on a crawl"},
	})

	missing := UndeclaredConditionalKeys("test_action_broken_spec")
	if len(missing) != 1 || missing[0] != "max_pages" {
		t.Errorf("got %v, want [max_pages] — a conditional key absent from ConfigKeys is a spec error", missing)
	}
}

// TestCheckConfigOptsInWithoutConfigKeys pins the second opt-in route, added
// 2026-07-28 because the first one had stalled adoption at 1 action of 152.
//
// ConfigKeys means "settings rather than references". For the large class of
// actions whose every key is an ExtractActionInputs field there was no honest
// way to opt in — you had to duplicate keys into a list they do not belong in.
// CheckConfig separates "I want to be checked" from "here are my non-field
// settings", which were never the same statement.
func TestCheckConfigOptsInWithoutConfigKeys(t *testing.T) {
	RegisterActionInputSpec("test_action_checkconfig", ActionInputSpec{
		CheckConfig: true,
		Required:    []string{"site_id"},
		Optional:    []string{"pages_field"},
		// ConfigKeys deliberately EMPTY — that is the whole point of this test.
	})

	unknown, checked := UnknownConfigKeys("test_action_checkconfig", map[string]interface{}{
		"site_id":       "x",
		"pages_field":   "a.b",
		"action":        "noop", // framework key, never unknown
		"totally_bogus": 1,
	})

	// The NEGATIVE CONTROL is the load-bearing assertion. A change that made
	// checksConfig() always false would still satisfy "no false positives" — the
	// audit would print a clean bill over an unexamined fleet, which is exactly
	// the failure this whole mechanism exists to prevent. So assert that the
	// detector FIRES, not merely that it stays quiet.
	if !checked {
		t.Fatal("CheckConfig:true must opt the action in even with empty ConfigKeys")
	}
	if !reflect.DeepEqual(unknown, []string{"totally_bogus"}) {
		t.Errorf("expected exactly [totally_bogus], got %v", unknown)
	}
}

// TestCheckConfigRecognisesRequiredAndOptional guards the claim that made the
// bulk opt-in safe: ExtractActionInputs reads config[k] for every Required and
// Optional key, so for an action that passes its own spec to the extractor those
// lists ARE a verified statement of what it reads. If the recognised set ever
// stopped including them, 56 actions would start warning about their own
// working config on the same day.
func TestCheckConfigRecognisesRequiredAndOptional(t *testing.T) {
	RegisterActionInputSpec("test_action_checkconfig_union", ActionInputSpec{
		CheckConfig: true,
		Required:    []string{"req_key"},
		Optional:    []string{"opt_key"},
		Deprecated:  map[string]string{"old_key": "opt_key"},
	})

	unknown, checked := UnknownConfigKeys("test_action_checkconfig_union", map[string]interface{}{
		"req_key": 1, "opt_key": 2, "old_key": 3,
	})
	if !checked {
		t.Fatal("expected checked=true")
	}
	if len(unknown) != 0 {
		t.Errorf("Required/Optional/Deprecated must all be recognised, got unknown=%v", unknown)
	}
}

// TestListDeclaredConfigKeysIncludesCheckConfig pins the runtime detector and
// the OFFLINE audit to one definition of "opted in". They used to test
// len(ConfigKeys)==0 in three separate places; a CheckConfig action that the
// validator checked but the audit still counted as an uncovered gap would make
// the coverage number silently wrong in the safe-looking direction.
func TestListDeclaredConfigKeysIncludesCheckConfig(t *testing.T) {
	RegisterActionInputSpec("test_action_audit_join", ActionInputSpec{
		CheckConfig: true,
		Optional:    []string{"only_key"},
	})

	declared := ListDeclaredConfigKeys()
	keys, present := declared["test_action_audit_join"]
	if !present {
		t.Fatal("a CheckConfig action must appear in ListDeclaredConfigKeys, or the audit under-reports coverage")
	}
	if !reflect.DeepEqual(keys, []string{"only_key"}) {
		t.Errorf("expected [only_key], got %v", keys)
	}
}
