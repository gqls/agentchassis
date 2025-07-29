# variables.tf
variable "monitoring_namespace" {
  description = "The Kubernetes namespace to deploy the monitoring stack into."
  type        = string
  default     = "monitoring"
}

variable "grafana_admin_password" {
  description = "The admin password for the Grafana dashboard."
  type        = string
  sensitive   = true
  default     = ""
}

variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "/home/ant/.kube/config_production_uk001"
}

variable "storage_class" {
  description = "Storage class for persistent volumes"
  type        = string
  default     = "ssd-large"
}
