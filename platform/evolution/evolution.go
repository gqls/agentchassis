// platform/evolution/evolution.go
package evolution

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type EvolutionService struct {
	db    *pgxpool.Pool
	rules MutationRules
}

type MutationRules struct {
	MaxAgentsPerType   int
	MaxGroupSize       int
	MinPerformance     float64
	RequiredUsageCount int
}

type Mutation struct {
	Type   string
	Target string
	Config map[string]interface{}
	Reason string
}

type WorkflowMetrics struct {
	SuccessRate         float64
	AvgExecutionTime    time.Duration
	BottleneckAgent     string
	MissingCapabilities []string
	FailureRate         float64
	ResourceUsage       map[string]int
}

type AgentGroup struct {
	ID           string
	Name         string
	Version      string
	ParentID     string
	AgentConfigs json.RawMessage
	Workflow     json.RawMessage
	UsageCount   int
	CreatedAt    time.Time
}

func NewEvolutionService(db *pgxpool.Pool) *EvolutionService {
	return &EvolutionService{
		db: db,
		rules: MutationRules{
			MaxAgentsPerType:   5,
			MaxGroupSize:       10,
			MinPerformance:     0.7,
			RequiredUsageCount: 5,
		},
	}
}

func (e *EvolutionService) EvaluateGroup(ctx context.Context, groupID string, metrics WorkflowMetrics) (*Mutation, error) {
	// Get group details
	group, err := e.getGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Check if evolution is warranted
	if group.UsageCount < e.rules.RequiredUsageCount {
		return nil, nil // Too early
	}

	if metrics.SuccessRate > 0.95 {
		return nil, nil // Good enough
	}

	// Analyze bottlenecks
	if metrics.BottleneckAgent != "" {
		return &Mutation{
			Type:   "add_parallel",
			Target: metrics.BottleneckAgent,
			Reason: fmt.Sprintf("Agent %s is bottleneck", metrics.BottleneckAgent),
		}, nil
	}

	// Check for missing capabilities
	if len(metrics.MissingCapabilities) > 0 {
		return &Mutation{
			Type: "add_specialist",
			Config: map[string]interface{}{
				"capabilities": metrics.MissingCapabilities,
			},
			Reason: "Missing capabilities detected",
		}, nil
	}

	// Check for high failure rate
	if metrics.FailureRate > 0.2 {
		return &Mutation{
			Type:   "add_validator",
			Reason: "High failure rate detected",
		}, nil
	}

	return nil, nil
}

func (e *EvolutionService) ApplyMutation(ctx context.Context, groupID string, mutation *Mutation) (*AgentGroup, error) {
	// Implementation depends on mutation type
	switch mutation.Type {
	case "add_parallel":
		return e.addParallelAgent(ctx, groupID, mutation)
	case "add_specialist":
		return e.addSpecialistAgent(ctx, groupID, mutation)
	case "add_validator":
		return e.addValidatorAgent(ctx, groupID, mutation)
	default:
		return nil, fmt.Errorf("unknown mutation type: %s", mutation.Type)
	}
}

func (e *EvolutionService) addParallelAgent(ctx context.Context, groupID string, mutation *Mutation) (*AgentGroup, error) {
	// Get current group
	group, err := e.getGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Parse agent configs
	var agentConfigs []map[string]interface{}
	if err := json.Unmarshal(group.AgentConfigs, &agentConfigs); err != nil {
		return nil, fmt.Errorf("failed to parse agent configs: %w", err)
	}

	// Find the target agent and duplicate it
	var targetConfig map[string]interface{}
	for _, config := range agentConfigs {
		if config["agent_id"] == mutation.Target || config["agent_type"] == mutation.Target {
			targetConfig = config
			break
		}
	}

	if targetConfig == nil {
		return nil, fmt.Errorf("target agent %s not found", mutation.Target)
	}

	// Check if we've hit the limit
	agentType := targetConfig["agent_type"].(string)
	count := 0
	for _, config := range agentConfigs {
		if config["agent_type"] == agentType {
			count++
		}
	}

	if count >= e.rules.MaxAgentsPerType {
		return nil, fmt.Errorf("max agents per type reached for %s", agentType)
	}

	// Create parallel agent config
	parallelConfig := make(map[string]interface{})
	for k, v := range targetConfig {
		parallelConfig[k] = v
	}
	parallelConfig["role"] = fmt.Sprintf("%s_parallel_%d", targetConfig["role"], count+1)
	parallelConfig["parallel_index"] = count + 1

	// Add to configs
	agentConfigs = append(agentConfigs, parallelConfig)

	// Create new group version
	return e.createNewGroupVersion(ctx, group, agentConfigs, mutation)
}

