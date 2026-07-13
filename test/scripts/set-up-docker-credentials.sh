#!/bin/bash
# test/scripts/setup-docker-credentials.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Docker Hub Kubernetes Secret Setup ===${NC}"
echo ""
echo "This script will create a Kubernetes secret for pulling images from Docker Hub."
echo ""

# Check if secret already exists
if kubectl get secret docker-hub-creds -n ai-persona-system &> /dev/null; then
    echo -e "${YELLOW}Warning: Secret 'docker-hub-creds' already exists in namespace 'ai-persona-system'${NC}"
    read -p "Do you want to update it? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Exiting without changes."
        exit 0
    fi
fi

# Get Docker Hub credentials
echo "Please enter your Docker Hub credentials:"
read -p "Docker Hub username: " DOCKER_USERNAME
read -s -p "Docker Hub password/token: " DOCKER_PASSWORD
echo
read -p "Docker Hub email: " DOCKER_EMAIL

# Create the secret
echo -e "\n${YELLOW}Creating Kubernetes secret...${NC}"
kubectl create secret docker-registry docker-hub-creds \
    --docker-server=docker.io \
    --docker-username="$DOCKER_USERNAME" \
    --docker-password="$DOCKER_PASSWORD" \
    --docker-email="$DOCKER_EMAIL" \
    -n ai-persona-system \
    --dry-run=client -o yaml | kubectl apply -f -

# Verify the secret
if kubectl get secret docker-hub-creds -n ai-persona-system &> /dev/null; then
    echo -e "${GREEN}✓ Docker Hub secret created successfully${NC}"

    # Test the secret by creating a test pod
    echo -e "\n${YELLOW}Testing image pull...${NC}"

    kubectl run test-pull-${RANDOM} \
        --image=docker.io/aqls/test-harness:latest \
        --rm -it \
        --restart=Never \
        -n ai-persona-system \
        --image-pull-policy=Always \
        --command -- echo "Image pull successful" || \
        echo -e "${RED}Warning: Test pull failed. Please verify your credentials.${NC}"
else
    echo -e "${RED}✗ Failed to create Docker Hub secret${NC}"
    exit 1
fi

echo -e "\n${GREEN}Setup complete!${NC}"
echo "You can now deploy test harness with: make harness-deploy"