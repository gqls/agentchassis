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

# Reference the existing namespace
data "kubernetes_namespace" "ai_persona_system" {
  metadata {
    name = "ai-persona-system"
  }
}

module "apply_base_configs" {
  source = "../../../../modules/kustomize-apply"

  kustomize_path = "../../../../../kustomize/infrastructure/configs/development"

  # Use the namespace from the data source
  namespace        = data.kubernetes_namespace.ai_persona_system.metadata[0].name
  image_repository = ""
  service_name     = ""
}