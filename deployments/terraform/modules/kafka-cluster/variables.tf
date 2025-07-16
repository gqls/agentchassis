variable "kafka_cr_namespace" {
  description = "Namespace where the Kafka CR will be applied (must be watched by Strimzi operator)."
  type        = string
}

variable "kafka_cr_yaml_file_path" {
  description = "Path to the Kafka Custom Resource YAML file."
  type        = string
}

variable "kubeconfig_path" {
  description = "Path to the kubeconfig file for the target Kubernetes cluster."
  type        = string
  sensitive   = true
}

variable "kube_context_name" {
  description = "The kubectl context to use for applying resources. Must be valid for the provided kubeconfig_path."
  type        = string
  # This will be provided by the calling component
}

# Variables to construct output values, assuming fixed naming conventions from Strimzi
variable "kafka_cr_cluster_name" {
  description = "The metadata.name of the Kafka cluster defined in the CR YAML."
  type        = string
}