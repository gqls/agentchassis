// FILE: platform/orchestration/actions/render_site_components_action.go
// RenderSiteComponentsAction renders header/footer/head for a site
// and stores them in site_components table for reuse across all pages

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// RenderSiteComponentsAction renders site-level components and stores them
//
// Config:
//   - site_id_field: path to site_id in collected_data (default: "site_record.site_id")
//   - domain_field: path to domain in collected_data (default: "site_record.domain")
//   - slots: which slots to render (default: ["header", "footer", "head"])
//   - force_rerender: re-render even if already exists (default: false)
//
// Returns:
//   - rendered: map of slot_name -> success boolean
//   - site_id: the site ID
func RenderSiteComponentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RenderSiteComponentsAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// Get site_id using configurable field path (matches UpdateSiteDefaultsAction pattern)
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		// Fallback: try legacy input_fields approach
		inputFields := []string{"site_id", "domain"}
		if fields, ok := config["input_fields"].([]interface{}); ok {
			inputFields = make([]string, len(fields))
			for i, f := range fields {
				inputFields[i], _ = f.(string)
			}
		}
		extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)
		siteIDStr, _ = extracted["site_id"].(string)
	}

	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id at %s: %w (got: %q, len: %d)", siteIDField, err, siteIDStr, len(siteIDStr))
	}

	// Get slots to render
	slots := []string{"header", "footer", "head"}
	if s, ok := config["slots"].([]interface{}); ok {
		slots = make([]string, len(s))
		for i, v := range s {
			slots[i], _ = v.(string)
		}
	}

	forceRerender, _ := config["force_rerender"].(bool)

	// Load site data
	siteData, err := loadSiteDataFull(ctx, params.DB, siteID, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load site data: %w", err)
	}

	params.Logger.Info("RenderSiteComponentsAction: Rendering components",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteData.Domain),
		zap.Strings("slots", slots),
	)

	// Load navigation for header/footer
	/*	navItems := loadNavItems(ctx, params.DB, siteID, 6, params.Logger)
		footerNavItems := loadFooterNavItems(ctx, params.DB, siteID, 10, params.Logger)*/
	// NavFetchableOnly: everything loaded here becomes an <a href> in the chrome
	// that ships on EVERY page of the site, so a nav item whose target was never
	// built is a site-wide 404 (bugs_open/049 mechanism 2 — one gaswholesalers
	// utility item accounted for 28 broken anchors). The old argument here was
	// `false`, justified as "runs during build when pages may not be deployed
	// yet"; that case is now handled inside GetNavItems, which serves the
	// unfiltered nav and warns rather than ever emptying the chrome.
	// maxItems=0: no limit — PopulateNavTablesAction controls primary group membership.
	navItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, params.Logger)
	footerNavItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupPrimary, NavGroupUtility, NavGroupLegal}, NavFetchableOnly, 0, params.Logger)

	// Build render context
	year := fmt.Sprintf("%d", time.Now().Year())
	copyright := fmt.Sprintf("© %s %s", year, siteData.CompanyName)

	// Build pre-rendered nav HTML for templates that use {{.nav_items_html}}
	// Uses existing buildNavItemsHTML from component_library.go
	// Header templates use nav_items_html (primary only).
	navItemsHTML := buildNavItemsHTML(navItems)

	// Build quick links HTML for footer — includes primary + utility items.
	// Utility items are pages that overflowed from primary or were classified
	// as secondary (FAQ, Approach, Insights etc). They belong in footer nav
	// but not the header. Legal items (privacy, terms) get their own footer section.
	quickLinksItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupPrimary, NavGroupUtility}, NavFetchableOnly, 0, params.Logger)
	quickLinksHTML := buildNavItemsHTML(quickLinksItems)

	// Build services HTML for footer "Our Services" column
	// Query pages that represent services (linked from services page or service-named pages)
	servicesHTML := buildServicesHTML(ctx, params.DB, siteID, params.Logger)

	// Convert NavItems to categories format for templates that use {{range .categories}}
	categories := make([]map[string]interface{}, len(navItems))
	for i, item := range navItems {
		categories[i] = map[string]interface{}{
			"name":  item.Label,
			"slug":  strings.TrimSuffix(strings.TrimPrefix(item.URL, "/"), ".html"),
			"url":   item.URL,
			"label": item.Label, // alias
		}
	}

	// Convert footer nav items similarly
	footerLinks := make([]map[string]interface{}, len(footerNavItems))
	for i, item := range footerNavItems {
		footerLinks[i] = map[string]interface{}{
			"name":  item.Label,
			"slug":  strings.TrimSuffix(strings.TrimPrefix(item.URL, "/"), ".html"),
			"url":   item.URL,
			"label": item.Label,
		}
	}

	// Build company links (about, contact, careers).
	// Also capture the real contact page URL for the header CTA, so it points
	// at an existing page instead of the hardcoded phantom /contact.html.
	companyLinks := []map[string]interface{}{}
	ctaURL := ""
	for _, item := range footerNavItems {
		lowerLabel := strings.ToLower(item.Label)
		if lowerLabel == "about" || lowerLabel == "contact" || lowerLabel == "careers" {
			companyLinks = append(companyLinks, map[string]interface{}{
				"name": item.Label,
				"url":  item.URL,
			})
		}
		if lowerLabel == "contact" {
			ctaURL = item.URL
		}
	}

	// Validate the header CTA against the SAME policy as the nav rendered beside
	// it; when there is no contact nav item (or it points at a phantom), fall
	// back to the same ranking the internal-link resolver uses (interactive
	// pages first, then hubs) rather than rendering no button at all.
	//
	// bugs_open/191: this was loadResolverPageSet, the page-CONTENT set, which
	// carries no deployment predicate. So the header shipped a CTA button to a
	// never-deployed page while the nav in the SAME component had that page
	// filtered out — mortgagecalculator.co.uk, a 404 on the wire on every page.
	// ChromeLinkPolicy is the nav's own decision, escapes and all: on a first
	// build or a failed lookup it is unfiltered, because this chrome is
	// idempotence-gated (the EXISTS probe below) and a button dropped here may
	// never get a second chance to render.
	chromeLinks := LoadChromeLinkPolicy(ctx, params.DB, siteID, params.Logger)
	if ctaURL == "" || !chromeLinks.Allows(ctaURL) {
		hubs, err := loadContentHubs(ctx, params, siteID, params.Logger)
		if err != nil {
			params.Logger.Warn("RenderSiteComponentsAction: loadContentHubs failed for header CTA fallback", zap.Error(err))
		}
		interactive, err := loadInteractivePages(ctx, params, siteID, params.Logger)
		if err != nil {
			params.Logger.Warn("RenderSiteComponentsAction: loadInteractivePages failed for header CTA fallback", zap.Error(err))
		}
		primary, _ := chooseCTATargets("", "", interactive, hubs)
		if primary.URL != "" && chromeLinks.Allows(primary.URL) {
			ctaURL = primary.URL
		} else {
			ctaURL = ""
		}
	}

	// Owner-set override (opt-in, sites.content_data->>'header_cta_url').
	// Applied AFTER the derivation so absence changes nothing anywhere, and
	// validated against the SAME ChromeLinkPolicy as the derived CTA — an
	// owner-set target can go stale, and a stale override must degrade to the
	// derived behaviour rather than ship a dead button (bugs_open/191's class).
	ctaText := "Get Started"
	if v := strings.TrimSpace(siteData.HeaderCTAURL); v != "" {
		if chromeLinks.Allows(v) {
			ctaURL = v
		} else {
			params.Logger.Warn("RenderSiteComponentsAction: header_cta_url override rejected by chrome link policy; using derived CTA",
				zap.String("override", v), zap.String("derived", ctaURL))
			emitCTAOverrideRejectedItem(ctx, params.DB, siteID, v, ctaURL, params.Logger)
		}
	}
	if v := strings.TrimSpace(siteData.HeaderCTAText); v != "" {
		ctaText = v
	}

	// Build legal links from real pages classified into the legal nav group.
	// Was a hardcoded {/privacy.html, /terms.html} slice — those pages do not
	// necessarily exist, so it produced phantom links. Now: only pages that
	// actually exist appear; if none, the list is empty and the footer renders
	// no legal links.
	legalNavItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupLegal}, NavFetchableOnly, 0, params.Logger)
	legalLinks := make([]map[string]interface{}, 0, len(legalNavItems))
	for _, item := range legalNavItems {
		legalLinks = append(legalLinks, map[string]interface{}{
			"name": item.Label,
			"url":  item.URL,
		})
	}

	// Social links (empty for now - could be populated from site data)
	socialLinks := []map[string]interface{}{}

	renderCtx := &RenderContext{
		Domain:         siteData.Domain,
		CompanyName:    siteData.CompanyName,
		Tagline:        siteData.Tagline,
		Email:          siteData.Email,
		Phone:          siteData.Phone,
		LogoText:       siteData.LogoText,
		LogoURL:        siteData.LogoURL,
		NavItems:       navItems,
		FooterNavItems: quickLinksItems,
		Year:           year,

		// Colors from style collection (RenderContext struct fields feed contextToInterfaceMap defaults)
		PrimaryColor:    siteData.PrimaryColor,
		SecondaryColor:  siteData.SecondaryColor,
		AccentColor:     siteData.AccentColor,
		TextColor:       siteData.TextColor,
		BackgroundColor: siteData.BackgroundColor,

		// Theme CSS from css_themes table
		ThemeCSS: siteData.ThemeCSS,

		ContentData: map[string]interface{}{
			// Core site info
			"company_name":  siteData.CompanyName,
			"brand_name":    siteData.CompanyName, // alias
			"tagline":       siteData.Tagline,
			"domain":        siteData.Domain,
			"email":         siteData.Email,
			"contact_email": siteData.Email, // alias
			"phone":         siteData.Phone,
			"logo_text":     siteData.LogoText,
			"logo_url":      siteData.LogoURL,
			"year":          year,
			"copyright":     copyright,

			// Pre-rendered nav HTML for templates using {{.nav_items_html}}
			"nav_items_html":   navItemsHTML,
			"quick_links_html": quickLinksHTML, // primary + utility items for footer
			"services_html":    servicesHTML,

			// Navigation - multiple formats for different templates
			"categories":       categories,   // for {{range .categories}}
			"nav_items":        categories,   // alias
			"footer_nav_items": footerLinks,  // for footer
			"quick_links":      footerLinks,  // alias for footer
			"company_links":    companyLinks, // about, contact, careers
			"legal_links":      legalLinks,   // privacy, terms
			"social_links":     socialLinks,  // social media (empty for now)

			// CTA defaults — cta_url resolved from the real contact page above
			// (empty when there is no contact page; the gated header template
			// then renders no CTA button rather than a phantom).
			"cta_text":       ctaText,
			"cta_url":        ctaURL,
			"subscribe_text": "Subscribe",
			"show_subscribe": false,

			// Newsletter defaults (can be overridden)
			"newsletter_title":       "Stay Updated",
			"newsletter_description": "Get the latest news and updates.",
			"email_placeholder":      "Enter your email",
		},
	}

	params.Logger.Info("RenderSiteComponentsAction: Render context built",
		zap.Int("nav_items", len(navItems)),
		zap.Int("nav_items_html_len", len(navItemsHTML)),
		zap.Bool("has_theme_css", siteData.ThemeCSS != ""),
		zap.String("primary_color", siteData.PrimaryColor),
	)

	// Render each slot
	rendered := make(map[string]bool)
	lockedSlots := []string{}
	ineligibleChrome := map[string]string{}
	chromeRenderFailed := map[string]string{} // slot -> render error (bugs_open/260)
	chromeUnserved := []string{}              // of those, the slots with nothing stored to serve
	for _, slot := range slots {
		success, locked, degraded, renderErr := renderAndStoreSiteComponent(ctx, params.DB, siteID, slot, renderCtx,
			forceRerender, chromeLinks,
			recordAbsentRequiredFields(params.StepConfig.Config),
			refuseAbsentRequiredFields(params.StepConfig.Config),
			params.Logger)
		rendered[slot] = success
		if locked {
			lockedSlots = append(lockedSlots, slot)
		}
		if degraded != "" {
			ineligibleChrome[slot] = degraded
		}
		if renderErr != nil {
			// bugs_open/260. SURFACED, not just logged (council a44d9eb8 round
			// 1, bug_historian): a degraded chrome slot reported only in a log
			// line, inside a step that still returns success, is the shape
			// bugs_closed/028 and bugs_closed/040 were about — "reports
			// complete" while the artefact is stale. So it lands in the action
			// result, AND it files a human-review item, AND the one case with
			// nothing to fall back on fails the step outright.
			chromeRenderFailed[slot] = renderErr.Error()
			if !chromeSlotHasStoredHTML(ctx, params.DB, siteID, slot) {
				chromeUnserved = append(chromeUnserved, slot)
			}
		}
	}

	if len(chromeRenderFailed) > 0 {
		params.Logger.Error("RenderSiteComponentsAction: chrome slot(s) failed to render",
			zap.Any("slots", chromeRenderFailed),
			zap.Strings("with_nothing_serving", chromeUnserved))
	}
	if len(chromeUnserved) > 0 {
		// A site must not go live with an empty header, footer or head. This is
		// a deliberate behaviour change for GREENFIELD builds, flagged as such
		// to the site-provisioning path (council a44d9eb8 round 1, guardian):
		// before this, a first build with an unexecutable chrome template
		// shipped the mangled regex render instead, and reported success.
		return nil, fmt.Errorf("site chrome could not be rendered for slot(s) %s and this site has nothing stored to serve instead (bugs_open/260)",
			strings.Join(chromeUnserved, ", "))
	}

	params.Logger.Info("RenderSiteComponentsAction: Complete",
		zap.Any("rendered", rendered),
		zap.Strings("locked_slots_preserved", lockedSlots),
		zap.Any("ineligible_chrome", ineligibleChrome),
	)

	// locked_slots_preserved reports human locks that refused a re-render
	// (bugs_open/069), mirroring save_page_sections' locked_sections_preserved.
	// It is NOT a failure: each one has filed a lock_blocked_change item.
	// ineligible_chrome names, per slot, a component this site fell back to
	// because the library holds NO active, unforked, chrome-level component for
	// that slot's function (bugs_open/118). It is a library defect, not a site
	// one, and it is reported rather than hidden because the version that hid it
	// left 11 of 14 sites rendering a deactivated footer for months.
	return map[string]interface{}{
		"success":                true,
		"site_id":                siteIDStr,
		"domain":                 siteData.Domain,
		"rendered":               rendered,
		"locked_slots_preserved": lockedSlots,
		"ineligible_chrome":      ineligibleChrome,
		// Empty on every healthy run. Non-empty means the slot's stored chrome
		// is STALE-BUT-SERVING because this run's render would not execute — a
		// degraded success, named so a reader of the result can see it without
		// reading the logs (bugs_open/260).
		"chrome_render_failed": chromeRenderFailed,
	}, nil
}

