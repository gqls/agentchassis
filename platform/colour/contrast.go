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
