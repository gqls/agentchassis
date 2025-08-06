// test-website-builder.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"},
		Topic:   "system.agent.website-builder-orchestrator.process",
	})
	defer writer.Close()

	correlationID := uuid.New().String()

	payload := map[string]interface{}{
		"action": "create_website",
		"data": map[string]interface{}{
			"domain": "example-bakery.com",
			"business_name": "Example Bakery",
			"description": "Local artisan bakery specializing in sourdough",
			"requirements": map[string]interface{}{
				"style": "modern",
				"pages": ["home", "menu", "about", "contact"],
				"features": ["responsive", "seo-friendly"],
			},
		},
	}

	// Send to orchestrator
	// ...
}