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

	// Validate the header CTA against real pages; when there is no contact
	// nav item (or it points at a phantom), fall back to the same ranking the
	// internal-link resolver uses (interactive pages first, then hubs) rather
	// than rendering no button at all. Loader errors degrade to empty lists —
	// ctaURL then stays empty and the gated template renders no CTA.
	headerPages := loadResolverPageSet(ctx, params, siteID, params.Logger)
	if ctaURL == "" || !headerPages.Contains(ctaURL) {
		hubs, err := loadContentHubs(ctx, params, siteID, params.Logger)
		if err != nil {
			params.Logger.Warn("RenderSiteComponentsAction: loadContentHubs failed for header CTA fallback", zap.Error(err))
		}
		interactive, err := loadInteractivePages(ctx, params, siteID, params.Logger)
		if err != nil {
			params.Logger.Warn("RenderSiteComponentsAction: loadInteractivePages failed for header CTA fallback", zap.Error(err))
		}
		primary, _ := chooseCTATargets("", "", interactive, hubs)
		if primary.URL != "" && headerPages.Contains(primary.URL) {
			ctaURL = primary.URL
		} else {
			ctaURL = ""
		}
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
			"cta_text":       "Get Started",
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
	for _, slot := range slots {
		success := renderAndStoreSiteComponent(ctx, params.DB, siteID, slot, renderCtx, forceRerender, params.Logger)
		rendered[slot] = success
	}

	params.Logger.Info("RenderSiteComponentsAction: Complete",
		zap.Any("rendered", rendered),
	)

	return map[string]interface{}{
		"success":  true,
		"site_id":  siteIDStr,
		"domain":   siteData.Domain,
		"rendered": rendered,
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
	LogoText    string
	LogoURL     string

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
			sc.color_palette::text,
			ct.css_content
		FROM sites si
		LEFT JOIN style_collections sc ON si.style_collection_id = sc.id
		LEFT JOIN css_themes ct ON sc.css_theme_id = ct.id
		WHERE si.id = $1
	`, siteID).Scan(
		&s.Domain, &s.Name, &s.CompanyName, &s.Tagline,
		&s.Email, &s.Phone, &s.LogoText, &s.LogoURL,
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

// injectBrandHeadTags adds favicon + Open Graph / Twitter-card markup to a
// rendered <head>, once (idempotent — skips if a favicon link already
// exists). The favicon points at the derived favicon.png but falls back to
// the site logo when present; og:image points at the derived og-card.png.
// Absolute URLs are built from the site domain for the social tags (OG
// requires absolute), relative for the favicon link.
func injectBrandHeadTags(headHTML string, ctx *RenderContext, hasSpriteCSS bool, logger *zap.Logger) string {
	if strings.Contains(headHTML, "rel=\"icon\"") || strings.Contains(headHTML, "og:image") {
		return headHTML // already has brand head tags
	}
	idx := strings.Index(headHTML, "</head>")
	if idx == -1 {
		return headHTML // not a well-formed head; leave untouched
	}

	origin := "https://" + ctx.Domain
	title := htmlEscapeAttr(defaultString(ctx.CompanyName, ctx.Domain))
	desc := htmlEscapeAttr(ctx.Tagline)

	var b strings.Builder
	// Primary favicon = the derived square PNG; if the site has a logo we
	// also list it as a secondary icon so a mark always resolves even before
	// derive_brand_head_assets commits favicon.png.
	b.WriteString("  <link rel=\"icon\" href=\"/assets/images/favicon.png\">\n")
	if ctx.LogoURL != "" {
		b.WriteString("  <link rel=\"icon\" href=\"" + ctx.LogoURL + "\">\n")
	}
	b.WriteString("  <link rel=\"apple-touch-icon\" href=\"/assets/images/favicon.png\">\n")
	b.WriteString("  <meta property=\"og:type\" content=\"website\">\n")
	b.WriteString("  <meta property=\"og:site_name\" content=\"" + title + "\">\n")
	b.WriteString("  <meta property=\"og:title\" content=\"" + title + "\">\n")
	if desc != "" {
		b.WriteString("  <meta property=\"og:description\" content=\"" + desc + "\">\n")
	}
	b.WriteString("  <meta property=\"og:image\" content=\"" + origin + "/assets/images/og-card.png\">\n")
	b.WriteString("  <meta property=\"og:url\" content=\"" + origin + "/\">\n")
	b.WriteString("  <meta name=\"twitter:card\" content=\"summary_large_image\">\n")
	b.WriteString("  <meta name=\"twitter:image\" content=\"" + origin + "/assets/images/og-card.png\">\n")
	// Phase I2: link the site's committed sprite stylesheet (styled bullets,
	// nav accents) only when a sprite sheet exists.
	if hasSpriteCSS {
		b.WriteString("  <link rel=\"stylesheet\" href=\"/assets/css/sprites.css\">\n")
	}

	logger.Info("injectBrandHeadTags: added favicon + OG tags",
		zap.String("domain", ctx.Domain), zap.Bool("sprite_css", hasSpriteCSS))
	return headHTML[:idx] + b.String() + headHTML[idx:]
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

func renderAndStoreSiteComponent(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	slot string,
	renderCtx *RenderContext,
	force bool,
	logger *zap.Logger,
) bool {
	// Check if already rendered (unless force)
	if !force {
		var exists bool
		db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM site_components 
				WHERE site_id = $1 AND slot_name = $2 
				AND rendered_html IS NOT NULL AND rendered_html != ''
			)
		`, siteID, slot).Scan(&exists)

		if exists {
			logger.Debug("Site component already rendered, skipping",
				zap.String("slot", slot))
			return true
		}
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
		// No component assigned, try to find default
		slotToFunction := map[string]string{
			"header": "site-header",
			"footer": "site-footer",
			"head":   "head",
		}
		funcName := slot
		if mapped, ok := slotToFunction[slot]; ok {
			funcName = mapped
		}
		err = db.QueryRowContext(ctx, `
			SELECT id, html_template, COALESCE(input_schema, '{}'::jsonb)
			FROM content_components
			WHERE function = $1
			ORDER BY name LIMIT 1
		`, funcName).Scan(&componentID, &htmlTemplate, &inputSchemaRaw)

		if err != nil {
			logger.Warn("No component found for slot",
				zap.String("slot", slot),
				zap.Error(err))
			return false
		}

		// Insert the component assignment
		db.ExecContext(ctx, `
			INSERT INTO site_components (site_id, slot_name, component_id, build_status)
			VALUES ($1, $2, $3, 'pending')
			ON CONFLICT (site_id, slot_name) DO UPDATE SET component_id = $3
		`, siteID, slot, componentID)
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
	if len(inputSchemaRaw) > 0 {
		var raw map[string]interface{}
		if jErr := json.Unmarshal(inputSchemaRaw, &raw); jErr != nil {
			logger.Warn("site chrome: input_schema unparseable; fixed vocabulary only",
				zap.String("slot", slot), zap.Error(jErr))
		} else {
			// TWO live schema shapes: wrapped {"fields":{name:{...}}} (most
			// components) and FLAT {name:{...}} (e.g. "Document Head", the head
			// slot on most sites). Detect both; a shape that is neither is skipped.
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
	renderedHTML, missing, deadURLFields := RenderTemplateReportingMissing(htmlTemplate, renderCtx, logger)
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
		return false
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
		renderedHTML = injectBrandHeadTags(renderedHTML, renderCtx, spriteCount > 0, logger)
	}

	// Store the rendered HTML
	_, err = db.ExecContext(ctx, `
		UPDATE site_components 
		SET rendered_html = $1, build_status = 'rendered', updated_at = now()
		WHERE site_id = $2 AND slot_name = $3
	`, renderedHTML, siteID, slot)

	if err != nil {
		logger.Error("Failed to store rendered component",
			zap.String("slot", slot),
			zap.Error(err))
		return false
	}

	logger.Info("Site component rendered and stored",
		zap.String("slot", slot),
		zap.Int("html_length", len(renderedHTML)))

	return true
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
func buildServicesHTML(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) string {
	rows, err := db.QueryContext(ctx, `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages
		WHERE site_id = $1
		  AND status IN ('deployed', 'active')
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
