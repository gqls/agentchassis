// FILE: platform/orchestration/datahelpers/claims_fact_kind_test.go
//
// bugs_open/105 — EvidenceFact.Kind was declared, documented in the spec,
// written by nine live registers, and read by nothing. A fact declared
// kind:"capability" behaved identically to kind:"metric", kind:"banana", or the
// field being absent.
//
// The fixtures below are the LIVE vocabulary, measured on 2026-07-27 across the
// nine current evidence_base registers — metric 46, count 18, entity 11,
// capability 9, attestation 4 — not the documented one. `count` is used by four
// sites and appears in no spec, which is why "reject unknown kinds" (the bug
// file's candidate 1 as written) would have failed four sites' registers closed.
package datahelpers

import (
	"encoding/json"
	"testing"
)

func TestCanonicalKind(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// the documented vocabulary, unchanged
		{"metric", FactKindMetric},
		{"capability", FactKindCapability},
		{"entity", FactKindEntity},
		{"attestation", FactKindAttestation},
		// the live-but-undocumented spelling, used by four sites
		{"count", FactKindMetric},
		// tolerated shapes
		{"", FactKindMetric},
		{"  Metric  ", FactKindMetric},
		{"ATTESTATION", FactKindAttestation},
		// an unknown kind is treated as a metric — the safe default — rather
		// than failing the register closed
		{"banana", FactKindMetric},
	}
	for _, c := range cases {
		if got := (EvidenceFact{Kind: c.in}).CanonicalKind(); got != c.want {
			t.Errorf("CanonicalKind(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestKindIsRecognised is the half that makes the default honest: treating an
// unknown kind as a metric is only acceptable because the caller can still tell
// that is what happened.
func TestKindIsRecognised(t *testing.T) {
	for _, k := range []string{"metric", "count", "capability", "entity", "attestation", ""} {
		if !(EvidenceFact{Kind: k}).KindIsRecognised() {
			t.Errorf("kind %q should be recognised", k)
		}
	}
	for _, k := range []string{"banana", "Metrik", "promise"} {
		if (EvidenceFact{Kind: k}).KindIsRecognised() {
			t.Errorf("kind %q should NOT be recognised — it is being silently "+
				"treated as a metric and the caller must be able to say so", k)
		}
	}
}

// TestUnrecognisedKinds is what a caller reports. It must be EMPTY for the live
// vocabulary — if this fails, the alias table has fallen behind the fleet and
// somebody's register is being silently reinterpreted.
func TestUnrecognisedKinds(t *testing.T) {
	live := &EvidenceBase{Facts: []EvidenceFact{
		{Kind: "metric"}, {Kind: "count"}, {Kind: "entity"},
		{Kind: "capability"}, {Kind: "attestation"}, {Kind: ""},
	}}
	if got := live.UnrecognisedKinds(); len(got) != 0 {
		t.Errorf("the live 2026-07-27 vocabulary reports unrecognised kinds %v — "+
			"the alias table has fallen behind the fleet", got)
	}

	mixed := &EvidenceBase{Facts: []EvidenceFact{
		{Kind: "metric"}, {Kind: "banana"}, {Kind: "promise"}, {Kind: "banana"},
	}}
	got := mixed.UnrecognisedKinds()
	if len(got) != 2 || got[0] != "banana" || got[1] != "promise" {
		t.Errorf("UnrecognisedKinds() = %v, want [banana promise] deduplicated and sorted", got)
	}

	if (*EvidenceBase)(nil).UnrecognisedKinds() != nil {
		t.Error("a nil register must not panic — callers treat nil as 'site not opted in'")
	}
}

// TestParseEvidenceBaseDoesNotRewriteKind is the one that matters for safety.
// EvidenceBase is marshalled BACK to site_specs by refresh_evidence_base_action
// and evidence_citations, so normalising Kind at parse time would silently
// rewrite 18 stored facts across four sites from "count" to "metric" through a
// write path that never intended to touch them.
func TestParseEvidenceBaseDoesNotRewriteKind(t *testing.T) {
	raw := []byte(`{"audit_doc":"x","facts":[
		{"id":"C1","claim":"c","kind":"count"},
		{"id":"C2","claim":"c","kind":"banana"}
	]}`)
	eb, err := ParseEvidenceBase(raw)
	if err != nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	if eb.Facts[0].Kind != "count" {
		t.Errorf("stored Kind was rewritten to %q — parse must not mutate the register, "+
			"because it is marshalled back to site_specs", eb.Facts[0].Kind)
	}
	if eb.Facts[1].Kind != "banana" {
		t.Errorf("stored Kind was rewritten to %q", eb.Facts[1].Kind)
	}
	// …and a round-trip must be byte-stable on the field.
	out, _ := json.Marshal(eb)
	var back EvidenceBase
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("round-trip: %v", err)
	}
	if back.Facts[0].Kind != "count" || back.Facts[1].Kind != "banana" {
		t.Errorf("round-trip changed the stored kinds: %q, %q",
			back.Facts[0].Kind, back.Facts[1].Kind)
	}
}

// TestAliasedKinds answers the council's guardian objection: the count->metric
// mapping is an interpretive judgement over 18 facts on 4 live sites, and
// resolving it silently means nothing surfaces if the guess is wrong. The
// register must be able to say which kinds it had reinterpreted.
func TestAliasedKinds(t *testing.T) {
	eb := &EvidenceBase{Facts: []EvidenceFact{
		{Kind: "metric"},      // canonical, not an alias
		{Kind: "count"},       // aliased
		{Kind: "count"},       // same alias, reported once
		{Kind: "attestation"}, // canonical
		{Kind: "banana"},      // unrecognised, not an alias
		{Kind: ""},            // absent, not an alias
	}}
	got := eb.AliasedKinds()
	if len(got) != 1 || got[0] != "count -> metric" {
		t.Errorf("AliasedKinds() = %v, want [\"count -> metric\"]", got)
	}
	// An unrecognised kind is a DIFFERENT signal and must not be folded in here.
	if u := eb.UnrecognisedKinds(); len(u) != 1 || u[0] != "banana" {
		t.Errorf("UnrecognisedKinds() = %v, want [banana]", u)
	}
	if (*EvidenceBase)(nil).AliasedKinds() != nil {
		t.Error("nil register must not panic")
	}
}
