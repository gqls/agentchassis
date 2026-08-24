// internal/adapters/imagegenerator/routing_test.go
//
// bugs_open/011 R1. Covers the two behaviours with the widest blast radius
// (council gate e996bf0a, editquality): the hint-fallback path and the hero
// default — plus the unmigrated-kind guard, which is the mechanism behind two
// shipped defects and the one thing here that must never regress quietly.
package imagegenerator

import (
	"strings"
	"testing"
)

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

// bugs_open/382. This test REPLACES TestRouteProviderEmptyKindIsLegacyNotUnmigrated,
// which asserted the exact opposite and pinned the defect in place:
//
//	d := routeProvider("", "")
//	if d.Provider != providerStability { t.Errorf(...) }
//	if d.UnmigratedKind { t.Error("empty kind must not be flagged ...") }
//
// It was not a bad test — it was a faithful test of a premise that turned out
// to be false. The premise was that an empty kind means a legacy caller who
// chose Stability. Production said otherwise: 15 hero assets on 5 sites (as of
// 2026-08-24) came out of SDXL after the hero routing fix, and no site or agent
// had chosen it. Keeping the old assertion is what would have made this bug
// unfixable without an argument.
func TestRouteProviderMissingKindGetsTheStrongProviderAndIsFlagged(t *testing.T) {
	d := routeProvider("", "")
	if d.Provider != providerBanana {
		t.Errorf("routeProvider(\"\", \"\") = %q, want %q — the safe answer to \"nobody said\" is the provider that renders text and honours reference images", d.Provider, providerBanana)
	}
	if !d.MissingKind {
		t.Error("an absent kind MUST set MissingKind — an unflagged fallback is the whole defect of bugs_open/382, and serving it from the strong provider without saying so would hide the broken caller instead of fixing it")
	}
	if d.UnmigratedKind {
		t.Error("an absent kind is not an unmigrated one: the fixes differ (map kind at the caller vs add a row to the table), so the flags must not collapse")
	}
	// Whitespace is absence, not a kind. The table lookup already trims, so a
	// caller mapping an empty-ish field must not fall into the unrouted branch
	// and mis-report itself as a table gap.
	d = routeProvider("   ", "")
	if d.Provider != providerBanana || !d.MissingKind || d.UnmigratedKind {
		t.Errorf("routeProvider(\"   \", \"\") = %+v, want banana + MissingKind only", d)
	}
}

