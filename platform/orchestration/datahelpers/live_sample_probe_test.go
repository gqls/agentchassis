package datahelpers

import (
	"encoding/json"
	"os"
	"testing"
)

// TestLiveSampleEscalationRate is a PROBE, not an assertion: run it with
// NOVAL_SAMPLE=<file> to measure what the shipped rule would do to real
// prompts. Skipped without the env var so it never runs in CI.
func TestLiveSampleEscalationRate(t *testing.T) {
	path := os.Getenv("NOVAL_SAMPLE")
	if path == "" {
		t.Skip("set NOVAL_SAMPLE to a JSON dump of live prompt_rendered rows")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Agent, Step, Rendered string
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	var withHoles, escalated, totalOcc, totalAuth int
	byAgent := map[string]int{}
	for _, r := range rows {
		// tpl is unavailable for a logged row, so attribution is skipped; the
		// counts under test here are the ones read from the OUTPUT.
		rep := ScanMissingValues("", r.Rendered, nil)
		if rep.Empty() {
			continue
		}
		withHoles++
		totalOcc += rep.Occurrences
		totalAuth += rep.Authoritative
		if rep.Authoritative > 0 {
			escalated++
			byAgent[r.Agent]++
			if escalated <= 2 {
				t.Logf("ESCALATES  %s/%s  occurrences=%d authoritative=%d\n    context: %q",
					r.Agent, r.Step, rep.Occurrences, rep.Authoritative, rep.Contexts[0])
			}
		}
	}
	t.Logf("rows=%d with_holes=%d total_occurrences=%d in_authoritative_blocks=%d",
		len(rows), withHoles, totalOcc, totalAuth)
	t.Logf("WOULD ERROR on %d of %d prompts (%.0f%%); the rest log Warn", escalated, withHoles,
		100*float64(escalated)/float64(max(withHoles, 1)))
	t.Logf("escalating agents: %v", byAgent)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
