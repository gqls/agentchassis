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

// PlanAgentTeamAction - Used at the start of a workflow to plan the team
func PlanAgentTeamAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	taskType := config["task_type"].(string)
	requirements := config["requirements"].(map[string]interface{})

	// First, try to find an existing group that matches
	groupDiscovery := NewGroupDiscovery(params.DB)
	existingGroup, err := groupDiscovery.FindBestGroup(ctx, taskType, requirements)

	if err == nil && existingGroup != nil {
		// Found a good group, but check if it needs improvement
		if existingGroup.LastPerformance < 0.8 {
			// Suggest improvements but require human approval
			suggestions := analyzeGroupPerformance(ctx, params.DB, existingGroup)

			return map[string]interface{}{
				"action":                 "request_approval",
				"existing_group":         existingGroup,
				"suggested_improvements": suggestions,
				"message":                fmt.Sprintf("Found team '%s' but it could be improved. Review suggestions?", existingGroup.Name),
			}, nil
		}

		// Group is performing well, use as-is
		return map[string]interface{}{
			"action":   "use_existing",
			"group_id": existingGroup.ID,
			"agents":   existingGroup.AgentConfigs,
		}, nil
	}

	// No existing group, create a proposal
	proposal := createTeamProposal(taskType, requirements)

	return map[string]interface{}{
		"action":        "request_approval",
		"proposed_team": proposal,
		"message":       "No existing team found. Approve this new team composition?",
	}, nil
}

// ReviewPerformanceAction - Called after task completion to analyze if improvements needed
func ReviewPerformanceAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Get the execution metrics from collected data
	metricsRaw, ok := params.CollectedData["execution_metrics"]
	if !ok {
		return nil, fmt.Errorf("execution_metrics not found in collected data")
	}

	metrics, ok := metricsRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("execution_metrics is not a map")
	}

	groupIDRaw, ok := params.CollectedData["group_id"]
	if !ok {
		return nil, fmt.Errorf("group_id not found in collected data")
	}

	groupID, ok := groupIDRaw.(string)
	if !ok {
		return nil, fmt.Errorf("group_id is not a string")
	}

	// Extract metrics safely
	duration := float64(0)
	if d, ok := metrics["duration"].(float64); ok {
		duration = d
	}

	var failedSteps []string
	if fs, ok := metrics["failed_steps"].([]interface{}); ok {
		for _, step := range fs {
			if s, ok := step.(string); ok {
				failedSteps = append(failedSteps, s)
			}
		}
	}

	// Analyze what went well and what didn't
	analysis := performanceAnalysis{
		TotalDuration:   duration,
		BottleneckSteps: identifyBottlenecks(metrics),
		FailedSteps:     failedSteps,
		QualityScore:    calculateQuality(params.CollectedData),
	}

	// Only suggest changes if performance was suboptimal
	if analysis.QualityScore < 0.8 || len(analysis.BottleneckSteps) > 0 {
		suggestions := generateImprovementSuggestions(analysis)

		return map[string]interface{}{
			"needs_improvement": true,
			"analysis":          analysis,
			"suggestions":       suggestions,
			"action":            "pause_for_human_review",
		}, nil
	}

	// Performance was good, just record it
	recordGroupPerformance(ctx, params.DB, groupID, analysis)

	return map[string]interface{}{
		"needs_improvement": false,
		"analysis":          analysis,
	}, nil
}

