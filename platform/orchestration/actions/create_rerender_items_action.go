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
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CreateRerenderItemsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional: []string{
		"domain", "pages_field", "reason", "component_id",
		// Single-page mode (see singlePageFromScalars): a producer that made
		// exactly ONE page names it by scalar paths instead of a list.
		"page_id_field", "page_name_field", "filename_field",
	},
	Defaults:   map[string]interface{}{"pages_field": "rerender_pages.pages"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("create_rerender_items", CreateRerenderItemsInputSpec)
}

// singlePageFromScalars builds a one-page list from scalar config paths, for
// producers that create exactly ONE page and therefore never assemble a pages
// array. Returns nil unless page_id_field and page_name_field both resolve.
//
// Why this exists: tool-generator creates a component, a page, nav and a PLAN,
// and its create_tool_component result even carries "needs_rerender": true —
// but nothing ever acted on it, so a newly born tool page sat at
// build_status='planned' until an unrelated sweep happened to pick it up. All
// three tool births to date (xp-curve, drop-rate, loot-table) needed a
// hand-inserted work item shaped exactly like the one below. Rather than a
// second, near-identical action, the existing one learns to take a single
// page: the item shape, status and dedup key stay defined in one place.
//
// filename_field is optional and tolerates a leading "/" (create_result's
// page_url is "/tools/x.html" while the item spec wants "tools/x.html").
func singlePageFromScalars(collected map[string]interface{}, config map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	idField := datahelpers.GetStringField(config, "page_id_field", "")
	nameField := datahelpers.GetStringField(config, "page_name_field", "")
	if idField == "" || nameField == "" {
		return nil
	}
	pageID := datahelpers.ExtractNestedFieldString(collected, idField)
	pageName := datahelpers.ExtractNestedFieldString(collected, nameField)
	if pageID == "" || pageName == "" {
		// Configured for single-page mode but the producer wrote nothing —
		// say so, rather than silently creating no work.
		logger.Warn("CreateRerenderItemsAction: single-page mode configured but page id/name did not resolve",
			zap.String("page_id_field", idField), zap.String("page_name_field", nameField),
			zap.String("page_id", pageID), zap.String("page_name", pageName))
		return nil
	}
	filename := ""
	if f := datahelpers.GetStringField(config, "filename_field", ""); f != "" {
		filename = strings.TrimPrefix(datahelpers.ExtractNestedFieldString(collected, f), "/")
	}
	logger.Info("CreateRerenderItemsAction: single-page mode",
		zap.String("page_name", pageName), zap.String("filename", filename))
	return map[string]interface{}{
		"page_id":  pageID,
		"name":     pageName,
		"filename": filename,
	}
}

// pageRerenderItemKey builds the idx_swi_dedup key for a per-page rerender work
// item. It MUST carry the render-MODE discriminator (the stamped spec.reason, or
// "assemble" when none is stamped), because that reason is what page-rerender's
// check_rerender_mode branches on: a reason-bearing item drives a true template
// re-render (rerender_page_sections, the ONLY writer of page_components.rendered_html),
// while a reason-less item drives assemble-only (rerender_single_page — "Simple
// concatenation, no template re-rendering", which re-ships stored HTML).
//
// Keying both modes on page_rerender_<page>_<site> alone let a stale reason-less
// item sitting in the dispatch backlog SUPPRESS the correct reason-bearing one:
// create_rerender_items' INSERT ... ON CONFLICT DO NOTHING collided with the open
// reason-less row, created zero items, and the reason-less item then re-deployed
// stale HTML — bugs_open/024 defect 6, the six-month-invisible delivery blocker.
// Scoping the key by mode preserves dedup WITHIN a mode (two concurrent site-wide
// refreshes still collapse to one assemble-only item per page) while ensuring the
// two modes can never suppress each other.
func pageRerenderItemKey(pageName string, siteID uuid.UUID, keyReason string) string {
	if keyReason == "" {
		keyReason = "assemble"
	}
	return fmt.Sprintf("page_rerender_%s_%s_%s", pageName, siteID, keyReason)
}

// insertPageRerenderItem is THE one INSERT for page_rerender work items — the
// canonical shape (item_type page_rerender → page-rerender, priority 80,
// targetless ON CONFLICT DO NOTHING, key from pageRerenderItemKey). Shared so
// a second emitter can never drift into a textually-similar-but-different
// row: the news freshness emitter (queueNewsPageRerenders) previously carried
// its own hand-rolled copy of this shape with the WRONG item_type and spent
// four days LLM-regenerating live pages. One literal implementation, two
// callers — the ExtractHrefs/PageURLSet sharing pattern.
//
// Targetless ON CONFLICT DO NOTHING is deliberate: idx_swi_dedup is a
// PARTIAL unique index (non-terminal statuses only), and a targeted
// ON CONFLICT (site_id, item_key) without the index's exact WHERE predicate
// raises 42P10. The targetless form matches whatever constraint conflicts —
// the same idiom this INSERT has always used here.
//
// Returns true when a row was actually inserted (false = dedup-suppressed by
// an open item with the same key, which for recurring emitters is normal).
func insertPageRerenderItem(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageID uuid.UUID,
	source string,
	severity string,
	summary string,
	specJSON string,
	itemKey string,
	batchID uuid.UUID,
) (bool, error) {
	res, err := db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			page_id, priority, handler_agent, status, created_by,
			spec, item_key, batch_id
		) VALUES ($1, $2, 'build', 'page_rerender',
		          $3, $4, $5, 80, 'page-rerender', 'triaged',
		          $2, $6::jsonb, $7, $8)
		ON CONFLICT DO NOTHING
	`, siteID, source, severity, summary, pageID, specJSON, itemKey, batchID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
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
	// cta_links_stale is stamped WITHOUT component scoping: the CTA recompute
	// in rerender_page_sections is cheap and page-scoped, and a site-wide CTA
	// repair has no single triggering component to scope by.
	stampReason := scoped || reason == "cta_links_stale"

	// keyReason mirrors EXACTLY what the spec carries (empty spec.reason =>
	// assemble-only), so the dedup key discriminates the two render modes. See
	// pageRerenderItemKey — bugs_open/024 defect 6.
	keyReason := ""
	if stampReason {
		keyReason = reason
	}

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
	var pages []interface{}
	pagesRaw := datahelpers.ExtractNestedField(params.CollectedData, pagesField)
	if pagesRaw != nil {
		var ok bool
		pages, ok = pagesRaw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("pages field is not an array (got %T)", pagesRaw)
		}
	} else if single := singlePageFromScalars(params.CollectedData, config, logger); single != nil {
		pages = []interface{}{single}
	} else {
		logger.Info("CreateRerenderItemsAction: No pages found, nothing to create",
			zap.String("pages_field", pagesField))
		return map[string]interface{}{
			"items_created": 0,
			"reason":        "no pages found",
		}, nil
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
		if stampReason {
			// page-rerender gates the section re-render on this reason.
			spec["reason"] = reason
		}
		specJSON, _ := json.Marshal(spec)

		itemKey := pageRerenderItemKey(pageName, siteID, keyReason)

		_, err = insertPageRerenderItem(ctx, params.DB, siteID, pageUUID,
			"rerender-pages", "medium",
			fmt.Sprintf("Rerender page: %s", pageName),
			string(specJSON), itemKey, batchID)

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