func (e *EvolutionService) addSpecialistAgent(ctx context.Context, groupID string, mutation *Mutation) (*AgentGroup, error) {
	group, err := e.getGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var agentConfigs []map[string]interface{}
	if err := json.Unmarshal(group.AgentConfigs, &agentConfigs); err != nil {
		return nil, err
	}

	// Check group size limit
	if len(agentConfigs) >= e.rules.MaxGroupSize {
		return nil, fmt.Errorf("max group size reached")
	}

	// Create specialist config
	capabilities := mutation.Config["capabilities"].([]string)
	specialistConfig := map[string]interface{}{
		"role":         fmt.Sprintf("specialist_%s", capabilities[0]),
		"agent_type":   "specialist",
		"capabilities": capabilities,
		"config": map[string]interface{}{
			"specialization": capabilities,
			"workflow": map[string]interface{}{
				"start_step": "analyze",
				"steps": map[string]interface{}{
					"analyze": map[string]interface{}{
						"action":    "specialized_analysis",
						"next_step": "complete",
					},
					"complete": map[string]interface{}{
						"action": "complete_workflow",
					},
				},
			},
		},
	}

	agentConfigs = append(agentConfigs, specialistConfig)

	return e.createNewGroupVersion(ctx, group, agentConfigs, mutation)
}

func (e *EvolutionService) addValidatorAgent(ctx context.Context, groupID string, mutation *Mutation) (*AgentGroup, error) {
	group, err := e.getGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var agentConfigs []map[string]interface{}
	if err := json.Unmarshal(group.AgentConfigs, &agentConfigs); err != nil {
		return nil, err
	}

	// Add validator config
	validatorConfig := map[string]interface{}{
		"role":       "quality_validator",
		"agent_type": "validator",
		"config": map[string]interface{}{
			"validation_rules": []string{
				"completeness",
				"accuracy",
				"consistency",
			},
		},
	}

	agentConfigs = append(agentConfigs, validatorConfig)

	// Also update workflow to include validation step
	var workflow map[string]interface{}
	if err := json.Unmarshal(group.Workflow, &workflow); err != nil {
		return nil, err
	}

	// Insert validation step before completion
	if steps, ok := workflow["steps"].(map[string]interface{}); ok {
		// Find steps that lead to "complete" and redirect them to "validate"
		for _, step := range steps {
			if stepMap, ok := step.(map[string]interface{}); ok {
				if stepMap["next_step"] == "complete" {
					stepMap["next_step"] = "validate"
				}
			}
		}

		// Add validation step
		steps["validate"] = map[string]interface{}{
			"action":     "call_agent",
			"agent_type": "validator",
			"next_step":  "complete",
		}
	}

	return e.createNewGroupVersionWithWorkflow(ctx, group, agentConfigs, workflow, mutation)
}

func (e *EvolutionService) createNewGroupVersion(ctx context.Context, parentGroup *AgentGroup, newConfigs []map[string]interface{}, mutation *Mutation) (*AgentGroup, error) {
	var workflow map[string]interface{}
	if err := json.Unmarshal(parentGroup.Workflow, &workflow); err != nil {
		return nil, err
	}

	return e.createNewGroupVersionWithWorkflow(ctx, parentGroup, newConfigs, workflow, mutation)
}

