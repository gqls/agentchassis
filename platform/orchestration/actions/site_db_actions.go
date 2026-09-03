// FILE: platform/orchestration/actions/site_db_actions.go
// Database actions for site, page, and link management
// Integrates with the link management schema for multi-site/multi-network operations

package actions

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// SiteRecord represents a site in the database
type SiteRecord struct {
	ID           uuid.UUID              `json:"id"`
	NetworkID    uuid.UUID              `json:"network_id"`
	Domain       string                 `json:"domain"`
	Name         string                 `json:"name"`
	BrandDNA     map[string]interface{} `json:"brand_dna,omitempty"`
	ContentData  map[string]interface{} `json:"content_data,omitempty"`
	GithubRepo   string                 `json:"github_repo,omitempty"`
	GithubBranch string                 `json:"github_branch,omitempty"`
	Status       string                 `json:"status"`
	LastBuiltAt  *time.Time             `json:"last_built_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	CompanyName  string                 `json:"company_name,omitempty"`
	Tagline      string                 `json:"tagline,omitempty"`
	Email        string                 `json:"email,omitempty"`
	Phone        string                 `json:"phone,omitempty"`
	LogoText     string                 `json:"logo_text,omitempty"`
}

// PageRecord represents a page in the database
type PageRecord struct {
	ID              uuid.UUID  `json:"id"`
	SiteID          uuid.UUID  `json:"site_id"`
	Name            string     `json:"name"`
	URL             string     `json:"url"`
	Title           string     `json:"title"`
	PageType        string     `json:"page_type"`
	Status          string     `json:"status"`
	ContentHash     string     `json:"content_hash,omitempty"`
	MetaDescription string     `json:"meta_description,omitempty"`
	NavLabel        string     `json:"nav_label,omitempty"`
	NavOrder        int        `json:"nav_order"`
	InHeader        bool       `json:"in_header"`
	InFooter        bool       `json:"in_footer"`
	LastBuiltAt     *time.Time `json:"last_built_at,omitempty"`

	// SectionsKept is true when the upsert REFUSED an empty incoming sections
	// list over a non-empty stored one (bugs_open/204). Not persisted — it is
	// this call's own answer, and it exists because a silent keep would be a new
	// landmine: the write would report success and the plan would quietly not be
	// what the database holds. The caller records it durably.
	SectionsKept bool `json:"sections_kept,omitempty"`
}

// NavigationItem represents a single item in navigation
type NavigationItem struct {
	PageID   string           `json:"page_id"`
	Label    string           `json:"label"`
	URL      string           `json:"url"`
	Children []NavigationItem `json:"children,omitempty"`
}

// NavigationStructure represents the full navigation for a site
type NavigationStructure struct {
	Items []NavigationItem `json:"items"`
}

// LinkRegistryEntry represents a link extracted from a page
type LinkRegistryEntry struct {
	ID                        uuid.UUID  `json:"id,omitempty"`
	SourceComponentInstanceID *uuid.UUID `json:"source_component_instance_id,omitempty"`
	SourcePageID              uuid.UUID  `json:"source_page_id"`
	SourceSiteID              uuid.UUID  `json:"source_site_id"`
	TargetURL                 string     `json:"target_url"`
	TargetPageID              *uuid.UUID `json:"target_page_id,omitempty"`
	TargetSiteID              *uuid.UUID `json:"target_site_id,omitempty"`
	Scope                     string     `json:"scope"`
	LinkType                  string     `json:"link_type"`
	AnchorText                string     `json:"anchor_text,omitempty"`
	RelAttr                   string     `json:"rel_attr,omitempty"`
	Status                    string     `json:"status"`
}

// ============================================================================
// ACTION: ensure_site_record
// ============================================================================

// EnsureSiteRecordAction creates or updates a site record in the database
// This is the first step in the multipage-website-builder workflow
func EnsureSiteRecordAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("EnsureSiteRecordAction: Starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Handle initialization
	if params.ExecutionContext.Action == "initialize" {
		params.Logger.Info("EnsureSiteRecordAction: Handling initialization")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Check database availability
	if params.DB == nil {
		params.Logger.Warn("EnsureSiteRecordAction: Database not available, returning placeholder")
		return createPlaceholderSiteRecord(params)
	}

	// Extract domain from input_data
	params.Logger.Info("EnsureSiteRecordAction: Extracting domain")

	// Check for a config-declared destination domain override first.
	// Used by site-adoption-agent to separate the crawled source URL from
	// the destination domain being built. The override is a dot-path into
	// CollectedData (e.g. "input_data.destination_domain"). If the path is
	// missing or resolves to an empty string, we fall through to the
	// existing aggressive search in extractDomainFromInput — so callers
	// that don't set destination_domain behave exactly as before.
	var domain string
	if overridePath, ok := params.StepConfig.Config["domain_override_field"].(string); ok && overridePath != "" {
		if explicitDomain := datahelpers.ExtractNestedFieldString(params.CollectedData, overridePath); explicitDomain != "" {
			params.Logger.Info("EnsureSiteRecordAction: using destination domain override",
				zap.String("override_path", overridePath),
				zap.String("destination_domain", explicitDomain),
			)
			domain = explicitDomain
		} else {
			params.Logger.Info("EnsureSiteRecordAction: domain_override_field set but path empty, falling through to extractDomainFromInput",
				zap.String("override_path", overridePath),
			)
		}
	}
	if domain == "" {
		domain = extractDomainFromInput(params.CollectedData, params.Logger)
	}
	if domain == "" {
		params.Logger.Error("EnsureSiteRecordAction: Domain not found")
		return nil, fmt.Errorf("domain not found in input_data")
	}

	// Clean domain (remove protocol, trailing slashes)
	domain = cleanDomain(domain)

	// Validate the domain is a real DNS-looking string, not a config
	// placeholder or dot-path reference that leaked in via the aggressive
	// domain search. Without this, the action would silently write a junk
	// site row (observed case: literal "site_record.domain" stored as a
	// domain because FindDomainAggressive matched the config placeholder).
	if !isPlausibleDomain(domain) {
		params.Logger.Error("EnsureSiteRecordAction: extracted value does not look like a domain — refusing to write",
			zap.String("rejected_value", domain))
		return nil, fmt.Errorf("extracted domain %q is not a plausible DNS domain — likely a config placeholder or dot-path reference leaked into domain extraction", domain)
	}

	params.Logger.Info("EnsureSiteRecordAction: Domain cleaned and validated",
		zap.String("domain", domain))

	// Get or create default network ID
	params.Logger.Info("EnsureSiteRecordAction: Getting default network ID")
	networkID, err := getDefaultNetworkID(ctx, params.DB, params.Logger)
	if err != nil {
		params.Logger.Warn("EnsureSiteRecordAction: Failed to get default network, using placeholder",
			zap.Error(err))
		networkID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	}
	params.Logger.Info("EnsureSiteRecordAction: Got network ID",
		zap.String("network_id", networkID.String()))

	// Upsert site record
	params.Logger.Info("EnsureSiteRecordAction: Upserting site record")
	siteRecord, err := upsertSite(ctx, params.DB, domain, networkID, params.Logger)
	if err != nil {
		params.Logger.Error("EnsureSiteRecordAction: Failed to upsert site", zap.Error(err))
		return nil, fmt.Errorf("failed to upsert site: %w", err)
	}

	params.Logger.Info("EnsureSiteRecordAction: Site record ready",
		zap.String("site_id", siteRecord.ID.String()),
		zap.String("domain", siteRecord.Domain),
	)

	// Seed the network's default analytics container for a site with no site_config yet
	// (bugs_open/397 §6.2). Best-effort by contract: failure leaves the site exactly as
	// unseeded as before this mechanism existed, and the backfill census catches it.
	if seeded, err := seedAnalyticsDefault(ctx, params.DB, siteRecord.ID, networkID, params.Logger); err != nil {
		params.Logger.Warn("EnsureSiteRecordAction: analytics default seed failed (non-fatal)",
			zap.Error(err))
	} else if seeded {
		params.Logger.Info("EnsureSiteRecordAction: seeded network default analytics container",
			zap.String("site_id", siteRecord.ID.String()))
	}

	return map[string]interface{}{
		"site_id":      siteRecord.ID.String(),
		"domain":       siteRecord.Domain,
		"content_data": siteRecord.ContentData,
		"network_id":   siteRecord.NetworkID.String(),
		"status":       siteRecord.Status,
		"created":      siteRecord.CreatedAt.Format(time.RFC3339),
		"company_name": siteRecord.CompanyName,
		"tagline":      siteRecord.Tagline,
		"email":        siteRecord.Email,
		"phone":        siteRecord.Phone,
		"logo_text":    siteRecord.LogoText,
		// Empty for the overwhelming majority of sites, which deploy to the default
		// "sites" repo (→ B2). Set only on VM-hosted sites; resolveGitRepoName reads it
		// from site_record.github_repo to pick the deploy target.
		"github_repo": siteRecord.GithubRepo,
	}, nil
}

// ============================================================================
// ACTION: sync_pages_to_db
// ============================================================================

// SyncPagesToDBAction syncs pages from strategist output to database
// and builds/caches the navigation structure
func SyncPagesToDBAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SyncPagesToDBAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Handle initialization
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Extract site_id
	siteIDStr := extractSiteID(params.CollectedData, params.Logger)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found in collected data")
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Extract pages from page_plan
	pages := extractPagesFromPlan(params.CollectedData, params.Logger)
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found in page_plan")
	}

	// Canonicalise LLM-emitted page identities through the SAME pipeline the
	// plan-writer uses: ValidateRoles (cross-page role correction) THEN
	// CanonicalisePage (doc 029 Phase 0). This path previously called
	// CanonicalisePage alone, so a section hub the LLM typed "content"
	// (e.g. games-index) was written FLAT to pages (/games-index.html,
	// page_type content) while WriteSitePlanAction — which runs ValidateRoles
	// first — wrote it section-index (/games/index.html) to site_plan_pages.
	// The two canonicalisation surfaces diverged and pages regressed a correct
	// plan. Running ValidateRoles here makes both surfaces apply identical
	// rules, so they cannot disagree. Mirrors write_site_plan_action.go.
	//
	// ValidateRoles needs the full set (rule 3 reads cross-page parents) and
	// preserves order, so validated[i] aligns with pages[i].
	//
	// Mutates the page maps in place. CollectedData consumers that read
	// page_plan downstream will see the canonical name/url/page_type —
	// by design, since the canonical identity is the one that survives.
	llmPages := make([]datahelpers.LLMPlannedPage, 0, len(pages))
	for i, page := range pages {
		name := datahelpers.GetStringField(page, "name", "")
		if name == "" {
			name = fmt.Sprintf("page-%d", i+1)
		}
		llmPages = append(llmPages, datahelpers.LLMPlannedPage{
			Name:          name,
			Role:          firstNonEmptyField(page, "page_type", "type", "role"),
			Slug:          datahelpers.GetStringField(page, "slug", ""),
			URL:           datahelpers.GetStringField(page, "url", ""),
			ParentSection: datahelpers.GetStringField(page, "parent_section", ""),
		})
	}
	validated := datahelpers.ValidateRoles(llmPages)

	// Site-level URL shape, read through the ONE shared helper — this and
	// WriteSitePlanAction must agree on it or the plan and the pages table
	// carry different URLs for the same page (bugs_open/241; the divergence
	// comment above). Nil DB (the nav-from-plan-only path below) keeps the
	// nested default.
	flatURLs := siteUsesFlatURLs(ctx, params.DB, siteID, params.Logger)
	// Who owns a page's identity — same helper, same reason as the URL shape
	// above. This surface is where the phantom is actually INSERTed (upsertPage
	// conflicts on (site_id, name) and nothing else), so a re-derivation here
	// undoes the reconciler no matter what the plan says (bugs_open/215).
	identityPolicy := siteIdentityPolicyFor(ctx, params.DB, siteID, params.Logger)

	normalised := make([]map[string]interface{}, 0, len(pages))
	for i, page := range pages {
		v := validated[i]
		name, url, pageType := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
			Role:          v.Role,
			Slug:          firstNonEmpty(v.Slug, v.Name),
			ParentSection: v.ParentSection,
			FlatURLs:      flatURLs,
		})
		if identityPolicy.HonourRealisedIdentity {
			if rName, rURL, rType, ok := realisedIdentityOf(page); ok {
				if rName != name || rURL != url {
					params.Logger.Info("SyncPagesToDBAction: kept realised page identity over canonicalisation",
						zap.String("realised_name", rName), zap.String("realised_url", rURL),
						zap.String("would_have_been_name", name),
						zap.String("would_have_been_url", url))
				}
				name, url, pageType = rName, rURL, rType
			}
		}
		if name == "" {
			params.Logger.Warn("SyncPagesToDBAction: page failed canonicalisation, skipping",
				zap.String("raw_name", v.Name),
				zap.String("raw_role", v.Role))
			continue
		}
		if v.CorrectedFromRole != "" {
			params.Logger.Info("SyncPagesToDBAction: corrected page role",
				zap.String("name", name),
				zap.String("from", v.CorrectedFromRole),
				zap.String("to", v.Role))
		}
		page["name"] = name
		page["url"] = url
		page["type"] = pageType
		page["page_type"] = pageType
		normalised = append(normalised, page)
	}
	pages = normalised

	params.Logger.Info("SyncPagesToDBAction: Found pages to sync",
		zap.Int("page_count", len(pages)),
		zap.String("site_id", siteIDStr),
	)

	// If no DB, build navigation from pages directly
	if params.DB == nil {
		params.Logger.Warn("SyncPagesToDBAction: Database not available, building nav from plan only")
		navigation := buildNavigationFromPages(pages)
		return map[string]interface{}{
			"pages_synced": len(pages),
			"navigation":   navigation,
			"db_available": false,
		}, nil
	}

	// Resolve the current plan version so each synced page can record which
	// plan it was built from. reconcile_site_plan treats a NULL
	// built_from_plan_version as "stale" and re-emits the page every run, so
	// without this the hubs churn forever. Best-effort: greenfield callers
	// (pageflow-builder etc.) have no plan — currentPlanID stays invalid and
	// built_from_plan_version is left NULL, i.e. unchanged behaviour for them.
	var currentPlanID uuid.NullUUID
	if err := params.DB.QueryRowContext(ctx, `
		SELECT id FROM site_plans WHERE site_id = $1 AND is_current = true
	`, siteID).Scan(&currentPlanID); err != nil {
		params.Logger.Info("SyncPagesToDBAction: no current site_plan; built_from_plan_version left unset",
			zap.String("site_id", siteIDStr))
		currentPlanID = uuid.NullUUID{}
	}

	// bugs_open/204: the ONE channel that may legitimately empty a live page's
	// section list. recompose_pages already means "this page is released for
	// redesign" — reusing it means no operator has to remember a new flag, and a
	// reviewer of the CALLER can see the intent. A page not named here cannot be
	// emptied by a sync, whatever the plan says.
	recompose := recomposePagesFromSpec(params.CollectedData, params.Logger)

	// Sync each page to database
	syncedCount := 0
	var sectionsRefused []string
	for i, page := range pages {
		pageName, _ := page["name"].(string)
		pageRecord, err := upsertPage(ctx, params.DB, siteID, page, i, currentPlanID, recompose[pageName], params.Logger)
		if err != nil {
			params.Logger.Error("Failed to upsert page",
				zap.Any("page", page),
				zap.Error(err))
			continue
		}
		syncedCount++
		if pageRecord.SectionsKept {
			// Loud AND durable. This is the guard doing its job, but it also means
			// the plan and the database now disagree about this page, and whoever
			// reads the plan afterwards needs to know which one they are looking at.
			sectionsRefused = append(sectionsRefused, pageRecord.Name)
			params.Logger.Warn("SyncPagesToDBAction: refused an empty sections list over a non-empty stored one",
				zap.String("page", pageRecord.Name),
				zap.String("site_id", siteIDStr))
		}
		params.Logger.Debug("Page synced",
			zap.String("page_id", pageRecord.ID.String()),
			zap.String("name", pageRecord.Name),
		)
	}

	// Get navigation from database (trigger will have invalidated cache)
	/*navigation, err := getNavigationFromDB(ctx, params.DB, siteID, "header", params.Logger)
	if err != nil {
		params.Logger.Warn("Failed to get navigation from DB, building from plan",
			zap.Error(err))
		navigation = buildNavigationFromPages(pages)
	}*/
	navigation, err := GetNavigationStructure(ctx, params.DB, siteID, "header", params.Logger)
	if err != nil {
		params.Logger.Warn("Failed to get navigation, building from plan", zap.Error(err))
		navigation = buildNavigationFromPages(pages)
	}

	// bugs_open/204: a refusal is durable, not just a log line. A log line on a
	// service whose output rotates sub-second is not a record — and the whole
	// reason this guard exists is that the destruction it prevents was invisible
	// for three days. Best-effort by construction: a failed write must not change
	// a sync that has already happened.
	if len(sectionsRefused) > 0 {
		LogActionError(ctx, params, siteIDStr, "", "sync_pages_to_db",
			"PAGE_SECTIONS_EMPTY_OVERWRITE_REFUSED", "warning",
			fmt.Sprintf("%d page(s) proposed an EMPTY sections list over a non-empty stored one; the stored list was kept: %s",
				len(sectionsRefused), strings.Join(sectionsRefused, ", ")),
			map[string]interface{}{
				"pages":  sectionsRefused,
				"why":    "pages.sections is the only record of a decomposed page's composition; page_components keeps serving after it is emptied, so the damage is invisible until the next rebuild builds an empty page over a live one (measured 2026-08-20: 41 of 45 live pages on one site).",
				"remedy": "the plan and the database now disagree for these pages, and the DATABASE is what the site is built from. If the emptying was intended, name the page in the run's spec.recompose_pages release and re-run. If it was not, this is upstream: check whether validate_plan dropped section names for these pages (agent_error_log, error_code='PLAN_SECTION_NAME_DROPPED') — bugs_open/204.",
			},
			params.Logger)
	}

	params.Logger.Info("SyncPagesToDBAction: Complete",
		zap.Int("pages_synced", syncedCount),
		zap.Int("nav_items", len(navigation.Items)),
		zap.Int("sections_overwrites_refused", len(sectionsRefused)),
	)

	return map[string]interface{}{
		"pages_synced":                syncedCount,
		"navigation":                  navigation,
		"site_id":                     siteIDStr,
		"db_available":                true,
		"sections_overwrites_refused": len(sectionsRefused),
	}, nil
}

// ============================================================================
// ACTION: extract_and_sync_links
// ============================================================================

// ExtractAndSyncLinksAction extracts links from rendered HTML and syncs to link_registry
func ExtractAndSyncLinksAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ExtractAndSyncLinksAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Handle initialization
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Extract required data
	siteIDStr := extractSiteID(params.CollectedData, params.Logger)
	htmlContent := extractHTMLContentLocal(params.CollectedData, params.Logger)

	if htmlContent == "" {
		params.Logger.Warn("ExtractAndSyncLinksAction: No HTML content found")
		return map[string]interface{}{
			"links_extracted": 0,
			"warning":         "no HTML content found",
		}, nil
	}

	// Extract current page info
	currentPage := extractCurrentPage(params.CollectedData, params.Logger)
	pageName := ""
	if name, ok := currentPage["name"].(string); ok {
		pageName = name
	}

	// Parse site ID
	var siteID uuid.UUID
	if siteIDStr != "" {
		var err error
		siteID, err = uuid.Parse(siteIDStr)
		if err != nil {
			params.Logger.Warn("Invalid site_id, links won't be persisted",
				zap.String("site_id", siteIDStr),
				zap.Error(err))
		}
	}

	// Extract links from HTML
	links := extractLinksFromHTML(htmlContent, siteID, pageName, params.Logger)

	params.Logger.Info("ExtractAndSyncLinksAction: Links extracted",
		zap.Int("link_count", len(links)),
		zap.String("page_name", pageName),
	)

	// If no DB or no valid site ID, just return the extracted links
	if params.DB == nil || siteID == uuid.Nil {
		params.Logger.Warn("ExtractAndSyncLinksAction: Cannot persist links (no DB or site_id)")
		return map[string]interface{}{
			"links_extracted": len(links),
			"links":           links,
			"persisted":       false,
		}, nil
	}

	// Get page ID from database
	pageID, err := getPageID(ctx, params.DB, siteID, pageName, params.Logger)
	if err != nil {
		params.Logger.Warn("Failed to get page ID, links won't be persisted",
			zap.Error(err))
		return map[string]interface{}{
			"links_extracted": len(links),
			"links":           links,
			"persisted":       false,
		}, nil
	}

	// COMPLETENESS FLOOR (bugs_open/165 site C). syncLinksToDB deletes every
	// link_registry row for this page and re-inserts what the extractor returned.
	// goquery over truncated or partially-rendered HTML returns a short result
	// with no error, so this asks — before anything is destroyed — whether this
	// run saw enough of the page to be reconciling against it.
	//
	// A refusal ERRORS, the same contract as sites A and B. It is nested in
	// multipage-website-builder's generate_pages_loop, which carries no
	// continue_on_error, so this does fail the whole site build — a real
	// disproportion, filed at its own layer as bugs_open/173 rather than worked
	// around here by making the action never error (council c69e935a round 2, four
	// seats). That loop already fails the build on any substep error; this adds one
	// more, and the floor is inert while link_registry is empty.
	floorDetail, err := enforceLinkRegistryFloor(ctx, params, siteID, pageID, pageName, len(links))
	if err != nil {
		return nil, err
	}

	// Sync links to database
	syncedCount, err := syncLinksToDB(ctx, params.DB, siteID, pageID, links, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to sync links to database",
			zap.Error(err))
		return map[string]interface{}{
			"links_extracted": len(links),
			"links":           links,
			"persisted":       false,
			"error":           err.Error(),
		}, nil
	}

	result := map[string]interface{}{
		"links_extracted": len(links),
		"links_persisted": syncedCount,
		"page_id":         pageID.String(),
		"persisted":       true,
	}
	// Publish the floor's numbers beside the counts — a bare "links_persisted: 3"
	// is the alarm presented as output without its denominator.
	for k, v := range floorDetail {
		result[k] = v
	}
	return result, nil
}

// ============================================================================
// ACTION: update_site_timestamps
// ============================================================================

// UpdateSiteTimestampsAction updates last_built_at and last_deployed_at
func UpdateSiteTimestampsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteTimestampsAction: Starting")

	// Handle initialization
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	siteIDStr := extractSiteID(params.CollectedData, params.Logger)
	if siteIDStr == "" {
		params.Logger.Warn("UpdateSiteTimestampsAction: No site_id found")
		return map[string]interface{}{"updated": false}, nil
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteTimestampsAction: No database available")
		return map[string]interface{}{"updated": false}, nil
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	now := time.Now().UTC()
	err = updateSiteTimestamps(ctx, params.DB, siteID, now, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to update site timestamps", zap.Error(err))
		return map[string]interface{}{"updated": false, "error": err.Error()}, nil
	}

	return map[string]interface{}{
		"updated":          true,
		"last_built_at":    now.Format(time.RFC3339),
		"last_deployed_at": now.Format(time.RFC3339),
	}, nil
}

// ============================================================================
// ACTION: get_navigation_from_db
// ============================================================================

// GetNavigationFromDBAction retrieves cached navigation structure
func GetNavigationFromDBAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("GetNavigationFromDBAction: Starting")

	// Handle initialization
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	siteIDStr := extractSiteID(params.CollectedData, params.Logger)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found")
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	navType := "header"
	if nt, ok := params.StepConfig.Config["nav_type"].(string); ok {
		navType = nt
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	/*navigation, err := getNavigationFromDB(ctx, params.DB, siteID, navType, params.Logger)*/
	navigation, err := GetNavigationStructure(ctx, params.DB, siteID, navType, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get navigation: %w", err)
	}

	return map[string]interface{}{
		"navigation": navigation,
		"nav_type":   navType,
		"site_id":    siteIDStr,
	}, nil
}

// ============================================================================
// HELPER FUNCTIONS - Data Extraction
// ============================================================================

func extractDomainFromInput(data map[string]interface{}, logger *zap.Logger) string {
	// Use the unified extractor which handles deeply nested input_data structures
	logger.Info("Using unified extractor to find domain")

	// Method 1: Use datahelpers.ExtractFields with input_data
	extracted := datahelpers.ExtractFields(data, []string{"input_data"}, logger)
	if domain, ok := extracted["domain"].(string); ok && domain != "" {
		logger.Info("Found domain via ExtractFields", zap.String("domain", domain))
		return domain
	}

	// Method 2: Use aggressive domain search directly
	domain := datahelpers.FindDomainAggressive(data, logger)
	if domain != "" {
		logger.Info("Found domain via FindDomainAggressive", zap.String("domain", domain))
		return domain
	}

	// Method 3: Fallback - Try direct paths (for simpler cases)
	// Try input_data.domain
	if inputData, ok := data["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok && domain != "" {
			return domain
		}
	}

	// Try direct domain field
	if domain, ok := data["domain"].(string); ok && domain != "" {
		return domain
	}

	logger.Warn("Domain not found via any method")
	return ""
}

func cleanDomain(domain string) string {
	// Remove protocol
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	// Remove trailing slash
	domain = strings.TrimSuffix(domain, "/")
	// Remove www. prefix for consistency
	domain = strings.TrimPrefix(domain, "www.")
	return domain
}

// isPlausibleDomain returns true if the string looks like a real DNS domain
// rather than a config placeholder, dot-path reference, or other junk that
// may have leaked in via the aggressive domain search.
//
// Rejection reasons observed in production:
//   - "site_record.domain"     — config dot-path placeholder
//   - "input_data.domain"      — workflow config reference
//   - ""                       — empty (handled earlier but belt-and-braces)
//   - "localhost"              — single word, no TLD
//   - strings containing spaces or internal slashes
//
// This is deliberately permissive — it's a smoke test, not a DNS validator.
// Real DNS resolution is not done here because (a) it would make the action
// slow and network-dependent, and (b) we accept domains that aren't yet
// registered (the user may be building a site before they register).
func isPlausibleDomain(s string) bool {
	if s == "" {
		return false
	}
	// Must have at least one dot (TLD separator). Single-word values like
	// "localhost" or "sitename" are rejected — the system is always
	// dealing with fully-qualified domains.
	if !strings.Contains(s, ".") {
		return false
	}
	// Whitespace anywhere is disqualifying — DNS labels cannot contain it.
	if strings.ContainsAny(s, " \t\n\r") {
		return false
	}
	// Slashes mean we got a URL rather than a bare domain, or a path
	// fragment. cleanDomain should have stripped these, so anything left
	// is suspicious.
	if strings.ContainsAny(s, "/\\") {
		return false
	}
	// Known config-placeholder shapes: dot-path references to fields in
	// collected_data. These all start with a well-known root segment that
	// a real domain would never begin with.
	placeholderPrefixes := []string{
		"input_data.",
		"site_record.",
		"agent_config.",
		"__", // internal CollectedData keys like __raw_message__
	}
	for _, p := range placeholderPrefixes {
		if strings.HasPrefix(s, p) {
			return false
		}
	}
	// Known single-token placeholder strings that happen to contain a dot
	// but aren't domains. Extend this list if new patterns emerge.
	knownPlaceholders := map[string]bool{
		"site_record.domain":  true,
		"site_record.site_id": true,
		"input_data.domain":   true,
		"input_data.url":      true,
	}
	if knownPlaceholders[s] {
		return false
	}
	// Last defence: the TLD (everything after the final dot) should be
	// 2+ alphabetic characters. This rejects shapes like "foo.1" or
	// "something." that technically contain a dot but aren't domains.
	lastDot := strings.LastIndex(s, ".")
	tld := s[lastDot+1:]
	if len(tld) < 2 {
		return false
	}
	for _, r := range tld {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

func extractSiteID(data map[string]interface{}, logger *zap.Logger) string {
	// Try site_record.site_id first (most common after ensure_site_record)
	if siteRecord, ok := data["site_record"].(map[string]interface{}); ok {
		if siteID, ok := siteRecord["site_id"].(string); ok && siteID != "" {
			logger.Info("Found site_id in site_record", zap.String("site_id", siteID))
			return siteID
		}
	}

	// Try direct site_id
	if siteID, ok := data["site_id"].(string); ok && siteID != "" {
		return siteID
	}

	// Try db_sync.site_id
	if dbSync, ok := data["db_sync"].(map[string]interface{}); ok {
		if siteID, ok := dbSync["site_id"].(string); ok && siteID != "" {
			return siteID
		}
	}

	// Use unified extractor as fallback
	extracted := datahelpers.ExtractFields(data, []string{"site_record"}, logger)
	if siteRecord, ok := extracted["site_record"].(map[string]interface{}); ok {
		if siteID, ok := siteRecord["site_id"].(string); ok && siteID != "" {
			logger.Info("Found site_id via unified extractor", zap.String("site_id", siteID))
			return siteID
		}
	}

	logger.Debug("site_id not found in collected data")
	return ""
}

func extractPagesFromPlan(data map[string]interface{}, logger *zap.Logger) []map[string]interface{} {
	var pages []map[string]interface{}

	// Try multiple field names: page_plan (old) and site_plan (new v3 workflow)
	fieldNames := []string{"page_plan", "site_plan"}

	for _, fieldName := range fieldNames {
		extracted := datahelpers.ExtractFields(data, []string{fieldName}, logger)

		if planMap, ok := extracted[fieldName].(map[string]interface{}); ok {
			pages = extractPagesFromPagePlanMap(planMap, logger)
			if len(pages) > 0 {
				logger.Info("Extracted pages via unified extractor",
					zap.String("field", fieldName),
					zap.Int("count", len(pages)))
				return pages
			}
		}

		// Fallback: Try direct lookup in collected data
		if planMap, ok := data[fieldName].(map[string]interface{}); ok {
			pages = extractPagesFromPagePlanMap(planMap, logger)
			if len(pages) > 0 {
				logger.Info("Extracted pages from direct field",
					zap.String("field", fieldName),
					zap.Int("count", len(pages)))
				return pages
			}
		}
	}

	logger.Warn("No pages found in page_plan or site_plan")
	return pages
}

// extractPagesFromPagePlanMap extracts pages array from a page_plan or site_plan map
func extractPagesFromPagePlanMap(pagePlan map[string]interface{}, logger *zap.Logger) []map[string]interface{} {
	var pages []map[string]interface{}

	// Check if actual data is nested under "response" (call_agent response format)
	// This happens when call_agent preserves metadata and puts response data under "response" key
	if response, ok := pagePlan["response"].(map[string]interface{}); ok {
		logger.Info("Found response wrapper in plan data, extracting from response",
			zap.Strings("response_keys", datahelpers.GetMapKeys(response)))
		// Recursively extract from the response
		pages = extractPagesFromPagePlanMap(response, logger)
		if len(pages) > 0 {
			return pages
		}
	}

	// Try page_plan.plan_data.pages (v2 workflow format)
	if planData, ok := pagePlan["plan_data"].(map[string]interface{}); ok {
		if pagesArr, ok := planData["pages"].([]interface{}); ok {
			for _, p := range pagesArr {
				if pageMap, ok := p.(map[string]interface{}); ok {
					pages = append(pages, pageMap)
				}
			}
		}
		// Also try sections (older format)
		if len(pages) == 0 {
			if sectionsArr, ok := planData["sections"].([]interface{}); ok {
				for _, s := range sectionsArr {
					if sectionMap, ok := s.(map[string]interface{}); ok {
						pages = append(pages, sectionMap)
					}
				}
			}
		}
	}

	// Try validated_plan.pages (v3 site-planner format)
	if len(pages) == 0 {
		if validatedPlan, ok := pagePlan["validated_plan"].(map[string]interface{}); ok {
			if pagesArr, ok := validatedPlan["pages"].([]interface{}); ok {
				for _, p := range pagesArr {
					if pageMap, ok := p.(map[string]interface{}); ok {
						pages = append(pages, pageMap)
					}
				}
				if len(pages) > 0 {
					logger.Info("Found pages in validated_plan.pages", zap.Int("count", len(pages)))
				}
			}
		}
	}

	// Try llm_plan.result.pages (v3 site-planner raw LLM output)
	if len(pages) == 0 {
		if llmPlan, ok := pagePlan["llm_plan"].(map[string]interface{}); ok {
			if result, ok := llmPlan["result"].(map[string]interface{}); ok {
				if pagesArr, ok := result["pages"].([]interface{}); ok {
					for _, p := range pagesArr {
						if pageMap, ok := p.(map[string]interface{}); ok {
							pages = append(pages, pageMap)
						}
					}
					if len(pages) > 0 {
						logger.Info("Found pages in llm_plan.result.pages", zap.Int("count", len(pages)))
					}
				}
			}
		}
	}

	// Try direct pages under page_plan
	if len(pages) == 0 {
		if pagesArr, ok := pagePlan["pages"].([]interface{}); ok {
			for _, p := range pagesArr {
				if pageMap, ok := p.(map[string]interface{}); ok {
					pages = append(pages, pageMap)
				}
			}
		}
	}

	// Try result.pages (if LLM response wrapped in result)
	if len(pages) == 0 {
		if result, ok := pagePlan["result"].(map[string]interface{}); ok {
			if pagesArr, ok := result["pages"].([]interface{}); ok {
				for _, p := range pagesArr {
					if pageMap, ok := p.(map[string]interface{}); ok {
						pages = append(pages, pageMap)
					}
				}
			}
		}
	}

	return pages
}

func extractHTMLContentLocal(data map[string]interface{}, logger *zap.Logger) string {
	// Try page_html directly as string
	if html, ok := data["page_html"].(string); ok && html != "" {
		return html
	}

	// Try page_html as map with various content field names
	if pageHTML, ok := data["page_html"].(map[string]interface{}); ok {
		// Try common field names
		for _, fieldName := range []string{"html", "content", "result", "html_content"} {
			if html, ok := pageHTML[fieldName].(string); ok && html != "" {
				logger.Info("Found HTML in page_html map", zap.String("field", fieldName))
				return html
			}
		}
	}

	// Try html_content directly
	if html, ok := data["html_content"].(string); ok && html != "" {
		return html
	}

	// Use unified extractor as fallback
	extracted := datahelpers.ExtractFields(data, []string{"page_html"}, logger)
	if pageHTML, ok := extracted["page_html"]; ok {
		// Handle string result
		if html, ok := pageHTML.(string); ok && html != "" {
			logger.Info("Found HTML via unified extractor (string)")
			return html
		}
		// Handle map result
		if htmlMap, ok := pageHTML.(map[string]interface{}); ok {
			for _, fieldName := range []string{"html", "content", "result"} {
				if html, ok := htmlMap[fieldName].(string); ok && html != "" {
					logger.Info("Found HTML via unified extractor (map)", zap.String("field", fieldName))
					return html
				}
			}
		}
	}

	logger.Debug("No HTML content found in collected data")
	return ""
}

func extractCurrentPage(data map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	if currentPage, ok := data["current_page"].(map[string]interface{}); ok {
		return currentPage
	}
	return make(map[string]interface{})
}

// ============================================================================
// HELPER FUNCTIONS - Database Operations
// ============================================================================

func createPlaceholderSiteRecord(params ActionParams) (interface{}, error) {
	domain := extractDomainFromInput(params.CollectedData, params.Logger)
	if domain == "" {
		domain = "unknown-domain"
	}
	domain = cleanDomain(domain)

	return map[string]interface{}{
		"site_id":      uuid.New().String(),
		"domain":       domain,
		"network_id":   "00000000-0000-0000-0000-000000000002",
		"status":       "placeholder",
		"created":      time.Now().UTC().Format(time.RFC3339),
		"db_available": false,
	}, nil
}

func getDefaultNetworkID(ctx context.Context, db interface{}, logger *zap.Logger) (uuid.UUID, error) {
	logger.Info("getDefaultNetworkID: Querying for default network")

	// Add a 5-second timeout to prevent indefinite hanging
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id FROM networks WHERE slug = 'default' LIMIT 1`

	var networkID uuid.UUID

	switch d := db.(type) {
	case *sql.DB:
		logger.Debug("getDefaultNetworkID: Using *sql.DB")
		err := d.QueryRowContext(queryCtx, query).Scan(&networkID)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Warn("getDefaultNetworkID: No default network found, will create one")
				return createDefaultNetwork(ctx, db, logger)
			}
			if queryCtx.Err() == context.DeadlineExceeded {
				logger.Error("getDefaultNetworkID: Query timed out after 5 seconds")
			}
			return uuid.Nil, fmt.Errorf("query failed: %w", err)
		}
	case *pgxpool.Pool:
		logger.Debug("getDefaultNetworkID: Using *pgxpool.Pool")
		err := d.QueryRow(queryCtx, query).Scan(&networkID)
		if err != nil {
			if err.Error() == "no rows in result set" {
				logger.Warn("getDefaultNetworkID: No default network found, will create one")
				return createDefaultNetwork(ctx, db, logger)
			}
			if queryCtx.Err() == context.DeadlineExceeded {
				logger.Error("getDefaultNetworkID: Query timed out after 5 seconds")
			}
			return uuid.Nil, fmt.Errorf("query failed: %w", err)
		}
	default:
		return uuid.Nil, fmt.Errorf("unsupported database type: %T", db)
	}

	logger.Info("getDefaultNetworkID: Found default network",
		zap.String("network_id", networkID.String()))
	return networkID, nil
}

