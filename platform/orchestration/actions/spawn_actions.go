// platform/orchestration/actions/spawn_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/discovery"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// AgentMatch represents a discovered agent
type AgentMatch struct {
	AgentID          string
	AgentType        string
	AgentName        string
	Capabilities     []string
	PerformanceScore float64
}

// AgentRecommendation represents a recommended agent for a task
type AgentRecommendation struct {
	AgentType            string
	DisplayName          string
	Category             string
	Capabilities         []string
	PerformanceScore     float64
	RecommendationReason string
}

// EnhancedDiscovery wraps database connections and provides discovery methods
type EnhancedDiscovery struct {
	db interface{} // Can be *sql.DB or *pgxpool.Pool
}

// SpawnAgentAction creates an agent instance in DB and spawns a Kubernetes Job
func SpawnAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	agentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified")
	}

	clientID := params.Headers["client_id"]
	if clientID == "" {
		return nil, fmt.Errorf("client_id not specified in headers")
	}

	params.Logger.Info("Spawning agent",
		zap.String("agent_type", agentType),
		zap.String("client_id", clientID))

	// Check for existing agent
	existingAgent := checkExistingAgent(ctx, params.DB, clientID, agentType)
	if existingAgent != nil {
		// Check if its Kubernetes job is still running
		if isAgentJobRunning(ctx, existingAgent.ID, params.Logger) {
			params.Logger.Info("Reusing existing running agent",
				zap.String("agent_id", existingAgent.ID),
				zap.String("agent_type", agentType))

			return map[string]interface{}{
				"agent_id": existingAgent.ID,
				"topic":    fmt.Sprintf("system.agent.%s.process", agentType),
				"status":   "reused",
			}, nil
		}
		// Agent exists but job not running - will spawn new job below
		params.Logger.Info("Agent exists but job not running, will spawn new job",
			zap.String("agent_id", existingAgent.ID))
	}

	// Create or reuse agent ID
	agentID := uuid.New().String()
	if existingAgent != nil {
		agentID = existingAgent.ID
	} else {
		// Create new agent in database
		if err := createAgentInDB(ctx, params, agentID, agentType, clientID); err != nil {
			return nil, fmt.Errorf("failed to create agent in database: %w", err)
		}
	}

	// Spawn the Kubernetes Job
	jobName, err := spawnAgentKubernetesJob(ctx, agentID, agentType, clientID, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to spawn agent job",
			zap.Error(err),
			zap.String("agent_id", agentID),
			zap.String("agent_type", agentType))
		// Don't fail - agent exists in DB and can be spawned manually
		return map[string]interface{}{
			"agent_id": agentID,
			"topic":    fmt.Sprintf("system.agent.%s.process", agentType),
			"status":   "created_without_job",
			"error":    err.Error(),
		}, nil
	}

	params.Logger.Info("Successfully spawned agent job",
		zap.String("job_name", jobName),
		zap.String("agent_id", agentID),
		zap.String("agent_type", agentType))

	return map[string]interface{}{
		"agent_id": agentID,
		"topic":    fmt.Sprintf("system.agent.%s.process", agentType),
		"status":   "spawned",
		"job_name": jobName,
	}, nil
}

// SpawnGroupAction spawns a complete agent group
func SpawnGroupAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	groupType, ok := config["group_type"].(string)
	if !ok {
		return nil, fmt.Errorf("group_type not specified")
	}

	// Find the best matching group
	var groupID, groupName string
	var agentConfigs json.RawMessage
	var workflow json.RawMessage

	// Fix: Use QueryRowContext with context as first parameter
	err := params.DB.QueryRowContext(ctx, `
        SELECT id, name, agent_configs, orchestration_workflow
        FROM agent_groups
        WHERE group_type = $1
        ORDER BY usage_count DESC, version DESC
        LIMIT 1
    `, groupType).Scan(&groupID, &groupName, &agentConfigs, &workflow)

	if err != nil {
		return nil, fmt.Errorf("no group found for type %s: %w", groupType, err)
	}

	// Parse agent configs
	var agents []map[string]interface{}
	if err := json.Unmarshal(agentConfigs, &agents); err != nil {
		return nil, fmt.Errorf("failed to parse agent configs: %w", err)
	}

	// which agents are we about to spawn
	params.Logger.Info("Retrieved workflow from database",
		zap.String("group_id", groupID),
		zap.String("workflow_raw", string(workflow)),
		zap.String("agentConfigs full json", string(agentConfigs)))

	// Spawn each agent in the group
	spawnedAgents := make(map[string]string)

	for _, agentConfig := range agents {
		role, ok := agentConfig["role"].(string)
		if !ok {
			continue
		}

		agentType, ok := agentConfig["agent_type"].(string)
		if !ok {
			continue
		}

		result, err := SpawnAgentAction(ctx, ActionParams{
			StepConfig: models.Step{
				Config: map[string]interface{}{
					"agent_type": agentType,
				},
			},
			Headers: params.Headers,
			DB:      params.DB,
			Logger:  params.Logger,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to spawn %s: %w", role, err)
		}

		agentResult, ok := result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected result type from SpawnAgentAction")
		}

		agentID, ok := agentResult["agent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("agent_id not found in spawn result")
		}

		spawnedAgents[role] = agentID
		params.Logger.Info("Agent id for role.",
			zap.String("role", role),
			zap.String("agentID", agentID),
			zap.String("spawnedAgents", fmt.Sprintf("%v", spawnedAgents)),
		)
	}

	// Update group usage - Fix: Use ExecContext
	_, err = params.DB.ExecContext(ctx, `
        UPDATE agent_groups 
        SET usage_count = usage_count + 1, 
            last_used_at = NOW() 
        WHERE id = $1
    `, groupID)

	if err != nil && params.Logger != nil {
		params.Logger.Error("Failed to update group usage", zap.Error(err))
	}

	return map[string]interface{}{
		"group_id":   groupID,
		"group_name": groupName,
		"agents":     spawnedAgents,
		"workflow":   workflow,
	}, nil
}

