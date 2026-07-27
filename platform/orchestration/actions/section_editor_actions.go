// FILE: platform/orchestration/actions/section_editor_actions.go
//
// Actions for the section-editor agent. Enables granular edits to individual
// page sections without re-running the full content generation pipeline.
//
// Two actions:
//   - load_edit_context:  Gathers the target section + component + page metadata
//   - apply_section_edit: Performs the edit, updates page_components, reassembles page HTML
//
// Edit types supported:
//   - content_edit:    Update content_data fields, re-render from template + DB context
//   - component_swap:  Change component template, re-render with existing content_data
//
// Design principle: content_data is the source of truth. Every edit updates
// content_data first, then re-renders from template. This means edits survive
// future re-renders (nav updates, theme changes, etc.).

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ===========================================================================
// INPUT SPECS
// ===========================================================================

var LoadEditContextInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"page_component_id", "page_name", "slot_name", "domain"},
	Defaults: map[string]interface{}{},
	Deprecated: map[string]string{
		"site_id_field": "site_id",
	},
}

var ApplySectionEditInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"edit_type"},
	Optional: []string{
		"field_updates",            // content_edit: JSON object of fields to merge
		"replacement_content_data", // content_edit or component_swap: full replacement content_data
		"new_component_function",   // component_swap
		"page_component_id",        // target (can also come from edit_context)
	},
	Defaults: map[string]interface{}{},
	Deprecated: map[string]string{
		"edit_type_field": "edit_type",
		"content_data":    "replacement_content_data", // backward compat: old callers sending content_data
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_edit_context", LoadEditContextInputSpec)
	datahelpers.RegisterActionInputSpec("apply_section_edit", ApplySectionEditInputSpec)
}

// ===========================================================================
// ACTION: load_edit_context
// ===========================================================================
//
// Loads everything needed to perform or plan an edit on a page section.
//
// Target identification (one of):
//   - page_component_id: direct UUID of the page_components row
//   - page_name + slot_name: look up by page name and section slot
//
// Returns:
//   - page_component: {id, page_id, slot_name, position, rendered_html, content_data, component_id}
//   - component:      {id, function, name, html_template, input_schema} (from content_components)
//   - page:           {id, name, title, url, filename, domain, site_id, meta_desc}
//   - site_id, domain

func LoadEditContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("LoadEditContextAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadEditContextInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// --- Identify target page_component ---
	var pcRow pageComponentRow

	if pcIDStr := inputs.Get("page_component_id"); pcIDStr != "" {
		pcID, err := uuid.Parse(pcIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid page_component_id: %w", err)
		}
		pcRow, err = loadPageComponentByID(ctx, params.DB, pcID)
		if err != nil {
			return nil, fmt.Errorf("page_component not found: %w", err)
		}
	} else {
		pageName := inputs.Get("page_name")
		slotName := inputs.Get("slot_name")
		if pageName == "" || slotName == "" {
			return nil, fmt.Errorf("need either page_component_id or both page_name + slot_name")
		}
		// Normalize slot_name per naming contract
		slotName = NormalizeComponentFunction(slotName)
		pcRow, err = loadPageComponentBySlot(ctx, params.DB, siteID, pageName, slotName)
		if err != nil {
			return nil, fmt.Errorf("page_component not found for %s/%s: %w", pageName, slotName, err)
		}
	}

	// --- Load page info ---
	pageInfo, err := getPageInfo(ctx, params.DB, pcRow.PageID)
	if err != nil {
		return nil, fmt.Errorf("failed to load page info: %w", err)
	}

	// --- Load component template (if linked) ---
	var componentData map[string]interface{}
	if pcRow.ComponentID != nil {
		componentData, err = loadComponentForEdit(ctx, params.DB, *pcRow.ComponentID)
		if err != nil {
			logger.Warn("LoadEditContextAction: Failed to load component template",
				zap.String("component_id", pcRow.ComponentID.String()),
				zap.Error(err),
			)
		}
	}

	// If no component_id, try lookup by slot_name
	if componentData == nil && pcRow.SlotName != "" {
		componentData, err = loadComponentByFunction(ctx, params.DB, pcRow.SlotName)
		if err != nil {
			logger.Debug("LoadEditContextAction: No component found for slot_name",
				zap.String("slot_name", pcRow.SlotName),
			)
		}
	}

	// --- Build result ---
	pcMap := map[string]interface{}{
		"id":            pcRow.ID.String(),
		"page_id":       pcRow.PageID.String(),
		"slot_name":     pcRow.SlotName,
		"position":      pcRow.Position,
		"rendered_html": pcRow.RenderedHTML,
		"content_data":  pcRow.ContentData,
		"build_status":  pcRow.BuildStatus,
	}
	if pcRow.ComponentID != nil {
		pcMap["component_id"] = pcRow.ComponentID.String()
	}

	pageMap := map[string]interface{}{
		"id":        pageInfo.ID.String(),
		"name":      pageInfo.Name,
		"title":     pageInfo.Title,
		"url":       pageInfo.URL,
		"filename":  pageInfo.Filename,
		"domain":    pageInfo.Domain,
		"site_id":   pageInfo.SiteID.String(),
		"meta_desc": pageInfo.MetaDesc,
	}

	logger.Info("LoadEditContextAction: Complete",
		zap.String("page_component_id", pcRow.ID.String()),
		zap.String("page_name", pageInfo.Name),
		zap.String("slot_name", pcRow.SlotName),
		zap.Bool("has_component_template", componentData != nil),
		zap.Bool("has_content_data", pcRow.ContentData != nil),
	)

	return map[string]interface{}{
		"success":        true,
		"page_component": pcMap,
		"component":      componentData, // may be nil
		"page":           pageMap,
		"site_id":        siteID.String(),
		"domain":         pageInfo.Domain,
	}, nil
}

