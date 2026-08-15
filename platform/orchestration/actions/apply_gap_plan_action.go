// FILE: platform/orchestration/actions/apply_gap_plan_action.go
//
// Executes the content-gap-planner's LLM decision. Takes the plan output
// and creates the appropriate database records and work items.
//
// Handles five approaches:
//   - add_to_page:     creates content_rewrite work item for existing page
//   - new_page:        creates page record + needs_content_page work item
//   - retype_existing: flips a stranded page's page_type to the routing key
//     the originating check selects on (bugs_open/015)
//   - update_spec:     writes to site_specs via inline query
//   - not_actionable:  marks the original work item as wont_fix
//
// Registration:
//   "apply_gap_plan": {
//       Handler:     ApplyGapPlanAction,
//       Category:    "site",
//       Description: "Execute content gap plan — create pages, work items, or spec updates",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ApplyGapPlanInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"plan", "work_item_id", "domain"},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("apply_gap_plan", ApplyGapPlanInputSpec)
}

func ApplyGapPlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "apply_gap_plan"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		ApplyGapPlanInputSpec, logger,
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

	// Parse the original work_item_id so we can mark it complete/wont_fix
	var originalItemID *uuid.UUID
	if itemIDStr := inputs.Get("work_item_id"); itemIDStr != "" {
		if parsed, err := uuid.Parse(itemIDStr); err == nil {
			originalItemID = &parsed
		}
	}

	// Parse the plan — could be string or map
	planField := "gap_plan.result"
	if p, ok := params.StepConfig.Config["plan"].(string); ok && p != "" {
		planField = p
	}

	planRaw := datahelpers.ExtractNestedField(params.CollectedData, planField)
	if planRaw == nil {
		// Try fallbacks
		for _, alt := range []string{"gap_plan.result", "gap_plan", "plan"} {
			planRaw = datahelpers.ExtractNestedField(params.CollectedData, alt)
			if planRaw != nil {
				break
			}
		}
	}

	if planRaw == nil {
		return map[string]interface{}{
			"applied":  false,
			"approach": "none",
			"reason":   "no plan found",
		}, nil
	}

	// Parse plan into a map
	var plan map[string]interface{}

	switch v := planRaw.(type) {
	case string:
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
		if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
			return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
		}
	case map[string]interface{}:
		plan = v
	default:
		return nil, fmt.Errorf("unexpected plan type: %T", planRaw)
	}

	approach, _ := plan["approach"].(string)
	reasoning, _ := plan["reasoning"].(string)

	logger.Info("ApplyGapPlanAction: executing plan",
		zap.String("approach", approach),
		zap.String("reasoning", reasoning),
		zap.String("site_id", siteIDStr))

	switch approach {

	case "add_to_page":
		return applyAddToPage(ctx, params.DB, plan, siteID, originalItemID, logger)

	case "new_page":
		return applyNewPage(ctx, params.DB, plan, siteID, domain, originalItemID, logger)

	case "retype_existing":
		return applyRetypeExisting(ctx, params.DB, plan, siteID, originalItemID, logger)

	case "update_spec":
		return applyUpdateSpec(ctx, params.DB, plan, siteID, originalItemID, logger)

	case "not_actionable":
		return applyNotActionable(ctx, params.DB, plan, originalItemID, logger)

	default:
		logger.Warn("Unknown plan approach",
			zap.String("approach", approach))
		return map[string]interface{}{
			"applied":  false,
			"approach": approach,
			"reason":   "unknown approach",
		}, nil
	}
}

// ============================================================================
// add_to_page: create a content_rewrite work item for an existing page
// ============================================================================

