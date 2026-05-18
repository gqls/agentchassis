// FILE: platform/orchestration/actions/discovery_checks/check_unfulfilled_imagery_plan.go
//
// Discovery check: site_plan_imagery rows whose assets are missing.
//
// Phase 2G step 4. Reads the current plan's imagery rows (populated by
// build-site-planner step 3 + write_site_plan_action's flattenImageryBlock)
// and emits a needs_imagery work item for each row whose asset_key is not
// yet present in active assets for the site.
//
// Sibling of check_unfulfilled_image_prompt.go (Phase 1.1) which reads
// from site_specs.site_plan.image_prompts. Both checks coexist during the
// transition window; once the legacy image_prompts path ages out
// (Phase 2G step 6), the older check is de-registered.
//
// Design decisions (settled before code):
//
//   1. asset_key = imagery.key directly. The planner prompt enforces
//      key uniqueness within scope; cross-scope collisions are detected
//      and logged but do not block emission. Practical reliance: the
//      prompt encourages page-prefixed keys (hero_about, hero_tools)
//      precisely to avoid this.
//
//   2. New item_type "needs_imagery". The existing variant-routing
//      item_type "unfulfilled_hero_variant" assumes "hero" semantics in
//      its name, wrong for logos / illustrations / icons / infographics.
//      image-build-handler picks up needs_imagery via step 5's workflow
//      extension.
//
//   3. Per-pass cap: maxImageryWorkItemsPerPass items. Larger plans
//      complete over multiple discovery passes. Site-scope and
//      page.index hero are emitted first (highest priority),
//      section-scope decoratives last.
//
//   4. item_key = "needs_imagery:<scope>:<scope_ref|->:<key>". Fully
//      deterministic from the imagery row — no query needed to derive
//      it. The dash placeholder keeps the format positional for
//      site-scope rows where scope_ref is null.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// maxImageryWorkItemsPerPass caps how many needs_imagery items the check
// emits in a single run. Per-pass throttling prevents dispatch flooding
// for sites with rich imagery requirements (multi-page sites can have
// 20+ imagery rows once page-scope heroes land per page). Items deferred
// past the cap are picked up on subsequent discovery passes.
const maxImageryWorkItemsPerPass = 20

func init() { Register(&UnfulfilledImageryPlanCheck{}) }

type UnfulfilledImageryPlanCheck struct{}

func (c *UnfulfilledImageryPlanCheck) Name() string { return "unfulfilled_imagery_plan" }

// plannedImagery mirrors one row from site_plan_imagery for the current plan.
type plannedImagery struct {
	Scope       string
	ScopeRef    *string // nil for site-scope
	Key         string
	Kind        string
	Prompt      string
	StyleHints  json.RawMessage // raw JSON; len 0 if column was null
	Constraints json.RawMessage // raw JSON; len 0 if column was null
}

