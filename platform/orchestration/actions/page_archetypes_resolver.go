// FILE: platform/orchestration/actions/page_archetypes_resolver.go
//
// sectionsForPage replaces defaultSectionsForPage (apply_gap_plan_action.go)
// as the primary source of a page's default section list, backed by the
// page_archetypes table (689_theme_kits.sql) instead of a hardcoded Go
// switch. defaultSectionsForPage itself is KEPT — as a logged last-resort
// fallback for a (name, type) shape no seed row covers yet — so a missing
// seed can never break a build. Every fallback hit is logged at Warn so the
// fallback's remaining usage is visible and it can be retired once seed
// coverage is proven complete.
//
// Resolution order, first match wins:
//  1. A site-scoped page_archetypes row (site_id = this site)
//  2. The site's current theme kit's rows (theme_kit_id, via
//     theme_kit_adoption — loadSiteThemeKitDefaults)
//  3. A fleet-scoped row (theme_kit_id IS NULL AND site_id IS NULL)
//  4. defaultSectionsForPage (Go fallback, logged)
//
// Within a scope, match precedence mirrors the Go switch it replaces
// (bugs_open/015: page_type outranks page_name — a type is not localised,
// a name can be): page_type exact -> page_name exact ->
// page_name_contains -> page_name_suffix -> default.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// sectionsForPage is the drop-in replacement for defaultSectionsForPage's
// call sites. db may be nil in a context where no page_archetypes lookup is
// possible (defensive only — every real caller has a live *sql.DB); in that
// case it falls straight to the Go switch.
func sectionsForPage(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, pageType string, logger *zap.Logger) []string {
	if db != nil {
		if rows, ok := pageArchetypeRows(ctx, db, siteID, nil, logger); ok {
			if sections, matched := matchArchetypeRows(rows, pageName, pageType); matched {
				return sections
			}
		}
		kit, ok, kerr := loadSiteThemeKitDefaults(ctx, db, siteID)
		if kerr != nil && logger != nil {
			// Not fatal — the kit scope is skipped and the fleet row still
			// serves. Logged because a swallowed error here means the kit has
			// silently stopped steering this site's structure and nothing else
			// would ever say so.
			logger.Warn("sectionsForPage: theme-kit lookup failed — kit scope skipped",
				zap.Error(kerr), zap.String("site_id", siteID.String()))
		}
		if kerr == nil && ok {
			if rows, ok := pageArchetypeRows(ctx, db, uuid.Nil, &kit.ThemeKitID, logger); ok {
				if sections, matched := matchArchetypeRows(rows, pageName, pageType); matched {
					return sections
				}
			}
		}
		if rows, ok := pageArchetypeRows(ctx, db, uuid.Nil, nil, logger); ok {
			if sections, matched := matchArchetypeRows(rows, pageName, pageType); matched {
				return sections
			}
		}
	}
	if logger != nil {
		logger.Warn("sectionsForPage: no page_archetypes row matched — served by the legacy Go fallback",
			zap.String("page_name", pageName), zap.String("page_type", pageType))
	}
	return defaultSectionsForPage(pageName, pageType)
}

type archetypeRow struct {
	MatchKind  string
	MatchValue string
	Sections   []string
}

