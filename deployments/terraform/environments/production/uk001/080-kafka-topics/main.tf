terraform {
  required_providers {
    kafka = {
      source  = "mongey/kafka"
      version = "~> 0.11.0"
    }
  }
  backend "kubernetes" {
    secret_suffix = "tfstate-kafka-topics"
    config_path   = "~/.kube/config"
  }
}

# Read the outputs from the Kafka cluster layer
data "terraform_remote_state" "kafka_cluster" {
  backend = "kubernetes"
  config = {
    secret_suffix = "tfstate-kafka-cluster" # Assuming this is the suffix used in 040
    config_path   = "~/.kube/config"
  }
}

provider "kafka" {
  bootstrap_servers = data.terraform_remote_state.kafka_cluster.outputs.kafka_bootstrap_servers
  # Add TLS/SASL config here if your production Kafka cluster requires it
}

# Define all required Kafka topics for the platform
resource "kafka_topic" "topics" {
  for_each = toset(var.platform_topics)

  name               = each.key
  partitions         = 3  # A good starting point for production
  replication_factor = 3  # Should match the number of Kafka brokers for resilience
  config = {
    "retention.ms" = "-1" # Retain messages indefinitely by default
    "cleanup.policy" = "compact,delete"
  }
}