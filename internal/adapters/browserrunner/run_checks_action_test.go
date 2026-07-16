package browserrunner

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// fakePage implements browserPage for tests: canned status/console/overflow,
// configured selector counts, and a record of interaction steps. Selectors not
// primed count as 0 (absent), matching real Locator.Count.
type fakePage struct {
	status   int
	navErr   string
	console  []string
	counts   map[string]int
	overflow       bool
	ovInfo         overflowInfo // culprit + attribution
	ovErr          error
	containerAsked string // the container selector the check resolved to
	texts    map[string]string
	stepErr  map[string]error // selector -> error to return from Do
	steps    []criteriaStep   // recorded
	shotErr   error // error to return from Screenshot
	shotTaken bool  // recorded: Screenshot was called
}

func (f *fakePage) Status() int            { return f.status }
func (f *fakePage) NavError() string       { return f.navErr }
func (f *fakePage) ConsoleErrors() []string { return f.console }
func (f *fakePage) Count(sel string) int   { return f.counts[sel] }
func (f *fakePage) HorizontalOverflow(container string) (bool, overflowInfo, error) {
	f.containerAsked = container
	return f.overflow, f.ovInfo, f.ovErr
}
func (f *fakePage) Text(sel string) (string, error)   { return f.texts[sel], nil }
func (f *fakePage) Close()                            {}
func (f *fakePage) Do(step criteriaStep) error {
	f.steps = append(f.steps, step)
	if f.stepErr != nil {
		if err, ok := f.stepErr[step.Selector]; ok {
			return err
		}
	}
	return nil
}
func (f *fakePage) Screenshot(fullPage bool) ([]byte, error) {
	f.shotTaken = true
	if f.shotErr != nil {
		return nil, f.shotErr
	}
	return []byte("png-bytes"), nil
}

// fakeStore records screenshot saves; err makes every Save fail.
type fakeStore struct {
	keys []string
	err  error
}

func (s *fakeStore) Save(_ context.Context, key string, png []byte) (string, string, error) {
	if s.err != nil {
		return "", "", s.err
	}
	s.keys = append(s.keys, key)
	return "s3://test-bucket/" + key, "https://signed.example/" + key, nil
}

func actionWith(pages map[string]*fakePage) *RunChecksAction {
	return &RunChecksAction{
		logger: zap.NewNop(),
		open: func(_ context.Context, _ /*url*/ string, profile string, _ *zap.Logger) (browserPage, error) {
			if p, ok := pages[profile]; ok {
				return p, nil
			}
			return &fakePage{status: 200}, nil
		},
	}
}

func resultByID(rs []CheckResult, id, profile string) *CheckResult {
	for i := range rs {
		if rs[i].CheckID == id && rs[i].Profile == profile {
			return &rs[i]
		}
	}
	return nil
}

const criteriaDesktopMobile = `{
  "profiles": ["desktop","mobile"],
  "checks": [
    {"id":"boots","type":"selector_exists","selector":".tool-container"},
    {"id":"console","type":"no_console_errors"},
    {"id":"status","type":"page_status_ok"},
    {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]},
    {"id":"rows","type":"selector_exists","selector":"#tableWrap tr"},
    {"id":"later-EDIT","type":"selector_exists","selector":"#placeholder"},
    {"id":"calc","type":"interaction",
      "steps":[{"action":"fill","selector":"#hours","value":"10"},{"action":"click","selector":"#go"}],
      "expect":{"selector":"#result","text_matches":"\\d"}}
  ]}`

