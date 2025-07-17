terraform {
  required_providers {
    kafka = {
      source  = "mongey/kafka"
      version = "~> 0.11.0"
    }
  }
  backend "kubernetes" {
    secret_suffix = "tfstate-kafka-topics-dev"
    config_path   = "~/.kube/config"
  }
}

# Read the outputs from the dev Kafka cluster layer
data "terraform_remote_state" "kafka_cluster_dev" {
  backend = "kubernetes"
  config = {
    # This suffix must match the backend config of your dev 040-kafka-cluster layer
    secret_suffix = "tfstate-kafka-cluster-dev"
    config_path   = "~/.kube/config"
  }
}

provider "kafka" {
  bootstrap_servers = data.terraform_remote_state.kafka_cluster_dev.outputs.kafka_bootstrap_servers
}

# Define all required Kafka topics for the platform
resource "kafka_topic" "topics_dev" {
  for_each = toset(var.platform_topics)

  name               = each.key
  partitions         = var.default_partitions
  replication_factor = var.default_replication_factor
  config = {
    "retention.ms" = "604800000" # Retain messages for 7 days in dev
  }
}