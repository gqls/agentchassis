# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/030-strimzi-operator/main.tf

# Define/Ensure the namespaces exist
resource "kubernetes_namespace" "operator_ns" {
  metadata {
    name = var.strimzi_operator_target_namespace # "strimzi"
  }
}
resource "kubernetes_namespace" "kafka_ns_for_watch" {
  # Ensure this namespace exists if it's in the watched list and not created elsewhere
  # Could also use a data source if creation is handled by a different TF config
  metadata {
    name = "kafka"
  }
}
resource "kubernetes_namespace" "personae_ns_for_watch" {
  metadata {
    name = "personae"
  }
}

module "strimzi_operator_service" {
  source = "../../../../modules/strimzi_operator" # Path to your reusable module

  operator_namespace    = kubernetes_namespace.operator_ns.metadata[0].name
  watched_namespaces_list = var.watched_namespaces_for_uk001
  strimzi_yaml_source_path = var.strimzi_yaml_bundle_path_for_uk001
  cluster_kubeconfig_path = var.kubeconfig_path # Module needs this for its local-exec
}

output "operator_namespace" {
  value = module.strimzi_operator_service.operator_namespace_used
}
output "watched_namespaces" {
  value = module.strimzi_operator_service.watched_namespaces_configured
}


resource "kubernetes_role" "strimzi_ingress_reader_kafka_ns" {
  metadata {
    name      = "strimzi-ingress-reader"
    namespace = "kafka" # Permissions within the 'kafka' namespace
  }
  rule {
    api_groups = ["networking.k8s.io"]
    resources  = ["ingresses"]
    verbs      = ["get", "list", "watch"]
  }
  # Ensure this depends on the kafka namespace existing if created by this TF
  depends_on = [kubernetes_namespace.kafka_ns_for_watch]
}

resource "kubernetes_role_binding" "strimzi_ingress_reader_kafka_ns_binding" {
  metadata {
    name      = "strimzi-ingress-reader-binding"
    namespace = "kafka" # RoleBinding in the 'kafka' namespace
  }
  subject {
    kind      = "ServiceAccount"
    name      = "strimzi-cluster-operator"
    namespace = "strimzi" # The SA is in the 'strimzi' namespace
  }
  role_ref {
    kind      = "Role"
    name      = kubernetes_role.strimzi_ingress_reader_kafka_ns.metadata[0].name
    api_group = "rbac.authorization.k8s.io"
  }
  depends_on = [kubernetes_role.strimzi_ingress_reader_kafka_ns]
}

# Explicitly create/manage the RoleBinding in the 'kafka' namespace.
# This binds the strimzi-cluster-operator SA (from 'strimzi' namespace)
# to the 'strimzi-cluster-operator-namespaced' ClusterRole
# for actions *within* the 'kafka' namespace.
resource "kubernetes_role_binding" "strimzi_operator_permissions_in_kafka_ns" {
  metadata {
    name      = "strimzi-cluster-operator-kafka-namespace-permissions" # A descriptive name
    namespace = "kafka"                                                # Binding is in the 'kafka' namespace
  }

  subject {
    kind      = "ServiceAccount"
    name      = "strimzi-cluster-operator"            # Name of the ServiceAccount
    namespace = var.strimzi_operator_target_namespace # Namespace of the SA (e.g., "strimzi")
  }

  role_ref {
    kind      = "ClusterRole"
    name      = "strimzi-cluster-operator-namespaced" # The ClusterRole that has the necessary permissions
    # (including for "roles" and "rolebindings")
    api_group = "rbac.authorization.k8s.io"
  }

  depends_on = [
    kubernetes_namespace.kafka_ns_for_watch,    // Ensures 'kafka' namespace exists
    module.strimzi_operator_service             // Ensures Strimzi operator YAMLs (which define SA and ClusterRole) are applied first
  ]
}

# Explicitly create/manage the RoleBinding in the 'personae' namespace.
resource "kubernetes_role_binding" "strimzi_operator_permissions_in_personae_ns" {
  metadata {
    name      = "strimzi-cluster-operator-personae-namespace-permissions"
    namespace = "personae" # Binding is in the 'personae' namespace
  }

  subject {
    kind      = "ServiceAccount"
    name      = "strimzi-cluster-operator"
    namespace = var.strimzi_operator_target_namespace
  }

  role_ref {
    kind      = "ClusterRole"
    name      = "strimzi-cluster-operator-namespaced"
    api_group = "rbac.authorization.k8s.io"
  }

  depends_on = [
    kubernetes_namespace.personae_ns_for_watch, // Ensures 'personae' namespace exists
    module.strimzi_operator_service
  ]
}