module "kafka_cluster_dev" {
  source = "../../../../modules/kafka_cluster"

  kubeconfig_path         = abspath(pathexpand(var.kubeconfig_path))
  kube_context_name       = var.kube_context_name    // Pass the dev context name
  kafka_cr_namespace      = var.kafka_namespace_dev
  kafka_cr_yaml_file_path = var.kafka_cluster_cr_yaml_path_dev
  kafka_cr_cluster_name   = var.kafka_cluster_name_dev
}