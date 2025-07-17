terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-image-generator-adapter-dev"
    config_path   = "~/.kube/config"
  }
}

module "image_generator_adapter_deployment_dev" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the DEVELOPMENT overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/image-generator-adapter/overlays/development"
}