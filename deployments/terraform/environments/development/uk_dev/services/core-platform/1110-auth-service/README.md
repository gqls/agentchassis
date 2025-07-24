docker build -f ./build/docker/backend/auth-service.dockerfile -t aqls/auth-service:latest .
docker push aqls/auth-service:latest

cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/services/core-platform/1110-auth-service$ 
terraform taint module.auth_service_deployment_dev.null_resource.apply_kustomization

make create-dev-secrets
--

# Set this in your terminal for the current session
export REGISTRY=docker.io/aqls
export IMAGE_TAG=latest

# Build auth-service
make build-auth-service

# Build core-manager
make build-core-manager

# Push auth-service
make push-auth-service

# Push core-manager
make push-core-manager

# To deploy both core services, run:
make deploy-core ENVIRONMENT=development REGION=uk_dev

# To deploy only the auth-service, run
make deploy-auth-service ENVIRONMENT=development REGION=uk_dev

# verify deployments
kubectl get pods -n ai-persona-system

# Check auth-service logs
make logs-auth PROJECT_NAME=ai-persona-system

# Check core-manager logs
make logs-core PROJECT_NAME=ai-persona-system

# Test connectivity
# Forward the auth-service
kubectl port-forward -n ai-persona-system svc/auth-service 8081:8081

# Forward the core-manager
kubectl port-forward -n ai-persona-system svc/core-manager 8088:8088

Then test with curl http://localhost:8081/health and curl http://localhost:8088/health in separate terminals.

--

export TMPDIR=~/kind-tmp
kind load docker-image docker.io/aqls/auth-service:latest --name personae-dev


# Create the Docker Hub secret
kubectl create secret docker-registry dockerhub-secret \
--docker-server=https://index.docker.io/v1/ \
--docker-username=aqls \
--docker-password="AaD02432123!" \
--docker-email=aaa@designconsultancy.co.uk \
-n ai-persona-system

# Patch the existing deployment
kubectl patch deployment auth-service -n ai-persona-system -p '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"dockerhub-secret"}]}}}}'

# Force a rollout
kubectl rollout restart deployment/auth-service -n ai-persona-system