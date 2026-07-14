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
	overflow bool
	culprit  string // widest element crossing the viewport edge
	ovErr    error
	texts    map[string]string
	stepErr  map[string]error // selector -> error to return from Do
	steps    []criteriaStep   // recorded
}

func (f *fakePage) Status() int            { return f.status }
func (f *fakePage) NavError() string       { return f.navErr }
func (f *fakePage) ConsoleErrors() []string { return f.console }
func (f *fakePage) Count(sel string) int   { return f.counts[sel] }
func (f *fakePage) HorizontalOverflow() (bool, string, error) {
	return f.overflow, f.culprit, f.ovErr
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

// The overflow is measured on the DOCUMENT, so the failure must name the widest
// offender — that is what tells a reader (and the fixer) whether the tool broke
// or the site template did. Live case 2026-07-14: vonc.com's div.footer-legal
// overflowed every page, failing a QUIZ's mobile-fit check. Without the culprit
// in the detail, the ticket is indistinguishable from a real tool bug.
func TestP1OverflowDetailNamesTheCulprit(t *testing.T) {
	mobile := &fakePage{
		status:   200,
		counts:   map[string]int{".tool-container": 1, "#tableWrap tr": 1, "#result": 1},
		overflow: true,
		culprit:  "div.site-footer.footer-legal (506px)",
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
	if !strings.Contains(r.Detail, "div.site-footer.footer-legal (506px)") {
		t.Errorf("the failure must name the offending element so the bug can be attributed; got %q", r.Detail)
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
