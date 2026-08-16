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
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

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

// ── the residual: an unresolvable name is DROPPED, and that must be durable ──
//
// bugs_open/282's council round 1, bug_historian (gating-adjacent, HIGH): the
// menu arm restores acceptance for one opted-in caller and leaves the generic
// silent-drop untouched — "any name that resolves to neither the DB base NOR
// the menu still vanishes with zero error surface, byte-identical to the bug
// being fixed". These pin the durable record that answers it, for BOTH the
// opted-in and the un-opted-in caller.

func TestValidateSitePlan_AnUnresolvableNameIsDroppedButTheActionStillSucceeds(t *testing.T) {
	// The shape bugs_open/039 records (a section naming a missing component):
	// it must still be removed from the plan — persisting it would put a name
	// nothing can build into site_plan_sections — and the run must not fail.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	expectResolverQueries(mock)

	params := validateMenuParams("available_components")
	params.DB = db
	page := params.CollectedData["llm_plan"].(map[string]interface{})["pages"].([]interface{})[0].(map[string]interface{})
	page["sections"] = []interface{}{"hero", "a-component-that-does-not-exist", "faq"}

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("a drop must never fail the run: %v", err)
	}
	got := sectionsOfFirstPage(t, out)
	for _, s := range got {
		if s == "a-component-that-does-not-exist" {
			t.Fatalf("an unresolvable name must not survive into the plan, got %v", got)
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected hero+faq, got %v", got)
	}
}

func TestRecordDroppedSectionNames_CountsWhatItWrites(t *testing.T) {
	// The negative ("a clean plan writes nothing") asserted where it is
	// observable. A sqlmock cannot carry this claim: it fails an EXPECTED call
	// that never came, not an UNEXPECTED call that did, so a recorder that
	// wrote on every run would satisfy it. Proven by mutation — deleting the
	// empty-drops guard left the mock-based version green.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))

	params := validateMenuParams("available_components")
	params.DB = db

	if n := recordDroppedSectionNames(context.Background(), params, nil, "available_components"); n != 0 {
		t.Errorf("no drops must attempt no rows, got %d", n)
	}
	if n := recordDroppedSectionNames(context.Background(), params, []droppedSectionName{}, ""); n != 0 {
		t.Errorf("an empty slice must attempt no rows, got %d", n)
	}
	drops := []droppedSectionName{{Page: "index", Name: "x"}, {Page: "about", Name: "y"}}
	if n := recordDroppedSectionNames(context.Background(), params, drops, ""); n != 2 {
		t.Errorf("two drops must attempt two rows, got %d", n)
	}
}

// TestValidateSitePlan_ADropReachesTheDurableRecord asserts the WIRING, not the
// shape: that a name dropped inside the action's resolve pass actually reaches
// the findings door. Written because the first version of these tests did NOT
// cover it — removing the line that collects a drop left every test passing,
// which is precisely the silent-loss shape this arm exists to remove. The
// assertion is a POSITIVE one (this INSERT happened), which is what a mock can
// honestly prove.
func TestValidateSitePlan_ADropReachesTheDurableRecord(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	expectResolverQueries(mock)
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := validateMenuParams("available_components")
	params.DB = db
	page := params.CollectedData["llm_plan"].(map[string]interface{})["pages"].([]interface{})[0].(map[string]interface{})
	page["sections"] = []interface{}{"hero", "a-component-that-does-not-exist", "faq"}

	if _, err := ValidateSitePlanAction(context.Background(), params); err != nil {
		t.Fatalf("action failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a dropped section must reach the durable record: %v", err)
	}
}

func TestValidateSitePlan_NoDropMeansNoDurableWrite(t *testing.T) {
	// The control for the test above: a plan whose every name resolves must not
	// touch the findings table at all. Without this, the assertion above would
	// pass just as well against code that wrote a row on every run.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	expectResolverQueries(mock)

	params := validateMenuParams("available_components")
	params.DB = db
	page := params.CollectedData["llm_plan"].(map[string]interface{})["pages"].([]interface{})[0].(map[string]interface{})
	page["sections"] = []interface{}{"hero", "faq"}

	if _, err := ValidateSitePlanAction(context.Background(), params); err != nil {
		t.Fatalf("action failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a clean plan must issue no findings write: %v", err)
	}
}

func TestDroppedFindings_ShapeAndTheTwoRemedies(t *testing.T) {
	drops := []droppedSectionName{
		{Page: "index", Name: "tool-loan-repayment"},
		{Page: "about", Name: "a-typo-section"},
	}

	withMenu := droppedFindings(drops, "available_components")
	if len(withMenu) != 2 {
		t.Fatalf("one durable row per drop, got %d", len(withMenu))
	}
	f := withMenu[0]
	if f.ErrorCode != "PLAN_SECTION_NAME_DROPPED" {
		t.Errorf("error code = %q", f.ErrorCode)
	}
	if f.Severity != "warning" {
		// A drop IS a legal outcome — removing an unbuildable name is correct.
		// What it must not be is invisible. Matching recordRecomposeOutcomes.
		t.Errorf("severity = %q, want warning", f.Severity)
	}
	if f.Context["page"] != "index" || f.Context["section"] != "tool-loan-repayment" {
		t.Errorf("context must carry page+section, got %+v", f.Context)
	}
	if f.Context["menu_field"] != "available_components" {
		t.Errorf("context must record whether a menu was configured, got %+v", f.Context["menu_field"])
	}
	if !strings.Contains(f.Message, "tool-loan-repayment") || !strings.Contains(f.Message, "index") {
		t.Errorf("message must name both, got %q", f.Message)
	}

	// The two remedies are diagnostically different and must not be shared.
	configured := droppedRemedy("available_components")
	unconfigured := droppedRemedy("")
	if configured == unconfigured {
		t.Fatal("the configured and unconfigured remedies must differ")
	}
	if strings.Contains(configured, "bugs_open/282") {
		t.Error("a configured menu must not blame the 282 shape")
	}
	if !strings.Contains(unconfigured, "menu_field") {
		t.Errorf("an unconfigured menu must name the missing key, got %q", unconfigured)
	}
}

