# MySQL outputs
output "external_mysql_host" {
  value = var.external_mysql_host
  description = "MySQL host endpoint"
}

output "external_mysql_user" {
  value = var.external_mysql_user
  description = "MySQL username"
}

output "external_mysql_database" {
  value = var.external_mysql_database
  description = "MySQL database name"
}

output "external_mysql_password" {
  value = var.external_mysql_password
  sensitive = true
  description = "MySQL password"
}

# PostgreSQL outputs
output "postgres_clients_db_dev_service_endpoint" {
  value = module.postgres_clients_db_dev.service_endpoint
  description = "PostgreSQL clients database endpoint"
}

output "clients_db_password" {
  value = var.clients_db_password
  sensitive = true
  description = "PostgreSQL clients database password"
}

output "postgres_templates_db_dev_service_endpoint" {
  value = module.postgres_templates_db_dev.service_endpoint
  description = "PostgreSQL templates database endpoint"
}

output "templates_db_password" {
  value = var.templates_db_password
  sensitive = true
  description = "PostgreSQL templates database password"
}