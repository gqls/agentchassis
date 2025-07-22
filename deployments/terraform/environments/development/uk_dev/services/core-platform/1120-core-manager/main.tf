terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-core-manager-dev"
    config_path   = "~/.kube/config"
  }
}

module "core_manager_deployment_dev" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the DEVELOPMENT overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/core-manager/overlays/development"

  depends_on = [
    module.kafka_topics # Ensure this module name matches your setup
  ]
}