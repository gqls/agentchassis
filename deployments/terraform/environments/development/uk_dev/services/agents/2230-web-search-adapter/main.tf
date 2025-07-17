terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-web-search-adapter-dev"
    config_path   = "~/.kube/config"
  }
}

module "web_search_adapter_deployment_dev" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the DEVELOPMENT overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/web-search-adapter/overlays/development"
}