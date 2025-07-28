# modules/kafka_topics/main.tf

terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.36.0"
    }
  }
}

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
        service_account_name = "default"
        restart_policy       = "OnFailure"

        container {
          name  = "topic-creator"
          image = "confluentinc/cp-kafka:7.5.0"

          command = ["/bin/bash", "-c"]
          args = [<<-EOT
            set -ex
            KAFKA_BROKERS="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

            echo "Waiting for Kafka..."
            until kafka-topics --bootstrap-server $KAFKA_BROKERS --list >/dev/null 2>&1; do
              sleep 5
            done
            echo "Kafka is ready."

            create_topic() {
              local topic=$1
              local partitions=$${2:-3}
              local replication=$${3:-3}

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
            create_topic "system.commands.workflow.resume" 3 3
            create_topic "system.events.workflow.paused" 3 3
            create_topic "system.events.workflow.completed" 3 3
            create_topic "system.events" 3 3
            create_topic "system.errors" 3 3
            create_topic "audit.log" 3 3

            # Core Service Topics
            create_topic "requests.auth.user.create" 3 3
            create_topic "events.auth.user.created" 3 3

            # Agent Communication Topics
            create_topic "requests.agent.task.execute" 3 3
            create_topic "events.agent.task.completed" 3 3
            create_topic "events.agent.task.failed" 3 3
            create_topic "events.agent.task.progress" 3 3

            # Specialized Agent Topics
            create_topic "requests.agent.reasoning" 3 3
            create_topic "requests.agent.web-search" 3 3
            create_topic "requests.agent.image-generation" 3 3

            echo "Platform topic initialization complete."
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