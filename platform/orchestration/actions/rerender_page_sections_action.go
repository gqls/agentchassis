// FILE: platform/orchestration/actions/rerender_page_sections_action.go
//
// RerenderPageSectionsAction re-renders ALL of a page's sections from their
// STORED content_data plus FRESHLY re-resolved dynamic fields, WITHOUT invoking
// the content writer (no LLM). It is the lightweight path for "a resolved field
// changed, re-render the page" — specifically an image asset landing
// (hero/section image) or deferred section data becoming query-resolvable —
// where the page's copy is unchanged and only resolved fields (asset URLs,
// query-backed lists) need refreshing.
//
// WHY (and why not the writer): routing these through page-build-handler ->
// page-content-writer regenerates copy via the LLM (cost, latency) and exposes
// an asset swap to the content-regression guard. The writer already proved
// content_data is a complete render source: RenderComponentAction persists
// content_data = LLM copy overlaid with resolved_data. So we re-render each
// section by feeding STORED content_data as the content and FRESH resolved_data
// on top, through the same RenderTemplate path render_component uses.
//
// MECHANISM (all reuse, in-package):
//   - newSourceResolver + planSection (plan_sections_action.go) rebuild each
//     section's resolved_data (queryresolve for query.*, the page-aware
//     ensureAssets for the hero). planSection is SIDE-EFFECT-FREE — the
//     needs_new_component / deferred-item writes live in PlanSectionsAction's
//     caller loop, not in planSection itself.
//   - RenderTemplate renders html_template against a RenderContext built from a
//     minimal site base + stored content_data + fresh resolved_data
//     (resolved_data merged last so it wins, matching RenderComponentAction).
//   - Emits sections_metadata in the exact shape CompilePageSectionsAction
//     produces ({rendered_html, component_id, component_name,
//     component_function, content_data}), so save_page_sections ingests it
//     unchanged — no compile step needed.
//
// NULL content_data (older pages that predate the writer's content_data
// capture): re-render-all needs EVERY section to have stored content. If ANY
// section's content_data is missing, this escalates the WHOLE page to the
// content generator (emits needs_page -> page-build-handler), which regenerates
// AND backfills content_data; the light path then works on the next trigger.
//
// OUTPUT (output_field, e.g. "rerender_sections"):
//   { sections_metadata, page_id, site_id, domain, page_name,
//     escalated (bool), skipped (bool), section_count, rerendered, carried }
// page_id/site_id/domain are surfaced so the downstream render_page
// (rerender_single_page) finds them via ExtractFields' recursive search — the
// work item itself only carries page_name.
//
// WIRING (page-rerender workflow, as a pre-pass gated by spec.reason):
//   check_rerender_mode (conditional: reason==image_landed OR
//     reason==section_data_resolved) -> rerender_sections -> check_escalated
//     -> save_sections -> render_page -> (check_skipped -> deploy -> status)
//   else_step (no/other reason) -> render_page   (unchanged assemble-only path)
//
// REGISTRATION (registry.go):
//   "rerender_page_sections": {
//       Handler:     RerenderPageSectionsAction,
//       Category:    "site",
//       Description: "Re-render a page's sections from stored content_data + fresh resolved fields (no LLM)",
//       IsLocal:     true,
//   }

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

var RerenderPageSectionsInputSpec = datahelpers.ActionInputSpec{
	// target_site_id (NOT site_id) per 001 §Field name collisions: site_id is a
	// key on the nested-source objects (site_record.site_id, input_data.site_id),
	// so a bare site_id can be silently bound from the wrong source. The wiring
	// maps it explicitly: "target_site_id": "input_data.site_id". Same precedent
	// as reconcile_site_plan.
	// page_name moved Required -> Optional (bugs_open/094). Callers that know the
	// page by ID had no way in: the required-field check rejected the envelope
	// before the action ran, with "missing required fields: [page_name]", even
	// though the action's very next act is a DB lookup that could have resolved
	// it. 049b_deploy_single_page.sh publishes page_id/site_id/domain and could
	// not use its own documented section_data_resolved branch at all.
	//
	// EITHER page_name OR page_id is now sufficient and the action derives the
	// other. That is what makes the bad envelope unrepresentable rather than
	// merely documented — every existing caller is fixed at once, including any
	// not yet found.
	Required: []string{"target_site_id"},
	Optional: []string{"page_name", "page_id", "reason"},
	Defaults: map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("rerender_page_sections", RerenderPageSectionsInputSpec)
}

