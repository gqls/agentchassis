package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test RequestMessage marshaling/unmarshaling
func TestRequestMessage_MarshalUnmarshal(t *testing.T) {
	original := &types.RequestMessage{
		Headers: types.RequestHeaders{
			Sender: types.AgentIdentity{
				AgentType: "calculator",
				AgentID:   "calc-123",
				PodName:   "calc-pod-1",
				Role:      "compute",
			},
			CorrelationID:     "corr-123",
			OrchestrationID:   "orch-456",
			OrchestrationName: "dual-calc",
			MessageID:         "msg-789",
			MessageType:       "request",
			Action:            "calculate",
			Timestamp:         time.Now(),
		},
		Body: map[string]interface{}{
			"action": "calculate",
			"input_data": map[string]interface{}{
				"value1": 10,
				"value2": 20,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var parsed types.RequestMessage
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Verify key fields
	assert.Equal(t, original.Headers.CorrelationID, parsed.Headers.CorrelationID)
	assert.Equal(t, original.Headers.OrchestrationID, parsed.Headers.OrchestrationID)
	assert.Equal(t, original.Headers.Sender.AgentType, parsed.Headers.Sender.AgentType)
	assert.Equal(t, original.Headers.Action, parsed.Headers.Action)
}

// Test ResponseMessage marshaling/unmarshaling
func TestResponseMessage_MarshalUnmarshal(t *testing.T) {
	original := &types.ResponseMessage{
		Headers: types.ResponseHeaders{
			Sender: types.AgentIdentity{
				AgentType: "calculator",
				AgentID:   "calc-123",
				PodName:   "calc-pod-1",
			},

			// Where to route this response
			OrchestrationID:   "parent-orch-123",
			OrchestrationName: "parent-workflow",

			// My identity
			MyOrchestrationID:   "child-orch-456",
			MyOrchestrationName: "calc-workflow",

			// Response tracking
			InResponseToRequestID: "req-789",
			InResponseToStepID:    "step-1",
			InResponseToAction:    "calculate",

			CorrelationID: "corr-123",
			ClientID:      "demo_client",
			MessageType:   "response",
			Status:        "complete",
			IsComplete:    true,
			TimeSent:      time.Now(),
		},
		Body: types.ResponseBody{
			Success: true,
			Body: map[string]interface{}{
				"result":      30,
				"calculation": "10 + 20",
			},
			Error: nil,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var parsed types.ResponseMessage
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Verify routing fields
	assert.Equal(t, original.Headers.OrchestrationID, parsed.Headers.OrchestrationID)
	assert.Equal(t, original.Headers.MyOrchestrationID, parsed.Headers.MyOrchestrationID)
	assert.Equal(t, original.Headers.InResponseToRequestID, parsed.Headers.InResponseToRequestID)
	assert.Equal(t, original.Headers.Status, parsed.Headers.Status)
	assert.True(t, parsed.Headers.IsComplete)
}

// Test error response structure
func TestResponseMessage_ErrorFormat(t *testing.T) {
	errorResponse := &types.ResponseMessage{
		Headers: types.ResponseHeaders{
			OrchestrationID:   "", // This was the problem!
			MyOrchestrationID: "my-orch-123",
			Status:            "error_unrecoverable",
			IsError:           true,
			IsComplete:        false,
		},
		Body: types.ResponseBody{
			Success: false,
			Body:    nil,
			Error: &types.ErrorInfo{
				Code:        "WORKFLOW_INVALID",
				Message:     "Invalid workflow configuration",
				Recoverable: false,
			},
		},
	}

	data, err := json.Marshal(errorResponse)
	require.NoError(t, err)

	var parsed types.ResponseMessage
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// This should demonstrate the routing issue
	assert.Empty(t, parsed.Headers.OrchestrationID, "This is the problem - no routing target")
	assert.NotEmpty(t, parsed.Headers.MyOrchestrationID)
	assert.True(t, parsed.Headers.IsError)
	assert.Equal(t, "WORKFLOW_INVALID", parsed.Body.Error.Code)
}

// Test FromHeaders parsing
func TestFromHeaders_RequestMessage(t *testing.T) {
	headers := map[string]string{
		"correlation_id":     "corr-123",
		"orchestration_id":   "orch-456",
		"orchestration_name": "test-workflow",
		"message_id":         "msg-789",
		"message_type":       "request",
		"client_id":          "demo_client",
		"action":             "calculate",
		"sender_agent_type":  "orchestrator",
		"sender_agent_id":    "orch-001",
		"sender_pod_name":    "orch-pod-1",
	}

	execCtx, err := types.FromHeaders(headers)
	require.NoError(t, err)

	assert.Equal(t, "corr-123", execCtx.CorrelationID)
	assert.Equal(t, "orch-456", execCtx.OrchestrationID)
	assert.Equal(t, "request", execCtx.MessageType)
	assert.Equal(t, "calculate", execCtx.Action)
	assert.Equal(t, "orchestrator", execCtx.Sender.AgentType)
}

// Test FromHeaders parsing response
func TestFromHeaders_ResponseMessage(t *testing.T) {
	headers := map[string]string{
		"correlation_id":                         "corr-123",
		"orchestration_id":                       "", // Empty!
		"my_orchestration_id":                    "calc-orch-456",
		"message_type":                           "response",
		"client_id":                              "demo_client",
		"in_response_to_request_id":              "req-789",
		"in_response_to_parent_orchestration_id": "parent-orch-123",
		"status":                                 "complete",
		"is_complete":                            "true",
		"is_error":                               "false",
	}

	execCtx, err := types.FromHeaders(headers)
	require.NoError(t, err)

	assert.Empty(t, execCtx.OrchestrationID, "This demonstrates the issue")
	assert.Equal(t, "response", execCtx.MessageType)
	assert.NotNil(t, execCtx.InResponseTo)
	assert.Equal(t, "req-789", execCtx.InResponseTo.RequestID)
	assert.Equal(t, "parent-orch-123", execCtx.InResponseTo.ParentOrchestrationID)
}

// Test ToResponseHeaders with proper routing
func TestToResponseHeaders_ProperRouting(t *testing.T) {
	// Child context responding to parent
	childCtx := &types.ExecutionContext{
		CorrelationID:     "corr-123",
		OrchestrationID:   "child-orch-456",
		OrchestrationName: "calc-workflow",

		// Parent info
		ParentOrchestrationID:   "parent-orch-123",
		ParentOrchestrationName: "main-workflow",

		ClientID: "demo_client",
		Status:   "complete",
	}

	headers := childCtx.ToResponseHeaders()

	// The key assertion - response should route to parent
	assert.Equal(t, "parent-orch-123", headers.OrchestrationID, "Response should route to parent")
	assert.Equal(t, "child-orch-456", headers.MyOrchestrationID, "Should identify sender")
}

// Test response routing determination
func TestDetermineResponseOrchestrationTarget(t *testing.T) {
	tests := []struct {
		name       string
		context    *types.ExecutionContext
		wantOrchID string
		wantName   string
	}{
		{
			name: "child_to_parent",
			context: &types.ExecutionContext{
				OrchestrationID:         "child-123",
				ParentOrchestrationID:   "parent-456",
				ParentOrchestrationName: "parent-workflow",
			},
			wantOrchID: "parent-456",
			wantName:   "parent-workflow",
		},
		{
			name: "response_with_parent_info",
			context: &types.ExecutionContext{
				OrchestrationID: "responder-123",
				InResponseTo: &types.ResponseContext{
					ParentOrchestrationID:   "requester-456",
					ParentOrchestrationName: "requester-workflow",
				},
			},
			wantOrchID: "requester-456",
			wantName:   "requester-workflow",
		},
		{
			name: "no_parent_info",
			context: &types.ExecutionContext{
				OrchestrationID: "standalone-123",
			},
			wantOrchID: "",
			wantName:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotName := tt.context.DetermineResponseOrchestrationTarget()
			assert.Equal(t, tt.wantOrchID, gotID)
			assert.Equal(t, tt.wantName, gotName)
		})
	}
}

// Test the full flow: request -> process -> response
func TestMessageFlow_Integration(t *testing.T) {
	// 1. Parent creates request to child
	parentRequest := &types.RequestMessage{
		Headers: types.RequestHeaders{
			OrchestrationID:   "parent-orch-123",
			OrchestrationName: "parent-workflow",
			ResponsesTopic:    "system.agent.orchestrator.responses",
			Action:            "calculate",
		},
		Body: map[string]interface{}{
			"data": "test",
		},
	}

	// 2. Child receives and processes, creating its context
	childCtx := &types.ExecutionContext{
		OrchestrationID:         "child-orch-456",
		OrchestrationName:       "calc-workflow",
		ParentOrchestrationID:   parentRequest.Headers.OrchestrationID,
		ParentOrchestrationName: parentRequest.Headers.OrchestrationName,
		ResponsesTopic:          parentRequest.Headers.ResponsesTopic,
	}

	// 3. Child creates response
	responseHeaders := childCtx.ToResponseHeaders()

	// 4. Verify response routes back to parent
	assert.Equal(t, "parent-orch-123", responseHeaders.OrchestrationID)
	assert.Equal(t, "child-orch-456", responseHeaders.MyOrchestrationID)
	assert.Equal(t, "system.agent.orchestrator.responses", childCtx.ResponsesTopic)
}
