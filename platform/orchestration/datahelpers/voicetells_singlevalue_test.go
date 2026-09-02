// FILE: platform/orchestration/datahelpers/voicetells_singlevalue_test.go
//
// bugs_open/338 — the voice gate's corpus statistics applied to a single
// sentence, where they are not measurements.
//
// The gate for leopardessconsulting.co.uk is reproduced here from the LIVE
// site_specs row (aspect='voice', read 2026-09-02), not composed: a fixture
// invented to suit the fix would exercise its own rule rather than the one
// production runs. Only the banned_phrases list is trimmed, and the two kept
// are copied verbatim.
package datahelpers

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The real leopardess shape: NO thresholds set at all, so every trip is the
// package default (mean 22, long 25, share 0.30, em-dash 3/1000). That is what
// makes it one of the only two sites bugs_open/338 bites.
const liveLeopardessGateJSON = `{
  "voice_gate": {
    "enabled": true,
    "expect_contractions": true,
    "banned_phrases": [
      {"pattern": "\\bharness(ing)?\\b", "reason": "banned_language: hype verb"},
      {"pattern": "\\btrust(ed|worthy|s)?\\b", "reason": "owner 2026-07-18: overused; say what is checked/verified instead"}
    ]
  }
}`

func liveGate(t *testing.T) *VoiceGate {
	t.Helper()
	g, err := ParseVoiceGate([]byte(liveLeopardessGateJSON))
	if err != nil || g == nil {
		t.Fatalf("ParseVoiceGate: %v (gate nil: %v)", err, g == nil)
	}
	return g
}

// The exact candidate production refused, quoted from the bug file, which
// quoted it from orchestration_states.collected_data.save_result. 24 words
// against a default mean trip of 22.
const refusedCandidate = "Read research-backed insights on AI adoption across healthcare, finance, hiring and data security, with findings on what builds and breaks confidence in these systems."

// ARM 2 of bugs_open/338 §6 — the clean 24-word candidate must now be WRITTEN.
//
// The first assertion is the CONTROL: it proves the sentence really does trip
// the old path, so the second assertion is a change in behaviour and not a
// candidate that was always going to pass. Without it this test would be
// indistinguishable from one run against harmless copy.
func TestSingleValue_GoodLongDescriptionNoLongerRefused(t *testing.T) {
	g := liveGate(t)

	old := checksIn(g.ScanVoice([]string{refusedCandidate}, false))
	if old["long_sentences"] == 0 {
		t.Fatalf("CONTROL FAILED: the candidate no longer trips long_sentences on the page path, so this test proves nothing about the fix: %v", old)
	}

	got := checksIn(g.ScanVoiceSingleValue(refusedCandidate))
	if len(got) != 0 {
		t.Errorf("single-value scan must publish the candidate, got findings %v", got)
	}
}

// ARM 1 of §6 — a banned phrase must STILL be refused. A test that only
// exercised arm 2 is indistinguishable from one that removed the gate.
func TestSingleValue_BannedPhraseStillRefused(t *testing.T) {
	g := liveGate(t)
	got := checksIn(g.ScanVoiceSingleValue("A platform that helps teams harness their data and build trust with regulators."))
	if got["banned_phrase"] == 0 {
		t.Errorf("banned_phrase must survive the single-value filter, got %v", got)
	}
}

// The em-dash rule is KEPT, against bugs_open/338 §4's suggestion to replace it
// with a flat test. It is a rate over WORDS, so it already reduces to "contains
// an em dash" at this length — and unlike a flat test it still honours the
// seven sites that set the trip to 100000 to switch the rule off.
func TestSingleValue_EmDashRuleTravelsAndStillHonoursOptOut(t *testing.T) {
	const withEmDash = "Read research-backed insights on AI adoption — the findings on what builds confidence."

	if got := checksIn(liveGate(t).ScanVoiceSingleValue(withEmDash)); got["em_dash_density"] == 0 {
		t.Errorf("default-threshold site: an em dash in a single value must be refused, got %v", got)
	}

	// The opt-out profile the other seven sites actually carry.
	optedOut, err := ParseVoiceGate([]byte(`{"voice_gate":{"enabled":true,"em_dash_per_1000_words":100000,"banned_phrases":[]}}`))
	if err != nil || optedOut == nil {
		t.Fatalf("ParseVoiceGate(opt-out): %v", err)
	}
	if got := checksIn(optedOut.ScanVoiceSingleValue(withEmDash)); got["em_dash_density"] != 0 {
		t.Errorf("a site that switched the em-dash rule off must not be re-gated by the single-value path, got %v", got)
	}
}

