terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-platform-configs"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

provider "kubernetes" {
  config_context = "personae-uk001-prod-agent-chassis-cluster"
  config_path   = "/home/ant/.kube/config_production_uk001"
}

# Reference existing namespace
data "kubernetes_namespace" "ai_persona_system" {
  metadata {
    name = "ai-persona-system"
  }
}

# Platform-wide configuration
resource "kubernetes_config_map" "personae_prod_config" {
  metadata {
    name      = "personae-prod-config"
    namespace = data.kubernetes_namespace.ai_persona_system.metadata[0].name
  }

  data = {
    # Kafka
    KAFKA_BROKERS: "personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
    kafka_brokers: "personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

    # Clients Database
    CLIENTS_DB_HOST: "postgres-clients.ai-persona-system.svc.cluster.local"
    CLIENTS_DB_PORT: "5432"
    CLIENTS_DB_NAME: "clients_db"
    CLIENTS_DB_USER: "clients_user"
    clients_db_host: "postgres-clients.ai-persona-system.svc.cluster.local"
    clients_db_port: "5432"
    clients_db_name: "clients_db"
    clients_db_user: "clients_user"

    # Templates Database
    TEMPLATES_DB_HOST: "postgres-templates.ai-persona-system.svc.cluster.local"
    TEMPLATES_DB_PORT: "5432"
    TEMPLATES_DB_NAME: "templates_db"
    TEMPLATES_DB_USER: "templates_user"
    templates_db_host: "postgres-templates.ai-persona-system.svc.cluster.local"
    templates_db_port: "5432"
    templates_db_name: "templates_db"
    templates_db_user: "templates_user"

    # Auth Database
    AUTH_DB_HOST: var.auth_db_host
    AUTH_DB_PORT: "3306"
    AUTH_DB_NAME: var.auth_db_name
    AUTH_DB_USER: var.auth_db_user
    auth_db_host: var.auth_db_host
    auth_db_port: "3306"
    auth_db_name: var.auth_db_name
    auth_db_user: var.auth_db_user

    # Service URLs
    auth_service_url: "http://auth-service.ai-persona-system.svc.cluster.local:8081"
    core_manager_url: "http://core-manager.ai-persona-system.svc.cluster.local:8088"

    # Environment settings
    environment: "production"
    region: "uk001"
    go_env: "production"
    log_level: "info"
    tracing_enabled: "false"
    tracing_endpoint: "otel-collector.monitoring.svc.cluster.local:4317"
  }
}

# Platform service secrets (shared by all platform services)
resource "kubernetes_secret" "personae_platform_secrets" {
  metadata {
    name      = "personae-platform-secrets"
    namespace = data.kubernetes_namespace.ai_persona_system.metadata[0].name
  }

  data = {
    # Database passwords for platform services
    AUTH_DB_PASSWORD      = var.auth_db_user_password
    TEMPLATES_DB_PASSWORD = var.templates_db_user_password
    CLIENTS_DB_PASSWORD   = var.clients_db_user_password

    # JWT key for auth-service
    JWT_SECRET_KEY = var.jwt_secret_key

    # Platform agent bootstrap key
    agent-bootstrap-key = var.agent_bootstrap_key
  }
}

# Default API keys (will be replaced by per-user keys later)
resource "kubernetes_secret" "personae_default_api_keys" {
  metadata {
    name      = "personae-default-secrets"
    namespace = data.kubernetes_namespace.ai_persona_system.metadata[0].name
  }

  data = {
    ANTHROPIC_API_KEY = var.default_anthropic_api_key
    STABILITY_API_KEY = var.default_stability_api_key
    SERP_API_KEY      = var.default_serp_api_key
    SCRAPING_BEE_API_KEY = var.default_scraping_bee_api_key
    FIRECRAWL_API_KEY =  var.default_firecrawl_api_key
  }
}

# Docker registry credentials
resource "kubernetes_secret" "docker_hub_creds" {
  metadata {
    name      = "docker-hub-creds"
    namespace = data.kubernetes_namespace.ai_persona_system.metadata[0].name
  }

  type = "kubernetes.io/dockerconfigjson"

  data = {
    ".dockerconfigjson" = jsonencode({
      auths = {
        "docker.io" = {
          username = "aqls"
          password = var.docker_password
          email    = var.docker_email
          auth     = base64encode("aqls:${var.docker_password}")
        }
      }
    })
  }
}