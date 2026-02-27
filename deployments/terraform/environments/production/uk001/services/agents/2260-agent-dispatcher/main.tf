# deployments/terraform/environments/production/uk001/services/agents/2220-agent-dispatcher/main.tf
terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-agent-dispatcher"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

provider "kubernetes" {
  config_path    = "/home/ant/.kube/config_production_uk001"
}

module "agent_dispatcher_deployment" {
  source = "../../../../../../modules/kustomize-apply"

  # Path to the production overlay for this service
  kustomize_path = "../../../../../deployments/kustomize/services/agent-dispatcher/overlays/production/uk_001"
  image_repository = ""
  namespace = ""
  service_name = ""
}