func applyAddToPage(ctx context.Context, db *sql.DB, plan map[string]interface{}, siteID uuid.UUID, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error) {
	addPlan, ok := plan["add_to_page"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("add_to_page plan missing or invalid")
	}

	pageNameRaw, _ := addPlan["page_name"].(string)
	if pageNameRaw == "" {
		return nil, fmt.Errorf("add_to_page.page_name is required")
	}

	contentGuidance, _ := addPlan["content_guidance"].(string)
	reasoning, _ := plan["reasoning"].(string)

	// LLM sometimes produces comma-separated page names like "index, services".
	// Split and process each one individually.
	pageNames := splitPageNames(pageNameRaw)

	var created int
	var skipped []string

	for _, pageName := range pageNames {
		// Look up the page
		var pageID uuid.UUID
		err := db.QueryRowContext(ctx, `
			SELECT id FROM pages WHERE site_id = $1 AND name = $2 LIMIT 1
		`, siteID, pageName).Scan(&pageID)

		if err == sql.ErrNoRows {
			logger.Warn("Page not found, skipping",
				zap.String("page_name", pageName),
				zap.String("raw_input", pageNameRaw))
			skipped = append(skipped, pageName)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("query page %q: %w", pageName, err)
		}

		// Build spec for the content rewrite
		spec := map[string]interface{}{
			"page_name": pageName,
			// bugs_open/271: "suggestion" is the key every reader reads
			// (page-build-handler's "rewrite_guidance?" mapping →
			// page-content-writer's "## Rewrite Guidance" prompt block). This
			// used to be written as "content_guidance", which nothing read, so
			// the gap planner's whole reason for existing — the composed brief
			// — was discarded before any prompt was built. The value still
			// comes from the gap PLAN's content_guidance at :178: that is the
			// content-gap-planner LLM's output contract and is unchanged.
			"suggestion": contentGuidance,
			"source":     "content-gap-planner",
			// mode=edit_live opts this item into load_current_section_content
			// (bugs_open/178): the page was just resolved by name from the
			// pages table above, so it already exists — the writer should
			// edit its current stored content, not fabricate a replacement
			// with nothing to work from.
			"mode": "edit_live",
		}
		if addSections, ok := addPlan["add_sections"].([]interface{}); ok {
			// Resolve section names to canonical functions before handing
			// them to the build pipeline, so "faq-section"/"FAQ Section"
			// don't orphan downstream.
			resolver := loadComponentNameResolver(ctx, db, logger)
			if len(resolver.validFunctions) > 0 {
				resolved := make([]interface{}, 0, len(addSections))
				for _, s := range addSections {
					name, ok := s.(string)
					if !ok {
						resolved = append(resolved, s)
						continue
					}
					if fn, ok := resolver.resolve(name); ok {
						resolved = append(resolved, fn)
					} else {
						logger.Warn("applyAddToPage: dropped unresolvable section name",
							zap.String("page", pageName), zap.String("section", name))
					}
				}
				spec["add_sections"] = resolved
			} else {
				spec["add_sections"] = addSections
			}
		}
		specJSON, _ := json.Marshal(spec)

		summary := fmt.Sprintf("Add content to %s: %s", pageName, truncate(reasoning, 80))
		itemKey := fmt.Sprintf("gap_plan_add_%s_%s", pageName, siteID)

		_, err = withWorkItemTx(ctx, db, logger, gapPlanWorkItem(
			siteID, "content_rewrite", summary, string(specJSON), &pageID, 35, itemKey, originalItemID))

		if err != nil {
			logger.Warn("Failed to create work item for page",
				zap.String("page_name", pageName), zap.Error(err))
			continue
		}
		created++
	}

	if created == 0 && len(skipped) > 0 {
		return nil, fmt.Errorf("no valid pages found from %q — pages not found: %s", pageNameRaw, strings.Join(skipped, ", "))
	}

	// Mark original as complete
	markOriginalComplete(ctx, db, originalItemID)

	logger.Info("ApplyGapPlanAction: add_to_page applied",
		zap.String("raw_page_names", pageNameRaw),
		zap.Int("items_created", created),
		zap.Strings("skipped", skipped))

	return map[string]interface{}{
		"applied":       true,
		"approach":      "add_to_page",
		"pages":         pageNames,
		"items_created": created,
		"skipped":       skipped,
	}, nil
}

// ============================================================================
// new_page: create page record + needs_content_page work item
// ============================================================================

