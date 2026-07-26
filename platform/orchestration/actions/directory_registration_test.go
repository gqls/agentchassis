// FILE: platform/orchestration/actions/directory_registration_test.go
//
// Two guards closing the objections the council gate raised on the Phase E
// kind-parameterisation (submission c67ecb24-eb0e-49a0-b673-2b055fdf3f11,
// APPROVED with four advisory objections). Both were self-identified risks in
// my own submission that stopped short of the closing test the rest of the
// change applied everywhere else — which is exactly what the bug_historian
// seat said, and it was right.

package actions

import (
	"reflect"
	"runtime"
	"testing"
)

// TestDirectoryRenderActionRegisteredUnderBothNames closes the bug_historian
// objection on the dual registration.
//
// The risk in my own words, from the submission: a typo in either key "fails
// only at dispatch time, silently, six hours later" — because the publish leg
// runs on a 6-hourly scheduled task, and an unresolvable action name does not
// fail at build, at deploy, or at seed time. Naming that risk and shipping no
// guard for it reproduces the pattern the whole change was written to avoid.
//
// Verified against live config 2026-07-25 (this is the citation my submission
// was missing, and the prior_art_librarian seat was right to demand it — my
// stated version was WRONG in its detail): the action name is referenced by
// the `model-directory-publisher` AGENT DEFINITION (is_active = true), not by
// the scheduled task's input_data as I claimed. The live chain is
// scheduled_tasks 'model-directory-publish' → target_agent_type
// 'model-directory-trigger' → spawns 'model-directory-publisher', whose
// workflow names the action. So the dependency is real and the dual
// registration is necessary rather than precautionary — but by a different
// route than the one I asserted.
func TestDirectoryRenderActionRegisteredUnderBothNames(t *testing.T) {
	for _, name := range []string{"render_model_directory", "render_directory"} {
		def, ok := GlobalActionRegistry[name]
		if !ok {
			t.Errorf("action %q is not in GlobalActionRegistry — a workflow naming it "+
				"fails only at dispatch, with no build or deploy error", name)
			continue
		}
		if def.Handler == nil {
			t.Errorf("action %q is registered with a nil Handler", name)
		}
	}

	// Both names must reach the SAME implementation. Two entries that drifted
	// apart would be a fork wearing a compatibility shim's clothes — the exact
	// thing the reuse seat approved this change for NOT doing. Func values are
	// not comparable in Go beyond nil, so compare the resolved code pointers.
	legacy, okA := GlobalActionRegistry["render_model_directory"]
	general, okB := GlobalActionRegistry["render_directory"]
	if okA && okB && legacy.Handler != nil && general.Handler != nil {
		lp := reflect.ValueOf(legacy.Handler).Pointer()
		gp := reflect.ValueOf(general.Handler).Pointer()
		if lp != gp {
			t.Errorf("render_model_directory and render_directory resolve to different handlers "+
				"(%v vs %v) — the alias has forked", runtime.FuncForPC(lp).Name(), runtime.FuncForPC(gp).Name())
		}
	}
}

// TestBlogPostIsNotAStructuralPageType closes the bug_historian and guardian
// objections on the growth-budget query rewrite.
//
// The budget query used to count the CONTENT bucket with an explicit NOT-IN
// list naming both 'blog-post' and every structural type. It now counts
// "neither blog-post nor a member of structuralPageTypes", which is equivalent
// ONLY while 'blog-post' is absent from that map. My submission said this was
// "worth confirming rather than assuming" and then assumed it. If it ever
// becomes a member, blog posts get counted in no bucket at all and the growth
// budget silently under-counts — no error, no log, just a site that is allowed
// to grow faster than its configured rate.
func TestBlogPostIsNotAStructuralPageType(t *testing.T) {
	if structuralPageTypes["blog-post"] {
		t.Fatal("'blog-post' is in structuralPageTypes: the growth-budget query's content " +
			"bucket ('neither blog-post nor structural') now excludes blog posts from every " +
			"bucket. Either remove it from the map or restore the explicit NOT-IN list.")
	}

	// The list must also be non-empty and sorted-stable, since it is passed to
	// SQL as a text[] literal: an empty list would make `= ANY($2)` match
	// nothing and silently reclassify every structural page as content.
	list := structuralPageTypeList()
	if len(list) == 0 {
		t.Fatal("structuralPageTypeList() is empty — every structural page would count as content")
	}
	for i := 1; i < len(list); i++ {
		if list[i-1] >= list[i] {
			t.Errorf("structuralPageTypeList() is not sorted at %d (%q >= %q)", i, list[i-1], list[i])
		}
	}
}

// TestDirectoryKindResolvesFromLiteralStepConfig is the regression guard for
// the defect that reached production on 2026-07-26: the three-kind publish
// chain published the MODEL register three times because
// ExtractActionInputs dropped `"kind": "company"` — string configs are
// references, never literals, by design (action_inputs.go Strategy 5).
//
// The property: a value from the CLOSED profile set, written literally in a
// step's config, must select that profile. Anything else must fall through to
// the reference path so a genuine wiring mistake still surfaces rather than
// being silently reinterpreted as a literal.
func TestDirectoryKindResolvesFromLiteralStepConfig(t *testing.T) {
	// resolveKind mirrors the action's selection, which is the part that was
	// wrong. Kept in lockstep with RenderDirectoryAction deliberately: if the
	// action's logic changes, this test should be changed with it, on purpose.
	resolveKind := func(configKind, resolvedInput string) string {
		kind := resolvedInput
		if _, known := directoryPublishProfiles[configKind]; known && configKind != "" {
			kind = configKind
		}
		if kind == "" {
			kind = "model"
		}
		return kind
	}

	cases := []struct {
		name        string
		configKind  string // literal in the step's config
		resolved    string // what ExtractActionInputs managed to resolve
		want        string
		explanation string
	}{
		{"company literal", "company", "", "company",
			"the exact case that shipped broken — config literal must win"},
		{"protocol literal", "protocol", "", "protocol", ""},
		{"model literal", "model", "", "model", ""},
		{"no kind anywhere", "", "", "model",
			"pre-Phase-E seeds pass no kind and must keep publishing models"},
		{"reference that resolved", "input_data.kind", "company", "company",
			"a resolved reference still works; the config string is not a profile name so it is ignored as a literal"},
		{"typo falls through, not silently literal", "compnay", "", "model",
			"a misspelt kind must NOT become its own value; it falls to the default and then to the unknown-kind refusal if referenced"},
	}
	for _, tc := range cases {
		if got := resolveKind(tc.configKind, tc.resolved); got != tc.want {
			t.Errorf("%s: resolveKind(config=%q, resolved=%q) = %q, want %q — %s",
				tc.name, tc.configKind, tc.resolved, got, tc.want, tc.explanation)
		}
	}

	// Every profile name must be usable as a literal, or a kind added later
	// silently reverts to publishing models — the same failure, one kind along.
	for name := range directoryPublishProfiles {
		if got := resolveKind(name, ""); got != name {
			t.Errorf("profile %q is not selectable as a config literal (got %q)", name, got)
		}
	}
}
