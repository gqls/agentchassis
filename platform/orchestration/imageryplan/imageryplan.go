// FILE: platform/orchestration/imageryplan/imageryplan.go
//
// Shared selection + classification logic for the current site plan's imagery
// requests (site_plan_imagery rows). Used by both emitters:
//   - actions.EmitImageryItemsAction          (build-time, status 'triaged')
//   - discovery_checks.UnfulfilledImageryPlanCheck (loop, status 'detected')
//
// Extracted here so the two stop duplicating the query, the priority bands,
// the brand_update rule, the item_key shape, and the spec body. Each call site
// keeps only its own work-item struct construction (WorkItemSpec vs the
// actions.workItem) and its own status, because those types live in their
// respective packages.

package imageryplan

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// MaxPerPass caps how many needs_imagery items either emitter queues in one
// run. Larger plans complete over multiple passes (loop) or are topped up by
// the loop after a build (build-time emitter).
const MaxPerPass = 20

// Row mirrors one site_plan_imagery row on the current plan.
type Row struct {
	Scope       string
	ScopeRef    *string // nil for site-scope
	Key         string
	Kind        string
	Prompt      string
	StyleHints  json.RawMessage // raw JSON; len 0 when the column was null
	Constraints json.RawMessage // raw JSON; len 0 when the column was null
}

// Queryer is satisfied by *sql.DB and *sql.Tx.
type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// RowQueryer is satisfied by *sql.DB and *sql.Tx.
type RowQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// LoadCurrentPlan returns all imagery rows on the current plan for siteID,
// ordered site → page → section, then scope_ref, then ordering. Empty slice
// when there is no current plan or it carries no imagery rows. siteID is passed
// straight through as a query argument, so any type the driver accepts works.
func LoadCurrentPlan(ctx context.Context, db Queryer, siteID interface{}) ([]Row, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT spi.scope, spi.scope_ref, spi.key, spi.kind, spi.prompt,
		       COALESCE(spi.style_hints::text, ''),
		       COALESCE(spi.constraints::text, '')
		  FROM site_plan_imagery spi
		  JOIN site_plans sp ON sp.id = spi.plan_id
		 WHERE sp.site_id = $1
		   AND sp.is_current = true
		 ORDER BY
		     CASE spi.scope
		         WHEN 'site'    THEN 0
		         WHEN 'page'    THEN 1
		         WHEN 'section' THEN 2
		         ELSE 3
		     END,
		     spi.scope_ref NULLS FIRST,
		     spi.ordering
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("query site_plan_imagery: %w", err)
	}
	defer rows.Close()

	var out []Row
	for rows.Next() {
		var r Row
		var scopeRef sql.NullString
		var styleHints, constraints string
		if err := rows.Scan(&r.Scope, &scopeRef, &r.Key, &r.Kind, &r.Prompt, &styleHints, &constraints); err != nil {
			return nil, fmt.Errorf("scan site_plan_imagery row: %w", err)
		}
		if scopeRef.Valid {
			r.ScopeRef = &scopeRef.String
		}
		if styleHints != "" {
			r.StyleHints = json.RawMessage(styleHints)
		}
		if constraints != "" {
			r.Constraints = json.RawMessage(constraints)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate site_plan_imagery rows: %w", err)
	}
	return out, nil
}

// Classify returns the canonical (priority, severity) bands. Foundational
// imagery is high priority; variants and decoratives trail:
//
//	 65 — page.index hero
//	 70 — site.logo
//	 75 — other site-scope
//	 80 — page-scope hero (non-index)
//	 90 — page-scope non-hero
//	100 — section-scope
//
// Note: the build-time emitter clamps section-scope below the terminal
// needs_rerender priority (99) so it lands in the first deploy; the loop keeps
// 100 for its catch-up purpose. The clamp is applied at the call site, not here.
func Classify(scope, kind string, scopeRef *string) (priority int, severity string) {
	if scope == "page" && scopeRef != nil && *scopeRef == "index" && kind == "hero" {
		return 65, "high"
	}
	if scope == "site" && kind == "logo" {
		return 70, "high"
	}
	if scope == "site" {
		return 75, "high"
	}
	if scope == "page" && kind == "hero" {
		return 80, "medium"
	}
	if scope == "page" {
		return 90, "medium"
	}
	return 100, "low"
}

// BrandUpdate is true for site-scope imagery and the canonical index-page hero
// — the imagery that belongs in the site-wide brand asset slot.
func BrandUpdate(scope, kind string, scopeRef *string) bool {
	return scope == "site" ||
		(scope == "page" && scopeRef != nil && *scopeRef == "index" && kind == "hero")
}

// AssetKey maps a row to its asset_key (direct mapping from key; the planner
// prompt enforces key uniqueness within scope).
func AssetKey(r Row) string { return r.Key }

// ItemKey is the deterministic work-item dedup key for a row. The dash
// placeholder keeps the format positional for site-scope rows (null scope_ref).
func ItemKey(r Row) string {
	scopeRef := "-"
	if r.ScopeRef != nil {
		scopeRef = *r.ScopeRef
	}
	return fmt.Sprintf("needs_imagery:%s:%s:%s", r.Scope, scopeRef, r.Key)
}

