// FILE: platform/orchestration/actions/ch_llm_review_action.go
// Reviews ambiguous Companies House matches using an LLM.
// Processes pending_llm_review entries from ch_vet_companies, batches them
// into LLM calls, and updates match status based on the response.
//
// The action is industry-agnostic. Industry-specific context (corporate groups,
// formation addresses, industry name) comes from step config. The ai_service
// config comes from the agent definition's default_config, per guidelines.
//
// Actions:
//   - ch_llm_review: Review pending matches using LLM judgment

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// reviewCandidate holds a pending match for LLM review
type reviewCandidate struct {
	CompanyNumber string
	CompanyName   string
	CHPostcode    string
	BusinessID    string
	BusinessName  string
	BizPostcode   string
	BizTown       string
	BizGroupName  string
	Confidence    float64
	Index         int
}

// CHLLMReviewAction reviews pending_llm_review matches using an LLM.
func CHLLMReviewAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CHLLMReviewAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config

	batchSize := 15
	if bs, ok := config["llm_batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}

	// Read industry context from step config (keeps the action generic)
	industryName := "business"
	if in, ok := config["industry_name"].(string); ok && in != "" {
		industryName = in
	}

	industryContext := ""
	if ic, ok := config["industry_context"].(string); ok && ic != "" {
		industryContext = ic
	}

	// Load ai_service config from agent_config (per guidelines)
	// Priority: agent_config.ai_service → step config.ai_service
	aiServiceConfig := getAIServiceConfig(params)
	if aiServiceConfig == nil {
		return nil, fmt.Errorf("ai_service configuration not found in agent_config or step config")
	}

	// Load pending review candidates
	candidates, err := loadPendingReviewCandidates(ctx, params.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to load candidates: %w", err)
	}

	if len(candidates) == 0 {
		params.Logger.Info("CHLLMReview: no pending candidates")
		return map[string]interface{}{
			"status":    "complete",
			"reviewed":  0,
			"confirmed": 0,
			"rejected":  0,
			"uncertain": 0,
		}, nil
	}

	params.Logger.Info("CHLLMReview: loaded candidates",
		zap.Int("count", len(candidates)))

	// Create AI client from config
	aiClient, err := createAIClient(ctx, aiServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	totalReviewed := 0
	totalConfirmed := 0
	totalRejected := 0
	totalUncertain := 0

	// Process in batches
	for i := 0; i < len(candidates); i += batchSize {
		select {
		case <-ctx.Done():
			params.Logger.Warn("CHLLMReview: context cancelled",
				zap.Int("reviewed", totalReviewed))
			break
		default:
		}

		end := i + batchSize
		if end > len(candidates) {
			end = len(candidates)
		}
		batch := candidates[i:end]

		prompt := buildReviewPrompt(batch, industryName, industryContext)

		params.Logger.Info("CHLLMReview: sending batch to LLM",
			zap.Int("batch_start", i),
			zap.Int("batch_size", len(batch)))

		// One options map PER BATCH, because the client writes its forensics back
		// into the map it is handed (__sent_max_tokens, __usage_*) and a shared map
		// would have each batch overwrite the previous batch's llm_call_log row.
		//
		// It is built by the package's resolver, not left empty (bugs_open/257).
		// An empty map is not neutral: since Path A the client falls back to the
		// `ai_service` block it was CONSTRUCTED with, so the budget did arrive — but
		// the STEP's own `max_tokens` and `budget_tokens` never reached the wire,
		// because no client constructor is ever shown a step config. The helper
		// also emits the "nobody sized this step" warning, which an empty map
		// cannot.
		llmOptions := llmOptionsFromConfig(config, aiServiceConfig, params.Logger, "ch_llm_review")
		llmCallStart := time.Now()

		response, err := aiClient.GenerateText(ctx, prompt, llmOptions)

		llmLatencyMs := int(time.Since(llmCallStart).Milliseconds())

		// Capture what the provider actually sent, for llm_call_log
		var sentTemperature interface{}
		if t, ok := llmOptions["__sent_temperature"]; ok {
			sentTemperature = t
		}
		sentMaxTokens := 0
		if mt, ok := llmOptions["__sent_max_tokens"].(int); ok {
			sentMaxTokens = mt
		}

		// Extract model name for logging
		llmModel := ""
		if m, ok := aiServiceConfig["model"].(string); ok {
			llmModel = m
		}
		llmProvider := ""
		if p, ok := aiServiceConfig["provider"].(string); ok {
			llmProvider = p
		}

		if err != nil {
			params.Logger.Warn("CHLLMReview: LLM call failed",
				zap.Int("batch_start", i),
				zap.Error(err))

			// Log the failed call
			LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
				AgentType:       params.AgentType,
				AgentID:         params.Headers["agent_id"],
				StepName:        params.ExecutionContext.StepName,
				OrchestrationID: params.ExecutionContext.OrchestrationID,
				CorrelationID:   params.ExecutionContext.CorrelationID,
				Model:           llmModel,
				Provider:        llmProvider,
				PromptRendered:  prompt,
				LatencyMs:       llmLatencyMs,
				Success:         false,
				ErrorMessage:    err.Error(),
				Temperature:     sentTemperature,
				MaxTokens:       sentMaxTokens,
				Options:         llmOptions,
			})

			continue
		}

		// Log the successful call
		inputTokens := 0
		outputTokens := 0
		if it, ok := llmOptions["__usage_input_tokens"].(int); ok {
			inputTokens = it
		}
		if ot, ok := llmOptions["__usage_output_tokens"].(int); ok {
			outputTokens = ot
		}
		LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
			AgentType:       params.AgentType,
			AgentID:         params.Headers["agent_id"],
			StepName:        params.ExecutionContext.StepName,
			OrchestrationID: params.ExecutionContext.OrchestrationID,
			CorrelationID:   params.ExecutionContext.CorrelationID,
			Model:           llmModel,
			Provider:        llmProvider,
			PromptRendered:  prompt,
			ResponseText:    response,
			InputTokens:     inputTokens,
			OutputTokens:    outputTokens,
			LatencyMs:       llmLatencyMs,
			Success:         true,
			Temperature:     sentTemperature,
			MaxTokens:       sentMaxTokens,
			Options:         llmOptions,
		})

		decisions := parseReviewResponse(response, len(batch))

		for j, decision := range decisions {
			if j >= len(batch) {
				break
			}
			c := batch[j]

			switch decision {
			case "YES":
				err := updateReviewResult(ctx, params.DB, c.CompanyNumber, c.BusinessID, c.Confidence, "llm_confirmed")
				if err == nil {
					totalConfirmed++
					params.Logger.Info("CHLLMReview: confirmed",
						zap.String("business", c.BusinessName),
						zap.String("company", c.CompanyName))
				}
			case "NO":
				err := clearReviewMatch(ctx, params.DB, c.CompanyNumber)
				if err == nil {
					totalRejected++
					params.Logger.Info("CHLLMReview: rejected",
						zap.String("business", c.BusinessName),
						zap.String("company", c.CompanyName))
				}
			default:
				err := updateReviewResult(ctx, params.DB, c.CompanyNumber, c.BusinessID, c.Confidence, "llm_uncertain")
				if err == nil {
					totalUncertain++
				}
			}
			totalReviewed++
		}
	}

	// Notify scheduler
	taskName := "ch-llm-review"
	if tn, ok := config["task_name"].(string); ok && tn != "" {
		taskName = tn
	}
	_, _ = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`,
		taskName)

	params.Logger.Info("CHLLMReview: complete",
		zap.Int("total_reviewed", totalReviewed),
		zap.Int("confirmed", totalConfirmed),
		zap.Int("rejected", totalRejected),
		zap.Int("uncertain", totalUncertain))

	return map[string]interface{}{
		"status":    "complete",
		"reviewed":  totalReviewed,
		"confirmed": totalConfirmed,
		"rejected":  totalRejected,
		"uncertain": totalUncertain,
	}, nil
}

// getAIServiceConfig reads ai_service config following the standard pattern:
// 1. agent_config.ai_service (top-level in agent definition)
// 2. step config.ai_service (per-step override)
//
// NOTE: On multi-type pods (e.g. business-intel pod running ch-llm-reviewer workflow),
// agent_config comes from the POD's type (business-intel), not the message's type.
// The ch-llm-reviewer's default_config.ai_service won't be in agent_config.
// So for actions running on shared pods, ai_service should be in the step config.
func getAIServiceConfig(params ActionParams) map[string]interface{} {
	// Try agent_config first (standard location per guidelines)
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if aiService, ok := agentConfig["ai_service"].(map[string]interface{}); ok {
			return aiService
		}
	}

	// Fall back to step config
	if aiService, ok := params.StepConfig.Config["ai_service"].(map[string]interface{}); ok {
		return aiService
	}

	return nil
}

// loadPendingReviewCandidates loads all pending_llm_review matches with business details.
func loadPendingReviewCandidates(ctx context.Context, db *sql.DB) ([]reviewCandidate, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT ch.company_number, ch.company_name, COALESCE(ch.postcode, ''),
			   b.id::text, b.name, COALESCE(b.postcode, ''),
			   COALESCE(b.town, ''), COALESCE(b.group_name, 'Independent'),
			   COALESCE(ch.match_confidence, 0)
		FROM business_intel.ch_vet_companies ch
		JOIN business_intel.businesses b ON b.id = ch.matched_business_id
		WHERE ch.match_method = 'pending_llm_review'
		ORDER BY ch.match_confidence DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []reviewCandidate
	idx := 0
	for rows.Next() {
		var c reviewCandidate
		if err := rows.Scan(&c.CompanyNumber, &c.CompanyName, &c.CHPostcode,
			&c.BusinessID, &c.BusinessName, &c.BizPostcode,
			&c.BizTown, &c.BizGroupName, &c.Confidence); err != nil {
			continue
		}
		c.Index = idx
		idx++
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

// buildReviewPrompt creates a prompt for the LLM to review a batch of candidates.
// The generic matching structure is in Go; industry-specific context is injected
// from step config so the action works across verticals.
func buildReviewPrompt(batch []reviewCandidate, industryName, industryContext string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are an expert at matching UK %s trading names to their Companies House (CH) registered company names. Your task is to determine if each pair refers to the SAME business entity.

IMPORTANT CONTEXT:
- Many businesses trade under a different name than their CH registration. E.g. "Riverside Services" may be registered as "RIVERSIDE SERVICES LIMITED". These ARE the same entity.
- The registered office is often at an accountant, solicitor, or company formation agent — NOT at the trading address. Different postcodes alone do NOT mean different businesses.
- However, different BRAND names mean different businesses. Two businesses in the same town with different names are competitors, not the same entity.
`, industryName))

	// Inject industry-specific context from step config
	if industryContext != "" {
		sb.WriteString("\nINDUSTRY-SPECIFIC CONTEXT:\n")
		sb.WriteString(industryContext)
		sb.WriteString("\n")
	}

	sb.WriteString(`
MATCHING RULES:
- YES: The core distinctive name matches (ignoring Ltd/Group/Surgery/Clinic suffixes and minor variations).
- YES: Known abbreviation or name variant of the same business, in a plausible geographic area.
- NO: Different brand names that happen to share a location word. E.g. "Best Services In Liverpool" and "LIVERPOOL ACME LIMITED" — different brands.
- NO: The practice's group ownership doesn't match the CH company's brand. If listed as "Independent" but matched to a corporate chain, that's NO.
- UNCERTAIN: Name is plausible but you can't be confident without more information.

For each numbered pair, respond with ONLY the number followed by YES, NO, or UNCERTAIN. No explanations.

Pairs to review:
`)

	for i, c := range batch {
		sb.WriteString(fmt.Sprintf("\n%d.\n", i+1))
		sb.WriteString(fmt.Sprintf("  Business: \"%s\"\n", c.BusinessName))
		sb.WriteString(fmt.Sprintf("  Location: %s, %s\n", c.BizTown, c.BizPostcode))
		sb.WriteString(fmt.Sprintf("  Group: %s\n", c.BizGroupName))
		sb.WriteString(fmt.Sprintf("  CH Company: \"%s\" (registered: %s)\n", c.CompanyName, c.CHPostcode))
		sb.WriteString(fmt.Sprintf("  Name similarity: %.0f%%\n", c.Confidence*100))
	}

	return sb.String()
}

