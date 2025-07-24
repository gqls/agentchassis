terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-agent-chassis-dev"
    config_path   = "~/.kube/config"
  }
}

module "agent_chassis_deployment_dev" {
  source = "../../../../../../modules/kustomize-apply"

  # Path to the DEVELOPMENT overlay for this service
  kustomize_path = "../../../../../development/kustomize/services/agent-chassis/overlays/development"
}