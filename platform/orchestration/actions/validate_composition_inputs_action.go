// FILE: platform/orchestration/actions/validate_composition_inputs_action.go
//
// ValidateCompositionInputsAction checks that the two specs site-design-planner
// requires to do its job — identity and classification — exist in site_specs for
// the target site.
//
// This is the first step in site-design-planner's workflow. It runs before any
// resolver does work.
//
// Missing specs are treated as an upstream bug AND a self-heal opportunity:
//
//   1. Log loudly via logger.Error with explicit "upstream bug" message.
//      This is the primary signal — something that was supposed to run earlier
//      (classifier) did not. Logs are immediate but transient.
//
//   2. Insert a `needs_domain_research` work item for the dispatch loop to
//      pick up. This serves two purposes:
//
//      (a) Durable error visibility: if the same site keeps hitting this path,
//          the work item will accumulate failure history via the existing
//          two-strike rule in insertWorkItem(). Repeated failures become
//          "unresolved" items visible to dashboards, not just to log aggregators.
//
//      (b) Self-healing: the classifier runs, writes the missing spec, and the
//          parent `needs_composition` item (which depends_on this backfill item)
//          gets re-dispatched once the classifier item completes. The site
//          builds without manual intervention in the common case.
//
// The action never hard-fails — it returns ready=false so the workflow's
// conditional can short-circuit cleanly. The dispatch loop sees the composition
// item failed, and the newly-queued classifier item unblocks it.
//
// Returns:
//   {
//     "ready":             true | false,
//     "missing":           []string,    // aspects that were not found
//     "identity":          map,         // spec data if present, else nil
//     "classification":    map,         // spec data if present, else nil
//     "classifier_queued": string|null, // UUID of queued work item if we made one
//   }
//
// Registration (add to registry.go):
//
//   "validate_composition_inputs": {
//       Handler:     ValidateCompositionInputsAction,
//       Category:    "site",
//       Description: "Check identity + classification specs exist, queue classifier if not",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ValidateCompositionInputsInputSpec follows the path-resolution pattern used
// by other actions in this file set (see ReadSiteSpecAction).
var ValidateCompositionInputsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{},
	Defaults:    map[string]interface{}{},
	Deprecated: map[string]string{
		"site_id_field": "site_id",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("validate_composition_inputs", ValidateCompositionInputsInputSpec)
}

// ValidateCompositionInputsAction is the entry point invoked by the workflow
// engine. It reads identity + classification from site_specs and on miss
// logs loudly AND queues a classifier work item.
func ValidateCompositionInputsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "validate_composition_inputs"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// Resolve site_id via path extraction (same pattern as ReadSiteSpecAction)
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ValidateCompositionInputsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	// Domain for logging + work item summary — best-effort
	var domain string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)

	// Load both specs
	identityData, identityFound, err := loadCurrentSpecData(ctx, params.DB, siteID, "identity")
	if err != nil {
		return nil, fmt.Errorf("load identity spec: %w", err)
	}
	classificationData, classificationFound, err := loadCurrentSpecData(ctx, params.DB, siteID, "classification")
	if err != nil {
		return nil, fmt.Errorf("load classification spec: %w", err)
	}

	missing := make([]string, 0, 2)
	if !identityFound {
		missing = append(missing, "identity")
	}
	if !classificationFound {
		missing = append(missing, "classification")
	}

	if len(missing) == 0 {
		logger.Info("ValidateCompositionInputsAction: required specs present",
			zap.String("site_id", siteID.String()),
			zap.String("domain", domain),
		)
		return map[string]interface{}{
			"ready":             true,
			"missing":           missing,
			"identity":          identityData,
			"classification":    classificationData,
			"classifier_queued": nil,
		}, nil
	}

	// Loud log — upstream bug signal. Logs are the immediate channel.
	logger.Error("ValidateCompositionInputsAction: required specs missing — upstream bug",
		zap.String("site_id", siteID.String()),
		zap.String("domain", domain),
		zap.Strings("missing", missing),
		zap.String("remediation", "queueing classifier work item for dispatch loop"),
	)

	// Queue domain-research-classifier — durable signal and self-heal.
	// A persistent item with a stable item_key means:
	//   - Dashboard queries see repeated failures accumulate for this site
	//   - The two-strike rule in insertWorkItem() marks it `unresolved`
	//     after 2 attempts, making it visible for investigation
	//   - The classifier runs and writes the missing spec, unblocking
	//     the parent composition item via depends_on
	classifierItemID, err := queueClassifierForCompositionRecovery(
		ctx, params.DB, siteID, domain, missing, logger)
	if err != nil {
		// Even the queue failed — keep going, the logger.Error above is still
		// the primary signal. Return nil classifier_queued and let the caller
		// surface the state.
		logger.Error("ValidateCompositionInputsAction: failed to queue classifier recovery",
			zap.Error(err),
			zap.String("site_id", siteID.String()),
		)
	}

	var classifierQueuedVal interface{}
	if classifierItemID != nil {
		classifierQueuedVal = classifierItemID.String()
	}

	return map[string]interface{}{
		"ready":             false,
		"missing":           missing,
		"identity":          identityData,       // may be nil
		"classification":    classificationData, // may be nil
		"classifier_queued": classifierQueuedVal,
	}, nil
}

