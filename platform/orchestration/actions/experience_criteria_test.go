// FILE: platform/orchestration/actions/experience_criteria_test.go
//
// Tests for the experience-register criteria validator. Two kinds:
//
//   - behaviour tests, each pinned to a real template fragment from the nine
//     entries harvested on 2026-07-26, so the validator is exercised against
//     what it will actually be asked to validate rather than invented input;
//   - a SOURCE-LOCKSTEP test that reads the two checkers' own switch
//     statements and asserts the capability table matches. A hand-maintained
//     capability table drifts from the checker within one change, and a
//     validator trusting a stale table waves through exactly the checks it
//     exists to catch. (This is also the shape the WRONG_CALLS tally kept
//     asking for: a check that cannot share ground with the thing it checks.)

package actions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func mustJSON(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("bad test JSON: %v", err)
	}
	return m
}

func hasErrorContaining(v ExperienceCriteriaValidation, want string) bool {
	for _, e := range v.Errors {
		if strings.Contains(e.Detail, want) {
			return true
		}
	}
	return false
}

func hasDeferredFor(v ExperienceCriteriaValidation, checkID string) bool {
	for _, d := range v.Deferred {
		if d.CheckID == checkID {
			return true
		}
	}
	return false
}

// The executable core of CC-001 feed-driven-teaser-list, as harvested.
func TestValidateExperienceCriteria_AcceptsAHarvestedTemplate(t *testing.T) {
	tmpl := mustJSON(t, `{
	  "profiles": ["desktop","mobile"],
	  "container": "{{binding.list_section}}",
	  "checks": [
	    {"id":"list_exists","type":"selector_exists","selector":"{{binding.list_container}}"},
	    {"id":"feed_loads","type":"asset_loads","path":"{{binding.feed_path}}"},
	    {"id":"detail_opens_populated","type":"interaction",
	     "steps":[{"action":"click","selector":"{{binding.item_openable_selector}}"}],
	     "expect":{"selector":"{{binding.detail_selector}}","text_matches":".{200,}"}}
	  ]}`)
	schema := mustJSON(t, `{
	  "list_section":{"type":"selector"},
	  "list_container":{"type":"selector"},
	  "feed_path":{"type":"asset_path"},
	  "item_openable_selector":{"type":"selector"},
	  "detail_selector":{"type":"selector"}}`)

	v := ValidateExperienceCriteria(tmpl, schema)
	if !v.OK() {
		t.Fatalf("harvested template must validate, got %+v", v.Errors)
	}
	if v.Executable != 3 {
		t.Errorf("want 3 executable checks, got %d (%s)", v.Executable, v.Summary())
	}
	if len(v.Placeholders) != 5 {
		t.Errorf("want 5 placeholders, got %v", v.Placeholders)
	}
}

// P3/P4 — closure in both directions.
func TestValidateExperienceCriteria_PlaceholderClosure(t *testing.T) {
	undeclared := mustJSON(t, `{"checks":[
	  {"id":"a","type":"selector_exists","selector":"{{binding.mystery}}"}]}`)
	v := ValidateExperienceCriteria(undeclared, map[string]interface{}{})
	if !hasErrorContaining(v, "not declared in binding_schema") {
		t.Errorf("undeclared placeholder must fail, got %+v", v.Errors)
	}

	unused := mustJSON(t, `{"checks":[
	  {"id":"a","type":"selector_exists","selector":"{{binding.used}}"}]}`)
	schema := mustJSON(t, `{"used":{"type":"selector"},"forgotten":{"type":"selector"}}`)
	v = ValidateExperienceCriteria(unused, schema)
	if !hasErrorContaining(v, `binding "forgotten" is declared but never used`) {
		t.Errorf("unused binding must fail, got %+v", v.Errors)
	}

	// …unless it is deliberate.
	schema["forgotten"] = map[string]interface{}{"type": "selector", "unused_ok": true}
	if v = ValidateExperienceCriteria(unused, schema); !v.OK() {
		t.Errorf("unused_ok must silence it, got %+v", v.Errors)
	}
}

// A placeholder used only in the CONTRACT still has to close.
func TestValidateExperienceCriteria_ContractPlaceholdersCount(t *testing.T) {
	tmpl := mustJSON(t, `{"checks":[{"id":"a","type":"selector_exists","selector":"{{binding.card}}"}]}`)
	schema := mustJSON(t, `{"card":{"type":"selector"}}`)
	var contract interface{}
	if err := json.Unmarshal([]byte(`[{"control_role":"card","primitive":"navigate",
	  "destination_role":"{{binding.card_destination_role}}"}]`), &contract); err != nil {
		t.Fatal(err)
	}
	v := ValidateExperienceCriteria(tmpl, schema, contract)
	if !hasErrorContaining(v, "{{binding.card_destination_role}}") {
		t.Errorf("a placeholder used only in the contract must still close, got %+v", v.Errors)
	}
}

