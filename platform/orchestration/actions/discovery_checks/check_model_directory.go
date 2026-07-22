// FILE: platform/orchestration/actions/discovery_checks/check_model_directory.go
//
// Discovery checks for the model directory pipeline (model_directory_pipeline
// Phase D). Cloned from check_news_feed.go's MissingNewsSectionCheck /
// MissingNewsPageCheck — same opt-in shape
// (site_specs.classification.content_features.model_directory), same
// handler (content-gap-planner decides placement, page-build-handler
// executes).
//
// One structural difference from news, both checks account for: there is no
// per-site "missing_model_directory_sources" equivalent to
// missing_news_sources. The registry (directory_entities/directory_claims,
// migration 192) is global, not per-site — discovery (Phase B's
// directory-researcher, on its own scheduled cadence) is what populates it,
// not anything triggered by a site opting in. So both checks below gate on
// "does the global registry have ANY publishable data yet" rather than "does
// this site have its own sources" — a site can opt in before the registry
// has data; the checks simply wait (return no finding) until it does,
// rather than creating an empty section.
//
// Simplification vs. the news precedent (deliberate, not an oversight):
// MissingModelDirectoryPageCheck does NOT do news's "stranded nav page /
// retype, don't duplicate" check (bugs_open/015). That check exists because
// news pages had already been created under wrong page_types (section-index,
// noticias-index, ...) on MANY sites before the news-index gate existed.
// model-directory is a brand-new page_type with no such history — there is
// nothing to retype yet. Add the same protection here if that ever changes.
//
// Handler agent routing:
//   - missing_model_directory_section → content-gap-planner (decides where
//     to add the model-directory snippet)
//   - missing_model_directory_page     → content-gap-planner (creates a
//     dedicated page with model-directory-listing)
//
// To enable, add these check names to the completeness-discovery-agent's
// "checks" config array (see SEED_directory_discovery_checks.sql).

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() {
	Register(&MissingModelDirectorySectionCheck{})
	Register(&MissingModelDirectoryPageCheck{})
}

// modelDirectoryHasPublishableData reports whether the global registry has
// at least one status='found' claim — the same "is there anything to show
// yet" gate a per-site sources check would apply, re-keyed to a global
// registry with no site_id to check.
func modelDirectoryHasPublishableData(dctx DiscoveryCheckContext) (int, error) {
	var claimCount int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM directory_claims WHERE is_current AND status = 'found'
	`).Scan(&claimCount)
	return claimCount, err
}

// modelDirectorySpecFlags reads content_features.model_directory from a
// site's current classification spec. ok is false when the aspect, key, or
// site spec doesn't exist — treated as "not opted in", not an error.
func modelDirectorySpecFlags(dctx DiscoveryCheckContext) (recommended, separatePage bool, raw map[string]interface{}, ok bool) {
	var specData []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
	`, dctx.SiteID).Scan(&specData)
	if err != nil {
		return false, false, nil, false
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(specData, &spec); err != nil {
		return false, false, nil, false
	}
	contentFeatures, _ := spec["content_features"].(map[string]interface{})
	if contentFeatures == nil {
		return false, false, nil, false
	}
	modelDirectory, _ := contentFeatures["model_directory"].(map[string]interface{})
	if modelDirectory == nil {
		return false, false, nil, false
	}
	recommended, _ = modelDirectory["recommended"].(bool)
	separatePage, _ = modelDirectory["separate_page"].(bool)
	return recommended, separatePage, modelDirectory, true
}

// ===========================================================================
// MissingModelDirectorySectionCheck
// ===========================================================================
// Detects: site spec has content_features.model_directory.recommended =
// true, the global registry has publishable data, but no page has a
// model-directory page_component.

type MissingModelDirectorySectionCheck struct{}

func (c *MissingModelDirectorySectionCheck) Name() string { return "missing_model_directory_section" }

func (c *MissingModelDirectorySectionCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	recommended, _, config, ok := modelDirectorySpecFlags(dctx)
	if !ok || !recommended {
		return &CheckResult{}, nil
	}

	claimCount, err := modelDirectoryHasPublishableData(dctx)
	if err != nil {
		return nil, fmt.Errorf("missing_model_directory_section: registry query failed: %w", err)
	}
	if claimCount == 0 {
		// Site has opted in, but discovery hasn't produced anything yet.
		// Wait for the next sweep rather than creating an empty section.
		return &CheckResult{}, nil
	}

	var sectionCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND EXISTS (
		      SELECT 1 FROM content_components cc
		      WHERE cc.id = pc.component_id AND cc.function = 'model-directory'
		  )
	`, dctx.SiteID).Scan(&sectionCount)
	if err != nil {
		return nil, fmt.Errorf("missing_model_directory_section: section query failed: %w", err)
	}
	if sectionCount > 0 {
		return &CheckResult{}, nil
	}

	dctx.Logger.Info("MissingModelDirectorySectionCheck: registry has data but no model-directory section",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("claim_count", claimCount))

	specJSON, _ := json.Marshal(map[string]interface{}{
		"check":                  "missing_model_directory_section",
		"registry_claim_count":   claimCount,
		"page_name":              "index",
		"model_directory_config": config,
		"description":            "Site opted into the model directory and the registry has publishable entries, but the homepage has no model-directory section to display them.",
		"suggestion":             "Add a model-directory section to the homepage to show the cited AI model directory",
		"category":               "content_completeness",
	})

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":       "missing_model_directory_section",
			"message":     "Registry has publishable entries but homepage has no model-directory section",
			"claim_count": claimCount,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "missing_model_directory_section",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Site opted into the model directory (%d cited entries available) but has no homepage section for it", claimCount),
			SpecJSON:     string(specJSON),
			Priority:     65,
			HandlerAgent: "content-gap-planner",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("missing_model_directory_section:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}

// ===========================================================================
// MissingModelDirectoryPageCheck
// ===========================================================================
// Detects: content_features.model_directory.separate_page = true, the
// global registry has publishable data, but no page with
// page_type='model-directory' exists.

type MissingModelDirectoryPageCheck struct{}

func (c *MissingModelDirectoryPageCheck) Name() string { return "missing_model_directory_page" }

func (c *MissingModelDirectoryPageCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	recommended, separatePage, config, ok := modelDirectorySpecFlags(dctx)
	if !ok || !recommended || !separatePage {
		return &CheckResult{}, nil
	}

	claimCount, err := modelDirectoryHasPublishableData(dctx)
	if err != nil {
		return nil, fmt.Errorf("missing_model_directory_page: registry query failed: %w", err)
	}
	if claimCount == 0 {
		return &CheckResult{}, nil
	}

	var pageCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM pages
		WHERE site_id = $1 AND page_type = 'model-directory'
	`, dctx.SiteID).Scan(&pageCount)
	if err != nil {
		return nil, fmt.Errorf("missing_model_directory_page: page query failed: %w", err)
	}
	if pageCount > 0 {
		return &CheckResult{}, nil
	}

	var domain string
	_ = dctx.DB.QueryRowContext(dctx.Ctx, `SELECT domain FROM sites WHERE id = $1`, dctx.SiteID).Scan(&domain)

	dctx.Logger.Info("MissingModelDirectoryPageCheck: spec recommends a dedicated page but none exists",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("claim_count", claimCount))

	gapSpec := map[string]interface{}{
		"check":                  "missing_model_directory_page",
		"page_name":              "model-directory",
		"page_type":              "model-directory",
		"category":               "content_completeness",
		"registry_claim_count":   claimCount,
		"model_directory_config": config,
		"approach":               "new_page",
		"description":            fmt.Sprintf("Site %s opted into a dedicated model-directory page (separate_page=true) and the registry has %d cited entries, but no page of page_type 'model-directory' exists.", domain, claimCount),
		"suggestion":             "Create a new page named 'model-directory' with page_type 'model-directory', a hero section, the model-directory-listing component, and a call-to-action. Add it to header and footer navigation.",
	}
	specJSON, _ := json.Marshal(gapSpec)

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":   "missing_model_directory_page",
			"message": fmt.Sprintf("Site spec recommends a dedicated model-directory page but none exists for %s", domain),
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "missing_model_directory_page",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Site %s needs a dedicated model-directory page (%d cited entries available)", domain, claimCount),
			SpecJSON:     string(specJSON),
			Priority:     60,
			HandlerAgent: "content-gap-planner",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("missing_model_directory_page:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}
