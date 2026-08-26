package actions

// The refresher's write half. The WHERE is the whole safety, so the tests pin
// it by regex and by refusal: revoked and expired tokens must be untouchable,
// an empty presign must never blank live rows, and a benign zero-row race must
// NOT fail the run (a failed run reads as a broken refresher every time a token
// expires between pre_query and dispatch).

import (
	"context"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

func refreshConfig() map[string]interface{} {
	return map[string]interface{}{
		"site_id":        "input_data.site_id",
		"presigned_url":  "zip_result.presigned_url",
		"expiry_minutes": "zip_result.expiry_minutes",
	}
}

func refreshCollected(siteID uuid.UUID, url string) map[string]interface{} {
	return map[string]interface{}{
		"input_data": map[string]interface{}{"site_id": siteID.String()},
		"zip_result": map[string]interface{}{"presigned_url": url, "expiry_minutes": "10080"},
	}
}

// The UPDATE must carry all three protective predicates. The regex is the
// assertion: dropping revoked_at or expires_at from the WHERE fails here, and
// each of those is a real resurrection (a killed link coming back; a closed
// customer window quietly extended).
func TestRefreshZipLinkGuardsTheWhereClause(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectExec(`(?s)UPDATE customer_access_tokens.*purpose\s+=\s+'zip_download'.*revoked_at IS NULL.*expires_at >`).
		WillReturnResult(sqlmock.NewResult(0, 2))

	out, err := RefreshZipLinkAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: refreshConfig()},
		CollectedData:    refreshCollected(siteID, "https://bucket.example/zip?sig=new"),
	})
	if err != nil {
		t.Fatalf("RefreshZipLinkAction: %v", err)
	}
	if m, ok := out.(map[string]interface{}); !ok || m["refreshed"] != int64(2) {
		t.Errorf("output = %#v, want refreshed=2", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the UPDATE did not carry the protective WHERE: %v", err)
	}
}

// An empty presign must refuse BEFORE touching the database: writing '' would
// turn every live token stale at once — the link-born-broken shape.
func TestRefreshZipLinkRefusesAnEmptyPresign(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// No expectations: any DB touch is the failure.

	_, err = RefreshZipLinkAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: refreshConfig()},
		CollectedData:    refreshCollected(uuid.New(), ""),
	})
	// TWO refusals stand in series and either is correct: ExtractActionInputs
	// treats a path resolving to "" as a MISSING required field (the earlier
	// door), and the in-action guard catches a literal empty that slips past it
	// (belt and braces — the braces would matter if the spec ever demoted
	// presigned_url to optional). What must hold either way: refusal, no DB.
	if err == nil {
		t.Fatal("an empty presign was accepted")
	}
	if !strings.Contains(err.Error(), "refusing to blank") && !strings.Contains(err.Error(), "missing required fields") {
		t.Fatalf("expected an empty-presign refusal (either door), got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the database was touched with an empty presign: %v", err)
	}
}

// Zero rows is a benign race (token expired between pre_query and dispatch),
// and it must NOT error: a failed run here reads as a broken refresher.
func TestRefreshZipLinkTreatsZeroRowsAsBenign(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE customer_access_tokens`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := RefreshZipLinkAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: refreshConfig()},
		CollectedData:    refreshCollected(uuid.New(), "https://bucket.example/zip"),
	})
	if err != nil {
		t.Fatalf("zero rows errored: %v — a benign race would fail the run", err)
	}
	if m, ok := out.(map[string]interface{}); !ok || m["refreshed"] != int64(0) {
		t.Errorf("output = %#v, want refreshed=0", out)
	}
}
