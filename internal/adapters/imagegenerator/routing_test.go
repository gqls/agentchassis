// internal/adapters/imagegenerator/routing_test.go
//
// bugs_open/011 R1. Covers the two behaviours with the widest blast radius
// (council gate e996bf0a, editquality): the hint-fallback path and the hero
// default — plus the unmigrated-kind guard, which is the mechanism behind two
// shipped defects and the one thing here that must never regress quietly.
package imagegenerator

import "testing"

func TestRouteProviderByKind(t *testing.T) {
	// Every routed kind goes to Banana. hero is the bugs_open/011 change:
	// it was the last kind on Stability and the fleet's largest.
	for _, kind := range []string{
		"icon", "logo", "illustration", "infographic",
		"sprite_sheet", "content_hero", "hero",
	} {
		d := routeProvider(kind, "")
		if d.Provider != providerBanana {
			t.Errorf("routeProvider(%q, \"\") = %q, want %q", kind, d.Provider, providerBanana)
		}
		if d.UnmigratedKind {
			t.Errorf("routeProvider(%q, \"\") flagged UnmigratedKind — %q is in the routing table", kind, kind)
		}
	}
}

func TestRouteProviderEmptyKindIsLegacyNotUnmigrated(t *testing.T) {
	// Legacy callers predating the kind field are a deliberate Stability
	// path. Flagging them would make the unmigrated-kind warning noise, and a
	// warning that fires constantly is one nobody reads.
	d := routeProvider("", "")
	if d.Provider != providerStability {
		t.Errorf("routeProvider(\"\", \"\") = %q, want %q", d.Provider, providerStability)
	}
	if d.UnmigratedKind {
		t.Error("empty kind must not be flagged as unmigrated — it is the documented legacy path")
	}
}

func TestRouteProviderUnmigratedKindIsFlagged(t *testing.T) {
	// The mechanism behind content_hero and then hero: a kind nobody added to
	// the table lands on Stability with no signal. It still routes — image
	// generation must not fail on an unknown kind — but it must SAY SO.
	d := routeProvider("diagram", "")
	if d.Provider != providerStability {
		t.Errorf("routeProvider(diagram) = %q, want %q (fallback must still route)", d.Provider, providerStability)
	}
	if !d.UnmigratedKind {
		t.Error("an unrouted non-empty kind MUST set UnmigratedKind — this guard is the whole point of bugs_open/011's mechanism fix")
	}
}

func TestRouteProviderHintWins(t *testing.T) {
	// The site's own preference beats the per-kind default in both directions.
	if d := routeProvider("hero", "stability"); d.Provider != providerStability {
		t.Errorf("hint stability on hero = %q, want %q — a photographic site must be able to keep SDXL heroes", d.Provider, providerStability)
	}
	if d := routeProvider("", "banana"); d.Provider != providerBanana {
		t.Errorf("hint banana on empty kind = %q, want %q", d.Provider, providerBanana)
	}
	// A deliberate stability choice is not an unmigrated kind.
	if d := routeProvider("hero", "stability"); d.UnmigratedKind {
		t.Error("an explicit stability hint is a choice, not an unmigrated kind")
	}
}

func TestRouteProviderHintIsCaseAndSpaceInsensitive(t *testing.T) {
	// The field is hand-written into a JSON spec; " Banana" should not
	// silently become an unrecognised hint.
	for _, hint := range []string{"Banana", " banana ", "BANANA"} {
		d := routeProvider("", hint)
		if d.Provider != providerBanana || d.BadHint {
			t.Errorf("routeProvider(\"\", %q) = %+v, want banana with no BadHint", hint, d)
		}
	}
}

func TestRouteProviderBadHintFallsBackToKind(t *testing.T) {
	// A typo in one site's style guide must degrade to the kind default, not
	// fail generation for that site — but it must be visible.
	d := routeProvider("hero", "bananna")
	if !d.BadHint {
		t.Error("unrecognised hint must set BadHint")
	}
	if d.Provider != providerBanana {
		t.Errorf("bad hint on hero = %q, want the kind default %q", d.Provider, providerBanana)
	}
	// Bad hint on an unrouted kind must still surface BOTH problems.
	d2 := routeProvider("diagram", "bananna")
	if !d2.BadHint || !d2.UnmigratedKind {
		t.Errorf("bad hint + unrouted kind = %+v, want both flags set", d2)
	}
}

