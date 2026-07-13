// test/integration/database/concurrent_access_test.go
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gqls/agentchassis/test/unit/helpers"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentStateUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	correlationID := helpers.TestUUIDWithType("integration")

	// Initialize state with all required fields including current_step
	_, err := db.Exec(`
        INSERT INTO orchestrator_state 
        (correlation_id, client_id, status, current_step, workflow_plan, execution_metadata)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, correlationID, "test_client", "RUNNING", "processing", // Added current_step
		json.RawMessage(`{"start_step": "init", "steps": {"init": {"action": "init"}, "processing": {"action": "process"}}}`),
		json.RawMessage(`{"counter": 0}`))
	require.NoError(t, err)

	// Concurrent updates
	concurrency := 20
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	errors := make([]error, 0)
	var errorsMutex sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()

			// Increment counter atomically
			_, err := db.Exec(`
                UPDATE orchestrator_state 
                SET execution_metadata = jsonb_set(
                    execution_metadata,
                    '{counter}',
                    to_jsonb((execution_metadata->>'counter')::int + 1)
                ),
                updated_at = NOW()
                WHERE correlation_id = $1
            `, correlationID)

			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, err)
				errorsMutex.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		t.Errorf("Update error: %v", err)
	}

	// Verify final count
	var metadata json.RawMessage
	err = db.QueryRow(`
        SELECT execution_metadata FROM orchestrator_state WHERE correlation_id = $1
    `, correlationID).Scan(&metadata)
	require.NoError(t, err)

	var meta map[string]interface{}
	json.Unmarshal(metadata, &meta)
	assert.Equal(t, float64(concurrency), meta["counter"])
}

func TestDeadlockPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	// Create two states that will be updated in different orders
	id1 := helpers.TestUUIDWithType("integration")
	id2 := helpers.TestUUIDWithType("integration")

	for _, id := range []string{id1, id2} {
		_, err := db.Exec(`
            INSERT INTO orchestrator_state 
            (correlation_id, client_id, status, current_step, workflow_plan)
            VALUES ($1, $2, $3, $4, $5)
        `, id, "test_client", "RUNNING", "init", // Added current_step
			json.RawMessage(`{"start_step": "init", "steps": {"init": {"action": "init"}}}`))
		require.NoError(t, err)
	}

	// Concurrent transactions updating in different orders
	errors := make(chan error, 2)

	go func() {
		tx, err := db.Begin()
		if err != nil {
			errors <- err
			return
		}
		defer tx.Rollback()

		// Update id1 then id2
		_, err = tx.Exec(`
			UPDATE orchestrator_state 
			SET status = 'PROCESSING', updated_at = NOW() 
			WHERE correlation_id = $1`, id1)
		if err != nil {
			errors <- err
			return
		}

		time.Sleep(100 * time.Millisecond) // Increase chance of deadlock

		_, err = tx.Exec(`
			UPDATE orchestrator_state 
			SET status = 'PROCESSING', updated_at = NOW() 
			WHERE correlation_id = $1`, id2)
		if err != nil {
			errors <- err
			return
		}

		errors <- tx.Commit()
	}()

	go func() {
		tx, err := db.Begin()
		if err != nil {
			errors <- err
			return
		}
		defer tx.Rollback()

		// Update id2 then id1 (opposite order)
		_, err = tx.Exec(`
			UPDATE orchestrator_state 
			SET status = 'VALIDATING', updated_at = NOW() 
			WHERE correlation_id = $1`, id2)
		if err != nil {
			errors <- err
			return
		}

		time.Sleep(100 * time.Millisecond) // Increase chance of deadlock

		_, err = tx.Exec(`
			UPDATE orchestrator_state 
			SET status = 'VALIDATING', updated_at = NOW() 
			WHERE correlation_id = $1`, id1)
		if err != nil {
			errors <- err
			return
		}

		errors <- tx.Commit()
	}()

	// Collect results
	var succeeded int
	var deadlockDetected bool

	for i := 0; i < 2; i++ {
		select {
		case err := <-errors:
			if err == nil {
				succeeded++
			} else {
				// PostgreSQL should handle deadlock detection
				t.Logf("Transaction error: %v", err)
				if err.Error() == "pq: deadlock detected" ||
					err.Error() == "deadlock detected" {
					deadlockDetected = true
				}
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent updates")
		}
	}

	// At least one should succeed
	assert.GreaterOrEqual(t, succeeded, 1)

	// Log result
	if deadlockDetected {
		t.Log("Deadlock was properly detected and handled by PostgreSQL")
	} else {
		t.Log("Both transactions completed without deadlock")
	}
}

func TestConnectionPooling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// First try to use the standard test DB connection to get the right config
	testDB := helpers.TestDB(t)
	if testDB == nil {
		t.Skip("Database not available")
		return
	}

	// Close the test DB as we'll create our own pool
	testDB.Close()

	// Get connection parameters from environment or use defaults
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		// In Kubernetes, try the service name
		if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
			dbHost = "postgres-clients.ai-persona-system.svc.cluster.local"
		} else {
			dbHost = "localhost"
		}
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "clients_user"
	}

	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "password"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "clients_db"
	}

	// Test connection pool behavior
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbName)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("Could not open database: %v", err)
		return
	}
	defer db.Close()

	// Configure pool BEFORE testing connection
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Skipf("Database not reachable (host: %s): %v", dbHost, err)
		return
	}

	// Verify pool stats
	stats := db.Stats()
	assert.LessOrEqual(t, stats.OpenConnections, 10)
	t.Logf("Initial pool stats - Open: %d, InUse: %d, Idle: %d",
		stats.OpenConnections, stats.InUse, stats.Idle)

	// Execute concurrent queries
	concurrency := 50
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	successCount := int32(0)
	var errorsMutex sync.Mutex
	queryErrors := make([]error, 0)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()

			var count int
			err := db.QueryRow(`SELECT COUNT(*) FROM orchestrator_state`).Scan(&count)
			if err != nil {
				errorsMutex.Lock()
				queryErrors = append(queryErrors, err)
				errorsMutex.Unlock()
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// Report results
	t.Logf("Queries - Success: %d, Failed: %d", successCount, len(queryErrors))

	// Report first few errors if any
	if len(queryErrors) > 0 {
		maxErrors := 3
		if len(queryErrors) < maxErrors {
			maxErrors = len(queryErrors)
		}
		for i := 0; i < maxErrors; i++ {
			t.Logf("Query error %d: %v", i+1, queryErrors[i])
		}
	}

	// Check final stats
	finalStats := db.Stats()
	assert.LessOrEqual(t, finalStats.OpenConnections, 10)

	t.Logf("Final pool stats - Open: %d, InUse: %d, Idle: %d, MaxOpen: %d",
		finalStats.OpenConnections, finalStats.InUse, finalStats.Idle, finalStats.MaxOpenConnections)

	// At least some queries should succeed
	assert.Greater(t, successCount, int32(0), "At least some queries should succeed")
}

func TestOptimisticLocking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	correlationID := helpers.TestUUIDWithType("integration")

	// Initialize state with version
	_, err := db.Exec(`
        INSERT INTO orchestrator_state 
        (correlation_id, client_id, status, current_step, workflow_plan, 
         execution_metadata, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `, correlationID, "test_client", "RUNNING", "init",
		json.RawMessage(`{"start_step": "init", "steps": {"init": {"action": "init"}}}`),
		json.RawMessage(`{"version": 1}`))
	require.NoError(t, err)

	// Get initial version/timestamp
	var initialUpdatedAt time.Time
	err = db.QueryRow(`
		SELECT updated_at FROM orchestrator_state WHERE correlation_id = $1
	`, correlationID).Scan(&initialUpdatedAt)
	require.NoError(t, err)

	// Simulate two concurrent updates with optimistic locking
	wg := sync.WaitGroup{}
	wg.Add(2)
	successCount := 0
	var successMutex sync.Mutex

	for i := 0; i < 2; i++ {
		go func(index int) {
			defer wg.Done()

			// Try to update with optimistic lock check
			result, err := db.Exec(`
				UPDATE orchestrator_state 
				SET status = $1,
					execution_metadata = jsonb_set(execution_metadata, '{version}', to_jsonb($2::int)),
					updated_at = NOW()
				WHERE correlation_id = $3 
				AND updated_at = $4
			`, fmt.Sprintf("UPDATED_%d", index), index+2, correlationID, initialUpdatedAt)

			if err == nil {
				rowsAffected, _ := result.RowsAffected()
				if rowsAffected > 0 {
					successMutex.Lock()
					successCount++
					successMutex.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	// Only one should succeed due to optimistic locking
	assert.Equal(t, 1, successCount, "Only one update should succeed with optimistic locking")

	// Verify final state
	var status string
	var metadata json.RawMessage
	err = db.QueryRow(`
		SELECT status, execution_metadata FROM orchestrator_state WHERE correlation_id = $1
	`, correlationID).Scan(&status, &metadata)
	require.NoError(t, err)

	var meta map[string]interface{}
	json.Unmarshal(metadata, &meta)

	t.Logf("Final status: %s, version: %v", status, meta["version"])
	assert.Contains(t, []string{"UPDATED_0", "UPDATED_1"}, status)
	assert.Contains(t, []float64{2, 3}, meta["version"])
}
