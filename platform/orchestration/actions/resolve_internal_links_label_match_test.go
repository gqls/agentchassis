// FILE: platform/orchestration/actions/resolve_internal_links_label_match_test.go
//
// Pins the bugs_open/203 follow-on: a CTA's own label, when one is already
// published, should steer resolution toward the real page it actually names
// — instead of chooseCTATargets' pure position (NavOrder, Name) picking
// whatever ranks first regardless of what any button claims to do. Covers
// both call sites that now try a label match before falling back to the
// unchanged positional behaviour: setCTAField (write-time) and
// applyCTARecompute (the check_misdirected_cta repair path, which previously
// could not fix the exact defect it was triggered to fix — see that
// function's own comment).

package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func riskCheckerCandidates(t *testing.T) []datahelpers.LabelMatchCandidate {
	t.Helper()
	riskChecker, ok := datahelpers.NewLabelMatchCandidate(
		"1", "tool-risk-checker", "AI Data Risk Checker", "/tools/tool-ai-data-risk-checker.html",
		true, "tool-risk-checker", "AI Data Risk Checker")
	if !ok {
		t.Fatal("fixture candidate produced no tokens")
	}
	passwordEntropy, ok := datahelpers.NewLabelMatchCandidate(
		"2", "tool-password-entropy", "Password Strength Physics", "/tools/password-entropy.html",
		true, "tool-password-entropy", "Password Strength Physics")
	if !ok {
		t.Fatal("fixture candidate produced no tokens")
	}
	return []datahelpers.LabelMatchCandidate{riskChecker, passwordEntropy}
}

// TestSetCTAFieldPrefersLabelMatchOverPositionalTarget is the direct fix for
// the live defect: chooseCTATargets' positional pick (here standing in as
// "password-entropy", exactly as it was on the real finetuning.uk canary)
// disagrees with what the button's own existing label names. The label match
// must win.
func TestSetCTAFieldPrefersLabelMatchOverPositionalTarget(t *testing.T) {
	valid := datahelpers.NewPageURLSet([]string{
		"/tools/tool-ai-data-risk-checker.html", "/tools/password-entropy.html",
	})
	positionalTarget := contentHub{Name: "tool-password-entropy", Title: "Password Strength Physics", URL: "/tools/password-entropy.html"}
	candidates := riskCheckerCandidates(t)

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, "cta_url", positionalTarget, valid, "hero", "hero", "primary", &unresolved,
		"Run the Risk Checker", candidates)

	if got := resolved["cta_url"]; got != "/tools/tool-ai-data-risk-checker.html" {
		t.Errorf("cta_url = %v, want the risk-checker tool (label match), not the positional pick", got)
	}
	if got := resolved["cta_target_title"]; got != "AI Data Risk Checker" {
		t.Errorf("cta_target_title = %v, want the matched target's title", got)
	}
}

// TestSetCTAFieldFallsBackToPositionalWhenLabelIsGeneric: a vague label must
// never force a match — this is the false-positive guard, not a side effect.
func TestSetCTAFieldFallsBackToPositionalWhenLabelIsGeneric(t *testing.T) {
	valid := datahelpers.NewPageURLSet([]string{"/tools/password-entropy.html"})
	positionalTarget := contentHub{Name: "tool-password-entropy", Title: "Password Strength Physics", URL: "/tools/password-entropy.html"}
	candidates := riskCheckerCandidates(t)

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, "cta_url", positionalTarget, valid, "hero", "hero", "primary", &unresolved,
		"Get Started", candidates)

	if got := resolved["cta_url"]; got != positionalTarget.URL {
		t.Errorf("cta_url = %v, want the unchanged positional pick for a generic label", got)
	}
}