// P6/P7/P8 — the three ways a check can be beyond the platform, each an error
// when unmarked and a deferral when declared. These are the real cases from
// HARVEST_01 §§3.2-3.5.
func TestValidateExperienceCriteria_UnexecutableChecks(t *testing.T) {
	cases := []struct {
		name, check, wantErr string
	}{
		{"unknown type (attribute assertion — the anti-dead-control clause)",
			`{"id":"static_not_interactive","type":"attribute_absent","selector":"{{binding.row}}"}`,
			"not executable by any tier"},
		{"inert field (the 300ms-vs-8-23s wait)",
			`{"id":"turn_returns","type":"interaction","selector":"{{binding.row}}","expect_within_ms":60000}`,
			"not read by any checker"},
		{"unperformable step (the deep link)",
			`{"id":"deep_link","type":"interaction","steps":[{"action":"goto","selector":"{{binding.row}}"}]}`,
			"is not performed by the runner"},
		{"unevaluated expectation (visibility)",
			`{"id":"reveal","type":"interaction","expect":{"selector":"{{binding.row}}","visible":true}}`,
			"is not evaluated by any checker"},
	}
	schema := mustJSON(t, `{"row":{"type":"selector"}}`)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmpl := mustJSON(t, `{"checks":[`+c.check+`]}`)
			v := ValidateExperienceCriteria(tmpl, schema)
			if !hasErrorContaining(v, c.wantErr) {
				t.Fatalf("want error %q, got %+v", c.wantErr, v.Errors)
			}
			if v.Executable != 0 {
				t.Errorf("an unexecutable check must not be counted executable")
			}

			// Marked: recorded as deferred, never an error, never a pass.
			marked := mustJSON(t, `{"checks":[`+
				strings.TrimSuffix(c.check, "}")+`,"_unsupported":"HARVEST_01 dependency"}]}`)
			v = ValidateExperienceCriteria(marked, schema)
			if !v.OK() {
				t.Errorf("a marked check must validate, got %+v", v.Errors)
			}
			if v.Executable != 0 {
				t.Errorf("a deferred check must never count as executable")
			}
			if len(v.Deferred) == 0 {
				t.Errorf("a marked check must be recorded as deferred, not dropped")
			}
		})
	}
}

// P5 — base entries are site-agnostic.
func TestValidateExperienceCriteria_NoSiteSpecificValues(t *testing.T) {
	absolute := mustJSON(t, `{"checks":[
	  {"id":"a","type":"asset_loads","path":"https://vonc.com/data/provocations.json"}]}`)
	v := ValidateExperienceCriteria(absolute, map[string]interface{}{})
	if !hasErrorContaining(v, "absolute URL appears in the template") {
		t.Errorf("absolute URL must fail, got %+v", v.Errors)
	}

	literal := mustJSON(t, `{"checks":[
	  {"id":"a","type":"selector_exists","selector":".provocations-archive__list"}]}`)
	v = ValidateExperienceCriteria(literal, map[string]interface{}{})
	if !hasErrorContaining(v, "is a literal") {
		t.Errorf("literal selector must fail, got %+v", v.Errors)
	}

	justified := mustJSON(t, `{"checks":[
	  {"id":"a","type":"selector_exists","selector":"a[href]","literal_ok":"generic HTML, carries no site's own naming"}]}`)
	if v = ValidateExperienceCriteria(justified, map[string]interface{}{}); !v.OK() {
		t.Errorf("a justified literal must pass, got %+v", v.Errors)
	}
}

// P9 — the -EDIT trap: Tier 2 skips those silently, so they read as green.
func TestValidateExperienceCriteria_RejectsEditPlaceholderIDs(t *testing.T) {
	tmpl := mustJSON(t, `{"checks":[{"id":"selector-EDIT","type":"selector_exists","selector":"{{binding.x}}"}]}`)
	v := ValidateExperienceCriteria(tmpl, mustJSON(t, `{"x":{"type":"selector"}}`))
	if !hasErrorContaining(v, "-EDIT") {
		t.Errorf("an -EDIT id must fail, got %+v", v.Errors)
	}
}

func TestValidateExperienceCriteria_RequiresChecksAndIDs(t *testing.T) {
	if v := ValidateExperienceCriteria(mustJSON(t, `{"checks":[]}`), map[string]interface{}{}); v.OK() {
		t.Error("an empty checks array must fail")
	}
	dup := mustJSON(t, `{"checks":[
	  {"id":"a","type":"page_status_ok"},
	  {"id":"a","type":"page_status_ok"}]}`)
	if v := ValidateExperienceCriteria(dup, map[string]interface{}{}); !hasErrorContaining(v, "duplicate check id") {
		t.Error("duplicate ids must fail")
	}
	// An inherited-clause marker carries no assertions and is not a check.
	marker := mustJSON(t, `{"checks":[
	  {"_inherited":"asserted by the required component contract"},
	  {"id":"a","type":"page_status_ok"}]}`)
	if v := ValidateExperienceCriteria(marker, map[string]interface{}{}); !v.OK() {
		t.Errorf("an inherited marker must be allowed, got %+v", v.Errors)
	}
	if hasDeferredFor(ValidateExperienceCriteria(marker, map[string]interface{}{}), "") {
		t.Error("a marker must not be recorded as a deferred check")
	}
}

