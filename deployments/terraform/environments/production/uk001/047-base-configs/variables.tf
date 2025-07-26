# Database user passwords (not root)
variable "auth_db_user_password" {
  description = "Password for auth service database user"
  type        = string
  sensitive   = true
}

variable "templates_db_user_password" {
  description = "Password for templates database user"
  type        = string
  sensitive   = true
}

variable "clients_db_user_password" {
  description = "Password for clients database user"
  type        = string
  sensitive   = true
}

# Platform keys
variable "jwt_secret_key" {
  description = "JWT signing key for auth-service"
  type        = string
  sensitive   = true
}

variable "agent_bootstrap_key" {
  description = "Bootstrap key for platform agents"
  type        = string
  sensitive   = true
}

# Default API keys (temporary until per-user keys)
variable "default_anthropic_key" {
  description = "Default Anthropic API key"
  type        = string
  sensitive   = true
}

variable "default_stability_key" {
  description = "Default Stability API key"
  type        = string
  sensitive   = true
}

variable "default_serp_key" {
  description = "Default SERP API key"
  type        = string
  sensitive   = true
}

# Docker
variable "docker_password" {
  description = "Docker Hub password"
  type        = string
  sensitive   = true
}

variable "docker_email" {
  description = "Docker Hub email"
  type        = string
}