// Every corpus-only check must be dropped, and each is induced individually so
// a pass cannot come from the input failing to trip it in the first place.
func TestSingleValue_CorpusOnlyChecksAreDropped(t *testing.T) {
	cases := []struct {
		check string
		gate  string
		value string
	}{
		{"long_sentences", liveLeopardessGateJSON, refusedCandidate},
		// min_sentences lowered to 1 so no_contractions CAN fire at n=1 —
		// on today's default of 15 it never fires and the case would be vacuous.
		{"no_contractions", `{"voice_gate":{"enabled":true,"expect_contractions":true,"min_sentences_for_contraction_check":1}}`,
			"The platform records every check it runs and reports the result."},
		{"triad_density", `{"voice_gate":{"enabled":true,"triads_per_page":1}}`,
			"We build data, tools, and systems for teams that ship models, agents, and pipelines."},
		{"negation_density", `{"voice_gate":{"enabled":true,"contrasts_per_page":1}}`,
			"It is a record, not a claim, and a check rather than a promise, and it doesn't guess."},
	}

	for _, tc := range cases {
		t.Run(tc.check, func(t *testing.T) {
			g, err := ParseVoiceGate([]byte(tc.gate))
			if err != nil || g == nil {
				t.Fatalf("ParseVoiceGate: %v", err)
			}
			if before := checksIn(g.ScanVoice([]string{tc.value}, false)); before[tc.check] == 0 {
				t.Fatalf("CONTROL FAILED: %s did not fire on the page path, so dropping it proves nothing: %v", tc.check, before)
			}
			if after := checksIn(g.ScanVoiceSingleValue(tc.value)); after[tc.check] != 0 {
				t.Errorf("%s must not gate a single value, got %v", tc.check, after)
			}
		})
	}
}

// THE ANTI-DRIFT GUARD. bugs_open/338 §4 enumerated the check names and was
// already stale when picked up — negation_density had been added by
// bugs_open/305 and was in no list. This fails on the NEXT such addition, so
// the classification cannot silently miss a check.
//
// MUTATION THAT KILLS IT: delete any entry from voiceCheckKinds, or add a
// `Check: "whatever"` to ScanVoice without classifying it.
func TestEveryVoiceCheckIsClassified(t *testing.T) {
	src, err := os.ReadFile("voicetells.go")
	if err != nil {
		t.Fatalf("read voicetells.go: %v", err)
	}
	// Only the emission sites: `Check: "name"` inside a VoiceFinding literal.
	// The classification map uses `"name": VoiceCheck…`, so it cannot match
	// itself and vouch for its own completeness.
	emitted := map[string]bool{}
	for _, m := range regexp.MustCompile(`Check:\s*"([a-z_]+)"`).FindAllStringSubmatch(string(src), -1) {
		emitted[m[1]] = true
	}
	if len(emitted) == 0 {
		t.Fatal("found no Check: literals — the scan pattern has drifted from the source")
	}
	for check := range emitted {
		if _, ok := VoiceCheckKindOf(check); !ok {
			t.Errorf("check %q is emitted by ScanVoice but not classified in voiceCheckKinds — "+
				"decide whether it survives a single-value field and add it", check)
		}
	}
	for check := range voiceCheckKinds {
		if !emitted[check] {
			t.Errorf("voiceCheckKinds classifies %q, which ScanVoice no longer emits — stale entry", check)
		}
	}
	if t.Failed() {
		names := make([]string, 0, len(emitted))
		for k := range emitted {
			names = append(names, k)
		}
		sort.Strings(names)
		t.Logf("emitted checks: %v", strings.Join(names, ", "))
	}
}
