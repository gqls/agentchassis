# ~/projects/terraform/rackspace_generic/terraform/environments/production/sydney/040-kafka-cluster/outputs.tf
output "deployed_kafka_cluster_name" {
  value = module.kafka_cluster_service.cluster_name_applied
}
output "deployed_kafka_cluster_namespace" {
  value = module.kafka_cluster_service.cluster_namespace_applied
}
output "kafka_bootstrap_servers_plain" {
  value = module.kafka_cluster_service.bootstrap_servers_plain
}
output "kafka_bootstrap_servers_tls" {
  value = module.kafka_cluster_service.bootstrap_servers_tls
}

output "cluster_context_name" {
  description = "The kubectl context name for the production cluster."
  value       = "personae-uk001-prod-cluster" // This is usually derived or is a known value from the kubeconfig
  // You can often find this in the data.spot_kubeconfig.cluster_kubeconfig.kubeconfigs[0].name
  // value = data.spot_kubeconfig.cluster_kubeconfig.kubeconfigs[0].name
}