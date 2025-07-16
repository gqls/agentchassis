# ~/projects/terraform/rackspace_generic/terraform/environments/production/sydney/040-kafka-cluster/versions.tf
terraform {
  required_providers {
    kubernetes = { # Even if module doesn't use it directly, root config might for data sources
      source  = "hashicorp/kubernetes"
      version = "~> 2.36.0"
    }
    null = { # Because the module uses null_resource
      source  = "hashicorp/null"
      version = "~> 3.2.4"
    }
  }
  required_version = ">= 1.0"
}