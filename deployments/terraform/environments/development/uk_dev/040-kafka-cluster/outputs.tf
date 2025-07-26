output "dev_kafka_cluster_name" {
  description = "The name of the deployed Kafka cluster."
  value       = kubernetes_manifest.kafka_cluster.object.metadata.name
}

output "dev_kafka_cluster_namespace" {
  description = "The namespace of the deployed Kafka cluster."
  value       = kubernetes_manifest.kafka_cluster.object.metadata.namespace
}

output "dev_kafka_bootstrap_servers_plain" {
  description = "The plain bootstrap servers for the Kafka cluster."
  value       = "${var.kafka_cluster_name_dev}-kafka-bootstrap.${var.kafka_namespace_dev}.svc:9092"
}

output "dev_kafka_bootstrap_servers_tls" {
  description = "The TLS bootstrap servers for the Kafka cluster."
  value       = "${var.kafka_cluster_name_dev}-kafka-bootstrap.${var.kafka_namespace_dev}.svc:9093"
}