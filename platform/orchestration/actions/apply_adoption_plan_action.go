// FILE: platform/orchestration/actions/apply_adoption_plan_action.go
//
// ApplyAdoptionPlanAction takes the LLM's structured analysis of a crawled site
// and creates everything needed to rebuild it: site_specs, page records, and
// work items. This is the core "adoption" action.
//
// The LLM analysis step (execute_llm_prompt) runs before this action and
// produces structured JSON describing the site's identity, design, pages,
// and sections. This action turns that analysis into database records.
//
// Registration:
//   "apply_adoption_plan": {
//       Handler:     ApplyAdoptionPlanAction,
//       Category:    "site",
//       Description: "Create specs, pages, and work items from site adoption analysis",
//       IsLocal:     true,
//   },
//
// Data inputs (via ActionInputSpec):
//   - site_id (required)
//   - domain (required)
//   - adoption_plan (required) — structured LLM output
//
// The adoption_plan is expected to be a JSON object with:
//   {
//     "identity": { "company_name": "...", "tagline": "...", "industry": "...", ... },
//     "design": { "palette": {...}, "typography": {...}, "tone": "..." },
//     "pages": [
//       {
//         "name": "index",
//         "title": "Page Title",
//         "page_type": "content",
//         "url": "/index.html",
//         "sections": ["hero", "features", "call-to-action"],
//         "existing_content": { "hero": "...", "features": "...", ... },
//         "meta_description": "..."
//       }
//     ]
//   }

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

	// ── Parse adoption plan ─────────────────────────────────────────────
	planField := "adoption_analysis"
	if pf, ok := params.StepConfig.Config["adoption_plan"].(string); ok && pf != "" {
		planField = pf
	}

	planRaw := datahelpers.ExtractNestedField(params.CollectedData, planField)
	if planRaw == nil {
		// Try fallbacks
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

	var plan map[string]interface{}
	switch v := planRaw.(type) {
	case map[string]interface{}:
		plan = v
	case string:
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
		if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
			return nil, fmt.Errorf("failed to parse adoption plan JSON: %w", err)
		}
	default:
		return nil, fmt.Errorf("unexpected adoption plan type: %T", planRaw)
	}

	logger.Info("ApplyAdoptionPlanAction: parsed plan",
		zap.String("site_id", siteIDStr),
		zap.String("domain", domain),
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

	// ── 1. Write site_specs ─────────────────────────────────────────────

	specAspects := map[string]interface{}{
		"identity":        plan["identity"],
		"design":          plan["design"],
		"adoption_source": plan, // store the full crawl analysis for reference
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

		// Supersede existing
		_, _ = tx.ExecContext(ctx, `
			UPDATE site_specs SET is_current = false, superseded_at = now()
			WHERE site_id = $1 AND aspect = $2 AND is_current = true
		`, siteID, aspect)

		// Insert new
		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_specs (
				site_id, aspect, data, source, source_agent,
				is_current, created_by, notes
			) VALUES (
				$1, $2, $3::jsonb, 'adoption', 'site-adoption-agent',
				true, 'site-adoption-agent',
				$4
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

	// ── 2. Create page records ──────────────────────────────────────────

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

		// Parse sections
		sections := []string{}
		if sectionsRaw, ok := pm["sections"].([]interface{}); ok {
			for _, s := range sectionsRaw {
				if ss, ok := s.(string); ok {
					sections = append(sections, ss)
				}
			}
		}
		sectionsJSON, _ := json.Marshal(sections)

		// Navigation hints
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

		// Store existing_content in site_specs as page-level spec
		if existingContent, ok := pm["existing_content"]; ok && existingContent != nil {
			pageSpecData := map[string]interface{}{
				"existing_content": existingContent,
				"adopted_from":     domain,
				"mode":             "recreate",
			}
			pageSpecJSON, _ := json.Marshal(pageSpecData)

			specAspect := fmt.Sprintf("page_content_%s", pageName)
			_, _ = tx.ExecContext(ctx, `
				UPDATE site_specs SET is_current = false, superseded_at = now()
				WHERE site_id = $1 AND aspect = $2 AND is_current = true
			`, siteID, specAspect)

			_, _ = tx.ExecContext(ctx, `
				INSERT INTO site_specs (
					site_id, aspect, data, source, source_agent,
					is_current, created_by, notes
				) VALUES ($1, $2, $3::jsonb, 'adoption', 'site-adoption-agent',
				          true, 'site-adoption-agent', $4)
			`, siteID, specAspect, string(pageSpecJSON),
				fmt.Sprintf("Existing content from %s/%s", domain, pageName))
		}

		adoptedPages = append(adoptedPages, adoptedPage{
			ID:       pageID,
			Name:     pageName,
			PageType: pageType,
		})
		pagesCreated++
	}

	// ── 3. Create work items ────────────────────────────────────────────

	// Design item — generates CSS from extracted palette/typography
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

	// Content page items — one per page
	for _, page := range adoptedPages {
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
			10+pagesCreated, // priority: pages build in order
			fmt.Sprintf("adoption_page_%s_%s", page.Name, siteID),
			batchID)
		if err == nil {
			itemsCreated++
		}
	}

	// Rerender item — runs after all pages are built
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
	)

	return map[string]interface{}{
		"applied":       true,
		"domain":        domain,
		"specs_written": specsWritten,
		"pages_created": pagesCreated,
		"items_created": itemsCreated,
		"batch_id":      batchID.String(),
		"adopted_pages": adoptedPages,
	}, nil
}
