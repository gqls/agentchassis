// test/e2e/scenarios/website_builder_test.go
package scenarios

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Global test tracking
var (
	testsRun     []string
	testsFailed  []string
	testsSkipped []string
	testsPassed  []string
)

// TestMain provides detailed logging and panic recovery
func TestMain(m *testing.M) {
	fmt.Println("=== E2E TEST SUITE STARTING ===")
	fmt.Printf("Time: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Process ID: %d\n", os.Getpid())

	// Set up panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\n!!! PANIC DETECTED IN TEST SUITE !!!\n")
			fmt.Printf("Panic: %v\n", r)
			fmt.Printf("Stack trace:\n%s\n", debug.Stack())
			os.Exit(2)
		}
	}()

	// Run tests
	fmt.Println("\n>>> RUNNING TESTS...")
	code := m.Run()

	// Print summary
	fmt.Println("\n=== TEST EXECUTION SUMMARY ===")
	fmt.Printf("Tests Run: %d\n", len(testsRun))
	fmt.Printf("Tests Passed: %d\n", len(testsPassed))
	fmt.Printf("Tests Failed: %d\n", len(testsFailed))
	fmt.Printf("Tests Skipped: %d\n", len(testsSkipped))

	if len(testsFailed) > 0 {
		fmt.Println("\nFailed Tests:")
		for _, name := range testsFailed {
			fmt.Printf("  ✗ %s\n", name)
		}
	}

	if len(testsPassed) > 0 {
		fmt.Println("\nPassed Tests:")
		for _, name := range testsPassed {
			fmt.Printf("  ✓ %s\n", name)
		}
	}

	fmt.Println("\n=== E2E TEST SUITE COMPLETED ===")
	fmt.Printf("Exit code: %d\n", code)
	if code != 0 {
		fmt.Println("FAILURE DETECTED")
	} else {
		fmt.Println("SUCCESS")
	}

	os.Exit(code)
}

// testWrapper provides logging and panic recovery for individual tests
func testWrapper(t *testing.T, testName string, testFunc func(*testing.T)) {
	testsRun = append(testsRun, testName)

	fmt.Printf("\n>>> STARTING TEST: %s at %s\n", testName, time.Now().Format("15:04:05.000"))

	// Set up panic recovery for this test
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PANIC in %s: %v\nStack: %s", testName, r, debug.Stack())
			testsFailed = append(testsFailed, testName)
			fmt.Printf(">>> TEST FAILED (PANIC): %s\n", testName)
		}
	}()

	// Track test outcome
	defer func() {
		if t.Failed() {
			testsFailed = append(testsFailed, testName)
			fmt.Printf(">>> TEST FAILED: %s at %s\n", testName, time.Now().Format("15:04:05.000"))
		} else if t.Skipped() {
			testsSkipped = append(testsSkipped, testName)
			fmt.Printf(">>> TEST SKIPPED: %s\n", testName)
		} else {
			testsPassed = append(testsPassed, testName)
			fmt.Printf(">>> TEST PASSED: %s at %s\n", testName, time.Now().Format("15:04:05.000"))
		}
	}()

	// Run the actual test
	testFunc(t)
}

