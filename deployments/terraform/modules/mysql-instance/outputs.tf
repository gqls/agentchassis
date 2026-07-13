output "secret_name" {
  description = "The name of the secret containing the external database credentials."
  value       = kubernetes_secret.external_mysql_secret.metadata[0].name
}