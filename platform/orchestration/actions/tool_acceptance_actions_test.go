package actions

import (
	"strings"
	"testing"
)

// The awaited-reply shapes vary across the codebase; extractRunResults must
// find results through every fallback and recompute the verdict itself.
func TestExtractRunResultsFallbackPaths(t *testing.T) {
	results := []interface{}{
		map[string]interface{}{"check_id": "boots", "pass": true, "detail": "ok"},
		map[string]interface{}{"check_id": "rows", "pass": false, "detail": "no match"},
	}
	skipped := []interface{}{
		map[string]interface{}{"check_id": "mobile-fit", "detail": "P0"},
	}

	shapes := map[string]map[string]interface{}{
		"response.data": {"browser_run": map[string]interface{}{
			"response": map[string]interface{}{
				"data": map[string]interface{}{"results": results, "skipped": skipped}}}},
		"response": {"browser_run": map[string]interface{}{
			"response": map[string]interface{}{"results": results, "skipped": skipped}}},
		"flattened": {"browser_run": map[string]interface{}{
			"results": results, "skipped": skipped}},
	}

	for name, collected := range shapes {
		v := extractRunResults(collected, "browser_run")
		if len(v.Results) != 2 {
			t.Errorf("%s: expected 2 results, got %d", name, len(v.Results))
			continue
		}
		if len(v.Passed) != 1 || v.Passed[0] != "boots" {
			t.Errorf("%s: passed wrong: %v", name, v.Passed)
		}
		if len(v.Failed) != 1 || v.Failed[0] != "rows" {
			t.Errorf("%s: failed wrong: %v", name, v.Failed)
		}
		if len(v.Details) != 1 || v.Details[0] != "rows: no match" {
			t.Errorf("%s: details wrong: %v", name, v.Details)
		}
		if len(v.SkipList) != 1 || v.SkipList[0] != "mobile-fit" {
			t.Errorf("%s: skip list wrong: %v", name, v.SkipList)
		}
	}
}

// REGRESSION (live run af5a4ac5, 2026-07-14): a check runs once per profile, so
// a bare check id is not a unique result. The note used to report "1 skipped:
// mobile-fit" when mobile-fit had PASSED on mobile and was skipped only on
// desktop (correctly — it is a mobile-only check). A reader concluded mobile was
// never checked: the opposite of the truth, in the artifact the ladder exists to
// produce. Every result must be labelled id@profile.
func TestExtractRunResultsLabelsByProfile(t *testing.T) {
	// The exact shape the P1/P2 adapter returned for tool-xp-curve-designer.
	results := []interface{}{
		map[string]interface{}{"check_id": "boots", "profile": "desktop", "pass": true, "detail": "1 element(s) match .tool-container"},
		map[string]interface{}{"check_id": "curve-switch", "profile": "desktop", "pass": true, "detail": "interaction produced the expected result (#tableWrap tr)"},
		map[string]interface{}{"check_id": "boots", "profile": "mobile", "pass": true, "detail": "1 element(s) match .tool-container"},
		map[string]interface{}{"check_id": "mobile-fit", "profile": "mobile", "pass": true, "detail": "no horizontal overflow on mobile"},
		map[string]interface{}{"check_id": "curve-switch", "profile": "mobile", "pass": true, "detail": "interaction produced the expected result (#tableWrap tr)"},
	}
	skipped := []interface{}{
		map[string]interface{}{"check_id": "mobile-fit", "profile": "desktop", "detail": "SKIPPED: not run on profile desktop"},
	}
	collected := map[string]interface{}{"browser_run": map[string]interface{}{
		"results": results, "skipped": skipped}}

	v := extractRunResults(collected, "browser_run")

	if len(v.Passed) != 5 {
		t.Fatalf("expected 5 passing instances, got %d (%v)", len(v.Passed), v.Passed)
	}
	// mobile-fit passed on mobile AND appears in the skip list for desktop —
	// both must be visible, and distinguishable.
	if got, want := v.SkipList, []string{"mobile-fit@desktop"}; len(got) != 1 || got[0] != want[0] {
		t.Errorf("skip list must name the PROFILE that skipped: got %v, want %v", got, want)
	}
	if !contains(v.Passed, "mobile-fit@mobile") {
		t.Errorf("mobile-fit passed on mobile and must be reported as such: %v", v.Passed)
	}
	if !contains(v.Passed, "curve-switch@desktop") || !contains(v.Passed, "curve-switch@mobile") {
		t.Errorf("the interaction ran on both profiles; both must be reported: %v", v.Passed)
	}
	if got, want := v.Profiles, []string{"desktop", "mobile"}; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("profiles wrong: got %v, want %v", got, want)
	}
}

