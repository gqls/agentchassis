// FILE: platform/orchestration/actions/apply_adoption_plan_action.go
//
// ApplyAdoptionPlanAction takes the LLM's structural analysis of a crawled site
// and creates everything needed to rebuild it: site_specs, page records, and
// work items. All pages get rawHtml stored when available. Pages with interactive
// features carry those descriptions in the work item spec.
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
	"database/sql"
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

// crawlPageContent holds both markdown and rawHtml for a crawled page
type crawlPageContent struct {
	Markdown string
	RawHTML  string
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
			logger.Warn("ApplyAdoptionPlanAction: JSON parse failed, attempting repair",
				zap.Error(err), zap.Int("length", len(cleaned)))
			repaired := repairTruncatedJSON(cleaned)
			if repaired != "" {
				if err2 := json.Unmarshal([]byte(repaired), &plan); err2 == nil {
					logger.Info("ApplyAdoptionPlanAction: repaired truncated JSON")
				} else {
					return nil, fmt.Errorf("adoption plan JSON parse failed (repair also failed): %w (tail: %s)",
						err, truncateTail(cleaned, 100))
				}
			} else {
				return nil, fmt.Errorf("adoption plan JSON parse failed: %w (tail: %s)",
					err, truncateTail(cleaned, 100))
			}
		}
	default:
		return nil, fmt.Errorf("unexpected adoption plan type after unwrap: %T", unwrapped)
	}

	logger.Info("ApplyAdoptionPlanAction: parsed plan",
		zap.String("site_id", siteIDStr),
		zap.String("domain", domain),
		zap.Strings("plan_keys", mapKeys(plan)),
	)

	// ── Resolve source reference ───────────────────────────────────────
	// The adoption flow may separate the crawled source URL from the
	// destination domain being built. When that happens:
	//   - `domain` (above, from apply_plan.config.domain → site_record.domain)
	//      is the DESTINATION.
	//   - sourceURL and sourceDomain describe the site that was CRAWLED.
	// For legacy callers where source == destination, sourceURL fills
	// from input_data.url and sourceDomain == domain (same value).
	sourceURL := ""
	if srcField, ok := params.StepConfig.Config["source_url_field"].(string); ok && srcField != "" {
		sourceURL = datahelpers.ExtractNestedFieldString(params.CollectedData, srcField)
	}
	if sourceURL == "" {
		sourceURL = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.target_url")
	}
	if sourceURL == "" {
		sourceURL = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.url")
	}

	// Derive the host portion for URL matching. cleanDomain strips
	// scheme/www/trailing-slash; we additionally strip any path after
	// the host so "https://example.com/some/page" -> "example.com".
	sourceDomain := domain // legacy default — identical to destination
	if sourceURL != "" {
		stripped := cleanDomain(sourceURL)
		if idx := strings.Index(stripped, "/"); idx >= 0 {
			stripped = stripped[:idx]
		}
		if stripped != "" {
			sourceDomain = stripped
		}
	}

	if sourceDomain != domain {
		logger.Info("ApplyAdoptionPlanAction: source differs from destination",
			zap.String("source_url", sourceURL),
			zap.String("source_domain", sourceDomain),
			zap.String("destination_domain", domain),
		)
	}

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
	}

	// ── 2. Write site_specs ─────────────────────────────────────────────

	identityData, _ := plan["identity"].(map[string]interface{})
	if identityData == nil {
		identityData = make(map[string]interface{})
	}
	identityData["adopted_from"] = sourceDomain
	identityData["adopted_at"] = time.Now().UTC().Format(time.RFC3339)

	specAspects := map[string]interface{}{
		"identity": identityData,
	}

	// Design reference: prefer concrete fingerprint data over LLM guess.
	// The fingerprint is extracted from actual CSS in the crawled HTML by
	// extract_design_fingerprint (Go action, no LLM). The plan["design"]
	// is the LLM's vague description from analyze_site. Fingerprint wins.
	fingerprintRaw := datahelpers.ExtractNestedField(params.CollectedData, "design_fingerprint")
	if fingerprintRaw != nil {
		fpUnwrapped := datahelpers.UnwrapDeep(fingerprintRaw, logger)
		if fpData, ok := fpUnwrapped.(map[string]interface{}); ok {
			if status, _ := fpData["status"].(string); status == "extracted" {
				// Merge the LLM's design description in as supplementary context
				// (it has useful things like "visual_tone" that CSS can't express)
				if llmDesign, ok := plan["design"].(map[string]interface{}); ok {
					fpData["llm_description"] = llmDesign
				}
				specAspects["design_reference"] = fpData
				logger.Info("Writing design_reference from fingerprint",
					zap.Int("css_vars", countMapKeys(fpData, "css_variables")),
					zap.Int("suggested", countMapKeys(fpData, "suggested_mapping")),
				)
			} else {
				// Fingerprint extraction didn't produce results — fall back to LLM
				specAspects["design_reference"] = plan["design"]
				logger.Info("Fingerprint status not 'extracted', using LLM design as reference")
			}
		}
	} else {
		// No fingerprint at all — fall back to LLM design
		specAspects["design_reference"] = plan["design"]
		logger.Info("No fingerprint available, using LLM design as reference")
	}

	// Content direction from the writing style analysis (separate LLM call)
	directionRaw := datahelpers.ExtractNestedField(params.CollectedData, "content_direction_analysis")
	if directionRaw != nil {
		directionUnwrapped := datahelpers.UnwrapDeep(directionRaw, logger)
		var directionData map[string]interface{}

		switch d := directionUnwrapped.(type) {
		case map[string]interface{}:
			directionData = d
		case string:
			cleaned := strings.TrimSpace(d)
			cleaned = strings.TrimPrefix(cleaned, "```json")
			cleaned = strings.TrimPrefix(cleaned, "```")
			cleaned = strings.TrimSuffix(cleaned, "```")
			cleaned = strings.TrimSpace(cleaned)
			if err := json.Unmarshal([]byte(cleaned), &directionData); err != nil {
				repaired := repairTruncatedJSON(cleaned)
				if repaired != "" {
					if err2 := json.Unmarshal([]byte(repaired), &directionData); err2 == nil {
						logger.Info("Content direction JSON repaired")
					} else {
						logger.Warn("Failed to parse content_direction JSON", zap.Error(err2))
					}
				}
			}
		}

		if directionData != nil {
			// Format the entire structured spec into one readable text block.
			// Uses shared datahelpers formatter — same function write_site_spec calls.
			formatted := datahelpers.FormatContentDirection(directionData)
			if formatted != "" {
				directionData["formatted"] = formatted
			}

			specAspects["content_direction"] = directionData
			logger.Info("Content direction spec written",
				zap.Int("formatted_len", len(formatted)),
				zap.Int("spec_keys", len(directionData)))
		}
	}

	// Site archetype from the classification analysis (separate LLM call)
	archetypeRaw := datahelpers.ExtractNestedField(params.CollectedData, "site_archetype_analysis")
	if archetypeRaw != nil {
		archetypeUnwrapped := datahelpers.UnwrapDeep(archetypeRaw, logger)
		var archetypeData map[string]interface{}

		switch d := archetypeUnwrapped.(type) {
		case map[string]interface{}:
			archetypeData = d
		case string:
			cleaned := strings.TrimSpace(d)
			cleaned = strings.TrimPrefix(cleaned, "```json")
			cleaned = strings.TrimPrefix(cleaned, "```")
			cleaned = strings.TrimSuffix(cleaned, "```")
			cleaned = strings.TrimSpace(cleaned)
			if err := json.Unmarshal([]byte(cleaned), &archetypeData); err != nil {
				repaired := repairTruncatedJSON(cleaned)
				if repaired != "" {
					if err2 := json.Unmarshal([]byte(repaired), &archetypeData); err2 == nil {
						logger.Info("Site archetype JSON repaired")
					} else {
						logger.Warn("Failed to parse site_archetype_analysis JSON", zap.Error(err2))
					}
				}
			}
		}

		if archetypeData != nil {
			specAspects["site_archetype"] = archetypeData
			label, _ := archetypeData["label"].(string)
			logger.Info("Site archetype spec will be written",
				zap.String("label", label),
				zap.Int("spec_keys", len(archetypeData)))
		}
	}

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
			"adopted_from": sourceDomain,
		}
	}

	// The mission spec is written only if input_data carries a mission object
	// with at least one recognised field. Empty or absent input_data.mission
	// leaves the spec unwritten — adopted sites without explicit mission
	// hints get the identical behaviour as before.

	// Mission spec — optional; only written if input_data carries mission
	// hints. The shape is free-form (forward-compatible with future fields)
	// but recognises preferred_palette, preferred_typography, tone,
	// objective — the fields site-design-planner looks for in its cascade.
	missionRaw := datahelpers.ExtractNestedField(params.CollectedData, "input_data.mission")
	if missionRaw == nil {
		// Also check for top-level objective / direction if no nested mission.
		// POST /sites passes {"objective": "...", "direction": {...}} and
		// domain-submitter already writes mission from those. For adoption
		// through the same API, reconstruct the mission shape if a direction
		// was provided.
		objective := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.objective")
		directionRaw := datahelpers.ExtractNestedField(params.CollectedData, "input_data.direction")
		if objective != "" || directionRaw != nil {
			reconstructed := make(map[string]interface{})
			if objective != "" {
				reconstructed["objective"] = objective
			}
			if dMap, ok := directionRaw.(map[string]interface{}); ok {
				for k, v := range dMap {
					reconstructed[k] = v
				}
			}
			missionRaw = reconstructed
		}
	}

	if missionRaw != nil {
		missionUnwrapped := datahelpers.UnwrapDeep(missionRaw, logger)
		if missionData, ok := missionUnwrapped.(map[string]interface{}); ok && len(missionData) > 0 {
			specAspects["mission"] = missionData
			logger.Info("Writing mission spec from input_data",
				zap.Int("mission_keys", len(missionData)),
			)
		}
	}

	for aspect, data := range specAspects {
		if data == nil {
			continue
		}
		dataJSON, err := json.Marshal(data)
		if err != nil {
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
			) VALUES ($1, $2, $3::jsonb, 'adoption', 'site-adoption-agent',
			          true, 'site-adoption-agent', $4)
		`, siteID, aspect, string(dataJSON),
			fmt.Sprintf("Adopted from %s", sourceDomain))
		if err == nil {
			specsWritten++
		}
	}

	// ── 3. Build crawl index and per-page feature map ───────────────────

	crawlPages := buildCrawlPageIndex(params.CollectedData, params.DB, siteID, logger)

	// ── 3a. fidelity=locked: preserve the site instead of recreating it ──
	//
	// doc 028's fidelity dial, finally consumed. Everything below this branch
	// (sections 4-5) canonicalises page URLs into a new shape and queues LLM
	// recreation of every page; for a site adopted to be KEPT — hand-built,
	// already indexed, already working — both of those are damage. The locked
	// path writes the crawled documents through verbatim under their own URLs.
	//
	// Specs written above still stand: they describe the site and nothing in
	// them restyles it. The classifier handoff below is deliberately skipped —
	// see applyVerbatimAdoption.
	if fidelity := adoptionFidelity(params.CollectedData); fidelity == fidelityLocked {
		vPages, vItems, vErr := applyVerbatimAdoption(ctx, tx, siteID, domain,
			sourceDomain, crawlPages, batchID, logger)
		if vErr != nil {
			return nil, fmt.Errorf("verbatim adoption (fidelity=locked): %w", vErr)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
		logger.Info("ApplyAdoptionPlanAction: Complete (fidelity=locked, verbatim)",
			zap.String("domain", domain),
			zap.Int("specs_written", specsWritten),
			zap.Int("pages_created", vPages),
			zap.Int("items_created", vItems),
			zap.String("batch_id", batchID.String()),
		)
		return map[string]interface{}{
			"applied":       true,
			"domain":        domain,
			"fidelity":      fidelityLocked,
			"deploy_mode":   deployModeVerbatim,
			"specs_written": specsWritten,
			"pages_created": vPages,
			"items_created": vItems,
			"batch_id":      batchID.String(),
			"research_id":   researchID.String(),
		}, nil
	}

	pageFeatures := buildPageFeatureMap(plan)

	// ── 4. Create page records ──────────────────────────────────────────

	pagesRaw, _ := plan["pages"].([]interface{})

	type adoptedPage struct {
		ID       uuid.UUID
		Name     string
		PageType string
		Features []map[string]interface{}
	}
	var adoptedPages []adoptedPage

	for i, pageRaw := range pagesRaw {
		pm, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Canonicalise name / url / page_type via the shared helper so
		// adoption and the planner produce identical identities for the
		// same page (doc 029 Phase 0). Adoption's analyze_site LLM emits
		// names like "tool-ttk-calculator", "guides-index"; the planner's
		// LLM emits "ttk-calculator", "guides". CanonicalisePage absorbs
		// both shapes.
		rawName, _ := pm["name"].(string)
		if rawName == "" {
			rawName = fmt.Sprintf("page-%d", i+1)
		}
		rawType, _ := pm["page_type"].(string)

		pageName, pageURL, pageType := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
			Role: rawType,
			Slug: rawName,
		})
		if pageName == "" {
			logger.Warn("ApplyAdoptionPlanAction: page failed canonicalisation, skipping",
				zap.String("raw_name", rawName),
				zap.String("raw_type", rawType))
			continue
		}

		pageTitle, _ := pm["title"].(string)
		if pageTitle == "" {
			pageTitle = strings.Title(strings.ReplaceAll(pageName, "-", " "))
		}

		metaDesc, _ := pm["meta_description"].(string)

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
				title = EXCLUDED.title, url = EXCLUDED.url,
				page_type = EXCLUDED.page_type, sections = EXCLUDED.sections,
				meta_description = EXCLUDED.meta_description, updated_at = NOW()
			RETURNING id
		`, siteID, pageName, pageURL, pageTitle, pageType,
			string(sectionsJSON), metaDesc, navLabel, inHeader, inFooter,
		).Scan(&pageID)

		if err != nil {
			logger.Warn("Failed to create page",
				zap.String("name", pageName), zap.Error(err))
			continue
		}

		// Store crawl content — markdown always, rawHtml when available.
		// Use sourceDomain (not destination domain) because the crawl index
		// is keyed by the SOURCE host — "https://source-site.com/about".
		// When source == destination, sourceDomain == domain and behaviour
		// is unchanged.
		crawlContent := matchCrawlContent(crawlPages, pageURL, sourceDomain)
		if crawlContent != nil {
			contentData := map[string]interface{}{
				"page_name":    pageName,
				"page_id":      pageID.String(),
				"adopted_from": sourceDomain,
				"mode":         "recreate",
				"existing_content": map[string]interface{}{
					"raw_markdown": crawlContent.Markdown,
				},
			}
			if crawlContent.RawHTML != "" {
				contentData["existing_content"].(map[string]interface{})["raw_html"] = crawlContent.RawHTML
			}
			if features, ok := pageFeatures[pageName]; ok {
				contentData["interactive_features"] = features
			}

			pageContentJSON, _ := json.Marshal(contentData)
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
			Features: pageFeatures[pageName],
		})
		pagesCreated++
	}

	// ── 5. Create work items ────────────────────────────────────────────
	//
	// Hand off to the classifier. Per doc 028 ("The classifier is the
	// strategic brain — it always runs in full. Adoption does not shortcut
	// it"), adoption does NOT queue needs_composition or needs_design
	// directly. Those are produced by the cascade that the classifier
	// starts (strategist → briefing → site-planner → composition →
	// webdesign).
	//
	// The specs adoption wrote above — site_archetype, design_reference,
	// identity, content_direction, design_intent — are read by the
	// classifier (006 prompt) and preserved under doc 028's read-and-extend
	// ownership pattern. The classifier also runs its own vertical /
	// competitor research (doc 028 failure mode: "Adoption without
	// strategic analysis"), and produces the classification spec with the
	// category + industry_tags the layout matcher needs (migration 008).
	//
	// Priority 5 matches domain-submitter's default — from the classifier's
	// point of view, adoption is just another input shape. The adopted_from
	// hint in the spec lets the classifier's prompt distinguish when useful,
	// without needing a different handler agent.

	researchSpec, _ := json.Marshal(map[string]interface{}{
		"site_id":      siteID.String(),
		"domain":       domain, // destination — the site being built
		"reason":       "post_adoption_classification",
		"adopted_from": sourceDomain, // the crawled source — provenance only
		"objective":    "Classify adopted site; respect site_archetype constraints; produce category + industry_tags for the layout matcher.",
	})
	_, err = insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "adoption",
		pipeline:     "build",
		itemType:     "needs_domain_research",
		severity:     "high",
		summary:      fmt.Sprintf("Classify adopted site %s (adoption→classifier handoff)", domain),
		spec:         string(researchSpec),
		priority:     5, // matches domain-submitter's default
		handlerAgent: "domain-research-classifier",
		status:       "triaged",
		createdBy:    "site-adoption-agent",
		itemKey:      "needs_domain_research",
		batchID:      batchID,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("insert needs_domain_research: %w", err)
	}
	itemsCreated++

	logger.Info("Adoption: handoff to classifier queued",
		zap.String("site_id", siteID.String()),
		zap.String("domain", domain),
		zap.String("source_domain", sourceDomain),
		zap.String("handoff", "needs_domain_research → domain-research-classifier"),
	)

	// Content pages — route interactive pages to tool-recreation-handler
	for i, page := range adoptedPages {
		pageSpec := map[string]interface{}{
			"page_name": page.Name,
			"page_type": page.PageType,
			"source":    "adoption",
		}

		var itemType, handlerAgent, summary string
		var priority int

		if len(page.Features) > 0 {
			// Interactive page — tool recreation handler
			pageSpec["mode"] = "recreate" // load_existing_content checks for this value
			pageSpec["interactive_features"] = page.Features
			itemType = "needs_tool_recreation"
			handlerAgent = "tool-recreation-handler"
			summary = fmt.Sprintf("Recreate %s (interactive) from %s", page.Name, domain)
			priority = 5 + i // higher priority — tools are the site's value
		} else {
			// Static content page — normal page build handler
			pageSpec["mode"] = "recreate"
			itemType = "needs_content_page"
			handlerAgent = "page-build-handler"
			summary = fmt.Sprintf("Recreate %s page from %s", page.Name, domain)
			priority = 10 + i
		}

		pageSpecJSON, _ := json.Marshal(pageSpec)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by,
				item_key, batch_id
			) VALUES ($1, 'adoption', 'build', $2, 'medium',
			          $3, $4::jsonb, $5, $6, $7, 'triaged',
			          'site-adoption-agent', $8, $9)
			ON CONFLICT DO NOTHING
		`, siteID, itemType,
			summary, string(pageSpecJSON), page.ID, priority, handlerAgent,
			// Canonical work-item key (doc 029 Phase 0). Shared shape with
			// planner-emitted items so idx_swi_dedup catches collisions.
			fmt.Sprintf("needs_page:%s", page.Name),
			batchID)
		if err == nil {
			itemsCreated++
		}
	}

	// Rerender
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

	interactiveCount := 0
	for _, p := range adoptedPages {
		if len(p.Features) > 0 {
			interactiveCount++
		}
	}

	logger.Info("ApplyAdoptionPlanAction: Complete",
		zap.String("domain", domain),
		zap.Int("specs_written", specsWritten),
		zap.Int("pages_created", pagesCreated),
		zap.Int("items_created", itemsCreated),
		zap.Int("interactive_pages", interactiveCount),
		zap.String("batch_id", batchID.String()),
	)

	return map[string]interface{}{
		"applied":           true,
		"domain":            domain,
		"specs_written":     specsWritten,
		"pages_created":     pagesCreated,
		"items_created":     itemsCreated,
		"interactive_pages": interactiveCount,
		"batch_id":          batchID.String(),
		"research_id":       researchID.String(),
	}, nil
}

// ── Helpers ─────────────────────────────────────────────────────────────

// buildPageFeatureMap groups the LLM's interactive_features by the CANONICAL
// page name, so the routing loop's pageFeatures[pageName] lookup
// (pageName = datahelpers.CanonicalisePage(...)) resolves.
//
// interactive_features[].page references a page by its RAW name — the same
// shape as pages[].name ("drop-rate-simulator" for tools, "game-foo" for
// games). The routing loop canonicalises page names, and CanonicalisePage
// adds a role prefix for some roles (tool -> "tool-foo") but not others
// ("game-foo" is already prefixed). Keying this map by the raw name therefore
// missed every tool page (raw "drop-rate-simulator" vs canonical
// "tool-drop-rate-simulator") while games happened to match — so adopted tools
// were misrouted to page-build-handler and rebuilt as static description pages
// with no widget. Keying by the canonical name removes the divergence.
// (See doc 016 addendum / doc 029 canonicalisation.)
//
// Generic — works for tools, games, calculators, forms, search,
// visualisations, or any other interactive element.
func buildPageFeatureMap(plan map[string]interface{}) map[string][]map[string]interface{} {
	featuresByPage := make(map[string][]map[string]interface{})

	features, ok := plan["interactive_features"].([]interface{})
	if !ok {
		return featuresByPage
	}

	// Map each page's raw name -> canonical name, using the same
	// CanonicalisePage call the routing loop uses. interactive_features[].page
	// references a page by its raw name, so this lets us key features by the
	// canonical name the lookup will use.
	canonByRawName := make(map[string]string)
	if pages, ok := plan["pages"].([]interface{}); ok {
		for _, pr := range pages {
			pm, ok := pr.(map[string]interface{})
			if !ok {
				continue
			}
			rawName, _ := pm["name"].(string)
			rawType, _ := pm["page_type"].(string)
			if rawName == "" {
				continue
			}
			if cname, _, _ := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
				Role: rawType,
				Slug: rawName,
			}); cname != "" {
				canonByRawName[strings.ToLower(strings.TrimSpace(rawName))] = cname
			}
		}
	}

	for _, f := range features {
		fm, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		// NOTE: local renamed from `pageName` to `rawPage` — this is the RAW
		// feature reference, not the canonical pageName the routing loop uses;
		// the old name invited exactly the confusion behind this bug.
		rawPage, _ := fm["page"].(string)
		if rawPage == "" {
			continue
		}

		// Prefer the canonical name (matches the routing loop's lookup); fall
		// back to the raw ref if the feature names a page absent from
		// plan.pages, so nothing is silently dropped.
		key := rawPage
		if cname, found := canonByRawName[strings.ToLower(strings.TrimSpace(rawPage))]; found {
			key = cname
		}

		featuresByPage[key] = append(featuresByPage[key], fm)
	}

	return featuresByPage
}

// buildCrawlPageIndex extracts markdown and rawHtml from the crawl response,
// keyed by URL. rawHtml is present when the crawl requested it via
// scrapeOptions.formats.
//
// Added DB fallback to buildCrawlPageIndex. When the crawl data isn't in
// collected_data (paginated flow stores per-page in research_results),
// read from DB.
func buildCrawlPageIndex(collectedData map[string]interface{}, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) map[string]*crawlPageContent {
	index := make(map[string]*crawlPageContent)

	// ── Primary path: read from collected_data (single-message crawl) ───
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

	// Process pages from collected_data
	for _, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		markdown, _ := page["markdown"].(string)
		rawHTML, _ := page["rawHtml"].(string)

		if markdown == "" && rawHTML == "" {
			continue
		}
		if markdown != "" && strings.Contains(markdown, "NoSuchKey") {
			continue
		}

		metadata, _ := page["metadata"].(map[string]interface{})
		pageURL, _ := metadata["url"].(string)
		sourceURL, _ := metadata["sourceURL"].(string)
		statusCode, _ := metadata["statusCode"].(float64)

		if statusCode != 0 && statusCode != 200 {
			continue
		}

		content := &crawlPageContent{Markdown: markdown, RawHTML: rawHTML}

		if pageURL != "" {
			index[pageURL] = content
			for _, prefix := range []string{"https://", "http://"} {
				if strings.HasPrefix(pageURL, prefix) {
					afterScheme := pageURL[len(prefix):]
					slashIdx := strings.Index(afterScheme, "/")
					if slashIdx >= 0 {
						index[afterScheme[slashIdx:]] = content
					}
				}
			}
		}
		if sourceURL != "" && sourceURL != pageURL {
			index[sourceURL] = content
		}
	}

	// ── DB fallback: read from research_results (paginated crawl) ───────
	if len(index) == 0 && db != nil {
		logger.Info("buildCrawlPageIndex: no pages in collected_data, trying DB fallback")

		rows, err := db.QueryContext(context.Background(), `
			SELECT data FROM research_results
			WHERE site_id = $1 AND result_type = 'adoption_crawl_page'
			ORDER BY created_at
		`, siteID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var dataJSON []byte
				if err := rows.Scan(&dataJSON); err != nil {
					continue
				}
				var pageData map[string]interface{}
				if err := json.Unmarshal(dataJSON, &pageData); err != nil {
					continue
				}

				pageURL, _ := pageData["url"].(string)
				markdown, _ := pageData["markdown"].(string)
				rawHTML, _ := pageData["raw_html"].(string)

				if pageURL == "" || (markdown == "" && rawHTML == "") {
					continue
				}

				content := &crawlPageContent{Markdown: markdown, RawHTML: rawHTML}
				index[pageURL] = content

				// Also index by path
				for _, prefix := range []string{"https://", "http://"} {
					if strings.HasPrefix(pageURL, prefix) {
						afterScheme := pageURL[len(prefix):]
						slashIdx := strings.Index(afterScheme, "/")
						if slashIdx >= 0 {
							index[afterScheme[slashIdx:]] = content
						}
					}
				}
			}

			if len(index) > 0 {
				logger.Info("buildCrawlPageIndex: loaded from DB fallback",
					zap.Int("indexed_urls", len(index)))
			}
		}
	}

	htmlCount := 0
	for _, c := range index {
		if c.RawHTML != "" {
			htmlCount++
		}
	}

	logger.Info("Built crawl page index",
		zap.Int("indexed_urls", len(index)),
		zap.Int("with_raw_html", htmlCount),
		zap.Int("raw_pages", len(pages)))

	return index
}

// matchCrawlContent finds crawl content for a page URL
func matchCrawlContent(index map[string]*crawlPageContent, pageURL, domain string) *crawlPageContent {
	if c, found := index[pageURL]; found {
		return c
	}
	if c, found := index["https://"+domain+pageURL]; found {
		return c
	}
	if c, found := index["http://"+domain+pageURL]; found {
		return c
	}
	return nil
}

// repairTruncatedJSON fixes JSON cut off by max_tokens
func repairTruncatedJSON(s string) string {
	if len(s) == 0 {
		return ""
	}
	lastGood := -1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '}' || s[i] == ']' {
			lastGood = i + 1
			break
		}
	}
	if lastGood <= 0 {
		return ""
	}
	truncated := strings.TrimRight(s[:lastGood], " \t\n\r,")
	for strings.Count(truncated, "[") > strings.Count(truncated, "]") {
		truncated += "]"
	}
	for strings.Count(truncated, "{") > strings.Count(truncated, "}") {
		truncated += "}"
	}
	return truncated
}

func truncateTail(s string, n int) string {
	if len(s) > n {
		return s[len(s)-n:]
	}
	return s
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// countMapKeys returns len of a nested map field, or 0 if not found.
func countMapKeys(data map[string]interface{}, key string) int {
	if m, ok := data[key].(map[string]interface{}); ok {
		return len(m)
	}
	return 0
}
