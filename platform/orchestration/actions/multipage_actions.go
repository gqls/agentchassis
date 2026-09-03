// FILE: platform/orchestration/actions/multipage_actions.go
package actions

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// AssemblePageAction builds a single complete HTML page
// Takes HTML content and ensures it's a valid, complete page
func AssemblePageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Assembling single page")

	config := params.StepConfig.Config
	if config == nil {
		return nil, fmt.Errorf("config is required for assemble_page action")
	}

	contentField, _ := config["content_field"].(string)
	if contentField == "" {
		return nil, fmt.Errorf("content_field is required in config")
	}

	addNav, _ := config["add_navigation"].(bool)

	// --- Page-ownership guard (bugs_open/208) ---
	//
	// Every consumer of this action feeds a loop whose next step is deploy_page
	// (git_commit) — so this is the last point at which an owned page can be
	// spared before its regenerated prose replaces the live tool in the site repo.
	// The refusal is expressed as the action's EXISTING skip shape, not an error,
	// for two reasons: git_commit already honours it (checkUpstreamSkipped), so no
	// agent config changes; and none of the three loops sets continue_on_error, so
	// an error here would fail the whole workflow and strand every page after this
	// one. See owned_page_guard.go for why this seam rather than git_commit.
	if pageID, pageName, ok := resolveGuardedPage(ctx, params.DB, params.CollectedData, params.Logger); ok {
		refused, class, checked := pageRefusesGenericBuild(ctx, params.DB, pageID, params.Logger)

		// The fail-open window, made countable instead of silent (council
		// `bug_historian`, medium). This page is about to be composed and committed
		// WITHOUT an ownership check having succeeded — rare, and exactly the case
		// that would otherwise look identical to a generic page in every record.
		if !checked {
			LogActionError(ctx, params,
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id"),
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
				"assemble_page", "OWNED_PAGE_GUARD_UNCHECKED", "high",
				fmt.Sprintf("build policy for page %s (%s) could not be read; generic assembly proceeded without an ownership or tool-shell check",
					pageName, pageID),
				map[string]interface{}{"page_id": pageID.String(), "page_name": pageName},
				params.Logger)
		}

		if refused {
			reason := fmt.Sprintf(
				"%s: page %s is rebuild_policy=owned (tool/widget-owned); a generic recomposition "+
					"would be committed over the live page. Use the tool pipeline for rebuilds or "+
					"apply_section_edit for targeted edits.",
				ownedPageSkipReasonPrefix, pageName)
			if class == refusalToolPending {
				reason = fmt.Sprintf(
					"%s: page %s is page_type=tool with no tool component; a generic recomposition "+
						"would commit prose about a tool that is not there. The tool pipeline builds "+
						"the component and this refusal then lifts by itself.",
					ownedPageSkipReasonPrefix, pageName)
			}

			params.Logger.Warn("AssemblePageAction: PAGE REFUSES GENERIC BUILD — assembly refused before deploy",
				zap.String("page_name", pageName),
				zap.String("page_id", pageID.String()),
				zap.String("refusal_class", class),
			)

			if siteID, parseErr := uuid.Parse(
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id"),
			); parseErr == nil {
				emitOwnedPageReviewItem(ctx, params.DB, siteID, pageName,
					"assemble_page", reason, class, params.Logger)
			}

			return map[string]interface{}{
				"html":         "",
				"skipped":      true,
				"skip_reason":  reason,
				"assembled_at": params.ExecutionContext.Timestamp,
			}, nil
		}
	}

	// Check if upstream content generation failed before trying to extract HTML
	// contentField is typically "page_content.response.page_html"
	// We need to check "page_content.response" for status/error fields
	if upstreamFailed, failureReason := checkUpstreamContentFailure(params.CollectedData, contentField, params.Logger); upstreamFailed {
		params.Logger.Warn("Upstream content generation failed, skipping page assembly",
			zap.String("content_field", contentField),
			zap.String("reason", failureReason))

		// Return success with skipped flag - allows loop to continue to next page
		return map[string]interface{}{
			"html":         "",
			"skipped":      true,
			"skip_reason":  failureReason,
			"assembled_at": params.ExecutionContext.Timestamp,
		}, nil
	}

	// Extract content
	content := extractFieldValue(params.CollectedData, contentField, params.Logger)
	if content == "" {
		// No content found - treat as skipped rather than error
		// This allows the loop to continue with other pages
		params.Logger.Warn("No content found at specified field, skipping page",
			zap.String("content_field", contentField))

		// Council 3918db52 (bug_historian, medium): a legitimate writer skip and
		// a mis-set content_field used to produce IDENTICAL quiet skips once the
		// 408 crash was fixed. When the upstream step did not declare itself
		// skipped, content was expected here — make that case countable rather
		// than letting it hide among routine skips (the OWNED_PAGE_GUARD_UNCHECKED
		// pattern). The skip shape below is returned either way; this is
		// additive observability, not a behaviour change.
		if !upstreamDeclaredSkip(params.CollectedData, contentField) {
			LogActionError(ctx, params,
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id"),
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
				"assemble_page", "ASSEMBLE_CONTENT_FIELD_UNRESOLVED", "high",
				fmt.Sprintf("content_field %q resolved on no candidate form and the upstream step did not declare a skip — content was expected here (likely a mis-set content_field, or a writer that failed without status)", contentField),
				map[string]interface{}{
					"content_field": contentField,
					"page_name":     datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.name"),
				},
				params.Logger)
		}

		return map[string]interface{}{
			"html":         "",
			"skipped":      true,
			"skip_reason":  fmt.Sprintf("no content found at %s", contentField),
			"assembled_at": params.ExecutionContext.Timestamp,
		}, nil
	}

	params.Logger.Info("Extracted content",
		zap.String("field", contentField),
		zap.Int("length", len(content)),
	)

	// Clean HTML (remove markdown code blocks, etc.)
	html := datahelpers.CleanHTMLString(content)

	// Ensure valid HTML structure
	html = cleanHTMLStructure(html)

	// Optionally inject head/header/footer from component library
	injectHead, _ := config["inject_head"].(bool)
	injectHeader, _ := config["inject_header"].(bool)
	injectFooter, _ := config["inject_footer"].(bool)

	if params.DB != nil && (injectHead || injectHeader || injectFooter) {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		siteUUID := uuid.Nil
		if siteIDStr != "" {
			siteUUID, _ = uuid.Parse(siteIDStr)
		}

		renderCtx := buildRenderContextFromCollectedData(params.CollectedData, params.Logger)

		// Set page-specific title/description from current_page
		if cp := datahelpers.ExtractNestedField(params.CollectedData, "current_page"); cp != nil {
			if cpMap, ok := cp.(map[string]interface{}); ok {
				if t, ok := cpMap["title"].(string); ok && t != "" {
					renderCtx.Title = t
				} else if n, ok := cpMap["name"].(string); ok && n != "" {
					renderCtx.Title = strings.Title(strings.ReplaceAll(n, "-", " "))
				}
				if d, ok := cpMap["meta_description"].(string); ok && d != "" {
					renderCtx.Description = d
				}
			}
		}

		if injectHead {
			html = InjectHead(ctx, params.DB, html, siteUUID, renderCtx, params.Logger)
		}
		if injectHeader {
			html = InjectHeader(ctx, params.DB, html, siteUUID, renderCtx, params.Logger)
		}
		if injectFooter {
			html = InjectFooter(ctx, params.DB, html, siteUUID, renderCtx, params.Logger)
		}

		params.Logger.Info("AssemblePageAction: Injected components",
			zap.Bool("head", injectHead),
			zap.Bool("header", injectHeader),
			zap.Bool("footer", injectFooter),
		)
	}

	// Add navigation if requested
	if addNav {
		html = addSimpleNavigation(html, "index")
	}

	// bugs_open/328 — the LAST point before deploy_page on every loop this action
	// feeds, and the only one they share. It is placed after chrome injection so
	// the whole outbound string is judged, matching what the rerender seam already
	// does; and it writes nothing, so content_data keeps the authored href and the
	// anchor returns by itself once the target ships. Opt-in, default OFF: with
	// no `suppress_unshipped_links` in the step config this call returns `html`
	// byte-identical, which is what every consumer sees until a migration enables
	// it. See refused_link_targets.go for the policy and its two escapes.
	if params.DB != nil {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		siteUUID, siteErr := uuid.Parse(siteIDStr)
		switch {
		case siteErr != nil && configBoolOrDefault(config, suppressUnshippedLinksKey, false):
			// A SILENT skip here would reproduce this bug's own shape: the seam
			// that was found to have no outbound link protection at all fails to
			// apply the check and nothing says so. Only warn when the step asked
			// for suppression — an un-opted-in step skipping is not an event.
			params.Logger.Warn("assemble_page: outbound link suppression SKIPPED — site_record.site_id missing or unparseable; page assembled unsuppressed (bugs_open/328)",
				zap.String("site_id_raw", siteIDStr),
				zap.Error(siteErr))
		case siteErr == nil:
			html = suppressUnshippedOutboundLinks(ctx, params, siteUUID,
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
				datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.name"),
				datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.url"),
				html, params.Logger)
		}
	}

	params.Logger.Info("Page assembled successfully",
		zap.Int("final_length", len(html)),
		zap.Bool("added_navigation", addNav),
	)

	return map[string]interface{}{
		"html":         html,
		"skipped":      false,
		"assembled_at": params.ExecutionContext.Timestamp,
	}, nil
}

