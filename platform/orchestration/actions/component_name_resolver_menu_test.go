// FILE: platform/orchestration/actions/component_name_resolver_menu_test.go
//
// bugs_open/282 — validate_site_plan drops what the planner's menu offered.
//
// The defect was invisible in every direction: the planner's raw response held
// the tool sections, the persisted plan did not, and the only trace was one
// Warn per dropped name. So the tests here pin BOTH directions of the opt-in —
// with `menu_field` a menu component survives, without it the identical plan
// still loses it (today's behaviour, which the fix must not change for any
// caller that has not opted in) — plus the degradation cases, because a menu
// query that selects different columns must no-op rather than widen the
// resolver to accept "".
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// baseResolver is the resolver as loadComponentNameResolver builds it from the
// section/element base: hero and faq, nothing tool-level.
func baseResolver() *componentNameResolver {
	return &componentNameResolver{
		validFunctions: map[string]bool{"hero": true, "faq": true},
		displayToFunc:  map[string]string{"faq section": "faq"},
		nameToFunc:     map[string]string{"hero": "hero"},
	}
}

// toolMenuRow is a menu row exactly as build-site-planner's load_components
// query selects it (name, display_name, function, category, description).
func toolMenuRow() map[string]interface{} {
	return map[string]interface{}{
		"name":         "Standard Loan Repayment Calculator",
		"display_name": "Standard Loan Repayment Calculator",
		"function":     "tool-loan-repayment",
		"category":     "finance",
		"description":  "Monthly repayment calculator",
	}
}

// ── addMenu ─────────────────────────────────────────────────────────────────

func TestAddMenu_MenuFunctionBecomesResolvableAndIsMarked(t *testing.T) {
	r := baseResolver()

	if _, ok := r.resolve("tool-loan-repayment"); ok {
		t.Fatal("precondition failed: the section/element base must NOT resolve a tool function")
	}

	added := r.addMenu([]interface{}{toolMenuRow()})
	if added != 1 {
		t.Errorf("expected 1 function added beyond the base, got %d", added)
	}
	fn, ok := r.resolve("tool-loan-repayment")
	if !ok || fn != "tool-loan-repayment" {
		t.Errorf("menu function must resolve, got (%q, %v)", fn, ok)
	}
	if !r.resolvedViaMenu("tool-loan-repayment") {
		t.Error("a function that owes its validity to the menu must be marked as such")
	}
	if r.resolvedViaMenu("hero") {
		t.Error("a base function must NOT be attributed to the menu")
	}
}

func TestAddMenu_DisplayNameFromMenuResolves(t *testing.T) {
	r := baseResolver()
	r.addMenu([]interface{}{toolMenuRow()})

	fn, ok := r.resolve("Standard Loan Repayment Calculator")
	if !ok || fn != "tool-loan-repayment" {
		t.Errorf("a planner echoing the menu's display name must resolve, got (%q, %v)", fn, ok)
	}
}

func TestAddMenu_BaseIdentityWinsOverAMenuRow(t *testing.T) {
	// A menu row claiming an existing display name must not repoint it: the DB
	// base is the stronger statement of identity.
	r := baseResolver()
	r.addMenu([]interface{}{map[string]interface{}{
		"function":     "tool-impostor",
		"display_name": "FAQ Section",
	}})

	if fn, _ := r.resolve("FAQ Section"); fn != "faq" {
		t.Errorf("base display mapping must win, got %q", fn)
	}
}

func TestAddMenu_RowsWithoutAFunctionAreIgnored(t *testing.T) {
	r := baseResolver()
	added := r.addMenu([]interface{}{
		map[string]interface{}{"name": "no function here"},
		map[string]interface{}{"function": "   "},
		"a bare string, not a row",
		nil,
		42,
	})
	if added != 0 {
		t.Errorf("malformed rows must add nothing, got %d", added)
	}
	if r.validFunctions[""] {
		t.Error("the empty function must never become valid")
	}
	if _, ok := r.resolve(""); ok {
		t.Error("resolve(\"\") must stay false")
	}
}

func TestAddMenu_CountsOnlyWhatTheBaseLacked(t *testing.T) {
	r := baseResolver()
	added := r.addMenu([]interface{}{
		map[string]interface{}{"function": "hero"}, // already in the base
		toolMenuRow(), // new
	})
	if added != 1 {
		t.Errorf("only functions beyond the base count, got %d", added)
	}
	if r.resolvedViaMenu("hero") {
		t.Error("a function already valid must not be attributed to the menu")
	}
}

// ── menuRowsFrom ────────────────────────────────────────────────────────────