// callSpecificAgent calls a specific agent by ID
func callSpecificAgent(ctx context.Context, params ActionParams, agentID string) (interface{}, error) {
	clientID := params.Headers["client_id"]
	if clientID == "" {
		return nil, fmt.Errorf("client_id not specified")
	}

	// Get agent configuration to find its topic
	var agentConfig json.RawMessage

	query := fmt.Sprintf(`
        SELECT config FROM client_%s.agent_instances 
        WHERE id = $1
    `, clientID)

	err := params.DB.QueryRowContext(ctx, query, agentID).Scan(&agentConfig)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(agentConfig, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent config: %w", err)
	}

	topic, ok := config["topic"].(string)
	if !ok {
		return nil, fmt.Errorf("topic not found in agent config")
	}

	// Prepare payload
	payload := models.TaskRequest{
		Action: params.StepConfig.Action,
		Data:   params.CollectedData,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create headers
	newRequestID := uuid.New().String()
	outHeaders := make(map[string]string)
	for k, v := range params.Headers {
		outHeaders[k] = v
	}
	outHeaders["causation_id"] = params.Headers["request_id"]
	outHeaders["request_id"] = newRequestID
	outHeaders["target_agent_id"] = agentID

	// Send message
	err = params.Producer.Produce(ctx, topic, outHeaders,
		[]byte(params.Headers["correlation_id"]), payloadBytes)

	if err != nil {
		return nil, fmt.Errorf("failed to call agent: %w", err)
	}

	return map[string]interface{}{
		"agent_called": agentID,
		"request_id":   newRequestID,
		"topic":        topic,
	}, nil
}

// NewAgentDiscovery creates a discovery service from the database connection
func NewAgentDiscovery(db interface{}) *EnhancedDiscovery {
	return &EnhancedDiscovery{db: db}
}

// FindAgentsByCapability uses the SQL function to find agents
func (d *EnhancedDiscovery) FindAgentsByCapability(ctx context.Context, capabilities []string, clientID string) ([]AgentMatch, error) {
	// Convert capabilities to PostgreSQL array format
	capArray := "{" + strings.Join(capabilities, ",") + "}"

	query := `SELECT * FROM find_agents_by_capability($1::text[], $2)`

	var matches []AgentMatch

	// Handle both database types
	switch db := d.db.(type) {
	case *sql.DB:
		rows, err := db.QueryContext(ctx, query, capArray, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to find agents by capability: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var match AgentMatch
			var capabilitiesJSON []byte

			err := rows.Scan(
				&match.AgentID,
				&match.AgentType,
				&match.AgentName,
				&capabilitiesJSON,
				&match.PerformanceScore,
			)
			if err != nil {
				return nil, err
			}

			// Parse capabilities JSON
			if err := json.Unmarshal(capabilitiesJSON, &match.Capabilities); err != nil {
				return nil, err
			}

			matches = append(matches, match)
		}

		return matches, rows.Err()

	case *pgxpool.Pool:
		rows, err := db.Query(ctx, query, capArray, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to find agents by capability: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var match AgentMatch
			var capabilitiesJSON []byte

			err := rows.Scan(
				&match.AgentID,
				&match.AgentType,
				&match.AgentName,
				&capabilitiesJSON,
				&match.PerformanceScore,
			)
			if err != nil {
				return nil, err
			}

			// Parse capabilities JSON
			if err := json.Unmarshal(capabilitiesJSON, &match.Capabilities); err != nil {
				return nil, err
			}

			matches = append(matches, match)
		}

		return matches, rows.Err()

	default:
		return nil, fmt.Errorf("unsupported database type: %T", d.db)
	}
}

// RecommendAgentsForTask uses the SQL function to get agent recommendations
func (d *EnhancedDiscovery) RecommendAgentsForTask(ctx context.Context, taskType string, capabilities []string) ([]AgentRecommendation, error) {
	var capArray interface{}
	if len(capabilities) > 0 {
		capArray = "{" + strings.Join(capabilities, ",") + "}"
	} else {
		capArray = nil
	}

	query := `SELECT * FROM recommend_agents_for_task($1, $2::text[])`

	var recommendations []AgentRecommendation

	switch db := d.db.(type) {
	case *sql.DB:
		rows, err := db.QueryContext(ctx, query, taskType, capArray)
		if err != nil {
			return nil, fmt.Errorf("failed to get recommendations: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var rec AgentRecommendation
			var capabilitiesJSON []byte

			err := rows.Scan(
				&rec.AgentType,
				&rec.DisplayName,
				&rec.Category,
				&capabilitiesJSON,
				&rec.PerformanceScore,
				&rec.RecommendationReason,
			)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(capabilitiesJSON, &rec.Capabilities); err != nil {
				return nil, err
			}

			recommendations = append(recommendations, rec)
		}

		return recommendations, rows.Err()

	case *pgxpool.Pool:
		rows, err := db.Query(ctx, query, taskType, capArray)
		if err != nil {
			return nil, fmt.Errorf("failed to get recommendations: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var rec AgentRecommendation
			var capabilitiesJSON []byte

			err := rows.Scan(
				&rec.AgentType,
				&rec.DisplayName,
				&rec.Category,
				&capabilitiesJSON,
				&rec.PerformanceScore,
				&rec.RecommendationReason,
			)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(capabilitiesJSON, &rec.Capabilities); err != nil {
				return nil, err
			}

			recommendations = append(recommendations, rec)
		}

		return recommendations, rows.Err()

	default:
		return nil, fmt.Errorf("unsupported database type: %T", d.db)
	}
}

// DiscoverAgents provides backward compatibility with existing code
func (d *EnhancedDiscovery) DiscoverAgents(ctx context.Context, requirements discovery.Requirements) ([]discovery.AgentMatch, error) {
	// Use the new FindAgentsByCapability function
	matches, err := d.FindAgentsByCapability(ctx, requirements.Capabilities, requirements.ClientID)
	if err != nil {
		return nil, err
	}

	// Convert to the expected return type
	var result []discovery.AgentMatch
	for _, m := range matches {
		result = append(result, discovery.AgentMatch{
			AgentID:     m.AgentID,
			AgentType:   m.AgentType,
			Performance: m.PerformanceScore,
		})
	}

	// Filter by agent type if specified
	if requirements.AgentType != "" {
		var filtered []discovery.AgentMatch
		for _, m := range result {
			if m.AgentType == requirements.AgentType {
				filtered = append(filtered, m)
			}
		}
		result = filtered
	}

	return result, nil
}

// GetAgentPerformanceSummary gets performance metrics for agents
func (d *EnhancedDiscovery) GetAgentPerformanceSummary(ctx context.Context, agentType string, limit int) ([]map[string]interface{}, error) {
	query := `SELECT * FROM get_agent_performance_summary($1, $2)`

	var results []map[string]interface{}

	switch db := d.db.(type) {
	case *sql.DB:
		rows, err := db.QueryContext(ctx, query, agentType, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get performance summary: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var agentID string
			var agentType string
			var totalTasks int
			var successRate float64
			var avgResponseTime int
			var avgFuelPerTask int
			var avgQualityScore *float64

			err := rows.Scan(
				&agentID,
				&agentType,
				&totalTasks,
				&successRate,
				&avgResponseTime,
				&avgFuelPerTask,
				&avgQualityScore,
			)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"agent_id":          agentID,
				"agent_type":        agentType,
				"total_tasks":       totalTasks,
				"success_rate":      successRate,
				"avg_response_time": avgResponseTime,
				"avg_fuel_per_task": avgFuelPerTask,
			}

			if avgQualityScore != nil {
				result["avg_quality_score"] = *avgQualityScore
			}

			results = append(results, result)
		}

		return results, rows.Err()

	case *pgxpool.Pool:
		rows, err := db.Query(ctx, query, agentType, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get performance summary: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var agentID string
			var agentType string
			var totalTasks int
			var successRate float64
			var avgResponseTime int
			var avgFuelPerTask int
			var avgQualityScore *float64

			err := rows.Scan(
				&agentID,
				&agentType,
				&totalTasks,
				&successRate,
				&avgResponseTime,
				&avgFuelPerTask,
				&avgQualityScore,
			)
			if err != nil {
				return nil, err
			}

			result := map[string]interface{}{
				"agent_id":          agentID,
				"agent_type":        agentType,
				"total_tasks":       totalTasks,
				"success_rate":      successRate,
				"avg_response_time": avgResponseTime,
				"avg_fuel_per_task": avgFuelPerTask,
			}

			if avgQualityScore != nil {
				result["avg_quality_score"] = *avgQualityScore
			}

			results = append(results, result)
		}

		return results, rows.Err()

	default:
		return nil, fmt.Errorf("unsupported database type: %T", d.db)
	}
}

// Updated CallAgentAction to use the new discovery with recommendations
func CallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	// If specific agent_id provided, use it directly
	if agentID, ok := config["agent_id"].(string); ok {
		return callSpecificAgent(ctx, params, agentID)
	}

	// Otherwise, discover best agent
	agentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified")
	}

	// Try to use recommendations first
	discover := NewAgentDiscovery(params.DB)
	if discover == nil {
		return nil, fmt.Errorf("failed to create discovery service")
	}

	// Get capabilities from config if provided
	var capabilities []string
	if caps, ok := config["required_capabilities"].([]interface{}); ok {
		for _, cap := range caps {
			if c, ok := cap.(string); ok {
				capabilities = append(capabilities, c)
			}
		}
	}

	// Try to get recommendations
	recommendations, err := discover.RecommendAgentsForTask(ctx, agentType, capabilities)
	if err == nil && len(recommendations) > 0 {
		// Use the best recommendation
		bestRec := recommendations[0]
		params.Logger.Info("Using recommended agent",
			zap.String("agent_type", bestRec.AgentType),
			zap.String("reason", bestRec.RecommendationReason),
			zap.Float64("performance_score", bestRec.PerformanceScore))

		// Update agent type to the recommended one
		agentType = bestRec.AgentType
	}

	// Try to find existing agent using backward-compatible method
	matches, err := discover.DiscoverAgents(ctx, discovery.Requirements{
		AgentType:    agentType,
		ClientID:     params.Headers["client_id"],
		Capabilities: capabilities,
	})

	if err != nil || len(matches) == 0 {
		// No agent found, spawn one
		params.Logger.Info("No existing agent found, spawning new one",
			zap.String("agent_type", agentType))

		spawnResult, err := SpawnAgentAction(ctx, ActionParams{
			StepConfig: models.Step{
				Config: map[string]interface{}{
					"agent_type": agentType,
				},
			},
			Headers: params.Headers,
			DB:      params.DB,
			Logger:  params.Logger,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to spawn agent: %w", err)
		}

		sr, ok := spawnResult.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected spawn result type")
		}

		agentID, ok := sr["agent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("agent_id not found in spawn result")
		}

		return callSpecificAgent(ctx, params, agentID)
	}

	// Use best match
	bestMatch := matches[0]
	params.Logger.Info("Using existing agent",
		zap.String("agent_id", bestMatch.AgentID),
		zap.String("agent_type", bestMatch.AgentType),
		zap.Float64("performance_score", bestMatch.Performance))

	return callSpecificAgent(ctx, params, bestMatch.AgentID)
}

