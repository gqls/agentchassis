package actions

import (
	"encoding/json"
	"testing"
)

// winner builds an attributed declaration. Defaults are the vonc.com case:
// a verified page <style> winner, `.gauntlet-cta-section .gauntlet-btn-primary`
// = (0,2,0), theme linked before it.
func winner(mut func(*cascadeWinner)) *cascadeWinner {
	w := &cascadeWinner{
		Property:         "color",
		Surface:          winnerSurfaceStyleBlock,
		Selector:         ".gauntlet-cta-section .gauntlet-btn-primary",
		Decl:             "var(--color-primary, #1a1a2e)",
		VarName:          "--color-primary",
		Verified:         true,
		ThemeAfterWinner: false,
		Candidates:       3,
	}
	if mut != nil {
		mut(w)
	}
	return w
}

const filedSelector = "A.gauntlet-btn-primary" // (0,1,1)

// TestContrastRepairRouteTable is the whole decision, one row at a time.
//
// MUTATION PROOFS (each reverted afterwards, file diffed byte-identical):
//   - strict -> non-strict for a style_block winner: "page block, theme before"
//     stops requiring strictly-greater and this test goes RED.
//   - drop the !important parity assignment: "winner is important" RED.
//   - treat an important inline attribute as reachable: "important inline
//     attribute" RED.
//   - remove the Verified early return: "unverified is not a weak yes" RED.
func TestContrastRepairRouteTable(t *testing.T) {
	cases := []struct {
		name        string
		w           *cascadeWinner
		wantSurface string
		wantSpec    [3]int
		wantStrict  bool
		wantImp     bool
		wantNilReq  bool
	}{
		{
			// The measured majority. The theme is earlier in the document, so a
			// tie loses on source order and only strictly-greater wins.
			name: "page block, theme before", w: winner(nil),
			wantSurface: repairSurfaceTheme, wantSpec: [3]int{0, 2, 0}, wantStrict: true,
		},
		{
			// A site that links its theme after the block: a tie now wins, so
			// the requirement relaxes. Reads the recorded fact rather than
			// assuming the emission order.
			name:        "page block, theme after",
			w:           winner(func(w *cascadeWinner) { w.ThemeAfterWinner = true }),
			wantSurface: repairSurfaceTheme, wantSpec: [3]int{0, 2, 0}, wantStrict: false,
		},
		{
			name:        "winner is important",
			w:           winner(func(w *cascadeWinner) { w.Important = true }),
			wantSurface: repairSurfaceTheme, wantSpec: [3]int{0, 2, 0}, wantStrict: true, wantImp: true,
		},
		{
			// No stylesheet rule can outrank it, at any specificity, ever.
			name: "important inline attribute",
			w: winner(func(w *cascadeWinner) {
				w.Surface, w.Selector, w.Important = winnerSurfaceInlineAttr, "", true
			}),
			wantSurface: repairSurfaceUnreachable, wantNilReq: true,
		},
		{
			// Beatable, but only with !important - specificity is irrelevant
			// against a style attribute.
			name: "plain inline attribute",
			w: winner(func(w *cascadeWinner) {
				w.Surface, w.Selector = winnerSurfaceInlineAttr, ""
			}),
			wantSurface: repairSurfaceTheme, wantSpec: [3]int{0, 1, 1}, wantImp: true,
		},
		{
			name: "linked stylesheet winner",
			w: winner(func(w *cascadeWinner) {
				w.Surface, w.SheetHref = winnerSurfaceLinked, "/assets/css/other.css"
			}),
			wantSurface: repairSurfaceTheme, wantSpec: [3]int{0, 2, 0}, wantStrict: true,
		},
		{
			name: "nothing sets it",
			w: winner(func(w *cascadeWinner) {
				w.Surface, w.Selector = winnerSurfaceUADefault, ""
			}),
			wantSurface: repairSurfaceTheme, wantSpec: [3]int{0, 1, 1},
		},
		{
			// UNVERIFIED IS NOT A WEAK YES. Every other field is a guess.
			name:        "unverified is not a weak yes",
			w:           winner(func(w *cascadeWinner) { w.Verified = false }),
			wantSurface: repairSurfaceUnattributed, wantNilReq: true,
		},
		{
			name: "blinded by a cross-origin sheet",
			w: winner(func(w *cascadeWinner) {
				w.Surface, w.OpaqueSheets = winnerSurfaceOpaque, 2
			}),
			wantSurface: repairSurfaceUnattributed, wantNilReq: true,
		},
		{
			// A newer adapter than this binary. Say we do not know rather than
			// guessing a route from a word we have never seen.
			name:        "surface this binary has never heard of",
			w:           winner(func(w *cascadeWinner) { w.Surface = "shadow_dom_part" }),
			wantSurface: repairSurfaceUnattributed, wantNilReq: true,
		},
		{
			// The browser reported a selector cascadia cannot parse. We cannot
			// compute what must be beaten, so we must not claim to know - a zero
			// triple would silently mean "any rule beats this".
			name:        "unparseable winning selector",
			w:           winner(func(w *cascadeWinner) { w.Selector = ".a >>> .b" }),
			wantSurface: repairSurfaceUnattributed, wantNilReq: true,
		},
		{
			name: "no attribution at all", w: nil,
			wantSurface: repairSurfaceUnattributed, wantNilReq: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			surface, req := contrastRepairRoute(tc.w, filedSelector)
			if surface != tc.wantSurface {
				t.Fatalf("surface = %q, want %q", surface, tc.wantSurface)
			}
			if tc.wantNilReq {
				if req != nil {
					t.Fatalf("requirement = %+v, want nil: a route we cannot justify must carry no instruction", req)
				}
				return
			}
			if req == nil {
				t.Fatal("requirement is nil, but the surface is routable")
			}
			if req.MinSpecificity != tc.wantSpec {
				t.Errorf("min specificity = %v, want %v", req.MinSpecificity, tc.wantSpec)
			}
			if req.StrictlyGreater != tc.wantStrict {
				t.Errorf("strictly_greater = %v, want %v — a tie %s win here",
					req.StrictlyGreater, tc.wantStrict,
					map[bool]string{true: "must NOT", false: "must"}[tc.wantStrict])
			}
			if req.NeedsImportant != tc.wantImp {
				t.Errorf("needs_important = %v, want %v", req.NeedsImportant, tc.wantImp)
			}
			if req.Why == "" {
				t.Error("requirement carries no reason; it is read by humans on parked rows")
			}
		})
	}
}

