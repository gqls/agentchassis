# terraform/environments/production/uk001/040-kafka-cluster/variables.tf
variable "kubeconfig_path" {
  description = "Path to the kubeconfig file for the uk001 cluster."
  type        = string
  sensitive   = true
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
  default     = "../../../../modules/kafka-cluster/config/kafka-cluster-cr.yaml"
}

variable "kafka_cluster_name_uk001" { // Changed from _sydney
  description = "The metadata.name of the Kafka cluster being deployed in uk001 (must match name in YAML)."
  type        = string
  default     = "personae-kafka-cluster" # Assuming you want to use the same Kafka cluster name internally
}