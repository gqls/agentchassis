variable "kustomize_path" {
  description = "The path to the Kustomize overlay to apply."
  type        = string
}

variable "service_name" {
  description = "The name of the Kubernetes deployment resource."
  type        = string
}

variable "namespace" {
  description = "The Kubernetes namespace to deploy into."
  type        = string
}

variable "image_tag" {
  description = "The Docker image tag to apply to the deployment."
  type        = string
  default     = "latest"
}

variable "image_repository" {
  description = "The Docker image repository (e.g., 'aqls/personae-auth-service')."
  type        = string
}

variable "config_sha" {
  description = "A hash of the service's config file to trigger updates."
  type        = string
  default     = ""
}

variable "deployment_name" {
  description = "The name of the deployment to update. If empty, image update will be skipped."
  type        = string
  default     = ""
}