// Updated DiscoverAgentsAction to use the new enhanced discovery
func DiscoverAgentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	agentType, ok := config["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_type not specified")
	}

	discover := NewAgentDiscovery(params.DB)
	if discover == nil {
		return nil, fmt.Errorf("failed to create discovery service")
	}

	// Get capabilities if specified
	var capabilities []string
	if caps, ok := config["capabilities"].([]string); ok {
		capabilities = caps
	} else if caps, ok := config["capabilities"].([]interface{}); ok {
		for _, cap := range caps {
			if c, ok := cap.(string); ok {
				capabilities = append(capabilities, c)
			}
		}
	}

	// Try recommendations first
	recommendations, err := discover.RecommendAgentsForTask(ctx, agentType, capabilities)
	if err == nil && len(recommendations) > 0 {
		params.Logger.Info("Found agent recommendations",
			zap.Int("count", len(recommendations)))
	}

	// Also get existing agents
	requirements := discovery.Requirements{
		AgentType:    agentType,
		ClientID:     params.Headers["client_id"],
		Capabilities: capabilities,
	}

	matches, err := discover.DiscoverAgents(ctx, requirements)
	if err != nil {
		return nil, err
	}

	// Get performance summary if requested
	var performanceSummary []map[string]interface{}
	if includePerf, ok := config["include_performance"].(bool); ok && includePerf {
		performanceSummary, _ = discover.GetAgentPerformanceSummary(ctx, agentType, 10)
	}

	return map[string]interface{}{
		"found_agents":     len(matches),
		"agents":           matches,
		"recommendations":  recommendations,
		"performance_data": performanceSummary,
	}, nil
}

