// FILE: platform/orchestration/actions/rebuild_blog_listing_action.go
//
// RebuildBlogListingAction queries deployed blog-post pages for a site and
// renders a blog-listing page_component using the template from content_components.
//
// Data layer only — presentation comes from the component library template.
// Uses the existing content-listing content_component's html_template.
// Falls back to a minimal template if no component is found.
//
// No LLM needed — purely algorithmic. Runs as a step in the rerender-pages
// workflow (before get_pages), or triggered after blog post publishing.
//
// Registration:
//   "rebuild_blog_listing": {
//       Handler:     RebuildBlogListingAction,
//       Category:    "site",
//       Description: "Rebuild blog listing page_component from published posts",
//       IsLocal:     true,
//   },
//
// Data inputs (via ActionInputSpec):
//   - site_id (required)
//
// Changes from previous version:
//   - findBlogListingSlot: discovers actual slot name from page_components
//     instead of hardcoding "blog-listing". Handles sites where the blog page
//     was planned with article-grid, featured-article, content-listing, etc.
//   - Loads content-listing template (has proper input_schema with {{range}})
//     instead of blog-listing (which was CSS-only, empty input_schema).
//   - ensureArticleLinks: patches article titles missing <a href>.
//   - Writes content_data alongside rendered_html (source-of-truth principle).
//   - estimateReadTime: computed from blog post rendered content length.
//   - Removed loadBlogListingTemplate (was loading CSS-only template).
//
// v2 changes (April 2026):
//   - findBlogPage: falls back to name='blog' when page_type != 'blog-index'.
//     Fixes sites where the planner created the blog page as type 'content'
//     because blog-index was in the unavailableBuilders map.
//     When found by name fallback, corrects page_type to 'blog-index' so
//     future runs find it directly.
//   - Blog post query widened from build_status='deployed' to
//     build_status IN ('deployed','needs_rebuild'). Posts with needs_rebuild
//     are already in git with valid content — they just have pending rerender
//     items. Excluding them left the listing empty.
//
// v3 changes (2026-07-26, bugs_open/052):
//   - The v2 widening above was right about the symptom and wrong about the
//     test. `needs_rebuild` does not mean "still serves": a page carries that
//     flag whether it was deployed once and later flagged (it keeps serving
//     its old artefact) or was never deployed at all (it 404s). build_status
//     cannot tell those apart; `deployed_at` can. The query now carries the
//     shared floor instead of a hand-written build_status list — see
//     blogPostsQuery. Deliberately no count here: that population is live
//     state and moves (10 on 2026-07-26, 4 six days earlier), so a number
//     baked into a comment is wrong within the week — measure it if you need
//     it, don't quote this.
//   - Both findBlogPage strategies gained a `status` filter. They had none, so
//     an archived blog-index page could be selected as the listing target.

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// blogPostsQuery selects the posts this listing advertises.
//
// The floor is queryresolve.ListedPageEligibilitySQL — the SAME constant
// queryresolve's own `blog_posts` source uses — because this action and that
// resolver both derive the article set for the same blog page and must not
// disagree about it. Two hand-maintained copies of one predicate is the
// drift class bugs_closed/023 documents; sharing the constant is what stops
// it recurring here.
//
// Why the floor at all (bugs_open/052): a listing regenerated from the page
// set must never advertise a page that would 404. The previous predicate,
// `build_status IN ('deployed','needs_rebuild')`, admitted two populations it
// should not have — pages that are `needs_rebuild` but were never deployed
// (they 404), and, because there was no `status` filter at all, pages that had
// been deliberately `archived`. That second gap defeated archiving, which is
// the containment route the fleet relies on for a dead page.
//
// THE IMAGE IS THE SHARED PROJECTION, NOT A BLANK (bugs_open/384 decision 3,
// 2026-08-25). This query used to omit the image entirely and the scan loop
// wrote `"image": ""` for every article. That made this action a SECOND WRITER
// of `content_data.articles` on a field the 384 seam exists to keep correct:
// leopardessconsulting.co.uk's blog page carries `blog-listing_pre_037`, which
// declares `articles` ← `query.blog_posts` AND renders `.image`, so it is a 384
// consumer. A card landing there makes the seam re-resolve the array with the
// real image; the next `rerender-pages` run (42 in the 14 days to 2026-08-25,
// and this is an unconditional step in it) blanked it again. Last writer won.
//
// Measured before changing it, 2026-08-25: 3 live blog-index listings, 47
// listed articles, 47 blank images — and 0 of the 47 has a card asset or a plan
// hero. So the shared projection returns "" for every one of them TODAY and
// this change alters no stored byte. It is a door-closing fix, not a repair;
// do not quote it as having fixed a visible listing.
//
// CAPPED at queryresolve.PageListingHardCap, added 2026-08-25 on the council's
// objection (round 170147b4, guardian). The first cut spliced the projection
// into an UNCAPPED query, so PageImageJoinsSQL's per-row LATERAL hero lookup —
// designed around a 24-row result — ran once per post with no bound.
//
// And the cap is not only about cost: WITHOUT it the two writers of this same
// listing DISAGREE. `query.blog_posts` resolves through resolvePagesWhereType,
// which caps at 24; this action did not. [MEASURED 2026-08-25]
// webdesign.co.uk has **40** eligible blog posts, so the resolver would produce
// 24 items and this action 40 — a 16-item divergence on one listing, which is
// exactly the drift class this file already shares ListedPageEligibilitySQL to
// avoid. Sharing the cap closes it the same way: ONE definition, both callers.
//
// A measured no-op on what this action writes today: the three live blog-index
// listings carry 16, 20 and 11 posts, all under the cap.
//
// Alias contract: `p` for pages, as the shared constants require. `ca` and `ha`
// come from PageImageJoinsSQL and do not collide with the `pc` subquery below.
// A `var`, not a `const`: the LIMIT is composed from queryresolve.PageListingHardCap
// at init, and a const cannot call strconv.Itoa. Sharing the cap is worth the
// keyword — a second literal 24 here is exactly the drift this splice removes.
var blogPostsQuery = `
		SELECT p.id, p.name, p.url, p.title,
		       COALESCE(p.meta_description, ''),
		       p.created_at,
		       COALESCE(
		           (SELECT SUM(LENGTH(COALESCE(pc.rendered_html, '')))
		            FROM page_components pc
		            WHERE pc.page_id = p.id
		              AND COALESCE(pc.slot_name, '') NOT IN ('header', 'footer', 'head')),
		           0
		       ) as content_length,
		       ` + queryresolve.PageImageProjectionSQL + `
		FROM pages p
		` + queryresolve.PageImageJoinsSQL + `
		WHERE p.site_id = $1
		  AND p.page_type = 'blog-post'
		  AND p.status IN ('active', 'deployed')` + queryresolve.ListedPageEligibilitySQL + `
		ORDER BY p.created_at DESC
		LIMIT ` + strconv.Itoa(queryresolve.PageListingHardCap) + `
	`

var RebuildBlogListingInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("rebuild_blog_listing", RebuildBlogListingInputSpec)
}

// slotPriority is the ordered list of slot names we look for when finding
// the blog listing page_component. The first match wins.
// These cover the various names planners have used for blog listing sections.
var slotPriority = []string{
	"blog-listing",
	"article-grid",
	"content-listing",
	"guide-list",
	"featured-article",
}

func RebuildBlogListingAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "rebuild_blog_listing"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		RebuildBlogListingInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// ── Find blog page ──────────────────────────────────────────────────
	blogPageID, blogPageName, err := findBlogPage(ctx, params.DB, siteID, logger)
	if err != nil {
		return nil, err
	}
	if blogPageID == uuid.Nil {
		logger.Info("RebuildBlogListingAction: No blog page found, skipping")
		return map[string]interface{}{
			"rebuilt": false,
			"reason":  "no blog page",
		}, nil
	}

	// ── Find the correct slot, and decide whether we may write there ────
	slot := findBlogListingSlot(ctx, params.DB, blogPageID, logger)
	slotName := slot.Name
	existingComponentID := slot.Existing
	op, refusal := decideBlogListingWrite(slot)
	logger.Info("RebuildBlogListingAction: Resolved slot",
		zap.String("slot_name", slotName),
		zap.String("origin", slot.Origin.String()),
		zap.Int("occupants", slot.Occupants),
		zap.Bool("has_existing_component", existingComponentID != uuid.Nil),
	)

	// A refusal is NOT an error, and this is load-bearing (bugs_open/457).
	// rebuild_blog_listing is an unconditional step of the rerender-pages
	// workflow, sitting between render_site_components and get_pages, and the
	// workflow has no error_step: returning an error here aborts the run BEFORE
	// create_rerender_items, so a chrome refresh re-renders the three
	// site_components and then creates none of the page rerenders it exists to
	// create. That is exactly the 18-page outage this bug was filed for. So a
	// refusal is loud in the log, reported in the result, and the chain
	// continues.
	if op != opUpdate && op != opInsert {
		logger.Error("RebuildBlogListingAction: refusing to write the blog listing",
			zap.String("blog_page", blogPageName),
			zap.String("slot_name", slotName),
			zap.String("origin", slot.Origin.String()),
			zap.Int("occupants", slot.Occupants),
			zap.String("reason", refusal),
		)
		return map[string]interface{}{
			"rebuilt":     false,
			"skipped":     true,
			"reason":      refusal,
			"slot_name":   slotName,
			"slot_origin": slot.Origin.String(),
			"occupants":   slot.Occupants,
		}, nil
	}

	// ── Load blog posts ─────────────────────────────────────────────────
	rows, err := params.DB.QueryContext(ctx, blogPostsQuery, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query blog posts: %w", err)
	}
	defer rows.Close()

	articles, err := scanBlogArticles(rows, logger)
	if err != nil {
		// Loud by design (council 170147b4, bug_historian): a projection/Scan
		// divergence must not be laundered into "no posts", which the branch
		// below would treat as "leave the existing listing alone" and report as
		// success.
		return nil, fmt.Errorf("failed to read blog posts: %w", err)
	}

	// How many articles the listing carried BEFORE this rebuild. Read for the
	// shrink check below; -1 means "no previous set to compare against".
	previousCount := previousArticleCount(ctx, params.DB, existingComponentID, logger)

	if len(articles) == 0 {
		// An empty set leaves any EXISTING listing untouched, so the page keeps
		// advertising whatever it advertised before. That is the safe choice —
		// blanking a live listing on a transient empty read would be worse —
		// but it is not a no-op, so say so loudly when there is a component
		// standing: the eligibility floor above can newly empty a set that used
		// to fill, and the stale listing is then the thing serving the 404s
		// this fix exists to stop (bugs_open/052).
		if existingComponentID != uuid.Nil {
			logger.Warn("RebuildBlogListingAction: no eligible blog posts, but a listing component exists — it keeps its previous contents and may now be stale",
				zap.String("blog_page", blogPageName),
				zap.String("slot_name", slotName),
				zap.String("component_id", existingComponentID.String()),
				zap.Int("previous_article_count", previousCount),
			)
		} else {
			logger.Info("RebuildBlogListingAction: No blog posts found")
		}
		return map[string]interface{}{
			"rebuilt": false,
			"reason":  "no blog posts",
		}, nil
	}

	// A listing that shrinks is the failure mode this action cannot see from
	// its own success: the rebuild reports `rebuilt: true` either way, and
	// posts simply stop appearing. Guarding only the all-zero case above would
	// miss every partial erosion (16 posts -> 10), which is the same
	// no-error-no-warning signature as the bug this fix closes. The eligibility
	// floor is the plausible cause — ListedPageEligibilitySQL additionally
	// requires non-empty `sections`, so a deployed post that lands in that
	// state drops out silently — but the check is deliberately cause-agnostic:
	// a post deleted, archived or retyped upstream shows up here too, and all
	// of those are worth a line in the log.
	//
	// Warn only, never a refusal: the new set is the correct one by
	// construction, and refusing to write it would leave the KNOWN-stale
	// listing serving instead. This makes the shrink visible, it does not
	// second-guess it.
	if previousCount > len(articles) {
		logger.Warn("RebuildBlogListingAction: blog listing SHRANK — fewer posts are eligible than the listing previously carried",
			zap.String("blog_page", blogPageName),
			zap.String("slot_name", slotName),
			zap.Int("previous_article_count", previousCount),
			zap.Int("new_article_count", len(articles)),
			zap.Int("dropped", previousCount-len(articles)),
		)
	}

	// ── Load component template ─────────────────────────────────────────
	// Use content-listing function (has proper input_schema with {{range}}).
	// The old blog-listing component was CSS-only with empty input_schema.
	htmlTemplate, templateComponentID := loadContentListingTemplate(ctx, params.DB, logger)

	// ── Render template with post data ──────────────────────────────────
	templateData := map[string]interface{}{
		"section_title":    "Latest Articles",
		"section_subtitle": "",
		"articles":         articles,
		"show_load_more":   false,
		"load_more_text":   "Load More",
	}

	rendered, renderErr := renderBlogTemplate(htmlTemplate, templateData, logger)
	if renderErr != nil {
		return nil, fmt.Errorf("blog listing render failed, listing left unchanged: %w", renderErr)
	}
	if rendered == "" {
		return nil, fmt.Errorf("template rendering produced empty output")
	}

	// Patch: ensure article titles are clickable links
	rendered = ensureArticleLinks(rendered, articles)

	// ── Build content_data (source of truth) ────────────────────────────
	contentData := map[string]interface{}{
		"section_title":    "Latest Articles",
		"section_subtitle": "",
		"articles":         articles,
		"show_load_more":   false,
		"load_more_text":   "Load More",
	}
	contentDataJSON, err := json.Marshal(contentData)
	if err != nil {
		logger.Warn("Failed to marshal content_data", zap.Error(err))
		contentDataJSON = []byte("{}")
	}

	// ── Upsert page_component ───────────────────────────────────────────
	var componentID uuid.UUID

	if op == opUpdate {
		// Update existing component in the correct slot — unless a human has
		// locked it (bugs_open/058): the lock predicate on the WHERE makes the
		// refusal race-free, and the blocked refresh is surfaced as a work
		// item rather than silently skipped.
		// Classify before the overwrite (bugs_open/229): if the stored bytes
		// no longer match their stamp, a non-render writer changed them and
		// this refresh is about to destroy that content. Advisory; the 357
		// trigger archives the outgoing bytes regardless.
		divergent, classifyErr := classifyPageComponentArtefacts(ctx, params.DB, blogPageID)
		if classifyErr != nil {
			logger.Warn("RebuildBlogListingAction: divergence classification failed — refresh proceeds, the 357 trigger still archives (bugs_open/229)",
				zap.Error(classifyErr))
		}

		// rendered_html_digest stamped in the SAME statement as the bytes
		// (bugs_open/229): this is the render path for the listing component.
		// Writer stamp (bugs_open/355 A1): the archive row this UPDATE fires
		// names the action, not the connection's socket.
		res, updErr := stampedExecContext(ctx, params.DB, contentWriterRebuildBlogListing, `
			UPDATE page_components
			SET rendered_html = $1, rendered_html_digest = md5($1), content_data = $2::jsonb, updated_at = NOW()
			WHERE id = $3 AND `+pageComponentAgentWritableSQL("")+`
		`, rendered, string(contentDataJSON), existingComponentID)
		if updErr != nil {
			return nil, fmt.Errorf("failed to update blog listing component: %w", updErr)
		}
		if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
			logger.Warn("RebuildBlogListingAction: blog listing component is human-locked — refresh refused (bugs_open/058)",
				zap.String("component_id", existingComponentID.String()),
				zap.String("slot_name", slotName),
			)
			lock, lockErr := CheckComponentLock(ctx, params.DB, existingComponentID, logger)
			lockedBy, lockType := "", ""
			if lockErr == nil && lock.IsLocked {
				lockedBy, lockType = lock.LockedBy, lock.LockType
			}
			pcID := existingComponentID
			emitLockBlockedChangeItem(ctx, params.DB, siteID, &blogPageID, &pcID,
				blogPageName, slotName, lockedBy, lockType,
				"overwrite", "rebuild_blog_listing", logger)
			return map[string]interface{}{
				"rebuilt": false,
				"skipped": true,
				"locked":  true,
				"reason":  fmt.Sprintf("blog listing component %s is locked by %q — auto-refresh refused", existingComponentID, lockedBy),
			}, nil
		}
		componentID = existingComponentID
		logger.Info("RebuildBlogListingAction: Updated existing component",
			zap.String("component_id", componentID.String()),
			zap.String("slot_name", slotName),
		)

		// Emit only for THIS component (classify covered the whole page) and
		// only after the UPDATE reported a row written — the same
		// after-RowsAffected rule as the chrome emitter. No ledger read-back
		// here: a single-row UPDATE whose classify failed loses at most one
		// mislabelled log line, and the archive row still exists.
		if classifyErr == nil {
			var mine []pageComponentDivergence
			for _, d := range divergent {
				if d.ComponentID == existingComponentID {
					mine = append(mine, d)
				}
			}
			emitPageDivergenceItems(ctx, params.DB, blogPageID, blogPageName, mine, "rebuild_blog_listing", logger)
		}
	} else {
		// opInsert: the slot is empty AND the page's plan names it as a listing
		// slot. Both halves are required — see decideBlogListingWrite.
		//
		// position comes from the plan's own section index, never a literal.
		// The old hard-coded 3 collided with whatever legitimately sat there
		// (on the motivating page, a call-to-action). Note that position is
		// deliberately NOT part of migration 316's key, so this is a layout
		// correctness fix and not the duplicate fix — the occupancy check above
		// is the duplicate fix.
		position := slot.PlanPos
		if position < 1 {
			position = 1
		}

		// component_id is bound from the component that actually rendered these
		// bytes, so the row can be attributed, re-rendered by component and
		// swept by anything reasoning about components. NULL when the built-in
		// default template was used, because then there is no component.
		var componentIDPtr *uuid.UUID
		if templateComponentID != uuid.Nil {
			c := templateComponentID
			componentIDPtr = &c
		}

		err = params.DB.QueryRowContext(ctx, `
			INSERT INTO page_components (page_id, slot_name, component_id, position, rendered_html, rendered_html_digest, content_data, build_status)
			VALUES ($1, $2, $3, $4, $5, md5($5), $6::jsonb, 'deployed')
			RETURNING id
		`, blogPageID, slotName, componentIDPtr, position, rendered, string(contentDataJSON)).Scan(&componentID)
		if err != nil {
			// Do NOT abort the workflow (see the refusal note above): a losing
			// race against a concurrent run of this same step must not take
			// create_rerender_items down with it. Migration 316 makes that race
			// loud rather than silent, which is the behaviour we want — it just
			// must not be fatal here.
			logger.Error("RebuildBlogListingAction: insert of the blog listing component failed — listing left unchanged",
				zap.String("blog_page", blogPageName),
				zap.String("slot_name", slotName),
				zap.Int("position", position),
				zap.Error(err),
			)
			return map[string]interface{}{
				"rebuilt":   false,
				"skipped":   true,
				"reason":    fmt.Sprintf("insert into slot %q failed: %v", slotName, err),
				"slot_name": slotName,
			}, nil
		}
		logger.Info("RebuildBlogListingAction: Created new component",
			zap.String("page_component_id", componentID.String()),
			zap.String("slot_name", slotName),
			zap.Int("position", position),
			zap.Bool("component_id_bound", componentIDPtr != nil),
		)
	}

	logger.Info("RebuildBlogListingAction: Complete",
		zap.String("blog_page", blogPageName),
		zap.Int("post_count", len(articles)),
		zap.String("slot_name", slotName),
	)

	return map[string]interface{}{
		"rebuilt":      true,
		"blog_page_id": blogPageID.String(),
		"post_count":   len(articles),
		"component_id": componentID.String(),
		"slot_name":    slotName,
	}, nil
}

