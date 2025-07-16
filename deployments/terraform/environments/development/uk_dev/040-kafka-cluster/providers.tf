provider "kubernetes" {
  config_path    = abspath(pathexpand(var.kubeconfig_path_for_dev))
  config_context = var.kube_context_name
}

provider "helm" {
  kubernetes {
    config_path    = abspath(pathexpand(var.kubeconfig_path_for_dev))
    config_context = var.kube_context_name
  }
}
provider "null" {}