// createDefaultNetwork creates the default network if it doesn't exist
func createDefaultNetwork(ctx context.Context, db interface{}, logger *zap.Logger) (uuid.UUID, error) {
	logger.Info("createDefaultNetwork: Creating default network")

	query := `
		INSERT INTO networks (slug, name, status)
		VALUES ('default', 'Default Network', 'active')
		ON CONFLICT (slug) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`

	var networkID uuid.UUID

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query).Scan(&networkID)
		if err != nil {
			logger.Error("createDefaultNetwork: Failed to create network", zap.Error(err))
			// Return a hardcoded fallback UUID
			return uuid.MustParse("00000000-0000-0000-0000-000000000002"), nil
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query).Scan(&networkID)
		if err != nil {
			logger.Error("createDefaultNetwork: Failed to create network", zap.Error(err))
			return uuid.MustParse("00000000-0000-0000-0000-000000000002"), nil
		}
	default:
		return uuid.MustParse("00000000-0000-0000-0000-000000000002"), nil
	}

	logger.Info("createDefaultNetwork: Created default network",
		zap.String("network_id", networkID.String()))
	return networkID, nil
}

func upsertSite(ctx context.Context, db interface{}, domain string, networkID uuid.UUID, logger *zap.Logger) (*SiteRecord, error) {
	query := `
		INSERT INTO sites (domain, name, network_id, status)
		VALUES ($1, $1, $2, 'active')
		ON CONFLICT (domain) DO UPDATE SET
			updated_at = NOW()
		RETURNING id, network_id, domain, name, status, created_at,
				  COALESCE(content_data, '{}'::jsonb),
				  COALESCE(company_name, ''), COALESCE(tagline, ''),
				  COALESCE(email, ''), COALESCE(phone, ''),
				  COALESCE(logo_text, ''), COALESCE(github_repo, '')
	`

	var site SiteRecord
	var contentDataJSON []byte

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, domain, networkID).Scan(
			&site.ID, &site.NetworkID, &site.Domain, &site.Name, &site.Status, &site.CreatedAt,
			&contentDataJSON,
			&site.CompanyName, &site.Tagline, &site.Email, &site.Phone, &site.LogoText,
			&site.GithubRepo,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert site: %w", err)
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, domain, networkID).Scan(
			&site.ID, &site.NetworkID, &site.Domain, &site.Name, &site.Status, &site.CreatedAt,
			&contentDataJSON,
			&site.CompanyName, &site.Tagline, &site.Email, &site.Phone, &site.LogoText,
			&site.GithubRepo,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert site: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	// Parse content_data JSON
	if len(contentDataJSON) > 0 && string(contentDataJSON) != "null" {
		site.ContentData = make(map[string]interface{})
		if err := json.Unmarshal(contentDataJSON, &site.ContentData); err != nil {
			logger.Warn("Failed to parse content_data", zap.Error(err))
		}
	}

	return &site, nil
}

