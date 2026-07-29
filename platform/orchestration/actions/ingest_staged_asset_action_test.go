package actions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// tinyPNG returns a valid 2×2 PNG and its sha256 hex — the smallest honest
// stand-in for operator-supplied bytes.
func tinyPNG(t *testing.T) ([]byte, string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0xff, A: 0xff})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode tiny png: %v", err)
	}
	sum := sha256.Sum256(buf.Bytes())
	return buf.Bytes(), hex.EncodeToString(sum[:])
}

func ingestParams(t *testing.T, db *sql.DB, stagingID, siteID uuid.UUID) ActionParams {
	t.Helper()
	return ActionParams{
		DB:     db,
		Logger: zap.NewNop(),
		StepConfig: models.Step{
			Config: map[string]interface{}{},
		},
		CollectedData: map[string]interface{}{
			"staging_id": stagingID.String(),
			"site_id":    siteID.String(),
		},
	}
}

// claimColumns matches the RETURNING list of the atomic claim statement.
func claimColumns() []string {
	return []string{"site_id", "asset_key", "purpose", "content", "sha256", "note", "created_by"}
}

func resultMap(t *testing.T, got interface{}) map[string]interface{} {
	t.Helper()
	m, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("result is %T, want map", got)
	}
	return m
}

func TestIngestStagedAsset_RefusesConsumedStaging(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	stagingID, siteID := uuid.New(), uuid.New()
	mock.ExpectQuery(`UPDATE asset_ingest_staging`).
		WithArgs(stagingID).
		WillReturnError(sql.ErrNoRows)

	got, err := IngestStagedAssetAction(context.Background(), ingestParams(t, db, stagingID, siteID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := resultMap(t, got)
	if m["ingested"] != false || !strings.Contains(m["reason"].(string), "not pending") {
		t.Errorf("want not-pending refusal, got %v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestIngestStagedAsset_RefusesCrossSite(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	stagingID, siteID, otherSite := uuid.New(), uuid.New(), uuid.New()
	content, sha := tinyPNG(t)
	mock.ExpectQuery(`UPDATE asset_ingest_staging`).
		WithArgs(stagingID).
		WillReturnRows(sqlmock.NewRows(claimColumns()).
			AddRow(otherSite, "logo", nil, content, sha, nil, "operator"))
	mock.ExpectExec(`UPDATE asset_ingest_staging SET status`).
		WithArgs(stagingID, "refused", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := IngestStagedAssetAction(context.Background(), ingestParams(t, db, stagingID, siteID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := resultMap(t, got)
	if m["ingested"] != false || !strings.Contains(m["reason"].(string), "cross-site") {
		t.Errorf("want cross-site refusal, got %v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestIngestStagedAsset_RefusesShaMismatch(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	stagingID, siteID := uuid.New(), uuid.New()
	content, _ := tinyPNG(t)
	mock.ExpectQuery(`UPDATE asset_ingest_staging`).
		WithArgs(stagingID).
		WillReturnRows(sqlmock.NewRows(claimColumns()).
			AddRow(siteID, "logo", nil, content, "deadbeef", nil, "operator"))
	mock.ExpectExec(`UPDATE asset_ingest_staging SET status`).
		WithArgs(stagingID, "failed", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := IngestStagedAssetAction(context.Background(), ingestParams(t, db, stagingID, siteID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := resultMap(t, got)
	if m["ingested"] != false || !strings.Contains(m["reason"].(string), "sha256 mismatch") {
		t.Errorf("want sha-mismatch refusal, got %v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestIngestStagedAsset_RefusesNonImage(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	stagingID, siteID := uuid.New(), uuid.New()
	content := []byte("this is prose, not pixels")
	sum := sha256.Sum256(content)
	mock.ExpectQuery(`UPDATE asset_ingest_staging`).
		WithArgs(stagingID).
		WillReturnRows(sqlmock.NewRows(claimColumns()).
			AddRow(siteID, "logo", nil, content, hex.EncodeToString(sum[:]), nil, "operator"))
	mock.ExpectExec(`UPDATE asset_ingest_staging SET status`).
		WithArgs(stagingID, "refused", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := IngestStagedAssetAction(context.Background(), ingestParams(t, db, stagingID, siteID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := resultMap(t, got)
	if m["ingested"] != false || !strings.Contains(m["reason"].(string), "does not decode") {
		t.Errorf("want non-image refusal, got %v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestIngestStagedAsset_RefusesLockedAsset(t *testing.T) {
	// The lock refusal must fire on the PRE-check, before any S3 upload —
	// this test passes no storage client, so reaching the upload would panic
	// the assertion instead of refusing, which is exactly the regression this
	// guards against (an orphan object written for a locked, refused amend).
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	stagingID, siteID := uuid.New(), uuid.New()
	content, sha := tinyPNG(t)
	mock.ExpectQuery(`UPDATE asset_ingest_staging`).
		WithArgs(stagingID).
		WillReturnRows(sqlmock.NewRows(claimColumns()).
			AddRow(siteID, "logo", nil, content, sha, nil, "operator"))
	mock.ExpectQuery(`SELECT locked_at FROM assets`).
		WithArgs(siteID, "logo").
		WillReturnRows(sqlmock.NewRows([]string{"locked_at"}).AddRow(time.Now()))
	mock.ExpectExec(`UPDATE asset_ingest_staging SET status`).
		WithArgs(stagingID, "refused", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := IngestStagedAssetAction(context.Background(), ingestParams(t, db, stagingID, siteID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := resultMap(t, got)
	if m["ingested"] != false || !strings.Contains(m["reason"].(string), "locked") {
		t.Errorf("want locked refusal, got %v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestFormatToExtAndMIME(t *testing.T) {
	cases := []struct{ format, ext, mime string }{
		{"png", ".png", "image/png"},
		{"jpeg", ".jpg", "image/jpeg"},
		{"gif", ".gif", "image/gif"},
		{"webp", "", ""}, // no decoder registered — refused upstream
	}
	for _, c := range cases {
		ext, mime := formatToExtAndMIME(c.format)
		if ext != c.ext || mime != c.mime {
			t.Errorf("formatToExtAndMIME(%q) = (%q,%q), want (%q,%q)", c.format, ext, mime, c.ext, c.mime)
		}
	}
}