// ApproveAgentChangesAction - Human approves proposed agent changes
// Update ApproveAgentChangesAction to handle deactivateAgent with clientID
func ApproveAgentChangesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	approvalRaw, ok := params.CollectedData["human_approval"]
	if !ok {
		return nil, fmt.Errorf("human_approval not found in collected data")
	}

	approval, ok := approvalRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("human_approval is not a map")
	}

	approved, ok := approval["approved"].(bool)
	if !ok || !approved {
		// Human rejected changes, continue with existing setup
		reason := "No reason provided"
		if r, ok := approval["rejection_reason"].(string); ok {
			reason = r
		}
		return map[string]interface{}{
			"changes_applied": false,
			"reason":          reason,
		}, nil
	}

	// Apply the approved changes
	changesRaw, ok := approval["approved_changes"]
	if !ok {
		return nil, fmt.Errorf("approved_changes not found")
	}

	changes, ok := changesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("approved_changes is not an array")
	}

	newGroupID := uuid.New().String()
	clientID := params.Headers["client_id"] // Get clientID from headers

	for _, change := range changes {
		changeMap, ok := change.(map[string]interface{})
		if !ok {
			continue
		}

		// Ensure clientID is in the changeMap
		if _, ok := changeMap["client_id"]; !ok {
			changeMap["client_id"] = clientID
		}

		changeType, _ := changeMap["type"].(string)

		switch changeType {
		case "add_agent":
			// Create new agent with specified role
			if err := createApprovedAgent(ctx, params, changeMap); err != nil {
				params.Logger.Error("Failed to create approved agent", zap.Error(err))
			}
		case "modify_workflow":
			// Update workflow configuration
			if err := updateWorkflow(ctx, params.DB, changeMap); err != nil {
				params.Logger.Error("Failed to update workflow", zap.Error(err))
			}
		case "remove_agent":
			// Mark agent as inactive with client context
			if agentID, ok := changeMap["agent_id"].(string); ok {
				if err := deactivateAgent(ctx, params.DB, agentID, clientID); err != nil {
					params.Logger.Error("Failed to deactivate agent", zap.Error(err))
				}
			}
		}
	}

	// Create new version of the group
	parentGroupID, _ := approval["parent_group_id"].(string)
	parentVersion, _ := approval["parent_version"].(string)

	if err := createGroupVersion(ctx, params.DB, newGroupID, parentGroupID, changes); err != nil {
		return nil, fmt.Errorf("failed to create group version: %w", err)
	}

	return map[string]interface{}{
		"changes_applied": true,
		"new_group_id":    newGroupID,
		"version":         incrementVersion(parentVersion),
	}, nil
}

// ConditionalRouteAction routes workflow based on conditions
func ConditionalRouteActionOld(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	conditionFieldRaw, ok := config["condition_field"]
	if !ok {
		return nil, fmt.Errorf("condition_field not specified")
	}

	conditionField, ok := conditionFieldRaw.(string)
	if !ok {
		return nil, fmt.Errorf("condition_field is not a string")
	}

	routesRaw, ok := config["routes"]
	if !ok {
		return nil, fmt.Errorf("routes not specified")
	}

	routes, ok := routesRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("routes is not a map")
	}

	// Get the condition value from collected data
	value := getNestedValue(params.CollectedData, conditionField)

	// Determine next step
	var nextStep string
	switch v := value.(type) {
	case bool:
		if v {
			if step, ok := routes["true"].(string); ok {
				nextStep = step
			}
		} else {
			if step, ok := routes["false"].(string); ok {
				nextStep = step
			}
		}
	case string:
		if step, ok := routes[v].(string); ok {
			nextStep = step
		}
	default:
		if step, ok := routes["default"].(string); ok {
			nextStep = step
		}
	}

	if nextStep == "" {
		return nil, fmt.Errorf("no route found for condition value: %v", value)
	}

	// Update workflow to use the determined next step
	if err := updateWorkflowNextStep(ctx, params.DB, params.Headers["correlation_id"], nextStep); err != nil {
		return nil, fmt.Errorf("failed to update workflow next step: %w", err)
	}

	return map[string]interface{}{
		"routed_to":       nextStep,
		"condition_value": value,
	}, nil
}

// performanceAnalysis structure
type performanceAnalysis struct {
	TotalDuration   float64
	BottleneckSteps []string
	FailedSteps     []string
	QualityScore    float64
}

func identifyBottlenecks(metrics map[string]interface{}) []string {
	bottlenecks := []string{}

	stepDurationsRaw, ok := metrics["step_durations"]
	if !ok {
		return bottlenecks
	}

	stepDurations, ok := stepDurationsRaw.(map[string]interface{})
	if !ok {
		return bottlenecks
	}

	// Calculate average duration
	totalDuration := 0.0
	count := 0
	for _, durationRaw := range stepDurations {
		if duration, ok := durationRaw.(float64); ok {
			totalDuration += duration
			count++
		}
	}

	if count == 0 {
		return bottlenecks
	}

	avgDuration := totalDuration / float64(count)
	threshold := avgDuration * 2

	// Find steps that exceed threshold
	for step, durationRaw := range stepDurations {
		if duration, ok := durationRaw.(float64); ok {
			if duration > threshold {
				bottlenecks = append(bottlenecks, step)
			}
		}
	}

	return bottlenecks
}

func calculateQuality(collectedData map[string]interface{}) float64 {
	quality := 1.0

	// Check for errors
	if errorsRaw, ok := collectedData["errors"]; ok {
		if errors, ok := errorsRaw.([]interface{}); ok {
			quality -= float64(len(errors)) * 0.1
		}
	}

	// Check for human feedback
	if feedbackRaw, ok := collectedData["human_feedback"]; ok {
		if feedback, ok := feedbackRaw.(map[string]interface{}); ok {
			if rating, ok := feedback["rating"].(float64); ok {
				quality = rating / 5.0
			}
		}
	}

	// Ensure quality stays within bounds
	if quality < 0 {
		quality = 0
	} else if quality > 1 {
		quality = 1
	}

	return quality
}

