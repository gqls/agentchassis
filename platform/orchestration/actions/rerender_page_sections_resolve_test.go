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
