# Create the KafkaNodePool first
resource "kubernetes_manifest" "kafka_nodepool" {
  manifest = yamldecode(templatefile("${path.module}/../../../../modules/kafka-cluster/config/kafka-nodepool-cr-dev.yaml", {
    cluster_name = var.kafka_cluster_name_dev
    namespace    = var.kafka_namespace_dev
  }))
}

# Then create the Kafka cluster
resource "kubernetes_manifest" "kafka_cluster" {
  depends_on = [kubernetes_manifest.kafka_nodepool]

  manifest = yamldecode(templatefile("${path.module}/../../../../modules/kafka-cluster/config/kafka-cluster-cr-dev.yaml", {
    cluster_name = var.kafka_cluster_name_dev
    namespace    = var.kafka_namespace_dev
  }))
}