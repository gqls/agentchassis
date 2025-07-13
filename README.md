// FILE: README.md
# AI Persona System

A cloud-native, multi-agent microservices platform for AI-powered content generation and workflow orchestration.

## Architecture Overview

The system implements a composable multi-agent architecture with:
- **Decoupled Core Services**: Authentication, core data management, and frontend
- **AI Agent Fleet**: Generic chassis-based agents and specialized code-driven agents
- **Central Communication**: Kafka-based asynchronous messaging
- **Shared Platform**: Reusable Go libraries for common functionality
- **Purpose-Built Storage**: PostgreSQL with pgvector, MySQL, and MinIO

## Quick Start

### Prerequisites
- Kubernetes cluster (local or cloud)
- kubectl configured
- Docker
- Go 1.21+

### Setup
```bash
# 1. Clone the repository
git clone https://github.com/gqls/ai-persona-system.git
cd ai-persona-system

# 2. Set up environment variables
cp .env.example .env
# Edit .env with your API keys

# 3. Run the complete setup
make quickstart
```

This will:
- Build all Docker images
- Deploy to Kubernetes
- Run database migrations
- Seed initial data
- Set up port forwarding

### Accessing Services
After setup, services are available at:
- Auth API: http://localhost:8081
- Core API: http://localhost:8088
- Grafana: http://localhost:3000
- MinIO Console: http://localhost:9001

## Development

### Building Individual Services
```bash
# Build a specific service
docker build -f Dockerfile.auth-service -t ai-persona-system/auth-service:latest .

# Build all services
make build
```

### Running Tests
```bash
# Unit tests
make test

# Integration tests
make test-integration

# System tests
./scripts/test-system.sh
```

### Creating a New Client
```bash
make create-client
# Enter client ID when prompted
```

## Key Components

### Platform Libraries (`/platform`)
- **config**: Centralized configuration management
- **database**: PostgreSQL and MySQL connection utilities
- **kafka**: Message producer/consumer wrappers
- **storage**: S3/MinIO object storage client
- **orchestration**: Saga pattern coordinator
- **governance**: Fuel/cost control system
- **logger**: Structured logging with Zap
- **observability**: OpenTelemetry tracing

### Core Services
- **auth-service**: User authentication, JWT tokens, subscription management
- **core-manager**: Template and instance management, project handling
- **agent-chassis**: Generic agent runtime for configuration-driven agents

### Specialized Agents
- **reasoning-agent**: Complex reasoning and analysis tasks
- **image-generator-adapter**: Integration with image generation APIs
- **web-search-adapter**: Web search capabilities

## Monitoring

Access Grafana at http://localhost:3000 (admin/admin) to view:
- Workflow success rates
- Agent response times
- Kafka consumer lag
- System health metrics

## Troubleshooting

### View logs
```bash
make logs
# Enter service name when prompted
```

### Check system status
```bash
make status
```

### Port forwarding issues
```bash
make stop-port-forward
make port-forward
```

## License

Copyright (c) 2024 AI Persona System. All rights reserved.