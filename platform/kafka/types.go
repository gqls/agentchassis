// FILE: platform/kafka/types.go
package kafka

import (
	"github.com/segmentio/kafka-go"
)

// Message wraps kafka-go Message to implement our interface
type Message = kafka.Message

// Header wraps kafka-go Header
type Header = kafka.Header