// TestTheMeasuredCaseIsRefusedByTheRuleTheAgentWasTold is the bug, expressed as
// arithmetic: what migration 318's prompt instructed ("repeat the offending
// selector exactly") does NOT satisfy the requirement, and what the platform now
// computes does.
func TestTheMeasuredCaseIsRefusedByTheRuleTheAgentWasTold(t *testing.T) {
	_, req := contrastRepairRoute(winner(nil), filedSelector)
	if req == nil {
		t.Fatal("expected a requirement for the worked case")
	}
	// This is what css-patch-agent actually appended on vonc.com, and it is
	// what the prompt told it to write.
	if satisfiesRequirement(filedSelector, req) {
		t.Fatalf("%q satisfies the requirement, but on the live site it lost to %q — "+
			"if this passes, the arithmetic no longer describes the bug", filedSelector, req.Beats)
	}
	// Repeating the winner's own selector is a TIE, and a tie loses on source
	// order because the theme is linked first. This is the case a specificity-
	// only rule gets wrong.
	if satisfiesRequirement(req.Beats, req) {
		t.Error("an equal-specificity selector satisfies a strictly-greater requirement")
	}
	// Out-specifying by one does win.
	if !satisfiesRequirement("body .gauntlet-cta-section .gauntlet-btn-primary", req) {
		t.Error("a strictly higher specificity does not satisfy the requirement")
	}
}

// TestSpecificityIsComparedLeftToRight pins the comparison the bug turns on: a
// single class beats any number of type selectors, which is why an audited
// (0,1,1) lost to a page's (0,2,0).
func TestSpecificityIsComparedLeftToRight(t *testing.T) {
	cases := []struct {
		a, b [3]int
		want int
	}{
		{[3]int{0, 1, 1}, [3]int{0, 2, 0}, -1}, // the measured case
		{[3]int{0, 2, 0}, [3]int{0, 1, 1}, 1},
		{[3]int{1, 0, 0}, [3]int{0, 9, 9}, 1}, // one id beats nine classes
		{[3]int{0, 2, 0}, [3]int{0, 2, 0}, 0},
	}
	for _, c := range cases {
		if got := compareSpecificity(c.a, c.b); got != c.want {
			t.Errorf("compare(%v,%v) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

// TestSpecificityOfRefusesRatherThanReturningZero. A zero triple is the most
// PERMISSIVE possible requirement ("any rule beats this"), and it would be
// produced by exactly the input we understand least. It must be an error.
func TestSpecificityOfRefusesRatherThanReturningZero(t *testing.T) {
	for _, bad := range []string{"", ".a >>> .b", "["} {
		if _, err := specificityOf(bad); err == nil {
			t.Errorf("specificityOf(%q) returned no error; a zero triple would silently "+
				"mean every rule beats it", bad)
		}
	}
}

// TestCascadeWinnerMirrorParity is the tripwire for the hand-kept mirror. The
// adapter's CascadeWinner lives in another package and rolls in another image,
// so a field added there and forgotten here decodes to its zero value and reads
// exactly like an older adapter — silence, not an error.
//
// The literal below is the adapter's JSON with EVERY field non-zero and
// distinguishable from its zero value. If a field stops arriving, this fails.
func TestCascadeWinnerMirrorParity(t *testing.T) {
	const adapterJSON = `{
	  "property":"color",
	  "surface":"style_block",
	  "selector":".s .e",
	  "sheet_href":"/assets/css/styles.css",
	  "important":true,
	  "theme_after_winner":true,
	  "decl":"var(--x, #fff)",
	  "var_name":"--x",
	  "verified":true,
	  "opaque_sheets":2,
	  "candidates":3
	}`
	var w cascadeWinner
	if err := json.Unmarshal([]byte(adapterJSON), &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	checks := map[string]bool{
		"property":           w.Property == "color",
		"surface":            w.Surface == winnerSurfaceStyleBlock,
		"selector":           w.Selector == ".s .e",
		"sheet_href":         w.SheetHref == "/assets/css/styles.css",
		"important":          w.Important,
		"theme_after_winner": w.ThemeAfterWinner,
		"decl":               w.Decl == "var(--x, #fff)",
		"var_name":           w.VarName == "--x",
		"verified":           w.Verified,
		"opaque_sheets":      w.OpaqueSheets == 2,
		"candidates":         w.Candidates == 3,
	}
	for field, ok := range checks {
		if !ok {
			t.Errorf("field %q did not arrive through the mirror — the adapter sends it and "+
				"this struct drops it silently, which reads as an older adapter", field)
		}
	}
	// And the reverse direction: a field this side declares must survive a
	// round trip, or the spec written from it is not what the adapter meant.
	out, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back cascadeWinner
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if back != w {
		t.Errorf("round trip changed the winner:\n got %+v\nwant %+v", back, w)
	}
}
