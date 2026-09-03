// FILE: platform/orchestration/actions/execute_vision_prompt_action.go
//
// ExecuteVisionPromptAction is execute_llm_prompt's vision sibling: it renders
// a prompt template against collected data, downloads the screenshots a prior
// step captured (URIs in collected data — never bytes; the screenshots.go rule
// keeps presigned URLs and image payloads out of travelling docs), and calls a
// VisionCapable provider (aiservice.VisionCapable — anthropic and gemini; the
// design-critic model trial is a config switch between them).
//
// Deliberately THIN: config resolution (resolveAIServiceConfig overlay,
// getPromptWithPriority, createAIClient), prompt rendering, JSON parsing
// (stripMarkdownFromResponse + ParseLLMJSONWithProvenance) and llm_call_log
// forensics are the SAME helpers execute_llm_prompt uses, so the two paths
// cannot drift. What it does NOT carry from its sibling, deliberately:
// tolerate_truncation and the bugs_open/119 re-ask — a truncated or unusable
// vision critique fails the step loudly in v1; add those only when a real run
// demonstrates the need, not by reflex.
//
// A provider without eyes is a CONFIGURATION error, surfaced by name — never
// papered over with a text-only call that would "critique" pages it never saw.
//
// Registration:
//   "execute_vision_prompt": {
//       Handler:     ExecuteVisionPromptAction,
//       Category:    "ai",
//       Description: "Render a prompt and send it with downloaded screenshots to a vision-capable provider",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

var ExecuteVisionPromptInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{},
	Optional:    []string{"images_field", "max_images", "max_image_dimension", "prompt_template", "output_type"},
	// ai_service is read, but not by this file: resolveAIServiceConfig
	// (ai_actions.go:61) is handed params.StepConfig.Config below and overlays a
	// step-level `ai_service` block onto the agent's. It belongs in ConfigKeys
	// rather than Optional because it never passes through ExtractActionInputs
	// and is a settings block, not a reference.
	//
	// Declared 2026-08-08 (bugs_open/136 lane): this action opted into
	// unknown-key detection without it, so the audit reported `ai_service` as an
	// unknown key on a step that genuinely uses it — a FALSE POSITIVE, and the
	// worst thing that can happen to a report people are meant to act on. It is
	// also not this action's key alone: any action calling resolveAIServiceConfig
	// reads it, so the next one to opt in must declare it too.
	ConfigKeys: []string{"ai_service"},
	Defaults: map[string]interface{}{
		"images_field": "render_audit",
		"max_images":   16,
		// Per-image long-edge cap; images over it are box-scaled down and
		// re-encoded as JPEG q85 (vision_image_downscale.go — the three
		// provider errors that forced this are quoted there). 0 disables.
		"max_image_dimension": visionImageDimensionCapDefault,
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("execute_vision_prompt", ExecuteVisionPromptInputSpec)
}

// visionImageRef is the slice of ScreenshotRef this action consumes: the
// adapter's Renders entries, read structurally (uri + profile + url).
type visionImageRef struct {
	URI     string
	Profile string
	PageURL string
}

func ExecuteVisionPromptAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "execute_vision_prompt"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.StorageClient == nil {
		return nil, fmt.Errorf("execute_vision_prompt: no storage client — cannot download screenshots")
	}

	// ── Agent + AI-service config, the sibling's exact resolution ──────────
	agentConfig, _ := params.CollectedData["agent_config"].(map[string]interface{})
	if agentConfig == nil && params.AgentType != "" && params.DB != nil {
		agentDef, err := loadAgentDefinitionForAction(ctx, params.DB, params.AgentType)
		if err != nil {
			return nil, fmt.Errorf("execute_vision_prompt: load agent definition: %w", err)
		}
		agentConfig = agentDef.DefaultConfig
		params.CollectedData["agent_config"] = agentConfig
	}
	if agentConfig == nil {
		return nil, fmt.Errorf("execute_vision_prompt: agent configuration not found")
	}

	currentStep := params.ExecutionContext.StepName
	aiServiceConfig, aiSources := resolveAIServiceConfig(agentConfig, params.StepConfig.Config, currentStep, params.Logger)
	if len(aiServiceConfig) == 0 {
		return nil, fmt.Errorf("execute_vision_prompt: ai_service configuration not found")
	}
	if err := checkOverlayRequiredKeys(aiServiceConfig); err != nil {
		return nil, err
	}
	logger.Info("execute_vision_prompt: ai_service resolved", zap.Strings("sources", aiSources))

	promptTemplate, promptSource := getPromptWithPriority(params, agentConfig)
	if promptTemplate == "" {
		return nil, fmt.Errorf("execute_vision_prompt: no prompt template (source checked: %s)", promptSource)
	}

	aiClient, err := createAIClient(ctx, aiServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("execute_vision_prompt: create AI client: %w", err)
	}
	vision, ok := aiClient.(aiservice.VisionCapable)
	if !ok {
		// Configuration error by name — the capability roster is pinned in
		// aiservice's own tests; this is the runtime half of that contract.
		return nil, fmt.Errorf("execute_vision_prompt: provider %q (model %q) is not vision-capable — use anthropic or gemini",
			aiClient.Provider(), aiClient.Model())
	}

	// ── Screenshots: resolve refs from collected data, download bytes ──────
	imagesField := datahelpers.GetStringField(params.StepConfig.Config, "images_field", "render_audit")
	maxImages := datahelpers.GetIntField(params.StepConfig.Config, "max_images", 16)
	maxImageDim := datahelpers.GetIntField(params.StepConfig.Config, "max_image_dimension", visionImageDimensionCapDefault)
	refs, err := resolveVisionImageRefs(params.CollectedData, imagesField)
	if err != nil {
		return nil, fmt.Errorf("execute_vision_prompt: %w", err)
	}
	if len(refs) == 0 {
		// No pictures is not a critique of zero findings — fail loud (the
		// absent ≠ clean rule, same as write_render_audit_findings).
		return nil, fmt.Errorf("execute_vision_prompt: no renders under %q — did the capture step run with capture_renders?", imagesField)
	}
	dropped := 0
	downscaled := 0
	if len(refs) > maxImages {
		dropped = len(refs) - maxImages
		refs = refs[:maxImages]
		logger.Warn("execute_vision_prompt: image list capped",
			zap.Int("kept", maxImages), zap.Int("dropped", dropped))
	}

	images := make([]aiservice.ImageInput, 0, len(refs))
	var manifest []map[string]interface{}
	for _, ref := range refs {
		key := storage.ExtractKeyFromS3URI(ref.URI)
		rc, dErr := params.StorageClient.Download(ctx, key)
		if dErr != nil {
			return nil, fmt.Errorf("execute_vision_prompt: download %s: %w", ref.URI, dErr)
		}
		data, rErr := io.ReadAll(rc)
		rc.Close()
		if rErr != nil {
			return nil, fmt.Errorf("execute_vision_prompt: read %s: %w", ref.URI, rErr)
		}
		mediaType := "image/png"
		if scaledData, mt, scaled := downscaleVisionImage(data, mediaType, maxImageDim, params.Logger); scaled {
			data, mediaType = scaledData, mt
			downscaled++
		}
		images = append(images, aiservice.ImageInput{MediaType: mediaType, Data: data})
		manifest = append(manifest, map[string]interface{}{
			"index": len(images), "page_url": ref.PageURL, "profile": ref.Profile,
		})
	}

	// ── Prompt: template + image manifest, so the model can cite pages ─────
	templateData := extractDataForAiAgent(params).(map[string]interface{})
	// The constant, not a literal: template_context_contract.go declares this as
	// a root this action supplies, and an offline check reads that declaration
	// (bugs_open/453). Sharing the symbol makes a rename a compile error instead
	// of a silent divergence between what is injected and what is declared.
	templateData[VisionImageManifestKey] = manifest
	renderedPrompt, err := datahelpers.RenderPromptTemplate(promptTemplate, templateData, *params.Logger)
	if err != nil {
		return nil, fmt.Errorf("execute_vision_prompt: render prompt: %w", err)
	}
	renderedPrompt = appendOutputInstructions(renderedPrompt, aiServiceConfig, params.StepConfig.Config, params.Logger)
	declaredOutputType := resolveOutputType(params.StepConfig.Config, aiServiceConfig)

	// ── Options + model resolution, mirrored for llm_call_log parity ───────
	options := make(map[string]interface{})
	var modelAlias, resolvedModel string
	if model, mok := aiServiceConfig["model"].(string); mok {
		modelAlias = model
		resolvedModel = aiservice.ResolveModelAlias(model, params.Logger)
		options["model"] = resolvedModel
	}
	if mt, mok := aiServiceConfig["max_tokens"].(float64); mok {
		options["max_tokens"] = int(mt)
	}
	if bt, bok := aiServiceConfig["budget_tokens"].(float64); bok && bt > 0 {
		options["budget_tokens"] = int(bt)
	}
	provider, _ := aiServiceConfig["provider"].(string)

	llmCallStart := time.Now()
	result, err := vision.GenerateWithImages(ctx, renderedPrompt, images, options)
	latencyMs := int(time.Since(llmCallStart).Milliseconds())

	sentMaxTokens := 0
	if mt, mok := options["__sent_max_tokens"].(int); mok {
		sentMaxTokens = mt
	}
	var sentTemperature interface{}
	if tv, tok := options["__sent_temperature"]; tok {
		sentTemperature = tv
	}

	logMsg := ""
	if err != nil {
		logMsg = err.Error()
	}
	LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
		AgentType:       params.AgentType,
		AgentID:         params.Headers["agent_id"],
		StepName:        currentStep,
		OrchestrationID: params.ExecutionContext.OrchestrationID,
		CorrelationID:   params.ExecutionContext.CorrelationID,
		Model:           modelAlias,
		ModelResolved:   resolvedModel,
		Provider:        provider,
		PromptTemplate:  promptTemplate,
		PromptRendered:  renderedPrompt,
		LatencyMs:       latencyMs,
		Success:         err == nil,
		ErrorMessage:    logMsg,
		Temperature:     sentTemperature,
		MaxTokens:       sentMaxTokens,
		Options:         options,
	})
	if err != nil {
		// Truncation included: v1 fails loud rather than tolerating a partial
		// critique (see header).
		return nil, fmt.Errorf("execute_vision_prompt: vision call failed: %w", err)
	}

	cleaned := stripMarkdownFromResponse(result)
	parsed, provenance, parseErr := ParseLLMJSONWithProvenance(cleaned)
	if parseErr != nil {
		out := map[string]interface{}{
			"result":            cleaned,
			"type":              "text",
			"images_sent":       len(images),
			"images_dropped":    dropped,
			"images_downscaled": downscaled,
		}
		markJSONContractUnmet(out, declaredOutputType)
		return out, nil
	}
	if provenance != ProvenanceClean {
		logger.Warn("execute_vision_prompt: JSON recovered from a non-bare response",
			zap.String("provenance", provenance))
	}
	out := map[string]interface{}{
		"result":            parsed,
		"type":              "json",
		"images_sent":       len(images),
		"images_dropped":    dropped,
		"images_downscaled": downscaled,
	}
	markEnvelopeRecovered(out, provenance)
	return out, nil
}

