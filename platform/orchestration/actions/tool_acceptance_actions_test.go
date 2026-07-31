package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
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

// bugs_open/010: the adapter's drill-down attribution must reach the fix ticket.
// For a TOOL-scoped overflow the forcing element + reason ride on the verdict so
// the improve_tool spec can point the fixer past the ancestor it kept "fixing".
func TestExtractRunResultsCapturesToolDrillDown(t *testing.T) {
	results := []interface{}{
		map[string]interface{}{
			"check_id": "mobile-fit", "profile": "mobile", "pass": false, "scope": "tool",
			"detail":        "page overflows on mobile; widest offending element: fieldset (419px) — inside the tool; the width is forced by div.ltb-row-grid [grid layout]",
			"forced_by":     "div.ltb-row-grid",
			"forced_reason": "grid layout — set min-width:0 on the items or let the grid wrap",
		},
	}
	collected := map[string]interface{}{"browser_run": map[string]interface{}{"results": results}}

	v := extractRunResults(collected, "browser_run")
	if v.ForcedBy != "div.ltb-row-grid" {
		t.Errorf("verdict must carry the forcing element, got %q", v.ForcedBy)
	}
	if v.ForcedReason == "" {
		t.Errorf("verdict must carry the fix hint")
	}
	// Still a tool failure (not chrome), so it goes to improve_tool.
	if len(v.Failed) != 1 || len(v.Chrome) != 0 {
		t.Errorf("tool-scoped overflow must stay with the tool; got failed=%v chrome=%d", v.Failed, len(v.Chrome))
	}
}

// A site-chrome overflow carries the drill-down too, so component-template-fixer
// gets pointed at the forcing descendant, and the suggestion says so.
func TestChromeFailureCarriesDrillDown(t *testing.T) {
	results := []interface{}{
		map[string]interface{}{
			"check_id": "mobile-fit", "profile": "mobile", "pass": false, "scope": "chrome",
			"component": "site-footer", "culprit": "div.footer-legal (506px)",
			"culprit_selector": "div.footer-legal", "slot": "footer",
			"forced_by":     "ul.footer-links",
			"forced_reason": "flex row does not wrap (flex-wrap:nowrap) — allow wrapping or set min-width:0 on the items",
			"detail":        "…",
		},
	}
	collected := map[string]interface{}{"browser_run": map[string]interface{}{"results": results}}
	v := extractRunResults(collected, "browser_run")
	if len(v.Chrome) != 1 {
		t.Fatalf("expected 1 chrome failure, got %d", len(v.Chrome))
	}
	c := v.Chrome[0]
	if c.ForcedBy != "ul.footer-links" || c.ForcedReason == "" {
		t.Errorf("chrome failure lost its drill-down: %+v", c)
	}
	hint := chromeForcedHint(c)
	if !strings.Contains(hint, "ul.footer-links") || !strings.Contains(hint, "fix THAT element") {
		t.Errorf("chrome suggestion must point at the forcing element: %q", hint)
	}
	if chromeForcedHint(chromeFailure{}) != "" {
		t.Errorf("no drill-down → empty hint (no clutter)")
	}
}

// ── TL-035: renders — a look that does not need a failure to justify it ──────

// browserRunParams builds the minimum ActionParams for RequestBrowserRunAction
// with NO database: url_field resolves the page from collected data, which is
// what keeps the page lookup (and its uuid casts) out of a unit test.
func browserRunParams(producer *capturingProducer, config map[string]interface{}) ActionParams {
	config["url_field"] = "input_data.url"
	return ActionParams{
		Logger:     zap.NewNop(),
		Producer:   producer,
		StepConfig: models.Step{Config: config},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{
				"function": "tool-review-council-simulator",
				"url":      "https://fundamentallyai.com/tools/review-council-simulator.html",
			},
			"doc_context": map[string]interface{}{
				"criteria_json": `{"checks":[{"id":"boots","type":"exists","selector":".tool-container"}]}`,
			},
			"site_record": map[string]interface{}{
				"site_id": "11111111-1111-1111-1111-111111111111",
				"domain":  "fundamentallyai.com",
			},
		},
		ExecutionContext: &types.ExecutionContext{
			CorrelationID:   "corr-1",
			OrchestrationID: "orch-1",
			ClientID:        "client-1",
			ResponsesTopic:  "system.agent.test.responses",
			Sender:          types.AgentIdentity{AgentType: "tool-acceptance-agent"},
		},
	}
}

// The opt-in must actually reach the adapter, spelled the way the adapter's
// json tag spells it. A config key the payload never carries is the exact shape
// of a mechanism that is live, configured, and inert.
func TestRequestBrowserRunPayloadCarriesCaptureRenders(t *testing.T) {
	producer := &capturingProducer{}
	params := browserRunParams(producer, map[string]interface{}{"capture_renders": true})

	if _, err := RequestBrowserRunAction(context.Background(), params); err != nil {
		t.Fatalf("RequestBrowserRunAction returned error: %v", err)
	}
	data := producedSearchData(t, producer)
	if data["capture_renders"] != true {
		t.Fatalf("payload capture_renders = %v, want true — the opt-in never reached the adapter, so a passing page is still never photographed", data["capture_renders"])
	}
}