func generateImprovementSuggestions(analysis performanceAnalysis) []map[string]interface{} {
	suggestions := []map[string]interface{}{}

	if len(analysis.BottleneckSteps) > 0 {
		suggestions = append(suggestions, map[string]interface{}{
			"type":   "add_parallel_agents",
			"target": analysis.BottleneckSteps,
			"reason": "Steps taking too long",
		})
	}

	if analysis.QualityScore < 0.7 {
		suggestions = append(suggestions, map[string]interface{}{
			"type":   "add_quality_checker",
			"reason": "Output quality below threshold",
		})
	}

	if len(analysis.FailedSteps) > 0 {
		suggestions = append(suggestions, map[string]interface{}{
			"type":   "add_retry_mechanism",
			"target": analysis.FailedSteps,
			"reason": "Steps are failing",
		})
	}

	return suggestions
}

// recordGroupPerformance to handle both DB types
func recordGroupPerformance(ctx context.Context, db interface{}, groupID string, analysis performanceAnalysis) {
	performanceData := map[string]interface{}{
		"duration":     analysis.TotalDuration,
		"quality":      analysis.QualityScore,
		"bottlenecks":  analysis.BottleneckSteps,
		"failed_steps": analysis.FailedSteps,
		"timestamp":    time.Now(),
	}

	performanceJSON, err := json.Marshal(performanceData)
	if err != nil {
		return
	}

	query := `
        UPDATE agent_groups 
        SET performance_metrics = jsonb_set(
            COALESCE(performance_metrics, '{}'),
            '{last_execution}',
            $1::jsonb
        ),
        last_used_at = NOW()
        WHERE id = $2
    `

	switch d := db.(type) {
	case *sql.DB:
		_, err = d.ExecContext(ctx, query, performanceJSON, groupID)
	case *pgxpool.Pool:
		_, err = d.Exec(ctx, query, performanceJSON, groupID)
	}

	if err != nil {
		// Log error but don't fail
		// In production, use proper logging
	}
}

// Helper functions

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

// updateWorkflowNextStep to handle both DB types
func updateWorkflowNextStep(ctx context.Context, db interface{}, correlationID string, nextStep string) error {
	query := `
        UPDATE orchestrator_state 
        SET current_step = $1, updated_at = NOW()
        WHERE correlation_id = $2
    `

	var err error
	switch d := db.(type) {
	case *sql.DB:
		_, err = d.ExecContext(ctx, query, nextStep, correlationID)
	case *pgxpool.Pool:
		_, err = d.Exec(ctx, query, nextStep, correlationID)
	default:
		err = fmt.Errorf("unsupported database type: %T", db)
	}

	return err
}

func createTeamProposal(taskType string, requirements map[string]interface{}) map[string]interface{} {
	switch taskType {
	case "website_builder":
		return map[string]interface{}{
			"name": "Website Builder Team",
			"agents": []map[string]interface{}{
				{"role": "architect", "agent_type": "site-architect"},
				{"role": "designer", "agent_type": "visual-designer"},
				{"role": "developer", "agent_type": "html-developer"},
				{"role": "publisher", "agent_type": "site-publisher"},
			},
			"estimated_fuel": 1000,
		}
	case "content_creation":
		return map[string]interface{}{
			"name": "Content Creation Team",
			"agents": []map[string]interface{}{
				{"role": "researcher", "agent_type": "researcher"},
				{"role": "writer", "agent_type": "content-creator"},
				{"role": "editor", "agent_type": "editor"},
			},
			"estimated_fuel": 750,
		}
	default:
		return map[string]interface{}{
			"name": fmt.Sprintf("%s Team", taskType),
			"agents": []map[string]interface{}{
				{"role": "coordinator", "agent_type": "generic"},
			},
			"estimated_fuel": 500,
		}
	}
}

func analyzeGroupPerformance(ctx context.Context, db interface{}, group *discovery.AgentGroup) []map[string]interface{} {
	suggestions := []map[string]interface{}{}

	if group.LastPerformance < 0.8 {
		suggestions = append(suggestions, map[string]interface{}{
			"type":   "optimize_workflow",
			"reason": "Performance below threshold",
			"impact": "medium",
		})
	}

	// In a real implementation, would query historical data
	// and analyze patterns

	return suggestions
}

