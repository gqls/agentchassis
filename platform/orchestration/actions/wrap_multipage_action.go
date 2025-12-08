// FILE: platform/orchestration/actions/wrap_multipage_action.go
// Smart multipage wrapper with content-aware HTML search

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// WrapMultipageAction wraps a single-page site into a multi-page structure
func WrapMultipageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing WrapMultipageAction",
		zap.Any("config_keys", getConfigKeys(params.StepConfig.Config)),
	)

	// Get the configured path hint (optional)
	indexHTMLField, _ := params.StepConfig.Config["index_html_field"].(string)

	// Smart HTML search with fallback
	indexHTML := datahelpers.FindHTMLWithFallback(params.CollectedData, indexHTMLField, params.Logger)

	if indexHTML == "" {
		// Log what we have for debugging
		params.Logger.Error("Failed to find HTML content",
			zap.String("config_path", indexHTMLField),
			zap.Any("collected_data_keys", getCollectedDataKeys(params.CollectedData)),
		)
		return nil, fmt.Errorf("failed to extract index HTML from %s", indexHTMLField)
	}

	params.Logger.Info("Found index HTML",
		zap.Int("html_length", len(indexHTML)),
	)

	// Extract business info for page generation
	domain := extractDomainFromCollectedData(params.CollectedData)
	businessInfo := extractBusinessInfoMap(params.CollectedData)

	// Generate about page
	aboutHTML := generateAboutPage(domain, businessInfo)

	// Generate contact page
	contactHTML := generateContactPage(domain, businessInfo)

	// Create files map
	filesMap := map[string]interface{}{
		"index.html":   indexHTML,
		"about.html":   aboutHTML,
		"contact.html": contactHTML,
	}

	params.Logger.Info("Generated multipage site",
		zap.Int("page_count", len(filesMap)),
		zap.Int("index_size", len(indexHTML)),
		zap.Int("about_size", len(aboutHTML)),
		zap.Int("contact_size", len(contactHTML)),
	)

	return map[string]interface{}{
		"files":      filesMap,
		"page_count": len(filesMap),
		"pages":      []string{"index.html", "about.html", "contact.html"},
	}, nil
}

// generateAboutPage creates a simple about page
func generateAboutPage(domain string, businessInfo map[string]interface{}) string {
	objective := ""
	if obj, ok := businessInfo["objective"].(string); ok {
		objective = obj
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>About - %s</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6; 
            padding: 40px 20px; 
            max-width: 800px; 
            margin: 0 auto;
        }
        nav { margin-bottom: 40px; }
        nav a { margin-right: 20px; text-decoration: none; color: #0066cc; }
        nav a:hover { text-decoration: underline; }
        h1 { margin-bottom: 20px; color: #333; }
        p { margin-bottom: 15px; color: #666; }
    </style>
</head>
<body>
    <nav>
        <a href="index.html">Home</a>
        <a href="about.html">About</a>
        <a href="contact.html">Contact</a>
    </nav>
    <h1>About %s</h1>
    <p>%s</p>
    <p>We're dedicated to providing quality solutions for our customers.</p>
</body>
</html>`, domain, domain, ifEmpty(objective, "Learn more about our company and what we do."))
}

// generateContactPage creates a simple contact page
func generateContactPage(domain string, businessInfo map[string]interface{}) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Contact - %s</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6; 
            padding: 40px 20px; 
            max-width: 800px; 
            margin: 0 auto;
        }
        nav { margin-bottom: 40px; }
        nav a { margin-right: 20px; text-decoration: none; color: #0066cc; }
        nav a:hover { text-decoration: underline; }
        h1 { margin-bottom: 20px; color: #333; }
        form { margin-top: 30px; }
        label { display: block; margin-bottom: 5px; color: #333; font-weight: 500; }
        input, textarea { 
            width: 100%%; 
            padding: 10px; 
            margin-bottom: 20px; 
            border: 1px solid #ddd;
            border-radius: 4px;
            font-family: inherit;
        }
        button { 
            background: #0066cc; 
            color: white; 
            padding: 12px 30px; 
            border: none; 
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        button:hover { background: #0052a3; }
    </style>
</head>
<body>
    <nav>
        <a href="index.html">Home</a>
        <a href="about.html">About</a>
        <a href="contact.html">Contact</a>
    </nav>
    <h1>Contact Us</h1>
    <p>Get in touch with us for more information.</p>
    <form>
        <label for="name">Name</label>
        <input type="text" id="name" name="name" required>
        
        <label for="email">Email</label>
        <input type="email" id="email" name="email" required>
        
        <label for="message">Message</label>
        <textarea id="message" name="message" rows="5" required></textarea>
        
        <button type="submit">Send Message</button>
    </form>
</body>
</html>`, domain)
}

// Helper functions

// extractDomainFromCollectedData searches for domain in collected data
func extractDomainFromCollectedData(collectedData map[string]interface{}) string {
	// Try input_data first
	if inputData, ok := collectedData["input_data"]; ok {
		if inputMap, ok := inputData.(map[string]interface{}); ok {
			if domain, ok := inputMap["domain"].(string); ok && domain != "" {
				return domain
			}
		}
	}

	// Search recursively
	domain := findStringInMap(collectedData, "domain", 0)
	if domain != "" {
		return domain
	}

	return "Our Company"
}

// extractBusinessInfoMap extracts business info from collected data
func extractBusinessInfoMap(collectedData map[string]interface{}) map[string]interface{} {
	businessInfo := make(map[string]interface{})

	// Try to find domain
	domain := extractDomainFromCollectedData(collectedData)
	if domain != "" {
		businessInfo["domain"] = domain
	}

	// Try to find objective
	if inputData, ok := collectedData["input_data"]; ok {
		if inputMap, ok := inputData.(map[string]interface{}); ok {
			if objective, ok := inputMap["objective"].(string); ok && objective != "" {
				businessInfo["objective"] = objective
			}
		}
	}

	// Fallback: search recursively
	if businessInfo["objective"] == nil {
		objective := findStringInMap(collectedData, "objective", 0)
		if objective != "" {
			businessInfo["objective"] = objective
		}
	}

	return businessInfo
}

// findStringInMap recursively searches for a string field
func findStringInMap(data interface{}, key string, depth int) string {
	if depth > 10 {
		return ""
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return ""
	}

	// Direct match
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}

	// Recurse into values
	for _, val := range m {
		if result := findStringInMap(val, key, depth+1); result != "" {
			return result
		}
	}

	return ""
}

// extractDomain extracts domain from business info map

func getConfigKeys(config map[string]interface{}) []string {
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	return keys
}

func getCollectedDataKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		// Skip internal fields
		if !strings.HasPrefix(k, "__") {
			keys = append(keys, k)
		}
	}
	return keys
}
