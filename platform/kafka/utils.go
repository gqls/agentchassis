// FILE: platform/kafka/utils.go
package kafka

import "github.com/segmentio/kafka-go"

// HeadersToMap converts Kafka headers to a map
func HeadersToMap(headers []kafka.Header) map[string]string {
	result := make(map[string]string)
	for _, h := range headers {
		result[h.Key] = string(h.Value)
	}
	return result
}

// MapToHeaders converts a map to Kafka headers
func MapToHeaders(m map[string]string) []kafka.Header {
	headers := make([]kafka.Header, 0, len(m))
	for k, v := range m {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}
	return headers
}
