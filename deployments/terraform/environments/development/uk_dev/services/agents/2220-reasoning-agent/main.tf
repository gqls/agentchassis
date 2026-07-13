terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-reasoning-agent-dev"
    config_path   = "~/.kube/config"
  }
}

module "reasoning_agent_deployment_dev" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the DEVELOPMENT overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/reasoning-agent/overlays/development"
}