// ===========================================================================
// ACTION: apply_section_edit
// ===========================================================================
//
// Performs the edit on a page_components row, then reassembles the full page.
//
// Reads edit_context from collected_data (output of load_edit_context).
//
// Edit types:
//   - content_edit:    Updates content_data (merge or replace), then re-renders
//                      the component template with full site context from DB.
//                      content_data is the source of truth — this ensures edits
//                      survive future re-renders.
//   - component_swap:  Changes the component template, re-renders with existing
//                      content_data + site context.
//
// After editing:
//   1. Update content_data in page_components
//   2. Re-render component template with updated content_data + site context
//   3. UPDATE page_components.rendered_html
//   4. Reassemble full page via assemblePage()
//   5. Return assembled HTML + metadata for git_commit

func ApplySectionEditAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("ApplySectionEditAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ApplySectionEditInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	editType := inputs.Get("edit_type")
	if editType == "" {
		return nil, fmt.Errorf("edit_type is required (content_edit or component_swap)")
	}

	// --- Load edit context from collected_data ---
	editCtx, ok := params.CollectedData["edit_context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("edit_context not found in collected_data — run load_edit_context first")
	}

	pcData, _ := editCtx["page_component"].(map[string]interface{})
	if pcData == nil {
		return nil, fmt.Errorf("edit_context.page_component is missing")
	}

	pageData, _ := editCtx["page"].(map[string]interface{})
	if pageData == nil {
		return nil, fmt.Errorf("edit_context.page is missing")
	}

	pcIDStr, _ := pcData["id"].(string)
	pcID, err := uuid.Parse(pcIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page_component id: %w", err)
	}

	pageIDStr, _ := pageData["id"].(string)
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page id: %w", err)
	}

	siteIDStr, _ := editCtx["site_id"].(string)
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id in edit_context: %w", err)
	}

	slotName, _ := pcData["slot_name"].(string)
	pageName := getStringVal(pageData, "name")

	// ── Lock gate (bugs_open/058) ────────────────────────────────────────────
	// A human-locked component must not be overwritten by an automated edit
	// (this action is reachable from the tool-improver loop, not only from a
	// human request). Skip-result, not error: an error would fail/retry the
	// orchestration for a state only a human unlock can change. The check is
	// advisory for messaging; the race-free enforcement is the lock predicate
	// on the UPDATEs themselves (errComponentLocked below). A check failure is
	// non-fatal for the same reason.
	if lock, lockErr := CheckComponentLock(ctx, params.DB, pcID, logger); lockErr != nil {
		logger.Warn("ApplySectionEditAction: component lock check failed — relying on guarded UPDATE",
			zap.Error(lockErr))
	} else if lock.IsLocked {
		emitLockBlockedChangeItem(ctx, params.DB, siteID, &pageID, &pcID,
			pageName, slotName, lock.LockedBy, lock.LockType,
			"edit", "apply_section_edit", logger)
		logger.Warn("ApplySectionEditAction: refusing to edit human-locked component (bugs_open/058)",
			zap.String("page_component_id", pcIDStr),
			zap.String("slot_name", slotName),
			zap.String("locked_by", lock.LockedBy),
		)
		return map[string]interface{}{
			"success": true,
			"skipped": true,
			"locked":  true,
			"reason": fmt.Sprintf("component %s (%s) is locked by %q — unlock via the admin dashboard to edit it",
				pcIDStr, slotName, lock.LockedBy),
		}, nil
	}

	// --- Apply the edit ---
	var newHTML string
	var newContentData map[string]interface{}

	switch editType {
	case "content_edit":
		newHTML, newContentData, err = applyContentEdit(ctx, params.DB, siteID, pcData, editCtx, inputs, logger)
	case "component_swap":
		newHTML, newContentData, err = applyComponentSwap(ctx, params.DB, siteID, pcData, editCtx, inputs, logger)
	default:
		return nil, fmt.Errorf("unknown edit_type: %s (expected: content_edit, component_swap)", editType)
	}

	if err != nil {
		if errors.Is(err, errComponentLocked) {
			// Race window: the lock landed between the pre-check and the write.
			return map[string]interface{}{
				"success": true,
				"skipped": true,
				"locked":  true,
				"reason":  fmt.Sprintf("component %s (%s) is locked — unlock via the admin dashboard to edit it", pcIDStr, slotName),
			}, nil
		}
		return nil, fmt.Errorf("edit failed (%s): %w", editType, err)
	}

	// --- Update page_components row ---
	// Note: component_swap already updates the row (including component_id),
	// but content_edit needs this separate update
	if editType == "content_edit" {
		err = updatePageComponentAfterEdit(ctx, params.DB, pcID, newHTML, newContentData)
		if errors.Is(err, errComponentLocked) {
			return map[string]interface{}{
				"success": true,
				"skipped": true,
				"locked":  true,
				"reason":  fmt.Sprintf("component %s (%s) is locked — unlock via the admin dashboard to edit it", pcIDStr, slotName),
			}, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed to update page_component: %w", err)
		}
	}

	logger.Info("ApplySectionEditAction: Updated page_component",
		zap.String("page_component_id", pcIDStr),
		zap.String("edit_type", editType),
		zap.String("slot_name", slotName),
		zap.Int("new_html_length", len(newHTML)),
	)

	// --- Reassemble full page ---
	pageInfo, err := getPageInfo(ctx, params.DB, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to load page info for reassembly: %w", err)
	}

	fullHTML, assembly, err := assemblePage(ctx, params.DB, pageInfo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to reassemble page: %w", err)
	}

	// bugs_open/095, same class one surface over: this path returned
	// success=true with html="" if reassembly produced nothing — i.e. it would
	// report that a section edit had been applied while handing the deployer an
	// empty page.
	//
	// It uses the SAME discriminator as the re-renderer rather than failing on
	// any empty reassembly. The first version of this fix failed
	// unconditionally, on the reasoning that a section of this page has just
	// been edited so its components necessarily exist. Two council seats
	// objected at medium severity that this was asserted rather than evidenced,
	// and that an error with no escape hatch is a sharper change than the
	// re-render side, which keeps a legitimate-skip branch. They were right:
	// where the two surfaces agree on the rule, the rule is the thing that has
	// been reviewed. And it costs nothing here — the edit target IS a component
	// row, so ComponentRows > 0 holds whenever the edit actually happened.
	if fullHTML == "" && assembly.assembledToNothingDespiteComponents() {
		return nil, fmt.Errorf(
			"page %q reassembled to nothing after the section edit — %s",
			pageInfo.Name, assembly.describe())
	}

	domain, _ := pageData["domain"].(string)

	logger.Info("ApplySectionEditAction: Page reassembled",
		zap.String("page_name", pageInfo.Name),
		zap.String("filename", pageInfo.Filename),
		zap.Int("full_html_length", len(fullHTML)),
	)

	return map[string]interface{}{
		"success":           true,
		"html":              fullHTML,
		"domain":            domain,
		"filename":          pageInfo.Filename,
		"page_id":           pageIDStr,
		"page_name":         pageInfo.Name,
		"page_component_id": pcIDStr,
		"edit_type":         editType,
		"slot_name":         slotName,
	}, nil
}

