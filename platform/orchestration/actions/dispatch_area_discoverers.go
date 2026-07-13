// FILE: platform/orchestration/actions/dispatch_area_discoverers.go
//
// DispatchAreaDiscoverersAction reads the list of unswept areas from
// collected_data (output of load_unswept_areas) and produces one Kafka
// message per district to trigger area-sweep-discoverer agents.
//
// Complexity is here in Go; the workflow step is simple.
//
// Workflow config:
//
//	"dispatch_discoverers": {
//	    "action": "dispatch_area_discoverers",
//	    "config": {
//	        "input_fields": ["unswept_areas"]
//	    },
//	    "output_field": "dispatch_result",
//	    "next_step": "complete"
//	}
//
// Registration:
//   "dispatch_area_discoverers": DispatchAreaDiscoverersAction,

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

var DispatchAreaDiscoverersInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"unswept_areas"},
	Optional: []string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("dispatch_area_discoverers", DispatchAreaDiscoverersInputSpec)
}

func DispatchAreaDiscoverersAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("DispatchAreaDiscoverersAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.Producer == nil {
		return nil, fmt.Errorf("no Kafka producer available")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		DispatchAreaDiscoverersInputSpec,
		params.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	unsweptAreas := inputs.GetMap("unswept_areas")
	if unsweptAreas == nil {
		return nil, fmt.Errorf("unswept_areas not found in collected data")
	}

	// areas is the array inside the unswept_areas output
	areasRaw, ok := unsweptAreas["areas"].([]interface{})
	if !ok || len(areasRaw) == 0 {
		params.Logger.Info("DispatchAreaDiscoverersAction: no areas to dispatch")
		return map[string]interface{}{
			"dispatched": 0,
			"errors":     0,
			"message":    "no areas to sweep",
		}, nil
	}

	businessType, _ := unsweptAreas["business_type"].(string)
	if businessType == "" {
		businessType = "veterinary practice"
	}

	batchName := fmt.Sprintf("sweep-%s", time.Now().Format("20060102-150405"))
	topic := "system.agent.generic.requests"
	clientID := params.ExecutionContext.ClientID

	params.Logger.Info("DispatchAreaDiscoverersAction: dispatching",
		zap.Int("area_count", len(areasRaw)),
		zap.String("batch", batchName),
		zap.String("topic", topic))

	dispatched := 0
	errors := 0

	for _, areaRaw := range areasRaw {
		area, ok := areaRaw.(map[string]interface{})
		if !ok {
			errors++
			continue
		}

		districtCode, _ := area["district_code"].(string)
		areaName, _ := area["area_name"].(string)
		searchAreaID, _ := area["search_area_id"].(string)

		if districtCode == "" {
			errors++
			continue
		}

		orchID := uuid.New().String()
		reqID := uuid.New().String()
		orchName := fmt.Sprintf("%s-%s", batchName, districtCode)

		// Message body — same format the generic agent expects
		body := map[string]interface{}{
			"action": "orchestrate",
			"config": map[string]interface{}{
				"agent_type": "area-sweep-discoverer",
			},
			"input_data": map[string]interface{}{
				"district_code":  districtCode,
				"area_name":      areaName,
				"search_area_id": searchAreaID,
				"business_type":  businessType,
			},
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			params.Logger.Warn("Failed to marshal message",
				zap.String("district", districtCode), zap.Error(err))
			errors++
			continue
		}

		// Headers — same format as bulk_vet_verify.sh
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
			params.Logger.Warn("Failed to produce message",
				zap.String("district", districtCode), zap.Error(err))
			errors++
			continue
		}

		dispatched++
		params.Logger.Info("Dispatched discoverer",
			zap.String("district", districtCode),
			zap.String("orchestration", orchName))

		// Small delay to avoid flooding
		time.Sleep(100 * time.Millisecond)
	}

	params.Logger.Info("DispatchAreaDiscoverersAction: complete",
		zap.Int("dispatched", dispatched),
		zap.Int("errors", errors))

	return map[string]interface{}{
		"dispatched": dispatched,
		"errors":     errors,
		"batch_name": batchName,
		"topic":      topic,
	}, nil
}
