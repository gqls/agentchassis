// FILE: platform/orchestration/datahelpers/runtime_fill_test.go
//
// The element-scoped runtime-fill exemption (bugs_open/137).
//
// The load-bearing cases are the two DIRECTIONS, and a test suite that only
// proves one of them would have passed against the defect this replaces:
//
//	over-exemption  — a shell must NOT exempt its neighbours (the defect);
//	under-exemption — a shell MUST still exempt its own contents (the reason
//	                  the exemption exists at all).
//
// Both are asserted on every consumer, because "fix the over-exemption" and
// "break the exemption" produce identical results on a test that only checks the
// first.

package datahelpers

import (
	"strings"
	"testing"
)

// ── the two directions, on the raw span scanner ────────────────────────────

func TestRuntimeFillSpansCoverOnlyTheMarkedSubtree(t *testing.T) {
	html := `<section class="a"><a href="#">outside</a></section>` +
		`<section data-runtime-fill="true"><a href="#">inside</a></section>` +
		`<section class="b"><a href="#">after</a></section>`

	spans := RuntimeFillSpans(html)
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 span, got %d: %+v", len(spans), spans)
	}

	inside := strings.Index(html, `>inside<`)
	before := strings.Index(html, `>outside<`)
	after := strings.Index(html, `>after<`)

	if !spans.Contains(inside) {
		t.Errorf("the shell's own anchor is NOT covered — the exemption would stop working")
	}
	if spans.Contains(before) || spans.Contains(after) {
		t.Errorf("a neighbour is covered — this is the whole-document bug, unfixed")
	}
}

func TestRuntimeFillSpansNestedElementsCloseCorrectly(t *testing.T) {
	// A same-named descendant inside the shell must not close the shell early —
	// depth counting, not first-match. Without it the span ends at the inner
	// </div> and the trailing anchor stops being exempt.
	html := `<div data-runtime-fill><div class="inner"><a href="#">deep</a></div>` +
		`<a href="#">still inside</a></div><a href="#">out</a>`

	spans := RuntimeFillSpans(html)
	if !spans.Contains(strings.Index(html, `>deep<`)) {
		t.Errorf("nested descendant not covered")
	}
	if !spans.Contains(strings.Index(html, `>still inside<`)) {
		t.Errorf("shell closed early at the inner </div> — depth counting is broken")
	}
	if spans.Contains(strings.Index(html, `>out<`)) {
		t.Errorf("span leaked past the shell's own close tag")
	}
}

func TestRuntimeFillMarkerMustBeAnAttributeNotASubstring(t *testing.T) {
	cases := []struct {
		name   string
		html   string
		marked bool
	}{
		{"real attribute", `<div data-runtime-fill><a href="#">x</a></div>`, true},
		{"real attribute with value", `<div data-runtime-fill="true"><a href="#">x</a></div>`, true},
		{"longer attribute name", `<div data-runtime-filler="1"><a href="#">x</a></div>`, false},
		{"prefixed attribute name", `<div x-data-runtime-fill="1"><a href="#">x</a></div>`, false},
		{"inside a script string", `<script>var s="data-runtime-fill";</script><a href="#">x</a>`, false},
		{"in text content", `<p>we use data-runtime-fill here</p><a href="#">x</a>`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			off := strings.Index(tc.html, `>x<`)
			got := RuntimeFillSpans(tc.html).Contains(off)
			if got != tc.marked {
				t.Errorf("anchor exempt=%v, want %v — a substring test would say true for every case here", got, tc.marked)
			}
		})
	}
}

func TestRuntimeFillSpansSurviveAwkwardMarkup(t *testing.T) {
	// A '>' inside a quoted attribute value must not end the tag early. With a
	// naive IndexByte('>') the shell's start tag "ends" mid-attribute and every
	// subsequent span is misplaced.
	html := `<section title="a > b" data-runtime-fill><a href="#">in</a></section><a href="#">out</a>`
	spans := RuntimeFillSpans(html)
	if !spans.Contains(strings.Index(html, `>in<`)) {
		t.Errorf("quoted '>' broke the start-tag scan")
	}
	if spans.Contains(strings.Index(html, `>out<`)) {
		t.Errorf("span leaked past the shell")
	}
}

func TestUnclosedShellDegradesToWholeDocumentNotToNothing(t *testing.T) {
	// Malformed markup must never make the exemption NARROWER than it was — that
	// would turn a parser limitation into a wave of false dead-control findings.
	html := `<a href="#">before</a><section data-runtime-fill><a href="#">after</a>`
	spans := RuntimeFillSpans(html)
	if !spans.Contains(strings.Index(html, `>after<`)) {
		t.Errorf("unclosed shell must still exempt what follows it")
	}
	if spans.Contains(strings.Index(html, `>before<`)) {
		t.Errorf("an unclosed shell must not reach BACKWARDS")
	}
}

