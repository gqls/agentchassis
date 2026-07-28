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
