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
	Palette string `json:"palette"`
	Medium  string `json:"medium"`
	Mood    string `json:"mood"`
	Avoid   string `json:"avoid"`
	// Provider optionally pins image generation to a named provider
	// ("banana" | "stability"). Empty means no opinion and the adapter's
	// per-kind default applies. This is the site-owned escape hatch from
	// the adapter's hardcoded kind routing (bugs_open/011 R1) — the adapter
	// has no DB access, so a site that wants a different model for a kind
	// than the fleet default has no other way to say so.
	Provider           string   `json:"provider"`
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
	Provider           string   `json:"provider"`
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
		g.Avoid == "" && g.Provider == "" &&
		len(g.ReferenceAssetKeys) == 0 && len(g.Kinds) == 0 {
		return nil
	}
	return &g
}

// ============================================================================
// Fallback: the composed palette, for flat kinds on sites with no style guide
// ============================================================================
//
// THE DEFECT THIS CLOSES. A flat-vector kind (icon, sprite_sheet, content_hero)
// takes its colour ONLY from the style guide — directionAppliesToKind excludes
// these kinds from the free-text design_intent.imagery_direction fallback on
// purpose, because a photographic direction prepended to a flat prompt makes
// the model composite the flat subject onto a photo (icon_cycle_time,
// 2026-05-20). That exclusion is right. What it left behind is a hole: a site
// with no imagery_style_guide gives an icon NO colour direction at all.
//
// Into that hole went a literal. build-site-planner's prompt template pins
// every icon prompt to "a darker grey (#4A4A4A) line on a flat solid light grey
// (#EEEEEE) background", unconditionally, for every site. That literal was added
// deliberately (053_build_site_planner.sql, "icon background: transparent/plain
// -> flat selectable grey (embrace the chip)") to stop transparency and
// checkerboard artefacts, and the flatness half of it is load-bearing. The
// colour half is not — it is a LIGHT ground shipped to every site regardless of
// scheme.
//
// Measured fleet-wide 2026-07-29: 92 site_plan_imagery rows across 14 plans
// carry #EEEEEE, 62 of them on 9 sites' CURRENT plans; and 6 of the 10 sites
// with a resolved composition are dark-background (dartsonline #111520,
// fundamentallyai #080E1C, gamesdesign #121212, oufe #111922, robot-hands
// #0F1218, vonc #0a0a0f). On dartsonline the result was 17 icons generated
// exactly as instructed, every one a near-white tile, none usable.
//
// So: when no guide exists, derive the flat-kind palette clause from the palette
// the RENDERER actually emits. That is the same source check_palette_contrast
// reads, and it means a site gets a correct ground with no per-site config at
// all. A guide still wins outright when one exists — this is strictly a floor.

