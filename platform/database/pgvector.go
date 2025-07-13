// FILE: platform/database/pgvector.go
package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

// MemoryRecord represents a single entry in the agent_memory table
type MemoryRecord struct {
	ID              uuid.UUID
	AgentInstanceID uuid.UUID
	Content         string
	Embedding       []float32
	Metadata        map[string]interface{}
}

// MemoryRepository provides methods for storing and retrieving agent memories
type MemoryRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewMemoryRepository creates a new repository for memory operations
func NewMemoryRepository(pool *pgxpool.Pool, logger *zap.Logger) *MemoryRepository {
	return &MemoryRepository{pool: pool, logger: logger}
}

// StoreMemory saves a new memory record to the database for a specific agent
func (r *MemoryRepository) StoreMemory(ctx context.Context, agentID uuid.UUID, content string, embedding []float32, metadata map[string]interface{}) error {
	l := r.logger.With(zap.String("agent_id", agentID.String()))
	l.Info("Storing new memory")

	query := `
        INSERT INTO agent_memory (agent_instance_id, content, embedding, metadata)
        VALUES ($1, $2, $3, $4)
    `
	_, err := r.pool.Exec(ctx, query, agentID, content, pgvector.NewVector(embedding), metadata)
	if err != nil {
		l.Error("Failed to store agent memory", zap.Error(err))
		return fmt.Errorf("failed to insert memory record: %w", err)
	}

	l.Debug("Successfully stored memory record")
	return nil
}

// SearchMemory performs a semantic similarity search to find the most relevant memories
func (r *MemoryRepository) SearchMemory(ctx context.Context, agentID uuid.UUID, queryEmbedding []float32, limit int) ([]MemoryRecord, error) {
	l := r.logger.With(zap.String("agent_id", agentID.String()))
	l.Info("Searching for relevant memories", zap.Int("limit", limit))

	query := `
        SELECT id, content, metadata
        FROM agent_memory
        WHERE agent_instance_id = $1
        ORDER BY embedding <=> $2
        LIMIT $3
    `
	rows, err := r.pool.Query(ctx, query, agentID, pgvector.NewVector(queryEmbedding), limit)
	if err != nil {
		l.Error("Failed to execute memory search query", zap.Error(err))
		return nil, fmt.Errorf("failed to search memory: %w", err)
	}
	defer rows.Close()

	var results []MemoryRecord
	for rows.Next() {
		var record MemoryRecord
		record.AgentInstanceID = agentID
		if err := rows.Scan(&record.ID, &record.Content, &record.Metadata); err != nil {
			l.Error("Failed to scan memory search result", zap.Error(err))
			continue
		}
		results = append(results, record)
	}

	l.Info("Memory search completed", zap.Int("results_found", len(results)))
	return results, nil
}