// pageArchetypeRows loads one scope's active rows. Pass a real siteID for
// the site scope, or uuid.Nil + a non-nil themeKitID for the kit scope, or
// uuid.Nil + nil for the fleet scope. The second bool return is false only
// on a query error (treated as "no rows in this scope", not a hard failure
// — a page_archetypes outage must not break a page build).
func pageArchetypeRows(ctx context.Context, db *sql.DB, siteID uuid.UUID, themeKitID *uuid.UUID, logger *zap.Logger) ([]archetypeRow, bool) {
	// ORDER BY is load-bearing, not tidiness. Two rows in one scope can both
	// match (page_name_contains 'faq' and 'pricing' both match
	// "pricing-faq-guide"), and the Go switch this replaces was DETERMINISTIC
	// about it — first case in source order wins, faq before pricing. Without
	// an ORDER BY the resolver would serve whichever row the heap returned,
	// which can change after a VACUUM with no code change at all. Longest
	// match_value first is the defensible rule (the more specific pattern
	// wins), then oldest, then id as a total tiebreak so the result is stable.
	const selectCols = `SELECT match_kind, match_value, sections FROM page_archetypes `
	const orderBy = ` ORDER BY length(match_value) DESC, created_at ASC, id ASC`
	var query string
	var args []interface{}
	switch {
	case siteID != uuid.Nil:
		query = selectCols + `WHERE site_id = $1 AND is_active = true` + orderBy
		args = []interface{}{siteID}
	case themeKitID != nil:
		query = selectCols + `WHERE theme_kit_id = $1 AND is_active = true` + orderBy
		args = []interface{}{*themeKitID}
	default:
		query = selectCols + `WHERE theme_kit_id IS NULL AND site_id IS NULL AND is_active = true` + orderBy
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		if logger != nil {
			logger.Warn("pageArchetypeRows: query failed — this scope will be skipped",
				zap.Error(err), zap.String("scope", archetypeScopeName(siteID, themeKitID)))
		}
		return nil, false
	}
	defer rows.Close()

	var out []archetypeRow
	offered := 0
	for rows.Next() {
		offered++
		var r archetypeRow
		var sectionsJSON []byte
		if err := rows.Scan(&r.MatchKind, &r.MatchValue, &sectionsJSON); err != nil {
			// scan-loss:accepted: counted — ScanShortfall below refuses the
			// whole scope rather than silently matching against a thinned
			// row set that might be missing exactly the row that would
			// have matched this page (bugs_open/410's failure shape).
			continue
		}
		var sections []string
		if err := json.Unmarshal(sectionsJSON, &sections); err != nil || len(sections) == 0 {
			// scan-loss:accepted: same reasoning — a malformed sections
			// payload is a lost row, counted the same way.
			continue
		}
		r.Sections = sections
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, false
	}
	if shortfall := datahelpers.ScanShortfall(offered, len(out), "page_archetypes: candidate rows for one scope"); shortfall != nil {
		// A thinned scope must not be trusted to say "no match, fall
		// through" — the lost row could have been the match. Refuse the
		// whole scope; sectionsForPage's caller falls through to the next
		// scope (and ultimately the Go switch), same as a query error.
		//
		// Falling through is right; doing it SILENTLY was not. A thinned site
		// scope drops to the fleet row and ships a different page structure
		// with nothing recording that a more specific rule was lost — which is
		// the 410 shape this guard cites while reproducing it one level up.
		if logger != nil {
			logger.Warn("pageArchetypeRows: scan shortfall — refusing this scope, falling through to the next",
				zap.Error(shortfall), zap.String("scope", archetypeScopeName(siteID, themeKitID)),
				zap.Int("offered", offered), zap.Int("kept", len(out)))
		}
		return nil, false
	}
	return out, true
}

// archetypeScopeName names a scope for logs, so a warning says WHICH of the
// three lookups was skipped rather than just that one was.
func archetypeScopeName(siteID uuid.UUID, themeKitID *uuid.UUID) string {
	switch {
	case siteID != uuid.Nil:
		return "site:" + siteID.String()
	case themeKitID != nil:
		return "theme_kit:" + themeKitID.String()
	default:
		return "fleet"
	}
}

// matchArchetypeRows applies the switch's own precedence within one scope's
// rows: page_type exact -> page_name exact -> page_name_contains ->
// page_name_suffix -> default.
func matchArchetypeRows(rows []archetypeRow, pageName, pageType string) ([]string, bool) {
	typeKey := strings.ToLower(strings.TrimSpace(pageType))
	nameKey := strings.ToLower(strings.TrimSpace(pageName))

	for _, kind := range []string{"page_type", "page_name", "page_name_contains", "page_name_suffix"} {
		for _, r := range rows {
			if r.MatchKind != kind {
				continue
			}
			switch kind {
			case "page_type":
				if typeKey != "" && r.MatchValue == typeKey {
					return r.Sections, true
				}
			case "page_name":
				if r.MatchValue == nameKey {
					return r.Sections, true
				}
			case "page_name_contains":
				if r.MatchValue != "" && strings.Contains(nameKey, r.MatchValue) {
					return r.Sections, true
				}
			case "page_name_suffix":
				if r.MatchValue != "" && strings.HasSuffix(nameKey, r.MatchValue) {
					return r.Sections, true
				}
			}
		}
	}
	for _, r := range rows {
		if r.MatchKind == "default" {
			return r.Sections, true
		}
	}
	return nil, false
}
