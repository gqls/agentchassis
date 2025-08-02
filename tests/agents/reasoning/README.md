If the kafka-client-test pod doesn't have Go installed, let's use a temporary pod:

kubectl run reasoning-test --rm -it --image=golang:1.21-alpine -n ai-persona-system --restart=Never -- sh -c '
apk add --no-cache git
cat > /tmp/test.go << "EOF"
package main

import (
"context"
"encoding/json"
"fmt"
"log"
"time"
"github.com/segmentio/kafka-go"
)

func main() {
writer := &kafka.Writer{
Addr:     kafka.TCP("personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"),
Topic:    "system.agent.reasoning.process",
Balancer: &kafka.LeastBytes{},
}
defer writer.Close()

    correlationID := fmt.Sprintf("test-%d", time.Now().Unix())
    
    payload := map[string]interface{}{
        "action": "analyze",
        "data": map[string]interface{}{
            "content_to_review": "The new product launch strategy focuses on social media marketing, influencer partnerships, and targeted email campaigns. We plan to allocate 60% of the budget to digital channels and 40% to traditional media.",
            "review_criteria": []string{
                "Budget allocation appropriateness",
                "Channel mix effectiveness",
                "Target audience alignment",
                "ROI potential",
            },
            "brief_context": map[string]interface{}{
                "product_type":   "B2C software",
                "target_market":  "millennials and gen-z", 
                "budget":         100000,
                "timeline":       "3 months",
            },
        },
    }

    jsonBytes, _ := json.Marshal(payload)
    
    msg := kafka.Message{
        Key:   []byte(correlationID),
        Value: jsonBytes,
        Headers: []kafka.Header{
            {Key: "correlation_id", Value: []byte(correlationID)},
            {Key: "request_id", Value: []byte(fmt.Sprintf("req-%d", time.Now().Unix()))},
            {Key: "client_id", Value: []byte("test-client")},
            {Key: "agent_instance_id", Value: []byte("reasoning-001")},
            {Key: "fuel_budget", Value: []byte("100")},
        },
    }

    fmt.Printf("Sending message with correlation_id: %s\n", correlationID)
    fmt.Printf("Payload: %s\n", string(jsonBytes))
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    err := writer.WriteMessages(ctx, msg)
    if err != nil {
        log.Fatalf("Failed to write message: %v", err)
    }
    
    fmt.Println("Message sent successfully!")
}
EOF
cd /tmp && go mod init test && go get github.com/segmentio/kafka-go@v0.4.47 && go run test.go
'


=====
