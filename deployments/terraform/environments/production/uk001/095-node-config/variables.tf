variable "kubeconfig_path" {
  description = "Path to the kubeconfig for the production uk001 cluster."
  type        = string
  default     = "/home/ant/.kube/config_production_uk001"
}

variable "kustomize_path" {
  description = "The path to the node-config Kustomize overlay to apply."
  type        = string
  default     = "../../../../../kustomize/services/node-config/overlays/production/uk_001/"
}
