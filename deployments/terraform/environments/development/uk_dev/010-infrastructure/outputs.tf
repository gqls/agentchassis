output "kind_cluster_name_output" {
  description = "Name of the Kind cluster."
  value       = var.kind_cluster_name
}

output "kind_kube_context_name_output" {
  description = "The kubectl context name for this Kind cluster."
  value       = "kind-${var.kind_cluster_name}" # Standard Kind context naming
}

# No kubeconfig_raw output here as Kind manages the default kubeconfig file.
# Other components will use the context name.