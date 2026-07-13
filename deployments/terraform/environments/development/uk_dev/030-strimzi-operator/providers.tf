provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = var.kube_context_name
}

provider "null" {} // If your module uses null_resource