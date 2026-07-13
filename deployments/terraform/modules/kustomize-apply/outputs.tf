output "deployment_info" {
  description = "Information about the deployed service"
  value = {
    service_name     = var.service_name
    namespace        = var.namespace
    image_repository = var.image_repository
    image_tag        = var.image_tag
    kustomize_path   = var.kustomize_path
  }
}

output "deployment_complete" {
  description = "Indicates whether the deployment was triggered"
  value       = true
  depends_on  = [null_resource.apply_kustomization]
}