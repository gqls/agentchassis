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

variable "kube_context_name" {
  description = "Kubernetes context name"
  type        = string
  default     = "kind-personae-dev"
}

