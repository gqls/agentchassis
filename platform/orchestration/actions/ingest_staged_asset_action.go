// FILE: platform/orchestration/actions/ingest_staged_asset_action.go
//
// IngestStagedAssetAction is the operator's path for AMENDING an image asset —
// the one thing the platform never had (bugs_open/131 found three sites whose
// stored "logo" was not a logo, and no way for a human to supply a corrected
// one). Every image before this arrived from a generator; the admin API edits
// metadata only. This action closes the gap without ever handing storage
// credentials to the operator: bytes travel operator file → base64 → psql →
// asset_ingest_staging (BYTEA) → this action → S3.
//
// The work item carries ONLY the staging row id — bytes must not ride Kafka
// (the kafka-go writer caps messages at 1 MiB, and the recorded doctrine is
// "heavy artifacts live in the DB, retrievable by id"; precedent:
// chassis_intake_events.payload BYTEA).
//
// Behaviour, refusals-as-results per house style (a refused ingest completes
// the workflow and reports why; it does not error the orchestration):
//   1. Atomically claim the staging row (status pending → processing) — a
//      double-dispatch finds no pending row and refuses instead of racing.
//   2. Re-verify sha256 over the bytes; mismatch ⇒ staging 'failed'.
//   3. image.Decode to prove it IS an image; capture dimensions + format.
//   4. Upload to S3 at a NEW key (images/uploads/<site>/<date>/<uuid>.<ext>)
//      — never overwrite; the previous object survives, matching the
//      regenerate path, which never deletes either.
//   5. Update the assets row IN PLACE (id stable, references hold), honouring
//      locked_at exactly as StoreAssetAction does (approved assets are never
//      overwritten), recording the amendment in assets.alterations with the
//      previous url/storage_path. storage_path is always populated so the row
//      survives a later url-flip (the defect that broke idea.uk's derivation).
//   6. Mark the staging row 'ingested'. Return a summary, never the bytes.
//
// Runs inside a storage-enabled agent (asset-deployer, mode 'ingest_upload').
//
// Registration (registry.go):
//   "ingest_staged_asset": {
//       Handler:     IngestStagedAssetAction,
//       Category:    "image",
//       Description: "Ingest operator-supplied asset bytes from staging into S3 and the assets row",
//       IsLocal:     true,
//   }

package actions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

var IngestStagedAssetInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"staging_id", "site_id"},
	Optional:    []string{},
	Defaults:    map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("ingest_staged_asset", IngestStagedAssetInputSpec)
}

// maxIngestBytes is a sanity cap, not a design limit — the largest legitimate
// brand asset in the estate is well under 1 MB; 10 MB catches a mistaken file
// without refusing any plausible image.
const maxIngestBytes = 10 * 1024 * 1024

func IngestStagedAssetAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "ingest_staged_asset"))

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(params.CollectedData, params.StepConfig.Config, IngestStagedAssetInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	stagingID, err := uuid.Parse(inputs.Get("staging_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid staging_id %q: %w", inputs.Get("staging_id"), err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}

	// ── 1. Claim the staging row atomically (pending → processing) ──
	// One statement claims AND loads: a concurrent second dispatch for the
	// same row finds status != 'pending' and refuses rather than double-
	// uploading.
	var (
		rowSiteID uuid.UUID
		assetKey  string
		purpose   sql.NullString
		content   []byte
		wantSHA   string
		note      sql.NullString
		createdBy string
	)
	err = params.DB.QueryRowContext(ctx, `
		UPDATE asset_ingest_staging
		   SET status = 'processing'
		 WHERE id = $1 AND status = 'pending'
		RETURNING site_id, asset_key, purpose, content, sha256, note, created_by
	`, stagingID).Scan(&rowSiteID, &assetKey, &purpose, &content, &wantSHA, &note, &createdBy)
	if err == sql.ErrNoRows {
		return map[string]interface{}{
			"ingested": false,
			"reason":   "staging row not found or not pending (already consumed, failed, or mid-processing)",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim staging row: %w", err)
	}

	refuse := func(status, reason string) (interface{}, error) {
		if _, uerr := params.DB.ExecContext(ctx, `
			UPDATE asset_ingest_staging SET status = $2, error = $3 WHERE id = $1
		`, stagingID, status, reason); uerr != nil {
			logger.Warn("ingest_staged_asset: could not record refusal on staging row", zap.Error(uerr))
		}
		logger.Warn("ingest_staged_asset: refused", zap.String("reason", reason),
			zap.String("staging_id", stagingID.String()), zap.String("asset_key", assetKey))
		return map[string]interface{}{
			"ingested":  false,
			"asset_key": assetKey,
			"reason":    reason,
		}, nil
	}

	if rowSiteID != siteID {
		return refuse("refused", "staging row site_id does not match the work item's site — refusing a cross-site amend")
	}
	if len(content) == 0 || len(content) > maxIngestBytes {
		return refuse("refused", fmt.Sprintf("content size %d outside (0, %d] sanity bounds", len(content), maxIngestBytes))
	}

	// ── 2. Integrity: the sha the loader computed must match what arrived ──
	gotSHA := sha256.Sum256(content)
	if !strings.EqualFold(hex.EncodeToString(gotSHA[:]), strings.TrimSpace(wantSHA)) {
		return refuse("failed", "sha256 mismatch between staged bytes and loader-computed digest — bytes corrupted in transit, restage")
	}

	// ── 3. Prove it is an image; capture what it is ──
	imgCfg, format, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		return refuse("refused", fmt.Sprintf("content does not decode as an image: %v", err))
	}
	ext, contentType := formatToExtAndMIME(format)
	if ext == "" {
		return refuse("refused", fmt.Sprintf("unsupported image format %q (png, jpeg, gif accepted)", format))
	}

	// ── 3b. Lock pre-check, BEFORE the upload ──
	// The in-tx FOR UPDATE re-check below is the enforcement (TOCTOU-safe);
	// this early read exists so a locked asset refuses without first writing
	// an orphan object to S3. Mirrors StoreAssetAction's D5 wording.
	var preLockedAt sql.NullTime
	err = params.DB.QueryRowContext(ctx, `
		SELECT locked_at FROM assets
		 WHERE site_id = $1 AND asset_key = $2 AND status = 'active'
	`, siteID, assetKey).Scan(&preLockedAt)
	if err != nil && err != sql.ErrNoRows {
		return refuse("failed", fmt.Sprintf("lock pre-check: %v", err))
	}
	if preLockedAt.Valid {
		return refuse("refused", "asset is locked (locked_at set) — approved assets are never overwritten; clear the lock deliberately first, then restage")
	}

	// ── 4. Upload at a NEW key — the old object is never touched ──
	s3Client, ok := params.StorageClient.(*storage.S3Client)
	if !ok || s3Client == nil {
		// Not a refusal: this is a deployment error (wrong agent), and the
		// staging row should stay claimable after the wiring is fixed.
		if _, uerr := params.DB.ExecContext(ctx, `
			UPDATE asset_ingest_staging SET status = 'pending' WHERE id = $1
		`, stagingID); uerr != nil {
			logger.Warn("ingest_staged_asset: could not release staging row", zap.Error(uerr))
		}
		return nil, fmt.Errorf("storage client not available (this action must run in a storage-enabled agent, e.g. asset-deployer)")
	}
	key := fmt.Sprintf("images/uploads/%s/%s/%s%s",
		siteID.String(), time.Now().UTC().Format("20060102"), uuid.NewString(), ext)
	s3URI, err := s3Client.Upload(ctx, key, contentType, bytes.NewReader(content))
	if err != nil {
		return refuse("failed", fmt.Sprintf("s3 upload failed: %v", err))
	}
	presignedURL, err := s3Client.GetPresignedURL(ctx, key, 7*24*60)
	if err != nil {
		// The object is up but we cannot mint any HTTPS form of its address —
		// and a bare s3:// URI in assets.url BREAKS derive_brand_head_assets
		// (presignedURLToS3URI reads the first key segment as the bucket and
		// derives a wrong key — measured on the 131 og-card lane, 2026-07-29).
		// Refuse rather than store a poisoned row; the orphan object is
		// harmless and a re-run re-uploads.
		return refuse("failed", fmt.Sprintf("presign failed after upload (object at %s is orphaned, harmless): %v", s3URI, err))
	}
	// assets.url gets the DURABLE path-style HTTPS form — the presigned URL
	// with its query stripped (https://<endpoint>/<bucket>/<key>). The signed
	// form expires in 7 days and would leave the row looking alive then dead;
	// every platform consumer parses only the path (or uses storage_path), and
	// the operator gets the signed URL in the result for immediate verification.
	rowURL := presignedURL
	if i := strings.IndexByte(rowURL, '?'); i >= 0 {
		rowURL = rowURL[:i]
	}

	// ── 5. Amend the assets row in place, honouring locks ──
	alteration := map[string]interface{}{
		"type": "bytes_replaced",
		"at":   time.Now().UTC().Format(time.RFC3339),
		"by":   createdBy,
		"note": note.String,
		"new":  map[string]interface{}{"storage_path": key, "sha256": strings.ToLower(hex.EncodeToString(gotSHA[:]))},
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return refuse("failed", fmt.Sprintf("begin tx: %v", err))
	}
	defer tx.Rollback()

	var (
		existingID          uuid.UUID
		existingURL         sql.NullString
		existingStoragePath sql.NullString
		existingPurpose     sql.NullString
		lockedAt            sql.NullTime
	)
	err = tx.QueryRowContext(ctx, `
		SELECT id, url, storage_path, purpose, locked_at
		  FROM assets
		 WHERE site_id = $1 AND asset_key = $2 AND status = 'active'
		 FOR UPDATE
	`, siteID, assetKey).Scan(&existingID, &existingURL, &existingStoragePath, &existingPurpose, &lockedAt)

	switch {
	case err == sql.ErrNoRows:
		// No active row for this key — a fresh INSERT, first alteration entry.
		altJSON, _ := json.Marshal([]interface{}{alteration})
		newID := uuid.New()
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO assets (id, site_id, name, asset_type, purpose, asset_key, url,
			                    storage_provider, storage_path, mime_type, file_size, dimensions,
			                    origin_type, origin_model, alterations, created_at)
			VALUES ($1, $2, $3, 'image', $4, $5, $6, 'backblaze', $7, $8, $9, $10,
			        'uploaded', 'operator-supplied', $11::jsonb, NOW())
		`, newID, siteID, assetKey+" (operator-supplied)", nullString(purposeOrDefault(purpose, "")),
			assetKey, rowURL, key, contentType, len(content),
			dimensionsJSON(imgCfg.Width, imgCfg.Height), string(altJSON)); err != nil {
			return refuse("failed", fmt.Sprintf("insert assets row: %v", err))
		}
		existingID = newID
	case err != nil:
		return refuse("failed", fmt.Sprintf("read assets row: %v", err))
	case lockedAt.Valid:
		// Mirror StoreAssetAction's D5 refusal wording — approved assets are
		// never overwritten by machinery. Clearing the lock is a deliberate,
		// documented human step, not something this action does.
		return refuse("refused", "asset is locked (locked_at set) — approved assets are never overwritten; clear the lock deliberately first, then restage")
	default:
		alteration["previous"] = map[string]interface{}{
			"url":          existingURL.String,
			"storage_path": existingStoragePath.String,
		}
		altJSON, _ := json.Marshal(alteration)
		if _, err := tx.ExecContext(ctx, `
			UPDATE assets
			   SET url = $2,
			       storage_provider = 'backblaze',
			       storage_path = $3,
			       mime_type = $4,
			       file_size = $5,
			       dimensions = $6::jsonb,
			       purpose = COALESCE($7, purpose),
			       origin_type = 'uploaded',
			       origin_model = 'operator-supplied',
			       alterations = COALESCE(alterations, '[]'::jsonb) || $8::jsonb,
			       updated_at = NOW()
			 WHERE id = $1 AND locked_at IS NULL
		`, existingID, rowURL, key, contentType, len(content),
			dimensionsJSON(imgCfg.Width, imgCfg.Height),
			nullString(purpose.String), string(altJSON)); err != nil {
			return refuse("failed", fmt.Sprintf("update assets row: %v", err))
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE asset_ingest_staging
		   SET status = 'ingested', consumed_at = NOW(), error = NULL
		 WHERE id = $1
	`, stagingID); err != nil {
		// Roll back BEFORE refuse: the tx now holds this staging row's lock,
		// and refuse writes the same row on a fresh connection — without the
		// explicit rollback the two would block each other until timeout.
		tx.Rollback()
		return refuse("failed", fmt.Sprintf("mark staging ingested: %v", err))
	}
	if err := tx.Commit(); err != nil {
		return refuse("failed", fmt.Sprintf("commit: %v", err))
	}

	logger.Info("ingest_staged_asset: bytes amended",
		zap.String("asset_key", assetKey),
		zap.String("site_id", siteID.String()),
		zap.String("storage_path", key),
		zap.Int("bytes", len(content)),
		zap.String("format", format))

	return map[string]interface{}{
		"ingested":      true,
		"asset_id":      existingID.String(),
		"asset_key":     assetKey,
		"s3_uri":        s3URI,
		"storage_path":  key,
		"presigned_url": presignedURL,
		"width":         imgCfg.Width,
		"height":        imgCfg.Height,
		"bytes":         len(content),
		"format":        format,
	}, nil
}

// formatToExtAndMIME maps image.DecodeConfig's format name to the file
// extension and content type used for the S3 object.
func formatToExtAndMIME(format string) (ext, mime string) {
	switch format {
	case "png":
		return ".png", "image/png"
	case "jpeg":
		return ".jpg", "image/jpeg"
	case "gif":
		return ".gif", "image/gif"
	default:
		return "", ""
	}
}

// dimensionsJSON renders the assets.dimensions jsonb payload.
func dimensionsJSON(w, h int) string {
	return fmt.Sprintf(`{"width": %d, "height": %d}`, w, h)
}

// purposeOrDefault returns the staged purpose when set, else the fallback.
func purposeOrDefault(p sql.NullString, fallback string) string {
	if p.Valid && strings.TrimSpace(p.String) != "" {
		return p.String
	}
	return fallback
}
