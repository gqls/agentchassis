// FILE: platform/orchestration/actions/derive_card_asset_action.go
//
// DeriveCardAssetAction produces a content entity's CARD image FROM its
// existing hero asset — never as an independent generation (Phase I3, G3:
// "card image = the article's asset re-cropped", user-confirmed 2026-07-08;
// D11/D12 2026-07-16: JPG, purpose-only, no new plan kind). It finds the
// entity's hero (the page-scope plan hero, falling back to the site-scope
// brand hero), downloads the bytes from storage, cover-crops to the card
// purpose's exact dimensions, commits the file to the site's git repo, and
// upserts an entity-linked assets row (entity_type + entity_id, Lane B) with
// origin_asset_id recording the derivation lineage. Deterministic image
// processing — no LLM, no diffusion.
//
// v1 supports entity_type='page' (articles/guides — the blog-index card
// surface). News items (Phase I5) and products (Phase I6) extend the same
// entity link without schema change.
//
// Runs inside a storage-enabled agent (asset-deployer, mode 'content_card'):
// it needs the S3 client (to read the hero) and the Kafka producer (to reach
// the git adapter). Idempotent in effect: re-running overwrites the card
// with a freshly-derived copy from the current hero.
//
// Registration (registry.go):
//   "derive_card_asset": {
//       Handler:     DeriveCardAssetAction,
//       Category:    "image",
//       Description: "Derive a content entity's card image from its hero and commit",
//       IsLocal:     true,
//   }

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

var DeriveCardAssetInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id", "entity_id"},
	Optional:    []string{"domain", "entity_type", "page_name"},
	Defaults:    map[string]interface{}{"entity_type": "page"},
}

func init() {
	datahelpers.RegisterActionInputSpec("derive_card_asset", DeriveCardAssetInputSpec)
}

func DeriveCardAssetAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "derive_card_asset"))

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(params.CollectedData, params.StepConfig.Config, DeriveCardAssetInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}
	entityID, err := uuid.Parse(inputs.Get("entity_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid entity_id %q: %w", inputs.Get("entity_id"), err)
	}
	entityType := inputs.Get("entity_type")
	if entityType == "" {
		entityType = "page"
	}
	if entityType != "page" {
		return nil, fmt.Errorf("derive_card_asset: unsupported entity_type %q (v1 handles 'page'; news is Phase I5, products Phase I6)", entityType)
	}
	domain := inputs.Get("domain")
	pageName := inputs.Get("page_name")

	// ── Resolve the page (and verify it belongs to the site) ──
	var pageNameDB, domainDB string
	err = params.DB.QueryRowContext(ctx, `
		SELECT p.name, s.domain
		  FROM pages p JOIN sites s ON s.id = p.site_id
		 WHERE p.id = $1 AND p.site_id = $2
	`, entityID, siteID).Scan(&pageNameDB, &domainDB)
	if err != nil {
		return nil, fmt.Errorf("page %s not found on site %s: %w", entityID, siteID, err)
	}
	if pageName == "" {
		pageName = pageNameDB
	}
	if domain == "" {
		domain = domainDB
	}

	// ── The card's asset_key IS its artefact's identity: the repo path below and
	// the provenance upsert both derive from it, so it is what the lock is read
	// against. Computed here (not at the commit) because the lock decides
	// whether any of the work that follows should happen at all. ──
	cardKey := "card_" + strings.ReplaceAll(pageName, "-", "_")

	// ── A locked row is an owner approval — honour it BEFORE the git commit
	// (bugs_open/143). The upsert's own `WHERE assets.locked_at IS NULL` below
	// protects only the provenance row; by the time it runs the card file in the
	// site repo has already been replaced, so the approved row would survive and
	// its artefact would not. Checked here, before storage is even required, so
	// a locked card costs nothing and refuses visibly. Shared guard — see
	// asset_lock_guard.go for why there is no status filter and no expiry test.
	// An error is fail-closed: we do not overwrite an approval on a DB blip. ──
	locks, err := lockedAssetKeys(ctx, params.DB, siteID, cardKey)
	if err != nil {
		return nil, fmt.Errorf("check card asset lock (%s): %w", cardKey, err)
	}
	if locks.Locked(cardKey) {
		logger.Info("derive_card_asset: card asset is locked — refusing to overwrite",
			zap.String("domain", domain),
			zap.String("page", pageName),
			zap.String("lock", locks.Describe(cardKey)))
		return map[string]interface{}{
			"derived":   false,
			"locked":    true,
			"asset_key": cardKey,
			"reason": fmt.Sprintf("card asset is locked (%s) — approved assets are never overwritten; clear the lock deliberately first, then re-derive",
				locks.Describe(cardKey)),
		}, nil
	}

	// ── Find the source hero: page-scope plan hero, else the site-scope brand
	// hero. The plan row's key IS the asset_key convention (Lane A). ──
	sourceAssetID, sourceURL, sourceKey, err := findCardSourceHero(ctx, params.DB, siteID, pageName)
	if err != nil {
		return map[string]interface{}{
			"derived": false,
			"reason":  fmt.Sprintf("no hero asset to derive from: %v", err),
		}, nil
	}

	// ── Fetch the hero bytes from storage ──
	s3Client, ok := params.StorageClient.(*storage.S3Client)
	if !ok || s3Client == nil {
		return nil, fmt.Errorf("storage client not available (this action must run in a storage-enabled agent, e.g. asset-deployer)")
	}
	key := storage.ExtractKeyFromS3URI(presignedURLToS3URI(sourceURL))
	if key == "" {
		return nil, fmt.Errorf("could not derive storage key from hero url (asset %s; a url-flipped web path needs the presigned source restored or storage_path populated)", sourceAssetID)
	}
	rc, err := s3Client.Download(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("download hero bytes: %w", err)
	}
	heroBytes := new(bytes.Buffer)
	if _, err := heroBytes.ReadFrom(rc); err != nil {
		rc.Close()
		return nil, fmt.Errorf("read hero bytes: %w", err)
	}
	rc.Close()

	heroImg, _, err := image.Decode(bytes.NewReader(heroBytes.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("decode hero image: %w", err)
	}

	// ── Derive the card: exact cover-crop at the card purpose's geometry ──
	w, h, quality, _ := storage.GetImageConfig("card")
	cardImg := storage.CoverCropResize(heroImg, w, h)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, cardImg, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("encode card jpg: %w", err)
	}
	cardBytes := buf.Bytes()

	// ── Commit to the site repo (cardKey resolved and lock-checked above) ──
	repoPath := storage.DefaultAssetBasePath + "/" + storage.AssetKeyFilename(cardKey, ".jpg")
	files := map[string]interface{}{
		repoPath: map[string]interface{}{
			"content": base64.StdEncoding.EncodeToString(cardBytes), "encoding": "base64",
		},
	}
	if _, err := sendGitCommitRequest(ctx, params, domain, files, "content-card", logger); err != nil {
		return nil, fmt.Errorf("git commit card: %w", err)
	}
	webPath := storage.DeployedWebPath(cardKey, "card")

	// ── Entity-linked provenance row (Lane B). Unlike the brand-head rows this
	// is NOT best-effort: the entity link is what the query resolvers read, so
	// a card that committed but never linked would stay invisible forever. ──
	dims, _ := json.Marshal(map[string]int{"width": int(w), "height": int(h)})
	res, err := params.DB.ExecContext(ctx, `
		INSERT INTO assets (id, site_id, name, asset_type, purpose, asset_key, url,
		                    origin_type, origin_model, origin_asset_id,
		                    entity_type, entity_id, mime_type, file_size, dimensions, created_at)
		VALUES (gen_random_uuid(), $1, $2, 'image', 'card', $3, $4,
		        'generated', 'derived-from-hero', $5,
		        $6, $7, 'image/jpeg', $8, $9::jsonb, NOW())
		ON CONFLICT (site_id, asset_key) WHERE asset_key IS NOT NULL AND status = 'active' DO UPDATE SET
			url = EXCLUDED.url, origin_asset_id = EXCLUDED.origin_asset_id,
			entity_type = EXCLUDED.entity_type, entity_id = EXCLUDED.entity_id,
			file_size = EXCLUDED.file_size, dimensions = EXCLUDED.dimensions,
			updated_at = NOW()
		WHERE assets.locked_at IS NULL
	`, siteID, cardKey+" (derived card)", cardKey, webPath,
		sourceAssetID, entityType, entityID, len(cardBytes), string(dims))
	if err != nil {
		return nil, fmt.Errorf("card asset upsert: %w", err)
	}

	// The upsert's WHERE assets.locked_at IS NULL can suppress the DO UPDATE
	// silently — no error, no row. Before bugs_open/143 this result was
	// discarded, so a lock-suppressed provenance write was reported as a clean
	// success. It can now only happen in the TOCTOU window (a lock taken after
	// the pre-check above and before this statement), and that must be loud: the
	// artefact HAS been replaced, so the state is genuinely inconsistent and a
	// human has to reconcile it. Reported at the call boundary, not just logged.
	provenanceRecorded := true
	if res != nil {
		if affected, aerr := res.RowsAffected(); aerr == nil && affected == 0 {
			provenanceRecorded = false
			logger.Error("derive_card_asset: provenance write SUPPRESSED by the asset lock after the pre-check passed — the card artefact was already committed and now disagrees with its row",
				zap.String("domain", domain),
				zap.String("page", pageName),
				zap.String("card_key", cardKey))
		}
	}

	logger.Info("derive_card_asset: committed entity card",
		zap.String("domain", domain),
		zap.String("page", pageName),
		zap.String("card_key", cardKey),
		zap.String("source_key", sourceKey),
		zap.Bool("provenance_recorded", provenanceRecorded),
		zap.Int("bytes", len(cardBytes)))

	result := map[string]interface{}{
		"derived":         true,
		"card_url":        webPath,
		"asset_key":       cardKey,
		"entity_type":     entityType,
		"entity_id":       entityID.String(),
		"source_asset_id": sourceAssetID,
		"file_size":       len(cardBytes),
	}
	if !provenanceRecorded {
		result["provenance_recorded"] = false
		result["locked"] = true
		result["reason"] = "card artefact was committed, then the provenance write was blocked by a lock taken mid-derivation — the committed file and the assets row now disagree and need reconciling"
	}
	return result, nil
}

