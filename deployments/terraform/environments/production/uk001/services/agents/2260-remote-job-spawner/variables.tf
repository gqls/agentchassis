# deployments/terraform/environments/production/uk001/services/agents/2220-remote-job-spawner/variables.tf

variable "image_tag" {
  description = "Docker image tag for remote-job-spawner"
  type        = string
  default     = "v1.0.0"
}

variable "replicas" {
  description = "Number of replicas"
  type        = number
  default     = 1
}

variable "kustomize_path" {
  description = "Path to kustomize overlay"
  type        = string
  default     = "../../../../../deployments/kustomize/services/remote-job-spawner/overlays/production/uk_001"
}
