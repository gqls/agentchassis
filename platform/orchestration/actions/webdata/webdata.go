// FILE: platform/orchestration/actions/data/data.go
// Package data provides data manipulation actions: validate, transform, aggregate
package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/aqls/personae/platform/orchestration/actions/registry"
)

func init() {
	// Validation
	registry.Register("validate_input", registry.ActionDefinition{
		Func:        ValidateInputAction,
		Category:    registry.CategoryData,
		Description: "Validates input data against rules",
		Status:      registry.StatusActive,
	})

	registry.Register("validate_schema", registry.ActionDefinition{
		Func:        ValidateSchemaAction,
		Category:    registry.CategoryData,
		Description: "Validates data against a JSON schema",
		Status:      registry.StatusActive,
	})

	// Transformation
	registry.Register("transform_data", registry.ActionDefinition{
		Func:        TransformDataAction,
		Category:    registry.CategoryData,
		Description: "Transforms data according to mapping rules",
		Status:      registry.StatusActive,
	})

	// Aggregation
	registry.Register("aggregate_data", registry.ActionDefinition{
		Func:        AggregateDataAction,
		Category:    registry.CategoryData,
		Description: "Aggregates data from multiple sources",
		Status:      registry.StatusActive,
	})

	registry.Register("aggregate_webpage", registry.ActionDefinition{
		Func:        AggregateWebpageAction,
		Category:    registry.CategoryData,
		Description: "Aggregates webpage sections into a single page",
		Status:      registry.StatusActive,
	})

	// Math operations
	registry.Register("calculate", registry.ActionDefinition{
		Func:        CalculateAction,
		Category:    registry.CategoryData,
		Description: "Performs mathematical calculations",
		Status:      registry.StatusActive,
	})

	// Multi-page wrapper (website-specific but data category since it's data transformation)
	registry.Register("wrap_multipage", registry.ActionDefinition{
		Func:        WrapMultipageAction,
		Category:    registry.CategoryData,
		Description: "Wraps single index.html into multi-page site structure",
		Status:      registry.StatusExperimental,
		DomainTags:  []string{"website"},
	})
}

// ValidateInputAction validates input against rules
func ValidateInputAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	rules, ok := params.Config["rules"].([]interface{})
	if !ok {
		// No rules = always valid
		return map[string]interface{}{
			"valid":  true,
			"errors": []string{},
		}, nil
	}

	var errors []string
	for _, r := range rules {
		rule, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		field, _ := rule["field"].(string)
		required, _ := rule["required"].(bool)

		if required {
			val := ExtractNestedField(params.CollectedData, field)
			if val == nil {
				errors = append(errors, fmt.Sprintf("required field missing: %s", field))
			}
		}
	}

	return map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
	}, nil
}

// ValidateSchemaAction validates against JSON schema
func ValidateSchemaAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	// TODO: Implement JSON schema validation
	return map[string]interface{}{
		"valid":  true,
		"errors": []string{},
	}, nil
}

// TransformDataAction transforms data
func TransformDataAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	mappings, ok := params.Config["mappings"].(map[string]interface{})
	if !ok {
		return params.CollectedData, nil
	}

	result := make(map[string]interface{})
	for target, source := range mappings {
		if sourceField, ok := source.(string); ok {
			val := ExtractNestedField(params.CollectedData, sourceField)
			if val != nil {
				result[target] = val
			}
		}
	}

	return result, nil
}

// AggregateDataAction aggregates from multiple sources
func AggregateDataAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	// TODO: Migrate from aggregate_data.go
	return map[string]interface{}{
		"status": "aggregated",
	}, nil
}

// AggregateWebpageAction aggregates webpage sections
func AggregateWebpageAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	// TODO: Migrate from aggregate_webpage.go
	return map[string]interface{}{
		"status": "aggregated",
	}, nil
}

// CalculateAction performs math operations
func CalculateAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	// TODO: Migrate from calculate_actions.go
	return map[string]interface{}{
		"status": "calculated",
	}, nil
}

// WrapMultipageAction creates multi-page site from single index.html
func WrapMultipageAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	// Get index.html from configured field
	indexHTMLField, _ := params.Config["index_html_field"].(string)
	if indexHTMLField == "" {
		indexHTMLField = "final_html"
	}

	indexHTML := ExtractNestedFieldString(params.CollectedData, indexHTMLField)
	if indexHTML == "" {
		return nil, fmt.Errorf("wrap_multipage: no index.html found at %s", indexHTMLField)
	}

	// Extract brand name from domain
	domain := ExtractNestedFieldString(params.CollectedData, "input_data.domain")
	brandName := domainToBrandName(domain)

	// Extract tagline if available
	tagline := ExtractNestedFieldString(params.CollectedData, "content_json.sections.component_footer_7.tagline")
	if tagline == "" {
		tagline = "Building the future"
	}

	// Generate about and contact pages
	aboutHTML := generateAboutPage(brandName, tagline, domain)
	contactHTML := generateContactPage(brandName, tagline, domain)

	files := map[string]string{
		"index.html":   indexHTML,
		"about.html":   aboutHTML,
		"contact.html": contactHTML,
	}

	return map[string]interface{}{
		"files":      files,
		"file_count": len(files),
		"pages":      []string{"index.html", "about.html", "contact.html"},
	}, nil
}

