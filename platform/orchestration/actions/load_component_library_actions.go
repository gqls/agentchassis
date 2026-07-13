// ===========================================================================
// ACTION: load_component_library
// FILE: platform/orchestration/actions/load_component_library_action.go
// ===========================================================================
// Queries available components from content_components table and returns
// them organized for LLM consumption. Used by site-architect to know what
// components are available when planning page sections.
// ===========================================================================

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ComponentLibraryEntry represents a component available for use
type ComponentLibraryEntry struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Function       string   `json:"function"`
	DisplayName    string   `json:"display_name"`
	Description    string   `json:"description"`
	Category       string   `json:"category"`
	ComponentLevel string   `json:"component_level"`
	RenderMode     string   `json:"render_mode"`
	SemanticTags   []string `json:"semantic_tags,omitempty"`
}

// ComponentLibraryResult is the output of LoadComponentLibraryAction
type ComponentLibraryResult struct {
	// Components grouped by function (for easy lookup)
	ByFunction map[string]ComponentLibraryEntry `json:"by_function"`

	// Components grouped by category
	ByCategory map[string][]ComponentLibraryEntry `json:"by_category"`

	// Flat list of section-level components (most common use case)
	Sections []ComponentLibraryEntry `json:"sections"`

	// Available function names (for LLM prompt)
	AvailableFunctions []string `json:"available_functions"`

	// Formatted for LLM prompt (human-readable list)
	ForPrompt string `json:"for_prompt"`

	// Count
	TotalCount int `json:"total_count"`
}

// LoadComponentLibraryAction queries the content_components table and returns
// organized component data suitable for LLM consumption.
//
// Config:
//   - component_level: filter by level ("section", "header", "footer", "head", "element")
//     default: "section"
//   - categories: array of categories to include (optional, empty = all)
//   - include_inactive: bool, whether to include inactive components (default: false)
//   - format_for_prompt: bool, whether to generate the ForPrompt string (default: true)
func LoadComponentLibraryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadComponentLibraryAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		params.Logger.Error("LoadComponentLibraryAction: No database connection")
		return nil, fmt.Errorf("database connection required for component library")
	}

	config := params.StepConfig.Config

	// Parse config options
	componentLevel := "section"
	if level, ok := config["component_level"].(string); ok && level != "" {
		componentLevel = level
	}

	var categories []string
	if cats, ok := config["categories"].([]interface{}); ok {
		for _, c := range cats {
			if catStr, ok := c.(string); ok {
				categories = append(categories, catStr)
			}
		}
	}

	includeInactive := false
	if inc, ok := config["include_inactive"].(bool); ok {
		includeInactive = inc
	}

	formatForPrompt := true
	if fmt, ok := config["format_for_prompt"].(bool); ok {
		formatForPrompt = fmt
	}

	// Build and execute query
	components, err := queryComponentLibrary(ctx, params.DB, componentLevel, categories, includeInactive, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query component library: %w", err)
	}

	// Organize results
	result := organizeComponents(components, formatForPrompt)

	params.Logger.Info("LoadComponentLibraryAction: Loaded components",
		zap.Int("total", result.TotalCount),
		zap.Int("categories", len(result.ByCategory)),
		zap.Strings("functions", result.AvailableFunctions),
	)

	return result, nil
}

// queryComponentLibrary fetches components from the database
func queryComponentLibrary(
	ctx context.Context,
	db interface{},
	componentLevel string,
	categories []string,
	includeInactive bool,
	logger *zap.Logger,
) ([]ComponentLibraryEntry, error) {

	// Build query
	query := `
		SELECT 
			id::text,
			name,
			function,
			COALESCE(display_name, name) as display_name,
			COALESCE(description, '') as description,
			COALESCE(category, '') as category,
			COALESCE(component_level, 'section') as component_level,
			COALESCE(render_mode, 'template') as render_mode,
			COALESCE(semantic_tags, '[]'::jsonb) as semantic_tags
		FROM content_components
		WHERE 1=1
	`

	var args []interface{}
	argNum := 1

	// Filter by component_level
	if componentLevel != "" && componentLevel != "all" {
		query += fmt.Sprintf(" AND component_level = $%d", argNum)
		args = append(args, componentLevel)
		argNum++
	}

	// Filter by categories
	if len(categories) > 0 {
		placeholders := make([]string, len(categories))
		for i, cat := range categories {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, cat)
			argNum++
		}
		query += fmt.Sprintf(" AND category IN (%s)", strings.Join(placeholders, ", "))
	}

	// Filter active
	if !includeInactive {
		query += " AND (is_active = true OR is_active IS NULL)"
	}

	query += " ORDER BY category, sort_order NULLS LAST, function"

	logger.Debug("LoadComponentLibraryAction: Executing query",
		zap.String("query", query),
		zap.Any("args", args),
	)

	// Execute query
	var rows interface{ Next() bool }
	var err error

	switch d := db.(type) {
	case *sql.DB:
		r, e := d.QueryContext(ctx, query, args...)
		if e != nil {
			return nil, e
		}
		defer r.Close()
		return scanComponentRows(r, logger)

	case *pgxpool.Pool:
		r, e := d.Query(ctx, query, args...)
		if e != nil {
			return nil, e
		}
		defer r.Close()
		return scanPgxComponentRows(r, logger)

	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	_ = rows
	_ = err
	return nil, nil
}

