# HITL API Endpoint - Complete Implementation Guide

## For Future Implementation (when time permits)

This is a complete working example you can use later.

---

## 1. Create Handler File

**File:** `internal/gateway/hitl_handler.go`

```go
package gateway

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

type HITLHandler struct {
	producer kafka.Producer
	logger   *zap.Logger
}

func NewHITLHandler(producer kafka.Producer, logger *zap.Logger) *HITLHandler {
	return &HITLHandler{
		producer: producer,
		logger:   logger,
	}
}

// HITLResponseRequest for submitting human input
type HITLResponseRequest struct {
	// Required fields
	CorrelationID   string                 `json:"correlation_id" binding:"required" example:"c9694bbb-2f4e-4791-a07f-4a6bb1047aa8"`
	OrchestrationID string                 `json:"orchestration_id" binding:"required" example:"6e22f90a-55a4-4976-8678-069cdb4b0be7"`
	RequestID       string                 `json:"request_id" binding:"required" example:"dcd1b4c6-1224-4adc-b112-903f99353d03"`
	StepName        string                 `json:"step_name" binding:"required" example:"hitl_review_brief"`
	
	// Response data
	Data            map[string]interface{} `json:"data" binding:"required"`
	
	// Optional fields
	ClientID        string                 `json:"client_id,omitempty" example:"demo_client"`
	RespondedBy     string                 `json:"responded_by,omitempty" example:"user@example.com"`
}

// @Summary Submit HITL Response
// @Description Submit a human-in-the-loop response to continue a paused workflow
// @Tags HITL
// @Accept json
// @Produce json
// @Param request body HITLResponseRequest true "HITL Response"
// @Success 200 {object} map[string]interface{} "Response submitted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Failed to send response"
// @Router /api/v1/hitl/respond [post]
func (h *HITLHandler) SubmitHITLResponse(c *gin.Context) {
	var req HITLResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid HITL response request", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Generate message ID if not provided
	messageID := uuid.New().String()
	timestamp := time.Now().UTC()

	// Default client ID
	clientID := req.ClientID
	if clientID == "" {
		clientID = "demo_client"
	}

	// Build response message
	message := types.ResponseMessage{
		Headers: types.MessageHeaders{
			CorrelationID:           req.CorrelationID,
			OrchestrationID:         req.OrchestrationID,
			MessageID:               messageID,
			MessageType:             "response",
			ClientID:                clientID,
			InResponseToRequestID:   req.RequestID,
			InResponseToStepName:    req.StepName,
			InResponseToAction:      "request_human_input",
			Status:                  "complete",
			IsComplete:              true,
			IsError:                 false,
			Sender: types.AgentInfo{
				AgentID:   "api-user",
				AgentType: "human",
				PodName:   "api",
			},
			Timestamp: timestamp,
		},
		Body: types.MessageBody{
			Success: true,
			Body:    req.Data,
		},
	}

	// Marshal message
	messageBytes, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal HITL response", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to marshal response"})
		return
	}

	// Build Kafka headers
	headers := map[string]string{
		"correlation_id":           req.CorrelationID,
		"orchestration_id":         req.OrchestrationID,
		"message_id":               messageID,
		"message_type":             "response",
		"client_id":                clientID,
		"in_response_to_request_id": req.RequestID,
		"in_response_to_step_name": req.StepName,
		"status":                   "complete",
		"sender_agent_type":        "human",
		"sender_agent_id":          "api-user",
		"timestamp":                timestamp.Format(time.RFC3339),
	}

	// Send to Kafka
	ctx := context.Background()
	topic := "system.agent.generic.responses"
	key := []byte(req.CorrelationID)

	err = h.producer.Produce(ctx, topic, headers, key, messageBytes)
	if err != nil {
		h.logger.Error("Failed to send HITL response to Kafka",
			zap.Error(err),
			zap.String("request_id", req.RequestID),
			zap.String("correlation_id", req.CorrelationID),
		)
		c.JSON(500, gin.H{"error": "Failed to send response"})
		return
	}

	h.logger.Info("HITL response submitted successfully",
		zap.String("request_id", req.RequestID),
		zap.String("correlation_id", req.CorrelationID),
		zap.String("orchestration_id", req.OrchestrationID),
		zap.String("step_name", req.StepName),
	)

	c.JSON(200, gin.H{
		"status":      "submitted",
		"message_id":  messageID,
		"request_id":  req.RequestID,
		"timestamp":   timestamp.Format(time.RFC3339),
	})
}

// @Summary Get Pending HITL Requests
// @Description Get list of pending HITL requests (requires database query)
// @Tags HITL
// @Produce json
// @Success 200 {array} map[string]interface{} "List of pending requests"
// @Failure 500 {object} map[string]interface{} "Failed to query requests"
// @Router /api/v1/hitl/pending [get]
func (h *HITLHandler) GetPendingRequests(c *gin.Context) {
	// TODO: Query awaited_requests table
	// SELECT * FROM awaited_requests WHERE status = 'waiting' ORDER BY sent_at DESC;
	
	c.JSON(200, gin.H{
		"requests": []map[string]interface{}{},
		"message": "Not implemented - requires database integration",
	})
}
```

---

## 2. Register Routes

**File:** `cmd/production-api/main.go`

```go
// After other handlers are initialized...

// Initialize HITL handler
hitlHandler := gateway.NewHITLHandler(producer, appLogger)

// Register HITL routes
hitlGroup := router.Group("/api/v1/hitl")
{
	hitlGroup.POST("/respond", hitlHandler.SubmitHITLResponse)
	hitlGroup.GET("/pending", hitlHandler.GetPendingRequests)
}

appLogger.Info("HITL endpoints registered")
```

