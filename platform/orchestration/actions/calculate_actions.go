// platform/orchestration/actions/calculate_actions.go
package actions

import (
	"context"
	"fmt"
)

func CalculateAction(ctx context.Context, params ActionParams) (interface{}, error) {
	data := params.CollectedData

	operation, _ := data["operation"].(string)
	operands, _ := data["operands"].([]interface{})

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
