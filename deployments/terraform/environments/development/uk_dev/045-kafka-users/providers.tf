provider "kubernetes" {
  config_path    = abspath(pathexpand(var.kubeconfig_path))
  config_context = var.kube_context_name
}

provider "null" {}