// AssembleMultipageSiteAction creates a complete multi-page site
// Takes pages from loop output, adds navigation, generates standard pages
func AssembleMultipageSiteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Assembling multi-page site with component-based headers")

	config := params.StepConfig.Config
	if config == nil {
		return nil, fmt.Errorf("config is required for assemble_multipage_site action")
	}

	pagesField, _ := config["pages_field"].(string)
	if pagesField == "" {
		return nil, fmt.Errorf("pages_field is required in config")
	}

	generateStandard, _ := config["generate_standard_pages"].(bool)

	// Extract pages from loop output
	pages := extractPagesFromLoop(params.CollectedData, pagesField, params.Logger)

	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found at %s", pagesField)
	}

	params.Logger.Info("Extracted pages from loop",
		zap.Int("page_count", len(pages)),
		zap.Strings("page_names", getPageNames(pages)),
	)

	// Generate standard pages if requested and missing
	if generateStandard {
		domain := extractDomainFromData(params.CollectedData)

		if _, hasAbout := pages["about.html"]; !hasAbout {
			pages["about.html"] = generateAboutPage(domain)
			params.Logger.Info("Generated about page")
		}

		if _, hasContact := pages["contact.html"]; !hasContact {
			pages["contact.html"] = generateContactPage(domain)
			params.Logger.Info("Generated contact page")
		}
	}

	// Build list of all page names for link fixing
	pageNamesList := make([]string, 0, len(pages))
	for name := range pages {
		pageNamesList = append(pageNamesList, name)
	}

	// Get site ID for component lookup
	siteID := extractSiteIDFromCollectedData(params.CollectedData)

	// Build shared RenderContext from collected data
	// This is used by component_library.go functions
	baseRenderCtx := buildRenderContextFromCollectedData(params.CollectedData, params.Logger)

	// POST-PROCESS ALL PAGES
	for name, html := range pages {
		// Step 1: Clean HTML structure (fix double DOCTYPE)
		html = cleanHTMLStructure(html)

		// Step 2: Fix anchor links to page links
		html = fixAnchorLinks(html, pageNamesList)

		// Step 3: Build page-specific render context
		currentPage := strings.TrimSuffix(name, ".html")
		renderCtx := copyRenderContext(baseRenderCtx)
		renderCtx.CurrentPage = currentPage
		renderCtx.NavItems = setActiveNavItems(baseRenderCtx.NavItems, currentPage)

		// Step 4: Inject header and footer
		// Use component_library.go functions if DB available, otherwise fallback
		if params.DB != nil {
			// Use component-based injection from component_library.go
			html = InjectHeader(ctx, params.DB, html, siteID, renderCtx, params.Logger)
			html = InjectFooter(ctx, params.DB, html, siteID, renderCtx, params.Logger)
		} else {
			// Fallback to hardcoded templates (existing behavior)
			headerConfig := convertRenderContextToHeaderConfig(renderCtx)
			html = injectConsistentHeader(html, headerConfig, params.Logger)
			// Note: existing code doesn't inject footer, add if needed
		}

		pages[name] = html

		params.Logger.Debug("Post-processed page",
			zap.String("page", name),
			zap.Int("nav_items", len(renderCtx.NavItems)),
			zap.Int("final_length", len(html)),
		)
	}

	params.Logger.Info("Multi-page site assembled with component-based headers",
		zap.Int("total_pages", len(pages)),
		zap.Int("total_bytes", calculateTotalSize(pages)),
	)

	return map[string]interface{}{
		"files":        pages,
		"page_count":   len(pages),
		"total_bytes":  calculateTotalSize(pages),
		"page_names":   getPageNames(pages),
		"assembled_at": params.ExecutionContext.Timestamp,
	}, nil
}

