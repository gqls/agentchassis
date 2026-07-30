// FILE: platform/orchestration/actions/report_charts.go
//
// Dependency-free inline-SVG chart helpers for the gripper dossier report
// (create_report_page). Deliberately NOT go-echarts: the report is a static
// deployed page, two simple comparison charts do not justify a new
// fleet-wide dependency plus a JS runtime tag, and inline SVG is
// golden-file-testable and cannot break at page-view time
// (DESIGN_2026-07-24_gripper_dossier_pilot.md §3 A6; the CAPABILITIES
// "[LIVE] go-echarts" claim was corrected 2026-07-24 — no chart renderer
// existed before this file).
//
// Doctrine preserved: every bar is drawn from a real figure the scoring
// action computed or a manufacturer published; nothing here invents data.
// Colours ride the site's CSS variables with plain fallbacks.

package actions

import (
	"fmt"
	"html"
	"sort"
	"strings"
)

type chartBar struct {
	Label string  // candidate name (escaped here)
	Value float64 // the real figure
	Text  string  // printed value label, e.g. "1520 N" or "7.6×"
	Pass  bool    // colours the bar: pass-ish vs fail-ish
}

// renderBarChartSVG draws horizontal bars with optional vertical reference
// lines (value → caption). Values are clamped to maxValue for geometry only;
// the printed Text always carries the true figure.
//
// Two layout rules here exist because the 2026-07-26 fixture pages got them
// wrong, visibly, and an SVG viewBox CLIPS rather than overflows — so the
// damage looked like mangled content rather than a layout bug:
//
//   - The value label is drawn outside the bar, so the right-hand gutter is
//     sized from the widest label rather than fixed. It was a fixed 90 units
//     while "6.42× (Insufficient data)" needs ~150, so every capped bar's
//     label was cut off mid-word ("6.42× (Insufficien").
//   - Reference captions are stacked into lanes when they would overprint.
//     1.0× and 1.25× are 0.25 apart on a 3× axis — about 32 user units — and
//     their captions are ~110 and ~140 wide, so on one line they rendered as
//     "reqmiaegimealttufh1r.0ex/f)old (1.25×)".
//
// A capped bar also gets a pointed tip, because two bars clipped to the same
// length otherwise read as equal when the figures beside them differ.
func renderBarChartSVG(title, subtitle string, bars []chartBar, maxValue float64, refs map[float64]string) string {
	if len(bars) == 0 || maxValue <= 0 {
		return ""
	}
	const (
		width     = 700.0
		labelW    = 230.0
		rowH      = 30.0
		topPad    = 46.0
		bottomPad = 26.0
		valueFont = 11.0
		refFont   = 10.0
		refLaneH  = 12.0
		minGutter = 90.0
		minPlotW  = 120.0
	)

	// Right-hand gutter: wide enough for the longest value label.
	gutter := minGutter
	for _, bar := range bars {
		if w := estTextWidth(bar.Text, valueFont) + 12; w > gutter {
			gutter = w
		}
	}
	if maxGutter := width - labelW - minPlotW; gutter > maxGutter {
		gutter = maxGutter
	}
	plotW := width - labelW - gutter
	maxTextW := gutter - 12

	plotBottom := topPad + rowH*float64(len(bars))

	// Reference lines behind the bars — keys sorted so the SVG is
	// byte-stable across runs (map iteration order is random; an unstable
	// chart would re-diff every committed report page).
	refKeys := make([]float64, 0, len(refs))
	for v := range refs {
		refKeys = append(refKeys, v)
	}
	sort.Float64s(refKeys)

	type refPlacement struct {
		x, textX float64
		lane     int
		caption  string
	}
	var placed []refPlacement
	var laneRight []float64 // right edge of the last caption placed in each lane
	for _, v := range refKeys {
		if v <= 0 || v > maxValue {
			continue
		}
		caption := refs[v]
		x := labelW + plotW*(v/maxValue)
		w := estTextWidth(caption, refFont)
		textX := x
		if textX-w/2 < 0 {
			textX = w / 2
		}
		if textX+w/2 > width {
			textX = width - w/2
		}
		lane := 0
		for lane < len(laneRight) && textX-w/2 < laneRight[lane]+6 {
			lane++
		}
		if lane == len(laneRight) {
			laneRight = append(laneRight, 0)
		}
		laneRight[lane] = textX + w/2
		placed = append(placed, refPlacement{x: x, textX: textX, lane: lane, caption: caption})
	}
	extraLanes := 0.0
	if len(laneRight) > 1 {
		extraLanes = float64(len(laneRight)-1) * refLaneH
	}
	height := plotBottom + bottomPad + extraLanes

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %.0f %.0f" role="img" xmlns="http://www.w3.org/2000/svg" style="max-width:100%%;height:auto;font-family:inherit;">`, width, height)
	fmt.Fprintf(&b, `<title>%s</title><desc>%s</desc>`, html.EscapeString(title), html.EscapeString(subtitle))
	fmt.Fprintf(&b, `<text x="0" y="18" font-size="15" font-weight="600" fill="var(--color-text, #1a1a1a)">%s</text>`, html.EscapeString(title))
	fmt.Fprintf(&b, `<text x="0" y="36" font-size="12" fill="var(--color-text-muted, #666)">%s</text>`, html.EscapeString(subtitle))

	for _, p := range placed {
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%.0f" x2="%.1f" y2="%.1f" stroke="var(--color-text-muted, #888)" stroke-dasharray="4 3" stroke-width="1"/>`,
			p.x, topPad-6, p.x, plotBottom+6)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="%.0f" text-anchor="middle" fill="var(--color-text-muted, #666)">%s</text>`,
			p.textX, plotBottom+18+float64(p.lane)*refLaneH, refFont, html.EscapeString(p.caption))
	}

	for i, bar := range bars {
		y := topPad + rowH*float64(i)
		v := bar.Value
		if v > maxValue {
			v = maxValue
		}
		if v < 0 {
			v = 0
		}
		w := plotW * (v / maxValue)
		fill := "var(--color-primary, #2b6cb0)"
		if !bar.Pass {
			fill = "var(--color-text-muted, #b23b3b)"
		}
		fmt.Fprintf(&b, `<text x="%.0f" y="%.1f" font-size="12" text-anchor="end" fill="var(--color-text, #1a1a1a)">%s</text>`,
			labelW-10, y+rowH/2+4, html.EscapeString(truncateLabel(bar.Label, 32)))

		// A bar clipped for scale ends in a point, so it cannot be read as an
		// exact length. The rect stops short and the tip makes up the width.
		rectW := w
		capped := bar.Value > maxValue
		if capped && rectW > 10 {
			rectW -= 8
		}
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.1f" width="%.1f" height="%.0f" rx="3" fill="%s" fill-opacity="0.85"/>`,
			labelW, y+6, rectW, rowH-12, fill)
		if capped {
			fmt.Fprintf(&b, `<path d="M%.1f %.1f L%.1f %.1f L%.1f %.1f Z" fill="%s" fill-opacity="0.85"/>`,
				labelW+rectW, y+6, labelW+w, y+rowH/2, labelW+rectW, y+rowH-6, fill)
		}

		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="%.0f" fill="var(--color-text, #1a1a1a)">%s</text>`,
			labelW+w+6, y+rowH/2+4, valueFont, html.EscapeString(fitText(bar.Text, maxTextW, valueFont)))
	}
	b.WriteString(`</svg>`)
	return b.String()
}

func truncateLabel(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

// estTextWidth approximates a rendered string's width without a font metric
// table. 0.55em per rune is a deliberate OVER-estimate for the proportional
// faces these reports use, because under-estimating is the direction that
// clips text, and clipped text in a report about honest numbers reads as
// corrupted data rather than as a layout fault.
func estTextWidth(s string, fontSize float64) float64 {
	return float64(len([]rune(s))) * fontSize * 0.55
}

// fitText shortens s until estTextWidth fits maxW, ellipsis included. Only
// reachable when a label is far longer than any real verdict string — but it
// fails visibly (an ellipsis) rather than by silent clipping at the viewBox.
func fitText(s string, maxW, fontSize float64) string {
	if maxW <= 0 || estTextWidth(s, fontSize) <= maxW {
		return s
	}
	r := []rune(s)
	for n := len(r) - 1; n > 0; n-- {
		if candidate := string(r[:n]) + "…"; estTextWidth(candidate, fontSize) <= maxW {
			return candidate
		}
	}
	return "…"
}

// renderHeadroomChart: capacity/need headroom per candidate, reference lines
// at 1.0 (bare requirement) and 1.25 (the Marginal threshold the verdicts
// use). Geometry capped at 3× so one huge gripper doesn't flatten the rest.
//
// Returns the omitted candidate names alongside the SVG. A candidate with no
// comparable capacity figure gets no bar — correct, because inventing one
// would be the exact dishonesty this pilot exists to prevent — but a silent
// skip makes two very different situations look identical: a figure the
// manufacturer genuinely does not publish, and a figure lost to an upstream
// fetch or normalisation bug. The caller names the omissions on the page, so
// the second case is visible to a reader who knows the product (council
// 7ed137d1, render_guardian + bug_historian: fail loud, not silent).
func renderHeadroomChart(cands []assessment) (svg string, omitted []string) {
	var bars []chartBar
	for _, a := range cands {
		if a.Headroom <= 0 {
			omitted = append(omitted, a.Name)
			continue // no comparable capacity figure — never draw an invented bar
		}
		bars = append(bars, chartBar{
			Label: a.Name,
			Value: a.Headroom,
			Text:  fmt.Sprintf("%.2f× (%s)", a.Headroom, a.Verdict),
			Pass:  a.Rank <= 1,
		})
	}
	return renderBarChartSVG(
		"Capacity headroom against your requirement",
		"Published capacity ÷ computed requirement, per candidate; ≥1.25× clears the marginal band. Bars capped at 3× for scale.",
		bars, 3.0,
		map[float64]string{1.0: "requirement (1.0×)", 1.25: "marginal threshold (1.25×)"}), omitted
}
