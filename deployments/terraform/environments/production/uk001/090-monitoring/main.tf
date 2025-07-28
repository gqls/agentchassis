terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.36.0"
    }
  }
  backend "kubernetes" {
    secret_suffix = "tfstate-monitoring"
    config_path   = "~/.kube/config"
  }
}

# Add the prometheus-community Helm repository
data "helm_repository" "prometheus_community" {
  name = "prometheus-community"
  url  = "https://prometheus-community.github.io/helm-charts"
}

# Create a dedicated namespace for monitoring tools
resource "kubernetes_namespace" "monitoring_ns" {
  metadata {
    name = var.monitoring_namespace
  }
}

# Deploy the kube-prometheus-stack Helm chart
resource "helm_release" "prometheus_stack" {
  name       = "prometheus-stack"
  repository = data.helm_repository.prometheus_community.metadata[0].name
  chart      = "kube-prometheus-stack"
  namespace  = kubernetes_namespace.monitoring_ns.metadata[0].name
  version    = "51.8.0" # Pin to a specific chart version for consistency

  values = [
    templatefile("${path.module}/values.yaml.tpl", {
      grafana_admin_password = var.grafana_admin_password
    })
  ]

  depends_on = [kubernetes_namespace.monitoring_ns]
}