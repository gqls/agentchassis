package actions

import (
	"context"
	"fmt"
)

// Implementation
func QueryDatabaseAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	query, ok := config["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query_database requires 'query' in config")
	}

	outputFormat := "array"
	if of, ok := config["output_format"].(string); ok {
		outputFormat = of
	}

	// Execute query
	rows, err := params.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, _ := rows.Columns()

	// Build results
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if outputFormat == "array" {
		return results, nil
	}
	return map[string]interface{}{"rows": results, "count": len(results)}, nil
}
