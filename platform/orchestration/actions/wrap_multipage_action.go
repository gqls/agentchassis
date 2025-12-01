// wrap_multipage_action.go
// Wraps a single-page site into a multi-page structure with about and contact pages
package actions

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// WrapMultipageAction takes the assembled index.html and creates about/contact pages
// Returns a files map ready for the git deployer
func WrapMultipageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	logger.Info("Executing WrapMultipageAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.Any("config_keys", getMapKeys(config)),
	)

	// Extract the main HTML content
	indexHTMLField, _ := config["index_html_field"].(string)
	if indexHTMLField == "" {
		indexHTMLField = "input_data.final_html.assemble_html.final_html"
	}

	indexHTML := datahelpers.ExtractNestedField(params.CollectedData, indexHTMLField)
	if indexHTML == nil {
		return nil, fmt.Errorf("failed to extract index HTML from %s", indexHTMLField)
	}

	indexHTMLStr, ok := indexHTML.(string)
	if !ok {
		return nil, fmt.Errorf("index HTML is not a string")
	}

	// Extract brand info for about/contact pages
	brandName := extractBrandNameForMultipage(params.CollectedData, config)
	domain := extractDomainForMultipage(params.CollectedData, config)
	tagline := extractTaglineForMultipage(params.CollectedData, config)

	logger.Info("Building multi-page site",
		zap.String("brand_name", brandName),
		zap.String("domain", domain),
		zap.Int("index_html_length", len(indexHTMLStr)),
	)

	// Generate about.html
	aboutHTML, err := generateAboutPage(brandName, domain, tagline)
	if err != nil {
		return nil, fmt.Errorf("failed to generate about page: %w", err)
	}

	// Generate contact.html
	contactHTML, err := generateContactPage(brandName, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to generate contact page: %w", err)
	}

	// Build the files map
	files := map[string]string{
		"index.html":   indexHTMLStr,
		"about.html":   aboutHTML,
		"contact.html": contactHTML,
	}

	logger.Info("Multi-page site assembled",
		zap.Int("file_count", len(files)),
		zap.Int("index_size", len(indexHTMLStr)),
		zap.Int("about_size", len(aboutHTML)),
		zap.Int("contact_size", len(contactHTML)),
	)

	return map[string]interface{}{
		"files":      files,
		"file_count": len(files),
		"pages":      []string{"index.html", "about.html", "contact.html"},
	}, nil
}

// extractBrandNameForMultipage tries to find brand name from various places in collected data
func extractBrandNameForMultipage(data map[string]interface{}, config map[string]interface{}) string {
	// Try config first
	if brand, ok := config["brand_name"].(string); ok && brand != "" {
		return brand
	}

	// Try input_data.domain and convert to title case
	if domain := datahelpers.ExtractNestedField(data, "input_data.domain"); domain != nil {
		if domainStr, ok := domain.(string); ok {
			// Convert domain to brand name: "ai-agent-orchestration.com" -> "AI Agent Orchestration"
			name := strings.TrimSuffix(domainStr, ".com")
			name = strings.TrimSuffix(name, ".io")
			name = strings.TrimSuffix(name, ".co")
			name = strings.ReplaceAll(name, "-", " ")
			name = strings.ReplaceAll(name, "_", " ")
			return strings.Title(name)
		}
	}

	return "Our Company"
}

// extractDomainForMultipage gets the domain from collected data
func extractDomainForMultipage(data map[string]interface{}, config map[string]interface{}) string {
	if domain := datahelpers.ExtractNestedField(data, "input_data.domain"); domain != nil {
		if domainStr, ok := domain.(string); ok {
			return domainStr
		}
	}
	if domain := datahelpers.ExtractNestedField(data, "domain"); domain != nil {
		if domainStr, ok := domain.(string); ok {
			return domainStr
		}
	}
	return "example.com"
}

// extractTaglineForMultipage tries to find tagline from collected data
func extractTaglineForMultipage(data map[string]interface{}, config map[string]interface{}) string {
	// Try to find tagline in content JSON
	if tagline := datahelpers.ExtractNestedField(data, "input_data.content_json.sections.component_footer_7.tagline"); tagline != nil {
		if taglineStr, ok := tagline.(string); ok && taglineStr != "" {
			return taglineStr
		}
	}
	return "Building the future, one step at a time"
}

