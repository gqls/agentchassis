// FILE: platform/orchestration/actions/load_page_record_action.go
//
// Loads a page record from the pages table by site_id + page_name or page_id.
// Replaces inline query_database SQL in workflows — column names are
// compiled Go, not fragile JSON strings.
//
// Lookup priority:
//   0. authoritative_page_id (if provided) — lookup by site_id + id, and the
//      name inputs are IGNORED. Opt-in, absent everywhere by default; see below.
//   1. page_name (if provided) — lookup by site_id + name
//   2. page_id (if page_name is empty) — lookup by site_id + id
//   At least one of the three must be provided.
//
// authoritative_page_id (bugs_open/220). The work item's page_id COLUMN names the
// ACTIONABLE page — for a cross-page item type (unbuilt_internal_link) that is the
// page to build, while spec.page_name/spec.page_id name the page CONTAINING the
// defect. Under the name-first priority above, forwarding the right id changes
// nothing: the container's name always resolves, always wins, and the handler
// rebuilds the wrong page while reporting success. The dispatcher therefore
// forwards the column as input_data.page_id, and a step that wants it decisive
// maps it into THIS input. It is a separate opt-in field, not a priority flip,
// per the owner ruling of 2026-08-02 (RFC_010 §2): the id-wins authority is
// visible in the config of the step that asserts it, and every config that does
// not name it keeps today's behaviour exactly.
//
// Handles:
//   - Page found: returns full page record with parsed sections
//   - Page not found: returns {found: false} (not an error)
//   - Page name is "new page needed" / "site-wide" / empty: falls back to page_id
//
// Registration:
//   "load_page_record": {
//       Handler:     LoadPageRecordAction,
//       Category:    "site",
//       Description: "Load a page record by site_id and page name",
//       IsLocal:     true,
//   },
//
// Workflow config example:
//
//   "load_page_record": {
//       "action": "load_page_record",
//       "config": {
//           "site_id": "site_record.site_id",
//           "page_name": "input_data.page_name"
//       },
//       "next_step": "spawn_content_writer",
//       "output_field": "page_record"
//   }

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

var LoadPageRecordInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"page_name", "page_id", "authoritative_page_id"},
	// refuse_owned_page (bugs_open/301): when true, a loaded page whose
	// pages.rebuild_policy is 'owned' is REFUSED here — before the content
	// writer is spawned — instead of at save_page_sections, the last step,
	// where the same refusal today costs a completed LLM chain first. A
	// setting, not a data reference, so it lives in ConfigKeys; note that the
	// RFC_022 optional-key budget counts Optional only, a trade-off named in
	// the council submission for this change. Opt-in with the unsafe default
	// OFF (owner ruling 2026-08-02 §2): of the two live carriers of this
	// action, page-build-handler opts in (migration 488) and
	// tool-recreation-handler — the tool pipeline, the legitimate owner of
	// owned pages — must never carry it.
	ConfigKeys: []string{"refuse_owned_page"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_page_record", LoadPageRecordInputSpec)
}

// Names that indicate the audit finding is about a new page or site-wide issue,
// not an existing page in the database.
var nonPageNames = map[string]bool{
	"new page needed": true,
	"site-wide":       true,
	"new page":        true,
	"general":         true,
	"":                true,
}

func LoadPageRecordAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_page_record"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		LoadPageRecordInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	// Priority 0: an authoritative id supplied by the dispatcher from the work
	// item's page_id column. When present it IS the page this step is about —
	// the name inputs describe where a defect was seen, not what to load.
	if authIDStr := strings.TrimSpace(inputs.Get("authoritative_page_id")); authIDStr != "" {
		authID, err := uuid.Parse(authIDStr)
		if err != nil {
			// The only wiring for this input is a work-item column, which is a
			// uuid or absent — a malformed value means a misconfigured path, and
			// falling through to the name would silently load a different page,
			// which is the exact defect this input exists to close.
			return nil, fmt.Errorf("invalid authoritative_page_id %q: %w", authIDStr, err)
		}
		res, err := loadPageRecordByID(ctx, params, siteID, authID, logger)
		if err != nil {
			return nil, err
		}
		return refuseOwnedPageIfConfigured(ctx, params, res, siteID, logger)
	}

	pageName := strings.TrimSpace(inputs.Get("page_name"))

	// Contract-compliant fallback: if page_name didn't resolve from config,
	// look for it under input_data.spec.page_name (primary contract path)
	// or other common locations. This protects against workflow configs
	// that use stale path references.
	if pageName == "" {
		fallbackPaths := []string{
			"input_data.spec.page_name",
			"input_data.spec.page.name",
			"current_page.name",
			"page_record.name",
		}
		for _, path := range fallbackPaths {
			if v := datahelpers.ExtractNestedField(params.CollectedData, path); v != nil {
				if s, ok := v.(string); ok && s != "" {
					pageName = strings.TrimSpace(s)
					logger.Info("LoadPageRecordAction: page_name recovered via fallback path",
						zap.String("path", path),
						zap.String("page_name", pageName))
					break
				}
			}
		}
	}

	if pageName == "" {
		logger.Warn("LoadPageRecordAction: page_name not found in any expected location",
			zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)))
		// Let the existing nonPageNames check handle empty string below
	}

	pageIDInput := strings.TrimSpace(inputs.Get("page_id"))

	// Check for non-page names (audit findings that describe gaps, not existing pages)
	if pageName != "" && nonPageNames[strings.ToLower(pageName)] {
		if pageIDInput != "" {
			// Have a real page_id — clear the bogus name and fall through to ID lookup
			logger.Info("LoadPageRecordAction: page_name is not a real page, falling back to page_id",
				zap.String("page_name", pageName),
				zap.String("page_id", pageIDInput))
			pageName = ""
		} else {
			logger.Info("LoadPageRecordAction: page_name is not a real page",
				zap.String("page_name", pageName))
			return map[string]interface{}{
				"found":     false,
				"page_name": pageName,
				"reason":    "page_name is a description, not an existing page",
			}, nil
		}
	}

	// Need at least one identifier
	if pageName == "" && pageIDInput == "" {
		return nil, fmt.Errorf("either page_name or page_id is required")
	}

	res, err := queryPageRecordRow(ctx, params.DB, siteID, pageName, pageIDInput, logger)
	if err != nil {
		return nil, err
	}
	return refuseOwnedPageIfConfigured(ctx, params, res, siteID, logger)
}

