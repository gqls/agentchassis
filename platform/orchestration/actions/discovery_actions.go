// platform/orchestration/actions/discovery_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/discovery"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================================
// DISCOVERY ACTIONS - Updated for agent_definitions architecture
// ============================================================================

// DiscoverBestAgentsAction finds agents matching requirements
// This replaces PlanAgentTeamAction with a simpler, more focused approach
//
// Config:
//   - capabilities: []string - required capabilities
//   - task_type: string - optional, used for proposal generation if no agents found
//   - min_version: int - optional minimum version
//
// Returns:
//   - agents: []AgentDefinitionResult - matching agents
//   - proposal: map - if no agents found, a proposal for creating them
func DiscoverBestAgentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	// Extract capabilities to search for
	var capabilities []string
	if caps, ok := config["capabilities"].([]interface{}); ok {
		for _, c := range caps {
			if s, ok := c.(string); ok {
				capabilities = append(capabilities, s)
			}
		}
	}

	if len(capabilities) == 0 {
		return nil, fmt.Errorf("capabilities required for agent discovery")
	}

	// Get database connection
	sqlDB := getSQLDB(params.DB)
	if sqlDB == nil {
		return nil, fmt.Errorf("database connection required for discovery")
	}

	agentDiscovery := discovery.NewAgentDefinitionDiscovery(sqlDB)

	// Find agents by capabilities
	agents, err := agentDiscovery.FindByCapabilities(ctx, capabilities, params.Logger)
	if err != nil {
		params.Logger.Warn("Error finding agents by capabilities",
			zap.Strings("capabilities", capabilities),
			zap.Error(err))
	}

	if len(agents) > 0 {
		// Found matching agents
		agentList := make([]map[string]interface{}, len(agents))
		for i, agent := range agents {
			agentList[i] = map[string]interface{}{
				"id":           agent.ID,
				"type":         agent.Type,
				"display_name": agent.DisplayName,
				"capabilities": agent.Capabilities,
				"version":      agent.Version,
				"usage_count":  agent.UsageCount,
			}
		}

		return map[string]interface{}{
			"found":  true,
			"agents": agentList,
			"count":  len(agents),
		}, nil
	}

	// No agents found - create a proposal
	taskType, _ := config["task_type"].(string)
	proposal := createAgentProposal(taskType, capabilities)

	return map[string]interface{}{
		"found":    false,
		"proposal": proposal,
		"message":  "No agents found with required capabilities. Review proposal?",
		"action":   "request_approval",
	}, nil
}

// ReviewPerformanceAction analyzes execution performance and suggests improvements
// Records metrics to entity_state_log and creates improvement_proposals if needed
//
// Reads from CollectedData:
//   - execution_metrics: map with duration, step_durations, failed_steps
//   - agent_type or agent_definition.type: the agent being reviewed
//   - entity_id: the entity this execution was for (e.g., domain name)
//
// Returns:
//   - needs_improvement: bool
//   - analysis: performance details
//   - proposal_id: if improvement needed, ID of created proposal
func ReviewPerformanceAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Get execution metrics
	metricsRaw, ok := params.CollectedData["execution_metrics"]
	if !ok {
		return nil, fmt.Errorf("execution_metrics not found in collected data")
	}

	metrics, ok := metricsRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("execution_metrics is not a map")
	}

	// Get agent type being reviewed
	agentType := extractAgentType(params.CollectedData)
	if agentType == "" {
		return nil, fmt.Errorf("agent_type not found in collected data")
	}

	// Get entity ID for state logging
	entityID := extractEntityID(params.CollectedData)

	// Extract and analyze metrics
	duration := extractFloat(metrics, "duration")
	failedSteps := extractStringSlice(metrics, "failed_steps")

	analysis := performanceAnalysis{
		TotalDuration:   duration,
		BottleneckSteps: identifyBottlenecks(metrics),
		FailedSteps:     failedSteps,
		QualityScore:    calculateQuality(params.CollectedData),
	}

	// Record performance to entity_state_log
	if entityID != "" {
		recordPerformanceToEntityState(ctx, params.DB, entityID, agentType, analysis, params.Logger)
	}

	// Update agent usage count
	sqlDB := getSQLDB(params.DB)
	if sqlDB != nil {
		agentDiscovery := discovery.NewAgentDefinitionDiscovery(sqlDB)
		go agentDiscovery.UpdateUsageCount(context.Background(), agentType)
	}

	// Check if improvements needed
	if analysis.QualityScore < 0.8 || len(analysis.BottleneckSteps) > 0 || len(analysis.FailedSteps) > 0 {
		suggestions := generateImprovementSuggestions(analysis)

		// Create improvement proposal
		proposalID, err := createImprovementProposal(ctx, params.DB, agentType, analysis, suggestions, params.Logger)
		if err != nil {
			params.Logger.Warn("Failed to create improvement proposal", zap.Error(err))
		}

		return map[string]interface{}{
			"needs_improvement": true,
			"analysis":          analysis,
			"suggestions":       suggestions,
			"proposal_id":       proposalID,
			"action":            "pause_for_human_review",
		}, nil
	}

	return map[string]interface{}{
		"needs_improvement": false,
		"analysis":          analysis,
	}, nil
}