// A check failing on ONE profile must label the instance, while the improve_tool
// spec keeps the bare, deduped criteria id (the fixer matches ids to the PLAN).
func TestExtractRunResultsFailedIDsStayBare(t *testing.T) {
	results := []interface{}{
		map[string]interface{}{"check_id": "rows", "profile": "desktop", "pass": true, "detail": "ok"},
		map[string]interface{}{"check_id": "rows", "profile": "mobile", "pass": false, "detail": "0 elements match"},
		map[string]interface{}{"check_id": "fit", "profile": "mobile", "pass": false, "detail": "overflows by 40px"},
	}
	collected := map[string]interface{}{"browser_run": map[string]interface{}{"results": results}}

	v := extractRunResults(collected, "browser_run")

	if got, want := v.Failed, []string{"rows@mobile", "fit@mobile"}; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("failed instances must carry the profile: got %v, want %v", got, want)
	}
	if got, want := v.FailedIDs, []string{"rows", "fit"}; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("failing_checks must stay bare criteria ids: got %v, want %v", got, want)
	}
	if !contains(v.Details, "rows@mobile: 0 elements match") {
		t.Errorf("details must pin the failing profile: %v", v.Details)
	}
	// rows passed on desktop — a mobile-only failure, and the note must say so.
	if !contains(v.Passed, "rows@desktop") {
		t.Errorf("rows passed on desktop; that must survive into the note: %v", v.Passed)
	}
}

// A document-level failure the adapter attributed to SITE CHROME must never be
// counted against the tool: it belongs to the site template (routed to
// component-template-fixer). Live case: vonc.com's div.footer-legal overflowed
// every page — blaming the quiz would have sent the fixer to edit a component
// that cannot reach the footer.
func TestExtractRunResultsSeparatesSiteChromeFromToolFailures(t *testing.T) {
	results := []interface{}{
		map[string]interface{}{"check_id": "boots", "profile": "mobile", "pass": true, "detail": "ok"},
		map[string]interface{}{
			"check_id": "mobile-fit", "profile": "mobile", "pass": false,
			"scope": "chrome", "component": "site-footer",
			"culprit": "div.footer-legal (506px)",
			"url":     "https://vonc.com/tools/x.html",
			"detail":  "page overflows horizontally on mobile; widest offending element: div.footer-legal (506px) — OUTSIDE the tool container: site chrome (site-footer)",
		},
		map[string]interface{}{
			"check_id": "rows", "profile": "mobile", "pass": false,
			"scope": "tool", "detail": "no element matches #tableWrap tr",
		},
	}
	collected := map[string]interface{}{"browser_run": map[string]interface{}{"results": results}}

	v := extractRunResults(collected, "browser_run")

	// The chrome failure must NOT appear among the tool's failures.
	if len(v.Failed) != 1 || v.Failed[0] != "rows@mobile" {
		t.Errorf("only the tool-scoped failure may be blamed on the tool; got %v", v.Failed)
	}
	if contains(v.FailedIDs, "mobile-fit") {
		t.Error("a site-chrome overflow must never reach the improve_tool spec's failing_checks")
	}
	if len(v.Chrome) != 1 {
		t.Fatalf("expected 1 site-chrome failure, got %d", len(v.Chrome))
	}
	c := v.Chrome[0]
	if c.Component != "site-footer" || c.Culprit != "div.footer-legal (506px)" {
		t.Errorf("chrome failure lost its attribution: %+v", c)
	}
	if !strings.Contains(chromeSummary(v.Chrome), "mobile-fit@mobile") {
		t.Errorf("chrome summary must label the instance: %q", chromeSummary(v.Chrome))
	}
}

// An UNATTRIBUTED failure (container not found) must fall back to the tool —
// never be silently written off as somebody else's chrome.
func TestUnattributedOverflowStaysWithTheTool(t *testing.T) {
	results := []interface{}{
		map[string]interface{}{
			"check_id": "mobile-fit", "profile": "mobile", "pass": false,
			"scope": "unknown", "detail": "overflows; attribution unknown",
		},
	}
	collected := map[string]interface{}{"browser_run": map[string]interface{}{"results": results}}

	v := extractRunResults(collected, "browser_run")
	if len(v.Chrome) != 0 {
		t.Error("an unattributed failure must NOT be routed to site chrome on a guess")
	}
	if len(v.Failed) != 1 || v.Failed[0] != "mobile-fit@mobile" {
		t.Errorf("unattributed failure must stay with the tool; got %v", v.Failed)
	}
}

