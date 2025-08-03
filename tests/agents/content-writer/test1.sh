kubectl run -n ai-persona-system test-content-creator --rm -i --tty \
  --image=golang:1.21-alpine \
  --restart=Never -- sh

--
apk add --no-cache vim
cd /
mkdir app
cd app

vim to edit test.go

package main

import (
"context"
"fmt"
"log"
"time"

"github.com/segmentio/kafka-go"
)

func main() {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"},
			Topic:   "system.agent.content-creator.process",
		})
		defer writer.Close()

		msg := kafka.Message{
		Headers: []kafka.Header{
	{Key: "correlation_id", Value: []byte("323e4567-e89b-12d3-a456-426614174000")},
	{Key: "request_id", Value: []byte("423e4567-e89b-12d3-a456-426614174000")},
		{Key: "client_id", Value: []byte("demo_client")},
		{Key: "agent_instance_id", Value: []byte("content-creator-instance")},
		{Key: "fuel_budget", Value: []byte("100")},
		},
		Key:   []byte("content-test"),
		Value: []byte(`{
            "action": "generate_content",
            "data": {
                "topic": "AI and the Future of Work",
                "content_type": "blog_post",
                "style": "informative",
                "length": "short",
                "keywords": ["AI", "automation", "jobs", "future"]
            }
        }`),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := writer.WriteMessages(ctx, msg)
		if err != nil {
		log.Fatal("failed to write message:", err)
		}

		fmt.Println("Content generation request sent!")
		}
EOL
' > test.go

go mod init test
go get github.com/segmentio/kafka-go
go run test.go

--
# check the response
  kubectl exec -it kafka-client-test -n kafka -- kafka-console-consumer \
    --bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
    --topic system.responses.content-creator \
    --from-beginning