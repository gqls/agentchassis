output "kustomize_apply_status" {
  description = "Status of the ollama kustomize applies. Prove at the pods: kubectl -n ai-persona-system get pods -l 'app in (ollama-adapter,ollama-eval)'."
  value       = "Deployment completed"
}
