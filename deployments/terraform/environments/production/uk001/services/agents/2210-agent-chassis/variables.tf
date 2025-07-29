# infrastructure/environments/production/uk001/100-svc-agent-chassis/variables.tf

variable "image_tag" {
  description = "Docker image tag for agent-chassis"
  type        = string
  default     = "v1.0.0"
}

variable "replicas" {
  description = "Number of replicas"
  type        = number
  default     = 3
}

variable "kustomize_path" {
  description = "The path to the Kustomize overlay to apply."
  type        = string
  default     = "../../../../../deployments/kustomize/services/agent-chassis/overlays/production/uk_001"

}