func upsertPage(ctx context.Context, db interface{}, siteID uuid.UUID, page map[string]interface{}, index int, planID uuid.NullUUID, allowEmptySections bool, logger *zap.Logger) (*PageRecord, error) {
	// Extract page fields with defaults
	name := datahelpers.GetStringField(page, "name", fmt.Sprintf("page-%d", index))
	title := datahelpers.GetStringField(page, "title", name)
	url := datahelpers.GetStringField(page, "url", "")
	if url == "" {
		if name == "index" {
			url = "/index.html"
		} else {
			url = "/" + name + ".html"
		}
	}
	pageType := datahelpers.GetStringField(page, "type", "content")
	navLabel := datahelpers.GetStringField(page, "nav_label", title)
	navOrder := datahelpers.GetIntField(page, "nav_order", (index+1)*10)
	inHeader := datahelpers.GetBoolField(page, "in_header", true)
	inFooter := datahelpers.GetBoolField(page, "in_footer", true)
	metaDescription := datahelpers.GetStringField(page, "meta_description", "")

	// Extract and serialize sections array from site plan
	var sectionsJSON []byte
	if sections, ok := page["sections"].([]interface{}); ok && len(sections) > 0 {
		sectionsJSON, _ = json.Marshal(sections)
	} else if sections, ok := page["sections"].([]string); ok && len(sections) > 0 {
		sectionsJSON, _ = json.Marshal(sections)
	} else {
		sectionsJSON = []byte("[]")
	}

	// Special handling for privacy/terms pages
	if strings.Contains(name, "privacy") || strings.Contains(name, "terms") {
		inHeader = false
		inFooter = true
	}

	// build_status / built_from_plan_version behaviour on conflict:
	//   - built_from_plan_version is filled only when the existing row has none
	//     (COALESCE existing-first). The authoritative stamp is written at deploy
	//     time by UpdatePageStatusAction; sync must not overwrite it, or drift
	//     detection across re-plans breaks (a v1-built page would look v2-current).
	//     Pre-plan deploys (tool-recreation) arrive with NULL and are adopted into
	//     the current plan here.
	//   - build_status is NOT flipped deployed->needs_rebuild. Rebuild-on-plan-change
	//     is the reconciler's job via drift detection (built_from_plan_version !=
	//     current plan id; 029/030 item 9, decideEmit). The old flip fired on every
	//     deployed page on every sync and churned pages deployed before the plan
	//     existed (tools). NULL is normalised to 'planned'; other states are left
	//     intact for the reconciler / dispatch to act on.
	//   - meta_description is COALESCE(NULLIF(EXCLUDED,''), existing) — an incoming
	//     BLANK never destroys a description the page already has (bugs_open/320).
	//     It used to be a bare `= EXCLUDED.meta_description`, three lines below a
	//     nav_label clause that WAS guarded, which is what made the asymmetry read
	//     as deliberate.
	//
	//     ⚠ CORRECTED 2026-08-19 (council round 2, editquality): this comment and
	//     bugs_open/320 both said the new clause "matches nav_label three lines
	//     above". IT IS THE MIRROR IMAGE, and the difference is the whole policy:
	//         nav_label        = COALESCE(NULLIF(pages.nav_label,''), EXCLUDED…)
	//                            -> the EXISTING value wins; effectively write-once.
	//         meta_description = COALESCE(NULLIF(EXCLUDED…,''), pages.meta_description)
	//                            -> the INCOMING value wins unless it is blank.
	//     Both are deliberate and they are deliberately different. A description is
	//     content the plan owns and should be able to REVISE, so a real incoming
	//     value must win; only a blank is refused. A nav label, once a human or an
	//     earlier build has set it, should not be churned by a replan. Do not
	//     "make them consistent" — that would either freeze descriptions for ever
	//     or hand nav labels back to every replan.
	//
	//     The underlying defect was not the clause shape but the input: metaDescription above defaults to "" whenever
	//     the incoming page map omits the key, and build-site-planner's page object
	//     never carried the key at all — so every replan of an existing page wrote
	//     a blank over whatever was there. Measured 2026-08-19: four robot-hands.com
	//     pages holding 97/120/169/329 chars in an April site_snapshot read 0 today.
	//     A NON-blank incoming value still wins, so a plan that DOES supply one
	//     continues to update the page; only the destructive direction is closed.
	query := `
		INSERT INTO pages (site_id, name, url, title, page_type, nav_label, nav_order, in_header, in_footer, meta_description, sections, build_status, status, built_from_plan_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'planned', 'active', $12)
		ON CONFLICT (site_id, name) DO UPDATE SET
			url = EXCLUDED.url,
			title = EXCLUDED.title,
			page_type = EXCLUDED.page_type,
			nav_label = COALESCE(NULLIF(pages.nav_label, ''), EXCLUDED.nav_label),
			nav_order = EXCLUDED.nav_order,
			in_header = EXCLUDED.in_header,
			in_footer = EXCLUDED.in_footer,
			meta_description = COALESCE(NULLIF(EXCLUDED.meta_description, ''), pages.meta_description),
			-- bugs_open/204: an EMPTY incoming list may not overwrite a real one
			-- unless the caller declares the emptying deliberate ($13).
			--
			-- This is the same asymmetry that licensed the nav_label and
			-- meta_description guards two clauses up (added 2026-08-19 after blank
			-- overwrites were measured on robot-hands.com), and sections is the
			-- worse case of the three: pages.sections is the ONLY record of a
			-- decomposed page's composition. page_components keeps serving after it
			-- is emptied, so nothing looks wrong — and the next rebuild has nothing
			-- to rebuild and builds an empty page over a live one. Measured
			-- 2026-08-20: one replan emptied 41 of 45 live pages this way and queued
			-- 20 needs_page items against them.
			--
			-- The plan stays authoritative for every NON-empty proposal, so a
			-- recomposition still wins — exactly as meta_description's guard lets a
			-- real value win and refuses only a blank. Only the non-empty -> empty
			-- transition is intercepted. Deliberate emptying travels through the
			-- recompose_pages release, the channel that already means "redesign this
			-- page", so no operator has to remember a new flag.
			sections = CASE
				WHEN $13::bool THEN EXCLUDED.sections
				WHEN COALESCE(jsonb_array_length(EXCLUDED.sections), 0) > 0 THEN EXCLUDED.sections
				WHEN COALESCE(jsonb_array_length(pages.sections), 0) = 0 THEN EXCLUDED.sections
				ELSE pages.sections
			END,
			built_from_plan_version = COALESCE(pages.built_from_plan_version, EXCLUDED.built_from_plan_version),
			build_status = CASE
				WHEN pages.build_status IS NULL THEN 'planned'
				ELSE pages.build_status
			END,
			updated_at = NOW()
		RETURNING id, site_id, name, url, title, page_type, nav_label, nav_order, in_header, in_footer, status,
		          (COALESCE(jsonb_array_length(sections), 0) > 0
		             AND COALESCE(jsonb_array_length($11::jsonb), 0) = 0) AS sections_kept
	`

	var pageRecord PageRecord

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query,
			siteID, name, url, title, pageType, navLabel, navOrder, inHeader, inFooter, metaDescription, sectionsJSON, planID, allowEmptySections,
		).Scan(
			&pageRecord.ID, &pageRecord.SiteID, &pageRecord.Name, &pageRecord.URL,
			&pageRecord.Title, &pageRecord.PageType, &pageRecord.NavLabel, &pageRecord.NavOrder,
			&pageRecord.InHeader, &pageRecord.InFooter, &pageRecord.Status, &pageRecord.SectionsKept,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert page: %w", err)
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query,
			siteID, name, url, title, pageType, navLabel, navOrder, inHeader, inFooter, metaDescription, sectionsJSON, planID, allowEmptySections,
		).Scan(
			&pageRecord.ID, &pageRecord.SiteID, &pageRecord.Name, &pageRecord.URL,
			&pageRecord.Title, &pageRecord.PageType, &pageRecord.NavLabel, &pageRecord.NavOrder,
			&pageRecord.InHeader, &pageRecord.InFooter, &pageRecord.Status, &pageRecord.SectionsKept,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert page: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	logger.Debug("Page upserted",
		zap.String("page_id", pageRecord.ID.String()),
		zap.String("name", pageRecord.Name),
		zap.Int("sections_bytes", len(sectionsJSON)),
	)

	return &pageRecord, nil
}

