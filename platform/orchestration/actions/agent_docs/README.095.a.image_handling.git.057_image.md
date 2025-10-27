# Image Generation Integration - Implementation Summary

## Overview
This document summarizes the work to integrate image generation capabilities into the agent workflow system, enabling agents to generate images as part of their workflows.

## Architecture Context

### Current System
- **Workflows**: Composed of steps that execute actions like `spawn_agent`, `call_agent`, `execute_llm_prompt`, `aggregate_data`, `complete_workflow`
- **Actions**: Receive `ActionParams` struct with context, data, and resources; return `map[string]interface{}` results
- **Data Flow**: Results stored in `CollectedData` and available to subsequent steps
- **Message Building**: New `data_helpers.go` provides utilities for extracting/building messages using existing types

### Old Image Generator Adapter
- Separate standalone service with own Kafka topics
- Not integrated with workflow system
- Located at: `/internal/adapters/imagegenerator/adapter.go`

## Implementation Created

### 1. Image Actions File (`image_actions.go`)
**Location**: `platform/orchestration/actions/image_actions.go`

**Key Function**: `GenerateImageAction(params ActionParams) (map[string]interface{}, error)`

**Features**:
- Extracts prompts from multiple sources (step config, collected data, input data)
- Supports template rendering (e.g., `{{.business_name}}`)
- Calls external API (Stability AI) for image generation
- Stores images in object storage (S3-compatible)
- Returns image URI and metadata

**Configuration Support**:
```go
config := map[string]interface{}{
    "aspect_ratio": "1:1",      // or "16:9", "9:16", etc.
    "style": "photographic",     // optional style preset
    "seed": 12345,              // optional seed for reproducibility
    "output_format": "png",     // png, jpeg, webp
}
```

**Environment Variables Required**:
```bash
# Image API
STABILITY_API_KEY=sk-...
STABILITY_API_ENDPOINT=https://api.stability.ai/v2beta/stable-image/generate/core

# Object Storage
STORAGE_PROVIDER=s3              # or gcs, azure
STORAGE_BUCKET=generated-images
STORAGE_REGION=us-east-1
STORAGE_ENDPOINT=               # optional, for S3-compatible services
STORAGE_ACCESS_KEY_ID=
STORAGE_SECRET_ACCESS_KEY=
```

### 2. Usage in Workflows

**Example Workflow Step**:
```json
{
  "generate_hero_image": {
    "action": "generate_image",
    "description": "Generate hero section image",
    "config": {
      "prompt": "A warm, inviting image of freshly baked artisanal bread for {{.business_name}}",
      "aspect_ratio": "16:9",
      "style": "photographic"
    },
    "next_step": "complete"
  }
}
```

**With Template Rendering**:
```json
{
  "generate_hero_image": {
    "action": "generate_image",
    "description": "Generate hero section image",
    "config": {
      "prompt": "Create an image showing {{.business_type}} with {{.style_description}}",
      "aspect_ratio": "16:9"
    },
    "next_step": "complete"
  }
}
```

**Using Input Fields**:
```json
{
  "generate_hero_image": {
    "action": "generate_image",
    "description": "Generate image based on content",
    "config": {
      "input_fields": ["generate_hero", "call_researcher"],
      "aspect_ratio": "16:9"
    },
    "next_step": "complete"
  }
}
```

## Next Steps

### 1. Register Action in Coordinator
**File to modify**: `platform/orchestration/coordinator.go`

Find the action dispatch logic (likely in `executeLocalAction` method) and add:
```go
case "generate_image":
    result, err = actions.GenerateImageAction(params)
```

Look for existing cases like:
```go
case "execute_llm_prompt":
    result, err = actions.ExecuteLLMPromptAction(params)
case "spawn_agent":
    result, err = actions.SpawnAgentAction(params)
case "call_agent":
    result, err = actions.CallAgentAction(params)
```

### 2. Fix Import Issues
The `image_actions.go` file needs to be reviewed for:
- Correct import paths for the codebase
- String manipulation functions (replace simplified `replaceAll` with `strings.ReplaceAll`)
- Template rendering (consider using `text/template` for robust rendering)

### 3. Test Integration
Create test workflows to verify:
1. **Basic Generation**: Simple prompt generates image
2. **Template Rendering**: Prompt with {{.variables}} renders correctly
3. **Data Flow**: Image URI propagates to subsequent steps
4. **Error Handling**: API failures handled gracefully
5. **Storage**: Images stored and accessible via URI

### 4. Integrate with Web Builder
Update the multi-section website builder to include images:

```json
{
  "generate_hero": {
    "action": "call_agent",
    "config": {
      "agent_type": "content-creator-hero",
      "target_role": "hero_writer"
    },
    "next_step": "generate_hero_image"
  },
  "generate_hero_image": {
    "action": "generate_image",
    "description": "Generate hero image",
    "config": {
      "prompt": "Create a hero image for {{.input_data.business_name}}, a {{.input_data.business_type}}",
      "aspect_ratio": "16:9",
      "input_fields": ["generate_hero"]
    },
    "next_step": "generate_features"
  },
  "generate_features": {
    "action": "call_agent",
    "config": {
      "agent_type": "content-creator-features",
      "target_role": "features_writer"
    },
    "next_step": "aggregate_all_sections"
  },
  "aggregate_all_sections": {
    "action": "aggregate_data",
    "config": {
      "response_fields": [
        "generate_hero",
        "generate_hero_image",
        "generate_features"
      ]
    },
    "next_step": "complete"
  }
}
```

### 5. Enhanced Prompts with LLM
For better image prompts, consider creating a prompt generation step:

```json
{
  "generate_image_prompt": {
    "action": "execute_llm_prompt",
    "description": "Generate detailed image prompt",
    "config": {
      "prompt": "Based on this content: {{.generate_hero}}, create a detailed image generation prompt that captures the essence and mood.",
      "input_fields": ["generate_hero"]
    },
    "next_step": "generate_hero_image"
  },
  "generate_hero_image": {
    "action": "generate_image",
    "description": "Generate hero image",
    "config": {
      "prompt": "{{.generate_image_prompt.result}}",
      "aspect_ratio": "16:9"
    },
    "next_step": "aggregate_sections"
  }
}
```

## Design Patterns Used

### 1. Data Extraction
Uses new `data_helpers.go` functions:
```go
inputData := orchestration.ExtractDataFromMessage(params.CollectedData, params.Logger)
```

### 2. Priority-based Configuration
Searches for configuration in order:
1. Step config (`params.StepConfig.Config`)
2. Collected data (`params.CollectedData`)
3. Input data (extracted from message)
4. Input fields (gathered from previous steps)

### 3. Template Rendering
Supports Go template syntax:
```
"prompt": "Image for {{.business_name}} showing {{.business_type}}"
```

### 4. Error Handling
Uses platform error wrapping with context:
```go
return nil, errors.WrapWithAgentContext(err, "failed to generate image",
    params.ExecutionContext.Sender.AgentType,
    params.ExecutionContext.Sender.AgentID,
    params.ExecutionContext.OrchestrationID,
    params.ExecutionContext.StepName,
    params.ExecutionContext.Action)
```

## Benefits

### For Content Creators
- Can request images inline in workflows
- Images generated based on content context
- No separate service to manage

### For System
- Consistent with existing action patterns
- Reuses workflow orchestration
- Maintains data lineage (images tracked in CollectedData)

### For Users
- Automated image generation for multi-section websites
- Context-aware images (based on content)
- Seamless integration with text generation

## Potential Issues to Watch

1. **API Rate Limits**: Stability AI may have rate limits
2. **Cost**: Image generation APIs charge per image
3. **Storage Costs**: Generated images consume storage
4. **Timeout**: 90-second timeout may not be sufficient for all images
5. **Large Images**: Base64 decoding of large images may be memory-intensive

## Future Enhancements

1. **Multiple Images**: Support generating multiple images in one action
2. **Image Editing**: Support img2img, inpainting, outpainting
3. **Caching**: Cache generated images to avoid regeneration
4. **Alternative Providers**: Support DALL-E, Midjourney, local Stable Diffusion
5. **Image Optimization**: Automatically resize/optimize images
6. **CDN Integration**: Push images to CDN for faster delivery

## Testing Checklist

- [ ] Action registers successfully in coordinator
- [ ] Simple prompt generates image
- [ ] Template variables render correctly
- [ ] Image stored in object storage
- [ ] Image URI returned in result
- [ ] Result accessible in next step via CollectedData
- [ ] Error handling works (bad API key, timeout, etc.)
- [ ] Integration with web builder workflow works
- [ ] Multiple images can be generated in sequence
- [ ] Parallel image generation works (if needed)

## Files Created/Modified

### Created
- `/home/claude/image_actions.go` - Image generation action implementation

### To Modify
- `platform/orchestration/coordinator.go` - Register action
- `platform/orchestration/actions/actions.go` - Export function if needed
- Workflow JSON files - Add image generation steps

### Dependencies
Uses existing:
- `platform/orchestration/types` - ExecutionContext, etc.
- `platform/orchestration` - ExtractDataFromMessage, etc.
- `platform/storage` - S3Client
- `platform/errors` - Error wrapping
- `go.uber.org/zap` - Logging
- `github.com/google/uuid` - UUID generation