# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/020-ingress-nginx/variables.tf
variable "ingress_helm_chart_version_override" {
  description = "Specific NGINX Ingress Helm chart version for Sydney (optional)."
  type        = string
  default     = null # Module will use its default if this is null
}

variable "ingress_custom_values_yaml_path" {
  description = "Path to a custom Helm values YAML file for NGINX Ingress for Sydney."
  type        = string
  # Path from this TF config to the actual YAML file in your modules directory
  default     = "../../../../modules/nginx-ingress/config/ingress-nginx-values.yaml" # ADJUST THIS PATH
}

variable "ingress_target_namespace" {
  description = "Target namespace for NGINX Ingress in Sydney."
  type        = string
  default     = "ingress-nginx"
}

variable "kubeconfig_path" {
  description = "Path to the kubeconfig file for the target Kubernetes cluster. This is typically set by the Makefile."
  type        = string
  sensitive   = true
  # No default is needed as the Makefile will pass it.
}