package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// The colour sweep.
//
// Both source sites scatter literal hex colours through page-local <style>
// blocks and inline style="" attributes, ignoring their own design tokens
// (website-design.com: 183x #fff, 142x #666, 99x #555; websitedesign.com's
// sub-pages: 59x #333, 25x #111, 17x #0f0). Those literals are what pins each
// page to its old skin, so they are what has to move.
//
// Two rules make this safe rather than a blind find-and-replace:
//
//  1. PROPERTY-AWARE. #111 means "ink" under `color:` and "a dark decorative
//     panel" under `background:`. The same literal maps to different targets
//     depending on the declaration it sits in, so the sweep parses declarations
//     rather than scanning raw text.
//
//  2. CSS CONTEXTS ONLY. A hex inside JavaScript is data, not skin — half these
//     tools generate colours for a living, and rewriting a canvas fill or a
//     palette seed would break the tool while looking like a successful reskin.
//     Script bodies are never touched; suspicious literals in them are reported.
//
// The paired flip is the subtle case: turning `background:#111` into a light
// panel orphans a `color:#fff` in the same declaration block as white-on-white.
// So a rule may declare flips_block, and any text colour in that same block is
// then mapped through the "flipped" context instead of "text".

// ColourMap is port/colour_map.json.
type ColourMap struct {
	// Allowlist holds literals that are functional, not decorative: a colour
	// tool's default swatch, sample data. Never rewritten, never reported.
	Allowlist []string `json:"allowlist"`
	Rules     []Rule   `json:"rules"`
}

// Rule maps one literal in one set of property contexts to one replacement.
type Rule struct {
	Hex string `json:"hex"`
	// Contexts: any of text, background, border, shadow, flipped.
	// Empty means every context.
	Contexts []string `json:"contexts,omitempty"`
	To       string   `json:"to"`
	// FlipsBlock marks a dark->light background flip, which re-contexts sibling
	// text declarations in the same block (see above).
	FlipsBlock bool   `json:"flips_block,omitempty"`
	Why        string `json:"why,omitempty"`
}

type colourEngine struct {
	// byHex[hex][context] = replacement
	byHex map[string]map[string]string
	// flips[hex] = true when a background match on this literal flips the block
	flips     map[string]bool
	allowlist map[string]bool
}

func newColourEngine(cm *ColourMap) (*colourEngine, error) {
	e := &colourEngine{
		byHex:     map[string]map[string]string{},
		flips:     map[string]bool{},
		allowlist: map[string]bool{},
	}
	for _, a := range cm.Allowlist {
		e.allowlist[normaliseHex(a)] = true
	}
	for _, r := range cm.Rules {
		h := normaliseHex(r.Hex)
		if h == "" {
			return nil, fmt.Errorf("colour_map: rule with unparseable hex %q", r.Hex)
		}
		if e.byHex[h] == nil {
			e.byHex[h] = map[string]string{}
		}
		ctxs := r.Contexts
		if len(ctxs) == 0 {
			ctxs = []string{"*"}
		}
		for _, c := range ctxs {
			if prev, dup := e.byHex[h][c]; dup && prev != r.To {
				return nil, fmt.Errorf("colour_map: %s/%s mapped twice (%s and %s)", h, c, prev, r.To)
			}
			e.byHex[h][c] = r.To
		}
		if r.FlipsBlock {
			e.flips[h] = true
		}
	}
	return e, nil
}

