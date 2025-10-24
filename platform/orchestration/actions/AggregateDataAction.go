package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

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