---

## 3. Usage Examples

### Using curl

```bash
# Submit HITL response
curl -X POST http://localhost:8080/api/v1/hitl/respond \
  -H "Content-Type: application/json" \
  -d '{
    "correlation_id": "c9694bbb-2f4e-4791-a07f-4a6bb1047aa8",
    "orchestration_id": "6e22f90a-55a4-4976-8678-069cdb4b0be7",
    "request_id": "dcd1b4c6-1224-4adc-b112-903f99353d03",
    "step_name": "hitl_review_brief",
    "data": {
      "company_name": "Leopardess Consulting",
      "tagline": "Agile AI Agents",
      "about_us": "We build AI systems...",
      "services": [...]
    }
  }'
```

### Using Python

```python
import requests

def submit_hitl_response(correlation_id, orchestration_id, request_id, step_name, data):
    url = "http://localhost:8080/api/v1/hitl/respond"
    
    payload = {
        "correlation_id": correlation_id,
        "orchestration_id": orchestration_id,
        "request_id": request_id,
        "step_name": step_name,
        "data": data
    }
    
    response = requests.post(url, json=payload)
    response.raise_for_status()
    
    return response.json()

# Example usage
data = {
    "company_name": "Leopardess Consulting",
    "tagline": "Agile AI Agents",
    # ... more fields
}

result = submit_hitl_response(
    correlation_id="c9694bbb-2f4e-4791-a07f-4a6bb1047aa8",
    orchestration_id="6e22f90a-55a4-4976-8678-069cdb4b0be7",
    request_id="dcd1b4c6-1224-4adc-b112-903f99353d03",
    step_name="hitl_review_brief",
    data=data
)

print(f"Submitted: {result['message_id']}")
```

### Using JavaScript/Node.js

```javascript
async function submitHITLResponse(correlationId, orchestrationId, requestId, stepName, data) {
  const response = await fetch('http://localhost:8080/api/v1/hitl/respond', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      correlation_id: correlationId,
      orchestration_id: orchestrationId,
      request_id: requestId,
      step_name: stepName,
      data: data,
    }),
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return await response.json();
}

// Example usage
const data = {
  company_name: "Leopardess Consulting",
  tagline: "Agile AI Agents",
  // ... more fields
};

submitHITLResponse(
  "c9694bbb-2f4e-4791-a07f-4a6bb1047aa8",
  "6e22f90a-55a4-4976-8678-069cdb4b0be7",
  "dcd1b4c6-1224-4adc-b112-903f99353d03",
  "hitl_review_brief",
  data
).then(result => {
  console.log('Submitted:', result.message_id);
});
```

---

## 4. Testing

**File:** `internal/gateway/hitl_handler_test.go`

```go
package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Produce(ctx context.Context, topic string, headers map[string]string, key []byte, value []byte) error {
	args := m.Called(ctx, topic, headers, key, value)
	return args.Error(0)
}

func TestSubmitHITLResponse_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockProducer := new(MockProducer)
	mockProducer.On("Produce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	logger, _ := zap.NewDevelopment()
	handler := NewHITLHandler(mockProducer, logger)

	req := HITLResponseRequest{
		CorrelationID:   "test-correlation",
		OrchestrationID: "test-orchestration",
		RequestID:       "test-request",
		StepName:        "test-step",
		Data: map[string]interface{}{
			"field1": "value1",
		},
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/hitl/respond", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SubmitHITLResponse(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockProducer.AssertExpectations(t)
}

func TestSubmitHITLResponse_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockProducer := new(MockProducer)
	logger, _ := zap.NewDevelopment()
	handler := NewHITLHandler(mockProducer, logger)

	// Missing required fields
	req := map[string]interface{}{
		"data": map[string]interface{}{},
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/hitl/respond", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SubmitHITLResponse(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

---

## 5. Documentation

Add to your API docs:

```yaml
paths:
  /api/v1/hitl/respond:
    post:
      summary: Submit HITL Response
      description: |
        Submit a human-in-the-loop response to continue a paused workflow.
        
        Use this endpoint when a workflow is waiting for human input.
        You'll need the request_id from the HITL notification.
      tags:
        - HITL
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - correlation_id
                - orchestration_id
                - request_id
                - step_name
                - data
              properties:
                correlation_id:
                  type: string
                  description: Correlation ID from HITL notification
                orchestration_id:
                  type: string
                  description: Orchestration ID from HITL notification
                request_id:
                  type: string
                  description: Request ID from HITL notification
                step_name:
                  type: string
                  description: Step name from HITL notification
                data:
                  type: object
                  description: Response data
      responses:
        '200':
          description: Response submitted successfully
        '400':
          description: Invalid request
        '500':
          description: Failed to send response
```

---

## 6. Benefits

With this API endpoint:

✅ **Clean Interface**
```bash
# Before (manual):
kubectl -n kafka run ... kcat -P ... <<JSON
{"headers":{"correlation_id":"$CORRELATION_ID",...}}
JSON

# After (API):
curl -X POST .../hitl/respond -d '{"correlation_id":"...","data":{...}}'
```

✅ **Built-in Validation**
- Required fields checked automatically
- Type validation
- Better error messages

✅ **Easy Integration**
- Call from any language
- Works with web UI
- Works with CLI tools
- Works with scripts

✅ **Better Logging**
- Every request logged
- Errors tracked
- Easy debugging

✅ **Can Add Features**
- Authentication
- Rate limiting
- Request history
- Notifications

---
