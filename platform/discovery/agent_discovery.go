// platform/discovery/agent_discovery.go
package discovery

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Requirements defines what capabilities and constraints an agent must meet
type Requirements struct {
	Capabilities   []string
	AgentType      string
	ClientID       string
	MaxCost        int
	MinPerformance float64
}

// AgentMatch represents a discovered agent that matches requirements
type AgentMatch struct {
	AgentID     string
	AgentType   string
	AgentName   string
	Topic       string
	Performance float64
	Available   bool
}

// AgentGroup represents a collection of agents that work together
type AgentGroup struct {
	ID              string
	Name            string
	GroupType       string
	AgentConfigs    json.RawMessage
	Workflow        json.RawMessage
	Capabilities    []string
	UsageCount      int
	LastPerformance float64
	Version         string
}

// AgentDiscovery handles finding individual agents
type AgentDiscovery struct {
	db *sql.DB
}

// GroupDiscovery handles finding agent groups
type GroupDiscovery struct {
	db *sql.DB
}

// NewAgentDiscovery creates a new agent discovery service
func NewAgentDiscovery(db *sql.DB) *AgentDiscovery {
	return &AgentDiscovery{db: db}
}

// NewGroupDiscovery creates a new group discovery service
func NewGroupDiscovery(db *sql.DB) *GroupDiscovery {
	return &GroupDiscovery{db: db}
}

// DiscoverAgents finds agents matching the given requirements
func (d *AgentDiscovery) DiscoverAgents(ctx context.Context, requirements Requirements) ([]AgentMatch, error) {
	// Build dynamic query based on client schema
	schemaName := fmt.Sprintf("client_%s", requirements.ClientID)

	query := fmt.Sprintf(`
        SELECT 
            ai.id,
            ai.name,
            ai.config->>'agent_type' as agent_type,
            ai.config->>'topic' as topic,
            COALESCE(am.success_rate, 0.5) as performance,
            CASE 
                WHEN ai.config->>'last_activity' IS NOT NULL 
                AND (ai.config->>'last_activity')::timestamp > NOW() - INTERVAL '5 minutes'
                THEN true 
                ELSE false 
            END as available
        FROM %s.agent_instances ai
        LEFT JOIN agent_metrics am ON ai.id = am.agent_id
        WHERE ai.is_active = true
        AND ($1::text[] IS NULL OR ai.config->'capabilities' ?| $1)
        AND ($2::text IS NULL OR ai.config->>'agent_type' = $2)
        ORDER BY 
            am.success_rate DESC NULLS LAST,
            am.avg_response_time ASC NULLS LAST
        LIMIT 10
    `, schemaName)

	rows, err := d.db.Query(ctx, query, requirements.Capabilities, requirements.AgentType)
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents: %w", err)
	}
	defer rows.Close()

	var matches []AgentMatch
	for rows.Next() {
		var m AgentMatch
		err := rows.Scan(&m.AgentID, &m.AgentName, &m.AgentType,
			&m.Topic, &m.Performance, &m.Available)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent match: %w", err)
		}

		// Apply additional filters
		if requirements.MinPerformance > 0 && m.Performance < requirements.MinPerformance {
			continue
		}

		matches = append(matches, m)
	}

	return matches, nil
}

// FindBestGroup finds the most suitable group for a task type
// Now supports optional version specification
func (d *GroupDiscovery) FindBestGroup(ctx context.Context, taskType string, version string) (*AgentGroup, error) {
	var group AgentGroup
	var lastPerf sql.NullFloat64
	var capabilities []byte

	var query string
	var args []interface{}

	if version != "" {
		query = `
            SELECT id, name, group_type, agent_configs, orchestration_workflow,
                   usage_count, version
            FROM agent_group_definitions
            WHERE group_type = $1
            AND version >= $2
            ORDER BY 
                version DESC,
                usage_count DESC
            LIMIT 1
        `
		args = []interface{}{taskType, version}
	} else {
		query = `
            SELECT id, name, group_type, agent_configs, orchestration_workflow,
                   usage_count, version
            FROM agent_group_definitions
            WHERE group_type = $1
            ORDER BY 
                usage_count DESC, 
                version DESC
            LIMIT 1
        `
		args = []interface{}{taskType}
	}

	err := d.db.QueryRowContext(ctx, query, args...).Scan(
		&group.ID, &group.Name, &group.GroupType,
		&group.AgentConfigs, &group.Workflow,
		&capabilities, &group.UsageCount, &group.Version,
		&lastPerf)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no group found for task type: %s", taskType)
		}
		return nil, fmt.Errorf("failed to find group: %w", err)
	}

	// Parse capabilities
	if err := json.Unmarshal(capabilities, &group.Capabilities); err != nil {
		// Don't fail, just log and continue
		group.Capabilities = []string{}
	}

	group.LastPerformance = lastPerf.Float64
	return &group, nil
}

