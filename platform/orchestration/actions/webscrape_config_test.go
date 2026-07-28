// FILE: platform/orchestration/actions/webscrape_config_test.go
//
// bugs_open/101 — scrape_web advertised four config keys that nothing read.

package actions

import (
	"reflect"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func TestBuildScrapeConfigHonoursExtractMode(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		wantFormats interface{}
	}{
		{
			name:        "text",
			config:      map[string]interface{}{"extract_mode": "text"},
			wantFormats: []interface{}{"markdown"},
		},
		{
			name:        "case and whitespace tolerant",
			config:      map[string]interface{}{"extract_mode": "  HTML "},
			wantFormats: []interface{}{"html"},
		},
		{
			// An unknown mode must not be coerced to a default — that would lose
			// the caller's instruction silently, which is the defect being fixed.
			name:        "unknown mode leaves formats alone",
			config:      map[string]interface{}{"extract_mode": "interpretive_dance"},
			wantFormats: nil,
		},
		{
			// The specific dialect beats the alias.
			name: "explicit scrape_config.formats wins",
			config: map[string]interface{}{
				"extract_mode":  "text",
				"scrape_config": map[string]interface{}{"formats": []interface{}{"markdown", "links"}},
			},
			wantFormats: []interface{}{"markdown", "links"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildScrapeConfig(tc.config, "scrape", zap.NewNop())
			if !reflect.DeepEqual(got["formats"], tc.wantFormats) {
				t.Errorf("formats = %#v, want %#v", got["formats"], tc.wantFormats)
			}
		})
	}
}

func TestBuildScrapeConfigMapsCrawlShaping(t *testing.T) {
	// The live vet-practice-verifier config, but as a crawl — where these keys
	// can actually take effect.
	got := buildScrapeConfig(map[string]interface{}{
		"max_pages":    float64(3), // JSONB decodes numbers as float64
		"follow_links": []interface{}{"fees", "prices", "about"},
	}, "crawl", zap.NewNop())

	if got["limit"] != 3 {
		t.Errorf("limit = %#v, want 3 (max_pages must reach the adapter's crawl limit)", got["limit"])
	}
	want := []interface{}{"fees", "prices", "about"}
	if !reflect.DeepEqual(got["include_paths"], want) {
		t.Errorf("include_paths = %#v, want %#v", got["include_paths"], want)
	}
}

func TestBuildScrapeConfigExplicitValuesWin(t *testing.T) {
	got := buildScrapeConfig(map[string]interface{}{
		"max_pages":    float64(3),
		"follow_links": []interface{}{"fees"},
		"scrape_config": map[string]interface{}{
			"limit":         float64(80),
			"include_paths": []interface{}{"news"},
		},
	}, "crawl", zap.NewNop())

	if got["limit"] != float64(80) {
		t.Errorf("limit = %#v, want 80 — a hand-written scrape_config must not be overwritten by an alias", got["limit"])
	}
	if !reflect.DeepEqual(got["include_paths"], []interface{}{"news"}) {
		t.Errorf("include_paths = %#v, want [news]", got["include_paths"])
	}
}

// TestBuildScrapeConfigDoesNotMutateStepConfig guards a leak that would be very
// hard to trace: step config is shared state, so a derived key written back into
// it would follow the definition into later steps and later loop iterations.
func TestBuildScrapeConfigDoesNotMutateStepConfig(t *testing.T) {
	explicit := map[string]interface{}{"only_main_content": false}
	config := map[string]interface{}{
		"extract_mode":  "text",
		"max_pages":     float64(3),
		"scrape_config": explicit,
	}

	buildScrapeConfig(config, "crawl", zap.NewNop())

	if len(config) != 3 {
		t.Errorf("step config gained keys: %#v", config)
	}
	if len(explicit) != 1 {
		t.Errorf("nested scrape_config was mutated: %#v", explicit)
	}
	if _, leaked := explicit["limit"]; leaked {
		t.Error("derived limit leaked into the caller's scrape_config map")
	}
}

