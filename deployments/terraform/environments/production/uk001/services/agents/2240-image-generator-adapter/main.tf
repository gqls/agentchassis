terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-image-generator-adapter"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

module "image_generator_adapter_deployment" {
  source = "../../../../../../modules/kustomize-apply"

  # Path to the DEVELOPMENT overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/image-generator-adapter/overlays/production/uk_001"
  image_repository = ""
  namespace = ""
  service_name = ""
}