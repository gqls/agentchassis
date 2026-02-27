# deployments/terraform/environments/production/uk001/services/agents/2220-agent-dispatcher/variables.tf

variable "image_tag" {
  description = "Docker image tag for agent-dispatcher"
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
  default     = "../../../../../deployments/kustomize/services/agent-dispatcher/overlays/production/uk_001"
}
