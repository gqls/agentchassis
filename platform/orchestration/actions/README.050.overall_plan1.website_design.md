Looking at your agent definitions, you have a sophisticated website-building ecosystem. Here's a step-by-step plan to build a basic website using your multi-agent system:

## Phase 1: Foundation Setup

### Step 1: Test Individual Agent Actions
Before orchestrating the full workflow, verify each agent's core action works:

1. **Test execute_llm_prompt action** - ensure it can call the LLM API
2. **Test upload_to_s3 action** - verify S3 connectivity and permissions
3. **Test agent spawning** - confirm agents can be dynamically created

### Step 2: Simplify the Website Builder Workflow
Start with a minimal workflow that uses fewer agents:

```yaml
workflow:
  start_step: "analyze_domain"
  steps:
    analyze_domain:
      action: "call_agent"
      config:
        agent_type: "domain-analyst"
        input_field: "business_info"
      next_step: "create_content"
    
    create_content:
      action: "call_agent"
      config:
        agent_type: "content-creator"
        input_field: "content_requirements"
      next_step: "develop_html"
    
    develop_html:
      action: "call_agent"
      config:
        agent_type: "html-developer"
        input_field: "site_content"
      next_step: "complete"
    
    complete:
      action: "complete_workflow"
```

## Phase 2: Core Actions Implementation

### Step 3: Implement execute_llm_prompt Action
```go
func ExecuteLLMPromptAction(ctx context.Context, params ActionParams) (interface{}, error) {
    // Extract prompt template from config
    // Replace template variables with actual data
    // Call Anthropic API
    // Return structured response
}
```

### Step 4: Implement upload_to_s3 Action
```go
func UploadToS3Action(ctx context.Context, params ActionParams) (interface{}, error) {
    // Extract HTML content from collected data
    // Connect to S3/B2
    // Upload files
    // Return public URL
}
```

## Phase 3: Data Flow Management

### Step 5: Structure Data Between Agents
Define clear data contracts between agents:

```json
// Input to domain-analyst
{
  "business_type": "restaurant",
  "business_name": "Joe's Pizza"
}

// Output from domain-analyst → Input to content-creator
{
  "target_audience": "families, local community",
  "key_features": ["online ordering", "menu display", "location info"],
  "content_tone": "friendly, casual"
}

// Output from content-creator → Input to html-developer
{
  "page_content": {
    "home": "Welcome to Joe's Pizza...",
    "about": "Family-owned since 1985..."
  },
  "meta_descriptions": {...}
}

// Output from html-developer
{
  "html_code": "<!DOCTYPE html>...",
  "css_code": "body { ... }",
  "js_code": "function init() { ... }"
}
```

### Step 6: Implement Data Extraction in CallAgentAction
Ensure CallAgentAction properly extracts and passes data:

```go
// In CallAgentAction
inputField := config["input_field"].(string)
inputData := params.CollectedData[inputField]

// Build message with proper data structure
requestBody := map[string]interface{}{
    "input_data": inputData,
    "action": "process",
}
```

## Phase 4: Testing Strategy

### Step 7: Test with Simple Input
Start with minimal input:

```json
{
  "action": "process",
  "input_data": {
    "business_type": "pizza restaurant",
    "business_name": "Test Pizza"
  }
}
```

### Step 8: Monitor Each Step
Add comprehensive logging at each stage:
- Log when each agent is spawned
- Log input/output for each agent
- Track orchestration state transitions

## Phase 5: Progressive Enhancement

### Step 9: Add More Agents
Once basic flow works, add:
1. **site-architect** - for navigation structure
2. **visual-designer** - for styling
3. **site-publisher** - for deployment

### Step 10: Add Error Handling
- Implement retry logic for failed agents
- Add fallback strategies
- Create error aggregation

## Phase 6: Final Integration

### Step 11: Complete Website Builder Workflow
```yaml
workflow:
  start_step: "spawn_team"
  steps:
    spawn_team:
      action: "spawn_group"
      config:
        agents:
          analyst: { type: "domain-analyst" }
          architect: { type: "site-architect" }
          content: { type: "content-creator" }
          developer: { type: "html-developer" }
      next_step: "analyze_domain"
    
    # ... rest of the workflow
```

### Step 12: Production Readiness
- Add timeout handling
- Implement cleanup for failed workflows
- Add monitoring/metrics
- Create deployment pipeline

## Implementation Priority

1. **Week 1**: Steps 1-4 (Foundation and core actions)
2. **Week 2**: Steps 5-8 (Data flow and basic testing)
3. **Week 3**: Steps 9-10 (Enhancement and error handling)
4. **Week 4**: Steps 11-12 (Full integration and production)

## Key Success Factors

1. **Start small** - Get domain-analyst → content-creator working first
2. **Mock LLM responses initially** - Test orchestration without API calls
3. **Use structured logging** - Track data flow at every step
4. **Test incrementally** - Verify each agent before adding the next
5. **Monitor resource usage** - Ensure K8s pods have sufficient resources

This plan builds on your successful calculator orchestration, applying the same patterns to a more complex, real-world use case. The key is maintaining clear data contracts between agents and ensuring each agent knows exactly what data structure to expect and produce.