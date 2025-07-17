terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-agent-chassis"
    config_path   = "~/.kube/config"
  }
}

module "agent_chassis_deployment" {
  source = "../../../../../modules/kustomize-apply"

  # Path to the production overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/agent-chassis/overlays/production/uk_001"
}