// scanBlogArticles projects blogPostsQuery's rows into the listing item shape.
//
// EXTRACTED from the action's inline loop 2026-08-25 (bugs_open/384 decision 3)
// so the SCAN CONTRACT is testable. The image arrives as three columns spliced
// in from queryresolve.PageImageProjectionSQL, and a projection that gains or
// reorders a column breaks this Scan at RUNTIME, not at compile time — the same
// failure class the alias-contract test already guards for the eligibility
// fragment. Driving this function with mock rows is the only DB-free way to
// prove the column count and order still line up.
//
// ONE bad row is skipped; EVERY row failing is an ERROR, and the difference is
// the whole point (council round 170147b4, bug_historian, gating, 2026-08-25).
//
// The original loop logged-and-skipped unconditionally. Combined with the
// caller — which treats an empty article set as "leave the existing listing
// alone" — that made a PROJECTION-SHAPE MISMATCH silent: if
// PageImageProjectionSQL ever gains or reorders a column, every Scan fails,
// every row is skipped, the listing comes back empty, the caller keeps the
// stale listing, the step reports success and NOBODY IS TOLD. The reviewer
// named it as this council's recurring shape and was right: the exposure was
// documented in prose here and not closed.
//
// So the two cases are now distinguished, because they have different causes:
//   - SOME rows scanned: one malformed post. Skip it, log at Warn, carry on —
//     a single bad row must not blank a live listing.
//   - NO rows scanned but rows were offered: the Scan destinations no longer
//     match the SELECT list. That is a code defect, not data, and it cannot be
//     repaired by retrying. Return an error so the step FAILS loudly instead of
//     handing the caller an empty set that looks like "no posts".
//
// A genuinely empty result set (no eligible posts) is NOT an error and never
// reaches this branch — attempted stays 0, and the caller's own no-posts path
// handles it with its existing warning.
func scanBlogArticles(rows *sql.Rows, logger *zap.Logger) ([]map[string]interface{}, error) {
	var articles []map[string]interface{}
	// CONVERGE ON TOUCH (bugs_open/410, council round c8385154, reuse_agent):
	// these hand-rolled offered/kept counters are the graded sibling of
	// datahelpers.ScanShortfall, which shipped after this function and extracted
	// the same concept. Next time this function is edited for any other reason,
	// adopt those counters (a graded variant of the helper, added with this as
	// its first caller) rather than leaving a third parallel implementation.
	// Deliberately not converted in the 410 commit itself: this function is
	// already guarded, and a behaviour-neutral refactor of the listing rebuild
	// would have widened a bug fix's blast radius for nothing.
	attempted, scanFailures := 0, 0
	for rows.Next() {
		attempted++
		var id uuid.UUID
		var name, url, title, metaDesc string
		var createdAt time.Time
		var contentLength int
		var img queryresolve.PageImageCols
		if err := rows.Scan(&id, &name, &url, &title, &metaDesc, &createdAt, &contentLength,
			&img.CardKey, &img.HeroKey, &img.HeroPurpose); err != nil {
			scanFailures++
			logger.Warn("Failed to scan blog post", zap.Error(err))
			continue
		}

		// SHARED with the resolver, never re-spelled here (bugs_open/425).
		// These two rules lived only in this loop; `query.blog_posts` feeds the
		// SAME content-listing component from resolvePagesWhereType and applied
		// neither, so one site served two spellings of one card — a clean
		// headline plus a deck on the blog index, the raw document <title> and
		// an empty <p> on the home page. The byte-slice truncation that stood
		// here is gone with it: it could cut a multi-byte character in half
		// (bugs_open/423's failure), and every meta description on this estate
		// is a candidate.
		cleanTitle := queryresolve.ListItemTitle(title)
		excerpt := queryresolve.ListItemExcerpt(metaDesc)

		articles = append(articles, map[string]interface{}{
			"title": cleanTitle,
			"url":   url,
			// BOTH deck keys, because the two producers of this shape had two
			// vocabularies and the component library reads both: content-listing
			// renders {{.excerpt}}, blog-listing_pre_037 renders
			// {{.meta_description}}. Emitting only one starves whichever template
			// reads the other, and which template that is depends on the
			// PRODUCER that happened to run — the same defect as bugs_open/425,
			// pointing the other way.
			//
			// [MEASURED 2026-09-02] latent, not live: this action loads the
			// content-listing template by lookup rather than following the
			// page_component's own component_id, so blog-listing_pre_037's
			// {{.meta_description}} is not reached on this path today. That is a
			// door left open by an unrelated mechanism, not a guarantee — closing
			// it here costs one key.
			"excerpt":          excerpt,
			"meta_description": metaDesc,
			"date":             createdAt.Format("Jan 2, 2006"),
			"category":         "",
			// The SHARED projection, never a blank (bugs_open/384 decision 3):
			// card crop first, plan hero second, "" when the page has neither.
			"image":     img.WebPath(),
			"read_time": estimateReadTime(contentLength),
		})
	}

	if attempted > 0 && len(articles) == 0 {
		logger.Error("RebuildBlogListingAction: EVERY blog post row failed to scan — the projection's column list no longer matches the Scan destinations; refusing to report an empty listing as 'no posts'",
			zap.Int("rows_offered", attempted),
			zap.Int("scan_failures", scanFailures),
			zap.String("hint", "blogPostsQuery splices queryresolve.PageImageProjectionSQL; a column added or reordered there must change scanBlogArticles in the same commit"))
		return nil, fmt.Errorf("blog listing scan failed for all %d offered rows (%d scan errors): the query projection and the Scan destinations have diverged", attempted, scanFailures)
	}
	if scanFailures > 0 {
		logger.Warn("RebuildBlogListingAction: some blog posts were skipped by the scan — the listing is being rebuilt WITHOUT them",
			zap.Int("scanned_ok", len(articles)), zap.Int("scan_failures", scanFailures))
	}
	return articles, nil
}

