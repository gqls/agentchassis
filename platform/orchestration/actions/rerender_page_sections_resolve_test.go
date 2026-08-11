// FILE: platform/orchestration/actions/rerender_page_sections_resolve_test.go
//
// bugs_open/094 — the single-page deploy script's `section_data_resolved` branch
// could not run at all. `rerender_page_sections` declared page_name REQUIRED, so
// the input-spec check rejected the envelope with
//
//	step rerender_sections failed: failed to execute action rerender_page_sections:
//	input extraction failed: missing required fields: [page_name]
//
// before the action did any work — even though its very next act is a DB lookup
// that could have resolved the name from the page_id the caller DID send.
//
// The fix is candidate 1: either key is sufficient and the action derives the
// other, which fixes every caller at once rather than the one that was noticed.
// These tests pin both directions and the site scoping.
package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// resolveParams builds the ActionParams shape the action reads. The live caller
// (page-rerender's rerender_sections step) maps page_name explicitly via config;
// the 049b script's envelope carries page_id in input_data and no page_name.
func resolveParams(db *sql.DB, collected map[string]interface{}, cfg map[string]interface{}) ActionParams {
	return ActionParams{
		Context:       context.Background(),
		DB:            db,
		Logger:        zap.NewNop(),
		CollectedData: collected,
		StepConfig:    models.Step{Config: cfg},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "22222222-2222-2222-2222-222222222222",
			StepName:        "rerender_sections",
		},
	}
}

// TestRerenderPageSections_ResolvesByPageIDWhenNameAbsent is the regression: the
// 049b envelope shape, which used to be rejected outright.
func TestRerenderPageSections_ResolvesByPageIDWhenNameAbsent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	// The lookup must be BY ID and scoped to the site.
	mock.ExpectQuery("SELECT p.id, s.domain").
		WithArgs(siteID, pageID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain", "url", "name"}).
			AddRow(pageID, "example.com", "/tools/x.html", "tool-x"))
	// Everything after resolution is out of scope here; an empty section set
	// short-circuits the action without further queries we need to model.
	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"component_id", "slot_name", "content_data", "rendered_html", "position"}))

	p := resolveParams(db, map[string]interface{}{
		"input_data": map[string]interface{}{
			"page_id": pageID.String(),
			"site_id": siteID.String(),
			"domain":  "example.com",
			"spec":    map[string]interface{}{"reason": "section_data_resolved"},
		},
	}, map[string]interface{}{
		"target_site_id": "input_data.site_id",
		"reason":         "input_data.spec.reason",
		// deliberately NO page_name mapping — this is the 049b shape
	})

	_, err = RerenderPageSectionsAction(context.Background(), p)
	if err != nil && strings.Contains(err.Error(), "missing required fields") {
		t.Fatalf("the envelope was rejected before the action ran — bugs_open/094 is not fixed: %v", err)
	}
	if err != nil && strings.Contains(err.Error(), "page_name is required") {
		t.Fatalf("still demanding page_name when page_id was supplied: %v", err)
	}
}

// TestRerenderPageSections_RefusesWhenNeitherKeyGiven: dropping page_name from
// Required must not mean "no key needed". The error has to name both options.
func TestRerenderPageSections_RefusesWhenNeitherKeyGiven(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	p := resolveParams(db, map[string]interface{}{
		"input_data": map[string]interface{}{"site_id": uuid.New().String()},
	}, map[string]interface{}{"target_site_id": "input_data.site_id"})

	_, err = RerenderPageSectionsAction(context.Background(), p)
	if err == nil {
		t.Fatal("expected a refusal when neither page_name nor page_id is supplied")
	}
	if !strings.Contains(err.Error(), "page_name") || !strings.Contains(err.Error(), "page_id") {
		t.Errorf("the refusal must name BOTH accepted keys, got: %v", err)
	}
}

