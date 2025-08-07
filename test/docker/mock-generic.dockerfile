# test/docker/mock-generic.dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Build a simple mock agent
RUN cat > mock-agent.go << 'EOF'
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"

    "github.com/segmentio/kafka-go"
)

func main() {
    topic := os.Getenv("KAFKA_TOPIC")
    if topic == "" {
        topic = "system.agent.generic.process"
    }

    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{os.Getenv("KAFKA_BROKERS")},
        Topic:   topic,
        GroupID: "mock-generic-agent",
    })

    writer := kafka.NewWriter(kafka.WriterConfig{
        Brokers: []string{os.Getenv("KAFKA_BROKERS")},
        Topic:   "system.responses.generic",
    })

    defer reader.Close()
    defer writer.Close()

    log.Printf("Mock generic agent listening on %s", topic)

    for {
        msg, err := reader.ReadMessage(context.Background())
        if err != nil {
            log.Printf("Error reading message: %v", err)
            continue
        }

        // Simple echo response
        response := map[string]interface{}{
            "success": true,
            "agent": "mock-generic",
            "echo": string(msg.Value),
        }

        respBytes, _ := json.Marshal(response)

        // Copy headers and send response
        headers := make([]kafka.Header, len(msg.Headers))
        copy(headers, msg.Headers)

        writer.WriteMessages(context.Background(), kafka.Message{
            Key:     msg.Key,
            Value:   respBytes,
            Headers: headers,
        })

        log.Printf("Processed message: %s", msg.Key)
    }
}
EOF

RUN go mod init mock-agent && \
    go get github.com/segmentio/kafka-go && \
    go build -o mock-agent mock-agent.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/mock-agent /usr/local/bin/
ENTRYPOINT ["mock-agent"]