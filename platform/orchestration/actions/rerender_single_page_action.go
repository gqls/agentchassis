// FILE: platform/orchestration/actions/rerender_single_page_action.go
// RerenderSinglePageAction assembles a page from stored components
// Uses site_components for header/footer, page_components for sections
// Simple concatenation - no template re-rendering

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/content"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// Precompiled regexps used by sectionHasVisibleContent to strip non-visible
// content (style/script blocks, tags, entities, whitespace) before measuring
// the remaining text length. Package-level so they compile once at startup
// rather than on every section.
var (
	reStyleBlocks  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	reScriptBlocks = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reHTMLTags     = regexp.MustCompile(`<[^>]+>`)
	reHTMLEntities = regexp.MustCompile(`&[a-zA-Z#0-9]+;`)
	reWhitespace   = regexp.MustCompile(`\s+`)

	// reHeadClose matches the closing </head> tag, case-insensitively, so
	// injectComponentCSS can place the component <style> block after the
	// stylesheet <link>. Stored head components are hand-authored and their
	// casing is not guaranteed.
	reHeadClose = regexp.MustCompile(`(?i)</head>`)

	// reRuntimeFill matches the data-runtime-fill marker on a section that is
	// intentionally empty at build time and populated client-side by a loader
	// (e.g. the daily provocation card). Such sections must NOT be dropped by
	// sectionHasVisibleContent for lacking build-time text.
	reRuntimeFill = regexp.MustCompile(`(?i)data-runtime-fill`)
)

// RerenderSinglePageAction assembles a page from pre-rendered components
//
// Config:
//   - input_fields: fields to extract (default: ["page_id", "site_id", "domain"])
//
// Returns:
//   - html: assembled page HTML
//   - domain: site domain
//   - filename: page filename (e.g., "about.html")
//   - page_id: the page ID
//   - page_name: the page name
//   - skipped: true if page had no sections
func RerenderSinglePageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RerenderSinglePageAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Extract input fields
	inputFields := []string{"page_id", "site_id", "domain"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}

	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	// Get page_id
	pageIDStr, _ := extracted["page_id"].(string)
	if pageIDStr == "" {
		return nil, fmt.Errorf("page_id not found in input")
	}
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page_id: %w", err)
	}

	// Get page info (includes site_id, area_id)
	pageInfo, err := getPageInfo(ctx, params.DB, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to load page info: %w", err)
	}

	params.Logger.Info("RerenderSinglePageAction: Assembling page",
		zap.String("page_id", pageIDStr),
		zap.String("page_name", pageInfo.Name),
		zap.String("domain", pageInfo.Domain),
	)

	// Assemble the page from stored components
	html, assembly, err := assemblePage(ctx, params.DB, pageInfo, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble page: %w", err)
	}

	if html == "" {
		// bugs_open/095. Every empty assembly used to return the same
		// skipped=true, and page-rerender's check_skipped conditional routes
		// that to complete_skipped — a terminal step whose name contains
		// "complete" and whose status is COMPLETED. A page whose components
		// all failed to render was therefore indistinguishable from a page
		// that legitimately had nothing to build: rendered_html stayed NULL,
		// build_status never advanced, nothing deployed, and no error was
		// recorded anywhere.
		//
		// Split the two. Component rows that exist and contribute nothing is
		// a defect and fails the step; no component rows at all is a page
		// that has not been built yet and is still a legitimate skip.
		if assembly.assembledToNothingDespiteComponents() {
			return nil, fmt.Errorf(
				"page %q has %d component row(s) and assembled to nothing — %s",
				pageInfo.Name, assembly.ComponentRows, assembly.describe())
		}

		// Legitimate skip — but name what the page wanted, so the outcome is
		// attributable instead of a bare "no components found".
		reason := "page has no component rows"
		if len(assembly.PlannedSections) > 0 {
			reason = fmt.Sprintf("page has no component rows yet, but plans %d section(s): %v",
				len(assembly.PlannedSections), assembly.PlannedSections)
		}
		params.Logger.Warn("RerenderSinglePageAction: nothing to assemble, skipping",
			zap.String("page_name", pageInfo.Name),
			zap.String("reason", reason),
			zap.Strings("planned_sections", assembly.PlannedSections))
		return map[string]interface{}{
			"success":          false,
			"skipped":          true,
			"reason":           reason,
			"planned_sections": assembly.PlannedSections,
			"component_rows":   assembly.ComponentRows,
			"html":             "",
			"domain":           pageInfo.Domain,
			"filename":         pageInfo.Filename,
			"page_id":          pageIDStr,
			"page_name":        pageInfo.Name,
		}, nil
	}

	// Strip tool-doc headers from the OUTBOUND page HTML (019 §Tool Doc
	// Header). DB rendered_html keeps the header (audit/parse parity with the
	// template); only what ships is stripped. No-op when absent.
	html = content.StripToolDocHeader(html)

	// Repair dead internal links against pages.url before the page ships —
	// the same repair the initial-build gate applies (bugs_open/079). Without
	// it a rebuilt page redeploys 404s the gate would have caught
	// (bugs_open/097, diagnosis 9543aaf1). Same outbound-only philosophy as
	// the strip above: DB rendered_html keeps the unrepaired form.
	pageIndex, pageIndexOK := loadValidPagePaths(ctx, params.DB, pageInfo.SiteID, params.Logger)
	html = repairOutboundPageLinks(ctx, params, pageInfo.SiteID, pageInfo.Domain,
		pageInfo.Name, pageInfo.URL, html, pageIndex, pageIndexOK, params.Logger)

	// Collect JS assets for components used on this page.
	// Components with js_content get deployed as /tools/assets/{function}.js
	jsAssets := collectJSAssets(ctx, params.DB, pageID, params.Logger)

	// Build files map: HTML page + any JS asset files.
	// The page-rerender workflow uses files_field for multi-file git commit.
	files := map[string]interface{}{
		pageInfo.Filename: html,
	}
	for assetPath, js := range jsAssets {
		files[assetPath] = js
	}

	params.Logger.Info("RerenderSinglePageAction: Complete",
		zap.String("page_name", pageInfo.Name),
		zap.Int("html_length", len(html)),
		zap.Int("js_assets", len(jsAssets)),
	)

	return map[string]interface{}{
		"success":   true,
		"html":      html,
		"files":     files,
		"domain":    pageInfo.Domain,
		"filename":  pageInfo.Filename,
		"page_id":   pageIDStr,
		"page_name": pageInfo.Name,
	}, nil
}

