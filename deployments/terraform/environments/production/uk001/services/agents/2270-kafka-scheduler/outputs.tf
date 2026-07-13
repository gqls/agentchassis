# deployments/terraform/environments/production/uk001/services/agents/2270-kafka-scheduler/outputs.tf

output "kustomize_apply_status" {
  description = "The status of the Kustomize for production kafka-scheduler."
  value       = "Deployment completed"
}
