# terraform/environments/production/uk001/020-ingress-nginx/versions.tf

terraform {
  required_providers {
    kubernetes = { source = "hashicorp/kubernetes", version = "~> 2.36.0" }
    helm       = { source = "hashicorp/helm", version = "~> 2.17.0" }
  }
  required_version = ">= 1.0"
}