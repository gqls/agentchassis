# FILE: deployments/terraform/environments/production/uk_001/services/agents/2250-content-creator-agent/variables.tf
variable "kubeconfig_path" {
  description = "Path to the Kubernetes kubeconfig file."
  type        = string
  # No default here, ensuring it's always explicitly provided (e.g., via -var-file or TF_VAR_kubeconfig_path)
}

