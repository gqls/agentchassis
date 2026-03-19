// FILE: platform/orchestration/actions/ch_llm_review_action.go
// Reviews ambiguous Companies House matches using an LLM.
// Processes pending_llm_review entries from ch_vet_companies, batches them
// into LLM calls, and updates match status based on the response.
//
// Actions:
//   - ch_llm_review: Review pending matches using LLM judgment

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

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
	Confidence    float64
	Index         int // position in the batch for parsing
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

	batchSize := 15 // candidates per LLM call
	if bs, ok := config["llm_batch_size"].(float64); ok && bs > 0 {
		batchSize = int(bs)
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

	// Create AI client
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	aiConfig := map[string]interface{}{
		"provider":        "anthropic",
		"model":           "claude-haiku-4-5",
		"api_key_env_var": "ANTHROPIC_API_KEY",
		"max_tokens":      2000,
		"temperature":     0.0,
	}
	aiClient, err := createAIClient(ctx, aiConfig)
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

		// Build prompt
		prompt := buildReviewPrompt(batch)

		params.Logger.Info("CHLLMReview: sending batch to LLM",
			zap.Int("batch_start", i),
			zap.Int("batch_size", len(batch)))

		// Call LLM
		response, err := aiClient.GenerateText(ctx, prompt, nil)
		if err != nil {
			params.Logger.Warn("CHLLMReview: LLM call failed",
				zap.Int("batch_start", i),
				zap.Error(err))
			continue
		}

		// Parse response
		decisions := parseReviewResponse(response, len(batch))

		// Apply decisions
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
				}
			default: // UNCERTAIN
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

// loadPendingReviewCandidates loads all pending_llm_review matches with business details.
func loadPendingReviewCandidates(ctx context.Context, db *sql.DB) ([]reviewCandidate, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT ch.company_number, ch.company_name, COALESCE(ch.postcode, ''),
			   b.id::text, b.name, COALESCE(b.postcode, ''),
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
			&c.BusinessID, &c.BusinessName, &c.BizPostcode, &c.Confidence); err != nil {
			continue
		}
		c.Index = idx
		idx++
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}

// buildReviewPrompt creates a prompt for the LLM to review a batch of candidates.
func buildReviewPrompt(batch []reviewCandidate) string {
	var sb strings.Builder

	sb.WriteString(`You are matching UK veterinary practice trading names to their Companies House registered company names.

For each pair below, determine if they are the SAME business entity. Consider:
- The practice may trade under a different name than its registered company
- The registered office may be at an accountant's address, not the practice
- Corporate groups (Vets4Pets, IVC, CVS) register branches under the parent or a location-specific company
- "Ltd"/"Limited"/"LLP" suffixes are legal forms, ignore them for matching
- Geographic proximity matters — a practice in Glasgow is unlikely to be registered in London unless it's a corporate group

For each numbered pair, respond with ONLY the number followed by YES, NO, or UNCERTAIN.
Example response format:
1: YES
2: NO
3: UNCERTAIN

Pairs to review:
`)

	for i, c := range batch {
		sb.WriteString(fmt.Sprintf("\n%d.\n", i+1))
		sb.WriteString(fmt.Sprintf("  Practice: \"%s\" (postcode: %s)\n", c.BusinessName, c.BizPostcode))
		sb.WriteString(fmt.Sprintf("  CH Company: \"%s\" (registered: %s)\n", c.CompanyName, c.CHPostcode))
		sb.WriteString(fmt.Sprintf("  Similarity score: %.2f\n", c.Confidence))
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

		// Parse "1: YES" or "1. YES" or "1 YES" patterns
		var idx int
		var decision string

		// Try "N: DECISION" format
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
