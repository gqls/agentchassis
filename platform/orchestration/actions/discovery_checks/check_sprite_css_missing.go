// FILE: platform/orchestration/actions/discovery_checks/check_sprite_css_missing.go
//
// Discovery check: a site has a VERIFIED sprite sheet deployed, but the
// stylesheet that slices it (/assets/css/sprites.css) has never been emitted —
// or was emitted from a grid that no longer matches the sheet. Emits a
// needs_sprite_css work item; asset-deployer's sprite_css mode re-runs
// emit_sprite_css, which commits the CSS and re-stamps the plan row.
//
// This is phase I2.4 of the imagery workstream — the fulfilment sibling of
// check_unfulfilled_imagery_plan. That check already covers the first half of
// the gap ("sprite_sheet planned but no asset" → needs_imagery, because it
// emits for ANY unfulfilled plan row regardless of kind). This check covers the
// second half: the asset exists, but the CSS that makes it usable does not.
//
// DB-ONLY, by house convention (see the header of check_image_url_404: discovery
// checks do not make HTTP calls). Fulfilment is therefore read from the stamp
// that emit_sprite_css writes into the plan row's style_hints.sprites_css:
//
//	{"emitted_at": "...", "sheet_path": "...", "signature": "3x3:check,gauge,..."}
//
// We re-emit when:
//   - no stamp exists (never emitted), OR
//   - the stamp's grid signature != the plan's current grid signature
//     (cell names re-verified, or the sheet regenerated at a new geometry —
//     the committed CSS now slices the sheet at the wrong offsets), OR
//   - the sheet asset was updated AFTER the CSS was emitted (sheet regenerated;
//     the CSS may reference a stale path or stale geometry).
//
// Deliberately silent when cell_names_verified is false: emit_sprite_css is
// itself guarded on that gate, so emitting an item would just produce a no-op
// handler run on every discovery pass.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"go.uber.org/zap"
)

func init() { Register(&SpriteCSSMissingCheck{}) }

type SpriteCSSMissingCheck struct{}

func (c *SpriteCSSMissingCheck) Name() string { return "sprite_css_missing" }

// spriteSheetPlanRow is the slice of the plan row this check reasons about.
type spriteSheetPlanRow struct {
	Key        string
	Rows       int      `json:"rows"`
	Cols       int      `json:"cols"`
	CellNames  []string `json:"cell_names"`
	Verified   bool     `json:"cell_names_verified"`
	SpritesCSS *struct {
		EmittedAt string `json:"emitted_at"`
		SheetPath string `json:"sheet_path"`
		Signature string `json:"signature"`
		Format    int    `json:"format"`
	} `json:"sprites_css"`
}

