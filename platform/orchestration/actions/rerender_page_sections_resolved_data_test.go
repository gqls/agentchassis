// FILE: platform/orchestration/actions/rerender_page_sections_resolved_data_test.go
//
// bugs_open/454 — the light re-render path silently stopped carrying FRESHLY
// RESOLVED field data from planSection into the render.
//
// WHAT BROKE. `94f81cc60` (2026-09-02, "035 P1: extract classifyStoredSection")
// moved `plan := planSection(...)` out of the per-section loop and into
// classifyStoredSection, declared a `plan sectionPlanItem` field on the returned
// sectionClassification, and had renderPlannedSection read it back
// (`comp, plan, htmlTemplate := cls.comp, cls.plan, cls.htmlTemplate`). The
// assignment `c.plan = plan` was never written, so the field was READ ONCE AND
// NEVER SET: every re-render since has rendered from `base ⊕ stored content_data`
// with plan.ResolvedData == nil, and persisted `mergedContent` = stored only.
//
// WHY IT IS INVISIBLE WITHOUT THIS TEST. Nothing errors, nothing is blanked and
// no section is carried: the stored content_data still renders, so the page
// keeps serving its last-good bytes and every count the action reports
// (`rerendered`, `carried`, `escalated`) is exactly what a healthy run reports.
// The only observable is a NEGATIVE — a re-render that changes nothing when the
// resolver's data has changed — which is precisely what bugs_open/427 hit:
// event-list's `items` (source `query.upcoming_events`) never populated, three
// dispatches running, byte-identical rendered_html every time.
//
// THE SHAPE OF THE ASSERTION. The field's source is not what matters — planSection
// writes every non-LLM resolution into the one ResolvedData map — but the test
// uses `query.upcoming_events` deliberately, because that reproduces 427's own
// case end to end and would have failed on the day the extraction landed.
package actions

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
)

// eventListComponent is boxingonline.com's live event-list contract, reduced to
// the one query-sourced field this test is about.
func eventListComponent() componentInfo {
	schema := map[string]interface{}{
		"fields": map[string]interface{}{
			"items": map[string]interface{}{
				"type":       "array",
				"source":     "query.upcoming_events",
				"required":   false,
				"on_missing": "skip_field",
				"limit":      float64(20),
			},
		},
	}
	tmpl := `<section class="event-list">` +
		`{{if .items}}{{range .items}}<li data-fact="{{.fact_id}}">{{.title}} — {{.date}}</li>{{end}}` +
		`{{else}}<p class="event-list-empty">No confirmed fixtures yet.</p>{{end}}</section>`
	return componentInfo{
		ID:          "3647c0c2-6564-489e-91ed-42145ad4f62d",
		Name:        "event-list",
		Function:    "event-list",
		InputSchema: schema,
		Raw: map[string]interface{}{
			"input_schema":  schema,
			"html_template": tmpl,
		},
	}
}

