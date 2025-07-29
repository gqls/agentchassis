# infrastructure/environments/production/uk001/100-svc-agent-chassis/outputs.tf

output "kustomize_apply_status" {
  description = "The status of the Kustomize for production agent-chassis."
  value       = "Deployment completed"  # Simple string output
}