func (c *SpriteCSSMissingCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	var key string
	var styleHints []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT spi.key, COALESCE(spi.style_hints, '{}'::jsonb)
		  FROM site_plan_imagery spi
		  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		 WHERE sp.site_id = $1 AND spi.kind = 'sprite_sheet' AND spi.scope = 'site'
		 ORDER BY spi.ordering
		 LIMIT 1
	`, dctx.SiteID).Scan(&key, &styleHints)
	if err == sql.ErrNoRows {
		return &CheckResult{}, nil // no sprite sheet planned for this site
	}
	if err != nil {
		return nil, fmt.Errorf("sprite_sheet plan row query failed: %w", err)
	}

	var row spriteSheetPlanRow
	if err := json.Unmarshal(styleHints, &row); err != nil {
		dctx.Logger.Warn("sprite_css_missing: unparseable style_hints; skipping",
			zap.String("key", key), zap.Error(err))
		return &CheckResult{}, nil
	}
	row.Key = key

	// The eyeball gate hasn't happened — emit_sprite_css would no-op. Stay quiet.
	if !row.Verified {
		dctx.Logger.Info("sprite_css_missing: sheet not yet cell-verified; no item",
			zap.String("key", key))
		return &CheckResult{}, nil
	}
	if row.Rows < 1 || row.Cols < 1 || len(row.CellNames) == 0 {
		return &CheckResult{}, nil
	}

	// No deployed sheet yet → not our gap; unfulfilled_imagery_plan owns it.
	hasAsset, err := hasActiveAssetForAssetKey(dctx, key)
	if err != nil {
		return nil, err
	}
	if !hasAsset {
		return &CheckResult{}, nil
	}

	// The sheet's mtime is only needed for the "regenerated after emit" case, but
	// fetching it up front keeps the decision itself a pure function (testable
	// without a DB, cf. missingRequiredValueFields).
	var assetUpdatedAt time.Time
	if err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COALESCE(MAX(updated_at), to_timestamp(0))
		  FROM assets
		 WHERE site_id = $1 AND asset_key = $2 AND status = 'active'
	`, dctx.SiteID, row.Key).Scan(&assetUpdatedAt); err != nil {
		dctx.Logger.Warn("sprite_css_missing: asset updated_at lookup failed; treating sheet as not-newer",
			zap.Error(err))
		assetUpdatedAt = time.Time{}
	}

	reason, stale := spriteCSSStaleness(row, assetUpdatedAt)
	if !stale {
		return &CheckResult{}, nil
	}

	specJSON, err := json.Marshal(map[string]interface{}{
		"mode":      "sprite_css",
		"check":     c.Name(),
		"asset_key": key,
		"reason":    reason,
	})
	if err != nil {
		return nil, err
	}

	dctx.Logger.Info("sprite_css_missing: emitting needs_sprite_css",
		zap.String("key", key), zap.String("reason", reason))

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":     c.Name(),
			"asset_key": key,
			"reason":    reason,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID: dctx.SiteID,
			Source: "discovery",
			// Pipeline is the destination, not the origin: needs_sprite_css is
			// handled by asset-deployer on the build pipeline (cf. the same note
			// in check_unfulfilled_imagery_plan).
			Pipeline:     "build",
			ItemType:     "needs_sprite_css",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Sprite sheet %q is live but sprites.css is %s", key, reason),
			SpecJSON:     string(specJSON),
			Priority:     70,
			HandlerAgent: "asset-deployer",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      "needs_sprite_css:" + key,
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// spriteCSSStaleness decides whether sprites.css needs (re-)emitting, and says
// why. Pure: all DB reads happen in the caller. A zero assetUpdatedAt means
// "unknown / not newer".
func spriteCSSStaleness(row spriteSheetPlanRow, assetUpdatedAt time.Time) (string, bool) {
	if row.SpritesCSS == nil || row.SpritesCSS.EmittedAt == "" {
		return "missing", true
	}

	want := imageryplan.SpriteGridSignature(row.Rows, row.Cols, row.CellNames)
	if row.SpritesCSS.Signature != want {
		// The committed CSS slices the sheet at offsets that no longer match the
		// plan's grid/vocabulary — worse than missing, because it renders the
		// WRONG glyphs rather than none.
		return "stale (grid or cell names changed since it was emitted)", true
	}

	// The sheet is unchanged but the EMITTER has moved on (new rules, fixed
	// selectors). Without this, a site would serve a stylesheet from an old
	// emitter version forever, because the grid signature still matches.
	if row.SpritesCSS.Format != imageryplan.SpriteCSSFormat {
		return fmt.Sprintf("stale (emitted by CSS format v%d; current is v%d)",
			row.SpritesCSS.Format, imageryplan.SpriteCSSFormat), true
	}

	emittedAt, err := time.Parse(time.RFC3339, row.SpritesCSS.EmittedAt)
	if err != nil {
		// Can't reason about freshness — re-emit rather than leave a possibly
		// stale stylesheet live. Re-emitting is idempotent.
		return "stale (unparseable emit timestamp)", true
	}
	if assetUpdatedAt.After(emittedAt) {
		return "stale (sheet regenerated after the CSS was emitted)", true
	}

	return "", false
}
