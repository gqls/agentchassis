# --- External MySQL Variables ---
variable "external_mysql_host" {
  description = "The endpoint for the external MySQL database."
  type        = string
  default     = "rs17.uk-noc.com"
}

variable "external_mysql_user" {
  description = "Username for the external MySQL database."
  type        = string
  default     = "catalogu_agent-chassis"
}

variable "external_mysql_database" {
  description = "Database name for production."
  type        = string
  default     = "catalogu_vectordb"  # Note: different from dev
}

variable "external_mysql_password" {
  description = "Password for the external MySQL database."
  type        = string
  sensitive   = true
}