Excellent question. You're absolutely right - **we should overwrite at the top level**.

Here's why:

## The Data Flow Model

Each agent's orchestration has its **own independent CollectedData**. When an agent calls a child:

1. **Parent decides** what data to send (via template rendering)
2. **Child receives** that data as **its new context**
3. **Child doesn't need** the parent's original data
4. **Child's `input_data`** is the data the parent explicitly sent

## Example Flow

```yaml
# Generic orchestrator receives:
input_data:
  business_name: "Golden Crust Bakery"
  business_type: "artisanal bakery"

# Generic's CollectedData after initialization:
CollectedData:
  input_data:
    business_name: "Golden Crust Bakery"
    business_type: "artisanal bakery"

# Generic calls content-creator-hero, passing everything:
# (no input_data spec, so passes whole input_data)

# Content-creator-hero's CollectedData after initialization:
CollectedData:
  input_data:
    business_name: "Golden Crust Bakery"  # Same as parent
    business_type: "artisanal bakery"     # Same as parent

# Content-creator-hero calls researcher with SELECTIVE data:
call_researcher:
  config:
    input_data:
      business_type: "{{.input_data.business_type}}"  # Only this field
      research_focus: "hero section"                   # Plus new field

# Researcher's CollectedData after initialization:
CollectedData:
  input_data:
    business_type: "artisanal bakery"  # From parent template
    research_focus: "hero section"      # New field from parent
  # NO business_name - parent didn't send it
```

## Why This Works

1. **Clean slate**: Each agent starts fresh with only what it needs
2. **Explicit contract**: Parent declares exactly what child receives
3. **No pollution**: Child can't accidentally reference parent's irrelevant data
4. **Transformations work**: Parent can add/modify/filter data for child

## The Code Flow

**In `call_agent.go` when building request:**
```go
// 1. Get input_data spec from workflow config
inputDataSpec := stepConfig.Config["input_data"].(map[string]interface{})

// 2. Render templates against parent's CollectedData
renderedData := renderTemplates(inputDataSpec, params.CollectedData)
// Result: {business_type: "artisanal bakery", research_focus: "hero section"}

// 3. Send to child in message body
message.Body["input_data"] = renderedData
```

**In `state.go` when child initializes:**
```go
// Extract from incoming message and put at TOP LEVEL
if body, ok := message["body"].(map[string]interface{}); ok {
    if inputData, ok := body["input_data"].(map[string]interface{}); ok {
        // Overwrite top level - this is the child's NEW context
        collectedData["input_data"] = inputData
    }
}
```

**Result**: Every agent has `CollectedData.input_data` containing exactly what its parent intended it to have.

## When Would You NOT Overwrite?

Only if you wanted the child to **see both** its received data AND the parent's original data. But that's an anti-pattern because:
- Creates confusion about which data is "current"
- Violates encapsulation
- Makes debugging harder ("where did this field come from?")
- Prevents data transformation

So yes: **Always overwrite `input_data` at the top level**. Each agent's `input_data` is its own context, derived from but independent of its parent's.


