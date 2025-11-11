# Data Helpers Quick Reference for Website Builder

## Using data_helpers.go in Website Builder Agents

The `data_helpers.go` provides essential functions for managing data flow between agents in the website builder system.

## Common Usage Patterns

### 1. In the Orchestrator - Building Child Requests

```go
func (c *SagaCoordinator) spawnCaptureAgent(ctx context.Context, state *OrchestrationState) error {
    // Extract current input data
    inputData := GetInputData(state.CollectedData, c.logger)
    
    // Prepare config for child agent
    agentConfig := map[string]interface{}{
        "capture_mode": "full_site",
        "include_interactions": true,
    }
    
    // Build initialization request for spawned agent
    initRequest := BuildInitializationRequest(
        state.ExecutionContext,
        "website-capture",        // child agent type
        "capture_specialist",      // functional role
        inputData,                 // pass input data to child
        agentConfig,              // agent-specific config
        c.logger,
    )
    
    // Send via Kafka
    return c.producer.Send(initRequest)
}
```

### 2. In Child Agents - Processing Requests

```go
func (a *CaptureAgent) processRequest(msg interface{}) error {
    // Build CollectedData from incoming message
    collectedData := BuildCollectedData(
        msg,
        a.executionContext,
        a.logger,
    )
    
    // Extract clean input data
    inputData := GetInputData(collectedData, a.logger)
    
    // Get specific fields
    targetURL := inputData["target_url"].(string)
    captureConfig := inputData["capture_config"].(map[string]interface{})
    
    // Process the capture...
    result := a.performCapture(targetURL, captureConfig)
    
    // Update CollectedData with results
    UpdateCollectedData(
        collectedData,
        "capture_result",
        result,
        a.logger,
    )
    
    return nil
}
```

### 3. Building Agent-to-Agent Requests

```go
func (a *CaptureAgent) callPlaywrightAdapter(ctx context.Context) error {
    // Get execution context
    execCtx := a.collectedData["__execution_context__"].(*types.ExecutionContext)
    
    // Prepare data for adapter
    adapterData := map[string]interface{}{
        "url": a.targetURL,
        "viewport": map[string]int{
            "width": 1920,
            "height": 1080,
        },
        "capture_options": map[string]interface{}{
            "full_page": true,
            "wait_until": "networkidle",
        },
    }
    
    // Build request message
    requestMsg := BuildRequestMessage(
        execCtx,
        "playwright-adapter",
        "capture",
        adapterData,
        nil, // no special config
        a.logger,
    )
    
    // Send and track
    requestID := uuid.New().String()
    a.awaitedRequests[requestID] = requestMsg
    
    return a.kafkaProducer.Send("system.adapter.playwright.requests", requestMsg)
}
```

### 4. Processing Responses from Child Agents

```go
func (c *SagaCoordinator) handleChildResponse(msg *types.ResponseMessage) error {
    // Extract response data
    responseData := NormalizeResponseData(msg.Body, c.logger)
    
    // Find the orchestration state
    orchID := msg.Headers["parent_orchestration_id"]
    state := c.stateManager.Get(orchID)
    
    // Determine which step this response is for
    stepName := msg.Headers["in_response_to_step"]
    
    // Update CollectedData with child's results
    UpdateCollectedData(
        state.CollectedData,
        stepName,
        responseData,
        c.logger,
    )
    
    // Check if we have all needed data
    if stepData, ok := GetStepData(state.CollectedData, stepName, c.logger); ok {
        // Process the step data
        c.processStepResults(stepData)
    }
    
    return nil
}
```

### 5. Aggregating Multiple Agent Results

```go
func (a *SynthesisAgent) aggregateAnalysisResults() map[string]interface{} {
    // Get data from multiple previous steps
    stepNames := []string{
        "capture_desktop",
        "capture_mobile", 
        "analyze_visuals",
        "analyze_code",
    }
    
    allStepData := GetMultipleStepData(
        a.collectedData,
        stepNames,
        a.logger,
    )
    
    // Extract specific fields using path notation
    desktopScreenshot, _ := GetFieldFromPath(
        a.collectedData,
        "capture_desktop.screenshot_base64",
        a.logger,
    )
    
    visualMap, _ := GetFieldFromPath(
        a.collectedData,
        "analyze_visuals.visual_map",
        a.logger,
    )
    
    // Merge data for synthesis
    synthesisInput := MergeInputData(
        allStepData["analyze_visuals"].(map[string]interface{}),
        allStepData["analyze_code"].(map[string]interface{}),
        a.logger,
    )
    
    return synthesisInput
}
```

