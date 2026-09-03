// FILE: platform/orchestration/actions/ai_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gqls/agentchassis/platform/voicestyle"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Fake response
func ExecuteLLMPromptActionFAKE(ctx context.Context, params ActionParams) (interface{}, error) {
	// Fake parsed result to return instead of calling the AI
	parsedResult := map[string]interface{}{
		"summary": "This is a fake response for testing.",
		"status":  "success",
		"results": map[string]interface{}{
			"business_name": "Test Company",
			"domain":        "test.com",
			"description":   "A test company",
			"analysis":      "THIS IS PLACEHOLDER ANALYSIS DATA.",
		},
	}

	return map[string]interface{}{
		"result": parsedResult,
		"type":   "json",
	}, nil

}

// resolveAIServiceConfig builds the effective ai_service block for a step.
// Precedence, least to most specific, later wins PER KEY: the agent's root
// ai_service block (fleet default: provider, api_key_env_var, model), the
// workflow step's config.ai_service block, then a runtime StepConfig block.
// A block contributes only the keys it declares, so a step that sets only
// max_tokens still inherits the root provider/model. The returned sources
// slice names the blocks that contributed, in overlay order.
//
// The per-key overlay itself reuses datahelpers.MergeInputData (the platform's
// standard "copy existing, overlay new per key" map merge) rather than a
// bespoke loop — this function only LOCATES the three blocks and records which
// contributed. A shallow overlay is the correct config-precedence semantic: a
// step declaring a key replaces root's value for that key wholesale. The
// deep-merge variants (siteSpecDeepMerge, deepMergeNewsFeed) would instead
// blend nested maps, which is wrong for precedence; and no ai_service key holds
// a nested value in any live agent_definitions row (checked 2026-07-20), so the
// shallow copy cannot alias a shared sub-map either.
//
// This replaces first-found-wins, under which the ENTIRE step block was dead
// config whenever a root block existed — per-step max_tokens overrides
// silently ran at the root (or hardcoded 2048) value (bugs_open/009).
func resolveAIServiceConfig(agentConfig, runtimeStepConfig map[string]interface{}, currentStep string, logger *zap.Logger) (map[string]interface{}, []string) {
	merged := make(map[string]interface{})
	var sources []string

	overlay := func(block map[string]interface{}, source string) {
		if len(block) == 0 {
			return
		}
		merged = datahelpers.MergeInputData(merged, block, logger)
		sources = append(sources, source)
	}

	if agentConfig != nil {
		if root, ok := agentConfig["ai_service"].(map[string]interface{}); ok {
			overlay(root, "root")
		}
		if workflow, ok := agentConfig["workflow"].(map[string]interface{}); ok {
			if steps, ok := workflow["steps"].(map[string]interface{}); ok {
				if stepDef, ok := steps[currentStep].(map[string]interface{}); ok {
					if stepConfig, ok := stepDef["config"].(map[string]interface{}); ok {
						if block, ok := stepConfig["ai_service"].(map[string]interface{}); ok {
							overlay(block, "workflow_step")
						}
					}
				}
			}
		}
	}
	if runtimeStepConfig != nil {
		if block, ok := runtimeStepConfig["ai_service"].(map[string]interface{}); ok {
			overlay(block, "step_config")
		}
	}

	return merged, sources
}

// checkOverlayRequiredKeys fails loud when a required ai_service key is present
// in the effective config but empty — the shape an overlay creates if a more-
// specific block declares e.g. `"provider": ""`, silently blanking a good root
// default (bugs_open/009 council round 2). It fires ONLY on present-but-empty:
// an absent key is not an error here (it keeps its normal downstream default),
// so a legitimately partial block is unaffected.
func checkOverlayRequiredKeys(cfg map[string]interface{}) error {
	for _, key := range []string{"provider", "model", "api_key_env_var"} {
		v, present := cfg[key]
		if !present {
			continue
		}
		s, isStr := v.(string)
		if !isStr || strings.TrimSpace(s) == "" {
			return fmt.Errorf("ai_service overlay produced an empty %q — a step or runtime block cleared it; declare it with a value or omit it to inherit the root default", key)
		}
	}
	return nil
}

func ExecuteLLMPromptAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing LLM prompt action",
		zap.String("agent_type", params.AgentType),
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
		zap.String("action", params.ExecutionContext.Action),
		zap.Bool("has_db", params.DB != nil),
		zap.Any("DEBUGaa: the full params in ExecuteLLMPromptAction", params),
	)

	// initialise
	if params.ExecutionContext.Action == "initialize" {
		params.Logger.Info("handling initialization")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Method 2: Check for initialization flag in collected data
	if isInit, ok := params.CollectedData["is_initialization"].(bool); ok && isInit {
		params.Logger.Info("initialization detected via flag")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Method 3: Check the action from collected data (if passed through)
	if action, ok := params.CollectedData["action"].(string); ok && action == "initialize" {
		params.Logger.Info("initialization detected via action field")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	/*	// Normalize the collected data using the helper
		normalizedData := datahelpers.NormalizeCollectedData(
			params.CollectedData,
			params.ExecutionContext,
			params.ExecutionContext.RequestsTopic,
			params.Logger,
		)

		// Update params.CollectedData with normalized version
		params.CollectedData = normalizedData*/

	// removed NormalizeCollectedData was destroying accumulated state (loop vars, render_context, etc.)
	// The params.CollectedData already has the correct state from the coordinator.
	// Only ensure essential topic fields are present if missing:
	if params.ExecutionContext.RequestsTopic != "" {
		if _, exists := params.CollectedData["__my_requests_topic__"]; !exists {
			params.CollectedData["__my_requests_topic__"] = params.ExecutionContext.RequestsTopic
		}
	}
	if params.ExecutionContext.ReplyToTopic != "" {
		if _, exists := params.CollectedData["__parent_responses_topic__"]; !exists {
			params.CollectedData["__parent_responses_topic__"] = params.ExecutionContext.ReplyToTopic
		}
	}

	params.Logger.Info("ExecuteLLMPromptAction: preserving accumulated CollectedData",
		zap.Strings("keys", datahelpers.GetMapKeys(params.CollectedData)),
		zap.Bool("has_current_section", params.CollectedData["current_section"] != nil),
	)

	// Get the agent's configuration
	var agentConfig map[string]interface{}
	var ok bool

	// First try CollectedData (for orchestrated agents)
	agentConfig, ok = params.CollectedData["agent_config"].(map[string]interface{})
	params.Logger.Info("Checking agent_config in CollectedData",
		zap.Bool("found", ok),
		zap.Bool("is_nil", agentConfig == nil),
		zap.Any("agentConfig in first try", agentConfig),
		zap.Any("Config in first try", agentConfig),
		zap.String("step name is", params.ExecutionContext.StepName),
		//zap.Any("DEBUGaa: collected Data at this stage is:", params.CollectedData),
	)

	currentStep := params.ExecutionContext.StepName

	// If not found, load it directly from the database
	if !ok && params.AgentType != "" {
		params.Logger.Info("Agent config not in collected data, loading from database",
			zap.String("agent_type", params.AgentType))

		agentDef, err := loadAgentDefinitionForAction(ctx, params.DB, params.AgentType)
		if err != nil {
			// Defensive fallback (bugs_open/060): now that owner_agent_type (→
			// params.AgentType) records the REAL resolved agent type rather than
			// always 'generic', an abnormal value (a legacy group name, or a type
			// with no active row) could fail this load where 'generic' used to
			// succeed. Fall back to 'generic' rather than hard-failing the action —
			// the config we actually need may still arrive via CollectedData below.
			params.Logger.Warn("Failed to load agent definition by owner type; falling back to generic",
				zap.String("agent_type", params.AgentType),
				zap.Error(err))
			if params.AgentType != "generic" {
				if genDef, gerr := loadAgentDefinitionForAction(ctx, params.DB, "generic"); gerr == nil {
					agentDef = genDef
					err = nil
				}
			}
			if err != nil {
				return nil, fmt.Errorf("failed to load agent definition: %w", err)
			}
		}

		agentConfig = agentDef.DefaultConfig
		params.CollectedData["agent_config"] = agentConfig
	}

	if agentConfig == nil {
		params.Logger.Error("Agent configuration is nil after all attempts")
		return nil, fmt.Errorf("agent configuration not found")
	}

	config, ok := params.CollectedData["config"].(map[string]interface{})
	if !ok {
		params.Logger.Warn("Failed to load normal config (overridden config)",
			zap.String("agent_type", params.AgentType),
		)
	}

	params.Logger.Info("Checking normal just config in CollectedData",
		zap.Any("config", config),
	)

	// Build the effective AI service configuration by overlay (bugs_open/009):
	// the root block is the fleet default, the current step's block overrides
	// it key-by-key, and a runtime StepConfig block overrides both. First-found
	// -wins made the step block dead config whenever a root block existed.
	aiServiceConfig, aiServiceSources := resolveAIServiceConfig(agentConfig, params.StepConfig.Config, currentStep, params.Logger)

	if len(aiServiceConfig) == 0 {
		params.Logger.Error("ai_service configuration not found after checking all locations",
			zap.String("checked_locations", "agent_config top-level, workflow.steps.config, StepConfig"))
		return nil, fmt.Errorf("ai_service configuration not found")
	}

	// Overlay guard (bugs_open/009 council round 2, bug_historian): the len==0
	// check above only catches TOTAL absence. A more-specific block that declares
	// a required key as an empty string would overlay a good root default with
	// nothing — the platform's recurring "present-but-empty passes silently"
	// shape. Fail loud HERE, naming the key, rather than downstream with an
	// obscure message. Only present-but-empty is an error; an absent key is left
	// to its normal default, so this changes no behaviour for legitimate configs
	// (createAIClient already rejects an empty provider, but model and
	// api_key_env_var were unguarded, and an early named error beats a late one).
	if err := checkOverlayRequiredKeys(aiServiceConfig); err != nil {
		params.Logger.Error("ai_service: required key emptied by overlay",
			zap.Error(err), zap.Strings("sources", aiServiceSources), zap.String("step", currentStep))
		return nil, err
	}

	if len(aiServiceSources) > 1 {
		params.Logger.Info("ai_service: step overlay applied",
			zap.Strings("sources", aiServiceSources),
			zap.String("step", currentStep))
	} else {
		params.Logger.Info("ai_service: single source",
			zap.Strings("sources", aiServiceSources),
			zap.String("step", currentStep))
	}

	// Extract prompt template
	// THREE-TIER PRIORITY:
	// Get prompt using three-tier priority system
	promptTemplate, promptSource := getPromptWithPriority(params, agentConfig)

	params.Logger.Info("Selected prompt for execution",
		zap.String("source", promptSource),
		zap.String("agent_type", params.AgentType),
		zap.String("prompt_preview", datahelpers.TruncateString(promptTemplate, 350)))

	// Create AI client based on provider
	aiClient, err := createAIClient(ctx, aiServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	params.Logger.Info("in ExecuteLLMPromptAction data from which were trying to extract templatedata",
		zap.Any("DEBUGaa: params sent to extractDataForAIAgent hoping for the correct cleandata", params),
	)

	extractedData := extractDataForAiAgent(params)
	templateData := extractedData.(map[string]interface{})

	// The house voice, injected for EVERY prompt template rather than pasted
	// into each one. A template opts in by writing {{.voice_style}}; one that
	// does not mention it is unaffected. Single source: bugs_open/121.
	//
	// Deliberately does not overwrite a value the step already supplied — a
	// caller that computed its own voice_style outranks the platform default,
	// which is the request-level override the owner asked for.
	injectPlatformBlocks(ctx, templateData, func(ctx context.Context, name string) (string, bool) {
		return voicestyle.GetBlock(ctx, name, func(ctx context.Context) (string, error) {
			if params.DB == nil {
				return "", sql.ErrConnDone
			}
			var t sql.NullString
			err := params.DB.QueryRowContext(ctx, voicestyle.SQLByName, name).Scan(&t)
			return t.String, err
		})
	})

	validateTemplateData(templateData, params.StepConfig.Config, params.Logger)

	params.Logger.Info("in ExecuteLLMPromptAction Template Data",
		// zap.Any("template_data DEBUGaa", templateData),
		zap.Any("DEBUGaa template_data ai_actions this is what I want to pass - should be good ", templateData), // is good
		zap.Any("agent_config", agentConfig),
		//zap.Any("DEBUGaa: params.CollectedData[input_data] when extracting data in ai actions", params.CollectedData["input_data"]), // neither data or template is not in here
		zap.Any("promptTemplate", promptTemplate), // good
	)

	// Render the prompt template
	renderedPrompt, err := datahelpers.RenderPromptTemplate(promptTemplate, templateData, *params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt template: %w", err)
	}

	params.Logger.Info("Rendered prompt template",
		zap.String("template_preview", datahelpers.TruncateString(promptTemplate, 300)),
		zap.String("rendered_preview - renderedPrompt", datahelpers.TruncateString(renderedPrompt, 400)))

	// Append output format instructions based on the declared output type
	// Check both step config and ai_service config
	renderedPrompt = appendOutputInstructions(renderedPrompt, aiServiceConfig, params.StepConfig.Config, params.Logger)

	// What this step SAID it must produce, resolved by the same single rule the
	// instructions above use. Captured here so the parse-failure path can tell a
	// step that needs JSON and did not get it from one that is happy with prose
	// (bugs_open/119).
	declaredOutputType := resolveOutputType(params.StepConfig.Config, aiServiceConfig)

	// Prepare AI service options
	options := make(map[string]interface{})
	if model, ok := aiServiceConfig["model"].(string); ok {
		options["model"] = model
	}
	if temp, ok := agentConfig["temperature"].(float64); ok {
		options["temperature"] = temp
	}

	var maxTokens float64
	if maxTokens, ok = agentConfig["max_tokens"].(float64); ok {
		options["max_tokens"] = int(maxTokens)
	} else if maxTokens, ok = aiServiceConfig["max_tokens"].(float64); ok {
		options["max_tokens"] = int(maxTokens)
	} else {
		// No cap at any level: the provider client's hardcoded fallback will
		// apply (anthropic.go: 2048 — the smallest number in the estate). That
		// is a transport safety net, not a sizing decision; a step reaching it
		// means nobody chose this step's output budget, and the first oversized
		// document meets a silent cliff (bugs_open/205: 8 of 126 active LLM
		// steps ran unconfigured, 64 truncations before anything said so).
		params.Logger.Warn("max_tokens not configured at any level; provider hardcoded default will apply",
			zap.String("agent_type", params.AgentType),
			zap.String("step", params.ExecutionContext.StepName))
	}

	// Pass through budget_tokens for extended thinking
	// Config: "ai_service": {"budget_tokens": 10000}
	if budgetTokens, ok := aiServiceConfig["budget_tokens"].(float64); ok && budgetTokens > 0 {
		options["budget_tokens"] = int(budgetTokens)
	}

	// Resolve model alias to actual API model name
	var modelAlias string // preserve original alias for logging
	var resolvedModel string
	if model, ok := options["model"].(string); ok {
		modelAlias = model
		resolvedModel = aiservice.ResolveModelAlias(model, params.Logger)
		options["model"] = resolvedModel
	}

	// Track provider for logging
	provider, _ := aiServiceConfig["provider"].(string)

	// ── Extract flywheel context for training data ─────────────────────
	// These fields link LLM calls to work item outcomes and industry verticals,
	// enabling training data export filtered by quality and vertical.
	flywheelWorkItemID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.work_item_id")
	flywheelVertical := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_specs.specs.identity.industry")
	if flywheelVertical == "" {
		flywheelVertical = datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.industry")
	}
	if flywheelVertical == "" {
		flywheelVertical = datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.vertical")
	}
	flywheelRAG := datahelpers.ExtractNestedField(params.CollectedData, "rag_results") != nil

	// Start timing the LLM call
	llmCallStart := time.Now()

	// Set when a truncated response is tolerated per step config (bugs_open/019),
	// so the returned result can be MARKED partial for the consumer.
	truncationTolerated := false
	truncatedTokens := 0

	// Call the AI service
	result, err := aiClient.GenerateText(ctx, renderedPrompt, options)

	// ── One escalated retry when the output cap cut the response (bugs_open/337) ──
	// Decision table + contract in truncation_escalation.go. This runs BEFORE
	// the error block below, so that block only ever sees the FINAL attempt's
	// error: if the escalated call also fails — truncated at the ceiling, a
	// 5xx, anything — everything downstream (tolerate_truncation and its
	// bugs_open/076 guard, isAIUnavailable, the transient ladder, the
	// catch-all) applies to it verbatim, with no second code path to drift.
	if err != nil {
		// The `ok` is carried, not discarded: an absent key must not read as a
		// 0 baseline, which would make the ceiling>cap refusal vacuous
		// (council round 1, editquality). See truncationEscalationApplies.
		sentCap, sentCapKnown := options["__sent_max_tokens"].(int)
		if ceiling, escalate := truncationEscalationApplies(err, aiServiceConfig, sentCap, sentCapKnown); escalate {
			// One call, one forensic row: the cut first call is recorded
			// success=false, prefixed so a census can count escalations without
			// joining anything. Logged before options is reused — LogLLMCall
			// reads the map synchronously, so the second call's telemetry
			// cannot race this row's.
			LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
				AgentType:       params.AgentType,
				AgentID:         params.Headers["agent_id"],
				StepName:        params.ExecutionContext.StepName,
				OrchestrationID: params.ExecutionContext.OrchestrationID,
				CorrelationID:   params.ExecutionContext.CorrelationID,
				Model:           modelAlias,
				ModelResolved:   resolvedModel,
				Provider:        provider,
				PromptTemplate:  promptTemplate,
				PromptRendered:  renderedPrompt,
				LatencyMs:       int(time.Since(llmCallStart).Milliseconds()),
				Success:         false,
				ErrorMessage: fmt.Sprintf(
					"ESCALATED (bugs_open/337: cap %d cut the response; retrying ONCE at max_tokens_ceiling=%d): %s",
					sentCap, ceiling, err.Error()),
				Temperature:    options["__sent_temperature"],
				MaxTokens:      sentCap,
				WorkItemID:     flywheelWorkItemID,
				Vertical:       flywheelVertical,
				RAGContextUsed: flywheelRAG,
				Options:        options,
			})
			params.Logger.Warn("LLM response truncated at the configured cap — retrying ONCE at max_tokens_ceiling",
				zap.String("step_name", params.ExecutionContext.StepName),
				zap.String("agent_type", params.AgentType),
				zap.Int("sent_max_tokens", sentCap),
				zap.Int("max_tokens_ceiling", ceiling))
			options["max_tokens"] = ceiling
			llmCallStart = time.Now()
			result, err = aiClient.GenerateText(ctx, renderedPrompt, options)
		}
	}

	if err != nil {
		llmLatencyMs := int(time.Since(llmCallStart).Milliseconds())
		params.Logger.Info("AI call failed once",
			zap.Error(err),
		)

		// Capture what the provider actually sent, for llm_call_log
		var sentTemperature interface{}
		if t, ok := options["__sent_temperature"]; ok {
			sentTemperature = t
		}
		sentMaxTokens := 0
		if mt, ok := options["__sent_max_tokens"].(int); ok {
			sentMaxTokens = mt
		}

		// Detect a tolerated truncation BEFORE logging, so the one forensic row
		// this call gets says whether the chain continued past it. Without the
		// prefix, a tolerated cut and a fatal one are indistinguishable in
		// llm_call_log, and per-step success-rate queries misread tolerated
		// calls as pure failures (council round 2eed453a, three seats).
		truncErr, isTruncatedCall := aiservice.IsTruncated(err)
		tolerateTruncation := datahelpers.GetBoolField(params.StepConfig.Config, "tolerate_truncation", false)

		// The guard's verdict is resolved HERE, before the log line, because the
		// prefix is a claim about what happened next and opting in is no longer
		// enough to make it true (bugs_open/076). Deciding the prefix from config
		// alone wrote "step continued on the partial" onto calls the guard had
		// just refused — the exact inverse of the misreading the prefix exists to
		// prevent, and caught by the induced-fault run that verified the guard
		// rather than by any test. Computed once and reused at the tolerate
		// branch below, so the log and the behaviour cannot disagree.
		guardStep, guardedConsumer := "", false
		if isTruncatedCall && tolerateTruncation {
			guardStep, guardedConsumer = findTruncationAwareConsumer(
				params.WorkflowSteps, params.CurrentStep, params.ExecutionContext.StepName)
		}

		logErrMsg := err.Error()
		if isTruncatedCall && tolerateTruncation {
			if guardedConsumer {
				logErrMsg = "TOLERATED (step continued on the partial): " + logErrMsg
			} else {
				logErrMsg = "REFUSED (bugs_open/076: tolerate_truncation set, but no step in this workflow reads the __truncated marker): " + logErrMsg
			}
		}

		// Options is handed over on the FAILURE path too, and deliberately.
		// gemini.go records usage into the map BEFORE it inspects finishReason,
		// so a call cut short by thinking still carries the thinking count that
		// explains why — the single most useful row in the table for sizing the
		// reserve (bugs_open/107 was diagnosed as an incapable model precisely
		// because this number was invisible). A failed row whose thinking_tokens
		// is close to wire_max_output_tokens is the starvation signature,
		// readable without any arithmetic on output_tokens.

		// Log the failed call
		LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
			AgentType:       params.AgentType,
			AgentID:         params.Headers["agent_id"],
			StepName:        params.ExecutionContext.StepName,
			OrchestrationID: params.ExecutionContext.OrchestrationID,
			CorrelationID:   params.ExecutionContext.CorrelationID,
			Model:           modelAlias,
			ModelResolved:   resolvedModel,
			Provider:        provider,
			PromptTemplate:  promptTemplate,
			PromptRendered:  renderedPrompt,
			LatencyMs:       llmLatencyMs,
			Success:         false,
			ErrorMessage:    logErrMsg,
			Temperature:     sentTemperature,
			MaxTokens:       sentMaxTokens,
			WorkItemID:      flywheelWorkItemID,
			Vertical:        flywheelVertical,
			RAGContextUsed:  flywheelRAG,
			Options:         options,
		})

		// ── Tolerated truncation: degrade the STEP, don't void the RUN ──
		// bugs_open/019. A step whose output cap is reached fails, its error_step
		// routes the whole workflow to a terminal state, and every sibling step's
		// completed work is discarded. For a council that is catastrophic and
		// routine at once: all 13 seats route error_step->complete_invalid and run
		// sequentially, so the FIRST seat to overrun ends the round before any
		// verdict exists — 9 rounds lost in 10 days, every one of them paid for.
		//
		// Opt-in per step (tolerate_truncation: true), never a default: turning a
		// hard error into a usable result is only safe where the consumer knows the
		// output is partial and treats it accordingly. The council decider does
		// (it degrades an unreadable seat to a loud abstention and refuses to let
		// an approve stand alongside one); a page renderer emphatically does not,
		// which is why the catch-all below must keep failing loud for everyone else.
		//
		// The partial then flows through the SAME parse/repair path as any other
		// response — no bypass — so it gets ParseLLMJSON's repair attempt, and the
		// llm_call_log row above stays success=false, keeping the forensic trail
		// the case file's own diagnostic queries rely on.
		if isTruncatedCall && tolerateTruncation {
			// The marker is only HALF the contract, and until bugs_open/076 the
			// other half was enforced by nothing: tolerating is safe only where a
			// consumer reads __truncated, yet any step could set
			// tolerate_truncation: true and be believed. Check the workflow
			// actually contains a reader before keeping the fragment, so a new
			// call site is safe by omission rather than by vigilance — the same
			// inversion that already makes tolerance itself opt-in.
			//
			// Refusing here costs the run the partial and returns the loud
			// failure the step would have had before it opted in. That is the
			// correct trade: an unguarded consumer does not fail, it succeeds
			// with a half-answer, which is the strictly worse bug this whole
			// mechanism exists to avoid.
			if !guardedConsumer {
				params.Logger.Error("LLM response was TRUNCATED and the step sets tolerate_truncation, but NO step in this workflow reads the __truncated marker — refusing to tolerate (bugs_open/076)",
					zap.String("step_name", params.ExecutionContext.StepName),
					zap.String("agent_type", params.AgentType),
					zap.Int("output_tokens", truncErr.OutputTokens),
					zap.Int("workflow_steps", len(params.WorkflowSteps)),
				)
				return nil, fmt.Errorf("step %q sets tolerate_truncation but no step in this workflow consumes the __truncated marker, so a partial response would be indistinguishable from a complete one (bugs_open/076): raise max_tokens, drop tolerate_truncation, or give the consuming step accepts_truncated: true once it handles a partial: %w",
					params.ExecutionContext.StepName, err)
			}

			truncationTolerated = true
			truncatedTokens = truncErr.OutputTokens
			result = truncErr.Partial

			params.Logger.Warn("LLM response was TRUNCATED — tolerating per step config; downstream must treat this result as partial",
				zap.String("step_name", params.ExecutionContext.StepName),
				zap.String("guarded_by", guardStep),
				zap.String("reason", truncErr.Reason),
				zap.Int("output_tokens", truncErr.OutputTokens),
				zap.Int("partial_chars", len(truncErr.Partial)),
				zap.Int("sent_max_tokens", sentMaxTokens),
			)

			err = nil // handled: fall through to the normal result path
		}

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		// ── Back-to-triage: fast-fail on infrastructure unavailability ──
		// Connection refused, DNS failure, timeout, credit exhaustion (401/402).
		// These won't be fixed by retrying. Return AIUnavailableError so the
		// coordinator releases the item back to triaged without counting attempts.
		if isAIUnavailable(err) {
			params.Logger.Warn("AI endpoint unavailable — item will be released back to queue",
				zap.String("provider", provider),
				zap.String("model", resolvedModel),
				zap.Error(err))

			// Reactively update health table (especially for Claude credit failures)
			if params.DB != nil {
				apiURL, _ := aiServiceConfig["api_url"].(string)
				if apiURL == "" && provider == "anthropic" {
					apiURL = "https://api.anthropic.com/v1/messages"
				}
				if apiURL != "" {
					go func() {
						rctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
						defer cancel()
						params.DB.ExecContext(rctx,
							`SELECT update_endpoint_health($1, false, $2)`,
							apiURL, err.Error())
					}()
				}
			}

			return nil, &AIUnavailableError{
				Provider: provider,
				Model:    resolvedModel,
				Endpoint: fmt.Sprintf("%v", aiServiceConfig["api_url"]),
				Cause:    err,
			}
		}

		// Check for model-related errors
		if strings.Contains(errStr, "model") || strings.Contains(errStr, "404") || strings.Contains(errStr, "not found") {
			modelUsed := fmt.Sprintf("%v", options["model"])

			params.Logger.Error("Model error - possibly invalid model name",
				zap.String("model_used", modelUsed),
				zap.Strings("available_aliases", aiservice.GetAvailableAliases()),
				zap.Error(err),
			)

			return nil, fmt.Errorf("model '%s' not found. Use aliases like: claude-sonnet-4-6, claude-opus-4-6, claude-haiku-4-5. Error: %w",
				modelUsed, err)
		}

		if strings.Contains(errStr, "529") || // overloaded
			strings.Contains(errStr, "503") || // service unavailable
			strings.Contains(errStr, "502") || // bad gateway
			strings.Contains(errStr, "500") { // internal server error

			retryDelays := []time.Duration{5 * time.Second, 15 * time.Second, 30 * time.Second, 60 * time.Second}
			var lastErr error
			for attempt, delay := range retryDelays {
				params.Logger.Warn("LLM call failed with transient error, retrying",
					zap.Int("attempt", attempt+1),
					zap.Int("max_attempts", len(retryDelays)),
					zap.Duration("backoff", delay),
					zap.String("error_type", errStr[:min(len(errStr), 80)]),
				)
				time.Sleep(delay)
				result, lastErr = aiClient.GenerateText(ctx, renderedPrompt, options)
				if lastErr == nil {
					params.Logger.Info("LLM retry succeeded",
						zap.Int("attempt", attempt+1))
					break
				}
			}
			if lastErr != nil {
				return nil, fmt.Errorf("AI call failed after %d retries. Last error: %w", len(retryDelays), lastErr)
			}
			err = nil // Clear the original error since retry succeeded
		}

		// Catch-all: any error not handled above (model check or transient retry)
		// is a hard failure. This catches genuinely malformed requests (invalid
		// parameters, context window exceeded) and other unexpected status codes.
		// Without this, the error is silently dropped and the pipeline continues
		// with an empty result — which can overwrite good page content with empty shells.
		if err != nil {
			params.Logger.Error("AI call failed with unhandled error — failing workflow",
				zap.Error(err),
			)
			return nil, fmt.Errorf("AI call failed with unhandled error: %w", err)
		}
	}

	params.Logger.Info("LLM response received",
		zap.String("result_preview", datahelpers.TruncateString(result, 200)))

	// ── Endpoint health: the SYMMETRIC half of the failure write above ──
	// bugs_open/243. Until this existed, `update_endpoint_health` had exactly one
	// caller — the isAIUnavailable branch above — and it only ever passed FALSE.
	// The sole writer of `true` was the hourly prober. So live traffic could mark
	// the endpoint down but nothing succeeding could mark it up, and
	// claim_work_item (claim_work_item_action.go, the AI-endpoint-health gate)
	// gates EVERY work-item claim fleet-wide on that one boolean. One refused call
	// therefore stopped all dispatch until the prober's next tick — measured at
	// 60m25s on 2026-08-17 while 93 of 99 live calls in the same window succeeded.
	//
	// The prober cannot close this itself: pingClaude returns healthy for ANY
	// non-auth status (check_endpoint_health_action.go), and a usage cap is a 400,
	// so for this condition the prober is a timer, not a health check. A real
	// successful call is the only signal that actually observes recovery.
	//
	// `AND NOT healthy` is load-bearing, not an optimisation: it makes this a
	// zero-row no-op on the primary key in the normal case, and it stops
	// last_healthy being rewritten on every call in the fleet.
	if params.DB != nil {
		apiURL, _ := aiServiceConfig["api_url"].(string)
		if apiURL == "" && provider == "anthropic" {
			apiURL = "https://api.anthropic.com/v1/messages"
		}
		if apiURL != "" {
			go func() {
				rctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				if _, uerr := params.DB.ExecContext(rctx, `
					UPDATE ai_endpoint_health
					   SET healthy = true, last_healthy = NOW(), error = NULL, updated_at = NOW()
					 WHERE endpoint_url = $1 AND NOT healthy`, apiURL); uerr != nil {
					params.Logger.Warn("failed to clear endpoint health after a successful call",
						zap.String("endpoint", apiURL), zap.Error(uerr))
				}
			}()
		}
	}

	// Log the successful LLM call
	llmLatencyMs := int(time.Since(llmCallStart).Milliseconds())
	inputTokens := 0
	outputTokens := 0
	if it, ok := options["__usage_input_tokens"].(int); ok {
		inputTokens = it
	}
	if ot, ok := options["__usage_output_tokens"].(int); ok {
		outputTokens = ot
	}
	var sentTemperature interface{}
	if t, ok := options["__sent_temperature"]; ok {
		sentTemperature = t
	}
	sentMaxTokens := 0
	if mt, ok := options["__sent_max_tokens"].(int); ok {
		sentMaxTokens = mt
	}
	// A TOLERATED truncation was already logged in the error path above
	// (success=false, stop_reason in error_message) — do not log it a second
	// time as a success. A success=true row with output_tokens at exactly the
	// cap is the pre-bugs_open/008 signature of a silent cut reading as a
	// finished completion, and it would poison the headroom queries the 019
	// case file documents (`success=t` rows are where the margins are measured).
	// One call, one row: the failed row IS this call's record.
	if !truncationTolerated {
		LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
			AgentType:       params.AgentType,
			AgentID:         params.Headers["agent_id"],
			StepName:        params.ExecutionContext.StepName,
			OrchestrationID: params.ExecutionContext.OrchestrationID,
			CorrelationID:   params.ExecutionContext.CorrelationID,
			Model:           modelAlias,
			ModelResolved:   resolvedModel,
			Provider:        provider,
			PromptTemplate:  promptTemplate,
			PromptRendered:  renderedPrompt,
			ResponseText:    result,
			InputTokens:     inputTokens,
			OutputTokens:    outputTokens,
			LatencyMs:       llmLatencyMs,
			Success:         true,
			Temperature:     sentTemperature,
			MaxTokens:       sentMaxTokens,
			WorkItemID:      flywheelWorkItemID,
			Vertical:        flywheelVertical,
			RAGContextUsed:  flywheelRAG,
			Options:         options,
		})
	}

	// Strip markdown code blocks from response before processing
	cleanedResult := stripMarkdownFromResponse(result)

	// Parse as JSON, repairing the escaping-only malformations models emit
	// constantly (raw newlines / unescaped quotes inside string values). A
	// TRUNCATED response — the model hit its output-token ceiling mid-object —
	// is genuinely incomplete and stays unparseable: it falls through to the
	// text path here, and the renderer's required-field check then refuses to
	// ship a blank section rather than silently emptying it. See
	// json_envelope.go and HANDOFF_2026-07-14_article_body_json_envelope.md.
	parsedResult, provenance, parseErr := ParseLLMJSONWithProvenance(cleanedResult)

	// ── One corrective re-ask when a JSON-declaring step got unusable output ──
	// bugs_open/119. Until this existed the platform retried a call that never
	// ARRIVED (the 500/502/503/529 ladder ~100 lines above, four attempts with
	// backoff) and accepted without question a call that arrived UNUSABLE. That
	// asymmetry is the whole defect: for a step that declared it needs JSON, a
	// response that will not parse is not a lesser answer, it is no answer.
	//
	// What it costs the council, measured 2026-08-01 over all 424 rounds in
	// diagnosis_artifacts: 23 were decided by "unreadable reviewer(s)" rather
	// than by any judgement — 15 of them in the last week alone, across SEVEN
	// different seats. One submission (correlation c5219a69) burned three
	// consecutive rounds to the same seat in ten minutes, because nothing
	// re-asked and every retry was a fresh round of 10-13 seats.
	//
	// Scope is deliberately narrow, so this cannot become a general "retry
	// everything" that doubles fleet cost:
	//   - only when the step DECLARED output_format/output_type json, so a step
	//     that legitimately returns prose is untouched;
	//   - only when the answer failed to parse, i.e. only on a path that is
	//     already producing a result the consumer cannot use;
	//   - exactly ONE extra attempt, never a ladder.
	// Measured cost of that scope: across 785 JSON-declared step outputs in the
	// live orchestration window, 2 failed to parse. This fires twice, not 785
	// times.
	//
	// The re-ask is CORRECTIVE rather than identical, because an identical
	// re-ask reproduces the failure — c5219a69 proved that three times running.
	// Note what it deliberately does NOT do: raise max_tokens. That is not an
	// oversight, it is aiservice/truncation.go:26-29's stated position —
	// "whatever the number, the step that writes most approaches it on the work
	// most worth doing" — so the retry asks for the same judgement SHORTER
	// instead of buying headroom that the next long review will eat anyway.
	//
	// If the re-ask also fails, every line below runs exactly as it did before
	// this block existed. bugs_open/019's downgrade contract and 138's salvage
	// path are untouched: an unreadable seat still blocks an approval, it just
	// gets one chance to be readable first.
	if parseErr != nil && declaredOutputType == "json" {
		params.Logger.Warn("LLM response would not parse and this step declared output json — re-asking ONCE before giving up (bugs_open/119)",
			zap.String("step_name", params.ExecutionContext.StepName),
			zap.String("agent_type", params.AgentType),
			zap.Bool("first_attempt_truncated", truncationTolerated),
			zap.Int("response_len", len(cleanedResult)),
			zap.Error(parseErr),
		)

		reaskPrompt := correctiveReaskPrompt(renderedPrompt, truncationTolerated)
		reaskStart := time.Now()
		reaskResult, reaskErr := aiClient.GenerateText(ctx, reaskPrompt, options)

		// The re-ask can itself truncate. Honour the same opt-in the first
		// attempt did — a step that tolerates a partial tolerates it twice, and
		// one that does not still fails loud.
		reaskTruncated := false
		reaskTruncatedTokens := 0
		if reaskErr != nil {
			if te, isTrunc := aiservice.IsTruncated(reaskErr); isTrunc &&
				datahelpers.GetBoolField(params.StepConfig.Config, "tolerate_truncation", false) {
				reaskResult = te.Partial
				reaskTruncated = true
				reaskTruncatedTokens = te.OutputTokens
				reaskErr = nil
			}
		}

		// The provider client writes usage back into `options`, so re-read it
		// AFTER the call — these are the re-ask's numbers, not the first
		// attempt's. Carried explicitly rather than left to default: a row with
		// NULL output_tokens is invisible to every p95/max/at-the-cap query over
		// this table, which is exactly the blind spot that made truncations
		// unmeasurable before (llm_call_log's known trap). A retry that is not
		// counted cannot be shown to be rare, and "it should stay small" is this
		// change's whole cost argument.
		reaskInputTokens, reaskOutputTokens := 0, 0
		if it, ok := options["__usage_input_tokens"].(int); ok {
			reaskInputTokens = it
		}
		if ot, ok := options["__usage_output_tokens"].(int); ok {
			reaskOutputTokens = ot
		}
		if reaskTruncated && reaskOutputTokens == 0 {
			// A truncated call may report usage only on the error. Prefer a real
			// number over a NULL for the reason above.
			reaskOutputTokens = reaskTruncatedTokens
		}

		// One call, one forensic row — the same rule the first attempt follows.
		// Marked in error_message so a census can separate re-asks from first
		// attempts without joining anything (llm_call_log has no attempt column).
		// NOTE the marker is present on SUCCESSFUL retries too, so a query that
		// treats a non-empty error_message as a failure will miscount: filter
		// failures on `success = false`, and find retries with
		// `error_message LIKE 'RETRY (bugs_open/119%'`.
		reaskLogErr := "RETRY (bugs_open/119: first attempt did not parse)"
		if reaskErr != nil {
			reaskLogErr = "RETRY (bugs_open/119) FAILED: " + reaskErr.Error()
		} else if reaskTruncated {
			reaskLogErr = "RETRY (bugs_open/119) TRUNCATED and tolerated"
		}
		LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
			AgentType:       params.AgentType,
			AgentID:         params.Headers["agent_id"],
			StepName:        params.ExecutionContext.StepName,
			OrchestrationID: params.ExecutionContext.OrchestrationID,
			CorrelationID:   params.ExecutionContext.CorrelationID,
			Model:           modelAlias,
			ModelResolved:   resolvedModel,
			Provider:        provider,
			PromptTemplate:  promptTemplate,
			PromptRendered:  reaskPrompt,
			ResponseText:    reaskResult,
			InputTokens:     reaskInputTokens,
			OutputTokens:    reaskOutputTokens,
			LatencyMs:       int(time.Since(reaskStart).Milliseconds()),
			Success:         reaskErr == nil && !reaskTruncated,
			ErrorMessage:    reaskLogErr,
			Temperature:     sentTemperature,
			MaxTokens:       sentMaxTokens,
			WorkItemID:      flywheelWorkItemID,
			Vertical:        flywheelVertical,
			RAGContextUsed:  flywheelRAG,
			Options:         options,
		})

		if reaskErr != nil {
			params.Logger.Warn("bugs_open/119 re-ask failed — falling through to the pre-existing text path",
				zap.Error(reaskErr))
		} else {
			reaskCleaned := stripMarkdownFromResponse(reaskResult)
			if rp, rprov, rerr := ParseLLMJSONWithProvenance(reaskCleaned); rerr == nil {
				params.Logger.Info("bugs_open/119 re-ask RECOVERED a parsable response — the round is not lost",
					zap.String("step_name", params.ExecutionContext.StepName),
					zap.String("provenance", rprov),
					zap.Bool("retry_truncated", reaskTruncated),
				)
				// Adopt the recovered answer wholesale, including its truncation
				// state: the marker must describe the response actually returned,
				// not the one that was thrown away, or a consumer degrades on the
				// wrong evidence.
				parsedResult, provenance, parseErr = rp, rprov, nil
				cleanedResult = reaskCleaned
				truncationTolerated = reaskTruncated
				truncatedTokens = reaskTruncatedTokens
			} else {
				params.Logger.Warn("bugs_open/119 re-ask still would not parse — falling through to the pre-existing text path",
					zap.Error(rerr),
					zap.Int("retry_response_len", len(reaskCleaned)),
				)
			}
		}
	}

	if parseErr != nil {
		params.Logger.Warn("LLM response holds no complete JSON value — returning as text (a required-content step will fail loud downstream)",
			zap.Error(parseErr),
			zap.Int("response_len", len(cleanedResult)),
			zap.String("response_tail", datahelpers.TruncateString(lastRunes(cleanedResult, 80), 80)),
		)
		out := map[string]interface{}{
			"result": cleanedResult,
			"type":   "text",
		}
		markJSONContractUnmet(out, declaredOutputType)
		markTruncated(out, truncationTolerated, truncatedTokens)
		return out, nil
	}
	if provenance != ProvenanceClean {
		// Log at Warn, not Info: the model was told to return the object and
		// nothing else, and did not. The step still succeeds — the answer was
		// complete — but the prompt that produced it wants fixing, and until
		// bugs_open/088 this class was invisible because nothing counted it.
		params.Logger.Warn("LLM response was not a bare JSON value — recovered",
			zap.String("provenance", provenance),
			zap.String("agent_type", params.AgentType),
			zap.String("step_name", params.ExecutionContext.StepName),
			zap.Int("response_len", len(cleanedResult)),
		)
	}

	out := map[string]interface{}{
		"result": parsedResult,
		"type":   "json",
	}
	markEnvelopeRecovered(out, provenance)
	markTruncated(out, truncationTolerated, truncatedTokens)
	return out, nil
}

