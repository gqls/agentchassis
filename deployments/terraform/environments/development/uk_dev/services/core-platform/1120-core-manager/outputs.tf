output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the development core-manager."
  value       = module.core_manager_deployment_dev.status
}