// FILE: platform/health/server.go
package health

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// CheckFunc is a health check function
type CheckFunc func(ctx context.Context) error

// Checkers is a map of named health checks
type Checkers map[string]CheckFunc

// Config for health server
type Config struct {
	HealthPort  string
	MetricsPort string
}

// Server handles health and metrics endpoints
type Server struct {
	serviceName string
	config      Config
	checkers    Checkers
	logger      *zap.Logger
}

// NewServer creates a new health server
func NewServer(serviceName string, config Config, checkers Checkers, logger *zap.Logger) *Server {
	return &Server{
		serviceName: serviceName,
		config:      config,
		checkers:    checkers,
		logger:      logger,
	}
}

// Start starts the health and metrics servers
func (s *Server) Start() {
	go s.startMetricsServer()
	go s.startHealthServer()
}

func (s *Server) startMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	s.logger.Info("Starting metrics server", zap.String("port", s.config.MetricsPort))
	if err := http.ListenAndServe(":"+s.config.MetricsPort, mux); err != nil {
		s.logger.Error("Metrics server failed", zap.Error(err))
	}
}

func (s *Server) startHealthServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)

	s.logger.Info("Starting health server", zap.String("port", s.config.HealthPort))
	if err := http.ListenAndServe(":"+s.config.HealthPort, mux); err != nil {
		s.logger.Error("Health server failed", zap.Error(err))
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	checks := make(map[string]interface{})
	healthy := true

	for name, checker := range s.checkers {
		if err := checker(ctx); err != nil {
			checks[name] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			healthy = false
		} else {
			checks[name] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	response := map[string]interface{}{
		"service": s.serviceName,
		"status":  "healthy",
		"checks":  checks,
	}

	if !healthy {
		response["status"] = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	for _, checker := range s.checkers {
		if err := checker(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("NOT READY"))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}
