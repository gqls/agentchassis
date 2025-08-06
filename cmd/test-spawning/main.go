// cmd/test-spawning/main.go
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Connect to database
	dbURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Connect to Kafka
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	producer, err := kafka.NewProducer(brokers, logger)
	if err != nil {
		log.Fatal("Failed to create Kafka producer:", err)
	}
	defer producer.Close()

	clientID := os.Getenv("CLIENT_ID")
	if clientID == "" {
		clientID = "demo_client"
	}

	fmt.Println("🧪 Testing Agent Spawning System")
	fmt.Println("================================")

	// Test 1: Direct spawn
	testDirectSpawn(db, logger, clientID)

	// Test 2: Spawn with orchestration
	testSpawnWithOrchestration(db, producer, logger, clientID)

	// Test 3: Group spawning
	testGroupSpawning(db, producer, logger, clientID)

	fmt.Println("\n✅ All tests completed!")
}

func testDirectSpawn(db *sql.DB, logger *zap.Logger, clientID string) {
	fmt.Println("\n📌 Test 1: Direct Agent Spawn")

	params := actions.ActionParams{
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"agent_type": "researcher",
			},
		},
		Headers: map[string]string{
			"client_id": clientID,
			"user_id":   "test_user",
		},
		DB:     db,
		Logger: logger,
	}

	result, err := actions.SpawnAgentAction(context.Background(), params)
	if err != nil {
		log.Printf("❌ SpawnAgentAction failed: %v", err)
		return
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("✅ Agent spawned successfully:\n%s\n", resultJSON)

	// Verify in database
	var count int
	err = db.QueryRow(fmt.Sprintf(`
        SELECT COUNT(*) FROM client_%s.agent_instances 
        WHERE config->>'agent_type' = 'researcher' AND is_active = true
    `, clientID)).Scan(&count)

	if err == nil && count > 0 {
		fmt.Printf("✅ Verified: %d researcher agent(s) in database\n", count)
	}
}

func testSpawnWithOrchestration(db *sql.DB, producer kafka.Producer, logger *zap.Logger, clientID string) {
	fmt.Println("\n📌 Test 2: Spawn via Orchestration")

	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	workflow := models.WorkflowPlan{
		StartStep: "spawn",
		Steps: map[string]models.Step{
			"spawn": {
				Action: "spawn_agent",
				Config: map[string]interface{}{
					"agent_type": "developer",
				},
				NextStep: "notify",
			},
			"notify": {
				Action:   "send_notification",
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := fmt.Sprintf("test-%d", time.Now().Unix())
	headers := map[string]string{
		"correlation_id": correlationID,
		"request_id":     "test-req-001",
		"client_id":      clientID,
		"user_id":        "test_user",
		"fuel_budget":    "1000",
	}

	err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, nil)
	if err != nil {
		log.Printf("❌ Workflow execution failed: %v", err)
		return
	}

	// Wait for workflow to complete
	time.Sleep(2 * time.Second)

	// Check workflow state
	var status string
	err = db.QueryRow(`
        SELECT status FROM orchestrator_state 
        WHERE correlation_id = $1
    `, correlationID).Scan(&status)

	if err == nil {
		fmt.Printf("✅ Workflow status: %s\n", status)
	}
}

func testGroupSpawning(db *sql.DB, producer kafka.Producer, logger *zap.Logger, clientID string) {
	fmt.Println("\n📌 Test 3: Group Spawning")

	params := actions.ActionParams{
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"group_type": "website-builder",
			},
		},
		Headers: map[string]string{
			"client_id": clientID,
			"user_id":   "test_user",
		},
		DB:       db,
		Producer: producer,
		Logger:   logger,
	}

	result, err := actions.SpawnGroupAction(context.Background(), params)
	if err != nil {
		log.Printf("❌ SpawnGroupAction failed: %v", err)
		return
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("✅ Group spawned successfully:\n%s\n", resultJSON)
}
