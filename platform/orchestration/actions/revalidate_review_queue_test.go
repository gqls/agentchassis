package actions

import (
	"encoding/json"
	"os"
	"strings"
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
//
// needs_page joined the set on 2026-08-03 (bugs_open/187): 28 items of that type
// were parked and no revalidator existed for it, so items whose page was later
// built by another route sat for ever. Its verdicts are pinned in
// page_section_satisfiability_test.go.
func TestRevalidatorCoverageIsDeliberate(t *testing.T) {
	want := []string{"unresolved_cta", "required_fields_missing", "needs_section_data", "needs_page"}
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

// ============================================================================
// The three gates (2026-08-04) — widening `unresolved`/`failed` into the sweep
// ============================================================================

// `status` is checked in THREE places in revalidate_review_queue_action.go: the
// selection in loadParkedReviewItems and the two write-time CAS guards in
// recordRevalidation. Widening only the selection selects the new rows and then
// silently updates nothing — the dispatcher-with-two-gates shape that
// LANDMINES.md records for input_mapping vs a claim query's RETURNING.
//
// This reads the SOURCE rather than exercising the queries because the failure
// is a missed EDIT, not a wrong result: a fourth gate added later, or one gate
// reverted to a literal, is exactly what this must catch. It cannot pass
// vacuously — it fails if it finds no gates at all.
func TestAllThreeStatusGatesUseTheSharedList(t *testing.T) {
	src, err := os.ReadFile("revalidate_review_queue_action.go")
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	body := string(src)

	shared := strings.Count(body, "sqlInList(workItemRevalidatableStatuses)")
	if shared == 0 {
		t.Fatal("found no uses of the shared status list — this test's needle has stopped " +
			"matching, so it can no longer fail; fix it rather than trusting the pass")
	}
	if shared != 3 {
		t.Errorf("the shared status list is interpolated %d times, want exactly 3 "+
			"(selection + 2 CAS guards). A gate that does not carry it will select rows it "+
			"then fails to update, which looks like 'the sweep found nothing to do'.", shared)
	}

	// The literal that used to be in all three. Any survivor is a drifted gate.
	if n := strings.Count(body, "status = 'needs_human_review'"); n != 0 {
		t.Errorf("%d gate(s) still hard-code `status = 'needs_human_review'` — they will not "+
			"see the `unresolved`/`failed` rows the other gates now select", n)
	}
}

// The list's CONTENTS carry the safety argument, so pin them: `unresolved` and
// `failed` are in deliberately (they mean "we gave up" and "the handler errored",
// never "this stopped being a problem" — RFC_010 Decision 2), and no already-
// closed status may be here or the sweep would re-open settled rows.
func TestRevalidatableStatusesAreTheIntendedSet(t *testing.T) {
	got := map[string]bool{}
	for _, s := range workItemRevalidatableStatuses {
		got[s] = true
	}
	for _, want := range []string{"needs_human_review", "unresolved"} {
		if !got[want] {
			t.Errorf("%q must be revalidatable — dropping it re-creates the queue this sweep "+
				"cannot drain (its own closes feed insertWorkItem's two-strike counter)", want)
		}
	}
	for _, closed := range workItemClosedStatuses {
		if got[closed] {
			t.Errorf("%q is a CLOSED status and must never be revalidated — a settled row "+
				"would be re-examined and could be re-stamped", closed)
		}
	}
	// `failed` must STAY OUT: 17 needs_page rows sit in it, parked by
	// FailWorkItemAction's status_override branch, which this action's header
	// defers to owner decision 033 D2. Adding it here would overrule that from
	// inside an unrelated change.
	for _, s := range workItemRevalidatableStatuses {
		if s == "failed" {
			t.Error("`failed` must not be revalidatable — it pulls in the 033 D2 population " +
				"(17 needs_page rows) that this sweep's own header defers by name")
		}
	}
	if len(workItemRevalidatableStatuses) != 2 {
		t.Errorf("unexpected size %d: every addition widens what an automated sweep may close, "+
			"so state the reason in the list's comment before growing it",
			len(workItemRevalidatableStatuses))
	}
}
