// FILE: cmd/dispatcher/main.go
//
// Remote Agent Dispatcher
//
// A lightweight service that runs in each remote cluster.
// It consumes agent dispatch requests from Kafka (system.dispatch.requests)
// and creates local Kubernetes Jobs.
//
// The dispatcher doesn't need access to PostgreSQL — the parent agent in the
// originating cluster already created all DB records. This service only needs:
//   - Kafka access (shared cluster or local federated broker)
//   - Local Kubernetes API access (to create Jobs)
//   - Access to the same container images (docker.io/aqls/agent-chassis)
//
// Configuration is via environment variables:
//   KAFKA_BROKERS          — comma-separated broker list
//   CLUSTER_ID             — identifier for this cluster (e.g. "cluster-b")
//   NAMESPACE              — K8s namespace for Jobs (default: "ai-persona-system")
//   CONSUMER_GROUP         — Kafka consumer group (default: "dispatcher-{CLUSTER_ID}")
//   DISPATCH_TOPIC         — topic to consume (default: "system.dispatch.requests")
//   DISPATCH_RESPONSE_TOPIC — topic for confirmations (default: "system.dispatch.responses")
//
// Infrastructure overrides (if the remote cluster should use different DB/Kafka
// endpoints than what the originating cluster sent in the dispatch message):
//   OVERRIDE_KAFKA_BROKERS       — Kafka brokers the spawned agent should use
//   OVERRIDE_DATABASE_HOST       — Postgres host for spawned agents
//   OVERRIDE_DATABASE_PORT       — Postgres port
//   OVERRIDE_DATABASE_USER       — Postgres user
//   OVERRIDE_DATABASE_NAME       — Postgres DB name
//   OVERRIDE_TEMPLATES_DB_HOST   — Templates DB host
//   OVERRIDE_TEMPLATES_DB_PORT   — Templates DB port
//   OVERRIDE_TEMPLATES_DB_USER   — Templates DB user
//   OVERRIDE_TEMPLATES_DB_NAME   — Templates DB name
//   OVERRIDE_CORE_MANAGER_URL    — Core manager URL

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// DispatchRequest mirrors the struct from dispatch_actions.go.
// Kept as a separate definition so the dispatcher has no import dependency
// on the chassis — it's a standalone binary.
type DispatchRequest struct {
	AgentID              string          `json:"agent_id"`
	AgentType            string          `json:"agent_type"`
	AgentName            string          `json:"agent_name"`
	Role                 string          `json:"role"`
	ClientID             string          `json:"client_id"`
	ImageRepository      string          `json:"image_repository"`
	ImageTag             string          `json:"image_tag"`
	Command              []string        `json:"command"`
	Resources            json.RawMessage `json:"resources"`
	HealthConfig         json.RawMessage `json:"health_config"`
	EnvVars              json.RawMessage `json:"env_vars"`
	Category             string          `json:"category"`
	RequestsTopic        string          `json:"requests_topic"`
	ResponsesTopic       string          `json:"responses_topic"`
	ParentResponsesTopic string          `json:"parent_responses_topic"`
	TargetCluster        string          `json:"target_cluster"`
	KafkaBrokers         string          `json:"kafka_brokers,omitempty"`
	DatabaseHost         string          `json:"database_host,omitempty"`
	DatabasePort         string          `json:"database_port,omitempty"`
	DatabaseUser         string          `json:"database_user,omitempty"`
	DatabaseName         string          `json:"database_name,omitempty"`
	TemplatesDBHost      string          `json:"templates_db_host,omitempty"`
	TemplatesDBPort      string          `json:"templates_db_port,omitempty"`
	TemplatesDBUser      string          `json:"templates_db_user,omitempty"`
	TemplatesDBName      string          `json:"templates_db_name,omitempty"`
	DispatchedAt         string          `json:"dispatched_at"`
}

// ResourceSpec matches the agent_definitions.resources JSON structure
type ResourceSpec struct {
	Requests map[string]string `json:"requests"`
	Limits   map[string]string `json:"limits"`
}

// HealthSpec matches the agent_definitions.health_config JSON structure
type HealthSpec struct {
	Port                int    `json:"port"`
	LivenessPath        string `json:"liveness_path"`
	ReadinessPath       string `json:"readiness_path"`
	InitialDelaySeconds int    `json:"initial_delay_seconds"`
}

