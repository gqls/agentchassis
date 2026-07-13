// FILE: platform/orchestration/actions/discovery_checks/check_image_source_unsatisfiable.go
//
// Discovery check: components asking for images the pipeline cannot supply.
//
// Walks every page component's input_schema and collects image-typed fields
// (type "image" / "image_url") whose source is "site_assets.<path>". A path is
// satisfiable when a literal asset key, a plan imagery row, or an image-role
// alias (imageryplan.ImageRoleForPath — the same table plan_sections resolves
// with, so the two cannot drift) can supply it, or the legacy content_data
// hero_url/logo_url fallback exists. Anything else renders src="" or bakes a
// placeholder — the 2026-07-09 robot-hands finding (empty product-detail
// images, one shared hero on eight pages), generalised so every FUTURE domain
// gets flagged automatically instead of eyeballed.
//
// Flag-only: items carry no handler agent and are emitted with status
// needs_human_review so the dispatch loop never claims them. The dedup key
// (site, page, component function, path) keeps one open flag per gap.

package discovery_checks

import (
	"encoding/json"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"go.uber.org/zap"
)

func init() { Register(&ImageSourceUnsatisfiableCheck{}) }

type ImageSourceUnsatisfiableCheck struct{}

func (c *ImageSourceUnsatisfiableCheck) Name() string { return "image_source_unsatisfiable" }

// maxUnsatisfiableFlagsPerPass bounds noise on badly-shaped sites; remaining
// gaps surface on later passes once earlier flags are resolved.
const maxUnsatisfiableFlagsPerPass = 25

func (c *ImageSourceUnsatisfiableCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// ── What the pipeline can supply ──
	imagery, err := imageryplan.LoadCurrentPlan(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	literalKeys := make(map[string]bool, len(imagery))
	pageHero := make(map[string]bool)
	siteHero, siteLogo := false, false
	for _, row := range imagery {
		literalKeys[row.Key] = true
		switch {
		case row.Scope == "page" && row.Kind == "hero" && row.ScopeRef != nil:
			pageHero[*row.ScopeRef] = true
		case row.Scope == "site" && row.Kind == "hero":
			siteHero = true
		case row.Scope == "site" && row.Kind == "logo":
			siteLogo = true
		}
	}

	// Active asset keys satisfy literal paths even without a plan row
	// (adopted/legacy sites).
	assetRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT asset_key FROM assets
		 WHERE site_id = $1 AND status = 'active' AND asset_key IS NOT NULL
	`, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	defer assetRows.Close()
	for assetRows.Next() {
		var key string
		if err := assetRows.Scan(&key); err == nil && key != "" {
			literalKeys[key] = true
		}
	}
	if err := assetRows.Err(); err != nil {
		return nil, err
	}

	// Legacy content_data fallback (what ensureAssets gap-fills from).
	var contentHeroURL, contentLogoURL string
	if err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COALESCE(content_data->>'hero_url', ''),
		       COALESCE(content_data->>'logo_url', '')
		  FROM sites WHERE id = $1
	`, dctx.SiteID).Scan(&contentHeroURL, &contentLogoURL); err != nil {
		dctx.Logger.Warn("image_source_unsatisfiable: content_data lookup failed", zap.Error(err))
	}

	heroSatisfiable := func(page string) bool {
		return pageHero[page] || siteHero || contentHeroURL != ""
	}

	// ── What components ask for ──
	compRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT DISTINCT p.name, cc.function, cc.input_schema::text
		  FROM page_components pc
		  JOIN pages p ON p.id = pc.page_id
		  JOIN content_components cc ON cc.id = pc.component_id
		 WHERE p.site_id = $1
		   AND cc.input_schema IS NOT NULL
	`, dctx.SiteID)
	if err != nil {
		return nil, err
	}
	defer compRows.Close()

	result := &CheckResult{}
	emitted := 0

	for compRows.Next() {
		var pageName, function, schemaText string
		if err := compRows.Scan(&pageName, &function, &schemaText); err != nil {
			continue
		}
		var schema map[string]interface{}
		if err := json.Unmarshal([]byte(schemaText), &schema); err != nil {
			continue
		}
		fields, ok := schema["fields"].(map[string]interface{})
		if !ok {
			continue
		}

		for fieldName, defRaw := range fields {
			def, ok := defRaw.(map[string]interface{})
			if !ok {
				continue
			}
			fieldType, _ := def["type"].(string)
			if fieldType != "image" && fieldType != "image_url" {
				continue
			}
			source, _ := def["source"].(string)
			if !strings.HasPrefix(source, "site_assets.") {
				continue
			}
			path := strings.TrimPrefix(source, "site_assets.")

			satisfiable := literalKeys[path]
			if !satisfiable {
				switch {
				case path == "hero":
					satisfiable = heroSatisfiable(pageName)
				case path == "logo":
					satisfiable = siteLogo || contentLogoURL != ""
				default:
					if role, ok := imageryplan.ImageRoleForPath(path); ok && role == "hero" {
						satisfiable = heroSatisfiable(pageName)
					}
				}
			}
			if satisfiable {
				continue
			}

			if emitted >= maxUnsatisfiableFlagsPerPass {
				dctx.Logger.Info("image_source_unsatisfiable: per-pass cap reached",
					zap.Int("cap", maxUnsatisfiableFlagsPerPass))
				return result, nil
			}

			result.Findings = append(result.Findings, map[string]interface{}{
				"check":              "image_source_unsatisfiable",
				"page":               pageName,
				"component_function": function,
				"field":              fieldName,
				"source_path":        path,
			})

			spec, err := json.Marshal(map[string]interface{}{
				"page":               pageName,
				"component_function": function,
				"field":              fieldName,
				"source":             source,
				"reason":             "no asset key, plan imagery row, or image-role alias can supply this source; the field renders empty or falls back to a placeholder",
			})
			if err != nil {
				continue
			}

			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:   dctx.SiteID,
				Source:   "discovery",
				Pipeline: "build",
				ItemType: "image_source_unsatisfiable",
				Severity: "low",
				Summary: "Component '" + function + "' on page '" + pageName +
					"' sources field '" + fieldName + "' from site_assets." + path +
					" which nothing generates",
				SpecJSON: string(spec),
				Priority: 150,
				// Flag-only: no handler, and needs_human_review keeps it out
				// of the dispatch loop while the dedup key holds it open.
				HandlerAgent: "",
				Status:       "needs_human_review",
				CreatedBy:    dctx.AgentType,
				ItemKey:      "image_source_unsatisfiable:" + pageName + ":" + function + ":" + path,
				BatchID:      dctx.BatchID,
			})
			emitted++
		}
	}
	if err := compRows.Err(); err != nil {
		return nil, err
	}

	if emitted > 0 {
		dctx.Logger.Info("image_source_unsatisfiable: flagged unsatisfiable image sources",
			zap.Int("count", emitted))
	}
	return result, nil
}
