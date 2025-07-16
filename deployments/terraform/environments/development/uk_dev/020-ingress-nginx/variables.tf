
variable "kube_context_name" {
  description = "The Kubernetes context name for Kind."
  type        = string
  default     = "kind-personae-dev"
}

variable "kubeconfig_path" {
  description = "Optional path to kubeconfig YAML file."
  type        = string
  default     = null # If null, a default single-node cluster is created
}

variable "ingress_nginx_dev_namespace" { // Renamed for clarity and consistency
  description = "Namespace for Nginx Ingress in dev."
  type        = string
  default     = "ingress-nginx"
}

variable "ingress_nginx_dev_chart_version" { // Renamed
  description = "Helm chart version for Nginx Ingress in dev."
  type        = string
  default     = "4.10.1"
}

variable "ingress_nginx_dev_http_node_port" {
  description = "NodePort for HTTP for Nginx Ingress in dev."
  type        = number
  default     = 30080
}

variable "ingress_nginx_dev_https_node_port" {
  description = "NodePort for HTTPS for Nginx Ingress in dev."
  type        = number
  default     = 30443
}