// correctiveReaskPrompt builds the single re-ask sent when a JSON-declaring step
// got an unusable answer (bugs_open/119).
//
// It appends to the ORIGINAL prompt rather than replacing it: the seat still has
// to answer the same question, and a prompt that carried only "fix your JSON"
// would ask a reviewer to re-judge a plan it can no longer see. It also does not
// echo the broken output back — the two failure modes are a response that was cut
// off (there is nothing useful to quote) and one that is complete but malformed
// (quoting it invites the model to edit its own mistake rather than re-answer).
//
// The corrective text is chosen by WHICH failure occurred, because the two have
// opposite remedies and a single generic "please return valid JSON" would be
// useless against a truncation:
//
//   - truncated  -> the answer was too LONG and was lost; ask for the same
//     judgement, materially shorter. Deliberately NOT a token-cap raise; see
//     platform/aiservice/truncation.go:26-29.
//   - malformed  -> the answer was complete but not parsable; ask for the same
//     content as one valid JSON value. This is the case bugs_open/119 documents
//     verbatim — a review object closed one bracket early, leaving a stray "]".
func correctiveReaskPrompt(basePrompt string, wasTruncated bool) string {
	if wasTruncated {
		return basePrompt + `

── RETRY — YOUR PREVIOUS ANSWER WAS DISCARDED ──
Your previous response to this exact prompt was CUT OFF at the output token limit
and could not be used, so it counted for nothing. This is your one retry.

Give the SAME judgement again, materially SHORTER.
Keep every finding you believe in: cut words, never findings.
Keep every citation, quote, id and file:line that supports a finding — a shorter
answer must not become a less evidenced one. Cut explanation, never evidence.
Drop preamble, restate nothing back to the author, and shorten prose fields first.
Output ONLY the single JSON value the prompt asks for: start with { or [, end with
} or ], no markdown fences, no commentary.`
	}
	return basePrompt + `

── RETRY — YOUR PREVIOUS ANSWER WAS DISCARDED ──
Your previous response to this exact prompt was NOT VALID JSON and could not be
parsed, so it counted for nothing. This is your one retry.

Give the SAME answer again as ONE valid JSON value. Check the structure before you
finish: every { closed by exactly one }, every [ closed by exactly one ], no stray
or duplicated brackets, no trailing commas, all strings quoted and escaped. Output
ONLY that JSON value: start with { or [, end with } or ], no markdown fences, no
commentary before or after.`
}