// These functions bridge between existing code and component_library.go
// extractSiteIDFromCollectedData gets the site UUID from collected data
func extractSiteIDFromCollectedData(collectedData map[string]interface{}) uuid.UUID {
	if siteRecord, ok := collectedData["site_record"].(map[string]interface{}); ok {
		if idStr, ok := siteRecord["site_id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

// buildRenderContextFromCollectedData creates a RenderContext for component_library.go
// This bridges the existing data extraction to the new RenderContext type
func buildRenderContextFromCollectedData(collectedData map[string]interface{}, logger *zap.Logger) *RenderContext {
	ctx := &RenderContext{}

	// Extract from input_data
	if inputData, ok := collectedData["input_data"].(map[string]interface{}); ok {
		// Domain
		if domain, ok := inputData["domain"].(string); ok {
			ctx.Domain = domain
			// Default logo from domain
			if parts := strings.Split(domain, "."); len(parts) > 0 && len(parts[0]) > 0 {
				ctx.LogoText = datahelpers.UpperFirst(parts[0])
			}
		}

		// From reviewed_brief
		if brief, ok := inputData["reviewed_brief"].(map[string]interface{}); ok {
			// Try company_name first (most common), then business_name
			if name, ok := brief["company_name"].(string); ok && name != "" {
				ctx.LogoText = name
				ctx.CompanyName = name
			} else if name, ok := brief["business_name"].(string); ok && name != "" {
				ctx.LogoText = name
				ctx.CompanyName = name
			}
			if tagline, ok := brief["tagline"].(string); ok {
				ctx.Tagline = tagline
			}
			if email, ok := brief["contact_email"].(string); ok {
				ctx.Email = email
			}
			if phone, ok := brief["contact_phone"].(string); ok {
				ctx.Phone = phone
			}
			// Colors from color_scheme
			if colorScheme, ok := brief["color_scheme"].(string); ok {
				ctx.PrimaryColor, ctx.AccentColor = parseColorScheme(colorScheme)
			}
		}

		// Direct fields fallback
		if name, ok := inputData["business_name"].(string); ok && name != "" && ctx.LogoText == "" {
			ctx.LogoText = name
			ctx.CompanyName = name
		}
	}

	// Extract navigation — prefer nav_data (from populate_nav step), then db_sync
	ctx.NavItems = extractNavItemsFromCollectedData(collectedData, logger)

	// Fallback navigation if none found
	if len(ctx.NavItems) == 0 {
		logger.Warn("No navigation found in nav_data or db_sync, using defaults")
		ctx.NavItems = []NavItem{
			{Label: "Home", URL: "/index.html"},
			{Label: "About", URL: "/about.html"},
			{Label: "Services", URL: "/services.html"},
			{Label: "Contact", URL: "/contact.html"},
		}
	}

	// Extract logo_url from collected_data (may have been set by deploy or asset steps)
	if logoURL := datahelpers.ExtractNestedFieldString(collectedData, "logo_url"); logoURL != "" {
		ctx.LogoURL = logoURL
	}
	// Also check site_record path
	if ctx.LogoURL == "" {
		if logoURL := datahelpers.ExtractNestedFieldString(collectedData, "site_record.logo_url"); logoURL != "" {
			ctx.LogoURL = logoURL
		}
	}

	// bugs_open/420: NO fallback email from the domain — see the same removal in
	// section_editor_actions.go. A synthesised "info@<domain>" is a fabricated
	// contact for a business that was never asked whether it wanted one
	// published, and it would make the post-fix default ("the site publishes no
	// contact") false on the full-page build path specifically.

	// Fallback company name
	if ctx.CompanyName == "" {
		ctx.CompanyName = ctx.LogoText
	}

	return ctx
}

// copyRenderContext creates a shallow copy of RenderContext
func copyRenderContext(src *RenderContext) *RenderContext {
	if src == nil {
		return &RenderContext{}
	}
	// copies all struct fields
	cpy := *src
	// Deep copy NavItems slice
	cpy.NavItems = make([]NavItem, len(src.NavItems))
	copy(cpy.NavItems, src.NavItems)
	return &cpy
}

// setActiveNavItems returns nav items with correct active state for current page
func setActiveNavItems(items []NavItem, currentPage string) []NavItem {
	result := make([]NavItem, len(items))
	for i, item := range items {
		result[i] = item
		urlPage := strings.TrimSuffix(strings.TrimPrefix(item.URL, "/"), ".html")
		result[i].IsActive = urlPage == currentPage ||
			(currentPage == "index" && (urlPage == "home" || urlPage == "index"))
	}
	return result
}

// convertRenderContextToHeaderConfig converts RenderContext to HeaderConfig for fallback
// This allows using the existing injectConsistentHeader when DB is not available
func convertRenderContextToHeaderConfig(ctx *RenderContext) *HeaderConfig {
	config := &HeaderConfig{
		LogoText:     ctx.LogoText,
		LogoURL:      ctx.LogoURL,
		PrimaryColor: ctx.PrimaryColor,
		AccentColor:  ctx.AccentColor,
		CurrentPage:  ctx.CurrentPage,
		IsHomePage:   ctx.CurrentPage == "index" || ctx.CurrentPage == "home",
	}

	// Convert NavItems
	for _, item := range ctx.NavItems {
		config.NavItems = append(config.NavItems, NavItem{
			Label:    item.Label,
			URL:      item.URL,
			IsActive: item.IsActive,
		})
	}

	// Set defaults if missing
	if config.PrimaryColor == "" {
		config.PrimaryColor = "#1a1a2e"
	}
	if config.AccentColor == "" {
		config.AccentColor = "#16a085"
	}

	// Extract logo URL from ContentData
	if ctx.ContentData != nil {
		if logoURL, ok := ctx.ContentData["logo_url"].(string); ok && logoURL != "" {
			config.LogoURL = logoURL
		}
	}

	return config
}

// ============================================================================
// Helper Functions
// ============================================================================
func cleanHTMLStructure(html string) string {
	html = strings.TrimSpace(html)

	// Remove duplicate DOCTYPEs (case-insensitive)
	// Pattern: <!DOCTYPE html>...<!doctype html> or variations
	doctypeRe := regexp.MustCompile(`(?i)<!doctype\s+html\s*>`)
	matches := doctypeRe.FindAllStringIndex(html, -1)

	if len(matches) > 1 {
		// Keep only the first DOCTYPE, remove others
		// Work backwards to preserve indices
		for i := len(matches) - 1; i > 0; i-- {
			start := matches[i][0]
			end := matches[i][1]
			html = html[:start] + html[end:]
		}
	}

	// Remove duplicate <html> tags
	htmlTagRe := regexp.MustCompile(`(?i)<html[^>]*>`)
	htmlMatches := htmlTagRe.FindAllStringIndex(html, -1)
	if len(htmlMatches) > 1 {
		for i := len(htmlMatches) - 1; i > 0; i-- {
			start := htmlMatches[i][0]
			end := htmlMatches[i][1]
			html = html[:start] + html[end:]
		}
	}

	// Remove duplicate <head> sections — keep the one before <body>, remove any after.
	// Previous version kept the LARGER one regardless of position, which could
	// remove a correct <head> (before <body>) in favour of a misplaced one (inside <body>)
	// from LLM-generated section content.
	// Use regex with [\s>] to match <head> or <head ...> but NOT <header>
	headTagRe := regexp.MustCompile(`(?i)<head[\s>]`)
	headPositions := headTagRe.FindAllStringIndex(html, -1)
	if len(headPositions) > 1 {
		lowerHTML := strings.ToLower(html)

		// Find <body> position to determine which <head> is correctly placed
		bodyPos := strings.Index(lowerHTML, "<body")

		// Remove all <head>...</head> blocks that are AFTER <body> — they're misplaced.
		// Work backwards to preserve string indices.
		for i := len(headPositions) - 1; i >= 0; i-- {
			headStart := headPositions[i][0]
			if bodyPos >= 0 && headStart > bodyPos {
				// This <head> is after <body> — it's inside body content, remove it
				headEndIdx := strings.Index(lowerHTML[headStart:], "</head>")
				if headEndIdx >= 0 {
					headEnd := headStart + headEndIdx + 7 // include </head>
					html = html[:headStart] + html[headEnd:]
					lowerHTML = strings.ToLower(html) // refresh after modification
				}
			}
		}

		// If no <body> tag found, fall back to keeping only the first <head>
		if bodyPos < 0 {
			// Re-scan after possible modifications
			headPositions = headTagRe.FindAllStringIndex(html, -1)
			if len(headPositions) > 1 {
				lowerHTML = strings.ToLower(html)
				// Remove all but the first <head>...</head>
				for i := len(headPositions) - 1; i > 0; i-- {
					headStart := headPositions[i][0]
					headEndIdx := strings.Index(lowerHTML[headStart:], "</head>")
					if headEndIdx >= 0 {
						headEnd := headStart + headEndIdx + 7
						html = html[:headStart] + html[headEnd:]
						lowerHTML = strings.ToLower(html)
					}
				}
			}
		}
	}

	// Remove duplicate <body> tags
	bodyTagRe := regexp.MustCompile(`(?i)<body[^>]*>`)
	bodyMatches := bodyTagRe.FindAllStringIndex(html, -1)
	if len(bodyMatches) > 1 {
		for i := len(bodyMatches) - 1; i > 0; i-- {
			start := bodyMatches[i][0]
			end := bodyMatches[i][1]
			html = html[:start] + html[end:]
		}
	}

	// Ensure we have a DOCTYPE at the start
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(html)), "<!DOCTYPE") {
		html = "<!DOCTYPE html>\n" + html
	}

	return html
}

func fixAnchorLinks(html string, pageNames []string) string {
	// Build set of known pages
	knownPages := make(map[string]bool)
	for _, name := range pageNames {
		cleanName := strings.TrimSuffix(strings.ToLower(name), ".html")
		knownPages[cleanName] = true
	}

	// Add common page names
	commonPages := []string{
		"home", "index", "about", "services", "contact", "insights", "blog",
		"careers", "team", "portfolio", "case-studies", "privacy", "terms",
		"pricing", "faq", "support", "features", "solutions", "resources",
		"testimonials", "clients", "work", "projects",
	}
	for _, name := range commonPages {
		knownPages[name] = true
	}

	result := html

	for pageName := range knownPages {
		targetURL := "/" + pageName + ".html"
		if pageName == "home" || pageName == "index" {
			targetURL = "/index.html"
		}

		// Replace various patterns
		patterns := []struct {
			old string
			new string
		}{
			// Double quotes
			{fmt.Sprintf(`href="#%s"`, pageName), fmt.Sprintf(`href="%s"`, targetURL)},
			{fmt.Sprintf(`href="#%s"`, strings.Title(pageName)), fmt.Sprintf(`href="%s"`, targetURL)},
			// Single quotes
			{fmt.Sprintf(`href='#%s'`, pageName), fmt.Sprintf(`href='%s'`, targetURL)},
			{fmt.Sprintf(`href='#%s'`, strings.Title(pageName)), fmt.Sprintf(`href='%s'`, targetURL)},
			// No quotes (minified)
			{fmt.Sprintf(`href=#%s>`, pageName), fmt.Sprintf(`href=%s>`, targetURL)},
			{fmt.Sprintf(`href=#%s `, pageName), fmt.Sprintf(`href=%s `, targetURL)},
		}

		for _, p := range patterns {
			result = strings.ReplaceAll(result, p.old, p.new)
		}
	}

	return result
}

// extractNavItemsFromCollectedData gets nav items from collectedData.
// Checks nav_data (from populate_nav step) first, then db_sync.
// Both store navigation in the same shape: {navigation: {items: [{label, url}]}}.
func extractNavItemsFromCollectedData(collectedData map[string]interface{}, logger *zap.Logger) []NavItem {
	// Priority 1: nav_data from populate_nav step (authoritative)
	if items := extractNavFromKey(collectedData, "nav_data", logger); len(items) > 0 {
		logger.Debug("Using navigation from nav_data (populate_nav)",
			zap.Int("items", len(items)),
		)
		return items
	}

	// Priority 2: db_sync from sync_pages_to_db step (legacy)
	if items := extractNavFromKey(collectedData, "db_sync", logger); len(items) > 0 {
		logger.Debug("Using navigation from db_sync",
			zap.Int("items", len(items)),
		)
		return items
	}

	return nil
}

// extractNavFromKey extracts []NavItem from collectedData[key].navigation.items
func extractNavFromKey(collectedData map[string]interface{}, key string, logger *zap.Logger) []NavItem {
	container, ok := collectedData[key].(map[string]interface{})
	if !ok {
		return nil
	}
	nav, ok := container["navigation"].(map[string]interface{})
	if !ok {
		return nil
	}
	rawItems, ok := nav["items"].([]interface{})
	if !ok {
		return nil
	}
	var items []NavItem
	for _, item := range rawItems {
		if itemMap, ok := item.(map[string]interface{}); ok {
			label, _ := itemMap["label"].(string)
			url, _ := itemMap["url"].(string)
			if label != "" && url != "" {
				items = append(items, NavItem{Label: label, URL: url})
			}
		}
	}
	return items
}

func extractCanonicalNavigation(collectedData map[string]interface{}, logger *zap.Logger) []map[string]string {
	var navItems []map[string]string

	// Priority 1: db_sync.navigation (from sync_pages_to_db)
	if dbSync, ok := collectedData["db_sync"].(map[string]interface{}); ok {
		if nav, ok := dbSync["navigation"].(map[string]interface{}); ok {
			if items, ok := nav["items"].([]interface{}); ok {
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						label, _ := itemMap["label"].(string)
						url, _ := itemMap["url"].(string)
						if label != "" && url != "" {
							navItems = append(navItems, map[string]string{
								"label": label,
								"url":   url,
							})
						}
					}
				}
				if len(navItems) > 0 {
					logger.Info("Using navigation from db_sync",
						zap.Int("nav_items", len(navItems)),
					)
					return navItems
				}
			}
		}
	}

	// Priority 2: page_plan.plan_data.sitemap
	if pagePlan, ok := collectedData["page_plan"].(map[string]interface{}); ok {
		var sitemap []interface{}

		if planData, ok := pagePlan["plan_data"].(map[string]interface{}); ok {
			if sm, ok := planData["sitemap"].([]interface{}); ok {
				sitemap = sm
			}
		}
		if sitemap == nil {
			if sm, ok := pagePlan["sitemap"].([]interface{}); ok {
				sitemap = sm
			}
		}

		for _, entry := range sitemap {
			if e, ok := entry.(map[string]interface{}); ok {
				label, _ := e["label"].(string)
				if label == "" {
					label, _ = e["title"].(string)
				}
				if label == "" {
					label, _ = e["name"].(string)
				}

				url, _ := e["url"].(string)
				if url == "" {
					if name, ok := e["name"].(string); ok {
						if name == "index" || name == "home" {
							url = "/index.html"
						} else {
							url = "/" + name + ".html"
						}
					}
				}

				inHeader := true
				if ih, ok := e["in_header"].(bool); ok {
					inHeader = ih
				}

				// Skip footer-only pages
				nameLower := strings.ToLower(label)
				if nameLower == "privacy" || nameLower == "terms" ||
					strings.Contains(nameLower, "privacy") || strings.Contains(nameLower, "terms") {
					inHeader = false
				}

				if inHeader && label != "" && url != "" {
					navItems = append(navItems, map[string]string{
						"label": label,
						"url":   url,
					})
				}
			}
		}

		if len(navItems) > 0 {
			logger.Info("Using navigation from page_plan sitemap",
				zap.Int("nav_items", len(navItems)),
			)
			return navItems
		}
	}

	logger.Warn("No canonical navigation found")
	return nil
}

