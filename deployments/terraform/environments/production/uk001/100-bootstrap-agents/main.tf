# deployments/terraform/environments/production/uk001/100-bootstrap-agents/main.tf
provider "kubernetes" {
  config_path    = "~/.kube/config_production_uk001"  # Path to your kubeconfig file
}

# Data sources for existing resources
data "kubernetes_namespace" "system" {
  metadata {
    name = var.namespace
  }
}

data "kubernetes_config_map" "prod_config" {
  metadata {
    name      = "personae-prod-config"
    namespace = var.namespace
  }
}

data "kubernetes_secret" "platform_secrets" {
  metadata {
    name      = "personae-platform-secrets"
    namespace = var.namespace
  }
}

resource "kubernetes_config_map" "agent_config" {
  metadata {
    name      = "agent-chassis-config"
    namespace = var.namespace
  }
  data = {
    "agent-chassis.yaml" = file("${path.module}/agent-chassis.yaml")
  }
}

# StatefulSet for the generic orchestrator (bootstrap agent)
resource "kubernetes_stateful_set" "generic_orchestrator" {
  metadata {
    name      = "generic-orchestrator"
    namespace = var.namespace
    labels = {
      app         = "generic-orchestrator"
      component   = "bootstrap"
      agent-type  = "generic"
      managed-by  = "terraform"
    }
  }

  spec {
    service_name = "generic-orchestrator"
    replicas     = 1

    selector {
      match_labels = {
        app = "generic-orchestrator"
      }
    }

    template {
      metadata {
        labels = {
          app             = "generic-orchestrator"
          agent-type      = "generic"
          agent-id        = "00000000-0000-0000-0000-000000000001"
          component       = "agent"
          role            = "bootstrap"
        }

        annotations = {
          "prometheus.io/scrape" = "true"
          "prometheus.io/port"   = "9090"
          "prometheus.io/path"   = "/metrics"
        }
      }

      spec {
        service_account_name = "ai-persona-app"

        image_pull_secrets {
          name = "docker-hub-creds"
        }

        container {
          name  = "orchestrator"
          image = "${var.registry}/agent-chassis:${var.image_tag}"

          # Core configuration
          env {
            name  = "AGENT_TYPE"
            value = "generic"
          }

          env {
            name  = "AGENT_ID"
            value = "00000000-0000-0000-0000-000000000001"
          }

          env {
            name  = "CLIENT_ID"
            value = "demo_client"
          }

          env {
            name  = "KAFKA_TOPIC"
            value = "system.agent.generic.process"
          }

          env {
            name  = "KAFKA_CONSUMER_GROUP"
            value = "generic-orchestrator-group"
          }

          env {
            name  = "HEALTH_PORT"
            value = "8080"
          }

          env {
            name  = "METRICS_PORT"
            value = "9090"
          }

          # Database passwords from secrets
          env {
            name = "CLIENTS_DB_PASSWORD"
            value_from {
              secret_key_ref {
                name = data.kubernetes_secret.platform_secrets.metadata[0].name
                key  = "CLIENTS_DB_PASSWORD"
              }
            }
          }

          env {
            name = "TEMPLATES_DB_PASSWORD"
            value_from {
              secret_key_ref {
                name = data.kubernetes_secret.platform_secrets.metadata[0].name
                key  = "TEMPLATES_DB_PASSWORD"
              }
            }
          }

          env {
            name = "AUTH_DB_PASSWORD"
            value_from {
              secret_key_ref {
                name = data.kubernetes_secret.platform_secrets.metadata[0].name
                key  = "AUTH_DB_PASSWORD"
              }
            }
          }

          # API keys from secrets
          env {
            name = "ANTHROPIC_API_KEY"
            value_from {
              secret_key_ref {
                name = "personae-default-secrets"
                key  = "ANTHROPIC_API_KEY"
              }
            }
          }

          # Import all config from ConfigMap
          env_from {
            config_map_ref {
              name = data.kubernetes_config_map.prod_config.metadata[0].name
            }
          }

          # Ports
          port {
            name           = "health"
            container_port = 8080
            protocol       = "TCP"
          }

          port {
            name           = "metrics"
            container_port = 9090
            protocol       = "TCP"
          }

          # Health checks
          liveness_probe {
            http_get {
              path   = "/health"
              port   = "health"
              scheme = "HTTP"
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            success_threshold     = 1
            failure_threshold     = 3
          }

          readiness_probe {
            http_get {
              path   = "/ready"
              port   = "health"
              scheme = "HTTP"
            }
            initial_delay_seconds = 10
            period_seconds        = 5
            timeout_seconds       = 3
            success_threshold     = 1
            failure_threshold     = 3
          }

          # Resources
          resources {
            requests = {
              cpu    = "100m"
              memory = "256Mi"
            }
            limits = {
              cpu    = "500m"
              memory = "512Mi"
            }
          }

          # Volume mounts if needed
          volume_mount {
            name       = "agent-config-volume"
            mount_path = "/app/configs"
            read_only  = true
          }
        }

        # Volumes
        volume {
          name = "agent-config-volume"
          config_map {
            name = data.kubernetes_config_map.prod_config.metadata[0].name
          }
        }

        # Pod settings
        restart_policy = "Always"

        # Tolerations for scheduling
        toleration {
          key      = "node-role.kubernetes.io/master"
          operator = "Exists"
          effect   = "NoSchedule"
        }
      }
    }

    # Update strategy
    update_strategy {
      type = "RollingUpdate"
      rolling_update {
        partition = 0
      }
    }

    # Pod management policy
    pod_management_policy = "OrderedReady"
  }
}

# Service for the generic orchestrator (for health checks and metrics)
resource "kubernetes_service" "generic_orchestrator" {
  metadata {
    name      = "generic-orchestrator"
    namespace = var.namespace
    labels = {
      app        = "generic-orchestrator"
      component  = "bootstrap"
    }
  }

  spec {
    selector = {
      app = "generic-orchestrator"
    }

    port {
      name        = "health"
      port        = 8080
      target_port = "health"
      protocol    = "TCP"
    }

    port {
      name        = "metrics"
      port        = 9090
      target_port = "metrics"
      protocol    = "TCP"
    }

    type = "ClusterIP"
  }
}

# Optional: HorizontalPodAutoscaler (disabled for bootstrap agent)
# We keep replicas at 1 for the bootstrap orchestrator
resource "kubernetes_horizontal_pod_autoscaler_v2" "generic_orchestrator" {
  count = var.enable_autoscaling ? 1 : 0

  metadata {
    name      = "generic-orchestrator"
    namespace = var.namespace
  }

  spec {
    scale_target_ref {
      api_version = "apps/v1"
      kind        = "StatefulSet"
      name        = kubernetes_stateful_set.generic_orchestrator.metadata[0].name
    }

    min_replicas = 1
    max_replicas = 1  # Keep at 1 for bootstrap agent

    metric {
      type = "Resource"
      resource {
        name = "cpu"
        target {
          type                = "Utilization"
          average_utilization = 80
        }
      }
    }
  }
}

# Outputs
output "generic_orchestrator_name" {
  value       = kubernetes_stateful_set.generic_orchestrator.metadata[0].name
  description = "Name of the generic orchestrator StatefulSet"
}