// EnvVarSpec for custom env vars from agent_definitions
type EnvVarSpec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// DispatchResponse is sent back to confirm job creation
type DispatchResponse struct {
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	ClusterID string `json:"cluster_id"`
	JobName   string `json:"job_name"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
}

func main() {
	// Setup logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Configuration from environment
	kafkaBrokers := getEnvOrDefault("KAFKA_BROKERS",
		"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092")
	clusterID := getEnvOrDefault("CLUSTER_ID", "cluster-b")
	namespace := getEnvOrDefault("NAMESPACE", "ai-persona-system")
	consumerGroup := getEnvOrDefault("CONSUMER_GROUP", fmt.Sprintf("dispatcher-%s", clusterID))
	dispatchTopic := getEnvOrDefault("DISPATCH_TOPIC", "system.dispatch.requests")
	responseTopic := getEnvOrDefault("DISPATCH_RESPONSE_TOPIC", "system.dispatch.responses")

	logger.Info("Remote Agent Dispatcher starting",
		zap.String("cluster_id", clusterID),
		zap.String("namespace", namespace),
		zap.String("kafka_brokers", kafkaBrokers),
		zap.String("consumer_group", consumerGroup),
		zap.String("dispatch_topic", dispatchTopic))

	// Get K8s client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Fatal("Failed to get in-cluster K8s config", zap.Error(err))
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Fatal("Failed to create K8s client", zap.Error(err))
	}

	// Create Kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        strings.Split(kafkaBrokers, ","),
		Topic:          dispatchTopic,
		GroupID:        consumerGroup,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})
	defer reader.Close()

	// Create Kafka writer for responses
	writer := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(kafkaBrokers, ",")...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}
	defer writer.Close()

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	logger.Info("Dispatcher ready, consuming from dispatch topic")

	// Main consume loop
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Info("Context cancelled, shutting down")
				return
			}
			logger.Error("Failed to read message", zap.Error(err))
			continue
		}

		// Check if this message is for our cluster
		targetCluster := getHeader(msg.Headers, "target_cluster")
		if targetCluster != "" && targetCluster != clusterID && targetCluster != "any" {
			logger.Debug("Message not for this cluster, skipping",
				zap.String("target_cluster", targetCluster),
				zap.String("my_cluster", clusterID))
			continue
		}

		// Parse dispatch request
		var req DispatchRequest
		if err := json.Unmarshal(msg.Value, &req); err != nil {
			logger.Error("Failed to parse dispatch request",
				zap.Error(err),
				zap.String("raw", string(msg.Value)))
			continue
		}

		logger.Info("Received dispatch request",
			zap.String("agent_id", req.AgentID),
			zap.String("agent_type", req.AgentType),
			zap.String("target_cluster", req.TargetCluster),
			zap.String("requests_topic", req.RequestsTopic))

		// Create the K8s Job
		jobName, err := createAgentJob(ctx, clientset, namespace, clusterID, &req, logger)

		// Send confirmation response
		resp := DispatchResponse{
			AgentID:   req.AgentID,
			AgentType: req.AgentType,
			ClusterID: clusterID,
			JobName:   jobName,
			Success:   err == nil,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if err != nil {
			resp.Error = err.Error()
			logger.Error("Failed to create agent job",
				zap.String("agent_id", req.AgentID),
				zap.Error(err))
		} else {
			logger.Info("Successfully created agent job",
				zap.String("agent_id", req.AgentID),
				zap.String("job_name", jobName))
		}

		respJSON, _ := json.Marshal(resp)
		writeErr := writer.WriteMessages(ctx, kafka.Message{
			Topic: responseTopic,
			Key:   []byte(req.AgentID),
			Value: respJSON,
			Headers: []kafka.Header{
				{Key: "agent_id", Value: []byte(req.AgentID)},
				{Key: "agent_type", Value: []byte(req.AgentType)},
				{Key: "cluster_id", Value: []byte(clusterID)},
				{Key: "correlation_id", Value: []byte(getHeader(msg.Headers, "correlation_id"))},
				{Key: "message_type", Value: []byte("dispatch_response")},
			},
		})
		if writeErr != nil {
			logger.Error("Failed to send dispatch response", zap.Error(writeErr))
		}
	}
}

// createAgentJob creates a K8s Job in the local cluster.
// This mirrors spawnAgentKubernetesJobFromDefinition from spawn_actions.go
// but receives all parameters via the DispatchRequest instead of reading
// from local environment and K8s API.
func createAgentJob(ctx context.Context, clientset *kubernetes.Clientset, namespace, clusterID string, req *DispatchRequest, logger *zap.Logger) (string, error) {
	jobName := fmt.Sprintf("agent-%s-%s", req.AgentType, req.AgentID[:8])

	// Check for existing job
	existingJob, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err == nil {
		if existingJob.Status.Failed > 0 || existingJob.Status.Succeeded > 0 {
			logger.Info("Deleting completed/failed job before recreating",
				zap.String("job_name", jobName))
			deletePolicy := metav1.DeletePropagationForeground
			_ = clientset.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})
			time.Sleep(2 * time.Second)
		} else if existingJob.Status.Active > 0 {
			logger.Info("Job already running", zap.String("job_name", jobName))
			return jobName, nil
		}
	}

	// Parse specs from the dispatch request
	resources := parseResourceSpec(req.Resources)
	healthConfig := parseHealthSpec(req.HealthConfig)
	envVars := parseEnvVarSpecs(req.EnvVars)

	// Resolve infrastructure endpoints.
	// Use overrides from this cluster's env if set, otherwise use what the
	// originating cluster sent in the dispatch message.
	kafkaBrokers := resolveConfig("OVERRIDE_KAFKA_BROKERS", req.KafkaBrokers,
		"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092")
	dbHost := resolveConfig("OVERRIDE_DATABASE_HOST", req.DatabaseHost,
		"pgbouncer-clients.ai-persona-system.svc.cluster.local")
	dbPort := resolveConfig("OVERRIDE_DATABASE_PORT", req.DatabasePort, "6432")
	dbUser := resolveConfig("OVERRIDE_DATABASE_USER", req.DatabaseUser, "clients_user")
	dbName := resolveConfig("OVERRIDE_DATABASE_NAME", req.DatabaseName, "clients_db")
	templatesDBHost := resolveConfig("OVERRIDE_TEMPLATES_DB_HOST", req.TemplatesDBHost,
		"postgres-templates.ai-persona-system.svc.cluster.local")
	templatesDBPort := resolveConfig("OVERRIDE_TEMPLATES_DB_PORT", req.TemplatesDBPort, "5432")
	templatesDBUser := resolveConfig("OVERRIDE_TEMPLATES_DB_USER", req.TemplatesDBUser, "templates_user")
	templatesDBName := resolveConfig("OVERRIDE_TEMPLATES_DB_NAME", req.TemplatesDBName, "templates_db")
	coreManagerURL := resolveConfig("OVERRIDE_CORE_MANAGER_URL", "",
		"http://core-manager.ai-persona-system.svc.cluster.local:8088")

	dbURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbName)

	// Build environment variables — same structure as spawnAgentKubernetesJobFromDefinition
	envList := []corev1.EnvVar{
		// Core config
		{Name: "AGENT_TYPE", Value: req.AgentType},
		{Name: "AGENT_ID", Value: req.AgentID},
		{Name: "CLIENT_ID", Value: req.ClientID},

		// Topics
		{Name: "KAFKA_TOPIC", Value: req.RequestsTopic},
		{Name: "KAFKA_TOPICS", Value: req.RequestsTopic},
		{Name: "RESPONSES_TOPIC", Value: req.ResponsesTopic},
		{Name: "REQUESTS_TOPIC", Value: req.RequestsTopic},
		{Name: "PARENT_RESPONSES_TOPIC", Value: req.ParentResponsesTopic},
		{Name: "KAFKA_CONSUMER_GROUP", Value: fmt.Sprintf("%s-group-%s", req.AgentType, req.AgentID[:8])},

		// Infrastructure
		{Name: "KAFKA_BROKERS", Value: kafkaBrokers},
		{Name: "SERVICE_INFRASTRUCTURE_KAFKA_BROKERS", Value: kafkaBrokers},
		{Name: "DATABASE_URL", Value: dbURL},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_HOST", Value: dbHost},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_PORT", Value: dbPort},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_USER", Value: dbUser},
		{Name: "SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_DB_NAME", Value: dbName},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_HOST", Value: templatesDBHost},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_PORT", Value: templatesDBPort},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_USER", Value: templatesDBUser},
		{Name: "SERVICE_INFRASTRUCTURE_TEMPLATES_DATABASE_DB_NAME", Value: templatesDBName},

		// Health and metrics
		{Name: "HEALTH_PORT", Value: fmt.Sprintf("%d", healthConfig.Port)},
		{Name: "METRICS_PORT", Value: "9090"},
		{Name: "CORE_MANAGER_URL", Value: coreManagerURL},

		// Dispatch metadata — so the agent knows it was remotely dispatched
		{Name: "DISPATCHED_BY_CLUSTER", Value: "true"},
		{Name: "DISPATCHER_CLUSTER_ID", Value: clusterID},
	}

	// Add custom env vars from the agent definition
	for _, ev := range envVars {
		envList = append(envList, corev1.EnvVar{Name: ev.Name, Value: ev.Value})
	}

	// Add secrets — these must exist in the remote cluster's namespace.
	// NOTE: you need to replicate the same secrets to cluster B:
	//   personae-platform-secrets (DB passwords, bootstrap key)
	//   personae-default-secrets  (ANTHROPIC_API_KEY)
	envList = append(envList, []corev1.EnvVar{
		secretEnvVar("CLIENTS_DB_PASSWORD", "personae-platform-secrets", "CLIENTS_DB_PASSWORD"),
		secretEnvVar("PGPASSWORD", "personae-platform-secrets", "CLIENTS_DB_PASSWORD"),
		secretEnvVar("TEMPLATES_DB_PASSWORD", "personae-platform-secrets", "TEMPLATES_DB_PASSWORD"),
		secretEnvVar("AUTH_DB_PASSWORD", "personae-platform-secrets", "AUTH_DB_PASSWORD"),
		secretEnvVar("ANTHROPIC_API_KEY", "personae-default-secrets", "ANTHROPIC_API_KEY"),
		secretEnvVar("AGENT_BOOTSTRAP_KEY", "personae-platform-secrets", "agent-bootstrap-key"),
	}...)

	// Define the Job — same structure as spawnAgentKubernetesJobFromDefinition
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":                "dynamic-agent",
				"agent-type":         req.AgentType,
				"agent-id":           req.AgentID,
				"client-id":          req.ClientID,
				"spawned-by":         "dispatcher",
				"component":          "agent",
				"category":           req.Category,
				"dispatcher-cluster": clusterID,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: int32Ptr(3600),
			BackoffLimit:            int32Ptr(3),
			ActiveDeadlineSeconds:   int64Ptr(86400),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                "dynamic-agent",
						"agent-type":         req.AgentType,
						"agent-id":           req.AgentID,
						"client-id":          req.ClientID,
						"dispatcher-cluster": clusterID,
					},
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9090",
						"prometheus.io/path":   "/metrics",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: "ai-persona-app",
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "docker-hub-creds"},
					},
					Containers: []corev1.Container{
						{
							Name:    "agent",
							Image:   fmt.Sprintf("%s:%s", req.ImageRepository, req.ImageTag),
							Command: req.Command,
							Ports: []corev1.ContainerPort{
								{
									Name:          "health",
									ContainerPort: int32(healthConfig.Port),
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "metrics",
									ContainerPort: 9090,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: envList,
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "personae-prod-config",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(resources.Requests["cpu"]),
									corev1.ResourceMemory: resource.MustParse(resources.Requests["memory"]),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(resources.Limits["cpu"]),
									corev1.ResourceMemory: resource.MustParse(resources.Limits["memory"]),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   healthConfig.LivenessPath,
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: int32(healthConfig.InitialDelaySeconds),
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   healthConfig.ReadinessPath,
										Port:   intstr.FromString("health"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
								TimeoutSeconds:      3,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
						},
					},
				},
			},
		},
	}

	// Create the Job
	createdJob, err := clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			logger.Warn("Job already exists (race condition), continuing",
				zap.String("job_name", jobName))
			return jobName, nil
		}
		return "", fmt.Errorf("failed to create kubernetes job: %w", err)
	}

	logger.Info("Created Kubernetes Job",
		zap.String("job_name", createdJob.Name),
		zap.String("namespace", createdJob.Namespace),
		zap.String("uid", string(createdJob.UID)))

	return createdJob.Name, nil
}

// --- Helper functions ---

func secretEnvVar(name, secretName, key string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
				Key:                  key,
			},
		},
	}
}

func parseResourceSpec(raw json.RawMessage) ResourceSpec {
	var spec ResourceSpec
	if err := json.Unmarshal(raw, &spec); err != nil || spec.Requests == nil {
		return ResourceSpec{
			Requests: map[string]string{"cpu": "100m", "memory": "256Mi"},
			Limits:   map[string]string{"cpu": "500m", "memory": "512Mi"},
		}
	}
	if spec.Limits == nil {
		spec.Limits = map[string]string{"cpu": "500m", "memory": "512Mi"}
	}
	return spec
}

func parseHealthSpec(raw json.RawMessage) HealthSpec {
	var spec HealthSpec
	if err := json.Unmarshal(raw, &spec); err != nil || spec.Port == 0 {
		return HealthSpec{
			Port:                8080,
			LivenessPath:        "/health",
			ReadinessPath:       "/ready",
			InitialDelaySeconds: 30,
		}
	}
	if spec.LivenessPath == "" {
		spec.LivenessPath = "/health"
	}
	if spec.ReadinessPath == "" {
		spec.ReadinessPath = "/ready"
	}
	if spec.InitialDelaySeconds == 0 {
		spec.InitialDelaySeconds = 30
	}
	return spec
}

func parseEnvVarSpecs(raw json.RawMessage) []EnvVarSpec {
	var specs []EnvVarSpec
	if err := json.Unmarshal(raw, &specs); err != nil {
		return nil
	}
	return specs
}

// resolveConfig returns: env override if set, then dispatch value if non-empty,
// then fallback default.
func resolveConfig(envOverride, dispatchValue, fallback string) string {
	if v := os.Getenv(envOverride); v != "" {
		return v
	}
	if dispatchValue != "" {
		return dispatchValue
	}
	return fallback
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getHeader(headers []kafka.Header, key string) string {
	for _, h := range headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
