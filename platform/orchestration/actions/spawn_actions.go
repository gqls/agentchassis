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
	"github.com/gqls/agentchassis/platform/orchestration/types"
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

// AgentDefinition represents the database agent definition
type AgentDefinition struct {
	ID              string          `db:"id"`
	Type            string          `db:"type"`
	DisplayName     string          `db:"display_name"`
	Description     string          `db:"description"`
	Category        string          `db:"category"`
	ImageRepository string          `db:"image_repository"`
	ImageTag        string          `db:"image_tag"`
	Command         []string        `db:"command"`
	Resources       json.RawMessage `db:"resources"`
	DefaultConfig   json.RawMessage `db:"default_config"`
	Capabilities    json.RawMessage `db:"capabilities"`
	Topics          json.RawMessage `db:"topics"`
	HealthConfig    json.RawMessage `db:"health_config"`
	EnvVars         json.RawMessage `db:"env_vars"`
	IsActive        bool            `db:"is_active"`
}

// ResourceSpec represents Kubernetes resource requirements
type ResourceSpec struct {
	Requests map[string]string `json:"requests"`
	Limits   map[string]string `json:"limits"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	LivenessPath        string `json:"liveness_path"`
	ReadinessPath       string `json:"readiness_path"`
	Port                int    `json:"port"`
	InitialDelaySeconds int    `json:"initial_delay_seconds"`
}

// TopicConfig represents Kafka topic configuration
type TopicConfig struct {
	Process  string `json:"process"`
	Response string `json:"response"`
	Error    string `json:"error"`
	DLQ      string `json:"dlq"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EnhancedDiscovery wraps database connections and provides discovery methods
type EnhancedDiscovery struct {
	db interface{} // Can be *sql.DB or *pgxpool.Pool
}

// SpawnGroupAction spawns a complete agent group
func SpawnGroupAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	groupType, ok := config["group_type"].(string)
	if !ok {
		// Try to get from input data
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			groupType, ok = inputData["group_type"].(string)
			params.Logger.Info("Got group_type from input_data",
				zap.String("DEBUG_SPAWN_1: group_type", groupType),
				zap.Any("DEBUG_SPAWN_1: input_data", inputData))
		}
		if !ok {
			params.Logger.Error("group_type not found anywhere",
				zap.Any("DEBUG_SPAWN_2: config", config),
				zap.Any("DEBUG_SPAWN_2: collected_data", params.CollectedData))
			return nil, fmt.Errorf("group_type not specified")
		}
	}

	params.Logger.Info("SpawnGroupAction starting",
		zap.String("group_type", groupType),
		zap.Any("headers", params.Headers),
		zap.Bool("has_producer", params.Producer != nil),
		zap.Bool("has_db", params.DB != nil))

	// Find the best matching group
	var groupID, groupName string
	var agentConfigs json.RawMessage
	var workflow json.RawMessage

	err := params.DB.QueryRowContext(ctx, `
        SELECT id, name, agent_configs, orchestration_workflow
        FROM agent_group_definitions
        WHERE group_type = $1
        ORDER BY usage_count DESC, version DESC
        LIMIT 1
    `, groupType).Scan(&groupID, &groupName, &agentConfigs, &workflow)

	if err != nil {
		params.Logger.Error("DEBUG_SPAWN_3: Failed to find group",
			zap.String("DEBUG_SPAWN_3: group_type", groupType),
			zap.Error(err))
		return nil, fmt.Errorf("no group found for type %s: %w", groupType, err)
	}

	params.Logger.Info("Found group in database",
		zap.String("DEBUG_SPAWN_4: group_id", groupID),
		zap.String("DEBUG_SPAWN_4: group_name", groupName),
		zap.Int("DEBUG_SPAWN_4: agent_configs_size", len(agentConfigs)))

	// Parse agent configs
	var agents []map[string]interface{}
	if err := json.Unmarshal(agentConfigs, &agents); err != nil {
		return nil, fmt.Errorf("failed to parse agent configs: %w", err)
	}

	params.Logger.Info("Parsed agent configs",
		zap.Int("DEBUG_SPAWN_5: agent_count", len(agents)),
		zap.Any("DEBUG_SPAWN_5: agents", agents))

	// Spawn each agent in the group
	spawnedAgents := make(map[string]string)

	for i, agentConfig := range agents {
		params.Logger.Info("Processing agent config",
			zap.Int("DEBUG_SPAWN_6: index", i),
			zap.Any("DEBUG_SPAWN_6: config", agentConfig))

		role, ok := agentConfig["role"].(string)
		if !ok {
			params.Logger.Warn("Skipping agent - no role",
				zap.Int("DEBUG_SPAWN_7: index", i))
			continue
		}

		agentType, ok := agentConfig["agent_type"].(string)
		if !ok {
			params.Logger.Warn("Skipping agent - no agent_type",
				zap.String("role", role))
			continue
		}

		// Build config overrides
		configOverrides := make(map[string]interface{})
		if role == "orchestrator" {
			configOverrides["processing_mode"] = "orchestrator"
			configOverrides["listen_to_responses"] = true
		}

		params.Logger.Info("About to spawn agent",
			zap.String("DEBUG_SPAWN_8: role", role),
			zap.String("DEBUG_SPAWN_8: agent_type", agentType),
			zap.Any("DEBUG_SPAWN_8: config_overrides", configOverrides),
			zap.Bool("DEBUG_SPAWN_8: has_producer", params.Producer != nil))

		result, err := SpawnAgentAction(ctx, ActionParams{
			StepConfig: models.Step{
				Config: map[string]interface{}{
					"agent_type":       agentType,
					"config_overrides": configOverrides,
				},
			},
			Headers:  params.Headers,
			DB:       params.DB,
			Logger:   params.Logger,
			Producer: params.Producer,
		})

		if err != nil {
			params.Logger.Error("Failed to spawn agent",
				zap.String("DEBUG_SPAWN_9: role", role),
				zap.String("DEBUG_SPAWN_9: agent_type", agentType),
				zap.Error(err))
			return nil, fmt.Errorf("failed to spawn %s: %w", role, err)
		}

		agentResult, ok := result.(map[string]interface{})
		if !ok {
			params.Logger.Error("Unexpected result type",
				zap.String("DEBUG_SPAWN_10: role", role),
				zap.Any("DEBUG_SPAWN_10: result", result))
			return nil, fmt.Errorf("unexpected result type from SpawnAgentAction")
		}

		agentID, ok := agentResult["agent_id"].(string)
		if !ok {
			params.Logger.Error("No agent_id in result",
				zap.String("DEBUG_SPAWN_11: role", role),
				zap.Any("DEBUG_SPAWN_11: result", agentResult))
			return nil, fmt.Errorf("agent_id not found in spawn result")
		}

		spawnedAgents[role] = agentID
		params.Logger.Info("Successfully spawned agent",
			zap.String("DEBUG_SPAWN_12: role", role),
			zap.String("DEBUG_SPAWN_12: agent_id", agentID),
			zap.String("DEBUG_SPAWN_12: agent_type", agentType),
			zap.Int("DEBUG_SPAWN_12: spawned_so_far", len(spawnedAgents)))
	}

	params.Logger.Info("All agents spawned",
		zap.Int("DEBUG_SPAWN_14: total_spawned", len(spawnedAgents)),
		zap.Any("DEBUG_SPAWN_14: spawned_agents", spawnedAgents))

	// The request_id that the parent is waiting for
	requestID := params.Headers["request_id"]
	if requestID == "" {
		// This shouldn't happen if executeLocalAction pre-generated it
		requestID = uuid.New().String()
		params.Logger.Warn("No request_id in headers, generating new one",
			zap.String("request_id", requestID))
	}

	params.Logger.Info("SpawnGroupAction using request_id",
		zap.String("request_id", requestID),
		zap.String("orchestration_id", params.Headers["orchestration_id"]))

	// Parse and validate workflow before returning
	var workflowPlan map[string]interface{}
	if err := json.Unmarshal(workflow, &workflowPlan); err != nil {
		params.Logger.Error("Failed to parse workflow from group",
			zap.Error(err),
			zap.String("group_id", groupID))
		// Create a minimal valid workflow
		workflowPlan = map[string]interface{}{
			"start_step": "execute",
			"steps": map[string]interface{}{
				"execute": map[string]interface{}{
					"action":      "execute_llm_prompt",
					"description": "Execute task",
					"next_step":   "complete",
				},
				"complete": map[string]interface{}{
					"action":      "complete_workflow",
					"description": "Complete workflow",
				},
			},
		}
	}

	// Validate workflow has required fields
	if _, hasStart := workflowPlan["start_step"]; !hasStart {
		workflowPlan["start_step"] = "execute"
	}
	if _, hasSteps := workflowPlan["steps"]; !hasSteps {
		workflowPlan["steps"] = map[string]interface{}{
			"execute": map[string]interface{}{
				"action": "execute_llm_prompt",
			},
		}
	}

	// Re-marshal the validated workflow
	validatedWorkflow, _ := json.Marshal(workflowPlan)

	// Before returning, trace what we're setting up
	if params.Tracer != nil {
		execCtx, _ := types.FromHeaders(params.Headers)
		if execCtx != nil {
			params.Tracer.TraceMessage(execCtx, "spawn_group_complete", "", 0)
		}
	}

	// Make sure to log the actual request_id that will be awaited
	params.Logger.Info("SPAWN_GROUP_REQUEST_ID",
		zap.String("request_id_in_headers", params.Headers["request_id"]),
		zap.String("requestID variable send back", requestID),
		zap.String("group_id", groupID))

	params.Logger.Info("CRITICAL_FLOW: SpawnGroup returning",
		zap.Bool("await_response", true),
		zap.String("request_id_returning", requestID),
		zap.String("request_id_in_headers", params.Headers["request_id"]),
		zap.String("orchestration_id", params.Headers["orchestration_id"]))

	return map[string]interface{}{
		"await_response": true,
		"request_id":     requestID,
		"group_id":       groupID,
		"group_name":     groupName,
		"group_type":     groupType,
		"agents":         spawnedAgents,
		"workflow":       validatedWorkflow,
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
			return nil, fmt.Errorf("Failed to find agents by capability: %w", err)
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
		return nil, fmt.Errorf("Unsupported database type: %T", d.db)
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
			return nil, fmt.Errorf("Failed to get recommendations: %w", err)
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
			return nil, fmt.Errorf("Failed to get performance summary: %w", err)
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
		return nil, fmt.Errorf("Unsupported Database type: %T", d.db)
	}
}

