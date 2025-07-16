# This module assumes the namespaces (operator_namespace and watched_namespaces_list)
# are created by a separate configuration or exist.
# The main.tf in the instance directory (e.g., 030-strimzi-operator) will create them.

resource "null_resource" "apply_strimzi_operator_yaml" {
  triggers = {
    operator_deployment_sha1 = fileexists("${var.strimzi_yaml_source_path}/${var.operator_deployment_yaml_filename}") ? filesha1("${var.strimzi_yaml_source_path}/${var.operator_deployment_yaml_filename}") : ""
    watched_namespaces_trigger = join(",", var.watched_namespaces_list)
  }

  provisioner "local-exec" {
    command = "kubectl apply --namespace ${var.operator_namespace} --filename ${var.strimzi_yaml_source_path}/"
    environment = {
      KUBECONFIG = var.cluster_kubeconfig_path
    }
  }
}

