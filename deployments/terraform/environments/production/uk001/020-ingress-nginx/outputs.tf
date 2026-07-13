# terraform/environments/production/uk001/020-ingress-nginx/outputs.tf

output "ingress_loadbalancer_ip" {
  description = "External IP of the NGINX Ingress for uk001."
  value       = module.nginx_ingress.loadbalancer_ip
}