Here's a concrete example using the helpers to extract data and package it for a child agent:

```go
// Example: Content Creator Hero agent processing a request and calling Content Researcher

func (a *ContentCreatorAgent) ProcessAndSpawnResearcher(params ActionParams) error {
    logger := params.Logger
    
    // ============================================================================
    // STEP 1: EXTRACT THE DATA (what you called "businessData")
    // ============================================================================
    
    // Get the clean data from whatever nested structure we received
    currentData := GetInputData(params.CollectedData, logger)
    
    // currentData now contains the actual data, e.g.:
    // {
    //   "business_name": "Golden Crust Bakery",
    //   "business_type": "artisanal bakery",
    //   "location": "San Francisco",
    //   ...any other fields...
    // }
    
    logger.Info("Extracted data from parent",
        zap.Any("data_fields", getMapKeys(currentData)))
    
    // ============================================================================
    // STEP 2: PREPARE DATA FOR THE CHILD (transform if needed)
    // ============================================================================
    
    // Option A: Pass all data as-is to child
    childData := currentData
    
    // Option B: Select specific fields for the child
    childData := map[string]interface{}{
        "business_type": currentData["business_type"],
        "business_name": currentData["business_name"],
        // Only pass what the researcher needs
    }
    
    // Option C: Transform/enrich the data for the child
    childData := TransformDataForAction(
        currentData,
        map[string]interface{}{
            "include_fields": []interface{}{"business_name", "business_type"},
            "add_fields": map[string]interface{}{
                "research_depth": "comprehensive",
                "timestamp": time.Now().Unix(),
            },
        },
        logger,
    )
    
    // ============================================================================
    // STEP 3: PACKAGE THE MESSAGE FOR THE CHILD
    // ============================================================================
    
    // For spawning a new agent (initialization)
    if spawning {
        agentInfo := map[string]interface{}{
            "agent_id":   GenerateAgentID(),
            "agent_type": "content-researcher",
            "agent_name": "content-researcher-" + GenerateShortID(),
        }
        
        messageBody := PackageInitializationMessage(
            agentInfo,
            "researcher", // role
            childData,    // the data to pass
            logger,
        )
        
        // messageBody now contains:
        // {
        //   "action": "initialize",
        //   "is_initialization": true,
        //   "agent_info": {...},
        //   "role": "researcher",
        //   "data": {
        //     "business_name": "Golden Crust Bakery",
        //     "business_type": "artisanal bakery",
        //     ...
        //   }
        // }
    }
    
    // For calling an existing agent (process request)
    if calling {
        // Package the data with any workflow config
        messageBody := PackageDataForChild(
            childData,
            map[string]interface{}{
                "prompt_template": "Research {{.business_type}} industry trends",
                "max_tokens": 2000,
            },
            "process", // action
            logger,
        )
        
        // messageBody now contains:
        // {
        //   "action": "process",
        //   "data": {
        //     "business_name": "Golden Crust Bakery",
        //     "business_type": "artisanal bakery",
        //     ...
        //   },
        //   "config": {
        //     "prompt_template": "Research {{.business_type}} industry trends",
        //     "max_tokens": 2000
        //   }
        // }
    }
    
    // ============================================================================
    // STEP 4: SEND THE MESSAGE (with headers added by messaging layer)
    // ============================================================================
    
    // Build complete message with headers
    fullMessage := map[string]interface{}{
        "headers": buildHeaders(params.ExecutionContext), // Your existing header builder
        "body":    messageBody,
    }
    
    // Send via Kafka
    err := a.kafkaProducer.Send(childRequestsTopic, fullMessage)
    
    return err
}

// ============================================================================
// RECEIVING SIDE: Child agent receiving and extracting the data
// ============================================================================

func (r *ResearcherAgent) HandleIncomingRequest(rawMessage map[string]interface{}) error {
    logger := r.logger
    
    // Build CollectedData from the raw message
    collectedData := NormalizeCollectedData(
        rawMessage,
        r.executionContext,
        r.requestsTopic,
        logger,
    )
    
    // Extract the clean data - handles ANY level of nesting
    inputData := GetInputData(collectedData, logger)
    
    // inputData now contains:
    // {
    //   "business_name": "Golden Crust Bakery",
    //   "business_type": "artisanal bakery",
    //   ...
    // }
    
    // Use the data
    businessName := inputData["business_name"].(string)
    businessType := inputData["business_type"].(string)
    
    logger.Info("Researcher received data",
        zap.String("business_name", businessName),
        zap.String("business_type", businessType))
    
    // Do the research...
    researchResults := r.performResearch(inputData)
    
    // Package response back to parent
    responseBody := PackageResponseMessage(
        true, // success
        researchResults,
        "",   // no error
        logger,
    )
    
    return r.sendResponse(responseBody)
}
```

## Key Points:

1. **Getting the data**: Just call `GetInputData(params.CollectedData, logger)` - it finds the data regardless of how deeply nested it is

2. **Packaging for child**: Use `PackageDataForChild()` or `PackageInitializationMessage()` - these create clean message bodies

3. **No hardcoding**: Notice we never hardcode "businessData" - we just work with maps and the actual field names come from the data itself

4. **Dynamic transformation**: The `TransformDataForAction()` function lets you specify which fields to pass, add new fields, or map field names - all driven by configuration, not hardcoded

5. **Clean extraction on receiving side**: The child just calls `GetInputData()` and gets clean data, no matter how the parent packaged it

The beauty is that this works whether the data comes in as:
- `body.input_data`
- `body.data`
- `input_data.input_data`
- Or any other nesting level

The helpers handle all the complexity, and your agents just work with clean data maps.