# terraform/modules/kubernetes_cluster_rackspace/outputs.tf

output "kubeconfig_raw" {
  description = "Raw kubeconfig content for the cluster."
  value       = data.spot_kubeconfig.cluster_kubeconfig.raw
  sensitive   = true
}
output "cluster_name_actual" {
  description = "Name of the created Kubernetes cluster."
  value       = spot_cloudspace.cluster.cloudspace_name
}
output "cluster_endpoint_actual" {
  description = "API endpoint for the Kubernetes cluster."
  value       = data.spot_kubeconfig.cluster_kubeconfig.kubeconfigs[0].host
  sensitive   = true
}