// scanComponentRows scans sql.Rows into ComponentLibraryEntry slice
func scanComponentRows(rows *sql.Rows, logger *zap.Logger) ([]ComponentLibraryEntry, error) {
	var components []ComponentLibraryEntry

	for rows.Next() {
		var entry ComponentLibraryEntry
		var semanticTagsJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.Name,
			&entry.Function,
			&entry.DisplayName,
			&entry.Description,
			&entry.Category,
			&entry.ComponentLevel,
			&entry.RenderMode,
			&semanticTagsJSON,
		)
		if err != nil {
			logger.Warn("Failed to scan component row", zap.Error(err))
			continue
		}

		// Parse semantic tags
		if len(semanticTagsJSON) > 0 {
			datahelpers.SafeUnmarshal(semanticTagsJSON, &entry.SemanticTags)
		}

		components = append(components, entry)
	}

	return components, rows.Err()
}

// scanPgxComponentRows scans pgx rows into ComponentLibraryEntry slice
func scanPgxComponentRows(rows interface{ Next() bool }, logger *zap.Logger) ([]ComponentLibraryEntry, error) {
	var components []ComponentLibraryEntry

	// Type assert to get the actual pgx rows interface
	type pgxScanner interface {
		Next() bool
		Scan(dest ...interface{}) error
		Err() error
	}

	pgxRows, ok := rows.(pgxScanner)
	if !ok {
		return nil, fmt.Errorf("invalid pgx rows type")
	}

	for pgxRows.Next() {
		var entry ComponentLibraryEntry
		var semanticTagsJSON []byte

		err := pgxRows.Scan(
			&entry.ID,
			&entry.Name,
			&entry.Function,
			&entry.DisplayName,
			&entry.Description,
			&entry.Category,
			&entry.ComponentLevel,
			&entry.RenderMode,
			&semanticTagsJSON,
		)
		if err != nil {
			logger.Warn("Failed to scan component row", zap.Error(err))
			continue
		}

		// Parse semantic tags
		if len(semanticTagsJSON) > 0 {
			datahelpers.SafeUnmarshal(semanticTagsJSON, &entry.SemanticTags)
		}

		components = append(components, entry)
	}

	return components, pgxRows.Err()
}

// organizeComponents organizes the flat list into various useful structures
func organizeComponents(components []ComponentLibraryEntry, formatForPrompt bool) *ComponentLibraryResult {
	result := &ComponentLibraryResult{
		ByFunction:         make(map[string]ComponentLibraryEntry),
		ByCategory:         make(map[string][]ComponentLibraryEntry),
		Sections:           []ComponentLibraryEntry{},
		AvailableFunctions: []string{},
		TotalCount:         len(components),
	}

	// Track unique functions (some may have duplicates in different categories)
	seenFunctions := make(map[string]bool)

	for _, comp := range components {
		// Add to ByFunction (first one wins if duplicates)
		if _, exists := result.ByFunction[comp.Function]; !exists {
			result.ByFunction[comp.Function] = comp
		}

		// Add to ByCategory
		cat := comp.Category
		if cat == "" {
			cat = "general"
		}
		result.ByCategory[cat] = append(result.ByCategory[cat], comp)

		// Add to Sections list
		if comp.ComponentLevel == "section" {
			result.Sections = append(result.Sections, comp)
		}

		// Track unique functions
		if !seenFunctions[comp.Function] {
			seenFunctions[comp.Function] = true
			result.AvailableFunctions = append(result.AvailableFunctions, comp.Function)
		}
	}

	// Format for LLM prompt
	if formatForPrompt {
		result.ForPrompt = formatComponentsForPrompt(components)
	}

	return result
}

// formatComponentsForPrompt creates a human-readable component list for LLM consumption
func formatComponentsForPrompt(components []ComponentLibraryEntry) string {
	var sb strings.Builder

	// Group by category for cleaner output
	byCategory := make(map[string][]ComponentLibraryEntry)
	for _, comp := range components {
		cat := comp.Category
		if cat == "" {
			cat = "general"
		}
		byCategory[cat] = append(byCategory[cat], comp)
	}

	sb.WriteString("Available section components:\n\n")

	for category, comps := range byCategory {
		sb.WriteString(fmt.Sprintf("## %s\n", strings.Title(category)))
		for _, comp := range comps {
			desc := comp.Description
			if desc == "" {
				desc = comp.DisplayName
			}
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", comp.Function, desc))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ===========================================================================
// ACTION REGISTRATION
// ===========================================================================
// Add to your action registry:
//
// registry.Register("load_component_library", registry.ActionDefinition{
//     Handler:     LoadComponentLibraryAction,
//     Description: "Load available components from the component library",
//     Category:    "database",
//     Status:      registry.StatusActive,
// })
// ===========================================================================
