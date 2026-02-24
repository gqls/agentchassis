package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func QueryDatabaseAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("QueryDatabaseAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	query, ok := config["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query_database requires 'query' in config")
	}

	outputFormat := "array"
	if of, ok := config["output_format"].(string); ok && of != "" {
		outputFormat = of
	}

	if params.DB == nil {
		params.Logger.Warn("QueryDatabaseAction: No database connection")
		return map[string]interface{}{
			"error": "no database connection", "results": []interface{}{}, "count": 0,
		}, nil
	}

	rows, err := params.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range columns {
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}
		results = append(results, row)
	}

	params.Logger.Info("QueryDatabaseAction: Complete", zap.Int("count", len(results)))

	if outputFormat == "array" {
		return results, nil
	}
	// "object" format: include metadata + flatten first row for easy path access
	result := map[string]interface{}{
		"rows":    results,
		"count":   len(results),
		"columns": columns,
	}
	// Flatten first row fields to top level so paths like
	// "dispatchable.domain" work without array indexing
	if len(results) > 0 {
		for k, v := range results[0] {
			result[k] = v
		}
	}
	return result, nil
}