func createApprovedAgent(ctx context.Context, params ActionParams, config map[string]interface{}) error {
	clientID, ok := config["client_id"].(string)
	if !ok {
		return fmt.Errorf("client_id not found in config")
	}

	userID, ok := config["user_id"].(string)
	if !ok {
		userID = "system"
	}

	_, err := SpawnAgentAction(ctx, ActionParams{
		StepConfig: models.Step{
			Config: config,
		},
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

func updateWorkflow(ctx context.Context, db interface{}, config map[string]interface{}) error {
	agentID, ok := config["agent_id"].(string)
	if !ok {
		return fmt.Errorf("agent_id not found")
	}

	clientID, ok := config["client_id"].(string)
	if !ok {
		return fmt.Errorf("client_id not found")
	}

	workflowUpdates, ok := config["workflow_updates"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("workflow_updates not found")
	}

	updateJSON, err := json.Marshal(workflowUpdates)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
        UPDATE client_%s.agent_instances 
        SET config = jsonb_set(config, '{workflow}', $1::jsonb),
            updated_at = NOW()
        WHERE id = $2
    `, clientID)

	switch d := db.(type) {
	case *sql.DB:
		_, err = d.ExecContext(ctx, query, updateJSON, agentID)
	case *pgxpool.Pool:
		_, err = d.Exec(ctx, query, updateJSON, agentID)
	default:
		err = fmt.Errorf("unsupported database type: %T", db)
	}

	return err
}

// deactivateAgent needs client context
func deactivateAgent(ctx context.Context, db interface{}, agentID string, clientID string) error {
	if clientID == "" {
		return fmt.Errorf("client_id is required")
	}

	// Build client-specific query
	query := fmt.Sprintf(`
        UPDATE client_%s.agent_instances 
        SET is_active = false, updated_at = NOW()
        WHERE id = $1
    `, clientID)

	var err error
	switch d := db.(type) {
	case *sql.DB:
		_, err = d.ExecContext(ctx, query, agentID)
	case *pgxpool.Pool:
		_, err = d.Exec(ctx, query, agentID)
	default:
		err = fmt.Errorf("unsupported database type: %T", db)
	}

	return err
}

func createGroupVersion(ctx context.Context, db interface{}, newGroupID string, parentGroupID string, changes []interface{}) error {
	if parentGroupID == "" {
		return fmt.Errorf("parent group ID is required")
	}

	// Build mutation history entry
	mutationEntry := map[string]interface{}{
		"timestamp": time.Now(),
		"type":      "evolution",
		"changes":   changes,
		"parent_id": parentGroupID,
	}

	mutationJSON, err := json.Marshal([]interface{}{mutationEntry})
	if err != nil {
		return err
	}

	// Query to create new version with mutation history
	query := `
        INSERT INTO agent_groups (id, name, group_type, parent_id, version, 
                                  agent_configs, orchestration_workflow, capabilities,
                                  tags, mutation_history)
        SELECT 
            $1, 
            name || ' (v' || COALESCE(
                (SELECT COALESCE(MAX(CAST(SUBSTRING(version FROM '^[0-9]+') AS INT)), 0) + 1 
                 FROM agent_groups WHERE parent_id = $2), 
                2
            )::text || '.0.0)',
            group_type,
            $2,
            COALESCE(
                (SELECT COALESCE(MAX(CAST(SUBSTRING(version FROM '^[0-9]+') AS INT)), 0) + 1 
                 FROM agent_groups WHERE parent_id = $2), 
                2
            )::text || '.0.0',
            agent_configs,
            orchestration_workflow,
            capabilities,
            tags,
            $3::jsonb
        FROM agent_groups 
        WHERE id = $2
    `

	switch d := db.(type) {
	case *sql.DB:
		_, err = d.ExecContext(ctx, query, newGroupID, parentGroupID, mutationJSON)
	case *pgxpool.Pool:
		_, err = d.Exec(ctx, query, newGroupID, parentGroupID, mutationJSON)
	default:
		err = fmt.Errorf("unsupported database type: %T", db)
	}

	return err
}

func incrementVersion(currentVersion string) string {
	if currentVersion == "" {
		return "1.0.0"
	}

	var major, minor, patch int
	n, _ := fmt.Sscanf(currentVersion, "%d.%d.%d", &major, &minor, &patch)

	if n == 3 {
		return fmt.Sprintf("%d.%d.%d", major, minor+1, patch)
	}

	return currentVersion + ".1"
}

// NewGroupDiscovery creates a discovery service from the database connection
func NewGroupDiscovery(db interface{}) *discovery.GroupDiscovery {
	switch d := db.(type) {
	case *sql.DB:
		// Create a pgxpool from sql.DB
		connStr := "postgres://user:pass@localhost/db" // This needs proper config
		config, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			return nil
		}
		pool, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			return nil
		}
		return discovery.NewGroupDiscovery(pool)
	case *pgxpool.Pool:
		return discovery.NewGroupDiscovery(d)
	default:
		return nil
	}
}
