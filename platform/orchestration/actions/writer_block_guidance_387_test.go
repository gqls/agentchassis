// FILE: platform/orchestration/actions/writer_block_guidance_387_test.go
//
// writer_block_guidance: human-owned prose carried VERBATIM through writer-block
// regeneration, on BOTH composition paths (site-wide managed regeneration and
// the per-section scoped block, which REPLACES the site-wide block in that
// section's prompt). bugs_open/387: 13 of 19 writer_block sites stayed
// unmanaged because regeneration deleted their hand-written NEVER-STATE lists,
// and a stand-in token hand-typed into one of those unmanaged blocks reached
// the public. Built as agreed with the bugs_open/288 lane (owner of the
// regeneration path).
//
// The absent-key case pins BYTE-IDENTICAL output — the opt-in contract: zero
// consumers existed when this shipped, and nothing changes until a site sets
// the key.

package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func guidanceEB(withGuidance bool, facts ...map[string]interface{}) map[string]interface{} {
	fr := make([]interface{}, 0, len(facts))
	for _, f := range facts {
		fr = append(fr, f)
	}
	eb := map[string]interface{}{
		"facts":            fr,
		"allowed_entities": []interface{}{"Companies House"},
	}
	if withGuidance {
		eb["writer_block_guidance"] = "NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, satisfaction rates, uptime percentages."
	}
	return eb
}

func guidanceFact(id, line string, value interface{}) map[string]interface{} {
	f := map[string]interface{}{"id": id, "writer_line": line}
	if value != nil {
		f["value"] = value
	}
	return f
}

func TestWriterBlockGuidanceAbsentIsByteIdentical(t *testing.T) {
	fact := guidanceFact("F1", "{value} live sites in production", 27.0)
	without := composeWriterBlock(guidanceEB(false, fact))
	present := guidanceEB(true, fact)
	delete(present, "writer_block_guidance")
	if got := composeWriterBlock(present); got != without {
		t.Fatalf("absent key must be byte-identical to the pre-carry output:\n%q\nvs\n%q", got, without)
	}
	if strings.Contains(without, "NEVER STATE") {
		t.Fatalf("guidance leaked into a composition with no key: %q", without)
	}
}

func TestWriterBlockGuidanceAppendedVerbatimAndLast(t *testing.T) {
	fact := guidanceFact("F1", "{value} live sites in production", 27.0)
	got := composeWriterBlock(guidanceEB(true, fact))
	want := "NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, satisfaction rates, uptime percentages."
	if !strings.Contains(got, want) {
		t.Fatalf("guidance not carried verbatim: %q", got)
	}
	if !strings.HasSuffix(got, want) {
		t.Fatalf("guidance must be the FINAL section (after entities), got: %q", got)
	}
	if !strings.Contains(got, "27 live sites in production") {
		t.Fatalf("the phrased fact went missing alongside the guidance: %q", got)
	}
}

func TestWriterBlockGuidanceAloneNeverReplacesAHandBlock(t *testing.T) {
	// No phrased facts at all: the nothing-phrased early return must still win,
	// so managed regeneration leaves a hand-written block alone rather than
	// replacing it with bare guidance.
	eb := guidanceEB(true)
	if got := composeWriterBlock(eb); got != "" {
		t.Fatalf("guidance alone must not compose a block (the hand-written block would be replaced): %q", got)
	}
}

func TestScopedWriterBlockCarriesGuidance(t *testing.T) {
	fact := guidanceFact("F1", "{value} live sites in production", 27.0)
	eb := guidanceEB(true, fact)
	got := composeScopedWriterBlock(eb, []string{"F1"}, zap.NewNop(), "hero")
	if !strings.Contains(got, "NEVER STATE") {
		t.Fatalf("the scoped block REPLACES the site-wide block in the prompt; without the carry the section loses the guidance: %q", got)
	}
	// And the scoped absent-key contract holds too.
	ebNo := guidanceEB(false, fact)
	if got := composeScopedWriterBlock(ebNo, []string{"F1"}, zap.NewNop(), "hero"); strings.Contains(got, "NEVER STATE") {
		t.Fatalf("guidance appeared with no key set: %q", got)
	}
}