// TestSetCTAFieldFallsBackToPositionalWhenLabelMatchesNoCandidate: a specific
// label that simply names nothing in the candidate set must not block
// resolution — it degrades to today's behaviour, not to unresolved.
func TestSetCTAFieldFallsBackToPositionalWhenLabelMatchesNoCandidate(t *testing.T) {
	valid := datahelpers.NewPageURLSet([]string{"/tools/password-entropy.html"})
	positionalTarget := contentHub{Name: "tool-password-entropy", Title: "Password Strength Physics", URL: "/tools/password-entropy.html"}
	candidates := riskCheckerCandidates(t)

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, "cta_url", positionalTarget, valid, "hero", "hero", "primary", &unresolved,
		"Read Our Privacy Policy", candidates)

	if got := resolved["cta_url"]; got != positionalTarget.URL {
		t.Errorf("cta_url = %v, want the positional pick when the label matches no candidate", got)
	}
}

// TestApplyCTARecomputeOverridesValidButMisdirectedLink is the KEY repair
// fix. Before this change, a currently-stored URL that was merely valid,
// non-excluded and non-self was kept unconditionally — so a misdirected-but-
// otherwise-fine-looking link (exactly what check_misdirected_cta detects and
// this function is invoked to repair) survived every cta_links_stale pass
// forever. Mutation target: comment out the label-match block below and this
// test must fail.
func TestApplyCTARecomputeOverridesValidButMisdirectedLink(t *testing.T) {
	valid := datahelpers.NewPageURLSet([]string{
		"/tools/tool-ai-data-risk-checker.html", "/tools/password-entropy.html", "/index.html",
	})
	positionalTarget := contentHub{Name: "tool-password-entropy", Title: "Password Strength Physics", URL: "/tools/password-entropy.html"}
	candidates := riskCheckerCandidates(t)

	resolved := map[string]interface{}{}
	stored := map[string]interface{}{"cta_url": "/tools/password-entropy.html"} // valid, but wrong for the label
	applyCTARecompute(resolved, stored, "cta_url", positionalTarget, valid, "/index.html",
		"Run the Risk Checker", candidates)

	if got := resolved["cta_url"]; got != "/tools/tool-ai-data-risk-checker.html" {
		t.Errorf("cta_url = %v, want the label-matched risk-checker tool to override the valid-but-wrong stored link", got)
	}
}

// TestApplyCTARecomputeLeavesAlreadyCorrectLinkUnwritten: when the stored URL
// already IS the label's best match, nothing should be written — a real
// no-op, not a same-value overwrite that would spuriously show up as a
// change in the ownership-conflict log.
func TestApplyCTARecomputeLeavesAlreadyCorrectLinkUnwritten(t *testing.T) {
	valid := datahelpers.NewPageURLSet([]string{"/tools/tool-ai-data-risk-checker.html", "/index.html"})
	positionalTarget := contentHub{Name: "tool-password-entropy", URL: "/tools/password-entropy.html"}
	candidates := riskCheckerCandidates(t)

	resolved := map[string]interface{}{}
	stored := map[string]interface{}{"cta_url": "/tools/tool-ai-data-risk-checker.html"} // already correct
	applyCTARecompute(resolved, stored, "cta_url", positionalTarget, valid, "/index.html",
		"Run the Risk Checker", candidates)

	if len(resolved) != 0 {
		t.Errorf("expected no write when the stored link already matches the label, got %v", resolved)
	}
}

// TestApplyCTARecomputeFallsBackWhenLabelGeneric pins the unchanged half:
// a generic label keeps today's "authored valid link, keep it" behaviour.
func TestApplyCTARecomputeFallsBackWhenLabelGeneric(t *testing.T) {
	valid := datahelpers.NewPageURLSet([]string{"/tools/password-entropy.html", "/index.html"})
	positionalTarget := contentHub{Name: "tool-risk-checker", URL: "/tools/tool-ai-data-risk-checker.html"}
	candidates := riskCheckerCandidates(t)

	resolved := map[string]interface{}{}
	stored := map[string]interface{}{"cta_url": "/tools/password-entropy.html"}
	applyCTARecompute(resolved, stored, "cta_url", positionalTarget, valid, "/index.html",
		"Get Started", candidates)

	if len(resolved) != 0 {
		t.Errorf("generic label should keep the stored valid link untouched, got %v", resolved)
	}
}