// DEPRECATED
func getNavigationFromDB(ctx context.Context, db interface{}, siteID uuid.UUID, navType string, logger *zap.Logger) (*NavigationStructure, error) {
	// First try to get cached navigation
	cacheQuery := `
		SELECT structure FROM navigation_structures
		WHERE site_id = $1 AND nav_type = $2 AND is_current = true
	`

	var structureJSON []byte

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, cacheQuery, siteID, navType).Scan(&structureJSON)
		if err == nil && len(structureJSON) > 0 {
			var nav NavigationStructure
			if json.Unmarshal(structureJSON, &nav) == nil {
				return &nav, nil
			}
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, cacheQuery, siteID, navType).Scan(&structureJSON)
		if err == nil && len(structureJSON) > 0 {
			var nav NavigationStructure
			if json.Unmarshal(structureJSON, &nav) == nil {
				return &nav, nil
			}
		}
	}

	// Cache miss or invalid - build from pages
	logger.Info("Navigation cache miss, building from pages")
	return buildNavigationFromDB(ctx, db, siteID, navType, logger)
}

// DEPRECATED
func buildNavigationFromDB(ctx context.Context, db interface{}, siteID uuid.UUID, navType string, logger *zap.Logger) (*NavigationStructure, error) {
	query := `
		SELECT id, COALESCE(nav_label, title, name) as label, url
		FROM pages
		WHERE site_id = $1 
		  AND status = 'active'
		  AND CASE WHEN $2 = 'header' THEN in_header WHEN $2 = 'footer' THEN in_footer ELSE true END
		ORDER BY nav_order, name
	`

	nav := &NavigationStructure{Items: []NavigationItem{}}

	switch d := db.(type) {
	case *sql.DB:
		rows, err := d.QueryContext(ctx, query, siteID, navType)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var item NavigationItem
			if err := rows.Scan(&item.PageID, &item.Label, &item.URL); err != nil {
				logger.Warn("Failed to scan page row", zap.Error(err))
				continue
			}
			nav.Items = append(nav.Items, item)
		}
	case *pgxpool.Pool:
		rows, err := d.Query(ctx, query, siteID, navType)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var item NavigationItem
			if err := rows.Scan(&item.PageID, &item.Label, &item.URL); err != nil {
				logger.Warn("Failed to scan page row", zap.Error(err))
				continue
			}
			nav.Items = append(nav.Items, item)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	// Cache the result
	cacheNavigation(ctx, db, siteID, navType, nav, logger)

	return nav, nil
}