func TestRouteProviderExplicitStabilityStillWinsOnAMissingKind(t *testing.T) {
	// The escape hatch bugs_open/382's fix relies on: a caller that genuinely
	// wants Stability says so, and still gets it, kind or no kind. If this
	// breaks, the fix has removed a capability rather than fixed a default.
	d := routeProvider("", "stability")
	if d.Provider != providerStability {
		t.Errorf("routeProvider(\"\", \"stability\") = %q, want %q", d.Provider, providerStability)
	}
	if d.MissingKind {
		t.Error("an explicit provider choice decided this request; the table never fell through, so MissingKind would be noise on a caller that did nothing wrong")
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
	if got := reportedConditions(d, "hero", "", "p", 0); len(got) != 0 {
		t.Errorf("clean decision produced %d conditions, want 0: %v", len(got), got)
	}
	// bugs_open/382 — this block previously asserted the OPPOSITE ("legacy
	// empty kind is a documented path, not a condition"). A missing kind is
	// now a reported condition, because the durable record is the only thing
	// that finds the caller: this defect ran for six weeks with nothing to
	// query, and surfaced as a person noticing a face.
	d = routeProvider("", "")
	got := reportedConditions(d, "", "", "a wide banner photograph of a veterinary waiting room", 0)
	if len(got) != 1 {
		t.Fatalf("missing kind produced %d conditions, want exactly 1: %v", len(got), got)
	}
	if got[0]["code"] != "MISSING_IMAGE_KIND" {
		t.Errorf("code = %v, want MISSING_IMAGE_KIND", got[0]["code"])
	}
	if got[0]["severity"] != "warning" {
		t.Errorf("severity = %v, want warning — the image is fine, the caller is not", got[0]["severity"])
	}
	ctx, ok := got[0]["context"].(map[string]interface{})
	if !ok {
		t.Fatal("MISSING_IMAGE_KIND has no context map")
	}
	// The row lands on the image-generator's own orchestration, which carries
	// nothing that names the producer, and the parent is reaped in about a
	// day. The prompt opening is what makes the row answerable on its own.
	if opening, _ := ctx["prompt_opening"].(string); opening != "a wide banner photograph of a veterinary waiting room" {
		t.Errorf("context.prompt_opening = %q, want the prompt's opening — without it the row cannot be attributed to a producer once the parent orchestration is reaped", opening)
	}
	if kinds, ok := ctx["routed_kinds"].([]string); !ok || len(kinds) != len(kindProviderRouting) {
		t.Errorf("context.routed_kinds = %v, want the full routed set", ctx["routed_kinds"])
	}
}

func TestMissingKindConditionBoundsThePromptItCarries(t *testing.T) {
	// A condition record is not a place to copy a prompt to. The coordinator
	// caps the NUMBER of conditions per response, not their size, so the bound
	// has to be here.
	long := strings.Repeat("x", 500)
	got := reportedConditions(routeProvider("", ""), "", "", long, 0)
	if len(got) != 1 {
		t.Fatalf("want 1 condition, got %d", len(got))
	}
	ctx := got[0]["context"].(map[string]interface{})
	opening, _ := ctx["prompt_opening"].(string)
	if len([]rune(opening)) > 121 { // 120 runes + the ellipsis
		t.Errorf("prompt_opening is %d runes, want it bounded at 120 plus an ellipsis", len([]rune(opening)))
	}
	if !strings.HasSuffix(opening, "…") {
		t.Errorf("a truncated opening must SAY it was truncated, else a reader compares it against a full prompt and finds no match: %q", opening)
	}
}

func TestMissingKindIsNotReportedWhenAHintDecided(t *testing.T) {
	// Symmetry with UnmigratedKind: the flags describe a fall-through that
	// actually happened. A site that set provider:"stability" made a choice,
	// and a warning on every one of its requests is how a signal stops being
	// read — the lesson 016b records from this very table.
	for _, hint := range []string{"banana", "stability"} {
		if got := reportedConditions(routeProvider("", hint), "", hint, "p", 0); len(got) != 0 {
			t.Errorf("hint %q with no kind produced %v, want no conditions", hint, got)
		}
	}
	// A BAD hint is different: the fall-through DID happen, so both defects
	// are real and both must be reported.
	got := reportedConditions(routeProvider("", "bananna"), "", "bananna", "p", 0)
	codes := map[string]bool{}
	for _, c := range got {
		codes[c["code"].(string)] = true
	}
	if !codes["MISSING_IMAGE_KIND"] || !codes["UNRECOGNISED_PROVIDER_HINT"] {
		t.Errorf("bad hint + no kind reported %v, want both MISSING_IMAGE_KIND and UNRECOGNISED_PROVIDER_HINT", codes)
	}
}

func TestMissingKindDoesNotDropReferenceAnchors(t *testing.T) {
	// Before bugs_open/382 an empty kind went to Stability, so a request
	// carrying reference images ALSO lost its brand anchors and reported
	// REFERENCE_ANCHORS_DROPPED. Routing to Banana fixes that as a side
	// effect; pin it, because a later "tidy-up" of the default would silently
	// bring the anchor loss back.
	got := reportedConditions(routeProvider("", ""), "", "", "p", 3)
	for _, c := range got {
		if c["code"] == "REFERENCE_ANCHORS_DROPPED" {
			t.Error("a missing kind must not drop reference anchors any more — it routes to the provider that honours them")
		}
	}
}

func TestReportedConditionsUnroutedKind(t *testing.T) {
	d := routeProvider("diagram", "")
	got := reportedConditions(d, "diagram", "", "p", 0)
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
	got := reportedConditions(d, "diagram", "bananna", "p", 2)
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
	if got := reportedConditions(d, "hero", "banana", "p", 3); len(got) != 0 {
		t.Errorf("banana with references produced %v, want none", got)
	}
	// A deliberate stability opt-out WITH references does warn: the site
	// chose the provider, but the anchor loss is still real information.
	d = routeProvider("hero", "stability")
	got := reportedConditions(d, "hero", "stability", "p", 3)
	if len(got) != 1 || got[0]["code"] != "REFERENCE_ANCHORS_DROPPED" {
		t.Errorf("stability opt-out with references = %v, want exactly REFERENCE_ANCHORS_DROPPED", got)
	}
}