// ApproveImprovementAction handles human approval of proposed improvements
// Updates agent_definitions with approved changes, creates new version if needed
//
// Reads from CollectedData:
//   - human_approval.approved: bool
//   - human_approval.proposal_id: string
//   - human_approval.approved_changes: []changes to apply
//   - human_approval.rejection_reason: string (if rejected)
func ApproveImprovementAction(ctx context.Context, params ActionParams) (interface{}, error) {
	approvalRaw, ok := params.CollectedData["human_approval"]
	if !ok {
		return nil, fmt.Errorf("human_approval not found in collected data")
	}

	approval, ok := approvalRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("human_approval is not a map")
	}

	proposalID, _ := approval["proposal_id"].(string)

	approved, ok := approval["approved"].(bool)
	if !ok || !approved {
		// Rejected
		reason := "No reason provided"
		if r, ok := approval["rejection_reason"].(string); ok {
			reason = r
		}

		// Update proposal status
		updateProposalStatus(ctx, params.DB, proposalID, "rejected", reason, params.Logger)

		return map[string]interface{}{
			"changes_applied": false,
			"reason":          reason,
		}, nil
	}

	// Get approved changes
	changesRaw, ok := approval["approved_changes"]
	if !ok {
		return nil, fmt.Errorf("approved_changes not found")
	}

	changes, ok := changesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("approved_changes is not an array")
	}

	// Apply changes
	appliedChanges := []map[string]interface{}{}
	clientID := params.Headers["client_id"]

	for _, change := range changes {
		changeMap, ok := change.(map[string]interface{})
		if !ok {
			continue
		}

		changeType, _ := changeMap["type"].(string)
		result := applyChange(ctx, params, changeType, changeMap, clientID)
		appliedChanges = append(appliedChanges, result)
	}

	// Update proposal status
	updateProposalStatus(ctx, params.DB, proposalID, "approved", "", params.Logger)

	return map[string]interface{}{
		"changes_applied": true,
		"applied":         appliedChanges,
		"proposal_id":     proposalID,
	}, nil
}

// ============================================================================
// HELPER TYPES
// ============================================================================

type performanceAnalysis struct {
	TotalDuration   float64  `json:"total_duration"`
	BottleneckSteps []string `json:"bottleneck_steps"`
	FailedSteps     []string `json:"failed_steps"`
	QualityScore    float64  `json:"quality_score"`
}

// ============================================================================
// HELPER FUNCTIONS - Analysis
// ============================================================================

func identifyBottlenecks(metrics map[string]interface{}) []string {
	bottlenecks := []string{}

	stepDurations, ok := metrics["step_durations"].(map[string]interface{})
	if !ok {
		return bottlenecks
	}

	// Calculate average
	totalDuration := 0.0
	count := 0
	for _, d := range stepDurations {
		if duration, ok := d.(float64); ok {
			totalDuration += duration
			count++
		}
	}

	if count == 0 {
		return bottlenecks
	}

	avgDuration := totalDuration / float64(count)
	threshold := avgDuration * 2

	// Find steps exceeding threshold
	for step, d := range stepDurations {
		if duration, ok := d.(float64); ok && duration > threshold {
			bottlenecks = append(bottlenecks, step)
		}
	}

	return bottlenecks
}

