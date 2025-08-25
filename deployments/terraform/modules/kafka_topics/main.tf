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

            # --- SYSTEM & ORCHESTRATION TOPICS (Generally consistent) ---
            create_topic "orchestrator.state-changes" 12 3 "Orchestrator state change notifications"
            create_topic "human.approvals" 6 3 "Human approval workflow messages"
            create_topic "system.events" 3 3 "General system events"
            create_topic "system.notifications.ui" 3 3 "UI notifications"
            create_topic "system.commands.workflow.resume" 3 3 "Workflow resume commands"

            create_topic "system.errors" 3 3 "General system error messages"
            create_topic "audit.log" 3 3 "Audit trail for user actions"
            create_topic "system.metrics.agents" 3 3 "Agent performance metrics"
            create_topic "system.logs.errors" 3 3 "Error logs aggregation"
            create_topic "system.audit.actions" 6 3 "Audit trail for user actions"

            # --- CORE SERVICE TOPICS ---
            # These are from your original TF job. Confirm if they are still needed or if Core Manager generates events.
            create_topic "requests.auth.user.create" 3 3 # Check if actually used for internal requests
            create_topic "events.auth.user.created" 3 3  # Check if actually used for internal events

            # --- GENERIC AGENT CHASSIS TOPICS ---
            create_topic "system.agent.generic.requests" 6 3 "Generic agent chassis requests" # From Kustomize job
            create_topic "system.agent.generic.responses" 6 3
            create_topic "system.agent.generic.errors" 1 3
            create_topic "system.agent.generic.dlq" 1 3

            # --- SPECIALIZED AGENT & ADAPTER COMMUNICATION TOPICS (CRITICAL ALIGNMENT) ---
            # Reasoning Agent
            create_topic "system.agent.reasoning.requests" 6 3 "Reasoning agent requests"  # Matches internal/agents/reasoning/agent.go RequestTopic
            create_topic "system.agent.reasoning.responses" 6 3 "Reasoning agent responses"      # Matches internal/agents/reasoning/agent.go ResponseTopic
            create_topic "system.agent.reasoning.errors" 1 3 "Reasoning agent error DLQ"         # Matches platform/kafka/topic_manager.go error topic pattern
            create_topic "system.agent.reasoning.dlq" 1 3 "Reasoning agent dead letter queue"           # Consistent DLQ naming, used by Kafka Connect

            # Web Search Adapter
            create_topic "system.agent.websearch.requests" 3 3 "Web search adapter requests"     # Matches internal/adapters/websearch/adapter.go requestTopic
            create_topic "system.agent.websearch.responses" 6 3 "Web search adapter responses"   # Matches internal/adapters/websearch/adapter.go responseTopic
            create_topic "system.agent.websearch.errors" 1 3 "Web search adapter error DLQ"      # Matches platform/kafka/topic_manager.go error topic pattern
            create_topic "system.agent.websearch.dlq" 1 3 "Web search adapter dead letter queue"        # Consistent DLQ naming

            # Image Generator Adapter
            create_topic "system.agent.image-generator.requests" 3 3 "Image generation adapter requests" # Matches internal/adapters/imagegenerator/adapter.go requestTopic
            create_topic "system.agent.image-generator.responses" 6 3 "Image generation adapter responses"      # Matches internal/adapters/imagegenerator/adapter.go responseTopic
            create_topic "system.agent.image-generator.errors" 1 3 "Image generation adapter error DLQ"         # Matches platform/kafka/topic_manager.go error topic pattern
            create_topic "system.agent.image-generator.dlq" 1 3 "Image generation adapter dead letter queue"           # Consistent DLQ naming

            # Content Creator Agent (NEW)
            create_topic "system.agent.content-creator.requests" 6 3 "Content creator process requests" # Matches internal/agents/contentcreator/agent.go RequestTopic
            create_topic "system.agent.content-creator.responses" 6 3 "Content creator responses"      # Matches internal/agents/contentcreator/agent.go ResponseTopic
            create_topic "system.agent.content-creator.errors" 1 3 "Content creator error DLQ"         # Matches platform/kafka/topic_manager.go error topic pattern
            create_topic "system.agent.content-creator.dlq" 1 3 "Content creator dead letter queue"           # Consistent DLQ naming

            # --- TASK/PRIORITY QUEUES (for Data-Driven Agents) ---
            # These are typically created by topic_manager if category is 'data-driven'.
            # If you want to guarantee their existence from startup for all data-driven agents,
            # you can explicitly add them here. Otherwise, rely on Core Manager.
            # Example if you add them:
            create_topic "tasks.high.copywriter" 3 3
            create_topic "tasks.normal.copywriter" 3 3
            create_topic "tasks.low.copywriter" 3 3
            create_topic "system.agent.copywriter.responses." 6 3 # Response for copywriter tasks (from Kustomize job)
            create_topic "dlq.copywriter" 1 3 # DLQ for copywriter

            create_topic "tasks.high.researcher" 3 3
            create_topic "tasks.normal.researcher" 3 3
            create_topic "tasks.low.researcher" 3 3
            create_topic "system.agent.researcher.responses" 6 3 # Response for researcher tasks (from Kustomize job)
            create_topic "dlq.researcher" 1 3 # DLQ for researcher

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