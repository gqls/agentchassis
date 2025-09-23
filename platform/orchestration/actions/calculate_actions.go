// platform/orchestration/actions/calculate_actions.go
package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func CalculateAction(ctx context.Context, params ActionParams) (interface{}, error) {

	params.Logger.Info("Entering CalculateAction",
		zap.Any("collected_data_keys", GetMapKeys(params.CollectedData)),
	)

	// Try multiple paths to find the input data
	var operation string
	var operands []interface{}

	// Path 1: Direct path - after our response simplification
	// Check input_data.data (most direct path after fixes)
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if data, ok := inputData["data"].(map[string]interface{}); ok {
			operation, _ = data["operation"].(string)
			operands, _ = data["operands"].([]interface{})

			params.Logger.Info("Extracted data from Path 1 (simplified)",
				zap.String("operation", operation),
				zap.Any("operands", operands))
		}

		// Also check directly in input_data for even simpler structure
		if operation == "" {
			operation, _ = inputData["operation"].(string)
			operands, _ = inputData["operands"].([]interface{})

			if operation != "" {
				params.Logger.Info("Extracted data directly from input_data",
					zap.String("operation", operation),
					zap.Any("operands", operands))
			}
		}
	}

	// Path 2: Legacy nested structure (for backward compatibility during transition)
	if operation == "" {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if body, ok := inputData["body"].(map[string]interface{}); ok {
				// Check body.data
				if data, ok := body["data"].(map[string]interface{}); ok {
					operation, _ = data["operation"].(string)
					operands, _ = data["operands"].([]interface{})

					params.Logger.Info("Extracted from legacy body.data",
						zap.String("operation", operation),
						zap.Any("operands", operands))
				}

				// Check body.input_data.data (double nested legacy)
				if operation == "" {
					if innerInputData, ok := body["input_data"].(map[string]interface{}); ok {
						if data, ok := innerInputData["data"].(map[string]interface{}); ok {
							operation, _ = data["operation"].(string)
							operands, _ = data["operands"].([]interface{})

							params.Logger.Info("Extracted from legacy nested structure",
								zap.String("operation", operation),
								zap.Any("operands", operands))
						}
					}
				}
			}
		}
	}

	// Path 3: Check raw message as last resort
	if operation == "" {
		if rawMsg, ok := params.CollectedData["__raw_message__"].(map[string]interface{}); ok {
			// Try raw message body
			if body, ok := rawMsg["body"].(map[string]interface{}); ok {
				if data, ok := body["data"].(map[string]interface{}); ok {
					operation, _ = data["operation"].(string)
					operands, _ = data["operands"].([]interface{})

					params.Logger.Info("Extracted from raw message",
						zap.String("operation", operation),
						zap.Any("operands", operands))
				}
			}

			// Try raw message data directly
			if operation == "" {
				if data, ok := rawMsg["data"].(map[string]interface{}); ok {
					operation, _ = data["operation"].(string)
					operands, _ = data["operands"].([]interface{})

					params.Logger.Info("Extracted from raw message data",
						zap.String("operation", operation),
						zap.Any("operands", operands))
				}
			}
		}
	}

	// Validate we found the required data
	if operation == "" || len(operands) == 0 {
		params.Logger.Error("Failed to extract operation data",
			zap.String("operation", operation),
			zap.Int("operands_count", len(operands)),
			zap.Any("collected_data", params.CollectedData))
		return nil, fmt.Errorf("missing operation or operands")
	}

	// Perform the calculation
	switch operation {
	case "add":
		if len(operands) == 2 {
			a, ok1 := toFloat64(operands[0])
			b, ok2 := toFloat64(operands[1])

			if ok1 && ok2 {
				result := map[string]interface{}{
					"result":    a + b,
					"operation": operation,
					"operands":  operands,
				}
				params.Logger.Info("Addition successful",
					zap.Any("result from successful addition", result))
				return result, nil
			}
		}

	case "subtract":
		if len(operands) == 2 {
			a, ok1 := toFloat64(operands[0])
			b, ok2 := toFloat64(operands[1])

			if ok1 && ok2 {
				result := map[string]interface{}{
					"result":    a - b,
					"operation": operation,
					"operands":  operands,
				}
				params.Logger.Info("Subtraction successful",
					zap.Any("result from successful subtraction", result))
				return result, nil
			}
		}

	case "multiply":
		if len(operands) == 2 {
			a, ok1 := toFloat64(operands[0])
			b, ok2 := toFloat64(operands[1])

			if ok1 && ok2 {
				result := map[string]interface{}{
					"result":    a * b,
					"operation": operation,
					"operands":  operands,
				}
				params.Logger.Info("Multiplication successful",
					zap.Any("result from successful multiplication", result))
				return result, nil
			}
		}

	case "divide":
		if len(operands) == 2 {
			a, ok1 := toFloat64(operands[0])
			b, ok2 := toFloat64(operands[1])

			if ok1 && ok2 && b != 0 {
				result := map[string]interface{}{
					"result":    a / b,
					"operation": operation,
					"operands":  operands,
				}
				params.Logger.Info("Division successful",
					zap.Any("result from successful division", result))
				return result, nil
			}
			if b == 0 {
				return nil, fmt.Errorf("division by zero")
			}
		}
	}

	err := fmt.Errorf("unsupported operation: '%s' with operands: %v", operation, operands)
	params.Logger.Error("CalculateAction failed", zap.Error(err))
	return nil, err
}

// Helper function to convert interface{} to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case float32:
		return float64(val), true
	default:
		return 0, false
	}
}
