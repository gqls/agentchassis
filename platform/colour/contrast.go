// FILE: platform/colour/contrast.go
//
// The WCAG colour maths, in a package both `actions` and
// `actions/discovery_checks` can import.
//
// WHY IT LIVES HERE AND NOT IN actions/color_util.go
// --------------------------------------------------
// `actions` imports `discovery_checks` (actions/discovery_checks.go — the
// blank import that triggers every check's init()). The dependency cannot run
// the other way, so a discovery check that needs a contrast ratio has exactly
// two options: import a package neither of them owns, or keep a second copy of
// the formulas.
//
// A second copy is not a hypothetical cost here. Two implementations of one
// rule, with nothing checking they agree, is the drift class this platform has
// been unpicking all week — bugs_open/109 (four hand-maintained maps of one
// struct), bugs_open/113 (a palette and a layout that each fill the same slot),
// bugs_closed/072 (a stylesheet and a spec that disagree). A contrast checker
// that quietly disagrees with the renderer's own dark/light classification
// would be worse than no checker, because it would file confident findings
// about pages that are fine and stay silent on pages that are not.
//
// So: the formulas live once, here. actions/color_util.go keeps its
// unexported names as thin wrappers so no call site has to change, and the
// package-level tests in both places measure the same code.
//
// All functions take CSS hex strings (#rgb, #rrggbb, #rrggbbaa). Alpha is
// parsed and discarded — WCAG contrast is defined on composited colours, and a
// caller that has a translucent ink must composite it against its own
// background first. Returning a ratio for a colour that is 30% transparent
// would be a confident wrong answer, which is the failure mode this package
// exists to avoid.
package colour

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// WCAG 2.1 AA thresholds. "Large" is 18pt / 14pt bold or above.
const (
	AANormal = 4.5
	AALarge  = 3.0
	// AAAnormal is not used by any current caller and is here so a future one
	// does not reinvent it as a literal.
	AAANormal = 7.0
)

// ParseHex handles #rgb, #rrggbb and #rrggbbaa. Alpha is ignored; see the
// package comment for why it is not silently composited.
func ParseHex(hex string) (r, g, b uint8, err error) {
	s := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	switch len(s) {
	case 3:
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	case 6:
		// already in the wanted form
	case 8:
		s = s[:6] // drop alpha
	default:
		return 0, 0, 0, fmt.Errorf("invalid hex colour %q: want #rgb, #rrggbb or #rrggbbaa", hex)
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex colour %q: %w", hex, err)
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v), nil
}

// sRGBToLinear applies the WCAG transfer function to one 8-bit channel.
func sRGBToLinear(c uint8) float64 {
	f := float64(c) / 255.0
	if f <= 0.03928 {
		return f / 12.92
	}
	return math.Pow((f+0.055)/1.055, 2.4)
}

// RelativeLuminance returns WCAG relative luminance (0..1).
func RelativeLuminance(r, g, b uint8) float64 {
	return 0.2126*sRGBToLinear(r) + 0.7152*sRGBToLinear(g) + 0.0722*sRGBToLinear(b)
}

// ContrastRatio returns the WCAG contrast ratio between two hex colours:
// 1.0 for identical colours, 21.0 for black on white. Order does not matter.
func ContrastRatio(hex1, hex2 string) (float64, error) {
	r1, g1, b1, err := ParseHex(hex1)
	if err != nil {
		return 0, fmt.Errorf("parse colour %q: %w", hex1, err)
	}
	r2, g2, b2, err := ParseHex(hex2)
	if err != nil {
		return 0, fmt.Errorf("parse colour %q: %w", hex2, err)
	}
	l1 := RelativeLuminance(r1, g1, b1)
	l2 := RelativeLuminance(r2, g2, b2)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05), nil
}

// IsDark reports whether white contrasts better than black on this colour —
// i.e. "does this background need light text?". This is the classification the
// CSS renderer uses, and it is deliberately contrast-based rather than a
// luminance threshold so that it can never disagree with ContrastRatio about
// which ink wins. The crossover sits near #777777.
//
// An unparseable colour returns false: callers treat that as "assume light",
// which is the conservative answer because it leaves dark text in place.
func IsDark(hex string) bool {
	white, err := ContrastRatio("#ffffff", hex)
	if err != nil {
		return false
	}
	black, err := ContrastRatio("#000000", hex)
	if err != nil {
		return false
	}
	return white > black
}

