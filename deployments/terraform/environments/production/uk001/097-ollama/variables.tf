variable "kubeconfig_path" {
  description = "Path to the kubeconfig for the production uk001 cluster."
  type        = string
  default     = "/home/ant/.kube/config_production_uk001"
}

variable "adapter_kustomize_path" {
  description = "The path to the ollama-adapter Kustomize overlay."
  type        = string
  default     = "../../../../../kustomize/services/ollama-adapter/overlays/production/uk_001/"
}

variable "eval_kustomize_path" {
  description = "The path to the ollama-eval Kustomize overlay."
  type        = string
  default     = "../../../../../kustomize/services/ollama-eval/overlays/production/uk_001/"
}
