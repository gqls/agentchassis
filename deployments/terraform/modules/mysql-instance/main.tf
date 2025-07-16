terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.20"
    }
  }
}

resource "kubernetes_secret" "external_mysql_secret" {
  metadata {
    name      = "${var.instance_name}-secret"
    namespace = var.namespace
    labels = {
      app  = var.instance_name
      type = "external-db"
    }
  }

  # Note: The keys here (e.g., DB_HOST) must match what your application expects
  # to read from the environment.
  data = {
    "DB_HOST"     = var.db_host
    "DB_PORT"     = var.db_port
    "DB_USER"     = var.database_user
    "DB_PASSWORD" = var.database_pass
    "DB_NAME"     = var.database_name
  }

  type = "Opaque"
}