// generateAboutPage creates a simple about page
func generateAboutPage(brandName, domain, tagline string) (string, error) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>About - {{.BrandName}}</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; line-height: 1.6; color: #333; background: #f8fafc; }
    nav { background: #1a1a2e; padding: 1rem 2rem; position: fixed; top: 0; width: 100%; z-index: 1000; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
    nav ul { list-style: none; display: flex; gap: 2rem; align-items: center; max-width: 1200px; margin: 0 auto; }
    nav a { color: #fff; text-decoration: none; font-weight: 500; transition: color 0.3s; }
    nav a:hover { color: #4a9eff; }
    .content { max-width: 800px; margin: 6rem auto 4rem; padding: 3rem; background: white; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
    .content h1 { font-size: 2.5rem; color: #1a1a2e; margin-bottom: 1.5rem; }
    .content p { font-size: 1.1rem; margin-bottom: 1.5rem; line-height: 1.8; color: #555; }
    .content h2 { font-size: 1.5rem; color: #1a1a2e; margin: 2rem 0 1rem; }
    .highlight { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; font-weight: 600; }
    footer { background: #1a1a2e; color: white; text-align: center; padding: 2rem; margin-top: 4rem; }
    footer a { color: #4a9eff; text-decoration: none; }
  </style>
</head>
<body>
  <nav>
    <ul>
      <li><a href="index.html">Home</a></li>
      <li><a href="about.html">About</a></li>
      <li><a href="contact.html">Contact</a></li>
    </ul>
  </nav>
  
  <main class="content">
    <h1>About {{.BrandName}}</h1>
    
    <p>{{.Tagline}}</p>
    
    <h2>Our Mission</h2>
    <p>We're dedicated to delivering exceptional value to our customers through innovation, quality, and outstanding service. Every decision we make is guided by our commitment to excellence.</p>
    
    <h2>Why Choose Us</h2>
    <p>With a focus on results and customer satisfaction, we've built a reputation for reliability and expertise. Our team brings together diverse skills and perspectives to solve complex challenges.</p>
    
    <h2>Looking Forward</h2>
    <p>We're constantly evolving and improving. As technology and markets change, we adapt—always keeping our customers' needs at the center of everything we do.</p>
  </main>
  
  <footer>
    <p>&copy; 2025 {{.BrandName}}. All rights reserved.</p>
  </footer>
</body>
</html>`

	t, err := template.New("about").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	err = t.Execute(&buf, map[string]string{
		"BrandName": brandName,
		"Domain":    domain,
		"Tagline":   tagline,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateContactPage creates a simple contact page
func generateContactPage(brandName, domain string) (string, error) {
	// Generate email from domain
	email := "hello@" + domain

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Contact - {{.BrandName}}</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; line-height: 1.6; color: #333; background: #f8fafc; }
    nav { background: #1a1a2e; padding: 1rem 2rem; position: fixed; top: 0; width: 100%; z-index: 1000; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
    nav ul { list-style: none; display: flex; gap: 2rem; align-items: center; max-width: 1200px; margin: 0 auto; }
    nav a { color: #fff; text-decoration: none; font-weight: 500; transition: color 0.3s; }
    nav a:hover { color: #4a9eff; }
    .content { max-width: 800px; margin: 6rem auto 4rem; padding: 3rem; background: white; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
    .content h1 { font-size: 2.5rem; color: #1a1a2e; margin-bottom: 1.5rem; }
    .content p { font-size: 1.1rem; margin-bottom: 1.5rem; line-height: 1.8; color: #555; }
    .contact-card { background: linear-gradient(135deg, #f5f7fa 0%, #e4e8ec 100%); padding: 2rem; border-radius: 10px; margin: 2rem 0; }
    .contact-card h2 { color: #1a1a2e; margin-bottom: 1rem; font-size: 1.3rem; }
    .contact-item { margin: 1rem 0; }
    .contact-item strong { color: #1a1a2e; }
    .contact-item a { color: #4a9eff; text-decoration: none; }
    .contact-item a:hover { text-decoration: underline; }
    footer { background: #1a1a2e; color: white; text-align: center; padding: 2rem; margin-top: 4rem; }
    footer a { color: #4a9eff; text-decoration: none; }
  </style>
</head>
<body>
  <nav>
    <ul>
      <li><a href="index.html">Home</a></li>
      <li><a href="about.html">About</a></li>
      <li><a href="contact.html">Contact</a></li>
    </ul>
  </nav>
  
  <main class="content">
    <h1>Get in Touch</h1>
    
    <p>We'd love to hear from you! Whether you have questions, feedback, or just want to say hello, don't hesitate to reach out.</p>
    
    <div class="contact-card">
      <h2>Contact Information</h2>
      
      <div class="contact-item">
        <strong>Email:</strong> <a href="mailto:{{.Email}}">{{.Email}}</a>
      </div>
      
      <div class="contact-item">
        <strong>Website:</strong> <a href="https://{{.Domain}}">{{.Domain}}</a>
      </div>
    </div>
    
    <p>We typically respond within 24-48 hours during business days. For urgent matters, please indicate so in your message subject line.</p>
  </main>
  
  <footer>
    <p>&copy; 2025 {{.BrandName}}. All rights reserved.</p>
  </footer>
</body>
</html>`

	t, err := template.New("contact").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	err = t.Execute(&buf, map[string]string{
		"BrandName": brandName,
		"Domain":    domain,
		"Email":     email,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
