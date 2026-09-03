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
	"github.com/gqls/agentchassis/platform/livespec"
	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
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
//
// The string itself lives in discovery_checks.PageRerenderItemKey (exported
// 2026-08-24, bugs_open/384) so the page_list_stale sweep — which cannot import
// this package — files a byte-identical key and dedups against the event
// emitters here. This is a delegate, not a second spelling.
func pageRerenderItemKey(pageName string, siteID uuid.UUID, keyReason string) string {
	return discovery_checks.PageRerenderItemKey(pageName, siteID, keyReason)
}

// rerenderMode is the decision the two gates make, extracted so it is PURE and
// therefore exhaustively testable without a database — the real function the
// action calls, not a re-implementation a test could disagree with silently.
type rerenderMode struct {
	Scoped        bool
	StampReason   bool
	KeyReason     string // exactly what the spec will carry; "" means assemble-only
	UnknownReason string // non-empty when the caller passed a reason nobody declared

	// RoutingKey is the value for spec.routing_reason — the ROUTING half of the
	// spec.reason split (bugs_open/440, RFC_062, REB-008). This action is the
	// FIRST sanctioned producer of that key; REB-008 forbids a second before the
	// RFC lands.
	//
	// ⚠ IT IS SET EXACTLY WHEN KeyReason IS, AND THAT IS THE WHOLE CARE OF THIS
	// CHANGE. Stamping it whenever the reason is merely KNOWN would be a
	// behaviour change disguised as a foundation: image_landed WITHOUT a
	// component_id deliberately stamps nothing (REB-001's designed degrade to
	// assemble), so a routing key there would make the phase-3 gate route a page
	// that assembles today. Keeping the two fields in lockstep is what makes the
	// flip provably byte-neutral for every reason that works correctly now.
	RoutingKey string
	Warnings   []string
}

