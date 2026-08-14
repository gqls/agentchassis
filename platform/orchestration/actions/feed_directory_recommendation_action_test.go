// FILE: platform/orchestration/actions/feed_directory_recommendation_action_test.go
//
// Phase B (2026-08-13): the vertical→directory-kind matcher. Pure-function
// table tests — the DB write half reuses the news action's proven
// supersede-then-insert shape verbatim and is not re-tested here.

package actions

import (
	"testing"

	"go.uber.org/zap"
)

func TestMatchVerticalDirectory(t *testing.T) {
	cases := []struct {
		name     string
		industry string
		siteType string
		category string
		domain   string
		wantKind string // "" = no recommendation (nil or Recommended:false)
	}{
		{"exact mortgage industry", "mortgage", "", "", "example.co.uk", "mortgage-lender"},
		{"exact savings", "savings", "", "", "example.co.uk", "savings-provider"},
		{"banking aliases to savings-provider", "banking", "", "", "example.co.uk", "savings-provider"},
		{"exact health insurance", "health insurance", "", "", "example.co.uk", "health-insurer"},
		{"partial: mortgage broker industry", "mortgage brokerage", "", "", "example.co.uk", "mortgage-lender"},

		// The longest-key determinism rule this matcher adds over
		// matchVerticalNews: a signal containing BOTH "health insurance" and
		// "insurance" must always take the longer, more specific key. (Both
		// resolve to health-insurer today, but the rule is what's pinned —
		// when more insurer kinds land, the specific one must keep winning.)
		{"longest partial key wins", "private health insurance advice", "", "", "example.co.uk", "health-insurer"},

		// "finance" alone is an explicit NOT-recommended entry, not a miss:
		// too generic to choose a provider class.
		{"bare finance is refused", "finance", "", "", "example.co.uk", ""},

		// A signal containing both "insurance" (recommends) and "finance"
		// (refuses): "insurance" is longer, so the recommendation wins —
		// deterministically, not by map iteration luck.
		{"insurance beats finance by length", "insurance finance", "", "", "example.co.uk", "health-insurer"},

		{"domain-derived signal", "", "", "", "bestmortgagedeals.co.uk", "mortgage-lender"},
		{"no signal at all", "", "", "", "", ""},
		{"unrelated vertical", "veterinary", "clinic", "", "vetexample.co.uk", ""},
	}

	logger := zap.NewNop()
	for _, c := range cases {
		got := matchVerticalDirectory(c.industry, c.siteType, c.category, c.domain, logger)
		gotKind := ""
		if got != nil && got.Recommended {
			gotKind = got.Kind
		}
		if gotKind != c.wantKind {
			t.Errorf("%s: matchVerticalDirectory(%q,%q,%q,%q) kind = %q, want %q",
				c.name, c.industry, c.siteType, c.category, c.domain, gotKind, c.wantKind)
		}
	}
}

// The non-price ruling's mechanical half (council round-1 objection: a policy
// with no enforcement mechanism is a preference): a finance kind's field
// vocabulary is CLOSED at registration, so a price-shaped fact is
// structurally unregistrable however it was produced. Non-finance kinds are
// untouched — no closed vocabulary.
func TestRefuseDisallowedDirectoryField(t *testing.T) {
	cases := []struct {
		kind, field string
		wantRefused bool
	}{
		{"mortgage-lender", "regulator_status", false},
		{"mortgage-lender", "fca_firm_reference", false},
		{"mortgage-lender", "representative_apr", true}, // the exact exposure the ruling names
		{"mortgage-lender", "typical_rate", true},
		{"savings-provider", "protection_scheme", false},
		{"savings-provider", "aer", true},
		{"health-insurer", "underwriter", false},
		{"health-insurer", "monthly_premium", true},
		{"model", "price_input_per_mtok", false}, // open vocabulary: prices ALLOWED for models
		{"company", "roi_pct", false},            // open vocabulary
		{"unheard-of-kind", "anything", false},   // no vocabulary registered = open
	}
	for _, c := range cases {
		got := refuseDisallowedDirectoryField(c.kind, c.field)
		if (got != "") != c.wantRefused {
			t.Errorf("refuseDisallowedDirectoryField(%q, %q) refused=%v, want %v (detail %q)",
				c.kind, c.field, got != "", c.wantRefused, got)
		}
	}
}

// Every finance kind with a closed vocabulary must also be a real, renderable
// kind — an allowlist for a kind that doesn't exist would be dead policy.
func TestFinanceAllowlistKindsAreRenderable(t *testing.T) {
	for kind := range financeKindFieldAllowlist {
		if _, ok := directoryPublishProfiles[kind]; !ok {
			t.Errorf("financeKindFieldAllowlist names kind %q, which has no directoryPublishProfiles entry", kind)
		}
	}
}

// Every recommending entry must name a kind that actually exists in
// directoryPublishProfiles — a recommendation for an unrenderable kind would
// plan a page nothing can ever populate. This is the lockstep guard between
// the two tables, in the same spirit as the planner-vocabulary caveat in
// SEED_directory_components.sql.
func TestVerticalDirectoryMapKindsAreRenderable(t *testing.T) {
	for signal, cfg := range verticalDirectoryMap {
		if !cfg.Recommended {
			continue
		}
		if cfg.Kind == "" || cfg.SpecKey == "" {
			t.Errorf("entry %q recommends but is missing Kind/SpecKey", signal)
			continue
		}
		if _, ok := directoryPublishProfiles[cfg.Kind]; !ok {
			t.Errorf("entry %q recommends kind %q, which has no directoryPublishProfiles entry — unrenderable", signal, cfg.Kind)
		}
	}
}