func calculateQuality(collectedData map[string]interface{}) float64 {
	quality := 1.0

	// Deduct for errors
	if errors, ok := collectedData["errors"].([]interface{}); ok {
		quality -= float64(len(errors)) * 0.1
	}

	// Use human feedback if available
	if feedback, ok := collectedData["human_feedback"].(map[string]interface{}); ok {
		if rating, ok := feedback["rating"].(float64); ok {
			quality = rating / 5.0
		}
	}

	// Clamp to [0, 1]
	if quality < 0 {
		return 0
	}
	if quality > 1 {
		return 1
	}
	return quality
}

func generateImprovementSuggestions(analysis performanceAnalysis) []map[string]interface{} {
	suggestions := []map[string]interface{}{}

	if len(analysis.BottleneckSteps) > 0 {
		suggestions = append(suggestions, map[string]interface{}{
			"type":       "optimize_steps",
			"target":     analysis.BottleneckSteps,
			"reason":     "Steps taking longer than average",
			"impact":     "medium",
			"auto_apply": false,
		})
	}

	if analysis.QualityScore < 0.7 {
		suggestions = append(suggestions, map[string]interface{}{
			"type":       "add_quality_check",
			"reason":     "Output quality below threshold",
			"impact":     "high",
			"auto_apply": false,
		})
	}

	if len(analysis.FailedSteps) > 0 {
		suggestions = append(suggestions, map[string]interface{}{
			"type":       "add_retry_logic",
			"target":     analysis.FailedSteps,
			"reason":     "Steps are failing",
			"impact":     "high",
			"auto_apply": false,
		})
	}

	return suggestions
}

// ============================================================================
// HELPER FUNCTIONS - Data Extraction
// ============================================================================

func extractAgentType(data map[string]interface{}) string {
	// Try direct field
	if at, ok := data["agent_type"].(string); ok && at != "" {
		return at
	}

	// Try from agent_definition
	if ad, ok := data["agent_definition"].(map[string]interface{}); ok {
		if at, ok := ad["type"].(string); ok {
			return at
		}
	}

	// Try from agent_group (backward compat)
	if ag, ok := data["agent_group"].(map[string]interface{}); ok {
		if at, ok := ag["type"].(string); ok {
			return at
		}
	}

	return ""
}

func extractEntityID(data map[string]interface{}) string {
	// Try direct field
	if id, ok := data["entity_id"].(string); ok && id != "" {
		return id
	}

	// Try from input_data.domain
	if input, ok := data["input_data"].(map[string]interface{}); ok {
		if domain, ok := input["domain"].(string); ok {
			return domain
		}
	}

	return ""
}

func extractFloat(data map[string]interface{}, key string) float64 {
	if v, ok := data[key].(float64); ok {
		return v
	}
	return 0
}

func extractStringSlice(data map[string]interface{}, key string) []string {
	result := []string{}
	if arr, ok := data[key].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
	}
	return result
}

func getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	return nil
}

// ============================================================================
// HELPER FUNCTIONS - Database
// ============================================================================

func getSQLDB(db interface{}) *sql.DB {
	switch d := db.(type) {
	case *sql.DB:
		return d
	default:
		return nil
	}
}

func execQuery(ctx context.Context, db interface{}, query string, args ...interface{}) error {
	var err error
	switch d := db.(type) {
	case *sql.DB:
		_, err = d.ExecContext(ctx, query, args...)
	case *pgxpool.Pool:
		_, err = d.Exec(ctx, query, args...)
	default:
		err = fmt.Errorf("unsupported database type: %T", db)
	}
	return err
}

// ============================================================================
// HELPER FUNCTIONS - Entity State
// ============================================================================

