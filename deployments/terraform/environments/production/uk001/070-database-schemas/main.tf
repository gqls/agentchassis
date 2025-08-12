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
  config_path = "/home/ant/.kube/config_production_uk001"
}

# Read the outputs from the dev database creation layer
data "terraform_remote_state" "databases_dev" {
  backend = "kubernetes"
  config = {
    secret_suffix = "tfstate-databases"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

# Read all SQL files
data "local_file" "extensions_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/001_enable_extensions.sql"
}

data "local_file" "core_tables_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/002_core_tables.sql"
}

data "local_file" "agent_groups_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/003_agent_groups.sql"
}

data "local_file" "agent_metrics_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/004_agent_metrics.sql"
}

data "local_file" "website_builder_agents_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/005_website_builder_agents.sql"
}

data "local_file" "client_schema_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/006_client_schema.sql"
}

data "local_file" "initial_website_builder_group_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/007_initial_website_builder_group.sql"
}

data "local_file" "add_discovery_functions_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/008_add_discovery_functions.sql"
}

data "local_file" "seed_website_builder_definitions_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/009_seed_website_builder_definitions.sql"
}

data "local_file" "client_schema_additions_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/060_client_schema_additions.sql"
}

data "local_file" "clientsdb_content_creator_agent_definition_with_memory_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/080_clientsdb_content_creator_agent_definition_with_memory.sql"
}

# Auth database migrations
data "local_file" "auth_schema_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/100_auth_schema.sql"
}

data "local_file" "projects_schema_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/101_projects_schema.sql"
}

# Templates database migrations
data "local_file" "create_templates_schema_sql" {
  filename = "${path.module}/../../../../../../platform/database/migrations/200_create_templates_schema.sql"
}

# Create ConfigMap for PostgreSQL migrations
resource "kubernetes_config_map" "postgres_sql_migrations" {
  metadata {
    name      = "postgres-sql-migrations-${substr(sha1(timestamp()), 0, 8)}"
    namespace = "ai-persona-system"
  }

  data = {
    # Core migrations
    "001_enable_extensions.sql"                 = data.local_file.extensions_sql.content
    "002_core_tables.sql"                       = data.local_file.core_tables_sql.content
    "003_agent_groups.sql"                      = data.local_file.agent_groups_sql.content
    "004_agent_metrics.sql"                     = data.local_file.agent_metrics_sql.content
    "005_website_builder_agents.sql"            = data.local_file.website_builder_agents_sql.content
    "006_client_schema.sql"                     = data.local_file.client_schema_sql.content
    "007_initial_website_builder_group.sql"     = data.local_file.initial_website_builder_group_sql.content
    "008_add_discovery_functions.sql"           = data.local_file.add_discovery_functions_sql.content
    "009_seed_website_builder_definitions.sql"  = data.local_file.seed_website_builder_definitions_sql.content
    "060_client_schema_additions.sql"           = data.local_file.client_schema_additions_sql.content
    "080_clientsdb_content_creator_agent_definition_with_memory.sql"        = data.local_file.clientsdb_content_creator_agent_definition_with_memory_sql.content

    # Templates migrations
    "200_create_templates_schema.sql"               = data.local_file.create_templates_schema_sql.content
  }
}

# Create ConfigMap for MySQL migrations
resource "kubernetes_config_map" "mysql_sql_migrations" {
  metadata {
    name      = "mysql-sql-migrations-${substr(sha1(timestamp()), 0, 8)}"
    namespace = "ai-persona-system"
  }

  data = {
    "100_auth_schema.sql"    = data.local_file.auth_schema_sql.content
    "101_projects_schema.sql"  = data.local_file.projects_schema_sql.content
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
            name = kubernetes_config_map.postgres_sql_migrations.metadata[0].name
          }
        }

        # Init container to wait for databases
        init_container {
          name  = "wait-for-postgres"
          image = "postgres:16-alpine"

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
            until pg_isready -h postgres-clients -U clients_user; do
              echo "Waiting for clients database..."
              sleep 2
            done
            until pg_isready -h postgres-templates -U templates_user; do
              echo "Waiting for templates database..."
              sleep 2
            done
            echo "Databases are ready!"
            EOT
          ]
        }

        # Container to run migrations on clients database
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

            export PGHOST=postgres-clients
            export PGUSER=clients_user
            export PGDATABASE=clients_db

            # Apply in numerical order
            psql -f /migrations/001_enable_extensions.sql
            psql -f /migrations/002_core_tables.sql
            psql -f /migrations/003_agent_groups.sql
            psql -f /migrations/004_agent_metrics.sql
            psql -f /migrations/005_website_builder_agents.sql
            psql -f /migrations/006_client_schema.sql
            psql -f /migrations/007_initial_website_builder_group.sql
            psql -f /migrations/008_add_discovery_functions.sql
            psql -f /migrations/009_seed_website_builder_definitions.sql
            psql -f /migrations/060_client_schema_additions.sql
            psql -f /migrations/080_clientsdb_content_creator_agent_definition_with_memory.sql

            echo "Clients database migrations completed!"
            EOT
          ]
        }

        # Container to run templates migrations
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

            # Connect to database
            export PGHOST=postgres-templates
            export PGUSER=templates_user
            export PGDATABASE=templates_db

            # Apply only templates migrations (200+)
            for migration in $(ls /migrations/2*.sql | sort); do
              echo "Applying $migration..."
              psql -f "$migration"
            done

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
            name = kubernetes_config_map.mysql_sql_migrations.metadata[0].name
          }
        }

        # Init container to wait for MySQL
        init_container {
          name  = "wait-for-mysql"
          image = "mysql:8.0"

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
            until mysqladmin ping -h${data.terraform_remote_state.databases_dev.outputs.external_mysql_host} --silent; do
              echo "Waiting for MySQL..."
              sleep 2
            done
            echo "MySQL is ready!"
            EOT
          ]
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

            # Set connection parameters
            MYSQL_HOST="${data.terraform_remote_state.databases_dev.outputs.external_mysql_host}"
            MYSQL_USER="${data.terraform_remote_state.databases_dev.outputs.external_mysql_user}"
            MYSQL_DATABASE="${data.terraform_remote_state.databases_dev.outputs.external_mysql_database}"

            # Apply migrations in order
            for migration in $(ls /migrations/*.sql | sort); do
              echo "Applying $migration..."

              # Extract table name from migration for idempotency check
              TABLE_NAME=$(grep -oP "CREATE TABLE IF NOT EXISTS \K\w+" "$migration" | head -1)

              if [ -n "$TABLE_NAME" ]; then
                # Check if table exists
                if mysql -h"$MYSQL_HOST" -u"$MYSQL_USER" "$MYSQL_DATABASE" \
                  -e "SHOW TABLES LIKE '$TABLE_NAME';" | grep -q "$TABLE_NAME"; then
                  echo "Table $TABLE_NAME already exists, checking for updates..."
                fi
              fi

              mysql -h"$MYSQL_HOST" -u"$MYSQL_USER" "$MYSQL_DATABASE" < "$migration"
            done

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

  depends_on = [kubernetes_config_map.mysql_sql_migrations]
}

# Output the job names for reference
output "postgres_migration_job" {
  value = kubernetes_job.postgres_migrations.metadata[0].name
}

output "mysql_migration_job" {
  value = kubernetes_job.mysql_migrations.metadata[0].name
}