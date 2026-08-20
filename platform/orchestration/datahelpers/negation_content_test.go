// FILE: platform/orchestration/datahelpers/negation_content_test.go
//
// The content map here is the real shape a page-content-writer section returns
// for a call-to-action / listing component: headline + subheadline strings, a
// rich_text body carrying <p> markup, an items list of objects, and the
// non-prose siblings (cta_url, stat values) that must never reach a model.
//
// Mutation checks (by hand, recorded in the lane NOTES):
//   - drop the nonProseFieldRe test in prosey()  -> TestWalkerSkipsNonProse fails
//   - return GetMapKeys instead of sortedContentKeys -> TestWalkerOrderIsStable fails (flakily; run -count=20)
//   - drop the Set closure and rebuild the map    -> TestSetWritesThrough fails

package datahelpers

import "testing"

func writerContent() map[string]interface{} {
	return map[string]interface{}{
		"headline":         "The registry shows you what's possible, not what survives production.",
		"subheadline":      "A model directory tells you which agents exist. It doesn't tell you how they hold up.",
		"body":             "<p>We run 1,600 orchestrations a day.</p><p>The list is pulled from the production registry, not from provider marketing pages.</p>",
		"cta_text":         "Book a technical discovery call",
		"cta_url":          "/tools/password-entropy.html",
		"primary_cta_url":  "/contact.html",
		"stat_1_value":     "1,600",
		"icon":             "database",
		"items": []interface{}{
			map[string]interface{}{
				"name":        "orchestrator",
				"description": "Routes work to the agents that can do it, rather than to whichever is idle.",
				"url":         "/agents/orchestrator.html",
			},
			map[string]interface{}{
				"name":        "planner",
				"description": "Builds the section plan a page is rendered from.",
			},
		},
	}
}

func TestWalkerSkipsNonProse(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range WalkContentStrings(writerContent()) {
		seen[f.Path] = true
	}
	for _, must := range []string{"headline", "subheadline", "body", "cta_text", "items[0].description", "items[1].description"} {
		if !seen[must] {
			t.Errorf("prose field %q was not walked", must)
		}
	}
	for _, mustNot := range []string{"cta_url", "primary_cta_url", "stat_1_value", "icon", "items[0].url", "items[0].name", "items[1].name"} {
		if seen[mustNot] {
			t.Errorf("non-prose field %q was walked — a URL or a token must never be sent to a model as a sentence", mustNot)
		}
	}
}

func TestWalkerOrderIsStable(t *testing.T) {
	first := ""
	for i := 0; i < 20; i++ {
		var paths string
		for _, f := range WalkContentStrings(writerContent()) {
			paths += f.Path + "|"
		}
		if i == 0 {
			first = paths
			continue
		}
		if paths != first {
			t.Fatalf("walk order is unstable:\n%s\n%s", first, paths)
		}
	}
}

// The renderer reads the very map the walker was handed, so a write must land
// there — not in a rebuilt copy the renderer will never see.
func TestSetWritesThrough(t *testing.T) {
	content := writerContent()
	for _, f := range WalkContentStrings(content) {
		if f.Path == "headline" {
			f.Set("The registry lists every agent definition running in production today.")
		}
		if f.Path == "items[0].description" {
			f.Set("Routes work to the agents that can do it.")
		}
	}
	if got := content["headline"].(string); got != "The registry lists every agent definition running in production today." {
		t.Errorf("headline was not written through: %q", got)
	}
	items := content["items"].([]interface{})
	if got := items[0].(map[string]interface{})["description"].(string); got != "Routes work to the agents that can do it." {
		t.Errorf("list-item description was not written through: %q", got)
	}
}

func TestHeadlineClassification(t *testing.T) {
	for _, yes := range []string{"headline", "subheadline", "sub_headline", "title", "eyebrow", "tagline", "items[2].heading"} {
		if !IsHeadlineField(yes) {
			t.Errorf("%q should be headline-class", yes)
		}
	}
	for _, no := range []string{"body", "cta_text", "items[0].description", "intro", "summary"} {
		if IsHeadlineField(no) {
			t.Errorf("%q should NOT be headline-class", no)
		}
	}
}

func TestScanContentDataForNegation(t *testing.T) {
	findings := ScanContentDataForNegation(writerContent())
	if len(findings) == 0 {
		t.Fatal("expected findings on a content map built from live copy")
	}
	byField := map[string]bool{}
	headlineFlagged := false
	for _, f := range findings {
		byField[f["field"].(string)] = true
		if f["field"] == "headline" && f["headline"] == true {
			headlineFlagged = true
		}
	}
	for _, must := range []string{"headline", "subheadline", "body", "items[0].description"} {
		if !byField[must] {
			t.Errorf("expected a finding on %q, got %v", must, byField)
		}
	}
	if !headlineFlagged {
		t.Error("a headline hit must be marked as headline-class — that is what makes it repairable regardless of budget")
	}
	// Clean content yields nil, so the renderer attaches nothing.
	clean := map[string]interface{}{"headline": "Every agent definition running in production", "body": "<p>We run 1,600 orchestrations a day across 13 live systems.</p>"}
	if got := ScanContentDataForNegation(clean); got != nil {
		t.Errorf("clean content must produce no findings, got %v", got)
	}
}
