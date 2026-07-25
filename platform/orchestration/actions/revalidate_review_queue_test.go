package actions

import (
	"encoding/json"
	"testing"
)

// Real parked-item specs, copied from live rows on 2026-07-25 (bugs_open/033).
// The three producers emit three different shapes for "these fields are
// missing", which is why extraction is tested against all of them rather than a
// tidied-up single form.
const (
	// leopardessconsulting.co.uk /how-we-work, parked 2026-07-10, item d0d5f910
	liveUnresolvedCTASpec = `{
		"fix": "No real page exists to serve as this CTA's destination (no eligible content hub).",
		"source": "resolve_internal_links",
		"missing": ["cta_url", "secondary_cta_url"],
		"component": "hero",
		"page_name": "how-we-work",
		"section_name": "hero"
	}`

	// ai-agent-orchestration.com, parked 2026-07-24
	liveRequiredFieldsSpec = `{
		"check": "required_fields_missing",
		"reason": "schema declares these fields required with source llm, but content_data never received them",
		"page_name": "ai-agent-observability-2025-what-teams-are-actually-monitoring",
		"slot_name": "hero",
		"component_id": "2ae2009c-faae-48e6-b2da-a477c27ff4ab",
		"missing_fields": ["headline"],
		"component_function": "hero"
	}`

	// parked 2026-07-20; note missing[] is objects, not strings
	liveSectionDataSpec = `{
		"source": "plan_sections",
		"missing": [{
			"type": "text",
			"field": "email",
			"reason": "Business contact email address",
			"source": "site_specs.identity.email",
			"on_missing": "needs_human_review"
		}],
		"function": "contact-info",
		"page_name": "contact",
		"component_id": "0bd72302-e9bf-4dc0-a615-41a9c919bf17",
		"section_name": "contact-info"
	}`

	// 15 of the 45 live needs_section_data rows carry this — missing is null
	liveNullMissingSpec = `{
		"source": "plan_sections",
		"missing": null,
		"function": "",
		"page_name": "use-cases",
		"component_id": "",
		"section_name": "call_to_action"
	}`
)

func mustSpec(t *testing.T, raw string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("bad test spec: %v", err)
	}
	return m
}

func TestExtractMissingFieldNames(t *testing.T) {
	t.Run("array of strings — unresolved_cta", func(t *testing.T) {
		got := extractMissingFieldNames(mustSpec(t, liveUnresolvedCTASpec)["missing"])
		if len(got) != 2 || got[0] != "cta_url" || got[1] != "secondary_cta_url" {
			t.Fatalf("want [cta_url secondary_cta_url], got %v", got)
		}
	})

	t.Run("array of strings — required_fields_missing", func(t *testing.T) {
		got := extractMissingFieldNames(mustSpec(t, liveRequiredFieldsSpec)["missing_fields"])
		if len(got) != 1 || got[0] != "headline" {
			t.Fatalf("want [headline], got %v", got)
		}
	})

	t.Run("array of objects — needs_section_data reads .field", func(t *testing.T) {
		got := extractMissingFieldNames(mustSpec(t, liveSectionDataSpec)["missing"])
		if len(got) != 1 || got[0] != "email" {
			t.Fatalf("want [email], got %v", got)
		}
	})

	t.Run("null missing yields no names, not a panic", func(t *testing.T) {
		if got := extractMissingFieldNames(mustSpec(t, liveNullMissingSpec)["missing"]); len(got) != 0 {
			t.Fatalf("want no names from a null missing, got %v", got)
		}
	})

	t.Run("blank and malformed entries are skipped", func(t *testing.T) {
		raw := mustSpec(t, `{"missing":["  ", "real", {"field":" "}, {"nofield":"x"}, 7, null]}`)
		got := extractMissingFieldNames(raw["missing"])
		if len(got) != 1 || got[0] != "real" {
			t.Fatalf("want [real], got %v", got)
		}
	})
}

// The verdict hinges entirely on this predicate, and the finding it re-checks is
// "the template renders this as an empty string". Anything a template renders as
// nothing must read as NOT populated, or the sweep closes live findings.
func TestFieldPopulated(t *testing.T) {
	notPopulated := map[string]interface{}{
		"absent (nil)":  nil,
		"empty string":  "",
		"whitespace":    "   ",
		"empty list":    []interface{}{},
		"empty map":     map[string]interface{}{},
		"boolean false": false,
	}
	for name, v := range notPopulated {
		if fieldPopulated(v) {
			t.Errorf("%s must read as still-missing, read as populated", name)
		}
	}

	populated := map[string]interface{}{
		"url":          "/services.html",
		"list":         []interface{}{"a"},
		"map":          map[string]interface{}{"k": "v"},
		"number":       float64(0), // a real 0 is a supplied value
		"boolean true": true,
	}
	for name, v := range populated {
		if !fieldPopulated(v) {
			t.Errorf("%s must read as populated, read as still-missing", name)
		}
	}
}

func TestSpecString(t *testing.T) {
	t.Run("prefers the first key present", func(t *testing.T) {
		spec := mustSpec(t, liveRequiredFieldsSpec)
		if got := specString(spec, "slot_name", "section_name", "component_function"); got != "hero" {
			t.Fatalf("want hero, got %q", got)
		}
	})

	t.Run("falls through to a later key", func(t *testing.T) {
		spec := mustSpec(t, liveUnresolvedCTASpec) // has section_name, no slot_name
		if got := specString(spec, "slot_name", "section_name"); got != "hero" {
			t.Fatalf("want hero from section_name, got %q", got)
		}
	})

	t.Run("blank values do not count as present", func(t *testing.T) {
		spec := mustSpec(t, `{"slot_name":"   ","section_name":"contact-info"}`)
		if got := specString(spec, "slot_name", "section_name"); got != "contact-info" {
			t.Fatalf("want contact-info, got %q", got)
		}
	})

	t.Run("nothing present yields empty", func(t *testing.T) {
		if got := specString(mustSpec(t, `{"a":1}`), "slot_name", "section_name"); got != "" {
			t.Fatalf("want empty, got %q", got)
		}
	})
}

// Regression guard for the reason this bug exists at all. cta_names_unknown_destination
// is the queue's single largest class (69 live) and it belongs to the
// cta_link_integrity workstream / bugs_open/023, which already knows a chunk of
// them are false positives of its own excluded-area branch. Registering a
// revalidator for it here would put two threads on one check.
func TestRevalidatorCoverageIsDeliberate(t *testing.T) {
	want := []string{"unresolved_cta", "required_fields_missing", "needs_section_data"}
	for _, itemType := range want {
		if _, ok := reviewRevalidators[itemType]; !ok {
			t.Errorf("revalidator for %q is missing", itemType)
		}
	}
	if len(reviewRevalidators) != len(want) {
		t.Errorf("revalidator set changed: want %d entries %v, got %d — if that is deliberate, update this test and say why in the commit",
			len(want), want, len(reviewRevalidators))
	}
	if _, ok := reviewRevalidators["cta_names_unknown_destination"]; ok {
		t.Error("cta_names_unknown_destination is owned by bugs_open/023 / cta_link_integrity — do not revalidate it from here while that check is mid-flight")
	}
}