// FindBestGroup finds the most suitable group for a task type
func (d *GroupDiscovery) FindBestGroupOLD(ctx context.Context, taskType string, requirements map[string]interface{}) (*AgentGroup, error) {
	var group AgentGroup
	var lastPerf sql.NullFloat64
	var capabilities []byte

	err := d.db.QueryRow(ctx, `
        SELECT id, name, group_type, agent_configs, orchestration_workflow,
               capabilities, usage_count, version,
               COALESCE((performance_metrics->>'success_rate')::float, 0.5)
        FROM agent_groups
        WHERE group_type = $1
        AND is_active = true
        ORDER BY 
            COALESCE((performance_metrics->>'success_rate')::float, 0.5) DESC,
            usage_count DESC, 
            version DESC
        LIMIT 1
    `, taskType).Scan(
		&group.ID, &group.Name, &group.GroupType,
		&group.AgentConfigs, &group.Workflow,
		&capabilities, &group.UsageCount, &group.Version,
		&lastPerf)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no group found for task type: %s", taskType)
		}
		return nil, fmt.Errorf("failed to find group: %w", err)
	}

	// Parse capabilities
	if err := json.Unmarshal(capabilities, &group.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities: %w", err)
	}

	group.LastPerformance = lastPerf.Float64

	return &group, nil
}

// DiscoverGroups finds groups that can handle the specified task type or have required capabilities
func (d *GroupDiscovery) DiscoverGroups(ctx context.Context, taskType string, capabilities []string) ([]*AgentGroup, error) {
	capabilitiesJSON, _ := json.Marshal(capabilities)

	query := `
        SELECT id, name, group_type, agent_configs, orchestration_workflow,
               capabilities, usage_count, version,
               COALESCE((performance_metrics->>'success_rate')::float, 0.5) as performance
        FROM agent_groups
        WHERE (group_type = $1 OR capabilities @> $2::jsonb)
        AND is_active = true
        ORDER BY 
            CASE WHEN group_type = $1 THEN 0 ELSE 1 END,  -- Exact matches first
            performance DESC,
            usage_count DESC,
            version DESC
        LIMIT 10
    `

	rows, err := d.db.Query(ctx, query, taskType, capabilitiesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to discover groups: %w", err)
	}
	defer rows.Close()

	var groups []*AgentGroup
	for rows.Next() {
		var group AgentGroup
		var caps []byte
		var perf float64

		err := rows.Scan(
			&group.ID, &group.Name, &group.GroupType,
			&group.AgentConfigs, &group.Workflow,
			&caps, &group.UsageCount, &group.Version,
			&perf)

		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		// Parse capabilities
		if err := json.Unmarshal(caps, &group.Capabilities); err != nil {
			continue // Skip groups with invalid capabilities
		}

		group.LastPerformance = perf
		groups = append(groups, &group)
	}

	return groups, nil
}

// DiscoverAgentsByCapability finds all agents with specific capabilities
func (d *AgentDiscovery) DiscoverAgentsByCapability(ctx context.Context, clientID string, capabilities []string) ([]AgentMatch, error) {
	return d.DiscoverAgents(ctx, Requirements{
		ClientID:     clientID,
		Capabilities: capabilities,
	})
}

// UpdateGroupPerformance updates the performance metrics for a group
func (d *GroupDiscovery) UpdateGroupPerformance(ctx context.Context, groupID string, successRate float64) error {
	_, err := d.db.Exec(ctx, `
        UPDATE agent_groups 
        SET performance_metrics = jsonb_set(
            COALESCE(performance_metrics, '{}'::jsonb),
            '{success_rate}',
            to_jsonb($1::float)
        ),
        last_used_at = NOW()
        WHERE id = $2
    `, successRate, groupID)

	return err
}

// GetGroupHistory returns the evolution history of a group
func (d *GroupDiscovery) GetGroupHistory(ctx context.Context, groupID string) ([]*AgentGroup, error) {
	// Recursive CTE to get all versions in the lineage
	query := `
        WITH RECURSIVE group_history AS (
            -- Base case: start with the given group
            SELECT * FROM agent_groups WHERE id = $1
            
            UNION ALL
            
            -- Recursive case: find parent groups
            SELECT g.* 
            FROM agent_groups g
            INNER JOIN group_history h ON g.id = h.parent_id
        )
        SELECT id, name, group_type, agent_configs, orchestration_workflow,
               capabilities, usage_count, version,
               COALESCE((performance_metrics->>'success_rate')::float, 0.5) as performance
        FROM group_history
        ORDER BY created_at ASC
    `

	rows, err := d.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group history: %w", err)
	}
	defer rows.Close()

	var groups []*AgentGroup
	for rows.Next() {
		var group AgentGroup
		var caps []byte
		var perf float64

		err := rows.Scan(
			&group.ID, &group.Name, &group.GroupType,
			&group.AgentConfigs, &group.Workflow,
			&caps, &group.UsageCount, &group.Version,
			&perf)

		if err != nil {
			return nil, fmt.Errorf("failed to scan group history: %w", err)
		}

		if err := json.Unmarshal(caps, &group.Capabilities); err != nil {
			continue
		}

		group.LastPerformance = perf
		groups = append(groups, &group)
	}

	return groups, nil
}
