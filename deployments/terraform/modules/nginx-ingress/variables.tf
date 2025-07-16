variable "ingress_namespace" {
  description = "Namespace to deploy the NGINX Ingress controller into."
  type        = string
  default     = "ingress-nginx"
}

variable "helm_chart_version" {
  description = "Version of the ingress-nginx Helm chart to deploy."
  type        = string
  default     = "4.10.1" # Use a specific, known-good version
}

variable "helm_values_content" {
  description = "YAML content string for Helm values. Pass using file() function from root module."
  type        = string
  default     = ""
}