// ── round 3: the class, not the caller ──────────────────────────────────────
//
// Council round 2, bug_historian (gating, HIGH): the durable record was wired
// into validate only, while apply_gap_plan's three call sites share the SAME
// resolver and lost names exactly as silently — on the fleet's dominant
// placement path. These pin the shared recorder and its provenance.

func TestRecordDroppedSectionNamesFor_NamesItsProvenanceExplicitly(t *testing.T) {
	// The gap-plan sites have no ActionParams to inherit provenance from, so
	// the entry must SAY what it is. An entry whose agent_type/action were
	// never stated is one the writer's merge can fill for you — the trap the
	// estate's landmines flag, and the reason this path does not reuse the
	// inheriting door.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"content-gap-planner", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"apply_gap_plan:new_page", sqlmock.AnyArg(), "PLAN_SECTION_NAME_DROPPED",
			"warning", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	n := recordDroppedSectionNamesFor(context.Background(), db, zap.NewNop(),
		"11111111-1111-1111-1111-111111111111", "apply_gap_plan:new_page",
		[]droppedSectionName{{Page: "guides", Name: "ghost-section"}}, "")
	if n != 1 {
		t.Errorf("one drop must attempt one row, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the row must carry the stated agent_type/action/code/severity: %v", err)
	}
}

func TestRecordDroppedSectionNamesFor_QuietWhenThereIsNothingToSay(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	if n := recordDroppedSectionNamesFor(context.Background(), db, zap.NewNop(), "site", "act", nil, ""); n != 0 {
		t.Errorf("no drops must attempt no rows, got %d", n)
	}
	// A nil DB must be survivable: these call sites run in paths where the
	// findings door is best-effort and must never change the disposition.
	if n := recordDroppedSectionNamesFor(context.Background(), nil, zap.NewNop(), "site", "act",
		[]droppedSectionName{{Page: "p", Name: "n"}}, ""); n != 0 {
		t.Errorf("a nil DB must attempt no rows and not panic, got %d", n)
	}
}

func TestWarnUnrecordedDrops_OnlyWarnsWhenARecordWasLost(t *testing.T) {
	// The "a failed record is itself silent" objection (round 2, bug_historian
	// LOW). The findings door is best-effort by design, so the shortfall is
	// made loud rather than fatal — failing a plan because its REPORT failed
	// would be worse than the thing being reported.
	core, logs := observer.New(zap.WarnLevel)
	logger := zap.New(core)

	warnUnrecordedDrops(3, 3, logger)
	if logs.Len() != 0 {
		t.Errorf("a fully recorded batch must be silent, got %d entries", logs.Len())
	}
	warnUnrecordedDrops(3, 1, logger)
	if logs.Len() != 1 {
		t.Fatalf("a shortfall must warn exactly once, got %d", logs.Len())
	}
	if !strings.Contains(logs.All()[0].Message, "could not be recorded durably") {
		t.Errorf("the warning must say a record was lost, got %q", logs.All()[0].Message)
	}
	warnUnrecordedDrops(1, 0, nil) // must not panic
}

// TestApplyNewPage_ADroppedSectionReachesTheDurableRecord asserts the WIRING at
// the gap-plan site, not just at validate. Added because mutation showed the
// existing gap-plan tests stay green with the recorder call deleted — the same
// silence that let the original bug live. content-gap-planner is the fleet's
// dominant placement path (116 runs/30d against build-site-planner's handful),
// so this is the site where the class fix actually earns its keep.
func TestApplyNewPage_ADroppedSectionReachesTheDurableRecord(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	// The resolver's base: two real components, neither of them the name the
	// plan proposes.
	mock.ExpectQuery("component_level IN").
		WillReturnRows(sqlmock.NewRows([]string{"function", "name", "display_name"}).
			AddRow("hero", "hero", "Hero").
			AddRow("faq", "faq", "FAQ Section"))
	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	// The claim under test: the unresolvable name is recorded, not merely logged.
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))

	plan := map[string]interface{}{
		"approach": "new_page",
		"new_page": map[string]interface{}{
			"name":     "guides",
			"sections": []interface{}{"hero", "a-component-that-does-not-exist"},
		},
	}
	if _, err := applyNewPage(context.Background(), db, plan, uuid.New(), "example.com", nil, zap.NewNop()); err != nil {
		t.Fatalf("applyNewPage: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a dropped section on the gap-plan path must reach the durable record: %v", err)
	}
}