// rerenderModeFor resolves a (reason, component_id) pair against the ONE
// definition of the sections-rerender vocabulary — bugs_open/404.
//
// The two gates are DIFFERENT TESTS and that is the whole subtlety:
//
//	Scoped      narrows WHICH PAGES get items, and needs a component to scope BY
//	StampReason decides whether the item carries a reason AT ALL, which is what
//	            page-rerender's check_rerender_mode branches on
//
// An unknown reason yields neither, so the item carries no reason, so the gate
// reads it as assemble-only and re-ships the stored HTML verbatim — completing
// green and changing nothing. Both readers fail toward assemble; that direction
// is why this vocabulary drifted twice in one day without anyone noticing.
func rerenderModeFor(reason, componentID string) rerenderMode {
	var m rerenderMode
	if reason == "" {
		// The ordinary case, and it must stay silent: a site-wide refresh IS
		// supposed to be assemble-only. [MEASURED 2026-08-26] 17,844 items of
		// exactly this shape, correctly.
		return m
	}

	r, known := livespec.RerenderSectionReasonByName(reason)
	if !known {
		// LOUD, NOT REFUSED. An unknown routing key that completes green IS this
		// bug — but out-of-vocabulary reasons are also a live, legitimate practice:
		// adopt_verbatim stamps `verbatim_adoption_deploy` on items that are
		// SUPPOSED to assemble, because verbatim adoption re-ships stored HTML by
		// design. Refusing would be new authority on a shared seam with a standing
		// counter-example (owner ruling 2026-08-02 §2), so this reports and lets
		// the caller decide.
		m.UnknownReason = reason
		m.Warnings = append(m.Warnings,
			"create_rerender_items: reason is not in the sections-rerender vocabulary — these "+
				"items will be ASSEMBLE-ONLY and will re-ship the stored HTML unchanged. If that is "+
				"what you meant, good; if you expected a template or section re-render, this is "+
				"bugs_open/404.")
		return m
	}

	m.Scoped = r.ComponentScoped && componentID != ""
	m.StampReason = m.Scoped || r.StampAlways
	if m.StampReason {
		m.KeyReason = reason
		// In lockstep, deliberately — see RoutingKey's doc. `reason` is
		// in-vocabulary on this branch by construction (the !known branch
		// returned above), so an unknown value can never reach the routing key:
		// that is REB-008's no-bad-producer constraint enforced by control flow
		// rather than by a check that could be edited away.
		m.RoutingKey = reason
	}

	if r.StampAlways && r.ComponentScoped && componentID == "" {
		// Stamped but unscoped: every page in the caller's list gets a sections
		// re-render. Bounded by that list, and the RIGHT degrade direction —
		// assemble can never deliver a template change at all, so over-delivery
		// (which announces itself) beats silent under-delivery.
		m.Warnings = append(m.Warnings,
			"create_rerender_items: "+reason+" without component_id — stamping WITHOUT component "+
				"scoping, so every page in this list gets a sections re-render. Pass component_id "+
				"to narrow it to the pages that carry the component.")
	}
	return m
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
// rerenderItemExec is the slice of *sql.DB / *sql.Tx this INSERT needs, so the
// one canonical implementation can also be called from inside a transaction
// (verbatim adoption creates pages and queues their deploys atomically).
// Widening the parameter is what keeps that caller from hand-rolling a second
// copy of this row shape — the exact drift this function exists to prevent.
type rerenderItemExec interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func insertPageRerenderItem(
	ctx context.Context,
	db rerenderItemExec,
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

	// ── THE VOCABULARY HAS ONE DEFINITION NOW (bugs_open/404) ──────────────
	//
	// These two branches used to carry their own hardcoded copies of "which
	// reasons mean re-resolve", and page-rerender's live gate carried a third. On
	// 2026-08-18 two lanes appended a value each to the gate — template_changed
	// (migration 460) and literal_markdown (473) — and neither touched this file.
	// The gate knew five; this reader knew three.
	//
	// ⚠ AND THE FAILURE DIRECTION IS WHY THAT MATTERED. An unknown reason here
	// leaves keyReason empty, so the item carries NO reason, so the gate reads it
	// as assemble-only and re-ships the stored HTML verbatim — completing green
	// and changing nothing. Both readers fail toward assemble, which is the
	// estate's safe, cheap, preferred mode AND its silent-failure mode, so drift
	// in this vocabulary is invisible by construction.
	//
	// livespec.RerenderSectionReasons is the single definition, and the live gate
	// is asserted against it every morning by config-key-audit
	// --live-declaration-drift.
	mode := rerenderModeFor(reason, componentIDStr)
	scoped, stampReason, unknownReason := mode.Scoped, mode.StampReason, mode.UnknownReason
	for _, w := range mode.Warnings {
		logger.Warn(w, zap.String("reason", reason),
			zap.Strings("declared", livespec.RerenderSectionReasonNames()))
	}

	// keyReason mirrors EXACTLY what the spec carries (empty spec.reason =>
	// assemble-only), so the dedup key discriminates the two render modes. See
	// pageRerenderItemKey — bugs_open/024 defect 6.
	keyReason := mode.KeyReason

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
	var emptyPagesConverted []string

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

		// bugs_open/315 (reopened 2026-09-02): an assemble-only item on a page
		// with ZERO component rows is a guaranteed skip that still completes —
		// nine such completions accumulated on one empty page. Do not file the
		// useless rerender; file the build ask the page actually needs (same
		// deduped item fileBuildAskForEmptyPage files on the consumer side, so
		// both doors converge on ONE open needs_content_page item). Scoped
		// runs are unaffected: dependent pages have components by construction.
		if !scoped {
			var hasComponents bool
			if err := params.DB.QueryRowContext(ctx,
				`SELECT EXISTS(SELECT 1 FROM page_components WHERE page_id = $1)`,
				pageUUID).Scan(&hasComponents); err == nil && !hasComponents {
				fileBuildAskForEmptyPage(ctx, params.DB, siteID, pageUUID, pageName, nil, logger)
				emptyPagesConverted = append(emptyPagesConverted, pageName)
				continue
			}
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
			// And the routing half beside it (bugs_open/440 phase 1b). The gate
			// still reads `reason` today — this key is INERT until RFC_062's
			// phase-3 migration, and is written now so that when the gate flips
			// there is a populated key to flip TO, rather than a drain window in
			// which every in-flight item routes to assemble.
			//
			// ⚠ NOT part of the dedup key: pageRerenderItemKey still takes
			// keyReason alone, so this addition cannot change which items
			// dedupe against which (idx_swi_dedup) — the one way a "just add a
			// field" change could have altered live behaviour.
			spec[livespec.RoutingReasonSpecKey] = mode.RoutingKey
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
		zap.Strings("empty_pages_converted_to_build_asks", emptyPagesConverted),
		zap.String("batch_id", batchID.String()))

	out := map[string]interface{}{
		"items_created": itemsCreated,
		"pages_total":   len(pages),
		"batch_id":      batchID.String(),
	}
	if len(emptyPagesConverted) > 0 {
		// In the RESULT, not only a pod log (the unknown_reason rule below,
		// same reasoning): "why did my page get no rerender item?" must be
		// answerable from the step result. These pages have 0 component rows;
		// each got a deduped needs_content_page ask instead (bugs_open/315).
		out["empty_pages_converted_to_build_asks"] = emptyPagesConverted
	}
	if unknownReason != "" {
		// In the RESULT, not only in a pod log. A log line scrolls; a step result
		// is what an operator reads when they ask why nothing changed — which is
		// the question this bug exists to make answerable.
		out["unknown_reason"] = unknownReason
	}
	return out, nil
}
