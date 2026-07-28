package actions

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
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

	// Resolve query parameters from collected_data
	var queryArgs []interface{}
	if paramPaths, ok := config["params"].([]interface{}); ok {
		for _, p := range paramPaths {
			pathStr, ok := p.(string)
			if !ok {
				continue
			}
			// The $ctx. namespace resolves from the RUN's own identity rather
			// than from collected_data — see executionContextParam.
			if ctxValue, isCtx, err := executionContextParam(params.ExecutionContext, pathStr); isCtx {
				if err != nil {
					return nil, err
				}
				queryArgs = append(queryArgs, ctxValue)
				continue
			}
			value := datahelpers.ExtractNestedField(params.CollectedData, pathStr)
			if value == nil {
				// Try with input_data prefix
				value = datahelpers.ExtractNestedField(params.CollectedData, "input_data."+pathStr)
			}
			if value == nil {
				return nil, fmt.Errorf("query param path '%s' resolved to nil", pathStr)
			}
			// Convert to string for SQL parameters
			switch v := value.(type) {
			case string:
				queryArgs = append(queryArgs, v)
			default:
				queryArgs = append(queryArgs, fmt.Sprintf("%v", v))
			}
		}
		params.Logger.Info("QueryDatabaseAction: Resolved params",
			zap.Int("count", len(queryArgs)),
			zap.Strings("paths", func() []string {
				s := make([]string, len(paramPaths))
				for i, p := range paramPaths {
					s[i] = fmt.Sprintf("%v", p)
				}
				return s
			}()),
		)
	}

	if params.DB == nil {
		params.Logger.Warn("QueryDatabaseAction: No database connection")
		return map[string]interface{}{
			"error": "no database connection", "results": []interface{}{}, "count": 0,
		}, nil
	}

	rows, err := params.DB.QueryContext(ctx, query, queryArgs...)
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