// ===========================================================================
// buildRenderContextFromDB
// ===========================================================================
//
// Builds a full RenderContext from database state alone — no collected_data
// pipeline needed. This is the key function that enables section editing
// without re-running the entire content generation pipeline.
//
// Loads:
//   - Site data (company name, domain, email, phone, logo)
//   - Style collection (colors)
//   - CSS theme (theme_css)
//   - Navigation items (header + footer)
//   - Page metadata (title, description, current page)
//   - Content data (from page_components.content_data → RenderContext.ContentData)

func buildRenderContextFromDB(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageInfo *PageInfo,
	contentData map[string]interface{},
	logger *zap.Logger,
) (*RenderContext, error) {

	// 1. Load site data
	siteData, err := loadSiteDataFull(ctx, db, siteID, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load site data: %w", err)
	}

	// 2. Load style collection for colors
	var primaryColor, secondaryColor, accentColor string
	coll, err := GetStyleCollectionForSite(ctx, db, siteID, logger)
	if err != nil {
		logger.Warn("buildRenderContextFromDB: No style collection found, using defaults",
			zap.Error(err))
	} else if coll != nil && coll.ColorPalette != nil {
		primaryColor = coll.ColorPalette["primary"]
		secondaryColor = coll.ColorPalette["secondary"]
		accentColor = coll.ColorPalette["accent"]
	}

	// 3. Load CSS theme
	var themeCSS string
	if coll != nil && coll.CSSThemeID != nil {
		theme, err := getThemeByID(ctx, db, *coll.CSSThemeID, logger)
		if err != nil {
			logger.Warn("buildRenderContextFromDB: Failed to load CSS theme",
				zap.Error(err))
		} else if theme != nil {
			themeCSS = theme.CSSContent
		}
	}

	// 4. Load navigation
	navItems := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, logger)
	footerNavItems := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary, NavGroupUtility, NavGroupLegal}, NavFetchableOnly, 0, logger)

	// 5. Derive page-specific fields
	currentPage := ""
	if pageInfo != nil {
		currentPage = strings.TrimSuffix(pageInfo.Filename, ".html")
	}
	year := fmt.Sprintf("%d", time.Now().Year())

	// 6. Build the RenderContext
	renderCtx := &RenderContext{
		Domain:         siteData.Domain,
		SiteID:         siteID,
		CompanyName:    siteData.CompanyName,
		LogoText:       siteData.LogoText,
		LogoURL:        siteData.LogoURL,
		Tagline:        siteData.Tagline,
		Email:          siteData.Email,
		Phone:          siteData.Phone,
		PrimaryColor:   primaryColor,
		SecondaryColor: secondaryColor,
		AccentColor:    accentColor,
		ThemeCSS:       themeCSS,
		NavItems:       setActiveNavItems(navItems, currentPage),
		FooterNavItems: footerNavItems,
		CurrentPage:    currentPage,
		Year:           year,
	}

	if pageInfo != nil {
		renderCtx.Title = pageInfo.Title
		renderCtx.Description = pageInfo.MetaDesc
	}

	// Fallback email from domain
	if renderCtx.Email == "" && renderCtx.Domain != "" {
		renderCtx.Email = "info@" + renderCtx.Domain
	}

	// 7. Set ContentData — this is what templates use for section-specific content.
	//    contextToInterfaceMap() merges ContentData into the top-level template data,
	//    so templates access these as {{.headline}}, {{.features}}, etc.
	renderCtx.ContentData = make(map[string]interface{})

	// Site-level fields that templates might reference
	renderCtx.ContentData["company_name"] = siteData.CompanyName
	renderCtx.ContentData["brand_name"] = siteData.CompanyName
	renderCtx.ContentData["tagline"] = siteData.Tagline
	renderCtx.ContentData["domain"] = siteData.Domain
	renderCtx.ContentData["email"] = siteData.Email
	renderCtx.ContentData["contact_email"] = siteData.Email
	renderCtx.ContentData["phone"] = siteData.Phone
	renderCtx.ContentData["logo_text"] = siteData.LogoText
	renderCtx.ContentData["logo_url"] = siteData.LogoURL
	renderCtx.ContentData["year"] = year
	renderCtx.ContentData["copyright"] = fmt.Sprintf("© %s %s", year, siteData.CompanyName)

	// Navigation in the formats templates expect
	categories := make([]map[string]interface{}, len(navItems))
	for i, item := range navItems {
		categories[i] = map[string]interface{}{
			"name":  item.Label,
			"slug":  strings.TrimSuffix(strings.TrimPrefix(item.URL, "/"), ".html"),
			"url":   item.URL,
			"label": item.Label,
		}
	}
	renderCtx.ContentData["categories"] = categories
	renderCtx.ContentData["nav_items"] = categories

	footerLinks := make([]map[string]interface{}, len(footerNavItems))
	for i, item := range footerNavItems {
		footerLinks[i] = map[string]interface{}{
			"name":  item.Label,
			"slug":  strings.TrimSuffix(strings.TrimPrefix(item.URL, "/"), ".html"),
			"url":   item.URL,
			"label": item.Label,
		}
	}
	renderCtx.ContentData["footer_nav_items"] = footerLinks
	renderCtx.ContentData["quick_links"] = footerLinks

	// CTA defaults
	renderCtx.ContentData["cta_text"] = "Get Started"
	renderCtx.ContentData["cta_url"] = "/contact.html"

	// Now merge the actual section content_data on top — these take priority
	// over site-level defaults. This is where headline, features[],
	// testimonials[], body text etc. come from.
	for key, value := range contentData {
		renderCtx.ContentData[key] = value
	}

	logger.Info("buildRenderContextFromDB: Context built",
		zap.String("domain", siteData.Domain),
		zap.String("company", siteData.CompanyName),
		zap.Bool("has_colors", primaryColor != ""),
		zap.Bool("has_theme_css", themeCSS != ""),
		zap.Int("nav_items", len(navItems)),
		zap.Int("content_data_fields", len(contentData)),
	)

	return renderCtx, nil
}

