# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/020-ingress-nginx/main.tf
module "nginx_ingress" {
  source = "../../../../modules/nginx-ingress" # Relative path to your new reusable module

  ingress_namespace    = var.ingress_target_namespace
  helm_chart_version = var.ingress_helm_chart_version_override == null ? null : var.ingress_helm_chart_version_override # Pass override or let module use default
  helm_values_content  = fileexists(var.ingress_custom_values_yaml_path) ? file(var.ingress_custom_values_yaml_path) : ""
  # create_namespace     = true # Or false if namespace is created elsewhere (e.g., by operator config)
  # If the namespace is also defined in 030-strimzi-operator for watched ns,
  # set create_namespace = false here and add depends_on to ensure it exists.
}
