# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/030-strimzi-operator/versions.tf
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