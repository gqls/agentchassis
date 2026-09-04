// FILE: platform/orchestration/datahelpers/negation_content_test.go
//
// The content map here is the real shape a page-content-writer section returns
// for a call-to-action / listing component: headline + subheadline strings, a
// rich_text body carrying <p> markup, an items list of objects, and the
// non-prose siblings (cta_url, stat values) that must never reach a model.
//
// ⚠ THE FIXTURE'S `name` VALUES ARE PROSE-SHAPED ON PURPOSE, AND THAT IS THE
// WHOLE POINT (bugs_open/420, 2026-09-04). They used to be "orchestrator" and
// "planner" — single tokens with no whitespace — so isProseContentField's VALUE
// test excluded them whatever the field regex said. TestWalkerSkipsNonProse
// therefore asserted `items[N].name` was skipped and could not have failed if
// the field rule were deleted: a passing test vouching for nothing, while the
// header below went on claiming it was mutation-proven. Keep every `name` here
// prose-shaped, or these tests quietly stop testing.
//
// Mutation checks — each is the change to the production code that MUST make the
// named test fail. Re-run them when you touch this file; a guard whose test
// cannot fail is the defect this bug already contains once.
//   - put bare `name` back in neverProseFieldRe      -> TestWalkerSkipsNonProse fails
//   - drop the `_name` SUFFIX arm from that regex    -> TestWalkerSkipsNonProse fails
//   - delete the siblings["url"] test in identityContentField
//                                                    -> TestIdentityNameIsWalkedAndFlagged fails
//   - replace that sibling test with a value/whitespace test
//                                                    -> TestIdentityNameIsWalkedAndFlagged fails
//   - drop `name` from headlineFieldRe               -> TestHeadlineClassification fails
//   - drop the identity key in ScanContentDataForNegation
//                                                    -> TestScanContentDataForNegation fails
//   - drop the map case from walkContentSlice        -> TestWalkerReachesNestedItems fails
//   - pass nil instead of `parent` in walkContentSlice
//                                                    -> TestIdentityAppliesToStringsInsideANamedList fails
//   - return GetMapKeys instead of sortedContentKeys -> TestWalkerOrderIsStable fails (flakily; run -count=20)
//   - drop the Set closure and rebuild the map       -> TestSetWritesThrough fails

package datahelpers

import (
	"regexp"
	"strings"
	"testing"
)

func writerContent() map[string]interface{} {
	return map[string]interface{}{
		"headline":        "The registry shows you what's possible, not what survives production.",
		"subheadline":     "A model directory tells you which agents exist. It doesn't tell you how they hold up.",
		"body":            "<p>We run 1,600 orchestrations a day.</p><p>The list is pulled from the production registry, not from provider marketing pages.</p>",
		"cta_text":        "Book a technical discovery call",
		"cta_url":         "/tools/password-entropy.html",
		"primary_cta_url": "/contact.html",
		"stat_1_value":    "1,600",
		"icon":            "database",
		// Prose-shaped and carrying a shape, so it is ONLY the `_name` suffix arm
		// of neverProseFieldRe keeping it away from the model. Drop that arm and
		// this walks — which is what the mutation check above asserts.
		"company_name": "Acme Compliance Systems, not a reseller",
		"items": []interface{}{
			map[string]interface{}{
				// IDENTITY — a `name` beside a `url` is a listing item's page slug
				// and the stem of that url. Deliberately prose-shaped so the test
				// proves the guarantee comes from the SIBLING, not from the value
				// heuristic that happens to cover real slugs today.
				"name":        "Routes work, not requests",
				"description": "Routes work to the agents that can do it, rather than to whichever is idle.",
				"url":         "/agents/orchestrator.html",
			},
			map[string]interface{}{
				// DISPLAY — a feature card's `name` IS its rendered heading. No url
				// sibling, so it is scannable and repairable. This is bugs_open/420's
				// motivating shape.
				"name":        "A planner that sequences, not schedules",
				"description": "Builds the section plan a page is rendered from.",
			},
		},
		// Two levels down: an object inside a list inside an object inside a list.
		// A shallow walk over top-level arrays misses these, and a census that did
		// exactly that undercounted the live estate by 30 items (components lane,
		// 2026-09-03) — the deeper shapes are the odd ones.
		"sections": []interface{}{
			map[string]interface{}{
				"cards": []interface{}{
					map[string]interface{}{
						"name":        "Exact math, not simulation",
						"description": "Every figure is computed, and the method is stated.",
					},
				},
			},
		},
	}
}