// ===========================================================================
// EDIT IMPLEMENTATIONS
// ===========================================================================

// applyContentEdit updates content_data and re-renders the component template
// with full site context from DB.
//
// Two modes for specifying the update:
//   - field_updates: merge specific fields into existing content_data
//     (e.g. change headline, update a phone number)
//   - content_data:  replace entire content_data with new object
//     (e.g. rewrite a whole case study)
//
// After updating content_data, loads the component template and builds a
// full RenderContext from DB state, then calls RenderTemplate.
func applyContentEdit(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pcData map[string]interface{},
	editCtx map[string]interface{},
	inputs *datahelpers.ActionInputs,
	logger *zap.Logger,
) (string, map[string]interface{}, error) {

	// --- Build updated content_data ---
	existingContentData := make(map[string]interface{})
	if cd, ok := pcData["content_data"].(map[string]interface{}); ok {
		for k, v := range cd {
			existingContentData[k] = v
		}
	}

	// Check field_updates first (merge mode — more common, more specific)
	// This is checked BEFORE replacement_content_data because ExtractActionInputs
	// nested lookup can pick up site_record.content_data (the site plan) as a
	// false match for any field named "content_data".
	if fieldUpdates := inputs.GetRaw("field_updates"); fieldUpdates != nil {
		// Merge mode — update specific fields
		var updates map[string]interface{}
		switch v := fieldUpdates.(type) {
		case map[string]interface{}:
			updates = v
		case string:
			if err := json.Unmarshal([]byte(v), &updates); err != nil {
				return "", nil, fmt.Errorf("field_updates must be valid JSON object: %w", err)
			}
		default:
			return "", nil, fmt.Errorf("field_updates must be a JSON object, got %T", fieldUpdates)
		}
		for k, v := range updates {
			existingContentData[k] = v
			logger.Debug("applyContentEdit: Merged field", zap.String("field", k))
		}
		logger.Info("applyContentEdit: Merged field_updates into content_data",
			zap.Int("updated_fields", len(updates)),
			zap.Int("total_fields", len(existingContentData)))
	} else if fullReplace := inputs.GetRaw("replacement_content_data"); fullReplace != nil {
		// Full replacement mode
		switch v := fullReplace.(type) {
		case map[string]interface{}:
			existingContentData = v
			logger.Info("applyContentEdit: Full content_data replacement",
				zap.Int("field_count", len(v)))
		case string:
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err != nil {
				return "", nil, fmt.Errorf("replacement_content_data must be valid JSON object: %w", err)
			}
			existingContentData = parsed
			logger.Info("applyContentEdit: Full content_data replacement (from JSON string)",
				zap.Int("field_count", len(parsed)))
		}
	} else {
		return "", nil, fmt.Errorf("content_edit requires either 'field_updates' or 'replacement_content_data' parameter")
	}

	// --- Get component template ---
	componentData, _ := editCtx["component"].(map[string]interface{})
	if componentData == nil {
		return "", nil, fmt.Errorf("no component template available — cannot re-render (component_id may be NULL)")
	}
	htmlTemplate, _ := componentData["html_template"].(string)
	if htmlTemplate == "" {
		return "", nil, fmt.Errorf("component template is empty — cannot re-render")
	}

	// --- Build render context from DB ---
	pageData, _ := editCtx["page"].(map[string]interface{})
	var pageInfoForRender *PageInfo
	if pageData != nil {
		pageID, _ := uuid.Parse(getStringVal(pageData, "id"))
		pageInfoForRender = &PageInfo{
			ID:       pageID,
			Name:     getStringVal(pageData, "name"),
			Title:    getStringVal(pageData, "title"),
			Filename: getStringVal(pageData, "filename"),
			MetaDesc: getStringVal(pageData, "meta_desc"),
			Domain:   getStringVal(pageData, "domain"),
			SiteID:   siteID,
		}
	}

	renderCtx, err := buildRenderContextFromDB(ctx, db, siteID, pageInfoForRender, existingContentData, logger)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build render context: %w", err)
	}

	// --- Render ---
	rendered := RenderTemplate(htmlTemplate, renderCtx, logger)
	if rendered == "" {
		return "", nil, fmt.Errorf("template rendering produced empty output")
	}

	logger.Info("applyContentEdit: Re-rendered component from template",
		zap.Int("output_length", len(rendered)),
		zap.Int("content_data_fields", len(existingContentData)),
	)

	return rendered, existingContentData, nil
}