// findBlogPage locates the blog listing page for a site.
// Strategy 1: Look for page_type = 'blog-index' (canonical).
// Strategy 2: Fall back to name = 'blog' with page_type = 'content'
//
//	(created when blog-index was in unavailableBuilders).
//	When found by fallback, corrects page_type to 'blog-index'
//	so future runs and other actions (e.g. BlogEmptyCheck) behave consistently.
//
// Returns uuid.Nil, "", nil if no blog page exists (not an error).
func findBlogPage(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (uuid.UUID, string, error) {
	var blogPageID uuid.UUID
	var blogPageName string

	// Strategy 1: canonical page_type.
	// The status filter is load-bearing (bugs_open/052): without it an archived
	// blog-index page is a valid listing target, so the action would rebuild a
	// listing on a page the site has deliberately retired.
	err := db.QueryRowContext(ctx, `
		SELECT id, name FROM pages
		WHERE site_id = $1 AND page_type = 'blog-index'
		  AND status IN ('active', 'deployed')
		LIMIT 1
	`, siteID).Scan(&blogPageID, &blogPageName)

	if err == nil {
		return blogPageID, blogPageName, nil
	}
	if err != sql.ErrNoRows {
		return uuid.Nil, "", fmt.Errorf("failed to query blog-index page: %w", err)
	}

	// Strategy 2: page named 'blog' created as content type.
	// Same status filter, and here it also guards a WRITE: on a hit this
	// strategy stamps page_type='blog-index' onto the row it found.
	err = db.QueryRowContext(ctx, `
		SELECT id, name FROM pages
		WHERE site_id = $1 AND name = 'blog' AND page_type = 'content'
		  AND status IN ('active', 'deployed')
		LIMIT 1
	`, siteID).Scan(&blogPageID, &blogPageName)

	if err == sql.ErrNoRows {
		return uuid.Nil, "", nil
	}
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to query blog page by name: %w", err)
	}

	// Found by fallback — fix page_type for future consistency
	_, fixErr := db.ExecContext(ctx, `
		UPDATE pages SET page_type = 'blog-index' WHERE id = $1
	`, blogPageID)
	if fixErr != nil {
		logger.Warn("findBlogPage: failed to fix page_type to blog-index",
			zap.String("page_id", blogPageID.String()),
			zap.Error(fixErr))
	} else {
		logger.Info("findBlogPage: corrected page_type from content to blog-index",
			zap.String("page_id", blogPageID.String()),
			zap.String("page_name", blogPageName))
	}

	return blogPageID, blogPageName, nil
}

