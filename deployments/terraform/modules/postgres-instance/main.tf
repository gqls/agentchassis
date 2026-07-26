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
        security_context {
          fs_group = 999  # PostgreSQL group
          run_as_user = 999  # PostgreSQL user
          run_as_non_root = true
        }
        container {
          name  = "postgres"
          image = "pgvector/pgvector:pg15"
          security_context {
            run_as_user = 999
            run_as_non_root = true
            allow_privilege_escalation = false
          }
          port {
            container_port = 5432
            name           = "postgres"
          }
          env {
            name  = "PGDATA"
            value = "/var/lib/postgresql/data/pgdata"
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
          # bugs_open/082. Until 2026-07-26 this block did not exist, so every
          # database this module builds ran BestEffort: no CPU request means
          # cpu.shares = 2, the kernel minimum. A co-tenant with requests.cpu = 2
          # therefore outweighed the clients DB roughly 1024:1 on a contended
          # node, and the database was descheduled to a standstill while using
          # 30m CPU and 32Mi RSS. BestEffort is also the first QoS class evicted
          # under memory pressure. The request is the floor that matters; the
          # limits are burst headroom, deliberately well above observed use
          # (~210Mi RSS) so that a limit never becomes the thing that kills a
          # database.
          resources {
            requests = {
              cpu    = var.cpu_request
              memory = var.memory_request
            }
            limits = {
              cpu    = var.cpu_limit
              memory = var.memory_limit
            }
          }
          # An exec probe must fork a process inside the container. Neither
          # probe set timeout_seconds, so both inherited the Kubernetes default
          # of 1s — unachievable for a starved container, so kubelet killed a
          # healthy database every ~90s and the Service lost its only endpoint.
          # The resources block above makes 5s comfortably achievable; the
          # timeout is what stops a moment of contention reading as death.
          liveness_probe {
            exec {
              command = ["pg_isready", "-U", var.database_user, "-d", var.database_name]
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 6
          }
          # replicas = 1, so there is no second backend to fail over to.
          # Dropping the only endpoint does not route traffic anywhere better —
          # it converts "queries are slow" into "no such host" for the entire
          # fleet. Tolerating a slow probe here costs nothing that failing it
          # would have saved.
          readiness_probe {
            exec {
              command = ["pg_isready", "-U", var.database_user, "-d", var.database_name]
            }
            initial_delay_seconds = 5
            period_seconds        = 5
            timeout_seconds       = 5
            failure_threshold     = 6
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