// ── the lockstep test ──────────────────────────────────────────────────────

// TestExperienceCheckCapabilities_LockstepWithCheckers reads the switch
// statements out of both checkers and asserts the validator's capability table
// says exactly what they implement. If someone adds a check type to the browser
// runner, or removes one, this fails until the table moves with it — the
// dedup-index/Go-list idiom (v1.0.1127) and the doc-subject lockstep
// (bugs_closed/064) applied a third time, to the thing the register's whole
// acceptance side depends on.
func TestExperienceCheckCapabilities_LockstepWithCheckers(t *testing.T) {
	const (
		tier2Path = "discovery_checks/check_tool_acceptance.go"
		tier4Path = "../../../internal/adapters/browserrunner/run_checks_action.go"
	)
	caseRE := regexp.MustCompile(`(?m)^\s*case\s+((?:"[a-z_]+"\s*,?\s*)+):`)
	valueRE := regexp.MustCompile(`"([a-z_]+)"`)

	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Clean(rel))
		if err != nil {
			t.Fatalf("cannot read %s (this test runs from the package dir and needs both checkers in the checkout): %v", rel, err)
		}
		return string(raw)
	}

	// Tier 2: every case in evaluateStaticCriteria's switch is a check type.
	tier2 := map[string]bool{}
	src2 := read(tier2Path)
	body2 := betweenFuncs(src2, "func evaluateStaticCriteria(")
	for _, m := range caseRE.FindAllStringSubmatch(body2, -1) {
		for _, v := range valueRE.FindAllStringSubmatch(m[1], -1) {
			tier2[v[1]] = true
		}
	}
	if len(tier2) == 0 {
		t.Fatal("no check types found in evaluateStaticCriteria — the checker's shape changed; update this test rather than the table")
	}

	// Tier 4: applicability switch = check types; the step switch = actions.
	src4 := read(tier4Path)
	tier4 := map[string]bool{}
	for _, m := range caseRE.FindAllStringSubmatch(betweenFuncs(src4, "func applicableChecks("), -1) {
		for _, v := range valueRE.FindAllStringSubmatch(m[1], -1) {
			tier4[v[1]] = true
		}
	}
	if len(tier4) == 0 {
		// The applicability switch may live in a differently-named function;
		// fall back to the whole file, which is still the real source.
		for _, m := range caseRE.FindAllStringSubmatch(src4, -1) {
			for _, v := range valueRE.FindAllStringSubmatch(m[1], -1) {
				tier4[v[1]] = true
			}
		}
	}

	// Every type the validator claims is executable must appear in the tier it
	// claims, and no checker type may be missing from the table.
	for typ, tier := range experienceCheckTiers {
		switch tier {
		case 2:
			if !tier2[typ] {
				t.Errorf("table says %q is Tier 2 but %s does not implement it", typ, tier2Path)
			}
		case 4:
			if !tier4[typ] {
				t.Errorf("table says %q is Tier 4 but the browser runner does not implement it", typ)
			}
		default:
			t.Errorf("table gives %q an unknown tier %d", typ, tier)
		}
	}
	for typ := range tier2 {
		if _, known := experienceCheckTiers[typ]; !known {
			t.Errorf("%s implements %q but the validator's table does not know it — a template using it would be rejected as unexecutable", tier2Path, typ)
		}
	}
	for typ := range tier4 {
		if experienceStepActions[typ] {
			continue // step action, not a check type
		}
		if _, known := experienceCheckTiers[typ]; !known {
			t.Errorf("the browser runner implements %q but the validator's table does not know it", typ)
		}
	}

	// Step actions, same discipline.
	stepSwitch := betweenFuncs(src4, "func (p *chromiumPage) Do(")
	if stepSwitch == "" {
		stepSwitch = src4
	}
	found := map[string]bool{}
	for _, m := range caseRE.FindAllStringSubmatch(stepSwitch, -1) {
		for _, v := range valueRE.FindAllStringSubmatch(m[1], -1) {
			if experienceStepActions[v[1]] {
				found[v[1]] = true
			}
		}
	}
	var missing []string
	for a := range experienceStepActions {
		if !found[a] {
			missing = append(missing, a)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("step action(s) %v are in the table but not in the runner's step switch", missing)
	}
}

// betweenFuncs returns the source from the start of the named function to the
// next top-level `func ` — enough to scope a switch to its own function without
// parsing Go.
func betweenFuncs(src, funcStart string) string {
	i := strings.Index(src, funcStart)
	if i < 0 {
		return ""
	}
	rest := src[i+len(funcStart):]
	if j := strings.Index(rest, "\nfunc "); j >= 0 {
		return rest[:j]
	}
	return rest
}
