variable "kind_cluster_name" {
  description = "Name for the Kind cluster for development."
  type        = string
  default     = "personae-dev"
}

variable "kind_node_image" {
  description = "Node image for Kind cluster (e.g., kindest/node:v1.30.10)."
  type        = string
  default     = "kindest/node:v1.30.10" # Choose a version
}

variable "kind_config_path" {
  description = "Optional path to a Kind configuration YAML file."
  type        = string
  default     = null # If null, a default single-node cluster is created
}

variable "kubeconfig_path" {
  description = "Optional path to kubeconfig YAML file."
  type        = string
  default     = null # If null, a default single-node cluster is created
}

# This output is not directly from a resource, but reflects the context name
# that will be used by other components.
variable "kube_context_name" {
  description = "The kubectl context name to use for this Kind cluster."
  type        = string
  default     = "kind-personae-dev" # Must match KIND_CONTEXT_DEV in Makefile
}
