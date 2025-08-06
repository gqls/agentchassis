#!/bin/bash
# test/e2e/scripts/setup_e2e.sh

set -e

echo "Setting up E2E test environment..."

# Ensure namespace exists
kubectl create namespace test-e2e --dry-run=client -o yaml | kubectl apply -f -

# Apply test configurations
kubectl apply -f test/k8s/configmaps/test-configs.yaml -n test-e2e

# Start test agents if needed
kubectl apply -f test/k8s/pods/kafka-test-pod.yaml -n test-e2e
kubectl apply -f test/k8s/pods/agent-test-pod.yaml -n test-e2e

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=test -n test-e2e --timeout=60s

echo "E2E test environment ready"