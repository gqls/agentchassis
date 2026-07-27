// FILE: platform/colour/palette_audit.go
//
// Evaluates the foreground/background pairs a composed palette actually emits.
//
// WHAT THIS IS FOR
// ----------------
// features_open/026 Phase 2. Every quality check the build pipeline runs reads
// a SOURCE — a component template, a palette row, a token vocabulary. None of
// them evaluates the COMPOSITION, and on 2026-07-27 that gap put a live
// commercial site into production with card headings at 1.21:1 (near-white on
// white) and section eyebrows at 1.11:1 (near-black on near-black), across five
// pages, for three days. Every input was individually valid.
//
// This is deliberately NOT a renderer. It answers one question — given the
// slots a palette resolves to, are the pairings the layout will emit legible? —
// which is families 1 and 2 of that feature's three, deterministically and in
// microseconds. Family 3 (a component hard-coding an ink over a themed fill)
// is invisible here by construction and needs the rendered check; see
// scripts/render_audit.py.
//
// THE PAIR LIST IS THE CONTRACT, AND IT IS DELIBERATELY SHORT
// ----------------------------------------------------------
// Only pairings that a stylesheet demonstrably emits are listed. A speculative
// pair produces a finding nobody can act on, and this platform already has more
// detection than consumption (bugs_open/115: three correct findings sat unread
// for three days). Each entry below is one that has actually failed in
// production, or is the direct inverse of one that has.
package colour

import "sort"

// PaletteSlots is a composed palette: slot name → hex. Slot names are the
// renderer's, not CSS variable names — "card_bg", not "--color-card-bg".
type PaletteSlots map[string]string

// Failure is one pairing that does not meet AA.
type Failure struct {
	Pair  Pair
	Ratio float64
	Need  float64
}

// auditPairs are the pairings a stylesheet emits, each with the AA level that
// applies to the text that lands there.
//
// Provenance for every entry, so the next author can tell a measured pair from
// a guessed one — all measured on fundamentallyai.com, 2026-07-27:
//
//	text/background      5.97:1 — passed throughout; the control that proves
//	                     a green result here is not vacuous
//	text/card_bg         1.21:1 — the owner's own example, "Every decision
//	                     leaves a record", invisible on 5 of 10 pages
//	text_muted/card_bg   3.23:1 — same cause, body copy
//	text_muted/background 5.97:1 — passed; the pair that made the failure
//	                     confusing, because the section text was fine
//	primary/background   1.11:1 — near-black eyebrow on near-black section
//	primary_text/primary 2.32:1 AFTER the palette was repaired: primary
//	                     flipped near-black → light blue and every ink over it
//	                     inverted. Repairing a palette can break this pair, so
//	                     it is checked in both directions.
//	accent/background    the CTA/highlight ink, same shape as primary
var auditPairs = []struct {
	role     string
	fg, bg   string
	large    bool
	optional bool // skip silently when the slot is absent
}{
	{role: "body text on the page background", fg: "text", bg: "background"},
	{role: "muted text on the page background", fg: "text_muted", bg: "background"},
	{role: "card title on a card", fg: "text", bg: "card_bg", optional: true},
	{role: "card body on a card", fg: "text_muted", bg: "card_bg", optional: true},
	{role: "text on a raised surface", fg: "text", bg: "surface", optional: true},
	{role: "primary used as an ink on the page background", fg: "primary", bg: "background", large: true},
	{role: "accent used as an ink on the page background", fg: "accent", bg: "background", large: true},
	{role: "the ink on a primary-filled control", fg: "primary_text", bg: "primary", optional: true},
	{role: "the ink on a CTA band", fg: "cta_text", bg: "cta_bg", optional: true},
}

// AuditPalette returns every pairing that fails AA, worst first.
//
// A slot that is absent is skipped when the pair is marked optional and
// reported as a failure otherwise — an absent `background` or `text` is not a
// passing palette, it is an unusable one, and returning "no failures" for it
// would be the vacuous green this whole exercise exists to prevent.
//
// A slot whose value cannot be parsed is likewise a failure, not a skip. In
// practice that means a `var(...)` reference or a gradient reached a slot that
// must hold a literal, which is worth a human look.
func AuditPalette(p PaletteSlots) []Failure {
	var out []Failure
	for _, ap := range auditPairs {
		fg, okFG := p[ap.fg]
		bg, okBG := p[ap.bg]
		if (!okFG || fg == "") || (!okBG || bg == "") {
			if ap.optional {
				continue
			}
			out = append(out, Failure{
				Pair: Pair{Role: ap.role, ForegroundSlot: ap.fg, BackgroundSlot: ap.bg,
					Foreground: fg, Background: bg, Large: ap.large},
				Ratio: 0, Need: 0,
			})
			continue
		}
		pair := Pair{Role: ap.role, ForegroundSlot: ap.fg, BackgroundSlot: ap.bg,
			Foreground: fg, Background: bg, Large: ap.large}
		ratio, ok, err := pair.Check()
		if err != nil {
			out = append(out, Failure{Pair: pair, Ratio: 0, Need: pair.Threshold()})
			continue
		}
		if !ok {
			out = append(out, Failure{Pair: pair, Ratio: ratio, Need: pair.Threshold()})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ratio < out[j].Ratio })
	return out
}
