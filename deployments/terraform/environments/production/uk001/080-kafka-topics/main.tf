terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-kafka-topics"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.36.0"
    }
  }
}

provider "kubernetes" {
  config_path    = var.kubeconfig_path
}

module "kafka_topics" {
  source = "../../../../modules/kafka_topics"
  namespace         = var.namespace  # or wherever your Kafka is deployed
}