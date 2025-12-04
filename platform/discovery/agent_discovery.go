// FILE: platform/discovery/agent_discovery.go
// Every agent is an orchestrator - queries agent_definitions only

package discovery

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
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

// AgentDefinitionResult represents an agent definition lookup result
type AgentDefinitionResult struct {
	ID                    string
	Type                  string
	DisplayName           string
	Description           string
	Category              string
	Workflow              json.RawMessage // from default_config->'workflow'
	DefaultConfig         json.RawMessage // full default_config
	Capabilities          []string
	BriefingQuestionnaire json.RawMessage
	UsageCount            int
	Version               int
	IsSnapshot            bool
}

// AgentDiscovery handles finding individual agent instances
type AgentDiscovery struct {
	db *sql.DB
}

// AgentDefinitionDiscovery handles finding agent definitions
type AgentDefinitionDiscovery struct {
	db *sql.DB
}

// NewAgentDiscovery creates a new agent discovery service
func NewAgentDiscovery(db *sql.DB) *AgentDiscovery {
	return &AgentDiscovery{db: db}
}

// NewAgentDefinitionDiscovery creates a new agent definition discovery service
func NewAgentDefinitionDiscovery(db *sql.DB) *AgentDefinitionDiscovery {
	return &AgentDefinitionDiscovery{db: db}
}

// ============================================================================
// Agent Definition Discovery
// ============================================================================

// FindByType finds an agent definition by type, optionally with minimum version
func (d *AgentDefinitionDiscovery) FindByType(ctx context.Context, agentType string, minVersion int, logger *zap.Logger) (*AgentDefinitionResult, error) {
	var result AgentDefinitionResult
	var capabilitiesJSON []byte
	var defaultConfigJSON []byte
	var briefingQuestionnaireJSON sql.NullString

	logger.Info("AgentDefinitionDiscovery.FindByType",
		zap.String("agent_type", agentType),
		zap.Int("min_version", minVersion),
	)

	var query string
	var args []interface{}

	if minVersion > 0 {
		query = `
			SELECT 
				id::text, type, display_name, COALESCE(description, ''), COALESCE(category, ''),
				default_config, 
				default_config->'workflow' as workflow,
				COALESCE(capabilities, '[]'::jsonb) as capabilities,
				briefing_questionnaire,
				COALESCE(usage_count, 0) as usage_count,
				version,
				COALESCE(is_snapshot, false) as is_snapshot
			FROM agent_definitions
			WHERE type = $1
			AND version >= $2
			AND is_active = true
			ORDER BY version DESC
			LIMIT 1
		`
		args = []interface{}{agentType, minVersion}
	} else {
		query = `
			SELECT 
				id::text, type, display_name, COALESCE(description, ''), COALESCE(category, ''),
				default_config,
				default_config->'workflow' as workflow,
				COALESCE(capabilities, '[]'::jsonb) as capabilities,
				briefing_questionnaire,
				COALESCE(usage_count, 0) as usage_count,
				version,
				COALESCE(is_snapshot, false) as is_snapshot
			FROM agent_definitions
			WHERE type = $1
			AND is_active = true
			ORDER BY version DESC
			LIMIT 1
		`
		args = []interface{}{agentType}
	}

	err := d.db.QueryRowContext(ctx, query, args...).Scan(
		&result.ID,
		&result.Type,
		&result.DisplayName,
		&result.Description,
		&result.Category,
		&defaultConfigJSON,
		&result.Workflow,
		&capabilitiesJSON,
		&briefingQuestionnaireJSON,
		&result.UsageCount,
		&result.Version,
		&result.IsSnapshot,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no agent definition found for type: %s", agentType)
		}
		return nil, fmt.Errorf("failed to find agent definition: %w", err)
	}

	result.DefaultConfig = defaultConfigJSON

	if briefingQuestionnaireJSON.Valid {
		result.BriefingQuestionnaire = json.RawMessage(briefingQuestionnaireJSON.String)
	} else {
		result.BriefingQuestionnaire = json.RawMessage("{}")
	}

	if err := json.Unmarshal(capabilitiesJSON, &result.Capabilities); err != nil {
		result.Capabilities = []string{}
	}

	logger.Info("AgentDefinitionDiscovery.FindByType: found",
		zap.String("agent_type", agentType),
		zap.String("agent_id", result.ID),
		zap.String("display_name", result.DisplayName),
		zap.Int("version", result.Version),
	)

	return &result, nil
}

