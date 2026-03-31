// FILE: platform/orchestration/actions/apply_adoption_plan_action.go
//
// ApplyAdoptionPlanAction takes the LLM's structural analysis of a crawled site
// and creates everything needed to rebuild it: site_specs, page records, and
// work items.
//
// The LLM provides structure (pages, identity, design) — lightweight classification.
// Content extraction is done here in Go from the raw crawl markdown — no LLM needed,
// no token limits, deterministic.
//
// Registration:
//   "apply_adoption_plan": {
//       Handler:     ApplyAdoptionPlanAction,
//       Category:    "site",
//       Description: "Create specs, pages, and work items from site adoption analysis",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ApplyAdoptionPlanInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id", "domain"},
	Optional:   []string{"adoption_plan"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("apply_adoption_plan", ApplyAdoptionPlanInputSpec)
}

func ApplyAdoptionPlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "apply_adoption_plan"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		ApplyAdoptionPlanInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	domain := inputs.Get("domain")

	// ── Parse adoption plan (LLM output) ────────────────────────────────
	planField := "adoption_analysis"
	if pf, ok := params.StepConfig.Config["adoption_plan"].(string); ok && pf != "" {
		planField = pf
	}

	planRaw := datahelpers.ExtractNestedField(params.CollectedData, planField)
	if planRaw == nil {
		for _, alt := range []string{"adoption_analysis", "adoption_plan", "analyze_site"} {
			planRaw = datahelpers.ExtractNestedField(params.CollectedData, alt)
			if planRaw != nil {
				break
			}
		}
	}
	if planRaw == nil {
		return nil, fmt.Errorf("adoption plan not found — check adoption_plan config path")
	}

	// Use existing UnwrapDeep to handle {"type":"text","result":"...json..."}
	// This handles both map results and JSON string results
	unwrapped := datahelpers.UnwrapDeep(planRaw, logger)

	var plan map[string]interface{}
	switch v := unwrapped.(type) {
	case map[string]interface{}:
		plan = v
	case string:
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
		if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
			// JSON is truncated (max_tokens hit). Try to repair by closing brackets.
			logger.Warn("ApplyAdoptionPlanAction: JSON parse failed, attempting repair",
				zap.Error(err),
				zap.Int("length", len(cleaned)))
			repaired := repairTruncatedJSON(cleaned)
			if repaired != "" {
				if err2 := json.Unmarshal([]byte(repaired), &plan); err2 == nil {
					logger.Info("ApplyAdoptionPlanAction: repaired truncated JSON successfully")
				} else {
					return nil, fmt.Errorf("failed to parse adoption plan JSON (repair also failed): %w (tail: %s)",
						err, func() string {
							if len(cleaned) > 100 {
								return cleaned[len(cleaned)-100:]
							}
							return cleaned
						}())
				}
			} else {
				return nil, fmt.Errorf("failed to parse adoption plan JSON: %w (tail: %s)",
					err, func() string {
						if len(cleaned) > 100 {
							return cleaned[len(cleaned)-100:]
						}
						return cleaned
					}())
			}
		}
	default:
		return nil, fmt.Errorf("unexpected adoption plan type after unwrap: %T", unwrapped)
	}

	logger.Info("ApplyAdoptionPlanAction: parsed plan",
		zap.String("site_id", siteIDStr),
		zap.String("domain", domain),
		zap.Strings("plan_keys", func() []string {
			keys := make([]string, 0, len(plan))
			for k := range plan {
				keys = append(keys, k)
			}
			return keys
		}()),
	)

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	batchID := uuid.New()
	specsWritten := 0
	pagesCreated := 0
	itemsCreated := 0

	// ── 1. Store full analysis in research_results ──────────────────────

	planJSON, _ := json.Marshal(plan)
	var researchID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO research_results (
			site_id, query, topic, result_type,
			findings, data, summary,
			researched_by, research_agent_type
		) VALUES (
			$1, $2, 'adoption', 'adoption_crawl',
			$3::jsonb, $3::jsonb, $4,
			'site-adoption-agent', 'site-adoption-agent'
		) RETURNING id
	`, siteID,
		fmt.Sprintf("Adoption crawl of %s", domain),
		string(planJSON),
		fmt.Sprintf("Site adoption analysis: %s", domain),
	).Scan(&researchID)
	if err != nil {
		logger.Warn("Failed to store adoption research", zap.Error(err))
	} else {
		logger.Info("Adoption research stored",
			zap.String("research_id", researchID.String()))
	}

	// ── 2. Write site_specs ─────────────────────────────────────────────

	identityData, _ := plan["identity"].(map[string]interface{})
	if identityData == nil {
		identityData = make(map[string]interface{})
	}
	identityData["adopted_from"] = domain
	identityData["adopted_at"] = time.Now().UTC().Format(time.RFC3339)

	specAspects := map[string]interface{}{
		"identity": identityData,
		"design":   plan["design"],
	}

	// Build structure spec from pages
	if pages, ok := plan["pages"].([]interface{}); ok {
		pageNames := make([]string, 0, len(pages))
		for _, p := range pages {
			if pm, ok := p.(map[string]interface{}); ok {
				if name, ok := pm["name"].(string); ok {
					pageNames = append(pageNames, name)
				}
			}
		}
		specAspects["structure"] = map[string]interface{}{
			"pages":        pageNames,
			"source":       "adoption",
			"adopted_from": domain,
		}
	}

	for aspect, data := range specAspects {
		if data == nil {
			continue
		}

		dataJSON, err := json.Marshal(data)
		if err != nil {
			logger.Warn("Failed to marshal spec",
				zap.String("aspect", aspect), zap.Error(err))
			continue
		}

		_, _ = tx.ExecContext(ctx, `
			UPDATE site_specs SET is_current = false, superseded_at = now()
			WHERE site_id = $1 AND aspect = $2 AND is_current = true
		`, siteID, aspect)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_specs (
				site_id, aspect, data, source, source_agent,
				is_current, created_by, notes
			) VALUES (
				$1, $2, $3::jsonb, 'adoption', 'site-adoption-agent',
				true, 'site-adoption-agent', $4
			)
		`, siteID, aspect, string(dataJSON),
			fmt.Sprintf("Adopted from %s", domain))

		if err != nil {
			logger.Warn("Failed to write spec",
				zap.String("aspect", aspect), zap.Error(err))
			continue
		}

		specsWritten++
		logger.Info("Spec written", zap.String("aspect", aspect))
	}

	// ── 3. Build crawl page index for content extraction ────────────────
	// Extract existing_content directly from crawl markdown (Go, no LLM)

	crawlPages := buildCrawlPageIndex(params.CollectedData, logger)

	// ── 4. Create page records ──────────────────────────────────────────

	pagesRaw, _ := plan["pages"].([]interface{})

	type adoptedPage struct {
		ID       uuid.UUID
		Name     string
		PageType string
	}
	var adoptedPages []adoptedPage

	for i, pageRaw := range pagesRaw {
		pm, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		pageName, _ := pm["name"].(string)
		if pageName == "" {
			pageName = fmt.Sprintf("page-%d", i+1)
		}
		pageName = strings.ToLower(strings.ReplaceAll(pageName, " ", "-"))

		pageTitle, _ := pm["title"].(string)
		if pageTitle == "" {
			pageTitle = strings.Title(strings.ReplaceAll(pageName, "-", " "))
		}

		pageType, _ := pm["page_type"].(string)
		if pageType == "" {
			pageType = "content"
		}

		pageURL, _ := pm["url"].(string)
		if pageURL == "" {
			switch pageType {
			case "blog-post":
				pageURL = "/blog/" + pageName + ".html"
			case "tool":
				pageURL = "/tools/" + pageName + ".html"
			default:
				pageURL = "/" + pageName + ".html"
			}
		}

		metaDesc, _ := pm["meta_description"].(string)

		// Parse sections from LLM plan
		sections := []string{}
		if sectionsRaw, ok := pm["sections"].([]interface{}); ok {
			for _, s := range sectionsRaw {
				if ss, ok := s.(string); ok {
					sections = append(sections, ss)
				}
			}
		}
		sectionsJSON, _ := json.Marshal(sections)

		navLabel, _ := pm["nav_label"].(string)
		inHeader := true
		if ih, ok := pm["in_header"].(bool); ok {
			inHeader = ih
		}
		inFooter := true
		if inf, ok := pm["in_footer"].(bool); ok {
			inFooter = inf
		}

		var pageID uuid.UUID
		err := tx.QueryRowContext(ctx, `
			INSERT INTO pages (
				site_id, name, url, title, page_type, build_status,
				sections, meta_description, nav_label, in_header, in_footer
			) VALUES ($1, $2, $3, $4, $5, 'planned',
			          $6::jsonb, $7, $8, $9, $10)
			ON CONFLICT (site_id, name) DO UPDATE SET
				title = EXCLUDED.title,
				url = EXCLUDED.url,
				page_type = EXCLUDED.page_type,
				sections = EXCLUDED.sections,
				meta_description = EXCLUDED.meta_description,
				updated_at = NOW()
			RETURNING id
		`, siteID, pageName, pageURL, pageTitle, pageType,
			string(sectionsJSON), metaDesc, navLabel, inHeader, inFooter,
		).Scan(&pageID)

		if err != nil {
			logger.Warn("Failed to create page",
				zap.String("name", pageName), zap.Error(err))
			continue
		}

		// Extract existing_content from crawl markdown (Go, not LLM)
		// Match by URL — the LLM told us the URL, the crawl has markdown per URL
		existingContent := pm["existing_content"] // LLM may have included some
		if existingContent == nil {
			// Try to extract from crawl data by matching URL
			if crawlMarkdown, found := crawlPages[pageURL]; found {
				existingContent = map[string]interface{}{
					"raw_markdown": crawlMarkdown,
				}
			} else if crawlMarkdown, found := crawlPages[domain+pageURL]; found {
				existingContent = map[string]interface{}{
					"raw_markdown": crawlMarkdown,
				}
			}
		}

		if existingContent != nil {
			pageContentJSON, _ := json.Marshal(map[string]interface{}{
				"page_name":        pageName,
				"page_id":          pageID.String(),
				"existing_content": existingContent,
				"adopted_from":     domain,
				"mode":             "recreate",
			})

			_, _ = tx.ExecContext(ctx, `
				INSERT INTO research_results (
					site_id, page_id, query, topic, result_type,
					data, summary,
					researched_by, research_agent_type
				) VALUES (
					$1, $2, $3, 'adoption_page_content', 'adoption_page',
					$4::jsonb, $5,
					'site-adoption-agent', 'site-adoption-agent'
				)
			`, siteID, pageID,
				fmt.Sprintf("Existing content for %s on %s", pageName, domain),
				string(pageContentJSON),
				fmt.Sprintf("Adopted page content: %s", pageName),
			)
		}

		adoptedPages = append(adoptedPages, adoptedPage{
			ID:       pageID,
			Name:     pageName,
			PageType: pageType,
		})
		pagesCreated++
	}

	// ── 5. Create work items ────────────────────────────────────────────

	designSpec := map[string]interface{}{
		"mode":         "recreate",
		"adopted_from": domain,
	}
	if designData, ok := plan["design"].(map[string]interface{}); ok {
		designSpec["adopt_from"] = designData
	}
	designSpecJSON, _ := json.Marshal(designSpec)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, created_by,
			item_key, batch_id
		) VALUES ($1, 'adoption', 'build', 'needs_design', 'high',
		          $2, $3::jsonb, 8, 'webdesign-agent', 'triaged',
		          'site-adoption-agent', $4, $5)
		ON CONFLICT DO NOTHING
	`, siteID,
		fmt.Sprintf("Generate stylesheet matching %s design", domain),
		string(designSpecJSON),
		fmt.Sprintf("adoption_design_%s", siteID),
		batchID)
	if err == nil {
		itemsCreated++
	}

	for i, page := range adoptedPages {
		pageSpec := map[string]interface{}{
			"page_name": page.Name,
			"page_type": page.PageType,
			"mode":      "recreate",
			"source":    "adoption",
		}
		pageSpecJSON, _ := json.Marshal(pageSpec)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by,
				item_key, batch_id
			) VALUES ($1, 'adoption', 'build', 'needs_content_page', 'medium',
			          $2, $3::jsonb, $4, $5, 'page-build-handler', 'triaged',
			          'site-adoption-agent', $6, $7)
			ON CONFLICT DO NOTHING
		`, siteID,
			fmt.Sprintf("Recreate %s page from %s", page.Name, domain),
			string(pageSpecJSON),
			page.ID,
			10+i,
			fmt.Sprintf("adoption_page_%s_%s", page.Name, siteID),
			batchID)
		if err == nil {
			itemsCreated++
		}
	}

	rerenderSpec, _ := json.Marshal(map[string]interface{}{
		"refresh_site_components": true,
		"reason":                  "post_adoption_assembly",
	})

	_, err = tx.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, created_by,
			item_key, batch_id
		) VALUES ($1, 'adoption', 'build', 'needs_rerender', 'medium',
		          'Assemble and deploy all pages after adoption build',
		          $2::jsonb, 99, 'rerender-pages', 'triaged',
		          'site-adoption-agent', $3, $4)
		ON CONFLICT DO NOTHING
	`, siteID, string(rerenderSpec),
		fmt.Sprintf("adoption_rerender_%s", siteID),
		batchID)
	if err == nil {
		itemsCreated++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("ApplyAdoptionPlanAction: Complete",
		zap.String("domain", domain),
		zap.Int("specs_written", specsWritten),
		zap.Int("pages_created", pagesCreated),
		zap.Int("items_created", itemsCreated),
		zap.String("batch_id", batchID.String()),
		zap.String("research_id", researchID.String()),
	)

	return map[string]interface{}{
		"applied":       true,
		"domain":        domain,
		"specs_written": specsWritten,
		"pages_created": pagesCreated,
		"items_created": itemsCreated,
		"batch_id":      batchID.String(),
		"research_id":   researchID.String(),
	}, nil
}

// buildCrawlPageIndex extracts markdown content from the raw crawl response
// keyed by URL path. This avoids needing the LLM to echo back content.
func buildCrawlPageIndex(collectedData map[string]interface{}, logger *zap.Logger) map[string]string {
	index := make(map[string]string)

	// Try multiple paths to find crawl pages
	paths := []string{
		"crawl_result.pages",
		"crawl_result.response.data.pages",
		"crawl_result.body.data.pages",
		"crawl_result.data.pages",
	}

	var pages []interface{}
	for _, path := range paths {
		raw := datahelpers.ExtractNestedField(collectedData, path)
		if raw != nil {
			if arr, ok := raw.([]interface{}); ok && len(arr) > 0 {
				pages = arr
				break
			}
		}
	}

	for _, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		markdown, _ := page["markdown"].(string)
		if markdown == "" {
			continue
		}

		// Skip error pages
		if strings.Contains(markdown, "NoSuchKey") {
			continue
		}

		metadata, _ := page["metadata"].(map[string]interface{})
		pageURL, _ := metadata["url"].(string)
		sourceURL, _ := metadata["sourceURL"].(string)
		statusCode, _ := metadata["statusCode"].(float64)

		if statusCode != 0 && statusCode != 200 {
			continue
		}

		// Index by multiple URL forms for matching
		if pageURL != "" {
			index[pageURL] = markdown
			// Also index by path only (strip domain)
			for _, prefix := range []string{"https://", "http://"} {
				if strings.HasPrefix(pageURL, prefix) {
					afterScheme := pageURL[len(prefix):]
					slashIdx := strings.Index(afterScheme, "/")
					if slashIdx >= 0 {
						index[afterScheme[slashIdx:]] = markdown
					}
				}
			}
		}
		if sourceURL != "" && sourceURL != pageURL {
			index[sourceURL] = markdown
		}
	}

	logger.Info("Built crawl page index",
		zap.Int("indexed_urls", len(index)),
		zap.Int("raw_pages", len(pages)))

	return index
}

// repairTruncatedJSON attempts to fix JSON that was cut off by max_tokens.
// Strategy: find the last complete JSON element (ending with }, ], or a quoted string
// that's a value not a key), strip the incomplete tail, close open brackets/braces.
// This salvages identity, design, and pages even if interactive_features was truncated.
func repairTruncatedJSON(s string) string {
	if len(s) == 0 {
		return ""
	}

	// Find the last position where a complete JSON object or array element ends.
	// We look for } or ] which definitively close a complete structure.
	lastGood := -1
	for i := len(s) - 1; i >= 0; i-- {
		ch := s[i]
		if ch == '}' || ch == ']' {
			lastGood = i + 1
			break
		}
	}

	if lastGood <= 0 {
		return ""
	}

	truncated := s[:lastGood]

	// Strip trailing commas after the last complete element
	truncated = strings.TrimRight(truncated, " \t\n\r,")

	// Count open vs close brackets and braces
	openBraces := strings.Count(truncated, "{") - strings.Count(truncated, "}")
	openBrackets := strings.Count(truncated, "[") - strings.Count(truncated, "]")

	// Close them (brackets first, then braces — innermost first)
	for openBrackets > 0 {
		truncated += "]"
		openBrackets--
	}
	for openBraces > 0 {
		truncated += "}"
		openBraces--
	}

	return truncated
}