// storedSection is one page_components row as loaded for re-render.
type storedSection struct {
	componentID  string
	slotName     string
	contentData  map[string]interface{}
	renderedHTML string
	position     int
}

func RerenderPageSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "rerender_page_sections"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		RerenderPageSectionsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("target_site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid target_site_id %q: %w", inputs.Get("target_site_id"), err)
	}
	pageName := inputs.Get("page_name")
	pageIDIn := inputs.Get("page_id")
	if pageName == "" && pageIDIn == "" {
		return nil, fmt.Errorf("need page_name or page_id (bugs_open/094): got neither")
	}
	reason := inputs.Get("reason")

	// ── Resolve page_id + domain (also surfaced for the downstream render_page,
	//    which finds them via ExtractFields' recursive search); page url feeds
	//    the CTA recompute's self-link test ────────────────────────────────────
	//
	// bugs_open/094: resolve by whichever key the caller supplied. Both branches
	// are scoped to the site, so a page_id belonging to a DIFFERENT site is a
	// not-found rather than a cross-site re-render — target_site_id stays
	// authoritative and page_id cannot be used to reach past it.
	var pageID uuid.UUID
	var domain, pageURL string
	if pageName != "" {
		err = params.DB.QueryRowContext(ctx, `
			SELECT p.id, s.domain, COALESCE(p.url, ''), p.name
			FROM pages p
			JOIN sites s ON s.id = p.site_id
			WHERE p.site_id = $1 AND p.name = $2
		`, siteID, pageName).Scan(&pageID, &domain, &pageURL, &pageName)
	} else {
		parsed, perr := uuid.Parse(pageIDIn)
		if perr != nil {
			return nil, fmt.Errorf("invalid page_id %q: %w", pageIDIn, perr)
		}
		err = params.DB.QueryRowContext(ctx, `
			SELECT p.id, s.domain, COALESCE(p.url, ''), p.name
			FROM pages p
			JOIN sites s ON s.id = p.site_id
			WHERE p.site_id = $1 AND p.id = $2
		`, siteID, parsed).Scan(&pageID, &domain, &pageURL, &pageName)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			// Name the key that was actually used, so "not found" is actionable
			// rather than ambiguous about which lookup failed.
			if pageIDIn != "" && inputs.Get("page_name") == "" {
				return nil, fmt.Errorf("page_id %s not found for site %s", pageIDIn, siteID)
			}
			return nil, fmt.Errorf("page %q not found for site %s", pageName, siteID)
		}
		return nil, fmt.Errorf("resolve page: %w", err)
	}

	// bugs_open/094, council objection (bug_historian, medium): page_id is not
	// mapped by any step config, so it arrives via ExtractActionInputs Strategy 2,
	// whose own comment warns it "uses aggressive recursive search that can find
	// stale values". Before this change a missing page_name failed LOUDLY at the
	// input-spec gate; now a stale page_id could in principle resolve a DIFFERENT
	// page of the same site and re-render it silently.
	//
	// The site scoping stops it crossing sites; it does not stop it picking the
	// wrong page WITHIN a site. So record which key was used and what it resolved
	// to. This does not prevent the wrong resolution — it makes it attributable
	// instead of invisible, which is the difference between a bug you can find
	// and one you cannot.
	resolvedBy := "page_name"
	if inputs.Get("page_name") == "" {
		resolvedBy = "page_id (recursive search — no step config maps it)"
	}
	logger.Info("rerender_page_sections: page resolved",
		zap.String("resolved_by", resolvedBy),
		zap.String("page_name", pageName),
		zap.String("page_id", pageID.String()),
		zap.String("site_id", siteID.String()),
		zap.String("url", pageURL))

	out := map[string]interface{}{
		"page_id":   pageID.String(),
		"site_id":   siteID.String(),
		"domain":    domain,
		"page_name": pageName,
		"escalated": false,
		"skipped":   false,
	}

	// ── Load stored sections ────────────────────────────────────────────────
	stored, err := loadStoredSections(ctx, params.DB, pageID, logger)
	if err != nil {
		return nil, fmt.Errorf("load stored sections: %w", err)
	}
	if len(stored) == 0 {
		logger.Info("rerender_page_sections: page has no stored components, nothing to re-render",
			zap.String("page", pageName))
		out["skipped"] = true
		out["section_count"] = 0
		return out, nil
	}

	// One component-schema load for all sections (loadComponentSchemas keys by
	// both name and function — slot_name matches either). Loaded before the
	// pre-check so the check can consult each section's required-field contract.
	names := make([]string, 0, len(stored))
	for _, s := range stored {
		names = append(names, s.slotName)
	}
	schemas := loadComponentSchemas(ctx, params.DB, names, logger)

	// ── Content pre-check: re-render-all renders each section from its STORED
	//    content_data, so every section must actually carry its content. Two
	//    ways it can fail, both of which would render an empty section and
	//    OVERWRITE good HTML with a blank shell (the exact defect that silently
	//    blanked live article bodies):
	//      (a) content_data is entirely absent (older pages predating capture);
	//      (b) content_data is present but MISSING a schema-required source:"llm"
	//          field — e.g. the stored {type,result} envelope that was never
	//          unwrapped, so it has no `content` key the article-body template
	//          needs.
	//    In either case escalate the WHOLE page to the writer (regenerate +
	//    backfill) and do NOT re-render here, leaving the existing HTML intact. ─
	for _, s := range stored {
		// A self-contained TOOL section legitimately has no content_data: a tool
		// is complete HTML with no LLM-authored fields, so content_data={} is
		// its correct shape, not the missing-content defect this pre-check
		// exists to catch. Escalating it bypasses save_sections — the ONLY
		// writer of rendered_html — so the re-render is computed and thrown
		// away, and a durable template fix never reaches the page
		// (bugs_open/024).
		//
		// Keyed on the EXPLICIT component_level='tool' marker plus an empty
		// input_schema, never on a heuristic about field shape: predicating on
		// "has no required LLM fields" would also exempt components declaring
		// OPTIONAL source:"llm" fields, a broader class than is justified.
		if comp, ok := schemas[s.slotName]; ok && isSelfContainedSection(comp) {
			logger.Info("rerender_page_sections: self-contained tool section, no content_data expected — rendering from template",
				zap.String("page", pageName),
				zap.String("section", s.slotName))
			continue
		}

		reason := ""
		if len(s.contentData) == 0 {
			reason = "no stored content_data"
		} else if comp, ok := schemas[s.slotName]; ok && len(comp.InputSchema) > 0 {
			// The rerender path reaches the gate WITHOUT a plan_sections pass, so
			// this is where a re-rendered legacy-dialect component would otherwise
			// be enforced silently — fire the fail-loud tripwire here.
			datahelpers.WarnIfLegacyDialect(comp.InputSchema, logger, "render-gate", comp.Function)
			if missing := missingRequiredLLMFields(comp.InputSchema, s.contentData); len(missing) > 0 {
				reason = fmt.Sprintf("stored content_data missing required field(s) %v", missing)
			}
		}
		if reason != "" {
			logger.Warn("rerender_page_sections: section content incomplete — escalating page to writer instead of blanking it",
				zap.String("page", pageName),
				zap.String("section", s.slotName),
				zap.String("reason", reason))
			if err := escalateRerenderToWriter(ctx, params.DB, siteID, pageName, logger); err != nil {
				return nil, fmt.Errorf("escalate to writer: %w", err)
			}
			out["escalated"] = true
			out["section_count"] = len(stored)
			return out, nil
		}
	}

	// ── Re-resolve + re-render each section (no LLM) ────────────────────────
	resolver := newSourceResolver(siteID, params.DB, logger, pageName)

	// Minimal render-context base from sites.content_data (company/contact/etc).
	// Section templates take colours from CSS vars and copy from content_data,
	// so this base is small; it only matters for a section that reads an ambient
	// field. Built once and re-merged per section.
	baseData := buildRerenderBaseData(ctx, params.DB, siteID, domain, pageName, logger)

	sectionsMetadata := make([]map[string]interface{}, 0, len(stored))
	reRendered := 0
	carried := 0

	// CTA targets for reason=cta_links_stale, loaded once on first CTA section.
	var cta *rerenderCTAState

	for _, s := range stored {
		comp, haveComp := schemas[s.slotName]

		// Can't load the component → carry the stored HTML untouched.
		if !haveComp {
			logger.Warn("rerender_page_sections: component not found, carrying stored HTML",
				zap.String("section", s.slotName))
			sectionsMetadata = append(sectionsMetadata, carryStoredSection(s))
			carried++
			continue
		}

		// Reuse planSection to rebuild resolved_data (side-effect-free).
		plan := planSection(ctx, s.slotName, comp, resolver, logger)
		if plan.Status != "ready" {
			// A required non-LLM field can't resolve now — carry the stored HTML
			// rather than render a broken/empty section.
			logger.Info("rerender_page_sections: section not ready, carrying stored HTML",
				zap.String("section", s.slotName),
				zap.String("status", plan.Status))
			sectionsMetadata = append(sectionsMetadata, carryStoredSection(s))
			carried++
			continue
		}

		htmlTemplate, _ := comp.Raw["html_template"].(string)
		if htmlTemplate == "" {
			logger.Warn("rerender_page_sections: empty html_template, carrying stored HTML",
				zap.String("section", s.slotName))
			sectionsMetadata = append(sectionsMetadata, carryStoredSection(s))
			carried++
			continue
		}

		// CTA recompute — ONLY for reason=cta_links_stale, so image_landed /
		// section_data_resolved rerenders behave byte-identically to before.
		// After migrations 091/098 the schema no longer sources CTA urls, so a
		// stale url survives in stored content_data; writing the recomputed
		// target into plan.ResolvedData wins the merge below (resolved_data last).
		if reason == "cta_links_stale" {
			fn := comp.Function
			if fn == "" {
				fn = s.slotName
			}
			if fields, isCTA := ctaFieldNames[fn]; isCTA {
				if cta == nil {
					cta = loadRerenderCTAState(ctx, params, siteID, pageName, logger)
				}
				if plan.ResolvedData == nil {
					plan.ResolvedData = map[string]interface{}{}
				}
				applyCTARecompute(plan.ResolvedData, s.contentData, fields[0], cta.primary, cta.validPages, pageURL)
				applyCTARecompute(plan.ResolvedData, s.contentData, fields[1], cta.secondary, cta.validPages, pageURL)
			}
		}

		// OBSERVE-ONLY (council trail 2525f980): this merge is where a
		// resolver-written CTA destination is actually lost — stored
		// content_data (the resolver's last write) merges FIRST, fresh
		// plan.ResolvedData merges LAST and wins. Log each derived CTA field
		// where the fresh value would replace a differing stored one,
		// carrying the rerender reason so deliberate cta_links_stale
		// recomputes are distinguishable from silent clobbers. No behaviour
		// change; the precedence flip returns to the council gate with this
		// log as its evidence. (An earlier sketch placed this log inside
		// planSection, where resolvedData is a fresh local map and the
		// condition could never fire — doc_notes correction, b6e374fc2.)
		if schema := datahelpers.ParseInputSchemaValue(comp.Raw["input_schema"]); schema != nil {
			for _, cf := range datahelpers.DeriveCTAURLFields(schema) {
				stored, hasStored := s.contentData[cf.URLField]
				fresh, hasFresh := plan.ResolvedData[cf.URLField]
				if hasStored && hasFresh && stored != fresh {
					logger.Info("rerender_page_sections: cta ownership conflict (observe-only)",
						zap.String("section", s.slotName),
						zap.String("field", cf.URLField),
						zap.String("source", cf.Source),
						zap.String("reason", reason))
				}
			}
		}

		// Render context: base ⊕ stored content_data ⊕ fresh resolved_data
		// (resolved_data merged last so it overrides stale values — matching
		// RenderComponentAction's content_from-then-merge_with ordering).
		rc := &RenderContext{Year: fmt.Sprintf("%d", time.Now().Year())}
		mergeIntoRenderContext(rc, baseData)
		mergeIntoRenderContext(rc, s.contentData)
		if plan.ResolvedData != nil {
			mergeIntoRenderContext(rc, plan.ResolvedData)
		}
		if rc.ContentData == nil {
			rc.ContentData = make(map[string]interface{})
		}
		rc.ContentData["ComponentID"] = comp.ID

		rendered := RenderTemplate(htmlTemplate, rc, logger)

		// Persisted content_data = stored ⊕ fresh resolved_data, mirroring
		// RenderComponentAction so the row remains a complete render source.
		mergedContent := make(map[string]interface{}, len(s.contentData)+len(plan.ResolvedData))
		for k, v := range s.contentData {
			mergedContent[k] = v
		}
		for k, v := range plan.ResolvedData {
			mergedContent[k] = v
		}

		entry := map[string]interface{}{
			"rendered_html":      rendered,
			"component_name":     s.slotName,
			"component_function": comp.Function,
			"content_data":       mergedContent,
		}
		if comp.ID != "" {
			entry["component_id"] = comp.ID
		} else if s.componentID != "" {
			entry["component_id"] = s.componentID
		}
		sectionsMetadata = append(sectionsMetadata, entry)
		reRendered++
	}

	out["sections_metadata"] = sectionsMetadata
	out["section_count"] = len(sectionsMetadata)
	out["rerendered"] = reRendered
	out["carried"] = carried

	logger.Info("rerender_page_sections: done",
		zap.String("page", pageName),
		zap.String("reason", reason),
		zap.Int("sections", len(sectionsMetadata)),
		zap.Int("rerendered", reRendered),
		zap.Int("carried", carried))

	return out, nil
}

