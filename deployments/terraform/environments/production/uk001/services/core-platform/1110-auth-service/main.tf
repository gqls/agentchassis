terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-auth"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

module "auth_service_deployment" {
  source = "../../../../../../modules/kustomize-apply"

  # Required arguments
  service_name     = "auth-service"
  namespace        = "ai-persona-system"
  image_repository = "docker.io/aqls/auth-service"
  image_tag        = var.image_tag  # Use the variable instead of hardcoding

  # Path to the production overlay for this service
  kustomize_path = "../../../../../../../kustomize/services/auth-service/overlays/production/uk_001/"
}