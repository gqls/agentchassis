# deployments/terraform/environments/production/uk001/services/agents/2270-kafka-scheduler/variables.tf

variable "image_tag" {
  description = "Docker image tag for kafka-scheduler"
  type        = string
  default     = "v1.0.1"
}

variable "replicas" {
  description = "Number of replicas"
  type        = number
  default     = 1
}

variable "kustomize_path" {
  description = "Path to kustomize overlay"
  type        = string
  default     = "../../../../../deployments/kustomize/services/kafka-scheduler/overlays/production/uk_001"
}
