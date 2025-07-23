terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-base-configs-dev"
    config_path   = "~/.kube/config"
  }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "kind-personae-dev"
}

# 1. Explicitly create the namespace using the kubernetes provider.
resource "kubernetes_namespace" "ai_persona_system" {
  metadata {
    name = "ai-persona-system"
    labels = {
      name       = "ai-persona-system"
      monitoring = "enabled"
    }
  }
}

# 2. Apply the rest of the base configs using Kustomize.
#    This resource now depends on the namespace being created first.
module "apply_base_configs" {
  source = "../../../../modules/kustomize-apply"

  # This ensures the namespace is created before this module runs.
  depends_on = [
    kubernetes_namespace.ai_persona_system
  ]

  # Point to the directory containing your new kustomization.yaml
  kustomize_path = "../../../../../kustomize/infrastructure/configs/development"
  image_repository = ""
  namespace = ""
  service_name = ""
}