// previousArticleCount returns how many articles the existing listing component
// carries, or -1 when there is nothing to compare against (no component yet, or
// the row cannot be read). -1 rather than 0 so a genuinely empty previous set is
// distinguishable from an absent one — 0 would make every first build look like
// a shrink from nothing.
//
// Failure is deliberately quiet and non-fatal: this feeds a diagnostic warning,
// so a malformed content_data must not stop the rebuild it is only commenting
// on. jsonb_array_length would ERROR on an object-shaped value, hence the
// jsonb_typeof guard — the same guard, for the same reason, as
// queryresolve.ListedPageEligibilitySQL.
func previousArticleCount(ctx context.Context, db *sql.DB, componentID uuid.UUID, logger *zap.Logger) int {
	if componentID == uuid.Nil {
		return -1
	}
	var n int
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(
		         CASE WHEN jsonb_typeof(content_data->'articles') = 'array'
		              THEN jsonb_array_length(content_data->'articles') END, -1)
		FROM page_components WHERE id = $1
	`, componentID).Scan(&n)
	if err != nil {
		logger.Debug("previousArticleCount: could not read previous listing size",
			zap.String("component_id", componentID.String()), zap.Error(err))
		return -1
	}
	return n
}

// blogSlotOrigin records HOW a listing slot name was resolved, because that —
// and not the presence of a component id — is what licenses writing there.
//
// bugs_open/457. The old findBlogListingSlot returned (name, componentID) and
// the caller read a nil id as "this slot is free". It never meant that: only
// strategy 1 looked in page_components at all, so a name resolved from the
// page's plan, or defaulted to, arrived with a nil id whether or not rows
// already occupied that slot. The caller then took the INSERT arm and appended
// a row — every run, for as long as resolution kept missing. On
// boxingonline.com's articles-index that was six orphan rows in two days, all
// at the hard-coded position 3, all with a NULL component_id, each freezing the
// listing template of its own birthday into rendered_html. The page served
// thirty-six cards where six belong. Migration 316's byte-identical guard did
// not prevent the accumulation; it only reported it, on the first run whose
// render happened to match a row already there, by failing the whole action.
//
// So the origin is the authority:
//
//   - a row we FOUND may be refreshed — that is the whole job;
//   - a slot the PLAN names as a listing slot may be created, because the plan
//     is what declares a page's sections;
//   - a name we GUESSED, or DEFAULTED to, may not be written at all. A guessed
//     name is by construction not a listing slot — strategy 1 or 2a would have
//     matched if it were — and on the motivating page it resolved to
//     `generic-text-block`, a slot that page uses for prose. Writing there
//     appends to, or overwrites, someone else's content.
type blogSlotOrigin int

const (
	slotOriginNone         blogSlotOrigin = iota // nothing resolved
	slotOriginExistingRow                        // strategy 1: a listing-class row is already present
	slotOriginPlanListing                        // strategy 2a: the plan names a listing-class slot
	slotOriginPlanFallback                       // strategy 2b: first non-skip section — a GUESS
	slotOriginDefault                            // strategy 3: the page declares no sections
)

func (o blogSlotOrigin) String() string {
	switch o {
	case slotOriginExistingRow:
		return "existing_row"
	case slotOriginPlanListing:
		return "plan_listing_slot"
	case slotOriginPlanFallback:
		return "plan_fallback_guess"
	case slotOriginDefault:
		return "default"
	default:
		return "none"
	}
}

// blogListingSlot is the resolved slot plus everything the write decision needs.
// Occupants is the count of non-removed rows already in that slot; Existing is
// the single occupant's row id and is set ONLY when Occupants == 1 — never a
// silently-picked first row out of several (the ambiguity rule
// portedPageComponentID states at adopt_verbatim.go).
type blogListingSlot struct {
	Name           string
	Origin         blogSlotOrigin
	Existing       uuid.UUID
	Occupants      int
	PlanPos        int  // 1-based index in pages.sections; 0 = not named by the plan
	OccupancyKnown bool // false when the lookup errored — must NOT read as "empty"
}

// blogListingSlotOccupancy counts the non-removed rows already in a slot.
//
// It is deliberately WIDER than migration 316's own partial index, whose
// `build_status <> 'removed'` is NULL-blind (NULL <> 'removed' is NULL, so a
// NULL-build_status row is not indexed and cannot collide). Using the shared
// tombstone predicate instead means a row the index cannot see is still counted
// here, so the error direction is refusal, never an append.
func blogListingSlotOccupancy(ctx context.Context, db *sql.DB, pageID uuid.UUID, slotName string) (uuid.UUID, int, error) {
	var occupants int
	var firstID string
	err := db.QueryRowContext(ctx, `
		SELECT count(*), COALESCE((array_agg(id ORDER BY position, id))[1]::text, '')
		  FROM page_components
		 WHERE page_id = $1 AND slot_name = $2
		   AND `+pageComponentNotRemovedSQL+`
	`, pageID, slotName).Scan(&occupants, &firstID)
	if err != nil {
		return uuid.Nil, 0, err
	}
	if occupants != 1 {
		return uuid.Nil, occupants, nil
	}
	parsed, parseErr := uuid.Parse(firstID)
	if parseErr != nil {
		return uuid.Nil, occupants, fmt.Errorf("unparseable page_components id %q: %w", firstID, parseErr)
	}
	return parsed, occupants, nil
}

// findBlogListingSlot resolves the listing slot for a blog-index page and
// reports how it got there, plus who already occupies it.
//
// Every exit path now runs the occupancy lookup — that is the fix. Previously
// only strategy 1 did, which is why "no component id" and "nobody is here" got
// confused (bugs_open/457).
func findBlogListingSlot(ctx context.Context, db *sql.DB, blogPageID uuid.UUID, logger *zap.Logger) blogListingSlot {
	resolve := func(name string, origin blogSlotOrigin, planPos int) blogListingSlot {
		slot := blogListingSlot{Name: name, Origin: origin, PlanPos: planPos}
		existing, occupants, err := blogListingSlotOccupancy(ctx, db, blogPageID, name)
		if err != nil {
			// An error is NOT an empty slot. Leaving OccupancyKnown false makes
			// the caller refuse rather than write blind.
			logger.Error("findBlogListingSlot: occupancy lookup failed — refusing rather than assuming the slot is free",
				zap.String("slot_name", name), zap.Error(err))
			return slot
		}
		slot.Existing, slot.Occupants, slot.OccupancyKnown = existing, occupants, true
		return slot
	}

	// Strategy 1: an existing page_components row in a known listing slot.
	for _, candidate := range slotPriority {
		var probe uuid.UUID
		err := db.QueryRowContext(ctx, `
			SELECT id FROM page_components
			WHERE page_id = $1 AND slot_name = $2 AND `+pageComponentNotRemovedSQL+`
			LIMIT 1
		`, blogPageID, candidate).Scan(&probe)
		if err == nil {
			logger.Info("findBlogListingSlot: found an existing listing component by slot_name",
				zap.String("slot_name", candidate))
			return resolve(candidate, slotOriginExistingRow, 0)
		}
	}

	// Strategy 2: the page's own plan.
	var sectionsJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT sections FROM pages WHERE id = $1 AND sections IS NOT NULL
	`, blogPageID).Scan(&sectionsJSON)
	if err == nil && len(sectionsJSON) > 0 {
		var sections []string
		if json.Unmarshal(sectionsJSON, &sections) == nil {
			// 2a: the plan names a listing-class slot. This is the one path that
			// may legitimately CREATE a row — the plan declares the slot.
			for _, candidate := range slotPriority {
				for i, section := range sections {
					if section == candidate {
						logger.Info("findBlogListingSlot: the page plan names a listing slot",
							zap.String("slot_name", candidate), zap.Int("plan_position", i+1))
						return resolve(candidate, slotOriginPlanListing, i+1)
					}
				}
			}
			// 2b: the old "first non-skip section" guess. Resolved so the refusal
			// can name it, never written to — see blogSlotOrigin.
			skipSections := map[string]bool{
				"hero": true, "header": true, "footer": true,
				"head": true, "call-to-action": true, "cta": true,
			}
			for i, section := range sections {
				if !skipSections[section] {
					logger.Info("findBlogListingSlot: no listing slot in the plan; first content section is only a guess",
						zap.String("guessed_slot", section), zap.Int("plan_position", i+1))
					return resolve(section, slotOriginPlanFallback, i+1)
				}
			}
		}
	}

	// Strategy 3: the page declares no sections at all.
	logger.Info("findBlogListingSlot: the page declares no sections; 'blog-listing' is a default, not a plan")
	return resolve("blog-listing", slotOriginDefault, 0)
}

