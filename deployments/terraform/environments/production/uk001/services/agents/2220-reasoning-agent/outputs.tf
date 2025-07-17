output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the reasoning-agent."
  value       = module.reasoning_agent_deployment.status
}