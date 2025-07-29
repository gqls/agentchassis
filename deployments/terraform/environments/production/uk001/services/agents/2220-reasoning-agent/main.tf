terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-reasoning-agent"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

module "reasoning_agent_deployment" {
  source = "../../../../../../modules/kustomize-apply"

  # Path to the production overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/reasoning-agent/overlays/production/uk_001"
  image_repository = ""
  namespace = ""
  service_name = ""
}