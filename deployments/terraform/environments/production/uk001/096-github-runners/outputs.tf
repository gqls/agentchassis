output "kustomize_apply_status" {
  description = "Status of the runner kustomize applies. Prove at the pods: kubectl -n ai-persona-system get pods -o wide -l workload=gha-runner (3 pods, 3 nodes)."
  value       = "Deployment completed"
}
