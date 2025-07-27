# terraform/environments/production/uk001/040-kafka-cluster/variables.tf
variable "kubeconfig_path" {
  description = "Path to the kubeconfig file for the uk001 cluster."
  type        = string
  sensitive   = true
}

# kubectl config get-contexts --kubeconfig=~/.kube/config_production_uk001
variable "kube_context_name" {
  description = "Kubernetes context name from kubeconfig"
  type        = string
  default     = "uk001-prod-agent-chassis-cluster"
}

variable "target_kafka_namespace" {
  description = "Namespace where the Kafka CR for uk001 will be deployed."
  type        = string
  default     = "kafka"
}

variable "kafka_cluster_cr_yaml_path_uk001" { // Changed from _sydney
  description = "Path to the Kafka CR YAML file for the uk001 instance."
  type        = string
  # This relative path should still correctly point to the shared module's config
  default     = "../../../../modules/kafka-cluster/templates/kafka-cluster-cr-prod.yaml.tpl"
}

variable "kafka_cluster_nodepool_cr_yaml_path_uk001" { // Changed from _sydney
  description = "Path to the Kafka CR YAML file for the uk001 instance."
  type        = string
  # This relative path should still correctly point to the shared module's config
  default     = "../../../../modules/kafka-cluster/templates/kafka-nodepool.yaml.tpl"
}



variable "kafka_cluster_name_uk001" { // Changed from _sydney
  description = "The metadata.name of the Kafka cluster being deployed in uk001 (must match name in YAML)."
  type        = string
  default     = "personae-kafka-cluster" # Assuming you want to use the same Kafka cluster name internally
}

variable "use_node_pools" {
  description = "Whether to use Kafka node pools"
  type        = bool
  default     = true
}

variable "node_pool_name" {
  description = "Name of the Kafka node pool"
  type        = string
}

variable "node_pool_replicas" {
  description = "Number of replicas for the Kafka node pool"
  type        = number
}

variable "node_pool_storage_size" {
  description = "Storage size for each Kafka node"
  type        = string
}

variable "kafka_version" {
  description = "Kafka version to deploy"
  type        = string
}

variable "replication_factor" {
  description = "Replication factor for Kafka topics"
  type        = number
}

variable "min_insync_replicas" {
  description = "Minimum in-sync replicas for Kafka"
  type        = number
}

variable "storage_class" {
  description = "Storage class to use for Kafka volumes"
  type        = string
}

variable "delete_claim" {
  description = "Whether to delete PVC when Kafka is deleted"
  type        = bool
}