// Summary is the human-readable work-item summary for a row.
func Summary(r Row) string {
	return fmt.Sprintf("Imagery %s/%s (kind=%s) requested but no asset for %s",
		r.Scope, r.Key, r.Kind, AssetKey(r))
}

// SpriteCSSFormat is the shape-version of the stylesheet emit_sprite_css
// produces. The grid signature below tracks the sheet's geometry/vocabulary, but
// the emitter can change the CSS itself without the grid moving — I2.5 added the
// `.sprite-bullets` container opt-in, for example. Without a version, the
// sprite_css_missing check would compare an unchanged signature and conclude the
// committed stylesheet was still current, so sites would keep serving the old CSS
// forever.
//
// BUMP THIS whenever buildSpriteCSS changes what it emits. Every site's next
// discovery pass then re-emits once and re-stamps.
//
//	1 = base + .sprite-<glyph> + ul.sprite-list bullets (scoped per-item overrides)
//	2 = adds the .sprite-bullets container opt-in (themes lists in generated content)
//	3 = default list bullet = arrow (was check); check now explicit-only
const SpriteCSSFormat = 3

// SpriteGridSignature identifies the geometry + cell vocabulary that a sprite
// stylesheet was built from, e.g. "3x3:check,gauge,gripper,...".
//
// It lives here because BOTH sides of the sprite-CSS loop depend on it and must
// not drift: the build side (actions.EmitSpriteCSSAction) stamps it into the
// plan row's style_hints.sprites_css after committing sprites.css, and the loop
// side (discovery_checks.SpriteCSSMissingCheck) recomputes it from the current
// plan to decide whether the committed CSS is stale. If the two ever disagreed,
// the check would either re-emit forever or never notice a regenerated sheet.
// (Same reasoning as Classify/ItemKey/BuildSpec above.)
func SpriteGridSignature(rows, cols int, names []string) string {
	return fmt.Sprintf("%dx%d:%s", rows, cols, strings.Join(names, ","))
}

// BuildSpec returns the JSON spec body image-build-handler reads. checkName
// identifies the emitter ("emit_imagery_items" or "unfulfilled_imagery_plan").
func BuildSpec(r Row, checkName string) (string, error) {
	spec := map[string]interface{}{
		"check":        checkName,
		"scope":        r.Scope,
		"key":          r.Key,
		"kind":         r.Kind,
		"asset_key":    AssetKey(r),
		"purpose":      r.Kind,
		"prompt":       r.Prompt,
		"brand_update": BrandUpdate(r.Scope, r.Kind, r.ScopeRef),
	}
	if r.ScopeRef != nil {
		spec["scope_ref"] = *r.ScopeRef
	}
	if len(r.StyleHints) > 0 {
		spec["style_hints"] = r.StyleHints
	}
	if len(r.Constraints) > 0 {
		spec["constraints"] = r.Constraints
	}
	b, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("marshal imagery spec: %w", err)
	}
	return string(b), nil
}

// HasActiveAsset reports whether an active asset exists for siteID + assetKey.
func HasActiveAsset(ctx context.Context, db RowQueryer, siteID interface{}, assetKey string) (bool, error) {
	var n int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM assets
		WHERE site_id = $1 AND asset_key = $2 AND status = 'active'
	`, siteID, assetKey).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// imageRoleAliases maps generic image field paths (the <path> in a component
// schema's source "site_assets.<path>") onto the image ROLE the pipeline can
// satisfy. Preset and imported components name their image fields freely
// (background, product_screenshot, image, ...) while the plan pipeline
// generates per-page heroes plus a site logo — without this mapping those
// fields resolve to nothing and templates render src="" (2026-07-09 finding,
// PLAN_imagery_best_in_class.md Phase I0).
//
// Shared here so the section resolver (plan_sections) and the
// image_source_unsatisfiable discovery check cannot drift. Literal asset keys
// always win — callers consult this only after an exact lookup misses, so a
// future dedicated product image (content-imagery lane, Phase I3) takes
// precedence automatically the moment it exists under the literal key.
var imageRoleAliases = map[string]string{
	"background":       "hero",
	"background_image": "hero",
	"image":            "hero",
	"hero_image":       "hero",
	"hero_background":  "hero",
	"banner":           "hero",
	"header_image":     "hero",
	// Product/secondary imagery has no dedicated generator yet (Phase I3).
	// Interim: the page hero, so nothing renders an empty src.
	"product_screenshot": "hero",
	"product_image":      "hero",
	"screenshot":         "hero",
}

// ImageRoleForPath maps a generic site_assets.<path> image field name to the
// image role the pipeline can satisfy (currently only "hero"). ok is false
// for paths with no alias — the caller treats those as literal asset keys.
func ImageRoleForPath(path string) (string, bool) {
	role, ok := imageRoleAliases[path]
	return role, ok
}