// rerenderCTAState holds the per-page CTA targets for a cta_links_stale
// recompute — computed once, shared by every CTA section on the page.
type rerenderCTAState struct {
	primary, secondary contentHub
	validPages         datahelpers.PageURLSet
}

// loadRerenderCTAState reuses the internal-link-resolver's loaders and ranking
// (interactive pages first, then hubs). Loader failures degrade to empty
// candidate lists: applyCTARecompute then leaves fields untouched rather than
// aborting the rerender.
func loadRerenderCTAState(ctx context.Context, params ActionParams, siteID uuid.UUID, pageName string, logger *zap.Logger) *rerenderCTAState {
	hubs, err := loadContentHubs(ctx, params, siteID, logger)
	if err != nil {
		logger.Warn("rerender_page_sections: loadContentHubs failed for CTA recompute", zap.Error(err))
	}
	interactive, err := loadInteractivePages(ctx, params, siteID, logger)
	if err != nil {
		logger.Warn("rerender_page_sections: loadInteractivePages failed for CTA recompute", zap.Error(err))
	}
	primary, secondary := chooseCTATargets("", pageName, interactive, hubs)
	return &rerenderCTAState{
		primary:    primary,
		secondary:  secondary,
		validPages: loadResolverPageSet(ctx, params, siteID, logger),
	}
}

