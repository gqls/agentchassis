output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the development auth-service."
  value       = module.auth_service_deployment_dev.status
}