// FILE: platform/orchestration/actions/contrast_cascade_route.go
//
// Decide, from what the render audit ATTRIBUTED, whether a contrast repair
// aimed at the site theme stylesheet can actually govern the pixel — and if so,
// what the appended rule has to do to win.
//
// WHY (bugs_open/390). `write_render_audit_findings` files every contrast_failure
// with `handlerAgent: "css-patch-agent"`, which appends a rule to the site theme.
// Until the audit learned to attribute the winning declaration
// (internal/adapters/browserrunner/cascade_attribution.go) the platform had no
// way to know whether that was the right surface. [MEASURED 2026-08-25] of 40
// completed repairs sampled across 7 sites, the winning declaration was in the
// page's own <style> block in 39 and in the theme in 0, and it out-specified the
// audited selector in 33. So the repair was authored, deployed, marked complete
// and inert — 75 of 151 pairings that ever completed were filed again, 97
// re-filings with byte-identical colours.
//
// TWO RULES OF THE AUTHOR CASCADE ARE ALL THIS NEEDS, and both are recorded
// facts rather than assumptions by the time they get here:
//
//  1. An author !important declaration beats every non-important author
//     declaration, whatever its specificity or position. So an appended rule
//     must match the winner's importance to have any chance.
//  2. Between declarations of equal importance, higher specificity wins; on a
//     TIE, the later one in document order wins. The audit records whether the
//     theme's <link> follows the winner (ThemeAfterWinner), so a tie is only
//     good enough when the theme genuinely comes later — which, for a page
//     <style> block, it never does.
//
// SPECIFICITY IS COMPUTED HERE, NOT IN THE BROWSER. The probe reports the
// winning rule's selector TEXT and uses only a rough ordering to decide what to
// try first (its answer is proven by removal, not by that ordering). The
// authoritative arithmetic is cascadia's — a real CSS3 implementation, already a
// direct dependency, and unit-testable on this side. One authority, so there is
// no second implementation to drift.

package actions

import (
	"fmt"

	"github.com/andybalholm/cascadia"
)

// Repair surfaces. These are the routing decision, and the vocabulary is
// deliberately small: the question is only ever "can the file css-patch-agent
// edits govern this pixel, and do we know?".
const (
	// repairSurfaceTheme: the theme stylesheet can win, given the requirement.
	repairSurfaceTheme = "theme"
	// repairSurfaceUnreachable: no stylesheet rule can ever win, because the
	// winner is an !important declaration in the element's own style attribute.
	repairSurfaceUnreachable = "unreachable"
	// repairSurfaceUnattributed: we do not know. The probe could not prove a
	// winner, or this reply predates attribution. Behave exactly as before.
	repairSurfaceUnattributed = "unattributed"
)

// cascadeWinner mirrors the adapter's CascadeWinner over the wire.
//
// ⚠ HAND-KEPT MIRROR. platform/ must not import internal/adapters, so this is a
// copy of a struct that lives in another package and rolls in another image. A
// field added there and forgotten here decodes to its zero value and reads
// exactly like an older adapter — silence, not an error. The parity test
// (contrast_cascade_route_test.go) marshals a fully-populated adapter-shaped
// literal and asserts every field arrives, which is the only thing that turns
// that silence into a build failure.
type cascadeWinner struct {
	Property         string `json:"property"`
	Surface          string `json:"surface"`
	Selector         string `json:"selector,omitempty"`
	SheetHref        string `json:"sheet_href,omitempty"`
	Important        bool   `json:"important,omitempty"`
	ThemeAfterWinner bool   `json:"theme_after_winner,omitempty"`
	Decl             string `json:"decl,omitempty"`
	VarName          string `json:"var_name,omitempty"`
	Verified         bool   `json:"verified"`
	OpaqueSheets     int    `json:"opaque_sheets,omitempty"`
	Candidates       int    `json:"candidates,omitempty"`
}

