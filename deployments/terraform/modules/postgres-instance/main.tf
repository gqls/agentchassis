resource "kubernetes_secret" "postgres_secret" {
  metadata {
    name      = "${var.instance_name}-secret"
    namespace = var.namespace
    labels = {
      app = var.instance_name
    }
  }
  data = {
    "POSTGRES_USER"     = var.database_user
    "POSTGRES_PASSWORD" = var.database_pass
    "POSTGRES_DB"       = var.database_name
  }
  type = "Opaque"
}

resource "kubernetes_stateful_set" "postgres_sts" {
  metadata {
    name      = var.instance_name
    namespace = var.namespace
  }
  spec {
    service_name = "${var.instance_name}-headless"
    replicas     = 1

    selector {
      match_labels = {
        app = var.instance_name
      }
    }

    template {
      metadata {
        labels = {
          app = var.instance_name
        }
      }
      spec {
        termination_grace_period_seconds = 10
        container {
          name  = "postgres"
          image = "postgres:15-alpine"
          port {
            container_port = 5432
            name           = "postgres"
          }
          env_from {
            secret_ref {
              name = kubernetes_secret.postgres_secret.metadata[0].name
            }
          }
          volume_mount {
            name       = "postgres-storage"
            mount_path = "/var/lib/postgresql/data"
          }
          liveness_probe {
            exec {
              command = ["pg_isready", "-U", var.database_user, "-d", var.database_name]
            }
            initial_delay_seconds = 30
            period_seconds        = 10
          }
          readiness_probe {
            exec {
              command = ["pg_isready", "-U", var.database_user, "-d", var.database_name]
            }
            initial_delay_seconds = 5
            period_seconds        = 5
          }
        }
      }
    }
    volume_claim_template {
      metadata {
        name = "postgres-storage"
      }
      spec {
        access_modes       = ["ReadWriteOnce"]
        storage_class_name = var.storage_class_name
        resources {
          requests = {
            storage = var.storage_size
          }
        }
      }
    }
  }
  depends_on = [kubernetes_secret.postgres_secret]
}

resource "kubernetes_service" "postgres_service" {
  metadata {
    name      = var.instance_name
    namespace = var.namespace
  }
  spec {
    selector = {
      app = var.instance_name
    }
    port {
      port        = 5432
      target_port = 5432
    }
    type = "ClusterIP"
  }
}