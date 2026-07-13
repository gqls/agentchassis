terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-web-search-adapter"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

module "web_search_adapter_deployment" {
  source = "../../../../../../modules/kustomize-apply"

  # Path to the DEVELOPMENT overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/web-search-adapter/overlays/production/uk_001"
  image_repository = ""
  namespace = ""
  service_name = ""
}