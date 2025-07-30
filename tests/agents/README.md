
https://claude.ai/chat/878cdc53-1a28-4598-beb9-0e0926b63bb8

~/projects/agent-chassis/tests/agents$ kubectl apply -f content-creator-test.yaml 

# Copy the test file into the pod
kubectl cp tests/agents/test_content_creator.go ai-persona-system/content-creator-test:/test.go

# Exec into the pod
kubectl exec -it -n ai-persona-system content-creator-test -- /bin/sh

# Inside the pod
cd /
go mod init test
go get github.com/segmentio/kafka-go
go get github.com/google/uuid
go get go.uber.org/zap

# Run the test
go run test.go

# Check if the content creator topics exist
kubectl exec -it -n kafka kafka-client-test -- kafka-topics --bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 --list | grep content-creator

# Check consumer groups
kubectl exec -it -n kafka kafka-client-test -- kafka-consumer-groups --bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 --list

# Check if the message is in the topic
kubectl exec -it -n kafka kafka-client-test -- kafka-console-consumer --bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 --topic system.agent.content-creator.process --from-beginning --max-messages 1

kubectl exec -it -n ai-persona-system content-creator-test -- /bin/sh
# Send a request and manually create the response
cd /
cat > full-flow-test.go << 'EOF'
package main

import (
"bytes"
"context"
"encoding/json"
"fmt"
"io"
"net/http"
"time"

    "github.com/google/uuid"
    "github.com/segmentio/kafka-go"
)

func main() {
bootstrapServers := "personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

    // 1. Send a request
    correlationID := uuid.NewString()
    fmt.Printf("Sending request with correlation ID: %s\n", correlationID)
    
    request := map[string]interface{}{
        "action": "generate_content",
        "data": map[string]interface{}{
            "topic": "Benefits of Remote Work",
            "content_type": "blog_post",
            "length": "short",
            "style": "informative",
        },
    }
    
    writer := &kafka.Writer{
        Addr:     kafka.TCP(bootstrapServers),
        Topic:    "system.agent.content-creator.process",
        Balancer: &kafka.LeastBytes{},
    }
    
    reqBytes, _ := json.Marshal(request)
    err := writer.WriteMessages(context.Background(), kafka.Message{
        Key:   []byte(correlationID),
        Value: reqBytes,
        Headers: []kafka.Header{
            {Key: "correlation_id", Value: []byte(correlationID)},
            {Key: "fuel_budget", Value: []byte("1000")},
            {Key: "client_id", Value: []byte("demo_client")},
        },
    })
    
    if err != nil {
        fmt.Printf("Failed to send request: %v\n", err)
        return
    }
    
    fmt.Println("Request sent, waiting 5 seconds for agent to process...")
    time.Sleep(5 * time.Second)
    
    // 2. Call Anthropic directly
    fmt.Println("Calling Anthropic API...")
    apiKey := ""
    
    payload := map[string]interface{}{
        "model": "claude-3-5-sonnet-20241022",
        "messages": []map[string]interface{}{
            {"role": "user", "content": "Write a brief blog post introduction about the benefits of home veterinary (100 words)"},
        },
        "max_tokens": 200,
    }
    
    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
    req.Header.Set("x-api-key", apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")
    req.Header.Set("content-type", "application/json")
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Printf("API Error: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    respBody, _ := io.ReadAll(resp.Body)
    var result map[string]interface{}
    json.Unmarshal(respBody, &result)
    
    text := result["content"].([]interface{})[0].(map[string]interface{})["text"].(string)
    fmt.Printf("\nGenerated: %s\n", text)
    
    // 3. Send the response
    response := map[string]interface{}{
        "success": true,
        "generated_text": text,
        "metadata": map[string]interface{}{
            "model": "claude-3-5-sonnet-20241022",
            "manual_flow": true,
        },
    }
    
    respBytes, _ := json.Marshal(response)
    err = writer.WriteMessages(context.Background(), kafka.Message{
        Key:   []byte(correlationID),
        Value: respBytes,
    })
    
    fmt.Printf("\nResponse sent with same correlation ID: %s\n", correlationID)
}
EOF
go mod init test
go get github.com/segmentio/kafka-go
go get github.com/google/uuid
go get go.uber.org/zap

go run full-flow-test.go


--
kubectl exec -it -n kafka kafka-client-test -- bash

kafka-console-consumer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.responses.content-creator \
--property print.key=true


==

# see what's in the topic
# Check for old messages without headers
kubectl exec -it -n kafka kafka-client-test -- kafka-console-consumer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.agent.content-creator.process \
--from-beginning \
--max-messages 5 \
--property print.headers=true \
--property print.key=true

# check topics current retention settings (didn't work)
kubectl exec -it -n kafka kafka-client-test -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--describe \
--topic system.agent.content-creator.process | grep -E "(retention|segment)"

# delete and recreate the topic if no important data
# Delete the topic
kubectl exec -it -n kafka kafka-client-test -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--delete \
--topic system.agent.content-creator.process

# Recreate it
kubectl exec -it -n kafka kafka-client-test -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--create \
--topic system.agent.content-creator.process \
--partitions 6 \
--replication-factor 1 \
--config retention.ms=86400000