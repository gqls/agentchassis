output "grafana_password" {
  value     = "admin"  # This should be changed for production
  sensitive = true
}

output "monitoring_namespace" {
  value = kubernetes_namespace.monitoring.metadata[0].name
}

output "grafana_service" {
  value = "kube-prometheus-stack-grafana.${kubernetes_namespace.monitoring.metadata[0].name}.svc.cluster.local"
}

output "prometheus_service" {
  value = "kube-prometheus-stack-prometheus.${kubernetes_namespace.monitoring.metadata[0].name}.svc.cluster.local"
}