func TestSpanSetContainsIsSortedAndMerged(t *testing.T) {
	html := `<div data-runtime-fill><div data-runtime-fill><a href="#">x</a></div></div>`
	spans := RuntimeFillSpans(html)
	if len(spans) != 1 {
		t.Fatalf("nested shells should merge to one span, got %d: %+v", len(spans), spans)
	}
	for i := 1; i < len(spans); i++ {
		if spans[i].Start < spans[i-1].Start {
			t.Errorf("spans not sorted — Contains' early return would be wrong")
		}
	}
}

// ── the consumer that judges control liveness ──────────────────────────────

func TestDeadControlAnchorsOutsideRuntimeFill(t *testing.T) {
	t.Run("a shell does not exempt its neighbours", func(t *testing.T) {
		// The measured live shape: vonc.com/index, where lobby-grid's marker
		// suppressed gauntlet-cta's controls on the assembled page.
		html := `<section data-runtime-fill><a href="#" class="tpl">Template</a></section>` +
			`<section class="cta"><a href="#">Enter the Gauntlet</a></section>`
		dead := DeadControlAnchorsOutsideRuntimeFill(html)
		if len(dead) != 1 {
			t.Fatalf("expected exactly the neighbour's control, got %d: %+v", len(dead), dead)
		}
		if dead[0].Text != "Enter the Gauntlet" {
			t.Errorf("wrong anchor reported: %+v", dead[0])
		}
	})

	t.Run("a shell still exempts its own contents", func(t *testing.T) {
		html := `<section data-runtime-fill><a href="#">Template</a></section>`
		if dead := DeadControlAnchorsOutsideRuntimeFill(html); len(dead) != 0 {
			t.Fatalf("the exemption is gone: %+v", dead)
		}
	})

	t.Run("no shell behaves exactly like the unscoped helper", func(t *testing.T) {
		html := `<section><a href="#">One</a><a href="/real.html">Two</a></section>`
		scoped := DeadControlAnchorsOutsideRuntimeFill(html)
		plain := DeadControlAnchors(html)
		if len(scoped) != len(plain) {
			t.Fatalf("diverged on shell-free markup: %d vs %d", len(scoped), len(plain))
		}
		for i := range scoped {
			if scoped[i] != plain[i] {
				t.Errorf("anchor %d differs: %+v vs %+v", i, scoped[i], plain[i])
			}
		}
	})
}

// ── the consumer that must stay byte-identical ─────────────────────────────

func TestRepairPageLinksExemptsPerAnchorNotPerDocument(t *testing.T) {
	index := NewPageURLIndex([]string{"/real.html"})

	t.Run("a shell no longer suppresses the whole page", func(t *testing.T) {
		html := `<section data-runtime-fill><a href="">shell placeholder</a></section>` +
			`<section><a href="">Enter the Gauntlet</a></section>`
		out, repairs := RepairPageLinks(html, index)
		if len(repairs) != 1 {
			t.Fatalf("expected the neighbour's empty href to be repaired, got %d: %+v", len(repairs), repairs)
		}
		if !strings.Contains(out, `<a href="">shell placeholder</a>`) {
			t.Errorf("the shell's own placeholder was repaired — the exemption is broken:\n%s", out)
		}
		if strings.Contains(out, `<a href="">Enter the Gauntlet</a>`) {
			t.Errorf("the neighbour was left unrepaired — this is the bug, unfixed:\n%s", out)
		}
	})

	t.Run("a section whose root is a shell is untouched, byte for byte", func(t *testing.T) {
		html := `<section data-runtime-fill><a href="">a</a><a href="/ghost.html">b</a></section>`
		out, repairs := RepairPageLinks(html, index)
		if len(repairs) != 0 || out != html {
			t.Fatalf("per-section callers must see no change: %d repairs\n%s", len(repairs), out)
		}
	})

	t.Run("clean markup is returned byte-identical", func(t *testing.T) {
		html := `<section data-runtime-fill><a href="">x</a></section><a href="/real.html">y</a>`
		out, repairs := RepairPageLinks(html, index)
		if len(repairs) != 0 || out != html {
			t.Fatalf("byte-identity broken on a page needing no repair:\n%q", out)
		}
	})
}

// ── the whole-input predicate is still available, and still means that ─────

func TestHasRuntimeFillMarkerIsStillWholeInput(t *testing.T) {
	// The "is this SECTION a shell?" question keeps its whole-input answer;
	// check_empty_sections and friends depend on it. Pinned so a later tidy-up
	// does not quietly redirect them at the element-scoped predicate, which asks
	// a different question.
	html := `<section class="a"></section><section data-runtime-fill></section>`
	if !HasRuntimeFillMarker(html) {
		t.Errorf("whole-input predicate changed meaning")
	}
}
