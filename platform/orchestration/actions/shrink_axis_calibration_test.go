package actions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// THE CALIBRATION HARNESS FOR THE SHRINK FLOOR'S AXIS — reusable, in the repo,
// and measuring with the REAL functions.
//
// WHY IT IS A COMMITTED TEST AND NOT A SCRATCH SCRIPT (bugs_open/293).
// section_visible_text.go's header ends "RE-RUN THAT CALIBRATION before
// changing it", and until now that was not a thing anyone could do: the 117-pair
// run that justified the section editor's axis lived in one session's scratch
// directory and its numbers survive only as prose in a doc. A calibration a
// later reader cannot re-run is a claim, not evidence — and this floor's whole
// history is of axes that looked right and measured the wrong thing.
//
// It also removes the trap that produced this file's own first wrong numbers:
// the 285 lane measured the axis with a REGEX APPROXIMATION of the parser and
// published 2,754 → 68 for a pair that really reads 2,143 → 16, and a refusal
// count of 1 where the real implementation refuses 3. This harness calls
// visibleTextLength and evaluateSectionShrink themselves, so "measured with the
// code" is structural rather than a promise.
//
// SKIPPED unless SHRINK_CALIBRATION_JSONL names an export — the data is 11 MB of
// live page HTML and does not belong in the repo. `go test ./...` is unaffected.
//
// Export it with docs/agent_docs/docs024_key_docs_latest/bugfix_293_whole_page_shrink_axis/
// export_pairs.sh (the join and its disconfirming control are documented there).
// Each line: {"domain","page_name","slot_name","del_at","existing","incoming"}.
//
//	SHRINK_CALIBRATION_JSONL=/path/pairs.jsonl \
//	  go test ./platform/orchestration/actions/ -run TestShrinkAxisCalibration -v
type calibrationPair struct {
	Domain   string `json:"domain"`
	PageName string `json:"page_name"`
	Slot     string `json:"slot_name"`
	At       string `json:"del_at"`
	Existing string `json:"existing"`
	Incoming string `json:"incoming"`

	// GapS is seconds from the archive event to the state paired with it. On the
	// terminal export it is the delete→re-insert lag, which is under 5 s for
	// 1,109 of 1,123 rebuilds; on the intermediate export it can be days, and a
	// long gap means the slot was NOT re-inserted by that rebuild. The guard
	// treats an absent slot as a DROP and declines to judge it, so a long-gap
	// pair is not a write the guard would ever have seen — counting one as a
	// false refusal invents a cost the change does not have.
	GapS float64 `json:"gap_s"`
}

// guardJudgedGapSeconds separates the two. A whole-page rebuild deletes and
// re-inserts in one transaction; anything beyond a few minutes is a slot that
// spent the interval ABSENT.
const guardJudgedGapSeconds = 300

func (p calibrationPair) id() string {
	return fmt.Sprintf("%s %s / %s @ %s", p.Domain, p.PageName, p.Slot, p.At)
}

// axis is one way of reducing a stored section to a length. The floor's whole
// defect class is that this choice was made independently at each call site, so
// the harness makes it the parameter it always was.
type axis struct {
	name    string
	measure func(string) int
}

func TestShrinkAxisCalibration(t *testing.T) {
	path := os.Getenv("SHRINK_CALIBRATION_JSONL")
	if path == "" {
		t.Skip("set SHRINK_CALIBRATION_JSONL to an exported pair file to run the calibration")
	}
	pairs, err := loadCalibrationPairs(path)
	if err != nil {
		t.Fatalf("loading pairs: %v", err)
	}
	if len(pairs) == 0 {
		t.Fatal("no pairs loaded — an empty calibration reports 'no refusals' and means nothing")
	}

	axes := []axis{
		{"tag-stripped (RETIRED 2026-08-17, kept as comparator)", tagStrippedLengthForCalibration},
		{"visible text (LIVE on all three floors)", visibleTextLength},
	}

	// Refusals are keyed by pair id per axis so the agreement between axes is
	// counted rather than asserted — the section editor's calibration found the
	// two axes agreeing on ZERO of its 117 pairs, and that is the finding that
	// makes this a correction rather than an additional floor.
	refused := make([]map[string]bool, len(axes))
	inScope := make([]int, len(axes))
	details := make([][]string, len(axes))

	for i, a := range axes {
		refused[i] = map[string]bool{}
		for _, p := range pairs {
			existing := map[string]int{p.Slot: a.measure(p.Existing)}
			incoming := map[string]int{p.Slot: a.measure(p.Incoming)}
			if existing[p.Slot] >= minShrinkGuardVisibleChars {
				inScope[i]++
			}
			for _, v := range evaluateSectionShrink(defaultSectionShrinkFloor, minShrinkGuardVisibleChars, existing, incoming) {
				refused[i][p.id()] = true
				details[i] = append(details[i], fmt.Sprintf("    %s: %d→%d (%.1f%% kept)",
					p.id(), v.Existing, v.Incoming, v.ratio()*100))
			}
		}
	}

	t.Logf("pairs: %d (floor %.2f, minimum %d chars on the EXISTING side)",
		len(pairs), defaultSectionShrinkFloor, minShrinkGuardVisibleChars)
	for i, a := range axes {
		sort.Strings(details[i])
		t.Logf("axis %-44s in scope %4d/%4d   refuses %3d",
			a.name, inScope[i], len(pairs), len(refused[i]))
		for _, d := range details[i] {
			t.Log(d)
		}
	}
	both := 0
	for id := range refused[0] {
		if refused[1][id] {
			both++
		}
	}
	t.Logf("pairs both axes refuse: %d", both)
	for i, a := range axes {
		for j, b := range axes {
			if i >= j {
				continue
			}
			only := 0
			for id := range refused[i] {
				if !refused[j][id] {
					only++
				}
			}
			t.Logf("refused by %q but NOT by %q: %d", a.name, b.name, only)
		}
	}
}

