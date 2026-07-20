// Tests for SafeCut and the rune-safety of TruncateString (bugs_open/027 §4b).
// The defect class: a raw byte slice at a fixed cap splits a multi-byte rune
// landing on the boundary and emits invalid UTF-8 — into logs via
// TruncateString, and into an image-generation prompt via the direction cut.

package datahelpers

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSafeCutIsRuneSafe(t *testing.T) {
	// A curly quote (3 bytes) straddling the boundary must not be split.
	s := strings.Repeat("a", 199) + "’tail"
	got := SafeCut(s, 200)
	if !utf8.ValidString(got) {
		t.Errorf("SafeCut produced invalid UTF-8: %q", got)
	}
	if len(got) > 200 {
		t.Errorf("SafeCut exceeded its byte budget: %d", len(got))
	}
}

func TestSafeCutEdges(t *testing.T) {
	if got := SafeCut("abc", 0); got != "" {
		t.Errorf("n=0 must return empty, got %q", got)
	}
	if got := SafeCut("abc", -1); got != "" {
		t.Errorf("n<0 must return empty, got %q", got)
	}
	if got := SafeCut("abc", 3); got != "abc" {
		t.Errorf("n=len(s) must return s unchanged, got %q", got)
	}
	if got := SafeCut("abc", 10); got != "abc" {
		t.Errorf("n>len(s) must return s unchanged, got %q", got)
	}
	// ASCII boundary: exact cut, no back-off.
	if got := SafeCut("abcdef", 4); got != "abcd" {
		t.Errorf("ASCII cut wrong: %q", got)
	}
}

func TestTruncateStringIsRuneSafe(t *testing.T) {
	s := strings.Repeat("a", 99) + "éx" // é (2 bytes) straddles a 100-byte cap
	got := TruncateString(s, 100)
	if !utf8.ValidString(got) {
		t.Errorf("TruncateString produced invalid UTF-8: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("TruncateString must keep its ellipsis contract: %q", got)
	}
	// Under the cap: unchanged, no ellipsis.
	if got := TruncateString("short", 100); got != "short" {
		t.Errorf("under-cap string must be untouched: %q", got)
	}
}