// DEPRECATED
func cacheNavigation(ctx context.Context, db interface{}, siteID uuid.UUID, navType string, nav *NavigationStructure, logger *zap.Logger) {
	structureJSON, err := json.Marshal(nav)
	if err != nil {
		logger.Warn("Failed to marshal navigation for cache", zap.Error(err))
		return
	}

	query := `
		INSERT INTO navigation_structures (site_id, nav_type, structure, is_current)
		VALUES ($1, $2, $3, true)
		ON CONFLICT (site_id, nav_type, version) DO UPDATE SET
			structure = EXCLUDED.structure,
			is_current = true
	`

	switch d := db.(type) {
	case *sql.DB:
		_, err = d.ExecContext(ctx, query, siteID, navType, structureJSON)
	case *pgxpool.Pool:
		_, err = d.Exec(ctx, query, siteID, navType, structureJSON)
	}

	if err != nil {
		logger.Warn("Failed to cache navigation", zap.Error(err))
	}
}

func buildNavigationFromPages(pages []map[string]interface{}) *NavigationStructure {
	nav := &NavigationStructure{Items: []NavigationItem{}}

	for i, page := range pages {
		name := datahelpers.GetStringField(page, "name", fmt.Sprintf("page-%d", i))
		title := datahelpers.GetStringField(page, "title", name)
		url := datahelpers.GetStringField(page, "url", "")
		if url == "" {
			if name == "index" || name == "home" {
				url = "/index.html"
			} else {
				url = "/" + name + ".html"
			}
		}
		navLabel := datahelpers.GetStringField(page, "nav_label", title)
		inHeader := datahelpers.GetBoolField(page, "in_header", true)

		// Skip non-header pages for navigation
		if !inHeader {
			continue
		}

		// Skip privacy/terms from header
		nameLower := strings.ToLower(name)
		if strings.Contains(nameLower, "privacy") || strings.Contains(nameLower, "terms") {
			continue
		}

		// Simplify the nav label - remove verbose descriptions
		simpleLabel := datahelpers.SimplifyNavLabel(navLabel, name)

		nav.Items = append(nav.Items, NavigationItem{
			Label: simpleLabel,
			URL:   url,
		})
	}

	return nav
}