// Helper functions

// Add this updated function to spawn_actions.go

func buildWorkflowForType(agentType string, db interface{}) map[string]interface{} {
	ctx := context.Background()

	var workflowJSON []byte
	var err error

	// Handle both database types
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, `
            SELECT default_config->'workflow'
            FROM agent_definitions
            WHERE type = $1
        `, agentType).Scan(&workflowJSON)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, `
            SELECT default_config->'workflow'
            FROM agent_definitions
            WHERE type = $1
        `, agentType).Scan(&workflowJSON)
	default:
		// Fallback to default workflow
		return getDefaultWorkflowForType(agentType)
	}

	if err == nil && len(workflowJSON) > 0 {
		var workflow map[string]interface{}
		if err := json.Unmarshal(workflowJSON, &workflow); err == nil {
			return workflow
		}
	}

	// Fallback
	return getDefaultWorkflowForType(agentType)
}

// getDefaultWorkflowForType returns a default workflow when database lookup fails
func getDefaultWorkflowForType(agentType string) map[string]interface{} {
	// Define specific defaults for known agent types
	switch agentType {
	case "domain-analyst":
		return map[string]interface{}{
			"start_step": "analyze",
			"steps": map[string]interface{}{
				"analyze": map[string]interface{}{
					"action":      "execute_llm_prompt",
					"description": "Analyze the domain",
					"config": map[string]interface{}{
						"prompt_template": "Analyze the domain {{.input.domain}} and provide business insights.",
					},
					"next_step": "complete",
				},
				"complete": map[string]interface{}{
					"action":      "complete_workflow",
					"description": "Complete the analysis",
				},
			},
		}

	case "site-architect":
		return map[string]interface{}{
			"start_step": "design",
			"steps": map[string]interface{}{
				"design": map[string]interface{}{
					"action":      "execute_llm_prompt",
					"description": "Design site structure",
					"config": map[string]interface{}{
						"prompt_template": "Design a website structure for {{.input.business_name}}.",
					},
					"next_step": "complete",
				},
				"complete": map[string]interface{}{
					"action": "complete_workflow",
				},
			},
		}

	case "content-creator":
		return map[string]interface{}{
			"start_step": "generate",
			"steps": map[string]interface{}{
				"generate": map[string]interface{}{
					"action":      "execute_llm_prompt",
					"description": "Generate content",
					"config": map[string]interface{}{
						"prompt_template": "Create website content for {{.input.business_name}}.",
					},
					"next_step": "complete",
				},
				"complete": map[string]interface{}{
					"action": "complete_workflow",
				},
			},
		}

	case "html-developer":
		return map[string]interface{}{
			"start_step": "develop",
			"steps": map[string]interface{}{
				"develop": map[string]interface{}{
					"action":      "execute_llm_prompt",
					"description": "Generate HTML code",
					"config": map[string]interface{}{
						"prompt_template": "Generate HTML for {{.input.page_name}} page.",
					},
					"next_step": "complete",
				},
				"complete": map[string]interface{}{
					"action": "complete_workflow",
				},
			},
		}

	case "visual-designer":
		return map[string]interface{}{
			"start_step": "design_visuals",
			"steps": map[string]interface{}{
				"design_visuals": map[string]interface{}{
					"action":      "execute_llm_prompt",
					"description": "Create visual design specs",
					"config": map[string]interface{}{
						"prompt_template": "Create visual design specifications for {{.input.business_name}}.",
					},
					"next_step": "complete",
				},
				"complete": map[string]interface{}{
					"action": "complete_workflow",
				},
			},
		}

	case "site-publisher":
		return map[string]interface{}{
			"start_step": "publish",
			"steps": map[string]interface{}{
				"publish": map[string]interface{}{
					"action":      "deploy_to_hosting",
					"description": "Deploy the website",
					"config": map[string]interface{}{
						"platform": "netlify",
						"auto_ssl": true,
					},
					"next_step": "complete",
				},
				"complete": map[string]interface{}{
					"action": "complete_workflow",
				},
			},
		}

	case "website-builder":
		// Orchestrator workflow
		return map[string]interface{}{
			"start_step": "validate",
			"steps": map[string]interface{}{
				"validate": map[string]interface{}{
					"action":      "validate_input",
					"description": "Validate the request",
					"next_step":   "spawn_team",
				},
				"spawn_team": map[string]interface{}{
					"action":      "spawn_group",
					"description": "Spawn the website builder team",
					"config": map[string]interface{}{
						"group_type": "website-builder",
					},
					"next_step": "orchestrate",
				},
				"orchestrate": map[string]interface{}{
					"action":      "start_orchestration",
					"description": "Start the orchestration",
					"next_step":   "complete",
				},
				"complete": map[string]interface{}{
					"action": "complete_workflow",
				},
			},
		}

	case "orchestrator", "generic":
		// Generic orchestrator workflow
		return map[string]interface{}{
			"start_step": "process",
			"steps": map[string]interface{}{
				"process": map[string]interface{}{
					"action":      "validate_input",
					"description": "Process the request",
					"next_step":   "execute",
				},
				"execute": map[string]interface{}{
					"action":      "transform_data",
					"description": "Execute the task",
					"config": map[string]interface{}{
						"transformation": "uppercase",
					},
					"next_step": "respond",
				},
				"respond": map[string]interface{}{
					"action":      "send_notification",
					"description": "Send response",
					"next_step":   "complete",
				},
				"complete": map[string]interface{}{
					"action": "complete_workflow",
				},
			},
		}

	case "reasoning", "web-search", "image-generator":
		// Adapter agents that call external services
		return map[string]interface{}{
			"start_step": "call_service",
			"steps": map[string]interface{}{
				"call_service": map[string]interface{}{
					"action":      "http_request",
					"description": "Call external service",
					"config": map[string]interface{}{
						"url":    getServiceURL(agentType),
						"method": "POST",
					},
					"next_step": "complete",
				},
				"complete": map[string]interface{}{
					"action": "complete_workflow",
				},
			},
		}

	default:
		// Generic fallback for unknown agent types
		return map[string]interface{}{
			"start_step": "process",
			"steps": map[string]interface{}{
				"process": map[string]interface{}{
					"action":      "execute_llm_prompt",
					"description": "Process the task",
					"config": map[string]interface{}{
						"prompt_template": "Process this request: {{.input}}",
					},
					"next_step": "respond",
				},
				"respond": map[string]interface{}{
					"action":      "send_notification",
					"description": "Send response",
					"next_step":   "complete",
				},
				"complete": map[string]interface{}{
					"action":      "complete_workflow",
					"description": "Complete the workflow",
				},
			},
		}
	}
}

