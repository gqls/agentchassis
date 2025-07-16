output "mysql_auth_db_endpoint" {
  description = "The connection endpoint for the auth service MySQL database."
  value       = module.mysql_auth_db.db_instance_endpoint
}

output "postgres_templates_db_endpoint" {
  description = "The connection endpoint for the templates PostgreSQL database."
  value       = module.postgres_templates_db.db_instance_endpoint
}

output "postgres_clients_db_endpoint" {
  description = "The connection endpoint for the clients PostgreSQL database."
  value       = module.postgres_clients_db.db_instance_endpoint
}