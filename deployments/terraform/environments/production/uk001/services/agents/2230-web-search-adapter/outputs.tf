output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the web-search-adapter."
  value       = module.web_search_adapter_deployment.status
}