// ===========================================================================
// NEW FUNCTION: injectCanonicalNavigation
// ===========================================================================
// Replaces header navigation with canonical navigation

func injectCanonicalNavigation(html string, navItems []map[string]string, currentPageName string, logger *zap.Logger) string {
	// Build the navigation HTML
	var navLinks []string
	currentPage := strings.TrimSuffix(currentPageName, ".html")

	for _, item := range navItems {
		label := item["label"]
		url := item["url"]

		// Add active class for current page
		activeClass := ""
		itemPage := strings.TrimSuffix(strings.TrimPrefix(url, "/"), ".html")
		if itemPage == currentPage || (currentPage == "index" && (itemPage == "home" || itemPage == "index")) {
			activeClass = ` class="active"`
		}

		navLinks = append(navLinks, fmt.Sprintf(`<a href="%s"%s>%s</a>`, url, activeClass, label))
	}

	navHTML := strings.Join(navLinks, "\n            ")

	// Try to find and replace existing nav content
	// Pattern: <nav...>...</nav>
	navRe := regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`)

	if navRe.MatchString(html) {
		// Replace the nav content but keep any nav attributes
		html = navRe.ReplaceAllStringFunc(html, func(match string) string {
			// Extract opening nav tag
			openTagEnd := strings.Index(match, ">")
			if openTagEnd < 0 {
				return match
			}
			openTag := match[:openTagEnd+1]

			return fmt.Sprintf(`%s
        <ul>
            %s
        </ul>
    </nav>`, strings.TrimSuffix(openTag, ">")+" id=\"main-nav\">",
				strings.ReplaceAll(navHTML, "<a ", "<li><a "))
		})

		logger.Debug("Replaced nav content",
			zap.String("page", currentPageName),
		)
	} else {
		// No nav found, try to insert after header tag
		headerEnd := strings.Index(strings.ToLower(html), "<header")
		if headerEnd >= 0 {
			// Find the end of header opening tag
			closeIdx := strings.Index(html[headerEnd:], ">")
			if closeIdx >= 0 {
				insertPoint := headerEnd + closeIdx + 1
				newNav := fmt.Sprintf(`
    <nav id="main-nav">
        <ul>
            %s
        </ul>
    </nav>`, strings.ReplaceAll(navHTML, "<a ", "<li><a "))
				html = html[:insertPoint] + newNav + html[insertPoint:]
			}
		}

		logger.Debug("Inserted new nav",
			zap.String("page", currentPageName),
		)
	}

	return html
}

// extractPagesFromLoop handles different loop output formats
func extractPagesFromLoop(data map[string]interface{}, fieldPath string, logger *zap.Logger) map[string]string {
	pages := make(map[string]string)

	value := extractNestedField(data, fieldPath)
	if value == nil {
		logger.Warn("No value found at field path", zap.String("path", fieldPath))
		return pages
	}

	logger.Info("Extracting pages from loop output",
		zap.String("value_type", fmt.Sprintf("%T", value)),
	)

	// Format 1: Map - could be direct pages OR loop_complete output
	if pagesMap, ok := value.(map[string]interface{}); ok {
		// Check for loop_complete output format: {iterations: N, results: [...]}
		if results, hasResults := pagesMap["results"].([]interface{}); hasResults {
			logger.Info("Detected loop_complete format, extracting from results array",
				zap.Int("results_count", len(results)))

			// Process the results array (same as Format 2)
			for i, item := range results {
				if itemMap, ok := item.(map[string]interface{}); ok {
					name := fmt.Sprintf("page_%d", i)

					// Try to get name from item
					if itemName, ok := itemMap["name"].(string); ok && itemName != "" {
						name = itemName
					} else if itemName, ok := itemMap["page_name"].(string); ok && itemName != "" {
						name = itemName
					}

					// Extract HTML
					html := extractHTMLFromValue(itemMap, logger)
					if html != "" {
						filename := name
						if !strings.HasSuffix(filename, ".html") {
							filename = filename + ".html"
						}
						pages[filename] = html
						logger.Debug("Extracted page from loop_complete results",
							zap.String("name", filename),
							zap.Int("index", i),
							zap.Int("length", len(html)),
						)
					}
				}
			}
			return pages
		}

		// Otherwise, treat as direct map of pages {"index": "...", "about": "...", ...}
		for name, content := range pagesMap {
			html := extractHTMLFromValue(content, logger)
			if html != "" {
				filename := name
				if !strings.HasSuffix(filename, ".html") {
					filename = filename + ".html"
				}
				pages[filename] = html
				logger.Debug("Extracted page from map",
					zap.String("name", filename),
					zap.Int("length", len(html)),
				)
			}
		}
		return pages
	}

	// Format 2: Array of page objects [{"name": "index", "page_html": "..."}, ...]
	if pagesArray, ok := value.([]interface{}); ok {
		for i, item := range pagesArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				name := fmt.Sprintf("page_%d", i)

				// Try to get name from item
				if itemName, ok := itemMap["name"].(string); ok && itemName != "" {
					name = itemName
				} else if itemName, ok := itemMap["page_name"].(string); ok && itemName != "" {
					name = itemName
				}

				// Extract HTML
				html := extractHTMLFromValue(itemMap, logger)
				if html != "" {
					filename := name
					if !strings.HasSuffix(filename, ".html") {
						filename = filename + ".html"
					}
					pages[filename] = html
					logger.Debug("Extracted page from array",
						zap.String("name", filename),
						zap.Int("index", i),
						zap.Int("length", len(html)),
					)
				}
			}
		}
		return pages
	}

	logger.Warn("Could not extract pages from value",
		zap.String("type", fmt.Sprintf("%T", value)))
	return pages
}

// extractHTMLFromValue tries to extract HTML string from various value types
func extractHTMLFromValue(value interface{}, logger *zap.Logger) string {
	// Direct string
	if html, ok := value.(string); ok {
		return html
	}

	// Map with HTML field
	if m, ok := value.(map[string]interface{}); ok {
		// Try common field names
		fieldNames := []string{"html", "page_html", "content", "result", "output"}
		for _, fieldName := range fieldNames {
			if html, ok := m[fieldName].(string); ok && html != "" {
				return html
			}
		}
	}

	return ""
}

// ensureValidHTML ensures HTML has proper structure
func ensureValidHTML(html string) string {
	// Just use cleanHTMLStructure - it handles everything
	return cleanHTMLStructure(html)
}

// addSimpleNavigation adds a simple navigation bar
func addSimpleNavigation(html string, currentPage string) string {
	// Determine current page for active state
	pageClass := func(page string) string {
		if page == currentPage {
			return ` class="active"`
		}
		return ""
	}

	nav := fmt.Sprintf(`<nav style="padding: 20px; background: #f5f5f5; margin-bottom: 20px;">
    <a href="index.html"%s style="margin-right: 20px; color: #0066cc; text-decoration: none;">Home</a>
    <a href="about.html"%s style="margin-right: 20px; color: #0066cc; text-decoration: none;">About</a>
    <a href="contact.html"%s style="color: #0066cc; text-decoration: none;">Contact</a>
</nav>`,
		pageClass("index"),
		pageClass("about"),
		pageClass("contact"),
	)

	// Insert after <body> tag
	bodyIdx := strings.Index(html, "<body>")
	if bodyIdx >= 0 {
		insertPoint := bodyIdx + 6 // len("<body>")
		html = html[:insertPoint] + "\n" + nav + "\n" + html[insertPoint:]
	}

	return html
}

// addNavigationToAllPages adds navigation to each page
func addNavigationToAllPages(pages map[string]string, logger *zap.Logger) map[string]string {
	result := make(map[string]string)

	for name, html := range pages {
		// Determine current page name (remove .html)
		currentPage := strings.TrimSuffix(name, ".html")
		result[name] = addSimpleNavigation(html, currentPage)

		logger.Debug("Added navigation to page",
			zap.String("page", name),
			zap.String("current", currentPage),
		)
	}

	return result
}

// generateAboutPage creates a standard about page
func generateAboutPage(domain string) string {
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
        h1 { margin-bottom: 20px; color: #333; }
        p { margin-bottom: 15px; color: #666; }
    </style>
</head>
<body>
    <h1>About %s</h1>
    <p>Learn more about our company and what we do.</p>
    <p>We're dedicated to providing quality solutions for our customers.</p>
</body>
</html>`, domain, domain)
}