// applyComponentSwap changes the component template for this section.
// Looks up the new component, then re-renders with existing content_data
// using full site context from DB.
func applyComponentSwap(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pcData map[string]interface{},
	editCtx map[string]interface{},
	inputs *datahelpers.ActionInputs,
	logger *zap.Logger,
) (string, map[string]interface{}, error) {

	newFunction := inputs.Get("new_component_function")
	if newFunction == "" {
		return "", nil, fmt.Errorf("component_swap requires 'new_component_function' parameter")
	}

	// Normalize per naming contract
	newFunction = NormalizeComponentFunction(newFunction)

	// Look up the new component
	comp, err := GetComponentWithFallback(ctx, db, newFunction, logger)
	if err != nil {
		return "", nil, fmt.Errorf("component %q not found: %w", newFunction, err)
	}

	if comp.HTMLTemplate == "" {
		return "", nil, fmt.Errorf("component %q has no HTML template", newFunction)
	}

	// Determine content_data to render with.
	// Priority: replacement_content_data from caller > existing content_data from page_component
	contentData := make(map[string]interface{})

	if rcd := inputs.GetMap("replacement_content_data"); rcd != nil {
		// Caller provided replacement content (component_swap with new data)
		for k, v := range rcd {
			contentData[k] = v
		}
		logger.Info("applyComponentSwap: Using replacement_content_data from caller",
			zap.Int("field_count", len(rcd)))
	} else if cd, ok := pcData["content_data"].(map[string]interface{}); ok {
		// No replacement — use existing content_data from page_component
		for k, v := range cd {
			contentData[k] = v
		}
		logger.Info("applyComponentSwap: Using existing content_data from page_component",
			zap.Int("field_count", len(cd)))
	}

	// Build render context from DB with existing content_data
	pageData, _ := editCtx["page"].(map[string]interface{})
	var pageInfoForRender *PageInfo
	if pageData != nil {
		pageID, _ := uuid.Parse(getStringVal(pageData, "id"))
		pageInfoForRender = &PageInfo{
			ID:       pageID,
			Name:     getStringVal(pageData, "name"),
			Title:    getStringVal(pageData, "title"),
			Filename: getStringVal(pageData, "filename"),
			MetaDesc: getStringVal(pageData, "meta_desc"),
			Domain:   getStringVal(pageData, "domain"),
			SiteID:   siteID,
		}
	}

	renderCtx, err := buildRenderContextFromDB(ctx, db, siteID, pageInfoForRender, contentData, logger)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build render context for swap: %w", err)
	}

	rendered := RenderTemplate(comp.HTMLTemplate, renderCtx, logger)
	if rendered == "" {
		return "", nil, fmt.Errorf("template rendering produced empty output after swap")
	}

	logger.Info("applyComponentSwap: Swapped and re-rendered component",
		zap.String("old_slot", getStringVal(pcData, "slot_name")),
		zap.String("new_function", comp.Function),
		zap.String("new_component_id", comp.ID),
		zap.Int("output_length", len(rendered)),
	)

	// Update component_id, slot_name, rendered_html, content_data in DB
	compID, _ := uuid.Parse(comp.ID)
	err = updatePageComponentSwap(ctx, db,
		mustParseUUID(getStringVal(pcData, "id")),
		compID,
		comp.Function,
		rendered,
		contentData,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to update page_component for swap: %w", err)
	}

	return rendered, contentData, nil
}

