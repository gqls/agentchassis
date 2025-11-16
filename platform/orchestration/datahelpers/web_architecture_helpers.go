package datahelpers

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ValidateWorkflowTransition validates data before passing to next agent
func ValidateWorkflowTransition(
	fromStep string,
	toStep string,
	collectedData map[string]interface{},
	logger *zap.Logger,
) error {

	transitions := map[string][]string{
		"chief-strategist->site-architect": {"build_plan_data.build_plan_json"},
		"site-architect->content-creator": {"template_data.stitched_html_template",
			"template_data.content_requirements"},
		"content-creator->deployer": {"final_site_data.final_html"},
	}

	key := fmt.Sprintf("%s->%s", fromStep, toStep)
	requiredFields, ok := transitions[key]
	if !ok {
		return nil // No validation defined
	}

	for _, field := range requiredFields {
		if _, err := GetFieldFromPath(collectedData, field, logger); err != nil {
			return fmt.Errorf("missing required field %s for transition %s", field, key)
		}
	}

	return nil
}

func recordMetric(action string, agentType string, duration time.Duration) {
	// Record to your metrics system
}

/*func BuildWebsiteCreationContext(
	collectedData map[string]interface{},
	logger *zap.Logger,
) *WebsiteCreationContext {
	return &WebsiteCreationContext{
		Domain: GetFieldFromPathWithDefault(collectedData, "input_data.domain", "", logger).(string),
		BuildPlan: GetFieldFromPath(collectedData, "build_plan_data.build_plan_json", logger),
		HTMLTemplate: GetFieldFromPath(collectedData, "template_data.stitched_html_template", logger),
		ContentRequirements: GetFieldFromPath(collectedData, "template_data.content_requirements", logger),
		FinalHTML: GetFieldFromPath(collectedData, "final_site_data.final_html", logger),
	}
}*/
