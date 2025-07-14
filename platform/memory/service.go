// FILE: platform/memory/service.go
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Service handles memory operations for agents
type Service struct {
	pool       *pgxpool.Pool
	aiClient   aiservice.AIService
	logger     *zap.Logger
	memoryRepo *database.MemoryRepository
}

// NewService creates a new memory service
func NewService(pool *pgxpool.Pool, aiClient aiservice.AIService, logger *zap.Logger) *Service {
	return &Service{
		pool:       pool,
		aiClient:   aiClient,
		logger:     logger,
		memoryRepo: database.NewMemoryRepository(pool, logger),
	}
}

// StoreMemory stores a memory entry for an agent
func (s *Service) StoreMemory(ctx context.Context, agentID uuid.UUID, entry models.MemoryEntry) error {
	// Generate embedding
	embedding, err := s.aiClient.GenerateEmbedding(ctx, entry.Content)
	if err != nil {
		s.logger.Error("Failed to generate embedding", zap.Error(err))
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Add timestamp to metadata
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]interface{})
	}
	entry.Metadata["timestamp"] = entry.Timestamp
	entry.Metadata["type"] = entry.Type

	// Store in database
	return s.memoryRepo.StoreMemory(ctx, agentID, entry.Content, embedding, entry.Metadata)
}

// RetrieveRelevantMemories retrieves memories relevant to a query
func (s *Service) RetrieveRelevantMemories(ctx context.Context, agentID uuid.UUID, query string, limit int) ([]models.MemoryEntry, error) {
	// Generate query embedding
	queryEmbedding, err := s.aiClient.GenerateEmbedding(ctx, query)
	if err != nil {
		s.logger.Error("Failed to generate query embedding", zap.Error(err))
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search for similar memories
	records, err := s.memoryRepo.SearchMemory(ctx, agentID, queryEmbedding, limit)
	if err != nil {
		return nil, err
	}

	// Convert to memory entries
	entries := make([]models.MemoryEntry, len(records))
	for i, record := range records {
		entries[i] = models.MemoryEntry{
			Content:  record.Content,
			Metadata: record.Metadata,
		}

		// Extract type and timestamp from metadata
		if typeStr, ok := record.Metadata["type"].(string); ok {
			entries[i].Type = typeStr
		}
		if timestampStr, ok := record.Metadata["timestamp"].(string); ok {
			if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				entries[i].Timestamp = t
			}
		}
	}

	return entries, nil
}

// ProcessWorkflowMemory handles memory storage for workflow steps
func (s *Service) ProcessWorkflowMemory(ctx context.Context, agentID uuid.UUID, config models.MemoryConfiguration, step models.Step, input, output interface{}) error {
	// Check if memory is enabled and this step should store memory
	if !config.Enabled || !step.StoreMemory {
		return nil
	}

	// Create memory entry
	entry := models.MemoryEntry{
		Type:      "workflow_step",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"step_action":      step.Action,
			"step_description": step.Description,
		},
	}

	// Format content based on input and output
	content := map[string]interface{}{
		"action": step.Action,
		"input":  input,
		"output": output,
	}

	contentBytes, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory content: %w", err)
	}
	entry.Content = string(contentBytes)

	// Store the memory
	return s.StoreMemory(ctx, agentID, entry)
}

// GetMemoryContext retrieves relevant memories for a given context
func (s *Service) GetMemoryContext(ctx context.Context, agentID uuid.UUID, config models.MemoryConfiguration, currentContext string) ([]models.MemoryEntry, error) {
	if !config.Enabled {
		return nil, nil
	}

	count := config.RetrievalCount
	if count == 0 {
		count = 5 // default
	}

	return s.RetrieveRelevantMemories(ctx, agentID, currentContext, count)
}
