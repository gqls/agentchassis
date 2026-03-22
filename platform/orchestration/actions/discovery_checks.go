// FILE: platform/orchestration/actions/run_discovery_checks_action.go
//
// AFTER migration — the main function becomes a simple loop.
// This replaces the ~480-line if-chain.
//
// The import of the discovery_checks package triggers all init()
// functions, registering every check automatically.

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"

	// This import triggers init() in every check_*.go file
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// --- InputSpec (carried over from existing file) ---

var RunDiscoveryChecksInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("run_discovery_checks", RunDiscoveryChecksInputSpec)
}

func RunDiscoveryChecksAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("RunDiscoveryChecksAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Extract inputs (unchanged) ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		RunDiscoveryChecksInputSpec,
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

	config := params.StepConfig.Config
	checkDomain, _ := config["check_domain"].(string)
	if checkDomain == "" {
		checkDomain = "design"
	}

	// --- Parse enabled checks from config ---
	enabledChecks := []string{"empty_sections", "undeployed_assets", "missing_css", "duplicate_palette"}
	if configChecks, ok := config["checks"].([]interface{}); ok && len(configChecks) > 0 {
		enabledChecks = make([]string, len(configChecks))
		for i, c := range configChecks {
			enabledChecks[i] = fmt.Sprintf("%v", c)
		}
	}

	var siteDomain string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID,
	).Scan(&siteDomain)

	logger.Info("RunDiscoveryChecksAction: Running checks",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteDomain),
		zap.Strings("checks", enabledChecks),
		zap.Strings("registered", checks.Names()),
	)

	// --- Shared state ---
	batchID := uuid.New()
	agentType := params.ExecutionContext.Sender.AgentType
	if agentType == "" {
		agentType = "discovery-agent"
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	dctx := checks.DiscoveryCheckContext{
		Ctx:       ctx,
		DB:        params.DB,
		TX:        tx,
		SiteID:    siteID,
		Domain:    checkDomain,
		AgentType: agentType,
		BatchID:   batchID,
		Logger:    logger,
	}

	// --- Run each enabled check ---
	var allFindings []interface{}
	inserted := 0
	skipped := 0

	for _, checkName := range enabledChecks {
		check := checks.Get(checkName)
		if check == nil {
			logger.Warn("Unknown discovery check — not registered",
				zap.String("check", checkName))
			continue
		}

		result, err := check.Run(dctx)
		if err != nil {
			logger.Warn("Discovery check failed",
				zap.String("check", checkName),
				zap.Error(err))
			continue
		}

		// Append findings
		for _, f := range result.Findings {
			allFindings = append(allFindings, f)
		}

		// Insert work items
		for _, wi := range result.WorkItems {
			ok, err := insertWorkItem(ctx, tx, workItem{
				siteID:       wi.SiteID,
				pageID:       wi.PageID,
				source:       wi.Source,
				pipeline:     wi.Pipeline,
				itemType:     wi.ItemType,
				severity:     wi.Severity,
				summary:      wi.Summary,
				spec:         wi.SpecJSON,
				priority:     wi.Priority,
				handlerAgent: wi.HandlerAgent,
				status:       wi.Status,
				createdBy:    wi.CreatedBy,
				itemKey:      wi.ItemKey,
				batchID:      wi.BatchID,
			}, logger)
			if err != nil {
				logger.Warn("Failed to insert work item",
					zap.String("check", checkName),
					zap.Error(err))
			} else if ok {
				inserted++
			} else {
				skipped++
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit work items: %w", err)
	}

	logger.Info("RunDiscoveryChecksAction: Complete",
		zap.Int("findings", len(allFindings)),
		zap.Int("items_inserted", inserted),
		zap.Int("items_skipped", skipped),
	)

	return map[string]interface{}{
		"site_id":        siteIDStr,
		"domain":         siteDomain,
		"checks_run":     enabledChecks,
		"findings":       allFindings,
		"items_inserted": inserted,
		"items_skipped":  skipped,
		"batch_id":       batchID.String(),
	}, nil
}
