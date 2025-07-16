output "kubeconfig_raw" {
  description = "Raw kubeconfig content for the cluster."
  value       = data.spot_kubeconfig.cluster_kubeconfig.raw
  sensitive   = true
}

output "cluster_name" {
  description = "Name of the created Kubernetes cluster."
  value       = spot_cloudspace.cluster.cloudspace_name
}

output "cluster_endpoint_actual" {
  description = "API endpoint for the Kubernetes cluster."
  value       = data.spot_kubeconfig.cluster_kubeconfig.kubeconfigs[0].host
  sensitive   = true
}