func getPageID(ctx context.Context, db interface{}, siteID uuid.UUID, pageName string, logger *zap.Logger) (uuid.UUID, error) {
	query := `SELECT id FROM pages WHERE site_id = $1 AND name = $2`

	var pageID uuid.UUID

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, siteID, pageName).Scan(&pageID)
		if err != nil {
			return uuid.Nil, err
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, siteID, pageName).Scan(&pageID)
		if err != nil {
			return uuid.Nil, err
		}
	default:
		return uuid.Nil, fmt.Errorf("unsupported database type: %T", db)
	}

	return pageID, nil
}

func updateSiteTimestamps(ctx context.Context, db interface{}, siteID uuid.UUID, timestamp time.Time, logger *zap.Logger) error {
	query := `
		UPDATE sites 
		SET last_built_at = $2, last_deployed_at = $2, updated_at = $2
		WHERE id = $1
	`

	switch d := db.(type) {
	case *sql.DB:
		_, err := d.ExecContext(ctx, query, siteID, timestamp)
		return err
	case *pgxpool.Pool:
		_, err := d.Exec(ctx, query, siteID, timestamp)
		return err
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

// ============================================================================
// HELPER FUNCTIONS - Link Extraction
// ============================================================================

func extractLinksFromHTML(htmlContent string, siteID uuid.UUID, pageName string, logger *zap.Logger) []LinkRegistryEntry {
	var links []LinkRegistryEntry

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		logger.Warn("Failed to parse HTML for link extraction", zap.Error(err))
		return links
	}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		// Skip javascript: and mailto: links
		if strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
			return
		}

		anchorText := strings.TrimSpace(s.Text())
		rel, _ := s.Attr("rel")

		link := LinkRegistryEntry{
			SourceSiteID: siteID,
			TargetURL:    href,
			AnchorText:   anchorText,
			RelAttr:      rel,
			Scope:        classifyLinkScope(href),
			LinkType:     classifyLinkTypeFromContext(s),
			Status:       "active",
		}

		links = append(links, link)
	})

	logger.Info("Links extracted from HTML",
		zap.Int("count", len(links)),
		zap.String("page", pageName),
	)

	return links
}

