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
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{"site_id_field": "site_id"},
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

	// Load style collection component IDs for this site
	var headerCompID, footerCompID, headCompID sql.NullString
	err = params.DB.QueryRowContext(ctx, `
		SELECT
			scol.header_component_id::text,
			scol.footer_component_id::text,
			(SELECT cc.id::text FROM content_components cc
			 JOIN site_components sc ON sc.component_id = cc.id
			 WHERE sc.site_id = s.id AND sc.slot_name = 'head'
			 LIMIT 1) as head_component_id
		FROM sites s
		LEFT JOIN style_collections scol ON s.style_collection_id = scol.id
		WHERE s.id = $1
	`, siteID).Scan(&headerCompID, &footerCompID, &headCompID)
	if err != nil {
		return nil, fmt.Errorf("failed to load style collection: %w", err)
	}

	// If style collection doesn't specify header/footer, try to find by function name
	if !headerCompID.Valid {
		var id string
		err := params.DB.QueryRowContext(ctx, `
			SELECT id::text FROM content_components
			WHERE function = 'site-header' AND is_active = true
			ORDER BY name LIMIT 1
		`).Scan(&id)
		if err == nil {
			headerCompID = sql.NullString{String: id, Valid: true}
			logger.Info("Resolved header component by function lookup",
				zap.String("component_id", id))
		}
	}

	if !footerCompID.Valid {
		var id string
		err := params.DB.QueryRowContext(ctx, `
			SELECT id::text FROM content_components
			WHERE function = 'site-footer' AND is_active = true
			ORDER BY name LIMIT 1
		`).Scan(&id)
		if err == nil {
			footerCompID = sql.NullString{String: id, Valid: true}
			logger.Info("Resolved footer component by function lookup",
				zap.String("component_id", id))
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
