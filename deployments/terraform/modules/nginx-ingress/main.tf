resource "kubernetes_namespace" "ns" {
  metadata {
    name = var.ingress_namespace
  }
}

resource "helm_release" "ingress_nginx" {
  name       = "ingress-nginx"
  repository = "https://kubernetes.github.io/ingress-nginx"
  chart      = "ingress-nginx"
  namespace  = kubernetes_namespace.ns.metadata[0].name
  version    = var.helm_chart_version

  # Pass the entire values file content. This is cleaner than many 'set' blocks.
  values = [var.helm_values_content]

  depends_on = [kubernetes_namespace.ns]
}
