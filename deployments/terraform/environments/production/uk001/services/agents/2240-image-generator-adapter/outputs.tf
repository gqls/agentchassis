output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the image-generator-adapter."
  value       = module.image_generator_adapter_deployment.status
}