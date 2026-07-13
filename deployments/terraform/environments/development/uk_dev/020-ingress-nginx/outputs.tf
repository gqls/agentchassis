
output "http_node_port" {
  description = "HTTP NodePort for the Nginx Ingress controller in dev."
  value       = var.ingress_nginx_dev_http_node_port
}
output "https_node_port" {
  description = "HTTPS NodePort for the Nginx Ingress controller in dev."
  value       = var.ingress_nginx_dev_https_node_port
}
output "namespace" {
  description = "Namespace of the Nginx Ingress controller in dev."
  value       = module.nginx_ingress_dev.namespace // Assuming your module outputs this
}