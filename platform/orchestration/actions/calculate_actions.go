// platform/orchestration/actions/calculate_actions.go
package actions

import (
	"context"
	"fmt"
	"math"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func CalculateAction(ctx context.Context, params ActionParams) (interface{}, error) {

	params.Logger.Info("Entering CalculateAction",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
		zap.String("DEBUGaa: in CalculateAction action", params.ExecutionContext.Action),
		//zap.Any("DEBUGaa: in CalculateAction CollectedData", params.CollectedData),
	)

	// Check if this is an initialization call
	// Method 1: Check ExecutionContext action
	if params.ExecutionContext.Action == "initialize" {
		params.Logger.Info("CalculateAction handling initialization")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Method 2: Check for initialization flag in collected data
	if isInit, ok := params.CollectedData["is_initialization"].(bool); ok && isInit {
		params.Logger.Info("CalculateAction initialization detected via flag")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Method 3: Check the action from collected data (if passed through)
	if action, ok := params.CollectedData["action"].(string); ok && action == "initialize" {
		params.Logger.Info("CalculateAction initialization detected via action field")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Try multiple paths to find the input data
	var operation string
	var operands []interface{}

	// Path 1.0: from where we put it in callagentaction in executeLocalAction
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if body, ok := inputData["body"].(map[string]interface{}); ok {
			if data, ok := body["input_data"].(map[string]interface{}); ok {
				operation, _ = data["operation"].(string)
				operands, _ = data["operands"].([]interface{})

				params.Logger.Info("Extracted data from Path 1 (simplified)",
					zap.String("operation", operation),
					zap.Any("operands", operands))
			}
		}
	}

	// Path 1.1: This logic now correctly targets the message from "CallAgentAction"
	if operation == "" {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if op, ok := inputData["operation"].(string); ok {
				if opr, ok := inputData["operands"].([]interface{}); ok {
					operation = op
					operands = opr

					params.Logger.Info("Extracted calculation data direct from input_data[operation] etc",
						zap.String("operation", operation),
						zap.Any("operands", operands))
				}
			}
		}
	}

	// Path 1.2: Direct path - after our response simplification
	// Check input_data.data (most direct path after fixes)
	if operation == "" {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if data, ok := inputData["data"].(map[string]interface{}); ok {
				operation, _ = data["operation"].(string)
				operands, _ = data["operands"].([]interface{})

				params.Logger.Info("Extracted data from Path 1.2 - [input_data][data]",
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

	case "modulo":
		if len(operands) == 2 {
			// Modulo requires integer operands in Go
			a, ok1 := toInt64(operands[0])
			b, ok2 := toInt64(operands[1])

			if ok1 && ok2 {
				if b == 0 {
					return nil, fmt.Errorf("division by zero for modulo")
				}
				result := map[string]interface{}{
					"result":    a % b, // Use the % operator
					"operation": operation,
					"operands":  operands,
				}
				params.Logger.Info("Modulo successful",
					zap.Any("result from successful modulo", result))
				return result, nil
			}
		}

	case "power":
		if len(operands) == 2 {
			base, ok1 := toFloat64(operands[0])
			exponent, ok2 := toFloat64(operands[1])

			if ok1 && ok2 {
				result := map[string]interface{}{
					"result":    math.Pow(base, exponent), // Use math.Pow
					"operation": operation,
					"operands":  operands,
				}
				params.Logger.Info("Power calculation successful",
					zap.Any("result from successful power calc", result))
				return result, nil
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

func toInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int:
		return int64(val), true
	case int64:
		return val, true
	case int32:
		return int64(val), true
	case float32:
		return int64(val), true
	default:
		return 0, false
	}
}
