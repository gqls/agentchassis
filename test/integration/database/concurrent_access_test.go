// test/integration/database/concurrent_access_test.go
package database

import (
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentStateUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	correlationID := "test-concurrent-" + uuid.New().String()

	// Initialize state
	_, err := db.Exec(`
        INSERT INTO orchestrator_state 
        (correlation_id, client_id, status, workflow_plan, execution_metadata)
        VALUES ($1, $2, $3, $4, $5)
    `, correlationID, "test_client", "RUNNING",
		json.RawMessage(`{"start_step": "init"}`),
		json.RawMessage(`{"counter": 0}`))
	require.NoError(t, err)

	// Concurrent updates
	concurrency := 20
	wg := sync.WaitGroup{}
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			// Increment counter atomically
			_, err := db.Exec(`
                UPDATE orchestrator_state 
                SET execution_metadata = jsonb_set(
                    execution_metadata,
                    '{counter}',
                    to_jsonb((execution_metadata->>'counter')::int + 1)
                )
                WHERE correlation_id = $1
            `, correlationID)

			if err != nil {
				t.Errorf("Update error: %v", err)
			}
		}()
	}

	wg.Wait()

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
	id1 := "test-deadlock-1"
	id2 := "test-deadlock-2"

	for _, id := range []string{id1, id2} {
		_, err := db.Exec(`
            INSERT INTO orchestrator_state 
            (correlation_id, client_id, status, workflow_plan)
            VALUES ($1, $2, $3, $4)
        `, id, "test_client", "RUNNING", json.RawMessage(`{}`))
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
		_, err = tx.Exec(`UPDATE orchestrator_state SET status = 'PROCESSING' WHERE correlation_id = $1`, id1)
		if err != nil {
			errors <- err
			return
		}

		time.Sleep(100 * time.Millisecond) // Increase chance of deadlock

		_, err = tx.Exec(`UPDATE orchestrator_state SET status = 'PROCESSING' WHERE correlation_id = $1`, id2)
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
		_, err = tx.Exec(`UPDATE orchestrator_state SET status = 'VALIDATING' WHERE correlation_id = $1`, id2)
		if err != nil {
			errors <- err
			return
		}

		time.Sleep(100 * time.Millisecond) // Increase chance of deadlock

		_, err = tx.Exec(`UPDATE orchestrator_state SET status = 'VALIDATING' WHERE correlation_id = $1`, id1)
		if err != nil {
			errors <- err
			return
		}

		errors <- tx.Commit()
	}()

	// Collect results
	var succeeded int
	for i := 0; i < 2; i++ {
		select {
		case err := <-errors:
			if err == nil {
				succeeded++
			} else {
				// PostgreSQL should handle deadlock detection
				assert.Contains(t, err.Error(), "deadlock")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent updates")
		}
	}

	// At least one should succeed
	assert.GreaterOrEqual(t, succeeded, 1)
}

func TestConnectionPooling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Test connection pool behavior
	dbURL := "postgres://clients_user:password@localhost:5432/clients_db?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Configure pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify pool stats
	stats := db.Stats()
	assert.LessOrEqual(t, stats.OpenConnections, 10)

	// Execute concurrent queries
	concurrency := 50
	wg := sync.WaitGroup{}
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer wg.Done()

			var count int
			err := db.QueryRow(`SELECT COUNT(*) FROM orchestrator_state`).Scan(&count)
			if err != nil {
				t.Errorf("Query error: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Check final stats
	finalStats := db.Stats()
	assert.LessOrEqual(t, finalStats.OpenConnections, 10)
	assert.GreaterOrEqual(t, finalStats.InUse+finalStats.Idle, 1)
}
