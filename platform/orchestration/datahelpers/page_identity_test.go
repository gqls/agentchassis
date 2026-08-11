package datahelpers

import "testing"

// TestPageItemStem_InvertsCanonicalisePagePrefixes pins the PREMISE of the stem
// key: that the prefixes it strips are exactly the prefixes CanonicalisePage
// adds. The two used to live in different packages under a hand-sync comment
// (bugs_open/215). If the canonicaliser ever adds a fourth prefixed role, or
// renames one, this test fails rather than leaving the stem key quietly blind
// to a whole family — which would present as "the reconciler stopped noticing
// twins", with no error anywhere.
func TestPageItemStem_InvertsCanonicalisePagePrefixes(t *testing.T) {
	for _, role := range []string{"tool", "guide", "game"} {
		const bare = "widget-sizer"
		name, _, _ := CanonicalisePage(PageDescriptor{Role: role, Slug: bare})
		if name != role+"-"+bare {
			t.Fatalf("premise changed: CanonicalisePage(%q, %q) produced name %q, expected %q\n"+
				"if the prefix convention moved, PageItemStem's prefix list must move with it",
				role, bare, name, role+"-"+bare)
		}
		if got := PageItemStem(name); got != bare {
			t.Fatalf("PageItemStem(%q) = %q, want %q — the stem key no longer inverts the canonicaliser for role %q",
				name, got, bare, role)
		}
		// And the bare form must be a fixed point, or a bare page and its
		// prefixed twin would not share a stem at all.
		if got := PageItemStem(bare); got != bare {
			t.Fatalf("PageItemStem(%q) = %q, want it unchanged", bare, got)
		}
	}
}

// TestPageItemStem_DoesNotPairTwoPrefixedNames is the guard that makes the
// weakest key safe to use at all. "tool-pricing" beside "guide-pricing" is the
// on-record false positive the reconciler's Pass C2 has refused to risk since
// 2026-07-20: two genuinely different pages on one topic. They DO share a stem
// — that is the point of this test — so the stem key alone can never be the
// decision, and the reconciler's bare-vs-prefixed guard is what makes the pair
// unmatchable. If someone deletes that guard, this test still passes; the
// reconciler-side test is the one that fails. This test exists so the reader of
// the key knows the hazard is real, not theoretical.
func TestPageItemStem_DoesNotPairTwoPrefixedNames(t *testing.T) {
	a, b := PageItemStem("tool-pricing"), PageItemStem("guide-pricing")
	if a != b {
		t.Fatalf("premise changed: %q and %q no longer share a stem (%q vs %q).\n"+
			"The bare-vs-prefixed guard at the reconciler was written because they DO — re-check it is still needed.",
			"tool-pricing", "guide-pricing", a, b)
	}
	if a == "tool-pricing" || a == "guide-pricing" {
		t.Fatalf("PageItemStem stripped nothing: got %q", a)
	}
}

func TestPagePathKey(t *testing.T) {
	cases := []struct{ url, want string }{
		// The flat/nested divergence this key exists for: the legacy tool-deploy
		// arm's shape and the canonicaliser's shape claim one path.
		{"/tools/llm-cost-calculator.html", "/tools/llm-cost-calculator"},
		{"/tools/llm-cost-calculator/index.html", "/tools/llm-cost-calculator"},
		// The homepage is a path, not an absence of one.
		{"/index.html", "/"},
		// Section index vs flat section page.
		{"/news/index.html", "/news"},
		{"/news.html", "/news"},
		// Neither suffix: passes through, so it can only match an exact twin.
		{"/tools/", "/tools/"},
		{"", "/"},
	}
	for _, c := range cases {
		if got := PagePathKey(c.url); got != c.want {
			t.Errorf("PagePathKey(%q) = %q, want %q", c.url, got, c.want)
		}
	}
}

// TestPageCanonicalNameForRow_SkipsTheHomepageCollapseFamily pins the one skip
// that stops this key manufacturing false identities. CanonicalisePage collapses
// role "index" to the homepage whatever the slug says — that is a merge of two
// writer conventions, not a spelling correction, so it must never be read as
// evidence that two stored rows are the same page.
func TestPageCanonicalNameForRow_SkipsTheHomepageCollapseFamily(t *testing.T) {
	if got := PageCanonicalNameForRow("anything-at-all", "index"); got != "" {
		t.Fatalf("PageCanonicalNameForRow(_, \"index\") = %q, want \"\" — "+
			"an index-typed row must contribute no name signal, or every one of them "+
			"collides with the real homepage", got)
	}
	// A row whose identity cannot be canonicalised at all also contributes
	// nothing, rather than grouping with every other empty.
	if got := PageCanonicalNameForRow("", "tool"); got != "" {
		t.Fatalf("PageCanonicalNameForRow(\"\", \"tool\") = %q, want \"\"", got)
	}
	// The signal itself: a bare tool-typed row names its canonical twin.
	if got := PageCanonicalNameForRow("llm-cost-calculator", "tool"); got != "tool-llm-cost-calculator" {
		t.Fatalf("PageCanonicalNameForRow(\"llm-cost-calculator\", \"tool\") = %q, want %q",
			got, "tool-llm-cost-calculator")
	}
	// And a row already canonical is a fixed point, so the key never claims a
	// page collides with itself under another spelling.
	if got := PageCanonicalNameForRow("tool-llm-cost-calculator", "tool"); got != "tool-llm-cost-calculator" {
		t.Fatalf("canonical row not a fixed point: got %q", got)
	}
}
