// FILE: platform/orchestration/actions/discovery_checks/default_brand_prompt_test.go
//
// bugs_open/210, OWNER RULING 2026-08-09. The default must ALWAYS produce a
// usable, site-specific prompt, because it is what stands between a site with no
// planned logo prompt and either (a) an item that can never be handled, or (b)
// IMG-069's refusal. At ~2,000 domains the degraded shapes are not edge cases —
// a site with nothing but a domain row is the COMMON case early in its life, so
// that is the case tested hardest.

package discovery_checks

import (
	"strings"
	"testing"
)

// The load-bearing property. A builder that could return "" would hand the
// caller straight back to the generic-fallback refusal this lane exists to keep
// unreachable — so "never empty" is not a nicety, it is the contract.
func TestDefaultBrandPrompt_NeverEmptyForARealSite(t *testing.T) {
	cases := []struct {
		name  string
		facts siteBrandFacts
	}{
		{"nothing but a domain", siteBrandFacts{Domain: "robot-hands.com"}},
		{"domain and name only", siteBrandFacts{Domain: "cookly.uk", Name: "Cookly"}},
		{"a .co.uk with no name", siteBrandFacts{Domain: "mortgagecalculator.co.uk"}},
		{"completely empty", siteBrandFacts{}},
		{"fully populated", siteBrandFacts{
			Domain: "vetcomparison.uk", Name: "Vet Comparison", Industry: "Veterinary price comparison",
			Tagline: "Compare UK vet prices", Audience: "UK pet owners", Tone: "Plain, trustworthy"}},
	}

	for _, c := range cases {
		for _, purpose := range []string{"logo", "hero"} {
			t.Run(c.name+"/"+purpose, func(t *testing.T) {
				got := composeBrandImagePrompt(c.facts, purpose)
				if strings.TrimSpace(got) == "" {
					t.Fatal("empty prompt — the caller would fall through to the generic-fallback refusal")
				}
				if len(got) < 40 {
					t.Errorf("prompt is too thin to generate from (%d chars): %q", len(got), got)
				}
			})
		}
	}
}

// The owner's ruling names three inputs — mission, target market, domain
// character. This pins that each actually reaches the prompt, so a future
// refactor cannot quietly drop one and still pass "not empty".
func TestDefaultBrandPrompt_CarriesTheOwnersThreeInputs(t *testing.T) {
	f := siteBrandFacts{
		Domain:   "vetcomparison.uk",
		Name:     "Vet Comparison",
		Industry: "Veterinary price comparison",
		Tagline:  "Compare UK vet prices honestly",
		Audience: "UK pet owners facing unexpected vet bills",
		Tone:     "Plain-spoken and trustworthy",
	}
	got := composeBrandImagePrompt(f, "logo")

	for _, want := range []string{
		"Vet Comparison",                            // who it is
		"vetcomparison.uk",                          // domain character
		"Veterinary price comparison",               // sector / mission
		"Compare UK vet prices honestly",            // positioning
		"UK pet owners facing unexpected vet bills", // target market
		"Plain-spoken and trustworthy",              // character
	} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt does not carry %q:\n%s", want, got)
		}
	}
}

// The contamination lesson, pinned. Logos must not acquire photographic or
// imagery-style direction — that exclusion is deliberate
// (imagery_style_guide.go: "logos get nothing"). This builder reads brand
// identity only, and the logo branch must actively say so.
func TestDefaultBrandPrompt_LogoAsksForAFlatMarkAndNoLettering(t *testing.T) {
	got := composeBrandImagePrompt(siteBrandFacts{Domain: "oufe.com", Name: "Oufe"}, "logo")

	for _, want := range []string{"Flat vector", "favicon size", "no lettering", "no photographic"} {
		if !strings.Contains(got, want) {
			t.Errorf("logo prompt is missing the %q constraint:\n%s", want, got)
		}
	}
	// A hero legitimately may be photographic; a logo may not. If this ever
	// fails, someone has unified the two branches and reintroduced the 2026-05-20
	// contamination.
	if strings.Contains(got, "Photographic or illustrative") {
		t.Error("the logo branch picked up the hero's photographic clause")
	}
}

func TestDefaultBrandPrompt_DomainCharacter(t *testing.T) {
	cases := map[string]string{
		"robot-hands.com":            "robot hands",
		"mortgagecalculator.co.uk":   "mortgagecalculator",
		"finetuning.uk":              "finetuning",
		"ai-agent-orchestration.com": "ai agent orchestration",
		"vonc.com":                   "vonc",
		"":                           "",
	}
	for in, want := range cases {
		if got := domainCharacter(in); got != want {
			t.Errorf("domainCharacter(%q) = %q, want %q", in, got, want)
		}
	}
}

// A long free-text identity field must not crowd out the craft instructions —
// some sites carry hundreds of characters of positioning prose in these keys.
func TestDefaultBrandPrompt_LongClausesAreTrimmed(t *testing.T) {
	long := strings.Repeat("a strategic positioning statement that goes on and on ", 12)
	got := composeBrandImagePrompt(siteBrandFacts{Domain: "x.com", Tagline: long}, "logo")

	if !strings.Contains(got, "Flat vector") {
		t.Fatal("craft constraints were lost behind the long clause")
	}
	if len(got) > 900 {
		t.Errorf("prompt ran to %d chars; the trim is not holding", len(got))
	}
}