// Helper function to get service URLs for adapter agents
func getServiceURL(agentType string) string {
	switch agentType {
	case "reasoning":
		return "http://reasoning-service.ai-persona-system.svc.cluster.local:8090/reason"
	case "web-search":
		return "http://web-search-service.ai-persona-system.svc.cluster.local:8091/search"
	case "image-generator":
		return "http://image-generator-service.ai-persona-system.svc.cluster.local:8092/generate"
	default:
		return "http://localhost:8080/process"
	}
}

// Update createAgentInDB to use the new function
func createAgentInDB(ctx context.Context, params ActionParams, agentID, agentType, clientID string) error {
	// Pass the database connection to buildWorkflowForType
	workflow := buildWorkflowForType(agentType, params.DB)

	agentConfig := map[string]interface{}{
		"agent_type":   agentType,
		"workflow":     workflow,
		"topic":        fmt.Sprintf("system.agent.%s.process", agentType),
		"capabilities": getCapabilitiesForType(agentType),
	}

	configJSON, err := json.Marshal(agentConfig)
	if err != nil {
		return err
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO client_%s.agent_instances 
		(id, template_id, owner_user_id, name, config, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, clientID)

	userID := params.Headers["user_id"]
	if userID == "" {
		userID = "system"
	}

	_, err = params.DB.ExecContext(ctx, insertQuery,
		agentID,
		"2a540b98-85d5-4762-a692-538bcf1be395", // generic template
		userID,
		fmt.Sprintf("%s-%s", agentType, time.Now().Format("20060102-150405")),
		configJSON,
	)

	return err
}

func getCapabilitiesForType(agentType string) []string {
	capabilities := map[string][]string{
		"copywriter":      {"writing", "marketing", "creative"},
		"researcher":      {"research", "analysis", "data"},
		"developer":       {"coding", "html", "css", "javascript"},
		"designer":        {"design", "ui", "ux", "graphics"},
		"domain-analyst":  {"analysis", "research", "categorization"},
		"site-architect":  {"planning", "structure", "navigation"},
		"html-developer":  {"html", "css", "javascript", "frontend"},
		"visual-designer": {"design", "graphics", "branding"},
		"site-publisher":  {"deployment", "hosting", "publishing"},
	}

	if caps, ok := capabilities[agentType]; ok {
		return caps
	}
	return []string{agentType}
}

func spawnAgentJob(ctx context.Context, agentID, agentType, clientID string, logger *zap.Logger) (string, error) {
	// Get Kubernetes client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Generate unique job name
	jobName := fmt.Sprintf("agent-%s-%s-%d", agentType, agentID[:8], time.Now().Unix())

	// Define the Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "ai-persona-system",
			Labels: map[string]string{
				"app":        "dynamic-agent",
				"agent-type": agentType,
				"agent-id":   agentID,
				"client-id":  clientID,
				"spawned-by": "orchestrator",
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: int32Ptr(3600), // Clean up after 1 hour
			BackoffLimit:            int32Ptr(3),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "dynamic-agent",
						"agent-type": agentType,
						"agent-id":   agentID,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: "ai-persona-app", // Same SA as your other pods
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "docker-hub-creds"},
					},
					Containers: []corev1.Container{
						{
							Name:  "agent",
							Image: "docker.io/aqls/agent-chassis:v1.0.25",
							Env: []corev1.EnvVar{
								// Core configuration
								{Name: "AGENT_TYPE", Value: agentType},
								{Name: "AGENT_ID", Value: agentID},
								{Name: "CLIENT_ID", Value: clientID},

								// Kafka configuration
								{Name: "KAFKA_TOPIC", Value: fmt.Sprintf("system.agent.%s.process", agentType)},
								{Name: "KAFKA_CONSUMER_GROUP", Value: fmt.Sprintf("%s-group", agentType)},

								// Database passwords from secrets
								{
									Name: "CLIENTS_DB_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "personae-platform-secrets",
											},
											Key: "CLIENTS_DB_PASSWORD",
										},
									},
								},
								{
									Name: "TEMPLATES_DB_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "personae-platform-secrets",
											},
											Key: "TEMPLATES_DB_PASSWORD",
										},
									},
								},
								// Add other secrets as needed
								{
									Name: "ANTHROPIC_API_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "personae-platform-secrets",
											},
											Key: "ANTHROPIC_API_KEY",
										},
									},
								},
							},
							// Add config from ConfigMap
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "personae-prod-config",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							// Add liveness/readiness probes
							// Liveness probe using your health server
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromString("health"), // Using named port
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							// Readiness probe
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ready",
										Port:   intstr.FromString("health"), // Using named port
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
								TimeoutSeconds:      3,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
						},
					},
				},
			},
		},
	}

	// Create the Job
	createdJob, err := clientset.BatchV1().Jobs("ai-persona-system").Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	logger.Info("Kubernetes Job created",
		zap.String("job_name", createdJob.Name),
		zap.String("namespace", createdJob.Namespace))

	return createdJob.Name, nil
}

