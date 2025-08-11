// test/tools/db-inspector/check_state.go
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	_ "github.com/lib/pq"
)

var (
	dbURL         = flag.String("db", "", "Database URL")
	correlationID = flag.String("correlation", "", "Correlation ID to inspect")
	clientID      = flag.String("client", "demo_client", "Client ID")
	listActive    = flag.Bool("active", false, "List active workflows")
	listStuck     = flag.Bool("stuck", false, "List stuck workflows")
	watch         = flag.Bool("watch", false, "Watch workflow progress")
)

func main() {
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
		if *dbURL == "" {
			log.Fatal("Database URL required (-db or DATABASE_URL env)")
		}
	}

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer db.Close()

	switch {
	case *correlationID != "":
		if *watch {
			watchWorkflow(db, *correlationID)
		} else {
			inspectWorkflow(db, *correlationID)
		}
	case *listActive:
		listActiveWorkflows(db, *clientID)
	case *listStuck:
		listStuckWorkflows(db, *clientID)
	default:
		flag.Usage()
	}
}

func inspectWorkflow(db *sql.DB, correlationID string) {
	var (
		clientID      string
		status        string
		currentStep   string
		workflowPlan  json.RawMessage
		collectedData json.RawMessage
		metadata      json.RawMessage
		executionPath json.RawMessage
		createdAt     time.Time
		updatedAt     time.Time
	)

	err := db.QueryRow(`
        SELECT client_id, status, current_step, workflow_plan, 
               collected_data, execution_metadata, execution_path,
               created_at, updated_at
        FROM orchestrator_state
        WHERE correlation_id = $1
    `, correlationID).Scan(
		&clientID, &status, &currentStep, &workflowPlan,
		&collectedData, &metadata, &executionPath,
		&createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		fmt.Printf("Workflow %s not found\n", correlationID)
		return
	}
	if err != nil {
		log.Fatal("Query failed:", err)
	}

	fmt.Printf("=== Workflow: %s ===\n", correlationID)
	fmt.Printf("Client ID: %s\n", clientID)
	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Current Step: %s\n", currentStep)
	fmt.Printf("Created: %s\n", createdAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s (%.1f seconds ago)\n",
		updatedAt.Format(time.RFC3339),
		time.Since(updatedAt).Seconds())

	// Pretty print workflow plan
	fmt.Println("\n--- Workflow Plan ---")
	prettyPrint(workflowPlan)

	// Pretty print execution metadata
	fmt.Println("\n--- Execution Metadata ---")
	prettyPrint(metadata)

	// Pretty print execution path
	fmt.Println("\n--- Execution Path ---")
	prettyPrint(executionPath)

	// Pretty print collected data
	if len(collectedData) > 2 {
		fmt.Println("\n--- Collected Data ---")
		prettyPrint(collectedData)
	}
}

func listActiveWorkflows(db *sql.DB, clientID string) {
	rows, err := db.Query(`
        SELECT correlation_id, status, current_step,
               EXTRACT(EPOCH FROM (NOW() - updated_at)) as seconds_since_update,
               execution_metadata->>'completed_steps' as completed,
               execution_metadata->>'total_steps' as total
        FROM orchestrator_state
        WHERE client_id = $1
          AND status IN ('RUNNING', 'AWAITING_RESPONSES', 'PAUSED_FOR_HUMAN')
        ORDER BY updated_at DESC
        LIMIT 20
    `, clientID)
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CORRELATION_ID\tSTATUS\tCURRENT_STEP\tPROGRESS\tLAST_UPDATE")

	for rows.Next() {
		var (
			correlationID string
			status        string
			currentStep   string
			secondsAgo    float64
			completed     sql.NullString
			total         sql.NullString
		)

		err := rows.Scan(&correlationID, &status, &currentStep, &secondsAgo, &completed, &total)
		if err != nil {
			continue
		}

		progress := "N/A"
		if completed.Valid && total.Valid {
			progress = fmt.Sprintf("%s/%s", completed.String, total.String)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.0fs ago\n",
			correlationID, status, currentStep, progress, secondsAgo)
	}
	w.Flush()
}

func watchWorkflow(db *sql.DB, correlationID string) {
	fmt.Printf("Watching workflow %s (Ctrl+C to stop)...\n\n", correlationID)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastStatus string

	for {
		select {
		case <-ticker.C:
			var status, currentStep string
			var metadata json.RawMessage

			err := db.QueryRow(`
                SELECT status, current_step, execution_metadata
                FROM orchestrator_state
                WHERE correlation_id = $1
            `, correlationID).Scan(&status, &currentStep, &metadata)

			if err != nil {
				continue
			}

			if status != lastStatus {
				fmt.Printf("[%s] Status: %s, Step: %s\n",
					time.Now().Format("15:04:05"),
					status, currentStep)

				// Parse and show progress
				var meta map[string]interface{}
				if json.Unmarshal(metadata, &meta) == nil {
					if completed, ok := meta["completed_steps"].(float64); ok {
						if total, ok := meta["total_steps"].(float64); ok {
							fmt.Printf("         Progress: %.0f/%.0f steps (%.1f%%)\n",
								completed, total, (completed/total)*100)
						}
					}
				}

				lastStatus = status

				if status == "COMPLETED" || status == "FAILED" {
					fmt.Println("\nWorkflow finished!")
					return
				}
			}
		}
	}
}

func prettyPrint(data json.RawMessage) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err == nil {
		pretty, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(pretty))
	}
}

func listStuckWorkflows(db *sql.DB, clientID string) {
	rows, err := db.Query(`
        SELECT correlation_id, status, current_step,
               EXTRACT(EPOCH FROM (NOW() - updated_at)) as seconds_since_update
        FROM orchestrator_state
        WHERE client_id = $1
          AND status IN ('RUNNING', 'AWAITING_RESPONSES')
          AND updated_at < NOW() - INTERVAL '1 hour'
        ORDER BY updated_at ASC
        LIMIT 50
    `, clientID)
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CORRELATION_ID\tSTATUS\tCURRENT_STEP\tLAST_UPDATE")

	for rows.Next() {
		var (
			correlationID string
			status        string
			currentStep   string
			secondsAgo    float64
		)

		err := rows.Scan(&correlationID, &status, &currentStep, &secondsAgo)
		if err != nil {
			continue
		}

		// Format the time since last update
		lastUpdate := fmt.Sprintf("%.0fm ago", secondsAgo/60)
		if secondsAgo > 3600 {
			lastUpdate = fmt.Sprintf("%.1fh ago", secondsAgo/3600)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			correlationID, status, currentStep, lastUpdate)
	}
	w.Flush()
}
