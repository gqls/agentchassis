resource "kubernetes_role_binding" "strimzi_watched_namespace_permissions" {
  # This for_each creates a binding only in the 'kafka' namespace
  for_each = toset([for ns in var.watched_namespaces_for_uk001 : ns if ns != var.strimzi_operator_target_namespace])

  metadata {
    name      = "strimzi-cluster-operator-watched-ns"
    namespace = each.key # This will be 'kafka'
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "strimzi-cluster-operator-watched"
  }

  subject {
    kind      = "ServiceAccount"
    name      = "strimzi-cluster-operator"
    namespace = var.strimzi_operator_target_namespace # This should be 'strimzi'
  }

  depends_on = [
    kubernetes_manifest.strimzi_resources
  ]
}

resource "kubernetes_role_binding" "strimzi_core_resource_permissions" {
  for_each = toset([for ns in var.watched_namespaces_for_uk001 : ns if ns != var.strimzi_operator_target_namespace])

  metadata {
    name      = "strimzi-cluster-operator-core-ns"
    namespace = each.key
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    # This is the role that allows creation of Secrets, Pods, Services, etc.
    name      = "strimzi-cluster-operator-namespaced"
  }

  subject {
    kind      = "ServiceAccount"
    name      = "strimzi-cluster-operator"
    namespace = var.strimzi_operator_target_namespace
  }

  depends_on = [
    kubernetes_manifest.strimzi_resources
  ]
}

# This binding allows the Cluster Operator to delegate permissions
# to the Entity Operator that it creates.
resource "kubernetes_cluster_role_binding" "strimzi_entity_operator_delegation" {
  metadata {
    name = "strimzi-cluster-operator-entity-operator-delegation"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    # This is the role containing permissions for KafkaTopic and KafkaUser
    name      = "strimzi-entity-operator"
  }

  subject {
    kind      = "ServiceAccount"
    name      = "strimzi-cluster-operator"
    namespace = var.strimzi_operator_target_namespace
  }

  depends_on = [
    kubernetes_manifest.strimzi_resources
  ]
}