// SiteDataFull contains all site data needed for rendering
type SiteDataFull struct {
	ID          uuid.UUID
	Domain      string
	Name        string
	CompanyName string
	Tagline     string
	Email       string
	Phone       string
	// Owner-set header CTA override, read from sites.content_data
	// (header_cta_url / header_cta_text). Opt-in: empty = the derived
	// behaviour (contact nav item, then the resolver-ranked fallback)
	// that every existing site gets today.
	HeaderCTAURL  string
	HeaderCTAText string
	LogoText      string
	LogoURL       string

	// Style collection data (loaded via sites.style_collection_id)
	PrimaryColor    string
	SecondaryColor  string
	AccentColor     string
	TextColor       string
	TextLightColor  string
	BackgroundColor string
	BackgroundAlt   string
	ThemeCSS        string // from css_themes table
	FontURL         string // Google Fonts URL if set
}

func loadSiteDataFull(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (*SiteDataFull, error) {
	var s SiteDataFull
	s.ID = siteID

	var colorPaletteJSON sql.NullString
	var themeCSSContent sql.NullString

	err := db.QueryRowContext(ctx, `
		SELECT 
			si.domain,
			COALESCE(si.name, si.domain),
			COALESCE(si.company_name, si.name, si.domain),
			COALESCE(si.tagline, ''),
			COALESCE(si.email, ''),
			COALESCE(si.phone, ''),
			COALESCE(si.logo_text, si.company_name, si.name, si.domain),
			COALESCE(si.logo_url, ''),
			COALESCE(si.content_data->>'header_cta_url', ''),
			COALESCE(si.content_data->>'header_cta_text', ''),
			sc.color_palette::text,
			ct.css_content
		FROM sites si
		LEFT JOIN style_collections sc ON si.style_collection_id = sc.id
		LEFT JOIN css_themes ct ON sc.css_theme_id = ct.id
		WHERE si.id = $1
	`, siteID).Scan(
		&s.Domain, &s.Name, &s.CompanyName, &s.Tagline,
		&s.Email, &s.Phone, &s.LogoText, &s.LogoURL,
		&s.HeaderCTAURL, &s.HeaderCTAText,
		&colorPaletteJSON, &themeCSSContent,
	)
	if err != nil {
		return nil, err
	}

	// Parse color palette from style collection
	if colorPaletteJSON.Valid && colorPaletteJSON.String != "" {
		var palette map[string]string
		if jsonErr := json.Unmarshal([]byte(colorPaletteJSON.String), &palette); jsonErr == nil {
			s.PrimaryColor = palette["primary"]
			s.SecondaryColor = palette["secondary"]
			s.AccentColor = palette["accent"]
			s.TextColor = palette["text"]
			s.TextLightColor = palette["text_light"]
			s.BackgroundColor = palette["background"]
			s.BackgroundAlt = palette["background_alt"]
		}
	}

	// Load theme CSS
	if themeCSSContent.Valid {
		s.ThemeCSS = themeCSSContent.String
	}

	// Phase I1 (closes the "logo-in-header" gap): resolve the site logo from
	// the current plan's imagery rows joined to the deployed asset, exactly
	// as plan_sections resolves page heroes. The legacy sites.logo_url value
	// (loaded above) remains the fallback for adopted/guide-less sites. The
	// resolved value is the DERIVED committed git path
	// (storage.DeployedWebPath) — never assets.url, which holds an expiring
	// presigned URL. Only override when the asset row exists AND is active,
	// so headers never reference a file the deployer hasn't committed.
	var logoKey, logoPurpose string
	logoErr := db.QueryRowContext(ctx, `
		SELECT a.asset_key, a.purpose
		  FROM site_plan_imagery spi
		  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		  JOIN assets a ON a.site_id = sp.site_id
		               AND a.asset_key = spi.key
		               AND a.status = 'active'
		 WHERE sp.site_id = $1
		   AND spi.scope = 'site'
		   AND spi.kind = 'logo'
		 ORDER BY spi.ordering
		 LIMIT 1
	`, siteID).Scan(&logoKey, &logoPurpose)
	if logoErr == nil && logoKey != "" {
		s.LogoURL = storage.DeployedWebPath(logoKey, logoPurpose)
	} else if logoErr != nil && logoErr != sql.ErrNoRows {
		// Non-fatal: keep the legacy fallback and continue.
		logger.Warn("loadSiteDataFull: plan logo lookup failed — falling back to sites.logo_url",
			zap.Error(logoErr))
	}

	return &s, err
}

// declaresHeadTag reports whether a rendered head already states a given
// property/name, in either quote style and at any attribute order — the same
// attribute-not-full-tag discipline injectCanonicalLink uses, because chrome
// templates are hand-authored and their quoting varies.
func declaresHeadTag(head, attr, value string) bool {
	for _, q := range []string{`"`, `'`} {
		if strings.Contains(head, attr+"="+q+value+q) {
			return true
		}
	}
	return false
}

// injectBrandHeadTags adds favicon + Open Graph / Twitter-card markup to a
// rendered <head>. The favicon points at the derived favicon.png but falls back
// to the site logo when present; og:image points at the derived og-card.png.
// Absolute URLs are built from the site domain for the social tags (OG requires
// absolute), relative for the favicon link.
//
// IDEMPOTENCE IS PER TAG, and that is the whole point of this shape
// (bugs_open/322 item 4). It used to be wholesale: `if the head contains
// rel="icon" OR og:image, return it untouched`. One foreign tag therefore
// disabled the entire block, and the consequences were not theoretical —
//
//   - webdesign.co.uk carries a hand-authored `rel="icon"`, so it received NO
//     og:image, no twitter card and no apple-touch-icon at all, on 117 pages,
//     the most of any site in the fleet. Every caller reported success.
//   - the guard could not see a BLANK tag, so on four sites the head-seo-standard
//     template's empty `og:title`/`og:description` sat alongside a filled pair
//     this function appended — the duplicate tags that made bugs_closed/252
//     visible.
//
// So: a tag the head ALREADY declares is left exactly as authored — this must
// never fight a hand-authored head — and only the missing ones are added. That
// also makes re-running safe on this function's own output, which the wholesale
// guard achieved by accident and only while og:image happened to be present.
//
// ⚠ NO og:url, and no other PAGE-scoped value, ever belongs here — see the
// comment at the og:image line below.
// RETURNS (head, declinedReason). declinedReason is "" on the normal path; when
// non-empty the caller MUST record it durably. That is deliberate plumbing, not
// ceremony: council round 2 (bug_historian, corr 54c660f8) observed that a
// zap.Warn is not a fail-loud signal on this platform — chassis pod logs retain
// on the order of a couple of MINUTES, so a warning nobody writes down is
// write-only. The reason is returned rather than recorded here because this
// function has no DB handle and its one caller does; widening its dependencies
// to reach the error log would be the larger change.
func injectBrandHeadTags(headHTML string, ctx *RenderContext, hasSpriteCSS bool, logger *zap.Logger) (string, string) {
	idx := strings.Index(headHTML, "</head>")
	if idx == -1 {
		// A head with no close tag: decline, but LOUDLY. Unlike this function's
		// siblings (injectCanonicalLink, injectPageJSONLD, spliceOpenGraph)
		// which append, this one declines, because appending brand markup after
		// the head boundary is not obviously right on a hand-authored head.
		//
		// The Warn is the point (council round 1, bug_historian: "'documented'
		// in a comment is not a fail-loud guard"). A silent skip here is how
		// webdesign.co.uk lost every brand tag on 117 pages while every caller
		// reported success — the exact shape this function has just been fixed
		// to stop. So the decline names the site and says what it cost.
		//
		// Currently unexercised: the one fragment head in the fleet was wrapped
		// by migration 529 (bugs_closed/347). If this fires, a new hand-authored
		// head component has arrived without a <head> element, and
		// site-locale-unset-check's finding B will be reporting it too.
		logger.Warn("injectBrandHeadTags: DECLINED — head has no </head> close tag, so no brand tags were added",
			zap.String("domain", ctx.Domain),
			zap.Int("head_bytes", len(headHTML)),
			zap.String("consequence", "this site serves no og:image, twitter card or apple-touch-icon until its head component gains a <head> element (see bugs_closed/347, migration 529)"))
		return headHTML, "head has no </head> close tag, so no brand tags were added"
	}

	origin := "https://" + ctx.Domain
	title := htmlEscapeAttr(defaultString(ctx.CompanyName, ctx.Domain))
	desc := htmlEscapeAttr(ctx.Tagline)

	var b strings.Builder
	// Favicons, and the asymmetry here is deliberate — read it before "tidying".
	//
	// When the head declares NO icon we write TWO: the derived square PNG, and
	// the site logo as a SECONDARY icon when there is one. That is unchanged
	// pre-existing behaviour and it is correct — multiple <link rel="icon"> is
	// valid, browsers pick, and the logo guarantees a mark resolves even before
	// derive_brand_head_assets has committed favicon.png.
	//
	// When the head ALREADY declares any rel="icon" we write NONE. A
	// hand-authored favicon is a deliberate choice, and the rule this enforces
	// is "never append one beside an AUTHORED icon" — not "never emit two".
	// (Council round 1, editquality, read the earlier wording as contradicting
	// itself; it was the comment that was wrong, not the code.)
	if !strings.Contains(headHTML, "rel=\"icon\"") && !strings.Contains(headHTML, "rel='icon'") {
		b.WriteString("  <link rel=\"icon\" href=\"/assets/images/favicon.png\">\n")
		if ctx.LogoURL != "" {
			b.WriteString("  <link rel=\"icon\" href=\"" + ctx.LogoURL + "\">\n")
		}
	}
	if !declaresHeadTag(headHTML, "rel", "apple-touch-icon") {
		b.WriteString("  <link rel=\"apple-touch-icon\" href=\"/assets/images/favicon.png\">\n")
	}
	if !declaresHeadTag(headHTML, "property", "og:type") {
		b.WriteString("  <meta property=\"og:type\" content=\"website\">\n")
	}
	if !declaresHeadTag(headHTML, "property", "og:site_name") {
		b.WriteString("  <meta property=\"og:site_name\" content=\"" + title + "\">\n")
	}
	// og:title and og:description are SITE-level fallbacks only. assemblePage's
	// spliceOpenGraph strips and restates both per page (bugs_closed/252), so
	// what this writes is what a consumer of the stored head alone would see.
	if !declaresHeadTag(headHTML, "property", "og:title") {
		b.WriteString("  <meta property=\"og:title\" content=\"" + title + "\">\n")
	}
	if desc != "" && !declaresHeadTag(headHTML, "property", "og:description") {
		b.WriteString("  <meta property=\"og:description\" content=\"" + desc + "\">\n")
	}
	if !declaresHeadTag(headHTML, "property", "og:image") {
		b.WriteString("  <meta property=\"og:image\" content=\"" + origin + "/assets/images/og-card.png\">\n")
	}
	// NO og:url here, deliberately. This block is written into the PER-SITE
	// stored head (site_components.rendered_html) and reused by every page
	// assemblePage builds, so an origin-rooted og:url made every assembled
	// SUBPAGE claim the homepage's identity: 22 of 24 heads, 700 pages,
	// 26 sites (bugs_open/252, verified on the wire 2026-08-19 beside a
	// canonical that correctly named the page). A per-site artefact cannot
	// carry a per-page value, and this function has no page to ask.
	// Per-page Open Graph identity belongs to assembly: spliceOpenGraph in
	// head_assembly.go, which strips this property and states the page's own.
	if !declaresHeadTag(headHTML, "name", "twitter:card") {
		b.WriteString("  <meta name=\"twitter:card\" content=\"summary_large_image\">\n")
	}
	if !declaresHeadTag(headHTML, "name", "twitter:image") {
		b.WriteString("  <meta name=\"twitter:image\" content=\"" + origin + "/assets/images/og-card.png\">\n")
	}
	// Phase I2: link the site's committed sprite stylesheet (styled bullets,
	// nav accents) only when a sprite sheet exists.
	if hasSpriteCSS && !strings.Contains(headHTML, "/assets/css/sprites.css") {
		b.WriteString("  <link rel=\"stylesheet\" href=\"/assets/css/sprites.css\">\n")
	}

	// Nothing missing: return the head BYTE-IDENTICAL rather than splicing an
	// empty string at the boundary. This is the ordinary steady state once a
	// site's chrome has rendered once, so it must be free and must not churn
	// site_components.rendered_html (the archive trigger fires on a real change).
	if b.Len() == 0 {
		return headHTML, ""
	}

	logger.Info("injectBrandHeadTags: added missing brand head tags",
		zap.String("domain", ctx.Domain),
		zap.Bool("sprite_css", hasSpriteCSS),
		zap.Int("bytes_added", b.Len()))
	return headHTML[:idx] + b.String() + headHTML[idx:], ""
}

// htmlEscapeAttr minimally escapes a string for use inside a double-quoted
// HTML attribute.
func htmlEscapeAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// repointRetiredChromeSlot moves a chrome slot off a component that has been
// RETIRED — `is_active = false` — and onto the one the library says should serve
// its function. It returns the name of the component it moved away from, and
// separately the name of one it wanted to move and could not.
//
// > **NARROWED at council round 1 (corr e242e9d3), and the narrowing is the whole
// > point of the review.** The first version repointed anything that was not
// > ELIGIBLE chrome — deactivated OR a fork OR the wrong component_level — reusing
// > chromeEligibleSQL. Re-running the census at revision time, as the
// > debug_historian seat demanded, found THREE live rows that criterion would have
// > moved and must not: `idea.uk`'s header and footer (active, unforked,
// > `component_level='section'`) and **`leopardessconsulting.co.uk`'s header, which
// > is that site's own ACTIVE FORK.** A fork is illegitimate as a library DEFAULT
// > and entirely legitimate as a deliberate per-site ASSIGNMENT by component_id —
// > it is, in fact, the supported way to give one site its own chrome, which this
// > lane's own 118 close message said out loud six hours before this function was
// > written to undo it. The level question belongs to bugs_open/167 and is
// > fleet-visible, which this deliberately is not.
// >
// > So: eligibility decides what may be CHOSEN as a default; retirement decides
// > what may no longer be SERVED. Only the second licenses touching an existing
// > assignment, and it is exactly what `deactivated_site_components` detects.
//
// It is the missing half of bugs_open/166. `deactivated_site_components` has
// detected this state correctly since 2026-07-17 and routes the repair to
// `rerender-pages`, which re-renders whatever `component_id` already points at —
// so the routed repair could never satisfy its own finding, and the items aged to
// `unresolved` after two strikes. Nothing was wrong with the detection; the
// repair had no way to change the thing being detected.
//
// Three deliberate refusals, each of which would otherwise turn a repair into a
// different bug:
//
//   - It repoints ONLY when an ELIGIBLE alternative exists. Fleet-wide today the
//     `head` function has none (both candidates are is_active=false), so the 13
//     head slots are left exactly as they are rather than being churned on every
//     build for a library gap no site can fix.
//   - It writes through pageComponentAgentWritableSQL, so a human-locked slot is
//     skipped. That predicate is not borrowed from page_components on a hunch:
//     `site_components` carries all four lock columns (locked_at, locked_by,
//     lock_type, lock_expires_at — verified against information_schema at review
//     time), and bugs_closed/069 put them there precisely so one predicate could
//     guard both. The UPDATE at the end of this file already uses it.
//     It does NOT file a lock_blocked_change item — the 069 gate below the
//     idempotence exit owns that decision — but it does REPORT the refusal to its
//     caller, which the first version did not. A locked slot pointing at a retired
//     component otherwise falls through the exit and re-renders the retired
//     component in silence, which is the original bug wearing a lock (bug_historian
//     seat, round 1).
//   - It sets build_status='pending' rather than clearing rendered_html. The slot
//     keeps serving its old chrome until the new render succeeds; blanking it
//     first would leave a window with no chrome at all, and a failed render would
//     make that permanent.
func repointRetiredChromeSlot(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	slot string,
	logger *zap.Logger,
) (movedFrom string, blockedOn string) {
	var currentID, currentName string
	var retired bool
	err := db.QueryRowContext(ctx, `
		SELECT cc.id::text, cc.name, NOT cc.is_active
		FROM site_components sc
		JOIN content_components cc ON cc.id = sc.component_id
		WHERE sc.site_id = $1 AND sc.slot_name = $2
	`, siteID, slot).Scan(&currentID, &currentName, &retired)
	if err != nil || !retired {
		// No assignment, no such component, or the assignment is still in
		// service. An unassigned slot is the fallback path's job, not this
		// one's; a FORK or a section-level assignment is a deliberate choice
		// and is not this function's business (see the header).
		return "", ""
	}

	want, wantEligible, resolveErr := ResolveChromeComponent(ctx, db, ChromeSlotFunction(slot), logger)
	if resolveErr != nil || !wantEligible || want.ID == currentID {
		// ResolveChromeComponent has already logged the library gap at ERROR.
		// Leaving the slot alone is the correct outcome here: there is nothing
		// better to point it at. This is every `head` slot in the fleet today.
		return "", ""
	}

	res, execErr := db.ExecContext(ctx, `
		UPDATE site_components
		SET component_id = $3, build_status = 'pending', updated_at = now()
		WHERE site_id = $1 AND slot_name = $2 AND `+pageComponentAgentWritableSQL("")+`
	`, siteID, slot, want.ID)
	if execErr != nil {
		logger.Warn("site chrome: failed to repoint a retired slot (bugs_open/166)",
			zap.String("slot", slot), zap.String("from", currentName), zap.Error(execErr))
		return "", currentName
	}
	if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
		// Locked, or the row vanished. No work item here — the 069 gate owns
		// that — but this MUST reach the caller: the slot goes on serving a
		// retired component and the idempotence exit below will wave it through
		// as "already rendered", so a silent return here reproduces the very bug
		// this function exists to fix, for the locked case.
		logger.Warn("site chrome: retired slot NOT repointed — locked or absent (bugs_open/166)",
			zap.String("slot", slot), zap.String("from", currentName), zap.String("wanted", want.Name))
		return "", currentName
	}

	logger.Warn("site chrome: repointed a slot off a RETIRED component (bugs_open/166)",
		zap.String("slot", slot),
		zap.String("from", currentName),
		zap.String("to", want.Name),
		zap.String("reason", "assigned component is is_active=false"),
	)
	return currentName, ""
}

