package discovery_checks

import (
	"context"
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestMissingRequiredValueFields(t *testing.T) {
	// Modelled on gripper-detail's product-details_pre_037 schema: value
	// fields required from the LLM, chrome fields optional, one image field
	// and one query-sourced field that must be ignored.
	fields := map[string]interface{}{
		"product_name":  map[string]interface{}{"type": "text", "source": "llm", "required": true},
		"product_price": map[string]interface{}{"type": "text", "source": "llm", "required": true},
		"feature_1":     map[string]interface{}{"type": "text", "source": "llm", "required": "true"}, // string encoding
		"product_sku":   map[string]interface{}{"type": "text", "required": true},                    // no source → llm-ish, checked
		"sku_label":     map[string]interface{}{"type": "text", "source": "llm", "required": false},
		"main_image":    map[string]interface{}{"type": "image", "source": "site_assets.hero", "required": true},
		"products":      map[string]interface{}{"type": "array", "source": "query.affiliate_products", "required": true},
	}

	t.Run("chrome-only content_data misses every value field", func(t *testing.T) {
		content := map[string]interface{}{
			"sku_label":         "SKU:",
			"add_to_cart_label": "Add to Cart",
			"product_price":     "  ", // whitespace-only counts as missing
		}
		got := missingRequiredValueFields(fields, content)
		want := []string{"feature_1", "product_name", "product_price", "product_sku"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("filled content_data passes", func(t *testing.T) {
		content := map[string]interface{}{
			"product_name":  "PG-90 Parallel Gripper",
			"product_price": "£1,240",
			"feature_1":     "90mm stroke",
			"product_sku":   "PG-90",
		}
		if got := missingRequiredValueFields(fields, content); len(got) != 0 {
			t.Errorf("expected no missing fields, got %v", got)
		}
	})

	t.Run("zero and false are values, empty collections are not", func(t *testing.T) {
		fields := map[string]interface{}{
			"count":  map[string]interface{}{"type": "number", "source": "llm", "required": true},
			"active": map[string]interface{}{"type": "bool", "source": "llm", "required": true},
			"items":  map[string]interface{}{"type": "array", "source": "llm", "required": true},
		}
		content := map[string]interface{}{
			"count":  float64(0),
			"active": false,
			"items":  []interface{}{},
		}
		got := missingRequiredValueFields(fields, content)
		want := []string{"items"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

// ============================================================================
// Retraction guards (RFC_010 Decision 1 — second adopter)
// ============================================================================

const (
	rfPageID = "8c1d4e2f-9a3b-4c5d-8e7f-1a2b3c4d5e6f"
	rfSlot   = "product-details"
	// One required LLM value field — the smallest schema that can be missing.
	rfSchema = `{"fields":{"product_name":{"type":"text","source":"llm","required":true}}}`
	rfFilled = `{"product_name":"PG-90 Parallel Gripper"}`
	rfHollow = `{"sku_label":"SKU:","add_to_cart_label":"Add to Cart"}`
)

var rfKey = fmt.Sprintf("required_fields_missing:%s:%s", rfPageID, rfSlot)

var rfFilingColumns = []string{
	"id", "page_id", "name", "slot_name", "function", "input_schema", "content_data", "runtime_fill"}

// rfObs is one row of the retraction observation query. A nil componentID
// models the LEFT JOIN miss (slot holds no deployed component); a nil schema
// models a NULL input_schema.
type rfObs struct {
	componentID interface{}
	schema      interface{}
	content     string
	runtimeFill bool
}

func rfRetractionRows(obs ...rfObs) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"item_key", "id", "input_schema", "content_data", "runtime_fill"})
	for _, o := range obs {
		r.AddRow(rfKey, o.componentID, o.schema, o.content, o.runtimeFill)
	}
	return r
}

// rfFilingRows builds n components that each miss their one required field, so
// the filing half emits n findings (used to drive the per-pass cap).
func rfFilingRows(n int) *sqlmock.Rows {
	r := sqlmock.NewRows(rfFilingColumns)
	for i := 0; i < n; i++ {
		r.AddRow(uuid.New().String(), rfPageID, "shop", fmt.Sprintf("%s-%d", rfSlot, i),
			"product-details", rfSchema, rfHollow, false)
	}
	return r
}

func runRFCheck(t *testing.T, filing, retraction *sqlmock.Rows) (*CheckResult, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(filing)
	mock.ExpectQuery("FROM site_work_items").WillReturnRows(retraction)

	check := &RequiredFieldsMissingCheck{}
	res, err := check.Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(), Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return res, mock, func() { db.Close() }
}

// A site with ZERO missing-field findings is the site whose stale items most
// need closing. If someone adds a `len(findings) == 0` early return above the
// retraction, this is the test that fails.
func TestRequiredFieldsRetractionRunsWhenThereAreNoFindingsAtAll(t *testing.T) {
	res, mock, done := runRFCheck(t,
		sqlmock.NewRows(rfFilingColumns),
		rfRetractionRows(rfObs{componentID: "c1", schema: rfSchema, content: rfFilled}))
	defer done()

	if len(res.Findings) != 0 || len(res.WorkItems) != 0 {
		t.Fatalf("expected no findings/work items, got %d/%d", len(res.Findings), len(res.WorkItems))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("a filled slot must retract even when the run found nothing: got %d resolved", len(res.Resolved))
	}
	if res.Resolved[0].ItemKey != rfKey {
		t.Errorf("ItemKey = %q, want %q", res.Resolved[0].ItemKey, rfKey)
	}
	if res.Resolved[0].ItemType != "required_fields_missing" {
		t.Errorf("ItemType = %q, want required_fields_missing", res.Resolved[0].ItemType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// THE TRAP UNIQUE TO THIS CHECK. maxRequiredFieldFlagsPerPass used to `return`,
// which skipped the retraction on exactly the sites that hit the cap — the
// badly-shaped ones carrying the most stale items. It is invisible to every
// test that stays under the cap, which is all of the others here.
func TestRequiredFieldsRetractionSurvivesThePerPassCap(t *testing.T) {
	res, mock, done := runRFCheck(t,
		rfFilingRows(maxRequiredFieldFlagsPerPass+5),
		rfRetractionRows(rfObs{componentID: "c1", schema: rfSchema, content: rfFilled}))
	defer done()

	if len(res.WorkItems) != maxRequiredFieldFlagsPerPass {
		t.Errorf("filing must still stop at the cap: got %d work items, want %d",
			len(res.WorkItems), maxRequiredFieldFlagsPerPass)
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("hitting the per-pass cap must not skip the retraction: got %d resolved", len(res.Resolved))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the retraction query must still run after the cap: %v", err)
	}
}

// The refusals. `healthy` is incremented only where missingRequiredValueFields
// actually ran and returned nothing; everything else falls through to the
// healthy != deployed gate.
func TestRequiredFieldsRetractionRefusesEverythingItHasNotPositivelyObserved(t *testing.T) {
	cases := []struct {
		name        string
		obs         []rfObs
		wantResolve bool
		why         string
	}{
		{
			name:        "slot holds no deployed component",
			obs:         []rfObs{{componentID: nil, schema: nil, content: "{}"}},
			wantResolve: false,
			why: "absence is equally a fix and a silently deleted component " +
				"(bugs_open/032) — 3 of 59 open items sit in this state",
		},
		{
			// SYNTHETIC ROW SHAPE, AND DELIBERATELY SO. Today's query joins
			// content_components THROUGH page_components, so a LEFT JOIN miss
			// always arrives with a NULL schema too and the schema refusal below
			// would catch it anyway — the two guards sit in series. That makes the
			// realistic case above unable to tell them apart: deleting the
			// componentID check leaves every test green (verified by mutation).
			// This row — a miss whose other columns look perfectly healthy — is
			// the one that pins the componentID guard on its own, so a future
			// change to the join (resolving the component from spec.component_id,
			// say) cannot quietly make correctness depend on the other guard.
			name:        "left join miss is refused even when the joined columns look healthy",
			obs:         []rfObs{{componentID: nil, schema: rfSchema, content: rfFilled}},
			wantResolve: false,
			why:         "no deployed component was observed at all, whatever else the row carries",
		},
		{
			name:        "required field still absent",
			obs:         []rfObs{{componentID: "c1", schema: rfSchema, content: rfHollow}},
			wantResolve: false,
			why:         "the finding still reproduces — 50 of 59",
		},
		{
			name:        "input_schema is NULL",
			obs:         []rfObs{{componentID: "c1", schema: nil, content: rfFilled}},
			wantResolve: false,
			why: "no schema means no required set, which reads as 'nothing missing' — " +
				"it is the inverse: the predicate could not be re-run at all",
		},
		{
			name:        "input_schema is unparseable",
			obs:         []rfObs{{componentID: "c1", schema: `{not json`, content: rfFilled}},
			wantResolve: false,
			why:         "same blindness as a NULL schema, arriving by a different route",
		},
		{
			name:        "input_schema carries no recognised field wrapper",
			obs:         []rfObs{{componentID: "c1", schema: `{"title":"legacy"}`, content: rfFilled}},
			wantResolve: false,
			why:         "SchemaContentFields refuses it, so there is no required set to check against",
		},
		{
			name:        "content_data is unparseable",
			obs:         []rfObs{{componentID: "c1", schema: rfSchema, content: `{not json`}},
			wantResolve: false,
			why:         "the filing half falls back to an empty map; the retraction half must not guess",
		},
		{
			name:        "runtime-fill shell",
			obs:         []rfObs{{componentID: "c1", schema: rfSchema, content: rfHollow, runtimeFill: true}},
			wantResolve: false,
			why: "the filing half skips these, but that is a reason not to FILE — it is " +
				"not evidence the fields arrived, so this side has observed nothing",
		},
		{
			name: "mixed slot: one of three still missing",
			obs: []rfObs{
				{componentID: "c1", schema: rfSchema, content: rfFilled},
				{componentID: "c2", schema: rfSchema, content: rfHollow},
				{componentID: "c3", schema: rfSchema, content: rfFilled}},
			wantResolve: false,
			why: "bugs_open/156 means (page_id, slot_name) is not unique; the finding " +
				"may still be true of the unfilled one",
		},
		{
			name: "every component in a multi-component slot is filled",
			obs: []rfObs{
				{componentID: "c1", schema: rfSchema, content: rfFilled},
				{componentID: "c2", schema: rfSchema, content: rfFilled}},
			wantResolve: true,
			why:         "nothing in the slot reproduces the finding",
		},
		{
			name: "a filled component next to an unreadable one",
			obs: []rfObs{
				{componentID: "c1", schema: rfSchema, content: rfFilled},
				{componentID: "c2", schema: nil, content: rfFilled}},
			wantResolve: false,
			why:         "one positive observation does not cover a sibling it could not read",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, _, done := runRFCheck(t, sqlmock.NewRows(rfFilingColumns), rfRetractionRows(tc.obs...))
			defer done()
			if got := len(res.Resolved) > 0; got != tc.wantResolve {
				t.Errorf("retracted = %v, want %v — %s", got, tc.wantResolve, tc.why)
			}
		})
	}
}

// AllOfType is the wide, destructive branch: it would close every open
// required_fields_missing item for the site at once. This check observes one
// slot at a time and must never reach for it.
func TestRequiredFieldsRetractionNeverUsesTheWideBranch(t *testing.T) {
	res, _, done := runRFCheck(t, sqlmock.NewRows(rfFilingColumns),
		rfRetractionRows(rfObs{componentID: "c1", schema: rfSchema, content: rfFilled}))
	defer done()

	if len(res.Resolved) == 0 {
		t.Fatal("expected a retraction to inspect")
	}
	for _, r := range res.Resolved {
		if r.AllOfType {
			t.Error("required_fields_missing retraction must address one item_key, never AllOfType")
		}
		if r.ItemKey == "" {
			t.Error("ItemKey must be set — an empty key with AllOfType false is refused by the runner")
		}
		if r.Reason == "" {
			t.Error("Reason must be set — the runner refuses a retraction with no stated cause")
		}
	}
}

// A retraction fault must cost a missed CLOSURE, never a missed DEFECT: the
// runner skips a check's inserts when Run returns an error.
func TestRequiredFieldsRetractionFailureDoesNotSuppressFindings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM page_components pc").WillReturnRows(
		sqlmock.NewRows(rfFilingColumns).
			AddRow(uuid.New().String(), rfPageID, "shop", rfSlot, "product-details", rfSchema, rfHollow, false))
	mock.ExpectQuery("FROM site_work_items").WillReturnError(sql.ErrConnDone)

	check := &RequiredFieldsMissingCheck{}
	res, err := check.Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(), Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("a retraction fault must not fail the check: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Errorf("the missing-field finding must still be filed: got %d work items", len(res.WorkItems))
	}
	if len(res.Resolved) != 0 {
		t.Errorf("a failed observation must retract nothing: got %d", len(res.Resolved))
	}
}

// This check's filing half never sets WorkItemSpec.PageID, so every one of its
// 70 rows carries a NULL page_id column and the spec fallback is the arm that
// actually fires (measured 2026-08-03). The column arm still has to come first:
// if a future filing change populates it, a spec-only read would silently stop
// retracting. Dropping the COALESCE fails this.
func TestRequiredFieldsRetractionPrefersTheFirstClassPageIDColumn(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM page_components pc").WillReturnRows(sqlmock.NewRows(rfFilingColumns))
	mock.ExpectQuery(`COALESCE\(page_id::text, spec->>'page_id'\)`).
		WillReturnRows(rfRetractionRows(rfObs{componentID: "c1", schema: rfSchema, content: rfFilled}))

	check := &RequiredFieldsMissingCheck{}
	if _, err := check.Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(), Logger: zap.NewNop(),
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the retraction query must resolve page_id from the first-class column: %v", err)
	}
}

// Anti-vacuity: the table above is mostly negative cases, so a retraction that
// never fires at all would pass nine of its ten subtests. This pins that the
// mechanism does something, and that the reason names what was observed.
func TestRequiredFieldsRetractionIsNotVacuous(t *testing.T) {
	res, _, done := runRFCheck(t, sqlmock.NewRows(rfFilingColumns),
		rfRetractionRows(rfObs{componentID: "c1", schema: rfSchema, content: rfFilled}))
	defer done()

	if len(res.Resolved) != 1 {
		t.Fatalf("a filled, readable, single-component slot must retract: got %d", len(res.Resolved))
	}
	if reason := res.Resolved[0].Reason; reason == "" ||
		!containsAll(reason, "re-observed filled", "1 deployed component") {
		t.Errorf("reason must state what was observed, got %q", reason)
	}
}

// The whole refusal design rests on ONE discipline: obs.healthy is incremented
// only where missingRequiredValueFields actually ran and returned nothing, so
// every refusal works by NOT counting rather than by remembering to veto. That
// held by review discipline alone, which council 64430363's bug_historian seat
// objected to (low) as "the same class of latent exposure, just caught by
// review rather than mechanism". This is the mechanism.
//
// It parses the real source with go/ast rather than grepping, because a grep
// would also match the string inside a comment — and this file's comments talk
// about obs.healthy repeatedly. It cannot pass vacuously: it t.Fatal's if it
// cannot find the function, and again if it counts zero increments (which would
// mean the needle stopped matching, not that the discipline is held).
func TestHealthyIsIncrementedFromExactlyOnePlace(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "check_required_fields_missing.go", nil, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	var fn *ast.FuncDecl
	for _, d := range file.Decls {
		if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == "findResolvedRequiredFields" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatal("findResolvedRequiredFields not found — this test's needle has stopped matching, " +
			"so it would pass vacuously; fix the test, do not delete it")
	}

	writes := 0
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		isHealthy := func(e ast.Expr) bool {
			sel, ok := e.(*ast.SelectorExpr)
			return ok && sel.Sel.Name == "healthy"
		}
		switch s := n.(type) {
		case *ast.IncDecStmt: // obs.healthy++ / --
			if isHealthy(s.X) {
				writes++
			}
		case *ast.AssignStmt: // obs.healthy = x / += x
			for _, lhs := range s.Lhs {
				if isHealthy(lhs) {
					writes++
				}
			}
		}
		return true
	})

	if writes == 0 {
		t.Fatal("counted zero writes to .healthy — the AST needle has stopped matching, " +
			"so this test can no longer fail; fix it rather than trusting the pass")
	}
	if writes != 1 {
		t.Errorf("`.healthy` is written from %d places, want exactly 1.\n"+
			"Every refusal in findResolvedRequiredFields works by NOT incrementing healthy and "+
			"falling through to the `healthy != deployed` gate. A second write site means some "+
			"path now counts as a positive observation without having run the predicate — which "+
			"is how an unreadable schema or a runtime-fill shell starts reading as healthy.", writes)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