func applyNewPage(ctx context.Context, db *sql.DB, plan map[string]interface{}, siteID uuid.UUID, domain string, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error) {
	newPlan, ok := plan["new_page"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("new_page plan missing or invalid")
	}

	pageNameRaw, _ := newPlan["name"].(string)
	if pageNameRaw == "" {
		return nil, fmt.Errorf("new_page.name is required")
	}
	pageTypeRaw, _ := newPlan["page_type"].(string)

	// Canonicalise name / url / page_type through the shared helper, so this
	// surface produces the same identity as the planner
	// (write_site_plan_action.go), adoption (apply_adoption_plan_action.go) and
	// the tool path (create_tool_component_action.go) for the same logical page
	// — doc 029 Phase 0, bugs_open/080.
	//
	// Before this, new_page lowercased the LLM's name and synthesised
	// "/<name>.html" by hand, so a gap-planned news listing became
	// name="news" url="/news.html" while every other surface produced
	// name="news-index" url="/news/index.html". Because the INSERT below
	// upserts ON CONFLICT (site_id, name), a disagreeing name does not
	// conflict — it inserts a SECOND row. robot-hands.com carries both, live
	// and both serving 200, one of them orphaned from the nav.
	//
	// For the common content case CanonicalisePage returns exactly what the
	// hand-rolled code produced (name=slug, url=/<slug>.html, type=content),
	// so only the typed roles — the section-index family, tool/guide/game,
	// blog-post, entity-page, and the home/index collapse — change shape.
	pageName, url, pageType := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
		Role:     pageTypeRaw,
		Slug:     pageNameRaw,
		FlatURLs: siteUsesFlatURLs(ctx, db, siteID, logger),
	})
	if pageName == "" {
		return nil, fmt.Errorf("new_page name %q (page_type %q) failed canonicalisation", pageNameRaw, pageTypeRaw)
	}
	if pageName != strings.ToLower(strings.ReplaceAll(pageNameRaw, " ", "-")) {
		logger.Info("applyNewPage: canonicalised page identity",
			zap.String("raw_name", pageNameRaw), zap.String("raw_page_type", pageTypeRaw),
			zap.String("page_name", pageName), zap.String("url", url),
			zap.String("page_type", pageType))
	}

	title, _ := newPlan["title"].(string)
	if title == "" {
		title = strings.Title(strings.ReplaceAll(pageName, "-", " "))
		if domain != "" {
			title = title + " | " + domain
		}
	}

	purpose, _ := newPlan["purpose"].(string)

	// Parse sections
	// Archetype-aware default: a recognised page type gets its archetype
	// shape rather than a generic text block that competes with structured
	// content. Unknown types keep the generic default.
	sections := defaultSectionsForPage(pageName, pageType)
	if sectionsRaw, ok := newPlan["sections"].([]interface{}); ok && len(sectionsRaw) > 0 {
		raw := make([]string, 0, len(sectionsRaw))
		for _, s := range sectionsRaw {
			if ss, ok := s.(string); ok {
				raw = append(raw, ss)
			}
		}
		// Resolve names to canonical functions; drop unresolvable. This
		// catches display names ("FAQ Section") and underscore/case
		// variants before they orphan the page_component downstream.
		resolver := loadComponentNameResolver(ctx, db, logger)
		if len(resolver.validFunctions) > 0 {
			resolved := make([]string, 0, len(raw))
			for _, name := range raw {
				if fn, ok := resolver.resolve(name); ok {
					resolved = append(resolved, fn)
				} else {
					logger.Warn("applyNewPage: dropped unresolvable section name",
						zap.String("page", pageName), zap.String("section", name))
				}
			}
			if len(resolved) > 0 {
				sections = resolved
			}
		} else if len(raw) > 0 {
			sections = raw // resolver unavailable — use as-is rather than lose them
		}
	}
	sectionsJSON, _ := json.Marshal(sections)

	navLabel, _ := newPlan["nav_label"].(string)
	inHeader := true
	if ih, ok := newPlan["in_header"].(bool); ok {
		inHeader = ih
	}
	inFooter := true
	if inf, ok := newPlan["in_footer"].(bool); ok {
		inFooter = inf
	}

	// url comes from CanonicalisePage above — it is no longer synthesised here.

	// Check page growth budget
	budget, budgetErr := CheckPageGrowthBudget(ctx, db, siteID, pageType, logger)
	if budgetErr != nil {
		logger.Warn("Page growth budget check failed, allowing by default", zap.Error(budgetErr))
	} else if !budget.Allowed {
		logger.Info("ApplyGapPlanAction: page creation throttled by growth budget",
			zap.String("page_name", pageName),
			zap.String("reason", budget.Reason),
			zap.Int("current_total", budget.CurrentTotal),
			zap.Int("recent_content", budget.RecentContent))

		// Mark original item as blocked (not failed — can be retried next week)
		if originalItemID != nil {
			db.ExecContext(ctx, `
				UPDATE site_work_items
				SET status = 'blocked',
				    error = $2,
				    updated_at = NOW()
				WHERE id = $1
			`, *originalItemID,
				fmt.Sprintf("Page growth budget: %s (total: %d, weekly content: %d/%d, blog: %d/%d, structural: %d/%d)",
					budget.Reason, budget.CurrentTotal,
					budget.RecentContent, budget.Config.WeeklyContentPagesMax,
					budget.RecentBlog, budget.Config.WeeklyBlogPostsMax,
					budget.RecentStructural, budget.Config.WeeklyStructuralPagesMax))
		}

		return map[string]interface{}{
			"applied":  false,
			"approach": "new_page",
			"reason":   "growth_budget_" + budget.Reason,
			"budget":   budget,
		}, nil
	}

	// Create the page record.
	//
	// bugs_open/081. This was a blind upsert —
	//     ON CONFLICT (site_id, name) DO UPDATE SET title, sections, updated_at
	// — which silently turned a CREATE arm into a PARTIAL update: it overwrote
	// the conflicting row's title and sections while leaving page_type, url and
	// nav at their old values. On a DEPLOYED page that is two defects at once.
	// The live page's content is replaced by a plan written for a role it does
	// not hold; and the mistype the plan existed to repair survives, so the
	// check that asked for the page fires again on the next sweep. Measured
	// looping since 2026-05-01 on ai-agent-orchestration.com — three months,
	// one item that ran out of attempts and one still sitting in 'detected'.
	//
	// So the arm now only ever CREATES, and a name already held by a DEPLOYED
	// page of a DIFFERENT type is refused rather than mutated. Re-typing a live
	// page changes what it serves the moment it happens, and 081 measured that
	// no predicate distinguishes a real news listing from a catalog index that
	// embeds one — robot-hands.com carries both, byte-identical at
	// sections=["news-listing"]. That choice belongs to a human, so this fails
	// closed: refuse, block the originating item with a message that names the
	// conflict, and file it for review.
	//
	// SCOPE IS DELIBERATELY THE SHIPPED CASE ONLY. Measured 2026-07-31 against
	// the live DB: all 5 mistyped pages fleet-wide are `deployed` — 0 `planned`,
	// 0 `needs_rebuild` — so extending the re-type to never-shipped rows would
	// repair nothing that exists while handing a generic arm broad mutation
	// authority for free. That widening is 081's fix candidate 1, and it is
	// declined here on the measurement rather than on taste.
	//
	// > **CORRECTED 2026-08-02: "deployed" was the wrong word for the boundary,
	// > though the boundary itself was right.** The guard tested
	// > `build_status = 'deployed'`, which is not "has this page been served" —
	// > a `needs_rebuild` row is still serving its old artefact
	// > (`bugs_closed/037`; 35 of 46 carry a `deployed_at`). It now asks
	// > `datahelpers.NeverDeployedPagePredicate`, negated. Re-measured on the day
	// > of the change: the mistyped population is unchanged at 5 rows, all
	// > `deployed`, so nothing's treatment changes today — the hole was
	// > prospective, and so is the fix. See architecture_review/RFC_010.
	var pageID uuid.UUID
	pageCreated := true
	err := db.QueryRowContext(ctx, `
		INSERT INTO pages (site_id, name, url, title, page_type, build_status,
		                   sections, nav_label, in_header, in_footer)
		VALUES ($1, $2, $3, $4, $5, 'planned', $6::jsonb, $7, $8, $9)
		ON CONFLICT (site_id, name) DO NOTHING
		RETURNING id
	`, siteID, pageName, url, title, pageType,
		string(sectionsJSON), navLabel, inHeader, inFooter,
	).Scan(&pageID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		// The name is taken. Everything from here is one transaction — see
		// resolveNewPageConflict.
		pageCreated = false
		resolvedID, refusal, rerr := resolveNewPageConflict(ctx, db, siteID, domain,
			pageName, title, string(sectionsJSON), pageType, originalItemID, logger)
		if rerr != nil {
			return nil, rerr
		}
		pageID = resolvedID
		if refusal != nil {
			return refusal, nil
		}

	case err != nil:
		return nil, fmt.Errorf("create page: %w", err)
	}

	// Create work item for page-build-handler
	spec := map[string]interface{}{
		"page_name": pageName,
		"page_type": pageType,
		"title":     title,
		"purpose":   purpose,
		"sections":  sections,
		"source":    "content-gap-planner",
	}
	specJSON, _ := json.Marshal(spec)

	reasoning, _ := plan["reasoning"].(string)
	summary := fmt.Sprintf("Build new page: %s — %s", title, truncate(reasoning, 60))

	itemKey := fmt.Sprintf("gap_plan_new_%s_%s", pageName, siteID)

	// bugs_open/091, same shape as the filed case: the write is dedup-guarded, so
	// it writes nothing when an OPEN item already holds this key — and this used
	// to report `"item_created": true` regardless, i.e. "no error" rather than
	// "a record exists". insertWorkItem's own boolean is the thing that knows.
	itemCreated, err := withWorkItemTx(ctx, db, logger, gapPlanWorkItem(
		siteID, "needs_content_page", summary, string(specJSON), &pageID, 40, itemKey, originalItemID))
	if err != nil {
		return nil, fmt.Errorf("create work item: %w", err)
	}

	// Mark original as complete
	markOriginalComplete(ctx, db, originalItemID)

	logger.Info("ApplyGapPlanAction: new_page applied",
		zap.String("page_name", pageName),
		zap.String("page_id", pageID.String()),
		zap.Bool("page_created", pageCreated),
		zap.Bool("item_created", itemCreated))

	return map[string]interface{}{
		"applied":   true,
		"approach":  "new_page",
		"page_name": pageName,
		"page_id":   pageID.String(),
		// bugs_open/081, same shape as 091's `item_created`: the arm is named
		// for creation and reached the conflict branch silently, so "applied"
		// could not tell a new page from an overwritten one. Say which.
		"page_created": pageCreated,
		"item_created": itemCreated,
	}, nil
}

