// FILE: platform/kafka/topic_manager_controller_test.go
package kafka

import (
	"testing"

	"github.com/segmentio/kafka-go"
)

// bugs_open/040-kafka-dial: controller.Host coming back empty must be
// rejected, not formatted into the dial target ":<port>".
func TestControllerAddress(t *testing.T) {
	tests := []struct {
		name    string
		broker  kafka.Broker
		want    string
		wantErr bool
	}{
		{
			name:   "valid host",
			broker: kafka.Broker{Host: "personae-kafka-cluster-combined-pool-prod-0.personae-kafka-cluster-kafka-brokers.kafka.svc", Port: 9092},
			want:   "personae-kafka-cluster-combined-pool-prod-0.personae-kafka-cluster-kafka-brokers.kafka.svc:9092",
		},
		{
			name:    "empty host is rejected, not formatted into :port",
			broker:  kafka.Broker{Host: "", Port: 9092},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := controllerAddress(tt.broker)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("controllerAddress(%+v) = %q, nil; want an error", tt.broker, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("controllerAddress(%+v) unexpected error: %v", tt.broker, err)
			}
			if got != tt.want {
				t.Fatalf("controllerAddress(%+v) = %q, want %q", tt.broker, got, tt.want)
			}
		})
	}
}
