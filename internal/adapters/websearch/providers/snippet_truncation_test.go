package providers

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// The two failure shapes below are not invented: both were measured in
// production over 30 days of content_feed_items on 2026-09-03 (5,863 rows;
// 943 exactly 200 bytes long, 288 carrying an unclosed "](url" tail, 2
// carrying U+FFFD). bugs_open/332 is the visitor-facing symptom.

func TestTruncateSnippetLeavesShortInputAlone(t *testing.T) {
	for _, s := range []string{"", "short", strings.Repeat("a", snippetMaxBytes)} {
		if got := truncateSnippet(s, snippetMaxBytes); got != s {
			t.Errorf("input of %d bytes was modified: %q -> %q", len(s), s, got)
		}
	}
}

func TestTruncateSnippetNeverSplitsARune(t *testing.T) {
	// A cut that lands mid-rune is what put U+FFFD into 2 live rows. Sweep
	// every boundary so this cannot pass by luck of one alignment.
	base := strings.Repeat("é", 400) // 2 bytes per rune
	for max := 4; max <= 300; max++ {
		got := truncateSnippet(base, max)
		if !utf8.ValidString(got) {
			t.Fatalf("max=%d produced invalid UTF-8: %q", max, got)
		}
		if strings.ContainsRune(got, '�') {
			t.Fatalf("max=%d produced U+FFFD", max)
		}
		if len(got) > max {
			t.Fatalf("max=%d produced %d bytes", max, len(got))
		}
	}
}

// The mutation guard for the rune half: the OLD implementation must fail this
// test. If someone reverts truncateSnippet to a byte slice, this is the test
// that goes red.
func TestByteSliceTruncationWouldSplitARune(t *testing.T) {
	base := strings.Repeat("é", 400)
	old := base[:197] + "..." // the implementation this replaced
	if utf8.ValidString(old) {
		t.Fatal("the byte-slice cut produced valid UTF-8 — this test no longer " +
			"proves the rune-safety fix does anything; pick a boundary that splits")
	}
}

func TestTruncateSnippetDropsAHalfWrittenLink(t *testing.T) {
	// Cut lands inside the URL: the exact 288-row shape. The downstream
	// literal-markdown strip cannot match this (its pattern needs the closing
	// paren), so the raw URL would reach visitors as article text.
	s := "Rangers beat Celtic in a match report by " +
		"[our correspondent](https://www.example.com/sport/football/2026/09/03/" +
		strings.Repeat("a-very-long-slug-segment/", 8) + "index.html) " +
		"and more prose follows."

	// The fixture only tests what it claims if the cut lands INSIDE the URL.
	// Asserting that first: an earlier version of this test was 194 bytes long,
	// truncated nothing, and passed its own premise by never exercising it.
	if len(s) <= snippetMaxBytes {
		t.Fatalf("fixture is %d bytes, under the %d budget — it would not truncate",
			len(s), snippetMaxBytes)
	}
	if closeParen := strings.IndexByte(s, ')'); closeParen < snippetMaxBytes {
		t.Fatalf("fixture's link closes at byte %d, before the cut — the half-written "+
			"case is not exercised", closeParen)
	}

	got := truncateSnippet(s, snippetMaxBytes)

	if strings.Contains(got, "](") {
		t.Errorf("kept a link opening with no closing paren: %q", got)
	}
	if strings.Contains(got, "https://") {
		t.Errorf("leaked a bare URL into the summary: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("lost the ellipsis: %q", got)
	}
	if !strings.Contains(got, "Rangers beat Celtic") {
		t.Errorf("threw away the prose before the link: %q", got)
	}
}

func TestTruncateSnippetKeepsACompleteLink(t *testing.T) {
	// A link that survives the cut intact is left alone — the downstream strip
	// handles it, and trimming it here would be the contested lossy transform.
	s := "See [the report](https://ex.com/a) for detail. " +
		strings.Repeat("padding words to exceed the budget. ", 10)
	got := truncateSnippet(s, snippetMaxBytes)
	if !strings.Contains(got, "[the report](https://ex.com/a)") {
		t.Errorf("dropped a complete link: %q", got)
	}
}

func TestTruncateSnippetKeepsABareBracket(t *testing.T) {
	// A footnote marker carries no URL. Trimming on a bare "[" would discard
	// real prose for nothing — the over-trim this implementation avoids.
	s := "The regulator said [1] that advertisers must substantiate claims. " +
		strings.Repeat("more prose to exceed the budget here. ", 10)
	got := truncateSnippet(s, snippetMaxBytes)
	if !strings.Contains(got, "[1]") {
		t.Errorf("trimmed a bare bracket that carried no link: %q", got)
	}
}

func TestTruncateSnippetOutputIsAlwaysWithinBudget(t *testing.T) {
	// Including the ellipsis — the old code returned 200 bytes for a 197-byte
	// cut plus "...", which is why 943 rows are exactly 200 long.
	inputs := []string{
		strings.Repeat("word ", 200),
		"[a](" + strings.Repeat("u", 500) + ")",
		strings.Repeat("é", 300) + "[x](https://e.com/" + strings.Repeat("p", 100),
	}
	for _, in := range inputs {
		got := truncateSnippet(in, snippetMaxBytes)
		if len(got) > snippetMaxBytes {
			t.Errorf("output %d bytes exceeds budget %d for input of %d bytes",
				len(got), snippetMaxBytes, len(in))
		}
	}
}

// Closes the council's one advisory objection (editquality, round 1, c93e71a6):
// the first version short-circuited on `max < 4` and returned the input
// UNMODIFIED even when it exceeded the budget, so a tiny max produced an
// unbounded result. Harmless while snippetMaxBytes was the only caller, and
// exactly the kind of latent dependency on a constant that outlives the reason
// for it. A budget the function may ignore is not a budget.
func TestTruncateSnippetHonoursEvenATinyBudget(t *testing.T) {
	long := strings.Repeat("word ", 100)
	for max := 1; max <= 6; max++ {
		got := truncateSnippet(long, max)
		if len(got) > max {
			t.Errorf("max=%d returned %d bytes — the budget was ignored", max, len(got))
		}
		if !utf8.ValidString(got) {
			t.Errorf("max=%d produced invalid UTF-8: %q", max, got)
		}
	}
	// And the multi-byte case at the same tiny budgets: cutting inside a
	// 3-byte rune is the failure this must not reintroduce.
	multi := strings.Repeat("→", 50) // 3 bytes per rune
	for max := 1; max <= 6; max++ {
		got := truncateSnippet(multi, max)
		if len(got) > max || !utf8.ValidString(got) || strings.ContainsRune(got, '�') {
			t.Errorf("max=%d on 3-byte runes returned %q (%d bytes)", max, got, len(got))
		}
	}
}
