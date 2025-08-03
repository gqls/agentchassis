// FILE: platform/health/monitoring.go
package health

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration"
)

// AddMonitoringEndpoints adds workflow monitoring endpoints to the health server
func (s *Server) AddMonitoringEndpoints(monitor *orchestration.WorkflowMonitor) {
	// Active workflows endpoint
	s.AddHandler("/monitor/workflows", func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			http.Error(w, `{"error": "client_id parameter required"}`, http.StatusBadRequest)
			return
		}

		workflows, err := monitor.GetActiveWorkflows(r.Context(), clientID)
		if err != nil {
			s.logger.Error("Failed to get active workflows",
				"error", err,
				"client_id", clientID)
			http.Error(w, `{"error": "failed to retrieve workflows"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"client_id": clientID,
			"workflows": workflows,
			"timestamp": time.Now().UTC(),
		})
	}, "GET")

	// Stuck workflows endpoint
	s.AddHandler("/monitor/stuck", func(w http.ResponseWriter, r *http.Request) {
		// Parse hours parameter (default to 1)
		hoursStr := r.URL.Query().Get("hours")
		hours := 1
		if hoursStr != "" {
			if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
				hours = h
			}
		}

		stuck, err := monitor.GetStuckWorkflows(r.Context(), time.Duration(hours)*time.Hour)
		if err != nil {
			s.logger.Error("Failed to get stuck workflows", "error", err)
			http.Error(w, `{"error": "failed to retrieve stuck workflows"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"stuck_after_hours": hours,
			"workflows":         stuck,
			"timestamp":         time.Now().UTC(),
		})
	}, "GET")

	// Metrics endpoint
	s.AddHandler("/monitor/metrics", func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			http.Error(w, `{"error": "client_id parameter required"}`, http.StatusBadRequest)
			return
		}

		// Parse duration parameter (default to 24h)
		durationStr := r.URL.Query().Get("duration")
		duration := 24 * time.Hour
		if durationStr != "" {
			if d, err := time.ParseDuration(durationStr); err == nil {
				duration = d
			}
		}

		metrics, err := monitor.GetWorkflowMetrics(r.Context(), clientID, time.Now().Add(-duration))
		if err != nil {
			s.logger.Error("Failed to get workflow metrics",
				"error", err,
				"client_id", clientID)
			http.Error(w, `{"error": "failed to retrieve metrics"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"client_id": clientID,
			"duration":  duration.String(),
			"metrics":   metrics,
			"timestamp": time.Now().UTC(),
		})
	}, "GET")

	// Workflow details endpoint
	s.AddHandler("/monitor/workflow/{correlation_id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		correlationID := vars["correlation_id"]

		details, err := monitor.GetWorkflowDetails(r.Context(), correlationID)
		if err != nil {
			if err.Error() == "state not found" {
				http.Error(w, `{"error": "workflow not found"}`, http.StatusNotFound)
				return
			}
			s.logger.Error("Failed to get workflow details",
				"error", err,
				"correlation_id", correlationID)
			http.Error(w, `{"error": "failed to retrieve workflow details"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(details)
	}, "GET")

	s.logger.Info("Monitoring endpoints added to health server")
}
