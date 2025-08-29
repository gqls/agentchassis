package messaging

import (
	"context"
	"time"

	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

type LoggingMiddleware struct {
	logger *zap.Logger
	next   MessageHandler
}

type MessageHandler interface {
	ProcessMessage(ctx context.Context, msg kafka.Message) error
}

func NewLoggingMiddleware(logger *zap.Logger, next MessageHandler) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
		next:   next,
	}
}

func (m *LoggingMiddleware) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	startTime := time.Now()
	headers := kafka.HeadersToMap(msg.Headers)

	// Create ExecutionContext for structured logging
	execCtx, err := types.FromHeaders(headers)
	if err != nil {
		// Log with raw headers if context creation fails
		m.logger.Error("Failed to create ExecutionContext",
			zap.Error(err),
			zap.Any("headers", headers))
		return m.next.ProcessMessage(ctx, msg)
	}

	// Create logger with context fields
	contextLogger := m.logger.With(execCtx.LogContext()...)

	// Log message receipt
	contextLogger.Info("Message received",
		zap.String("topic", msg.Topic),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
		zap.Int("payload_size", len(msg.Value)))

	// Process the message
	err = m.next.ProcessMessage(ctx, msg)

	// Log completion
	duration := time.Since(startTime)
	if err != nil {
		contextLogger.Error("Message processing failed",
			zap.Error(err),
			zap.Duration("duration", duration))
	} else {
		contextLogger.Info("Message processed successfully",
			zap.Duration("duration", duration))
	}

	return err
}
