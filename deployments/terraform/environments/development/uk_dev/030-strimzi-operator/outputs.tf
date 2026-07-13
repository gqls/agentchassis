output "operator_namespace_used" {
  description = "Namespace where the Strimzi operator was deployed for dev."
  value       = module.strimzi_operator.operator_namespace_used
}

output "watched_namespaces_configured" {
  description = "Namespaces the Strimzi operator is configured to watch for dev."
  value       = module.strimzi_operator.watched_namespaces_configured
}