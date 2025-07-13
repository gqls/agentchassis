#!/bin/bash
# FILE: scripts/generate-jwt-secret.sh

# Generate a secure JWT secret
JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')

echo "Generated JWT Secret (save this securely):"
echo "$JWT_SECRET"
echo ""

# Create/update Kubernetes secret
kubectl create secret generic auth-secrets \
  --from-literal=jwt-secret="$JWT_SECRET" \
  -n ai-persona-system \
  --dry-run=client -o yaml | kubectl apply -f -

echo "JWT secret stored in Kubernetes"