// Surface values as the adapter spells them. Kept as constants here so a typo
// is a compile error rather than a route that silently never matches.
const (
	winnerSurfaceLinked     = "linked"
	winnerSurfaceStyleBlock = "style_block"
	winnerSurfaceInlineAttr = "inline_attr"
	winnerSurfaceUADefault  = "ua_default"
	winnerSurfaceOpaque     = "opaque"
)

// overrideRequirement is what an appended rule must satisfy to take effect. It
// is written into the finding's spec and rendered into the repair agent's
// prompt, so the agent is told what to beat instead of guessing.
type overrideRequirement struct {
	// Property is the single property to correct.
	Property string `json:"property"`
	// MinSpecificity is the winning selector's specificity as [ids, classes,
	// types]. The appended rule's selector must reach at least this.
	MinSpecificity [3]int `json:"min_specificity"`
	// MinSpecificityText is the same triple as "0,2,0" - a prompt template can
	// render a string but not an array, and a human reading a parked row should
	// not have to decode JSON.
	MinSpecificityText string `json:"min_specificity_text"`
	// StrictlyGreater is true when matching MinSpecificity is NOT enough,
	// because the winner is later in document order and would win the tie.
	StrictlyGreater bool `json:"strictly_greater"`
	// NeedsImportant is true when the winner carries !important, so nothing
	// without it can win at any specificity.
	NeedsImportant bool `json:"needs_important"`
	// Beats is the winning selector, repeated here so the requirement is
	// self-describing wherever it is read.
	Beats string `json:"beats,omitempty"`
	// Why states, in one sentence, what makes this the requirement. It is read
	// by humans on parked rows and by the repair prompt.
	Why string `json:"why"`
}

// specificityOf is the authoritative computation, and the ONLY one in the
// platform. An unparseable selector is an explicit error rather than a zero
// triple: a zero would silently become "any rule beats this", which is the most
// permissive possible requirement and would be produced by exactly the input we
// understand least.
func specificityOf(selector string) ([3]int, error) {
	if selector == "" {
		return [3]int{}, fmt.Errorf("empty selector")
	}
	sel, err := cascadia.Parse(selector)
	if err != nil {
		return [3]int{}, fmt.Errorf("parse %q: %w", selector, err)
	}
	s := sel.Specificity()
	return [3]int{s[0], s[1], s[2]}, nil
}