func (e *EvolutionService) createNewGroupVersionWithWorkflow(ctx context.Context, parentGroup *AgentGroup, newConfigs []map[string]interface{}, workflow map[string]interface{}, mutation *Mutation) (*AgentGroup, error) {
	// Generate new version number
	newVersion := e.incrementVersion(parentGroup.Version)
	newID := uuid.New().String()

	// Marshal configs
	configsJSON, err := json.Marshal(newConfigs)
	if err != nil {
		return nil, err
	}

	workflowJSON, err := json.Marshal(workflow)
	if err != nil {
		return nil, err
	}

	// Create mutation record
	mutationRecord := map[string]interface{}{
		"type":      mutation.Type,
		"target":    mutation.Target,
		"reason":    mutation.Reason,
		"timestamp": time.Now(),
	}
	mutationJSON, _ := json.Marshal(mutationRecord)

	// Insert new group version
	_, err = e.db.Exec(ctx, `
        INSERT INTO agent_groups 
        (id, name, group_type, version, parent_id, agent_configs, 
         orchestration_workflow, capabilities, created_by, mutation_history)
        SELECT 
            $1, 
            name || ' ' || $2,
            group_type,
            $2,
            $3,
            $4,
            $5,
            capabilities,
            'evolution_service',
            COALESCE(mutation_history, '[]'::jsonb) || $6::jsonb
        FROM agent_groups
        WHERE id = $3
    `, newID, newVersion, parentGroup.ID, configsJSON, workflowJSON, mutationJSON)

	if err != nil {
		return nil, fmt.Errorf("failed to create new group version: %w", err)
	}

	// Return the new group
	return e.getGroup(ctx, newID)
}

func (e *EvolutionService) getGroup(ctx context.Context, groupID string) (*AgentGroup, error) {
	var group AgentGroup
	err := e.db.QueryRow(ctx, `
        SELECT id, name, version, COALESCE(parent_id, ''), 
               agent_configs, orchestration_workflow, usage_count, created_at
        FROM agent_groups 
        WHERE id = $1
    `, groupID).Scan(
		&group.ID, &group.Name, &group.Version, &group.ParentID,
		&group.AgentConfigs, &group.Workflow, &group.UsageCount, &group.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (e *EvolutionService) incrementVersion(version string) string {
	if version == "" {
		return "1.0.0"
	}

	// Simple implementation - in production, parse and increment properly
	var major, minor, patch int
	fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)

	// For mutations, increment minor version
	return fmt.Sprintf("%d.%d.%d", major, minor+1, patch)
}

// GetEvolutionHistory returns the evolution history of a group
func (e *EvolutionService) GetEvolutionHistory(ctx context.Context, groupID string) ([]EvolutionRecord, error) {
	query := `
        WITH RECURSIVE group_lineage AS (
            SELECT id, parent_id, version, mutation_history, created_at
            FROM agent_groups
            WHERE id = $1
            
            UNION ALL
            
            SELECT g.id, g.parent_id, g.version, g.mutation_history, g.created_at
            FROM agent_groups g
            INNER JOIN group_lineage l ON g.id = l.parent_id
        )
        SELECT id, version, mutation_history, created_at
        FROM group_lineage
        ORDER BY created_at ASC
    `

	rows, err := e.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []EvolutionRecord
	for rows.Next() {
		var record EvolutionRecord
		var mutationJSON json.RawMessage

		err := rows.Scan(&record.GroupID, &record.Version, &mutationJSON, &record.Timestamp)
		if err != nil {
			return nil, err
		}

		// Parse mutations
		json.Unmarshal(mutationJSON, &record.Mutations)
		history = append(history, record)
	}

	return history, nil
}

type EvolutionRecord struct {
	GroupID   string
	Version   string
	Mutations []map[string]interface{}
	Timestamp time.Time
}
