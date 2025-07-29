# FILE: deployments/terraform/environments/production/uk_001/services/agents/2250-content-creator-agent/outputs.tf
output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the production content-creator-agent."
  value       = "Deployment completed"
}