// IsPerceptuallyDark is the stricter luminance<0.2 test. It is NOT a synonym
// for IsDark and the two disagree in the mid greys by design: IsDark answers
// "which ink wins", this answers "is this colour dark in itself". Callers
// choosing between them should prefer IsDark for anything about legibility.
func IsPerceptuallyDark(hex string) bool {
	r, g, b, err := ParseHex(hex)
	if err != nil {
		return false
	}
	return RelativeLuminance(r, g, b) < 0.2
}

// Pair is one foreground/background combination a stylesheet actually emits,
// named so a finding can say which slot pairing failed rather than quoting two
// hex values with no provenance.
type Pair struct {
	Role           string // e.g. "card title", human-facing
	ForegroundSlot string // e.g. "text"
	BackgroundSlot string // e.g. "card_bg"
	Foreground     string
	Background     string
	Large          bool // large text, so AALarge applies
}

// Threshold returns the AA ratio this pair must meet.
func (p Pair) Threshold() float64 {
	if p.Large {
		return AALarge
	}
	return AANormal
}

// Check evaluates one pair. A pair whose colours cannot be parsed is reported
// as an error rather than as a pass — an unreadable value is exactly the case
// where a silent "fine" is most expensive.
func (p Pair) Check() (ratio float64, ok bool, err error) {
	ratio, err = ContrastRatio(p.Foreground, p.Background)
	if err != nil {
		return 0, false, err
	}
	return ratio, ratio >= p.Threshold(), nil
}

// ============================================================================
// Hue-preserving legibility — HSL lightness search
// ============================================================================
//
// WHY THIS EXISTS, and it is a correction to a shipped mechanism rather than a
// new capability.
//
// `legibleInkFor` (actions/palette_specialised_slots.go) answers "can this
// colour be READ on these grounds", and when the answer is no it substitutes
// the first palette colour that can. Its documented intent was that the
// substitution "prefers a palette colour so the site keeps its character".
//
// MEASURED 2026-08-13, at the served artefact, across all 18 palette-driven
// live sites: it never keeps any character. The walk is
// {text, accent, text_muted, secondary, primary} and `text` is first — and
// `text` is BY CONSTRUCTION the slot chosen to be legible on `background`, so
// it clears the grounds whenever any candidate does. All 16 divergences between
// an ink companion and its source slot resolved to that site's own
// --color-text. Zero exceptions. The other four candidates are unreachable in
// production, and --color-<x>-ink was --color-text under another name.
//
// That mattered because repointing a component onto the ink companion was
// therefore equivalent to writing `color: var(--color-text)`, and the fleet's
// only contrast instrument (scripts/render_audit.py) scores that a CLEAN PASS —
// near-black text on a pale ground has excellent contrast. A change that
// stripped a brand colour from every affected element would have graded green.
// See bugs_open/122 contribution 2026-08-13, and the LANDMINES entry of the
// same date.
//
// THE REPAIR: before substituting a different colour, try to make the source
// colour itself legible by moving ONLY its HSL lightness. Hue and saturation
// are preserved exactly, so #1A1F2E becomes a lighter navy rather than becoming
// body text, and the smallest sufficient change wins.
//
// WHY LIGHTNESS AND NOT A BLEND TOWARD WHITE/BLACK. Blending desaturates: it
// pulls the colour toward grey on the way, so a brand colour rescued by
// blending arrives visibly washed out. An HSL lightness move keeps S intact, so
// what comes back reads as the same colour at a different level. That is the
// whole point of the exercise — a legible colour nobody recognises is the
// failure this function exists to stop.

// rgbToHSL converts 8-bit sRGB to HSL with h in [0,360) and s,l in [0,1].
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l // achromatic: hue is undefined, saturation is zero
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case rf:
		h = (gf - bf) / d
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/d + 2
	default:
		h = (rf-gf)/d + 4
	}
	return h * 60, s, l
}

// hslToRGB is the inverse of rgbToHSL. h is taken modulo 360.
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	h = math.Mod(math.Mod(h, 360)+360, 360) / 360
	if s == 0 {
		v := uint8(math.Round(l * 255))
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	channel := func(t float64) uint8 {
		if t < 0 {
			t++
		}
		if t > 1 {
			t--
		}
		switch {
		case t < 1.0/6.0:
			return uint8(math.Round((p + (q-p)*6*t) * 255))
		case t < 1.0/2.0:
			return uint8(math.Round(q * 255))
		case t < 2.0/3.0:
			return uint8(math.Round((p + (q-p)*(2.0/3.0-t)*6) * 255))
		default:
			return uint8(math.Round(p * 255))
		}
	}
	return channel(h + 1.0/3.0), channel(h), channel(h - 1.0/3.0)
}