// resolveNewPageConflict runs the whole read-decide-write for a `new_page` name
// collision inside ONE transaction, with the conflicting row locked FOR UPDATE.
//
// Returns (pageID, nil, nil) when the row was refreshed, and (pageID, refusal,
// nil) when the collision was refused — bugs_open/081.
//
// THE TRANSACTION IS NOT TIDINESS. Two reasons, both raised by the council on
// correlation ccd4384c:
//
//   - the decision reads `pages.build_status` and then writes based on it, and
//     that column races the sweep that publishes a page. Unlocked, a page in the
//     middle of a deploy could be read as not-yet-deployed and then overwritten
//     as it goes live — the exact damage this fix exists to prevent, arriving
//     through the fix itself. `FOR UPDATE` closes that window.
//   - the shared work-item creator takes a *sql.Tx, which is the point: it
//     carries insertWorkItem's two-strike anti-churn and the ON CONFLICT clause
//     that matches idx_swi_dedup's partial predicate. A hand-rolled bare
//     `ON CONFLICT DO NOTHING` has neither and silently misbehaves once the
//     prior row goes terminal.
//
// Every statement's error is checked and propagated, deliberately. A discarded
// Exec error is invisible under sqlmock too, so a swallowed write here would
// also make the tests that guard this branch unable to see it (LANDMINES.md,
// "mock.ExpectationsWereMet() is NOT 'no database call happened'").
func resolveNewPageConflict(
	ctx context.Context, db *sql.DB, siteID uuid.UUID, domain, pageName,
	title, sectionsJSON, wantType string,
	originalItemID *uuid.UUID, logger *zap.Logger,
) (uuid.UUID, map[string]interface{}, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("begin conflict resolution for %q: %w", pageName, err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// "Has this page been served?" comes from the estate's ONE definition,
	// datahelpers.NeverDeployedPagePredicate, negated — never from a status test
	// spelled out here.
	//
	// > **CORRECTED 2026-08-02 (owner ruling, architecture_review/RFC_010).** This
	// > read `existingBuild == "deployed"`, and bugs_closed/037 is an entire filed
	// > case about what that misses: a `needs_rebuild` page HAS shipped and is still
	// > serving its old artefact — 35 of 46 such rows carry a non-null `deployed_at`.
	// > The guard could therefore re-type a live page it believed was unshipped.
	// > Measured before changing it: the mistyped population is still 5 rows, all
	// > `deployed`, so **no live row's treatment changes today** — this closes the
	// > hole prospectively rather than repairing damage. The predicate deliberately
	// > does NOT name `needs_rebuild` (its own test forbids that; singling out the
	// > status produced a 34-page false-positive class for the nav lane) —
	// > `deployed_at IS NULL` is what carries it.
	var pageID uuid.UUID
	var existingType, existingBuild string
	var hasShipped bool
	if qerr := tx.QueryRowContext(ctx, `
		SELECT id, COALESCE(page_type, ''), COALESCE(build_status, ''),
		       NOT (`+datahelpers.NeverDeployedPagePredicate+`)
		FROM pages WHERE site_id = $1 AND name = $2
		FOR UPDATE
	`, siteID, pageName).Scan(&pageID, &existingType, &existingBuild, &hasShipped); qerr != nil {
		return uuid.Nil, nil, fmt.Errorf("resolve conflicting page %q: %w", pageName, qerr)
	}

	if hasShipped && existingType != wantType {
		refusal, rerr := refuseDeployedPageTypeConflict(ctx, tx, siteID, domain,
			pageName, pageID, existingType, wantType, originalItemID, logger)
		if rerr != nil {
			return uuid.Nil, nil, rerr
		}
		if cerr := tx.Commit(); cerr != nil {
			return uuid.Nil, nil, fmt.Errorf("commit refusal for %q: %w", pageName, cerr)
		}
		committed = true
		return pageID, refusal, nil
	}

	// Either the page has never shipped, or it already holds the type the plan
	// asked for. Refreshing the plan's own fields is then coherent — the plan is
	// for this page's actual role — and is what this arm has always done.
	// page_type is deliberately NOT updated here; see applyNewPage.
	if _, uerr := tx.ExecContext(ctx, `
		UPDATE pages SET title = $3, sections = $4::jsonb, updated_at = NOW()
		WHERE site_id = $1 AND name = $2
	`, siteID, pageName, title, sectionsJSON); uerr != nil {
		return uuid.Nil, nil, fmt.Errorf("refresh existing page %q: %w", pageName, uerr)
	}
	if cerr := tx.Commit(); cerr != nil {
		return uuid.Nil, nil, fmt.Errorf("commit refresh for %q: %w", pageName, cerr)
	}
	committed = true
	return pageID, nil, nil
}

// refuseDeployedPageTypeConflict handles the bugs_open/081 case: a gap plan
// asked for a page under a name a DEPLOYED page already holds under a different
// page_type.
//
// It mutates the page not at all. Re-typing a live page changes what it serves
// immediately (render_news_section starts emitting data/news-archive.json for
// it, and the check that raised the gap stops firing), and 081 established that
// the choice of WHICH page holds a role cannot be made from the data: on
// robot-hands.com the real news listing and the gripper-catalog index are
// byte-identical at sections=["news-listing"], and a third page of that shape is
// archived. Guessing wrong breaks a working page, which is worse than the loop.
//
// The originating item is BLOCKED rather than completed — mirroring the growth
// budget branch in applyNewPage, and for the same reason: 'complete' on an item
// whose defect is untouched is the false-green this estate keeps relearning.
// Blocked stops the silent re-fire while leaving the item recoverable once a
// human has re-typed the page.
//
// `recurrenceExpected` is deliberately LEFT FALSE, and that is a decision rather
// than an omission. It exempts an item whose RE-REQUEST is normal; this is a
// DETECTED DEFECT, not an action request. While the decision is outstanding the
// row is non-terminal, so idx_swi_dedup blocks a duplicate and the two-strike
// rule is never reached. If a human DOES resolve it and the same collision
// returns, that genuinely is "we fixed this and it came back" — which is exactly
// what the two-strike label is for, so the default behaviour is the wanted one.
//
// handlerAgent is empty ON PURPOSE: nothing can resolve this but a person.
// Measured 2026-08-01 — claim_work_item_action.go:102 and
// load_work_item_actions.go:632 both claim `status IN ('triaged','approved')`,
// so a needs_human_review row cannot be picked up by the dispatch loop and
// re-run through content-gap-planner.
func refuseDeployedPageTypeConflict(
	ctx context.Context, tx *sql.Tx, siteID uuid.UUID, domain, pageName string,
	pageID uuid.UUID, existingType, wantType string,
	originalItemID *uuid.UUID, logger *zap.Logger,
) (map[string]interface{}, error) {
	reason := fmt.Sprintf(
		"page %q is deployed as page_type=%q; the plan wants %q. Re-typing a live page changes what it serves, so it is not done automatically (bugs_open/081)",
		pageName, existingType, wantType)

	// Keyed per (site, page, wanted type) so a check that re-fires while the
	// decision is outstanding dedups onto the open item rather than filing a
	// second one.
	//
	// The filing itself lives in page_role_upsert.go (bugs_open/175), shared with
	// the constant-role arms' refusal path, so both produce ONE item shape and
	// ONE item_key — a gap-plan refusal and a tool-deploy refusal on the same
	// page dedupe onto a single open decision instead of two.
	itemCreated, ierr := fileMistypedLivePageItem(ctx, tx, mistypedPageConflict{
		SiteID:       siteID,
		Domain:       domain,
		PageID:       pageID,
		PageName:     pageName,
		ExistingType: existingType,
		WantedType:   wantType,
		Source:       "content-gap-planner",
		Bug:          "bugs_open/081",
	}, logger)
	if ierr != nil {
		return nil, ierr
	}

	if originalItemID != nil {
		if _, uerr := tx.ExecContext(ctx, `
			UPDATE site_work_items
			SET status = 'blocked', error = $2, updated_at = NOW()
			WHERE id = $1
		`, *originalItemID, reason); uerr != nil {
			return nil, fmt.Errorf("block originating item for %q: %w", pageName, uerr)
		}
	}

	logger.Warn("ApplyGapPlanAction: new_page refused — deployed page holds this name under a different type",
		zap.String("domain", domain),
		zap.String("page_name", pageName),
		zap.String("existing_page_type", existingType),
		zap.String("wanted_page_type", wantType),
		zap.String("page_id", pageID.String()),
		zap.Bool("item_created", itemCreated))

	return map[string]interface{}{
		"applied":      false,
		"approach":     "new_page",
		"page_name":    pageName,
		"page_id":      pageID.String(),
		"page_created": false,
		"item_created": itemCreated,
		"reason":       "deployed_page_type_conflict",
		"detail":       reason,
	}, nil
}

// gapPlanWorkItem builds the work item every gap-plan arm raises, so all three
// arms describe the same thing the same way.
//
// WHY THIS EXISTS AT ALL (bugs_open/091, council correlation a5b70424). All three
// arms used to hand-roll `INSERT INTO site_work_items … ON CONFLICT DO NOTHING`
// against a *sql.DB, which two seats objected to: the `guidelines` seat because a
// bare conflict target does not match idx_swi_dedup's partial predicate, and
// `reuse_agent` because these should inherit insertWorkItem's semantics rather
// than reimplement half of them. Both were right, and the REASON the fork
// existed is the interesting part: these items set `parent_item_id`, and the
// shared `workItem` struct had no field for it — so the shared door was unusable
// to anyone who needed a parent, and three call sites walked round it. The field
// exists now; this routes through the door.
//
// recurrenceExpected is set DELIBERATELY, and it is what makes this adoption
// behaviour-preserving rather than a behaviour change riding in on a bug patch.
// A gap plan asking for a page to be built is an ACTION REQUEST: a completed
// predecessor means the request SUCCEEDED. Without the flag, adoption would newly
// suppress an item within 3h of a terminal predecessor and brand it `unresolved`
// after two — the exact regression bugs_open/024 recorded on the tool fix loop.
// Dedup itself is NOT waived: idx_swi_dedup still refuses a second OPEN row.
func gapPlanWorkItem(
	siteID uuid.UUID, itemType, summary, spec string, pageID *uuid.UUID,
	priority int, itemKey string, parentItemID *uuid.UUID,
) workItem {
	return workItem{
		siteID:             siteID,
		source:             "content-gap-planner",
		pipeline:           "build",
		itemType:           itemType,
		severity:           "medium",
		summary:            summary,
		spec:               spec,
		pageID:             pageID,
		priority:           priority,
		handlerAgent:       "page-build-handler",
		status:             "triaged",
		createdBy:          "content-gap-planner",
		itemKey:            itemKey,
		parentItemID:       parentItemID,
		recurrenceExpected: true,
	}
}

// ============================================================================
// retype_existing: flip a stranded page onto the right page_type
// ============================================================================

// applyRetypeExisting re-types a page the discovery check identified as
// "occupying the role under the wrong page_type" (bugs_open/015:
// relojistas.com's news listing was emitted as section-index, orphaning
// it from every gate that keys on news-index). Creating a page here would
// mint a duplicate; the fix is to flip the existing one.
//
// The plan only NAMES the page. Everything that AUTHORISES the re-type
// comes from the original work item's spec, written deterministically by
// the discovery check, never from the LLM:
//   - retype_candidates: the structural stranded-set (nav-visible,
//     sectionless, never deployed — findStrandedNavPages). A page outside
//     that set is refused, so a plan cannot re-type a live, working page.
//   - page_type: the routing key the check selects on (e.g. news-index).
//
// Refusals return applied=false and leave the original item untouched.
//
// Deliberately NO growth-budget check: nothing is created — the page
// already exists and already holds its nav slot.
func applyRetypeExisting(ctx context.Context, db *sql.DB, plan map[string]interface{}, siteID uuid.UUID, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error) {
	retypePlan, ok := plan["retype_existing"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("retype_existing plan missing or invalid")
	}

	pageName, _ := retypePlan["page_name"].(string)
	if pageName == "" {
		return nil, fmt.Errorf("retype_existing.page_name is required")
	}
	pageName = strings.ToLower(strings.TrimSpace(pageName))

	refuse := func(reason string) (interface{}, error) {
		logger.Warn("ApplyGapPlanAction: retype_existing refused",
			zap.String("page_name", pageName),
			zap.String("reason", reason))
		return map[string]interface{}{
			"applied":  false,
			"approach": "retype_existing",
			"reason":   reason,
		}, nil
	}

	if originalItemID == nil {
		return refuse("no original work item — retype is only authorised by a discovery-check item carrying retype_candidates")
	}

	var itemSpecJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT spec FROM site_work_items WHERE id = $1 AND site_id = $2
	`, *originalItemID, siteID).Scan(&itemSpecJSON)
	if err != nil {
		return refuse("original work item spec unreadable: " + err.Error())
	}
	var itemSpec map[string]interface{}
	if err := json.Unmarshal(itemSpecJSON, &itemSpec); err != nil {
		return refuse("original work item spec is not valid JSON: " + err.Error())
	}

	targetType, _ := itemSpec["page_type"].(string)
	if targetType == "" {
		return refuse("original item spec carries no page_type — nothing deterministic to re-type to")
	}

	candidatesRaw, _ := itemSpec["retype_candidates"].([]interface{})
	if len(candidatesRaw) == 0 {
		return refuse("original item spec carries no retype_candidates — this item did not authorise a re-type")
	}

	var candidateID uuid.UUID
	var previousType string
	var candidateNames []string
	found := false
	for _, c := range candidatesRaw {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := cm["name"].(string)
		candidateNames = append(candidateNames, name)
		if strings.ToLower(strings.TrimSpace(name)) != pageName {
			continue
		}
		idStr, _ := cm["id"].(string)
		id, perr := uuid.Parse(idStr)
		if perr != nil {
			return refuse(fmt.Sprintf("candidate %q has an unparseable id %q", name, idStr))
		}
		candidateID = id
		previousType, _ = cm["page_type"].(string)
		found = true
		break
	}
	if !found {
		return refuse(fmt.Sprintf("page %q is not in retype_candidates [%s] — refusing to re-type a page the check did not identify as stranded",
			pageName, strings.Join(candidateNames, ", ")))
	}

	// Sections: plan's list resolved to canonical component functions
	// (same treatment as new_page), else the archetype default for the
	// target type.
	sections := defaultSectionsForPage(pageName, targetType)
	if sectionsRaw, ok := retypePlan["sections"].([]interface{}); ok && len(sectionsRaw) > 0 {
		raw := make([]string, 0, len(sectionsRaw))
		for _, s := range sectionsRaw {
			if ss, ok := s.(string); ok {
				raw = append(raw, ss)
			}
		}
		resolver := loadComponentNameResolver(ctx, db, logger)
		if len(resolver.validFunctions) > 0 {
			resolved := make([]string, 0, len(raw))
			for _, name := range raw {
				if fn, ok := resolver.resolve(name); ok {
					resolved = append(resolved, fn)
				} else {
					logger.Warn("applyRetypeExisting: dropped unresolvable section name",
						zap.String("page", pageName), zap.String("section", name))
				}
			}
			if len(resolved) > 0 {
				sections = resolved
			}
		} else if len(raw) > 0 {
			sections = raw // resolver unavailable — use as-is rather than lose them
		}
	}
	sectionsJSON, _ := json.Marshal(sections)

	res, err := db.ExecContext(ctx, `
		UPDATE pages
		SET page_type = $2, sections = $3::jsonb, updated_at = NOW()
		WHERE id = $1 AND site_id = $4
	`, candidateID, targetType, string(sectionsJSON), siteID)
	if err != nil {
		return nil, fmt.Errorf("retype page %q: %w", pageName, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return refuse(fmt.Sprintf("page %q (id %s) no longer exists — candidate list was stale", pageName, candidateID))
	}

	// Hand the now-correctly-typed page to the build pipeline, mirroring
	// new_page's work item so it actually gets built.
	reasoning, _ := plan["reasoning"].(string)
	spec := map[string]interface{}{
		"page_name":    pageName,
		"page_type":    targetType,
		"sections":     sections,
		"source":       "content-gap-planner",
		"retyped_from": previousType,
	}
	specJSON, _ := json.Marshal(spec)
	summary := fmt.Sprintf("Build re-typed page: %s (now %s) — %s", pageName, targetType, truncate(reasoning, 60))
	itemKey := fmt.Sprintf("gap_plan_retype_%s_%s", pageName, siteID)

	// bugs_open/091, third site of the same shape — see gapPlanWorkItem.
	itemCreated, err := withWorkItemTx(ctx, db, logger, gapPlanWorkItem(
		siteID, "needs_content_page", summary, string(specJSON), &candidateID, 40, itemKey, originalItemID))
	if err != nil {
		return nil, fmt.Errorf("create build work item for re-typed page: %w", err)
	}

	markOriginalComplete(ctx, db, originalItemID)

	logger.Info("ApplyGapPlanAction: retype_existing applied",
		zap.String("page_name", pageName),
		zap.String("page_id", candidateID.String()),
		zap.String("from_type", previousType),
		zap.String("to_type", targetType))

	return map[string]interface{}{
		"applied":      true,
		"approach":     "retype_existing",
		"page_name":    pageName,
		"page_id":      candidateID.String(),
		"from_type":    previousType,
		"to_type":      targetType,
		"item_created": itemCreated,
	}, nil
}

// defaultSectionsForPage returns archetype-appropriate default sections
// for a new page when the planner LLM provides none. A recognised page
// type gets its archetype shape (e.g. faq -> structured faq, not a
// generic-text-block that would compete with it). Unrecognised pages keep
// the generic content shape.
func defaultSectionsForPage(pageName, pageType string) []string {
	// page_type outranks the name heuristics: a typed page has an
	// archetype regardless of what it is called — the name is localised
	// ("noticias"), the type is not (bugs_open/015).
	switch strings.ToLower(strings.TrimSpace(pageType)) {
	case "news-index":
		return []string{"hero", "news-listing", "call-to-action"}
	case "entity-directory":
		// bugs_open/206: directory-listing resolves query.business_directory
		// at plan time (queryresolve package) — a real, per-site business
		// list, not fabricated content. Requires the site to have its own
		// directory-export-json config; an unconfigured site's section
		// resolves to an empty (not fabricated) list.
		return []string{"hero", "directory-listing"}
	}
	key := strings.ToLower(strings.TrimSpace(pageName))
	switch {
	case key == "faq" || strings.Contains(key, "faq"):
		return []string{"hero", "faq", "call-to-action"}
	case key == "contact":
		return []string{"contact-hero", "contact-form", "contact-info"}
	case key == "pricing" || strings.Contains(key, "pricing"):
		return []string{"hero", "pricing", "faq", "call-to-action"}
	case key == "about":
		return []string{"hero-about", "about-content", "call-to-action"}
	case key == "guides-index" || key == "guide-index":
		// Matches the ALREADY-PROVEN fleet pattern exactly (verified live,
		// bugs_open/206): mortgagecalculator.co.uk, idea.uk, gamesdesign.co.uk
		// and relojistas.com all deploy this layout for this page shape,
		// backed by guide-list's own query.pages_where_type:guide.
		return []string{"hero", "guide-list"}
	case key == "tools-index" || key == "tool-index":
		// Same proven pattern, tool-list's sibling case (gamesdesign.co.uk,
		// robot-hands.com, finetuning.uk, ai-agent-orchestration.com).
		return []string{"hero", "tool-list"}
	default:
		return []string{"hero", "generic-text-block", "call-to-action"}
	}
}

// insertSitePlanSectionRows writes one site_plan_sections row per section,
// in order, for one page under one plan, with no per-section fact
// assignment (assigned_fact_ids stays NULL — "unscoped", the documented
// default shape, PBP-037). Used by ensure_page_section_layout_action.go.
//
// bugs_open/206 council round 1 (reuse_agent, constitution, architecture all
// independently flagged the same "two writers, same target" shape): this was
// meant to also replace write_site_plan_action.go's own equivalent loop, but
// that file has SINCE gained per-section assigned_fact_ids handling
// (PBP-037, a separate, actively-revised lane — council REJECTED on breadth,
// re-sliced per RFC_016) mid-way through this review round. Forcing that
// file onto this 4-column helper right now would either drop its new
// fact-assignment column or require guessing at RFC_016's still-moving
// shape — neither is this lane's call to make unilaterally. Left as a
// genuine follow-up for whichever lane lands second: extend this helper
// with an optional assignedFacts []interface{} parameter (nil per index =
// today's NULL/unscoped behaviour) and have write_site_plan_action.go call
// it too, once PBP-037 settles.
func insertSitePlanSectionRows(ctx context.Context, tx *sql.Tx, planID uuid.UUID, pageName string, sections []string) error {
	for ord, componentName := range sections {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
			VALUES ($1, $2, $3, $4)
		`, planID, pageName, ord, componentName); err != nil {
			return fmt.Errorf("insert site_plan_sections for %q[%d]: %w", pageName, ord, err)
		}
	}
	return nil
}