// applyCTARecompute writes the recomputed CTA target into resolved (which the
// caller merges LAST, beating stale stored content_data) — unless the stored
// value is an explicitly authored link worth keeping. Keep = resolves to a
// real page AND is not an excluded destination (contact/legal/about) AND is
// not the page linking to itself. Everything else — phantom, empty, circular,
// or excluded — is replaced, provided a valid target exists.
func applyCTARecompute(resolved, stored map[string]interface{}, field string, target contentHub,
	validPages datahelpers.PageURLSet, pageURL string) {

	if field == "" {
		return // single-URL component — no field in this slot
	}
	if current, ok := stored[field].(string); ok && current != "" &&
		validPages.Contains(current) &&
		!ctaExcludedDestination(current) &&
		datahelpers.NormalizePagePath(current) != datahelpers.NormalizePagePath(pageURL) {
		return // authored link to a real, sensible destination — keep it
	}

	if target.URL == "" || !validPages.Contains(target.URL) {
		return // nothing valid to write — leave the field as stored
	}
	resolved[field] = target.URL
	if title := targetTitle(target); title != "" {
		resolved[ctaTargetTitleField(field)] = title
	}
}

// loadStoredSections reads the page's current page_components rows in order.
func loadStoredSections(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) ([]storedSection, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(component_id::text, ''),
		       COALESCE(slot_name, ''),
		       content_data,
		       COALESCE(rendered_html, ''),
		       position
		FROM page_components
		WHERE page_id = $1
		ORDER BY position ASC
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []storedSection
	for rows.Next() {
		var s storedSection
		var cdJSON []byte
		if err := rows.Scan(&s.componentID, &s.slotName, &cdJSON, &s.renderedHTML, &s.position); err != nil {
			logger.Warn("rerender_page_sections: row scan failed", zap.Error(err))
			continue
		}
		if len(cdJSON) > 0 {
			_ = json.Unmarshal(cdJSON, &s.contentData)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// carryStoredSection builds a sections_metadata entry that re-emits a section's
// stored render unchanged (used when a section can't be safely re-rendered).
// Shape matches CompilePageSectionsAction so save_page_sections ingests it.
func carryStoredSection(s storedSection) map[string]interface{} {
	m := map[string]interface{}{
		"rendered_html":  s.renderedHTML,
		"component_name": s.slotName,
	}
	if s.componentID != "" {
		m["component_id"] = s.componentID
	}
	if len(s.contentData) > 0 {
		m["content_data"] = s.contentData
	}
	return m
}

// buildRerenderBaseData assembles the minimal ambient render-context data from
// sites.content_data (top-level + reviewed_brief), plus domain and year. The
// section templates examined take colours from CSS vars and copy from
// content_data, so this base is rarely read — it only covers a section that
// references an ambient field (company name, contact, etc).
func buildRerenderBaseData(ctx context.Context, db *sql.DB, siteID uuid.UUID, domain string, pageName string, logger *zap.Logger) map[string]interface{} {
	base := map[string]interface{}{
		"domain": domain,
		"year":   fmt.Sprintf("%d", time.Now().Year()),
	}

	// The page's own identity. bugs_open/085 was fixed on the page-BUILD path
	// (BuildRenderContextAction); this is the same defect on the scoped
	// section-re-render path, and it was found by firing a re-render on the
	// fixed binary and watching a page still render three charts assigned to
	// two different pages. The page name was already in scope here — it is
	// passed to newSourceResolver on the line above the call — so the identity
	// was available and simply never reached the render base.
	//
	// mergeIntoRenderContext restores this into RenderContext.CurrentPage, so
	// setting the map key is all that is needed. Trimmed to match
	// buildHeaderConfig, which is the form every other producer uses.
	if pageName != "" {
		base["current_page"] = strings.TrimSuffix(pageName, ".html")
	} else {
		logger.Warn("rerender_page_sections: no page name for the render base — every section will see an empty current_page and cannot vary per page (bugs_open/085)")
	}

	// Read content_data AND the canonical sites.email COLUMN together. The
	// full-writer render path (loadSiteDataFull) sources ctx.Email from the
	// sites.email column; this light path historically read only
	// content_data.email, which is empty or stale on most sites (idea.uk held a
	// stale idea-uk@leopardess.uk while its column carried the current address).
	// So a section re-render could not convert a dead contact form to a mailto
	// even where the site had a real address. We now prefer the column, applied
	// AFTER the content_data merge so it wins — making both render paths agree.
	// See bugs_open/006 §B.
	var cdJSON []byte
	var siteEmail string
	if err := db.QueryRowContext(ctx, `SELECT content_data, COALESCE(email, '') FROM sites WHERE id = $1`, siteID).Scan(&cdJSON, &siteEmail); err != nil {
		if err != sql.ErrNoRows {
			logger.Warn("rerender_page_sections: load sites row failed", zap.Error(err))
		}
	} else if len(cdJSON) > 0 {
		var cd map[string]interface{}
		if err := json.Unmarshal(cdJSON, &cd); err != nil {
			logger.Warn("rerender_page_sections: parse sites.content_data failed", zap.Error(err))
		} else {
			// reviewed_brief first so its keys are present, then top-level wins on overlap.
			if rb, ok := cd["reviewed_brief"].(map[string]interface{}); ok {
				for k, v := range rb {
					if _, exists := base[k]; !exists {
						base[k] = v
					}
				}
			}
			for k, v := range cd {
				if k == "reviewed_brief" {
					continue
				}
				base[k] = v
			}
		}
	}

	// The sites.email column is canonical (matches loadSiteDataFull) — applied
	// last so it overrides any stale/empty content_data.email merged above.
	// The sites.email column is canonical (matches loadSiteDataFull) — applied
	// last so it overrides any stale/empty content_data.email merged above.
	if siteEmail != "" {
		base["email"] = siteEmail
		base["contact_email"] = siteEmail
	}

	return base
}

// escalateRerenderToWriter emits a needs_page work item so page-build-handler
// rebuilds the page through the writer (regenerate + backfill content_data).
// Keyed needs_page:<page> so it co-dedups with reconcile_site_plan's items.
// isSelfContainedSection reports whether a section's component renders entirely
// from its own template, with no LLM-authored content_data to supply.
//
// Both signals are required and both are explicit:
//   - component_level == "tool", the marker set at component creation. It is
//     already SELECTed by loadSectionComponents (COALESCE(component_level,
//     'section')) and carried on componentInfo.Raw, so reading it here costs no
//     extra query and no struct change.
//   - an empty input_schema, i.e. the component declares no content fields at
//     all.
//
// Deliberately NOT a heuristic over field shape (e.g. "has no REQUIRED llm
// fields"), which would also exempt components declaring optional source:"llm"
// fields — a broader class than the evidence justifies. As of 2026-07-20 this
// matches 12 of 122 active components.
func isSelfContainedSection(comp componentInfo) bool {
	if len(comp.InputSchema) > 0 {
		return false
	}
	level, _ := comp.Raw["component_level"].(string)
	return level == "tool"
}

func escalateRerenderToWriter(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string, logger *zap.Logger) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	batchID := uuid.New()
	spec := fmt.Sprintf(`{"reason":"content_data_backfill","page_name":%q}`, pageName)
	if _, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "page-rerender",
		pipeline:     "build",
		itemType:     "needs_page",
		severity:     "medium",
		summary:      fmt.Sprintf("Full rebuild of %s — a section had no stored content_data", pageName),
		spec:         spec,
		priority:     90,
		handlerAgent: "page-build-handler",
		status:       "triaged",
		createdBy:    "page-rerender",
		itemKey:      fmt.Sprintf("needs_page:%s", pageName),
		batchID:      batchID,
	}, logger); err != nil {
		return fmt.Errorf("emit needs_page: %w", err)
	}
	return tx.Commit()
}
