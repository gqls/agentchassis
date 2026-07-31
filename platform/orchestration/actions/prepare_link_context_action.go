// FILE: platform/orchestration/actions/prepare_link_context_action.go
// PrepareLinkContextAction prepares available pages context for content generation
// Used by page-content-writer to constrain internal links

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

// Source labels for link_context.source, so a run says where its allow-list came
// from rather than leaving it to be inferred from a count.
const (
	linkSourceDatabase      = "database"
	linkSourceCollectedData = "collected_data"
	linkSourceNone          = "none"
	linkSourceDisabled      = "disabled"
)

// linkContextUnavailableCode is the agent_error_log error_code for "this run
// could not establish what pages exist". It is deliberately distinct from "this
// site has no pages": the two states have opposite remedies (fix the plumbing
// vs. build some pages) and they used to be indistinguishable — both produced an
// empty list and a logger.Warn nobody reads. bugs_open/092.
const linkContextUnavailableCode = "LINK_CONTEXT_UNAVAILABLE"

// maxLinkablePagesInPrompt bounds how many pages may be listed in one prompt.
// Inert today — the largest site in the fleet has 99 linkable pages (measured
// 2026-07-31, mean 30) — and present so a news or tool site that grows into the
// thousands cannot silently turn the writer prompt into a page index. When it
// fires it is REPORTED, in the log and in the action output: a cap that
// truncates quietly reads as "these are all the pages", which is precisely the
// false statement this action exists to stop making.
const maxLinkablePagesInPrompt = 200