// markJSONContractUnmet stamps a step result that DECLARED it needs JSON and,
// after the corrective re-ask, still has none.
//
// Answers the council's standing objection to the retry (bugs_open/119, round 1,
// seat bug_historian at medium): a re-ask lowers the FREQUENCY of the silent
// degrade but does not change its SHAPE — the step still returns
// `{result: <text>, type: "text"}` and still SUCCEEDS, which is
// `bugs_closed/076`'s title exactly ("truncated LLM responses tolerated at
// unguarded call sites") and the `missingkey=zero` family: one call site
// improved, the generic silent-success left intact.
//
// The remedy is deliberately a MARKER and not a hard error. Making a declared-
// json step fail loud would convert ~91 currently-succeeding steps into
// failing ones on a path they did not author, which is `bugs_closed/073`'s
// defect (a page became unbuildable for telling the truth) — the same reason
// 119's own candidate 2 was declined. So: keep the behaviour, end the silence.
// Same "__" convention as markTruncated/markEnvelopeRecovered, so it cannot
// collide with a model's own keys, and its ABSENCE is meaningful — it is only
// ever stamped where a contract was declared AND unmet.
//
// What it buys, concretely: this class is currently invisible to everything
// except a pod log line. With the marker a consumer can refuse, and a census can
// count it — `collected_data @> '{"<field>":{"__json_contract_unmet":true}}'`.
func markJSONContractUnmet(out map[string]interface{}, declaredOutputType string) {
	if declaredOutputType != "json" {
		return
	}
	out["__json_contract_unmet"] = true
}

