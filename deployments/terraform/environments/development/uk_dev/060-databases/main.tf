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

# Create the namespace first
resource "kubernetes_namespace" "db_namespace" {
  metadata {
    name = var.k8s_namespace
  }
}

# --- External MySQL Secret ---
module "external_mysql_auth_db_dev" {
  source = "../../../../modules/mysql-instance"

  instance_name = "personae-dev-mysql"
  namespace     = var.k8s_namespace
  db_host       = var.external_mysql_host
  db_port       = "3306"
  database_name = var.external_mysql_database
  database_user = var.external_mysql_user
  database_pass = var.external_mysql_password

  depends_on = [kubernetes_namespace.db_namespace]
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

  depends_on = [kubernetes_namespace.db_namespace]
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

  depends_on = [kubernetes_namespace.db_namespace]
}

# Create secrets for database passwords
resource "kubernetes_secret" "postgres_passwords" {
  metadata {
    name      = "postgres-passwords"
    namespace = var.k8s_namespace
  }

  data = {
    clients-password   = var.clients_db_password
    templates-password = var.templates_db_password
  }

  depends_on = [kubernetes_namespace.db_namespace]
}

resource "kubernetes_secret" "mysql_password" {
  metadata {
    name      = "mysql-password"
    namespace = var.k8s_namespace
  }

  data = {
    password = var.external_mysql_password
  }

  depends_on = [kubernetes_namespace.db_namespace]
}