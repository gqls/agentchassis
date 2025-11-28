// FILE: platform/orchestration/actions/route_by_field.go
package actions

import (
"context"
"encoding/json"
"fmt"
"strings"

	"go.uber.org/zap"
)

// RouteByFieldAction evaluates a field in collected data and returns the next step
// based on a routing table. This enables conditional workflow branching.
//
// Config:
//   - field_path: dot-notation path to the field to evaluate (e.g., "brief_data.structured_brief.result.site_type")
//   - routes: map of field_value -> next_step_name
//   - default_route: step to use if field value doesn't match any route
//
// Returns:
//   - next_step: the step name to execute next
//   - routed_by: the field value that determined the route
func RouteByFieldAction(ctx context.Context, params ActionParams) (interface{}, error) {
params.Logger.Info("RouteByFieldAction starting",
zap.String("step_name", params.ExecutionContext.StepName),
)

	config := params.StepConfig.Config

	// Get field path
	fieldPath, ok := config["field_path"].(string)
	if !ok || fieldPath == "" {
		return nil, fmt.Errorf("route_by_field requires 'field_path' in config")
	}

	// Get routes map
	routesRaw, ok := config["routes"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("route_by_field requires 'routes' map in config")
	}

	routes := make(map[string]string)
	for k, v := range routesRaw {
		if s, ok := v.(string); ok {
			routes[k] = s
		}
	}

	// Get default route
	defaultRoute, _ := config["default_route"].(string)

	// Extract the field value from collected data
	fieldValue, err := extractFieldValue(params.CollectedData, fieldPath, params.Logger)
	if err != nil {
		params.Logger.Warn("Could not extract field value, using default route",
			zap.String("field_path", fieldPath),
			zap.Error(err),
		)
		if defaultRoute == "" {
			return nil, fmt.Errorf("field not found and no default_route specified: %w", err)
		}
		return map[string]interface{}{
			"next_step":  defaultRoute,
			"routed_by":  nil,
			"route_type": "default",
			"reason":     err.Error(),
		}, nil
	}

	params.Logger.Info("Extracted field value for routing",
		zap.String("field_path", fieldPath),
		zap.String("field_value", fieldValue),
	)

	// Look up the route
	nextStep, found := routes[fieldValue]
	routeType := "matched"
	if !found {
		if defaultRoute == "" {
			return nil, fmt.Errorf("no route found for value '%s' and no default_route specified", fieldValue)
		}
		nextStep = defaultRoute
		routeType = "default"
		params.Logger.Info("Using default route",
			zap.String("field_value", fieldValue),
			zap.String("next_step", nextStep),
		)
	} else {
		params.Logger.Info("Route matched",
			zap.String("field_value", fieldValue),
			zap.String("next_step", nextStep),
		)
	}

	return map[string]interface{}{
		"next_step":   nextStep,
		"routed_by":   fieldValue,
		"route_type":  routeType,
		"field_path":  fieldPath,
		"routes_available": routes,
	}, nil
}

// extractFieldValue navigates a nested map using dot notation and returns the string value
func extractFieldValue(data map[string]interface{}, path string, logger *zap.Logger) (string, error) {
parts := strings.Split(path, ".")
current := interface{}(data)

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return "", fmt.Errorf("path part '%s' not found at position %d", part, i)
			}
			current = val

		case string:
			// The value might be a JSON string that needs parsing
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err != nil {
				return "", fmt.Errorf("cannot navigate into string value at '%s': not valid JSON", part)
			}
			val, exists := parsed[part]
			if !exists {
				return "", fmt.Errorf("path part '%s' not found in parsed JSON at position %d", part, i)
			}
			current = val

		default:
			return "", fmt.Errorf("cannot navigate into type %T at path part '%s'", current, part)
		}
	}

	// Convert final value to string
	switch v := current.(type) {
	case string:
		return v, nil
	case float64:
		return fmt.Sprintf("%v", v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	default:
		return "", fmt.Errorf("final value is not a scalar type: %T", current)
	}
}

// ============================================================================
// REGISTRY UPDATE
// ============================================================================
// Add to GlobalActionRegistry in registry.go:
//
// "route_by_field": RouteByFieldAction,
//



and this one:

// FILE: platform/orchestration/actions/conditional_call_agent.go
package actions

import (
"context"
"encoding/json"
"fmt"
"strings"

	"go.uber.org/zap"
)

