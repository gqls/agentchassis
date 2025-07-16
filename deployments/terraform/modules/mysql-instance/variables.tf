variable "instance_name" {
  description = "A logical name for this database instance (used for naming the secret)."
  type        = string
}

variable "namespace" {
  description = "The Kubernetes namespace to create the secret in."
  type        = string
}

variable "db_host" {
  description = "The hostname or IP address of the external MySQL database."
  type        = string
}

variable "db_port" {
  description = "The port number of the external MySQL database."
  type        = string
  default     = "3306"
}

variable "database_name" {
  description = "The name of the database to connect to."
  type        = string
}

variable "database_user" {
  description = "The username for the external database."
  type        = string
}

variable "database_pass" {
  description = "The password for the external database user."
  type        = string
  sensitive   = true
}