func TestWalkerSkipsNonProse(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range WalkContentStrings(writerContent()) {
		seen[f.Path] = true
	}
	must := []string{
		"headline", "subheadline", "body", "cta_text",
		"items[0].description", "items[1].description",
		// bugs_open/420: a card's `name` is its heading and MUST be scanned.
		// items[0].name is an identity — still walked, so both consumers count
		// it; the repair refuses it via the Identity flag, not by skipping it.
		"items[0].name", "items[1].name",
	}
	for _, m := range must {
		if !seen[m] {
			t.Errorf("prose field %q was not walked", m)
		}
	}
	mustNot := []string{
		"cta_url", "primary_cta_url", "stat_1_value", "icon", "items[0].url",
		// The `_name` suffix arm. These are identifiers with no url sibling, so
		// nothing else would keep them out.
		"company_name",
	}
	for _, m := range mustNot {
		if seen[m] {
			t.Errorf("non-prose field %q was walked — a URL, a token or a stored identifier must never be sent to a model as a sentence", m)
		}
	}
}

// The walker must reach an object nested inside a list inside an object inside a
// list, and must carry the right siblings down with it — the identity rule is
// decided against the NEAREST enclosing object, not the top-level map.
func TestWalkerReachesNestedItems(t *testing.T) {
	var got *NegationTextField
	for _, f := range WalkContentStrings(writerContent()) {
		if f.Path == "sections[0].cards[0].name" {
			field := f
			got = &field
		}
	}
	if got == nil {
		t.Fatal("a name two levels down was not walked at all — a shallow walk over top-level arrays is what undercounted the live census")
	}
	if got.Identity != "" {
		t.Errorf("sections[0].cards[0].name has no url sibling, so it is display copy; got Identity=%q", got.Identity)
	}
}

// The `parent` parameter on walkContentSlice earns its place here and only here:
// a string sitting DIRECTLY in a list has no enclosing object of its own, so the
// identity rule has to be decided against the nearest one above it. No live
// producer emits this shape today (measured 2026-09-03), which is the reason it
// needs a test rather than a census — the day one does, the identity guarantee
// must already hold rather than being noticed afterwards.
func TestIdentityAppliesToStringsInsideANamedList(t *testing.T) {
	content := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"name": []interface{}{"Routes work, not requests"},
				"url":  "/agents/orchestrator.html",
			},
		},
	}
	var found bool
	for _, f := range WalkContentStrings(content) {
		if f.Path != "items[0].name[0]" {
			continue
		}
		found = true
		if f.Identity != IdentityNameWithURLSibling {
			t.Errorf("a string in a `name` list beside a url is still an identity; got Identity=%q", f.Identity)
		}
	}
	if !found {
		t.Fatal("items[0].name[0] was not walked at all")
	}
}

func TestIdentityNameIsWalkedAndFlagged(t *testing.T) {
	byPath := map[string]NegationTextField{}
	for _, f := range WalkContentStrings(writerContent()) {
		byPath[f.Path] = f
	}

	id, ok := byPath["items[0].name"]
	if !ok {
		t.Fatal("items[0].name must still be WALKED — the annotation counts it; only the repair refuses it")
	}
	if id.Identity != IdentityNameWithURLSibling {
		t.Errorf("items[0].name has a url sibling and must be flagged %q, got %q", IdentityNameWithURLSibling, id.Identity)
	}

	disp, ok := byPath["items[1].name"]
	if !ok {
		t.Fatal("items[1].name is a card heading and must be walked")
	}
	if disp.Identity != "" {
		t.Errorf("items[1].name has no url sibling, so a writer may repair it; got Identity=%q", disp.Identity)
	}

	// The guarantee must come from the SIBLING, not from the value. Both names
	// above are prose-shaped, so a value-based rule would flag both or neither —
	// this pair can only be told apart structurally.
	if id.Text == "" || disp.Text == "" || !strings.Contains(id.Text, " ") || !strings.Contains(disp.Text, " ") {
		t.Fatal("fixture regression: both names must be prose-shaped or this test cannot distinguish a sibling rule from a value rule")
	}
}