// collectJSAssets queries content_components for js_content associated
// with the page's sections. Returns a map of asset path → JS content.
// Only returns entries for components that have non-empty js_content.
func collectJSAssets(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) map[string]string {
	assets := make(map[string]string)

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT cc.function, cc.js_content
		FROM page_components pc
		JOIN content_components cc ON pc.component_id = cc.id
		WHERE pc.page_id = $1
		  AND cc.js_content IS NOT NULL
		  AND cc.js_content != ''
		UNION
		-- Chrome (header/footer/head) is reached only through site_components,
		-- so its js_content was never published and its <script src> always
		-- 404'd (bugs_open/041) — the mobile menu is dead on every page. Mirrors
		-- the UNION already used by render_js_snippets_for_site_action.go:203-219
		-- for the snippets bundle, so the two JS paths agree that chrome exists.
		SELECT DISTINCT cc.function, cc.js_content
		FROM site_components sc
		JOIN content_components cc ON sc.component_id = cc.id
		WHERE sc.site_id = (SELECT site_id FROM pages WHERE id = $1)
		  AND cc.js_content IS NOT NULL
		  AND cc.js_content != ''
	`, pageID)
	if err != nil {
		logger.Warn("collectJSAssets: query failed", zap.Error(err))
		return assets
	}
	defer rows.Close()

	for rows.Next() {
		var function, jsContent string
		if err := rows.Scan(&function, &jsContent); err != nil {
			continue
		}
		assetPath := fmt.Sprintf("tools/assets/%s.js", function)
		// Strip any tool-doc header — the asset ships verbatim to the public
		// CDN path, and separateInlineJS copies script bodies into js_content
		// unmodified (019 §Tool Doc Header).
		assets[assetPath] = content.StripToolDocHeader(jsContent)
		logger.Info("collectJSAssets: found JS asset",
			zap.String("function", function),
			zap.Int("js_length", len(jsContent)))
	}

	return assets
}

// componentCSSMarker identifies the injected block. Also the idempotency
// guard: a head that already carries the marker is left alone, so an
// assembly path that runs twice cannot stack duplicate blocks.
const componentCSSMarker = `data-component-css="1"`

// collectComponentCSS returns a <style> block holding the css_snippets that
// apply to this page's components, or "" when none apply.
//
// Why this exists (bugs_open/072): a site's assets/css/styles.css is written
// ONLY by a webdesign-agent design run, and it carries the css_snippets that
// matched the site's component list AT THAT INSTANT. Nothing re-renders it
// when the site later gains a component, so a page can ship markup whose CSS
// was never written — measured on 2 of the 5 sites emitting .news-card, one
// of them 80 days after its stylesheet was last generated. Collecting the
// snippets HERE puts a component's CSS on the same path as its markup, so
// the two cannot drift apart: whatever assembles the page also styles it.
//
// Deliberately excludes any component whose stored rendered_html already
// carries its own <style> block. 86 of the 94 component functions in use
// ship their own CSS in html_template (the house pattern), and for those the
// snippet would be a second copy of the same rules. The exclusion is per
// component function, so one self-styling component on a page does not
// suppress the snippet another component on that page still needs.
func collectComponentCSS(ctx context.Context, db *sql.DB, pageID string, logger *zap.Logger) string {
	rows, err := db.QueryContext(ctx, `
		WITH page_funcs AS (
			SELECT cc.function AS fn,
			       bool_or(COALESCE(pc.rendered_html, '') LIKE '%<style%') AS ships_own_css
			FROM page_components pc
			JOIN content_components cc ON cc.id = pc.component_id
			WHERE pc.page_id = $1
			  AND cc.function IS NOT NULL
			  AND cc.function != ''
			GROUP BY cc.function
		)
		SELECT s.name, s.css_content
		FROM css_snippets s
		WHERE jsonb_typeof(s.applies_to) = 'array'
		  AND EXISTS (
		    SELECT 1
		    FROM jsonb_array_elements_text(s.applies_to) AS a(elem)
		    JOIN page_funcs f ON f.fn = a.elem
		    WHERE NOT f.ships_own_css
		  )
		ORDER BY s.name
	`, pageID)
	if err != nil {
		// Warn, never fail the assembly: a page that ships slightly
		// under-styled is better than a page that does not ship. The
		// jsonb_typeof guard above keeps one malformed applies_to row from
		// taking the query down for every page on the fleet.
		logger.Warn("collectComponentCSS: query failed", zap.Error(err))
		return ""
	}
	defer rows.Close()

	var css strings.Builder
	var names []string
	for rows.Next() {
		var name, cssContent string
		if err := rows.Scan(&name, &cssContent); err != nil {
			continue
		}
		if strings.TrimSpace(cssContent) == "" {
			continue
		}
		css.WriteString("\n/* component css_snippet: " + name + " */\n")
		css.WriteString(cssContent)
		names = append(names, name)
	}

	if css.Len() == 0 {
		return ""
	}

	logger.Info("collectComponentCSS: injecting component snippets",
		zap.String("page_id", pageID),
		zap.Strings("snippets", names),
		zap.Int("css_length", css.Len()))

	return "<style " + componentCSSMarker + ">" + css.String() + "\n</style>\n"
}

// injectComponentCSS places the block immediately before </head>, so it
// follows the <link> to the site stylesheet in document order. That ordering
// is load-bearing: where a site's frozen styles.css already holds an older
// copy of the same snippet, the fresher copy from css_snippets is the one
// that wins the tie.
//
// A head with no </head> (a truncated or hand-written component) still gets
// the block, prepended, rather than silently losing it.
func injectComponentCSS(headHTML, cssBlock string) string {
	if cssBlock == "" {
		return headHTML
	}
	if strings.Contains(headHTML, componentCSSMarker) {
		return headHTML
	}
	if loc := reHeadClose.FindStringIndex(headHTML); loc != nil {
		return headHTML[:loc[0]] + cssBlock + headHTML[loc[0]:]
	}
	return cssBlock + headHTML
}

// getPageInfo loads page metadata including site and area
func getPageInfo(ctx context.Context, db *sql.DB, pageID uuid.UUID) (*PageInfo, error) {
	var p PageInfo
	var areaID sql.NullString

	err := db.QueryRowContext(ctx, `
		SELECT 
			p.id, p.site_id, p.site_area_id, 
			p.name, COALESCE(p.title, p.name), p.url,
			COALESCE(p.meta_description, ''),
			s.domain
		FROM pages p
		JOIN sites s ON p.site_id = s.id
		WHERE p.id = $1
	`, pageID).Scan(
		&p.ID, &p.SiteID, &areaID,
		&p.Name, &p.Title, &p.URL,
		&p.MetaDesc,
		&p.Domain,
	)
	if err != nil {
		return nil, err
	}

	if areaID.Valid {
		id, _ := uuid.Parse(areaID.String)
		p.AreaID = &id
	}

	// Derive filename from URL
	if p.URL == "/" || p.URL == "" || p.Name == "index" {
		p.Filename = "index.html"
	} else {
		p.Filename = strings.TrimPrefix(p.URL, "/")
		if !strings.HasSuffix(p.Filename, ".html") {
			p.Filename = p.Filename + ".html"
		}
	}

	return &p, nil
}

// assemblePage combines site/area/page components into full HTML
func assemblePage(ctx context.Context, db *sql.DB, page *PageInfo, logger *zap.Logger) (string, pageAssembly, error) {
	// 1. Get site-level components
	siteComponents, err := getSiteComponents(ctx, db, page.SiteID)
	if err != nil {
		logger.Warn("Failed to load site components", zap.Error(err))
	}

	// 2. Get area-level overrides (if page is in an area)
	areaComponents := map[string]string{}
	if page.AreaID != nil {
		areaComponents, err = getAreaComponents(ctx, db, *page.AreaID)
		if err != nil {
			logger.Warn("Failed to load area components", zap.Error(err))
		}
	}

	// 3. Get page sections
	sections, assembly, err := getPageSections(ctx, db, page.ID, logger)
	if err != nil {
		logger.Warn("Failed to load page sections", zap.Error(err))
	}

	// No page-level content? Skip — we never deploy a page that is just
	// header + empty <main> + footer. siteComponents being non-empty is
	// expected on every page (header/footer/head live at site level), so
	// it isn't a useful signal here; sections is what determines whether
	// this page has anything to say.
	//
	// The caller decides what an empty assembly MEANS — see pageAssembly. It
	// is returned alongside so the caller can tell a page that has nothing to
	// build from one whose components all failed to contribute; this function
	// deliberately does not make that judgement, because the section editor
	// and the re-renderer answer it differently.
	if len(sections) == 0 {
		logger.Info("assemblePage: page assembled to nothing",
			zap.String("page_name", page.Name),
			zap.String("page_id", page.ID.String()),
			zap.Int("site_components", len(siteComponents)),
			zap.Int("component_rows", assembly.ComponentRows),
			zap.Strings("planned_sections", assembly.PlannedSections),
			zap.Strings("unrendered_slots", assembly.UnrenderedSlots),
			zap.Strings("blank_slots", assembly.BlankSlots),
		)
		return "", assembly, nil
	}

	// 4. Resolve components (area overrides site)
	head := resolveComponent(areaComponents, siteComponents, "head")
	header := resolveComponent(areaComponents, siteComponents, "header")
	footer := resolveComponent(areaComponents, siteComponents, "footer")

	// 5. Build page-specific head if we don't have one stored
	if head == "" {
		head = buildDefaultHead(page)
	} else {
		// Inject page-specific title into stored head component
		// The site-level head has <title></title> — replace with this page's title
		if page.Title != "" {
			titleRe := regexp.MustCompile(`<title>[^<]*</title>`)
			head = titleRe.ReplaceAllString(head, fmt.Sprintf("<title>%s</title>", page.Title))
		}
		// Inject meta description if the page has one and the head has an empty content=""
		if page.MetaDesc != "" {
			head = strings.Replace(head,
				`content="">`,
				fmt.Sprintf(`content="%s">`, page.MetaDesc), 1)
		}
	}

	// 5a. Inject per-page JSON-LD (schema.org). Structured data is how an answer
	// engine knows what a URL IS rather than guessing from prose, and the estate had
	// none: measured 2026-07-28, ZERO of 14 live sites emitted any application/ld+json.
	//
	// This is the head, deliberately. The obvious alternative — a page SECTION that
	// emits the script — cannot work: getPageSections drops any section with no
	// VISIBLE content (sectionHasVisibleContent strips <script> then requires >10
	// chars), so a metadata-only section renders correctly into page_components and is
	// silently discarded at assembly. Proven on relojistas, backed out, see
	// FLEET_GUIDANCE_discoverability.md.
	//
	// WebPage only, and nothing richer. We have Name/Title/URL/MetaDesc here and
	// nothing that tells us a page is an Article, a DefinedTerm or a product — and a
	// wrong @type is a false claim about the page, which is worse than no claim.
	// Richer per-type markup needs page_type plumbed through PageInfo first.
	head = injectPageJSONLD(head, page, logger)

	// 5b. Inject the CSS for this page's components (bugs_open/072). The site
	// stylesheet is frozen at the last design run, so a component added since
	// then has markup on the page and no rules anywhere; this puts its CSS on
	// the same path as its markup.
	head = injectComponentCSS(head, collectComponentCSS(ctx, db, page.ID.String(), logger))

	// 6. Assemble
	var html strings.Builder
	html.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n")
	html.WriteString(head)
	html.WriteString("\n<body>\n")

	if header != "" {
		html.WriteString(header)
		html.WriteString("\n")
	}

	html.WriteString("<main>\n")
	html.WriteString(sections)
	html.WriteString("\n</main>\n")

	if footer != "" {
		html.WriteString(footer)
		html.WriteString("\n")
	}

	html.WriteString("</body>\n</html>")

	logger.Debug("assemblePage: Complete",
		zap.String("page", page.Name),
		zap.Bool("has_header", header != ""),
		zap.Bool("has_footer", footer != ""),
		zap.Int("sections_length", len(sections)),
	)

	return html.String(), assembly, nil
}

// getSiteComponents loads site-level components (header, footer, head)
func getSiteComponents(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]string, error) {
	components := make(map[string]string)

	rows, err := db.QueryContext(ctx, `
		SELECT slot_name, rendered_html 
		FROM site_components 
		WHERE site_id = $1 AND rendered_html IS NOT NULL AND rendered_html != ''
	`, siteID)
	if err != nil {
		return components, err
	}
	defer rows.Close()

	for rows.Next() {
		var slot, html string
		if err := rows.Scan(&slot, &html); err != nil {
			continue
		}
		components[slot] = html
	}

	return components, nil
}

// getAreaComponents loads area-level component overrides
func getAreaComponents(ctx context.Context, db *sql.DB, areaID uuid.UUID) (map[string]string, error) {
	components := make(map[string]string)

	rows, err := db.QueryContext(ctx, `
		SELECT slot_name, rendered_html 
		FROM area_components 
		WHERE area_id = $1 AND rendered_html IS NOT NULL AND rendered_html != ''
	`, areaID)
	if err != nil {
		return components, err
	}
	defer rows.Close()

	for rows.Next() {
		var slot, html string
		if err := rows.Scan(&slot, &html); err != nil {
			continue
		}
		components[slot] = html
	}

	return components, nil
}

// pageAssembly records what the assembler actually saw, so that an EMPTY
// assembly can explain itself (bugs_open/095).
//
// An empty assembled page is ambiguous, and that ambiguity is the bug. Two
// states produce the identical empty string:
//
//   - the page has no component rows at all — nothing has been built yet.
//     Skipping is correct.
//   - the page HAS component rows and not one of them reached the output,
//     because none carries usable rendered_html or every one was dropped as
//     visually empty. That is a defect, and skipping it reports a page that
//     silently vanished as a success.
//
// Before this type existed both returned `skipped: true, reason: "no
// components found for page"`, which page-rerender's check_skipped
// conditional routes to complete_skipped — a terminal step whose name
// contains "complete" and whose status is COMPLETED. So the second case was
// shaped exactly like a success and left no error anywhere: not on the
// orchestration, not on the work item, not on the page row.
type pageAssembly struct {
	ComponentRows   int      // page_components rows for this page, unfiltered
	Contributed     int      // rows whose HTML reached the assembled output
	UnrenderedSlots []string // rows whose rendered_html is NULL or empty
	BlankSlots      []string // rows dropped as having no visible content
	PlannedSections []string // pages.sections — what the page says it wants
}

// assembledToNothingDespiteComponents is the defect shape: rows exist, none
// contributed. A page with no rows at all is a legitimate skip and is
// deliberately excluded — measured 2026-07-27, 17 active pages fleet-wide have
// planned sections and zero component rows (5 of them never built), and
// failing those would convert a correct no-op into a fleet-wide error.
func (a pageAssembly) assembledToNothingDespiteComponents() bool {
	return a.Contributed == 0 && a.ComponentRows > 0
}

// describe renders the two lists the operator needs and cannot otherwise see:
// what the page wanted, and what it actually had.
func (a pageAssembly) describe() string {
	return fmt.Sprintf("planned sections %v; %d component row(s), %d contributed; unrendered slots %v; blank slots %v",
		a.PlannedSections, a.ComponentRows, a.Contributed, a.UnrenderedSlots, a.BlankSlots)
}

// getPageSections loads and concatenates page sections in order,
// filtering out sections whose rendered_html has no visible text content.
// Empty sections (e.g. a hero with <h1></h1>, a CTA with no copy) produce
// blank space on the live site, so they're excluded from assembly and
// logged.
//
// The SQL no longer pre-filters on rendered_html (bugs_open/095): the rows it
// used to discard are exactly the evidence needed to tell "nothing built yet"
// from "everything failed to build", so they are now read and counted, and
// filtered here instead. The assembled output is unchanged.
func getPageSections(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) (string, pageAssembly, error) {
	var diag pageAssembly

	// What the page says it wants. Read separately from the component rows so
	// a page with sections planned and no rows at all is still describable.
	var sectionsJSON []byte
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(sections, '[]'::jsonb) FROM pages WHERE id = $1`, pageID,
	).Scan(&sectionsJSON); err == nil {
		_ = json.Unmarshal(sectionsJSON, &diag.PlannedSections)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(rendered_html, ''), COALESCE(slot_name, '')
		FROM page_components
		WHERE page_id = $1
		ORDER BY position ASC
	`, pageID)
	if err != nil {
		return "", diag, err
	}
	defer rows.Close()

	var sections strings.Builder
	sectionIdx := 0
	for rows.Next() {
		var html, slotName string
		if err := rows.Scan(&html, &slotName); err != nil {
			continue
		}
		diag.ComponentRows++

		if html == "" {
			// Never rendered. This is the row shape behind bugs_open/095: a
			// component row that looks deliberate — correct component, correct
			// position — but that nothing ever populated.
			diag.UnrenderedSlots = append(diag.UnrenderedSlots, slotName)
			continue
		}
		if !sectionHasVisibleContent(html) {
			// Dropping a blank section from the assembled page is correct — it
			// would only render as empty space — but doing it silently is how 9
			// blanked article bodies vanished unnoticed. Name the section so the
			// drop is attributable; the durable owner of this signal is the
			// empty_sections / required_fields_missing discovery checks, which
			// emit the work item.
			logger.Warn("getPageSections: dropping section with no visible content from assembled page",
				zap.String("page_id", pageID.String()),
				zap.String("slot_name", slotName),
				zap.Int("section_index", sectionIdx),
				zap.Int("html_length", len(html)),
			)
			diag.BlankSlots = append(diag.BlankSlots, slotName)
			sectionIdx++
			continue
		}
		sections.WriteString(html)
		sections.WriteString("\n")
		diag.Contributed++
		sectionIdx++
	}

	if len(diag.BlankSlots) > 0 || len(diag.UnrenderedSlots) > 0 {
		logger.Warn("getPageSections: filtered sections from assembled page",
			zap.String("page_id", pageID.String()),
			zap.Int("component_rows", diag.ComponentRows),
			zap.Int("contributed", diag.Contributed),
			zap.Strings("blank_slots", diag.BlankSlots),
			zap.Strings("unrendered_slots", diag.UnrenderedSlots),
		)
	}

	return sections.String(), diag, nil
}