// Request tracking parameters
type trackRequestParams struct {
	RequestID       string
	OrchestrationID string
	ToAgentID       string
	Timeout         time.Duration
}

// trackRequest records the request in the database
func trackRequest(ctx context.Context, db *sql.DB, requestID, orchestrationID, toAgentID string) error {
	query := `
        INSERT INTO pending_requests 
        (request_id, orchestration_id, to_agent_id, status, timeout_at, created_at)
        VALUES ($1, $2, $3, 'pending', $4, NOW())
    `

	timeout := time.Now().Add(60 * time.Second)
	_, err := db.ExecContext(ctx, query,
		requestID,
		orchestrationID,
		toAgentID,
		timeout,
	)

	return err
}

// failRequest marks a request as failed
func failRequest(ctx context.Context, db *sql.DB, requestID string) error {
	query := `
		UPDATE pending_requests 
		SET status = 'failed', 
		    completed_at = NOW()
		WHERE request_id = $1
	`

	_, err := db.ExecContext(ctx, query, requestID)
	return err
}

// getAgentImage - Database-driven version
func getAgentImage(agentType string, db interface{}) string {
	ctx := context.Background()

	query := `
        SELECT image_repository, image_tag 
        FROM agent_definitions 
        WHERE type = $1 AND is_active = true AND deleted_at IS NULL
        LIMIT 1
    `

	var imageRepo, imageTag string

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, agentType).Scan(&imageRepo, &imageTag)
		if err != nil {
			// Log the error and return default
			return fmt.Sprintf("docker.io/aqls/agent-chassis:latest")
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, agentType).Scan(&imageRepo, &imageTag)
		if err != nil {
			return fmt.Sprintf("docker.io/aqls/agent-chassis:latest")
		}
	default:
		return fmt.Sprintf("docker.io/aqls/agent-chassis:latest")
	}

	return fmt.Sprintf("%s:%s", imageRepo, imageTag)
}