func TestP0DesktopHealthy(t *testing.T) {
	desktop := &fakePage{status: 200, counts: map[string]int{".tool-container": 1, "#tableWrap tr": 26, "#result": 1}, texts: map[string]string{"#result": "42"}}
	a := actionWith(map[string]*fakePage{"desktop": desktop})
	out, err := a.Execute(context.Background(), RunChecksRequest{
		RunID: "d", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"boots", "console", "status", "rows", "calc"} {
		if r := resultByID(out.Results, id, "desktop"); r == nil || !r.Pass {
			t.Errorf("%s should pass on desktop, got %+v", id, r)
		}
	}
	// mobile-fit is mobile-only, later-EDIT is skipped → both skipped on desktop
	if out.Summary.Skipped != 2 {
		t.Errorf("expected 2 skipped on desktop, got %d: %+v", out.Summary.Skipped, out.Skipped)
	}
	for _, s := range out.Skipped {
		if s.Pass {
			t.Errorf("a skipped check must never read as pass: %+v", s)
		}
	}
}

func TestP1MobileProfileRunsOverflowAndSkipsDesktopOnly(t *testing.T) {
	base := map[string]int{".tool-container": 1, "#tableWrap tr": 5, "#result": 1}
	desktop := &fakePage{status: 200, counts: base, texts: map[string]string{"#result": "7"}}
	mobile := &fakePage{status: 200, counts: base, overflow: false, texts: map[string]string{"#result": "7"}}
	a := actionWith(map[string]*fakePage{"desktop": desktop, "mobile": mobile})

	out, err := a.Execute(context.Background(), RunChecksRequest{
		RunID: "dm", URLs: []string{"u"}, Profiles: []string{"desktop", "mobile"}, CriteriaJSON: criteriaDesktopMobile,
	})
	if err != nil {
		t.Fatal(err)
	}
	// mobile-fit evaluates on mobile (pass, no overflow), skips on desktop.
	if r := resultByID(out.Results, "mobile-fit", "mobile"); r == nil || !r.Pass {
		t.Errorf("mobile-fit should pass on mobile, got %+v", r)
	}
	if r := resultByID(out.Results, "mobile-fit", "desktop"); r != nil {
		t.Errorf("mobile-fit must NOT be a desktop result, got %+v", r)
	}
	// boots runs on BOTH profiles.
	if resultByID(out.Results, "boots", "desktop") == nil || resultByID(out.Results, "boots", "mobile") == nil {
		t.Error("boots should be evaluated on both profiles")
	}
}

func TestP1MobileOverflowFails(t *testing.T) {
	mobile := &fakePage{status: 200, counts: map[string]int{".tool-container": 1, "#tableWrap tr": 1, "#result": 1}, overflow: true, texts: map[string]string{"#result": "1"}}
	a := actionWith(map[string]*fakePage{"mobile": mobile})
	out, _ := a.Execute(context.Background(), RunChecksRequest{
		RunID: "m", URLs: []string{"u"}, Profiles: []string{"mobile"}, CriteriaJSON: criteriaDesktopMobile,
	})
	if r := resultByID(out.Results, "mobile-fit", "mobile"); r == nil || r.Pass {
		t.Errorf("mobile-fit should FAIL when the page overflows, got %+v", r)
	}
}

// The overflow is measured on the DOCUMENT, but the run is scoped to ONE tool.
// Each failure must therefore say WHERE the offender lives, or a site footer
// raises an unfixable tool ticket for every tool on the site — the live vonc.com
// case (div.footer-legal, 506px at 390px, on every page).
func TestP1OverflowAttributesTheCulprit(t *testing.T) {
	cases := []struct {
		name      string
		info      overflowInfo
		wantScope string
		wantIn    string // substring the detail must carry
	}{
		{
			name:      "site chrome — must NOT be blamed on the tool",
			info:      overflowInfo{Culprit: "div.footer-legal (506px)", Component: "site-footer", Located: true, InTool: false},
			wantScope: ScopeChrome,
			wantIn:    "OUTSIDE the tool container",
		},
		{
			name:      "the tool's own element — a real tool bug",
			info:      overflowInfo{Culprit: "div.quiz-option-btn (420px)", Component: "tool-quiz-section", Located: true, InTool: true},
			wantScope: ScopeTool,
			wantIn:    "inside the tool",
		},
		{
			name:      "container not found — never guess 'chrome'",
			info:      overflowInfo{Culprit: "div.mystery (500px)", Located: false},
			wantScope: ScopeUnknown,
			wantIn:    "attribution unknown",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mobile := &fakePage{
				status:   200,
				counts:   map[string]int{".tool-container": 1, "#tableWrap tr": 1, "#result": 1},
				overflow: true,
				ovInfo:   tc.info,
				texts:    map[string]string{"#result": "1"},
			}
			a := actionWith(map[string]*fakePage{"mobile": mobile})
			out, _ := a.Execute(context.Background(), RunChecksRequest{
				RunID: "m", URLs: []string{"u"}, Profiles: []string{"mobile"}, CriteriaJSON: criteriaDesktopMobile,
			})
			r := resultByID(out.Results, "mobile-fit", "mobile")
			if r == nil || r.Pass {
				t.Fatalf("mobile-fit should FAIL, got %+v", r)
			}
			if r.Scope != tc.wantScope {
				t.Errorf("scope = %q, want %q (detail: %s)", r.Scope, tc.wantScope, r.Detail)
			}
			if !strings.Contains(r.Detail, tc.info.Culprit) {
				t.Errorf("detail must name the offender %q; got %q", tc.info.Culprit, r.Detail)
			}
			if !strings.Contains(r.Detail, tc.wantIn) {
				t.Errorf("detail must say %q; got %q", tc.wantIn, r.Detail)
			}
			if r.Culprit != tc.info.Culprit {
				t.Errorf("culprit not carried to the judge: %q", r.Culprit)
			}
		})
	}
}

