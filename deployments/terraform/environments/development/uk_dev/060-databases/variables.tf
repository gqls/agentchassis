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
  sensitive   = true
}

variable "external_mysql_password" {
  description = "Password for the external MySQL database."
  type        = string
  sensitive   = true
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