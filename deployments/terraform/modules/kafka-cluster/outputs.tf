# terraform/modules/kafka_cluster/outputs.tf
output "cluster_name_applied" {
  description = "The name of the Kafka cluster that was applied."
  value       = var.kafka_cr_cluster_name
}

output "cluster_namespace_applied" {
  description = "The namespace where the Kafka cluster CR was applied."
  value       = var.kafka_cr_namespace
}

output "bootstrap_servers_plain" {
  description = "Assumed Internal Plaintext Bootstrap Servers."
  value       = "${var.kafka_cr_cluster_name}-kafka-bootstrap.${var.kafka_cr_namespace}.svc:9092"
}

output "bootstrap_servers_tls" {
  description = "Assumed Internal TLS Bootstrap Servers."
  value       = "${var.kafka_cr_cluster_name}-kafka-bootstrap.${var.kafka_cr_namespace}.svc:9093"
}