func TestMenuRowsFrom_ShapeVariance(t *testing.T) {
	rows := []interface{}{toolMenuRow()}
	cases := []struct {
		name      string
		collected map[string]interface{}
		path      string
		wantOK    bool
	}{
		{"live shape: array under the output field", map[string]interface{}{"available_components": rows}, "available_components", true},
		{"absent path", map[string]interface{}{"available_components": rows}, "not_a_field", false},
		{"empty path", map[string]interface{}{"available_components": rows}, "", false},
		{"nil collected data", nil, "available_components", false},
		{"a JSON string, not an array", map[string]interface{}{"available_components": `[{"function":"tool-x"}]`}, "available_components", false},
		{"a map, not an array", map[string]interface{}{"available_components": map[string]interface{}{"rows": rows}}, "available_components", false},
		{"an empty array", map[string]interface{}{"available_components": []interface{}{}}, "available_components", false},
	}
	for _, tc := range cases {
		got, ok := menuRowsFrom(tc.collected, tc.path)
		if ok != tc.wantOK {
			t.Errorf("%s: ok = %v, want %v", tc.name, ok, tc.wantOK)
		}
		if !ok && got != nil {
			t.Errorf("%s: rows must be nil when not ok", tc.name)
		}
	}
}

// ── ValidateSitePlanAction end to end (sqlmock) ─────────────────────────────

// validateMenuParams builds the action params for a one-page plan whose
// sections mix a base component and a tool-level one. menuField == "" is the
// un-opted-in control.
func validateMenuParams(menuField string) ActionParams {
	config := map[string]interface{}{
		"plan_field":          "llm_plan",
		"validate_components": true,
	}
	if menuField != "" {
		config["menu_field"] = menuField
	}
	return ActionParams{
		Context: context.Background(),
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"llm_plan": map[string]interface{}{
				"pages": []interface{}{
					map[string]interface{}{
						"name": "index", "page_type": "index",
						"sections": []interface{}{"hero", "tool-loan-repayment", "faq"},
					},
				},
			},
			"available_components": []interface{}{
				map[string]interface{}{"function": "hero", "name": "hero"},
				toolMenuRow(),
			},
		},
		StepConfig: models.Step{Config: config},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "55555555-5555-5555-5555-555555555555",
			StepName:        "validate_plan",
		},
	}
}

// expectResolverQueries arms the two content_components reads validate makes
// before the resolve pass: the site-chrome strip and the resolver base.
func expectResolverQueries(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("component_level = 'site'").
		WillReturnRows(sqlmock.NewRows([]string{"name"}))
	mock.ExpectQuery("component_level IN").
		WillReturnRows(sqlmock.NewRows([]string{"function", "name", "display_name"}).
			AddRow("hero", "hero", "Hero").
			AddRow("faq", "faq", "FAQ Section"))
}

// sectionsOfFirstPage pulls the validated section list back out of the action's
// result — the artefact the plan is written from.
func sectionsOfFirstPage(t *testing.T, out interface{}) []string {
	t.Helper()
	plan, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a plan map: %T", out)
	}
	pages, ok := plan["pages"].([]interface{})
	if !ok || len(pages) == 0 {
		t.Fatalf("no pages in result: %+v", plan)
	}
	page := pages[0].(map[string]interface{})
	raw, _ := page["sections"].([]interface{})
	names := make([]string, 0, len(raw))
	for _, s := range raw {
		if str, ok := s.(string); ok {
			names = append(names, str)
		}
	}
	return names
}

func TestValidateSitePlan_MenuFieldKeepsAToolSectionTheBaseWouldDrop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	expectResolverQueries(mock)

	params := validateMenuParams("available_components")
	params.DB = db

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	got := sectionsOfFirstPage(t, out)
	want := []string{"hero", "tool-loan-repayment", "faq"}
	if len(got) != len(want) {
		t.Fatalf("sections = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sections = %v, want %v", got, want)
		}
	}
}

func TestValidateSitePlan_WithoutMenuFieldTheToolSectionIsStillDropped(t *testing.T) {
	// The negative control, and the whole point of the opt-in: an un-opted-in
	// caller (content-gap-planner's path, site-planner today) behaves EXACTLY
	// as it did before this change.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	expectResolverQueries(mock)

	params := validateMenuParams("")
	params.DB = db

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	got := sectionsOfFirstPage(t, out)
	for _, s := range got {
		if s == "tool-loan-repayment" {
			t.Fatalf("without menu_field the tool section must still be dropped, got %v", got)
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected hero+faq only, got %v", got)
	}
}

func TestValidateSitePlan_MenuFieldPointingAtNothingFallsBackToTheBase(t *testing.T) {
	// A misconfigured menu_field must degrade to today's behaviour, never to
	// "accept everything" and never to a hard failure of the planner run.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	expectResolverQueries(mock)

	params := validateMenuParams("no_such_step_output")
	params.DB = db

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action must not fail on a misconfigured menu_field: %v", err)
	}
	got := sectionsOfFirstPage(t, out)
	if len(got) != 2 {
		t.Fatalf("expected the base's hero+faq, got %v", got)
	}
}
