// FILE: platform/orchestration/actions/discovery_checks/check_cta_rank_anomaly_test.go
//
// The predicate is tested over ALREADY-RANKED input because the ranking it
// consumes is datahelpers.RankCTAPositionalCandidates — the writers' own,
// tested where it lives — and this check must add no ordering opinion of its
// own. Each negative case here is a REAL estate state, named, so a threshold
// edit that un-silences one of them has to argue with the state it re-flags.

package discovery_checks

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func rankedTools(navOrders ...int) []datahelpers.CTAPositionalCandidate {
	names := []string{"tool-a", "tool-b", "tool-c", "tool-d"}
	out := make([]datahelpers.CTAPositionalCandidate, 0, len(navOrders))
	for i, n := range navOrders {
		out = append(out, datahelpers.CTAPositionalCandidate{
			Name: names[i], URL: "/tools/" + names[i] + ".html", Area: "tools", NavOrder: n,
		})
	}
	return out
}

func TestCTARankAnomalyPredicate(t *testing.T) {
	cases := []struct {
		name string
		in   []datahelpers.CTAPositionalCandidate
		want bool
	}{
		// 391's fossil: nav_order 1 set at page creation 2026-03-13, pack at
		// the default. The case this check exists for.
		{"the 391 fossil fires", rankedTools(1, 100, 100), true},
		// The same sites after the owner's demotion (1 -> 900): pack leads.
		{"the demoted state is silent", rankedTools(100, 100, 900), false},
		// webdesign's shape: every tool at the default, winner is alphabetical.
		// Arbitrary but NOT anomalous — candidate 3's business, out of scope.
		{"all-default is silent", rankedTools(100, 100, 100, 100), false},
		// A deliberate curated ordering must not be second-guessed.
		{"a curated ladder is silent", rankedTools(10, 20, 30), false},
		// Two curated leaders: unique minimum but no fossil-sized lead.
		{"two close leaders are silent", rankedTools(1, 2, 100), false},
		// "Anomalous against its siblings" needs siblings.
		{"two candidates are too few to judge", rankedTools(1, 100), false},
		{"an empty ranking is silent", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, detail := ctaRankAnomaly(tc.in)
			if got != tc.want {
				t.Errorf("ctaRankAnomaly(%v) = %v (%s), want %v", tc.in, got, detail, tc.want)
			}
			if detail == "" {
				t.Error("detail must always be populated — it is the Resolved reason and the work item body")
			}
		})
	}
}

// TestCTARankAnomalySilencedByTheLever pins the pairing the owner approved:
// setting eligible_as_cta_target=false on the fossil removes it from the
// RANKING (the lever), and the alarm then observes a healthy rank-1 — correct
// silencing, because the flagged page can no longer be the primary button.
// The ranking is the real one, not a re-implementation, so this test also
// fails if the check ever grows its own candidate filtering.
func TestCTARankAnomalySilencedByTheLever(t *testing.T) {
	supply := []datahelpers.CTAPositionalCandidate{
		{Name: "tool-fossil", URL: "/tools/fossil.html", Area: "tools", NavOrder: 1},
		{Name: "tool-b", URL: "/tools/b.html", Area: "tools", NavOrder: 100},
		{Name: "tool-c", URL: "/tools/c.html", Area: "tools", NavOrder: 100},
	}
	if got, _ := ctaRankAnomaly(datahelpers.RankCTAPositionalCandidates("", supply)); !got {
		t.Fatal("fixture is inert: the fossil must fire before the lever is pulled")
	}
	supply[0].IneligibleAsCTATarget = true
	anomalous, detail := ctaRankAnomaly(datahelpers.RankCTAPositionalCandidates("", supply))
	if anomalous {
		t.Errorf("opting the fossil out must silence the alarm (the condition is gone, "+
			"not blinded): %s", detail)
	}
	if !strings.Contains(detail, "candidate") {
		t.Errorf("detail should describe the healthy observation, got %q", detail)
	}
}
