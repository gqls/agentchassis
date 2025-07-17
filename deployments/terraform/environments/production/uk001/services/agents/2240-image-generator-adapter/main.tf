terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-image-generator-adapter"
    config_path   = "~/.kube/config"
  }
}

module "image_generator_adapter_deployment" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the production overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/image-generator-adapter/overlays/production/uk_001"
}