// tagStrippedLengthForCalibration is the whole-page path's live axis, expressed
// as a measure so the harness can compare it against visibleTextLength on the
// same pairs. It reproduces strippedIncomingBySlot's arithmetic exactly
// (TrimSpace then len of the tag-stripped string); keep the two in step, and if
// this file's copy ever drifts the harness is measuring a third axis nobody
// runs. Asserted against the real function in TestCalibrationAxisMatchesGuard.
func tagStrippedLengthForCalibration(html string) int {
	return strippedIncomingBySlot([]SectionData{{ComponentName: "s", HTML: html}})["s"]
}

// TestCalibrationAxisMatchesGuard is the control that keeps the harness honest:
// it proves the calibration's tag-stripped measure IS the shipped one rather
// than a lookalike. It runs in the normal suite, without the export.
func TestCalibrationAxisMatchesGuard(t *testing.T) {
	samples := []string{
		"",
		"   ",
		"<p>hello world</p>",
		"<style>.a{color:#fff;padding:0}</style><article></article>",
		"<div class=\"x\"><script>var a=1;</script>text</div>",
	}
	for _, s := range samples {
		want := strippedIncomingBySlot([]SectionData{{ComponentName: "slot", HTML: s}})["slot"]
		if got := tagStrippedLengthForCalibration(s); got != want {
			t.Fatalf("calibration axis drifted from the guard's own measure for %q: got %d want %d", s, got, want)
		}
	}
	// And the two axes must actually differ on the shape this bug is about —
	// a stylesheet where the prose used to be. Without this, a harness whose
	// visible measure silently equalled the tag-stripped one would report
	// perfect agreement and read as reassuring.
	poison := "<style>.wrap{display:grid;grid-template-columns:repeat(3,1fr);gap:2rem;padding:4rem}</style><article></article>"
	if tagStrippedLengthForCalibration(poison) <= visibleTextLength(poison) {
		t.Fatalf("the two axes do not disagree on a CSS-for-prose swap (tag-stripped %d, visible %d) — "+
			"the harness cannot detect what it exists to measure",
			tagStrippedLengthForCalibration(poison), visibleTextLength(poison))
	}
}

