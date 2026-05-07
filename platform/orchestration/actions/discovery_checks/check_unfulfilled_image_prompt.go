// FILE: platform/orchestration/actions/discovery_checks/check_unfulfilled_image_prompt.go
//
// Discovery check: site_specs.site_plan asks for an image but no asset exists
// for the corresponding purpose.
//
// Three categories of prompt key are handled:
//
//   1. "logo"            → routes to needs_logo, handler image-build-handler.
//   2. "hero_home"       → routes to needs_hero_image, handler image-build-handler.
//   3. "hero_<page>"     → emits flag-only unfulfilled_hero_variant. NO handler
//                          today because image-build-handler's deploy step is
//                          hardcoded to assets/images/hero.jpg and the
//                          assets.UNIQUE(site_id,purpose) constraint blocks
//                          having both a hero_home and a hero_about asset for
//                          the same site. When PLAN Phase 2 lands (asset_key,
//                          relaxed unique constraint, parameterised deploy
//                          paths), these items become routable. Until then
//                          we record the gap so it's visible in audits.
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

// classifyPromptKey returns (purpose, itemType, handlerAgent, priority, severity).
// handlerAgent == "" means the item is flag-only — it gets recorded but the
// dispatch loop will not pick it up. Returns ok=false for unrecognised keys.
func classifyPromptKey(promptKey string) (purpose, itemType, handlerAgent string, priority int, severity string, ok bool) {
	switch promptKey {
	case "logo":
		return "logo", "needs_logo", "image-build-handler", 70, "high", true
	case "hero_home":
		return "hero", "needs_hero_image", "image-build-handler", 65, "high", true
	}
	if strings.HasPrefix(promptKey, "hero_") {
		// Page-named hero variant. Purpose is the prompt key itself so the
		// future Phase-2 routing can disambiguate without re-deriving from
		// the spec. Severity medium because the page renders fine without
		// the variant (existing components fall back to hero_url or skip).
		return promptKey, "unfulfilled_hero_variant", "", 80, "medium", true
	}
	return "", "", "", 0, "", false
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

		purpose, itemType, handlerAgent, priority, severity, ok := classifyPromptKey(promptKey)
		if !ok {
			continue // unrecognised key shape; ignore conservatively
		}

		hasAsset, err := hasActiveAssetForPurpose(dctx, purpose)
		if err != nil {
			dctx.Logger.Warn("unfulfilled_image_prompt: asset existence check failed",
				zap.String("purpose", purpose),
				zap.Error(err))
			continue
		}
		if hasAsset {
			continue // asset already exists; nothing to do
		}

		// Build the work item spec. For routable items (logo, hero_home) the
		// handler reads input_data.spec.image_prompts.{key} so we preserve
		// the planner's keying. For flag-only variants we still include the
		// prompt so any future tooling can pick it up without re-querying
		// site_specs.
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":      "unfulfilled_image_prompt",
			"purpose":    purpose,
			"prompt_key": promptKey,
			"image_prompts": map[string]interface{}{
				promptKey: prompt,
			},
		})

		summary := fmt.Sprintf("Planner asked for %s but no asset exists", purpose)
		if handlerAgent == "" {
			summary = fmt.Sprintf("Planner asked for %s — recorded, deferred until multi-image asset path lands", purpose)
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":         "unfulfilled_image_prompt",
			"purpose":       purpose,
			"prompt_key":    promptKey,
			"prompt_length": len(prompt),
			"flag_only":     handlerAgent == "",
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
			HandlerAgent: handlerAgent, // empty for hero variants
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("unfulfilled_image_prompt:%s", purpose),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}
