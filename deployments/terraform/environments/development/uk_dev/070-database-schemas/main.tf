terraform {
  required_providers {
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "~> 1.20.0"
    }
    mysql = {
      source  = "drarko/mysql"
      version = "2.0.0"
    }
    null = {
      source = "hashicorp/null"
      version = "~> 3.2.1"
    }
  }

  backend "kubernetes" {
    secret_suffix = "tfstate-db-schemas-dev"
    config_path   = "~/.kube/config"
  }
}

# Read the outputs from the dev database creation layer
data "terraform_remote_state" "databases_dev" {
  backend = "kubernetes"
  config = {
    secret_suffix = "tfstate-databases-dev"
    config_path   = "~/.kube/config"
  }
}

# --- MySQL Provider ---
provider "mysql" {
  endpoint = "${data.terraform_remote_state.databases_dev.outputs.external_mysql_host}:3306"
  username = "auth_user_dev"
  password = data.terraform_remote_state.databases_dev.outputs.external_mysql_password
}

# --- Apply PostgreSQL Schemas ---
# Read the SQL file for the pgvector extension
data "local_file" "pgvector_sql" {
  filename = "${path.module}/../../../../sql/001_enable_pgvector.sql"
}

# Apply the pgvector schema using a local psql command
resource "null_resource" "pgvector_extension_dev" {
  # Trigger a re-run if the SQL file content changes
  triggers = {
    content_sha1 = sha1(data.local_file.pgvector_sql.content)
  }

  provisioner "local-exec" {
    # This command uses the psql client to run the schema.
    # It securely passes the password via the PGPASSWORD environment variable.
    command = "psql -h ${data.terraform_remote_state.databases_dev.outputs.postgres_clients_db_dev_service_endpoint} -U clients_user_dev -d clientsdb_dev -f ${data.local_file.pgvector_sql.filename}"

    environment = {
      PGPASSWORD = data.terraform_remote_state.databases_dev.outputs.clients_db_password
    }
  }
}

# --- Apply MySQL Schema ---
# Read the SQL file for the auth service database schema
data "local_file" "auth_db_schema" {
  filename = "${path.module}/../../../../sql/auth_schema.sql"
}

# Apply the schema to the external dev database using the mysql_script resource
resource "mysql_script" "auth_db_schema_apply" {
  database = "authservicedb_dev"
  script_path = data.local_file.auth_db_schema.filename

  # This ensures the database is created if it doesn't exist before running the script
  depends_on = [
    resource.mysql_database.auth_db_from_schema
  ]
}

resource "mysql_database" "auth_db_from_schema" {
  name = "authservicedb_dev"
  default_character_set = "utf8mb4"
  default_collation     = "utf8mb4_unicode_ci"
}