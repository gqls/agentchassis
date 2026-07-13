# terraform/modules/nginx-ingress/main.tf

resource "kubernetes_namespace" "ns" {
  count = var.create_namespace ? 1 : 0 # Create namespace only if variable is true
  metadata {
    name = var.ingress_namespace
    labels = {
      name = var.ingress_namespace
    }
  }
}

resource "helm_release" "ingress_nginx" {
  name       = "ingress-nginx"
  repository = "https://kubernetes.github.io/ingress-nginx"
  chart      = "ingress-nginx"
  namespace  = var.create_namespace ? kubernetes_namespace.ns[0].metadata[0].name : var.ingress_namespace
  version    = var.helm_chart_version

  values = var.helm_values_content != "" ? [var.helm_values_content] : []

  # Common overrides if not in values file, especially LoadBalancer type
  set {
    name  = "controller.service.type"
    value = "NodePort"
  }
/*  set {
    name  = "controller.replicaCount"
    value = "2" # Default to 2 replicas
  }*/

  depends_on = [kubernetes_namespace.ns] # Depends on namespace if created by module
}

data "kubernetes_service" "ingress_controller_service" {
  # This data source might fail if the service is not immediately available.
  # Consider making it optional or using a more robust way to get the IP if needed for immediate output.
  # For now, it assumes the service name convention from the chart.
  # The actual service name might vary based on the release name and chart.
  # Usually it's <release-name>-controller, so "ingress-nginx-controller".
  metadata {
    name      = "${helm_release.ingress_nginx.name}-controller"
    namespace = var.create_namespace ? kubernetes_namespace.ns[0].metadata[0].name : var.ingress_namespace
  }
  depends_on = [helm_release.ingress_nginx]
}