// contrastRepairRoute decides the repair surface and, when the theme can reach
// the defect, what an appended rule must satisfy.
//
// PURE by design — no DB, no clock, no I/O — so the whole decision table is
// unit-testable and every arm can be mutated one at a time.
//
// filedSelector is the selector the audit filed (what the repair will be written
// against). winner is the attributed declaration, or nil when there is none.
func contrastRepairRoute(winner *cascadeWinner, filedSelector string) (string, *overrideRequirement) {
	// No attribution at all: an old adapter, a capped page, or an abandoned one.
	// Behave exactly as the platform did before this existed.
	if winner == nil {
		return repairSurfaceUnattributed, nil
	}
	// UNVERIFIED IS NOT A WEAK YES. The probe proves a winner by removing it and
	// watching the computed value move; when that did not happen, every other
	// field is a guess. Routing on a guess is how a confident wrong answer gets
	// acted on, which is worse than the blindness this whole change replaces.
	if !winner.Verified {
		return repairSurfaceUnattributed, nil
	}
	switch winner.Surface {
	case winnerSurfaceOpaque, "":
		return repairSurfaceUnattributed, nil
	case winnerSurfaceInlineAttr:
		if winner.Important {
			// An !important inline style attribute is the one thing no
			// stylesheet rule can outrank. Nothing css-patch-agent can write
			// will ever apply, and pretending otherwise is what files a repair
			// that completes and does nothing.
			return repairSurfaceUnreachable, nil
		}
		// A non-important inline style beats every normal rule regardless of
		// specificity, but loses to an author !important one.
		spec, err := specificityOf(filedSelector)
		if err != nil {
			return repairSurfaceUnattributed, nil
		}
		return repairSurfaceTheme, &overrideRequirement{
			Property:           winner.Property,
			MinSpecificity:     spec,
			MinSpecificityText: specText(spec),
			StrictlyGreater:    false,
			NeedsImportant:     true,
			Beats:              "the element's own style attribute",
			Why: "the winning declaration is on the element's style attribute, which outranks " +
				"every normal rule whatever its specificity; only !important can beat it",
		}
	case winnerSurfaceUADefault:
		// Nothing in the document sets it, so any matching rule wins.
		spec, err := specificityOf(filedSelector)
		if err != nil {
			return repairSurfaceUnattributed, nil
		}
		return repairSurfaceTheme, &overrideRequirement{
			Property:           winner.Property,
			MinSpecificity:     spec,
			MinSpecificityText: specText(spec),
			StrictlyGreater:    false,
			NeedsImportant:     false,
			Why:                "no author declaration sets this property, so any matching rule takes effect",
		}
	case winnerSurfaceLinked, winnerSurfaceStyleBlock:
		spec, err := specificityOf(winner.Selector)
		if err != nil {
			// We cannot compute what must be beaten, so we do not claim to
			// know. Downgrading here is what keeps the browser's rough ordering
			// from ever becoming the contract.
			return repairSurfaceUnattributed, nil
		}
		// A tie is only good enough when the theme genuinely comes later in the
		// document. For a page <style> block it never does — assemblePage emits
		// those inside <main>, always after the <link> in <head> — but this
		// reads the recorded fact rather than assuming it, so a site that
		// orders its head differently is judged on its own evidence.
		strict := !winner.ThemeAfterWinner
		why := fmt.Sprintf("the winning declaration is %s and the theme is linked %s it, so an appended rule needs %s specificity",
			describeSurface(winner.Surface),
			map[bool]string{true: "after", false: "before"}[winner.ThemeAfterWinner],
			map[bool]string{true: "strictly higher", false: "at least equal"}[strict])
		if winner.Important {
			why += ", and !important to match the winner's importance"
		}
		return repairSurfaceTheme, &overrideRequirement{
			Property:           winner.Property,
			MinSpecificity:     spec,
			MinSpecificityText: specText(spec),
			StrictlyGreater:    strict,
			NeedsImportant:     winner.Important,
			Beats:              winner.Selector,
			Why:                why,
		}
	}
	// An unknown surface is a newer adapter than this binary. Say we do not
	// know, rather than guessing a route from a word we have never seen.
	return repairSurfaceUnattributed, nil
}

func specText(s [3]int) string { return fmt.Sprintf("%d,%d,%d", s[0], s[1], s[2]) }

func describeSurface(surface string) string {
	switch surface {
	case winnerSurfaceLinked:
		return "in a linked stylesheet"
	case winnerSurfaceStyleBlock:
		return "in a <style> block in the page"
	case winnerSurfaceInlineAttr:
		return "on the element's style attribute"
	case winnerSurfaceUADefault:
		return "a browser default"
	}
	return surface
}

// satisfiesRequirement reports whether a candidate selector would actually beat
// the winner under req. Exported to the package so the filer can offer the agent
// a worked example it has CHECKED rather than one it hopes is right.
func satisfiesRequirement(candidate string, req *overrideRequirement) bool {
	if req == nil {
		return false
	}
	spec, err := specificityOf(candidate)
	if err != nil {
		return false
	}
	cmp := compareSpecificity(spec, req.MinSpecificity)
	if req.StrictlyGreater {
		return cmp > 0
	}
	return cmp >= 0
}

// compareSpecificity returns -1, 0 or 1. CSS compares the triple
// left-to-right — ids, then classes, then types — and a single higher id count
// beats any number of classes, which is exactly the trap that made the audited
// (0,1,1) lose to a page's (0,2,0).
func compareSpecificity(a, b [3]int) int {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			if a[i] > b[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}