// CORROBORATION, NOT A DISCRIMINATOR (components lane, 2026-09-03). Fleet-wide
// today, 908/908 identity names are lowercase slugs and 825/825 display names
// contain an uppercase letter — three signals partitioning the same 1,729 items
// identically. The design deliberately uses only the structural one, because
// case is a property of who the current producers are. This test asserts the
// lexical agreement on real live values so that WHEN it breaks we are told, and
// find out that a producer changed rather than inferring it from a silent miss.
func TestLexicalAgreementIsCorroborationOnly(t *testing.T) {
	slug := regexp.MustCompile(`^[a-z0-9-]+$`)
	// Real values pulled from page_components.content_data, 2026-09-03.
	for _, s := range []string{"cruiserweight-boxings-best-kept-secret", "tool-fight-countdown-guide", "advertising-on-a-small-budget"} {
		if !slug.MatchString(s) {
			t.Errorf("identity names are expected to be lowercase slugs today; %q is not — a producer has changed, so re-read IdentityNameWithURLSibling's census before trusting any value-shaped reasoning about them", s)
		}
	}
	for _, s := range []string{"Freedom Health Insurance", "170+ Agents Running in Production", "Darlington Building Society"} {
		if s == strings.ToLower(s) {
			t.Errorf("display names are expected to carry uppercase today; %q does not", s)
		}
	}
	// The load-bearing half: the structural rule must not depend on either.
	lower := map[string]interface{}{"name": "routes work, not requests", "url": "/x.html"}
	if identityContentField("name", lower) != IdentityNameWithURLSibling {
		t.Error("a lowercase identity name must still be flagged — the rule is the url sibling, not the case")
	}
	upper := map[string]interface{}{"name": "Routes Work, Not Requests"}
	if identityContentField("name", upper) != "" {
		t.Error("an uppercase display name must stay repairable — the rule is the url sibling, not the case")
	}
}

// An empty url is still a listing item: the predicate tests key PRESENCE because
// its job is to fail toward exclusion. Measured 2026-09-03 the state does not
// occur live, which is exactly why it needs a test rather than a census.
func TestIdentityUsesKeyPresenceNotValue(t *testing.T) {
	for _, siblings := range []map[string]interface{}{
		{"name": "Routes work, not requests", "url": ""},
		{"name": "Routes work, not requests", "url": nil},
	} {
		if got := identityContentField("name", siblings); got != IdentityNameWithURLSibling {
			t.Errorf("a name beside an empty/null url is still a listing item; got %q", got)
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
	yes := []string{"headline", "subheadline", "sub_headline", "title", "eyebrow", "tagline", "items[2].heading",
		// bugs_open/420: a card's name is a heading surface.
		"items[2].name", "features[0].name"}
	for _, y := range yes {
		if !IsHeadlineField(y) {
			t.Errorf("%q should be headline-class", y)
		}
	}
	for _, no := range []string{"body", "cta_text", "items[0].description", "intro", "summary", "company_name"} {
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
	byField := map[string]map[string]interface{}{}
	headlineFlagged := false
	for _, f := range findings {
		byField[f["field"].(string)] = f
		if f["field"] == "headline" && f["headline"] == true {
			headlineFlagged = true
		}
	}
	for _, must := range []string{"headline", "subheadline", "body", "items[0].description",
		// The annotation reports BOTH names. It is read-only, so it has no reason
		// to hide either — and hiding one is what would make its count stop
		// reconciling with the repair's.
		"items[0].name", "items[1].name"} {
		if byField[must] == nil {
			t.Errorf("expected a finding on %q", must)
		}
	}
	if !headlineFlagged {
		t.Error("a headline hit must be marked as headline-class — that is what makes it repairable regardless of budget")
	}
	if got := byField["items[0].name"]["identity"]; got != IdentityNameWithURLSibling {
		t.Errorf("an identity finding must say so, for the reader who wonders why it was never repaired; got %v", got)
	}
	if _, present := byField["items[1].name"]["identity"]; present {
		t.Error("ordinary copy must carry no identity key — a finding on repairable copy stays byte-identical to what this returned before 420")
	}
	if _, present := byField["headline"]["identity"]; present {
		t.Error("ordinary copy must carry no identity key")
	}
	// Clean content yields nil, so the renderer attaches nothing.
	clean := map[string]interface{}{"headline": "Every agent definition running in production", "body": "<p>We run 1,600 orchestrations a day across 13 live systems.</p>"}
	if got := ScanContentDataForNegation(clean); got != nil {
		t.Errorf("clean content must produce no findings, got %v", got)
	}
}
