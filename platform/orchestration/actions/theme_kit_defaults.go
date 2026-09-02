// FILE: platform/orchestration/actions/theme_kit_defaults.go
//
// loadSiteThemeKitDefaults is the one shared, siteID-keyed lookup every
// composition-resolution consumer of theme_kits uses. It reads the site's
// CURRENT 'theme_kit_adoption' site_specs row and joins theme_kits — a
// plain DB read, not something threaded through collected_data, so it
// works identically whether the caller was invoked via apply_theme_kit's
// own needs_composition dispatch or completely independently (a normal
// site build that happens to have adopted a kit earlier).
//
// Returns ok=false when the site has no current theme_kit_adoption spec,
// or when the referenced kit (or the row it points at) is no longer active
// — a kit deactivated after a site adopted it must not silently keep
// steering that site's builds.

package actions

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// chromeComponentEligible checks ONE component id against the same
// eligibility predicate ResolveChromeComponent/fork_theme_from_site use
// (chromePinEligibleSQL, component_library.go) — is_active AND
// component_level IN ('site','header','footer','head'). Used when pinning a
// theme kit's chrome so a kit can never pin a row that would fail this
// check anyway (e.g. a component_level='section' row with a
// confusingly-matching name — see the 686 migration's header note).
func chromeComponentEligible(ctx context.Context, tx queryRowContexter, id uuid.UUID) (bool, error) {
	var eligible bool
	err := tx.QueryRowContext(ctx,
		`SELECT (`+chromePinEligibleSQL("")+`) FROM content_components WHERE id = $1`,
		id,
	).Scan(&eligible)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return eligible, err
}

// queryRowContexter is satisfied by both *sql.DB and *sql.Tx — this helper
// is called from install_site_composition_action.go inside a transaction.
type queryRowContexter interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type siteThemeKitDefaults struct {
	ThemeKitID        uuid.UUID
	ThemeKitName      string
	LayoutID          uuid.NullUUID
	HeaderComponentID uuid.NullUUID
	FooterComponentID uuid.NullUUID
}

func loadSiteThemeKitDefaults(ctx context.Context, db *sql.DB, siteID uuid.UUID) (siteThemeKitDefaults, bool, error) {
	var out siteThemeKitDefaults
	row := db.QueryRowContext(ctx, `
		SELECT tk.id, tk.name, tk.layout_id, tk.header_component_id, tk.footer_component_id
		FROM site_specs ss
		JOIN theme_kits tk ON tk.id = (ss.data->>'theme_kit_id')::uuid
		WHERE ss.site_id = $1 AND ss.aspect = 'theme_kit_adoption' AND ss.is_current = true
		  AND tk.is_active = true
	`, siteID)
	var layoutID, headerID, footerID sql.NullString
	if err := row.Scan(&out.ThemeKitID, &out.ThemeKitName, &layoutID, &headerID, &footerID); err != nil {
		if err == sql.ErrNoRows {
			return siteThemeKitDefaults{}, false, nil
		}
		return siteThemeKitDefaults{}, false, err
	}
	if layoutID.Valid {
		if id, perr := uuid.Parse(layoutID.String); perr == nil {
			out.LayoutID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}
	if headerID.Valid {
		if id, perr := uuid.Parse(headerID.String); perr == nil {
			out.HeaderComponentID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}
	if footerID.Valid {
		if id, perr := uuid.Parse(footerID.String); perr == nil {
			out.FooterComponentID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}
	return out, true, nil
}
