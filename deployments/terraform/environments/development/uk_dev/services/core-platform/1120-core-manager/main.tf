terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-core-manager-dev"
    config_path   = "~/.kube/config"
  }
}

# main.tf for the core-manager service

module "core_manager_deployment_dev" {
  # The source should point to your reusable service deployment module.
  # Based on your structure, it's likely a kustomize-apply or similar module.
  source = "../../../../../../modules/kustomize-apply"

#  /home/ant/projects/agent-chassis/deployments/terraform/modules/kustomize-apply/main.tf
#  /home/ant/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/services/core-platform/1120-core-manager/main.tf

  # --- Add these required variables ---
  service_name     = "core-manager"
  namespace        = "ai-persona-system"
  image_repository = "aqls/core-manager" # Assuming 'aqls' is your Docker Hub username. Adjust if needed.
  image_tag        = "latest" # You can change this tag as needed.

  # The path to the specific kustomize overlay for this service
  kustomize_path   = "../../../../../../../kustomize/services/core-manager/overlays/development"
}