// Helper function to check if an agent pod is running
func isAgentPodRunning(ctx context.Context, agentID string) bool {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return false
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return false
	}

	// List pods with the agent-id label
	pods, err := clientset.CoreV1().Pods("ai-persona-system").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("agent-id=%s", agentID),
	})

	if err != nil || len(pods.Items) == 0 {
		return false
	}

	// Check if any pod is running
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			return true
		}
	}

	return false
}

// Helper functions
func int32Ptr(i int32) *int32 { return &i }

func checkExistingAgent(ctx context.Context, db *sql.DB, clientID, agentType string) *AgentInfo {
	var id string
	query := fmt.Sprintf(`
        SELECT id FROM client_%s.agent_instances 
        WHERE config->>'agent_type' = $1 
        AND is_active = true
        LIMIT 1
    `, clientID)

	err := db.QueryRowContext(ctx, query, agentType).Scan(&id)
	if err != nil {
		return nil
	}

	return &AgentInfo{ID: id}
}

type AgentInfo struct {
	ID string
}

func spawnAgentKubernetesJob(ctx context.Context, agentID, agentType, clientID string, logger *zap.Logger) (string, error) {
	// Get in-cluster config
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Generate unique job name (must be DNS-1123 compliant)
	jobName := fmt.Sprintf("agent-%s-%s", agentType, agentID[:8])

	// Check if job already exists
	existingJob, err := clientset.BatchV1().Jobs("ai-persona-system").Get(ctx, jobName, metav1.GetOptions{})
	if err == nil {
		// Job exists - check if it's failed and needs to be recreated
		if existingJob.Status.Failed > 0 || existingJob.Status.Succeeded > 0 {
			logger.Info("Deleting completed/failed job before recreating",
				zap.String("job_name", jobName),
				zap.Int32("failed", existingJob.Status.Failed),
				zap.Int32("succeeded", existingJob.Status.Succeeded))

			// Delete the old job
			deletePolicy := metav1.DeletePropagationForeground
			err = clientset.BatchV1().Jobs("ai-persona-system").Delete(ctx, jobName, metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})
			if err != nil {
				logger.Warn("Failed to delete old job", zap.Error(err))
			}

			// Wait a bit for deletion to propagate
			time.Sleep(2 * time.Second)
		} else if existingJob.Status.Active > 0 {
			// Job is still running
			logger.Info("Job is already running",
				zap.String("job_name", jobName),
				zap.Int32("active", existingJob.Status.Active))
			return jobName, nil
		}
	}

	// Define the Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName, // Use the jobName variable here
			Namespace: "ai-persona-system",
			Labels: map[string]string{
				"app":        "dynamic-agent",
				"agent-type": agentType,
				"agent-id":   agentID,
				"client-id":  clientID,
				"spawned-by": "orchestrator",
				"component":  "agent",
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: int32Ptr(3600), // Clean up after 1 hour
			BackoffLimit:            int32Ptr(3),
			ActiveDeadlineSeconds:   int64Ptr(86400), // 24 hours max runtime
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "dynamic-agent",
						"agent-type": agentType,
						"agent-id":   agentID,
						"client-id":  clientID,
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9090", // Metrics port
						"prometheus.io/path":   "/metrics",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: "ai-persona-app",
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "docker-hub-creds"},
					},
					Containers: []corev1.Container{
						{
							Name:    "agent",
							Image:   getAgentImage(agentType),
							Command: getAgentCommand(agentType), // Add this helper function
							Ports: []corev1.ContainerPort{
								{
									Name:          "health",
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "metrics",
									ContainerPort: 9090,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								// Core configuration
								{Name: "AGENT_TYPE", Value: agentType},
								{Name: "AGENT_ID", Value: agentID},
								{Name: "CLIENT_ID", Value: clientID},

								// Dynamic topic configuration
								{Name: "KAFKA_TOPIC", Value: fmt.Sprintf("system.agent.%s.process", agentType)},
								{Name: "KAFKA_CONSUMER_GROUP", Value: fmt.Sprintf("%s-group-%s", agentType, agentID[:8])},

								// Health server ports
								{Name: "HEALTH_PORT", Value: "8080"},
								{Name: "METRICS_PORT", Value: "9090"},

								// Secrets
								{
									Name: "CLIENTS_DB_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "personae-platform-secrets",
											},
											Key: "CLIENTS_DB_PASSWORD",
										},
									},
								},
								{
									Name: "TEMPLATES_DB_PASSWORD", // ADD THIS
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "personae-platform-secrets",
											},
											Key: "TEMPLATES_DB_PASSWORD",
										},
									},
								},
								{
									Name: "AUTH_DB_PASSWORD", // MIGHT NEED THIS TOO
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "personae-platform-secrets",
											},
											Key: "AUTH_DB_PASSWORD",
										},
									},
								},
								{
									Name: "ANTHROPIC_API_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "personae-default-secrets",
											},
											Key: "ANTHROPIC_API_KEY",
										},
									},
								},
								// bootstrap key for agent registration
								{
									Name: "AGENT_BOOTSTRAP_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "personae-platform-secrets",
											},
											Key: "agent-bootstrap-key",
										},
									},
								},
								{
									Name:  "CORE_MANAGER_URL",
									Value: "http://core-manager.ai-persona-system.svc.cluster.local:8088",
								},
							},
							// Add all config from ConfigMap
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "personae-prod-config",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(getResourceRequest(agentType, "cpu")),
									corev1.ResourceMemory: resource.MustParse(getResourceRequest(agentType, "memory")),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(getResourceLimit(agentType, "cpu")),
									corev1.ResourceMemory: resource.MustParse(getResourceLimit(agentType, "memory")),
								},
							},
							// Liveness probe using your health server
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							// Readiness probe
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ready",
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
								TimeoutSeconds:      3,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							/*							// Use exec probe instead of HTTP
														LivenessProbe: &corev1.Probe{
															ProbeHandler: corev1.ProbeHandler{
																Exec: &corev1.ExecAction{
																	Command: []string{"sh", "-c", "ps aux | grep agent-chassis | grep -v grep"},
																},
															},
															InitialDelaySeconds: 30,
															PeriodSeconds:       10,
														},*/
						},
					},
				},
			},
		},
	}

	// Create the Job
	createdJob, err := clientset.BatchV1().Jobs("ai-persona-system").Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		// If it still exists (race condition), just return success
		if strings.Contains(err.Error(), "already exists") {
			logger.Warn("Job already exists (race condition), continuing",
				zap.String("job_name", jobName))
			return jobName, nil
		}
		return "", fmt.Errorf("failed to create kubernetes job: %w", err)
	}

	logger.Info("Created Kubernetes Job",
		zap.String("job_name", createdJob.Name),
		zap.String("namespace", createdJob.Namespace),
		zap.String("uid", string(createdJob.UID)))

	return createdJob.Name, nil
}