// getAgentCommand - Database-driven version
func getAgentCommand(agentType string, db interface{}) []string {
	ctx := context.Background()

	query := `
        SELECT command 
        FROM agent_definitions 
        WHERE type = $1 AND is_active = true AND deleted_at IS NULL
        LIMIT 1
    `

	switch d := db.(type) {
	case *sql.DB:
		var commandStr string
		err := d.QueryRowContext(ctx, query, agentType).Scan(&commandStr)
		if err != nil {
			return []string{"./agent-chassis", "-config", "configs/agent-chassis.yaml"}
		}

		// Parse the PostgreSQL array string
		command := parsePostgresArray(commandStr)
		if len(command) > 0 {
			return command
		}
	case *pgxpool.Pool:
		var command []string
		err := d.QueryRow(ctx, query, agentType).Scan(&command)
		if err != nil || len(command) == 0 {
			return []string{"./agent-chassis", "-config", "configs/agent-chassis.yaml"}
		}
		return command
	}

	return []string{"./agent-chassis", "-config", "configs/agent-chassis.yaml"}
}

// Helper to parse PostgreSQL array format
func parsePostgresArray(s string) []string {
	// PostgreSQL arrays look like: {value1,value2,value3}
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
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
		zap.String("DEBUG_SPAWN_28: agent_type", agentType),
		zap.String("DEBUG_SPAWN_28: client_id", clientID))

	// Get agent definition from database
	agentDef, err := getAgentDefinition(ctx, params.DB, agentType, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent definition: %w", err)
	}

	if !agentDef.IsActive {
		return nil, fmt.Errorf("agent type %s is not active", agentType)
	}

	// Check for existing agent
	existingAgent := checkExistingAgent(ctx, params.DB, clientID, agentType)
	if existingAgent != nil {
		// Check if its Kubernetes job is still running
		if isAgentJobRunning(ctx, existingAgent.ID, params.Logger) {
			params.Logger.Info("Reusing existing running agent",
				zap.String("DEBUG_SPAWN_29: agent_id", existingAgent.ID),
				zap.String("DEBUG_SPAWN_29: agent_type", agentType))

			// Get topic from definition
			topics := parseTopicConfig(agentDef.Topics)
			processTopic := strings.ReplaceAll(topics.Process, "{type}", agentType)

			return map[string]interface{}{
				"agent_id": existingAgent.ID,
				"topic":    processTopic,
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
		// Create new agent in database using definition
		if err := createAgentInDBFromDefinition(ctx, params, agentID, agentDef, clientID); err != nil {
			return nil, fmt.Errorf("failed to create agent in database: %w", err)
		}
	}

	// Spawn the Kubernetes Job using definition
	jobName, err := spawnAgentKubernetesJobFromDefinition(ctx, agentID, agentDef, clientID, params.Logger)
	if err != nil {
		params.Logger.Error("Failed to spawn agent job",
			zap.Error(err),
			zap.String("DEBUG_SPAWN_30: agent_id", agentID),
			zap.String("DEBUG_SPAWN_30: agent_type", agentType))
		// Don't fail - agent exists in DB and can be spawned manually
		topics := parseTopicConfig(agentDef.Topics)
		processTopic := strings.ReplaceAll(topics.Process, "{type}", agentType)

		return map[string]interface{}{
			"agent_id": agentID,
			"topic":    processTopic,
			"status":   "created_without_job",
			"error":    err.Error(),
		}, nil
	}

	params.Logger.Info("Successfully spawned agent job",
		zap.String("DEBUG_SPAWN_31: job_name", jobName),
		zap.String("DEBUG_SPAWN_31: agent_id", agentID),
		zap.String("DEBUG_SPAWN_31: agent_type", agentType))

	topics := parseTopicConfig(agentDef.Topics)
	processTopic := strings.ReplaceAll(topics.Process, "{type}", agentType)

	return map[string]interface{}{
		"agent_id": agentID,
		"topic":    processTopic,
		"status":   "spawned",
		"job_name": jobName,
	}, nil
}

// getAgentDefinition retrieves agent definition from database
func getAgentDefinition(ctx context.Context, db interface{}, agentType string, logger *zap.Logger) (*AgentDefinition, error) {
	query := `
		SELECT id, type, display_name, description, category,
		       image_repository, image_tag, command,
		       resources, default_config, capabilities, topics,
		       health_config, env_vars, is_active
		FROM agent_definitions
		WHERE type = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var def AgentDefinition
	var command sql.NullString

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, agentType).Scan(
			&def.ID, &def.Type, &def.DisplayName, &def.Description, &def.Category,
			&def.ImageRepository, &def.ImageTag, &command,
			&def.Resources, &def.DefaultConfig, &def.Capabilities, &def.Topics,
			&def.HealthConfig, &def.EnvVars, &def.IsActive,
		)
		if err != nil {
			return nil, err
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, agentType).Scan(
			&def.ID, &def.Type, &def.DisplayName, &def.Description, &def.Category,
			&def.ImageRepository, &def.ImageTag, &def.Command,
			&def.Resources, &def.DefaultConfig, &def.Capabilities, &def.Topics,
			&def.HealthConfig, &def.EnvVars, &def.IsActive,
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	logger.Info("Agent definition loaded for spawn",
		zap.String("DEBUG_SPAWN_32: agent_type", agentType),
		zap.String("DEBUG_SPAWN_32: raw_default_config", string(def.DefaultConfig)))

	return &def, nil
}

// parseTopicConfig parses the topic configuration from JSON
func parseTopicConfig(topicsJSON json.RawMessage) TopicConfig {
	var topics TopicConfig
	if err := json.Unmarshal(topicsJSON, &topics); err != nil {
		// Return defaults
		return TopicConfig{
			Process:  "system.agent.{type}.process",
			Response: "system.agent.{type}.responses",
			Error:    "system.agent.{type}.errors",
			DLQ:      "system.agent.{type}.dlq",
		}
	}
	return topics
}

// parseResourceSpec parses resource configuration
func parseResourceSpec(resourcesJSON json.RawMessage) ResourceSpec {
	var spec ResourceSpec
	if err := json.Unmarshal(resourcesJSON, &spec); err != nil {
		// Return defaults
		return ResourceSpec{
			Requests: map[string]string{"cpu": "100m", "memory": "256Mi"},
			Limits:   map[string]string{"cpu": "500m", "memory": "1Gi"},
		}
	}
	return spec
}

// parseHealthConfig parses health check configuration
func parseHealthConfig(healthJSON json.RawMessage) HealthCheckConfig {
	var config HealthCheckConfig
	if err := json.Unmarshal(healthJSON, &config); err != nil {
		// Return defaults
		return HealthCheckConfig{
			LivenessPath:        "/health",
			ReadinessPath:       "/ready",
			Port:                8080,
			InitialDelaySeconds: 30,
		}
	}
	return config
}

// parseEnvVars parses environment variables configuration
func parseEnvVars(envJSON json.RawMessage) []EnvVar {
	var envVars []EnvVar
	json.Unmarshal(envJSON, &envVars)
	return envVars
}

// buildWorkflowForType fetches workflow from agent definition in database
// This replaces the old function that had hardcoded workflows
func buildWorkflowForType(ctx context.Context, db interface{}, agentType string) (map[string]interface{}, error) {
	// Query the workflow from agent_definitions
	query := `
		SELECT default_config->'workflow'
		FROM agent_definitions
		WHERE type = $1 
		AND is_active = true 
		AND deleted_at IS NULL
		LIMIT 1
	`

	var workflowJSON json.RawMessage

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, agentType).Scan(&workflowJSON)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("agent type '%s' not found in definitions", agentType)
			}
			return nil, fmt.Errorf("failed to fetch workflow for agent type '%s': %w", agentType, err)
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, agentType).Scan(&workflowJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch workflow for agent type '%s': %w", agentType, err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	// Parse the workflow
	if workflowJSON == nil || len(workflowJSON) == 0 {
		// No workflow defined, return minimal
		return buildMinimalWorkflow(agentType), nil
	}

	var workflow map[string]interface{}
	if err := json.Unmarshal(workflowJSON, &workflow); err != nil {
		// Failed to parse, return minimal
		return buildMinimalWorkflow(agentType), nil
	}

	// Validate workflow has required fields
	if _, hasStartStep := workflow["start_step"]; !hasStartStep {
		return buildMinimalWorkflow(agentType), nil
	}
	if _, hasSteps := workflow["steps"]; !hasSteps {
		return buildMinimalWorkflow(agentType), nil
	}

	return workflow, nil
}

// validateWorkflow ensures a workflow has all required fields
func validateWorkflow(workflow map[string]interface{}) error {
	// Check for start_step
	startStep, ok := workflow["start_step"].(string)
	if !ok || startStep == "" {
		return fmt.Errorf("workflow missing required field 'start_step'")
	}

	// Check for steps
	steps, ok := workflow["steps"].(map[string]interface{})
	if !ok || len(steps) == 0 {
		return fmt.Errorf("workflow missing required field 'steps' or steps is empty")
	}

	// Verify start_step exists in steps
	if _, exists := steps[startStep]; !exists {
		return fmt.Errorf("start_step '%s' not found in workflow steps", startStep)
	}

	// Check each step has required fields
	for stepName, stepData := range steps {
		step, ok := stepData.(map[string]interface{})
		if !ok {
			return fmt.Errorf("step '%s' is not a valid object", stepName)
		}

		// Every step must have an action
		if _, hasAction := step["action"]; !hasAction {
			return fmt.Errorf("step '%s' missing required field 'action'", stepName)
		}
	}

	return nil
}

// buildMinimalWorkflow creates a minimal fallback workflow
// This is only used when database doesn't have a workflow defined
func buildMinimalWorkflow(agentType string) map[string]interface{} {
	// Determine the primary action based on agent category
	// We can fetch this from the DB or use a simple default
	primaryAction := "execute_llm_prompt"

	// You could make this smarter by checking the agent type
	switch {
	case strings.Contains(agentType, "generic"):
		primaryAction = "validate_input"
	case strings.Contains(agentType, "adapt"):
		primaryAction = "http_request"
	case strings.Contains(agentType, "publish"):
		primaryAction = "deploy_to_hosting"
	default:
		primaryAction = "execute_llm_prompt"
	}

	return map[string]interface{}{
		"start_step": "process",
		"steps": map[string]interface{}{
			"process": map[string]interface{}{
				"action":      primaryAction,
				"description": fmt.Sprintf("Process %s task", agentType),
				"next_step":   "complete",
			},
			"complete": map[string]interface{}{
				"action":      "complete_workflow",
				"description": "Complete the workflow",
			},
		},
	}
}

// createAgentInDBFromDefinition creates agent instance using definition from database
func createAgentInDBFromDefinition(ctx context.Context, params ActionParams, agentID string, agentDef *AgentDefinition, clientID string) error {
	// Parse the default config from the agent definition
	var defaultConfig map[string]interface{}
	if err := json.Unmarshal(agentDef.DefaultConfig, &defaultConfig); err != nil {
		// If parsing fails, start with empty config
		defaultConfig = make(map[string]interface{})
		params.Logger.Warn("Failed to parse default config, using empty config",
			zap.String("agent_type", agentDef.Type),
			zap.Error(err))
	}

	// Ensure we have a workflow - either from default_config or build a minimal one
	if _, hasWorkflow := defaultConfig["workflow"]; !hasWorkflow {
		// Try to extract workflow from the default_config if it exists at a different level
		if workflowRaw, ok := defaultConfig["workflow"]; ok {
			defaultConfig["workflow"] = workflowRaw
		} else {
			// Build a minimal fallback workflow
			params.Logger.Info("No workflow in default config, using minimal workflow",
				zap.String("agent_type", agentDef.Type))
			defaultConfig["workflow"] = buildMinimalWorkflow(agentDef.Type)
		}
	}

	// Parse and set capabilities
	var capabilities []string
	if err := json.Unmarshal(agentDef.Capabilities, &capabilities); err == nil {
		defaultConfig["capabilities"] = capabilities
	} else {
		// Fallback to agent type as capability
		defaultConfig["capabilities"] = []string{agentDef.Type}
	}

	// Parse topics configuration
	topics := parseTopicConfig(agentDef.Topics)
	processTopic := strings.ReplaceAll(topics.Process, "{type}", agentDef.Type)

	// Add runtime configuration that's always needed
	runtimeConfig := map[string]interface{}{
		"agent_id":     agentID,
		"agent_type":   agentDef.Type,
		"display_name": agentDef.DisplayName,
		"category":     agentDef.Category,
		"topic":        processTopic,
		"topics": map[string]string{
			"process":  strings.ReplaceAll(topics.Process, "{type}", agentDef.Type),
			"response": strings.ReplaceAll(topics.Response, "{type}", agentDef.Type),
			"error":    strings.ReplaceAll(topics.Error, "{type}", agentDef.Type),
			"dlq":      strings.ReplaceAll(topics.DLQ, "{type}", agentDef.Type),
		},
	}

	// Merge runtime config with default config
	for k, v := range runtimeConfig {
		defaultConfig[k] = v
	}

	// Add any overrides from the spawn request
	if overrides, ok := params.StepConfig.Config["config_overrides"].(map[string]interface{}); ok {
		for k, v := range overrides {
			defaultConfig[k] = v
		}
		params.Logger.Info("Applied config overrides",
			zap.String("agent_id", agentID),
			zap.Int("override_count", len(overrides)))
	}

	// Marshal the final configuration
	configJSON, err := json.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal agent config: %w", err)
	}

	// Prepare the insert query for the client-specific schema
	insertQuery := fmt.Sprintf(`
		INSERT INTO client_%s.agent_instances 
		(id, template_id, owner_user_id, name, config, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			config = EXCLUDED.config,
			updated_at = NOW()
	`, clientID)

	// Get user ID from headers or default to system
	userID := params.Headers["user_id"]
	if userID == "" {
		userID = "system"
	}

	// Generate a descriptive name for the instance
	instanceName := fmt.Sprintf("%s-%s", agentDef.DisplayName, time.Now().Format("20060102-150405"))
	if customName, ok := params.StepConfig.Config["instance_name"].(string); ok && customName != "" {
		instanceName = customName
	}

	// Execute the insert
	_, err = params.DB.ExecContext(ctx, insertQuery,
		agentID,
		agentDef.ID, // Reference to the agent_definitions table
		userID,
		instanceName,
		configJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to insert agent instance: %w", err)
	}

	params.Logger.Info("Agent instance created in database",
		zap.String("DEBUG_SPAWN_33: agent_id", agentID),
		zap.String("DEBUG_SPAWN_33: agent_type", agentDef.Type),
		zap.String("DEBUG_SPAWN_33: instance_name", instanceName),
		zap.String("DEBUG_SPAWN_33: client_id", clientID))

	return nil
}

// spawnAgentKubernetesJobFromDefinition spawns job using database definition
func spawnAgentKubernetesJobFromDefinition(ctx context.Context, agentID string, agentDef *AgentDefinition, clientID string, logger *zap.Logger) (string, error) {
	// Get in-cluster config
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Generate unique job name - use only first 8 chars for K8s naming constraints
	// But the AGENT_ID env var will have the FULL UUID
	jobName := fmt.Sprintf("agent-%s-%s", agentDef.Type, agentID[:8])

	// Check if job already exists and handle accordingly
	existingJob, err := clientset.BatchV1().Jobs("ai-persona-system").Get(ctx, jobName, metav1.GetOptions{})
	if err == nil {
		if existingJob.Status.Failed > 0 || existingJob.Status.Succeeded > 0 {
			logger.Info("Deleting completed/failed job before recreating",
				zap.String("DEBUG_SPAWN_34: job_name", jobName))

			deletePolicy := metav1.DeletePropagationForeground
			err = clientset.BatchV1().Jobs("ai-persona-system").Delete(ctx, jobName, metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})
			if err != nil {
				logger.Warn("Failed to delete old job", zap.Error(err))
			}
			time.Sleep(2 * time.Second)
		} else if existingJob.Status.Active > 0 {
			logger.Info("Job is already running",
				zap.String("DEBUG_SPAWN_35: job_name", jobName))
			return jobName, nil
		}
	}

	// Parse configurations
	resources := parseResourceSpec(agentDef.Resources)
	healthConfig := parseHealthConfig(agentDef.HealthConfig)
	envVars := parseEnvVars(agentDef.EnvVars)
	topics := parseTopicConfig(agentDef.Topics)

	// Build topic name
	processTopic := strings.ReplaceAll(topics.Process, "{type}", agentDef.Type)

	// For orchestrator agents, they need to listen to responses too
	kafkaTopics := processTopic
	if agentDef.Category == "orchestrator" || agentDef.Type == "website-builder" {
		// Orchestrators must listen to both their process topic AND responses
		kafkaTopics = fmt.Sprintf("%s,system.orchestrator.responses", processTopic)
	}

	// Build environment variables
	envList := []corev1.EnvVar{
		// Core configuration
		{Name: "AGENT_TYPE", Value: agentDef.Type},
		{Name: "AGENT_ID", Value: agentID},
		{Name: "CLIENT_ID", Value: clientID},

		// Dynamic topic configuration
		{Name: "KAFKA_TOPIC", Value: processTopic},
		{Name: "KAFKA_TOPICS", Value: kafkaTopics},
		// Consumer group can use truncated ID for readability
		{Name: "KAFKA_CONSUMER_GROUP", Value: fmt.Sprintf("%s-group-%s", agentDef.Type, agentID[:8])},

		// Health server ports
		{Name: "HEALTH_PORT", Value: fmt.Sprintf("%d", healthConfig.Port)},
		{Name: "METRICS_PORT", Value: "9090"},

		// Core Manager URL
		{Name: "CORE_MANAGER_URL", Value: "http://core-manager.ai-persona-system.svc.cluster.local:8088"},
	}

	// Add custom env vars from definition
	for _, envVar := range envVars {
		envList = append(envList, corev1.EnvVar{
			Name:  envVar.Name,
			Value: envVar.Value,
		})
	}

	// Add infrastructure configuration from orchestrator environment
	envList = append(envList, []corev1.EnvVar{
		{Name: "SERVICE_INFRASTRUCTURE_KAFKA_BROKERS", Value: os.Getenv("SERVICE_INFRASTRUCTURE_KAFKA_BROKERS")},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST", Value: os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST")},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT", Value: os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT")},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER", Value: os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER")},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME", Value: os.Getenv("SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME")},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_HOST", Value: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_HOST")},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_PORT", Value: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_PORT")},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_USER", Value: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_USER")},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_DB_NAME", Value: os.Getenv("SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_DB_NAME")},
	}...)

	// Add database passwords and other secrets
	envList = append(envList, []corev1.EnvVar{
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
		{
			Name: "AUTH_DB_PASSWORD",
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
	}...)

	// Check and add storage configuration for storage-enabled agents
	storageSecret := os.Getenv("AGENT_STORAGE_SECRET")
	storageConfigMap := os.Getenv("AGENT_STORAGE_CONFIGMAP")

	logger.Info("Checking storage configuration",
		zap.String("DEBUG_SPAWN_36: agent_type", agentDef.Type),
		zap.String("DEBUG_SPAWN_36: storage_secret", storageSecret),
		zap.String("DEBUG_SPAWN_36: storage_configmap", storageConfigMap),
		zap.Bool("DEBUG_SPAWN_36: is_storage_agent", isStorageEnabledAgent(agentDef.Type)))

	if isStorageEnabledAgent(agentDef.Type) || agentDef.Category == "orchestrator" || agentDef.Category == "code-driven" { // Get storage credentials from orchestrator's environment
		awsKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
		awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		b2KeyId := os.Getenv("B2_APPLICATION_KEY_ID")
		b2Key := os.Getenv("B2_APPLICATION_KEY")

		logger.Info("Injecting storage credentials as direct values",
			zap.String("DEBUG_SPAWN_37: agent_type", agentDef.Type),
			zap.Int("DEBUG_SPAWN_37: aws_key_len", len(awsKeyId)),
			zap.Int("DEBUG_SPAWN_37: aws_secret_len", len(awsSecretKey)),
			zap.Int("DEBUG_SPAWN_37: b2_key_id_len", len(b2KeyId)),
			zap.Int("DEBUG_SPAWN_37: b2_key_len", len(b2Key)))

		// Add storage credentials as direct values (not secret references)
		if awsKeyId != "" {
			envList = append(envList, corev1.EnvVar{
				Name:  "AWS_ACCESS_KEY_ID",
				Value: awsKeyId,
			})
		}

		if awsSecretKey != "" {
			envList = append(envList, corev1.EnvVar{
				Name:  "AWS_SECRET_ACCESS_KEY",
				Value: awsSecretKey,
			})
		}

		if b2KeyId != "" {
			envList = append(envList, corev1.EnvVar{
				Name:  "B2_APPLICATION_KEY_ID",
				Value: b2KeyId,
			})
		}

		if b2Key != "" {
			envList = append(envList, corev1.EnvVar{
				Name:  "B2_APPLICATION_KEY",
				Value: b2Key,
			})
		}
	}

	// Storage configuration from ConfigMap (keep these as references since they work)
	if storageConfigMap != "" && isStorageEnabledAgent(agentDef.Type) {
		logger.Info("Injecting storage config from ConfigMap",
			zap.String("DEBUG_SPAWN_38: agent_type", agentDef.Type),
			zap.String("DEBUG_SPAWN_38: configmap_name", storageConfigMap))

		envList = append(envList, []corev1.EnvVar{
			{
				Name: "S3_ENDPOINT",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "S3-ENDPOINT",
					},
				},
			},
			{
				Name: "S3_REGION",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "S3-REGION",
					},
				},
			},
			{
				Name: "IMAGE_BUCKET",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "image_bucket",
					},
				},
			},
			{
				Name: "ASSETS_BUCKET",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "assets_bucket",
					},
				},
			},
			{
				Name: "S3_USE_PATH_STYLE",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "S3_USE_PATH_STYLE",
					},
				},
			},
		}...)
	}

	if storageConfigMap != "" && isStorageEnabledAgent(agentDef.Type) {
		logger.Info("Injecting storage config",
			zap.String("DEBUG_SPAWN_39: agent_type", agentDef.Type),
			zap.String("DEBUG_SPAWN_39: configmap_name", storageConfigMap))

		envList = append(envList, []corev1.EnvVar{
			{
				Name: "S3_ENDPOINT",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "S3-ENDPOINT",
					},
				},
			},
			{
				Name: "S3_REGION",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "S3-REGION",
					},
				},
			},
			{
				Name: "IMAGE_BUCKET",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "image_bucket",
					},
				},
			},
			{
				Name: "ASSETS_BUCKET",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "assets_bucket",
					},
				},
			},
			{
				Name: "S3_USE_PATH_STYLE",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: storageConfigMap,
						},
						Key: "S3_USE_PATH_STYLE",
					},
				},
			},
		}...)
	}

	logger.Info("Total environment variables configured",
		zap.String("agent_type", agentDef.Type),
		zap.Int("env_count", len(envList)))

	// Define the Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "ai-persona-system",
			Labels: map[string]string{
				"app":        "dynamic-agent",
				"agent-type": agentDef.Type,
				"agent-id":   agentID,
				"client-id":  clientID,
				"spawned-by": "orchestrator",
				"component":  "agent",
				"category":   agentDef.Category,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: int32Ptr(3600),
			BackoffLimit:            int32Ptr(3),
			ActiveDeadlineSeconds:   int64Ptr(86400),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "dynamic-agent",
						"agent-type": agentDef.Type,
						"agent-id":   agentID,
						"client-id":  clientID,
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9090",
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
							Image:   fmt.Sprintf("%s:%s", agentDef.ImageRepository, agentDef.ImageTag),
							Command: agentDef.Command,
							Ports: []corev1.ContainerPort{
								{
									Name:          "health",
									ContainerPort: int32(healthConfig.Port),
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "metrics",
									ContainerPort: 9090,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: envList,
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
									corev1.ResourceCPU:    resource.MustParse(resources.Requests["cpu"]),
									corev1.ResourceMemory: resource.MustParse(resources.Requests["memory"]),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(resources.Limits["cpu"]),
									corev1.ResourceMemory: resource.MustParse(resources.Limits["memory"]),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   healthConfig.LivenessPath,
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: int32(healthConfig.InitialDelaySeconds),
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   healthConfig.ReadinessPath,
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
						},
					},
				},
			},
		},
	}

	// Right before creating the Job, add this debug code:
	// Debug: Verify env vars are complete before job creation
	for _, env := range envList {
		if strings.Contains(env.Name, "AWS") || strings.Contains(env.Name, "B2") {
			logger.Info("DEBUG: Final env var check before job creation",
				zap.String("name", env.Name),
				zap.Bool("has_value", env.Value != ""),
				zap.Bool("has_value_from", env.ValueFrom != nil))

			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				logger.Info("DEBUG: SecretKeyRef is present",
					zap.String("env_name", env.Name),
					zap.String("secret_name", env.ValueFrom.SecretKeyRef.Name),
					zap.String("key", env.ValueFrom.SecretKeyRef.Key))
			}
		}
	}

	// Create the Job
	createdJob, err := clientset.BatchV1().Jobs("ai-persona-system").Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
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

	// After creating the job struct, verify it has the env vars:
	if len(job.Spec.Template.Spec.Containers) > 0 {
		containerEnvs := job.Spec.Template.Spec.Containers[0].Env
		for _, env := range containerEnvs {
			if strings.Contains(env.Name, "AWS") || strings.Contains(env.Name, "B2") {
				logger.Info("DEBUG: Env in job struct",
					zap.String("name", env.Name),
					zap.Bool("has_value_from", env.ValueFrom != nil))
			}
		}
	}

	return createdJob.Name, nil
}

// Keep all existing helper functions...
func int32Ptr(i int32) *int32 { return &i }
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

// isStorageEnabledAgent checks if an agent type requires storage access
func isStorageEnabledAgent(agentType string) bool {
	storageAgents := []string{
		"site-publisher",
		"html-developer",
		"visual-designer",
		"image-generator",
		"content-creator",
		"content-researcher",
		"domain-analyst",
		"site-architect",
		"website-builder",
	}

	for _, t := range storageAgents {
		if t == agentType {
			return true
		}
	}
	return false
}

// findOrSpawnAgent finds an existing agent or spawns a new one
func findOrSpawnAgent(ctx context.Context, params ActionParams, agentType string) (string, error) {
	// Try to find existing agent
	discover := NewAgentDiscovery(params.DB)
	matches, err := discover.DiscoverAgents(ctx, discovery.Requirements{
		AgentType: agentType,
		ClientID:  params.Headers["client_id"],
	})

	if err == nil && len(matches) > 0 {
		return matches[0].AgentID, nil
	}

	// Spawn new agent
	spawnResult, err := SpawnAgentAction(ctx, ActionParams{
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"agent_type": agentType,
			},
		},
		Headers:  params.Headers,
		DB:       params.DB,
		Logger:   params.Logger,
		Producer: params.Producer,
	})

	if err != nil {
		return "", fmt.Errorf("failed to spawn agent: %w", err)
	}

	sr := spawnResult.(map[string]interface{})
	return sr["agent_id"].(string), nil
}
