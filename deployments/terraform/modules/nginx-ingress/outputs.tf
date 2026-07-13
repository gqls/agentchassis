# terraform/modules/nginx-ingress/outputs.tf
# ~/projects/terraform/rackspace_generic/terraform/modules/nginx-ingress/outputs.tf

output "namespace" {
  description = "Namespace where the ingress controller is deployed."
  value       = var.create_namespace ? kubernetes_namespace.ns[0].metadata[0].name : var.ingress_namespace
}

output "release_name" {
  description = "Helm release name for the ingress controller."
  value       = helm_release.ingress_nginx.name
}

output "loadbalancer_ip" {
  description = "External IP or Hostname of the NGINX Ingress controller LoadBalancer."
  value = try(
    # Attempt to get IP from the first ingress of the first load_balancer
    data.kubernetes_service.ingress_controller_service.status[0].load_balancer[0].ingress[0].ip,
    # Fallback to hostname if IP isn't present
    data.kubernetes_service.ingress_controller_service.status[0].load_balancer[0].ingress[0].hostname,
    "IP/Hostname pending or not a LoadBalancer" # Generic fallback
  )
}

# For debugging the structure, you can add:
output "debug_ingress_load_balancer_status_block" {
  description = "The raw load_balancer status block for debugging."
  value       = try(data.kubernetes_service.ingress_controller_service.status[0].load_balancer, null)
}