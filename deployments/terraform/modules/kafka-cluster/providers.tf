# terraform/environments/production/services-sydney/030-kafka-cluster/providers.tf
provider "kubernetes" {
  config_path = var.kubeconfig_path
}