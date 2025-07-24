terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-databases"
    config_path   = "~/.kube/config"
  }
}

# --- External MySQL Secret ---
module "external_mysql_auth_db" {
  source = "../../../modules/mysql-instance"

  instance_name = "personae-prod-uk001-auth-db"
  namespace     = var.k8s_namespace
  db_host       = var.external_mysql_host
  database_name = "authservicedb"
  database_user = "auth_user"
  database_pass = var.external_mysql_password
}

# --- In-Cluster PostgreSQL for Templates ---
module "postgres_templates_db" {
  source = "../../../modules/postgres-instance"

  instance_name      = "postgres-templates"
  namespace          = var.k8s_namespace
  database_name      = "templates_db"
  database_user      = "templates_user"
  database_pass      = var.templates_db_password
  storage_class_name = var.postgres_storage_class
  storage_size       = "5Gi"
}

# --- In-Cluster PostgreSQL for Client Data ---
module "postgres_clients_db" {
  source = "../../../modules/postgres-instance"

  instance_name      = "postgres-clients"
  namespace          = var.k8s_namespace
  database_name      = "clientsdb"
  database_user      = "clients_user"
  database_pass      = var.clients_db_password
  storage_class_name = var.postgres_storage_class
  storage_size       = "20Gi"
}