// cmd/workflow-monitor/main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration"
)

func main() {
	var (
		dbURL      = flag.String("db", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
		clientID   = flag.String("client", os.Getenv("CLIENT_ID"), "Client ID to monitor")
		stuckHours = flag.Int("stuck-hours", 1, "Hours to consider a workflow stuck")
		//outputFmt  = flag.String("format", "text", "Output format: text, json")
	)
	flag.Parse()

	// Allow environment variables to provide defaults
	if *dbURL == "" {
		*dbURL = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("CLIENTS_DB_USER"),
			os.Getenv("CLIENTS_DB_PASSWORD"),
			os.Getenv("CLIENTS_DB_HOST"),
			os.Getenv("CLIENTS_DB_PORT"),
			os.Getenv("CLIENTS_DB_NAME"))
	}

	if *clientID == "" {
		log.Fatal("Client ID is required (use -client flag or CLIENT_ID env var)")
	}

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	monitor := orchestration.NewWorkflowMonitor(db)
	ctx := context.Background()

	// Get active workflows
	fmt.Println("\n=== Active Workflows ===")
	active, err := monitor.GetActiveWorkflows(ctx, *clientID)
	if err != nil {
		log.Fatal("Failed to get active workflows:", err)
	}

	for _, w := range active {
		fmt.Printf("ID: %s | Step: %s | Progress: %.0f%% | Duration: %s\n",
			w.CorrelationID[:8], w.CurrentStep, w.Progress, w.Duration)
	}

	// Get stuck workflows
	fmt.Printf("\n=== Workflows Stuck > %d hours ===\n", *stuckHours)
	stuck, err := monitor.GetStuckWorkflows(ctx, time.Duration(*stuckHours)*time.Hour)
	if err != nil {
		log.Fatal("Failed to get stuck workflows:", err)
	}

	for _, w := range stuck {
		fmt.Printf("ID: %s | Client: %s | Stuck at: %s | For: %s\n",
			w.CorrelationID[:8], w.ClientID, w.CurrentStep, w.StuckDuration)
	}

	// Get metrics
	fmt.Println("\n=== 24-Hour Metrics ===")
	metrics, err := monitor.GetWorkflowMetrics(ctx, *clientID, time.Now().Add(-24*time.Hour))
	if err != nil {
		log.Fatal("Failed to get metrics:", err)
	}

	fmt.Printf("Total: %d | Success Rate: %.1f%% | Active: %d | Failed: %d\n",
		metrics.TotalWorkflows, metrics.SuccessRate, metrics.ActiveWorkflows, metrics.FailedWorkflows)
}
