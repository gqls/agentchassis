// FILE: platform/orchestration/actions/check_tool_completeness_action.go
//
// Checks that LLM-generated tool HTML is complete (not truncated).
// Looks for the completion marker <!-- tool-recreation-complete -->
// and verifies balanced script/style tags.
//
// Registration (add to registry.go):
//   "check_tool_completeness": {
//       Handler:     CheckToolCompletenessAction,
//       Category:    "validation",
//       Description: "Verify LLM tool output is complete and not truncated",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CheckToolCompletenessInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"html_field"},
	Defaults:   map[string]interface{}{"html_field": "tool_recreation.result"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("check_tool_completeness", CheckToolCompletenessInputSpec)
}

const toolCompleteMarker = "<!-- tool-recreation-complete -->"

func CheckToolCompletenessAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "check_tool_completeness"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	htmlField := "tool_recreation.result"
	if f, ok := config["html_field"].(string); ok && f != "" {
		htmlField = f
	}

	html := datahelpers.ExtractNestedFieldString(params.CollectedData, htmlField)
	if html == "" {
		return nil, fmt.Errorf("no HTML found at %s — tool recreation produced empty output", htmlField)
	}

	issues := []string{}

	// Check for completion marker
	hasMarker := strings.Contains(html, toolCompleteMarker)
	if !hasMarker {
		issues = append(issues, "missing completion marker — output likely truncated")
	}

	// Check for balanced script tags
	scriptOpens := strings.Count(html, "<script")
	scriptCloses := strings.Count(html, "</script>")
	if scriptOpens > 0 && scriptOpens != scriptCloses {
		issues = append(issues, fmt.Sprintf("unbalanced script tags: %d opens, %d closes", scriptOpens, scriptCloses))
	}

	// Check for balanced style tags
	styleOpens := strings.Count(html, "<style")
	styleCloses := strings.Count(html, "</style>")
	if styleOpens > 0 && styleOpens != styleCloses {
		issues = append(issues, fmt.Sprintf("unbalanced style tags: %d opens, %d closes", styleOpens, styleCloses))
	}

	// Check the HTML has substance (not just a wrapper div)
	if len(html) < 500 {
		issues = append(issues, fmt.Sprintf("suspiciously short output: %d chars", len(html)))
	}

	// Strip the completion marker from the HTML before passing downstream
	cleanHTML := strings.Replace(html, toolCompleteMarker, "", 1)
	cleanHTML = strings.TrimSpace(cleanHTML)

	isComplete := len(issues) == 0

	logger.Info("CheckToolCompleteness",
		zap.Bool("complete", isComplete),
		zap.Int("html_length", len(html)),
		zap.Bool("has_marker", hasMarker),
		zap.Int("script_tags", scriptOpens),
		zap.Int("issues", len(issues)),
	)

	if !isComplete {
		logger.Warn("Tool recreation output has completeness issues",
			zap.Strings("issues", issues))
		// Don't fail — let it through but flag it. The tool might still be
		// partially functional and worth deploying for review.
	}

	return map[string]interface{}{
		"complete":   isComplete,
		"clean_html": cleanHTML,
		"html_size":  len(cleanHTML),
		"has_marker": hasMarker,
		"issues":     issues,
	}, nil
}
