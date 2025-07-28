terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-kafka-topics-dev"
    config_path   = var.kubeconfig_path
  }
}

provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.context_name
}

module "kafka_topics" {
  source = "../../../../modules/kafka_topics"

  namespace         = var.namespace  # or wherever your Kafka is deployed
  kube_context_name = var.context_name
}