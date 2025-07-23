# Since the kustomize-apply module doesn't provide outputs,
# we'll create meaningful outputs based on the inputs we provided

output "service_name" {
  description = "The name of the deployed service"
  value       = "auth-service"
}

output "namespace" {
  description = "The namespace where the service is deployed"
  value       = "ai-persona-system"
}

output "image" {
  description = "The full image reference used for deployment"
  value       = "docker.io/aqls/auth-service:${var.image_tag}"
}

output "kustomize_path" {
  description = "The path to the Kustomize overlay that was applied"
  value       = "../../../../../../../kustomize/services/auth-service/overlays/development"
}