// evidenceBaseWithOneFutureFixture is the register row resolveUpcomingEvents
// reads. The date is computed, not literal, so the test cannot rot into a
// vacuous pass the way a hardcoded 2026-10-31 would once that date is past.
func evidenceBaseWithOneFutureFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]interface{}{
		"facts": []interface{}{
			map[string]interface{}{
				"id":           "CIT-testfixture000001",
				"claim":        "Canelo Alvarez faces Christian Mbilli",
				"event_date":   time.Now().UTC().AddDate(0, 0, 30).Format("2006-01-02"),
				"venue":        "Allegiant Stadium",
				"participants": []interface{}{"Canelo Alvarez", "Christian Mbilli"},
				"source": map[string]interface{}{
					"citation": map[string]interface{}{
						"url":   "https://example.test/fight-announced",
						"quote": "The bout is confirmed for the announced date.",
						"title": "Fight announced",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal evidence_base: %v", err)
	}
	return raw
}

// TestClassifyStoredSection_ReturnsTheResolvedDataItComputed is the direct
// regression: classifyStoredSection calls planSection and must hand the result
// back, or every caller renders against stored data alone.
func TestClassifyStoredSection_ReturnsTheResolvedDataItComputed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("FROM site_specs").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow(evidenceBaseWithOneFutureFixture(t)))

	comp := eventListComponent()
	resolver := newSourceResolver(siteID, db, zap.NewNop(), "tool-fight-calendar")
	s := storedSection{
		componentID: comp.ID,
		slotName:    "event-list",
		contentData: map[string]interface{}{"heading": "Fight calendar", "content": "old prose"},
		position:    2,
	}

	cls := classifyStoredSection(context.Background(), s, newSectionRef("event-list", 0),
		func(storedSection) (componentInfo, bool, bool) { return comp, false, true },
		resolver, zap.NewNop())

	if cls.carryKind != "" {
		t.Fatalf("section was carried (%s) — this test is about the RENDER path, not a carry", cls.carryKind)
	}
	if cls.plan.ResolvedData == nil {
		t.Fatal("classifyStoredSection computed a plan and threw its ResolvedData away — " +
			"every re-render renders from stored content_data alone (bugs_open/454)")
	}
	items, ok := cls.plan.ResolvedData["items"].([]map[string]interface{})
	if !ok || len(items) != 1 {
		t.Fatalf("resolved items = %#v, want the one qualifying future fixture", cls.plan.ResolvedData["items"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}

// TestRerenderFlatSections_FreshResolvedDataReachesHTMLAndPersistedRow is the
// same defect asserted where it is USER-VISIBLE: the fixture must appear in the
// rendered bytes AND in the content_data the save writes back. Asserting only
// the classification would leave the read side (renderPlannedSection's merge)
// unguarded.
func TestRerenderFlatSections_FreshResolvedDataReachesHTMLAndPersistedRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()
	mock.ExpectQuery("FROM site_specs").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow(evidenceBaseWithOneFutureFixture(t)))

	comp := eventListComponent()
	resolver := newSourceResolver(siteID, db, zap.NewNop(), "tool-fight-calendar")
	stored := []storedSection{{
		componentID:  comp.ID,
		slotName:     "event-list",
		contentData:  map[string]interface{}{"heading": "Fight calendar"},
		renderedHTML: `<section class="event-list"><p class="event-list-empty">No confirmed fixtures yet.</p></section>`,
		position:     2,
	}}

	params := ActionParams{
		Context:    context.Background(),
		DB:         db,
		Logger:     zap.NewNop(),
		StepConfig: models.Step{Config: map[string]interface{}{}},
	}

	outcome := rerenderFlatSections(context.Background(), params, stored,
		func(storedSection) (componentInfo, bool, bool) { return comp, false, true },
		resolver, map[string]interface{}{}, siteID, pageID,
		"tool-fight-calendar", "boxingonline.com", "/tools/fight-calendar.html",
		"section_data_resolved", nil, zap.NewNop())

	if outcome.reRendered != 1 || outcome.carried != 0 {
		t.Fatalf("rerendered=%d carried=%d, want 1/0", outcome.reRendered, outcome.carried)
	}
	entry := outcome.sectionsMetadata[0]

	html, _ := entry["rendered_html"].(string)
	if !strings.Contains(html, "CIT-testfixture000001") {
		t.Errorf("the rendered section does not carry the qualifying fixture — it rendered the "+
			"guarded empty state from stored data alone (bugs_open/454, the live symptom of "+
			"bugs_open/427). html = %q", html)
	}
	if strings.Contains(html, "event-list-empty") {
		t.Errorf("the rendered section still shows its empty state with one qualifying fixture resolved: %q", html)
	}

	merged, _ := entry["content_data"].(map[string]interface{})
	if merged["items"] == nil {
		t.Error("the persisted content_data gained no `items` key — the row is no longer a complete " +
			"render source, so the next re-render starts from the same empty state")
	}
	if merged["heading"] != "Fight calendar" {
		t.Errorf("the stored half of the merge was lost: %#v", merged)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}
