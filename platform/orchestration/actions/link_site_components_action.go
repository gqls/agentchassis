// FILE: platform/orchestration/actions/link_site_components_action.go
//
// LinkSiteComponentsAction ensures site_components rows have their component_id
// set correctly from the site's style collection. Without this linkage,
// renderAndStoreSiteComponent falls through to a hardcoded fallback that ignores
// the style collection's templates (see 003b Site Component Linkage Contract).
//
// This action is idempotent — safe to run multiple times. It only updates rows
// where component_id is NULL or differs from the style collection.
//
// Both of its by-function fallbacks resolve through ResolveChromeComponent
// (component_library.go) so that this action, render_site_components and the
// library lookup cannot disagree about which component serves a chrome slot —
// bugs_open/118, where they disagreed three ways.
//
// Config:
//   - site_id_field: path to site_id in collected_data (default: "site_record.site_id")
//
// Registration:
//   "link_site_components": {
//       Handler:     LinkSiteComponentsAction,
//       Category:    "site",
//       Description: "Link site_components to content_components from style collection",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LinkSiteComponentsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{"site_id_field": "site_id"},
}

func init() {
	datahelpers.RegisterActionInputSpec("link_site_components", LinkSiteComponentsInputSpec)
}

func LinkSiteComponentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "link_site_components"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LinkSiteComponentsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Load style collection component IDs for this site, WITH the eligibility of
	// each pin decided by the same query that reads it (bugs_open/170).
	//
	// The pins used to arrive unqualified, and the `!Valid` fall-through below was
	// the only gate on them — so a style collection pinning a DEACTIVATED component
	// was written straight into site_components.component_id, bypassing the one
	// chrome predicate bugs_open/118 installed and overwriting the repair
	// bugs_open/166 had performed on the same column. Three deployed sites were
	// pinned to a deactivated header and four to a deactivated footer on
	// 2026-08-01, while their assignments were already correct: this action is
	// what would have undone that.
	//
	// chromePinEligibleSQL, not chromeEligibleSQL: a pin naming the site's own
	// FORK is legitimate and must survive — see the predicate's own comment.
	var headerCompID, footerCompID, headCompID sql.NullString
	var headerPinName, footerPinName sql.NullString
	var headerPinOK, footerPinOK sql.NullBool
	err = params.DB.QueryRowContext(ctx, `
		SELECT
			scol.header_component_id::text,
			scol.footer_component_id::text,
			hc.name, (`+chromePinEligibleSQL("hc.")+`),
			fc.name, (`+chromePinEligibleSQL("fc.")+`),
			(SELECT cc.id::text FROM content_components cc
			 JOIN site_components sc ON sc.component_id = cc.id
			 WHERE sc.site_id = s.id AND sc.slot_name = 'head'
			 LIMIT 1) as head_component_id
		FROM sites s
		LEFT JOIN style_collections scol ON s.style_collection_id = scol.id
		LEFT JOIN content_components hc ON hc.id = scol.header_component_id
		LEFT JOIN content_components fc ON fc.id = scol.footer_component_id
		WHERE s.id = $1
	`, siteID).Scan(&headerCompID, &footerCompID,
		&headerPinName, &headerPinOK, &footerPinName, &footerPinOK, &headCompID)
	if err != nil {
		return nil, fmt.Errorf("failed to load style collection: %w", err)
	}

	// An ineligible pin is DISCARDED, which hands the slot to the eligible-only
	// library lookup below — the same path a collection with no pin at all takes.
	// Discarding rather than refusing is deliberate: refusing would leave the slot
	// pointing at whatever it already held, which for these sites is the very
	// component the pin is wrong about.
	for _, pin := range []struct {
		slot string
		id   *sql.NullString
		name sql.NullString
		ok   sql.NullBool
	}{
		{"header", &headerCompID, headerPinName, headerPinOK},
		{"footer", &footerCompID, footerPinName, footerPinOK},
	} {
		if pin.id.Valid && !pin.ok.Bool {
			logger.Warn("link_site_components: style collection pins an ineligible component — falling back to the library (bugs_open/170)",
				zap.String("slot", pin.slot),
				zap.String("pinned_component", pin.name.String),
				zap.String("pinned_component_id", pin.id.String),
			)
			*pin.id = sql.NullString{}
		}
	}

	// If the style collection doesn't specify header/footer — or specified one that
	// is no longer eligible chrome, discarded just above (bugs_open/170) — resolve
	// the library default through the ONE chrome predicate (bugs_open/118). These
	// two lookups
	// used to filter on `is_active` alone, which reads as careful and is not: a
	// FORK carries its parent's `function`, so `header-leopardess` — one client's
	// forked header — sorted first among active `site-header` rows and is what this
	// would have linked into any other site whose style collection has no header.
	//
	// Unlike the render path, this one ASSIGNS without rendering, so an ineligible
	// component here would be pinned in site_components and never noticed. It
	// therefore links only what the library says is eligible chrome, and leaves the
	// slot alone otherwise.
	if !headerCompID.Valid {
		if comp, eligible, err := ResolveChromeComponent(ctx, params.DB, ChromeSlotFunction("header"), logger); err == nil && eligible {
			headerCompID = sql.NullString{String: comp.ID, Valid: true}
			logger.Info("Resolved header component by function lookup",
				zap.String("component_id", comp.ID),
				zap.String("component_name", comp.Name))
		}
	}

	if !footerCompID.Valid {
		if comp, eligible, err := ResolveChromeComponent(ctx, params.DB, ChromeSlotFunction("footer"), logger); err == nil && eligible {
			footerCompID = sql.NullString{String: comp.ID, Valid: true}
			logger.Info("Resolved footer component by function lookup",
				zap.String("component_id", comp.ID),
				zap.String("component_name", comp.Name))
		}
	}

	// Map slot → component_id
	slotMapping := map[string]sql.NullString{
		"header": headerCompID,
		"footer": footerCompID,
	}
	if headCompID.Valid {
		slotMapping["head"] = headCompID
	}

	linked := 0
	lockedSlots := []string{}
	for slot, compID := range slotMapping {
		if !compID.Valid {
			logger.Warn("No component found for slot",
				zap.String("slot", slot))
			continue
		}

		compUUID, err := uuid.Parse(compID.String)
		if err != nil {
			logger.Warn("Invalid component UUID",
				zap.String("slot", slot),
				zap.String("id", compID.String))
			continue
		}

		// Human lock gate (bugs_open/069). Unlike every other writer in this
		// family the pre-check here is LOAD-BEARING, not advisory: the upsert's
		// own WHERE means "0 rows affected" is the NORMAL, expected outcome
		// (the slot already points at the right component), so RowsAffected
		// cannot tell "nothing to do" from "a lock refused me". Only a read of
		// the row before the write can. Do not "simplify" this away.
		//
		// It matters more here than anywhere else in this fix: the DO UPDATE
		// arm sets rendered_html = NULL, so an ungated relink does not
		// overwrite a locked human artefact — it ERASES it.
		lock, lockErr := CheckSiteComponentLock(ctx, params.DB, siteID, slot, logger)
		if lockErr != nil {
			logger.Warn("link_site_components: chrome lock check failed — relying on the guarded upsert",
				zap.String("slot", slot), zap.Error(lockErr))
		} else if lock.RowExists && lock.ComponentID.Valid && lock.ComponentID.UUID == compUUID {
			// Already linked correctly: no write was pending, so a lock blocks
			// nothing and there is nothing to report.
			continue
		} else if lock.IsLocked {
			logger.Warn("link_site_components: refusing to relink human-locked chrome slot (bugs_open/069)",
				zap.String("slot", slot),
				zap.String("locked_by", lock.LockedBy),
				zap.String("wanted_component_id", compID.String),
			)
			emitChromeLockBlockedChangeItem(ctx, params.DB, siteID, slot, lock,
				"relink", "link_site_components", logger)
			lockedSlots = append(lockedSlots, slot)
			continue
		}

		// Upsert: create the site_components row if missing, update
		// component_id if wrong. Guarded — see relinkSiteComponent.
		changed, err := relinkSiteComponent(ctx, params.DB, siteID, slot, compUUID)
		if err != nil {
			logger.Warn("Failed to link site component",
				zap.String("slot", slot),
				zap.Error(err))
			continue
		}

		if changed {
			linked++
			logger.Info("Linked site component",
				zap.String("slot", slot),
				zap.String("component_id", compID.String))
		}
	}

	logger.Info("LinkSiteComponentsAction: Complete",
		zap.String("site_id", siteIDStr),
		zap.Int("linked", linked),
		zap.Strings("locked_slots_skipped", lockedSlots),
	)

	return map[string]interface{}{
		"linked":               linked,
		"site_id":              siteIDStr,
		"locked_slots_skipped": lockedSlots,
	}, nil
}
