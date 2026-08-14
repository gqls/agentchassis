# 096-github-runners — both self-hosted CI runner deployments, as an install
# step (owner directive 2026-08-14: the whole framework installs via
# terraform+kustomize; the runners previously existed only in the day-to-day
# makefile path).
#
# Two module calls, one per deployment: github-actions-runner (2 replicas,
# repo "sites") and github-actions-runner-vmsites (1 replica). They are ONE
# disk-consumption pool — the shared pod label workload=gha-runner and the
# maxSkew-1 spread constraint in their manifests keep them on separate nodes
# (bugs_open/252 candidate 1b) — but they are separate Deployments with
# separate configmaps and DIFFERENT pinned image tags, so they stay separate
# kustomize overlays and separate module calls.
#
# deployment_name is left empty in BOTH calls, deliberately: the runners' image
# tags are PINNED in their overlays (v1.0.948 / v1.0.1126 at time of writing)
# and do not follow the platform IMAGE_TAG — build-backend/push-backend never
# build a runner image, so the module's set-image block would point them at a
# tag that was never pushed and both would ImagePullBackOff. Same reasoning as
# the makefile's deploy-github-runners target (apply, no sed). To move a
# runner to a new image: make release-github-runner.
#
# Day-to-day the same overlays are applied by `make deploy-github-runners`
# (wired into deploy-agents); this step is what makes a FRESH INSTALL carry
# them. Both paths apply the same manifests and are idempotent in either
# order. Note the module's re-apply trigger hashes only files UNDER the
# overlay dir, so base/ edits do not re-trigger terraform — the makefile path
# is the re-apply mechanism; this step exists for install completeness.
# Depends on 047-base-configs (personae-platform-secrets: GITHUB_TOKEN;
# docker-hub-creds), which runs earlier in the sequence.
#
# Verify after apply: kubectl -n ai-persona-system get pods -o wide \
#   -l workload=gha-runner   # 3 pods, 3 distinct nodes

terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-github-runners"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

provider "kubernetes" {
  config_path = var.kubeconfig_path
}

module "github_actions_runner" {
  source = "../../../../modules/kustomize-apply"

  kustomize_path   = var.runner_kustomize_path
  service_name     = "github-actions-runner"
  namespace        = "ai-persona-system"
  image_repository = ""
}

module "github_actions_runner_vmsites" {
  source = "../../../../modules/kustomize-apply"

  kustomize_path   = var.vmsites_kustomize_path
  service_name     = "github-actions-runner-vmsites"
  namespace        = "ai-persona-system"
  image_repository = ""
}
