variable "k8s_namespace" {
  description = "The Kubernetes namespace to deploy database resources into."
  type        = string
  default     = "personae-prod-db"
}

variable "postgres_storage_class" {
  description = "The name of the Kubernetes StorageClass to use for PostgreSQL volumes."
  type        = string
  # Note: Change this to your actual production storage class name.
  default     = "premium-storage"
}

# --- External MySQL Variables ---
variable "external_mysql_host" {
  description = "The endpoint for the external MySQL database."
  type        = string
  sensitive   = true
}

variable "external_mysql_password" {
  description = "Password for the external MySQL database."
  type        = string
  sensitive   = true
}

# --- In-Cluster PostgreSQL Variables ---
variable "templates_db_password" {
  description = "Password for the templates PostgreSQL database."
  type        = string
  sensitive   = true
}

variable "clients_db_password" {
  description = "Password for the multi-tenant clients PostgreSQL database."
  type        = string
  sensitive   = true
}