var (
	hexRe = regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`)
	// A CSS declaration block body: everything between { and }. Good enough for
	// the flat, un-nested stylesheets these pages carry (no @media nesting of
	// rules inside rules, no CSS nesting syntax).
	blockRe = regexp.MustCompile(`\{[^{}]*\}`)
)

// normaliseHex lowercases and expands #abc to #aabbcc. Returns "" if not a hex.
func normaliseHex(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if !strings.HasPrefix(s, "#") {
		return ""
	}
	d := s[1:]
	for _, c := range d {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return ""
		}
	}
	switch len(d) {
	case 3:
		return "#" + string([]byte{d[0], d[0], d[1], d[1], d[2], d[2]})
	case 4: // #rgba -> drop alpha for matching purposes
		return "#" + string([]byte{d[0], d[0], d[1], d[1], d[2], d[2]})
	case 6:
		return "#" + d
	case 8: // #rrggbbaa -> match on the rgb part
		return "#" + d[:6]
	}
	return ""
}

// isBrandToken reports whether a CSS value paints a saturated brand colour, and
// therefore needs light text on top of it.
func isBrandToken(val string) bool {
	v := strings.ToLower(val)
	for _, t := range []string{"var(--primary)", "var(--accent)", "var(--secondary)",
		"var(--danger)", "var(--primary-hover)", "var(--ok)", "var(--warn)", "var(--bad)"} {
		if strings.Contains(v, t) {
			return true
		}
	}
	return false
}

// contextOf classifies a CSS property name.
func contextOf(prop string) string {
	p := strings.ToLower(strings.TrimSpace(prop))
	switch {
	case p == "color", p == "caret-color", p == "-webkit-text-fill-color", p == "fill":
		return "text"
	case strings.HasPrefix(p, "background"):
		return "background"
	case strings.HasPrefix(p, "border"), strings.HasPrefix(p, "outline"), p == "stroke":
		return "border"
	case strings.HasSuffix(p, "shadow"):
		return "shadow"
	default:
		return "other"
	}
}

// sweepReport records what a page's sweep did and what it declined to do.
type sweepReport struct {
	replaced map[string]int // "hex->replacement" -> count
	unmapped map[string]int // hex seen in a CSS context with no rule
	inScript map[string]int // hex seen inside a <script> — reported, never touched
}

func newSweepReport() *sweepReport {
	return &sweepReport{
		replaced: map[string]int{},
		unmapped: map[string]int{},
		inScript: map[string]int{},
	}
}

func (r *sweepReport) sortedUnmapped() []string {
	var out []string
	for k := range r.unmapped {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// sweepStylesheet rewrites every declaration block in a stylesheet body.
func (e *colourEngine) sweepStylesheet(css string, rep *sweepReport) string {
	return blockRe.ReplaceAllStringFunc(css, func(block string) string {
		inner := strings.TrimSuffix(strings.TrimPrefix(block, "{"), "}")
		return "{" + e.sweepDeclarations(inner, rep) + "}"
	})
}

// sweepDeclarations rewrites one `prop: value; prop: value` run. This is also
// exactly the shape of an inline style="" attribute, which is why the same
// function serves both.
func (e *colourEngine) sweepDeclarations(decls string, rep *sweepReport) string {
	parts := strings.Split(decls, ";")

	// Pass 1: classify the block by what its background is doing.
	//
	//   flipped — a background this map turns from dark to light. Text in the
	//             same block was written for a dark panel and must be re-mapped.
	//   onBrand — the background is a saturated brand colour. Site B's buttons
	//             and tags say `background: var(--primary); color: #000`, which
	//             was legible against its old light-sage/orange but is not
	//             against anything: black on a mid-tone is poor either way. Ten
	//             blocks across six tools do this. Detected on the RAW value,
	//             because the background is usually already a var() and so has
	//             no hex for the rules to match.
	flipped, onBrand := false, false
	for _, p := range parts {
		prop, val, ok := splitDecl(p)
		if !ok || contextOf(prop) != "background" {
			continue
		}
		for _, raw := range hexRe.FindAllString(val, -1) {
			h := normaliseHex(raw)
			if e.flips[h] {
				flipped = true
			}
			if to, hit := e.byHex[h]["background"]; hit && isBrandToken(to) {
				onBrand = true
			}
			if to, hit := e.byHex[h]["*"]; hit && isBrandToken(to) {
				onBrand = true
			}
		}
		if isBrandToken(val) {
			onBrand = true
		}
	}

	// Pass 2: rewrite.
	for i, p := range parts {
		prop, val, ok := splitDecl(p)
		if !ok {
			continue
		}
		ctx := contextOf(prop)
		if ctx == "text" {
			switch {
			case onBrand:
				ctx = "on-brand"
			case flipped:
				ctx = "flipped"
			}
		}
		newVal := hexRe.ReplaceAllStringFunc(val, func(raw string) string {
			h := normaliseHex(raw)
			if h == "" || e.allowlist[h] {
				return raw
			}
			m := e.byHex[h]
			if m == nil {
				rep.unmapped[h]++
				return raw
			}
			to, hit := m[ctx]
			if !hit {
				to, hit = m["*"]
			}
			if !hit {
				rep.unmapped[h+" ("+ctx+")"]++
				return raw
			}
			rep.replaced[h+" -> "+to]++
			return to
		})
		if newVal != val {
			// Preserve the original leading whitespace so the CSS stays readable.
			lead := p[:len(p)-len(strings.TrimLeft(p, " \t\r\n"))]
			parts[i] = lead + strings.TrimSpace(prop) + ": " + strings.TrimSpace(newVal)
		}
	}
	return strings.Join(parts, ";")
}

// splitDecl splits "  color: #fff " into ("color", " #fff ").
func splitDecl(d string) (prop, val string, ok bool) {
	i := strings.Index(d, ":")
	if i < 0 {
		return "", "", false
	}
	prop = d[:i]
	val = d[i+1:]
	if strings.TrimSpace(prop) == "" {
		return "", "", false
	}
	return prop, val, true
}

// noteScriptColours records hex literals inside script bodies without touching
// them. A genuine skin colour hiding in JS is then a deliberate override, not a
// silent miss.
func (e *colourEngine) noteScriptColours(js string, rep *sweepReport) {
	for _, raw := range hexRe.FindAllString(js, -1) {
		h := normaliseHex(raw)
		if h == "" || e.allowlist[h] {
			continue
		}
		if _, mapped := e.byHex[h]; mapped {
			rep.inScript[h]++
		}
	}
}