// TestRerenderPageSections_PageIDCannotReachPastTheSite is the scoping test.
// page_id is globally unique, so without the site predicate it would be a way to
// re-render another site's page through an envelope naming this one.
func TestRerenderPageSections_PageIDCannotReachPastTheSite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	foreignPageID := uuid.New()

	// Scoped lookup finds nothing: the page belongs to a different site.
	mock.ExpectQuery("SELECT p.id, s.domain").
		WithArgs(siteID, foreignPageID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain", "url", "name"}))

	p := resolveParams(db, map[string]interface{}{
		"input_data": map[string]interface{}{
			"page_id": foreignPageID.String(),
			"site_id": siteID.String(),
		},
	}, map[string]interface{}{"target_site_id": "input_data.site_id"})

	_, err = RerenderPageSectionsAction(context.Background(), p)
	if err == nil {
		t.Fatal("a page_id from another site must not resolve")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected a not-found refusal, got: %v", err)
	}
}

// ============================================================================
// bugs_open/182 — a re-render that resolved NO component for a section used
// to carry the stored HTML and complete successfully; nothing distinguished
// that from a real re-render. These tests pin the fix: resolve by
// component_id first (loadComponentSchemasByID), fall back to slot_name, and
// fail the step — naming every unresolved slot — when a section resolves via
// NEITHER route. "Not ready" and "empty template" carries stay legitimate and
// non-fatal.
// ============================================================================

// componentColumns is the column order scanSectionComponentRow expects, once,
// so every fixture below stays honest about column position.
var componentColumns = []string{
	"id", "name", "display_name", "function", "category", "semantic_tags",
	"description", "html_template", "input_schema", "render_mode", "agent_type",
	"component_level",
}

func emptyComponentRows() *sqlmock.Rows {
	return sqlmock.NewRows(componentColumns)
}

// expectPageAndSections wires the two queries every test below needs first:
// the page resolve, then loadStoredSections' page_components read.
func expectPageAndSections(mock sqlmock.Sqlmock, siteID uuid.UUID, pageID uuid.UUID, pageName, domain string, section sqlmock.Rows) {
	mock.ExpectQuery("SELECT p.id, s.domain").
		WithArgs(siteID, pageName).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain", "url", "name"}).
			AddRow(pageID, domain, "/x.html", pageName))
	mock.ExpectQuery("FROM page_components").WillReturnRows(&section)
}

func rerenderParams(db *sql.DB, siteID uuid.UUID, pageName string) ActionParams {
	return resolveParams(db, map[string]interface{}{
		"input_data": map[string]interface{}{
			"site_id": siteID.String(),
			"spec":    map[string]interface{}{"reason": "section_data_resolved", "page_name": pageName},
		},
	}, map[string]interface{}{
		"target_site_id": "input_data.site_id",
		"page_name":      "input_data.spec.page_name",
		"reason":         "input_data.spec.reason",
	})
}

func expectBaseSiteData(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("FROM sites WHERE id").
		WillReturnRows(sqlmock.NewRows([]string{"content_data", "email"}).AddRow(nil, ""))
}

