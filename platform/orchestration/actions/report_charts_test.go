// FILE: platform/orchestration/actions/report_charts_test.go
//
// Layout regressions for the gripper dossier's inline SVG charts.
//
// Both cases here were live on the 2026-07-26 fixture pages and were found by
// LOOKING at them, not by any check — which is the point of this file. An SVG
// viewBox clips rather than overflows, so a label that does not fit is not a
// visibly broken layout: it is a sentence cut off mid-word, and in a report
// whose whole doctrine is "every figure is real", clipped text reads as
// corrupted data. Nothing in the pipeline could see it.
//
// These assertions are geometric, not golden-file, deliberately: a golden file
// would fail on any spacing change and teach the next reader to re-bless it,
// which is how a real regression gets waved through.

package actions

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const svgViewBoxW = 700.0

var (
	// Value labels beside each bar (font-size 11), reference captions under the
	// plot (font-size 10, centred). Both carry x, y and their text.
	reValueLabel = regexp.MustCompile(`<text x="([0-9.]+)" y="([0-9.]+)" font-size="11" fill="[^"]*">([^<]*)</text>`)
	reRefCaption = regexp.MustCompile(`<text x="([0-9.]+)" y="([0-9.]+)" font-size="10" text-anchor="middle" fill="[^"]*">([^<]*)</text>`)
	reCapTip     = regexp.MustCompile(`<path d="M[0-9.]+ [0-9.]+ L`)
)

func mustFloat(t *testing.T, s string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("unparseable coordinate %q: %v", s, err)
	}
	return f
}

// The fixture that produced the clipped label: capacity headroom well past the
// 3× cap, with the longest verdict string the scorer emits.
func clippedLabelBars() []chartBar {
	return []chartBar{
		{Label: "Schmalz SGM-HP 50", Value: 6.42, Text: "6.42× (Insufficient data)", Pass: true},
		{Label: "Zimmer Group GEP5010IO-00-A", Value: 7.60, Text: "7.60× (No match)"},
		{Label: "OnRobot VG10", Value: 2.45, Text: "2.45× (Insufficient data)"},
		{Label: "OnRobot Gecko SP5", Value: 0.82, Text: "0.82× (No match)"},
	}
}

func headroomRefs() map[float64]string {
	return map[float64]string{1.0: "requirement (1.0×)", 1.25: "marginal threshold (1.25×)"}
}

// BUG 1 (live on the 07-26 fixtures): the right-hand gutter was a fixed 90
// units while "6.42× (Insufficient data)" needs ~150, so every capped bar's
// value label ran past the viewBox and was clipped mid-word — the page served
// "6.42× (Insufficien".
func TestNoValueLabelIsClippedByTheViewBox(t *testing.T) {
	svg := renderBarChartSVG("t", "s", clippedLabelBars(), 3.0, headroomRefs())
	if svg == "" {
		t.Fatal("no chart rendered")
	}
	bars := clippedLabelBars()
	matches := reValueLabel.FindAllStringSubmatch(svg, -1)
	if len(matches) != len(bars) {
		t.Fatalf("found %d value labels, want %d — the regex no longer matches the emitted shape, so this test is not checking anything",
			len(matches), len(bars))
	}
	for i, m := range matches {
		x, text := mustFloat(t, m[1]), m[3]
		right := x + estTextWidth(text, 11)
		if right > svgViewBoxW {
			t.Errorf("value label %q starts at x=%.1f and ends at %.1f, past the %.0f viewBox — it will be clipped mid-word",
				text, x, right, svgViewBoxW)
		}
		// Fitting inside the viewBox by SHORTENING the label is not a fix: the
		// figure is the content of the report. Mutation testing caught this —
		// reverting the computed gutter left the geometry assertion above
		// perfectly green, because fitText quietly truncated instead.
		if want := bars[i].Text; text != want {
			t.Errorf("value label rendered as %q, want %q — it fitted by losing content, not by sizing the gutter", text, want)
		}
	}
}

