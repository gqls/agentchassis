# deployments/terraform/environments/production/uk001/services/agents/2220-agent-dispatcher/outputs.tf

output "kustomize_apply_status" {
  description = "The status of the Kustomize for production agent-dispatcher."
  value       = "Deployment completed"
}
