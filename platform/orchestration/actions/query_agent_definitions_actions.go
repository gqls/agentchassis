// FILE: platform/orchestration/actions/query_agent_definitions.go
// Action to query agent_definitions table for dynamic discovery

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// QueryAgentDefinitionsAction queries agent_definitions to discover available agents
//
// This enables self-describing systems where agents can discover what other agents
// exist without hardcoding. Primary use case: classifier discovering available builders.
//
// Config:
//   - filter: map of field conditions (category, capabilities, type_pattern)
//   - fields: []string of fields to return (defaults to type, display_name, description)
//   - active_only: bool (default true) - only return is_active=true
//   - order_by: string (default "usage_count DESC, display_name ASC")
//   - limit: int (default 20)
//
// Example config:
//
//	{
//	  "filter": {
//	    "category": "builder",
//	    "type_pattern": "%-builder"
//	  },
//	  "fields": ["type", "display_name", "description", "capabilities"]
//	}
//
// Returns:
//
//	{
//	  "agents": [
//	    {"type": "landing-page-builder", "display_name": "Landing Page Builder", ...},
//	    {"type": "content-site-builder", "display_name": "Content Site Builder", ...}
//	  ],
//	  "count": 2
//	}
func QueryAgentDefinitionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger

	logger.Info("QueryAgentDefinitionsAction starting",
		zap.String("step_name", params.ExecutionContext.StepName))

	// Parse config
	filter := extractMapConfig(config, "filter")
	fields := extractStringSliceConfig(config, "fields", []string{"type", "display_name", "description"})
	activeOnly := extractBoolConfig(config, "active_only", true)
	orderBy := extractStringConfig(config, "order_by", "usage_count DESC, display_name ASC")
	limit := extractIntConfig(config, "limit", 20)

	// Build query
	query, args := buildAgentDefinitionsQuery(fields, filter, activeOnly, orderBy, limit)

	logger.Debug("Executing agent definitions query",
		zap.String("query", query),
		zap.Any("args", args))

	// Execute query
	rows, err := queryRows(ctx, params.DB, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent_definitions: %w", err)
	}
	defer rows.Close()

	// Scan results
	agents := []map[string]interface{}{}
	for rows.Next() {
		agent, err := scanAgentRow(rows, fields)
		if err != nil {
			logger.Warn("Failed to scan agent row", zap.Error(err))
			continue
		}
		agents = append(agents, agent)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating agent rows: %w", err)
	}

	logger.Info("Query completed",
		zap.Int("count", len(agents)),
		zap.Any("filter", filter))

	return map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	}, nil
}

// buildAgentDefinitionsQuery constructs the SQL query
func buildAgentDefinitionsQuery(fields []string, filter map[string]interface{}, activeOnly bool, orderBy string, limit int) (string, []interface{}) {
	// Validate and sanitize field names
	validFields := map[string]bool{
		"id": true, "type": true, "display_name": true, "description": true,
		"category": true, "capabilities": true, "version": true,
		"usage_count": true, "is_snapshot": true, "created_at": true,
	}

	selectFields := []string{}
	for _, f := range fields {
		if validFields[f] {
			selectFields = append(selectFields, f)
		}
	}
	if len(selectFields) == 0 {
		selectFields = []string{"type", "display_name", "description"}
	}

	// Build WHERE clauses
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if activeOnly {
		whereClauses = append(whereClauses, "is_active = true")
	}

	// Filter by category
	if category, ok := filter["category"].(string); ok && category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, category)
		argIndex++
	}

	// Filter by type pattern (LIKE)
	if typePattern, ok := filter["type_pattern"].(string); ok && typePattern != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("type LIKE $%d", argIndex))
		args = append(args, typePattern)
		argIndex++
	}

	// Filter by exact type
	if exactType, ok := filter["type"].(string); ok && exactType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, exactType)
		argIndex++
	}

	// Filter by capability (JSON contains)
	if capability, ok := filter["capability"].(string); ok && capability != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("capabilities @> $%d::jsonb", argIndex))
		capJSON, _ := json.Marshal([]string{capability})
		args = append(args, string(capJSON))
		argIndex++
	}

	// Filter by multiple capabilities (must have all)
	if capabilities, ok := filter["capabilities"].([]interface{}); ok && len(capabilities) > 0 {
		caps := []string{}
		for _, c := range capabilities {
			if s, ok := c.(string); ok {
				caps = append(caps, s)
			}
		}
		if len(caps) > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("capabilities @> $%d::jsonb", argIndex))
			capJSON, _ := json.Marshal(caps)
			args = append(args, string(capJSON))
			argIndex++
		}
	}

	// Construct query
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Sanitize orderBy (basic protection)
	safeOrderBy := sanitizeOrderBy(orderBy)

	query := fmt.Sprintf(`
		SELECT %s
		FROM agent_definitions
		%s
		ORDER BY %s
		LIMIT %d
	`, strings.Join(selectFields, ", "), whereClause, safeOrderBy, limit)

	return query, args
}