func TestBuildScrapeConfigPreservesExplicitConfig(t *testing.T) {
	got := buildScrapeConfig(map[string]interface{}{
		"scrape_config": map[string]interface{}{
			"only_main_content": false,
			"max_age":           float64(0),
		},
	}, "scrape", zap.NewNop())

	// only_main_content:false is the key that bugs_open/101's [UNSETTLED] box was
	// about — it must survive this merge to reach the provider at all.
	if got["only_main_content"] != false {
		t.Errorf("only_main_content = %#v, want false", got["only_main_content"])
	}
	if got["max_age"] != float64(0) {
		t.Errorf("max_age = %#v, want 0", got["max_age"])
	}
}

// TestScrapeWebDeclaresEveryKeyItReads keeps the declaration honest. If someone
// teaches WebscrapeAction a new config key and forgets the spec, the key becomes
// invisible to the validator — the exact failure this bug is about, reintroduced.
func TestScrapeWebDeclaresEveryKeyItReads(t *testing.T) {
	spec, ok := datahelpers.GetActionInputSpec("scrape_web")
	if !ok {
		t.Fatal("scrape_web registers no ActionInputSpec — unknown-key detection is off for it")
	}

	// Every key read by WebscrapeAction or buildScrapeConfig.
	mustDeclare := []string{
		"url_field", "fallback_url_field", "url",
		"upload_results", "scrape_config",
		"max_pages", "follow_links", "extract_mode",
	}

	declared := make(map[string]bool, len(spec.ConfigKeys))
	for _, k := range spec.ConfigKeys {
		declared[k] = true
	}
	for _, k := range mustDeclare {
		if !declared[k] && !datahelpers.IsFrameworkStepConfigKey(k) {
			t.Errorf("scrape_web reads %q but does not declare it — it would be reported as unknown", k)
		}
	}

	// And the live config must now come back clean.
	liveConfig := map[string]interface{}{
		"max_pages":          float64(3),
		"extract_mode":       "text",
		"url_field":          "business_record.business.website_url",
		"fallback_url_field": "search_results.results.0.url",
		"follow_links":       []interface{}{"fees", "prices", "about", "team", "contact", "services"},
	}
	unknown, checked := datahelpers.UnknownConfigKeys("scrape_web", liveConfig)
	if !checked {
		t.Fatal("scrape_web is not opted in to unknown-key detection")
	}
	if len(unknown) != 0 {
		t.Errorf("live vet-practice-verifier config still reports unknown keys: %v", unknown)
	}
}

// TestScrapeWebStillSurfacesTheOriginalCase is the discipline this fix had to learn
// the hard way (WRONG_CALLS.md 2026-07-28): after adding any declaration or
// allow-list, re-run the detector against the case that MOTIVATED it and confirm it
// still fires. Declaring max_pages/follow_links made them recognised, which silenced
// the audit on the two live steps that still advertise a crawl they cannot perform.
func TestScrapeWebStillSurfacesTheOriginalCase(t *testing.T) {
	// The live vet-practice-verifier config, verbatim from agent_definitions.
	liveConfig := map[string]interface{}{
		"max_pages":          float64(3),
		"extract_mode":       "text",
		"url_field":          "business_record.business.website_url",
		"fallback_url_field": "search_results.results.0.url",
		"follow_links":       []interface{}{"fees", "prices", "about", "team", "contact", "services"},
	}

	// It must NOT report as unknown — the keys are implemented or mapped.
	unknown, checked := datahelpers.UnknownConfigKeys("scrape_web", liveConfig)
	if !checked || len(unknown) != 0 {
		t.Fatalf("unknown=%v checked=%v, want none/true", unknown, checked)
	}

	// ...but it MUST still be surfaced, or the fix has merely hidden it.
	conditional, condChecked := datahelpers.ConditionalConfigKeys("scrape_web", liveConfig)
	if !condChecked {
		t.Fatal("scrape_web declares no ConditionalKeys — the audit would report this config clean")
	}
	for _, key := range []string{"max_pages", "follow_links"} {
		if conditional[key] == "" {
			t.Errorf("%q is not surfaced as conditionally honoured; a step carrying it "+
				"still describes a crawl that /scrape cannot perform, and the audit would call it clean", key)
		}
	}

	// And the spec must be internally consistent.
	if missing := datahelpers.UndeclaredConditionalKeys("scrape_web"); len(missing) != 0 {
		t.Errorf("conditional keys missing from ConfigKeys: %v", missing)
	}
}