// PrepareLinkContextAction extracts available pages and builds link constraint text
// Config:
//   - pages_field: path to pages array (default: "db_sync.pages")
//   - max_links_per_section: int (default: 3)
//   - enabled: bool (default: true)
//   - site_id: optional literal or dot-path override for site resolution
//
// Output:
//   - available_pages: []map{url,title,name,description}
//   - link_constraint_text: string (for prompt inclusion)
//   - page_count: int
//   - source: where the list came from (database | collected_data | none | disabled)
//   - db_consulted: bool — whether the authoritative table was actually read
//   - degraded: bool — true only when the list could not be ESTABLISHED
//   - reason: human-readable statement of what happened
//
// ── bugs_open/092 ────────────────────────────────────────────────────────────
// This action used to look for its page list ONLY in collected_data, at a
// configured field plus three fallbacks. None of the four has ever existed on
// page-content-writer's orchestration — its sole live consumer — so it returned
// an empty list on 26 of 26 runs (100%, measured 2026-07-31, unchanged since the
// defect was first recorded as concept-register LNK-017 on 2026-06-12). An empty
// list produced an empty constraint text, the prompt template's
// {{if .link_context.link_constraint_text}} guard then elided the whole
// "## Internal Linking" block, and the model wrote links with no idea what
// exists. Every layer of that failure was silent.
//
// Two things changed:
//
//  1. The DATABASE is now the authoritative source, read here, under exactly the
//     predicate validate_page_content's loadValidPagePaths uses to decide what
//     counts as a phantom_link. The writer's allow-list and the gate's
//     accept-set are therefore the same set by construction. Any other source
//     can disagree with the gate, and a writer punished for obeying its own
//     instructions is the drift class this codebase keeps paying for.
//     collected_data remains the fallback, so a workflow that does declare a
//     page list still has it honoured.
//
//  2. An empty list now produces an EXPLICIT "do not create internal links"
//     instruction instead of silence. Returning "" meant the one input where
//     guidance matters most produced no guidance at all — a fail-open dressed as
//     a no-op.
func PrepareLinkContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("PrepareLinkContextAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Check if enabled. An explicit opt-out is the one case where saying nothing
	// is right: the workflow has declared it does not want this block.
	enabled := true
	if v, ok := config["enabled"].(bool); ok {
		enabled = v
	}
	if !enabled {
		return map[string]interface{}{
			"enabled":              false,
			"available_pages":      []interface{}{},
			"link_constraint_text": "",
			"page_count":           0,
			"source":               linkSourceDisabled,
			"db_consulted":         false,
			"degraded":             false,
			"reason":               "step disabled by config",
		}, nil
	}

	// Get max links per section. JSON round-trips give float64; a hand-built
	// step config in a test gives int. Accept both rather than silently keeping
	// the default on one of them.
	maxLinks := 3
	switch v := config["max_links_per_section"].(type) {
	case float64:
		maxLinks = int(v)
	case int:
		maxLinks = v
	}

	var (
		pages       []linkablePage
		source      = linkSourceNone
		reason      string
		dbConsulted bool
		dbFailure   string
		truncated   int
		siteIDStr   = resolveLinkContextSiteID(params)
	)

	// ── 1. the authoritative source ──────────────────────────────────────────
	switch {
	case params.DB == nil:
		dbFailure = "no database handle on this action"
	case siteIDStr == "":
		dbFailure = "site_id could not be resolved from config or collected_data"
	default:
		siteID, err := uuid.Parse(siteIDStr)
		if err != nil {
			dbFailure = fmt.Sprintf("site_id %q is not a uuid", siteIDStr)
			break
		}
		loaded, over, err := loadLinkablePages(ctx, params.DB, siteID, maxLinkablePagesInPrompt, logger)
		if err != nil {
			dbFailure = "page query failed: " + err.Error()
			break
		}
		dbConsulted = true
		truncated = over
		if len(loaded) > 0 {
			pages = loaded
			source = linkSourceDatabase
			reason = fmt.Sprintf("%d linkable page(s) read from the pages table", len(pages))
		}
	}

	// ── 2. the fallback ──────────────────────────────────────────────────────
	// Reached when the database could not be consulted, or when it was and the
	// site genuinely has no pages yet — during an initial build a planned page
	// list is better than nothing, and it is what the configured field is for.
	if len(pages) == 0 {
		if fallback := extractPagesForLinking(params.CollectedData, config, logger); len(fallback) > 0 {
			pages = fallback
			source = linkSourceCollectedData
			reason = fmt.Sprintf("%d page(s) from collected_data (database %s)",
				len(pages), describeDBOutcome(dbConsulted, dbFailure))
		}
	}

	// ── 3. what an empty list means, said out loud ───────────────────────────
	// degraded is NOT "the list is empty" — it is "the list could not be
	// established". A brand-new site with no pages is a correct empty list and
	// must not read as a broken one.
	degraded := len(pages) == 0 && !dbConsulted
	if reason == "" {
		switch {
		case degraded:
			reason = "link context UNAVAILABLE — " + dbFailure
		case len(pages) == 0:
			reason = "site has no linkable pages yet"
		}
	}

	constraintText := buildLinkConstraintText(pages, maxLinks)

	// A durable record whenever the authority could not be read. Best-effort:
	// this is context for an operator, never a reason to fail the build.
	if dbFailure != "" {
		recordLinkContextUnavailable(ctx, params, siteIDStr, dbFailure, source, len(pages), degraded, logger)
	}

	logFields := []zap.Field{
		zap.Int("pages_found", len(pages)),
		zap.Int("max_links", maxLinks),
		zap.String("source", source),
		zap.Bool("db_consulted", dbConsulted),
		zap.Bool("degraded", degraded),
		zap.String("site_id", siteIDStr),
		zap.String("reason", reason),
	}
	// != 0, not > 0: -1 is the "over the cap, exact figure unreadable" marker and
	// must warn like any other truncation, not slip through as "no truncation".
	if truncated != 0 {
		logFields = append(logFields, zap.Int("pages_omitted_by_cap", truncated))
		logger.Warn("PrepareLinkContextAction: page list TRUNCATED by the prompt cap — the writer is being told these are all the pages and they are not",
			logFields...)
	} else if degraded {
		logger.Warn("PrepareLinkContextAction: link context unavailable — the writer is being told to emit NO internal links",
			logFields...)
	} else {
		logger.Info("PrepareLinkContextAction: Complete", logFields...)
	}

	out := map[string]interface{}{
		"enabled":              true,
		"available_pages":      linkablePagesAsMaps(pages),
		"link_constraint_text": constraintText,
		"page_count":           len(pages),
		"source":               source,
		"db_consulted":         dbConsulted,
		"degraded":             degraded,
		"reason":               reason,
	}
	if siteIDStr != "" {
		out["site_id"] = siteIDStr
	}
	if truncated != 0 {
		out["pages_omitted_by_cap"] = truncated
	}
	return out, nil
}

