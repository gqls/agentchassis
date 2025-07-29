terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.36.0"
    }
  }

  backend "kubernetes" {
    secret_suffix = "tfstate-db-schemas"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

provider "kubernetes" {
  config_path    = "/home/ant/.kube/config_production_uk001"
}

# Read the outputs from the dev database creation layer (for hostnames, usernames, etc)
data "terraform_remote_state" "databases_dev" {
  backend = "kubernetes"
  config = {
    secret_suffix = "tfstate-databases"
    config_path   = "/home/ant/.kube/config_production_uk001"
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

data "local_file" "content_creator_agent_definition_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/contentcreator/001_clientsdb_agent_definition_with_memory.sql"
}

# Create a ConfigMap with all SQL files
resource "kubernetes_config_map" "sql_migrations" {
  metadata {
    name      = "sql-migrations-${substr(sha1(timestamp()), 0, 8)}"
    namespace = "ai-persona-system"
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
    namespace = "ai-persona-system"
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
            name = "PGPASSWORD"
            value_from {
              secret_key_ref {
                name = "postgres-passwords"
                key  = "clients-password"
              }
            }
          }

          command = ["/bin/sh", "-c"]
          args = [
            <<-EOT
            set -e
            echo "Applying migrations to clients database..."

            # Apply pgvector extension
            psql -h postgres-clients -U clients_user -d clients_db -f /migrations/001_enable_pgvector.sql

            # Apply client schema but stop before the template section
            sed '/-- This should be run for each new client/,$d' /migrations/003_create_client_schema.sql > /tmp/clean_client_schema.sql
            psql -h postgres-clients -U clients_user -d clients_db -f /tmp/clean_client_schema.sql

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
            name = "PGPASSWORD"
            value_from {
              secret_key_ref {
                name = "postgres-passwords"
                key  = "templates-password"
              }
            }
          }

          command = ["/bin/sh", "-c"]
          args = [
            <<-EOT
            set -e
            echo "Applying migrations to templates database..."

            # Apply templates schema
            psql -h postgres-templates -U templates_user -d templates_db -f /migrations/002_create_templates_schema.sql

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
    namespace = "ai-persona-system"
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
            name = "MYSQL_PWD"
            value_from {
              secret_key_ref {
                name = "mysql-password"
                key  = "password"
              }
            }
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

            # Check if schema is already applied
            if mysql -h ${data.terraform_remote_state.databases_dev.outputs.external_mysql_host} \
              -u ${data.terraform_remote_state.databases_dev.outputs.external_mysql_user} \
              ${data.terraform_remote_state.databases_dev.outputs.external_mysql_database} \
              -e "SHOW TABLES LIKE 'users';" | grep -q users; then
              echo "Auth schema already exists, skipping..."
            else
              echo "Applying auth schema..."
              mysql -h ${data.terraform_remote_state.databases_dev.outputs.external_mysql_host} \
                -u ${data.terraform_remote_state.databases_dev.outputs.external_mysql_user} \
                ${data.terraform_remote_state.databases_dev.outputs.external_mysql_database} < /tmp/004_auth_schema.sql
            fi

            # Check if projects table exists
            if mysql -h ${data.terraform_remote_state.databases_dev.outputs.external_mysql_host} \
              -u ${data.terraform_remote_state.databases_dev.outputs.external_mysql_user} \
              ${data.terraform_remote_state.databases_dev.outputs.external_mysql_database} \
              -e "SHOW TABLES LIKE 'projects';" | grep -q projects; then
              echo "Projects schema already exists, skipping..."
            else
              echo "Applying projects schema..."
              mysql -h ${data.terraform_remote_state.databases_dev.outputs.external_mysql_host} \
                -u ${data.terraform_remote_state.databases_dev.outputs.external_mysql_user} \
                ${data.terraform_remote_state.databases_dev.outputs.external_mysql_database} < /tmp/005_projects_schema.sql
            fi

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