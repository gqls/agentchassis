package actions

import (
	"strings"
	"testing"
)

// The fixtures below are shaped from REAL rows in component_versions /
// content_components as of 2026-07-18 (bugs_open/012). Lengths are reproduced
// faithfully because the collapse threshold is calibrated against them; the
// tails are the actual recorded tails. If a future change makes one of these
// cases flip, the calibration in component_write_guard.go is no longer true and
// the header's simulation must be re-run against live data before shipping.

// pad grows body to n bytes with filler that does not affect any check
// (no tags, no angle brackets), preserving the supplied tail.
func pad(n int, head, tail string) string {
	filler := n - len(head) - len(tail)
	if filler < 0 {
		filler = 0
	}
	return head + strings.Repeat(" ", filler) + tail
}

func TestComponentRegressionIssues(t *testing.T) {
	// The working loot-table component: 10,272 chars, closed <script>,
	// ending on </section>.
	lootGood := pad(10272, "<style>.a{}</style><section><div><script>var t=1;</script>", "</div></section>")

	cases := []struct {
		name       string
		current    string
		next       string
		wantBlock  bool
		wantReason string // substring of the expected issue, "" if none
	}{
		{
			// bugs_open/012 FINAL write: 10,272 → 1,253 chars of CSS only,
			// ending "font-weight: bold;". Caught by collapse AND mid-token.
			name:       "012 final wreck — CSS-only fragment",
			current:    lootGood,
			next:       pad(1253, "<style>.a{", "font-weight: bold;"),
			wantBlock:  true,
			wantReason: "retained",
		},
		{
			// bugs_open/012 INTERMEDIATE write: 6,765 chars = 66% retained,
			// inside the legitimate size band, but <script> left open and
			// ending mid-JS literal. This is the case a size-only guard misses.
			name:       "012 intermediate — 66% retained but script left open",
			current:    lootGood,
			next:       pad(6765, "<style>.a{}</style><section><div><script>var tiers=['Common','Rare','Epic", ""),
			wantBlock:  true,
			wantReason: "unterminated",
		},
		{
			// provocation-card v1→v2, 10,300 → 6,618 (64% retained). A real
			// rewrite that deliberately dropped its JavaScript and ends
			// cleanly. MUST NOT block — this is the case that killed the
			// earlier "lost a </script> region" check.
			name:      "legitimate — dropped JS, shrank 36%, ends cleanly",
			current:   pad(10300, "<style>.a{}</style><section><div><script>var x=1;</script>", "</div></section>"),
			next:      pad(6618, "<style>.a{}</style><section><div>", "</div></section>"),
			wantBlock: false,
		},
		{
			// tool-list v2→v3, 9,290 → 11,588 (GREW 25%) while dropping its
			// <script src=...>. Truncation cannot grow an artifact.
			name:      "legitimate — dropped JS but grew 25%",
			current:   pad(9290, "<section><div>", "<script src=\"/a.js\"></script></section>"),
			next:      pad(11588, "<section><div>", "</div></section>"),
			wantBlock: false,
		},
		{
			// tool-list v3→v4, 11,588 → 4,535 — retains only 39% yet ends
			// cleanly on "{{end}}</a></div></div></section>". This is the
			// transition that appeared AFTER the collapse floor was first set
			// at 50% and forced it down to 30%: a hard but legitimate shrink.
			// It is the reason the floor is where it is — if it starts failing,
			// re-run the simulation in the guard's header rather than nudging
			// the constant.
			name:      "legitimate — shrank to 39% but ends cleanly",
			current:   pad(11588, "<section><div>", "</div></section>"),
			next:      pad(4535, "<section><div>", "</div></section>"),
			wantBlock: false,
		},
		{
			// header-with-search_pre_037, 2,919 → 11,043. Grew 4x and does not
			// end on a tag; still a deliberate rewrite, not a truncation.
			name:      "legitimate — grew 4x, untidy tail",
			current:   pad(2919, "<section>", "</section>"),
			next:      pad(11043, "<section><div>x", "trailing text"),
			wantBlock: false,
		},
		{
			// An ordinary small edit.
			name:      "legitimate — minor edit",
			current:   lootGood,
			next:      pad(10250, "<style>.a{}</style><section><div><script>var t=2;</script>", "</div></section>"),
			wantBlock: false,
		},
		{
			// Comparative rule: a component that is ALREADY unbalanced must
			// stay repairable, otherwise the guard traps it forever.
			name:      "already-broken component stays repairable",
			current:   pad(5000, "<section><div><script>var t=1;", ""),
			next:      pad(4800, "<section><div><script>var t=1;", ""),
			wantBlock: false,
		},
		{
			// Birth path: no current template to regress against.
			name:      "empty current template is not this guard's business",
			current:   "",
			next:      pad(1253, "<style>.a{", "font-weight: bold;"),
			wantBlock: false,
		},
		{
			// Case folding — an uppercase <SCRIPT> must not slip past.
			name:      "uppercase SCRIPT left unterminated is still caught",
			current:   pad(9000, "<SECTION><DIV><SCRIPT>var t=1;</SCRIPT>", "</DIV></SECTION>"),
			next:      pad(6000, "<SECTION><DIV><SCRIPT>var t=1;", ""),
			wantBlock: true,
			// no reason asserted: both the balance and mid-token checks apply
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issues := componentRegressionIssues(tc.current, tc.next)
			gotBlock := len(issues) > 0

			if gotBlock != tc.wantBlock {
				t.Fatalf("block = %v (issues %v), want %v", gotBlock, issues, tc.wantBlock)
			}
			if tc.wantReason != "" {
				joined := strings.Join(issues, " | ")
				if !strings.Contains(joined, tc.wantReason) {
					t.Errorf("issues %q do not mention %q", joined, tc.wantReason)
				}
			}
		})
	}
}

func TestEndsCleanly(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"<div></div>", true},
		{"<div></div>\n  \n", true},
		{"font-weight: bold;", false},
		{"var tiers = ['Common', 'Epic", false},
		{"", false},
	}
	for _, c := range cases {
		if got := endsCleanly(c.in); got != c.want {
			t.Errorf("endsCleanly(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// hasUnbalancedStructuralTags is the absolute tag-imbalance signal used by the
// BIRTH gates (bugs_open/021 INSTANCE 1). It must fire on a generation cut
// mid-stream and stay quiet on a whole artifact — INCLUDING the bugs_open/046
// shape where a truncated <script> sits after a legitimately-closed </section>,
// which the old <section>-only TemplateClosed check let through.
func TestHasUnbalancedStructuralTags(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"whole: all pairs balanced", "<style>.a{}</style><section><div><script>var t=1;</script></div></section>", false},
		{"truncated <script> (046 signature)", "<section><div><script>var t=1; // cut", true},
		{"046 shape: cut <script> AFTER a closed </section>", "<section><h1>x</h1></section><script>var t=['a','b", true},
		{"012 wreck: CSS only, <style> left open", "<style>.a{font-weight: bold;", true},
		{"unterminated <div>", "<section><div><div></div></section>", true},
		{"case-insensitive: <SCRIPT> open", "<SECTION><SCRIPT>var t=1;</SECTION>", true},
		{"no structural tags at all", "just some plain text", false},
		{"empty", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasUnbalancedStructuralTags(c.in); got != c.want {
				t.Fatalf("hasUnbalancedStructuralTags(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
