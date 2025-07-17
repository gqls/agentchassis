terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-web-search-adapter"
    config_path   = "~/.kube/config"
  }
}

module "web_search_adapter_deployment" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the production overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/web-search-adapter/overlays/production/uk_001"
}