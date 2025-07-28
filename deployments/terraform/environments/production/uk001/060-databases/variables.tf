variable "k8s_namespace" {
  description = "The Kubernetes namespace to deploy prod database resources into."
  type        = string
  default     = "ai-persona-system"
}

variable "postgres_storage_class" {
  description = "The name of the StorageClass for dev PostgreSQL volumes (e.g., your local-path-provisioner)."
  type        = string
  default     = "ssd-large"  ## or ssd
}

variable "postgres_storage_size" {
  description = "The size of the StorageClass for PostgreSQL volumes 50 - 1024 Gi for ssd-large."
  type        = string
  default     = "100Gi"
}

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
  default     = "catalogu_vectordb_chassis"
}

variable "external_mysql_password" {
  description = "Password for the external MySQL database."
  type        = string
  sensitive   = true
}

variable "external_mysql_port" {
  description = "Database port for production."
  type        = number
  default     = 3306
}

# --- In-Cluster PostgreSQL Variables ---
variable "templates_db_password" {
  description = "Password for the prod templates PostgreSQL database."
  type        = string
  sensitive   = true
}

variable "clients_db_password" {
  description = "Password for the prod clients PostgreSQL database."
  type        = string
  sensitive   = true
}