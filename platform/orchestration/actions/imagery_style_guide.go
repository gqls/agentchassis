// FILE: platform/orchestration/actions/imagery_style_guide.go
//
// Phase I1 (imagery best-in-class): the per-site imagery style guide.
//
// A site_specs aspect `imagery_style_guide` holds the structured,
// brand-approved imagery signal:
//
//	{
//	  "palette":  "deep charcoal, electric blue accents, light grey",
//	  "medium":   "industrial photography, dark atmospheric lighting",
//	  "mood":     "precise, technical, engineered",
//	  "avoid":    "stock-photo people, generic technology abstractions",
//	  "reference_asset_keys": ["hero_canonical"]
//	}
//
// generate_image composes a per-KIND direction from it (photographic kinds
// get medium+mood+palette; icons get palette only; logos get nothing — the
// 2026-05-20 contamination lesson), sends `avoid` to the negative prompt,
// and resolves reference_asset_keys to s3:// URIs as style anchors for
// providers that accept reference images (Banana). When the guide yields a
// direction it supersedes the free-text design_intent.imagery_direction
// fallback — one coherent brand voice, no double prepend.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type imageryStyleGuide struct {
	Palette            string   `json:"palette"`
	Medium             string   `json:"medium"`
	Mood               string   `json:"mood"`
	Avoid              string   `json:"avoid"`
	ReferenceAssetKeys []string `json:"reference_asset_keys"`
}

// getImageryStyleGuideForSite loads the current imagery_style_guide spec for
// a site. Returns nil when the site has none (the common case today) or on
// any error — the guide is enrichment, never a requirement.
func getImageryStyleGuideForSite(ctx context.Context, db interface{}, siteID string, logger *zap.Logger) *imageryStyleGuide {
	if siteID == "" || db == nil {
		return nil
	}
	const query = `
		SELECT data
		FROM site_specs
		WHERE site_id = $1
		  AND aspect = 'imagery_style_guide'
		  AND is_current = true
		LIMIT 1
	`
	var raw []byte
	switch d := db.(type) {
	case *sql.DB:
		if err := d.QueryRowContext(ctx, query, siteID).Scan(&raw); err != nil {
			if err != sql.ErrNoRows {
				logger.Warn("getImageryStyleGuideForSite: query failed",
					zap.String("site_id", siteID), zap.Error(err))
			}
			return nil
		}
	case *pgxpool.Pool:
		if err := d.QueryRow(ctx, query, siteID).Scan(&raw); err != nil {
			if !strings.Contains(err.Error(), "no rows") {
				logger.Warn("getImageryStyleGuideForSite: query failed",
					zap.String("site_id", siteID), zap.Error(err))
			}
			return nil
		}
	default:
		logger.Warn("getImageryStyleGuideForSite: unsupported database type",
			zap.String("site_id", siteID))
		return nil
	}

	var g imageryStyleGuide
	if err := json.Unmarshal(raw, &g); err != nil {
		logger.Warn("getImageryStyleGuideForSite: unparseable spec data",
			zap.String("site_id", siteID), zap.Error(err))
		return nil
	}
	if g.Palette == "" && g.Medium == "" && g.Mood == "" &&
		g.Avoid == "" && len(g.ReferenceAssetKeys) == 0 {
		return nil
	}
	return &g
}

// directionForKind builds the prompt-prefix direction appropriate to an
// image kind. Mirrors directionAppliesToKind's gating philosophy:
//   - photographic kinds (hero, illustration, infographic, legacy default)
//     get the full brand voice: medium, mood, palette;
//   - icons get ONLY the palette — a photographic medium prepended to an
//     icon prompt makes the model composite an icon onto a photo
//     (icon_cycle_time, 2026-05-20);
//   - logos get nothing: generated once, human-approved, then locked.
func (g *imageryStyleGuide) directionForKind(kind string) string {
	if g == nil {
		return ""
	}
	switch kind {
	case "logo":
		return ""
	case "icon":
		if g.Palette == "" {
			return ""
		}
		return "Colour palette: " + g.Palette
	default:
		parts := make([]string, 0, 3)
		if g.Medium != "" {
			parts = append(parts, g.Medium)
		}
		if g.Mood != "" {
			parts = append(parts, g.Mood)
		}
		if g.Palette != "" {
			parts = append(parts, "colour palette: "+g.Palette)
		}
		return strings.Join(parts, ". ")
	}
}

// resolveReferenceAssetURIs maps the guide's reference asset keys to s3://
// URIs the image adapter's ReferenceFetcher can read. assets.url holds a
// presigned URL whose signature expires — presignedURLToS3URI strips it back
// to the stable s3:// form so anchors keep working long after generation.
// Missing keys are skipped silently: references are enhancement only.
func resolveReferenceAssetURIs(ctx context.Context, db interface{}, siteID string, keys []string, logger *zap.Logger) []string {
	sqlDB, ok := db.(*sql.DB)
	if !ok || siteID == "" || len(keys) == 0 {
		return nil
	}
	const query = `
		SELECT url FROM assets
		WHERE site_id = $1 AND asset_key = $2 AND status = 'active'
		LIMIT 1
	`
	uris := make([]string, 0, len(keys))
	for _, key := range keys {
		var url string
		if err := sqlDB.QueryRowContext(ctx, query, siteID, key).Scan(&url); err != nil {
			if err != sql.ErrNoRows {
				logger.Warn("resolveReferenceAssetURIs: lookup failed",
					zap.String("asset_key", key), zap.Error(err))
			}
			continue
		}
		if uri := presignedURLToS3URI(url); uri != "" {
			uris = append(uris, uri)
		}
	}
	return uris
}
