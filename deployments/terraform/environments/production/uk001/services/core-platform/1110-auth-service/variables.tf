# Variables for auth-service deployment
variable "image_tag" {
  description = "Docker image tag for auth-service"
  type        = string
  default     = "latest"
}