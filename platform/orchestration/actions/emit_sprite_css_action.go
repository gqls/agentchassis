// FILE: platform/orchestration/actions/emit_sprite_css_action.go
//
// EmitSpriteCSSAction generates a site's sprite stylesheet from its verified
// sprite-sheet plan row and commits it to /assets/css/sprites.css (Phase I2,
// G4). The "slicing" is pure CSS background-position arithmetic computed from
// the grid geometry (rows/cols) and the fixed sheet size — there is no image
// cropping. Deterministic; no LLM, no storage client (DB read + git-commit
// via the git adapter only).
//
// Emits, from a 3×3 (etc.) sheet at fixed WxH with cells W/cols × H/rows,
// scaled to a small bullet display size T:
//   - `.sprite` base + `.sprite-<name>` per verified cell (inline/icon use),
//   - a themed list-bullet style (`ul.sprite-list li::before`) with a default
//     glyph and per-item `li.sprite-b-<name>` overrides.
//
// GUARD: only emits when the plan row's style_hints.cell_names_verified is
// true — the cell→name map is human-assigned at the eyeball gate, and CSS
// keyed to an unverified (possibly wrong) map would mislabel every glyph.
//
// Runs anywhere with a Kafka producer (route via asset-deployer's sprite_css
// mode). Idempotent: overwrites sprites.css with a freshly-computed copy.
//
// Registration (registry.go):
//   "emit_sprite_css": { Handler: EmitSpriteCSSAction, Category: "site",
//     Description: "Generate and commit sprites.css from the sprite sheet plan",
//     IsLocal: true }

package actions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

var EmitSpriteCSSInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"domain"},
	Defaults: map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("emit_sprite_css", EmitSpriteCSSInputSpec)
}

// spriteBulletDisplayPx is the rendered size of a sprite bullet/icon. Cells
// (256px on a 768² 3×3 sheet) scale down to this; the whole sheet is drawn at
// (cols×T)×(rows×T) via background-size so each cell lands at T×T.
const spriteBulletDisplayPx = 20

// spriteDefaultBulletGlyph is the glyph a themed list item falls back to when it
// specifies none. Neutral by design: the container opt-in themes every content
// list, so the default is a plain list marker (arrow), with `check` reserved for
// explicit `sprite-b-check` where affirmation is wanted (user decision, 2026-07-15).
const spriteDefaultBulletGlyph = "arrow"

func EmitSpriteCSSAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "emit_sprite_css"))

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(params.CollectedData, params.StepConfig.Config, EmitSpriteCSSInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}
	domain := inputs.Get("domain")

	// ── Load the verified sprite-sheet plan row (+ domain fallback) ──
	var planRowID string
	var key string
	var styleHints []byte
	var domainDB string
	err = params.DB.QueryRowContext(ctx, `
		SELECT spi.id, spi.key, spi.style_hints, COALESCE(si.domain, '')
		  FROM site_plan_imagery spi
		  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		  JOIN sites si ON si.id = sp.site_id
		 WHERE sp.site_id = $1 AND spi.kind = 'sprite_sheet' AND spi.scope = 'site'
		 ORDER BY spi.ordering
		 LIMIT 1
	`, siteID).Scan(&planRowID, &key, &styleHints, &domainDB)
	if err != nil {
		return map[string]interface{}{"emitted": false, "reason": "no sprite_sheet plan row"}, nil
	}
	if domain == "" {
		domain = domainDB
	}

	var hints struct {
		Rows      int      `json:"rows"`
		Cols      int      `json:"cols"`
		CellNames []string `json:"cell_names"`
		Verified  bool     `json:"cell_names_verified"`
	}
	if err := json.Unmarshal(styleHints, &hints); err != nil {
		return nil, fmt.Errorf("unparseable style_hints: %w", err)
	}
	if !hints.Verified {
		return map[string]interface{}{"emitted": false, "reason": "cell_names_verified is false — awaiting the eyeball gate"}, nil
	}
	if hints.Rows < 1 || hints.Cols < 1 || len(hints.CellNames) == 0 {
		return map[string]interface{}{"emitted": false, "reason": "incomplete grid plan (rows/cols/cell_names)"}, nil
	}

	// ── Require the deployed sheet to actually exist as an active asset ──
	var assetCount int
	if err := params.DB.QueryRowContext(ctx, `
		SELECT count(*) FROM assets
		 WHERE site_id = $1 AND asset_key = $2 AND status = 'active'
	`, siteID, key).Scan(&assetCount); err != nil {
		return nil, fmt.Errorf("asset existence check: %w", err)
	}
	if assetCount == 0 {
		return map[string]interface{}{"emitted": false, "reason": "no active sprite_sheet asset yet"}, nil
	}

	sheetPath := storage.DeployedWebPath(key, "sprite_sheet") // /assets/images/sprite-sheet-main.jpg
	css := buildSpriteCSS(sheetPath, hints.Rows, hints.Cols, hints.CellNames, domain)

	files := map[string]interface{}{
		"assets/css/sprites.css": map[string]interface{}{
			"content":  base64.StdEncoding.EncodeToString([]byte(css)),
			"encoding": "base64",
		},
	}
	if _, err := sendGitCommitRequest(ctx, params, domain, files, "sprite-css", logger); err != nil {
		return nil, fmt.Errorf("git commit sprites.css: %w", err)
	}

	// Stamp fulfilment state onto the plan row. The sprite_css_missing discovery
	// check is DB-only (house convention — see check_image_url_404), so without
	// this stamp it cannot tell "already emitted" from "never emitted" and would
	// re-emit on every discovery pass. Recording the grid signature alongside the
	// timestamp also makes STALENESS detectable: regenerate the sheet or change
	// the cell names, and the signature no longer matches, so the check re-emits.
	// Best-effort: the CSS is already committed, so a stamp failure must not fail
	// the action — it degrades to a re-emit on the next pass, which is idempotent.
	stamp := map[string]interface{}{
		"emitted_at": time.Now().UTC().Format(time.RFC3339),
		"sheet_path": sheetPath,
		"signature":  imageryplan.SpriteGridSignature(hints.Rows, hints.Cols, hints.CellNames),
		// The signature tracks the SHEET; format tracks the STYLESHEET's shape. An
		// emitter change (e.g. adding the .sprite-bullets opt-in) moves the format
		// without moving the signature, and the check must still re-emit.
		"format": imageryplan.SpriteCSSFormat,
	}
	stampJSON, _ := json.Marshal(stamp)
	if _, err := params.DB.ExecContext(ctx, `
		UPDATE site_plan_imagery
		   SET style_hints = jsonb_set(COALESCE(style_hints, '{}'::jsonb), '{sprites_css}', $2::jsonb, true)
		 WHERE id = $1
	`, planRowID, string(stampJSON)); err != nil {
		logger.Warn("emit_sprite_css: fulfilment stamp failed (CSS committed; check will re-emit next pass)",
			zap.Error(err))
	}

	logger.Info("emit_sprite_css: committed sprites.css",
		zap.String("domain", domain),
		zap.String("sheet", sheetPath),
		zap.Int("cells", len(hints.CellNames)),
		zap.Int("css_bytes", len(css)))

	return map[string]interface{}{
		"emitted":    true,
		"css_path":   "/assets/css/sprites.css",
		"sheet_path": sheetPath,
		"cell_count": len(hints.CellNames),
	}, nil
}