// loadCurrentSpecData returns the data map for a single aspect at is_current=true.
// Returns (data, found, err). found=false with nil data and nil err if the row
// does not exist.
func loadCurrentSpecData(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	aspect string,
) (map[string]interface{}, bool, error) {
	var dataJSON []byte
	err := db.QueryRowContext(ctx,
		`SELECT data FROM site_specs
		 WHERE site_id = $1 AND aspect = $2 AND is_current = true`,
		siteID, aspect,
	).Scan(&dataJSON)

	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("query %s spec: %w", aspect, err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		return nil, false, fmt.Errorf("unmarshal %s spec: %w", aspect, err)
	}
	return data, true, nil
}

// queueClassifierForCompositionRecovery inserts a needs_domain_research work
// item using the existing insertWorkItem helper. Returns the new item's UUID,
// or nil if dedup suppression prevented insertion (an open item with the same
// key already exists — that's fine, the dispatch loop will process it).
//
// Uses item_key `backfill_classification_for_composition` — distinct from the
// manual backfill key (`backfill_classification`) used by the one-shot SQL
// script run during the rollout, so the two coexist in history without
// collision.
func queueClassifierForCompositionRecovery(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	domain string,
	missingAspects []string,
	logger *zap.Logger,
) (*uuid.UUID, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	spec := map[string]interface{}{
		"domain":          domain,
		"reason":          "classification_spec_missing",
		"queued_by":       "site-design-planner:validate_composition_inputs",
		"missing_aspects": missingAspects,
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal spec: %w", err)
	}

	item := workItem{
		siteID:       siteID,
		source:       "side_effect",
		pipeline:     "build",
		itemType:     "needs_domain_research",
		severity:     "high",
		summary:      fmt.Sprintf("Backfill classification for %s (requested by site-design-planner)", domain),
		spec:         string(specJSON),
		priority:     5,
		handlerAgent: "domain-research-classifier",
		status:       "triaged",
		createdBy:    "site-design-planner",
		itemKey:      "backfill_classification_for_composition",
	}

	inserted, err := insertWorkItem(ctx, tx, item, logger)
	if err != nil {
		return nil, fmt.Errorf("insert work item: %w", err)
	}
	if !inserted {
		// Dedup: an open backfill item already exists for this site. Fine —
		// dispatch will process the existing one. insertWorkItem logs the
		// suppression reason already.
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit tx (suppressed): %w", err)
		}
		return nil, nil
	}

	// Look up the ID we just inserted (insertWorkItem doesn't return it).
	// The partial unique index guarantees at most one open row for this key.
	var newItemID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM site_work_items
		WHERE site_id = $1
		  AND item_key = 'backfill_classification_for_composition'
		  AND status NOT IN ('complete','verified','rejected','wont_fix','failed')
		ORDER BY created_at DESC
		LIMIT 1
	`, siteID).Scan(&newItemID)
	if err != nil {
		return nil, fmt.Errorf("fetch inserted item id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	logger.Info("Queued classifier recovery item",
		zap.String("site_id", siteID.String()),
		zap.String("new_item_id", newItemID.String()),
	)
	return &newItemID, nil
}
