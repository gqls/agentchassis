// test/performance/benchmarks/agent_group_benchmark_test.go
package benchmarks

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/test/unit/helpers"
)

func BenchmarkAgentGroupSpawning(b *testing.B) {
	db := helpers.TestDB(&testing.T{})
	defer db.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		params := actions.ActionParams{
			StepConfig: models.Step{
				Config: map[string]interface{}{
					"group_type": "website-builder",
				},
			},
			Headers: map[string]string{
				"correlation_id": fmt.Sprintf("bench-group-%d", i),
				"client_id":      "demo_client",
				"user_id":        "bench_user",
			},
			DB:     db,
			Logger: zap.NewNop(),
		}

		_, err := actions.SpawnGroupAction(context.Background(), params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGroupCoordination(b *testing.B) {
	db := helpers.TestDB(&testing.T{})
	defer db.Close()

	// Pre-spawn a group
	groupID := setupTestGroup(b, db)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Benchmark coordinating tasks across group
		simulateGroupCoordination(b, db, groupID)
	}
}