// refuseOwnedPageIfConfigured is the opt-in EARLY ownership refusal
// (bugs_open/301). The owned-page guard's refusal already exists and is not in
// question — but in page-build-handler it lives at save_page_sections, the LAST
// step, so a build aimed at a rebuild_policy='owned' page runs the
// page-content-writer LLM chain to completion and is only then refused (39
// discarded chains in 2.5h measured on one site, 2026-08-18). rebuild_policy is
// knowable the moment this action has loaded the row; with refuse_owned_page
// set, the refusal happens here instead and the writer is never spawned.
//
// The three pieces are deliberately the guard's own, not copies:
//   - pageIsOwnedForGuard — the single ownership predicate (owned_page_guard.go;
//     a second predicate is the drift class the council's reuse_agent seat
//     already objected to once);
//   - emitOwnedPageReviewItem — same deduped owned_page_review row the save
//     path files, refused_by='load_page_record';
//   - ownedPageSkipReasonPrefix LEADING the error — routeToErrorStep copies the
//     message verbatim into __step_error.message, which is what
//     update_work_item_status' owned_page_refusal_status (migration 480) matches
//     to stamp the item wont_fix rather than failed. An early refusal composes
//     with that Tier 1 change with zero new vocabulary, via the step's
//     error_step (migration 488 sets both the key and the routing).
//
// Fail-open when the policy cannot be read, matching the predicate's own
// posture — and cheaply here: the save-path guard is still downstream as the
// backstop (removing it would re-open bugs_closed/295), so an unchecked page
// costs at most one wasted chain, never a clobber. The unset key is the
// unchanged default: every caller that does not name it — today that is
// tool-recreation-handler, the tool pipeline, which legitimately builds owned
// pages — behaves exactly as before (owner ruling 2026-08-02 §2).
func refuseOwnedPageIfConfigured(ctx context.Context, params ActionParams, res interface{}, siteID uuid.UUID, logger *zap.Logger) (interface{}, error) {
	refuse, _ := params.StepConfig.Config["refuse_owned_page"].(bool)
	if !refuse {
		return res, nil
	}
	m, ok := res.(map[string]interface{})
	if !ok {
		return res, nil
	}
	if found, _ := m["found"].(bool); !found {
		return res, nil
	}
	pageIDStr, _ := m["id"].(string)
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		// No usable id on a found row is a shape this action never produces;
		// stand down rather than block — the save-path guard still holds.
		logger.Warn("LoadPageRecordAction: refuse_owned_page set but page id unparsable — ownership unchecked here, save-path guard remains",
			zap.String("page_id", pageIDStr))
		return res, nil
	}
	pageName, _ := m["name"].(string)

	refused, class, checked := pageRefusesGenericBuild(ctx, params.DB, pageID, logger)
	if !checked {
		logger.Warn("LoadPageRecordAction: page build policy unreadable — early guard standing down, save-path guard remains",
			zap.String("page_id", pageID.String()),
			zap.String("page_name", pageName))
		return res, nil
	}
	if !refused {
		return res, nil
	}

	// This is the seam that matters most for bugs_open/450's BACKLOG. A write-time
	// door cannot touch the ~339 unbuilt_internal_link items already queued against
	// shell pages; they arrive here, and refusing at load spends no LLM tokens on a
	// page whose save would be refused anyway.
	reason := fmt.Sprintf(
		"%s: page %s is rebuild_policy=owned (tool/widget-owned); refused at load, before the "+
			"content writer runs — a generic section save would be refused at save_page_sections "+
			"anyway, after the LLM work was already spent. Use apply_section_edit for targeted "+
			"edits or the tool pipeline for rebuilds.",
		ownedPageSkipReasonPrefix, pageName)
	if class == refusalToolPending {
		reason = fmt.Sprintf(
			"%s: page %s is page_type=tool with no tool component; refused at load, before the "+
				"content writer runs — building it generically would publish prose about a tool "+
				"that is not there. The tool pipeline creates the component itself, after which "+
				"this refusal lifts by itself.",
			ownedPageSkipReasonPrefix, pageName)
	}

	logger.Warn("LoadPageRecordAction: PAGE REFUSES GENERIC BUILD — refused before the writer",
		zap.String("page_name", pageName),
		zap.String("page_id", pageID.String()),
		zap.String("refusal_class", class))

	emitOwnedPageReviewItem(ctx, params.DB, siteID, pageName, "load_page_record", reason, class, logger)

	return nil, fmt.Errorf("%s", reason)
}

// loadPageRecordByID serves the authoritative_page_id path: lookup strictly by
// id, never consulting the name inputs. Split out so the main flow's page_name
// fallback recovery (which would re-fill an emptied name from the spec and win
// the priority race all over again) cannot reach it.
//
// Zero rows is a FATAL ERROR here, not the {found:false} soft miss the
// name-based flow returns (council round 1 on bugs_open/220, bug_historian,
// HIGH). The id comes from the work item's page_id column, so a valid uuid
// matching no row means the target was deleted (or belongs to another site) —
// a state the caller cannot remedy by building anything. A soft miss would
// route the saga through the success-labelled complete_error path, which is
// the exact silent no-op shape this input exists to close for its own target
// case; failing loudly routes the item into the dispatcher's error handling
// instead.
func loadPageRecordByID(ctx context.Context, params ActionParams, siteID, pageID uuid.UUID, logger *zap.Logger) (interface{}, error) {
	logger.Info("LoadPageRecordAction: authoritative page_id supplied — loading by id, name inputs ignored",
		zap.String("authoritative_page_id", pageID.String()))
	res, err := queryPageRecordRow(ctx, params.DB, siteID, "", pageID.String(), logger)
	if err != nil {
		return nil, err
	}
	if m, ok := res.(map[string]interface{}); ok {
		if found, _ := m["found"].(bool); !found {
			return nil, fmt.Errorf("authoritative_page_id %s names no page row for site %s — the id came from the work item's page_id column, so this target is deleted or foreign, and falling back to a name lookup would load a different page", pageID, siteID)
		}
	}
	return res, nil
}