// composedPaletteDirection builds the palette clause for a flat-vector kind on
// a site with no imagery_style_guide, from the site's resolved composition.
//
// Returns "" when the site has no resolved composition, no palette row, or no
// background/text slots — every one of which means "nothing better to say than
// the prompt already says", and the prompt still ships. It is a floor, never a
// requirement, in keeping with getImageryStyleGuideForSite above.
//
// The SQL is deliberately the same shape as
// discovery_checks.loadComposedPalette (check_palette_contrast.go:156-168) —
// site_specs.resolved_composition.palette_id joined to palettes.colours. They
// are two copies because discovery_checks is imported BY actions and its helper
// is unexported and takes a DiscoveryCheckContext; if either query changes, the
// other is wrong, and the failure is silent in both (an image generated against
// a palette the page does not use looks deliberate).
func composedPaletteDirection(ctx context.Context, db interface{}, siteID string, logger *zap.Logger) string {
	if siteID == "" || db == nil {
		return ""
	}
	const query = `
		SELECT p.colours
		  FROM site_specs ss
		  JOIN palettes p ON p.id = (ss.data->>'palette_id')::uuid
		 WHERE ss.site_id = $1
		   AND ss.aspect  = 'resolved_composition'
		   AND ss.is_current
		 LIMIT 1`

	var raw []byte
	switch d := db.(type) {
	case *sql.DB:
		if err := d.QueryRowContext(ctx, query, siteID).Scan(&raw); err != nil {
			if err != sql.ErrNoRows {
				logger.Warn("composedPaletteDirection: query failed",
					zap.String("site_id", siteID), zap.Error(err))
			}
			return ""
		}
	case *pgxpool.Pool:
		if err := d.QueryRow(ctx, query, siteID).Scan(&raw); err != nil {
			if !strings.Contains(err.Error(), "no rows") {
				logger.Warn("composedPaletteDirection: query failed",
					zap.String("site_id", siteID), zap.Error(err))
			}
			return ""
		}
	default:
		return ""
	}

	var colours map[string]string
	if err := json.Unmarshal(raw, &colours); err != nil {
		logger.Warn("composedPaletteDirection: unparseable palettes.colours",
			zap.String("site_id", siteID), zap.Error(err))
		return ""
	}

	// DARK SITES ONLY, and this is a deliberate narrowing rather than an
	// oversight. It is the same boundary fillDarkSchemeSpecialisedSlots draws,
	// for the same reason and stated in its own words: "for a light site the
	// layout's literals are broadly right, and four live palettes rely on them
	// today. Deriving for them too would repaint live sites to no benefit."
	//
	// A light grey icon ground on a light site is unremarkable; on a dark site
	// it is a hole in the page. And because these two mechanisms are a PAIR —
	// the generated ground and the chip it lands in must agree — their gates
	// have to match, or a site gets one half of the treatment. Keeping the
	// boundaries identical is the property to preserve if either is changed.
	//
	// Measured 2026-07-29: 6 of the 10 sites with a resolved composition are
	// dark (dartsonline, fundamentallyai, gamesdesign, oufe, robot-hands, vonc);
	// the other 4 (idea.uk, relojistas, vetcomparison, webdesign) keep the
	// planner's literal and are unaffected by this change.
	if bg := colours["background"]; bg == "" || !isDarkHex(bg) {
		return ""
	}
	return flatKindPaletteClause(colours)
}

// flatKindPaletteClause turns a palette into the one sentence a flat-vector
// prompt needs: what colour the ground is, what colour the linework is, and
// what may be used for a single accent stroke.
//
// GROUND IS `surface`, NOT `background`, and that is not a detail. An icon is
// rendered inside a styled chip on a card, and the chip takes surface_alt which
// itself derives from surface (palette_specialised_slots.go). If the generated
// ground and the chip disagree, the icon reads as a tile pasted onto the card —
// which is the SAME visual defect as the pale icons, merely inverted. The two
// must be derived from one slot, and this is that slot.
//
// Falls back to `background` when a palette omits `surface`, and returns "" when
// it has neither — a clause naming a colour we do not have would be worse than
// the prompt's own literal.
func flatKindPaletteClause(colours map[string]string) string {
	ground := colours["surface"]
	if ground == "" {
		ground = colours["background"]
	}
	if ground == "" {
		return ""
	}

	// Linework must be legible ON the ground, so it is picked the same way an
	// ink slot is rather than assumed to be `text`: a palette whose text does
	// not clear AA against its own surface exists (this is why pickInkOn takes
	// a candidate list at all), and an unreadable icon is exactly the failure
	// being fixed.
	//
	// pickInkOn never returns empty — when nothing in the palette clears AA it
	// falls back to whichever of black/white contrasts better. So there is no
	// "no ink" branch to guard, and adding one would be dead code that reads
	// like a real case. (I wrote that guard first and a test caught it: the
	// helper returned "#ffffff" where I had asserted "".)
	ink, _ := pickInkOn(ground, colours)

	clause := "flat solid " + ground + " background, " + ink + " linework"
	if accent := colours["accent"]; accent != "" {
		clause += ", " + accent + " permitted for a single accent stroke only"
	}
	// Said explicitly because the model is also being handed a prompt that may
	// still carry the planner's own colour words. Naming the ground twice, once
	// as an instruction and once as a prohibition, is what makes the derived
	// value win over a stale literal in the subject prompt.
	clause += "; one single uniform background colour, and no other background colour"
	return clause
}

