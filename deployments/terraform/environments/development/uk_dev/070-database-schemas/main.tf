terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.20"
    }
  }

  backend "kubernetes" {
    secret_suffix = "tfstate-db-schemas-dev"
    config_path   = "~/.kube/config"
  }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "kind-personae-dev"
}

# Read the outputs from the dev database creation layer
data "terraform_remote_state" "databases_dev" {
  backend = "kubernetes"
  config = {
    secret_suffix = "tfstate-databases-dev"
    config_path   = "~/.kube/config"
  }
}

# Read all SQL files
data "local_file" "pgvector_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/001_enable_pgvector.sql"
}

data "local_file" "templates_schema_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/002_create_templates_schema.sql"
}

data "local_file" "client_schema_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/003_create_client_schema.sql"
}

data "local_file" "auth_db_schema" {
  filename = "${path.module}/../../../../../../platform/database/migrations/004_auth_schema.sql"
}

data "local_file" "projects_schema_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/005_projects_schema.sql"
}

# Create a ConfigMap with all SQL files
resource "kubernetes_config_map" "sql_migrations" {
  metadata {
    name      = "sql-migrations-${substr(sha1(timestamp()), 0, 8)}"
    namespace = "personae-dev-db"
  }

  data = {
    "001_enable_pgvector.sql"       = data.local_file.pgvector_sql.content
    "002_create_templates_schema.sql" = data.local_file.templates_schema_sql.content
    "003_create_client_schema.sql"   = data.local_file.client_schema_sql.content
    "004_auth_schema.sql"           = data.local_file.auth_db_schema.content
    "005_projects_schema.sql"       = data.local_file.projects_schema_sql.content
  }
}

# Job to run PostgreSQL migrations
resource "kubernetes_job" "postgres_migrations" {
  metadata {
    name      = "postgres-migrations-${substr(sha1(timestamp()), 0, 8)}"
    namespace = "personae-dev-db"
  }

  spec {
    template {
      metadata {
        labels = {
          app = "postgres-migrations"
        }
      }

      spec {
        restart_policy = "Never"

        volume {
          name = "sql-scripts"
          config_map {
            name = kubernetes_config_map.sql_migrations.metadata[0].name
          }
        }

        # Container to run pgvector and client schema on clients database
        container {
          name  = "migrate-clients"
          image = "postgres:16-alpine"

          volume_mount {
            name       = "sql-scripts"
            mount_path = "/migrations"
          }

          env {
            name  = "PGPASSWORD"
            value = data.terraform_remote_state.databases_dev.outputs.clients_db_password
          }

          command = ["/bin/sh", "-c"]
          args = [
            <<-EOT
            set -e
            echo "Applying migrations to clients database..."

            # Apply pgvector extension
            psql -h postgres-clients-dev -U clients_user_dev -d clientsdb_dev -f /migrations/001_enable_pgvector.sql

            # Apply client schema
            psql -h postgres-clients-dev -U clients_user_dev -d clientsdb_dev -f /migrations/003_create_client_schema.sql

            echo "Clients database migrations completed!"
            EOT
          ]
        }

        # Container to run templates schema on templates database
        container {
          name  = "migrate-templates"
          image = "postgres:16-alpine"

          volume_mount {
            name       = "sql-scripts"
            mount_path = "/migrations"
          }

          env {
            name  = "PGPASSWORD"
            value = data.terraform_remote_state.databases_dev.outputs.templates_db_password
          }

          command = ["/bin/sh", "-c"]
          args = [
            <<-EOT
            set -e
            echo "Applying migrations to templates database..."

            # Apply templates schema
            psql -h postgres-templates-dev -U templates_user_dev -d templatesdb_dev -f /migrations/002_create_templates_schema.sql

            echo "Templates database migrations completed!"
            EOT
          ]
        }
      }
    }

    backoff_limit = 3
  }

  wait_for_completion = true
  timeouts {
    create = "10m"
  }
}

# Job to run MySQL migrations
resource "kubernetes_job" "mysql_migrations" {
  metadata {
    name      = "mysql-migrations-${substr(sha1(timestamp()), 0, 8)}"
    namespace = "personae-dev-db"
  }

  spec {
    template {
      metadata {
        labels = {
          app = "mysql-migrations"
        }
      }

      spec {
        restart_policy = "Never"

        volume {
          name = "sql-scripts"
          config_map {
            name = kubernetes_config_map.sql_migrations.metadata[0].name
          }
        }

        container {
          name  = "migrate-mysql"
          image = "mysql:8.0"

          volume_mount {
            name       = "sql-scripts"
            mount_path = "/migrations"
          }

          env {
            name  = "MYSQL_PWD"
            value = data.terraform_remote_state.databases_dev.outputs.external_mysql_password
          }

          command = ["/bin/sh", "-c"]
          args = [
            <<-EOT
            set -e
            echo "Applying migrations to MySQL database..."
            echo "Host: ${data.terraform_remote_state.databases_dev.outputs.external_mysql_host}"
            echo "User: ${data.terraform_remote_state.databases_dev.outputs.external_mysql_user}"
            echo "Database: ${data.terraform_remote_state.databases_dev.outputs.external_mysql_database}"

            # Copy files to a writable location
            cp /migrations/*.sql /tmp/

            # Fix SQL syntax issues
            sed -i 's|// FILE:.*||g' /tmp/005_projects_schema.sql

            # Apply auth schema
            mysql -h ${data.terraform_remote_state.databases_dev.outputs.external_mysql_host} \
              -u ${data.terraform_remote_state.databases_dev.outputs.external_mysql_user} \
              ${data.terraform_remote_state.databases_dev.outputs.external_mysql_database} < /tmp/004_auth_schema.sql || {
                echo "Failed to apply auth schema"
                echo "Error details above"
                exit 1
              }

            # Apply projects schema
            mysql -h ${data.terraform_remote_state.databases_dev.outputs.external_mysql_host} \
              -u ${data.terraform_remote_state.databases_dev.outputs.external_mysql_user} \
              ${data.terraform_remote_state.databases_dev.outputs.external_mysql_database} < /tmp/005_projects_schema.sql || {
                echo "Failed to apply projects schema"
                echo "Error details above"
                exit 1
              }

            echo "MySQL migrations completed!"
            EOT
          ]
        }
      }
    }

    backoff_limit = 3
  }

  wait_for_completion = true
  timeouts {
    create = "10m"
  }

  depends_on = [kubernetes_config_map.sql_migrations]
}