// ============================================================================
// update_spec: write a value to site_specs
// ============================================================================

func applyUpdateSpec(ctx context.Context, db *sql.DB, plan map[string]interface{}, siteID uuid.UUID, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error) {
	specPlan, ok := plan["update_spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("update_spec plan missing or invalid")
	}

	aspect, _ := specPlan["aspect"].(string)
	if aspect == "" {
		aspect = "identity"
	}

	field, _ := specPlan["field"].(string)
	suggestedValue := specPlan["suggested_value"]

	if field == "" || suggestedValue == nil {
		return nil, fmt.Errorf("update_spec needs field and suggested_value")
	}

	// Read current spec
	var currentDataJSON []byte
	var oldID *uuid.UUID
	err := db.QueryRowContext(ctx, `
		SELECT id, data FROM site_specs
		WHERE site_id = $1 AND aspect = $2 AND is_current = true
	`, siteID, aspect).Scan(&oldID, &currentDataJSON)

	var currentData map[string]interface{}
	if err == sql.ErrNoRows {
		currentData = make(map[string]interface{})
	} else if err != nil {
		return nil, fmt.Errorf("read current spec: %w", err)
	} else {
		json.Unmarshal(currentDataJSON, &currentData)
	}

	// Merge
	currentData[field] = suggestedValue
	mergedJSON, _ := json.Marshal(currentData)

	// Supersede old
	if oldID != nil {
		db.ExecContext(ctx, `
			UPDATE site_specs SET is_current = false, superseded_at = now() WHERE id = $1
		`, *oldID)
	}

	// Insert new
	_, err = db.ExecContext(ctx, `
		INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by)
		VALUES ($1, $2, $3::jsonb, 'content-gap-planner', 'content-gap-planner', true, 'content-gap-planner')
	`, siteID, aspect, string(mergedJSON))

	if err != nil {
		return nil, fmt.Errorf("write spec: %w", err)
	}

	// Mark original as complete
	markOriginalComplete(ctx, db, originalItemID)

	logger.Info("ApplyGapPlanAction: update_spec applied",
		zap.String("aspect", aspect),
		zap.String("field", field),
		zap.String("value_preview", truncate(fmt.Sprintf("%v", suggestedValue), 60)))

	return map[string]interface{}{
		"applied":  true,
		"approach": "update_spec",
		"aspect":   aspect,
		"field":    field,
	}, nil
}

