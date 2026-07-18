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
