output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the development reasoning-agent."
  value       = module.reasoning_agent_deployment_dev.status
}