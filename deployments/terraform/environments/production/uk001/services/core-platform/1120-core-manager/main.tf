terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-core-manager"
    config_path   = "~/.kube/config"
  }
}

module "core_manager_deployment" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the production overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/core-manager/overlays/production/uk_001"
}