// TestRerenderPageSections_FailsWhenComponentUnresolvedByNameOrID is the 182
// regression: a positional slot_name (no component_id at all, matching
// loancalculator.co.uk's shape) resolves via neither name/function nor id.
func TestRerenderPageSections_FailsWhenComponentUnresolvedByNameOrID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	expectPageAndSections(mock, siteID, pageID, "tool-loan-vs-savings", "loancalculator.co.uk",
		*sqlmock.NewRows([]string{"component_id", "slot_name", "content_data", "rendered_html", "position"}).
			AddRow("", "tool-2", []byte(`{"foo":"bar"}`), "<section>old</section>", 2))

	// loadComponentSchemas: name pass then function pass, both miss — "tool-2"
	// is a positional slot name, not any component's identity.
	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())
	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())
	// component_id is empty, so loadComponentSchemasByID issues no query at all.

	expectBaseSiteData(mock)

	p := rerenderParams(db, siteID, "tool-loan-vs-savings")

	_, err = RerenderPageSectionsAction(context.Background(), p)
	if err == nil {
		t.Fatal("expected the step to fail when a section's component resolves via neither name nor id")
	}
	if !strings.Contains(err.Error(), "tool-2") {
		t.Errorf("error must name the unresolved slot, got: %v", err)
	}
	if !strings.Contains(err.Error(), "bugs_open/182") {
		t.Errorf("error must cite bugs_open/182, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestRerenderPageSections_ResolvesToolByComponentIDWithoutEscalating is the
// loancalculator repair shape (candidate 2) combined with the 024 end-state
// flip (candidate 4 from the plan): a positional slot_name with NO name/
// function match, but a component_id that resolves to an active, self-
// contained tool component — it must render from the template, not escalate
// the page to the writer for "missing content_data".
func TestRerenderPageSections_ResolvesToolByComponentIDWithoutEscalating(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()
	componentID := uuid.New().String()

	expectPageAndSections(mock, siteID, pageID, "tool-loan-vs-savings", "loancalculator.co.uk",
		*sqlmock.NewRows([]string{"component_id", "slot_name", "content_data", "rendered_html", "position"}).
			AddRow(componentID, "tool-2", []byte(`{}`), "<section>old</section>", 2))

	// name/function passes both miss.
	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())
	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())

	// by-id resolution hits: an active, self-contained tool component.
	mock.ExpectQuery("FROM content_components").WillReturnRows(
		emptyComponentRows().AddRow(
			componentID, "tool-loan-vs-savings", "Pay Off Loan or Save?", "tool-loan-vs-savings",
			"", nil, nil, "<section>hi</section>", nil, "template", nil, "tool"))

	expectBaseSiteData(mock)

	p := rerenderParams(db, siteID, "tool-loan-vs-savings")

	out, err := RerenderPageSectionsAction(context.Background(), p)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	m := out.(map[string]interface{})
	if m["escalated"] == true {
		t.Error("a self-contained tool with resolvable component_id must not escalate to the writer")
	}
	if got := m["rerendered"]; got != 1 {
		t.Errorf("expected rerendered=1, got %v", got)
	}
	if got := m["carried"]; got != 0 {
		t.Errorf("expected carried=0, got %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestRerenderPageSections_ComponentIDWinsOverNameWhenBothResolve pins the
// resolution ORDER: when a slot_name coincidentally matches a generic
// component by name AND the row's own component_id resolves to a different
// component, the id's template must win — the row is telling you exactly
// which component it is; the name match is a coincidence.
func TestRerenderPageSections_ComponentIDWinsOverNameWhenBothResolve(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()
	nameMatchID := uuid.New().String()
	pinnedID := uuid.New().String()

	expectPageAndSections(mock, siteID, pageID, "index", "webdesign.co.uk",
		*sqlmock.NewRows([]string{"component_id", "slot_name", "content_data", "rendered_html", "position"}).
			AddRow(pinnedID, "hero", []byte(`{"headline":"hi"}`), "<section>old</section>", 0))

	// Pass 1 (name) resolves — to a different, generic component.
	mock.ExpectQuery("FROM content_components").WillReturnRows(
		emptyComponentRows().AddRow(
			nameMatchID, "hero", "Generic Hero", "hero", "", nil, nil, "<section>GENERIC</section>", nil, "template", nil, "section"))

	// by-id resolves to the page's own pinned, different component.
	mock.ExpectQuery("FROM content_components").WillReturnRows(
		emptyComponentRows().AddRow(
			pinnedID, "webdesign.co.uk Two-Column Hero", "webdesign.co.uk Two-Column Hero", "hero",
			"", nil, nil, "<section>PINNED</section>", nil, "template", nil, "section"))

	expectBaseSiteData(mock)

	p := rerenderParams(db, siteID, "index")

	out, err := RerenderPageSectionsAction(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]interface{})
	sections, _ := m["sections_metadata"].([]map[string]interface{})
	if len(sections) != 1 {
		t.Fatalf("expected exactly one section, got %d", len(sections))
	}
	html, _ := sections[0]["rendered_html"].(string)
	if !strings.Contains(html, "PINNED") {
		t.Errorf("expected the page's own pinned component to win, got: %s", html)
	}
	if strings.Contains(html, "GENERIC") {
		t.Error("the generic name-matched template must not be used when component_id resolves")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestRerenderPageSections_InvalidTemplateByID_IsFatalAndNamed covers
// bugs_open/024's second route into an empty schemas map: the row's pinned
// component_id resolves to a row that EXISTS but fails the template-
// truncation guard. This must be fatal (not a silent carry) and named
// distinctly from "no component at all", because the remediations differ.
func TestRerenderPageSections_InvalidTemplateByID_IsFatalAndNamed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()
	componentID := uuid.New().String()
	// Long, with an opened <div> and no closing </div> — fails
	// toolTemplateValid's tag-balance check (component_level="tool").
	brokenTemplate := "<div>" + strings.Repeat("x", 120)

	expectPageAndSections(mock, siteID, pageID, "tool-x", "example.com",
		*sqlmock.NewRows([]string{"component_id", "slot_name", "content_data", "rendered_html", "position"}).
			AddRow(componentID, "tool-2", []byte(`{"foo":"bar"}`), "<section>old</section>", 0))

	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())
	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())

	mock.ExpectQuery("FROM content_components").WillReturnRows(
		emptyComponentRows().AddRow(
			componentID, "tool-x", "Tool X", "tool-x", "", nil, nil, brokenTemplate, nil, "template", nil, "tool"))

	expectBaseSiteData(mock)

	p := rerenderParams(db, siteID, "tool-x")

	_, err = RerenderPageSectionsAction(context.Background(), p)
	if err == nil {
		t.Fatal("expected the step to fail when the pinned component fails the template guard")
	}
	if !strings.Contains(err.Error(), "tool-2") {
		t.Errorf("error must name the slot, got: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid template") {
		t.Errorf("error must distinguish an invalid-template drop from a plain miss, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestRerenderPageSections_EmptyTemplateCarriesWithoutFailing pins the other
// side of the boundary: an empty html_template is a legitimate, evidenced
// fallback (an intentional stub), not the silent-no-op class 182 is about —
// it must carry, surface itself in the output, and NOT fail the step.
func TestRerenderPageSections_EmptyTemplateCarriesWithoutFailing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()
	componentID := uuid.New().String()

	expectPageAndSections(mock, siteID, pageID, "tool-x", "example.com",
		*sqlmock.NewRows([]string{"component_id", "slot_name", "content_data", "rendered_html", "position"}).
			AddRow(componentID, "tool-2", []byte(`{"foo":"bar"}`), "<section>old</section>", 0))

	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())
	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())

	// html_template NULL: scanSectionComponentRow omits the map key entirely
	// for a NULL/empty template, which is what drives the empty-template
	// branch (as opposed to the component simply not being found).
	mock.ExpectQuery("FROM content_components").WillReturnRows(
		emptyComponentRows().AddRow(
			componentID, "tool-x", "Tool X", "tool-x", "", nil, nil, nil, nil, "template", nil, "tool"))

	expectBaseSiteData(mock)

	p := rerenderParams(db, siteID, "tool-x")

	out, err := RerenderPageSectionsAction(context.Background(), p)
	if err != nil {
		t.Fatalf("an empty html_template must be a legitimate, non-fatal carry: %v", err)
	}
	m := out.(map[string]interface{})
	if got := m["carried"]; got != 1 {
		t.Errorf("expected carried=1, got %v", got)
	}
	if got := m["rerendered"]; got != 0 {
		t.Errorf("expected rerendered=0, got %v", got)
	}
	reasons, ok := m["carried_empty_template"].([]string)
	if !ok || len(reasons) != 1 || !strings.Contains(reasons[0], "tool-2") {
		t.Errorf("expected carried_empty_template to name the slot, got: %v", m["carried_empty_template"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// ── rerenderResolution predicate/describe unit tests (no DB) ────────────────

func TestRerenderResolution_FatalOnlyOnUnresolvedOrInvalidTemplate(t *testing.T) {
	cases := []struct {
		name  string
		r     rerenderResolution
		fatal bool
	}{
		{"nothing carried", rerenderResolution{}, false},
		{"only legitimate carries", rerenderResolution{
			NotReadySlots: []string{"a (pos 0)"}, EmptyTemplateSlots: []string{"b (pos 1)"},
		}, false},
		{"one unresolved slot", rerenderResolution{UnresolvedSlots: []string{"c (pos 2)"}}, true},
		{"one invalid-template slot", rerenderResolution{InvalidTemplateSlots: []string{"d (pos 3)"}}, true},
		{"mixed fatal and legitimate", rerenderResolution{
			UnresolvedSlots: []string{"e (pos 4)"}, NotReadySlots: []string{"f (pos 5)"},
		}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.r.fatal(); got != c.fatal {
				t.Errorf("fatal() = %v, want %v", got, c.fatal)
			}
		})
	}
}

func TestRerenderResolution_DescribeNamesEveryList(t *testing.T) {
	r := rerenderResolution{
		UnresolvedSlots:      []string{"tool-2 (pos 2)"},
		InvalidTemplateSlots: []string{"tool-3 (pos 3)"},
		NotReadySlots:        []string{"prose-1 (pos 1)"},
		EmptyTemplateSlots:   []string{"prose-4 (pos 4)"},
	}
	desc := r.describe()
	for _, want := range []string{"tool-2 (pos 2)", "tool-3 (pos 3)", "prose-1 (pos 1)", "prose-4 (pos 4)"} {
		if !strings.Contains(desc, want) {
			t.Errorf("describe() missing %q: %s", want, desc)
		}
	}
}

func TestSlotLabel_NamesPositionForDuplicateSlotNames(t *testing.T) {
	a := slotLabel(storedSection{slotName: "generic-text-block", position: 0})
	b := slotLabel(storedSection{slotName: "generic-text-block", position: 3})
	if a == b {
		t.Errorf("two sections sharing a slot_name must not produce the same label: %q", a)
	}
}

// TestRerenderPageSections_StructuralCarryMakesANotReadySectionRerender pins the
// one DELIBERATE behaviour change bugs_open/238's plan-time carry makes on this
// path, in the direction that is easy to lose by accident.
//
// Before the carry, a section whose required non-llm field could not resolve
// planned `deferred`, took the not-ready branch above, and had its STORED HTML
// carried unchanged — which on a page already damaged means faithfully
// re-shipping the damage, and on an undamaged page means a template fix silently
// never lands. With the carry, the page's own stored content_data satisfies the
// field, the section plans `ready`, and it genuinely re-renders.
//
// This is pinned rather than merely allowed because both outcomes report success
// (`complete`, no error) and differ only in `carried` vs `rerendered` — the exact
// pair of counters bugs_open/182 exists because nobody was reading.
func TestRerenderPageSections_StructuralCarryMakesANotReadySectionRerender(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// planSection iterates the schema fields map, so query order is randomised.
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	pageID := uuid.New()
	componentID := uuid.New().String()

	// The stored row still holds the URL its declared source can no longer
	// resolve — the finetuning shape, one section wide.
	expectPageAndSections(mock, siteID, pageID, "index", "example.com",
		*sqlmock.NewRows([]string{"component_id", "slot_name", "content_data", "rendered_html", "position"}).
			AddRow(componentID, "case-studies-grid",
				[]byte(`{"card1_image_url":"/assets/images/case-study-facilities.jpg"}`),
				"<section>stale</section>", 0))

	mock.ExpectQuery("FROM content_components").WillReturnRows(
		emptyComponentRows().AddRow(
			componentID, "case-studies-grid", "Case studies", "case-studies-grid", "", nil, nil,
			`<img src="{{.card1_image_url}}" />`,
			[]byte(`{"fields":{"card1_image_url":{"type":"url","source":"site_assets.image","required":true}}}`),
			"template", nil, "section"))

	// The declared source resolves to nothing: no plan imagery, no content_data
	// hero. (No site_specs read — a site_assets.* source never reaches
	// ensureSpecs, and sqlmock would fail on an expectation nothing satisfies.)
	mock.ExpectQuery("site_plan_imagery").WillReturnRows(sqlmock.NewRows([]string{"key", "url", "scope", "scope_ref", "kind"}))
	mock.ExpectQuery("FROM sites").WillReturnRows(sqlmock.NewRows([]string{"content_data"}))
	// The carry preload — the page's own deployed row.
	mock.ExpectQuery("build_status = 'deployed'").
		WillReturnRows(sqlmock.NewRows([]string{"slot_name", "content_data"}).
			AddRow("case-studies-grid", []byte(`{"card1_image_url":"/assets/images/case-study-facilities.jpg"}`)))
	expectBaseSiteData(mock)

	out, err := RerenderPageSectionsAction(context.Background(), rerenderParams(db, siteID, "index"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]interface{})
	if got := m["rerendered"]; got != 1 {
		t.Errorf("expected the carried key to make this section RE-RENDER (rerendered=1), got rerendered=%v carried=%v not_ready=%v",
			got, m["carried"], m["carried_not_ready"])
	}
	if got := m["carried"]; got != 0 {
		t.Errorf("expected carried=0 — carrying the stored HTML here re-ships whatever the stored HTML already had wrong; got %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