### 6. Building Final Response

```go
func (a *WebsiteBuilderOrchestrator) completeWorkflow() *types.ResponseMessage {
    // Gather all results
    finalWebsite := map[string]interface{}{
        "html": a.collectedData["generated_html"],
        "css": a.collectedData["generated_css"],
        "js": a.collectedData["generated_js"],
        "assets": a.collectedData["extracted_assets"],
        "metadata": map[string]interface{}{
            "generated_at": time.Now().UTC(),
            "source_url": a.collectedData["input_data"].(map[string]interface{})["target_url"],
        },
    }
    
    // Build response message
    response := BuildResponseMessage(
        a.executionContext,
        true, // success
        finalWebsite,
        nil, // no error
        a.logger,
    )
    
    return response
}
```

### 7. Transforming Data Between Steps

```go
func (c *SagaCoordinator) prepareDataForVisionAnalysis() map[string]interface{} {
    // Define transformation spec
    transformSpec := map[string]interface{}{
        "field_mappings": map[string]interface{}{
            "screenshot": "capture_desktop.screenshot_base64",
            "mobile_screenshot": "capture_mobile.screenshot_base64",
        },
        "include_fields": []interface{}{
            "viewport_dimensions",
            "capture_metadata",
        },
        "add_fields": map[string]interface{}{
            "analysis_type": "ui_elements",
            "detection_threshold": 0.8,
        },
    }
    
    // Transform data for vision agent
    visionInput := TransformDataForAction(
        c.state.CollectedData,
        transformSpec,
        c.logger,
    )
    
    return visionInput
}
```

### 8. Error Handling with Data Helpers

```go
func (a *Agent) safeDataExtraction(msg interface{}) (map[string]interface{}, error) {
    // Safely extract data with fallbacks
    inputData := ExtractDataFromMessage(msg, a.logger)
    
    if len(inputData) == 0 {
        a.logger.Warn("No input data found, using defaults")
        inputData = map[string]interface{}{
            "target_url": "https://example.com",
            "capture_mode": "basic",
        }
    }
    
    // Use default values for missing fields
    targetURL := GetFieldFromPathWithDefault(
        inputData,
        "target_url",
        "https://example.com",
        a.logger,
    ).(string)
    
    captureMode := GetFieldFromPathWithDefault(
        inputData,
        "capture_mode",
        "full",
        a.logger,
    ).(string)
    
    return map[string]interface{}{
        "target_url": targetURL,
        "capture_mode": captureMode,
    }, nil
}
```

## Key Principles

1. **Always Extract First**: Use `ExtractDataFromMessage()` to get clean data from any message format
2. **Build Structured Messages**: Use `BuildRequestMessage()` and `BuildResponseMessage()` for consistency
3. **Update Incrementally**: Use `UpdateCollectedData()` to add results from each step
4. **Access Safely**: Use `GetFieldFromPath()` with error handling or `GetFieldFromPathWithDefault()`
5. **Transform When Needed**: Use `TransformDataForAction()` to prepare data for specific agents
6. **Maintain Context**: Always preserve `__execution_context__` in CollectedData

## Common Patterns

### Pattern 1: Parent-Child Communication
```go
Parent: BuildInitializationRequest() → Child
Child: BuildCollectedData() → Process → BuildResponseMessage() → Parent
Parent: NormalizeResponseData() → UpdateCollectedData()
```

### Pattern 2: Async Adapter Calls
```go
Agent: BuildRequestMessage() → Kafka Topic
Adapter: Process → Send Response
Agent: ExtractDataFromMessage() → UpdateCollectedData()
```

### Pattern 3: Multi-Step Aggregation
```go
Step1: UpdateCollectedData(collected, "step1", result1)
Step2: UpdateCollectedData(collected, "step2", result2)
Final: GetMultipleStepData(collected, ["step1", "step2"])
```

## Debugging Tips

1. **Log Data Extraction**:
```go
data := ExtractDataFromMessage(msg, logger)
logger.Debug("Extracted data", zap.Any("data", data))
```

2. **Verify CollectedData State**:
```go
logger.Debug("Current CollectedData",
    zap.Int("fields", len(collectedData)),
    zap.Any("keys", getMapKeys(collectedData)))
```

3. **Track Data Flow**:
```go
// Before transformation
logger.Debug("Pre-transform", zap.Any("data", sourceData))

// After transformation  
transformed := TransformDataForAction(sourceData, spec, logger)
logger.Debug("Post-transform", zap.Any("data", transformed))
```

This reference guide shows practical usage of every major function in data_helpers.go within the website builder context.