// linkablePage is one entry of the writer's allow-list. Deliberately narrower
// than PageInfo: only the four fields the prompt actually uses, so a run's
// collected_data does not carry a zero uuid and two untagged struct fields into
// every orchestration row.
type linkablePage struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func linkablePagesAsMaps(pages []linkablePage) []interface{} {
	out := make([]interface{}, 0, len(pages))
	for _, p := range pages {
		out = append(out, map[string]interface{}{
			"url":         p.URL,
			"title":       p.Title,
			"name":        p.Name,
			"description": p.Description,
		})
	}
	return out
}

// resolveLinkContextSiteID finds the site this run is writing for.
//
// It reuses the package's shared extractSiteID first, then adds the two
// locations that extractSiteID does not know about. That order is deliberate:
// widening the SHARED extractor would change what every one of its five other
// callers resolves — several of which treat "" as "skip this work" — and a
// silent behaviour change to five unrelated actions is not something a bug patch
// gets to make. The extra locations are added here, where the blast radius is
// this action.
//
// input_data.site_id is the one that matters: it is present on 26 of 26
// page-content-writer orchestrations, while site_record, top-level site_id and
// db_sync are present on 0 (measured 2026-07-31).
func resolveLinkContextSiteID(params ActionParams) string {
	// An explicit config value wins, and may itself be a dot-path — same
	// convention as validate_page_content's site_id.
	if id := resolveConfigString(params.StepConfig.Config, "site_id", params.CollectedData, params.Logger); id != "" {
		return id
	}
	if id := extractSiteID(params.CollectedData, params.Logger); id != "" {
		return id
	}
	for _, path := range []string{"input_data.site_id", "site_specs.site_id", "site_record.id"} {
		if id := extractNestedString(params.CollectedData, path); id != "" {
			return id
		}
	}
	return ""
}

// loadLinkablePages returns the pages this site's writer may link to, newest
// truth from the pages table.
//
// The predicate is COPIED FROM loadValidPagePaths (validate_page_content.go) on
// purpose, and the two must stay identical: that function decides what the
// deploy gate calls a phantom_link, so any divergence produces links the writer
// was told to emit and the gate then flags. Measured 2026-07-31, the fleet has
// exactly two page statuses — active (449) and archived (23) — so this predicate
// and loadActivePagesForLinkContext's status='active' are the same set today;
// this one is used because it is the gate's.
//
// Returns (pages, omittedByCap, error). An error means the list is NOT
// trustworthy and the caller must treat the context as unavailable rather than
// as an empty site: a truncated allow-list would quietly forbid links to pages
// that exist, which is harder to notice than having no list at all.
func loadLinkablePages(ctx context.Context, db *sql.DB, siteID uuid.UUID, limit int, logger *zap.Logger) ([]linkablePage, int, error) {
	if limit <= 0 {
		limit = maxLinkablePagesInPrompt
	}

	// limit+1 so an overflow is DETECTED rather than inferred from a full page.
	rows, err := db.QueryContext(ctx, `
		SELECT name,
		       url,
		       COALESCE(title, '')            AS title,
		       COALESCE(meta_description, '') AS description
		FROM pages
		WHERE site_id = $1
		  AND status NOT IN ('deleted', 'archived')
		  AND COALESCE(url, '') <> ''
		ORDER BY COALESCE(nav_order, 100) ASC, name ASC
		LIMIT $2
	`, siteID, limit+1)
	if err != nil {
		return nil, 0, fmt.Errorf("query pages: %w", err)
	}
	defer rows.Close()

	var pages []linkablePage
	for rows.Next() {
		var p linkablePage
		if err := rows.Scan(&p.Name, &p.URL, &p.Title, &p.Description); err != nil {
			return nil, 0, fmt.Errorf("scan page: %w", err)
		}
		// No URL is ever synthesised from a name (bugs_open/092 trap 2): a
		// constructed "/name.html" is a plausible-but-wrong address, which is
		// the bugs_closed/029 failure mode one layer upstream. The WHERE clause
		// already excludes empty urls; this is the belt to its braces.
		if p.URL == "" {
			continue
		}
		if p.Title == "" {
			p.Title = humanisePageName(p.Name)
		}
		pages = append(pages, p)
	}
	// A mid-iteration failure silently TRUNCATES the set, which is the same
	// hazard wearing a disguise — the same reasoning loadValidPagePaths records.
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("page list truncated by a row error: %w", err)
	}

	omitted := 0
	if len(pages) > limit {
		pages = pages[:limit]
		// The limit+1 probe only proves "there is at least one more". Reporting
		// that arithmetic as the omitted count would say 1 when 4,000 were
		// dropped — a truthful-looking number that is wrong by three orders of
		// magnitude. Ask for the real total instead; this costs one extra query
		// only on a site that has actually outgrown the cap.
		var total int
		if err := db.QueryRowContext(ctx, `
			SELECT count(*) FROM pages
			WHERE site_id = $1 AND status NOT IN ('deleted', 'archived') AND COALESCE(url, '') <> ''
		`, siteID).Scan(&total); err != nil || total <= limit {
			// Could not establish the true figure. -1 is the "more than the cap,
			// count unknown" marker — better than a confident wrong number.
			omitted = -1
			logger.Warn("loadLinkablePages: over the cap and the exact total could not be read",
				zap.String("site_id", siteID.String()), zap.Error(err))
		} else {
			omitted = total - limit
		}
		logger.Warn("loadLinkablePages: site has more linkable pages than the prompt cap",
			zap.String("site_id", siteID.String()),
			zap.Int("cap", limit),
			zap.Int("omitted", omitted))
	}

	logger.Info("loadLinkablePages: loaded",
		zap.Int("count", len(pages)),
		zap.String("site_id", siteID.String()))
	return pages, omitted, nil
}

