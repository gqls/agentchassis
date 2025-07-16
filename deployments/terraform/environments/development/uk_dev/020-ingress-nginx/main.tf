
module "nginx_ingress_dev" {
  source = "../../../../modules/nginx_ingress"

  ingress_namespace    = var.ingress_nginx_dev_namespace
  helm_chart_version   = var.ingress_nginx_dev_chart_version
  create_namespace     = true # Let the module create the namespace

  # For Kind, you usually don't need custom values unless you want specific NodePorts
  # or to disable LoadBalancer service type (which isn't typically used with Kind directly).
  # The module default of NodePort service type for the controller is fine for Kind.
  helm_values_content = yamlencode({
    controller = {
      kind = "Deployment" # DaemonSet is fine too, Deployment is often simpler for local Kind
      replicaCount = 1
      service = {
        type = "NodePort" # Exposes on NodePorts
        nodePorts = {
          http = var.ingress_nginx_dev_http_node_port
          https = var.ingress_nginx_dev_https_node_port
        }
      }
      # Disable admission webhooks for simpler Kind setup if they cause issues.
      admissionWebhooks = {
         enabled = false
      }
    }
  })
}