// parseReviewResponse extracts YES/NO/UNCERTAIN decisions from the LLM response.
func parseReviewResponse(response string, expectedCount int) []string {
	decisions := make([]string, expectedCount)
	for i := range decisions {
		decisions[i] = "UNCERTAIN" // default if parsing fails
	}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var idx int
		var decision string

		if n, err := fmt.Sscanf(line, "%d: %s", &idx, &decision); n == 2 && err == nil {
			// ok
		} else if n, err := fmt.Sscanf(line, "%d. %s", &idx, &decision); n == 2 && err == nil {
			// ok
		} else if n, err := fmt.Sscanf(line, "%d %s", &idx, &decision); n == 2 && err == nil {
			// ok
		} else {
			continue
		}

		decision = strings.ToUpper(strings.TrimRight(decision, ".,;:"))

		if idx >= 1 && idx <= expectedCount {
			switch decision {
			case "YES", "NO", "UNCERTAIN":
				decisions[idx-1] = decision
			}
		}
	}

	return decisions
}

// updateReviewResult updates the match method after LLM review.
func updateReviewResult(ctx context.Context, db *sql.DB, companyNumber, businessID string, confidence float64, method string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE business_intel.ch_vet_companies
		SET match_method = $1,
			matched_at = NOW(),
			updated_at = NOW()
		WHERE company_number = $2
		  AND matched_business_id = $3::uuid`,
		method,
		companyNumber,
		businessID,
	)
	return err
}

// clearReviewMatch removes a rejected match.
func clearReviewMatch(ctx context.Context, db *sql.DB, companyNumber string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE business_intel.ch_vet_companies
		SET matched_business_id = NULL,
			matched_at = NULL,
			match_confidence = NULL,
			match_method = NULL,
			updated_at = NOW()
		WHERE company_number = $1`,
		companyNumber,
	)
	return err
}
