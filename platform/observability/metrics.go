// FILE: platform/observability/metrics.go
package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Workflow metrics
	WorkflowsStarted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_workflows_started_total",
		Help: "Total number of workflows started",
	}, []string{"agent_type", "workflow_type", "client_id"})

	WorkflowsCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_workflows_completed_total",
		Help: "Total number of workflows completed",
	}, []string{"agent_type", "workflow_type", "status", "client_id"})

	WorkflowDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ai_persona_workflow_duration_seconds",
		Help:    "Duration of workflow execution",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1s to ~100s
	}, []string{"agent_type", "workflow_type", "status"})

	// Agent metrics
	AgentTasksReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_agent_tasks_received_total",
		Help: "Total number of tasks received by agents",
	}, []string{"agent_type", "action"})

	AgentTasksProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_agent_tasks_processed_total",
		Help: "Total number of tasks processed by agents",
	}, []string{"agent_type", "action", "status"})

	AgentProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ai_persona_agent_processing_duration_seconds",
		Help:    "Duration of agent task processing",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
	}, []string{"agent_type", "action"})

	// Fuel metrics
	FuelConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_fuel_consumed_total",
		Help: "Total fuel consumed by operations",
	}, []string{"agent_type", "action", "client_id"})

	FuelExhausted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_fuel_exhausted_total",
		Help: "Total number of operations that failed due to insufficient fuel",
	}, []string{"agent_type", "action", "client_id"})

	// Kafka metrics
	KafkaMessagesProduced = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_kafka_messages_produced_total",
		Help: "Total number of messages produced to Kafka",
	}, []string{"topic"})

	KafkaMessagesConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_kafka_messages_consumed_total",
		Help: "Total number of messages consumed from Kafka",
	}, []string{"topic", "consumer_group"})

	KafkaConsumerLag = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ai_persona_kafka_consumer_lag",
		Help: "Current consumer lag in messages",
	}, []string{"topic", "consumer_group", "partition"})

	// bugs_open/040-kafka-dial: until these existed the only way to measure the
	// intermittent broker dial timeouts was grepping pod logs — and the pods that
	// flake are ephemeral spawned Jobs whose logs GC before anyone looks. Every
	// Kafka dial in the fleet now goes through platform/kafka.SharedDialer and
	// lands here. "outcome" separates a DNS failure from a TCP one, which is the
	// open question the case turns on.
	KafkaDialTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_kafka_dial_total",
		Help: "Kafka broker dial attempts by outcome (ok/timeout/refused/dns/dns_timeout/error)",
	}, []string{"broker", "outcome"})

	KafkaDialDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ai_persona_kafka_dial_duration_seconds",
		Help:    "Kafka broker dial latency, including DNS resolution",
		Buckets: prometheus.ExponentialBuckets(0.005, 2, 12), // 5ms -> ~20s
	}, []string{"outcome"})

	// Database metrics
	DatabaseQueries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_database_queries_total",
		Help: "Total number of database queries",
	}, []string{"database", "operation", "table"})

	DatabaseQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ai_persona_database_query_duration_seconds",
		Help:    "Duration of database queries",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
	}, []string{"database", "operation"})

	// AI Service metrics
	AIServiceRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_ai_service_requests_total",
		Help: "Total number of AI service requests",
	}, []string{"provider", "model", "operation"})

	AIServiceTokensUsed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_ai_service_tokens_used_total",
		Help: "Total number of tokens used by AI services",
	}, []string{"provider", "model", "type"}) // type: input/output

	AIServiceLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ai_persona_ai_service_latency_seconds",
		Help:    "Latency of AI service requests",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 100ms to ~100s
	}, []string{"provider", "model", "operation"})

	// Memory/Vector DB metrics
	VectorSearchQueries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_vector_search_queries_total",
		Help: "Total number of vector similarity searches",
	}, []string{"agent_type"})

	VectorSearchLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ai_persona_vector_search_latency_seconds",
		Help:    "Latency of vector similarity searches",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 8), // 10ms to ~2.5s
	}, []string{"agent_type"})

	MemoriesStored = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_memories_stored_total",
		Help: "Total number of memories stored",
	}, []string{"agent_type", "memory_type"})

	// Circuit breaker metrics
	CircuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ai_persona_circuit_breaker_state",
		Help: "Current state of circuit breakers (0=closed, 1=open, 2=half-open)",
	}, []string{"service"})

	CircuitBreakerTrips = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_circuit_breaker_trips_total",
		Help: "Total number of circuit breaker trips",
	}, []string{"service"})

	// System health metrics
	ActiveWorkflows = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ai_persona_active_workflows",
		Help: "Number of currently active workflows",
	}, []string{"agent_type"})

	AgentPoolSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ai_persona_agent_pool_size",
		Help: "Current size of agent pools",
	}, []string{"agent_type"})

	SystemErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_system_errors_total",
		Help: "Total number of system errors",
	}, []string{"service", "error_code"})

	// Message processing metrics (used by agent.go)
	MessagesProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_messages_processed_total",
		Help: "Total number of messages processed",
	}, []string{"agent_type", "message_type"})

	MessagesFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_messages_failed_total",
		Help: "Total number of messages that failed processing",
	}, []string{"agent_type", "message_type"})

	MessagesDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_messages_dropped_total",
		Help: "Total number of messages dropped",
	}, []string{"agent_type", "reason"})

	MessageProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ai_persona_message_processing_duration_seconds",
		Help:    "Duration of message processing",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
	}, []string{"agent_type", "message_type"})

	// Agent health metric
	AgentHealth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ai_persona_agent_health",
		Help: "Health status of agents (1=healthy, 0=unhealthy)",
	}, []string{"agent_type", "pod_name"})

	// Response metrics (add this new metric)
	ResponsesCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ai_persona_responses_completed_total",
		Help: "Total number of responses completed",
	}, []string{"agent_type", "status"})
)

// MetricsServer provides an HTTP server for Prometheus metrics
type MetricsServer struct {
	port string
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port string) *MetricsServer {
	return &MetricsServer{port: port}
}

// Start starts the metrics HTTP server
func (m *MetricsServer) Start() error {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return http.ListenAndServe(":"+m.port, nil)
}

// WorkflowTimer helps track workflow execution time
type WorkflowTimer struct {
	agentType    string
	workflowType string
	startTime    time.Time
	timer        *prometheus.Timer
}

// StartWorkflowTimer starts timing a workflow execution
func StartWorkflowTimer(agentType, workflowType string) *WorkflowTimer {
	return &WorkflowTimer{
		agentType:    agentType,
		workflowType: workflowType,
		timer: prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			WorkflowDuration.WithLabelValues(agentType, workflowType, "unknown").Observe(v)
		})),
	}
}

// Complete marks the workflow as completed with the given status
func (wt *WorkflowTimer) Complete(status string) {
	// The ObserveDuration method records the observation on the histogram provided to NewTimer.
	// It's designed to be called once. By calling it here, we override the "unknown" status.
	duration := wt.timer.ObserveDuration()
	WorkflowDuration.WithLabelValues(wt.agentType, wt.workflowType, status).Observe(duration.Seconds())
}

// CircuitBreakerStateValue converts circuit breaker state to numeric value
func CircuitBreakerStateValue(state string) float64 {
	switch state {
	case "closed":
		return 0
	case "open":
		return 1
	case "half-open":
		return 2
	default:
		return -1
	}
}
