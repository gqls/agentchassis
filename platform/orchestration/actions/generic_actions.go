// FILE: platform/orchestration/actions/generic_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ValidateSchemaAction validates data against a JSON schema
func ValidateSchemaAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	// Get the schema from config
	schema, ok := config["schema"].(map[string]interface{})
	if !ok {
		// If no schema specified, just pass through
		return map[string]interface{}{
			"validated": true,
			"data":      params.CollectedData,
		}, nil
	}

	// Get validation type
	validationType, _ := config["validation_type"].(string)

	// For now, do basic validation
	// In production, you'd use a proper JSON schema validator
	result := map[string]interface{}{
		"validated":       true,
		"validation_type": validationType,
	}

	// Check required fields if specified
	if requiredFields, ok := schema["required"].([]interface{}); ok {
		for _, field := range requiredFields {
			fieldName, _ := field.(string)
			if _, exists := params.CollectedData[fieldName]; !exists {
				return nil, fmt.Errorf("required field missing: %s", fieldName)
			}
		}
	}

	return result, nil
}

// RetrieveMemoryAction retrieves relevant memories for context
func RetrieveMemoryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	retrievalCount := 5
	if count, ok := config["retrieval_count"].(float64); ok {
		retrievalCount = int(count)
	}

	// In a real implementation, this would query a vector database
	// For now, return empty memory context
	return map[string]interface{}{
		"memory_context":  []interface{}{},
		"retrieval_count": retrievalCount,
	}, nil
}

// StoreMemoryAction stores data in memory for future retrieval
func StoreMemoryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// In a real implementation, this would store to a vector database
	params.Logger.Info("Storing memory",
		zap.String("agent_type", params.AgentType),
		zap.Int("data_size", len(params.CollectedData)))

	return map[string]interface{}{
		"stored":    true,
		"memory_id": "mem_" + params.Headers["correlation_id"],
	}, nil
}

// ValidateAssetsAction checks if required assets are present
func ValidateAssetsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	requiredAssets, ok := config["required_assets"].([]interface{})
	if !ok {
		return map[string]interface{}{"validated": true}, nil
	}

	missingAssets := []string{}
	for _, asset := range requiredAssets {
		assetName, _ := asset.(string)
		if _, exists := params.CollectedData[assetName]; !exists {
			missingAssets = append(missingAssets, assetName)
		}
	}

	if len(missingAssets) > 0 {
		return nil, fmt.Errorf("missing required assets: %v", missingAssets)
	}

	return map[string]interface{}{
		"validated": true,
		"assets":    requiredAssets,
	}, nil
}

// DeployToHostingAction simulates deployment to a hosting platform
func DeployToHostingAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	platform, _ := config["platform"].(string)
	autoSSL, _ := config["auto_ssl"].(bool)

	// In a real implementation, this would deploy to actual hosting
	params.Logger.Info("Deploying to hosting platform",
		zap.String("platform", platform),
		zap.Bool("auto_ssl", autoSSL))

	// Simulate deployment
	return map[string]interface{}{
		"deployed": true,
		"platform": platform,
		"url": fmt.Sprintf("https://%s.%s.app",
			params.Headers["correlation_id"][:8], platform),
		"ssl_enabled":   autoSSL,
		"deployment_id": "deploy_" + params.Headers["correlation_id"][:8],
	}, nil
}

// HTTPRequestAction makes HTTP requests to external services
func HTTPRequestAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	url, _ := config["url"].(string)
	method, _ := config["method"].(string)

	if url == "" {
		return nil, fmt.Errorf("url is required for http_request action")
	}

	if method == "" {
		method = "GET"
	}

	// In a real implementation, make actual HTTP request
	// For now, return mock response
	return map[string]interface{}{
		"status": 200,
		"body": map[string]interface{}{
			"success": true,
			"data":    "mock response",
		},
		"headers": map[string]string{
			"content-type": "application/json",
		},
	}, nil
}

