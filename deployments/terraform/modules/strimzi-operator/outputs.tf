output "operator_namespace_used" {
  description = "Namespace where the Strimzi operator was deployed."
  value       = var.operator_namespace
}

output "watched_namespaces_configured" {
  description = "Namespaces the Strimzi operator is configured to watch."
  value       = var.watched_namespaces_list
}