// The container selector is resolved per-check → per-document → convention. The
// convention must match BOTH delivery paths, since PLANs predate attribution.
func TestToolContainerResolution(t *testing.T) {
	doc := criteriaDoc{Container: ".doc-level"}
	if got := toolContainer(doc, criteriaCheck{Container: ".per-check"}); got != ".per-check" {
		t.Errorf("per-check container must win, got %q", got)
	}
	if got := toolContainer(doc, criteriaCheck{}); got != ".doc-level" {
		t.Errorf("document container must apply, got %q", got)
	}
	got := toolContainer(criteriaDoc{}, criteriaCheck{})
	if !strings.Contains(got, ".tool-container") || !strings.Contains(got, "-section") {
		t.Errorf("the fallback must cover both the generator (.tool-container) and page-section conventions, got %q", got)
	}
}

func TestP2InteractionPassAndFail(t *testing.T) {
	// Pass: steps run, #result exists and matches \d.
	ok := &fakePage{status: 200, counts: map[string]int{".tool-container": 1, "#tableWrap tr": 1, "#result": 1}, texts: map[string]string{"#result": "42"}}
	a := actionWith(map[string]*fakePage{"desktop": ok})
	out, _ := a.Execute(context.Background(), RunChecksRequest{RunID: "i", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	if r := resultByID(out.Results, "calc", "desktop"); r == nil || !r.Pass {
		t.Errorf("interaction should pass when result matches, got %+v", r)
	}
	if len(ok.steps) != 2 {
		t.Errorf("both steps should have run, got %d", len(ok.steps))
	}

	// Fail: expect text doesn't match (result stayed non-numeric).
	bad := &fakePage{status: 200, counts: map[string]int{".tool-container": 1, "#tableWrap tr": 1, "#result": 1}, texts: map[string]string{"#result": "—"}}
	a2 := actionWith(map[string]*fakePage{"desktop": bad})
	out2, _ := a2.Execute(context.Background(), RunChecksRequest{RunID: "i2", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	if r := resultByID(out2.Results, "calc", "desktop"); r == nil || r.Pass {
		t.Errorf("interaction should FAIL when result doesn't match, got %+v", r)
	}
}

func TestP2InteractionStepFailure(t *testing.T) {
	// The #go button is missing → click step errors → interaction fails.
	p := &fakePage{status: 200, counts: map[string]int{".tool-container": 1, "#result": 1}, texts: map[string]string{"#result": "1"},
		stepErr: map[string]error{"#go": context.DeadlineExceeded}}
	a := actionWith(map[string]*fakePage{"desktop": p})
	out, _ := a.Execute(context.Background(), RunChecksRequest{RunID: "s", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	r := resultByID(out.Results, "calc", "desktop")
	if r == nil || r.Pass {
		t.Errorf("interaction should FAIL when a step can't run, got %+v", r)
	}
}

func TestConsoleCapturesInteractionErrors(t *testing.T) {
	// An error appears in console; no_console_errors must fail even though it
	// appears before the interaction in the criteria order.
	p := &fakePage{status: 200, console: []string{"TypeError: boom"}, counts: map[string]int{".tool-container": 1, "#tableWrap tr": 1, "#result": 1}, texts: map[string]string{"#result": "5"}}
	a := actionWith(map[string]*fakePage{"desktop": p})
	out, _ := a.Execute(context.Background(), RunChecksRequest{RunID: "c", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	if r := resultByID(out.Results, "console", "desktop"); r == nil || r.Pass {
		t.Errorf("console should fail with an error present, got %+v", r)
	}
}

func TestNavigationFailure(t *testing.T) {
	p := &fakePage{navErr: "net::ERR_NAME_NOT_RESOLVED"}
	a := actionWith(map[string]*fakePage{"desktop": p})
	out, err := a.Execute(context.Background(), RunChecksRequest{RunID: "n", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	if err != nil {
		t.Fatalf("nav failure must be a check failure, not infra error: %v", err)
	}
	if r := resultByID(out.Results, "status", "desktop"); r == nil || r.Pass {
		t.Errorf("status should fail on nav error, got %+v", r)
	}
	for _, id := range []string{"boots", "rows", "calc", "console"} {
		if r := resultByID(out.Results, id, "desktop"); r == nil || r.Pass {
			t.Errorf("%s must fail (not evaluated) when navigation failed, got %+v", id, r)
		}
	}
}

func TestEmptyCriteriaAndNoURLsError(t *testing.T) {
	a := actionWith(nil)
	if _, err := a.Execute(context.Background(), RunChecksRequest{URLs: []string{"u"}, CriteriaJSON: " "}); err == nil {
		t.Error("empty criteria must error, not fake a pass")
	}
	if _, err := a.Execute(context.Background(), RunChecksRequest{URLs: []string{}, CriteriaJSON: criteriaDesktopMobile}); err == nil {
		t.Error("no urls must error")
	}
}

func TestResolveProfiles(t *testing.T) {
	cases := map[string][]string{}
	cases["empty→desktop"] = resolveProfiles(nil)
	if len(cases["empty→desktop"]) != 1 || cases["empty→desktop"][0] != "desktop" {
		t.Error("empty profiles must default to desktop")
	}
	if got := resolveProfiles([]string{"mobile", "desktop"}); len(got) != 2 || got[0] != "desktop" || got[1] != "mobile" {
		t.Errorf("stable order desktop,mobile expected, got %v", got)
	}
	if got := resolveProfiles([]string{"mobile"}); len(got) != 1 || got[0] != "mobile" {
		t.Errorf("mobile-only should run mobile, got %v", got)
	}
}

// ── P3: screenshots on failure ──────────────────────────────────────────────

func TestP3ScreenshotCapturedOnFailure(t *testing.T) {
	// mobile overflows → mobile-fit fails → one screenshot for that page.
	mobile := &fakePage{status: 200, counts: map[string]int{".tool-container": 1, "#tableWrap tr": 1, "#result": 1}, overflow: true, texts: map[string]string{"#result": "1"}}
	a := actionWith(map[string]*fakePage{"mobile": mobile})
	st := &fakeStore{}
	a.store = st
	out, err := a.Execute(context.Background(), RunChecksRequest{
		RunID: "run-1", URLs: []string{"u"}, Profiles: []string{"mobile"},
		CriteriaJSON: criteriaDesktopMobile, Function: "tool-xp-curve-designer", SiteID: "e33263f4",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !mobile.shotTaken {
		t.Fatal("a failing run must capture a screenshot when a store is configured")
	}
	if len(out.Screenshots) != 1 {
		t.Fatalf("expected 1 screenshot ref, got %+v", out.Screenshots)
	}
	ref := out.Screenshots[0]
	if ref.Profile != "mobile" || ref.URI == "" || ref.ViewURL == "" {
		t.Errorf("ref must carry profile + durable uri + view url: %+v", ref)
	}
	if !strings.HasPrefix(ref.URI, "s3://test-bucket/acceptance-evidence/e33263f4/tool-xp-curve-designer/run-1_mobile") {
		t.Errorf("unexpected evidence key: %s", ref.URI)
	}
	found := false
	for _, f := range ref.FailingChecks {
		if f == "mobile-fit@mobile" {
			found = true
		}
	}
	if !found {
		t.Errorf("ref must name the failing instances, got %v", ref.FailingChecks)
	}
}

func TestP3NoScreenshotWhenAllPass(t *testing.T) {
	ok := &fakePage{status: 200, counts: map[string]int{".tool-container": 1, "#tableWrap tr": 5, "#result": 1}, texts: map[string]string{"#result": "9"}}
	a := actionWith(map[string]*fakePage{"desktop": ok})
	st := &fakeStore{}
	a.store = st
	out, err := a.Execute(context.Background(), RunChecksRequest{RunID: "p", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	if err != nil {
		t.Fatal(err)
	}
	if ok.shotTaken || len(out.Screenshots) != 0 || len(st.keys) != 0 {
		t.Errorf("a clean pass must not photograph anything: taken=%v refs=%d saves=%d", ok.shotTaken, len(out.Screenshots), len(st.keys))
	}
}

func TestP3NoStoreMeansNoScreenshot(t *testing.T) {
	// store nil (P0 deploys, or storage misconfigured) → failing run unchanged.
	failing := &fakePage{status: 500, counts: map[string]int{}}
	a := actionWith(map[string]*fakePage{"desktop": failing})
	out, err := a.Execute(context.Background(), RunChecksRequest{RunID: "n", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	if err != nil {
		t.Fatal(err)
	}
	if failing.shotTaken {
		t.Error("no store configured — Screenshot must not even be attempted")
	}
	if len(out.Screenshots) != 0 {
		t.Errorf("no store → no refs, got %+v", out.Screenshots)
	}
	if out.Summary.Failed == 0 {
		t.Error("sanity: this run should have failures")
	}
}

func TestP3EvidenceErrorsNeverFailTheRun(t *testing.T) {
	// Capture error and upload error each degrade to nothing — verdict intact.
	captureFails := &fakePage{status: 500, counts: map[string]int{}, shotErr: context.DeadlineExceeded}
	a := actionWith(map[string]*fakePage{"desktop": captureFails})
	a.store = &fakeStore{}
	out, err := a.Execute(context.Background(), RunChecksRequest{RunID: "e1", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	if err != nil || len(out.Screenshots) != 0 {
		t.Fatalf("capture error must not fail the run or emit refs: err=%v refs=%+v", err, out.Screenshots)
	}

	uploadFails := &fakePage{status: 500, counts: map[string]int{}}
	a2 := actionWith(map[string]*fakePage{"desktop": uploadFails})
	a2.store = &fakeStore{err: context.DeadlineExceeded}
	out2, err := a2.Execute(context.Background(), RunChecksRequest{RunID: "e2", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	if err != nil || len(out2.Screenshots) != 0 {
		t.Fatalf("upload error must not fail the run or emit refs: err=%v refs=%+v", err, out2.Screenshots)
	}
	if out2.Summary.Failed == 0 {
		t.Error("sanity: the run itself should still report its failures")
	}
}

func TestP3NoScreenshotOnNavigationFailure(t *testing.T) {
	p := &fakePage{navErr: "net::ERR_NAME_NOT_RESOLVED"}
	a := actionWith(map[string]*fakePage{"desktop": p})
	a.store = &fakeStore{}
	out, err := a.Execute(context.Background(), RunChecksRequest{RunID: "nav", URLs: []string{"u"}, Profiles: []string{"desktop"}, CriteriaJSON: criteriaDesktopMobile})
	if err != nil {
		t.Fatal(err)
	}
	if p.shotTaken || len(out.Screenshots) != 0 {
		t.Error("nothing loaded — there is no page state worth photographing")
	}
}

func TestScreenshotKeySanitizes(t *testing.T) {
	key := screenshotKey("site/../id", "tool fn", "", "mobile", 1)
	if strings.Contains(key, "..") || strings.Contains(key, " ") {
		t.Errorf("unsafe characters must not reach the object key: %s", key)
	}
	if !strings.Contains(key, "unknown-run") {
		t.Errorf("empty run id must fall back, got %s", key)
	}
	if !strings.HasSuffix(key, "_1.png") {
		t.Errorf("urlIdx>0 must disambiguate, got %s", key)
	}
}
