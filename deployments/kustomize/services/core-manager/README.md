kubectl -n ai-persona-system describe deployment core-manager

kubectl apply -f deployments/kustomize/base/rbac-security.yaml -n ai-persona-system