// findCardSourceHero locates the active hero asset a card derives from, in
// the canonical preference order (same as plan_sections.ensureAssets and the
// content_image_missing check, so all three converge on one source): the
// page's own plan hero, then the Lane B content hero (D13, literal
// ContentHeroKey convention), then the site-scope brand hero.
func findCardSourceHero(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string) (assetID, url, assetKey string, err error) {
	const q = `
		SELECT a.id::text, a.url, a.asset_key
		  FROM site_plan_imagery spi
		  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		  JOIN assets a ON a.site_id = sp.site_id AND a.asset_key = spi.key AND a.status = 'active'
		 WHERE sp.site_id = $1 AND spi.kind = 'hero' AND spi.scope = $2
		   AND ($2 = 'site' OR spi.scope_ref = $3)
		 ORDER BY spi.ordering
		 LIMIT 1`
	err = db.QueryRowContext(ctx, q, siteID, "page", pageName).Scan(&assetID, &url, &assetKey)
	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, `
			SELECT a.id::text, a.url, a.asset_key
			  FROM assets a
			 WHERE a.site_id = $1 AND a.asset_key = $2 AND a.status = 'active'
			 LIMIT 1
		`, siteID, imageryplan.ContentHeroKey(pageName)).Scan(&assetID, &url, &assetKey)
	}
	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, q, siteID, "site", "").Scan(&assetID, &url, &assetKey)
	}
	if err == sql.ErrNoRows {
		return "", "", "", fmt.Errorf("no active page, content, or site hero for %q", pageName)
	}
	if err != nil {
		return "", "", "", fmt.Errorf("hero lookup failed: %w", err)
	}
	return assetID, url, assetKey, nil
}