// BUG 2 (live on the 07-26 fixtures): 1.0× and 1.25× are 0.25 apart on a 3×
// axis, about 32 user units, while their captions are ~110 and ~140 wide. Drawn
// on one baseline they overprinted into "reqmiaegimealttufh1r.0ex/f)old (1.25×)".
func TestReferenceCaptionsDoNotOverprint(t *testing.T) {
	svg := renderBarChartSVG("t", "s", clippedLabelBars(), 3.0, headroomRefs())
	matches := reRefCaption.FindAllStringSubmatch(svg, -1)
	if len(matches) != 2 {
		t.Fatalf("found %d reference captions, want 2 — the regex no longer matches the emitted shape, so this test is not checking anything", len(matches))
	}

	type span struct {
		lo, hi float64
		text   string
	}
	byBaseline := map[string][]span{}
	for _, m := range matches {
		x, y, text := mustFloat(t, m[1]), m[2], m[3]
		w := estTextWidth(text, 10)
		byBaseline[y] = append(byBaseline[y], span{lo: x - w/2, hi: x + w/2, text: text})
	}
	for y, spans := range byBaseline {
		for i := 0; i < len(spans); i++ {
			for j := i + 1; j < len(spans); j++ {
				if spans[i].lo < spans[j].hi && spans[j].lo < spans[i].hi {
					t.Errorf("captions %q and %q share baseline y=%s and overlap horizontally — they will render on top of each other",
						spans[i].text, spans[j].text, y)
				}
			}
		}
	}
	// And the fix must not have worked by dropping one of them.
	if !strings.Contains(svg, "requirement (1.0×)") || !strings.Contains(svg, "marginal threshold (1.25×)") {
		t.Error("a reference caption is missing — overlap must be solved by stacking, not by omission")
	}
}

// Two bars clipped to the cap are the same length while the figures beside them
// differ (6.42× and 7.60×). The pointed tip is what stops that reading as equal.
func TestCappedBarsAreMarkedAndUncappedOnesAreNot(t *testing.T) {
	bars := clippedLabelBars() // 6.42, 7.60 over the cap; 2.45, 0.82 under it
	svg := renderBarChartSVG("t", "s", bars, 3.0, headroomRefs())
	if got, want := len(reCapTip.FindAllString(svg, -1)), 2; got != want {
		t.Errorf("%d capped-bar tips drawn, want %d (one per bar exceeding the cap)", got, want)
	}

	none := renderBarChartSVG("t", "s", []chartBar{{Label: "A", Value: 1.5, Text: "1.50×"}}, 3.0, nil)
	if got := len(reCapTip.FindAllString(none, -1)); got != 0 {
		t.Errorf("%d tips drawn for a chart with no capped bar, want 0", got)
	}
}

// The geometry work must not have cost the properties the chart already had.
func TestChartStaysStableEscapedAndScriptFree(t *testing.T) {
	bars := append(clippedLabelBars(), chartBar{
		Label: `Evil <script>alert(1)</script>`,
		Value: 1.0,
		Text:  `1.00× ("&" <b>)`,
	})
	a := renderBarChartSVG("t", "s", bars, 3.0, headroomRefs())
	b := renderBarChartSVG("t", "s", bars, 3.0, headroomRefs())
	if a != b {
		t.Error("chart SVG is not byte-stable across runs")
	}
	if strings.Contains(a, "<script") || strings.Contains(a, "<b>") {
		t.Error("candidate-supplied markup reached the SVG unescaped")
	}
	if strings.Count(a, "<svg") != 1 || !strings.HasSuffix(a, "</svg>") {
		t.Error("malformed SVG envelope")
	}
}

// fitText is the last-resort guard behind the computed gutter. It should be
// unreachable for real verdict strings — assert that, so nobody "fixes" a
// clipped label by leaning on truncation instead of on the gutter.
func TestRealVerdictLabelsAreNeverTruncated(t *testing.T) {
	for _, verdict := range []string{"Match", "Marginal", "No match", "Insufficient data"} {
		text := "12.34× (" + verdict + ")"
		svg := renderBarChartSVG("t", "s",
			[]chartBar{{Label: "Some Maker SOME-MODEL-1234", Value: 12.34, Text: text}}, 3.0, headroomRefs())
		if !strings.Contains(svg, text) {
			t.Errorf("label %q was shortened; the gutter should have fitted it", text)
		}
	}
}

func TestFitTextShortensOnlyWhenItMustAndSaysSo(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := fitText(long, 60, 11)
	if got == long {
		t.Fatal("fitText returned an over-long string unchanged")
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("shortened text %q does not end in an ellipsis — a silent cut is the failure mode being avoided", got)
	}
	if estTextWidth(got, 11) > 60 {
		t.Errorf("shortened text still exceeds the budget: %.1f > 60", estTextWidth(got, 11))
	}
	if short := "6.42× (No match)"; fitText(short, 300, 11) != short {
		t.Error("fitText altered a string that already fitted")
	}
}