// ============================================================================
// not_actionable: mark original item as wont_fix
// ============================================================================

func applyNotActionable(ctx context.Context, db *sql.DB, plan map[string]interface{}, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error) {
	reason := "Not actionable"
	if naPlan, ok := plan["not_actionable"].(map[string]interface{}); ok {
		if r, ok := naPlan["reason"].(string); ok {
			reason = r
		}
	}

	if originalItemID != nil {
		db.ExecContext(ctx, `
			UPDATE site_work_items
			SET status = 'wont_fix', error = $2, completed_at = NOW()
			WHERE id = $1
		`, *originalItemID, reason)
	}

	logger.Info("ApplyGapPlanAction: not_actionable",
		zap.String("reason", reason))

	return map[string]interface{}{
		"applied":  true,
		"approach": "not_actionable",
		"reason":   reason,
	}, nil
}

// ============================================================================
// Helpers
// ============================================================================

func markOriginalComplete(ctx context.Context, db *sql.DB, itemID *uuid.UUID) {
	if itemID != nil {
		db.ExecContext(ctx, `
			UPDATE site_work_items
			SET status = 'complete', completed_at = NOW(), handled_by = 'content-gap-planner'
			WHERE id = $1 AND status IN ('triaged', 'claimed')
		`, *itemID)
	}
}

// splitPageNames handles LLM output that sometimes produces comma-separated,
// "and"-separated, or slash-separated page lists like "index, services",
// "index and services", or "index/services".
// Returns cleaned, trimmed individual page names.
func splitPageNames(raw string) []string {
	// Normalise separators
	normalized := strings.ReplaceAll(raw, " and ", ",")
	normalized = strings.ReplaceAll(normalized, " & ", ",")
	normalized = strings.ReplaceAll(normalized, "/", ",")

	parts := strings.Split(normalized, ",")
	var result []string
	seen := map[string]bool{}
	for _, p := range parts {
		name := strings.TrimSpace(p)
		name = strings.ToLower(name)
		if name != "" && !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	return result
}