// markEnvelopeRecovered stamps how a non-clean response was recovered, so a
// consumer — and a census — can tell a bare answer from one dug out of
// commentary. Same "__" convention as markTruncated: platform-set, cannot collide
// with a model's own field names. A clean parse stamps nothing, so the marker's
// presence is the signal (bugs_open/088).
func markEnvelopeRecovered(out map[string]interface{}, provenance string) {
	if provenance == "" || provenance == ProvenanceClean {
		return
	}
	out["__envelope_recovered"] = provenance
}

// markTruncated stamps a tolerated-truncation marker onto an LLM step result.
//
// The marker is the whole point of tolerating a truncation rather than failing:
// the step now SUCCEEDS, so without it a consumer cannot distinguish a complete
// answer from a fragment, and bugs_open/019 would trade a loud void for a silent
// half-answer — a strictly worse bug. Keyed with the "__" prefix used elsewhere
// in collected_data for platform-set fields, so it cannot collide with a model's
// own output keys.
func markTruncated(out map[string]interface{}, tolerated bool, outputTokens int) {
	if !tolerated {
		return
	}
	out["__truncated"] = true
	out["__truncated_output_tokens"] = outputTokens
}

// lastRunes returns up to n trailing runes of s, for logging a response tail
// without splitting a multi-byte rune.
func lastRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[len(r)-n:])
}

