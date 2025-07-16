provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = var.kube_context_name
}

provider "helm" {
  kubernetes {
    config_path    = "~/.kube/config"
    config_context = var.kube_context_name
  }
}