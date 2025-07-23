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

# main.tf for 010-infrastructure

resource "kubernetes_namespace" "ai_persona_system" {
  metadata {
    name = "ai-persona-system"
    labels = {
      name       = "ai-persona-system"
      monitoring = "enabled"
    }
  }
}

module "apply_base_configs" {
  source = "../../../../modules/kustomize-apply"

  depends_on = [
    kubernetes_namespace.ai_persona_system
  ]

  kustomize_path = "../../../../../kustomize/infrastructure/configs/development"

  # Add the namespace variable here.
  # We are NOT setting deployment_name, so it will be skipped.
  namespace      = kubernetes_namespace.ai_persona_system.metadata[0].name
  image_repository = ""
  service_name = ""
}