// stripMarkdownFromResponse removes markdown code fences from LLM responses
// Handles ```json, ```html, ``` and similar patterns
func stripMarkdownFromResponse(s string) string {
	s = strings.TrimSpace(s)

	// Handle ```json, ```html, ```text, etc. at start
	if strings.HasPrefix(s, "```") {
		// Find end of first line (the language identifier line)
		newlineIdx := strings.Index(s, "\n")
		if newlineIdx > 0 {
			s = s[newlineIdx+1:] // Skip past ```json\n
		} else {
			s = strings.TrimPrefix(s, "```")
		}
		s = strings.TrimSpace(s)
	}

	// Remove trailing ```
	if strings.HasSuffix(s, "```") {
		// Find the last occurrence and remove it
		lastFence := strings.LastIndex(s, "```")
		if lastFence > 0 {
			s = s[:lastFence]
		} else {
			s = strings.TrimSuffix(s, "```")
		}
		s = strings.TrimSpace(s)
	}

	return s
}

// validateTemplateData checks if template data has fields referenced in the template
// Logs warnings for missing fields to help debug <no value> issues
func validateTemplateData(templateData map[string]interface{}, stepConfig map[string]interface{}, logger *zap.Logger) {
	// Get input_fields from config to know what we expected to find
	inputFields := []string{}
	if fields, ok := stepConfig["input_fields"].([]interface{}); ok {
		for _, f := range fields {
			if field, ok := f.(string); ok {
				inputFields = append(inputFields, field)
			}
		}
	}

	logger.Info("Template data validation",
		zap.Strings("expected_fields", inputFields),
		zap.Strings("available_fields", getTemplateDataKeys(templateData)),
	)

	// Check each expected field.
	//
	// The key to look for is the LAST dotted segment, not the first. ExtractFields
	// stores an entry under datahelpers.TemplateRootForInputField(field) — its leaf
	// — so `reviewed_brief.company_name` arrives as `company_name` and nothing named
	// `reviewed_brief` is written by that entry at all. Splitting on parts[0] here
	// meant a dotted entry that extracted SUCCESSFULLY was reported as missing, and
	// a genuine absence looked identical to it: the one cross-check the runtime has
	// over this pair was answering about a key the extractor never writes.
	//
	// Byte-identical for an undotted entry (leaf == whole), which is every live step
	// as of 2026-09-03 (0 of 1,474 declare a dotted entry) — so this corrects the
	// latent case without changing any log the fleet emits today. Called out by the
	// council's bug_historian seat on corr 54abc24b as the shape 016b §9 keeps
	// recording: two halves of one contract, each green in isolation, meeting for
	// the first time in production. bugs_open/453.
	missingFields := []string{}
	for _, field := range inputFields {
		rootField := datahelpers.TemplateRootForInputField(field)

		if _, ok := templateData[rootField]; !ok {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		logger.Error("TEMPLATE DATA VALIDATION FAILED - Missing fields will render as <no value>",
			zap.Strings("missing_fields", missingFields),
			zap.Strings("available_fields", getTemplateDataKeys(templateData)),
		)
	}

	// Also check specific commonly-needed fields
	commonFields := []string{"reviewed_brief", "site_record", "input_data"}
	for _, field := range commonFields {
		if val, ok := templateData[field]; ok {
			if valMap, isMap := val.(map[string]interface{}); isMap {
				logger.Info("Template field contents",
					zap.String("field", field),
					zap.Strings("keys", getTemplateDataKeys(valMap)),
				)
			}
		}
	}
}

func getTemplateDataKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// resolveOutputType applies the step-then-ai_service precedence in ONE place.
//
// Deliberately extracted rather than inlined twice: the parse-failure retry
// (bugs_open/119) has to ask the same question this function answers — "did this
// step declare that it needs JSON?" — and two implementations of one resolution
// rule that must agree is precisely the drift class the council gate exists to
// catch. One rule, two callers.
func resolveOutputType(stepConfig, aiConfig map[string]interface{}) string {
	// Step config first (where it is typically defined in the workflow).
	if outputType := getOutputType(stepConfig); outputType != "" {
		return outputType
	}
	return getOutputType(aiConfig)
}

// appendOutputInstructions adds format-specific instructions based on the
// declared output type (see getOutputType for which config keys that means).
func appendOutputInstructions(prompt string, aiConfig map[string]interface{}, stepConfig map[string]interface{}, logger *zap.Logger) string {
	outputType := resolveOutputType(stepConfig, aiConfig)

	if outputType == "" {
		// No specific output type, use default clean output instructions
		return prompt + getDefaultOutputInstructions()
	}

	instructions := getOutputInstructions(outputType)
	if instructions == "" {
		return prompt
	}

	logger.Info("Appending output format instructions",
		zap.String("output_type", outputType))

	return prompt + instructions
}

// llmOutputVocabulary is the set of output types this action knows how to write
// instructions for. It exists to make the output_format fallback below SAFE, and
// it is load-bearing rather than defensive tidiness: `output_format` is not a
// name this action owns. `query_database` reads the SAME KEY with a DIFFERENT
// vocabulary — "array"/"object" (database_actions.go:26) — so a blind
// pass-through would one day hand getOutputInstructions a value that means
// something else entirely and select a wrong instruction set. An unrecognised
// value falls through to the default instructions, which is exactly what it
// gets today.
var llmOutputVocabulary = map[string]bool{
	"json": true, "text": true, "html": true, "markdown": true,
}

// getOutputType extracts the declared output type from a step or ai_service config.
//
// It reads TWO keys, and the second one is a repair rather than a convenience
// (bugs_open/119). The code has always read `output_type`; the fleet has almost
// always written `output_format`. Measured 2026-08-01 across all 135 active
// `execute_llm_prompt` steps (134 top-level + 1 nested in a loop's
// `sub_workflow`): `output_type` is set on **6**, `output_format` on **101** —
// 91 of them "json", across 33 agents, every council seat among them.
//
// DEPTH MATTERS HERE and was measured rather than assumed, because the council
// gated round 1 on exactly this: a step's prompt and its token cap are known to
// sit at DIFFERENT depths in `default_config`, so a census taken at one depth can
// silently overstate a fix's reach. All 101 declarations sit at
// `config.output_format` — the map `params.StepConfig.Config` is — with ZERO
// under `config.ai_service` and ZERO at the step root. The ai_service fallback
// below is therefore unexercised today; it is kept because the precedence rule
// belongs with the key, not because anything uses it.
//
// So for as long as this function read one key, ~75% of LLM steps declared an
// output contract that nothing honoured, and silently received
// getDefaultOutputInstructions() — which says nothing whatsoever about JSON. The
// instruction those 90 steps were missing is "Ensure valid JSON syntax (proper
// quotes, commas, brackets)", which is the exact failure bugs_open/119 was filed
// for: a seat closing its review object one bracket early.
//
// This is the `bugs_open/134` class — a config key that is declared and inert
// looks identical to one that is live — and the remedy is to make the key the
// fleet actually writes mean what it says, not to migrate 100 config rows to a
// spelling the fleet has never used.
func getOutputType(config map[string]interface{}) string {
	if outputType, ok := config["output_type"].(string); ok && outputType != "" {
		return outputType
	}
	// The key the fleet actually writes. Vocabulary-gated — see above.
	if outputFormat, ok := config["output_format"].(string); ok {
		if llmOutputVocabulary[strings.ToLower(strings.TrimSpace(outputFormat))] {
			return strings.ToLower(strings.TrimSpace(outputFormat))
		}
	}
	return ""
}

// getOutputInstructions returns format-specific instructions
func getOutputInstructions(outputType string) string {
	switch outputType {
	case "json":
		return getJSONOutputInstructions()
	case "html":
		return getHTMLOutputInstructions()
	case "text":
		return getTextOutputInstructions()
	case "markdown":
		return getMarkdownOutputInstructions()
	default:
		return getDefaultOutputInstructions()
	}
}

// getJSONOutputInstructions returns JSON formatting instructions
func getJSONOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT - JSON:
- Output ONLY the raw JSON object or array
- Do NOT wrap in markdown code fences (no ` + "```" + ` or ` + "```json" + `)
- Do NOT add explanatory text before or after the JSON
- Start your response with { or [ and end with } or ]
- Ensure valid JSON syntax (proper quotes, commas, brackets)

Example CORRECT:
{"site_type": "brochure", "recommended_builder": "multipage-website-builder"}

Example INCORRECT:
` + "```json\n{\"site_type\": \"brochure\"}\n```\n\nNote: This classification..."
}

// getHTMLOutputInstructions returns HTML formatting instructions
func getHTMLOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT - HTML:
- Output ONLY the raw HTML code
- Do NOT wrap in markdown code fences (no ` + "```" + ` or ` + "```html" + `)
- Do NOT add explanatory text before or after the HTML
- Start with <!DOCTYPE html> or the appropriate opening tag
- Include complete, valid HTML structure

Example CORRECT:
<!DOCTYPE html>
<html lang="en">
<head>...</head>
<body>...</body>
</html>

Example INCORRECT:
` + "```html\n<!DOCTYPE html>...\n```\n\nHere's the HTML for your site..."
}

// getTextOutputInstructions returns text formatting instructions
func getTextOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT - TEXT:
- Output ONLY the actual text content
- Do NOT wrap in markdown code fences
- Do NOT add meta-commentary like "Here's the content..." or "I've created..."
- Start directly with the content itself

Example CORRECT:
Welcome to our website. We provide excellent services...

Example INCORRECT:
Here's the content you requested:
` + "```\nWelcome to our website...\n```"
}

