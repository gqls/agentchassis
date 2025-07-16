resource "null_resource" "apply_kafka_cluster_cr" {
  triggers = {
    yaml_file_sha1 = fileexists(var.kafka_cr_yaml_file_path) ? filesha1(var.kafka_cr_yaml_file_path) : ""
    # Adding context and namespace to triggers to ensure re-apply if they change for some reason
    context_trigger   = var.kube_context_name
    namespace_trigger = var.kafka_cr_namespace
  }

  provisioner "local-exec" {
    command = "kubectl --kubeconfig=${var.kubeconfig_path} --context=${var.kube_context_name} apply --namespace ${var.kafka_cr_namespace} --filename ${var.kafka_cr_yaml_file_path}"
    # The KUBECONFIG env var is redundant if --kubeconfig is used in the command, but harmless.
    environment = {
      KUBECONFIG = var.kubeconfig_path
    }
  }
}