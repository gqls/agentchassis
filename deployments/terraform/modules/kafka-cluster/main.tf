# Create the KafkaNodePool if using node pools
resource "kubernetes_manifest" "kafka_nodepool" {
  count = var.use_node_pools ? 1 : 0

  manifest = yamldecode(templatefile(var.kafka_nodepool_cr_yaml_file_path, {
    pool_name      = var.node_pool_name
    cluster_name   = var.kafka_cluster_name
    namespace      = var.kafka_namespace
    replicas       = var.node_pool_replicas
    storage_size   = var.node_pool_storage_size
    storage_class  = var.storage_class
    delete_claim   = var.delete_claim
  }))
}

# Create the Kafka cluster
resource "kubernetes_manifest" "kafka_cluster" {
  depends_on = [kubernetes_manifest.kafka_nodepool]

  manifest = yamldecode(templatefile(var.kafka_cr_yaml_file_path, {
    cluster_name        = var.kafka_cluster_name
    namespace           = var.kafka_namespace
    use_node_pools      = var.use_node_pools
    kafka_version       = var.kafka_version
    replicas            = var.node_pool_replicas
    replication_factor  = var.replication_factor
    min_insync_replicas = var.min_insync_replicas
    storage_size        = var.node_pool_storage_size
    storage_class       = var.storage_class
    delete_claim        = var.delete_claim
    message_max_bytes   = var.message_max_bytes
  }))
}