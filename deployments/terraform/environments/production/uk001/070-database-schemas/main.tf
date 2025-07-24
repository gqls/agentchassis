terraform {
  required_providers {
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "~> 1.20.0"
    }
    # We don't need a MySQL provider if we use the job runner for it.
  }

  backend "kubernetes" {
    secret_suffix = "tfstate-db-schemas"
    config_path   = "~/.kube/config"
  }
}

# Read the outputs from the database creation layer
data "terraform_remote_state" "databases" {
  backend = "kubernetes"
  config = {
    secret_suffix = "tfstate-databases"
    config_path   = "~/.kube/config"
  }
}

# --- PostgreSQL Provider for the 'templates' database ---
provider "postgresql" {
  alias    = "templates_db_provider"
  host     = data.terraform_remote_state.databases.outputs.postgres_templates_db_endpoint
  port     = 5432
  database = "templates_db"
  username = "templates_user"
  password = data.terraform_remote_state.databases.outputs.templates_db_password
  sslmode  = "disable" # Change to "require" if you configure SSL
}

# --- PostgreSQL Provider for the 'clients' database ---
provider "postgresql" {
  alias    = "clients_db_provider"
  host     = data.terraform_remote_state.databases.outputs.postgres_clients_db_endpoint
  port     = 5432
  database = "clientsdb"
  username = "clients_user"
  password = data.terraform_remote_state.databases.outputs.clients_db_password
  sslmode  = "disable" # Change to "require" if you configure SSL
}

# Read the SQL file for the pgvector extension
data "local_file" "pgvector_sql" {
  filename = "${path.module}/../../../../sql/001_enable_pgvector.sql" # Assuming this path
}

# Apply the pgvector extension to the clients database
resource "postgresql_query" "pgvector_extension" {
  provider = postgresql.clients_db_provider
  query    = data.local_file.pgvector_sql.content
}