// ===========================================================================
// DB HELPERS
// ===========================================================================

type pageComponentRow struct {
	ID           uuid.UUID
	PageID       uuid.UUID
	ComponentID  *uuid.UUID
	Position     int
	SlotName     string
	RenderedHTML string
	ContentData  map[string]interface{}
	BuildStatus  string
}

func loadPageComponentByID(ctx context.Context, db *sql.DB, id uuid.UUID) (pageComponentRow, error) {
	var row pageComponentRow
	var componentID sql.NullString
	var contentDataJSON []byte

	err := db.QueryRowContext(ctx, `
		SELECT id, page_id, component_id, position, slot_name,
		       COALESCE(rendered_html, ''), COALESCE(content_data, '{}'::jsonb),
		       COALESCE(build_status, 'pending')
		FROM page_components
		WHERE id = $1
	`, id).Scan(
		&row.ID, &row.PageID, &componentID, &row.Position, &row.SlotName,
		&row.RenderedHTML, &contentDataJSON, &row.BuildStatus,
	)
	if err != nil {
		return row, err
	}

	if componentID.Valid {
		parsed, _ := uuid.Parse(componentID.String)
		row.ComponentID = &parsed
	}

	if len(contentDataJSON) > 0 {
		json.Unmarshal(contentDataJSON, &row.ContentData)
	}

	return row, nil
}