func TestKnownRoutedKindsIsSortedAndComplete(t *testing.T) {
	kinds := knownRoutedKinds()
	if len(kinds) != len(kindProviderRouting) {
		t.Errorf("knownRoutedKinds() returned %d kinds, table has %d", len(kinds), len(kindProviderRouting))
	}
	for i := 1; i < len(kinds); i++ {
		if kinds[i-1] > kinds[i] {
			t.Errorf("knownRoutedKinds() not sorted: %v", kinds)
			break
		}
	}
}

// bugs_open/011 §4 residual — the flags must become condition RECORDS, not
// just log lines, or detection depends on someone tailing the right pod.

func TestReportedConditionsCleanDecisionReportsNothing(t *testing.T) {
	// The healthy case must produce an ABSENT field, not an empty list —
	// noise in every response body is how a signal stops being read.
	d := routeProvider("hero", "")
	if got := reportedConditions(d, "hero", "", 0); len(got) != 0 {
		t.Errorf("clean decision produced %d conditions, want 0: %v", len(got), got)
	}
	// Legacy empty kind is a documented path, not a condition.
	d = routeProvider("", "")
	if got := reportedConditions(d, "", "", 0); len(got) != 0 {
		t.Errorf("legacy empty kind produced %d conditions, want 0", len(got))
	}
}

func TestReportedConditionsUnroutedKind(t *testing.T) {
	d := routeProvider("diagram", "")
	got := reportedConditions(d, "diagram", "", 0)
	if len(got) != 1 {
		t.Fatalf("unrouted kind produced %d conditions, want 1", len(got))
	}
	c := got[0]
	if c["code"] != "UNROUTED_IMAGE_KIND" {
		t.Errorf("code = %v, want UNROUTED_IMAGE_KIND", c["code"])
	}
	if c["severity"] != "warning" {
		t.Errorf("severity = %v, want warning", c["severity"])
	}
	// The record must name the valid set — the row is read by someone who
	// does not have the source open.
	ctx, ok := c["context"].(map[string]interface{})
	if !ok {
		t.Fatal("condition has no context map")
	}
	kinds, ok := ctx["routed_kinds"].([]string)
	if !ok || len(kinds) != len(kindProviderRouting) {
		t.Errorf("context.routed_kinds = %v, want the full routed set", ctx["routed_kinds"])
	}
}

func TestReportedConditionsBadHintAndAnchorsDropped(t *testing.T) {
	// A bad hint on an unrouted kind with references carries all three
	// conditions — each is a separate defect with a separate fix.
	d := routeProvider("diagram", "bananna")
	got := reportedConditions(d, "diagram", "bananna", 2)
	if len(got) != 3 {
		t.Fatalf("produced %d conditions, want 3 (unrouted + bad hint + anchors dropped): %v", len(got), got)
	}
	codes := map[string]bool{}
	for _, c := range got {
		code, _ := c["code"].(string)
		codes[code] = true
	}
	for _, want := range []string{"UNROUTED_IMAGE_KIND", "UNRECOGNISED_PROVIDER_HINT", "REFERENCE_ANCHORS_DROPPED"} {
		if !codes[want] {
			t.Errorf("missing condition %s in %v", want, codes)
		}
	}
}

func TestReportedConditionsNoAnchorsDroppedOnBanana(t *testing.T) {
	// Banana honours references — routing there with references is the
	// healthy case whatever the kind's table entry says.
	d := routeProvider("hero", "banana")
	if got := reportedConditions(d, "hero", "banana", 3); len(got) != 0 {
		t.Errorf("banana with references produced %v, want none", got)
	}
	// A deliberate stability opt-out WITH references does warn: the site
	// chose the provider, but the anchor loss is still real information.
	d = routeProvider("hero", "stability")
	got := reportedConditions(d, "hero", "stability", 3)
	if len(got) != 1 || got[0]["code"] != "REFERENCE_ANCHORS_DROPPED" {
		t.Errorf("stability opt-out with references = %v, want exactly REFERENCE_ANCHORS_DROPPED", got)
	}
}