// A pre-P1 adapter reports no profile: labels degrade to bare ids, never "id@".
func TestCheckLabelDegradesWithoutProfile(t *testing.T) {
	if got := checkLabel("boots", ""); got != "boots" {
		t.Errorf("no profile must yield a bare id, got %q", got)
	}
	if got := checkLabel("boots", "mobile"); got != "boots@mobile" {
		t.Errorf("got %q", got)
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func TestExtractRunResultsSkippedRun(t *testing.T) {
	// request_browser_run's no-criteria no-op: {skipped: true, reason: ...}
	collected := map[string]interface{}{
		"browser_run": map[string]interface{}{"skipped": true, "reason": "needs_criteria"},
	}
	v := extractRunResults(collected, "browser_run")
	if !v.Skipped {
		t.Fatal("a skipped run must be recognised, never judged")
	}
	if len(v.Results) != 0 {
		t.Fatalf("skipped run must carry no results, got %d", len(v.Results))
	}
}

func TestExtractRunResultsEmpty(t *testing.T) {
	v := extractRunResults(map[string]interface{}{}, "browser_run")
	if v.Skipped || len(v.Results) != 0 {
		t.Fatalf("missing field must yield empty verdict, got %+v", v)
	}
}

// ── P3: screenshots on failure ──────────────────────────────────────────────

// The adapter attaches screenshots for failing (url, profile) runs; the judge
// must find them through the same response-shape fallbacks as the results.
func TestExtractRunResultsFindsScreenshots(t *testing.T) {
	shots := []interface{}{
		map[string]interface{}{
			"profile": "mobile", "url": "https://x/t.html",
			"uri":            "s3://bucket/acceptance-evidence/site/tool/run_mobile.png",
			"view_url":       "https://signed.example/k?sig=abc",
			"failing_checks": []interface{}{"mobile-fit@mobile"},
		},
	}
	results := []interface{}{
		map[string]interface{}{"check_id": "mobile-fit", "profile": "mobile", "pass": false, "detail": "overflow"},
	}

	shapes := map[string]map[string]interface{}{
		"response.data": {"browser_run": map[string]interface{}{
			"response": map[string]interface{}{
				"data": map[string]interface{}{"results": results, "screenshots": shots}}}},
		"flattened": {"browser_run": map[string]interface{}{
			"results": results, "screenshots": shots}},
	}
	for name, collected := range shapes {
		v := extractRunResults(collected, "browser_run")
		if len(v.Shots) != 1 {
			t.Errorf("%s: expected 1 screenshot ref, got %d", name, len(v.Shots))
			continue
		}
		s := v.Shots[0]
		if s.Profile != "mobile" || s.URI == "" || s.ViewURL == "" || len(s.Failing) != 1 {
			t.Errorf("%s: ref lost fields: %+v", name, s)
		}
	}
}

// A ref without a durable URI is unusable evidence and must be dropped.
func TestExtractRunResultsDropsURIlessScreenshots(t *testing.T) {
	collected := map[string]interface{}{"browser_run": map[string]interface{}{
		"results":     []interface{}{map[string]interface{}{"check_id": "x", "pass": false}},
		"screenshots": []interface{}{map[string]interface{}{"profile": "mobile", "view_url": "https://signed"}},
	}}
	if v := extractRunResults(collected, "browser_run"); len(v.Shots) != 0 {
		t.Errorf("a screenshot without a uri must be dropped, got %+v", v.Shots)
	}
}

// Notes are loaded into LLM prompt contexts by load_doc_context: the evidence
// line carries ONLY the durable s3:// URI — never the presigned signature.
func TestEvidenceLineDurableURIOnly(t *testing.T) {
	shots := []screenshotRef{
		{Profile: "mobile", URI: "s3://b/acceptance-evidence/s/t/r_mobile.png", ViewURL: "https://signed.example/k?sig=SECRETSIG"},
		{Profile: "desktop", URI: "s3://b/acceptance-evidence/s/t/r_desktop.png", ViewURL: "https://signed.example/k2?sig=SECRETSIG2"},
	}
	line := evidenceLine(shots)
	if !strings.Contains(line, "s3://b/acceptance-evidence/s/t/r_mobile.png (mobile)") {
		t.Errorf("evidence line must carry the durable uri + profile: %q", line)
	}
	if strings.Contains(line, "signed.example") || strings.Contains(line, "SECRETSIG") {
		t.Errorf("presigned URLs must NEVER enter a note body: %q", line)
	}
	if evidenceLine(nil) != "" {
		t.Error("no shots → no evidence line (clean pass / screenshots unconfigured)")
	}
}

// Item specs get both forms; a chrome item takes only its own profile's shot.
func TestShotsForSpecProfileFilter(t *testing.T) {
	shots := []screenshotRef{
		{Profile: "mobile", URI: "s3://b/m.png", ViewURL: "https://v/m"},
		{Profile: "desktop", URI: "s3://b/d.png", ViewURL: "https://v/d"},
	}
	all := shotsForSpec(shots, "")
	if len(all) != 2 {
		t.Fatalf("empty profile keeps everything, got %d", len(all))
	}
	if all[0]["uri"] != "s3://b/m.png" || all[0]["view_url"] != "https://v/m" {
		t.Errorf("spec shot lost fields: %+v", all[0])
	}
	mobile := shotsForSpec(shots, "mobile")
	if len(mobile) != 1 || mobile[0]["profile"] != "mobile" {
		t.Errorf("profile filter wrong: %+v", mobile)
	}
	if len(shotsForSpec(nil, "mobile")) != 0 {
		t.Error("no shots → empty spec list")
	}
}