func classifyLinkScope(href string) string {
	if strings.HasPrefix(href, "#") {
		return "internal" // anchor within page
	}
	if strings.HasPrefix(href, "/") {
		return "page" // relative, same site
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return "external"
	}
	// Assume relative URL
	return "page"
}

func classifyLinkTypeFromContext(s *goquery.Selection) string {
	// Check parent elements for hints
	parent := s.Parent()
	for i := 0; i < 5 && parent.Length() > 0; i++ {
		tagName := goquery.NodeName(parent)
		class, _ := parent.Attr("class")

		// Navigation elements
		if tagName == "nav" || tagName == "header" || tagName == "footer" {
			return "navigation"
		}
		if strings.Contains(class, "nav") || strings.Contains(class, "menu") {
			return "navigation"
		}

		// CTA elements
		if strings.Contains(class, "cta") || strings.Contains(class, "button") {
			return "content"
		}

		// Related content
		if strings.Contains(class, "related") || strings.Contains(class, "similar") {
			return "semantic"
		}

		parent = parent.Parent()
	}

	// Check the link itself
	class, _ := s.Attr("class")
	if strings.Contains(class, "btn") || strings.Contains(class, "button") || strings.Contains(class, "cta") {
		return "content"
	}

	// Default to content
	return "content"
}

