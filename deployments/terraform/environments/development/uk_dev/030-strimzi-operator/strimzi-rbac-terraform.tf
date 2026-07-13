resource "kubernetes_cluster_role" "strimzi_kafka_namespace" {
  metadata {
    name = "strimzi-cluster-operator-kafka-namespace"
  }

  rule {
    api_groups = [""]
    resources  = ["pods", "services", "endpoints", "persistentvolumeclaims", "configmaps", "secrets", "serviceaccounts"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }

  rule {
    api_groups = ["apps"]
    resources  = ["deployments", "statefulsets", "replicasets"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }

  rule {
    api_groups = ["networking.k8s.io"]
    resources  = ["ingresses", "networkpolicies"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }

  rule {
    api_groups = ["kafka.strimzi.io"]
    resources  = ["*"]
    verbs      = ["*"]
  }

  rule {
    api_groups = ["core.strimzi.io"]
    resources  = ["*"]
    verbs      = ["*"]
  }

  rule {
    api_groups = ["rbac.authorization.k8s.io"]
    resources  = ["roles", "rolebindings"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }

  rule {
    api_groups = ["policy"]
    resources  = ["poddisruptionbudgets"]
    verbs      = ["get", "list", "watch", "create", "update", "patch", "delete"]
  }
}

resource "kubernetes_cluster_role_binding" "strimzi_kafka_namespace" {
  metadata {
    name = "strimzi-cluster-operator-kafka-namespace"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.strimzi_kafka_namespace.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = "strimzi-cluster-operator"
    namespace = "strimzi"
  }
}

resource "kubernetes_role_binding" "strimzi_kafka_namespace" {
  metadata {
    name      = "strimzi-cluster-operator-kafka-namespace"
    namespace = "kafka"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.strimzi_kafka_namespace.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = "strimzi-cluster-operator"
    namespace = "strimzi"
  }
}

resource "kubernetes_cluster_role_binding" "strimzi_entity_operator_delegation" {
  metadata {
    name = "strimzi-cluster-operator-entity-operator-delegation"
    labels = {
      app = "strimzi"
    }
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "strimzi-entity-operator"
  }

  subject {
    kind      = "ServiceAccount"
    name      = "strimzi-cluster-operator"
    namespace = "strimzi"
  }
}