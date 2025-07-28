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
    # Service endpoints
    kafka_brokers     = "personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
    clients_db_host   = "postgres-clients.ai-persona-system.svc.cluster.local"
    templates_db_host = "postgres-templates.ai-persona-system.svc.cluster.local"
    auth_db_host      = "rs17.uk-noc.com"
    core_manager_url  = "http://core-manager.ai-persona-system.svc.cluster.local:8088"
    auth_service_url  = "http://auth-service.ai-persona-system.svc.cluster.local:8081"

    # Environment settings
    environment      = "production"
    region          = "uk001"
    log_level       = "info"
    go_env          = "production"
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
    auth-db-password      = var.auth_db_user_password
    templates-db-password = var.templates_db_user_password
    clients-db-password   = var.clients_db_user_password

    # JWT key for auth-service
    jwt-secret = var.jwt_secret_key

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
    anthropic-api-key = var.default_anthropic_key
    stability-api-key = var.default_stability_key
    serp-api-key      = var.default_serp_key
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