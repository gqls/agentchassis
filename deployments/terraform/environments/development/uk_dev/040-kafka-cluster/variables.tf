variable "kube_context_name" {
  description = "The Kubernetes context name for Kind (e.g., kind-personae-dev)."
  type        = string
  default     = "kind-personae-dev"
}

variable "kubeconfig_path" { // Specific name for this component's var
  description = "Path to the kubeconfig file to be used for this dev component."
  type        = string
  default     = "~/.kube/config" // Default for Kind, overridden by Makefile if necessary
}

variable "kafka_namespace_dev" {
  description = "Namespace where the Kafka CR for dev will be deployed."
  type        = string
  default     = "kafka"
}

variable "kafka_cluster_cr_yaml_path_dev" {
  description = "Path to the Kafka CR YAML file for the dev instance."
  type        = string
  default     = "../../../../modules/kafka-cluster/config/kafka-cluster-cr-dev.yaml" // Point to your DEV version
}

variable "kafka_cluster_name_dev" {
  description = "The metadata.name of the Kafka cluster for dev."
  type        = string
  default     = "personae-kafka-cluster" // Should match the name in kafka-cluster-cr-dev.yaml
}