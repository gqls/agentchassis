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
variable "default_anthropic_api_key" {
  description = "Default Anthropic API key"
  type        = string
  sensitive   = true
}

variable "default_stability_api_key" {
  description = "Default Stability API key"
  type        = string
  sensitive   = true
}

variable "default_banana_api_key" {
  description = "Default Gemini Banana API key"
  type        = string
  sensitive   = true
}

variable "default_gemini_content_api_key" {
  description = "Default Gemini Content API key"
  type        = string
  sensitive   = true
}

variable "default_serp_api_key" {
  description = "Default SERP API key"
  type        = string
  sensitive   = true
}

variable "default_scraping_bee_api_key" {
  description = "Default SCRAPING BEE API key"
  type        = string
  sensitive   = true
}

variable "default_firecrawl_api_key" {
  description = "Default FIRECRAWL API key"
  type        = string
  sensitive   = true
}

variable "default_perplexity_api_key" {
  description = "Default PERPLEXITY API key"
  type        = string
  sensitive   = true
}

variable "default_companies_house_api_key" {
  description = "Default COMPANIES HOUSE API key"
  type        = string
  sensitive   = true
}

variable "default_thunder_compute_api_key" {
  description = "Default THUNDER_COMPUTE_API_KEY key"
  type        = string
  sensitive   = true
}

variable "default_grok_api_key" {
  description = "Default GROK_API_KEY key"
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

variable "github_personal_access_token" {
    description = "Github personal access token"
    type = string
    sensitive = true
}

variable "github_read_token" {
    description = " Github personal access token for doc analyser READ ONLY"
    type = string
    sensitive = true
}

variable "auth_db_host" {
  description = "The endpoint for the external MySQL database."
  type        = string
  default     = "rs17.uk-noc.com"
}

variable "auth_db_name" {
  description = "The name of the external MySQL database."
  type        = string
  default     = "catalogu_vectordb_chassis"
}

variable "auth_db_user" {
  description = "Username for the external MySQL database."
  type        = string
  default     = "catalogu_personae"
}

variable "agent_image_tag" {
  description = "Docker image tag for agent chassis"
  type        = string
  default     = "v1.0.44"
}