// generateContactPage creates a standard contact page
func generateContactPage(domain string) string {
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
        h1 { margin-bottom: 20px; color: #333; }
        p { margin-bottom: 20px; color: #666; }
        form { margin-top: 30px; }
        label { display: block; margin: 15px 0 5px; font-weight: 500; }
        input, textarea { 
            width: 100%%; 
            padding: 10px; 
            border: 1px solid #ddd;
            border-radius: 4px;
            font-family: inherit;
        }
        button { 
            margin-top: 15px; 
            padding: 10px 24px; 
            background: #0066cc;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        button:hover { background: #0052a3; }
    </style>
</head>
<body>
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

// extractDomainFromData extracts domain from collected data
func extractDomainFromData(data map[string]interface{}) string {
	// Try input_data.domain first
	if inputData, ok := data["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok && domain != "" {
			return domain
		}
	}

	// Try direct domain field
	if domain, ok := data["domain"].(string); ok && domain != "" {
		return domain
	}

	// Fallback
	return "Our Company"
}

// extractNestedField navigates nested field paths like "step.result.html"
func extractNestedField(data map[string]interface{}, fieldPath string) interface{} {
	parts := strings.Split(fieldPath, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[part]; ok {
				current = val
				continue
			}
			// Try ExtractStepData for step results
			if extracted := datahelpers.ExtractStepData(v[part]); extracted != nil {
				current = extracted
				continue
			}
			return nil
		default:
			return nil
		}
	}

	return current
}

// upstreamDeclaredSkip reports whether the step result feeding contentField
// declared itself skipped — the legitimate no-content case. An unresolvable
// content_field WITHOUT that declaration is the misconfiguration signature:
// content was expected here (council 3918db52, bug_historian seat).
func upstreamDeclaredSkip(collectedData map[string]interface{}, contentField string) bool {
	top, ok := collectedData[strings.Split(contentField, ".")[0]].(map[string]interface{})
	if !ok {
		return false
	}
	if skipped, ok := top["skipped"].(bool); ok && skipped {
		return true
	}
	if response, ok := top["response"].(map[string]interface{}); ok {
		if skipped, ok := response["skipped"].(bool); ok && skipped {
			return true
		}
	}
	return false
}

// pathWalkOutcome classifies why a single walk over one candidate path ended.
// The three-way split is load-bearing (bugs_open/408): a missing key admits
// trying another candidate form, while a non-map met mid-path must end the
// lookup outright — collapsing the two would widen the old resolution
// semantics, which short-circuited on exactly that case.
type pathWalkOutcome int

const (
	pathWalkResolved       pathWalkOutcome = iota
	pathWalkKeyMissing                     // a map lacked the next segment — another candidate form may still resolve
	pathWalkNotTraversable                 // a non-map met mid-path — no further candidate is tried (matches the old behaviour)
)

// walkFieldPath resolves a dot-notation path against nested maps with a plain
// iterative walk — no fallbacks, no recursion (bugs_open/408: the previous
// shape here was two mutually-recursive fallbacks with no depth bound, and an
// unresolvable path crashed the pod with a stack overflow). It is log-free;
// it reports the segment that stopped it so the caller can log once.
func walkFieldPath(data map[string]interface{}, path string) (value interface{}, stoppedAt string, outcome pathWalkOutcome) {
	parts := strings.Split(path, ".")
	var current interface{} = data
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, part, pathWalkNotTraversable
		}
		val, ok := m[part]
		if !ok {
			return nil, part, pathWalkKeyMissing
		}
		current = val
	}
	return current, "", pathWalkResolved
}

