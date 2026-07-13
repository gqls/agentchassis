# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/020-ingress-nginx/providers.tf
provider "kubernetes" {
  config_path = var.kubeconfig_path
}

provider "helm" {
  kubernetes {
    config_path = var.kubeconfig_path
  }
}