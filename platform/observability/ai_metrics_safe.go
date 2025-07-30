// FILE: platform/observability/ai_metrics_safe.go
package observability

import (
	"github.com/gqls/agentchassis/platform/aiservice"
	"go.uber.org/zap"
	"sync"
)

// SafeAIMetrics provides safe AI metrics recording
type SafeAIMetrics struct {
	logger *zap.Logger
}

var (
	metricsLogger *zap.Logger
	loggerOnce    sync.Once
)

// NewSafeAIMetrics creates a new safe AI metrics recorder
func NewSafeAIMetrics(logger *zap.Logger) *SafeAIMetrics {
	SetMetricsLogger(logger)
	return &SafeAIMetrics{logger: logger}
}

// RecordAIServiceMetrics safely records AI service metrics
func (s *SafeAIMetrics) RecordAIServiceMetrics(client aiservice.AIService, operation string, tokensUsed int, tokenType string) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Failed to record AI metrics",
				zap.Any("panic", r),
				zap.String("operation", operation),
				zap.Int("tokens", tokensUsed),
				zap.String("type", tokenType),
			)
		}
	}()

	// Safely get provider and model
	provider := s.safeGetProvider(client)
	model := s.safeGetModel(client)

	// Record metrics using safe wrappers
	if tokensUsed > 0 {
		s.safeRecordTokens(provider, model, tokenType, float64(tokensUsed))
	}

	if operation != "" {
		s.safeRecordRequest(provider, model, operation)
	}
}

// safeGetProvider safely gets the provider from client
func (s *SafeAIMetrics) safeGetProvider(client aiservice.AIService) string {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Failed to get provider", zap.Any("panic", r))
		}
	}()

	if client == nil {
		return "unknown"
	}

	provider := client.Provider()
	if provider == "" {
		return "unknown"
	}

	return provider
}

// safeGetModel safely gets the model from client
func (s *SafeAIMetrics) safeGetModel(client aiservice.AIService) string {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Failed to get model", zap.Any("panic", r))
		}
	}()

	if client == nil {
		return "unknown"
	}

	model := client.Model()
	if model == "" {
		return "unknown"
	}

	return model
}

// safeRecordTokens safely records token metrics
func (s *SafeAIMetrics) safeRecordTokens(provider, model, tokenType string, tokens float64) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Failed to record token metrics",
				zap.Any("panic", r),
				zap.String("provider", provider),
				zap.String("model", model),
				zap.String("type", tokenType),
			)
		}
	}()

	// Ensure we have valid labels
	if provider == "" {
		provider = "unknown"
	}
	if model == "" {
		model = "unknown"
	}
	if tokenType == "" {
		tokenType = "unknown"
	}

	// Log what we're about to record for debugging
	s.logger.Debug("Recording token metrics",
		zap.String("provider", provider),
		zap.String("model", model),
		zap.String("type", tokenType),
		zap.Float64("tokens", tokens),
	)

	AIServiceTokensUsed.WithLabelValues(provider, model, tokenType).Add(tokens)
}

// safeRecordRequest safely records request metrics
func (s *SafeAIMetrics) safeRecordRequest(provider, model, operation string) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Failed to record request metrics",
				zap.Any("panic", r),
				zap.String("provider", provider),
				zap.String("model", model),
				zap.String("operation", operation),
			)
		}
	}()

	// Ensure we have valid labels
	if provider == "" {
		provider = "unknown"
	}
	if model == "" {
		model = "unknown"
	}
	if operation == "" {
		operation = "unknown"
	}

	AIServiceRequests.WithLabelValues(provider, model, operation).Inc()
}

// Global safe metrics instance
var (
	safeMetrics     *SafeAIMetrics
	safeMetricsOnce sync.Once
)

// GetSafeAIMetrics returns the global safe metrics instance
func GetSafeAIMetrics(logger *zap.Logger) *SafeAIMetrics {
	safeMetricsOnce.Do(func() {
		safeMetrics = NewSafeAIMetrics(logger)
	})
	return safeMetrics
}

// RecordAIServiceMetricsSafe is a safe global function for recording AI metrics
func RecordAIServiceMetricsSafe(logger *zap.Logger, client aiservice.AIService, operation string, tokensUsed int, tokenType string) {
	GetSafeAIMetrics(logger).RecordAIServiceMetrics(client, operation, tokensUsed, tokenType)
}

// SetMetricsLogger sets the logger for metrics errors
func SetMetricsLogger(logger *zap.Logger) {
	loggerOnce.Do(func() {
		metricsLogger = logger
	})
}

// getLogger returns the metrics logger or a noop logger
func getLogger() *zap.Logger {
	if metricsLogger != nil {
		return metricsLogger
	}
	// Return a noop logger if not set
	return zap.NewNop()
}
