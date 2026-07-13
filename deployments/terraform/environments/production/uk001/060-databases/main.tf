terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-databases"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.36.0"
    }
  }
}

provider "kubernetes" {
  config_path = "/home/ant/.kube/config_production_uk001"
}

data "kubernetes_namespace" "db_namespace" {
  metadata {
    name = var.k8s_namespace
  }
}

# --- External MySQL Secret ---
module "external_mysql_auth_db" {
  source = "../../../../modules/mysql-instance"

  instance_name = "personae-prod-uk001-auth-db"
  namespace     = var.k8s_namespace
  db_host       = var.external_mysql_host
  db_port       = var.external_mysql_port
  database_name = var.external_mysql_database
  database_user = var.external_mysql_user
  database_pass = var.external_mysql_password
}

# --- In-Cluster PostgreSQL for Templates ---
module "postgres_templates_db" {
  source = "../../../../modules/postgres-instance"

  instance_name      = "postgres-templates"
  namespace          = var.k8s_namespace
  database_name      = "templates_db"
  database_user      = "templates_user"
  database_pass      = var.templates_db_password
  storage_class_name = var.postgres_storage_class
  storage_size       = var.postgres_storage_size
}

# --- In-Cluster PostgreSQL for Client Data ---
module "postgres_clients_db" {
  source = "../../../../modules/postgres-instance"

  instance_name      = "postgres-clients"
  namespace          = var.k8s_namespace
  database_name      = "clients_db"
  database_user      = "clients_user"
  database_pass      = var.clients_db_password
  storage_class_name = var.postgres_storage_class
  storage_size       = var.postgres_storage_size
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