// composeDirection joins medium/mood/palette into the prompt-prefix voice.
func composeDirection(medium, mood, palette string) string {
	parts := make([]string, 0, 3)
	// Palette FIRST (2026-07-20, bugs_open/027 §4b): the composed direction is
	// truncated at maxImageryDirectionInPrompt, and the palette is both the
	// shortest field and the one carrying brand identity, so it must not be
	// the tail. Composed last, a verbose medium+mood silently evicted the
	// colours entirely and the model invented an accent per image — and the
	// output looked deliberate. medium/mood are prose and degrade gracefully
	// when clipped; every palette in the fleet fits inside the cap whole.
	if palette != "" {
		parts = append(parts, "colour palette: "+palette)
	}
	if medium != "" {
		parts = append(parts, medium)
	}
	if mood != "" {
		parts = append(parts, mood)
	}
	return strings.Join(parts, ". ")
}

// directionForKind builds the prompt-prefix direction appropriate to an
// image kind. Mirrors directionAppliesToKind's gating philosophy:
//   - photographic kinds (hero, illustration, infographic, legacy default)
//     get the full brand voice: medium, mood, palette;
//   - flat-vector kinds (icon, sprite_sheet, content_hero) get ONLY the
//     palette — a photographic medium prepended to a flat prompt makes the
//     model composite the flat subject onto a photo (icon_cycle_time,
//     2026-05-20);
//   - logos get nothing: generated once, human-approved, then locked.
//
// content_hero (D14, a flat duotone illustration kind) takes its full voice
// from a per-kind override when one exists — that is the intended path, hit by
// every site that has configured it. WITHOUT an override it must NOT fall to
// the photographic base voice (bugs_open/027 §5(a)): that is the same
// contamination as prepending the photographic medium to an icon. So it joins
// the palette-only branch, in lockstep with its exclusion from
// directionAppliesToKind — content_hero is flat in BOTH gating functions.
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
	case "icon", "sprite_sheet", "content_hero":
		// Flat-vector kinds take the brand palette only — a photographic
		// medium prepended to a flat prompt contaminates the output
		// (2026-05-20 lesson; sprite_sheet added Phase I2, content_hero
		// bugs_open/027 §5(a)).
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
	if g == nil || kind == "logo" {
		// Logos stay locked in all three accessors, override or not: a logo is
		// generated once, human-approved and then frozen, and anchoring it to
		// other brand imagery is how it drifts. directionForKind and
		// avoidForKind both suppress logo; this one did not, which would have
		// let a stray kinds.logo entry feed it anchors (council gate 098b29b8,
		// editquality).
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

// providerForKind returns the site's image-provider preference for a kind
// ("banana" | "stability"), or "" for no opinion — in which case the adapter
// applies its per-kind default.
//
// Mirrors avoidForKind's override semantics: a per-kind override replaces the
// guide-level value wholesale, INCLUDING when empty. A site whose base voice
// pins one provider may deliberately want a single kind left on the fleet
// default, and partial merging would make that impossible to express.
//
// Unlike the other three accessors, logo is NOT special-cased. Those suppress
// logo because prepending style to a locked logo CONTAMINATES it; choosing
// which model renders it is not contamination, and logo already routes to
// Banana either way.
func (g *imageryStyleGuide) providerForKind(kind string) string {
	if g == nil {
		return ""
	}
	if o, ok := g.Kinds[kind]; ok {
		return strings.TrimSpace(o.Provider)
	}
	return strings.TrimSpace(g.Provider)
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
