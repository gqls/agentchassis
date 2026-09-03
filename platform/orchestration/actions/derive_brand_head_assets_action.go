// FILE: platform/orchestration/actions/derive_brand_head_assets_action.go
//
// DeriveBrandHeadAssetsAction produces a site's favicon and Open Graph card
// FROM its approved logo — never as independent generations (Phase I1, G1/G8:
// "favicon derived from the logo; OG/social card images"). It fetches the
// logo's bytes from storage, resizes to a square favicon, composes the logo
// centred on a brand-colour background for the 1200×630 OG card, and commits
// both to the site's git repo. Deterministic image processing — no LLM, no
// diffusion.
//
// Runs inside a storage-enabled agent (asset-deployer): it needs both the
// S3 client (to read the logo) and the Kafka producer (to reach the git
// adapter). The head-tag wiring that references these files lives in
// render_site_components (favicon <link> + og:image <meta>), with graceful
// fallback to the logo / brand hero when these files don't exist yet.
//
// Idempotent in effect: re-running overwrites favicon.png / og-card.png with
// freshly-derived copies from the current logo — EXCEPT artefacts whose
// assets row is locked (locked_at set): those are owner approvals and are
// skipped before the git commit, not just at the provenance upsert.
//
// Registration (registry.go):
//   "derive_brand_head_assets": {
//       Handler:     DeriveBrandHeadAssetsAction,
//       Category:    "site",
//       Description: "Derive favicon + OG card from the site logo and commit",
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
	"image/color"
	"image/draw"
	"image/png"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"github.com/nfnt/resize"
	"go.uber.org/zap"
)

var DeriveBrandHeadAssetsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"domain"},
	Defaults:    map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("derive_brand_head_assets", DeriveBrandHeadAssetsInputSpec)
}

const (
	faviconSize = 64   // square favicon edge, px
	ogWidth     = 1200 // Open Graph card, px
	ogHeight    = 630
	ogLogoMax   = 420 // longest logo edge inside the OG card
)

func DeriveBrandHeadAssetsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "derive_brand_head_assets"))

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(params.CollectedData, params.StepConfig.Config, DeriveBrandHeadAssetsInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}
	domain := inputs.Get("domain")

	// ── Load the logo asset (prefer the locked/approved one) + brand colour ──
	var logoURL, logoStoragePath, domainDB string
	err = params.DB.QueryRowContext(ctx, `
		SELECT a.url, COALESCE(a.storage_path, ''), si.domain
		  FROM assets a
		  JOIN sites si ON si.id = a.site_id
		 WHERE a.site_id = $1 AND a.asset_key = 'logo' AND a.status = 'active'
		 ORDER BY (a.locked_at IS NOT NULL) DESC, a.updated_at DESC
		 LIMIT 1
	`, siteID).Scan(&logoURL, &logoStoragePath, &domainDB)
	if err != nil {
		return map[string]interface{}{"derived": false, "reason": "no active logo asset"}, nil
	}
	if domain == "" {
		domain = domainDB
	}

	bgColour := loadBrandBackgroundColour(ctx, params.DB, siteID)

	// ── Locked rows are owner approvals — honour them BEFORE the git commit ──
	// recordDerivedAsset's own locked_at guard protects only the provenance
	// row; by the time it runs the file in the site repo has already been
	// replaced. An approved artefact (locked_at set, Phase I1 D5) must never
	// be overwritten by a re-derivation, so the lock is checked here — before
	// storage is even required — and the locked artefact is left out of the
	// commit entirely.
	lockedKeys, err := lockedBrandHeadKeys(ctx, params.DB, siteID)
	if err != nil {
		return nil, fmt.Errorf("check locked brand-head assets: %w", err)
	}
	if lockedKeys.Locked("favicon") && lockedKeys.Locked("og_card") {
		logger.Info("derive_brand_head_assets: both artefacts locked — nothing to derive",
			zap.String("domain", domain))
		return map[string]interface{}{
			"derived": false,
			"reason":  "favicon and og_card are locked (owner-approved) — refusing to overwrite",
		}, nil
	}

	// ── Fetch the logo bytes from storage ──
	s3Client, ok := params.StorageClient.(*storage.S3Client)
	if !ok || s3Client == nil {
		return nil, fmt.Errorf("storage client not available (this action must run in a storage-enabled agent, e.g. asset-deployer)")
	}
	// Resolve the logo's SOURCE object from the row itself (storage_path
	// first, url as the legacy fallback). Reading url alone broke the moment
	// deploy_image_asset started flipping it to the local web path
	// (bugs_open/152) — measured 2026-08-06: five sites' active logo rows
	// carried a non-presigned url, four of them recoverable only via
	// storage_path.
	key := storage.ExtractKeyFromS3URI(storage.AssetSourceRef(logoStoragePath, logoURL))
	if key == "" {
		return nil, fmt.Errorf("could not derive a storage key for the logo source (url %q, storage_path %q): neither identifies a stored object", logoURL, logoStoragePath)
	}
	rc, err := s3Client.Download(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("download logo bytes: %w", err)
	}
	logoBytes := new(bytes.Buffer)
	if _, err := logoBytes.ReadFrom(rc); err != nil {
		rc.Close()
		return nil, fmt.Errorf("read logo bytes: %w", err)
	}
	rc.Close()

	logoImg, _, err := image.Decode(bytes.NewReader(logoBytes.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("decode logo image: %w", err)
	}

	// ── Derive favicon (square) and OG card (logo on brand background) ──
	files := map[string]interface{}{}
	// Captured from the bytes actually committed, so the provenance rows below
	// record what was stored rather than what the filename claims (bugs_open/433).
	var faviconMIME, ogCardMIME string
	if !lockedKeys.Locked("favicon") {
		faviconPNG, err := encodePNG(composeFavicon(logoImg))
		if err != nil {
			return nil, fmt.Errorf("encode favicon: %w", err)
		}
		_, faviconMIME = storage.SniffImageExtAndMIME(faviconPNG)
		files[storage.DefaultAssetBasePath+"/favicon.png"] = map[string]interface{}{
			"content": base64.StdEncoding.EncodeToString(faviconPNG), "encoding": "base64",
		}
	}
	if !lockedKeys.Locked("og_card") {
		ogPNG, err := composeOGCard(logoImg, bgColour)
		if err != nil {
			return nil, fmt.Errorf("compose og card: %w", err)
		}
		_, ogCardMIME = storage.SniffImageExtAndMIME(ogPNG)
		files[storage.DefaultAssetBasePath+"/og-card.png"] = map[string]interface{}{
			"content": base64.StdEncoding.EncodeToString(ogPNG), "encoding": "base64",
		}
	}
	if _, err := sendGitCommitRequest(ctx, params, domain, files, "brand-head", logger); err != nil {
		return nil, fmt.Errorf("git commit favicon/og: %w", err)
	}

	// ── Provenance rows (best-effort; derivation, origin = the logo) ──
	if !lockedKeys.Locked("favicon") {
		recordDerivedAsset(ctx, params.DB, siteID, "favicon", "/assets/images/favicon.png", faviconMIME, logger)
	}
	if !lockedKeys.Locked("og_card") {
		recordDerivedAsset(ctx, params.DB, siteID, "og_card", "/assets/images/og-card.png", ogCardMIME, logger)
	}

	logger.Info("derive_brand_head_assets: committed brand-head assets",
		zap.String("domain", domain),
		zap.String("background", bgColour),
		zap.Int("files", len(files)))

	// The partial-lock case must be visible at the call boundary, not only in
	// logs: URLs are reported only for artefacts actually derived, and
	// skipped_locked names the ones the lock excluded.
	result := map[string]interface{}{"derived": true}
	skipped := []string{}
	if lockedKeys.Locked("favicon") {
		skipped = append(skipped, "favicon")
	} else {
		result["favicon_url"] = "/assets/images/favicon.png"
	}
	if lockedKeys.Locked("og_card") {
		skipped = append(skipped, "og_card")
	} else {
		result["og_image_url"] = "/assets/images/og-card.png"
	}
	if len(skipped) > 0 {
		result["skipped_locked"] = skipped
	}
	return result, nil
}

// composeFavicon scales the logo to FIT the square favicon box, preserving
// aspect, centred on a transparent canvas. resize.Resize with both edges
// pinned stretches a non-square logo to the box — a wide wordmark comes out
// as an illegible horizontal smear — so the padding must come from the
// canvas, not the resample.
func composeFavicon(logo image.Image) image.Image {
	canvas := image.NewRGBA(image.Rect(0, 0, faviconSize, faviconSize))
	scaled := resize.Thumbnail(faviconSize, faviconSize, logo, resize.Lanczos3)
	b := scaled.Bounds()
	offX := (faviconSize - b.Dx()) / 2
	offY := (faviconSize - b.Dy()) / 2
	draw.Draw(canvas, image.Rect(offX, offY, offX+b.Dx(), offY+b.Dy()), scaled, b.Min, draw.Over)
	return canvas
}

// brandHeadAssetKeys are the two artefacts this action derives. Named so the
// lock check and the files/result blocks above cannot drift apart.
var brandHeadAssetKeys = []string{"favicon", "og_card"}

