// FILE: platform/orchestration/datahelpers/meta_description_test.go
//
// bugs_open/103. The strings below are not invented: the internal ones are
// taken from the live census that found the defect, so the guard is tested
// against what actually leaked rather than against what a leak might look like.
package datahelpers

import "testing"

// arenaBrief is the real thing — the first 200-odd characters of the 1,206
// published live at https://vonc.com/tools/arena/index.html as that page's
// meta description, i.e. the text Google was given.
const arenaBrief = "The Arena is Spark's competitive mode, v1 as a fully self-contained " +
	"client-side experience (no fetch calls, no backend). Four elements, in order: " +
	"(1) TODAY'S PROVOCATION — a bold prompt displayed prominently at the top " +
	"(embed 5 sample provocations in JS and pick one by day-of-date so the page " +
	"changes daily, e.g. …"

func TestMetaDescriptionLooksInternal(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		// --- the ones that leaked ---
		{"the live Arena brief", arenaBrief, true},
		{"short brief, marker only", "Fully self-contained client-side widget, no fetch calls.", true},
		{"numbered build steps", "Build it in order: (1) header, (2) form, (3) results table.", true},
		{"long spec under no marker", stringOfLength(400), true},

		// --- real copy that must survive ---
		{"the hand-fixed Arena copy, live now", "Read today's provocation, browse the archive, " +
			"then take a position into the Gauntlet and defend it against an AI opponent " +
			"on a twenty-minute clock.", false},
		{"the guide-page composed line", "A practical guide to LLM Cost Calculator: what it " +
			"means, how it works, and how to use the interactive llm cost calculator.", false},
		{"the tool composed line, house voice", "An interactive Fuel Cost Estimator, free to run " +
			"in the browser. The companion guide sets out the method behind it, so you can " +
			"check the working.", false},
		{"ordinary marketing copy", "Compare wholesale fuel prices across UK suppliers and " +
			"forecast your monthly spend.", false},
		{"exactly at the length limit", stringOfLength(maxPublicMetaDescription), false},
		{"empty is a different problem", "", false},
	}
	for _, c := range cases {
		if got := MetaDescriptionLooksInternal(c.in); got != c.want {
			t.Errorf("%s: MetaDescriptionLooksInternal() = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestPublicMetaDescriptionPrefersGoodCandidate: a tool whose description is
// genuinely public copy keeps it. The guard must not replace good text.
func TestPublicMetaDescriptionPrefersGoodCandidate(t *testing.T) {
	good := "Estimate the return on an AI agent deployment in under a minute."
	got, replaced := PublicMetaDescription(good, "composed fallback that is long enough to be real copy")
	if got != good {
		t.Errorf("a publishable candidate was replaced: got %q", got)
	}
	if replaced {
		t.Error("replaced=true for a candidate that was kept — the caller would log a rejection that did not happen")
	}
}

// TestPublicMetaDescriptionRefusesTheBrief is the regression: the exact live
// value must not be returned, and the composed fallback must be.
func TestPublicMetaDescriptionRefusesTheBrief(t *testing.T) {
	composed := "Use our free interactive arena in your browser, with a companion guide " +
		"explaining how it works."
	got, replaced := PublicMetaDescription(arenaBrief, composed)
	if got == arenaBrief {
		t.Fatal("the build brief was published — this is the defect bugs_open/103 describes")
	}
	if got != composed {
		t.Errorf("expected the composed fallback, got %q", got)
	}
	if !replaced {
		t.Error("replaced=false after rejecting a brief — the swap would be silent, which is " +
			"the objection the second return value exists to answer")
	}
}

// TestPublicMetaDescriptionRefusesABriefShapedFallback: if the CALLER's composed
// text is itself brief-shaped, returning it would quietly reintroduce the defect
// through the escape hatch. Empty is the correct answer — a page with no meta
// description is a worse SEO outcome than a good one and a better one than a
// published spec.
func TestPublicMetaDescriptionRefusesABriefShapedFallback(t *testing.T) {
	got, replaced := PublicMetaDescription(arenaBrief, stringOfLength(500))
	if got != "" {
		t.Errorf("a brief-shaped fallback was published: got %d chars", len(got))
	}
	if !replaced {
		t.Error("replaced=false though the candidate was rejected")
	}
}

// TestPublicMetaDescriptionFallsBackWhenCandidateEmpty covers the ordinary case
// of a tool with no description at all.
func TestPublicMetaDescriptionFallsBackWhenCandidateEmpty(t *testing.T) {
	composed := "Use our free interactive cost calculator in your browser."
	got, replaced := PublicMetaDescription("   ", composed)
	if got != composed {
		t.Errorf("got %q, want the composed fallback", got)
	}
	if replaced {
		t.Error("replaced=true for an ABSENT candidate — nothing was rejected, so a " +
			"caller must not log a rejection")
	}
}

// stringOfLength builds prose-shaped filler of exactly n runes, so a length
// assertion is not accidentally also a marker assertion.
func stringOfLength(n int) string {
	const word = "value "
	out := make([]rune, 0, n)
	for len(out) < n {
		for _, r := range word {
			if len(out) == n {
				break
			}
			out = append(out, r)
		}
	}
	return string(out)
}

// TestMetaDescriptionKnownGap_ShortPhraseFreeBrief documents, rather than
// asserts, the hole the council's editquality seat identified: the gate is a
// heuristic (length + measured phrases), so a brief that is BOTH under the
// length limit AND free of every marker is published.
//
// This is not a failing test, because the behaviour is the accepted trade — the
// bug file's candidate 1 (a distinct public-copy field) is what makes it
// structurally impossible, and that is a schema change owned elsewhere. It is
// here so the gap is visible in the test file rather than only in a risks
// paragraph nobody re-reads.
func TestMetaDescriptionKnownGap_ShortPhraseFreeBrief(t *testing.T) {
	shortBrief := "Renders a slider bound to the volume field and recomputes on input."
	if MetaDescriptionLooksInternal(shortBrief) {
		t.Fatal("this test is stale in a GOOD way: the gate now catches a short, " +
			"phrase-free brief. Tighten the documented gap or delete this test.")
	}
	t.Log("known gap (bugs_open/103): a brief under 320 runes with no measured marker " +
		"is published. Closing it needs candidate 1, a distinct public-copy field.")
}
