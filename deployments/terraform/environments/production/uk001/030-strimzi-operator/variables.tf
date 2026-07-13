# terraform/environments/production/uk001/030-strimzi-operator/variables.tf
variable "kubeconfig_path" {
  description = "Path to the kubeconfig file for the uk001 cluster."
  type        = string
  sensitive   = true
  default = "/home/ant/.kube/config_production_uk001"
}

variable "strimzi_operator_target_namespace" {
  description = "Namespace for the Strimzi operator in uk001."
  type        = string
  default     = "strimzi"
}

variable "watched_namespaces_for_uk001" { // Changed from _for_sydney
  description = "List of namespaces for the Strimzi operator to watch in uk001."
  type        = list(string)
  default     = ["kafka", "strimzi"] // Assuming same watched namespaces for now
}

variable "strimzi_yaml_bundle_path_for_uk001" { // Changed from _for_sydney
  description = "Path to the Strimzi YAML files directory for this instance."
  type        = string
  # This relative path should still be correct from the new uk001 directory
  default     = "../../../../modules/strimzi-operator/strimzi-0.47.0/"
}

variable "kube_context_name" {
  description = "Kubernetes context name to use"
  type        = string
}

variable "strimzi_version" {
  description = "Version of Strimzi to install"
  type        = string
  default     = "0.47.0"
}