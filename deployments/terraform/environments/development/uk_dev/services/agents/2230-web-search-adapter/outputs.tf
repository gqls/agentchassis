output "kustomize_apply_status" {
  description = "The status of the Kustomize deployment for the development web-search-adapter."
  value       = module.web_search_adapter_deployment_dev.status
}