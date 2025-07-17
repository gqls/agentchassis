output "grafana_service_name" {
  description = "The name of the Grafana service."
  value       = "${helm_release.prometheus_stack.name}-grafana"
}

output "prometheus_service_name" {
  description = "The name of the Prometheus service."
  value       = "${helm_release.prometheus_stack.name}-prometheus"
}

output "alertmanager_service_name" {
  description = "The name of the Alertmanager service."
  value       = "${helm_release.prometheus_stack.name}-alertmanager"
}