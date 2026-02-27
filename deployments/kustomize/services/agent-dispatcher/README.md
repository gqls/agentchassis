# deployments/kustomize/services/agent-dispatcher/README.md
Remote agent dispatcher. Consumes dispatch requests from Kafka and creates K8s Jobs locally.
Deployed to remote clusters to enable cross-cluster agent spawning. Uses the same ServiceAccount
(ai-persona-app) and RBAC as agent-chassis since it creates the same Job specs.
