
variable "namespace" {
  description = "The namespace for the Kafka topics job."
  type        = string
  default     = "kafka"
}

variable "default_partitions" {
  description = "Default number of partitions for topics"
  type        = number
  default     = 3
}

variable "default_replication_factor" {
  description = "Default replication factor for topics"
  type        = number
  default     = 3
}