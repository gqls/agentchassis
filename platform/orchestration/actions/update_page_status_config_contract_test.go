// FILE: platform/orchestration/actions/update_page_status_config_contract_test.go
//
// bugs_open/336. `deploy_result_field` was declared on RenderComponentInputSpec —
// the OTHER spec this file's init() registers, for an action that never reads the
// key — while the reader is UpdatePageStatusAction and UpdatePageStatusInputSpec
// sets StrictConfig: true. Nothing failed until migration 494 armed the key on
// live update_page_status steps (2026-08-20 06:49:49Z), at which point the strict
// branch of validation.checkStepConfigKeys hard-failed every workflow that stamps
// a page: 8 items across 4 item types, 123 page_rerender queued fleet-wide and
// none draining.
//
// Why no existing check saw it, which is what these tests are for:
//   - both specs are literals in one file, forty lines apart, each correct-looking
//     in isolation, and a diff shows only an added line in a list;
//   - a binary probe says the key is PRESENT whichever spec declares it (the
//     literal is in the reader and in two zap.String calls);
//   - `git log -S'"deploy_result_field",'` matches the zap call, so it names the
//     commit that shipped the READER and reads like the declaration's commit;
//   - the owning lane's arming precondition ("wait for a build carrying X") was
//     about the reader too, so ancestry said go.
//
// The third test is the general one: every key the handler READS must be declared
// on ITS OWN spec. That is the check whose absence cost the outage, and it is
// scoped to this one action deliberately — a fleet-wide version belongs in
// cmd/config-key-audit, where the source-scanning machinery already lives.
package actions

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func specHasKey(spec datahelpers.ActionInputSpec, key string) bool {
	for _, k := range spec.ConfigKeys {
		if k == key {
			return true
		}
	}
	return false
}

// TestDeployResultFieldIsDeclaredOnTheActionThatReadsIt pins BOTH halves. Asserting
// only that update_page_status declares it would still pass with the key left on
// render_component as well, which is the state that hid the defect: a key declared
// somewhere reads as declared to anyone grepping the file.
func TestDeployResultFieldIsDeclaredOnTheActionThatReadsIt(t *testing.T) {
	const key = "deploy_result_field"

	if !specHasKey(UpdatePageStatusInputSpec, key) {
		t.Errorf("%s is NOT declared on UpdatePageStatusInputSpec, which is the spec for the "+
			"action that reads it (v3_site_actions.go, UpdatePageStatusAction). Because that spec "+
			"is StrictConfig, any live step setting this key will fail workflow validation and take "+
			"the whole workflow down — bugs_open/336", key)
	}
	if specHasKey(RenderComponentInputSpec, key) {
		t.Errorf("%s is declared on RenderComponentInputSpec, whose action never reads it. That is "+
			"where bugs_open/336 put it; declaring a key on a spec that does not read it hides the "+
			"key from the reader's strict check and misstates whose contract it belongs to", key)
	}
}

// TestArmingDeployResultFieldPassesStrictValidation exercises the predicate the
// validator itself calls, with the exact config shape migration 494 writes. This is
// the behavioural test: it fails if the declaration moves back, whatever the lists
// happen to look like.
func TestArmingDeployResultFieldPassesStrictValidation(t *testing.T) {
	if !datahelpers.IsStrictConfigAction("update_page_status") {
		t.Fatal("update_page_status is no longer StrictConfig — if that was deliberate, this test's " +
			"premise is gone, but do NOT relax strictness to paper over a misplaced key: strictness " +
			"caught bugs_open/336 at validation instead of letting the deploy stamp silently skip " +
			"reading its evidence (the bugs_open/234 shape)")
	}

	// The live shape as armed by 494 on page-rerender / report-builder
	// (update_status) and section-editor (update_page_status).
	armed := map[string]interface{}{
		"status":              "deployed",
		"deploy_result_field": "deploy_result",
	}
	unknown, checked := datahelpers.UnknownConfigKeys("update_page_status", armed)
	if !checked {
		t.Fatal("update_page_status no longer opts into unknown-config-key checking (CheckConfig)")
	}
	if len(unknown) != 0 {
		t.Errorf("validation would reject an armed step: unrecognised keys %v. This is exactly the "+
			"WORKFLOW_INVALID that stopped every page-stamping workflow on 2026-08-20 — bugs_open/336",
			unknown)
	}

	// The negative arm: a key this action really does not read must still be
	// caught, so the test cannot pass by the check having been switched off.
	unknown, _ = datahelpers.UnknownConfigKeys("update_page_status",
		map[string]interface{}{"status": "deployed", "not_a_real_key": "x"})
	if len(unknown) != 1 || unknown[0] != "not_a_real_key" {
		t.Errorf("the unknown-key check is no longer discriminating: expected [not_a_real_key], got %v", unknown)
	}
}

// TestUpdatePageStatusDeclaresEveryConfigKeyItReads scans the handler body for
// direct config accesses and requires each one to be declared. Sound for THIS
// action specifically because it indexes params.StepConfig.Config directly, with no
// ResolveConfigSetting / GetStringField / ExtractActionInputs indirection for a key
// to hide behind — a property the spec's own comment records and this test would
// notice losing (a key resolved through a helper simply would not appear here, so
// the test is a floor, not a ceiling).
func TestUpdatePageStatusDeclaresEveryConfigKeyItReads(t *testing.T) {
	src, err := os.ReadFile("v3_site_actions.go")
	if err != nil {
		t.Fatalf("cannot read the source this test reasons about: %v", err)
	}

	lines := strings.Split(string(src), "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "func UpdatePageStatusAction(") {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatal("UpdatePageStatusAction not found — this test is scanning the wrong thing and would " +
			"otherwise pass by finding nothing")
	}
	end := -1
	for i := start + 1; i < len(lines); i++ {
		if lines[i] == "}" {
			end = i
			break
		}
	}
	if end < 0 {
		t.Fatal("could not find the end of UpdatePageStatusAction")
	}

	// Comment lines are stripped so that a key NAMED in prose (this file discusses
	// its own keys at length) cannot be mistaken for an access.
	re := regexp.MustCompile(`config\["([a-zA-Z0-9_]+)"\]`)
	read := map[string]int{}
	for i := start; i <= end; i++ {
		body := lines[i]
		if strings.HasPrefix(strings.TrimSpace(body), "//") {
			continue
		}
		for _, m := range re.FindAllStringSubmatch(body, -1) {
			read[m[1]] = i + 1
		}
	}
	if len(read) == 0 {
		t.Fatal("found no config accesses in UpdatePageStatusAction — the scan is broken, and a " +
			"broken scan passes silently")
	}

	for key, line := range read {
		if datahelpers.IsFrameworkStepConfigKey(key) {
			continue
		}
		if !specHasKey(UpdatePageStatusInputSpec, key) {
			t.Errorf("v3_site_actions.go:%d reads config[%q] but UpdatePageStatusInputSpec does not "+
				"declare it. On a StrictConfig action that is not an inert key: the first live step "+
				"to set it fails the whole workflow at validation (bugs_open/336)", line, key)
		}
	}
}