// sanitizeOrderBy validates order by clause
func sanitizeOrderBy(orderBy string) string {
	validColumns := map[string]bool{
		"type": true, "display_name": true, "category": true,
		"usage_count": true, "version": true, "created_at": true, "updated_at": true,
	}

	// Parse and validate each part
	parts := strings.Split(orderBy, ",")
	safeParts := []string{}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		tokens := strings.Fields(part)
		if len(tokens) == 0 {
			continue
		}

		col := strings.ToLower(tokens[0])
		if !validColumns[col] {
			continue
		}

		dir := "ASC"
		if len(tokens) > 1 && strings.ToUpper(tokens[1]) == "DESC" {
			dir = "DESC"
		}

		safeParts = append(safeParts, fmt.Sprintf("%s %s", col, dir))
	}

	if len(safeParts) == 0 {
		return "display_name ASC"
	}

	return strings.Join(safeParts, ", ")
}

// queryRows handles both sql.DB and pgxpool.Pool
func queryRows(ctx context.Context, db interface{}, query string, args ...interface{}) (RowScanner, error) {
	switch d := db.(type) {
	case *sql.DB:
		rows, err := d.QueryContext(ctx, query, args...)
		return &sqlRowScanner{rows}, err
	case *pgxpool.Pool:
		rows, err := d.Query(ctx, query, args...)
		return &pgxRowScanner{rows}, err
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}
}

// RowScanner interface for abstracting row scanning
type RowScanner interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
	Close() error
}

type sqlRowScanner struct {
	rows *sql.Rows
}

func (s *sqlRowScanner) Next() bool                     { return s.rows.Next() }
func (s *sqlRowScanner) Scan(dest ...interface{}) error { return s.rows.Scan(dest...) }
func (s *sqlRowScanner) Err() error                     { return s.rows.Err() }
func (s *sqlRowScanner) Close() error                   { return s.rows.Close() }

type pgxRowScanner struct {
	rows interface {
		Next() bool
		Scan(dest ...interface{}) error
		Err() error
		Close()
	}
}

func (p *pgxRowScanner) Next() bool                     { return p.rows.Next() }
func (p *pgxRowScanner) Scan(dest ...interface{}) error { return p.rows.Scan(dest...) }
func (p *pgxRowScanner) Err() error                     { return p.rows.Err() }
func (p *pgxRowScanner) Close() error                   { p.rows.Close(); return nil }

// scanAgentRow scans a row into a map based on requested fields
func scanAgentRow(rows RowScanner, fields []string) (map[string]interface{}, error) {
	// Create scan destinations
	values := make([]interface{}, len(fields))
	valuePtrs := make([]interface{}, len(fields))

	for i := range fields {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	// Build result map
	result := make(map[string]interface{})
	for i, field := range fields {
		val := values[i]

		// Handle []byte for JSONB fields
		if b, ok := val.([]byte); ok {
			var parsed interface{}
			if err := json.Unmarshal(b, &parsed); err == nil {
				val = parsed
			} else {
				val = string(b)
			}
		}

		result[field] = val
	}

	return result, nil
}

// Config extraction helpers
func extractMapConfig(config map[string]interface{}, key string) map[string]interface{} {
	if v, ok := config[key].(map[string]interface{}); ok {
		return v
	}
	return map[string]interface{}{}
}

func extractStringSliceConfig(config map[string]interface{}, key string, defaultVal []string) []string {
	if v, ok := config[key].([]interface{}); ok {
		result := []string{}
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultVal
}

func extractStringConfig(config map[string]interface{}, key string, defaultVal string) string {
	if v, ok := config[key].(string); ok && v != "" {
		return v
	}
	return defaultVal
}

func extractBoolConfig(config map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := config[key].(bool); ok {
		return v
	}
	return defaultVal
}

func extractIntConfig(config map[string]interface{}, key string, defaultVal int) int {
	if v, ok := config[key].(float64); ok {
		return int(v)
	}
	if v, ok := config[key].(int); ok {
		return v
	}
	return defaultVal
}