// ToHex formats an 8-bit sRGB triple as #rrggbb, lower case.
func ToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// CompositeOverGround returns `overlay` at `alpha` painted over `ground`, as an
// opaque hex — the colour a browser actually shows where a translucent overlay
// sits on a ground.
//
// WHY A CONTRAST DERIVATION NEEDS THIS. This package refuses to composite alpha
// for the caller (see the package header: returning a ratio for a 30%
// transparent colour would be a confident wrong answer). That is the right rule
// for ContrastRatio, and it leaves a gap: a caller that KNOWS its overlay must
// be able to composite deliberately and then measure the result. The renderer
// emits exactly one such overlay — `--section-surface: rgba(255,255,255,0.05)`
// from buildSectionDefaults — and an ink certified against the DECLARED surface
// while the visitor sees the COMPOSITED one is certified against the wrong
// colour.
//
// MEASURED 2026-08-14: that 5% white overlay costs 0.62 of contrast ratio on the
// fleet's dark palettes (robot-hands, dartsonline, vonc), against a first-hit
// search headroom of 0.02–0.09. So the gap was ~10x the margin, and inks that
// cleared 4.5 on the declared surface measured 3.93–4.03 on the composited one —
// a fresh contrast_failure for an element the repair had just fixed. Raised by
// the bugfix_122_contrast_ink_slots lane, whose objection this closes.
func CompositeOverGround(overlay string, alpha float64, ground string) string {
	or, og, ob, err := ParseHex(overlay)
	if err != nil {
		return ground
	}
	gr, gg, gb, err := ParseHex(ground)
	if err != nil {
		return ground
	}
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	mix := func(o, g uint8) uint8 {
		return uint8(math.Round(float64(g)*(1-alpha) + float64(o)*alpha))
	}
	return ToHex(mix(or, gr), mix(og, gg), mix(ob, gb))
}

// lightnessSearchStep is the HSL lightness increment the search walks. 0.01 is
// ~2.55 levels per 8-bit channel, so every reachable rounded value is visited
// without testing the same hex twice more than a couple of times. Coarser steps
// overshoot and return a colour further from the brand than necessary; finer
// steps cost iterations and cannot find anything new after rounding.
const lightnessSearchStep = 0.01

// LegibleVariant returns srcHex moved in HSL LIGHTNESS ONLY — hue and
// saturation preserved exactly — until it clears minRatio against EVERY ground,
// and returns the variant needing the SMALLEST lightness change. ok is false
// when no lightness whatsoever makes srcHex legible on all grounds, which is
// the caller's signal to fall back to substituting a different colour.
//
// GROUNDS IS A SLICE, and that is load-bearing for the same reason it is in
// legibleInkFor: one variable serving two grounds is only right if it is right
// for the WORSE of them. A component may place one ink on the page and on a
// card. Narrowing this to a single ground produces a value that looks correct
// on whichever surface the reviewer happens to open.
//
// An empty or wholly unmeasurable grounds slice returns ok=false rather than a
// vacuous pass: a colour certified against nothing has not been checked, and
// reporting success there is how an untested value reaches a stylesheet looking
// tested.
func LegibleVariant(srcHex string, grounds []string, minRatio float64) (hex string, ok bool) {
	r, g, b, err := ParseHex(srcHex)
	if err != nil {
		return "", false
	}
	h, s, l0 := rgbToHSL(r, g, b)

	// A colour with zero saturation has no hue to preserve, so there is nothing
	// this function can offer that the caller's achromatic fallback does not
	// already do better. Declining keeps the two mechanisms from both claiming
	// the same case.
	if s == 0 {
		return "", false
	}

	clearsAll := func(candidate string) bool {
		measured := false
		for _, ground := range grounds {
			if ground == "" {
				continue
			}
			ratio, err := ContrastRatio(candidate, ground)
			if err != nil || ratio < minRatio {
				return false
			}
			measured = true
		}
		return measured
	}

	// Walk outward from the source lightness, testing both directions at each
	// distance, so the FIRST hit is by construction the smallest change. Trying
	// one direction to exhaustion first would return a heavily darkened colour
	// when a slight lightening would have done.
	for delta := lightnessSearchStep; delta <= 1.0; delta += lightnessSearchStep {
		for _, l := range []float64{l0 + delta, l0 - delta} {
			if l < 0 || l > 1 {
				continue
			}
			candidate := ToHex(hslToRGB(h, s, l))
			if clearsAll(candidate) {
				return candidate, true
			}
		}
	}
	return "", false
}
