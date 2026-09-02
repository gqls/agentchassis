// FILE: platform/orchestration/datahelpers/utf8_safety_test.go
//
// bugs_open/423. Postgres refuses invalid UTF-8, so a Go string that cuts a
// multi-byte rune does not degrade quietly — it kills the statement that tries
// to persist it, and (before this bug's half 1) the failure was reported as
// nothing at all. These tests pin the two primitives that make the cut
// impossible at source and the failure legible when it happens anyway.
//
// Every test here is written to FAIL if the primitive is reverted to the byte
// idiom it replaces; the comment above each says which mutation it kills.
package datahelpers

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// MUTATION KILLED: UpperFirst reverted to `strings.ToUpper(s[:1]) + s[1:]`.
// That idiom hands ToUpper a lone lead byte, which decodes as U+FFFD, and then
// re-attaches the orphaned continuation bytes — so the em-dash below comes back
// as `ef bf bd 80 94` and the first invalid byte is 0x80, which is verbatim the
// byte Postgres named in the live capture.
func TestUpperFirstDoesNotSplitAMultiByteRune(t *testing.T) {
	for _, s := range []string{"—", "—dash", "“quote”", "école", "…", "£40"} {
		got := UpperFirst(s)
		if !utf8.ValidString(got) {
			t.Errorf("UpperFirst(%q) produced invalid UTF-8: % x", s, got)
		}
	}
	if got := UpperFirst("école"); got != "École" {
		t.Errorf("UpperFirst must upper-case the first RUNE: got %q, want %q", got, "École")
	}
	// A rune with no upper case must come back untouched, not mangled.
	if got := UpperFirst("—dash"); got != "—dash" {
		t.Errorf("UpperFirst(%q) altered a caseless leading rune: %q", "—dash", got)
	}
}

// The parity guarantee that made converting eight call sites in one pass safe:
// on ASCII, UpperFirst is byte-identical to the idiom it replaced, so no
// existing casing behaviour moved.
//
// MUTATION KILLED: an UpperFirst that upper-cases the whole string, or that
// lower-cases, or that drops the first rune.
func TestUpperFirstIsByteIdenticalToTheOldIdiomOnASCII(t *testing.T) {
	for _, s := range []string{"boxing", "Fight Calendar", "a", "guides-index", ""} {
		want := s
		if len(s) > 0 {
			want = strings.ToUpper(s[:1]) + s[1:]
		}
		if got := UpperFirst(s); got != want {
			t.Errorf("ASCII parity broken for %q: got %q, want %q", s, got, want)
		}
	}
}

// The live case, reproduced end to end at the shape that actually bit us
// (render_site_components_action.go's footer services column): a page title
// containing a STANDALONE em-dash, which strings.Fields makes its own word, so
// every byte of the word is the first byte of a multi-byte rune.
//
// MUTATION KILLED: reverting the buildServicesHTML loop to the byte idiom. This
// is the one test that fails on the exact production input.
func TestFooterServicesLabelCasingStaysValidUTF8(t *testing.T) {
	// boxingonline.com, pages.title for tool-boxing-trivia-quiz (2026-09-02).
	const label = "Boxing Quiz — Test Your Knowledge | Tools"

	words := strings.Fields(strings.ReplaceAll(label, "-", " "))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = UpperFirst(w)
		}
	}
	got := strings.Join(words, " ")

	if !utf8.ValidString(got) {
		off, window, _ := InvalidUTF8At(got)
		t.Fatalf("footer services label is not valid UTF-8 at byte %d (near %s): % x", off, window, got)
	}
	if !strings.Contains(got, "—") {
		t.Errorf("the em-dash was destroyed rather than preserved: %q", got)
	}
}

// MUTATION KILLED: an InvalidUTF8At that reports the wrong offset, or that
// returns found=true for clean input (which would refuse every healthy render
// at the store seam), or that returns a window still containing the raw bad
// bytes — a diagnostic that cannot itself be persisted is the defect this bug
// already hit once, in the failure-REPORTING path.
func TestInvalidUTF8AtNamesTheOffsetAndStaysPrintable(t *testing.T) {
	if _, _, found := InvalidUTF8At("<footer>Perfectly — fine £40</footer>"); found {
		t.Error("clean UTF-8 reported as invalid: the store seam would refuse every healthy render")
	}

	prefix := "<footer><ul><li>"
	bad := prefix + "\x80\x94 Test</li></ul></footer>"
	off, window, found := InvalidUTF8At(bad)
	if !found {
		t.Fatal("a bare continuation byte was not detected")
	}
	if off != len(prefix) {
		t.Errorf("wrong offset: got %d, want %d", off, len(prefix))
	}
	if !utf8.ValidString(window) {
		t.Errorf("the report window is not valid UTF-8, so the report cannot be stored: %q", window)
	}
	for i := 0; i < len(window); i++ {
		if window[i] > 0x7f {
			t.Fatalf("the report window is not pure ASCII at byte %d: %q", i, window)
		}
	}
	if !strings.Contains(window, `\x80`) {
		t.Errorf("the window must show the offending byte escaped, got %q", window)
	}
}
