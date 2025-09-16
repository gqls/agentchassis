// platform/orchestration/actions/calculate_actions.go
package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func CalculateAction(ctx context.Context, params ActionParams) (interface{}, error) {

	params.Logger.Info("Entering CalculateAction",
		zap.Any("action_params", params),
	)

	// Try multiple paths to find the input data
	var operation string
	var operands []interface{}

	params.Logger.Info("Completing workflow CompleteWorkflowAction",
		zap.Any("action params", params),
		zap.Any("action params.CollectedData", params.CollectedData),
	)

	// Path 1: Check input_data.body.input_data.data
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if body, ok := inputData["body"].(map[string]interface{}); ok {
			if innerInputData, ok := body["input_data"].(map[string]interface{}); ok {
				if data, ok := innerInputData["data"].(map[string]interface{}); ok {
					operation, _ = data["operation"].(string)
					operands, _ = data["operands"].([]interface{})

					params.Logger.Info("Extracted data from Path 1",
						zap.String("operation", operation),
						zap.Any("operands", operands),
					)
				}
			}
		}
	}

	// Path 2: Check __raw_message__.body.input_data.data
	if operation == "" {
		if rawMsg, ok := params.CollectedData["__raw_message__"].(map[string]interface{}); ok {
			if body, ok := rawMsg["body"].(map[string]interface{}); ok {
				if inputData, ok := body["input_data"].(map[string]interface{}); ok {
					if data, ok := inputData["data"].(map[string]interface{}); ok {
						operation, _ = data["operation"].(string)
						operands, _ = data["operands"].([]interface{})

						params.Logger.Info("Extracted data from Path 2",
							zap.String("operation", operation),
							zap.Any("operands", operands),
						)
					}
				}
			}
		}
	}

	// Path 3: Direct path for simpler cases
	if operation == "" {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if data, ok := inputData["data"].(map[string]interface{}); ok {
				operation, _ = data["operation"].(string)
				operands, _ = data["operands"].([]interface{})

				params.Logger.Info("Extracted data from Path 3",
					zap.String("operation", operation),
					zap.Any("operands", operands),
				)
			}
		}
	}

	// Perform the calculation
	if operation == "add" && len(operands) == 2 {
		// Convert operands to float64
		a, ok1 := toFloat64(operands[0])
		b, ok2 := toFloat64(operands[1])

		if ok1 && ok2 {
			return map[string]interface{}{
				"result":    a + b,
				"operation": operation,
				"operands":  operands,
			}, nil
		}
	} else if operation == "subtract" && len(operands) == 2 {
		a, ok1 := toFloat64(operands[0])
		b, ok2 := toFloat64(operands[1])

		if ok1 && ok2 {
			result := map[string]interface{}{
				"result":    a + b,
				"operation": operation,
				"operands":  operands,
			}
			// Step 5: Add a log for successful calculation.
			params.Logger.Info("Calculation successful",
				zap.Any("CALCULATION RESULT: addition ", result),
			)
			return result, nil
		}
	} else if operation == "multiply" && len(operands) == 2 {
		a, ok1 := toFloat64(operands[0])
		b, ok2 := toFloat64(operands[1])

		if ok1 && ok2 {
			return map[string]interface{}{
				"result":    a * b,
				"operation": operation,
				"operands":  operands,
			}, nil
		}
	} else if operation == "divide" && len(operands) == 2 {
		a, ok1 := toFloat64(operands[0])
		b, ok2 := toFloat64(operands[1])

		if ok1 && ok2 && b != 0 {
			return map[string]interface{}{
				"result":    a / b,
				"operation": operation,
				"operands":  operands,
			}, nil
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
	default:
		return 0, false
	}
}

func CalculateActionNormalFormat(ctx context.Context, params ActionParams) (interface{}, error) {
	// Extract input_data first
	inputData, ok := params.CollectedData["input_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("input_data not found in CollectedData")
	}

	operation, _ := inputData["operation"].(string)
	operands, _ := inputData["operands"].([]interface{})

	if operation == "add" && len(operands) == 2 {
		a, _ := operands[0].(float64)
		b, _ := operands[1].(float64)
		return map[string]interface{}{
			"result":    a + b,
			"operation": operation,
		}, nil
	}

	return nil, fmt.Errorf("unsupported operation")
}
