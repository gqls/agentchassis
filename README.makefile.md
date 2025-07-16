

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