// resolveVisionImageRefs pulls ScreenshotRef-shaped entries out of collected
// data. field names either the renders array directly, or an object holding it
// under .renders, or the coordinator's awaited-response wrapper (.response),
// so the step composes with request_render_audit's output_field unmodified.
func resolveVisionImageRefs(collected map[string]interface{}, field string) ([]visionImageRef, error) {
	node := datahelpers.ExtractNestedField(collected, field)
	if node == nil {
		return nil, fmt.Errorf("nothing under images_field %q", field)
	}
	// Descend: object → .response → .renders
	if m, ok := node.(map[string]interface{}); ok {
		if inner, ok2 := m["response"].(map[string]interface{}); ok2 {
			m = inner
		}
		if r, ok2 := m["renders"]; ok2 {
			node = r
		}
	}
	arr, ok := node.([]interface{})
	if !ok {
		return nil, fmt.Errorf("images_field %q does not resolve to a renders array", field)
	}
	var refs []visionImageRef
	for _, it := range arr {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		uri, _ := m["uri"].(string)
		if uri == "" {
			continue
		}
		profile, _ := m["profile"].(string)
		pageURL, _ := m["url"].(string)
		refs = append(refs, visionImageRef{URI: uri, Profile: profile, PageURL: pageURL})
	}
	return refs, nil
}
