terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-kafka-topics-dev"
    config_path   = "~/.kube/config"
  }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = var.kube_context_name
}

module "kafka_topics" {
  source = "../../../../modules/kafka_topics"

  namespace         = "kafka"  # or wherever your Kafka is deployed
  kube_context_name = var.kube_context_name
}