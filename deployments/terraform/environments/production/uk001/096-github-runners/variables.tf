variable "kubeconfig_path" {
  description = "Path to the kubeconfig for the production uk001 cluster."
  type        = string
  default     = "/home/ant/.kube/config_production_uk001"
}

variable "runner_kustomize_path" {
  description = "The path to the github-actions-runner Kustomize overlay."
  type        = string
  default     = "../../../../../kustomize/services/github-actions-runner/overlays/production/uk_001/"
}

variable "vmsites_kustomize_path" {
  description = "The path to the github-actions-runner-vmsites Kustomize overlay."
  type        = string
  default     = "../../../../../kustomize/services/github-actions-runner-vmsites/overlays/production/uk_001/"
}
