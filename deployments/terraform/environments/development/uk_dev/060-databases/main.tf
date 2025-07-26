terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-databases-dev"
    config_path   = "~/.kube/config"
  }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "kind-personae-dev"
}

# Reference the existing namespace instead of creating it
data "kubernetes_namespace" "db_namespace" {
  metadata {
    name = var.k8s_namespace
  }
}

# --- External MySQL Secret ---
module "external_mysql_auth_db_dev" {
  source = "../../../../modules/mysql-instance"

  instance_name = "personae-dev-mysql"
  namespace     = data.kubernetes_namespace.db_namespace.metadata[0].name
  db_host       = var.external_mysql_host
  db_port       = "3306"
  database_name = var.external_mysql_database
  database_user = var.external_mysql_user
  database_pass = var.external_mysql_password
}

# --- In-Cluster PostgreSQL for Templates ---
module "postgres_templates_db_dev" {
  source = "../../../../modules/postgres-instance"

  instance_name      = "postgres-templates-dev"
  namespace          = data.kubernetes_namespace.db_namespace.metadata[0].name
  database_name      = "templates_db"
  database_user      = "templates_user"
  database_pass      = var.templates_db_password
  storage_class_name = var.postgres_storage_class
  storage_size       = "2Gi" # Smaller size for dev
}

# --- In-Cluster PostgreSQL for Client Data ---
module "postgres_clients_db_dev" {
  source = "../../../../modules/postgres-instance"

  instance_name      = "postgres-clients-dev"
  namespace          = data.kubernetes_namespace.db_namespace.metadata[0].name
  database_name      = "clients_db"
  database_user      = "clients_user"
  database_pass      = var.clients_db_password
  storage_class_name = var.postgres_storage_class
  storage_size       = "5Gi" # Smaller size for dev
}

# Create secrets for database passwords
resource "kubernetes_secret" "postgres_passwords" {
  metadata {
    name      = "postgres-passwords"
    namespace = data.kubernetes_namespace.db_namespace.metadata[0].name
  }

  data = {
    clients-password   = var.clients_db_password
    templates-password = var.templates_db_password
  }
}

resource "kubernetes_secret" "mysql_password" {
  metadata {
    name      = "mysql-password"
    namespace = data.kubernetes_namespace.db_namespace.metadata[0].name
  }

  data = {
    password = var.external_mysql_password
  }
}