func syncLinksToDB(ctx context.Context, db interface{}, siteID uuid.UUID, pageID uuid.UUID, links []LinkRegistryEntry, logger *zap.Logger) (int, error) {
	// Delete existing links for this page
	deleteQuery := `DELETE FROM link_registry WHERE source_page_id = $1`

	switch d := db.(type) {
	case *sql.DB:
		_, err := d.ExecContext(ctx, deleteQuery, pageID)
		if err != nil {
			return 0, fmt.Errorf("failed to delete existing links: %w", err)
		}
	case *pgxpool.Pool:
		_, err := d.Exec(ctx, deleteQuery, pageID)
		if err != nil {
			return 0, fmt.Errorf("failed to delete existing links: %w", err)
		}
	default:
		return 0, fmt.Errorf("unsupported database type: %T", db)
	}

	// Insert new links
	insertQuery := `
		INSERT INTO link_registry (
			source_page_id, source_site_id, target_url,
			scope, link_type, anchor_text, rel_attr, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	syncedCount := 0
	for _, link := range links {
		var err error
		switch d := db.(type) {
		case *sql.DB:
			_, err = d.ExecContext(ctx, insertQuery,
				pageID, siteID, link.TargetURL,
				link.Scope, link.LinkType, link.AnchorText, link.RelAttr, link.Status,
			)
		case *pgxpool.Pool:
			_, err = d.Exec(ctx, insertQuery,
				pageID, siteID, link.TargetURL,
				link.Scope, link.LinkType, link.AnchorText, link.RelAttr, link.Status,
			)
		}

		if err != nil {
			logger.Warn("Failed to insert link", zap.String("url", link.TargetURL), zap.Error(err))
			continue
		}
		syncedCount++
	}

	return syncedCount, nil
}

// ============================================================================
// HELPER FUNCTIONS - Field Extraction
// ============================================================================

// computeContentHash generates MD5 hash of content for change detection
func computeContentHash(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

// isValidURL checks if a string is a valid URL
func isValidURL(str string) bool {
	// Simple validation - starts with / or http
	if strings.HasPrefix(str, "/") || strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://") {
		return true
	}
	// Check for relative URLs
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)*\.html?$`, str)
	return matched
}
