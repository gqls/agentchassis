terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-auth-dev"
    config_path   = "~/.kube/config"
  }
}

module "auth_service_deployment_dev" {
  source = "../../../../../../modules/kustomize-apply"

  # Required arguments
  service_name     = "auth-service"
  namespace        = "ai-persona-system"
  image_repository = "docker.io/aqls/auth-service"
  image_tag        = var.image_tag  # Use the variable instead of hardcoding

  # Path to the DEVELOPMENT overlay for this service
  # Adjust the path based on where we are in the directory structure
  # We're in: deployments/terraform/environments/development/uk_dev/services/core-platform/1110-auth-service
  # We need to get to: deployments/kustomize/services/auth-service/overlays/development
  kustomize_path = "../../../../../../../kustomize/services/auth-service/overlays/development"

  # Optional: Add config_sha if you want to trigger updates on config changes
  # config_sha = filesha256("${path.module}/../../../../../deployments/kustomize/services/auth-service/overlays/development/kustomization.yaml")
}