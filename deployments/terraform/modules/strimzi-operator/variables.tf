variable "strimzi_version" {
  description = "Version of Strimzi to install"
  type        = string
  default     = "0.47.0"
}

variable "operator_namespace" {
  description = "Namespace where the Strimzi operator will be deployed"
  type        = string
  default     = "strimzi"
}

variable "watched_namespaces_list" {
  description = "List of namespaces that Strimzi operator should watch"
  type        = list(string)
  default     = ["kafka", "strimzi"]
}

variable "cluster_kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = ""
}