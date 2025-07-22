variable "kube_context_name" {
  description = "The Kubernetes context name for Kind."
  type        = string
  default     = "kind-personae-dev"
}

variable "kubeconfig_path" {
  description = "Path to the kubeconfig file."
  type        = string
  default     = "~/.kube/config"
}

variable "kafka_cluster_name" {
  description = "Name of the Kafka cluster to associate users with."
  type        = string
  default     = "personae-kafka-cluster"
}

variable "kafka_namespace" {
  description = "Namespace where Kafka cluster and users are deployed."
  type        = string
  default     = "kafka"
}