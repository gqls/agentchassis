#!/bin/bash
# deploy-agents.sh

IMAGE_TAG=${1:-v1.0.44}

echo "Deploying agents with image tag: $IMAGE_TAG"

# Build and push
make build-agent-chassis push-agent-chassis IMAGE_TAG=$IMAGE_TAG

# Update ConfigMap
kubectl patch configmap personae-prod-config -n ai-persona-system --type merge -p "{\"data\":{\"AGENT_IMAGE_TAG\":\"$IMAGE_TAG\",\"agent_image_tag\":\"$IMAGE_TAG\"}}"

# Deploy agent-chassis
kubectl set image deployment/agent-chassis agent-chassis=docker.io/aqls/agent-chassis:$IMAGE_TAG -n ai-persona-system

# Wait for rollout
kubectl rollout status deployment/agent-chassis -n ai-persona-system

# Delete old jobs
kubectl delete jobs -n ai-persona-system -l spawned-by=orchestrator

echo "Deployment complete. Agents will be respawned with image: $IMAGE_TAG"