// blogListingOp is what the caller should do with a resolved slot.
type blogListingOp int

const (
	opRefuseUnknown         blogListingOp = iota // occupancy unreadable
	opUpdate                                     // exactly one occupant — refresh it
	opInsert                                     // empty slot the plan authorises
	opRefuseAmbiguous                            // several occupants
	opRefuseNoSlotAuthority                      // nothing licenses writing here
)

// decideBlogListingWrite is deliberately pure so the whole decision table can be
// tested without a database — including the cases that must NEVER write.
//
// The trap it exists to close (bugs_open/457): binding a real component_id on
// the INSERT would stop the new row colliding with the NULL-component_id rows
// already present, because uq_page_components_no_byte_identical_duplicate is
// NULLS NOT DISTINCT. That is the only thing currently reporting the
// duplication, so a fix that binds the id WITHOUT deciding the write in Go
// first would silently append a seventh row instead of failing loudly. The
// decision below never consults the constraint; it is taken before the write.
func decideBlogListingWrite(slot blogListingSlot) (blogListingOp, string) {
	if !slot.OccupancyKnown {
		return opRefuseUnknown, "could not read what already occupies the slot"
	}
	if slot.Occupants > 1 {
		return opRefuseAmbiguous, fmt.Sprintf("%d rows already occupy slot %q — refusing to guess which is the listing", slot.Occupants, slot.Name)
	}
	if slot.Occupants == 1 {
		return opUpdate, ""
	}
	switch slot.Origin {
	case slotOriginPlanListing:
		return opInsert, ""
	case slotOriginExistingRow:
		// Strategy 1 matched a row and the count then said zero: the row went
		// between the two queries. Do not fall through to a write.
		return opRefuseUnknown, "the listing row disappeared between resolution and the occupancy count"
	default:
		return opRefuseNoSlotAuthority, fmt.Sprintf("no listing slot is declared for this page (slot %q resolved by %s); refusing to invent one", slot.Name, slot.Origin)
	}
}