// getMarkdownOutputInstructions returns markdown formatting instructions
func getMarkdownOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT - MARKDOWN:
- Output ONLY the markdown content
- Do NOT wrap in code fences
- Do NOT add explanatory text before or after
- Use proper markdown syntax for headings, lists, links, etc.

Example CORRECT:
# Welcome

This is **bold** and this is *italic*.

Example INCORRECT:
` + "```markdown\n# Welcome\n```\n\nI've created markdown content for you..."
}

// getDefaultOutputInstructions returns general clean output instructions
func getDefaultOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT:
- Output ONLY the requested content
- Do NOT wrap in code fences or markdown formatting
- Do NOT add preambles like "Here's what you asked for..."
- Do NOT add post-ambles like "Hope this helps!"
- Start directly with the actual content`
}

// helper function to load agent definition
func loadAgentDefinitionForAction(ctx context.Context, db interface{}, agentType string) (*AgentDefinition, error) {

	fmt.Printf("DEBUG: loadAgentDefinitionForAction called with agentType=%s, db type=%T\n", agentType, db)

	query := `
		SELECT id, type, display_name, COALESCE(description, ''), category,
		       image_repository, image_tag, 
		       resources, default_config, capabilities, topics,
		       health_config, env_vars, is_active
		FROM agent_definitions
		WHERE type = $1 AND is_active = true
		ORDER BY version DESC
		LIMIT 1
	`

	var def AgentDefinition
	var defaultConfigJSON json.RawMessage // Read as RawMessage first
	var capabilitiesJSON json.RawMessage

	// Handle both *sql.DB and *pgxpool.Pool
	switch d := db.(type) {
	case *sql.DB:
		// For *sql.DB, we need to handle the Command field differently
		err := d.QueryRowContext(ctx, query, agentType).Scan(
			&def.ID,
			&def.Type,
			&def.DisplayName,
			&def.Description,
			&def.Category,
			&def.ImageRepository,
			&def.ImageTag,
			&def.Resources,
			&def.DefaultConfig,
			&def.Capabilities,
			&def.Topics,
			&def.HealthConfig,
			&def.EnvVars,
			&def.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query agent definition: %w", err)
		}
		// Note: Command field is not loaded here as it's not needed for LLM prompt
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, agentType).Scan(
			&def.ID,
			&def.Type,
			&def.DisplayName,
			&def.Description,
			&def.Category,
			&def.ImageRepository,
			&def.ImageTag,
			&def.Resources,
			&def.DefaultConfig,
			&def.Capabilities,
			&def.Topics,
			&def.HealthConfig,
			&def.EnvVars,
			&def.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query agent definition: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	// Now unmarshal the JSON into the map
	if len(defaultConfigJSON) > 0 && string(defaultConfigJSON) != "null" {
		if err := json.Unmarshal(defaultConfigJSON, &def.DefaultConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal default_config: %w", err)
		}
	}

	// Unmarshal capabilities
	if len(capabilitiesJSON) > 0 {
		json.Unmarshal(capabilitiesJSON, &def.Capabilities)
	}

	// Validate that we have a config
	if def.DefaultConfig == nil {
		return nil, fmt.Errorf("agent %s has no default config", agentType)
	}

	return &def, nil
}

// Keep all existing helper functions...
func createAIClient(ctx context.Context, aiServiceConfig map[string]interface{}) (aiservice.AIService, error) {
	return aiservice.NewClient(ctx, aiServiceConfig)
}

func ConditionalRouteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	conditionField := config["condition_field"].(string)
	routes := config["routes"].(map[string]interface{})

	// Evaluate the condition
	// Check for condition value in input_data first
	var conditionValue interface{}

	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		// Look in input_data first
		if val, exists := inputData[conditionField]; exists {
			conditionValue = val
		}
	}

	// Fall back to collected data
	if conditionValue == nil {
		conditionValue = params.CollectedData[conditionField]
	}

	// If no explicit condition value, evaluate it
	if conditionValue == nil {
		conditionValue = evaluateCondition(params)
		params.CollectedData[conditionField] = conditionValue
	}

	// Determine next step
	nextStep, ok := routes[fmt.Sprintf("%v", conditionValue)]
	if !ok {
		// Use default route if available
		nextStep = routes["default"]
	}

	return map[string]interface{}{
		"next_step": nextStep,
		"condition": conditionValue,
	}, nil
}

func evaluateCondition(params ActionParams) interface{} {
	// Simple complexity evaluation
	inputSize := len(params.CollectedData)

	if inputSize < 3 {
		return "simple"
	} else if inputSize < 10 {
		return "moderate"
	}
	return "complex"
}

func EvaluateTaskAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Evaluating task complexity")

	complexity := "simple" // Default

	// Check input data size and structure
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		dataSize := len(inputData)

		// Check for indicators of complexity
		hasNestedData := false

		for _, v := range inputData {
			switch v.(type) {
			case map[string]interface{}, []interface{}:
				hasNestedData = true
			}
		}

		// Determine complexity
		if dataSize > 10 || hasNestedData {
			complexity = "complex"
		} else if dataSize > 5 {
			complexity = "moderate"
		}
	}

	params.Logger.Info("Task complexity evaluated",
		zap.String("complexity", complexity))

	return map[string]interface{}{
		"complexity": complexity,
	}, nil
}

// extractDataForAiAgent merges data from multiple sources specified in the step's 'input_fields' config.
func extractDataForAiAgent(params ActionParams) interface{} {
	params.Logger.Info("Extracting data for AI agent using UNIFIED EXTRACTOR",
		zap.Any("available_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Determine which fields to extract
	var inputFields []string
	if fields, ok := params.StepConfig.Config["input_fields"].([]interface{}); ok {
		for _, fieldInterface := range fields {
			if field, ok := fieldInterface.(string); ok {
				inputFields = append(inputFields, field)
			}
		}
	} else {
		params.Logger.Warn("No 'input_fields' found in config, defaulting to ['input_data']")
		inputFields = []string{"input_data"}
	}

	params.Logger.Info("Processing input_fields", zap.Strings("fields", inputFields))

	// USE THE UNIFIED EXTRACTOR
	templateData := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	params.Logger.Info("Template data extracted",
		zap.Int("field_count", len(templateData)),
		zap.Strings("keys", datahelpers.GetMapKeys(templateData)),
	)

	return templateData
}

// getPromptWithPriority implements three-tier priority for prompt selection
func getPromptWithPriority(params ActionParams, agentConfig map[string]interface{}) (prompt string, source string) {
	logger := params.Logger

	logger.Info("in getPromptWithPriority")

	// PRIORITY 1: Check incoming message for prompt (from parent's call_agent)
	// Check in StepConfig.Config first (this is where call_agent passes it)
	if configPrompt, ok := params.StepConfig.Config["prompt"].(string); ok && configPrompt != "" {
		// Check if prompt contains template syntax and needs interpolation
		if strings.Contains(configPrompt, "{{") && strings.Contains(configPrompt, "}}") {
			logger.Info("Prompt contains template syntax, interpolating against CollectedData",
				zap.String("raw_prompt", configPrompt))

			interpolated, err := datahelpers.RenderPromptTemplate(configPrompt, params.CollectedData, *logger)
			if err != nil {
				logger.Warn("Failed to interpolate prompt template, using raw prompt",
					zap.Error(err),
					zap.String("raw_prompt", configPrompt))
				// Fall through to use raw prompt
			} else if interpolated != "" && interpolated != configPrompt {
				logger.Info("Prompt interpolated successfully",
					zap.String("interpolated_preview", datahelpers.TruncateString(interpolated, 200)))
				configPrompt = interpolated
			}
		}

		logger.Info("Using prompt from step config (Priority 1 - from parent)",
			zap.String("prompt_preview", datahelpers.TruncateString(configPrompt, 100)))
		return configPrompt, "parent_message"
	}

	// Also check in CollectedData["prompt"] as a fallback
	if collectedPrompt, ok := params.CollectedData["prompt"].(string); ok && collectedPrompt != "" {
		logger.Info("Using prompt from collected data (Priority 1 - from parent)",
			zap.String("prompt_preview", datahelpers.TruncateString(collectedPrompt, 100)))
		return collectedPrompt, "parent_message"
	}

	// PRIORITY 2: Check agent's own default_config.prompt_template
	// This comes from the agent_definitions table for this specific agent type
	if agentPrompt, ok := agentConfig["prompt_template"].(string); ok && agentPrompt != "" {
		logger.Info("Using prompt from agent config (Priority 2 - agent's default)",
			zap.String("agent_type", params.AgentType),
			zap.String("prompt_preview", datahelpers.TruncateString(agentPrompt, 100)))
		return agentPrompt, "agent_default"
	}

	// PRIORITY 3: Check workflow step config (fallback)
	// This is the hardcoded fallback in the workflow definition
	if stepConfig, ok := params.StepConfig.Config["prompt_template"].(string); ok && stepConfig != "" {
		logger.Info("Using prompt from workflow step config (Priority 3 - fallback)",
			zap.String("prompt_preview", datahelpers.TruncateString(stepConfig, 100)))
		return stepConfig, "workflow_fallback"
	}

	// Generic fallback if nothing found
	logger.Warn("No prompt found in any tier, using generic fallback")
	return "Generate content based on the provided context.", "generic_fallback"
}

// platformPromptBlocks maps each templateData key to the agent_default_configs
// row that backs it. Adding a third platform-wide block means one entry here, a
// const in platform/voicestyle, and a reviewed carrier migration — never an ad
// hoc injection elsewhere (council round b5a642b7, architecture note).
var platformPromptBlocks = map[string]string{
	"voice_style":    voicestyle.ConfigName,
	"build_standard": voicestyle.BuildStandardConfigName,
}

// injectPlatformBlocks adds the platform-wide prompt blocks to templateData.
// Semantics shared by every block, pinned by TestInjectPlatformBlocks: a
// step-supplied value is never overwritten (request-level override), a template
// that does not name the key is unaffected (opt-in is the template's mention;
// an extra map key is inert to text/template), and a missing or unfetchable row
// degrades to "no block" rather than a failed generation. Extracted from
// ExecuteLLMPromptAction so the opt-in + override semantics are testable
// without its machinery.
func injectPlatformBlocks(ctx context.Context, templateData map[string]interface{}, get func(ctx context.Context, name string) (string, bool)) {
	for key, name := range platformPromptBlocks {
		if _, already := templateData[key]; already {
			continue
		}
		if block, ok := get(ctx, name); ok {
			templateData[key] = block
		}
	}
}
