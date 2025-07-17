terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-auth-dev"
    config_path   = "~/.kube/config"
  }
}

module "auth_service_deployment_dev" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the DEVELOPMENT overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/auth-service/overlays/development"
}