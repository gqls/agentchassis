output "operator_namespace_used" {
  value = var.operator_namespace
  depends_on = [null_resource.install_strimzi]
}

output "watched_namespaces_configured" {
  value = var.watched_namespaces_list
  depends_on = [null_resource.install_strimzi]
}

output "strimzi_version" {
  value = var.strimzi_version
}