// Helper function to determine which image to use for each agent type
func getAgentImage(agentType string) string {
	imageTag := getImageTag()

	// Map agent types to their specific images
	imageMap := map[string]string{
		"content-creator": fmt.Sprintf("docker.io/aqls/content-creator-agent:%s", imageTag),
		"reasoning":       fmt.Sprintf("docker.io/aqls/reasoning-agent:%s", imageTag),
		"web-search":      fmt.Sprintf("docker.io/aqls/web-search-adapter:%s", imageTag),
		"image-generator": fmt.Sprintf("docker.io/aqls/image-generator-adapter:%s", imageTag),
	}

	if image, ok := imageMap[agentType]; ok {
		return image
	}

	// Default to generic agent-chassis
	return fmt.Sprintf("docker.io/aqls/agent-chassis:%s", imageTag)
}

// Helper function to get the correct command for each agent type
func getAgentCommand(agentType string) []string {
	// Map agent types to their specific commands
	commandMap := map[string][]string{
		"content-creator": {
			"./content-creator-agent",
			"-config",
			"configs/content-creator-agent.yaml",
		},
		"reasoning": {
			"./reasoning-agent",
			"-config",
			"configs/reasoning-agent.yaml",
		},
		"web-search": {
			"./web-search-adapter",
			"-config",
			"configs/web-search-adapter.yaml",
		},
		"image-generator": {
			"./image-generator-adapter",
			"-config",
			"configs/image-adapter.yaml",
		},
	}

	// For agent-chassis based agents (domain-analyst, site-architect, etc.)
	if cmd, ok := commandMap[agentType]; ok {
		return cmd
	}

	// Default command for generic agents using agent-chassis
	return []string{
		"./agent-chassis",
		"-config",
		"configs/agent-chassis.yaml",
	}
}

