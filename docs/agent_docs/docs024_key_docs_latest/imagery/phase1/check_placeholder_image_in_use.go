// FILE: platform/orchestration/actions/discovery_checks/check_placeholder_image_in_use.go
//
// Discovery check: rendered HTML contains the fallback image path
// (e.g. /assets/images/hero.jpg) AND no assets row exists for the matching
// purpose. The component's input_schema fallback resolved because the asset
// was never generated; the page is silently shipping the system default.
//
// This is the symptom you see when image-build-handler never ran for the
// site, or when its run completed in error. The unfulfilled_image_prompt
// check catches the upstream cause (planner asked, no asset); this check
// catches the downstream symptom (no asset, fallback rendered) — the two
// can fire together but for different reasons:
//   - unfulfilled_image_prompt: planner provided a prompt
//   - placeholder_image_in_use: a page references the fallback
// They produce identical work items in practice; deduping is by ItemKey.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&PlaceholderImageInUseCheck{}) }

type PlaceholderImageInUseCheck struct{}

func (c *PlaceholderImageInUseCheck) Name() string { return "placeholder_image_in_use" }

// placeholderPathMapping maps the documented fallback path to the assets.purpose
// value that, if missing, indicates we're rendering the placeholder for real.
// Paths come from 003_contracts_and_standards_v7.md.
var placeholderPathMapping = []struct {
	path     string // the fallback path that appears in rendered_html
	purpose  string // assets.purpose to check for
	itemType string // work item type for image-build-handler
	priority int
}{
	{"/assets/images/hero.jpg", "hero", "needs_hero_image", 65},
	{"/assets/images/logo.png", "logo", "needs_logo", 70},
}

func (c *PlaceholderImageInUseCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	for _, mapping := range placeholderPathMapping {
		referenced, err := isPathReferencedInPages(dctx, mapping.path)
		if err != nil {
			dctx.Logger.Warn("placeholder_image_in_use: rendered_html scan failed",
				zap.String("path", mapping.path),
				zap.Error(err))
			continue
		}
		if !referenced {
			continue
		}

		hasAsset, err := hasActiveAssetForPurpose(dctx, mapping.purpose)
		if err != nil {
			dctx.Logger.Warn("placeholder_image_in_use: asset existence check failed",
				zap.String("purpose", mapping.purpose),
				zap.Error(err))
			continue
		}
		if hasAsset {
			continue // asset exists; this isn't a placeholder use, the deployed
			// path just happens to match the canonical name for that purpose.
		}

		// Try to recover the original prompt from the site_plan so the
		// regenerate has the planner's intent. If the site_plan spec doesn't
		// have the prompt either, the handler will fall back to its default
		// prompt template — still useful, just less specific.
		prompts, _ := loadImagePromptsForSite(dctx)
		var prompt string
		switch mapping.purpose {
		case "logo":
			prompt = prompts["logo"]
		case "hero":
			prompt = prompts["hero_home"]
		}

		spec := map[string]interface{}{
			"check":   "placeholder_image_in_use",
			"purpose": mapping.purpose,
			"path":    mapping.path,
		}
		if prompt != "" {
			promptKey := mapping.purpose
			if mapping.purpose == "hero" {
				promptKey = "hero_home"
			}
			spec["image_prompts"] = map[string]interface{}{promptKey: prompt}
			spec["prompt_key"] = promptKey
		}
		specJSON, _ := json.Marshal(spec)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":        "placeholder_image_in_use",
			"purpose":      mapping.purpose,
			"path":         mapping.path,
			"prompt_known": prompt != "",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     dctx.Pipeline,
			ItemType:     mapping.itemType,
			Severity:     "high",
			Summary:      fmt.Sprintf("Pages reference fallback %s but no asset exists", mapping.path),
			SpecJSON:     string(specJSON),
			Priority:     mapping.priority,
			HandlerAgent: "image-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("placeholder_image_in_use:%s", mapping.purpose),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

// isPathReferencedInPages returns true if any deployed component for this
// site contains the given fallback path in its rendered HTML. Locked
// components are skipped: a human-locked component referencing the fallback
// is presumably intentional.
func isPathReferencedInPages(dctx DiscoveryCheckContext, path string) (bool, error) {
	var n int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html LIKE '%' || $2 || '%'
		LIMIT 1
	`, dctx.SiteID, path).Scan(&n)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return n > 0, nil
}
