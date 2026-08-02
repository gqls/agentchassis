// Parity between the Go action and the Python builder it replaces.
//
// WHY THIS TEST IS THE IMPORTANT ONE
// This action is a SECOND implementation of a feed that already has a proven
// first one (provocation_pipeline/builder/build_provocations.py, verified across
// 39 dates and currently the source of what vonc.com serves). Two implementations
// of one artefact is a real risk: they can drift in ways no unit test notices,
// because each is self-consistent. The invariant tests in
// provocation_feed_action_test.go check that the Go side obeys the RULES; this
// one checks it produces the same BYTES from the same content.
//
// The fixtures:
//   testdata/provocation_pool.json         — the nine live rows, dumped from the
//                                            `provocations` table (migration 282)
//   testdata/provocation_feed_golden.json  — build_provocations.py --date
//                                            2026-07-31, with generated_at
//                                            replaced by the literal GOLDEN
//
// Regenerate both together, never one alone — a golden refreshed from the code
// under test only ever agrees with it.

package actions

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

type poolFixtureRow struct {
	Slug       string `json:"slug"`
	PublishOn  string `json:"publish_on"`
	Title      string `json:"title"`
	Teaser     string `json:"teaser"`
	CardDesc   string `json:"card_desc"`
	DetailBody string `json:"detail_body"`
	Headline   string `json:"headline"`
	Body       string `json:"body"`
}

func loadPoolFixture(t *testing.T) []provocation {
	t.Helper()
	raw, err := os.ReadFile("testdata/provocation_pool.json")
	if err != nil {
		t.Fatalf("read pool fixture: %v", err)
	}
	var rows []poolFixtureRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("parse pool fixture: %v", err)
	}
	out := make([]provocation, 0, len(rows))
	for _, r := range rows {
		d, err := time.Parse("2006-01-02", r.PublishOn)
		if err != nil {
			t.Fatalf("bad publish_on %q: %v", r.PublishOn, err)
		}
		out = append(out, provocation{
			Slug: r.Slug, PublishOn: d, Title: r.Title, Teaser: r.Teaser,
			CardDesc: r.CardDesc, DetailBody: r.DetailBody,
			Headline: r.Headline, Body: r.Body,
		})
	}
	return out
}

// normalise round-trips a value through JSON so two structurally identical feeds
// compare equal regardless of how they were built. Go marshals map keys in sorted
// order and Python preserves insertion order, so a raw byte comparison would fail
// on formatting alone and tell us nothing about the content.
func normalise(t *testing.T, v interface{}) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var generic interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.MarshalIndent(generic, "", "  ")
	if err != nil {
		t.Fatalf("remarshal: %v", err)
	}
	return string(out)
}

func TestGoFeedMatchesThePythonBuilderExactly(t *testing.T) {
	schedule := loadPoolFixture(t)
	if len(schedule) != 9 {
		t.Fatalf("fixture has %d rows, expected the 9 live provocations", len(schedule))
	}

	feed, today, err := buildProvocationFeed(schedule, day("2026-07-31"), "GOLDEN")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	goldenRaw, err := os.ReadFile("testdata/provocation_feed_golden.json")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var golden interface{}
	if err := json.Unmarshal(goldenRaw, &golden); err != nil {
		t.Fatalf("parse golden: %v", err)
	}

	got, want := normalise(t, feed), normalise(t, golden)
	if got != want {
		// Report where, not just that — a 20KB diff of two JSON blobs is not a
		// usable failure message.
		t.Errorf("Go feed differs from the Python builder's output for 2026-07-31.\n"+
			"today=%s archive=%d\nFirst difference at byte %d:\n  go:     %s\n  python: %s",
			today.Slug, len(schedule)-1, firstDiff(got, want),
			excerpt(got, firstDiff(got, want)), excerpt(want, firstDiff(got, want)))
	}
}

// The parity test above is only meaningful if the two sides could actually
// disagree. This asserts the comparison has teeth: perturb the input and the
// same comparison must fail.
func TestParityComparisonCanFail(t *testing.T) {
	schedule := loadPoolFixture(t)
	schedule[len(schedule)-1].Headline = "something else entirely"

	feed, _, err := buildProvocationFeed(schedule, day("2026-07-31"), "GOLDEN")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	goldenRaw, _ := os.ReadFile("testdata/provocation_feed_golden.json")
	var golden interface{}
	_ = json.Unmarshal(goldenRaw, &golden)

	if normalise(t, feed) == normalise(t, golden) {
		t.Fatal("a changed headline still matched the golden — the parity test is vacuous")
	}
}

// TestFeedFileKeepsMarkupLiteral pins the ENCODING, which the parity test above
// structurally cannot see.
//
// Both tests read the same nine rows and the same builder. The parity test
// unmarshals before comparing, so `<em>` and `<em>` are the same value
// to it, and it passed for a week while the action shipped the escaped form to
// production. Only a diff of the committed artefact against the oracle's own
// bytes exposed it. So: assert on the bytes, and prove the assertion has teeth
// by checking the encoder we replaced still fails it.
func TestFeedFileKeepsMarkupLiteral(t *testing.T) {
	schedule := loadPoolFixture(t)
	feed, _, err := buildProvocationFeed(schedule, day("2026-07-31"), "GOLDEN")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	payload, err := marshalFeedFile(feed)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(payload)

	// Built by concatenation, never typed as an escape sequence: a literal
	// backslash-u-0-0-3-c in a Go source file is decoded by the compiler inside
	// a double-quoted string, and by more than one editor and tool on the way
	// in. Both decodings turn this needle into the very character it is meant to
	// prove absent, which makes the assertion silently vacuous. It happened
	// while writing this test — the control below is what caught it.
	escLT := `\` + "u003c"
	escAmp := `\` + "u0026"

	if strings.Contains(got, escLT) || strings.Contains(got, escAmp) {
		t.Errorf("feed file contains HTML-escaped markup (%s / %s); headlines carry <em> and must reach the file literally", escLT, escAmp)
	}
	if !strings.Contains(got, "<em>") {
		t.Errorf("feed file contains no literal <em>; the fixture is supposed to carry emphasis markup, so this test is no longer testing anything")
	}

	// The control. Without it this test only says "the code does what it does":
	// the encoder we moved away from must actually fail the assertion above,
	// otherwise the escaping was never happening and the fix is theatre.
	escaped, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		t.Fatalf("control marshal: %v", err)
	}
	if !strings.Contains(string(escaped), escLT) {
		t.Fatal("json.MarshalIndent did not escape the markup, so this test cannot detect the regression it was written for")
	}
}

func firstDiff(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func excerpt(s string, at int) string {
	start := at - 60
	if start < 0 {
		start = 0
	}
	end := at + 120
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}