// Check if an agent job is currently running
func isAgentJobRunning(ctx context.Context, agentID string, logger *zap.Logger) bool {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Warn("Failed to get k8s config", zap.Error(err))
		return false
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Warn("Failed to create k8s client", zap.Error(err))
		return false
	}

	// Check for jobs with this agent ID
	jobs, err := clientset.BatchV1().Jobs("ai-persona-system").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("agent-id=%s", agentID),
	})

	if err != nil || len(jobs.Items) == 0 {
		return false
	}

	// Check if any job is active
	for _, job := range jobs.Items {
		if job.Status.Active > 0 {
			logger.Info("Found active job for agent",
				zap.String("job_name", job.Name),
				zap.Int32("active_pods", job.Status.Active))
			return true
		}
	}

	return false
}

// Helper functions
func int64Ptr(i int64) *int64 { return &i }

func getResourceLimit(agentType, resource string) string {
	// Agent-specific resource limits
	resourceMap := map[string]map[string]string{
		"content-creator": {
			"cpu":    "1000m",
			"memory": "2Gi",
		},
		"reasoning": {
			"cpu":    "1000m",
			"memory": "2Gi",
		},
		"web-search": {
			"cpu":    "500m",
			"memory": "512Mi",
		},
		"image-generator": {
			"cpu":    "500m",
			"memory": "1Gi",
		},
		// Default for unknown types
		"default": {
			"cpu":    "500m",
			"memory": "1Gi",
		},
	}

	if agentResources, ok := resourceMap[agentType]; ok {
		if value, ok := agentResources[resource]; ok {
			return value
		}
	}

	return resourceMap["default"][resource]
}

// Helper functions for resource allocation based on agent type
func getResourceRequest(agentType, resource string) string {
	// Agent-specific resource requests
	resourceMap := map[string]map[string]string{
		"content-creator": {
			"cpu":    "200m",
			"memory": "512Mi",
		},
		"reasoning": {
			"cpu":    "250m",
			"memory": "512Mi",
		},
		"web-search": {
			"cpu":    "100m",
			"memory": "256Mi",
		},
		"image-generator": {
			"cpu":    "100m",
			"memory": "256Mi",
		},
		// Default for unknown types
		"default": {
			"cpu":    "100m",
			"memory": "256Mi",
		},
	}

	if agentResources, ok := resourceMap[agentType]; ok {
		if value, ok := agentResources[resource]; ok {
			return value
		}
	}

	return resourceMap["default"][resource]
}

// Helper function to get the image tag from environment or use default
func getImageTag() string {
	if tag := os.Getenv("AGENT_IMAGE_TAG"); tag != "" {
		return tag
	}
	return "latest" // or another sensible default
}
