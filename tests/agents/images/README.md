# 7. Check for errors in the error topic

kubectl exec -it personae-kafka-cluster-combined-pool-prod-0 -n kafka -- \
kafka-console-consumer.sh \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.errors.image \
--from-beginning \
--max-messages 10

kubectl exec -it kafka-client-test -n kafka -- sh
sh-5.1$ ls /usr/bin/kafka*

cat > test-image.go << 'EOF'
package main

import (
"context"
"encoding/json"
"fmt"
"log"
"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type ImageGenerationRequest struct {
Action string `json:"action"`
Data   struct {
Prompt      string  `json:"prompt"`
AspectRatio string  `json:"aspect_ratio,omitempty"`
Style       string  `json:"style,omitempty"`
Seed        float64 `json:"seed,omitempty"`
} `json:"data"`
}

func main() {
// Kafka configuration
brokers := []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"}
requestTopic := "system.adapter.image.generate"
responseTopic := "system.responses.image"

	// Create a unique correlation ID
	correlationID := uuid.New().String()
	requestID := uuid.New().String()

	// Create the request
	request := ImageGenerationRequest{
		Action: "generate_image",
	}
	request.Data.Prompt = "A beautiful sunset over mountains with a lake in the foreground, photorealistic style"

	requestBytes, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	// Create Kafka writer
	writer := kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    requestTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// Create headers
	headers := []kafka.Header{
		{Key: "correlation_id", Value: []byte(correlationID)},
		{Key: "request_id", Value: []byte(requestID)},
		{Key: "client_id", Value: []byte("test-client")},
		{Key: "agent_instance_id", Value: []byte("test-instance")},
		{Key: "timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
	}

	// Send the message
	msg := kafka.Message{
		Key:     []byte(correlationID),
		Value:   requestBytes,
		Headers: headers,
	}

	fmt.Printf("Sending image generation request...\n")
	fmt.Printf("Correlation ID: %s\n", correlationID)
	fmt.Printf("Prompt: %s\n", request.Data.Prompt)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := writer.WriteMessages(ctx, msg); err != nil {
		log.Fatalf("Failed to write message: %v", err)
	}

	fmt.Println("Message sent successfully!")
	fmt.Printf("\nNow listening for response on topic: %s\n", responseTopic)

	// Create a reader for responses
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   responseTopic,
		GroupID: "test-consumer-" + uuid.New().String(),
	})
	defer reader.Close()

	// Read the response
	ctx2, cancel2 := context.WithTimeout(context.Background(), 120*time.Second) // 2 minutes for image generation
	defer cancel2()

	for {
		msg, err := reader.FetchMessage(ctx2)
		if err != nil {
			log.Printf("Timeout or error waiting for response: %v", err)
			break
		}

		// Check if this is our response
		for _, h := range msg.Headers {
			if h.Key == "correlation_id" && string(h.Value) == correlationID {
				fmt.Println("\nReceived response!")

				var response map[string]interface{}
				if err := json.Unmarshal(msg.Value, &response); err != nil {
					log.Printf("Failed to unmarshal response: %v", err)
				} else {
					responseJSON, _ := json.MarshalIndent(response, "", "  ")
					fmt.Printf("Response: %s\n", responseJSON)
				}

				// Commit the message
				reader.CommitMessages(context.Background(), msg)
				return
			}
		}
	}
}

EOF

go mod init test-image
go get github.com/segmentio/kafka-go
go get github.com/google/uuid
go run test-image.go


# Generate IDs
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)

# Create a test that will at least show proper error handling
cat > /tmp/test-msg.json << EOF
{"action":"generate_image","data":{"prompt":"A beautiful mountain landscape"}}
EOF

# Send with key (correlation_id as key)
kafka-console-producer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.adapter.image.generate \
--property parse.key=true \
--property key.separator="|" << EOF
${CORRELATION_ID}|{"action":"generate_image","data":{"prompt":"A beautiful mountain landscape"}}
EOF

# 5. Monitor the response topic for errors:
# In another terminal, watch for responses
kubectl exec -it kafka-client-test -n kafka -- \
kafka-console-consumer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.responses.image \
--from-beginning \
--max-messages 5

# 6. Check the error topic:
kubectl exec -it kafka-client-test -n kafka -- \
kafka-console-consumer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.errors.image-generator \
--from-beginning \
--max-messages 5

# check the adapter is getting the key correctly
# Check the pod environment
kubectl exec -n ai-persona-system deployment/image-generator-adapter -- env | grep STABILITY

# 1. List all topics:
kubectl exec -it kafka-client-test -n kafka -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--list | grep -E "(image|adapter)"

# 2. Check if the topic exists exactly as expected:
kubectl exec -it kafka-client-test -n kafka -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--describe --topic system.adapter.image.generate

# 3. Check consumer group status:
kubectl exec -it kafka-client-test -n kafka -- kafka-consumer-groups \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--group image-generator-adapter-group \
--describe

# 4. List all consumer groups to see if your adapter is registered:
kubectl exec -it kafka-client-test -n kafka -- kafka-consumer-groups \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--list

# this one
# 5. Check if there are messages in the topic:
kubectl exec -it kafka-client-test -n kafka -- kafka-console-consumer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.adapter.image.generate \
--from-beginning \
--max-messages 5 \
--timeout-ms 5000

# 6. Check for any offset/lag issues:
kubectl exec -it kafka-client-test -n kafka -- kafka-consumer-groups \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--group image-generator-adapter-group \
--describe --members --verbose

# Good! The messages are in the topic. Now let's check why the consumer isn't picking them up:
# 1. Check the consumer group status:
# This will show if the consumer group exists and what offset it's at.
kubectl exec -it kafka-client-test -n kafka -- kafka-consumer-groups \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--group image-generator-adapter-group \
--describe

# 2. Reset the consumer group offset to beginning:
#   If the consumer group already exists and has consumed past these messages:
# First, stop the adapter
kubectl scale deployment/image-generator-adapter -n ai-persona-system --replicas=0

# Wait for pods to terminate
kubectl wait --for=delete pod -l app=image-generator-adapter -n ai-persona-system --timeout=60s

# Reset the consumer group offset
kubectl exec -it kafka-client-test -n kafka -- kafka-consumer-groups \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--group image-generator-adapter-group \
--reset-offsets \
--to-earliest \
--topic system.adapter.image.generate \
--execute

# Scale back up
kubectl scale deployment/image-generator-adapter -n ai-persona-system --replicas=3