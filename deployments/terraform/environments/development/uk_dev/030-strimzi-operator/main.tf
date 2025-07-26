# Ensure the namespaces Strimzi will operate in or watch exist.
# Strimzi operator's own namespace:
resource "kubernetes_namespace" "operator_ns" {
  metadata {
    name = var.strimzi_operator_dev_namespace // e.g., "strimzi"
  }
}

# Namespace for Kafka clusters (watched by Strimzi):
resource "kubernetes_namespace" "kafka_cluster_ns" {
  metadata {
    name = "kafka" // Assuming Kafka CRs will be in 'kafka' namespace
  }
}

# Namespace for Personae app (if Strimzi needs to manage KafkaUsers there):
resource "kubernetes_namespace" "personae_app_ns" {
  metadata {
    name = "personae" // Assuming Personae app and potentially KafkaUsers are in 'personae'
  }
}

module "strimzi_operator" {
  source = "../../../../modules/strimzi-operator"

  operator_namespace                = kubernetes_namespace.operator_ns.metadata[0].name
  watched_namespaces_list           = var.watched_namespaces_dev
  strimzi_yaml_source_path          = var.strimzi_yaml_bundle_path_dev
  operator_deployment_yaml_filename = var.strimzi_operator_deployment_yaml_filename_dev
  cluster_kubeconfig_path           = "" # Path to the kubeconfig Terraform should use

  depends_on = [
    kubernetes_namespace.operator_ns,
    kubernetes_namespace.kafka_cluster_ns,
    kubernetes_namespace.personae_app_ns
  ]
}

