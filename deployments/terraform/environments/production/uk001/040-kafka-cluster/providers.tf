# ~/projects/terraform/rackspace_generic/terraform/environments/production/sydney/040-kafka-cluster/providers.tf
provider "kubernetes" {
  config_path = var.kubeconfig_path # Will be provided by Makefile/tfvars
}