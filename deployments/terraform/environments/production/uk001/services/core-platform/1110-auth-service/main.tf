terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-auth"
    config_path   = "~/.kube/config"
  }
}

module "auth_service_deployment" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the production overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/auth-service/overlays/production/uk_001"
}