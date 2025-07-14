// FILE: platform/resilience/circuit_breaker.go
package resilience

import (
	"context"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// CircuitBreakerConfig holds configuration for circuit breakers
type CircuitBreakerConfig struct {
	Name                string
	MaxRequests         uint32
	Interval            time.Duration
	Timeout             time.Duration
	ConsecutiveFailures uint32
	FailureRatio        float64
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:                name,
		MaxRequests:         3,
		Interval:            60 * time.Second,
		Timeout:             60 * time.Second,
		ConsecutiveFailures: 5,
		FailureRatio:        0.6,
	}
}

// CircuitBreaker wraps the gobreaker implementation
type CircuitBreaker struct {
	breaker *gobreaker.CircuitBreaker
	logger  *zap.Logger
	config  CircuitBreakerConfig
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig, logger *zap.Logger) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= config.ConsecutiveFailures && failureRatio >= config.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("Circuit breaker state change",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	return &CircuitBreaker{
		breaker: gobreaker.NewCircuitBreaker(settings),
		logger:  logger,
		config:  config,
	}
}

// Execute runs a function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return cb.breaker.Execute(fn)
}

// ExecuteWithContext runs a function with context through the circuit breaker
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	return cb.breaker.Execute(func() (interface{}, error) {
		return fn(ctx)
	})
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() string {
	return cb.breaker.State().String()
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.breaker.State() == gobreaker.StateOpen
}

// Counts returns the current counts
func (cb *CircuitBreaker) Counts() gobreaker.Counts {
	return cb.breaker.Counts()
}

// HTTPClient wraps an HTTP client with circuit breaker functionality
type HTTPClientWithBreaker struct {
	client  HTTPDoer
	Breaker *CircuitBreaker
	logger  *zap.Logger
}

// HTTPDoer interface for HTTP client
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPClientWithBreaker creates a new HTTP client with circuit breaker
func NewHTTPClientWithBreaker(client HTTPDoer, config CircuitBreakerConfig, logger *zap.Logger) *HTTPClientWithBreaker {
	return &HTTPClientWithBreaker{
		client:  client,
		Breaker: NewCircuitBreaker(config, logger),
		logger:  logger,
	}
}

// Do executes an HTTP request through the circuit breaker
func (c *HTTPClientWithBreaker) Do(req *http.Request) (*http.Response, error) {
	result, err := c.Breaker.Execute(func() (interface{}, error) {
		return c.client.Do(req)
	})

	if err != nil {
		return nil, err
	}

	return result.(*http.Response), nil
}

// State returns the current state of the circuit breaker
func (c *HTTPClientWithBreaker) State() string {
	return c.Breaker.State()
}

// Counts returns the current counts
func (c *HTTPClientWithBreaker) Counts() gobreaker.Counts {
	return c.Breaker.Counts()
}

// IsCircuitBreakerError checks if an error is from a circuit breaker
func IsCircuitBreakerError(err error) bool {
	if err == nil {
		return false
	}
	return err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests
}
