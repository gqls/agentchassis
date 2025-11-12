# Website Builder Agent Orchestration Integration Guide

## Overview

This guide details the integration of the website builder agents into your existing orchestration framework. The system uses your established Kafka-based messaging patterns with the new data_helpers.go utilities.

## Architecture Components

### 1. Agent Hierarchy

```
Website Builder Orchestrator (Master)
├── Website Capture Agent
│   └── Playwright Adapter (Python)
├── Website Vision Agent  
│   └── Vision ML Adapter (Python)
├── Website Code Analyzer Agent
│   └── Code Analysis Adapter (Go/Python)
├── Website Synthesis Agent
├── Content Strategist Agent
└── Component Library Agent
    └── PostgreSQL Vector DB
```

### 2. Message Flow Using data_helpers.go

The new data_helpers functions facilitate clean message passing:

```go
// Example in coordinator.go executeStep function
case "capture_site":
    // Extract input data using data_helpers
    inputData := GetInputData(state.CollectedData, c.logger)
    
    // Build request message
    requestMsg := BuildRequestMessage(
        execCtx,
        "playwright",  // adapter type
        "capture",     // action
        inputData,     // data from CollectedData
        config,        // step config
        c.logger,
    )
    
    // Send via existing pattern
    result, err = actions.CaptureSiteAction(ctx, params)
    
    // Process response and update CollectedData
    if captureResult, ok := result.(*actions.CaptureSiteResult); ok {
        if captureResult.AwaitResponse {
            // Store awaited request info
            state.AwaitedRequests[captureResult.RequestID] = types.AwaitedRequest{
                RequestID:    captureResult.RequestID,
                StepName:     stepName,
                TargetAgent:  "playwright-adapter",
                ResponseTopic: execCtx.ResponsesTopic,
            }
        }
    }
```

## Integration Steps

### Step 1: Add Action Handlers to Coordinator

In `internal/backend/agent-chassis/platform/orchestration/coordinator.go`, add cases to the executeStep function:

```go
func (c *SagaCoordinator) executeStep(ctx context.Context, state *OrchestrationState, stepName string) error {
    // ... existing code ...
    
    switch step.Action {
    // ... existing cases ...
    
    // New Website Builder Actions
    case "capture_site":
        result, err = actions.CaptureSiteAction(ctx, params)
        
    case "capture_hover_states":
        result, err = actions.CaptureHoverStatesAction(ctx, params)
        
    case "capture_scroll_animation":
        result, err = actions.CaptureScrollAnimationAction(ctx, params)
        
    case "validate_url":
        result, err = actions.ValidateURLAction(ctx, params)
        
    case "extract_website_assets":
        result, err = actions.ExtractWebsiteAssetsAction(ctx, params)
        
    case "upload_to_s3":
        result, err = actions.UploadToS3Action(ctx, params)
        
    case "analyze_visuals":
        result, err = actions.AnalyzeVisualsAction(ctx, params)
        
    case "analyze_code":
        result, err = actions.AnalyzeCodeAction(ctx, params)
        
    case "synthesize_design":
        result, err = actions.SynthesizeDesignAction(ctx, params)
        
    case "store_component":
        result, err = actions.StoreComponentAction(ctx, params)
        
    case "parallel_section_generation":
        result, err = actions.ParallelSectionGenerationAction(ctx, params)
        
    case "analyze_input_type":
        result, err = actions.AnalyzeInputTypeAction(ctx, params)
    
    // ... rest of existing code ...
    }
}
```

### Step 2: Handle Async Responses

When an adapter sends a response back, it needs to be correlated with the waiting orchestration:

```go
// In processMessage or similar handler
func (p *MessageProcessor) handleAdapterResponse(msg *types.ResponseMessage) error {
    // Extract data from response using data_helpers
    responseData := ExtractDataFromMessage(msg, p.logger)
    
    // Find the waiting orchestration
    requestID := msg.Headers["request_id"]
    state := p.orchestrationStore.GetByAwaitedRequest(requestID)
    
    if state != nil {
        // Update CollectedData with response
        UpdateCollectedData(
            state.CollectedData,
            state.AwaitedRequests[requestID].StepName,
            responseData,
            p.logger,
        )
        
        // Remove from awaited
        delete(state.AwaitedRequests, requestID)
        
        // Continue workflow
        p.coordinator.ContinueExecution(state)
    }
}
```

### Step 3: Deploy Python Adapters

Create Kubernetes deployments for the Python adapters:

```yaml
# playwright-adapter-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: playwright-adapter
  namespace: agent-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: playwright-adapter
  template:
    metadata:
      labels:
        app: playwright-adapter
    spec:
      containers:
      - name: playwright-adapter
        image: your-registry/playwright-adapter:latest
        env:
        - name: KAFKA_BROKER
          value: "kafka:9092"
        - name: REQUEST_TOPIC
          value: "system.adapter.playwright.requests"
        - name: S3_ENDPOINT
          value: "https://s3.us-west-002.backblazeb2.com"
        - name: S3_BUCKET
          value: "website-captures"
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: s3-credentials
              key: access-key-id
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: s3-credentials
              key: secret-access-key
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
```

### Step 4: KEDA Autoscaling (Optional)

To scale adapters based on Kafka lag:

```yaml
# playwright-adapter-scaledobject.yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: playwright-adapter-scaler
  namespace: agent-system
spec:
  scaleTargetRef:
    name: playwright-adapter
  minReplicaCount: 0
  maxReplicaCount: 5
  triggers:
  - type: kafka
    metadata:
      bootstrapServers: kafka:9092
      consumerGroup: playwright-adapter-group
      topic: system.adapter.playwright.requests
      lagThreshold: "1"
```

## Message Flow Example

Here's a complete flow for capturing a website:

1. **User Request** → Generic Agent:
```json
{
  "action": "orchestrate",
  "config": {
    "group_type": "website-builder-orchestrator"
  },
  "input_data": {
    "target_url": "example.com",
    "business_name": "New Business",
    "business_type": "e-commerce"
  }
}
```

2. **Orchestrator** spawns Website Capture Agent:
```go
// Uses BuildRequestMessage from data_helpers.go
requestMsg := BuildRequestMessage(
    execCtx,
    "website-capture",
    "initialize",
    inputData,
    agentConfig,
    logger,
)
```

3. **Website Capture Agent** sends to Playwright adapter:
```go
// In CaptureSiteAction
requestPayload := map[string]interface{}{
    "request_id": requestID,
    "action": "capture",
    "url": url,
    "capture_config": captureConfig,
    "reply_to_topic": params.ExecutionContext.ResponsesTopic,
}
// Send to Kafka topic: system.adapter.playwright.requests
```

4. **Playwright Adapter** processes and responds:
```python
# Captures website
result = await self.handle_capture(request)

# Sends response back
response = {
    'request_id': request_id,
    'result': {
        'screenshot_base64': screenshot,
        'html_content': html,
        's3_paths': {...}
    }
}
# Send to reply_to_topic
```

5. **Website Capture Agent** receives response:
```go
// Response handled by coordinator
// Updates CollectedData using UpdateCollectedData from data_helpers.go
UpdateCollectedData(
    state.CollectedData,
    "capture_desktop",
    responseData,
    logger,
)
```

6. **Flow continues** to next step (capture_mobile, then analyze_visuals, etc.)

## Logging and Tracking

The system provides comprehensive logging at each stage:

```go
// Example logging in action
logger.Info("Executing CaptureSiteAction",
    zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
    zap.String("step_name", params.ExecutionContext.StepName),
    zap.String("request_id", requestID),
    zap.String("topic", adapterTopic))
```

Message tracing through the system:
- Request ID tracks individual adapter calls
- Orchestration ID tracks the entire workflow
- Correlation ID tracks the original user request
- Step Name tracks which workflow step generated the call

## Testing the Integration

### 1. Unit Test for Actions

```go
func TestCaptureSiteAction(t *testing.T) {
    // Setup mock producer
    mockProducer := &MockKafkaProducer{}
    
    // Create params
    params := actions.ActionParams{
        ExecutionContext: &types.ExecutionContext{
            OrchestrationID: "test-orch-123",
            ResponsesTopic: "test.responses",
        },
        CollectedData: map[string]interface{}{
            "input_data": map[string]interface{}{
                "target_url": "https://example.com",
            },
        },
        StepConfig: types.StepConfig{
            Config: map[string]interface{}{
                "capture_config": map[string]interface{}{
                    "viewport": map[string]int{
                        "width": 1920,
                        "height": 1080,
                    },
                },
            },
        },
        Producer: mockProducer,
        Logger: zap.NewNop(),
    }
    
    // Execute action
    result, err := actions.CaptureSiteAction(context.Background(), params)
    
    // Verify
    assert.NoError(t, err)
    assert.True(t, result.(*actions.CaptureSiteResult).Success)
    assert.True(t, result.(*actions.CaptureSiteResult).AwaitResponse)
}
```

### 2. Integration Test

```bash
# Send test message to orchestrator
curl -X POST http://localhost:8080/orchestrate \
  -H "Content-Type: application/json" \
  -d '{
    "action": "orchestrate",
    "config": {
      "workflow": "website-builder"
    },
    "input_data": {
      "target_url": "https://example.com",
      "business_name": "Test Business",
      "business_type": "retail"
    }
  }'
```

## Monitoring

Key metrics to track:

1. **Adapter Performance**:
    - Message processing time
    - Capture success rate
    - S3 upload success rate

2. **Orchestration Performance**:
    - Workflow completion time
    - Step execution time
    - Awaited request duration

3. **Kafka Metrics**:
    - Consumer lag
    - Message throughput
    - Error rates

## Next Steps

1. **Implement remaining agents**:
    - Vision Agent (UI element detection)
    - Code Analyzer Agent (HTML/CSS cleaning)
    - Synthesis Agent (correlation logic)
    - Content Strategy Agent

2. **Add more adapters**:
    - Vision ML Adapter (YOLOv8, OCR)
    - LLM Adapter (code generation)
    - Vector DB Adapter (component storage)

3. **Enhance error handling**:
    - Retry logic for failed captures
    - Fallback strategies
    - Circuit breakers for adapters

4. **Optimize performance**:
    - Batch processing for multiple URLs
    - Caching for repeated captures
    - Parallel adapter execution

## Summary

This integration:
- Maintains your existing orchestration patterns
- Uses data_helpers.go for clean message handling
- Provides async, scalable adapter architecture
- Enables comprehensive website analysis and generation
- Supports detailed logging and tracking as requested

The modular design allows incremental development - start with the capture functionality and gradually add more sophisticated analysis and generation capabilities.