// sectionHasVisibleContent reports whether the given rendered HTML should be kept in the
// assembled page. Sections explicitly marked data-runtime-fill are intentionally empty at
// build time (a client-side loader populates them) and are always kept. Otherwise the
// section is kept only if it contains more than 10 characters of visible text after
// stripping <style>/<script> blocks, HTML tags, HTML entities, and whitespace — sections
// with less than that aren't meaningful build-time content and represent a generation
// failure or a stale/not-yet-built empty shell, which shouldn't reach the deployed page.
func sectionHasVisibleContent(html string) bool {
	if html == "" {
		return false
	}
	// Runtime-filled sections (e.g. the daily provocation card) are intentionally empty
	// at build time; a browser-side loader fills them. Keep them regardless of build-time
	// text. Genuine empty shells do not carry this marker and are filtered below.
	if reRuntimeFill.MatchString(html) {
		return true
	}
	s := reStyleBlocks.ReplaceAllString(html, "")
	s = reScriptBlocks.ReplaceAllString(s, "")
	s = reHTMLTags.ReplaceAllString(s, "")
	s = reHTMLEntities.ReplaceAllString(s, "")
	s = reWhitespace.ReplaceAllString(s, "")
	return len(s) > 10
}

// injectPageJSONLD adds one schema.org WebPage block to the head, describing the page
// in terms of what we actually know about it. Idempotent, and a no-op when there is
// nothing truthful to say.
//
// Silent no-ops are deliberate here rather than partial output: a JSON-LD block naming
// a page with no title, or on a site with no domain, is a machine-readable assertion
// that is wrong, and wrong structured data is worse than none — search engines act on it.
func injectPageJSONLD(head string, page *PageInfo, logger *zap.Logger) string {
	// Every skip below is NAMED. The council's bug_historian seat objected that the
	// first version replaced one undetectable silence (metadata sections dropped by
	// sectionHasVisibleContent) with another (pages skipped for missing PageInfo
	// fields) — and it was right: the whole reason this change exists is that
	// JSON-LD's fleet-wide absence was only discovered by a manual measurement.
	// A skip here is legitimate, but it must never again be invisible.
	skip := func(reason string) string {
		if logger != nil && page != nil {
			logger.Debug("injectPageJSONLD: no structured data emitted",
				zap.String("reason", reason),
				zap.String("page", page.Name),
				zap.String("domain", page.Domain))
		}
		return head
	}
	if head == "" || page == nil {
		if logger != nil {
			logger.Debug("injectPageJSONLD: no structured data emitted", zap.String("reason", "empty head or nil page"))
		}
		return head
	}
	// Never emit twice: a stored head may already carry one from an earlier render.
	if strings.Contains(head, "application/ld+json") {
		return head
	}
	if page.Title == "" {
		return skip("page has no title")
	}
	if page.Domain == "" {
		return skip("page has no domain")
	}

	origin := "https://" + page.Domain
	pageURL := origin + page.URL

	doc := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "WebPage",
		"@id":      pageURL,
		"url":      pageURL,
		"name":     page.Title,
		"isPartOf": map[string]interface{}{
			"@type": "WebSite",
			"url":   origin,
			"name":  page.Domain,
		},
	}
	if page.MetaDesc != "" {
		doc["description"] = page.MetaDesc
	}

	// json.Marshal escapes <, > and & to \u003c/\u003e/\u0026 by default, which keeps
	// the payload safe to embed inside a <script> element — a title containing "</script>"
	// cannot break out. Do NOT switch to an Encoder with SetEscapeHTML(false).
	payload, err := json.Marshal(doc)
	if err != nil {
		return skip("json.Marshal failed: " + err.Error())
	}

	block := fmt.Sprintf("\n<script type=\"application/ld+json\">%s</script>\n", payload)
	if idx := strings.LastIndex(head, "</head>"); idx >= 0 {
		return head[:idx] + block + head[idx:]
	}
	return head + block
}

// resolveComponent returns area component if exists, otherwise site component
func resolveComponent(area, site map[string]string, slot string) string {
	if html, ok := area[slot]; ok && html != "" {
		return html
	}
	return site[slot]
}

// buildDefaultHead creates a basic head section for a page
// Used as fallback if no head stored in site_components
func buildDefaultHead(page *PageInfo) string {
	return fmt.Sprintf(`<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" href="/assets/css/styles.css">
</head>`, page.Title)
}
