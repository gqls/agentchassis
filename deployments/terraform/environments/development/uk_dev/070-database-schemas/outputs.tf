output "pgvector_extension_status" {
  description = "Status of the pgvector extension application on the dev clients database."
  value       = "Applied via Kubernetes Job"
  depends_on  = [kubernetes_job.postgres_migrations]
}

output "mysql_schema_status" {
  description = "Status of the MySQL schema application."
  value       = "Applied via Kubernetes Job"
  depends_on  = [kubernetes_job.mysql_migrations]
}

output "postgres_migration_job" {
  description = "Name of the PostgreSQL migration job"
  value       = kubernetes_job.postgres_migrations.metadata[0].name
}

output "mysql_migration_job" {
  description = "Name of the MySQL migration job"
  value       = kubernetes_job.mysql_migrations.metadata[0].name
}