// describeDBOutcome states, in one clause, what happened to the authoritative
// read — so the fallback's reason never reads as though the database agreed.
func describeDBOutcome(consulted bool, failure string) string {
	if consulted {
		return "returned none"
	}
	if failure == "" {
		return "not consulted"
	}
	return "NOT consulted: " + failure
}

// humanisePageName turns "learning-center" into "Learning Center" for the
// prompt's parenthetical. Replaces the deprecated strings.Title this file used
// to call; ASCII-only upper-casing of the first letter of each word is exactly
// what that call did for these inputs.
func humanisePageName(name string) string {
	if name == "" {
		return ""
	}
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' || r == ' ' })
	for i, w := range words {
		if w == "" {
			continue
		}
		r := []rune(w)
		if r[0] >= 'a' && r[0] <= 'z' {
			r[0] -= 'a' - 'A'
		}
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

// recordLinkContextUnavailable persists the fact that this run could not read
// the page list. Best-effort by design: a failure to record must never change
// the disposition the caller has already decided, so it returns nothing and the
// caller does not branch on it.
//
// Actions write agent_error_log with a direct INSERT rather than through
// orchestration.LogAgentError, which would be an import cycle (platform/
// orchestration imports this package).
func recordLinkContextUnavailable(
	ctx context.Context, params ActionParams, siteIDStr, failure, source string,
	pageCount int, degraded bool, logger *zap.Logger,
) {
	if params.DB == nil {
		return // nothing to write with; the log line above is the whole record
	}

	severity := "warning"
	outcome := fmt.Sprintf("fell back to %s (%d page(s))", source, pageCount)
	if degraded {
		// No list at all: the writer is being told to emit no internal links.
		// That is safe, and it is also a silently poorer page every time.
		severity = "error"
		outcome = "writer instructed to emit NO internal links"
	}

	contextJSON, err := json.Marshal(map[string]interface{}{
		"outcome":    outcome,
		"failure":    failure,
		"source":     source,
		"page_count": pageCount,
		"degraded":   degraded,
		"site_id":    siteIDStr,
		"bug":        "bugs_open/092",
	})
	if err != nil {
		logger.Warn("link context: could not marshal finding context", zap.Error(err))
		return
	}

	agentType := params.AgentType
	if agentType == "" {
		agentType = "unknown"
	}

	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO agent_error_log
		    (site_id, agent_type, step_name, action, error_message, error_code, severity, context, orchestration_id)
		VALUES (NULLIF($1,'')::uuid, $2, $3, 'prepare_link_context', $4, $5, $6, $7::jsonb, NULLIF($8,''))`,
		siteIDStr, agentType, params.ExecutionContext.StepName,
		fmt.Sprintf("Writer link context unavailable — %s; %s", failure, outcome),
		linkContextUnavailableCode, severity, string(contextJSON),
		params.ExecutionContext.OrchestrationID,
	); err != nil {
		logger.Warn("link context: failed to write finding record", zap.Error(err))
	}
}

// extractPagesForLinking gets pages from collected data. The FALLBACK source —
// see the action doc for why the database now comes first.
func extractPagesForLinking(collectedData map[string]interface{}, config map[string]interface{}, logger *zap.Logger) []linkablePage {
	var pages []linkablePage

	// Try configured field first
	pagesField := "db_sync.pages"
	if f, ok := config["pages_field"].(string); ok && f != "" {
		pagesField = f
	}

	// Extract pages array
	pagesData := datahelpers.ExtractNestedField(collectedData, pagesField)
	if pagesData == nil {
		// Try alternate locations
		alternates := []string{
			"site_plan.pages",
			"pages_to_build.pages",
			"render_context.available_pages",
		}
		for _, alt := range alternates {
			pagesData = datahelpers.ExtractNestedField(collectedData, alt)
			if pagesData != nil {
				logger.Debug("Found pages at alternate location", zap.String("field", alt))
				break
			}
		}
	}

	if pagesData == nil {
		logger.Debug("PrepareLinkContextAction: no page list in collected_data")
		return pages
	}

	// Convert to page slice
	pagesArray, ok := pagesData.([]interface{})
	if !ok {
		logger.Warn("PrepareLinkContextAction: pages is not an array")
		return pages
	}

	droppedNoURL := 0
	for _, p := range pagesArray {
		pageMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		page := linkablePage{
			URL:         datahelpers.GetStringField(pageMap, "url", ""),
			Title:       datahelpers.GetStringField(pageMap, "title", ""),
			Name:        datahelpers.GetStringField(pageMap, "name", ""),
			Description: datahelpers.GetStringField(pageMap, "description", ""),
		}

		// A page with no stored url is NOT a linkable target. This used to
		// synthesise "/" + name + ".html" — a hardcoded extension, not
		// NormalizePagePath and not the stored pages.url, which on a
		// dir/index.html site or a .html-less fleet is a confidently wrong
		// address handed to the writer as though it were real (bugs_open/092
		// trap 2). Dropping it is also what lets the database path supply the
		// truth instead.
		if page.URL == "" {
			droppedNoURL++
			continue
		}

		// Use name as title fallback
		if page.Title == "" && page.Name != "" {
			page.Title = humanisePageName(page.Name)
		}

		pages = append(pages, page)
	}

	if droppedNoURL > 0 {
		logger.Warn("PrepareLinkContextAction: dropped page entries with no url — a url is never synthesised from a name",
			zap.Int("dropped", droppedNoURL),
			zap.Int("kept", len(pages)))
	}

	return pages
}

// buildLinkConstraintText creates the constraint text for prompt inclusion.
//
// The empty case is the load-bearing one. It used to return "", which made the
// consuming template's {{if}} guard drop the section entirely — so a writer that
// knew nothing about the site was given no instruction at all and invented
// destinations. An empty list now says so explicitly, which is both true and the
// safest instruction available.
//
// No "## Internal Links" heading: page-content-writer's prompt template already
// emits "## Internal Linking" immediately above this text, so the old heading
// rendered a duplicate one line below it.
func buildLinkConstraintText(pages []linkablePage, maxLinks int) string {
	if len(pages) == 0 {
		return "There are NO pages available to link to on this site. " +
			"Do NOT create any internal links or hyperlinks in this content. " +
			"Write about topics in plain prose instead of linking to them.\n"
	}

	var sb strings.Builder

	sb.WriteString("When creating internal links, ONLY link to these pages:\n\n")

	for _, page := range pages {
		sb.WriteString(fmt.Sprintf("- %s", page.URL))
		if page.Title != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", page.Title))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nDo NOT create links to pages not in this list. ")
	sb.WriteString("Use each address EXACTLY as written above, including its file extension. ")
	sb.WriteString("If mentioning a topic without a corresponding page, write about it without making it a hyperlink.\n")

	if maxLinks > 0 {
		sb.WriteString(fmt.Sprintf("\nUse at most %d internal links per section.", maxLinks))
	}

	return sb.String()
}
