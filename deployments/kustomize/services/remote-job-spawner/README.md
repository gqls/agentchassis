# deployments/kustomize/services/remote-job-spawner/README.md
Remote job spawner. Consumes dispatch requests from Kafka and creates K8s Jobs locally.
Deployed to each cluster (including the primary) to enable cross-cluster agent spawning.
Uses the same ServiceAccount (ai-persona-app) and RBAC pattern as agent-chassis.
