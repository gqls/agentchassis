terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-core-manager"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

# main.tf for the core-manager service

module "core_manager_deployment" {
  # The source should point to your reusable service deployment module.
  # Based on your structure, it's likely a kustomize-apply or similar module.
  source = "../../../../../../modules/kustomize-apply"

#  /home/ant/projects/agent-chassis/deployments/terraform/modules/kustomize-apply/main.tf
#  /home/ant/projects/agent-chassis/deployments/terraform/environments/production/uk001/services/core-platform/1120-core-manager/main.tf

  # --- Add these required variables ---
  service_name     = "core-manager"
  namespace        = "ai-persona-system"
  image_repository = "aqls/core-manager"
  image_tag        = "latest"

  # The path to the specific kustomize overlay for this service
  kustomize_path   = "../../../../../../../kustomize/services/core-manager/overlays/production/uk_001/"
}