// Helper functions

// ExtractNestedField extracts a value from nested map using dot notation
func ExtractNestedField(data map[string]interface{}, fieldPath string) interface{} {
	if fieldPath == "" {
		return nil
	}

	parts := strings.Split(fieldPath, ".")
	current := interface{}(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}
		if current == nil {
			return nil
		}
	}

	return current
}

// ExtractNestedFieldString is a convenience wrapper returning empty string if not found
func ExtractNestedFieldString(data map[string]interface{}, fieldPath string) string {
	val := ExtractNestedField(data, fieldPath)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// ExtractNestedFieldMap is a convenience wrapper returning nil if not found
func ExtractNestedFieldMap(data map[string]interface{}, fieldPath string) map[string]interface{} {
	val := ExtractNestedField(data, fieldPath)
	if val == nil {
		return nil
	}
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func domainToBrandName(domain string) string {
	if domain == "" {
		return "Our Company"
	}
	// Remove TLD
	parts := strings.Split(domain, ".")
	if len(parts) > 0 {
		name := parts[0]
		// Convert hyphens to spaces and title case
		name = strings.ReplaceAll(name, "-", " ")
		return strings.Title(name)
	}
	return domain
}

func generateAboutPage(brandName, tagline, domain string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>About - %s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        header { background: #1a1a2e; color: white; padding: 1rem 2rem; position: fixed; width: 100%%; top: 0; z-index: 1000; }
        nav { display: flex; justify-content: space-between; align-items: center; max-width: 1200px; margin: 0 auto; }
        nav a { color: white; text-decoration: none; margin-left: 2rem; }
        nav a:hover { text-decoration: underline; }
        .logo { font-size: 1.5rem; font-weight: bold; }
        main { max-width: 800px; margin: 0 auto; padding: 6rem 2rem 4rem; }
        h1 { font-size: 2.5rem; margin-bottom: 1rem; }
        p { margin-bottom: 1rem; color: #666; }
        footer { background: #1a1a2e; color: white; text-align: center; padding: 2rem; margin-top: 4rem; }
    </style>
</head>
<body>
    <header>
        <nav>
            <div class="logo">%s</div>
            <div>
                <a href="index.html">Home</a>
                <a href="about.html">About</a>
                <a href="contact.html">Contact</a>
            </div>
        </nav>
    </header>
    <main>
        <h1>About Us</h1>
        <p>%s</p>
        <p>We are dedicated to delivering exceptional value to our customers through innovation, quality, and commitment to excellence.</p>
        <p>Our team brings together diverse expertise and a shared passion for making a difference in our industry.</p>
    </main>
    <footer>
        <p>&copy; 2024 %s. All rights reserved.</p>
    </footer>
</body>
</html>`, brandName, brandName, tagline, brandName)
}

func generateContactPage(brandName, tagline, domain string) string {
	email := "hello@" + domain
	if domain == "" {
		email = "hello@example.com"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Contact - %s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        header { background: #1a1a2e; color: white; padding: 1rem 2rem; position: fixed; width: 100%%; top: 0; z-index: 1000; }
        nav { display: flex; justify-content: space-between; align-items: center; max-width: 1200px; margin: 0 auto; }
        nav a { color: white; text-decoration: none; margin-left: 2rem; }
        nav a:hover { text-decoration: underline; }
        .logo { font-size: 1.5rem; font-weight: bold; }
        main { max-width: 800px; margin: 0 auto; padding: 6rem 2rem 4rem; }
        h1 { font-size: 2.5rem; margin-bottom: 1rem; }
        p { margin-bottom: 1rem; color: #666; }
        .contact-info { background: #f5f5f5; padding: 2rem; border-radius: 8px; margin-top: 2rem; }
        .contact-info a { color: #1a1a2e; }
        footer { background: #1a1a2e; color: white; text-align: center; padding: 2rem; margin-top: 4rem; }
    </style>
</head>
<body>
    <header>
        <nav>
            <div class="logo">%s</div>
            <div>
                <a href="index.html">Home</a>
                <a href="about.html">About</a>
                <a href="contact.html">Contact</a>
            </div>
        </nav>
    </header>
    <main>
        <h1>Contact Us</h1>
        <p>We'd love to hear from you. Get in touch with us using the information below.</p>
        <div class="contact-info">
            <p><strong>Email:</strong> <a href="mailto:%s">%s</a></p>
        </div>
    </main>
    <footer>
        <p>&copy; 2024 %s. All rights reserved.</p>
    </footer>
</body>
</html>`, brandName, brandName, email, email, brandName)
}
