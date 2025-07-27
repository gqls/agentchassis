# ~/projects/terraform/rackspace_generic/terraform/environments/production/sydney/040-kafka-cluster/main.tf

module "kafka_cluster_service" {
  source = "../../../../modules/kafka-cluster" # Path to your reusable module

  kafka_cr_namespace      = var.target_kafka_namespace
  kafka_cr_yaml_file_path = var.kafka_cluster_cr_yaml_path_uk001
  kafka_nodepool_cr_yaml_file_path = var.kafka_cluster_nodepool_cr_yaml_path_uk001
  kubeconfig_path = var.kubeconfig_path
  kafka_cr_cluster_name   = var.kafka_cluster_name_uk001
  kube_context_name       = var.kube_context_name

  # Pass the node pool configuration
  use_node_pools         = var.use_node_pools
  node_pool_name         = var.node_pool_name
  node_pool_replicas     = var.node_pool_replicas
  node_pool_storage_size = var.node_pool_storage_size
  kafka_version          = var.kafka_version
  replication_factor     = var.replication_factor
  min_insync_replicas    = var.min_insync_replicas
  storage_class          = var.storage_class
  delete_claim           = var.delete_claim
  kafka_cluster_name     = var.kafka_cluster_name_uk001
  kafka_namespace        = var.target_kafka_namespace
}
