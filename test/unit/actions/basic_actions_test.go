// test/unit/actions/basic_actions_test.go
package actions

import (
	"context"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateInputAction(t *testing.T) {
	tests := []struct {
		name      string
		inputData []byte
		wantErr   bool
	}{
		{
			name: "Valid input",
			inputData: []byte(`{
				"action": "test",
				"data": {
					"message": "test message"
				}
			}`),
			wantErr: false,
		},
		{
			name:      "No input data",
			inputData: nil,
			wantErr:   true,
		},
		{
			name:      "Invalid JSON",
			inputData: []byte(`{invalid json}`),
			wantErr:   true,
		},
		{
			name: "Missing data field",
			inputData: []byte(`{
				"action": "test"
			}`),
			wantErr: true,
		},
		{
			name: "Missing message in data",
			inputData: []byte(`{
				"action": "test",
				"data": {}
			}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := actions.ActionParams{
				Context:   context.Background(),
				InputData: tt.inputData,
			}

			result, err := actions.ValidateInputAction(context.Background(), params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)

				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.True(t, resultMap["validated"].(bool))
				assert.NotNil(t, resultMap["input"])
			}
		})
	}
}

func TestTransformDataAction(t *testing.T) {
	// Setup collected data from previous step
	collectedData := map[string]interface{}{
		"validate_input": map[string]interface{}{
			"validated": true,
			"input": map[string]interface{}{
				"message": "hello",
			},
		},
	}

	params := actions.ActionParams{
		Context: context.Background(),
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"transformation": "uppercase",
			},
		},
		CollectedData: collectedData,
	}

	result, err := actions.TransformDataAction(context.Background(), params)

	require.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "HELLO", resultMap["message_transformed"])
}