func recordPerformanceToEntityState(ctx context.Context, db interface{}, entityID, agentType string, analysis performanceAnalysis, logger *zap.Logger) {
	performanceData := map[string]interface{}{
		"duration":     analysis.TotalDuration,
		"quality":      analysis.QualityScore,
		"bottlenecks":  analysis.BottleneckSteps,
		"failed_steps": analysis.FailedSteps,
		"timestamp":    time.Now(),
	}

	dataJSON, err := json.Marshal(performanceData)
	if err != nil {
		logger.Warn("Failed to marshal performance data", zap.Error(err))
		return
	}

	query := `
		INSERT INTO entity_state_log (entity_id, entity_type, namespace, path, data, created_by_agent_type)
		VALUES ($1, 'domain', $2, 'performance.execution', $3::jsonb, $2)
	`

	if err := execQuery(ctx, db, query, entityID, agentType, dataJSON); err != nil {
		logger.Warn("Failed to record performance to entity_state_log", zap.Error(err))
	}
}

// ============================================================================
// HELPER FUNCTIONS - Improvement Proposals
// ============================================================================

func createImprovementProposal(ctx context.Context, db interface{}, agentType string, analysis performanceAnalysis, suggestions []map[string]interface{}, logger *zap.Logger) (string, error) {
	proposalID := uuid.New().String()

	proposedChanges := map[string]interface{}{
		"analysis":    analysis,
		"suggestions": suggestions,
	}
	changesJSON, err := json.Marshal(proposedChanges)
	if err != nil {
		return "", err
	}

	query := `
		INSERT INTO improvement_proposals (id, target_type, target_id, proposed_changes, source, status)
		VALUES ($1, 'agent_definition', $2, $3::jsonb, 'metrics', 'pending')
	`

	if err := execQuery(ctx, db, query, proposalID, agentType, changesJSON); err != nil {
		return "", err
	}

	logger.Info("Created improvement proposal",
		zap.String("proposal_id", proposalID),
		zap.String("agent_type", agentType))

	return proposalID, nil
}

func updateProposalStatus(ctx context.Context, db interface{}, proposalID, status, reason string, logger *zap.Logger) {
	if proposalID == "" {
		return
	}

	query := `
		UPDATE improvement_proposals 
		SET status = $1, 
		    reviewed_at = NOW(),
		    review_notes = $2
		WHERE id = $3
	`

	if err := execQuery(ctx, db, query, status, reason, proposalID); err != nil {
		logger.Warn("Failed to update proposal status", zap.Error(err))
	}
}

// ============================================================================
// HELPER FUNCTIONS - Apply Changes
// ============================================================================

func applyChange(ctx context.Context, params ActionParams, changeType string, config map[string]interface{}, clientID string) map[string]interface{} {
	result := map[string]interface{}{
		"type":    changeType,
		"success": false,
	}

	var err error

	switch changeType {
	case "add_agent":
		err = createApprovedAgent(ctx, params, config)

	case "modify_workflow":
		err = updateAgentWorkflow(ctx, params.DB, config)

	case "remove_agent", "deactivate_agent":
		if agentID, ok := config["agent_id"].(string); ok {
			err = deactivateAgentInstance(ctx, params.DB, agentID, clientID)
		} else {
			err = fmt.Errorf("agent_id required")
		}

	case "create_variant":
		err = createAgentVariant(ctx, params.DB, config, params.Logger)

	default:
		err = fmt.Errorf("unknown change type: %s", changeType)
	}

	if err != nil {
		result["error"] = err.Error()
		params.Logger.Warn("Failed to apply change",
			zap.String("type", changeType),
			zap.Error(err))
	} else {
		result["success"] = true
	}

	return result
}

func createApprovedAgent(ctx context.Context, params ActionParams, config map[string]interface{}) error {
	clientID, _ := config["client_id"].(string)
	userID, _ := config["user_id"].(string)
	if userID == "" {
		userID = "system"
	}

	_, err := SpawnAgentAction(ctx, ActionParams{
		StepConfig: models.Step{Config: config},
		Headers: map[string]string{
			"client_id": clientID,
			"user_id":   userID,
		},
		DB:       params.DB,
		Logger:   params.Logger,
		Producer: params.Producer,
	})

	return err
}

