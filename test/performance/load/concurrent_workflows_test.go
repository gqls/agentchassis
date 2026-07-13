// test/performance/load/concurrent_workflows_test.go
package load

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentWorkflowExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	producer := createRealKafkaProducer(t)
	defer producer.Close()

	coordinator := orchestration.NewSagaCoordinator(db, producer, zap.NewNop())

	// Test parameters
	concurrentWorkflows := 50
	workflowDuration := 30 * time.Second

	// Metrics
	var successCount int64
	var failureCount int64
	var totalDuration int64

	// Create workflow
	workflow := models.WorkflowPlan{
		StartStep: "init",
		Steps: map[string]models.Step{
			"init": {
				Action:   "validate_input",
				NextStep: "process",
			},
			"process": {
				Action: "call_agent",
				Topic:  "system.agent.generic.process",
				Config: map[string]interface{}{
					"simulate_duration": "1s",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	// Run concurrent workflows
	wg := sync.WaitGroup{}
	wg.Add(concurrentWorkflows)

	startTime := time.Now()

	for i := 0; i < concurrentWorkflows; i++ {
		go func(index int) {
			defer wg.Done()

			workflowStart := time.Now()
			correlationID := fmt.Sprintf("test-load-%d-%d", time.Now().Unix(), index)
			headers := helpers.TestHeaders(correlationID)

			err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, nil)

			duration := time.Since(workflowStart)
			atomic.AddInt64(&totalDuration, duration.Nanoseconds())

			if err != nil {
				atomic.AddInt64(&failureCount, 1)
				t.Logf("Workflow %d failed: %v", index, err)
			} else {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)

		// Stagger starts slightly
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// Calculate metrics
	successRate := float64(successCount) / float64(concurrentWorkflows) * 100
	avgDuration := time.Duration(totalDuration / int64(concurrentWorkflows))

	// Assertions
	assert.GreaterOrEqual(t, successRate, 95.0, "Success rate should be at least 95%")
	assert.Less(t, avgDuration, 5*time.Second, "Average workflow duration should be under 5 seconds")

	// Report
	t.Logf("Load Test Results:")
	t.Logf("  Total workflows: %d", concurrentWorkflows)
	t.Logf("  Successful: %d (%.1f%%)", successCount, successRate)
	t.Logf("  Failed: %d", failureCount)
	t.Logf("  Total time: %v", totalTime)
	t.Logf("  Average duration: %v", avgDuration)
	t.Logf("  Throughput: %.2f workflows/second", float64(concurrentWorkflows)/totalTime.Seconds())
}

func TestWorkflowScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	testCases := []int{10, 50, 100, 200}

	for _, workflowCount := range testCases {
		t.Run(fmt.Sprintf("%d_workflows", workflowCount), func(t *testing.T) {
			runScalabilityTest(t, db, workflowCount)
		})
	}
}

func TestResourceContention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource contention test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	// Monitor resource usage
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics := &ResourceMetrics{}
	go monitorResources(ctx, db, metrics)

	// Run workflows that compete for same resources
	concurrency := 20
	wg := sync.WaitGroup{}
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()

			// All workflows try to spawn same agent type
			workflow := models.WorkflowPlan{
				StartStep: "spawn",
				Steps: map[string]models.Step{
					"spawn": {
						Action: "spawn_agent",
						Config: map[string]interface{}{
							"agent_type": "shared-resource",
						},
						NextStep: "use_agent",
					},
					"use_agent": {
						Action: "call_agent",
						Config: map[string]interface{}{
							"agent_type": "shared-resource",
						},
						NextStep: "complete",
					},
					"complete": {
						Action: "complete_workflow",
					},
				},
			}

			correlationID := fmt.Sprintf("test-contention-%d", index)
			executeWorkflow(t, workflow, correlationID)
		}(i)
	}

	wg.Wait()

	// Verify no deadlocks or excessive contention
	assert.Less(t, metrics.MaxConnections, 50)
	assert.Less(t, metrics.MaxLockWaitTime, 5*time.Second)
}

type ResourceMetrics struct {
	MaxConnections  int
	MaxLockWaitTime time.Duration
	mu              sync.Mutex
}

func monitorResources(ctx context.Context, db *sql.DB, metrics *ResourceMetrics) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := db.Stats()

			metrics.mu.Lock()
			if stats.OpenConnections > metrics.MaxConnections {
				metrics.MaxConnections = stats.OpenConnections
			}
			metrics.mu.Unlock()
		}
	}
}
