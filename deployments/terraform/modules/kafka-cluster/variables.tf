variable "kafka_cr_namespace" {
  description = "Namespace where the Kafka CR will be applied (must be watched by Strimzi operator)."
  type        = string
}

variable "kafka_cr_yaml_file_path" {
  description = "Path to the Kafka Custom Resource YAML file."
  type        = string
}

variable "kafka_nodepool_cr_yaml_file_path" {
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

variable "use_node_pools" {
  description = "Whether to use Kafka node pools"
  type        = bool
  default     = true
}

variable "node_pool_name" {
  description = "Name of the Kafka node pool"
  type        = string
  default     = "combined-pool"
}

variable "node_pool_replicas" {
  description = "Number of replicas for the Kafka node pool"
  type        = number
  default     = 3
}

variable "node_pool_storage_size" {
  description = "Storage size for each Kafka node"
  type        = string
  default     = "50Gi"
}

variable "kafka_version" {
  description = "Kafka version to deploy"
  type        = string
  default     = "3.9.0"
}

variable "replication_factor" {
  description = "Replication factor for Kafka topics"
  type        = number
  default     = 3
}

variable "min_insync_replicas" {
  description = "Minimum in-sync replicas for Kafka"
  type        = number
  default     = 2
}

variable "message_max_bytes" {
  description = "Max size of messages for Kafka messaging"
  type        = number
  default     = 5242880
}

variable "storage_class" {
  description = "Storage class to use for Kafka volumes"
  type        = string
  default     = "standard"
}

variable "delete_claim" {
  description = "Whether to delete PVC when Kafka is deleted"
  type        = bool
  default     = false
}

variable "kafka_cluster_name" {
  description = "Name of the Kafka cluster"
  type        = string
  default     = "personae-kafka-cluster"
}

variable "kafka_namespace" {
  description = "Namespace for Kafka"
  type        = string
  default     = "kafka"
}