# modules/strimzi-operator/main.tf

resource "null_resource" "install_strimzi" {
  triggers = {
    version = var.strimzi_version
    namespace = var.operator_namespace
    watched_namespaces = join(",", var.watched_namespaces_list)
  }

  provisioner "local-exec" {
    command = <<-EOT
      set -e

      # Set kubeconfig if provided
      ${var.cluster_kubeconfig_path != "" ? "export KUBECONFIG=${var.cluster_kubeconfig_path}" : ""}

      echo "Installing Strimzi ${var.strimzi_version}..."

      # Create namespace
      kubectl create namespace ${var.operator_namespace} --dry-run=client -o yaml | kubectl apply -f -

      # Download and apply Strimzi
      kubectl apply -f "https://strimzi.io/install/latest?namespace=${var.operator_namespace}" -n ${var.operator_namespace}

      # Wait for deployment
      kubectl wait --for=condition=available --timeout=300s deployment/strimzi-cluster-operator -n ${var.operator_namespace}

      # Update watched namespaces
      kubectl -n ${var.operator_namespace} set env deployment/strimzi-cluster-operator \
        STRIMZI_NAMESPACE="${join(",", var.watched_namespaces_list)}"

      # Setup RBAC for other namespaces
      %{for ns in var.watched_namespaces_list ~}
      %{if ns != var.operator_namespace ~}
      echo "Setting up RBAC for namespace ${ns}..."
      kubectl create namespace ${ns} --dry-run=client -o yaml | kubectl apply -f -

      kubectl apply -n ${ns} -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: strimzi-cluster-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: strimzi-cluster-operator-namespaced
subjects:
- kind: ServiceAccount
  name: strimzi-cluster-operator
  namespace: ${var.operator_namespace}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: strimzi-entity-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: strimzi-entity-operator
subjects:
- kind: ServiceAccount
  name: strimzi-cluster-operator
  namespace: ${var.operator_namespace}
EOF
      %{endif ~}
      %{endfor ~}

      echo "Strimzi installation complete!"

      # Fix leader election permissions
      echo "Applying leader election permissions..."
      kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: strimzi-cluster-operator-leader-election
  namespace: ${var.operator_namespace}
  labels:
    app: strimzi
rules:
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: strimzi-cluster-operator-leader-election
  namespace: ${var.operator_namespace}
  labels:
    app: strimzi
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: strimzi-cluster-operator-leader-election
subjects:
- kind: ServiceAccount
  name: strimzi-cluster-operator
  namespace: ${var.operator_namespace}
EOF

      # Restart operator to pick up permissions
      kubectl rollout restart deployment/strimzi-cluster-operator -n ${var.operator_namespace}
      kubectl rollout status deployment/strimzi-cluster-operator -n ${var.operator_namespace} --timeout=60s

      echo "Strimzi installation with leader election fix complete!"
    EOT
  }
}