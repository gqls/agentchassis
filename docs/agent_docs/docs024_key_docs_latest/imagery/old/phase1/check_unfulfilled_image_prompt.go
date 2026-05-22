// FILE: platform/orchestration/actions/discovery_checks/check_unfulfilled_image_prompt.go
//
// Discovery check: site_specs.site_plan asks for an image (logo or hero_home)
// but no assets row exists for that purpose.
//
// This catches the case where the planner produced image_prompts but the
// build pipeline never generated the asset — e.g. generation failed silently,
// the work item was never dispatched, or a rebuild dropped the asset.
//
// Routes to image-build-handler with the existing needs_logo / needs_hero_image
// item types. The handler's workflow already reads input_data.spec.image_prompts
// for the prompt to use, so we pass the planner's prompt through unchanged.

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&UnfulfilledImagePromptCheck{}) }

type UnfulfilledImagePromptCheck struct{}

func (c *UnfulfilledImagePromptCheck) Name() string { return "unfulfilled_image_prompt" }

// purposeMapping maps the image_prompts JSON key to the assets.purpose value
// and the work item type that the image-build-handler dispatches on.
var imagePromptPurposeMapping = []struct {
	promptKey string // key in site_specs.data.image_prompts
	purpose   string // assets.purpose value
	itemType  string // site_work_items.item_type
	priority  int
}{
	{"logo", "logo", "needs_logo", 70},
	{"hero_home", "hero", "needs_hero_image", 65},
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

	for _, mapping := range imagePromptPurposeMapping {
		prompt, ok := prompts[mapping.promptKey]
		if !ok || prompt == "" {
			continue // planner did not request this image
		}

		hasAsset, err := hasActiveAssetForPurpose(dctx, mapping.purpose)
		if err != nil {
			dctx.Logger.Warn("unfulfilled_image_prompt: asset existence check failed",
				zap.String("purpose", mapping.purpose),
				zap.Error(err))
			continue
		}
		if hasAsset {
			continue // asset already exists; nothing to do
		}

		// Build the work item spec the way image-build-handler expects.
		// The handler reads input_data.spec.image_prompts.{key} so we need to
		// preserve the same shape the planner uses.
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":      "unfulfilled_image_prompt",
			"purpose":    mapping.purpose,
			"prompt_key": mapping.promptKey,
			"image_prompts": map[string]interface{}{
				mapping.promptKey: prompt,
			},
		})

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":         "unfulfilled_image_prompt",
			"purpose":       mapping.purpose,
			"prompt_key":    mapping.promptKey,
			"prompt_length": len(prompt),
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     dctx.Pipeline,
			ItemType:     mapping.itemType,
			Severity:     "high",
			Summary:      fmt.Sprintf("Planner asked for %s but no asset exists", mapping.purpose),
			SpecJSON:     string(specJSON),
			Priority:     mapping.priority,
			HandlerAgent: "image-build-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("unfulfilled_image_prompt:%s", mapping.purpose),
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}