func TestWebsiteBuilderWorkflow(t *testing.T) {
	testWrapper(t, "TestWebsiteBuilderWorkflow", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping E2E test in short mode")
			return
		}

		t.Log("Step 1: Setting up test environment")
		ctx := context.Background()

		// Setup with detailed error checking
		t.Log("Step 1.1: Setting up test database...")
		db := setupTestDB(t)
		if db == nil {
			t.Fatal("Failed to setup test database - db is nil")
			return
		}
		defer func() {
			t.Log("Cleanup: Closing database connection")
			if err := db.Close(); err != nil {
				t.Logf("Warning: Error closing database: %v", err)
			}
		}()

		// Test database connection
		if err := db.Ping(); err != nil {
			t.Fatalf("Database ping failed: %v", err)
			return
		}
		t.Log("Step 1.2: Database connection verified")

		t.Log("Step 1.3: Setting up test producer...")
		producer := setupTestProducer(t)
		if producer == nil {
			t.Fatal("Failed to setup test producer - producer is nil")
			return
		}

		logger := zap.NewNop()

		t.Log("Step 1.4: Creating saga coordinator...")
		coordinator := orchestration.NewSagaCoordinator(db, producer, logger)
		if coordinator == nil {
			t.Fatal("Failed to create saga coordinator - coordinator is nil")
			return
		}

		// Setup website builder group in database
		t.Log("Step 2: Setting up website builder group...")
		setupWebsiteBuilderGroup(t, db)

		// Create website builder workflow
		t.Log("Step 3: Creating workflow plan...")
		workflow := models.WorkflowPlan{
			StartStep: "validate_request",
			Steps: map[string]models.Step{
				"validate_request": {
					Action:      "validate_input",
					Description: "Validate website requirements",
					NextStep:    "spawn_builders",
				},
				"spawn_builders": {
					Action: "spawn_group",
					Config: map[string]interface{}{
						"group_type": "website-builder",
					},
					NextStep: "plan_site",
				},
				"plan_site": {
					Action: "call_agent",
					Config: map[string]interface{}{
						"agent_type": "site-architect",
						"action":     "create_site_plan",
						"data": map[string]interface{}{
							"domain":        "test-site.com",
							"business_name": "Test Business",
							"requirements":  []string{"responsive", "modern", "fast"},
						},
					},
					NextStep: "transform_plan",
				},
				"transform_plan": {
					Action: "transform_data",
					Config: map[string]interface{}{
						"transformation": "uppercase",
					},
					NextStep: "notify",
				},
				"notify": {
					Action: "send_notification",
					Config: map[string]interface{}{
						"topic": "system.agent.website.responses",
					},
					NextStep: "complete",
				},
				"complete": {
					Action: "complete_workflow",
				},
			},
		}

		correlationID := helpers.TestUUIDWithType("e2e")
		t.Logf("Step 4: Generated correlation ID: %s", correlationID)

		headers := helpers.TestHeaders(correlationID)
		t.Logf("Step 5: Created headers with %d entries", len(headers))

		initialData, err := json.Marshal(map[string]interface{}{
			"action": "build_website",
			"data": map[string]interface{}{
				"message":       "Build a website",
				"business_name": "Test Business",
				"domain":        "test-site.com",
			},
		})
		if err != nil {
			t.Fatalf("Failed to marshal initial data: %v", err)
			return
		}

		// Execute workflow
		t.Log("Step 6: Executing workflow...")
		err = coordinator.ExecuteWorkflow(ctx, workflow, headers, initialData)
		if err != nil {
			t.Fatalf("ExecuteWorkflow failed: %v", err)
			return
		}
		t.Log("Step 6.1: Workflow execution initiated successfully")

		// Check initial state
		t.Log("Step 7: Getting workflow state...")
		repo := orchestration.NewStateRepository(db, logger)
		if repo == nil {
			t.Fatal("Failed to create state repository")
			return
		}

		state, err := repo.GetState(ctx, correlationID)
		if err != nil {
			t.Fatalf("Failed to get workflow state: %v", err)
			return
		}

		if state == nil {
			t.Fatal("State is nil after GetState")
			return
		}

		t.Logf("Step 7.1: Got state - Step: %s, Status: %s", state.CurrentStep, state.Status)

		assert.Equal(t, correlationID, state.CorrelationID)
		assert.NotEqual(t, "", state.CurrentStep)

		// Check workflow status
		validStatuses := []string{
			string(orchestration.StatusRunning),
			string(orchestration.StatusAwaitingResponses),
			string(orchestration.StatusCompleted),
		}

		if !assert.Contains(t, validStatuses, string(state.Status)) {
			t.Fatalf("Invalid status: %s", state.Status)
			return
		}

		// Handle remote actions if needed
		if state.Status == orchestration.StatusAwaitingResponses {
			t.Logf("Step 8: Handling %d awaited responses...", len(state.AwaitedSteps))

			for i, awaitedStep := range state.AwaitedSteps {
				t.Logf("Step 8.%d: Simulating response for step: %s", i+1, awaitedStep)

				response := models.TaskResponse{
					Data: map[string]interface{}{
						"status":    "success",
						"site_plan": "Generated site plan",
						"result":    "success",
					},
					Error: "",
				}

				responseData, _ := json.Marshal(response)
				responseHeaders := make(map[string]string)
				for k, v := range headers {
					responseHeaders[k] = v
				}
				responseHeaders["causation_id"] = awaitedStep

				err = coordinator.HandleResponse(ctx, responseHeaders, responseData)
				if err != nil {
					t.Errorf("Failed to handle response for step %s: %v", awaitedStep, err)
				}
			}

			// Check state after responses
			time.Sleep(100 * time.Millisecond)
			state, _ = repo.GetState(ctx, correlationID)
			t.Logf("Step 9: Final state after responses: %s", state.Status)
		}

		t.Log("Step 10: Test completed successfully")
	})
}

