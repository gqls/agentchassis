// FILE: platform/orchestration/actions/navigation_from_pages.go
// Functions to build navigation from pages table (deployed pages only)
// Used by InjectHeader and InjectFooter to ensure nav matches reality

package actions

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NavItemFromDB represents a navigation item from pages table
type NavItemFromDB struct {
	Label    string
	URL      string
	NavOrder int
	InHeader bool
	InFooter bool
}

// GetHeaderNavFromPages queries pages table for header navigation
// Only includes pages with in_header=true AND status in deployed/active
func GetHeaderNavFromPages(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxItems int, logger *zap.Logger) []NavItem {
	if db == nil || siteID == uuid.Nil {
		logger.Debug("GetHeaderNavFromPages: No DB or site_id, returning empty nav")
		return []NavItem{}
	}

	if maxItems <= 0 {
		maxItems = 6 // Default max header items
	}

	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url,
			COALESCE(nav_order, position, 0) as nav_order
		FROM pages 
		WHERE site_id = $1 
		  AND in_header = true
		  AND status IN ('deployed', 'active')
		  AND deleted_at IS NULL
		ORDER BY nav_order ASC, created_at ASC
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, siteID, maxItems)
	if err != nil {
		logger.Warn("GetHeaderNavFromPages: Query failed", zap.Error(err))
		return []NavItem{}
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		var navOrder int
		if err := rows.Scan(&label, &url, &navOrder); err != nil {
			logger.Warn("GetHeaderNavFromPages: Scan failed", zap.Error(err))
			continue
		}

		// Clean up verbose labels
		label = simplifyNavLabel(label, extractNameFromURL(url))

		items = append(items, NavItem{
			Label: label,
			URL:   url,
		})
	}

	logger.Debug("GetHeaderNavFromPages: Built nav",
		zap.Int("items", len(items)),
		zap.String("site_id", siteID.String()),
	)

	return items
}

// GetFooterNavFromPages queries pages table for footer navigation
// Includes pages with in_footer=true OR legal pages (privacy, terms)
func GetFooterNavFromPages(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) []NavItem {
	if db == nil || siteID == uuid.Nil {
		return []NavItem{}
	}

	query := `
		SELECT 
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url,
			COALESCE(nav_order, position, 0) as nav_order
		FROM pages 
		WHERE site_id = $1 
		  AND (in_footer = true OR LOWER(name) LIKE '%privacy%' OR LOWER(name) LIKE '%terms%')
		  AND status IN ('deployed', 'active')
		  AND deleted_at IS NULL
		ORDER BY 
			CASE WHEN LOWER(name) LIKE '%privacy%' OR LOWER(name) LIKE '%terms%' THEN 1 ELSE 0 END,
			nav_order ASC
	`

	rows, err := db.QueryContext(ctx, query, siteID)
	if err != nil {
		logger.Warn("GetFooterNavFromPages: Query failed", zap.Error(err))
		return []NavItem{}
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		var navOrder int
		if err := rows.Scan(&label, &url, &navOrder); err != nil {
			continue
		}

		label = simplifyNavLabel(label, extractNameFromURL(url))
		items = append(items, NavItem{Label: label, URL: url})
	}

	return items
}

// simplifyNavLabel cleans up verbose labels
func simplifyNavLabel(label, name string) string {
	// If already short, keep it
	if len(label) <= 15 {
		return label
	}

	// Standard mappings
	nameLower := strings.ToLower(name)
	switch {
	case nameLower == "index" || nameLower == "home":
		return "Home"
	case strings.HasPrefix(nameLower, "about"):
		return "About"
	case strings.HasPrefix(nameLower, "service"):
		return "Services"
	case strings.HasPrefix(nameLower, "contact"):
		return "Contact"
	case strings.HasPrefix(nameLower, "case") || strings.HasPrefix(nameLower, "portfolio") || strings.HasPrefix(nameLower, "work"):
		return "Work"
	case strings.HasPrefix(nameLower, "team") || strings.HasPrefix(nameLower, "leadership"):
		return "Team"
	case strings.HasPrefix(nameLower, "privacy"):
		return "Privacy"
	case strings.HasPrefix(nameLower, "terms"):
		return "Terms"
	}

	// Take first word if still long
	if len(label) > 20 {
		words := strings.Fields(label)
		if len(words) > 0 {
			return words[0]
		}
	}

	return label
}

// extractNameFromURL gets page name from URL
func extractNameFromURL(url string) string {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, ".html")
	return url
}

// UpdateRenderContextNavFromPages updates RenderContext nav items from pages table
// Call this before header/footer injection to ensure nav is current
func UpdateRenderContextNavFromPages(ctx context.Context, db *sql.DB, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) {
	if db == nil || siteID == uuid.Nil {
		return
	}

	sqlDB, ok := db.(*sql.DB)
	if !ok {
		logger.Warn("UpdateRenderContextNavFromPages: DB is not *sql.DB")
		return
	}

	// Get header nav
	headerNav := GetHeaderNavFromPages(ctx, sqlDB, siteID, 6, logger)
	if len(headerNav) > 0 {
		renderCtx.NavItems = headerNav
	}

	// Get footer nav (store separately if RenderContext supports it)
	// For now, footer template can query directly
}

// ============================================================
// PATCH for InjectHeader - add before the existing function
// ============================================================
//
// To use this, update InjectHeader to call:
//   UpdateRenderContextNavFromPages(ctx, db.(*sql.DB), siteID, renderCtx, logger)
// before rendering the header template.
//
// Example modification to InjectHeader:
//
// func InjectHeader(ctx context.Context, db interface{}, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string {
//     // NEW: Update nav from deployed pages
//     if sqlDB, ok := db.(*sql.DB); ok && siteID != uuid.Nil {
//         UpdateRenderContextNavFromPages(ctx, sqlDB, siteID, renderCtx, logger)
//     }
//
//     headerHTML, err := RenderHeader(ctx, db, siteID, renderCtx, logger)
//     // ... rest of function
// }
