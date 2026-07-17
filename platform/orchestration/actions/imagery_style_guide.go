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
//	  "reference_asset_keys": ["hero_canonical"],
//	  "kinds": {
//	    "content_hero": {
//	      "medium": "flat duotone editorial illustration",
//	      "avoid":  "photorealism, gradients, text",
//	      "reference_asset_keys": []
//	    }
//	  }
//	}
//
// generate_image composes a per-KIND direction from it (photographic kinds
// get medium+mood+palette; icons get palette only; logos get nothing — the
// 2026-05-20 contamination lesson), sends `avoid` to the negative prompt,
// and resolves reference_asset_keys to s3:// URIs as style anchors for
// providers that accept reference images (Banana). When the guide yields a
// direction it supersedes the free-text design_intent.imagery_direction
// fallback — one coherent brand voice, no double prepend.
//
// Phase I3.1 (D14): the optional `kinds` map holds per-kind OVERRIDES for
// kinds whose visual language deliberately differs from the site's base
// voice — the driving case is content_hero, where article heroes/cards moved
// to flat duotone illustration because photographic SDXL output failed the
// card-size gate (colour drift, medium drift, text artefacts). An override
// REPLACES the guide-level fields wholesale for its kind — including empty
// values: the base `avoid` ("cartoonish rendering, decorative colour"…) and
// photographic reference anchors would actively fight a flat-illustration
// direction, so partial merging would reintroduce exactly the contamination
// the override exists to prevent.

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
	// Kinds holds per-kind overrides (Phase I3.1, D14). A present entry
	// replaces every guide-level field for that kind — see file header.
	Kinds map[string]imageryStyleGuideKindOverride `json:"kinds"`
}

// imageryStyleGuideKindOverride is one kind's complete replacement voice.
type imageryStyleGuideKindOverride struct {
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
		g.Avoid == "" && len(g.ReferenceAssetKeys) == 0 && len(g.Kinds) == 0 {
		return nil
	}
	return &g
}

// composeDirection joins medium/mood/palette into the prompt-prefix voice.
func composeDirection(medium, mood, palette string) string {
	parts := make([]string, 0, 3)
	if medium != "" {
		parts = append(parts, medium)
	}
	if mood != "" {
		parts = append(parts, mood)
	}
	if palette != "" {
		parts = append(parts, "colour palette: "+palette)
	}
	return strings.Join(parts, ". ")
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
	// Per-kind override wins outright (D14) — logos stay locked regardless.
	if kind != "logo" {
		if o, ok := g.Kinds[kind]; ok {
			return composeDirection(o.Medium, o.Mood, o.Palette)
		}
	}
	switch kind {
	case "logo":
		return ""
	case "icon", "sprite_sheet":
		// Flat-vector kinds take the brand palette only — a photographic
		// medium prepended to a glyph prompt contaminates the output
		// (2026-05-20 lesson; sprite_sheet added Phase I2).
		if g.Palette == "" {
			return ""
		}
		return "Colour palette: " + g.Palette
	default:
		return composeDirection(g.Medium, g.Mood, g.Palette)
	}
}

// avoidForKind returns the negative-prompt terms for a kind. A per-kind
// override replaces the guide-level avoid even when empty (the base avoid
// may contradict the override's visual language — robot-hands' base forbids
// "cartoonish rendering", which would fight a flat-illustration override).
// Logos get nothing: generated once, human-approved, then locked.
func (g *imageryStyleGuide) avoidForKind(kind string) string {
	if g == nil || kind == "logo" {
		return ""
	}
	if o, ok := g.Kinds[kind]; ok {
		return o.Avoid
	}
	return g.Avoid
}

// referenceKeysForKind returns the style-anchor asset keys for a kind. A
// per-kind override replaces the guide-level list even when empty —
// anchoring a flat-illustration kind to the site's photographic brand
// heroes reproduces the 2026-05-20 contamination failure. An override's
// keys are deliberate, so they flow ungated; the guide-level fallback keeps
// the directionAppliesToKind gate (implicit anchors are the site's
// photographic heroes, wrong for flat-vector kinds).
func (g *imageryStyleGuide) referenceKeysForKind(kind string) []string {
	if g == nil {
		return nil
	}
	if o, ok := g.Kinds[kind]; ok {
		return o.ReferenceAssetKeys
	}
	if !directionAppliesToKind(kind) {
		return nil
	}
	return g.ReferenceAssetKeys
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