func updateAgentWorkflow(ctx context.Context, db interface{}, config map[string]interface{}) error {
	agentType, ok := config["agent_type"].(string)
	if !ok {
		return fmt.Errorf("agent_type required")
	}

	workflowUpdates, ok := config["workflow_updates"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("workflow_updates required")
	}

	updateJSON, err := json.Marshal(workflowUpdates)
	if err != nil {
		return err
	}

	// Update the agent definition's workflow
	query := `
		UPDATE agent_definitions 
		SET default_config = jsonb_set(default_config, '{workflow}', $1::jsonb),
		    updated_at = NOW()
		WHERE type = $2 AND is_active = true
	`

	return execQuery(ctx, db, query, updateJSON, agentType)
}

func deactivateAgentInstance(ctx context.Context, db interface{}, agentID, clientID string) error {
	if clientID == "" {
		return fmt.Errorf("client_id required")
	}

	query := fmt.Sprintf(`
		UPDATE client_%s.agent_instances 
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`, clientID)

	return execQuery(ctx, db, query, agentID)
}

func createAgentVariant(ctx context.Context, db interface{}, config map[string]interface{}, logger *zap.Logger) error {
	baseType, ok := config["base_agent_type"].(string)
	if !ok {
		return fmt.Errorf("base_agent_type required")
	}

	variantName, ok := config["variant_name"].(string)
	if !ok {
		return fmt.Errorf("variant_name required")
	}

	overrides, _ := config["config_overrides"].(map[string]interface{})
	overridesJSON, _ := json.Marshal(overrides)

	// Create variant in agent_variants table (if exists) or as new agent_definition
	query := `
		INSERT INTO agent_definitions (
			type, display_name, description, category,
			default_config, is_active, version,
			capabilities
		)
		SELECT 
			$1,  -- new type (base + variant name)
			display_name || ' (' || $2 || ')',
			description || ' [Variant: ' || $2 || ']',
			category,
			default_config || $3::jsonb,  -- merge overrides
			true,
			1,
			capabilities
		FROM agent_definitions
		WHERE type = $4 AND is_active = true
		ORDER BY version DESC
		LIMIT 1
	`

	newType := baseType + "-" + strings.ReplaceAll(strings.ToLower(variantName), " ", "-")

	return execQuery(ctx, db, query, newType, variantName, overridesJSON, baseType)
}

// ============================================================================
// HELPER FUNCTIONS - Proposals
// ============================================================================

func createAgentProposal(taskType string, capabilities []string) map[string]interface{} {
	// Generate proposal based on task type
	switch taskType {
	case "website_builder", "site_builder":
		return map[string]interface{}{
			"name": "Website Builder Team",
			"agents": []map[string]interface{}{
				{"role": "strategist", "agent_type": "site-strategist", "capabilities": []string{"strategy", "planning"}},
				{"role": "architect", "agent_type": "landing-page-architect", "capabilities": []string{"build", "assemble"}},
				{"role": "writer", "agent_type": "content-writer", "capabilities": []string{"content", "writing"}},
				{"role": "assembler", "agent_type": "html-assembler", "capabilities": []string{"html", "assembly"}},
				{"role": "deployer", "agent_type": "site-deployer", "capabilities": []string{"deploy", "git"}},
			},
			"estimated_cost": 1000,
		}

	case "content_creation":
		return map[string]interface{}{
			"name": "Content Creation Team",
			"agents": []map[string]interface{}{
				{"role": "researcher", "agent_type": "researcher", "capabilities": []string{"research", "analysis"}},
				{"role": "writer", "agent_type": "content-writer", "capabilities": []string{"content", "writing"}},
				{"role": "editor", "agent_type": "editor", "capabilities": []string{"editing", "review"}},
			},
			"estimated_cost": 750,
		}

	default:
		// Generate based on capabilities
		agents := []map[string]interface{}{}
		for i, cap := range capabilities {
			agents = append(agents, map[string]interface{}{
				"role":         fmt.Sprintf("agent_%d", i+1),
				"agent_type":   "generic",
				"capabilities": []string{cap},
			})
		}
		return map[string]interface{}{
			"name":           fmt.Sprintf("%s Team", taskType),
			"agents":         agents,
			"estimated_cost": 500,
		}
	}
}
