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
		plotW     = width - labelW - 90.0
	)
	height := topPad + rowH*float64(len(bars)) + bottomPad

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %.0f %.0f" role="img" xmlns="http://www.w3.org/2000/svg" style="max-width:100%%;height:auto;font-family:inherit;">`, width, height)
	fmt.Fprintf(&b, `<title>%s</title><desc>%s</desc>`, html.EscapeString(title), html.EscapeString(subtitle))
	fmt.Fprintf(&b, `<text x="0" y="18" font-size="15" font-weight="600" fill="var(--color-text, #1a1a1a)">%s</text>`, html.EscapeString(title))
	fmt.Fprintf(&b, `<text x="0" y="36" font-size="12" fill="var(--color-text-muted, #666)">%s</text>`, html.EscapeString(subtitle))

	// Reference lines behind the bars — keys sorted so the SVG is
	// byte-stable across runs (map iteration order is random; an unstable
	// chart would re-diff every committed report page).
	refKeys := make([]float64, 0, len(refs))
	for v := range refs {
		refKeys = append(refKeys, v)
	}
	sort.Float64s(refKeys)
	for _, v := range refKeys {
		caption := refs[v]
		if v <= 0 || v > maxValue {
			continue
		}
		x := labelW + plotW*(v/maxValue)
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%.0f" x2="%.1f" y2="%.1f" stroke="var(--color-text-muted, #888)" stroke-dasharray="4 3" stroke-width="1"/>`,
			x, topPad-6, x, height-bottomPad+6)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.0f" font-size="10" text-anchor="middle" fill="var(--color-text-muted, #666)">%s</text>`,
			x, height-8, html.EscapeString(caption))
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
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.1f" width="%.1f" height="%.0f" rx="3" fill="%s" fill-opacity="0.85"/>`,
			labelW, y+6, w, rowH-12, fill)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="11" fill="var(--color-text, #1a1a1a)">%s</text>`,
			labelW+w+6, y+rowH/2+4, html.EscapeString(bar.Text))
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
