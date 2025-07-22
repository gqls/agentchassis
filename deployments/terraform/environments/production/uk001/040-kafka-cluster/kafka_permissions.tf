# Grant topic management permissions to the core-manager service account
resource "kubernetes_manifest" "core_manager_kafka_user" {
  manifest = {
    "apiVersion" = "kafka.strimzi.io/v1beta2"
    "kind"       = "KafkaUser"
    "metadata" = {
      "name"      = "core-manager-user"
      # The KafkaUser must be in the same namespace as the Kafka Cluster
      "namespace" = "kafka"
      "labels" = {
        "strimzi.io/cluster" = "personae-kafka-cluster"
      }
    }
    "spec" = {
      "authorization" = {
        "type" = "simple"
        "acls" = [
          {
            "resource" = {
              "type"        = "topic"
              "name"        = "*"
              "patternType" = "literal"
            }
            "operations" = [
              "Create",
              "Describe",
              "Alter"
            ]
            "host" = "*"
          }
        ]
      }
    }
  }
}