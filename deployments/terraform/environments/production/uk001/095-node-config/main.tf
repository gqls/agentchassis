# 095-node-config — the node-config DaemonSet (concept register BLD-021,
# bugs_open/252 candidates 3a+3b).
#
# Applies deployments/kustomize/services/node-config/, which sets each node's
# kubelet image-GC thresholds (85/80/0s -> 70/60/168h) by editing
# /var/lib/kubelet/config.yaml in place and restarting kubelet, idempotently,
# re-applying itself whenever a spot node is replaced.
#
# WHY THIS EXISTS AS AN INSTALL STEP: the kubeadm route — editing the
# kube-system/kubelet-config ConfigMap — is PROVIDER-PROTECTED on this hosted
# control plane (writes return 200 "patched" and are silently reverted;
# measured 2026-08-14, see LANDMINES.md). Node files are the only
# tenant-writable home for kubelet settings, and a DaemonSet is the only
# carrier that survives spot-node replacement. Without this step, a fresh
# cluster runs with the image-GC trigger (85% used) sitting exactly on the
# imagefs eviction line (15% free) — the confirmed root cause of
# bugs_open/252: pods rejected mid-roll whenever a deploy lands in the
# reclaim cycle's trough.
#
# No image, no tag, no rollout wait: deployment_name is left empty so the
# kustomize-apply module skips its image-set/rollout block — the DaemonSet
# runs pinned busybox and must never be swept to the platform IMAGE_TAG.
# Day-to-day the same overlay is applied by `make deploy-node-config` (wired
# into deploy-agents); this step is what makes a FRESH INSTALL carry it.
# Verify after apply: `make node-config-status` — every node's kubelet must
# read high 70 / low 60 / maxAge 168h.

terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-node-config"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

provider "kubernetes" {
  config_path = var.kubeconfig_path
}

module "node_config_daemonset" {
  source = "../../../../modules/kustomize-apply"

  kustomize_path   = var.kustomize_path
  service_name     = "node-config"
  namespace        = "ai-persona-system"
  image_repository = ""
  # deployment_name deliberately left at its "" default: this is a DaemonSet,
  # and the module's set-image/rollout-status block (a) targets a Deployment
  # and (b) would retag the pinned busybox image. Empty skips both.
}