// classifyImageryRow returns (priority, severity) using the same intuition
// as classifyPromptKey in check_unfulfilled_image_prompt.go: foundational
// imagery is high priority; variants and decoratives trail.
//
// Priority numbers match the legacy check's bands so the two checks
// produce comparable queue orderings during transition:
//
//	 65 — page.index hero        (mirrors hero_home)
//	 70 — site.logo               (mirrors logo)
//	 75 — other site-scope        (brand-level supporting imagery)
//	 80 — page-scope hero (non-index, mirrors hero variants)
//	 90 — page-scope non-hero
//	100 — section-scope decoratives
func classifyImageryRow(scope, kind string, scopeRef *string) (priority int, severity string) {
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

func (c *UnfulfilledImageryPlanCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	imagery, err := loadCurrentPlanImagery(dctx)
	if err != nil {
		return nil, err
	}
	if len(imagery) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{}
	emitted := 0
	seenAssetKeys := make(map[string]bool, len(imagery))

	for _, row := range imagery {
		if emitted >= maxImageryWorkItemsPerPass {
			dctx.Logger.Info("unfulfilled_imagery_plan: per-pass cap reached; remaining rows picked up next pass",
				zap.Int("cap", maxImageryWorkItemsPerPass),
				zap.Int("plan_rows_total", len(imagery)),
				zap.Int("emitted_this_pass", emitted))
			break
		}

		assetKey := row.Key // option (a): direct mapping

		// Cross-scope collision detection — visibility only, non-blocking.
		// The planner prompt steers toward unique keys; if two rows still
		// resolve to the same asset_key, the second emission collides on
		// the (site_id, item_key) dedup index for the work item, but the
		// asset_key collision in the assets table is the real risk and
		// only matters at image-build-handler time.
		if seenAssetKeys[assetKey] {
			dctx.Logger.Warn("unfulfilled_imagery_plan: asset_key collision across imagery rows",
				zap.String("asset_key", assetKey),
				zap.String("scope", row.Scope),
				zap.String("key", row.Key))
		}
		seenAssetKeys[assetKey] = true

		hasAsset, err := hasActiveAssetForAssetKey(dctx, assetKey)
		if err != nil {
			dctx.Logger.Warn("unfulfilled_imagery_plan: asset existence check failed",
				zap.String("asset_key", assetKey),
				zap.Error(err))
			continue
		}
		if hasAsset {
			continue
		}

		priority, severity := classifyImageryRow(row.Scope, row.Kind, row.ScopeRef)

		// brand_update rule (b): site-scope imagery, OR the canonical
		// index-page hero. The site_brand_assets JSON on sites table only
		// holds one logo / hero — site-wide brand identity. Page-scope
		// heroes on non-index pages and section-scope decoratives don't
		// belong there. Computed here so the workflow doesn't have to
		// know the rule; it just routes on spec.brand_update.
		brandUpdate := row.Scope == "site" ||
			(row.Scope == "page" && row.ScopeRef != nil && *row.ScopeRef == "index" && row.Kind == "hero")

		// Spec carries everything image-build-handler needs once step 5
		// adds the needs_imagery branch. The legacy hero variant chain
		// reads spec.prompt, spec.purpose, spec.asset_key; this spec
		// is a superset.
		spec := map[string]interface{}{
			"check":        "unfulfilled_imagery_plan",
			"scope":        row.Scope,
			"key":          row.Key,
			"kind":         row.Kind,
			"asset_key":    assetKey,
			"purpose":      row.Kind, // step 5 may refine kind→purpose mapping
			"prompt":       row.Prompt,
			"brand_update": brandUpdate,
		}
		if row.ScopeRef != nil {
			spec["scope_ref"] = *row.ScopeRef
		}
		if len(row.StyleHints) > 0 {
			spec["style_hints"] = row.StyleHints
		}
		if len(row.Constraints) > 0 {
			spec["constraints"] = row.Constraints
		}

		specJSON, err := json.Marshal(spec)
		if err != nil {
			dctx.Logger.Warn("unfulfilled_imagery_plan: spec marshal failed",
				zap.String("asset_key", assetKey),
				zap.Error(err))
			continue
		}

		scopeRefDisplay := "-"
		if row.ScopeRef != nil {
			scopeRefDisplay = *row.ScopeRef
		}
		itemKey := fmt.Sprintf("needs_imagery:%s:%s:%s", row.Scope, scopeRefDisplay, row.Key)

		summary := fmt.Sprintf("Imagery %s/%s (kind=%s) requested but no asset for %s",
			row.Scope, row.Key, row.Kind, assetKey)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "unfulfilled_imagery_plan",
			"scope":     row.Scope,
			"key":       row.Key,
			"kind":      row.Kind,
			"asset_key": assetKey,
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID: dctx.SiteID,
			Source: "discovery",
			// Pipeline is the destination, not the origin. needs_imagery routes
			// to image-build-handler (build pipeline) regardless of which
			// discovery agent invoked us. dctx.Pipeline gives us the origin
			// (e.g. "design" from design-discovery-agent), which is the wrong
			// axis for this field.
			Pipeline:     "build", // dctx.Pipeline,
			ItemType:     "needs_imagery",
			Severity:     severity,
			Summary:      summary,
			SpecJSON:     string(specJSON),
			Priority:     priority,
			HandlerAgent: "image-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      itemKey,
			BatchID:      dctx.BatchID,
		})
		emitted++
	}

	if emitted > 0 {
		dctx.Logger.Info("unfulfilled_imagery_plan: emitted work items",
			zap.Int("count", emitted),
			zap.Int("plan_rows", len(imagery)))
	}
	return result, nil
}

// loadCurrentPlanImagery returns all site_plan_imagery rows attached to the
// current site_plans row for this site, ordered by priority-relevant scope
// (site → page → section), then scope_ref, then ordering. Returns an empty
// slice if no current plan exists or the plan has no imagery rows.
func loadCurrentPlanImagery(dctx DiscoveryCheckContext) ([]plannedImagery, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
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
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("query site_plan_imagery: %w", err)
	}
	defer rows.Close()

	var out []plannedImagery
	for rows.Next() {
		var r plannedImagery
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