// Default OFF, and PRESENT: an absent key and an explicit false are the same to
// the adapter, but only the explicit form makes the setting readable in a
// captured payload when someone is asking why no render appeared.
func TestRequestBrowserRunCaptureRendersDefaultsOff(t *testing.T) {
	producer := &capturingProducer{}
	params := browserRunParams(producer, map[string]interface{}{})

	if _, err := RequestBrowserRunAction(context.Background(), params); err != nil {
		t.Fatalf("RequestBrowserRunAction returned error: %v", err)
	}
	data := producedSearchData(t, producer)
	v, present := data["capture_renders"]
	if !present {
		t.Fatal("capture_renders must be present in the payload even when off")
	}
	if v != false {
		t.Fatalf("capture_renders = %v with no config — every existing step config must keep its exact prior behaviour", v)
	}
}

// Renders arrive in their own key and must be found through the same envelope
// fallbacks as everything else in the reply.
func TestExtractRunResultsFindsRenders(t *testing.T) {
	renders := []interface{}{
		map[string]interface{}{
			"profile": "desktop", "url": "https://x/t.html",
			"uri":            "s3://bucket/acceptance-evidence/site/tool/run_desktop.png",
			"view_url":       "https://signed.example/k?sig=abc",
			"failing_checks": []interface{}{},
		},
	}
	results := []interface{}{
		map[string]interface{}{"check_id": "boots", "profile": "desktop", "pass": true, "detail": "ok"},
	}
	shapes := map[string]map[string]interface{}{
		"response.data": {"browser_run": map[string]interface{}{
			"response": map[string]interface{}{
				"data": map[string]interface{}{"results": results, "renders": renders}}}},
		"flattened": {"browser_run": map[string]interface{}{
			"results": results, "renders": renders}},
	}
	for name, collected := range shapes {
		v := extractRunResults(collected, "browser_run")
		if len(v.Renders) != 1 {
			t.Errorf("%s: expected 1 render ref, got %d", name, len(v.Renders))
			continue
		}
		if v.Renders[0].Profile != "desktop" || v.Renders[0].URI == "" {
			t.Errorf("%s: render ref lost fields: %+v", v.Renders[0], name)
		}
	}
}

// THE load-bearing property of the two-list design (see captureEvidence's own
// comment): three consumers attach Screenshots to a work item BECAUSE something
// failed, two of them unfiltered. A render leaking into Shots would put a
// photograph of a perfectly good page into a failure ticket as its evidence.
func TestRendersNeverEnterTheScreenshotList(t *testing.T) {
	collected := map[string]interface{}{"browser_run": map[string]interface{}{
		"results": []interface{}{
			map[string]interface{}{"check_id": "mobile-fit", "profile": "mobile", "pass": false, "detail": "overflow"},
			map[string]interface{}{"check_id": "boots", "profile": "desktop", "pass": true, "detail": "ok"},
		},
		"screenshots": []interface{}{map[string]interface{}{
			"profile": "mobile", "uri": "s3://b/fail_mobile.png",
			"failing_checks": []interface{}{"mobile-fit@mobile"},
		}},
		"renders": []interface{}{map[string]interface{}{
			"profile": "desktop", "uri": "s3://b/pass_desktop.png",
		}},
	}}
	v := extractRunResults(collected, "browser_run")
	if len(v.Shots) != 1 || v.Shots[0].URI != "s3://b/fail_mobile.png" {
		t.Fatalf("Shots must hold the failing profile ONLY: %+v", v.Shots)
	}
	if len(v.Renders) != 1 || v.Renders[0].URI != "s3://b/pass_desktop.png" {
		t.Fatalf("Renders must hold the passing profile ONLY: %+v", v.Renders)
	}
	// And the mixed run is the reason the FAILED note renders both lines.
	body := evidenceLine(v.Shots) + renderLine(v.Renders)
	if !strings.Contains(body, "fail_mobile.png") || !strings.Contains(body, "pass_desktop.png") {
		t.Errorf("a mixed run must report both the failure shot and the passing render: %q", body)
	}
}

// Same rule as evidenceLine: a note body is loaded into LLM prompt contexts, so
// only the durable s3:// uri may appear. And the line must not call a render
// "evidence" — every render is a photograph of a run that PASSED.
func TestRenderLineDurableURIOnlyAndClaimsNothing(t *testing.T) {
	renders := []screenshotRef{
		{Profile: "desktop", URI: "s3://b/acceptance-evidence/s/t/r_desktop.png", ViewURL: "https://signed.example/k?sig=SECRETSIG"},
	}
	line := renderLine(renders)
	if !strings.Contains(line, "s3://b/acceptance-evidence/s/t/r_desktop.png (desktop)") {
		t.Errorf("render line must carry the durable uri + profile: %q", line)
	}
	if strings.Contains(line, "signed.example") || strings.Contains(line, "SECRETSIG") {
		t.Errorf("presigned URLs must NEVER enter a note body: %q", line)
	}
	// Match the LABEL, not the whole line: the durable key prefix is literally
	// "acceptance-evidence/", so a bare substring search for "evidence" matches
	// the URI and fails a correct line. (It did, on this test's first run.)
	if !strings.HasPrefix(strings.TrimPrefix(line, "\n"), "Rendered:") {
		t.Errorf("the line must be labelled Rendered, not Evidence: %q", line)
	}
	if strings.Contains(line, "Evidence:") || strings.Contains(line, "at failure") {
		t.Errorf("a render is not evidence of a failure and must not say it is: %q", line)
	}
	if !strings.Contains(line, "not a verdict") {
		t.Errorf("the line must say plainly that a render decides nothing: %q", line)
	}
	if renderLine(nil) != "" {
		t.Error("no renders → no line (the default-off case must add nothing to the note)")
	}
}
