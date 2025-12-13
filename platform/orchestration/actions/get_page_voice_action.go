package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func GetPageVoiceAction(ctx context.Context, params ActionParams) (interface{}, error) {
	pageName := extractStringField(params.CollectedData,
		params.Config["page_field"].(string), params.Logger)

	// Simple logic: home = casual, others = professional
	if pageName == "index" || pageName == "home" {
		return map[string]interface{}{
			"formality":       0.5,
			"technical_depth": 0.3,
		}, nil
	}

	return map[string]interface{}{
		"formality":       0.7,
		"technical_depth": 0.5,
	}, nil
}
