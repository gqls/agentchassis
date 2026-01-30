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
	domain := extractDomainFromInput(params.CollectedData, params.Logger)
	if domain == "" {
		params.Logger.Error("EnsureSiteRecordAction: Domain not found")
		return nil, fmt.Errorf("domain not found in input_data")
	}

	// Clean domain (remove protocol, trailing slashes)
	domain = cleanDomain(domain)
	params.Logger.Info("EnsureSiteRecordAction: Domain cleaned",
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

	return map[string]interface{}{
		"site_id":      siteRecord.ID.String(),
		"domain":       siteRecord.Domain,
		"content_data": siteRecord.ContentData,
		"network_id":   siteRecord.NetworkID.String(),
		"status":       siteRecord.Status,
		"created":      siteRecord.CreatedAt.Format(time.RFC3339),
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

	// Sync each page to database
	syncedCount := 0
	for i, page := range pages {
		pageRecord, err := upsertPage(ctx, params.DB, siteID, page, i, params.Logger)
		if err != nil {
			params.Logger.Error("Failed to upsert page",
				zap.Any("page", page),
				zap.Error(err))
			continue
		}
		syncedCount++
		params.Logger.Debug("Page synced",
			zap.String("page_id", pageRecord.ID.String()),
			zap.String("name", pageRecord.Name),
		)
	}

	// Get navigation from database (trigger will have invalidated cache)
	navigation, err := getNavigationFromDB(ctx, params.DB, siteID, "header", params.Logger)
	if err != nil {
		params.Logger.Warn("Failed to get navigation from DB, building from plan",
			zap.Error(err))
		navigation = buildNavigationFromPages(pages)
	}

	params.Logger.Info("SyncPagesToDBAction: Complete",
		zap.Int("pages_synced", syncedCount),
		zap.Int("nav_items", len(navigation.Items)),
	)

	return map[string]interface{}{
		"pages_synced": syncedCount,
		"navigation":   navigation,
		"site_id":      siteIDStr,
		"db_available": true,
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

	return map[string]interface{}{
		"links_extracted": len(links),
		"links_persisted": syncedCount,
		"page_id":         pageID.String(),
		"persisted":       true,
	}, nil
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

	navigation, err := getNavigationFromDB(ctx, params.DB, siteID, navType, params.Logger)
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
		RETURNING id, network_id, domain, name, status, created_at, COALESCE(content_data, '{}'::jsonb)
	`

	var site SiteRecord
	var contentDataJSON []byte // NEW: to scan JSONB

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, domain, networkID).Scan(
			&site.ID, &site.NetworkID, &site.Domain, &site.Name, &site.Status, &site.CreatedAt,
			&contentDataJSON, // NEW
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert site: %w", err)
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, domain, networkID).Scan(
			&site.ID, &site.NetworkID, &site.Domain, &site.Name, &site.Status, &site.CreatedAt,
			&contentDataJSON, // NEW
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

func upsertPage(ctx context.Context, db interface{}, siteID uuid.UUID, page map[string]interface{}, index int, logger *zap.Logger) (*PageRecord, error) {
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

	query := `
		INSERT INTO pages (site_id, name, url, title, page_type, nav_label, nav_order, in_header, in_footer, meta_description, sections, build_status, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'planned', 'active')
		ON CONFLICT (site_id, name) DO UPDATE SET
			url = EXCLUDED.url,
			title = EXCLUDED.title,
			page_type = EXCLUDED.page_type,
			nav_label = COALESCE(NULLIF(pages.nav_label, ''), EXCLUDED.nav_label),
			nav_order = EXCLUDED.nav_order,
			in_header = EXCLUDED.in_header,
			in_footer = EXCLUDED.in_footer,
			meta_description = EXCLUDED.meta_description,
			sections = EXCLUDED.sections,
			build_status = CASE 
				WHEN pages.build_status = 'deployed' THEN 'needs_rebuild'
				WHEN pages.build_status IS NULL THEN 'planned'
				ELSE pages.build_status
			END,
			updated_at = NOW()
		RETURNING id, site_id, name, url, title, page_type, nav_label, nav_order, in_header, in_footer, status
	`

	var pageRecord PageRecord

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query,
			siteID, name, url, title, pageType, navLabel, navOrder, inHeader, inFooter, metaDescription, sectionsJSON, // Added sectionsJSON
		).Scan(
			&pageRecord.ID, &pageRecord.SiteID, &pageRecord.Name, &pageRecord.URL,
			&pageRecord.Title, &pageRecord.PageType, &pageRecord.NavLabel, &pageRecord.NavOrder,
			&pageRecord.InHeader, &pageRecord.InFooter, &pageRecord.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert page: %w", err)
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query,
			siteID, name, url, title, pageType, navLabel, navOrder, inHeader, inFooter, metaDescription, sectionsJSON, // Added sectionsJSON
		).Scan(
			&pageRecord.ID, &pageRecord.SiteID, &pageRecord.Name, &pageRecord.URL,
			&pageRecord.Title, &pageRecord.PageType, &pageRecord.NavLabel, &pageRecord.NavOrder,
			&pageRecord.InHeader, &pageRecord.InFooter, &pageRecord.Status,
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
