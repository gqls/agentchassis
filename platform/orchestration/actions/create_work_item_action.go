// FILE: platform/orchestration/actions/create_work_item_action.go
//
// Inserts a single work item into site_work_items. Used by handler agents
// to chain to the next pipeline stage after completing their work.
//
// Reuses the existing insertWorkItem() private helper for dedup behavior.
//
// Workflow config example (domain-research-classifier creating next item):
//
//   "create_next_item": {
//       "action": "create_work_item",
//       "config": {
//           "site_id":       "input_data.site_id",
//           "item_type":     "needs_briefing",
//           "handler_agent": "briefing-agent",
//           "item_pipeline":   "build",
//           "severity":      "high",
//           "source":        "domain-research-classifier",
//           "summary":       "Briefing needed",
//           "item_key_prefix": "briefing",
//           "spec_data":     "classification_result",
//           "parent_item_id": "input_data.work_item_id",
//           "priority":      10
//       },
//       "output_field": "next_item_created",
//       "next_step": "complete"
//   }
//
// Three optional config keys exist for callers whose item is an ACTION REQUEST
// rather than a detected defect (bugs_open/024 — a tool re-render request that
// was born dead three ways at once):
//
//   "spec_literal":          {"reason": "section_data_resolved"}
//       Constants written into the spec verbatim. spec_data is a PATH into
//       collected data and can never express a literal, so without this no
//       workflow step can stamp a value that a downstream gate reads.
//
//   "spec_paths":            {"component_id": "update_result.component_id"}
//       Per-key paths resolved individually. A configured path that does not
//       resolve is a HARD ERROR — an incomplete spec silently degrades the
//       downstream re-render to assemble-from-stale and still reports success.
//
//   "item_key_suffix_field": "update_result.component_id"
//       Appends a resolved value to the item_key. The default key is
//       '<prefix>_<domain>', which is SITE-wide: without a suffix, two
//       components fixed close together collide on idx_swi_dedup and one
//       request is lost.
//
//   "recurrence_expected":   true
//       Sets workItem.recurrenceExpected, which skips insertWorkItem's
//       anti-churn heuristics (dedup is NOT waived). For an action request a
//       'complete' predecessor is a SUCCESS, not a strike.
//
// Layering: spec_data, then spec_paths, then spec_literal — later wins.
//
// Registration (add to registry.go):
//   "create_work_item": { Handler: CreateWorkItemAction, IsLocal: true },

package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CreateWorkItemInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"spec_data", "parent_item_id", "page_id", "component_id", "summary"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("create_work_item", CreateWorkItemInputSpec)
}

func CreateWorkItemAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "create_work_item"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		CreateWorkItemInputSpec,
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

	// Config literals
	config := params.StepConfig.Config
	itemType, _ := config["item_type"].(string)
	if itemType == "" {
		return nil, fmt.Errorf("item_type config is required")
	}
	handlerAgent, _ := config["handler_agent"].(string)
	if handlerAgent == "" {
		return nil, fmt.Errorf("handler_agent config is required")
	}
	itemPipeline, _ := config["item_pipeline"].(string)
	if itemPipeline == "" {
		itemPipeline, _ = config["item_domain"].(string) // backwards compat
	}
	if itemPipeline == "" {
		itemPipeline = "build"
	}
	severity, _ := config["severity"].(string)
	if severity == "" {
		severity = "high"
	}
	source, _ := config["source"].(string)
	if source == "" {
		source = params.AgentType
	}
	summary := inputs.Get("summary")
	if summary == "" {
		summary, _ = config["summary"].(string)
	}
	if summary == "" {
		summary = itemType
	}
	status, _ := config["status"].(string)
	if status == "" {
		status = "triaged"
	}
	priority := datahelpers.GetIntField(config, "priority", 100)

	// Build item_key from prefix + domain
	itemKeyPrefix, _ := config["item_key_prefix"].(string)
	var itemKey string
	if itemKeyPrefix != "" {
		// Try to get domain for the key
		domain := inputs.Get("domain")
		if domain == "" {
			domain = siteID.String()[:8]
		}
		itemKey = fmt.Sprintf("%s_%s", itemKeyPrefix, domain)

		// Optional scoping suffix. '<prefix>_<domain>' is SITE-wide: two
		// components fixed close together on one site collide on
		// idx_swi_dedup (site_id, item_key) and one request is simply lost.
		// A step that knows what it is acting on names it here.
		// Additive by design — unset or unresolved leaves the key exactly as
		// it was, so no existing caller changes.
		if f, ok := config["item_key_suffix_field"].(string); ok && f != "" {
			if sfx := datahelpers.ExtractNestedFieldString(params.CollectedData, f); sfx != "" {
				itemKey = fmt.Sprintf("%s_%s", itemKey, sfx)
			} else {
				logger.Warn("create_work_item: item_key_suffix_field did not resolve, key stays site-wide",
					zap.String("field", f),
					zap.String("item_key", itemKey),
				)
			}
		}
	}

	// recurrence_expected marks an item whose RE-REQUEST is normal rather than
	// evidence that previous handlers failed — see workItem.recurrenceExpected.
	// Default false, so existing callers keep today's behaviour exactly.
	recurrenceExpected, _ := config["recurrence_expected"].(bool)

	// Build the spec JSONB in three layers, later winning over earlier:
	//   spec_data    — a PATH to a map already in collected data (as before)
	//   spec_paths   — {key: path}, each resolved individually
	//   spec_literal — {key: value}, verbatim
	// spec_data alone could never express a constant, which is why no workflow
	// step could stamp e.g. reason='section_data_resolved' — the value
	// create_rerender_items gates the section re-render on (bugs_open/024).
	specMap := map[string]interface{}{}
	if specData := inputs.GetMap("spec_data"); specData != nil {
		for k, v := range specData {
			specMap[k] = v
		}
	}

	// spec_paths: a configured path that does not resolve is a HARD ERROR.
	// The step author asked for this field; silently omitting it is exactly how
	// bugs_open/024 stayed invisible for three cycles — a spec missing
	// component_id makes create_rerender_items' scoped=false, which degrades
	// the re-render to assemble-from-stale and still reports success.
	if sp, ok := config["spec_paths"].(map[string]interface{}); ok {
		for key, rawPath := range sp {
			path, isStr := rawPath.(string)
			if !isStr || path == "" {
				return nil, fmt.Errorf("spec_paths[%q] must be a non-empty string path", key)
			}
			val := datahelpers.ExtractNestedField(params.CollectedData, path)
			if val == nil || val == "" {
				return nil, fmt.Errorf(
					"spec_paths[%q]: path %q did not resolve; refusing to create a work item with an incomplete spec",
					key, path)
			}
			specMap[key] = val
		}
	}

	// spec_literal: constants, no resolution attempted.
	if sl, ok := config["spec_literal"].(map[string]interface{}); ok {
		for k, v := range sl {
			specMap[k] = v
		}
	}

	specJSON := "{}"
	if len(specMap) > 0 {
		b, err := json.Marshal(specMap)
		if err != nil {
			return nil, fmt.Errorf("marshal spec: %w", err)
		}
		specJSON = string(b)
	}

	// Optional parent_item_id
	var dependsOn []uuid.UUID
	if parentIDStr := inputs.Get("parent_item_id"); parentIDStr != "" {
		if parentID, err := uuid.Parse(parentIDStr); err == nil {
			dependsOn = append(dependsOn, parentID)
		}
	}

	// Optional page_id
	var pageID *uuid.UUID
	if pageIDStr := inputs.Get("page_id"); pageIDStr != "" {
		if parsed, err := uuid.Parse(pageIDStr); err == nil {
			pageID = &parsed
		}
	}

	// Optional component_id
	var componentID *uuid.UUID
	if componentIDStr := inputs.Get("component_id"); componentIDStr != "" {
		if parsed, err := uuid.Parse(componentIDStr); err == nil {
			componentID = &parsed
		}
	}

	// Insert via transaction (insertWorkItem expects *sql.Tx)
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       source,
		pipeline:     itemPipeline,
		itemType:     itemType,
		severity:     severity,
		summary:      summary,
		spec:         specJSON,
		pageID:       pageID,
		componentID:  componentID,
		priority:     priority,
		handlerAgent: handlerAgent,
		status:       status,
		createdBy:    source,
		itemKey:      itemKey,
		dependsOn:    dependsOn,

		recurrenceExpected: recurrenceExpected,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("insert work item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("CreateWorkItemAction: complete",
		zap.String("item_type", itemType),
		zap.String("handler_agent", handlerAgent),
		zap.Bool("inserted", inserted),
		zap.String("item_key", itemKey),
	)

	return map[string]interface{}{
		"inserted":      inserted,
		"item_type":     itemType,
		"handler_agent": handlerAgent,
		"site_id":       siteID.String(),
		"item_key":      itemKey,
		"deduped":       !inserted,
	}, nil
}
