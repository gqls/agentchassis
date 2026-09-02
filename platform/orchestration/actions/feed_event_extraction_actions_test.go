// FILE: platform/orchestration/actions/feed_event_extraction_actions_test.go
//
// bugs_open/427 fix candidate #1. Two properties matter and neither is
// obvious from reading the SQL in isolation:
//
//  1. the loader's WHERE clause is status='relevant' AND event_extracted_at
//     IS NULL — NOT status='ingested' (that's feed-triage's own loader) and
//     NOT a plain "unprocessed" check on processed_at, which already means
//     something else (triaged). Getting either wrong either re-triages
//     un-scored items or silently never re-checks anything.
//  2. the marker only ever narrows (site_id + id = ANY(...) + still NULL) —
//     it can never stamp a row event_extracted_at twice, and it can never
//     touch a row outside the requested site.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TestLoadFeedItemsForEventExtraction_FiltersRelevantAndUnconsidered pins the
// WHERE clause. MUTATION THAT MUST BREAK IT: change 'relevant' to 'ingested'
// (collides with feed-triage's own loader, re-triaging unscored items) or
// drop the event_extracted_at IS NULL predicate (re-sends every relevant item
// to the LLM every cycle forever).
func TestLoadFeedItemsForEventExtraction_FiltersRelevantAndUnconsidered(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	itemID := uuid.New()

	mock.ExpectQuery(`SELECT(.|\n)*FROM content_feed_items cfi(.|\n)*WHERE(.|\n)*cfi\.site_id = \$1(.|\n)*cfi\.status = 'relevant'(.|\n)*cfi\.event_extracted_at IS NULL`).
		WithArgs(siteID.String(), 50).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "source_title", "source_summary", "source_content", "source_url",
			"published_at", "source_name", "topics",
		}).AddRow(
			itemID.String(), "Confirmed: title fight set for March 14", "summary text", "full body",
			"https://example.com/article", "2026-03-01", "Example Wire", `["boxing","fight-card"]`,
		))

	out, err := LoadFeedItemsForEventExtractionAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"site_id": siteID.String(),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result shape %#v", out)
	}
	if m["count"] != 1 {
		t.Errorf("count = %v, want 1", m["count"])
	}
	ids, ok := m["ids"].([]string)
	if !ok || len(ids) != 1 || ids[0] != itemID.String() {
		t.Errorf("ids = %#v, want [%s]", m["ids"], itemID.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet or unexpected queries: %v", err)
	}
}

// TestMarkFeedItemsEventExtracted_NarrowsToSiteAndStillNull pins the UPDATE's
// safety predicate. MUTATION THAT MUST BREAK IT: drop the
// "AND event_extracted_at IS NULL" clause — a re-run (e.g. a retried step
// after a partial failure) would then re-stamp an already-marked row, which
// is harmless today but signals the guard is gone; drop "AND site_id = $2" —
// an id collision across sites (uuids don't collide in practice, but the
// predicate is the stated invariant, and a test that doesn't check it lets a
// future edit remove it silently) would let one site's step touch another's row.
func TestMarkFeedItemsEventExtracted_NarrowsToSiteAndStillNull(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	id1, id2 := uuid.New(), uuid.New()

	mock.ExpectExec(`UPDATE content_feed_items(.|\n)*SET event_extracted_at = \$1(.|\n)*WHERE(.|\n)*site_id = \$2(.|\n)*id = ANY\(\$3::uuid\[\]\)(.|\n)*event_extracted_at IS NULL`).
		WithArgs(sqlmock.AnyArg(), siteID.String(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))

	out, err := MarkFeedItemsEventExtractedAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{"ids": "extracted.ids"}},
		CollectedData: map[string]interface{}{
			"site_id": siteID.String(),
			"extracted": map[string]interface{}{
				"ids": []interface{}{id1.String(), id2.String()},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result shape %#v", out)
	}
	if m["marked"] != int64(2) {
		t.Errorf("marked = %v, want 2", m["marked"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet or unexpected queries: %v", err)
	}
}

// TestMarkFeedItemsEventExtracted_EmptyIDsIsANoOp: an extraction pass over
// zero items must not issue an UPDATE at all — proves the empty-list guard
// rather than trusting an UPDATE ... = ANY('{}') to be harmless.
func TestMarkFeedItemsEventExtracted_EmptyIDsIsANoOp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	out, err := MarkFeedItemsEventExtractedAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{"ids": "extracted.ids"}},
		CollectedData: map[string]interface{}{
			"site_id":   siteID.String(),
			"extracted": map[string]interface{}{"ids": []interface{}{}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := out.(map[string]interface{})
	if m["marked"] != 0 {
		t.Errorf("marked = %v, want 0", m["marked"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("no query should have been issued for an empty id list: %v", err)
	}
}