// loadContentListingTemplate loads the content-listing template from
// content_components. This component has a proper input_schema with
// {{range .articles}} and article card markup.
// Falls back to a minimal default if not found.
// It also returns the id of the component the template came from, so a row this
// action creates can be attributed to the thing that actually rendered it
// (bugs_open/457 candidate 2; bugs_open/425 fix-candidate 5 names the same
// carelessness). uuid.Nil means "no honest id" and the caller must write NULL —
// the built-in default below is a Go constant with no content_components row,
// and inventing an id there would claim bytes came from a component that did
// not produce them.
//
// ⚠ `function = 'content-listing'` is NOT unique: measured 2026-09-03 there are
// two active rows, the shared `content-listing` and a per-site fork
// (`content-listing-guides-boxingonline-com`), byte-identical today. The old
// `ORDER BY created_at DESC LIMIT 1` therefore silently served the newest —
// i.e. one site's fork became every site's listing template, and would diverge
// the moment anyone edited it. Prefer the CANONICAL row (name = function) and
// log the ambiguity rather than refusing: RFC_034 makes several forks per
// function the intended future, so refusing here would stop every site's
// listing rebuilding the moment a second fork exists.
func loadContentListingTemplate(ctx context.Context, db *sql.DB, logger *zap.Logger) (string, uuid.UUID) {
	var tmpl, idStr string
	var candidates int

	// Primary: content-listing function (has proper input_schema).
	// name = function first, then oldest — the canonical row either way.
	err := db.QueryRowContext(ctx, `
		SELECT html_template, id::text,
		       count(*) OVER ()
		  FROM content_components
		 WHERE function = 'content-listing'
		   AND is_active = true
		   AND html_template IS NOT NULL
		   AND html_template != ''
		   AND html_template LIKE '%range%'
		 ORDER BY (name = function) DESC, created_at ASC
		 LIMIT 1
	`).Scan(&tmpl, &idStr, &candidates)
	if err == nil && tmpl != "" {
		parsed, _ := uuid.Parse(idStr)
		if candidates > 1 {
			logger.Warn("RebuildBlogListingAction: several active components share function 'content-listing' — using the canonical one",
				zap.Int("candidates", candidates),
				zap.String("chosen_component_id", idStr))
		}
		logger.Info("RebuildBlogListingAction: Using content-listing template",
			zap.String("component_id", idStr))
		return tmpl, parsed
	}

	// Fallback: article_grid by name
	err = db.QueryRowContext(ctx, `
		SELECT html_template, id::text FROM content_components
		WHERE name = 'article_grid'
		  AND is_active = true
		  AND html_template LIKE '%range%'
		LIMIT 1
	`).Scan(&tmpl, &idStr)
	if err == nil && tmpl != "" {
		parsed, _ := uuid.Parse(idStr)
		logger.Info("RebuildBlogListingAction: Using article_grid template as fallback",
			zap.String("component_id", idStr))
		return tmpl, parsed
	}

	// Last resort — minimal template. No DB row, so no honest component id.
	logger.Warn("RebuildBlogListingAction: No listing template found, using built-in default (row will carry a NULL component_id)")
	return defaultBlogListingTemplate, uuid.Nil
}

