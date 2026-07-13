// FILE: platform/orchestration/actions/queryresolve/section_index_for.go
//
// section_index_for:<type> — resolves the hub (section-index) page URL for a
// given page type, from real page relationships. Used by the list components
// (tool-list, game-list, guide-list) for their "Browse All X" button, so the
// hub link is derived from the actual pages — not a hardcoded /<area>/index.html
// or an unpopulated *_index_url spec.
//
// Returns a single URL string (interface{}), or nil when no hub exists. nil
// matters: the plan_sections field loop assigns any non-nil value, so returning
// "" would write href="" — a nil leaves the field unresolved and the (gated)
// template renders no button.
//
// Add to queryresolve.Resolve's switch:
//     case "section_index_for":
//         return resolveSectionIndexForType(ctx, db, req.SiteID, arg, logger)

package queryresolve

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// resolveSectionIndexForType returns the URL of the section-index page that is
// the hub for pages of the given type (e.g. "tool" -> the tools hub). Strategy:
//  1. Prefer the section-index page sharing a site_area with the typed pages.
//  2. Fall back to URL-prefix matching (typed pages live under /<area>/...; the
//     section-index at /<area>/index.html is the hub) when site_area_id is unset.
//
// Returns (url, nil) when found, (nil, nil) when no hub exists, (nil, err) on a
// real query error.
func resolveSectionIndexForType(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageType string,
	logger *zap.Logger,
) (interface{}, error) {
	if pageType == "" {
		return nil, fmt.Errorf("resolveSectionIndexForType: empty type argument")
	}

	// 1) Shared-area lookup: the section-index page whose site_area matches the
	//    area of the typed pages.
	var url string
	err := db.QueryRowContext(ctx, `
		SELECT idx.url
		FROM pages idx
		WHERE idx.site_id   = $1
		  AND idx.page_type = 'section-index'
		  AND idx.status    IN ('active', 'deployed')
		  AND idx.site_area_id = (
		      SELECT t.site_area_id FROM pages t
		      WHERE t.site_id = $1
		        AND t.page_type = $2
		        AND t.site_area_id IS NOT NULL
		        AND t.status IN ('active', 'deployed')
		      LIMIT 1
		  )
		ORDER BY COALESCE(idx.nav_order, 100)
		LIMIT 1
	`, siteID, pageType).Scan(&url)
	if err == nil && url != "" {
		logger.Info("queryresolve: resolved section_index_for (by area)",
			zap.String("type", pageType), zap.String("url", url),
			zap.String("site_id", siteID.String()))
		return url, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("resolveSectionIndexForType area lookup failed: %w", err)
	}

	// 2) URL-prefix fallback: the section-index at /<first-segment>/index.html,
	//    where the segment is taken from a typed page's URL (e.g. /tools/x/...).
	err = db.QueryRowContext(ctx, `
		SELECT idx.url
		FROM pages idx
		WHERE idx.site_id   = $1
		  AND idx.page_type = 'section-index'
		  AND idx.status    IN ('active', 'deployed')
		  AND idx.url = '/' || split_part(
		        ltrim((SELECT t.url FROM pages t
		               WHERE t.site_id = $1
		                 AND t.page_type = $2
		                 AND t.status IN ('active', 'deployed')
		               LIMIT 1), '/'),
		        '/', 1) || '/index.html'
		LIMIT 1
	`, siteID, pageType).Scan(&url)
	if err == sql.ErrNoRows {
		logger.Info("queryresolve: section_index_for found no hub",
			zap.String("type", pageType), zap.String("site_id", siteID.String()))
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("resolveSectionIndexForType prefix lookup failed: %w", err)
	}
	if url == "" {
		return nil, nil
	}

	logger.Info("queryresolve: resolved section_index_for (by url prefix)",
		zap.String("type", pageType), zap.String("url", url),
		zap.String("site_id", siteID.String()))
	return url, nil
}