// Problem: page_components rows sometimes have empty slot_name (e.g. placeholder rows,
// builds where metadata path fell through to HTML parsing with no data-component attribute).
// The current query only matches pc.slot_name = $3, so these rows are invisible to
// section-editor.
//
// Fix: Add two fallback match paths, preferring direct match:
//   1. pc.slot_name = $3                     (direct match — best)
//   2. cc.function = $3                       (component knows its function)
//   3. pages.sections[position-1] = $3        (plan-based position match)
//
// When a fallback match succeeds, the function also backfills the empty slot_name
// so subsequent queries use the direct path.
//
// ALSO: Returns the matched slot_name (COALESCE) so the caller always sees it.

func loadPageComponentBySlot(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, slotName string) (pageComponentRow, error) {
	row, err := loadPageComponentBySlotRO(ctx, db, siteID, pageName, slotName)
	if err != nil {
		return row, err
	}

	// Backfill: if we matched via fallback, update the slot_name in DB
	// so future queries use the direct path. Non-blocking — log and continue.
	if row.SlotName == slotName {
		// Check if the DB value was actually empty (we COALESCEd it)
		var dbSlotName sql.NullString
		db.QueryRowContext(ctx, `SELECT slot_name FROM page_components WHERE id = $1`, row.ID).Scan(&dbSlotName)
		if !dbSlotName.Valid || dbSlotName.String == "" || dbSlotName.String != slotName {
			_, _ = db.ExecContext(ctx, `UPDATE page_components SET slot_name = $2 WHERE id = $1`, row.ID, slotName)
		}
	}

	return row, nil
}

// loadPageComponentBySlotRO is loadPageComponentBySlot's matching logic with NO
// write side-effects — same three-way match (slot_name, component function,
// plan position), no slot_name backfill.
//
// Split out for read-only callers that must not mutate what they are inspecting:
// revalidate_review_queue sweeps the parked review queue in a dry_run mode whose
// whole contract is "report, change nothing", and the backfill above would break
// that. Keeping ONE copy of the match logic is the point — a second hand-rolled
// slot lookup is precisely the drift this repo keeps paying for (bugs_closed/041,
// section lookup never normalising).
func loadPageComponentBySlotRO(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, slotName string) (pageComponentRow, error) {
	var row pageComponentRow
	var componentID sql.NullString
	var contentDataJSON []byte

	err := db.QueryRowContext(ctx, `
		SELECT pc.id, pc.page_id, pc.component_id, pc.position,
		       COALESCE(NULLIF(pc.slot_name, ''), $3) as slot_name,
		       COALESCE(pc.rendered_html, ''),
		       COALESCE(pc.content_data, '{}'::jsonb),
		       COALESCE(pc.build_status, 'pending')
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		LEFT JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1 AND p.name = $2
		  AND (
		    pc.slot_name = $3
		    OR cc.function = $3
		    OR (
		      (pc.slot_name IS NULL OR pc.slot_name = '')
		      AND p.sections IS NOT NULL
		      AND pc.position > 0
		      AND pc.position <= jsonb_array_length(p.sections)
		      AND trim(both '"' from (p.sections->(pc.position - 1))::text) = $3
		    )
		  )
		ORDER BY
		  CASE WHEN pc.slot_name = $3 THEN 0
		       WHEN cc.function = $3 THEN 1
		       ELSE 2
		  END
		LIMIT 1
	`, siteID, pageName, slotName).Scan(
		&row.ID, &row.PageID, &componentID, &row.Position, &row.SlotName,
		&row.RenderedHTML, &contentDataJSON, &row.BuildStatus,
	)
	if err != nil {
		return row, err
	}

	if componentID.Valid {
		parsed, _ := uuid.Parse(componentID.String)
		row.ComponentID = &parsed
	}

	if len(contentDataJSON) > 0 {
		json.Unmarshal(contentDataJSON, &row.ContentData)
	}

	return row, nil
}

