variable "k8s_namespace" {
  description = "The Kubernetes namespace to deploy dev database resources into."
  type        = string
  default     = "personae-dev-db"
}

variable "postgres_storage_class" {
  description = "The name of the StorageClass for dev PostgreSQL volumes (e.g., your local-path-provisioner)."
  type        = string
  default     = "standard"
}

# --- External MySQL Variables ---
variable "external_mysql_host" {
  description = "The endpoint for the external MySQL database used for development."
  type        = string
  default     = "rs17.uk-noc.com"
}

variable "external_mysql_user" {
  description = "Username for the external MySQL database."
  type        = string
  default     = "catalogu_agent-chassis"
}

variable "external_mysql_database" {
  description = "Database name for development."
  type        = string
  default     = "catalogu_vectordbdev"
}

variable "external_mysql_password" {
  description = "Password for the external MySQL database."
  type        = string
  sensitive   = true
}

variable "external_mysql_port" {
  description = "Database port for development."
  type        = number
  default     = 3306
}

# --- In-Cluster PostgreSQL Variables ---
variable "templates_db_password" {
  description = "Password for the dev templates PostgreSQL database."
  type        = string
  sensitive   = true
}

variable "clients_db_password" {
  description = "Password for the dev clients PostgreSQL database."
  type        = string
  sensitive   = true
}