// extractFieldValue navigates nested field paths like "base_structure.result"
// and extracts a string value, with fallbacks for common content keys.
// Returns the value, or "" if it cannot be found — the sole caller
// (AssemblePageAction) treats "" as "skip this page".
func extractFieldValue(data map[string]interface{}, fieldPath string, logger *zap.Logger) string {
	// Build the ordered candidate paths. This is precisely the sequence the
	// old mutually-recursive fallbacks tried before they began to cycle: the
	// original path; the path with each ".response." stripped in turn; and
	// finally the fully-stripped path with ".response." inserted after its
	// first segment (skipped when identical to the original). A bounded list,
	// walked in a loop — non-termination is unrepresentable (bugs_open/408).
	candidates := []string{fieldPath}
	stripped := fieldPath
	for strings.Contains(stripped, ".response.") {
		stripped = strings.Replace(stripped, ".response.", ".", 1)
		candidates = append(candidates, stripped)
	}
	if parts := strings.Split(stripped, "."); len(parts) >= 2 {
		withResponse := parts[0] + ".response." + strings.Join(parts[1:], ".")
		if withResponse != fieldPath {
			candidates = append(candidates, withResponse)
		}
	}

	var current interface{}
	resolved := false
	firstMissingSegment := ""
	for _, candidate := range candidates {
		value, stoppedAt, outcome := walkFieldPath(data, candidate)
		switch outcome {
		case pathWalkResolved:
			current = value
			resolved = true
		case pathWalkKeyMissing:
			if firstMissingSegment == "" {
				firstMissingSegment = stoppedAt
			}
			logger.Debug("extractFieldValue: candidate path did not resolve",
				zap.String("candidate_path", candidate),
				zap.String("field", stoppedAt),
				zap.String("original_path", fieldPath))
			continue
		case pathWalkNotTraversable:
			// Preserved from the old mid-walk default branch: a non-map value
			// part-way along a path ends the lookup outright.
			logger.Warn("Cannot traverse further, value is not a map",
				zap.String("field", stoppedAt),
				zap.String("full_path", candidate),
			)
			return ""
		}
		break
	}
	if !resolved {
		// One Warn per failed lookup — the old shape logged this line once per
		// recursion and produced 12,654 identical lines on the way to the
		// crash (bugs_open/408 §2).
		logger.Warn("Field not found in path",
			zap.String("field", firstMissingSegment),
			zap.String("full_path", fieldPath),
			zap.Strings("paths_tried", candidates),
		)
		return ""
	}

	// Convert final value to string
	switch v := current.(type) {
	case string:
		return v
	case map[string]interface{}:
		// If it's still a map, try common content field names
		// Order: most specific first, then generic
		contentKeys := []string{
			"result",    // LLM action output
			"page_html", // page-content-writer output
			"html",      // generic HTML
			"content",   // generic content
			"text",      // text content
			"markdown",  // markdown content
			"body",      // response body
		}

		for _, key := range contentKeys {
			if val, ok := v[key].(string); ok && val != "" {
				logger.Debug("extractFieldValue: found content in map",
					zap.String("full_path", fieldPath),
					zap.String("key_used", key),
				)
				return val
			}
		}

		// Log available keys to help debug
		availableKeys := make([]string, 0, len(v))
		for k := range v {
			availableKeys = append(availableKeys, k)
		}
		logger.Warn("Final value is a map but couldn't extract string",
			zap.String("full_path", fieldPath),
			zap.Strings("available_keys", availableKeys),
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

// calculateTotalSize calculates total bytes across all pages
func calculateTotalSize(pages map[string]string) int {
	total := 0
	for _, content := range pages {
		total += len(content)
	}
	return total
}

// getPageNames returns sorted list of page names
func getPageNames(pages map[string]string) []string {
	names := make([]string, 0, len(pages))
	for name := range pages {
		names = append(names, name)
	}
	return names
}

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

// HeaderConfig holds configuration for header generation
type HeaderConfig struct {
	LogoText     string
	LogoAccent   string
	LogoURL      string
	NavItems     []NavItem
	PrimaryColor string
	AccentColor  string
	CurrentPage  string
	IsHomePage   bool
}

// buildHeaderConfig extracts header configuration from collected data
func buildHeaderConfig(collectedData map[string]interface{}, currentPageName string, logger *zap.Logger) *HeaderConfig {
	config := &HeaderConfig{
		LogoText:     "Company",
		LogoAccent:   "",
		PrimaryColor: "#1a1a2e",
		AccentColor:  "#16a085",
		CurrentPage:  strings.TrimSuffix(currentPageName, ".html"),
		IsHomePage:   currentPageName == "index.html" || currentPageName == "home.html",
	}

	// Try to get domain/business name for logo
	if inputData, ok := collectedData["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok && domain != "" {
			// Extract business name from domain
			parts := strings.Split(domain, ".")
			if len(parts) > 0 {
				name := parts[0]
				// Capitalize first letter, handle common suffixes
				if len(name) > 0 {
					config.LogoText = datahelpers.UpperFirst(name)
				}
			}
		}

		// Try to get colors from reviewed_brief
		if brief, ok := inputData["reviewed_brief"].(map[string]interface{}); ok {
			if colors, ok := brief["color_scheme"].(string); ok && colors != "" {
				config.PrimaryColor, config.AccentColor = parseColorScheme(colors)
			}
		}
	}

	// Get navigation items
	config.NavItems = extractNavItemsForHeader(collectedData, config.CurrentPage, logger)

	// Get logo URL if available
	if logoURL := datahelpers.ExtractNestedFieldString(collectedData, "logo_url"); logoURL != "" {
		config.LogoURL = logoURL
	}

	return config
}

// parseColorScheme extracts primary and accent colors from a description
func parseColorScheme(scheme string) (primary, accent string) {
	primary = "#1a1a2e"
	accent = "#16a085"

	schemeLower := strings.ToLower(scheme)

	if strings.Contains(schemeLower, "dark") {
		primary = "#1a1a2e"
	}
	if strings.Contains(schemeLower, "navy") {
		primary = "#1e3a5f"
	}
	if strings.Contains(schemeLower, "teal") {
		accent = "#16a085"
	}
	if strings.Contains(schemeLower, "gold") {
		accent = "#d4af37"
	}
	if strings.Contains(schemeLower, "blue") {
		accent = "#2563eb"
	}
	if strings.Contains(schemeLower, "green") {
		accent = "#059669"
	}
	if strings.Contains(schemeLower, "purple") {
		accent = "#7c3aed"
	}

	return primary, accent
}

// extractNavItemsForHeader gets navigation items with simple labels
func extractNavItemsForHeader(collectedData map[string]interface{}, currentPage string, logger *zap.Logger) []NavItem {
	var items []NavItem

	// Priority 1: db_sync.navigation
	if dbSync, ok := collectedData["db_sync"].(map[string]interface{}); ok {
		if nav, ok := dbSync["navigation"].(map[string]interface{}); ok {
			if navItems, ok := nav["items"].([]interface{}); ok {
				for _, item := range navItems {
					if itemMap, ok := item.(map[string]interface{}); ok {
						label, _ := itemMap["label"].(string)
						url, _ := itemMap["url"].(string)

						if label != "" && url != "" {
							urlPage := strings.TrimSuffix(strings.TrimPrefix(url, "/"), ".html")
							isActive := urlPage == currentPage ||
								(currentPage == "index" && (urlPage == "home" || urlPage == "index"))

							items = append(items, NavItem{
								Label:    label, // Already simplified by buildNavigationFromPages
								URL:      url,
								IsActive: isActive,
							})
						}
					}
				}
			}
		}
	}

	// Fallback: default navigation
	if len(items) == 0 {
		logger.Warn("No navigation found, using defaults")
		items = []NavItem{
			{Label: "Home", URL: "/index.html", IsActive: currentPage == "index"},
			{Label: "About", URL: "/about.html", IsActive: currentPage == "about"},
			{Label: "Services", URL: "/services.html", IsActive: currentPage == "services"},
			{Label: "Contact", URL: "/contact.html", IsActive: currentPage == "contact"},
		}
	}

	return items
}

// generateConsistentHeader creates the header HTML
func generateConsistentHeader(config *HeaderConfig) string {
	var navLinks []string
	for _, item := range config.NavItems {
		activeClass := ""
		if item.IsActive {
			activeClass = ` class="active"`
		}
		navLinks = append(navLinks, fmt.Sprintf(
			`<li><a href="%s"%s>%s</a></li>`,
			item.URL, activeClass, item.Label,
		))
	}

	navHTML := strings.Join(navLinks, "\n                ")

	var logoHTML string
	if config.LogoURL != "" {
		logoHTML = fmt.Sprintf(`<img src="%s" alt="%s" class="logo-img">`,
			config.LogoURL, config.LogoText)
	} else if config.LogoAccent != "" {
		logoHTML = fmt.Sprintf(`<span class="logo-text">%s</span><span class="logo-accent">%s</span>`,
			config.LogoText, config.LogoAccent)
	} else {
		logoHTML = fmt.Sprintf(`<span class="logo-text">%s</span>`, config.LogoText)
	}

	return fmt.Sprintf(`<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">
            %s
        </a>
        <button class="mobile-menu-toggle" aria-label="Toggle menu">
            <span></span><span></span><span></span>
        </button>
        <nav class="main-nav">
            <ul>
                %s
            </ul>
        </nav>
    </div>
</header>`, logoHTML, navHTML)
}

// generateHeaderStyles creates CSS for the header
func generateHeaderStyles(config *HeaderConfig) string {
	return fmt.Sprintf(`
/* ========== CONSISTENT HEADER STYLES ========== */
.site-header {
    background: %s;
    padding: 1rem 0;
    position: sticky;
    top: 0;
    z-index: 1000;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}
.header-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.logo {
    text-decoration: none;
    font-size: 1.5rem;
    font-weight: 700;
    color: white;
}
.logo-accent { color: %s; }
.logo-img {
    max-height: 40px;
    width: auto;
    display: block;
}
.main-nav ul {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
    gap: 2rem;
}
.main-nav a {
    color: rgba(255,255,255,0.9);
    text-decoration: none;
    font-weight: 500;
    padding: 0.5rem 0;
    transition: color 0.2s;
}
.main-nav a:hover,
.main-nav a.active { color: %s; }
.mobile-menu-toggle {
    display: none;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
}
.mobile-menu-toggle span {
    display: block;
    width: 24px;
    height: 2px;
    background: white;
    margin: 5px 0;
}
@media (max-width: 768px) {
    .mobile-menu-toggle { display: block; }
    .main-nav {
        position: absolute;
        top: 100%%;
        left: 0;
        right: 0;
        background: %s;
        padding: 1rem;
        display: none;
    }
    .main-nav.active { display: block; }
    .main-nav ul { flex-direction: column; gap: 0; }
    .main-nav a { display: block; padding: 0.75rem 0; border-bottom: 1px solid rgba(255,255,255,0.1); }
}
/* ========== END HEADER STYLES ========== */
`, config.PrimaryColor, config.AccentColor, config.AccentColor, config.PrimaryColor)
}

// injectConsistentHeader replaces the existing header with a consistent one
func injectConsistentHeader(html string, config *HeaderConfig, logger *zap.Logger) string {
	headerHTML := generateConsistentHeader(config)
	headerCSS := generateHeaderStyles(config)

	// Step 1: Remove existing header element
	headerRe := regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`)
	html = headerRe.ReplaceAllString(html, "<!-- HEADER_PLACEHOLDER -->")

	// Step 2: Insert new header after <body> tag
	bodyRe := regexp.MustCompile(`(?i)(<body[^>]*>)`)
	if bodyRe.MatchString(html) {
		html = bodyRe.ReplaceAllString(html, "$1\n"+headerHTML)
		html = strings.Replace(html, "<!-- HEADER_PLACEHOLDER -->", "", 1)
	} else {
		html = strings.Replace(html, "<!-- HEADER_PLACEHOLDER -->", headerHTML, 1)
	}

	// Step 3: Inject CSS into <style> or <head>
	styleRe := regexp.MustCompile(`(?i)(</style>)`)
	if styleRe.MatchString(html) {
		// Insert before first </style>
		replaced := false
		html = styleRe.ReplaceAllStringFunc(html, func(match string) string {
			if !replaced {
				replaced = true
				return headerCSS + "\n" + match
			}
			return match
		})
	} else {
		// No style tag, add one in head
		headCloseRe := regexp.MustCompile(`(?i)(</head>)`)
		if headCloseRe.MatchString(html) {
			html = headCloseRe.ReplaceAllString(html, "<style>"+headerCSS+"</style>\n$1")
		}
	}

	// Step 4: Add mobile menu JS
	if !strings.Contains(html, "mobile-menu-toggle") || !strings.Contains(strings.ToLower(html), "addeventlistener") {
		mobileJS := `<script>
document.addEventListener('DOMContentLoaded', function() {
    var toggle = document.querySelector('.mobile-menu-toggle');
    var nav = document.querySelector('.main-nav');
    if (toggle && nav) {
        toggle.addEventListener('click', function() {
            nav.classList.toggle('active');
        });
    }
});
</script>`
		bodyCloseRe := regexp.MustCompile(`(?i)(</body>)`)
		if bodyCloseRe.MatchString(html) {
			html = bodyCloseRe.ReplaceAllString(html, mobileJS+"\n$1")
		}
	}

	logger.Debug("Injected consistent header",
		zap.Int("nav_items", len(config.NavItems)),
		zap.String("primary_color", config.PrimaryColor),
	)

	return html
}

// checkUpstreamContentFailure checks if the content generation step failed
// by examining the response status/error fields in collected data
func checkUpstreamContentFailure(collectedData map[string]interface{}, contentField string, logger *zap.Logger) (bool, string) {
	// contentField is like "page_content.response.page_html"
	// We need to check "page_content.response.status" and "page_content.response.error"

	// Parse the field path to get the parent (e.g., "page_content.response")
	parts := strings.Split(contentField, ".")
	if len(parts) < 2 {
		return false, ""
	}

	// Get top-level key (e.g., "page_content")
	topKey := parts[0]

	// Check top-level object
	topData, ok := collectedData[topKey].(map[string]interface{})
	if !ok {
		logger.Debug("Top-level key not found or not a map",
			zap.String("key", topKey))
		return false, ""
	}

	// Check for response object
	response, ok := topData["response"].(map[string]interface{})
	if !ok {
		logger.Debug("No response object in top-level data",
			zap.String("key", topKey))
		return false, ""
	}

	// Check status field
	if status, ok := response["status"].(string); ok && status == "failed" {
		errorMsg := "content generation failed"
		if errStr, ok := response["error"].(string); ok && errStr != "" {
			errorMsg = errStr
		}
		logger.Info("Detected failed status in upstream response",
			zap.String("top_key", topKey),
			zap.String("status", status),
			zap.String("error", errorMsg))
		return true, errorMsg
	}

	// Check for error field directly
	if errStr, ok := response["error"].(string); ok && errStr != "" {
		logger.Info("Detected error in upstream response",
			zap.String("top_key", topKey),
			zap.String("error", errStr))
		return true, errStr
	}

	return false, ""
}
