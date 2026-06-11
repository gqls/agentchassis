// FILE: platform/orchestration/actions/discovery_checks/check_unfulfilled_image_prompt.go
//
// Discovery check: site_specs.site_plan asks for an image but no asset exists
// for the corresponding asset_key.
//
// Three categories of prompt key are handled:
//
//   1. "logo"            → routes to needs_logo, handler image-build-handler.
//                          purpose=logo, asset_key=logo.
//   2. "hero_home"       → routes to needs_hero_image, handler image-build-handler.
//                          purpose=hero, asset_key=hero (canonical hero).
//   3. "hero_<page>"     → Phase 2E: routes to unfulfilled_hero_variant with
//                          handler image-build-handler. purpose=hero,
//                          asset_key=<the prompt key, e.g. "hero_about">.
//                          Multi-image enabled by Phase 2A-D (asset_key column,
//                          relaxed unique constraint, store_asset writes
//                          asset_key) plus the asset_key-aware deploy_path
//                          derivation in deploy_image_asset_action.go.
//
// Other prompt keys: ignored. The check is conservative — unknown keys produce
// no work items, no findings.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func init() { Register(&UnfulfilledImagePromptCheck{}) }

type UnfulfilledImagePromptCheck struct{}

func (c *UnfulfilledImagePromptCheck) Name() string { return "unfulfilled_image_prompt" }

// classifyPromptKey returns (purpose, assetKey, itemType, handlerAgent, priority, severity).
// Returns ok=false for unrecognised keys.
//
// Phase 2E: variants now route through image-build-handler. (purpose, asset_key)
// is the post-Phase-2D natural key. Canonical logo and hero have asset_key=purpose
// for backward compat; variants have asset_key=<prompt_key> and purpose="hero".
func classifyPromptKey(promptKey string) (purpose, assetKey, itemType, handlerAgent string, priority int, severity string, ok bool) {
	switch promptKey {
	case "logo":
		return "logo", "logo", "needs_logo", "image-build-handler", 70, "high", true
	case "hero_home":
		return "hero", "hero", "needs_hero_image", "image-build-handler", 65, "high", true
	}
	if strings.HasPrefix(promptKey, "hero_") {
		// Page-named hero variant. asset_key is the prompt key itself
		// (e.g. "hero_about"), purpose is the conceptual category "hero".
		// This split lets image-build-handler render to a per-variant
		// deploy path (assets/images/hero-about.jpg) while keeping
		// purpose-based metadata (resize dimensions, audit grouping)
		// consistent across all hero variants.
		return "hero", promptKey, "unfulfilled_hero_variant", "image-build-handler", 80, "medium", true
	}
	return "", "", "", "", 0, "", false
}

func (c *UnfulfilledImagePromptCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	prompts, err := loadImagePromptsForSite(dctx)
	if err != nil {
		return nil, err
	}
	if len(prompts) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{}

	for promptKey, prompt := range prompts {
		if prompt == "" {
			continue // planner left the slot empty
		}

		purpose, assetKey, itemType, handlerAgent, priority, severity, ok := classifyPromptKey(promptKey)
		if !ok {
			continue // unrecognised key shape; ignore conservatively
		}

		hasAsset, err := hasActiveAssetForAssetKey(dctx, assetKey)
		if err != nil {
			dctx.Logger.Warn("unfulfilled_image_prompt: asset existence check failed",
				zap.String("asset_key", assetKey),
				zap.Error(err))
			continue
		}
		if hasAsset {
			continue // asset already exists; nothing to do
		}

		// Build the work item spec. The handler reads input_data.spec.prompt
		// directly (avoids workflow-engine dynamic key lookup) plus
		// input_data.spec.purpose and input_data.spec.asset_key for the
		// store_asset and deploy_image_asset config.
		//
		// image_prompts is included for backward compat / tracing only — the
		// existing pre-Phase-2E workflow steps for canonical logo/hero still
		// reference image_prompts.{key}. The Phase 2E workflow change
		// switches to reading spec.prompt; this dual format means the
		// migration order doesn't matter.
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":      "unfulfilled_image_prompt",
			"purpose":    purpose,
			"asset_key":  assetKey,
			"prompt_key": promptKey,
			"prompt":     prompt,
			"image_prompts": map[string]interface{}{
				promptKey: prompt,
			},
		})

		summary := fmt.Sprintf("Planner asked for %s (key=%s) but no asset exists", purpose, assetKey)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":         "unfulfilled_image_prompt",
			"purpose":       purpose,
			"asset_key":     assetKey,
			"prompt_key":    promptKey,
			"prompt_length": len(prompt),
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     dctx.Pipeline,
			ItemType:     itemType,
			Severity:     severity,
			Summary:      summary,
			SpecJSON:     string(specJSON),
			Priority:     priority,
			HandlerAgent: handlerAgent,
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			// item_key keyed on asset_key (post-Phase-2D natural key) so
			// hero variants get distinct keys and don't dedup-collide.
			ItemKey: fmt.Sprintf("unfulfilled_image_prompt:%s", assetKey),
			BatchID: dctx.BatchID,
		})
	}

	return result, nil
}
