// FILE: platform/orchestration/actions/dispatch_verifiers.go
//
// DispatchVerifiersAction queries businesses with verification_status='pending'
// and produces one Kafka message per business to trigger vet-practice-verifier
// agents. Follows the same dispatch pattern as dispatch_area_discoverers.go.
//
// Workflow config:
//
//	"dispatch_verifiers": {
//	    "action": "dispatch_verifiers",
//	    "config": {
//	        "input_fields": ["limit", "vertical_slug", "delay_ms"]
//	    },
//	    "output_field": "verify_dispatch_result",
//	    "next_step": "complete"
//	}
//
// Registration:
//   "dispatch_verifiers": DispatchVerifiersAction,

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DispatchVerifiersInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{"verify_limit", "vertical_slug", "delay_ms"},
	Defaults: map[string]interface{}{
		"verify_limit":  100,
		"vertical_slug": "veterinary",
		"delay_ms":      200,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("dispatch_verifiers", DispatchVerifiersInputSpec)
}

func DispatchVerifiersAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("DispatchVerifiersAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.Producer == nil {
		return nil, fmt.Errorf("no Kafka producer available")
	}
	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		DispatchVerifiersInputSpec,
		params.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	limit := inputs.GetInt("verify_limit", 100)
	verticalSlug := inputs.Get("vertical_slug")
	if verticalSlug == "" {
		verticalSlug = "veterinary"
	}
	delayMs := inputs.GetInt("delay_ms", 200)

	// Query businesses that need verification
	rows, err := params.DB.QueryContext(ctx, `
		SELECT b.id, b.name
		FROM business_intel.businesses b
		JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
		WHERE bv.slug = $1
		  AND b.verification_status = 'pending'
		  AND b.is_active = true
		ORDER BY b.created_at ASC
		LIMIT $2`, verticalSlug, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending businesses: %w", err)
	}
	defer rows.Close()

	type pendingBiz struct {
		ID   string
		Name string
	}

	var businesses []pendingBiz
	for rows.Next() {
		var b pendingBiz
		if err := rows.Scan(&b.ID, &b.Name); err != nil {
			params.Logger.Warn("DispatchVerifiersAction: scan error", zap.Error(err))
			continue
		}
		businesses = append(businesses, b)
	}

	if len(businesses) == 0 {
		params.Logger.Info("DispatchVerifiersAction: no pending businesses to verify")
		return map[string]interface{}{
			"dispatched": 0,
			"errors":     0,
			"message":    "no pending businesses",
		}, nil
	}

	batchName := fmt.Sprintf("verify-%s", time.Now().Format("20060102-150405"))
	topic := "system.agent.generic.requests"
	clientID := params.ExecutionContext.ClientID

	params.Logger.Info("DispatchVerifiersAction: dispatching",
		zap.Int("business_count", len(businesses)),
		zap.String("batch", batchName),
		zap.String("topic", topic))

	dispatched := 0
	errors := 0

	for _, biz := range businesses {
		orchID := uuid.New().String()
		reqID := uuid.New().String()
		orchName := fmt.Sprintf("%s-%s", batchName, biz.ID[:8])

		// Message body — same format the generic agent expects
		body := map[string]interface{}{
			"action": "orchestrate",
			"config": map[string]interface{}{
				"agent_type": "vet-practice-verifier",
			},
			"input_data": map[string]interface{}{
				"business_id": biz.ID,
			},
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			params.Logger.Warn("DispatchVerifiersAction: marshal failed",
				zap.String("business_id", biz.ID), zap.Error(err))
			errors++
			continue
		}

		// Headers — same format as dispatch_area_discoverers.go
		headers := map[string]string{
			"correlation_id":     uuid.New().String(),
			"request_id":         reqID,
			"message_id":         uuid.New().String(),
			"orchestration_id":   orchID,
			"orchestration_name": orchName,
			"step_name":          "start",
			"client_id":          clientID,
			"message_type":       "request",
			"action":             "orchestrate",
			"from_agent_type":    params.AgentType,
			"from_agent_id":      params.ExecutionContext.FromAgentID,
			"responses_topic":    "system.generic.responses",
		}

		err = params.Producer.Produce(ctx, topic, headers, []byte(reqID), bodyBytes)
		if err != nil {
			params.Logger.Warn("DispatchVerifiersAction: produce failed",
				zap.String("business_id", biz.ID), zap.Error(err))
			errors++
			continue
		}

		dispatched++
		params.Logger.Info("DispatchVerifiersAction: dispatched verifier",
			zap.String("business_id", biz.ID),
			zap.String("name", biz.Name),
			zap.String("orchestration", orchName))

		// Throttle to avoid flooding
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	params.Logger.Info("DispatchVerifiersAction: complete",
		zap.Int("dispatched", dispatched),
		zap.Int("errors", errors))

	return map[string]interface{}{
		"dispatched":    dispatched,
		"errors":        errors,
		"batch_name":    batchName,
		"topic":         topic,
		"vertical_slug": verticalSlug,
		"completed_at":  time.Now().UTC().Format(time.RFC3339),
	}, nil
}
