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
//           "item_domain":   "build",
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
	Optional:   []string{"spec_data", "parent_item_id", "page_id", "summary"},
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
	itemDomain, _ := config["item_domain"].(string)
	if itemDomain == "" {
		itemDomain = "build"
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
	}

	// Build spec JSONB from spec_data path
	specJSON := "{}"
	if specData := inputs.GetMap("spec_data"); specData != nil {
		if b, err := json.Marshal(specData); err == nil {
			specJSON = string(b)
		}
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

	// Insert via transaction (insertWorkItem expects *sql.Tx)
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       source,
		domain:       itemDomain,
		itemType:     itemType,
		severity:     severity,
		summary:      summary,
		spec:         specJSON,
		pageID:       pageID,
		priority:     priority,
		handlerAgent: handlerAgent,
		status:       status,
		createdBy:    source,
		itemKey:      itemKey,
		dependsOn:    dependsOn,
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
