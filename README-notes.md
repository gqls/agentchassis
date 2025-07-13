
# set up .env file with real keys

# Enable pgvector on clients database
psql -h localhost -U clients_user -d clients_db -f platform/database/migrations/001_enable_pgvector.sql

# Create schemas
psql -h localhost -U templates_user -d templates_db -f platform/database/migrations/002_create_templates_schema.sql

docker-compose up -d

# Clone and enter the project
cd ai-persona-system

# Run the setup script
chmod +x scripts/setup.sh
./scripts/setup.sh

# Build and deploy
make build
make deploy

# Access locally
make port-forward

Create a client
make create-client
# Enter: client_123

# Access the API
curl -X POST http://localhost:8081/api/v1/auth/register \
-H "Content-Type: application/json" \
-d '{"email":"user@example.

# Continue from registration...
curl -X POST http://localhost:8081/api/v1/auth/register \
-H "Content-Type: application/json" \
-d '{"email":"user@example.com","password":"securepass123","client_id":"client_123"}'

# Login to get token
curl -X POST http://localhost:8081/api/v1/auth/login \
-H "Content-Type: application/json" \
-d '{"email":"user@example.com","password":"securepass123"}' \
| jq -r '.access_token' > token.txt

# Create an agent instance
TOKEN=$(cat token.txt)
curl -X POST http://localhost:8088/api/v1/personas/instances \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{"template_id":"00000000-0000-0000-0000-000000000001","instance_name":"My Copywriter"}'

// Example: Send a copywriting task
producer, _ := kafka.NewProducer([]string{"localhost:9092"}, logger)

taskPayload := map[string]interface{}{
"action": "generate_blog_post",
"data": map[string]interface{}{
"topic": "AI and the Future of Work",
"tone": "professional",
"length": "1000 words",
},
}

headers := map[string]string{
"correlation_id": uuid.NewString(),
"request_id": uuid.NewString(),
"client_id": "client_123",
"agent_instance_id": "your-agent-id",
"fuel_budget": "100",
}

payloadBytes, _ := json.Marshal(taskPayload)
producer.Produce(ctx, "system.tasks.copywriter", headers, nil, payloadBytes)

# View logs
make logs
# Enter: agent-chassis

# Check Grafana dashboards
# Open http://localhost:3000
# Default login: admin/admin

==
# Using the Makefile
make deploy

# Or manually in order:
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/network-policies.yaml
kubectl apply -f k8s/configmap-common.yaml
# ... (rest of the files as shown in Makefile)
k8s/
├── namespace.yaml
├── configmap-common.yaml
├── network-policies.yaml
├── postgres-clients.yaml
├── postgres-templates.yaml
├── mysql-auth.yaml
├── kafka.yaml
├── minio.yaml
├── auth-service.yaml
├── core-manager.yaml
├── agent-chassis.yaml
├── reasoning-agent.yaml
├── image-generator-adapter.yaml
├── web-search-adapter.yaml
├── ingress.yaml
├── secrets-template.yaml
└── monitoring/
├── prometheus-config.yaml
├── prometheus.yaml
├── grafana.yaml
└── service-monitor.yaml


# Option 1: Do everything automatically
make quickstart

# Option 2: Step by step
make setup          # Run the setup.sh script
make build          # Build Docker images  
make deploy         # Deploy to Kubernetes
make migrate-up     # Run database migrations
make seed-data      # Add initial templates
make port-forward   # Access services locally

--

If Using Standard Prometheus (your current setup):

You don't need ServiceMonitor
Your prometheus-config.yaml already uses annotation-based discovery
Services just need the prometheus.io/scrape: "true" annotation

If Using Prometheus Operator:

Use the ServiceMonitor definitions provided
Install with: kubectl apply -f k8s/monitoring/service-monitor.yaml

To Enable Metrics:

Ensure services expose metrics port (9090)
Implement /metrics endpoint in your Go services
Add Prometheus metrics to your code (you already have this in platform/observability/metrics.go)

Quick Check:
# Verify metrics are exposed
kubectl port-forward -n ai-persona-system svc/auth-service 9090:9090
curl http://localhost:9090/metrics

The ServiceMonitor approach provides more fine-grained control over metric collection and is the preferred method when using Prometheus Operator in production environments.


# Complete deployment from scratch
make quickstart

# Step-by-step deployment
make setup
make build  
make deploy

# Create a new client
make create-client-schema CLIENT_ID=client_123

# Check system health
make system-check
make status

# Access services locally
make port-forward

# View logs
make logs
# (then type: auth-service)

# Register a new agent type
make register-agent

# Clean up everything
make clean