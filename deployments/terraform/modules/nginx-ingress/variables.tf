# terraform/modules/nginx_ingress/variables.tf

variable "ingress_namespace" {
  description = "Namespace to deploy the NGINX Ingress controller into."
  type        = string
  default     = "ingress-nginx"
}

variable "helm_chart_version" {
  description = "Version of the ingress-nginx Helm chart to deploy."
  type        = string
  default     = "4.10.1" # Example, use a known good/recent version
}

variable "helm_values_content" {
  description = "YAML content string for Helm values. Pass using file() function from root module."
  type        = string
  default     = "" # Empty by default, meaning chart defaults unless provided
}

variable "create_namespace" {
  description = "Whether the module should create the namespace for the ingress controller."
  type        = bool
  default     = true
}