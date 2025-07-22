output "kafka_users_created" {
  description = "List of Kafka users created"
  value = [
    "core-manager-user",
    "personae-app-anonymous"
  ]
}

output "kafka_cluster_associated" {
  description = "Kafka cluster the users are associated with"
  value       = var.kafka_cluster_name
}