func TestWebsiteBuilderWithErrors(t *testing.T) {
	testWrapper(t, "TestWebsiteBuilderWithErrors", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping E2E test in short mode")
			return
		}

		t.Log("Starting error handling tests")

		ctx := context.Background()

		db := setupTestDB(t)
		if db == nil {
			t.Fatal("Failed to setup test database")
			return
		}
		defer db.Close()

		producer := setupTestProducer(t)
		logger := zap.NewNop()
		coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

		tests := []struct {
			name          string
			invalidData   map[string]interface{}
			expectedError string
		}{
			{
				name: "Missing message field",
				invalidData: map[string]interface{}{
					"action": "build",
					"data": map[string]interface{}{
						"domain": "test.com",
					},
				},
				expectedError: "missing required field: message",
			},
			{
				name: "Invalid JSON",
				invalidData: map[string]interface{}{
					"invalid": "json_structure",
				},
				expectedError: "invalid",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Logf("Running subtest: %s", tt.name)

				workflow := models.WorkflowPlan{
					StartStep: "validate",
					Steps: map[string]models.Step{
						"validate": {
							Action:   "validate_input",
							NextStep: "complete",
						},
						"complete": {
							Action: "complete_workflow",
						},
					},
				}

				correlationID := helpers.TestUUIDWithType("e2e")
				headers := helpers.TestHeaders(correlationID)
				initialData, _ := json.Marshal(tt.invalidData)

				_ = coordinator.ExecuteWorkflow(ctx, workflow, headers, initialData)

				repo := orchestration.NewStateRepository(db, logger)
				state, stateErr := repo.GetState(ctx, correlationID)

				if stateErr == nil && state != nil {
					if state.Status == orchestration.StatusFailed {
						assert.Contains(t, state.Error, tt.expectedError)
						t.Logf("Correctly caught error: %s", state.Error)
					}
				}
			})
		}

		t.Log("Error handling tests completed")
	})
}

func setupWebsiteBuilderGroup(t *testing.T, db *sql.DB) {
	// Create agent configs properly
	agentConfigs := []map[string]interface{}{
		{
			"role":       "architect",
			"agent_type": "site-architect",
		},
		{
			"role":       "designer",
			"agent_type": "visual-designer",
		},
		{
			"role":       "developer",
			"agent_type": "html-developer",
		},
		{
			"role":       "publisher",
			"agent_type": "site-publisher",
		},
	}
	agentConfigsJSON, _ := json.Marshal(agentConfigs)

	// Create workflow steps properly
	planStep := map[string]interface{}{
		"action":    "create_plan",
		"next_step": "design",
	}

	designStep := map[string]interface{}{
		"action":    "create_design",
		"next_step": "develop",
	}

	developStep := map[string]interface{}{
		"action":    "build_html",
		"next_step": "publish",
	}

	publishStep := map[string]interface{}{
		"action": "publish_site",
	}

	workflowSteps := map[string]interface{}{
		"plan":    planStep,
		"design":  designStep,
		"develop": developStep,
		"publish": publishStep,
	}

	workflowData := map[string]interface{}{
		"start_step": "plan",
		"steps":      workflowSteps,
	}

	workflowJSON, _ := json.Marshal(workflowData)

	_, err := db.Exec(`
		INSERT INTO agent_groups (id, name, group_type, version, agent_configs, orchestration_workflow)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, uuid.New().String(), "Website Builder Team", "website-builder", "1.0.0",
		agentConfigsJSON, workflowJSON)

	if err != nil {
		t.Logf("Warning: Could not insert website builder group: %v", err)
	}
}