// lockedBrandHeadKeys reports which of the site's brand-head asset keys have
// ANY locked row. The predicate itself lives in asset_lock_guard.go
// (bugs_open/143) — this was one of several hand-written copies of the same
// "is this asset_key locked" query, and a duplicated safety predicate drifts
// silently (016b §9). Semantics are unchanged by the move: no status filter, no
// expiry test, a lock on any row of the key fails CLOSED and the artefact is
// skipped visibly via skipped_locked / the both-locked refusal.
func lockedBrandHeadKeys(ctx context.Context, db *sql.DB, siteID uuid.UUID) (assetLockSet, error) {
	return lockedAssetKeys(ctx, db, siteID, brandHeadAssetKeys...)
}

// composeOGCard draws the logo centred on a solid brand-colour 1200×630 card.
func composeOGCard(logo image.Image, bgHex string) ([]byte, error) {
	card := image.NewRGBA(image.Rect(0, 0, ogWidth, ogHeight))
	draw.Draw(card, card.Bounds(), &image.Uniform{C: parseHexColour(bgHex)}, image.Point{}, draw.Src)

	// Resize the logo so its longest edge is ogLogoMax, preserving aspect.
	scaled := resize.Thumbnail(ogLogoMax, ogLogoMax, logo, resize.Lanczos3)
	b := scaled.Bounds()
	offX := (ogWidth - b.Dx()) / 2
	offY := (ogHeight - b.Dy()) / 2
	draw.Draw(card, image.Rect(offX, offY, offX+b.Dx(), offY+b.Dy()), scaled, b.Min, draw.Over)

	return encodePNG(card)
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// loadBrandBackgroundColour picks the OG card background from the site's
// palette: header_bg (a dark brand surface on most themes) → primary → a
// neutral dark default. Keeps the card on-brand without a config knob.
// Gradient values (e.g. cta_bg) are rejected by parseHexColour downstream.
func loadBrandBackgroundColour(ctx context.Context, db *sql.DB, siteID uuid.UUID) string {
	var paletteJSON sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT sc.color_palette::text
		  FROM style_collections sc
		  JOIN sites si ON si.style_collection_id = sc.id
		 WHERE si.id = $1
	`, siteID).Scan(&paletteJSON)
	if err != nil || !paletteJSON.Valid || paletteJSON.String == "" {
		return "#1a1a2e"
	}
	var palette map[string]string
	if json.Unmarshal([]byte(paletteJSON.String), &palette) != nil {
		return "#1a1a2e"
	}
	for _, slot := range []string{"header_bg", "footer_bg", "background", "primary"} {
		if v := strings.TrimSpace(palette[slot]); strings.HasPrefix(v, "#") {
			return v
		}
	}
	return "#1a1a2e"
}

// parseHexColour parses #RGB or #RRGGBB into a color.Color, defaulting to a
// dark neutral when the value is a gradient/empty/unparseable (OG cards need
// a solid fill; gradients like the palette's cta_bg are not usable here).
func parseHexColour(hex string) color.Color {
	hex = strings.TrimSpace(hex)
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return color.RGBA{R: 0x1a, G: 0x1a, B: 0x2e, A: 0xff}
	}
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return color.RGBA{R: 0x1a, G: 0x1a, B: 0x2e, A: 0xff}
	}
	return color.RGBA{R: r, G: g, B: b, A: 0xff}
}

// recordDerivedAsset upserts a provenance row for a derived brand asset.
// Best-effort: derivation succeeds even if this bookkeeping fails. Uses the
// same (site_id, asset_key) active-upsert as store_asset; origin_type
// 'generated', origin_model records the derivation source.
// mimeType is read from the bytes the caller actually committed, not assumed
// from the .png filename (bugs_open/433). These artefacts genuinely ARE PNG —
// encodePNG produces them a few lines up — so a literal would have been true
// here; sniffing anyway means all four asset writers obey one rule instead of
// three plus an exception, and the rule survives someone changing the encoder.
func recordDerivedAsset(ctx context.Context, db *sql.DB, siteID uuid.UUID, assetKey, webPath, mimeType string, logger *zap.Logger) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO assets (id, site_id, name, asset_type, purpose, asset_key, url, origin_type, origin_model, mime_type, created_at)
		VALUES (gen_random_uuid(), $1, $2, 'image', $3, $3, $4, 'generated', 'derived-from-logo', NULLIF($5, ''), NOW())
		ON CONFLICT (site_id, asset_key) WHERE asset_key IS NOT NULL AND status = 'active' DO UPDATE SET
			url = EXCLUDED.url, mime_type = EXCLUDED.mime_type, updated_at = NOW()
		WHERE `+assetAgentWritableSQL("assets.")+`
	`, siteID, assetKey+" (derived)", assetKey, webPath, mimeType)
	if err != nil {
		logger.Warn("recordDerivedAsset: provenance upsert failed (non-fatal)",
			zap.String("asset_key", assetKey), zap.Error(err))
	}
}
