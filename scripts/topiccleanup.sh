#!/bin/bash
# FILE: scripts/cleanup-calculator-jobs.sh
# Script to cleanup infinite calculator jobs and pods

echo "Cleaning up calculator jobs and pods..."

# Get the namespace (adjust if different)
NAMESPACE="ai-persona-system"

# Delete all calculator jobs
echo "Deleting calculator jobs..."
kubectl delete jobs -n $NAMESPACE -l app=calculator --grace-period=0 --force 2>/dev/null || true
kubectl delete jobs -n $NAMESPACE --field-selector metadata.name=agent-calculator* --grace-period=0 --force 2>/dev/null || true

# Delete calculator jobs by pattern
for job in $(kubectl get jobs -n $NAMESPACE -o name | grep calculator); do
    echo "Deleting $job"
    kubectl delete $job -n $NAMESPACE --grace-period=0 --force
done

# Delete calculator pods
echo "Deleting calculator pods..."
kubectl delete pods -n $NAMESPACE -l app=calculator --grace-period=0 --force 2>/dev/null || true
kubectl delete pods -n $NAMESPACE --field-selector metadata.name=agent-calculator* --grace-period=0 --force 2>/dev/null || true

# Delete pods by pattern
for pod in $(kubectl get pods -n $NAMESPACE -o name | grep calculator); do
    echo "Deleting $pod"
    kubectl delete $pod -n $NAMESPACE --grace-period=0 --force
done

# Show remaining resources
echo ""
echo "Remaining jobs:"
kubectl get jobs -n $NAMESPACE | grep -E "(calculator|NAME)" || echo "No calculator jobs found"

echo ""
echo "Remaining pods:"
kubectl get pods -n $NAMESPACE | grep -E "(calculator|NAME)" || echo "No calculator pods found"

echo ""
echo "Cleanup completed!"

# Optional: Scale down the main agent to stop it from creating more
read -p "Do you want to scale down agent-chassis deployment to stop spawning? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    kubectl scale deployment agent-chassis -n $NAMESPACE --replicas=0
    echo "Agent chassis scaled down. Scale it back up with:"
    echo "kubectl scale deployment agent-chassis -n $NAMESPACE --replicas=1"
fi