// ConditionalBranchAction provides if/then/else logic in workflows
func ConditionalBranchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	condition, _ := config["condition"].(map[string]interface{})
	if condition == nil {
		return nil, fmt.Errorf("condition is required for conditional_branch")
	}

	// Evaluate condition (simplified)
	field, _ := condition["field"].(string)
	operator, _ := condition["operator"].(string)
	value := condition["value"]

	fieldValue := params.CollectedData[field]

	var conditionMet bool
	switch operator {
	case "equals":
		conditionMet = fieldValue == value
	case "not_equals":
		conditionMet = fieldValue != value
	case "exists":
		_, conditionMet = params.CollectedData[field]
	case "not_exists":
		_, exists := params.CollectedData[field]
		conditionMet = !exists
	default:
		conditionMet = false
	}

	// Determine next step based on condition
	var nextStep string
	if conditionMet {
		nextStep, _ = config["then_step"].(string)
	} else {
		nextStep, _ = config["else_step"].(string)
	}

	return map[string]interface{}{
		"condition_met":      conditionMet,
		"next_step_override": nextStep,
	}, nil
}

// AggregateDataAction combines data from multiple sources
/** how to use it - this might be out of date:
{
  "aggregate_results": {
    "action": "aggregate_data",
    "config": {
      "strategy": "group_responses",  // or "combine_artifacts" for HTML
      "index_by": "operation"  // optional: how to label results
    },
    "next_step": "complete"
  }
}
*/
func AggregateDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("AggregateDataAction in generic_actions")

	config := params.StepConfig.Config

	// Get aggregation strategy
	strategy, _ := config["strategy"].(string)
	if strategy == "" {
		strategy = "group_responses"
	}

	responseFields, _ := config["response_fields"].([]interface{})

	params.Logger.Info("AggregateDataAction strategy is",
		zap.String("strategy", strategy),
		zap.Any("params look whats in collected data", params),
		zap.Any("response fields are:", responseFields),
	)

	//aggregated := make(map[string]interface{})
	results := make([]interface{}, 0)
	responses := make(map[string]interface{})

	// Collect responses from specified fields
	for _, field := range responseFields {
		fieldName, ok := field.(string)
		if !ok {
			continue
		}

		params.Logger.Info("Looking for response in field",
			zap.String("field_name", fieldName))

		// Check if this step has a response
		if stepData, ok := params.CollectedData[fieldName].(map[string]interface{}); ok {
			if response, ok := stepData["response"].(map[string]interface{}); ok {
				responses[fieldName] = response
				results = append(results, response)

				params.Logger.Info("Found response for field",
					zap.String("field_name", fieldName),
					zap.Any("response", response))
			} else {
				params.Logger.Warn("No response found in step data",
					zap.String("field_name", fieldName),
					zap.Any("step_data", stepData))
			}
		} else {
			params.Logger.Warn("Field not found or not a map",
				zap.String("field_name", fieldName),
				zap.Any("value", params.CollectedData[fieldName]))
		}
	}

	// Build aggregated result based on strategy
	aggregated := map[string]interface{}{
		"timestamp":            time.Now(),
		"aggregation_strategy": strategy,
		"count":                len(results),
		"responses":            responses,
		"results_array":        results,
	}

	if strategy == "group_responses" {
		// Group by operation if they're calculations
		byOperation := make(map[string][]interface{})
		for _, result := range results {
			if calc, ok := result.(map[string]interface{}); ok {
				if op, ok := calc["operation"].(string); ok {
					byOperation[op] = append(byOperation[op], calc)
				}
			}
		}
		aggregated["by_operation"] = byOperation
	}

	switch strategy {
	case "group_responses":

		params.Logger.Info("AggregateDataAction in group_responses")

		// Group all response_* keys by type
		responses := make(map[string]interface{})
		operations := make(map[string]interface{})
		results := make([]interface{}, 0)

		// NEW: Look for responses in two places:
		// 1. Steps that have responses embedded
		// 2. Traditional response_ prefixed keys (backward compatibility)

		for key, value := range params.CollectedData {
			// Check if this is a step with an embedded response
			if stepMap, ok := value.(map[string]interface{}); ok {
				if response, hasResponse := stepMap["response"]; hasResponse {
					// This step has a response
					respMap, isMap := response.(map[string]interface{})
					if isMap {
						// Store by step name for clarity
						responses[key] = respMap

						// Group by operation type if present
						if op, hasOp := respMap["operation"]; hasOp {
							operations[fmt.Sprintf("%v", op)] = respMap
						}

						// Add to sequential results
						results = append(results, respMap)

						params.Logger.Info("Found response in step",
							zap.String("step_name", key),
							zap.Any("response", respMap))
					}
				}
			}

			// Also check traditional response_ prefix (backward compatibility)
			if strings.HasPrefix(key, "response_") {
				// Extract meaningful identifier from response key
				respID := strings.TrimPrefix(key, "response_")

				// Try to categorize the response
				if respMap, ok := value.(map[string]interface{}); ok {
					// Only add if not already found in step data
					if _, alreadyProcessed := responses[respID]; !alreadyProcessed {
						// Group by operation type if present
						if op, hasOp := respMap["operation"]; hasOp {
							operations[fmt.Sprintf("%v", op)] = respMap
						}

						// Add to sequential results
						results = append(results, respMap)

						// Store by ID
						responses[respID] = respMap
					}
				}
			}
		}

		aggregated["responses"] = responses
		aggregated["by_operation"] = operations
		aggregated["results_array"] = results
		aggregated["count"] = len(results)

	case "merge_outputs":

		params.Logger.Info("AggregateDataAction in merge_outputs")

		// Merge specific output fields from multiple steps
		outputFields, _ := config["output_fields"].([]interface{})
		merged := make(map[string]interface{})

		for _, field := range outputFields {
			fieldName, _ := field.(string)
			if data, exists := params.CollectedData[fieldName]; exists {
				// Check if this field has an embedded response
				if stepMap, ok := data.(map[string]interface{}); ok {
					if response, hasResponse := stepMap["response"]; hasResponse {
						merged[fieldName] = response
					} else {
						merged[fieldName] = data
					}
				} else {
					merged[fieldName] = data
				}
			}
		}
		aggregated["merged"] = merged

	case "combine_artifacts":

		params.Logger.Info("AggregateDataAction in combine_artifacts")

		// For combining generated artifacts (HTML, code, etc)
		artifacts := make(map[string]interface{})
		metadata := make(map[string]interface{})

		for key, value := range params.CollectedData {
			// Check for responses in step data
			if stepMap, ok := value.(map[string]interface{}); ok {
				if response, hasResponse := stepMap["response"]; hasResponse {
					if respMap, ok := response.(map[string]interface{}); ok {
						// Extract artifact type if present
						if artifactType, hasType := respMap["type"]; hasType {
							artifacts[fmt.Sprintf("%v", artifactType)] = respMap
						} else {
							artifacts[key] = respMap
						}

						// Extract metadata
						if meta, hasMeta := respMap["metadata"]; hasMeta {
							metadata[key] = meta
						}
					}
				}
			}

			// Also check traditional response_ or result_ patterns
			if strings.Contains(key, "response_") || strings.Contains(key, "result_") {
				if respMap, ok := value.(map[string]interface{}); ok {
					// Extract artifact type if present
					if artifactType, hasType := respMap["type"]; hasType {
						artifacts[fmt.Sprintf("%v", artifactType)] = respMap
					} else {
						artifacts[key] = value
					}

					// Extract metadata
					if meta, hasMeta := respMap["metadata"]; hasMeta {
						metadata[key] = meta
					}
				}
			}
		}

		aggregated["artifacts"] = artifacts
		aggregated["metadata"] = metadata

	case "indexed":

		params.Logger.Info("AggregateDataAction in indexed")

		// Create an indexed structure with semantic names
		indexedResults := make([]map[string]interface{}, 0)

		// Get index configuration
		indexBy, _ := config["index_by"].(string)
		if indexBy == "" {
			indexBy = "sequence"
		}

		idx := 1

		// First index responses from steps
		for key, value := range params.CollectedData {
			if stepMap, ok := value.(map[string]interface{}); ok {
				if response, hasResponse := stepMap["response"]; hasResponse {
					entry := map[string]interface{}{
						"index": idx,
						"key":   key,
						"data":  response,
					}

					// Add semantic label if possible
					if respMap, ok := response.(map[string]interface{}); ok {
						if label, hasLabel := respMap[indexBy]; hasLabel {
							entry["label"] = label
						}
					}

					indexedResults = append(indexedResults, entry)
					idx++
				}
			}
		}

		// Also index traditional response_ keys
		for key, value := range params.CollectedData {
			if strings.HasPrefix(key, "response_") {
				entry := map[string]interface{}{
					"index": idx,
					"key":   key,
					"data":  value,
				}

				// Add semantic label if possible
				if respMap, ok := value.(map[string]interface{}); ok {
					if label, hasLabel := respMap[indexBy]; hasLabel {
						entry["label"] = label
					}
				}

				indexedResults = append(indexedResults, entry)
				idx++
			}
		}

		aggregated["indexed_results"] = indexedResults

	case "flatten":

		params.Logger.Info("AggregateDataAction in flatten")

		// Flatten all response data into a single level
		for key, value := range params.CollectedData {
			// Check for responses in step data first
			if stepMap, ok := value.(map[string]interface{}); ok {
				if response, hasResponse := stepMap["response"]; hasResponse {
					if respMap, ok := response.(map[string]interface{}); ok {
						for k, v := range respMap {
							// Create unique key using step name
							flatKey := fmt.Sprintf("%s_%s", key, k)
							aggregated[flatKey] = v
						}
					}
				}
			}

			// Also handle traditional response_ prefix
			if strings.HasPrefix(key, "response_") {
				if respMap, ok := value.(map[string]interface{}); ok {
					for k, v := range respMap {
						// Create unique key
						flatKey := fmt.Sprintf("%s_%s", strings.TrimPrefix(key, "response_"), k)
						aggregated[flatKey] = v
					}
				}
			}
		}

	default:
		// Default: semantic grouping
		outputs := make([]interface{}, 0)

		// Collect from step responses
		for _, value := range params.CollectedData {
			if stepMap, ok := value.(map[string]interface{}); ok {
				if response, hasResponse := stepMap["response"]; hasResponse {
					outputs = append(outputs, response)
				}
			}
		}

		// Also collect traditional response_ keys
		for key, value := range params.CollectedData {
			if strings.HasPrefix(key, "response_") {
				outputs = append(outputs, value)
			}
		}

		aggregated["outputs"] = outputs
	}

	// Add summary information
	aggregated["aggregation_strategy"] = strategy
	aggregated["timestamp"] = time.Now()

	params.Logger.Info("AggregateDataAction aggregated is",
		zap.Any("aggregated is:", aggregated))

	return aggregated, nil
}

// CacheLookupAction checks cache for existing results
func CacheLookupAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	cacheKey, _ := config["cache_key"].(string)
	if cacheKey == "" {
		// Generate cache key from inputs
		inputData, _ := json.Marshal(params.CollectedData["input_data"])
		cacheKey = fmt.Sprintf("%s_%x", params.AgentType, inputData)
	}

	// In real implementation, check Redis or other cache
	// For now, always return cache miss
	return map[string]interface{}{
		"cache_hit": false,
		"cache_key": cacheKey,
	}, nil
}

// StoreResultAction stores results to database
func StoreResultAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	table, _ := config["table"].(string)
	if table == "" {
		table = "agent_results"
	}

	// In real implementation, store to database
	params.Logger.Info("Storing result to database",
		zap.String("table", table),
		zap.String("correlation_id", params.Headers["correlation_id"]))

	return map[string]interface{}{
		"stored": true,
		"table":  table,
		"id":     params.Headers["correlation_id"],
	}, nil
}
