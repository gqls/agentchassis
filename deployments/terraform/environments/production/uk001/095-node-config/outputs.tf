output "kustomize_apply_status" {
  description = "Status of the node-config DaemonSet kustomize apply. Prove the result at the kubelets, not here: make node-config-status."
  value       = "Deployment completed"
}
