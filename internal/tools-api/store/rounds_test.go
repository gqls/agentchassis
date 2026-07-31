package store

import (
	"strings"
	"testing"
)

// The slug is the public handle of a published round, so three properties matter
// and none of them is checked by the compiler: it must be the length the URL and
// the share card were designed around, it must only contain characters the
// validator accepts, and it must not be predictable.

func TestNewSlugShapeAndAlphabet(t *testing.T) {
	for i := 0; i < 200; i++ {
		s, err := newSlug()
		if err != nil {
			t.Fatalf("newSlug: %v", err)
		}
		if len(s) != SlugLength {
			t.Fatalf("slug %q has length %d, want %d", s, len(s), SlugLength)
		}
		for _, r := range s {
			if !strings.ContainsRune(SlugAlphabet, r) {
				t.Fatalf("slug %q contains %q, which is outside SlugAlphabet", s, r)
			}
		}
	}
}

// The generator and the validator were briefly two definitions (a regexp in the
// handler spelling out the same characters). Drift between them fails SILENTLY
// and totally: every newly published round 404s on read. This asserts the round
// trip so the two cannot be separated again without a red test.
func TestNewSlugIsAlwaysAcceptedByValidSlug(t *testing.T) {
	for i := 0; i < 500; i++ {
		s, err := newSlug()
		if err != nil {
			t.Fatalf("newSlug: %v", err)
		}
		if !ValidSlug(s) {
			t.Fatalf("ValidSlug rejected a slug newSlug produced: %q", s)
		}
	}
}

func TestValidSlugRejectsWhatItShould(t *testing.T) {
	cases := map[string]string{
		"":                      "empty",
		"abcdefghi":             "one short",
		"abcdefghijk":           "one long",
		"abcdefgh1j":            "contains 1, excluded to prevent mistyping",
		"abcdefghOj":            "contains O, excluded to prevent mistyping",
		"abcdefgh-j":            "hyphen is not in the alphabet",
		"ABCDEFGHJK":            "upper case is not in the alphabet",
		"abcdefgh j":            "space",
		"../../etc/pa":          "path traversal shape",
		"abcdefgh%2":            "percent-encoding shape",
		"39595461-245e-493e-84": "a uuid fragment, which is what the old mock URL used",
	}
	for in, why := range cases {
		if ValidSlug(in) {
			t.Errorf("ValidSlug(%q) = true, want false (%s)", in, why)
		}
	}
}

// A 32-character alphabet is load-bearing, not cosmetic: 256 % 32 == 0, so
// reducing a random byte with % is unbiased. If someone adds a character, the
// generator quietly starts favouring the first few — a weakness invisible in any
// individual slug.
func TestSlugAlphabetDivides256(t *testing.T) {
	if len(SlugAlphabet) != 32 {
		t.Fatalf("SlugAlphabet has %d characters; 32 is required for unbiased "+
			"reduction of a random byte (256 %% 32 == 0)", len(SlugAlphabet))
	}
	seen := map[rune]bool{}
	for _, r := range SlugAlphabet {
		if seen[r] {
			t.Errorf("SlugAlphabet repeats %q, which skews the distribution", r)
		}
		seen[r] = true
	}
	for _, r := range "01lIO" {
		if strings.ContainsRune(SlugAlphabet, r) {
			t.Errorf("SlugAlphabet contains %q, which is easily mistyped from a card", r)
		}
	}
}

// Not a statistical test — just a guard against the catastrophic case where the
// generator returns a constant (e.g. a fallback path swallowing a rand failure).
func TestNewSlugDoesNotRepeat(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		s, err := newSlug()
		if err != nil {
			t.Fatalf("newSlug: %v", err)
		}
		if seen[s] {
			t.Fatalf("newSlug returned %q twice in 1000 draws — not random", s)
		}
		seen[s] = true
	}
}
