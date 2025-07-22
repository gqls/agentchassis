terraform {
  required_version = ">= 1.0"
  # Remove backend block - modules shouldn't have backends

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
  }
}

# Job to create initial system topics only.
# Agent topics are created dynamically by the Core Manager service.
resource "kubernetes_job" "kafka_system_topics" {
  metadata {
    name      = "kafka-system-topics-init-${substr(sha1(timestamp()), 0, 8)}"
    namespace = var.namespace
  }

  spec {
    template {
      metadata {
        labels = {
          app = "kafka-topic-init"
        }
      }

      spec {
        service_account_name = "default"  # Change if you have a specific SA
        restart_policy       = "OnFailure"

        container {
          name  = "topic-creator"
          image = "confluentinc/cp-kafka:7.5.0"

          command = ["/bin/bash", "-c"]
          args = [<<-EOT
            set -ex
            KAFKA_BROKERS="personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092"

            echo "Waiting for Kafka..."
            until kafka-topics --bootstrap-server $KAFKA_BROKERS --list >/dev/null 2>&1; do
              sleep 5
            done
            echo "Kafka is ready."

            create_topic() {
              local topic=$1
              local partitions=$${2:-1}
              local replication=$${3:-1}

              if kafka-topics --bootstrap-server $KAFKA_BROKERS --list | grep -q "^$$topic$$"; then
                echo "Topic $$topic already exists."
              else
                echo "Creating topic: $$topic"
                kafka-topics --bootstrap-server $KAFKA_BROKERS \
                  --create --topic "$$topic" --partitions "$$partitions" --replication-factor "$$replication" \
                  --if-not-exists
              fi
            }

            # System & Orchestration Topics
            create_topic "system.commands.workflow.resume" 1 1
            create_topic "system.events.workflow.paused" 1 1
            create_topic "system.events.workflow.completed" 1 1
            create_topic "system.events" 3 1
            create_topic "system.errors" 3 1
            create_topic "audit.log" 3 1

            echo "System topic initialization complete."
            EOT
          ]
        }
      }
    }
    backoff_limit = 3
  }

  wait_for_completion = true

  timeouts {
    create = "5m"
  }
}