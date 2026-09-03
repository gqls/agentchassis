// FILE: platform/livespec/rerender_reason_stamping_test.go
//
// MUTATION CHECKS for whoever edits the stamping helpers:
//   - drop the `known` guard in RerenderReasonFields (stamp the routing key
//     unconditionally) → TestRerenderReasonFields_UnknownGetsNoRoutingKey fails.
//   - hand-write RerenderReasonJSONPrefix instead of deriving it →
//     TestRerenderReasonJSONPrefixMatchesTheMapForm is what catches the drift;
//     verify it fails by changing one key's spelling in either form.

package livespec

import (
	"encoding/json"
	"testing"
)

func TestRerenderReasonFields_EveryVocabularyValueGetsBoth(t *testing.T) {
	for _, r := range RerenderSectionReasons {
		f := RerenderReasonFields(r.Name)
		if f["reason"] != r.Name || f[RoutingReasonSpecKey] != r.Name {
			t.Errorf("%q: got %v — a vocabulary value must carry BOTH fields, identical, or the "+
				"phase-3 flip loses this producer's items to assemble", r.Name, f)
		}
	}
}

func TestRerenderReasonFields_UnknownGetsNoRoutingKey(t *testing.T) {
	// Real observed out-of-vocabulary values (bugs_open/440's census) plus free
	// prose of the shape migrations 696/693 mint. The annotation must survive
	// untouched (owner ruling D4: `reason` is never validated) and the routing
	// key must be ABSENT — not empty-but-present, which a phase-3 reader would
	// have to special-case.
	for _, unknown := range []string{
		"tool_retirement",
		"light_palette_chrome_replaced",
		"FCA rule citation corrected by migration 696 (owner decision 2026-09-02)",
	} {
		if _, declared := RerenderSectionReasonByName(unknown); declared {
			t.Fatalf("%q is now IN the vocabulary and no longer poisons this test — pick another "+
				"out-of-vocabulary value; do NOT delete the case", unknown)
		}
		f := RerenderReasonFields(unknown)
		if f["reason"] != unknown {
			t.Errorf("%q: the annotation must be written as given, got %q", unknown, f["reason"])
		}
		if _, present := f[RoutingReasonSpecKey]; present {
			t.Errorf("%q: produced a routing key — phase 3 would then refuse an item the estate's "+
				"own producer minted (REB-008)", unknown)
		}
	}
}

func TestRerenderReasonFields_EmptyReasonStampsNothing(t *testing.T) {
	// The assemble-only case: no reason in, no fields out. The item creator's
	// own measurement is that this is the overwhelming majority of traffic.
	if f := RerenderReasonFields(""); len(f) != 0 {
		t.Fatalf("an empty reason produced %v — an item with no reason is assemble-only and must "+
			"stay exactly that", f)
	}
	if s := RerenderReasonJSONPrefix(""); s != "" {
		t.Fatalf("an empty reason produced the fragment %q, which would corrupt a caller's spec", s)
	}
	// AND the empty case must still COMPOSE into valid JSON at a call site —
	// this is the assertion that makes the trailing-comma design load-bearing
	// rather than stylistic.
	var probe map[string]string
	if err := json.Unmarshal([]byte("{"+RerenderReasonJSONPrefix("")+`"page_name":"about"}`), &probe); err != nil {
		t.Fatalf("an empty reason broke a caller's spec JSON: %v — a helper must not have a shape "+
			"that only works while every caller passes a constant", err)
	}
}

func TestStampRerenderReason_WritesIntoASpecMap(t *testing.T) {
	spec := map[string]interface{}{"page_name": "about"}
	StampRerenderReason(spec, ReasonCTALinksStale)
	if spec["reason"] != ReasonCTALinksStale || spec[RoutingReasonSpecKey] != ReasonCTALinksStale {
		t.Fatalf("got %v, want both fields set", spec)
	}
	if spec["page_name"] != "about" {
		t.Fatal("the helper must not disturb the caller's own keys")
	}
	StampRerenderReason(spec, "")
	if spec["reason"] != ReasonCTALinksStale {
		t.Fatal("an empty reason must be a no-op, not an erasure")
	}
}

func TestRerenderReasonJSONPrefixMatchesTheMapForm(t *testing.T) {
	// The two renderings are one rule; this is what stops them drifting. Wrap
	// the fragment back into an object and require it to equal the map form.
	for _, reason := range []string{ReasonCTALinksStale, ReasonTemplateChanged, "tool_retirement"} {
		frag := RerenderReasonJSONPrefix(reason)
		// Compose it the way the call sites do — prefix then another field —
		// so the trailing comma is exercised rather than assumed.
		var got map[string]string
		if err := json.Unmarshal([]byte("{"+frag+`"page_name":"about"}`), &got); err != nil {
			t.Fatalf("%q: fragment %q does not parse when wrapped: %v — a caller splicing this "+
				"into a spec template would emit invalid JSON", reason, frag, err)
		}
		want := RerenderReasonFields(reason)
		if got["page_name"] != "about" {
			t.Errorf("%q: the caller's own field was lost composing %q", reason, frag)
		}
		delete(got, "page_name")
		if len(got) != len(want) {
			t.Fatalf("%q: fragment gave %v, map form gave %v", reason, got, want)
		}
		for k, v := range want {
			if got[k] != v {
				t.Errorf("%q: fragment[%s]=%q want %q", reason, k, got[k], v)
			}
		}
	}
}
