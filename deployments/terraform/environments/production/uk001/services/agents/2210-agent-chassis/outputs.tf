output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the agent-chassis."
  value       = module.agent_chassis_deployment.status
}