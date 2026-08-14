# 097-ollama — both ollama services (adapter + eval), as an install step
# (owner directive 2026-08-14: the whole framework installs via
# terraform+kustomize).
#
# Two module calls: ollama-adapter (the fleet's local-model endpoint) and
# ollama-eval (the evaluation instance). Both run the upstream ollama/ollama
# image — NOT an aqls-built image — so deployment_name is left empty in both
# calls: the module's set-image block must never sweep them to the platform
# IMAGE_TAG (there is no such ollama image in the registry, and the makefile's
# deploy-agents block already documents this exemption for the adapter).
#
# Day-to-day the same overlays are applied by deploy-agents (`make release`);
# this step is what makes a FRESH INSTALL carry them. Both paths apply the
# same manifests and are idempotent in either order. The module's re-apply
# trigger hashes only files UNDER the overlay dir, so base/ edits do not
# re-trigger terraform — the makefile path is the re-apply mechanism.
#
# Both deployments carry ephemeral-storage requests (bugs_open/252 candidate
# 1, applied 2026-08-12) — repo overlays verified byte-identical to live
# before this step was written (kubectl diff, all four services, 2026-08-14).

terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-ollama"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

provider "kubernetes" {
  config_path = var.kubeconfig_path
}

module "ollama_adapter" {
  source = "../../../../modules/kustomize-apply"

  kustomize_path   = var.adapter_kustomize_path
  service_name     = "ollama-adapter"
  namespace        = "ai-persona-system"
  image_repository = ""
}

module "ollama_eval" {
  source = "../../../../modules/kustomize-apply"

  kustomize_path   = var.eval_kustomize_path
  service_name     = "ollama-eval"
  namespace        = "ai-persona-system"
  image_repository = ""
}