// FindByCapabilities finds agent definitions that have all specified capabilities
func (d *AgentDefinitionDiscovery) FindByCapabilities(ctx context.Context, capabilities []string, logger *zap.Logger) ([]*AgentDefinitionResult, error) {
	capabilitiesJSON, err := json.Marshal(capabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	query := `
		SELECT 
			id::text, type, display_name, COALESCE(description, ''), COALESCE(category, ''),
			default_config, 
			default_config->'workflow' as workflow,
			COALESCE(capabilities, '[]'::jsonb) as capabilities,
			briefing_questionnaire,
			COALESCE(usage_count, 0) as usage_count,
			version,
			COALESCE(is_snapshot, false) as is_snapshot
		FROM agent_definitions
		WHERE capabilities @> $1::jsonb
		AND is_active = true
		ORDER BY usage_count DESC, version DESC
		LIMIT 10
	`

	rows, err := d.db.QueryContext(ctx, query, capabilitiesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to find agents by capabilities: %w", err)
	}
	defer rows.Close()

	var results []*AgentDefinitionResult
	for rows.Next() {
		var result AgentDefinitionResult
		var capabilitiesJSON []byte
		var defaultConfigJSON []byte
		var briefingQuestionnaireJSON sql.NullString

		err := rows.Scan(
			&result.ID,
			&result.Type,
			&result.DisplayName,
			&result.Description,
			&result.Category,
			&defaultConfigJSON,
			&result.Workflow,
			&capabilitiesJSON,
			&briefingQuestionnaireJSON,
			&result.UsageCount,
			&result.Version,
			&result.IsSnapshot,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent definition: %w", err)
		}

		result.DefaultConfig = defaultConfigJSON

		if briefingQuestionnaireJSON.Valid {
			result.BriefingQuestionnaire = json.RawMessage(briefingQuestionnaireJSON.String)
		} else {
			result.BriefingQuestionnaire = json.RawMessage("{}")
		}

		if err := json.Unmarshal(capabilitiesJSON, &result.Capabilities); err != nil {
			result.Capabilities = []string{}
		}

		results = append(results, &result)
	}

	return results, nil
}

// UpdateUsageCount increments the usage count for an agent definition
func (d *AgentDefinitionDiscovery) UpdateUsageCount(ctx context.Context, agentType string) error {
	_, err := d.db.ExecContext(ctx, `
		UPDATE agent_definitions 
		SET usage_count = COALESCE(usage_count, 0) + 1,
		    updated_at = NOW()
		WHERE type = $1
	`, agentType)
	return err
}

// ============================================================================
// Agent Instance Discovery (for client schemas)
// ============================================================================

// DiscoverAgents finds agent instances matching the given requirements
func (d *AgentDiscovery) DiscoverAgents(ctx context.Context, requirements Requirements) ([]AgentMatch, error) {
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

	rows, err := d.db.Query(query, requirements.Capabilities, requirements.AgentType)
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

		if requirements.MinPerformance > 0 && m.Performance < requirements.MinPerformance {
			continue
		}

		matches = append(matches, m)
	}

	return matches, nil
}

// DiscoverAgentsByCapability finds all agents with specific capabilities
func (d *AgentDiscovery) DiscoverAgentsByCapability(ctx context.Context, clientID string, capabilities []string) ([]AgentMatch, error) {
	return d.DiscoverAgents(ctx, Requirements{
		ClientID:     clientID,
		Capabilities: capabilities,
	})
}

// ============================================================================
// BACKWARD COMPATIBILITY: GroupDiscovery alias
// Allows existing code using GroupDiscovery to continue working
// ============================================================================

// GroupDiscovery is an alias for AgentDefinitionDiscovery
// Deprecated: Use AgentDefinitionDiscovery directly
type GroupDiscovery = AgentDefinitionDiscovery

// NewGroupDiscovery creates a new discovery service
// Deprecated: Use NewAgentDefinitionDiscovery
func NewGroupDiscovery(db *sql.DB) *GroupDiscovery {
	return NewAgentDefinitionDiscovery(db)
}

// AgentGroup is kept for code that expects this type
// The new code should use AgentDefinitionResult
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

// FindBestGroup provides backward compatibility
// Converts AgentDefinitionResult to AgentGroup format
func (d *AgentDefinitionDiscovery) FindBestGroup(ctx context.Context, agentType string, version string, logger zap.Logger) (*AgentGroup, error) {
	// Convert version string to int
	versionInt := 0
	if version != "" {
		fmt.Sscanf(version, "%d", &versionInt)
	}

	result, err := d.FindByType(ctx, agentType, versionInt, &logger)
	if err != nil {
		return nil, err
	}

	// Convert to legacy format
	return &AgentGroup{
		ID:              result.ID,
		Name:            result.DisplayName,
		GroupType:       result.Type,
		Workflow:        result.Workflow,
		Capabilities:    result.Capabilities,
		UsageCount:      result.UsageCount,
		Version:         fmt.Sprintf("%d", result.Version),
		AgentConfigs:    json.RawMessage("[]"),
		LastPerformance: 0.0,
	}, nil
}
