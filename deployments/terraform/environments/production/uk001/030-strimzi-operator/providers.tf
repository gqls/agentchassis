# terraform/environments/production/uk001/030-strimzi-operator/providers.tf

provider "kubernetes" {
  config_path = var.kubeconfig_path
}

terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.36.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2.4"
    }
  }
  required_version = ">= 1.0"
}