// FILE: platform/kafka/utils.go
package kafka

import "github.com/segmentio/kafka-go"

// HeadersToMap converts Kafka headers to a map for easier access
func HeadersToMap(headers []kafka.Header) map[string]string {
	result := make(map[string]string)
	for _, h := range headers {
		result[h.Key] = string(h.Value)
	}
	return result
}
