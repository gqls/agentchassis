terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-reasoning-agent"
    config_path   = "~/.kube/config"
  }
}

module "reasoning_agent_deployment" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the production overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/reasoning-agent/overlays/production/uk_001"
}