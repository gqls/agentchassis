terraform {
  required_providers {
    kubernetes = { # Needed if you want to add data sources for services later, but not strictly for null_resource
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