// queryPageRecordRow runs the page lookup — by name when pageName is non-empty,
// by id otherwise — and builds the page_record result map. Shared by the normal
// name-first flow and the authoritative-id path above.
func queryPageRecordRow(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, pageIDInput string, logger *zap.Logger) (interface{}, error) {
	var (
		pageID           string
		name             string
		title            sql.NullString
		pageType         sql.NullString
		sectionsJSON     sql.NullString
		url              sql.NullString
		buildStatus      sql.NullString
		navLabel         sql.NullString
		navOrder         sql.NullInt32
		contentDirection sql.NullString
	)

	// content_direction: optional per-page writer steering (bug 025). Nullable jsonb;
	// reaches page-content-writer as .current_page.content_direction because
	// page-build-handler maps current_page <- page_record (this row).
	selectCols := `SELECT id::text, name, title, page_type, sections::text, url, build_status, nav_label, nav_order, content_direction::text FROM pages`

	var err error
	if pageName != "" {
		err = db.QueryRowContext(ctx,
			selectCols+` WHERE site_id = $1 AND name = $2 LIMIT 1`,
			siteID, pageName,
		).Scan(&pageID, &name, &title, &pageType,
			&sectionsJSON, &url, &buildStatus, &navLabel, &navOrder, &contentDirection)
	} else {
		// Lookup by page_id — the work item column always has this
		err = db.QueryRowContext(ctx,
			selectCols+` WHERE site_id = $1 AND id = $2::uuid LIMIT 1`,
			siteID, pageIDInput,
		).Scan(&pageID, &name, &title, &pageType,
			&sectionsJSON, &url, &buildStatus, &navLabel, &navOrder, &contentDirection)
	}

	if err == sql.ErrNoRows {
		logger.Info("LoadPageRecordAction: page not found",
			zap.String("site_id", siteID.String()),
			zap.String("page_name", pageName),
			zap.String("page_id", pageIDInput))
		return map[string]interface{}{
			"found":     false,
			"page_name": pageName,
			"site_id":   siteID.String(),
			"reason":    "page not found in database",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query page: %w", err)
	}

	// Parse sections JSON into an array
	var sections []interface{}
	if sectionsJSON.Valid && sectionsJSON.String != "" {
		if err := json.Unmarshal([]byte(sectionsJSON.String), &sections); err != nil {
			logger.Warn("LoadPageRecordAction: failed to parse sections JSON",
				zap.String("sections_raw", sectionsJSON.String), zap.Error(err))
			sections = []interface{}{}
		}
	}

	result := map[string]interface{}{
		"found":         true,
		"id":            pageID,
		"name":          name,
		"site_id":       siteID.String(),
		"sections":      sections,
		"section_count": len(sections),
	}

	if title.Valid {
		result["title"] = title.String
	}
	if pageType.Valid {
		result["page_type"] = pageType.String
	}
	if url.Valid {
		result["url"] = url.String
	}
	if buildStatus.Valid {
		result["build_status"] = buildStatus.String
	}
	if navLabel.Valid {
		result["nav_label"] = navLabel.String
	}
	if navOrder.Valid {
		result["nav_order"] = navOrder.Int32
	}
	// Parse per-page content_direction (bug 025). Only set the key when a value is
	// present, so the writer's {{if .current_page.content_direction}} guard stays false
	// for the (currently every) page that has none.
	if contentDirection.Valid && contentDirection.String != "" {
		var cd interface{}
		if err := json.Unmarshal([]byte(contentDirection.String), &cd); err != nil {
			logger.Warn("LoadPageRecordAction: failed to parse content_direction JSON",
				zap.String("content_direction_raw", contentDirection.String), zap.Error(err))
		} else if cd != nil {
			result["content_direction"] = cd
		}
	}

	logger.Info("LoadPageRecordAction: page loaded",
		zap.String("page_id", pageID),
		zap.String("page_name", name),
		zap.String("page_type", pageType.String),
		zap.Int("sections", len(sections)),
	)

	return result, nil
}
