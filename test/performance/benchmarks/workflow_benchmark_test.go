// test/performance/benchmarks/workflow_benchmark_test.go
package benchmarks

import (
	"context"
	"fmt"
	"github.com/gqls/agentchassis/platform/discovery"
	"go.uber.org/zap"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
)

func BenchmarkWorkflowExecution(b *testing.B) {
	db := helpers.TestDB(&testing.T{})
	defer db.Close()

	producer := helpers.NewMockProducer()
	coordinator := orchestration.NewSagaCoordinator(db, producer, zap.NewNop())

	workflow := helpers.ValidWorkflow()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		correlationID := fmt.Sprintf("bench-%d-%d", b.N, i)
		headers := helpers.TestHeaders(correlationID)

		err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAgentDiscovery(b *testing.B) {
	db := helpers.TestDB(&testing.T{})
	defer db.Close()

	discovery := discovery.NewAgentDiscovery(convertToPool(db))

	requirements := discovery.Requirements{
		AgentType: "researcher",
		ClientID:  "demo_client",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		matches, err := discovery.DiscoverAgents(context.Background(), requirements)
		if err != nil {
			b.Fatal(err)
		}
		if len(matches) == 0 {
			b.Fatal("No agents found")
		}
	}
}
