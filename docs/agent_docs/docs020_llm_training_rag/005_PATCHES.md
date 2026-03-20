# Patch Instructions for Existing Go Files

## PATCH 01: platform/aiservice/anthropic.go

### What: Capture usage tokens from API response and write back to options map

The Anthropic API returns input/output token counts but the current client discards them.
After this patch, token counts are written back into the options map so ExecuteLLMPromptAction
can pass them to the logger.

### Find the response struct (around where JSON is unmarshaled):

Add a Usage field to the response struct:

```go
// In the response parsing section, ensure the struct includes:
type anthropicResponse struct {
    // ... existing fields ...
    Content []struct {
        Type string `json:"type"`
        Text string `json:"text"`
    } `json:"content"`
    Usage struct {
        InputTokens  int `json:"input_tokens"`
        OutputTokens int `json:"output_tokens"`
    } `json:"usage"`
}
```

### After successful response parsing, add token write-back:

```go
// After parsing the response, before returning:
if options != nil {
    options["__usage_input_tokens"] = response.Usage.InputTokens
    options["__usage_output_tokens"] = response.Usage.OutputTokens
}
```

---

## PATCH 02: platform/orchestration/actions/ai_actions.go

### What: Add LLM call logging after each execute_llm_prompt call + add ollama to createAIClient

### Part A: Add timing and logging around the LLM call

Find the `GenerateText` call in `ExecuteLLMPromptAction`. Add timing before it and logging after:

```go
// BEFORE the GenerateText call, add:
import "time"  // add to imports if not present

llmCallStart := time.Now()

// ... existing GenerateText call ...

// AFTER successful GenerateText, add:
latencyMs := int(time.Since(llmCallStart).Milliseconds())

inputTokens := 0
outputTokens := 0
if it, ok := options["__usage_input_tokens"].(int); ok {
    inputTokens = it
}
if ot, ok := options["__usage_output_tokens"].(int); ok {
    outputTokens = ot
}

LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
    AgentType:       params.Headers["agent_type"],
    AgentID:         params.Headers["agent_id"],
    StepName:        params.ExecutionContext.StepName,
    OrchestrationID: params.ExecutionContext.OrchestrationID,
    CorrelationID:   params.ExecutionContext.CorrelationID,
    Model:           model,
    ModelResolved:   resolvedModel,
    Provider:        provider,
    PromptTemplate:  promptTemplate,
    PromptRendered:  renderedPrompt,
    ResponseText:    responseText,
    InputTokens:     inputTokens,
    OutputTokens:    outputTokens,
    LatencyMs:       latencyMs,
    Success:         true,
})

// In the ERROR path (where GenerateText fails), add:
latencyMs := int(time.Since(llmCallStart).Milliseconds())
LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
    AgentType:       params.Headers["agent_type"],
    AgentID:         params.Headers["agent_id"],
    StepName:        params.ExecutionContext.StepName,
    OrchestrationID: params.ExecutionContext.OrchestrationID,
    CorrelationID:   params.ExecutionContext.CorrelationID,
    Model:           model,
    Provider:        provider,
    PromptTemplate:  promptTemplate,
    PromptRendered:  renderedPrompt,
    LatencyMs:       latencyMs,
    Success:         false,
    ErrorMessage:    err.Error(),
})
```

Note: `model`, `resolvedModel`, `provider`, `promptTemplate`, `renderedPrompt`, `responseText` 
are variable names from the existing code. Check what the actual variable names are in your 
version of ai_actions.go and adjust accordingly.

### Part B: Add ollama case to createAIClient

Find the `switch provider` block in `createAIClient`:

```go
// FIND:
switch provider {
case "anthropic":
    return aiservice.NewAnthropicClient(ctx, aiServiceConfig)
// ...
default:
    return nil, fmt.Errorf("unsupported AI provider: %s", provider)
}

// ADD the ollama case:
case "ollama":
    return aiservice.NewOllamaClient(ctx, aiServiceConfig)
```

---

## PATCH 03: platform/orchestration/actions/registry.go

### What: Register rag_lookup and rag_index in GlobalActionRegistry

Find the GlobalActionRegistry map. Add these entries in the storage section 
(after existing memory/cache actions):

```go
// RAG — retrieval-augmented generation
"rag_lookup": {
    Handler:     RAGLookupAction,
    Category:    "storage",
    Description: "Search the knowledge base for relevant content using vector similarity",
    IsLocal:     true,
},
"rag_index": {
    Handler:     RAGIndexAction,
    Category:    "storage",
    Description: "Chunk, embed, and store content in the knowledge base",
    IsLocal:     true,
},
```
