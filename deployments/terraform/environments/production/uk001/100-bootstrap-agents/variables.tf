# deployments/terraform/environments/production/uk001/100-bootstrap-agents/variables.tf

variable "namespace" {
  description = "Kubernetes namespace for deployment"
  type        = string
  default     = "ai-persona-system"
}

variable "registry" {
  description = "Docker registry for images"
  type        = string
  default     = "docker.io/aqls"
}

variable "image_tag" {
  description = "Docker image tag to deploy"
  type        = string
  default     = "v1.0.47"
}

variable "enable_autoscaling" {
  description = "Enable horizontal pod autoscaling"
  type        = bool
  default     = false  # Bootstrap agent should not autoscale
}

variable "environment" {
  description = "Environment name (production, staging, development)"
  type        = string
  default     = "production"
}

variable "region" {
  description = "Region identifier"
  type        = string
  default     = "uk001"
}