func renderAndStoreSiteComponent(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	slot string,
	renderCtx *RenderContext,
	force bool,
	chromeLinks ChromeLinkPolicy,
	// recordAbsentRequired arms the bugs_open/342 render-time record. Passed in
	// rather than read here because this function has no step config, and read
	// from config at the ONE caller so the arming decision stays visible next to
	// the other step settings rather than buried at the write.
	recordAbsentRequired bool,
	// refuseAbsentRequired arms the REFUSAL: the slot is not stored and the
	// previously stored bytes keep serving. Separate from recordAbsentRequired
	// because filing a note and declining to write are different authorities —
	// and it exists at ALL because the council's bug_historian seat was right
	// (council 3626629a round 1, medium): arming DETECTION here while only the
	// section editor got PROTECTION would reproduce bugs_open/342's own shape
	// on the sibling call site, which is 016b §9's "one call site of a shared
	// judgement gets the rigorous fix, the sibling stays heuristic".
	//
	// ⚠ NO MIGRATION ARMS THIS, deliberately, and that is not an oversight.
	// The measured population is ZERO — every component the chrome store
	// references declares no required source:"llm" field (2026-08-22) — so
	// arming it today would arm an unexercisable refusal, while leaving the
	// CAPABILITY absent would mean the first chrome component adopted with a
	// required field needs a code change, a review and a roll before it can be
	// protected. This way it needs a config flip. The trigger to flip it is
	// the record half firing: the first required_fields_missing item with
	// surface="site_component" is the day this becomes exercisable.
	refuseAbsentRequired bool,
	logger *zap.Logger,
) (ok bool, locked bool, degraded string, fatal error) {
	// degraded names the component this slot fell back to when the library held
	// no ELIGIBLE chrome for the slot's function (bugs_open/118). Empty on every
	// healthy path, including the ones that do no work at all.
	//
	// fatal is NON-NIL for exactly one condition (bugs_open/260): the slot's
	// template failed to EXECUTE and this site has no stored chrome for the slot
	// to fall back on — i.e. a first build that would otherwise go live with a
	// missing header, footer or head. Every other failure keeps its existing
	// report-and-continue behaviour, because every other failure leaves the
	// previously stored bytes serving.

	// Retire a RETIRED assignment before the idempotence exit below can
	// declare the slot finished (bugs_open/166). This has to run first: the exit
	// asks "does this slot hold HTML?", never "does it hold the RIGHT component's
	// HTML?", so a slot pointing at a deactivated component reads as done for
	// ever — which is why the platform's own deactivated_component items have sat
	// unrepaired since 2026-07-17 while their handler faithfully re-rendered the
	// deactivated component.
	//
	// It repoints ONLY when the library offers an eligible alternative, so the
	// `head` slots (no active head component exists fleet-wide) are untouched
	// rather than being re-rendered pointlessly on every build. It touches only
	// RETIRED assignments — never a fork, never a level mismatch; see its header.
	repointed, repointBlocked := repointRetiredChromeSlot(ctx, db, siteID, slot, logger)
	if repointBlocked != "" {
		// Surfaced in the action result as ineligible_chrome, beside the library
		// gap it shares a field with: in both cases this slot is serving chrome
		// the platform would not choose, and in both cases no site can fix it.
		degraded = repointBlocked
	}

	// Same position, same reason, one bug along: the exit below asks whether the
	// slot holds HTML, never whether it holds html that still satisfies the LINK
	// policy, so a header already serving a dead CTA reads as done for ever
	// (bugs_open/191). Marks `build_status = 'pending'` — the signal the exit
	// already honours for the repoint above — under the same lock guard.
	markStaleChromeLinkSlot(ctx, db, siteID, slot, chromeLinks, logger)

	// Check if already rendered (unless force)
	if !force {
		var exists bool
		// build_status is part of the question, not decoration: a slot that has
		// just been repointed still holds the OLD component's HTML, so "has
		// HTML" is true and "is up to date" is false. Without this clause a
		// repoint — by the line above, by an operator, or by any future writer —
		// silently never reaches the page.
		db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM site_components
				WHERE site_id = $1 AND slot_name = $2
				AND rendered_html IS NOT NULL AND rendered_html != ''
				AND COALESCE(build_status, '') <> 'pending'
			)
		`, siteID, slot).Scan(&exists)

		if exists {
			logger.Debug("Site component already rendered, skipping",
				zap.String("slot", slot))
			return true, false, "", nil
		}
	}

	// Human lock gate (bugs_open/069). DELIBERATELY below the !force check
	// above: that path already no-ops on a populated slot, so checking earlier
	// would file a "a writer wanted to change this" item for a call that was
	// never going to write. Everything after this point writes — the fallback
	// below repoints component_id at a GENERIC default template, and the render
	// at the end replaces rendered_html outright — so one bail here covers both.
	//
	// A check failure is non-fatal: the guarded statements below carry the lock
	// predicate themselves, so enforcement does not depend on this read.
	if lock, lockErr := CheckSiteComponentLock(ctx, db, siteID, slot, logger); lockErr != nil {
		logger.Warn("renderAndStoreSiteComponent: chrome lock check failed — relying on the guarded writes",
			zap.String("slot", slot), zap.Error(lockErr))
	} else if lock.IsLocked {
		logger.Warn("renderAndStoreSiteComponent: refusing to re-render human-locked chrome slot (bugs_open/069)",
			zap.String("slot", slot),
			zap.String("locked_by", lock.LockedBy),
			zap.String("lock_type", lock.LockType),
			zap.Bool("artefact_empty", !lock.HasHTML),
		)
		emitChromeLockBlockedChangeItem(ctx, db, siteID, slot, lock,
			"overwrite", "render_site_components", logger)
		// ok reports whether the slot still SERVES chrome, which is what the
		// caller's map means. A lock over a populated slot is a success from
		// the page's point of view; a lock over an empty one is not.
		return lock.HasHTML, true, "", nil
	}

	// Get component template + its declared input_schema. The schema is what
	// makes render_site_components honour a component's own field vocabulary
	// rather than only the hardcoded ContentData map below — bugs_open/018.
	var componentID uuid.UUID
	var htmlTemplate string
	var inputSchemaRaw []byte

	err := db.QueryRowContext(ctx, `
		SELECT sc.component_id, cc.html_template, COALESCE(cc.input_schema, '{}'::jsonb)
		FROM site_components sc
		JOIN content_components cc ON sc.component_id = cc.id
		WHERE sc.site_id = $1 AND sc.slot_name = $2
	`, siteID, slot).Scan(&componentID, &htmlTemplate, &inputSchemaRaw)

	if err != nil {
		// No component assigned — resolve the library default through the ONE
		// chrome predicate (bugs_open/118). What used to be here was a bare
		// `WHERE function = $1 ORDER BY name LIMIT 1`: no is_active, no fork
		// exclusion, no level filter. It is what assigned 11 of 14 sites a
		// DEACTIVATED footer and then pinned the choice in site_components, where
		// it has outlived every attempt to edit the active component instead.
		funcName := ChromeSlotFunction(slot)

		comp, eligible, resolveErr := ResolveChromeComponent(ctx, db, funcName, logger)
		if resolveErr != nil {
			logger.Warn("No component found for slot",
				zap.String("slot", slot),
				zap.String("function", funcName),
				zap.Error(resolveErr))
			return false, false, "", nil
		}

		parsedID, parseErr := uuid.Parse(comp.ID)
		if parseErr != nil {
			logger.Warn("Resolved chrome component has an unparseable id",
				zap.String("slot", slot),
				zap.String("component_id", comp.ID),
				zap.Error(parseErr))
			return false, false, "", nil
		}
		componentID = parsedID
		htmlTemplate = comp.HTMLTemplate
		inputSchemaRaw = nil
		if len(comp.InputSchema) > 0 {
			if encoded, mErr := json.Marshal(comp.InputSchema); mErr == nil {
				inputSchemaRaw = encoded
			}
		}

		// ResolveChromeComponent has already logged the detail at ERROR. Naming
		// it here too is what carries it out of the log and into the action's
		// result, where an operator reading the orchestration reads it.
		if !eligible {
			degraded = comp.Name
		}

		// Insert the component assignment. The DO UPDATE arm repoints an
		// EXISTING row at this generic default, so it carries the lock
		// predicate (bugs_open/069) — qualified, because bare column names in
		// an ON CONFLICT ... WHERE are ambiguous against EXCLUDED. The INSERT
		// arm needs no gate: a row that does not exist yet cannot be locked.
		// The error was previously discarded outright, which made a failed
		// assignment look like a successful render.
		if _, execErr := db.ExecContext(ctx, `
			INSERT INTO site_components (site_id, slot_name, component_id, build_status)
			VALUES ($1, $2, $3, 'pending')
			ON CONFLICT (site_id, slot_name) DO UPDATE SET component_id = $3
			WHERE `+pageComponentAgentWritableSQL("site_components.")+`
		`, siteID, slot, componentID); execErr != nil {
			logger.Warn("Failed to assign default component to slot",
				zap.String("slot", slot), zap.Error(execErr))
		}
	}

	// Schema-driven fill (bugs_open/018). renderCtx.ContentData above is a FIXED
	// vocabulary; a chrome component declaring any other field name previously
	// received "" for it, silently — idea.uk rendered 30 dead href="" controls
	// because its templates ask for nav_link_1_url etc. and the map has none of
	// them. Resolve the component's OWN declared schema fields through the
	// existing sourceResolver, GAP-FILLING only, so wherever the map already
	// supplies a value it stays authoritative and the nine sites on
	// header-bold-gradient/footer-4-column render byte-identically under every
	// caller. `unresolved` is at function scope so the report below (outside the
	// schema block) can read it.
	var unresolved []string
	// parsedInputSchema is at function scope for the same reason `unresolved`
	// is: the render call below needs it (bugs_open/342). ⚠ Coverage note, so a
	// silent result is not read as a clean one — missingRequiredLLMFields reads
	// the v2 `fields` and legacy `properties` dialects, so the FLAT chrome shape
	// (`{name:{...}}`, e.g. "Document Head") yields no field set and therefore no
	// absent-required report. Fail-open, stated.
	var parsedInputSchema map[string]interface{}
	if len(inputSchemaRaw) > 0 {
		var raw map[string]interface{}
		if jErr := json.Unmarshal(inputSchemaRaw, &raw); jErr != nil {
			logger.Warn("site chrome: input_schema unparseable; fixed vocabulary only",
				zap.String("slot", slot), zap.Error(jErr))
		} else {
			// TWO live schema shapes: wrapped {"fields":{name:{...}}} (most
			// components) and FLAT {name:{...}} (e.g. "Document Head", the head
			// slot on most sites). Detect both; a shape that is neither is skipped.
			parsedInputSchema = raw
			fields := raw
			if w, ok := raw["fields"].(map[string]interface{}); ok {
				fields = w
			}
			resolver := newSourceResolver(siteID, db, logger, "") // site-wide: no page
			for name, defAny := range fields {
				def, ok := defAny.(map[string]interface{})
				if !ok {
					continue // not a field descriptor (e.g. a stray scalar)
				}
				// GAP-FILL ONLY. Presence wins for EVERY non-string type, so a
				// set bool like show_subscribe is never overwritten; only a
				// genuinely empty string is filled.
				if existing, present := renderCtx.ContentData[name]; present {
					if s, isStr := existing.(string); !isStr || s != "" {
						continue
					}
				}
				source, _ := def["source"].(string)
				if source == "" || source == "static" || strings.HasPrefix(source, "static.") {
					// Declared static literal (e.g. nav_aria_label): honour its
					// fallback. A data-source MISS below never gets a fallback.
					if fb := def["fallback"]; fb != nil {
						renderCtx.ContentData[name] = fb
					} else {
						unresolved = append(unresolved, name)
					}
					continue
				}
				if source == "llm" || source == "renderer" || strings.HasPrefix(source, "renderer.") {
					unresolved = append(unresolved, name)
					continue
				}
				if v, found := resolver.resolve(ctx, source); found && v != nil {
					// Declared-type guard (council 56ab6e23, bug_historian
					// advisory). A non-array value reaching a {{range}} errors
					// the WHOLE template into the silent regex-fallback
					// renderer — one bad config key degrades the entire slot.
					// Refusing the fill instead means the gated block renders
					// ABSENT and the rest of the chrome renders normally.
					// Enforced only for array/list: measured 2026-08-02,
					// every array/list-declared field fleet-wide is either
					// {{range}}-consumed (53) or unreferenced (16), zero are
					// bare-output — so this can only ever un-degrade a render,
					// never break a working one.
					if declared, _ := def["type"].(string); !datahelpers.DeclaredTypeSatisfied(declared, v) {
						logger.Warn("site chrome: resolved value does not satisfy the field's declared type — refusing the fill (gated template renders without it)",
							zap.String("slot", slot),
							zap.String("field", name),
							zap.String("declared_type", declared),
							zap.String("actual_type", fmt.Sprintf("%T", v)),
							zap.String("component_id", componentID.String()),
							zap.String("site_id", siteID.String()))
						unresolved = append(unresolved, name)
						continue
					}
					renderCtx.ContentData[name] = v
					continue
				}
				// MISSED data-source resolution. Do NOT apply def["fallback"] —
				// header-bold-gradient's cta_url fallback is /contact.html, the
				// fossil of the phantom-CTA bug LNK-007. Correct-or-absent (LNK-005).
				unresolved = append(unresolved, name)
			}
		}
	}

	// Render the template, reporting which placeholders rendered empty so a dead
	// chrome control is named (Error, via the mechanism) and combined here with
	// the site/slot/component only this caller knows.
	// bugs_open/342 — the schema this function already parsed for the fill.
	renderCtx.InputSchema = parsedInputSchema
	renderedHTML, missing, deadURLFields, renderErr := RenderTemplate(htmlTemplate, renderCtx, logger)

	// bugs_open/260: this path has NO gate downstream — whatever it stores is
	// what the site serves — so a failed render must not be stored. Doing
	// nothing leaves the existing row serving, which is strictly better than
	// replacing working chrome with chrome carrying unexecuted {{if}}/{{range}}
	// directives. The one case that is not better is a site with no stored
	// chrome for this slot at all: there the alternative to storing is NOTHING,
	// and a site must not go live with a missing header, footer or head — so
	// that case, and only that case, escalates to the caller.
	//
	// Note the `renderedHTML == ""` guard below could never have caught this:
	// the deleted fallback returned MANGLED html, never empty.
	if renderErr != nil {
		serving := chromeSlotHasStoredHTML(ctx, db, siteID, slot)
		logger.Error("site chrome: template execution failed — not storing",
			zap.String("slot", slot),
			zap.String("component_id", componentID.String()),
			zap.String("site_id", siteID.String()),
			zap.Bool("existing_row_keeps_serving", serving),
			zap.Error(renderErr),
		)
		// File the human record whichever branch the caller takes: a chrome
		// template that cannot execute is not something a re-render fixes, and
		// a signal that exists only inside a failed step is a signal nobody
		// reads (the same reasoning as the dead-control emit above).
		emitChromeRenderFailedItem(ctx, db, siteID, componentID, slot, renderErr, serving, logger)
		return false, false, degraded, fmt.Errorf("chrome slot %q failed to render: %w", slot, renderErr)
	}

	if len(unresolved) > 0 || len(missing) > 0 {
		logger.Warn("site chrome: fields rendered empty — template must gate these",
			zap.String("slot", slot),
			zap.String("component_id", componentID.String()),
			zap.String("site_id", siteID.String()),
			zap.Strings("schema_unresolved", unresolved),
			zap.Strings("template_missing", missing),
			zap.Strings("dead_url_fields", deadURLFields),
		)
	}

	// bugs_open/054: the observability half (018) named a dead chrome URL control
	// loudly but still SHIPPED it — idea.uk shipped 30 empty-href nav links this
	// way, and the council held that "a named log is not escalation". Owner ruling
	// 2026-07-22: make it MEAN something. (1) DROP the dead control from the
	// rendered chrome so it never reaches a live page (LNK-005 correct-or-absent),
	// and (2) FILE it to the human-review queue (owner ruled 033 IS a worked queue)
	// with NO handler, mirroring the sibling check_dead_controls — a dropped
	// control is a human decision, not something an auto-re-render can invent.
	// Gated on deadURLFields, so a clean render is never touched and its
	// byte-identical output is preserved. data-runtime-fill shells hydrate their
	// own hrefs client-side, so an empty URL attribute there is intentional, not a
	// dead control — exempt them exactly as check_dead_controls does
	// (render_guardian council note, 2026-07-22).
	// bugs_open/342 — ESCALATE, don't just log (council bb7f5d0e round 1,
	// bug_historian, GATING: "a named log is not escalation", owner ruling on
	// bugs_open/054). The seam has published which schema-required fields
	// rendered EMPTY; this function has the database handle and the site
	// identity it lacks, so this is where the report becomes a queue entry.
	//
	// REUSES required_fields_missing — the item type that already exists for
	// this exact defect, with a router seeded by bugs_open/277 — rather than a
	// fourth chrome-specific type. ⚠ AND IT REACHES A POPULATION THE EXISTING
	// PRODUCER CANNOT: check_required_fields_missing scans
	// `pc.build_status = 'deployed'`, i.e. rows that made it. 342's damage is
	// the section that renders empty and is DROPPED, so it never becomes a
	// deployed row to be scanned — survivorship, the same shape as bugs_closed/
	// 260's "no live damage" headline. The item_key matches the check's
	// (required_fields_missing:<page>:<slot>) so the two producers co-dedup
	// instead of racing (owner ruling 2026-08-02 §1: name the producer set and
	// the key shape where the type is registered).
	//
	// OPT-IN, unsafe default OFF: this is a new DB write on a shared render
	// path, which is exactly what three seats made the dead-URL RECORD arm
	// justify (council 98852baa). Unset means today's behaviour, byte for byte.
	if recordAbsentRequired && len(renderCtx.AbsentRequiredFields) > 0 {
		compID := componentID
		// pageContext{} — EMPTY, and deliberately: a chrome slot hangs off the
		// SITE, not a page, so there is no page_name for the router's classify
		// step to resolve. The emitter reads that as "no page" and files the
		// item for human review rather than handing it to a page-resolving
		// router that would classify it `malformed` and burn three attempts —
		// which is exactly what happened on the editor route before
		// 2026-08-23 (bugs_open/342, item a31da7f3).
		emitRequiredFieldsMissing(ctx, db, siteID, pageContext{},
			&compID, slot,
			fmt.Sprintf("Chrome %s", slot), "site_component", "render_site_components",
			renderCtx.AbsentRequiredFields,
			map[string]interface{}{"slot_name": slot, "component_id": componentID.String()}, logger)
	}

	// The REFUSAL, when armed (see the parameter's own comment for why it
	// exists unarmed). Placed AFTER the emit so a refused slot still leaves its
	// queue entry, exactly as the section-editor gate does — refusing must
	// never be the reason a defect goes unrecorded.
	//
	// ⚠ NEITHER THIS NOR THE RECORD ABOVE IS REACHED ON THE `!force`
	// IDEMPOTENT-SKIP PATH, and that asymmetry is worth stating rather than
	// discovering (council 3626629a round 2, guardian, medium — the seat asked
	// where this sits relative to that exit, and the honest answer is "after
	// it, so it never runs when the exit fires"). For the REFUSAL that is
	// correct and not a gap: the exit fires precisely when the slot already
	// holds non-empty, non-pending HTML and nothing is about to be written, and
	// a gate that exists to prevent a write has nothing to prevent. For
	// DETECTION it is a real blind spot, and a pre-existing one: an already
	// populated slot whose content_data is missing a required field is never
	// inspected by a non-forced refresh, because no render happens to inspect.
	// The fleet's chrome rerenders pass force=true, so the common path does
	// reach here — but do not read this call site as "every chrome slot is
	// checked continuously". What is checked is every slot this function is
	// about to write.
	//
	// The fatal-vs-degraded split reuses chromeSlotHasStoredHTML, which the
	// execution-failure branch ~90 lines above already uses for exactly this
	// decision (and the caller uses at :333) — deliberately NOT a new
	// discriminator (same council round, bug_historian, medium: this estate has
	// twice had to re-harden a "does this hold real content?" judgement). Note
	// it is not that class of judgement at all: it asks whether any non-empty
	// rendered_html is STORED for the slot, never whether that HTML looks
	// meaningful, so there is no sparse-but-real content for it to misread.
	//
	// It declines to STORE, which on this path means the previously stored
	// bytes keep serving: the same disposition the execution-failure branch
	// above already chose, for the same reason (whatever this function stores
	// is what the site serves, and there is no gate downstream). The one case
	// where not storing is worse than storing is a slot with nothing stored at
	// all — a site must not go live with a missing header — so that case
	// escalates to the caller as fatal, mirroring the branch above rather than
	// inventing a second disposition for the same predicament.
	if refusePersistForAbsentRequired(
		map[string]interface{}{absentRequiredRefuseConfigKey: refuseAbsentRequired},
		renderCtx.AbsentRequiredFields,
	) {
		serving := chromeSlotHasStoredHTML(ctx, db, siteID, slot)
		logger.Error("site chrome: REQUIRED content field(s) absent — refusing to store (bugs_open/342)",
			zap.String("slot", slot),
			zap.String("component_id", componentID.String()),
			zap.String("site_id", siteID.String()),
			zap.Strings("absent_required_fields", renderCtx.AbsentRequiredFields),
			zap.Bool("existing_row_keeps_serving", serving),
		)
		if !serving {
			return false, false, degraded, fmt.Errorf(
				"site chrome %q: refusing to store — %d schema-required field(s) rendered empty (%s), and this site has no stored %s to fall back on (bugs_open/342)",
				slot, len(renderCtx.AbsentRequiredFields),
				strings.Join(renderCtx.AbsentRequiredFields, ", "), slot)
		}
		return false, false, degraded, nil
	}

	if len(deadURLFields) > 0 && !strings.Contains(renderedHTML, "data-runtime-fill") {
		beforeLen := len(renderedHTML)
		renderedHTML = DropDeadURLControls(renderedHTML)
		logger.Warn("site chrome: dropped dead URL control(s) before store — bugs_open/054",
			zap.String("slot", slot),
			zap.String("component_id", componentID.String()),
			zap.String("site_id", siteID.String()),
			zap.Strings("dead_url_fields", deadURLFields),
			zap.Int("bytes_removed", beforeLen-len(renderedHTML)),
		)
		emitChromeDeadControlItem(ctx, db, siteID, componentID, slot, deadURLFields, logger)
	}

	if renderedHTML == "" {
		logger.Warn("Template rendered to empty string",
			zap.String("slot", slot))
		return false, false, degraded, nil
	}

	// Phase I1 (G8): inject favicon + Open Graph tags into <head>. Head
	// templates predate brand-head assets and carry no favicon/og markup, so
	// we add it deterministically at render time — fleet-wide, no per-site
	// template regeneration. References the derived files
	// (/assets/images/favicon.png, /assets/images/og-card.png), which the
	// derive_brand_head_assets action commits; harmless if they 404 until
	// derivation runs, and the favicon degrades to the site logo.
	if slot == "head" {
		// Phase I2: only link sprites.css when the site actually has an active
		// sprite-sheet asset — otherwise the <link> would 404 on sites without
		// one.
		var spriteCount int
		_ = db.QueryRowContext(ctx, `
			SELECT count(*) FROM assets
			 WHERE site_id = $1 AND purpose = 'sprite_sheet' AND status = 'active'
		`, siteID).Scan(&spriteCount)
		var brandDeclined string
		renderedHTML, brandDeclined = injectBrandHeadTags(renderedHTML, renderCtx, spriteCount > 0, logger)
		if brandDeclined != "" {
			// Durable, because a log line here is not (council round 2,
			// bug_historian): chassis log retention is ~minutes, so the row is
			// the only thing a later triage can find. Best-effort by design —
			// agenterrors.Write never changes the disposition already decided,
			// and this slot still stores its rendered HTML either way.
			agenterrors.Write(ctx, db, logger, agenterrors.Entry{
				SiteID:       siteID.String(),
				Domain:       renderCtx.Domain,
				AgentType:    "render_site_components",
				StepName:     "inject_brand_head_tags",
				Action:       "render_site_components",
				ErrorCode:    "BRAND_HEAD_TAGS_DECLINED",
				Severity:     "warning",
				ErrorMessage: "brand head tags not injected: " + brandDeclined,
				Context: map[string]interface{}{
					"slot":        slot,
					"head_bytes":  len(renderedHTML),
					"consequence": "site serves no og:image, twitter card or apple-touch-icon until its head component gains a <head> element",
					"remedy":      "wrap the head component's template in <head>…</head> — bugs_closed/347, migration 529 is the worked example",
				},
			})
		}
	}

	// Divergence gate, read half (bugs_open/226). The artefact about to be
	// replaced may hold bytes the pipeline did not put there — a psql patch,
	// an admin edit, a chrome-fix artefact patch — and this store would
	// silently destroy them. The RECOVERY is not here: trg_site_component_archive
	// (a DB trigger, sql_for_agents/344 — invisible to any grep of Go, which
	// is why this comment exists) archives the outgoing bytes atomically with
	// every differing overwrite, from every writer. This read is the LOUD half
	// only, so a race with a concurrent writer costs at most a mislabelled log
	// line, never the artefact.
	//
	// The read must run BEFORE the store (it classifies the outgoing bytes),
	// but the WARN + work item fire only AFTER the store reports rows>0 —
	// a lock refusal between here and there destroys nothing, so filing
	// "bytes were overwritten" on that path would be a false record (council
	// round 1, render_guardian seat).
	divergence, divErr := classifySiteComponentArtefact(ctx, db, siteID, slot)
	if divErr != nil {
		logger.Warn("site chrome: divergence classification failed — overwrite proceeds, the 344 trigger still archives (bugs_open/226)",
			zap.String("slot", slot), zap.Error(divErr))
		divergence = nil
	}

	// Store the rendered HTML. The lock predicate makes the refusal race-free
	// (bugs_open/069) — the pre-check above is for the message and to save the
	// work, this is the enforcement.
	//
	// render_inputs is the render-provenance stamp (bugs_open/117): the shared
	// fingerprint expression digests every store this chrome renders from, in
	// the SAME statement that persists the artefact, so the stamp can never
	// describe a different render than the bytes beside it. The
	// stale_site_components discovery check recomputes the SAME expression and
	// fires when the two differ — which is what finally couples "the thing
	// chrome is built from changed" to "a chrome rebuild runs". The row alias
	// is load-bearing: the fingerprint correlates on `sc` (see the fragment's
	// header in datahelpers/chrome_render_inputs.go).
	//
	// rendered_html_digest is the ARTEFACT stamp (bugs_open/226), in the same
	// statement for the same reason: digest = md5(bytes) thereafter means
	// "these are exactly the bytes this path wrote", and a mismatch at the
	// next overwrite is how a hand patch is told from a machine render. Only
	// this path may write the digest — a stamp beside patched bytes silences
	// the detector.
	res, err := db.ExecContext(ctx, `
		UPDATE site_components AS sc
		SET rendered_html = $1, build_status = 'rendered', updated_at = now(),
		    rendered_html_digest = md5($1),
		    render_inputs = (`+datahelpers.ChromeRenderInputsSQL+`)
		WHERE sc.site_id = $2 AND sc.slot_name = $3 AND `+pageComponentAgentWritableSQL("sc.")+`
	`, renderedHTML, siteID, slot)

	if err != nil {
		logger.Error("Failed to store rendered component",
			zap.String("slot", slot),
			zap.Error(err))
		return false, false, degraded, nil
	}

	// Zero rows means the row is locked or gone. Before this gate the result
	// was discarded entirely, so a store that wrote nothing still logged
	// "rendered and stored" and reported success — the two cases must not be
	// folded together now that one of them is legitimate.
	if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
		lock, lockErr := CheckSiteComponentLock(ctx, db, siteID, slot, logger)
		if lockErr == nil && lock.IsLocked {
			logger.Warn("Chrome slot locked between check and store — render discarded (bugs_open/069)",
				zap.String("slot", slot), zap.String("locked_by", lock.LockedBy))
			emitChromeLockBlockedChangeItem(ctx, db, siteID, slot, lock,
				"overwrite", "render_site_components", logger)
			return lock.HasHTML, true, degraded, nil
		}
		logger.Error("Failed to store rendered component: no row matched",
			zap.String("slot", slot), zap.String("site_id", siteID.String()))
		return false, false, degraded, nil
	}

	// Divergence gate, loud half (bugs_open/226): the store wrote (rows>0), so
	// the trigger has archived the outgoing bytes — if they were hand-patched,
	// say so where an operator will see it.
	//
	// If the pre-store classification failed, the trigger's own verdict is on
	// the archive row it just wrote — read it back rather than letting a
	// transient SELECT error be the one silent path (council round 2).
	if divergence == nil {
		lv, lErr := readBackDivergenceFromLedger(ctx, db, siteID, slot)
		if lErr != nil {
			logger.Error("site chrome: divergence verdict UNKNOWN for this overwrite — classification and ledger read-back both failed; the archive itself is intact (bugs_open/226)",
				zap.String("slot", slot), zap.String("site_id", siteID.String()), zap.Error(lErr))
		} else {
			divergence = lv
		}
	}
	if divergence != nil && divergence.State == artefactHandPatched {
		logger.Warn("site chrome: artefact diverged from its render stamp — hand-patched bytes were overwritten and archived (bugs_open/226)",
			zap.String("slot", slot),
			zap.String("site_id", siteID.String()),
			zap.String("stamped_digest", divergence.StampedDigest),
			zap.String("current_digest", divergence.CurrentDigest),
			zap.Int("artefact_bytes", divergence.Bytes),
		)
		emitChromeDivergenceItem(ctx, db, siteID, slot, divergence, "render_site_components", logger)
	}

	logger.Info("Site component rendered and stored",
		zap.String("slot", slot),
		zap.Int("html_length", len(renderedHTML)),
		zap.String("repointed_from", repointed))

	return true, false, degraded, nil
}

// emitChromeDeadControlItem files ONE work item when a site-chrome component
// renders a dead URL control that DropDeadURLControls has just dropped
// (bugs_open/054). Routing MIRRORS the sibling check_dead_controls
// (page_components): status 'needs_human_review' with NO handler_agent. A dropped
// chrome control is a human decision — nothing can auto-invent a destination, and
// an auto-re-render handler would mark the item complete without verifying the
// field resolved (council REVISE, 2026-07-22, owner-confirmed). The owner ruled
// 033 IS a worked queue, so needs_human_review is the visible queue the dashboard
// surfaces (v1.0.1141), not the pre-033 void. The common data-lag case (idea.uk's
// nav fields resolved after a data fix) self-heals on the next normal chrome
// re-render regardless.
//
// Persistence goes through the shared insertWorkItem helper (the exact path
// check_dead_controls uses) rather than a hand-rolled INSERT, so it inherits the
// correct idx_swi_dedup-matched ON CONFLICT (keyed on the shared
// workItemTerminalStatuses list, no drift) and the two-strike anti-churn label.
// A failure is logged, never returned — the signal must not block the render, and
// DropDeadURLControls has already made the page safe. deadFields arrives
// pre-sorted from missingBareFields.
func emitChromeDeadControlItem(ctx context.Context, db *sql.DB, siteID, componentID uuid.UUID,
	slot string, deadFields []string, logger *zap.Logger) {

	fieldList := strings.Join(deadFields, ", ")

	spec := map[string]interface{}{
		"surface":         "site_component",
		"slot_name":       slot,
		"component_id":    componentID.String(),
		"dead_url_fields": deadFields,
		"source":          "render_site_components",
		"fix": "A site-chrome control (nav link / header CTA / logo) had no " +
			"resolvable destination and was DROPPED from the rendered page so it " +
			"would not ship as a dead href=\"\" control (bugs_open/054). Decide: " +
			"wire a real destination, build the missing target, or accept the drop. " +
			"A re-render of the site chrome restores it automatically if the " +
			"destination now exists. Never point it at /contact.html (the LNK-007 " +
			"fossil).",
	}
	specJSON, _ := json.Marshal(spec)

	summary := fmt.Sprintf("Dead chrome control on %s slot: no destination for %s (dropped)", slot, fieldList)
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}
	itemKey := fmt.Sprintf("chrome_dead_control:%s:%s:%s", siteID, slot, fieldList)

	compID := componentID
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("render_site_components: begin tx for chrome_dead_control failed",
			zap.String("slot", slot), zap.Error(err))
		return
	}
	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:      siteID,
		componentID: &compID,
		source:      "render-site-components",
		pipeline:    "build",
		itemType:    "chrome_dead_control",
		severity:    "high",
		summary:     summary,
		spec:        string(specJSON),
		priority:    40,
		status:      "needs_human_review",
		createdBy:   "render_site_components",
		itemKey:     itemKey,
	}, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("render_site_components: insert chrome_dead_control item failed",
			zap.String("slot", slot), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("render_site_components: commit chrome_dead_control item failed",
			zap.String("slot", slot), zap.Error(err))
		return
	}
	if inserted {
		logger.Info("render_site_components: chrome_dead_control item filed for review",
			zap.String("slot", slot), zap.Strings("dead_url_fields", deadFields))
	}
}

// loadNavItems loads navigation items for header
// DEPRECATED
func loadNavItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxItems int, logger *zap.Logger) []NavItem {
	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages 
		WHERE site_id = $1 
		  AND (in_header = true OR in_header IS NULL)
		  AND status IN ('deployed', 'active')
		  AND name NOT IN ('privacy', 'terms', 'cookies', '404', 'sitemap')
		ORDER BY 
			COALESCE(nav_order, 99),
			CASE name 
				WHEN 'index' THEN 1 
				WHEN 'home' THEN 1
				WHEN 'services' THEN 2
				WHEN 'about' THEN 3
				WHEN 'contact' THEN 10
				ELSE 5 
			END
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, siteID, maxItems)
	if err != nil {
		logger.Warn("loadNavItems: Query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			continue
		}
		// Simplify label
		label = strings.Title(strings.ReplaceAll(label, "-", " "))
		items = append(items, NavItem{Label: label, URL: url})
	}

	return items
}

// loadFooterNavItems loads navigation items for footer
// DEPRECATED
func loadFooterNavItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxItems int, logger *zap.Logger) []NavItem {
	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages 
		WHERE site_id = $1 
		  AND (in_footer = true OR in_footer IS NULL)
		  AND status IN ('deployed', 'active')
		  AND name NOT IN ('index', '404', 'sitemap')
		ORDER BY 
			COALESCE(nav_order, 99),
			CASE name 
				WHEN 'services' THEN 1
				WHEN 'about' THEN 2
				WHEN 'contact' THEN 3
				WHEN 'privacy' THEN 8
				WHEN 'terms' THEN 9
				ELSE 5 
			END
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, siteID, maxItems)
	if err != nil {
		logger.Warn("loadFooterNavItems: Query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			continue
		}
		label = strings.Title(strings.ReplaceAll(label, "-", " "))
		items = append(items, NavItem{Label: label, URL: url})
	}

	return items
}

// buildServicesHTML queries service-related pages and builds <li> HTML for the footer services column.
// Looks for pages that represent individual service offerings (excludes structural pages).
// Falls back to an empty string if no service pages found.
// buildServicesHTML builds the footer's "Our Services" column.
//
// It queries `pages` DIRECTLY rather than going through GetNavItems, which is
// why the NavVisibility change did not cover it — and it emits <a href> into
// chrome exactly like the nav does. Found the hard way (bugs_open/049): after the
// nav fix landed and gaswholesalers' chrome was re-rendered, its footer STILL
// carried /fuel-pricing-framework.html, because this column had put it back. The
// legal slot and the quick-links were clean; this one query was the remainder.
//
// The lesson worth keeping: "chrome links come from the nav loader" was an
// assumption, and the census that sized the bug (nav items -> unbuilt pages) could
// not have found this, because these hrefs are not nav items at all.
//
// LIMIT 6 stays inside the query and is correct here BECAUSE the deployment
// predicate is also in the query — unlike GetNavItems, nothing is dropped after
// the cap. An empty result renders no column, which is the honest outcome.
func buildServicesHTML(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) string {
	rows, err := db.QueryContext(ctx, `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages
		WHERE site_id = $1
		  AND status IN ('deployed', 'active')
		  AND NOT (`+datahelpers.NeverDeployedPagePredicate+`)
		  AND name NOT IN ('index', 'about', 'contact', 'privacy', 'terms', 'cookies', '404', 'sitemap', 'faq', 'careers', 'insights', 'blog', 'news')
		  AND name != 'services'
		  AND (in_header = true OR in_footer = true)
		ORDER BY COALESCE(nav_order, 99), name
		LIMIT 6
	`, siteID)
	if err != nil {
		logger.Warn("buildServicesHTML: Query failed", zap.Error(err))
		return ""
	}
	defer rows.Close()

	var parts []string
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			continue
		}
		label = strings.ReplaceAll(label, "-", " ")
		words := strings.Fields(label)
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
		}
		label = strings.Join(words, " ")
		parts = append(parts, fmt.Sprintf(`<li><a href="%s">%s</a></li>`, url, label))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n                ")
}

// resolvedValueSatisfiesDeclaredType MOVED 2026-08-19 to
// datahelpers.DeclaredTypeSatisfied (bugs_open/260, council a44d9eb8 round 1,
// reuse_agent seat). It is now the shared leaf primitive for BOTH declared-type
// questions on this estate — this path's "may a resolver fill this field?" and
// the new ContentTypeViolations' "did the writer produce the declared shape?" —
// rather than two independently-maintained answers to the same question. The
// seat's objection was exactly right: this file's own header cites it as the
// live precedent for the new checker, and citing a precedent while writing a
// parallel copy of it is the fork-path pattern.

// emitChromeRenderFailedItem files ONE work item when a site-chrome component
// template fails to EXECUTE (bugs_open/260). Sibling of
// emitChromeDeadControlItem and deliberately its shape: the shared
// insertWorkItem helper (so it inherits the idx_swi_dedup-matched ON CONFLICT
// and the two-strike anti-churn label), needs_human_review with NO handler —
// nothing automated can decide whether the template or the content is wrong —
// and every failure logged rather than returned, because the refusal to store
// must not depend on the record being written.
//
// The item key carries site and slot, never the component id: a repoint to a
// different component is not a new problem while the slot is still broken, and
// a key that changes under repointing would mint a second open item for one
// defect (the lesson emitSectionDeadControlItem's own header records from
// image_url_404's site-wide key going the other way).
func emitChromeRenderFailedItem(ctx context.Context, db *sql.DB, siteID, componentID uuid.UUID,
	slot string, renderErr error, stillServing bool, logger *zap.Logger) {

	if db == nil || siteID == uuid.Nil {
		logger.Warn("render_site_components: no site identity available, chrome_render_failed item not filed",
			zap.String("slot", slot))
		return
	}

	spec := map[string]interface{}{
		"surface":       "site_component",
		"slot_name":     slot,
		"component_id":  componentID.String(),
		"render_error":  renderErr.Error(),
		"still_serving": stillServing,
		"source":        "render_site_components",
		"fix": "This chrome component's template could not be executed, so the " +
			"render was NOT stored (bugs_open/260). If still_serving is true the " +
			"slot keeps its previous bytes and the site is stale, not broken; if " +
			"false the slot has never rendered and the build was failed. The " +
			"error names the template expression and the offending value: fix " +
			"the component template, or the site data the template reads, then " +
			"re-render the site chrome. Do not 'fix' it by reinstating a " +
			"fallback renderer — that is the defect this replaced.",
	}
	specJSON, _ := json.Marshal(spec)

	summary := fmt.Sprintf("Chrome %s failed to render: %v", slot, renderErr)
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}

	compID := componentID
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("render_site_components: begin tx for chrome_render_failed failed",
			zap.String("slot", slot), zap.Error(err))
		return
	}
	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:      siteID,
		componentID: &compID,
		source:      "render-site-components",
		pipeline:    "build",
		itemType:    "chrome_render_failed",
		severity:    "high",
		summary:     summary,
		spec:        string(specJSON),
		priority:    40,
		status:      "needs_human_review",
		createdBy:   "render_site_components",
		itemKey:     fmt.Sprintf("chrome_render_failed:%s:%s", siteID, slot),
	}, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("render_site_components: insert chrome_render_failed item failed",
			zap.String("slot", slot), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("render_site_components: commit chrome_render_failed item failed",
			zap.String("slot", slot), zap.Error(err))
		return
	}
	if inserted {
		logger.Info("render_site_components: chrome_render_failed item filed for review",
			zap.String("slot", slot), zap.Bool("still_serving", stillServing))
	}
}

// emitCTAOverrideRejectedItem files an owner-visible record that the owner's
// header CTA override (sites.content_data->>'header_cta_url') was refused by the
// chrome link policy and the render degraded to the derived CTA. Before this, the
// refusal was a Warn in a pod log: the owner's explicit request could silently
// degrade for ever, with the site reading as healthy (the derived button serves)
// and nothing owner-visible saying their choice was not honoured.
//
// severity is medium, not high like the two emitters above: nothing dead ships
// and nothing failed to render — the defect is that a stated intent is being
// ignored, which is urgent to the owner and invisible to everyone else.
//
// refreshOnConflict, not the siblings' insertWorkItem (bugs_open/184, the
// bugs_closed/091 class): the key is per SITE — one override slot per site —
// while the finding names the CURRENT refused value. Under the default policy an
// owner who edits a refused override to a second, also-refused value would keep
// an open row describing the first for ever.
//
// This is NOT a third dead-control emitter in the sense of the consolidation
// trigger recorded against emitChromeDeadControlItem (page-build-pipeline
// register, `reuse_agent` disposition): no control is dropped here and the
// conflict policy differs, so the shape it would consolidate into does not fit.
func emitCTAOverrideRejectedItem(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	override, derived string, logger *zap.Logger) {

	if db == nil || siteID == uuid.Nil {
		logger.Warn("render_site_components: no site identity available, cta_override_rejected item not filed",
			zap.String("override", override))
		return
	}

	spec := map[string]interface{}{
		"surface":      "site_component",
		"slot_name":    "header",
		"override_url": override,
		"derived_url":  derived,
		"config_key":   "sites.content_data->>'header_cta_url'",
		"source":       "render_site_components",
		"fix": "The owner-set header CTA override names a target the chrome link " +
			"policy refuses — usually a page that is not deployed or does not exist " +
			"(bugs_open/191's class: an owner-set target can go stale). The site is " +
			"NOT broken: the derived CTA serves meanwhile (derived_url; empty means " +
			"no CTA rendered at all). Decide: deploy or fix the target, point the " +
			"override at a real page, or clear the key. The next chrome re-render " +
			"applies the override automatically once its target is allowed.",
	}
	specJSON, _ := json.Marshal(spec)

	summary := fmt.Sprintf("Header CTA override %q refused by chrome link policy; serving derived CTA", override)
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("render_site_components: begin tx for cta_override_rejected failed",
			zap.String("override", override), zap.Error(err))
		return
	}
	w, err := writeWorkItem(ctx, tx, workItem{
		siteID:    siteID,
		source:    "render-site-components",
		pipeline:  "build",
		itemType:  "cta_override_rejected",
		severity:  "medium",
		summary:   summary,
		spec:      string(specJSON),
		priority:  40,
		status:    "needs_human_review",
		createdBy: "render_site_components",
		// The slot is in the key (STY-054's shape) even though only the header
		// override exists today: the council's bug_historian seat noted that a
		// per-site key makes "one override slot per site" a silent assumption —
		// a second surface sharing this item_type would refresh-overwrite the
		// first finding. With the slot in the key, that door is closed rather
		// than documented.
		itemKey:   "cta_override_rejected:" + siteID.String() + ":header",
	}, refreshOnConflict, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("render_site_components: write cta_override_rejected item failed",
			zap.String("override", override), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("render_site_components: commit cta_override_rejected item failed",
			zap.String("override", override), zap.Error(err))
		return
	}
	if w.Recorded() {
		logger.Info("render_site_components: cta_override_rejected item recorded for the owner",
			zap.String("override", override), zap.Bool("inserted", w.Inserted), zap.Bool("refreshed", w.Refreshed))
	}
}

// chromeSlotHasStoredHTML reports whether this site already serves stored chrome
// for a slot. It answers ONE question, asked at one place (bugs_open/260): when
// a chrome render fails, is doing nothing safe? It is safe exactly when there
// are bytes already serving; otherwise the slot would be empty on a live site.
//
// build_status is deliberately NOT part of the predicate, unlike the
// idempotence exit above: stale-but-serving bytes are still a fallback, and the
// question here is "is anything there", not "is it up to date". A query error
// answers false — the conservative side, because it escalates rather than
// silently accepting a missing slot.
func chromeSlotHasStoredHTML(ctx context.Context, db *sql.DB, siteID uuid.UUID, slot string) bool {
	var exists bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM site_components
			WHERE site_id = $1 AND slot_name = $2
			AND rendered_html IS NOT NULL AND rendered_html != ''
		)
	`, siteID, slot).Scan(&exists); err != nil {
		return false
	}
	return exists
}
