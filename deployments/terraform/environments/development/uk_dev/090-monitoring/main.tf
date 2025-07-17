terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.20"
    }
  }
  backend "kubernetes" {
    secret_suffix = "tfstate-monitoring-dev"
    config_path   = "~/.kube/config"
  }
}

data "helm_repository" "prometheus_community" {
  name = "prometheus-community"
  url  = "https://prometheus-community.github.io/helm-charts"
}

resource "kubernetes_namespace" "monitoring_ns_dev" {
  metadata {
    name = var.monitoring_namespace
  }
}

resource "helm_release" "prometheus_stack_dev" {
  name       = "prometheus-stack-dev"
  repository = data.helm_repository.prometheus_community.metadata[0].name
  chart      = "kube-prometheus-stack"
  namespace  = kubernetes_namespace.monitoring_ns_dev.metadata[0].name
  version    = "51.8.0" # Pin to the same chart version as production

  values = [
    templatefile("${path.module}/values.yaml.tpl", {
      grafana_admin_password = var.grafana_admin_password
    })
  ]

  depends_on = [kubernetes_namespace.monitoring_ns_dev]
}