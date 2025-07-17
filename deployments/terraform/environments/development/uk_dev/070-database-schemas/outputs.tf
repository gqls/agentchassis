output "pgvector_extension_status" {
  description = "Status of the pgvector extension application on the dev clients database."
  value       = "Applied"
  depends_on  = [postgresql_query.pgvector_extension_dev]
}