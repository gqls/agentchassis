// platform/orchestration/ui_notifications.go
package orchestration

import (
	"context"
	"encoding/json"
	"github.com/gqls/agentchassis/platform/kafka"
)

type UINotification struct {
	Type          string                 `json:"type"`
	CorrelationID string                 `json:"correlation_id"`
	Component     string                 `json:"component"`
	Data          map[string]interface{} `json:"data"`
	Actions       []UIAction             `json:"actions"`
}

type UIAction struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Type    string `json:"type"` // approve, reject, modify
	Default bool   `json:"default"`
}

func sendTeamApprovalRequest(ctx context.Context, producer kafka.Producer, correlationID string, proposal interface{}) error {
	notification := UINotification{
		Type:          "APPROVAL_REQUIRED",
		CorrelationID: correlationID,
		Component:     "team_builder",
		Data: map[string]interface{}{
			"proposal":       proposal,
			"estimated_cost": calculateEstimatedFuel(proposal),
			"similar_teams":  findSimilarTeams(proposal),
		},
		Actions: []UIAction{
			{ID: "approve", Label: "Approve Team", Type: "approve", Default: true},
			{ID: "modify", Label: "Modify Team", Type: "modify"},
			{ID: "reject", Label: "Use Default", Type: "reject"},
		},
	}

	return sendNotification(ctx, producer, notification)
}

// Ahelper functions
func sendNotification(ctx context.Context, producer kafka.Producer, notification UINotification) error {
	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"notification_type": notification.Type,
		"correlation_id":    notification.CorrelationID,
	}

	return producer.Produce(ctx, NotificationTopic, headers,
		[]byte(notification.CorrelationID), notificationBytes)
}

func calculateEstimatedFuel(proposal interface{}) int {
	// Simple estimation based on number of agents
	if p, ok := proposal.(map[string]interface{}); ok {
		if agents, ok := p["agents"].([]interface{}); ok {
			return len(agents) * 100 // 100 fuel per agent
		}
	}
	return 500 // default
}

func findSimilarTeams(proposal interface{}) []map[string]interface{} {
	// In a real implementation, this would query the database
	// For now, return empty
	return []map[string]interface{}{}
}
