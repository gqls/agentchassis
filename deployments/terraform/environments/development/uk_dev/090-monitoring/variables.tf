variable "monitoring_namespace" {
  description = "The Kubernetes namespace to deploy the dev monitoring stack into."
  type        = string
  default     = "monitoring-dev"
}

variable "grafana_admin_password" {
  description = "The admin password for the Grafana dashboard."
  type        = string
  sensitive   = true
}