// buildSpriteCSS computes the stylesheet. Cells map in reading order
// (left-to-right, top-to-bottom); cell i → row i/cols, col i%cols. The whole
// sheet is drawn at (cols×T)×(rows×T) so each cell renders at T×T
// (T = spriteBulletDisplayPx).
func buildSpriteCSS(sheetPath string, rows, cols int, names []string, domain string) string {
	T := spriteBulletDisplayPx
	sheetW, sheetH := cols*T, rows*T
	pos := func(i int) string {
		return fmt.Sprintf("%dpx %dpx", -(i%cols)*T, -(i/cols)*T)
	}

	// Default bullet for a list item that specifies no glyph. Since the container
	// opt-in themes EVERY content list, the default wants to be a neutral list
	// marker (arrow), not an affirmation (check) — check stays available as an
	// explicit `sprite-b-check`. Resolved by NAME (cell order varies per sheet);
	// falls back to the first cell if the sheet lacks the named glyph.
	defaultIdx := 0
	for i, name := range names {
		if sanitiseSpriteName(name) == spriteDefaultBulletGlyph {
			defaultIdx = i
			break
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "/* Auto-generated sprite CSS for %s — source %s (%dx%d, verified). Do not hand-edit; regenerated by emit_sprite_css. */\n",
		domain, sheetPath, rows, cols)

	// Base + per-glyph classes (inline/icon/nav use).
	fmt.Fprintf(&b, ".sprite{display:inline-block;background-image:url(%s);background-repeat:no-repeat;background-size:%dpx %dpx;width:%dpx;height:%dpx;vertical-align:middle}\n",
		sheetPath, sheetW, sheetH, T, T)
	for i, name := range names {
		n := sanitiseSpriteName(name)
		if n == "" {
			continue
		}
		fmt.Fprintf(&b, ".sprite-%s{background-position:%s}\n", n, pos(i))
	}

	// Themed list bullets, offered as TWO opt-ins over identical geometry:
	//
	//   ul.sprite-list      — the class sits ON the list. Precise, but only usable
	//                         where we author the markup.
	//   .sprite-bullets ul  — the class sits on a CONTAINER (e.g. a component
	//                         wrapper) and themes every list inside it. This is the
	//                         one that works for generated content: article bodies
	//                         are LLM-written HTML dropped into a template, so their
	//                         <ul>s never carry classes and never can.
	//
	// Both scopes are emitted from the same selector list so the two can't drift.
	listScopes := []string{"ul.sprite-list", "ol.sprite-list", ".sprite-bullets ul", ".sprite-bullets ol"}

	fmt.Fprintf(&b, "%s{list-style:none;padding-left:0}\n", strings.Join(listScopes, ","))
	fmt.Fprintf(&b, "%s{position:relative;padding-left:1.9em;margin-bottom:.4em}\n", joinScoped(listScopes, ">li"))
	fmt.Fprintf(&b, "%s{content:\"\";position:absolute;left:0;top:.15em;width:%dpx;height:%dpx;background-image:url(%s);background-repeat:no-repeat;background-size:%dpx %dpx;background-position:%s}\n",
		joinScoped(listScopes, ">li::before"), T, T, sheetPath, sheetW, sheetH, pos(defaultIdx))

	// Per-item overrides MUST stay scoped under the list/container class. A bare
	// `li.sprite-b-x::before` is specificity (0,1,2) and LOSES to the default rule
	// above (0,1,3), so every bullet silently rendered the default glyph — this
	// shipped once and was only caught by looking at the live page. Scoped, the
	// override is (0,2,3) and wins in both opt-ins.
	for i, name := range names {
		n := sanitiseSpriteName(name)
		if n == "" {
			continue
		}
		fmt.Fprintf(&b, "%s{background-position:%s}\n",
			joinScoped(listScopes, ">li.sprite-b-"+n+"::before"), pos(i))
	}
	return b.String()
}

// joinScoped suffixes every list scope and joins them into one selector list,
// e.g. suffix ">li::before" → "ul.sprite-list>li::before,ol.sprite-list>li::before,…".
func joinScoped(scopes []string, suffix string) string {
	out := make([]string, len(scopes))
	for i, s := range scopes {
		out[i] = s + suffix
	}
	return strings.Join(out, ",")
}

// sanitiseSpriteName lowercases and keeps only [a-z0-9-] so a cell name is a
// safe CSS class fragment.
func sanitiseSpriteName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == '_' || r == ' ':
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
