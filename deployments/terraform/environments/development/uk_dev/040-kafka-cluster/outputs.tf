output "dev_kafka_cluster_name" {
  description = "Name of the Kafka cluster deployed in dev."
  value       = module.kafka_cluster_dev.cluster_name_applied
}

output "dev_kafka_cluster_namespace" {
  description = "Namespace of the Kafka cluster in dev."
  value       = module.kafka_cluster_dev.cluster_namespace_applied
}

output "dev_kafka_bootstrap_servers_plain" {
  description = "Internal Plaintext Bootstrap Servers for dev Kafka."
  value       = module.kafka_cluster_dev.bootstrap_servers_plain
}

output "dev_kafka_bootstrap_servers_tls" {
  description = "Internal TLS Bootstrap Servers for dev Kafka."
  value       = module.kafka_cluster_dev.bootstrap_servers_tls
}