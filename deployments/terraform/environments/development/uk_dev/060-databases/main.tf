terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-databases-dev"
    config_path   = "~/.kube/config"
  }
}

# --- External MySQL Secret ---
module "external_mysql_auth_db_dev" {
  source = "../../../../modules/mysql-instance"

  instance_name = "personae-dev-uk-auth-db"
  namespace     = var.k8s_namespace
  db_host       = var.external_mysql_host
  database_name = "authservicedb_dev"
  database_user = "auth_user_dev"
  database_pass = var.external_mysql_password
}

# --- In-Cluster PostgreSQL for Templates ---
module "postgres_templates_db_dev" {
  source = "../../../../modules/postgres-instance"

  instance_name      = "postgres-templates-dev"
  namespace          = var.k8s_namespace
  database_name      = "templatesdb_dev"
  database_user      = "templates_user_dev"
  database_pass      = var.templates_db_password
  storage_class_name = var.postgres_storage_class
  storage_size       = "2Gi" # Smaller size for dev
}

# --- In-Cluster PostgreSQL for Client Data ---
module "postgres_clients_db_dev" {
  source = "../../../../modules/postgres-instance"

  instance_name      = "postgres-clients-dev"
  namespace          = var.k8s_namespace
  database_name      = "clientsdb_dev"
  database_user      = "clients_user_dev"
  database_pass      = var.clients_db_password
  storage_class_name = var.postgres_storage_class
  storage_size       = "5Gi" # Smaller size for dev
}