// ensureArticleLinks checks rendered HTML for article titles that lack
// <a href> links and patches them using the article data.
// This handles the case where the content-listing template has
// <h3 class="article-card__title">{{.title}}</h3> without a link.
func ensureArticleLinks(html string, articles []map[string]interface{}) string {
	for _, article := range articles {
		title, _ := article["title"].(string)
		url, _ := article["url"].(string)
		if title == "" || url == "" {
			continue
		}

		// Look for title text NOT already wrapped in an <a> tag
		// Pattern: <h3...>Title</h3> where Title is not inside <a>
		escapedTitle := regexp.QuoteMeta(title)
		// Match: >Title</h3  (title directly inside heading, no <a>)
		pattern := regexp.MustCompile(`(>[^<]*)` + escapedTitle + `([^<]*</h[1-6])`)
		if pattern.MatchString(html) {
			// Check it's not already linked
			linkedPattern := regexp.MustCompile(`<a[^>]*>` + escapedTitle + `</a>`)
			if !linkedPattern.MatchString(html) {
				// Replace bare title with linked title
				old := `>` + title + `</h`
				new := `><a href="` + url + `">` + title + `</a></h`
				html = strings.Replace(html, old, new, 1)
			}
		}
	}
	return html
}

// estimateReadTime returns a human-readable read time estimate
// based on approximate content length in characters.
// Assumes ~5 chars per word, 200 words per minute.
func estimateReadTime(contentLengthChars int) string {
	if contentLengthChars == 0 {
		return ""
	}
	words := contentLengthChars / 5
	minutes := words / 200
	if minutes < 1 {
		minutes = 1
	}
	return fmt.Sprintf("%d min read", minutes)
}

// renderBlogTemplate renders a Go template with blog data. It is the THIRD
// Go-template executor in this package and the second that renders a COMPONENT's
// html_template — found 2026-08-21 by render_seam_one_spelling_test.go, which is
// the whole reason that test exists.
//
// ⚠ ITS LANGUAGE IS NOT THE COMPONENT SEAM'S. No FuncMap and no
// missingkey=zero, so {{safe}}, {{default}} and {{isset}} — ordinary in every
// component template — are PARSE ERRORS here. Same divergence as
// RenderTemplateWithMap (bugs_closed/260 §13g), except this path is LIVE:
// `rebuild_blog_listing` is a registered action.
//
// THE SILENT SUBSTITUTION IS REMOVED (2026-08-21). A parse failure used to log
// at Warn and quietly render `defaultBlogListingTemplate` instead — so a site
// whose listing template the estate could not parse shipped a GENERIC listing,
// with its own design silently replaced and nothing downstream able to tell.
// That is bugs_closed/260's defect in a second place: output that is
// well-formed, plausible, and not what the component said. It now returns an
// error, and the caller already fails the step on empty output, so the two
// failure modes finally agree.
//
// Measured before changing it (2026-08-21): ONE live blog-listing component,
// 6,413 bytes, and it uses none of the missing FuncMap names — so no live
// template can trip the parse error today, and this closes the door before
// anyone edits one rather than after.
func renderBlogTemplate(htmlTemplate string, data map[string]interface{}, logger *zap.Logger) (string, error) {
	tmpl, err := template.New("blog_listing").Parse(htmlTemplate)
	if err != nil {
		logger.Error("RebuildBlogListingAction: blog listing template failed to PARSE — refusing to substitute the generic default listing",
			zap.Error(err),
			zap.String("hint", "this executor has no FuncMap: {{safe}}, {{default}} and {{isset}} parse fine in a section component and NOT here"))
		return "", fmt.Errorf("blog listing template failed to parse: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		logger.Error("RebuildBlogListingAction: blog listing template failed to EXECUTE",
			zap.Error(err))
		return "", fmt.Errorf("blog listing template failed to execute: %w", err)
	}

	return buf.String(), nil
}

var defaultBlogListingTemplate = `<section class="blog-listing" data-component="content-listing">
  <div class="container">
    {{if .section_title}}<h2 class="section-title">{{.section_title}}</h2>{{end}}
    {{if .section_subtitle}}<p class="section-subtitle">{{.section_subtitle}}</p>{{end}}
    <div class="article-grid">
      {{range .articles}}
      <article class="article-card">
        <div class="article-card__content">
          <h3 class="article-card__title"><a href="{{.url}}">{{.title}}</a></h3>
          {{if .date}}<time class="article-card__date">{{.date}}</time>{{end}}
          {{if .read_time}}<span class="article-card__read-time">{{.read_time}}</span>{{end}}
          {{if .excerpt}}<p class="article-card__excerpt">{{.excerpt}}</p>{{end}}
        </div>
      </article>
      {{end}}
    </div>
  </div>
</section>`