// TestShrinkAxisBlindness measures what the SHIPPED axis cannot see, on the
// real population, rather than waiting for an incident to prove it.
//
// The calibration above answers "would the new axis have refused any of the
// writes that actually happened" — and over 1,079 whole-page rebuild writes the
// answer is none, because no rebuild in the archived window hollowed a page.
// That is a measured absence of false refusals; it is NOT an argument for the
// change, and quoting it as one would be the "a post-fix zero needs a demand
// control" mistake in a new costume.
//
// So this asks the prospective question instead. For every existing section in
// the export it CONSTRUCTS the failure: the exact shape bugs_closed/285 shipped
// — every word of prose deleted, the wrapper markup and its <style>/<script>
// left intact — and asks each axis whether it would refuse that write. A pair
// the axis lets through is a live page where the whole of its prose can be
// deleted in one rebuild with nothing to stop it.
//
// The simulation carries its own controls (below), because a hollower that
// quietly did nothing would report the shipped axis as protective.
func TestShrinkAxisBlindness(t *testing.T) {
	path := os.Getenv("SHRINK_CALIBRATION_JSONL")
	if path == "" {
		t.Skip("set SHRINK_CALIBRATION_JSONL to an exported pair file to run the calibration")
	}
	pairs, err := loadCalibrationPairs(path)
	if err != nil {
		t.Fatalf("loading pairs: %v", err)
	}

	axes := []axis{
		{"tag-stripped (RETIRED 2026-08-17, kept as comparator)", tagStrippedLengthForCalibration},
		{"visible text (LIVE on all three floors)", visibleTextLength},
	}
	type outcome struct{ inScope, refused, allowed int }
	results := make([]outcome, len(axes))

	simulated, wroteProse := 0, 0
	for _, p := range pairs {
		hollow := hollowVisibleText(p.Existing)

		// CONTROL 1 — the hollower really removed the prose. Without this a
		// no-op hollower makes every axis look like it is protecting the page.
		if visibleTextLength(hollow) != 0 {
			t.Fatalf("hollower left %d visible chars on %s — the simulation is not simulating the failure",
				visibleTextLength(hollow), p.id())
		}
		// CONTROL 2 — it removed prose that was THERE. A section with no prose
		// to start with cannot demonstrate blindness either way, so it is
		// excluded from the denominator rather than counted as protected.
		if visibleTextLength(p.Existing) == 0 {
			continue
		}
		simulated++
		if visibleTextLength(p.Existing) >= minShrinkGuardVisibleChars {
			wroteProse++
		}

		for i, a := range axes {
			existing := map[string]int{p.Slot: a.measure(p.Existing)}
			incoming := map[string]int{p.Slot: a.measure(hollow)}
			if existing[p.Slot] < minShrinkGuardVisibleChars {
				continue // out of this axis's scope: it declines to judge
			}
			results[i].inScope++
			if len(evaluateSectionShrink(defaultSectionShrinkFloor, minShrinkGuardVisibleChars, existing, incoming)) > 0 {
				results[i].refused++
			} else {
				results[i].allowed++
			}
		}
	}

	// CONTROL 3 — a demand control on the simulation itself: the population must
	// actually contain prose-bearing sections, or "the axis allows the wipe" is
	// a statement about an empty set.
	if simulated == 0 || wroteProse == 0 {
		t.Fatalf("no prose-bearing sections in the export (%d simulated, %d over the minimum) — "+
			"this measurement cannot come out either way", simulated, wroteProse)
	}

	t.Logf("prose-bearing sections simulated: %d of %d pairs (%d carry ≥%d visible chars)",
		simulated, len(pairs), wroteProse, minShrinkGuardVisibleChars)
	for i, a := range axes {
		t.Logf("axis %-44s judged %4d   REFUSES the wipe %4d   ALLOWS it %4d",
			a.name, results[i].inScope, results[i].refused, results[i].allowed)
	}
}

// TestMinimumSweep sizes what the EXISTING-side minimum costs each axis. The
// 500-char minimum was chosen against tag-stripped lengths, where a stylesheet
// inflates every count; on visible text the same 500 excludes ordinary prose
// slots, so "which minimum" is a separate question from "which axis" and this
// keeps them separate. Reported, not decided here.
func TestMinimumSweep(t *testing.T) {
	path := os.Getenv("SHRINK_CALIBRATION_JSONL")
	if path == "" {
		t.Skip("set SHRINK_CALIBRATION_JSONL to an exported pair file to run the calibration")
	}
	pairs, err := loadCalibrationPairs(path)
	if err != nil {
		t.Fatalf("loading pairs: %v", err)
	}
	// The sweep drives the REAL decision at every minimum. It used to replicate
	// evaluateSectionShrink's arithmetic — because the minimum was a hardcoded
	// constant and a sweep had no other way in — and carried a pinning assertion so
	// the copy could not drift. bugs_open/293 made the minimum a parameter, which
	// deleted both the copy and the need for the assertion: there is now one rule
	// and the harness turns its dial.
	for _, min := range []int{500, 400, 300, 200, 150, 120, 100, 50} {
		inScope, refused, hollowCaught, refusedGuardJudged := 0, 0, 0, 0
		var judged []string
		for _, p := range pairs {
			ex := map[string]int{p.Slot: visibleTextLength(p.Existing)}
			if ex[p.Slot] >= min {
				inScope++
			}
			if v := evaluateSectionShrink(defaultSectionShrinkFloor, min, ex,
				map[string]int{p.Slot: visibleTextLength(p.Incoming)}); len(v) > 0 {
				refused++
				if p.GapS <= guardJudgedGapSeconds {
					refusedGuardJudged++
					judged = append(judged, fmt.Sprintf("        %s: %d→%d visible, gap %.0fs",
						p.id(), v[0].Existing, v[0].Incoming, p.GapS))
				}
			}
			if len(evaluateSectionShrink(defaultSectionShrinkFloor, min, ex,
				map[string]int{p.Slot: visibleTextLength(hollowVisibleText(p.Existing))})) > 0 {
				hollowCaught++
			}
		}
		marker := ""
		if min == minShrinkGuardVisibleChars {
			marker = "  ← SHIPPED"
		}
		t.Logf("minimum %4d visible chars: in scope %4d/%4d   refuses real writes %2d (of which the guard would actually have judged %2d)   catches the wipe %4d%s",
			min, inScope, len(pairs), refused, refusedGuardJudged, hollowCaught, marker)
		sort.Strings(judged)
		for _, j := range judged {
			t.Log(j)
		}
	}
}

