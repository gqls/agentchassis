output "service_name" {
  description = "The name of the PostgreSQL Kubernetes service."
  value       = kubernetes_service.postgres_service.metadata[0].name
}

output "service_endpoint" {
  description = "The internal DNS endpoint for the service."
  value       = "${kubernetes_service.postgres_service.metadata[0].name}.${var.namespace}.svc.cluster.local"
}

output "secret_name" {
  description = "The name of the secret containing the database credentials."
  value       = kubernetes_secret.postgres_secret.metadata[0].name
}