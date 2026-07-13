# FILE: deployments/terraform/environments/production/uk_001/services/agents/2250-content-creator-agent/main.tf
terraform {
  backend "kubernetes" {
    # Unique suffix for this service's tfstate
    secret_suffix = "tfstate-svc-content-creator-agent"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

module "content_creator_agent_deployment" {
  source = "../../../../../../modules/kustomize-apply" # Path to your kustomize-apply module

  # Path to the production overlay for this new service
  kustomize_path = "../../../../../deployments/kustomize/services/content-creator-agent/overlays/production/uk_001"

  # These variables are passed through to the kustomize-apply module.
  # They might be used for image overrides or namespace setting if your kustomize-apply module supports them.
  # For now, keep them as empty strings as your current module definition has them.
  image_repository = ""
  namespace = ""
  service_name = ""
}