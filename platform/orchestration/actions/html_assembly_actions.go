// FILE: platform/orchestration/actions/html_assembly_actions.go
package actions

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// AssembleHTMLPartsAction assembles separate HTML components into a complete document
// Takes structure, styles, and content and combines them intelligently
func AssembleHTMLPartsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Assembling HTML from parts")

	// Get configuration - Config is already map[string]interface{}
	config := params.StepConfig.Config
	if config == nil {
		return nil, fmt.Errorf("config is required for assemble_html_parts action")
	}

	// Extract field names from config
	structureField, _ := config["structure_field"].(string)
	stylesField, _ := config["styles_field"].(string)
	contentField, _ := config["content_field"].(string)

	if structureField == "" || stylesField == "" || contentField == "" {
		return nil, fmt.Errorf("assemble_html_parts requires structure_field, styles_field, and content_field in config")
	}

	// Extract parts from collected data
	structure := extractFieldValue(params.CollectedData, structureField, params.Logger)
	styles := extractFieldValue(params.CollectedData, stylesField, params.Logger)
	content := extractFieldValue(params.CollectedData, contentField, params.Logger)

	// Validate we have all parts
	if structure == "" {
		return nil, fmt.Errorf("structure HTML is empty (field: %s)", structureField)
	}
	if styles == "" {
		params.Logger.Warn("No styles provided, proceeding without CSS")
	}
	if content == "" {
		return nil, fmt.Errorf("content HTML is empty (field: %s)", contentField)
	}

	// Log the parts we're working with
	params.Logger.Info("HTML parts extracted",
		zap.Int("structure_length", len(structure)),
		zap.Int("styles_length", len(styles)),
		zap.Int("content_length", len(content)),
	)

	// Assemble the complete HTML
	completeHTML, err := assembleHTMLDocument(structure, styles, content, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble HTML: %w", err)
	}

	params.Logger.Info("HTML assembled successfully",
		zap.Int("final_length", len(completeHTML)),
	)

	return map[string]interface{}{
		"html":           completeHTML,
		"assembled_at":   params.ExecutionContext.Timestamp,
		"parts_combined": 3,
	}, nil
}

// extractFieldValue navigates nested field paths like "base_structure.result"
func extractFieldValue(data map[string]interface{}, fieldPath string, logger *zap.Logger) string {
	parts := strings.Split(fieldPath, ".")

	var current interface{} = data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			// First try direct access
			if val, ok := v[part]; ok {
				current = val
				continue
			}
			// Then try ExtractStepData if it looks like a step result
			if extracted := datahelpers.ExtractStepData(v[part]); extracted != nil {
				current = extracted
				continue
			}
			logger.Warn("Field not found in path",
				zap.String("field", part),
				zap.String("full_path", fieldPath),
			)
			return ""
		default:
			// If we're at a terminal value and still have more parts, something's wrong
			if len(parts) > 1 {
				logger.Warn("Cannot traverse further, value is not a map",
					zap.String("field", part),
					zap.String("full_path", fieldPath),
				)
				return ""
			}
		}
	}

	// Convert final value to string
	switch v := current.(type) {
	case string:
		return v
	case map[string]interface{}:
		// If it's still a map, try to get "result" or "html" or "content"
		if result, ok := v["result"].(string); ok {
			return result
		}
		if html, ok := v["html"].(string); ok {
			return html
		}
		if content, ok := v["content"].(string); ok {
			return content
		}
		logger.Warn("Final value is a map but couldn't extract string",
			zap.String("full_path", fieldPath),
		)
		return ""
	default:
		logger.Warn("Final value is not a string",
			zap.String("full_path", fieldPath),
			zap.String("type", fmt.Sprintf("%T", current)),
		)
		return ""
	}
}

// assembleHTMLDocument combines the three parts into a complete HTML document
func assembleHTMLDocument(structure, styles, content string, logger *zap.Logger) (string, error) {
	// Clean up inputs - remove markdown code blocks if present
	structure = datahelpers.CleanHTMLString(structure)
	styles = datahelpers.CleanHTMLString(styles)
	content = datahelpers.CleanHTMLString(content)

	// Parse the structure to find where to insert styles and content
	doc := structure

	// 1. Insert styles into <head> before </head>
	if styles != "" {
		// Ensure styles are wrapped in <style> tags if not already
		if !strings.Contains(styles, "<style") {
			styles = "<style>\n" + styles + "\n</style>"
		}

		// Find </head> and insert before it
		headCloseIdx := strings.Index(doc, "</head>")
		if headCloseIdx >= 0 {
			doc = doc[:headCloseIdx] + "\n" + styles + "\n" + doc[headCloseIdx:]
			logger.Info("Inserted styles into <head>")
		} else {
			logger.Warn("No </head> tag found, skipping style insertion")
		}
	}

	// 2. Insert content into <body>
	// Look for empty body or body with just comments
	bodyStartIdx := strings.Index(doc, "<body")
	if bodyStartIdx < 0 {
		return "", fmt.Errorf("no <body> tag found in structure")
	}

	// Find the end of the <body> opening tag
	bodyOpenEndIdx := strings.Index(doc[bodyStartIdx:], ">")
	if bodyOpenEndIdx < 0 {
		return "", fmt.Errorf("malformed <body> tag")
	}
	bodyContentStart := bodyStartIdx + bodyOpenEndIdx + 1

	// Find </body>
	bodyCloseIdx := strings.Index(doc, "</body>")
	if bodyCloseIdx < 0 {
		return "", fmt.Errorf("no </body> tag found in structure")
	}

	// Get current body content
	currentBodyContent := doc[bodyContentStart:bodyCloseIdx]

	// If body is empty or just whitespace/comments, replace it entirely
	trimmedBody := strings.TrimSpace(currentBodyContent)
	if trimmedBody == "" || isJustComments(trimmedBody) {
		// Replace entire body content
		doc = doc[:bodyContentStart] + "\n" + content + "\n" + doc[bodyCloseIdx:]
		logger.Info("Replaced empty body with content")
	} else {
		// Body has some content, try to find a placeholder or append
		// Look for common placeholders
		placeholders := []string{
			"<!-- CONTENT_HERE -->",
			"<!-- content -->",
			"<main></main>",
			"<main>",
		}

		replaced := false
		for _, placeholder := range placeholders {
			if strings.Contains(currentBodyContent, placeholder) {
				newBodyContent := strings.Replace(currentBodyContent, placeholder, content, 1)
				doc = doc[:bodyContentStart] + newBodyContent + doc[bodyCloseIdx:]
				logger.Info("Replaced placeholder with content",
					zap.String("placeholder", placeholder),
				)
				replaced = true
				break
			}
		}

		if !replaced {
			// Just insert content before </body>
			doc = doc[:bodyCloseIdx] + "\n" + content + "\n" + doc[bodyCloseIdx:]
			logger.Info("Appended content before </body>")
		}
	}

	// Ensure DOCTYPE is present
	if !strings.HasPrefix(strings.TrimSpace(doc), "<!DOCTYPE") {
		doc = "<!DOCTYPE html>\n" + doc
	}

	return doc, nil
}

// isJustComments checks if a string contains only HTML comments and whitespace
func isJustComments(s string) bool {
	// Remove all HTML comments
	commentRe := regexp.MustCompile("<!--[\\s\\S]*?-->")
	withoutComments := commentRe.ReplaceAllString(s, "")

	// Check if anything substantial remains
	return strings.TrimSpace(withoutComments) == ""
}