func loadComponentForEdit(ctx context.Context, db *sql.DB, componentID uuid.UUID) (map[string]interface{}, error) {
	var id, function, name string
	var htmlTemplate string
	var inputSchemaJSON []byte

	err := db.QueryRowContext(ctx, `
		SELECT id::text, function, name, html_template, COALESCE(input_schema, '{}'::jsonb)
		FROM content_components
		WHERE id = $1
	`, componentID).Scan(&id, &function, &name, &htmlTemplate, &inputSchemaJSON)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":            id,
		"function":      function,
		"name":          name,
		"html_template": htmlTemplate,
	}

	if len(inputSchemaJSON) > 0 {
		var schema interface{}
		if json.Unmarshal(inputSchemaJSON, &schema) == nil {
			result["input_schema"] = schema
		}
	}

	return result, nil
}

func loadComponentByFunction(ctx context.Context, db *sql.DB, function string) (map[string]interface{}, error) {
	var id, funcVal, name string
	var htmlTemplate string
	var inputSchemaJSON []byte

	err := db.QueryRowContext(ctx, `
		SELECT id::text, function, name, html_template, COALESCE(input_schema, '{}'::jsonb)
		FROM content_components
		WHERE function = $1 AND (is_active = true OR is_active IS NULL)
		LIMIT 1
	`, function).Scan(&id, &funcVal, &name, &htmlTemplate, &inputSchemaJSON)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":            id,
		"function":      funcVal,
		"name":          name,
		"html_template": htmlTemplate,
	}

	if len(inputSchemaJSON) > 0 {
		var schema interface{}
		if json.Unmarshal(inputSchemaJSON, &schema) == nil {
			result["input_schema"] = schema
		}
	}

	return result, nil
}

// errComponentLocked is returned by the page_components UPDATE helpers when
// the row exists but carries an active human lock (bugs_open/058). The
// lock predicate lives in the UPDATE's WHERE clause, so the refusal is
// race-free; callers convert this to a skip-result rather than a failure.
var errComponentLocked = errors.New("page component is locked or missing — automated edit refused (bugs_open/058)")

func updatePageComponentAfterEdit(ctx context.Context, db *sql.DB, pcID uuid.UUID, html string, contentData map[string]interface{}) error {
	var contentDataJSON []byte
	var err error
	var res sql.Result

	if contentData != nil {
		contentDataJSON, err = json.Marshal(contentData)
		if err != nil {
			return fmt.Errorf("failed to marshal content_data: %w", err)
		}
	}

	if contentDataJSON != nil {
		res, err = db.ExecContext(ctx, `
			UPDATE page_components
			SET rendered_html = $2,
			    content_data = $3::jsonb,
			    build_status = 'approved',
			    updated_at = NOW()
			WHERE id = $1 AND `+pageComponentAgentWritableSQL("")+`
		`, pcID, html, string(contentDataJSON))
	} else {
		res, err = db.ExecContext(ctx, `
			UPDATE page_components
			SET rendered_html = $2,
			    build_status = 'approved',
			    updated_at = NOW()
			WHERE id = $1 AND `+pageComponentAgentWritableSQL("")+`
		`, pcID, html)
	}

	if err != nil {
		return err
	}
	if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
		return errComponentLocked
	}
	return nil
}

func updatePageComponentSwap(ctx context.Context, db *sql.DB, pcID, componentID uuid.UUID, newSlotName, html string, contentData map[string]interface{}) error {
	contentDataJSON, err := json.Marshal(contentData)
	if err != nil {
		return fmt.Errorf("failed to marshal content_data: %w", err)
	}

	res, err := db.ExecContext(ctx, `
		UPDATE page_components
		SET component_id = $2,
		    slot_name = $3,
		    rendered_html = $4,
		    content_data = $5::jsonb,
		    build_status = 'approved',
		    updated_at = NOW()
		WHERE id = $1 AND `+pageComponentAgentWritableSQL("")+`
	`, pcID, componentID, newSlotName, html, string(contentDataJSON))

	if err != nil {
		return err
	}
	if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
		return errComponentLocked
	}
	return nil
}

// ===========================================================================
// UTILITY HELPERS
// ===========================================================================

func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

func getStringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
