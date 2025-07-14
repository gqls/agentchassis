// FILE: platform/infrastructure/connections.go
package infrastructure

import (
	"context"
	"fmt"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/database"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Connections holds all infrastructure connections
type Connections struct {
	ClientsDB     *pgxpool.Pool
	TemplatesDB   *pgxpool.Pool
	KafkaConsumer *kafka.Consumer
	KafkaProducer kafka.Producer
}

// Manager handles infrastructure lifecycle
type Manager struct {
	logger      *zap.Logger
	connections *Connections
}

// NewManager creates a new infrastructure manager
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		logger:      logger,
		connections: &Connections{},
	}
}

// Initialize sets up all infrastructure connections
func (m *Manager) Initialize(ctx context.Context, cfg *config.ServiceConfig, topic, consumerGroup string) error {
	// Initialize database
	clientsPool, err := database.NewPostgresConnection(ctx, cfg.Infrastructure.ClientsDatabase, m.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to clients database: %w", err)
	}
	m.connections.ClientsDB = clientsPool

	// Initialize Templates DB if needed
	if cfg.Infrastructure.TemplatesDatabase.Host != "" {
		templatesPool, err := database.NewPostgresConnection(ctx, cfg.Infrastructure.TemplatesDatabase, m.logger)
		if err != nil {
			m.Close()
			return fmt.Errorf("failed to connect to templates database: %w", err)
		}
		m.connections.TemplatesDB = templatesPool
	}

	// Initialize Kafka
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, topic, consumerGroup, m.logger)
	if err != nil {
		m.Close()
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	m.connections.KafkaConsumer = consumer

	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, m.logger)
	if err != nil {
		m.Close()
		return fmt.Errorf("failed to create producer: %w", err)
	}
	m.connections.KafkaProducer = producer

	return nil
}

// GetConnections returns the infrastructure connections
func (m *Manager) GetConnections() *Connections {
	return m.connections
}

// Close gracefully closes all connections
func (m *Manager) Close() error {
	var errs []error

	if m.connections.KafkaConsumer != nil {
		if err := m.connections.KafkaConsumer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close kafka consumer: %w", err))
		}
	}

	if m.connections.KafkaProducer != nil {
		if err := m.connections.KafkaProducer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close kafka producer: %w", err))
		}
	}

	if m.connections.ClientsDB != nil {
		m.connections.ClientsDB.Close()
	}

	if m.connections.TemplatesDB != nil {
		m.connections.TemplatesDB.Close()
	}

	if len(errs) > 0 {
		return fmt.Errorf("infrastructure shutdown errors: %v", errs)
	}
	return nil
}