// TestPageTotalGuardBlindness sizes the THIRD copy of the same axis — the
// page-total content-regression guard inlined in save_page_sections_action.go
// (~:549), which refuses a save whose whole-page tag-stripped text falls below a
// QUARTER of what is deployed. It is the oldest of the three and the same
// stylesheet-counts-as-text substitution applies to it, so "is the per-slot floor
// the only blind one?" is a measurement rather than a scoping opinion.
//
// APPROXIMATE, and marked so. The real guard sums every deployed row on the page
// and compares against the whole incoming section set; this sums the paired slots
// only, which is most of a page but not provably all of it. Directional evidence
// for a scope decision, not a calibration.
func TestPageTotalGuardBlindness(t *testing.T) {
	path := os.Getenv("SHRINK_CALIBRATION_JSONL")
	if path == "" {
		t.Skip("set SHRINK_CALIBRATION_JSONL to an exported pair file to run the calibration")
	}
	pairs, err := loadCalibrationPairs(path)
	if err != nil {
		t.Fatalf("loading pairs: %v", err)
	}
	// The live rule: existing page text over 200 chars, refuse below a quarter.
	const pageMinChars, pageFloor = 200, 0.25
	type totals struct{ tagEx, tagIn, tagHollow, visEx, visIn, visHollow int }
	byPage := map[string]*totals{}
	for _, p := range pairs {
		tp := byPage[p.PageName+"@"+p.Domain]
		if tp == nil {
			tp = &totals{}
			byPage[p.PageName+"@"+p.Domain] = tp
		}
		hollow := hollowVisibleText(p.Existing)
		tp.tagEx += tagStrippedLengthForCalibration(p.Existing)
		tp.tagIn += tagStrippedLengthForCalibration(p.Incoming)
		tp.tagHollow += tagStrippedLengthForCalibration(hollow)
		tp.visEx += visibleTextLength(p.Existing)
		tp.visIn += visibleTextLength(p.Incoming)
		tp.visHollow += visibleTextLength(hollow)
	}
	var tagJudged, tagRefuses, tagReal, visJudged, visRefuses, visReal int
	for _, tp := range byPage {
		if tp.tagEx > pageMinChars {
			tagJudged++
			if float64(tp.tagHollow) < float64(tp.tagEx)*pageFloor {
				tagRefuses++
			}
			if float64(tp.tagIn) < float64(tp.tagEx)*pageFloor {
				tagReal++
			}
		}
		if tp.visEx > pageMinChars {
			visJudged++
			if float64(tp.visHollow) < float64(tp.visEx)*pageFloor {
				visRefuses++
			}
			// The REAL rebuild, not the constructed one — the false-refusal
			// question for this guard, which the wipe column cannot answer.
			if float64(tp.visIn) < float64(tp.visEx)*pageFloor {
				visReal++
			}
		}
	}
	t.Logf("pages: %d  [APPROXIMATE — paired slots only, not every deployed row]", len(byPage))
	t.Logf("page-total axis tag-stripped (live): judged %3d  REFUSES a whole-page prose wipe %3d  ALLOWS it %3d   refuses REAL rebuilds %2d",
		tagJudged, tagRefuses, tagJudged-tagRefuses, tagReal)
	t.Logf("page-total axis visible text:        judged %3d  REFUSES a whole-page prose wipe %3d  ALLOWS it %3d   refuses REAL rebuilds %2d",
		visJudged, visRefuses, visJudged-visRefuses, visReal)
}

// hollowVisibleText builds bugs_closed/285's failure out of a real section:
// every text node a reader would see is emptied, and everything else — the tag
// structure, the class attributes, and the CONTENT of <style> and <script> — is
// left exactly as it was. That is what a wrapper-stylesheet replacement looks
// like to a guard, and it is the shape both live axes are asked to judge here.
func hollowVisibleText(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return htmlStr
	}
	var walk func(*html.Node)
	skip := map[string]bool{"script": true, "style": true, "noscript": true, "template": true, "head": true}
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && skip[n.Data] {
			return // leave stylesheet and script source untouched
		}
		if n.Type == html.TextNode {
			n.Data = ""
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return htmlStr
	}
	return buf.String()
}

func loadCalibrationPairs(path string) ([]calibrationPair, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var pairs []calibrationPair
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20) // live sections run to megabytes
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var p calibrationPair
		if err := json.Unmarshal(line, &p); err != nil {
			return nil, fmt.Errorf("line %d: %w", len(pairs)+1, err)
		}
		pairs = append(pairs, p)
	}
	return pairs, sc.Err()
}
