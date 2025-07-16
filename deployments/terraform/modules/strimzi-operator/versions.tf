
terraform {
  required_version = ">=1.0"
  required_providers {
    helm = { source = "hashicorp/helm", version = "~> 2.17.0" }
    kubernetes = { source = "hashicorp/kubernetes", version = "~> 2.36.0"}
    null = { source = "hashicorp/null", version = "~> 3.2.4" }
  }
}