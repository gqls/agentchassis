variable "instance_name" {
  description = "The unique name for the PostgreSQL StatefulSet and related resources."
  type        = string
}

variable "namespace" {
  description = "The Kubernetes namespace to deploy the resources into."
  type        = string
}

variable "database_name" {
  description = "The name of the database to create."
  type        = string
}

variable "database_user" {
  description = "The username for the database."
  type        = string
}

variable "database_pass" {
  description = "The password for the database user."
  type        = string
  sensitive   = true
}

variable "storage_class_name" {
  description = "The name of the StorageClass to use for the PersistentVolumeClaim."
  type        = string
}

variable "storage_size" {
  description = "The size of the persistent volume (e.g., '10Gi')."
  type        = string
}