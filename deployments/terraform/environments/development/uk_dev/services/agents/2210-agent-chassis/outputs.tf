output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the development agent-chassis."
  value       = module.agent_chassis_deployment_dev.status
}