// ConditionalCallAgentAction routes to and calls one of several agents based on
// a field value in collected data. This combines routing logic with agent calling
// in a single action, avoiding the need for coordinator changes.
//
// Config:
//   - field_path: dot-notation path to evaluate (e.g., "brief_data.structured_brief.result.site_type")
//   - agent_mapping: map of field_value -> agent_type (e.g., {"landing": "landing-page-architect"})
//   - default_agent: agent_type to use if field value doesn't match
//   - input_fields: fields to pass to the called agent
//   - timeout_seconds: timeout for the agent call
//
// Example config:
//
//	{
//	  "field_path": "brief_data.structured_brief.result.site_type",
//	  "agent_mapping": {
//	    "landing": "landing-page-architect",
//	    "content": "content-site-architect",
//	    "portfolio": "portfolio-architect"
//	  },
//	  "default_agent": "landing-page-architect",
//	  "input_fields": ["build_plan_data", "brief_data", "input_data"],
//	  "timeout_seconds": 120
//	}
func ConditionalCallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
params.Logger.Info("ConditionalCallAgentAction starting",
zap.String("step_name", params.ExecutionContext.StepName),
)

	config := params.StepConfig.Config

	// Get field path to evaluate
	fieldPath, ok := config["field_path"].(string)
	if !ok || fieldPath == "" {
		return nil, fmt.Errorf("conditional_call_agent requires 'field_path' in config")
	}

	// Get agent mapping
	agentMappingRaw, ok := config["agent_mapping"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("conditional_call_agent requires 'agent_mapping' in config")
	}

	agentMapping := make(map[string]string)
	for k, v := range agentMappingRaw {
		if s, ok := v.(string); ok {
			agentMapping[k] = s
		}
	}

	// Get default agent
	defaultAgent, _ := config["default_agent"].(string)

	// Extract the field value
	fieldValue, err := extractFieldValueForRouting(params.CollectedData, fieldPath, params.Logger)
	if err != nil {
		params.Logger.Warn("Could not extract field value, using default agent",
			zap.String("field_path", fieldPath),
			zap.Error(err),
		)
		fieldValue = ""
	}

	params.Logger.Info("Routing based on field value",
		zap.String("field_path", fieldPath),
		zap.String("field_value", fieldValue),
	)

	// Determine target agent
	targetAgentType, found := agentMapping[fieldValue]
	routeType := "matched"
	if !found {
		if defaultAgent == "" {
			return nil, fmt.Errorf("no agent mapping for value '%s' and no default_agent specified", fieldValue)
		}
		targetAgentType = defaultAgent
		routeType = "default"
	}

	params.Logger.Info("Selected target agent",
		zap.String("field_value", fieldValue),
		zap.String("target_agent_type", targetAgentType),
		zap.String("route_type", routeType),
	)

	// Now we need to call the agent
	// Update the step config to use the selected agent type
	modifiedConfig := make(map[string]interface{})
	for k, v := range config {
		modifiedConfig[k] = v
	}
	modifiedConfig["agent_type"] = targetAgentType

	// Create modified params for CallAgentAction
	modifiedStepConfig := params.StepConfig
	modifiedStepConfig.Config = modifiedConfig

	modifiedParams := ActionParams{
		Context:          params.Context,
		ExecutionContext: params.ExecutionContext,
		Headers:          params.Headers,
		StepConfig:       modifiedStepConfig,
		CollectedData:    params.CollectedData,
		Logger:           params.Logger,
		KafkaClient:      params.KafkaClient,
		ResponseChan:     params.ResponseChan,
		DB:               params.DB,
	}

	// Call the CallAgentAction with the selected agent
	result, err := CallAgentAction(ctx, modifiedParams)
	if err != nil {
		return nil, fmt.Errorf("failed to call agent %s: %w", targetAgentType, err)
	}

	// Wrap result with routing metadata
	resultMap := make(map[string]interface{})
	if r, ok := result.(map[string]interface{}); ok {
		resultMap = r
	} else {
		resultMap["result"] = result
	}

	resultMap["_routing"] = map[string]interface{}{
		"field_path":    fieldPath,
		"field_value":   fieldValue,
		"selected_agent": targetAgentType,
		"route_type":    routeType,
	}

	return resultMap, nil
}

// extractFieldValueForRouting navigates nested data using dot notation
// Handles both map[string]interface{} and JSON strings
func extractFieldValueForRouting(data map[string]interface{}, path string, logger *zap.Logger) (string, error) {
parts := strings.Split(path, ".")
current := interface{}(data)

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return "", fmt.Errorf("path part '%s' not found at position %d", part, i)
			}
			current = val

		case string:
			// Value might be a JSON string - try to parse it
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err != nil {
				// Not JSON, can't navigate further
				return "", fmt.Errorf("cannot navigate into non-JSON string at '%s'", part)
			}
			val, exists := parsed[part]
			if !exists {
				return "", fmt.Errorf("path part '%s' not found in parsed JSON at position %d", part, i)
			}
			current = val

		default:
			return "", fmt.Errorf("cannot navigate into type %T at path part '%s'", current, part)
		}
	}

	// Convert final value to string
	switch v := current.(type) {
	case string:
		return v, nil
	case float64:
		return fmt.Sprintf("%v", v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	default:
		return "", fmt.Errorf("final value is type %T, expected scalar", current)
	}
}

// ============================================================================
// REGISTRY UPDATE
// ============================================================================
// Add to GlobalActionRegistry in registry.go:
//
// "conditional_call_agent": ConditionalCallAgentAction,
//


