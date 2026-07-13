// FILE: platform/orchestration/actions/create_rerender_items_action.go
//
// CreateRerenderItemsAction takes the pages list from get_pages_for_rerender
// and creates one work item per page with handler_agent = 'page-rerender'.
//
// This replaces the old pattern where rerender-pages spawned one page-rerender
// pod and looped through all pages sequentially. That approach hit reaper
// timeouts on large sites (16+ pages = 30+ minutes in AWAITING_RESPONSES).
//
// The new pattern: rerender-pages creates work items, then completes.
// The dispatch loop processes each page independently — own pod, own retry,
// own logs. A failure on page 7 doesn't block or kill pages 1-6 or 8-16.
//
// Registration:
//   "create_rerender_items": {
//       Handler:     CreateRerenderItemsAction,
//       Category:    "site",
//       Description: "Create per-page rerender work items for dispatch loop",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CreateRerenderItemsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"domain", "pages_field", "reason", "component_id"},
	Defaults:   map[string]interface{}{"pages_field": "rerender_pages.pages"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("create_rerender_items", CreateRerenderItemsInputSpec)
}

func CreateRerenderItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "create_rerender_items"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, CreateRerenderItemsInputSpec, logger,
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
	if domain == "" {
		// Fallback: look up from DB
		params.DB.QueryRowContext(ctx, `SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)
	}

	// When the rerender was triggered by a specific component changing, the
	// inbound spec carries a section-rerender reason and the component_id. In
	// that case we (a) stamp the reason onto each page_rerender item so
	// page-rerender runs the section re-render (rerender_page_sections)
	// instead of assemble-only, and (b) scope the items to the pages that
	// actually use that component — so a component regen only re-renders its
	// own dependents, not every page on the site (which would re-render
	// unrelated sections and could surface other latent mismatches). Without
	// both signals we leave behaviour exactly as before: one assemble-only
	// item per page, which is what site-wide refreshes rely on.
	reason := inputs.Get("reason")
	componentIDStr := inputs.Get("component_id")
	// Reasons page-rerender gates the section re-render on.
	scoped := (reason == "section_data_resolved" || reason == "image_landed") && componentIDStr != ""

	var dependentPages map[string]bool
	if scoped {
		compUUID, perr := uuid.Parse(componentIDStr)
		if perr != nil {
			return nil, fmt.Errorf("invalid component_id: %w", perr)
		}
		dependentPages = map[string]bool{}
		rows, qerr := params.DB.QueryContext(ctx, `
			SELECT pc.page_id::text
			FROM page_components pc
			JOIN pages p ON p.id = pc.page_id
			WHERE pc.component_id = $1 AND p.site_id = $2
		`, compUUID, siteID)
		if qerr != nil {
			return nil, fmt.Errorf("dependent-page lookup failed: %w", qerr)
		}
		for rows.Next() {
			var pid string
			if serr := rows.Scan(&pid); serr == nil {
				dependentPages[pid] = true
			}
		}
		rows.Close()
		logger.Info("CreateRerenderItemsAction: scoping to component dependents",
			zap.String("component_id", componentIDStr),
			zap.String("reason", reason),
			zap.Int("dependent_pages", len(dependentPages)))
	}

	// Extract pages list
	pagesField := "rerender_pages.pages"
	if f, ok := config["pages_field"].(string); ok && f != "" {
		pagesField = f
	}
	pagesRaw := datahelpers.ExtractNestedField(params.CollectedData, pagesField)
	if pagesRaw == nil {
		logger.Info("CreateRerenderItemsAction: No pages found, nothing to create",
			zap.String("pages_field", pagesField))
		return map[string]interface{}{
			"items_created": 0,
			"reason":        "no pages found",
		}, nil
	}

	pages, ok := pagesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("pages field is not an array (got %T)", pagesRaw)
	}

	if len(pages) == 0 {
		return map[string]interface{}{
			"items_created": 0,
			"reason":        "empty pages list",
		}, nil
	}

	batchID := uuid.New()
	itemsCreated := 0

	for _, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		pageID, _ := page["page_id"].(string)
		pageName, _ := page["name"].(string)
		filename, _ := page["filename"].(string)

		if pageID == "" || pageName == "" {
			logger.Warn("CreateRerenderItemsAction: Skipping page with missing id or name",
				zap.Any("page", page))
			continue
		}

		pageUUID, err := uuid.Parse(pageID)
		if err != nil {
			logger.Warn("CreateRerenderItemsAction: Invalid page_id, skipping",
				zap.String("page_id", pageID), zap.Error(err))
			continue
		}

		// Scoped rerender: skip pages that don't use the changed component.
		if scoped && !dependentPages[pageID] {
			continue
		}

		spec := map[string]interface{}{
			"page_id":   pageID,
			"page_name": pageName,
			"filename":  filename,
			"domain":    domain,
		}
		if scoped {
			// page-rerender gates the section re-render on this reason.
			spec["reason"] = reason
		}
		specJSON, _ := json.Marshal(spec)

		itemKey := fmt.Sprintf("page_rerender_%s_%s", pageName, siteID)

		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				page_id, priority, handler_agent, status, created_by,
				spec, item_key, batch_id
			) VALUES ($1, 'rerender-pages', 'build', 'page_rerender',
			          'medium', $2, $3, 80, 'page-rerender', 'triaged',
			          'rerender-pages', $4::jsonb, $5, $6)
			ON CONFLICT DO NOTHING
		`, siteID,
			fmt.Sprintf("Rerender page: %s", pageName),
			pageUUID,
			string(specJSON),
			itemKey,
			batchID,
		)

		if err != nil {
			logger.Warn("CreateRerenderItemsAction: Failed to create work item",
				zap.String("page_name", pageName), zap.Error(err))
			continue
		}
		itemsCreated++
	}

	logger.Info("CreateRerenderItemsAction: Complete",
		zap.Int("pages_total", len(pages)),
		zap.Int("items_created", itemsCreated),
		zap.String("batch_id", batchID.String()))

	return map[string]interface{}{
		"items_created": itemsCreated,
		"pages_total":   len(pages),
		"batch_id":      batchID.String(),
	}, nil
}
