# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/010-infrastructure/outputs.tf
output "kubeconfig_raw" { // Used by Makefile
  description = "Raw Kubeconfig for the production cluster."
  value       = module.kubernetes_cluster.kubeconfig_raw
  sensitive   = true
}

output "cluster_endpoint" {
  description = "API Endpoint for the production cluster."
  value       = module.kubernetes_cluster.cluster_endpoint_actual
  sensitive   = true
}

output "cluster_name" {
  description = "Actual name of the created cluster."
  value       = module.kubernetes_cluster.cluster_name_actual
}