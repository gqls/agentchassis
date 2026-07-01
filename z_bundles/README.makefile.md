

## Usage Examples:

```bash
# Development workflow
make dev-up                              # Start local environment
make test                                # Run tests
make build-auth-service                  # Build single service
make dev-logs                           # Check logs

# Full deployment
make full-deploy                        # Build, push, and deploy everything

# Deploy infrastructure only
make deploy-infrastructure              # Deploy all infrastructure components

# Deploy applications only
make quick-deploy                       # Deploy apps using existing images

# Individual service workflow
make auth-service                       # Build, push, and deploy auth-service
make logs-auth                         # Check auth-service logs
make rollback-auth-service             # Rollback if needed

# Frontend deployment
make build-frontends                   # Build all frontends
make deploy-admin-dashboard           # Deploy just admin dashboard

# Monitoring
make status                           # Check deployment status
make port-forward-grafana            # Access Grafana locally

# Different environments
make deploy-all ENVIRONMENT=staging REGION=us001
make deploy-all ENVIRONMENT=production REGION=uk001 IMAGE_TAG=v1.2.3


This Makefile provides:
1. **Organized sections** for different types of operations
2. **Individual targets** for each service/component
3. **Composite targets** for common workflows
4. **Environment flexibility** through variables
5. **Colored output** for better readability
6. **Help system** showing all available commands
7. **Safety features** for destructive operations


# Create new agent type on the fly
make agent-create TYPE=market-analyzer NAME="Market Analyzer" DESC="Analyzes market data"

# Update configuration instantly
make agent-update-config TYPE=market-analyzer CONFIG='{"temperature": 0.8, "model": "claude-3-opus"}'

# Hot reload to running agents
make agent-hot-reload AGENT_ID=abc-123 CONFIG='{"interval": "1m"}'

# Spawn new agent instance
make agent-spawn TYPE=market-analyzer CLIENT=demo_client

# Check performance
make agent-performance

# Test discovery
make agent-recommend TASK=website-builder

--=== new ===---
# Complete deployment
make deploy-all

# Quick application update
make quick-deploy

# Build and push everything
make bp  # or make build-push

# Check everything
make status  # or make ps

# View logs
make logs SERVICE=core-manager
make logs TARGET=all

# Database operations
make db-console DB=clients
make db-query DB=templates SQL="SELECT * FROM agent_definitions"

# Port forwarding
make port-forward SERVICE=grafana PORT=3001:3000

----
up = "Bring system UP"

Alias for deploy-all
Like docker-compose up - starts everything
Usage: make up

down = "Bring system DOWN"

Alias for destroy-infra
Like docker-compose down - stops/removes everything
Usage: make down

ps = "Process Status"

Alias for status
Like docker ps or Unix ps - shows what's running
Usage: make ps

bp = "Build & Push"

Alias for build-push (which does build-all + push-all)
Common workflow shortened
Usage: make bp

qd = "Quick Deploy"

Alias for quick-deploy
Deploys just applications (not infrastructure)
Usage: make qd