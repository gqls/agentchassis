# This module properly handles both cluster-scoped and namespace-scoped resources

# Apply all Strimzi operator resources in the correct order
resource "null_resource" "apply_strimzi_operator" {
  triggers = {
    # Trigger on any change to force reapplication
    always_run = "${timestamp()}"
  }

  provisioner "local-exec" {
    command = <<-EOT
      set -e
      echo "Applying Strimzi operator resources..."

      # 1. Apply CRDs (cluster-scoped, no namespace)
      echo "Applying CRDs..."
      for file in ${var.strimzi_yaml_source_path}/04*-Crd-*.yaml; do
        if [ -f "$file" ]; then
          kubectl apply -f "$file"
        fi
      done

      # 2. Apply ClusterRoles (cluster-scoped, no namespace)
      echo "Applying ClusterRoles..."
      for file in ${var.strimzi_yaml_source_path}/02*-ClusterRole-*.yaml ${var.strimzi_yaml_source_path}/03*-ClusterRole-*.yaml; do
        if [ -f "$file" ]; then
          kubectl apply -f "$file"
        fi
      done

      # 3. Apply namespace-scoped resources
      echo "Applying namespace-scoped resources..."

      # ServiceAccount (010)
      if [ -f "${var.strimzi_yaml_source_path}/010-ServiceAccount-strimzi-cluster-operator.yaml" ]; then
        kubectl apply -f "${var.strimzi_yaml_source_path}/010-ServiceAccount-strimzi-cluster-operator.yaml" -n ${var.operator_namespace}
      fi

      # ClusterRoleBindings need special handling - they reference the namespace
      echo "Applying ClusterRoleBindings..."
      for file in ${var.strimzi_yaml_source_path}/021-ClusterRoleBinding-*.yaml ${var.strimzi_yaml_source_path}/030-ClusterRoleBinding-*.yaml ${var.strimzi_yaml_source_path}/033-ClusterRoleBinding-*.yaml; do
        if [ -f "$file" ]; then
          # Apply and update the subject namespace
          kubectl apply -f "$file"
          filename=$(basename "$file")
          binding_name=$(echo "$filename" | sed 's/.*-ClusterRoleBinding-//' | sed 's/.yaml//')
          kubectl patch clusterrolebinding "$binding_name" --type='json' \
            -p='[{"op": "replace", "path": "/subjects/0/namespace", "value": "'${var.operator_namespace}'"}]' || true
        fi
      done

      # RoleBindings (namespace-scoped)
      for file in ${var.strimzi_yaml_source_path}/020-RoleBinding-*.yaml ${var.strimzi_yaml_source_path}/022-RoleBinding-*.yaml ${var.strimzi_yaml_source_path}/023-RoleBinding-*.yaml ${var.strimzi_yaml_source_path}/031-RoleBinding-*.yaml; do
        if [ -f "$file" ]; then
          kubectl apply -f "$file" -n ${var.operator_namespace}
        fi
      done

      # ConfigMap
      if [ -f "${var.strimzi_yaml_source_path}/050-ConfigMap-strimzi-cluster-operator.yaml" ]; then
        kubectl apply -f "${var.strimzi_yaml_source_path}/050-ConfigMap-strimzi-cluster-operator.yaml" -n ${var.operator_namespace}
      fi

      # Finally, apply the deployment
      if [ -f "${var.strimzi_yaml_source_path}/${var.operator_deployment_yaml_filename}" ]; then
        kubectl apply -f "${var.strimzi_yaml_source_path}/${var.operator_deployment_yaml_filename}" -n ${var.operator_namespace}
      fi

      echo "Strimzi operator resources applied successfully"
    EOT

    environment = {
      KUBECONFIG = var.cluster_kubeconfig_path != "" ? var.cluster_kubeconfig_path : null
    }
  }
}

# Add this after your existing apply_strimzi_operator resource

resource "null_resource" "patch_strimzi_namespaces" {
  depends_on = [null_resource.apply_strimzi_operator]

  triggers = {
    watched_namespaces = join(",", var.watched_namespaces_list)
  }

  provisioner "local-exec" {
    command = <<-EOT
      echo "Patching Strimzi operator to watch namespaces: ${join(",", var.watched_namespaces_list)}"

      # Patch the deployment to set the correct watched namespaces
      kubectl patch deployment strimzi-cluster-operator -n ${var.operator_namespace} --type='json' \
        -p='[
          {
            "op": "replace",
            "path": "/spec/template/spec/containers/0/env",
            "value": [
              {
                "name": "STRIMZI_NAMESPACE",
                "value": "${join(",", var.watched_namespaces_list)}"
              },
              {
                "name": "STRIMZI_FULL_RECONCILIATION_INTERVAL_MS",
                "value": "120000"
              },
              {
                "name": "STRIMZI_OPERATION_TIMEOUT_MS",
                "value": "300000"
              },
              {
                "name": "STRIMZI_DEFAULT_KAFKA_EXPORTER_IMAGE",
                "value": "quay.io/strimzi/kafka:0.47.0-kafka-4.0.0"
              },
              {
                "name": "STRIMZI_DEFAULT_CRUISE_CONTROL_IMAGE",
                "value": "quay.io/strimzi/kafka:0.47.0-kafka-4.0.0"
              },
              {
                "name": "STRIMZI_KAFKA_IMAGES",
                "value": "3.9.0=quay.io/strimzi/kafka:0.47.0-kafka-3.9.0\n3.9.1=quay.io/strimzi/kafka:0.47.0-kafka-3.9.1\n4.0.0=quay.io/strimzi/kafka:0.47.0-kafka-4.0.0"
              },
              {
                "name": "STRIMZI_KAFKA_CONNECT_IMAGES",
                "value": "3.9.0=quay.io/strimzi/kafka:0.47.0-kafka-3.9.0\n3.9.1=quay.io/strimzi/kafka:0.47.0-kafka-3.9.1\n4.0.0=quay.io/strimzi/kafka:0.47.0-kafka-4.0.0"
              },
              {
                "name": "STRIMZI_KAFKA_MIRROR_MAKER_2_IMAGES",
                "value": "3.9.0=quay.io/strimzi/kafka:0.47.0-kafka-3.9.0\n3.9.1=quay.io/strimzi/kafka:0.47.0-kafka-3.9.1\n4.0.0=quay.io/strimzi/kafka:0.47.0-kafka-4.0.0"
              },
              {
                "name": "STRIMZI_DEFAULT_TOPIC_OPERATOR_IMAGE",
                "value": "quay.io/strimzi/operator:0.47.0"
              },
              {
                "name": "STRIMZI_DEFAULT_USER_OPERATOR_IMAGE",
                "value": "quay.io/strimzi/operator:0.47.0"
              },
              {
                "name": "STRIMZI_DEFAULT_KAFKA_INIT_IMAGE",
                "value": "quay.io/strimzi/operator:0.47.0"
              },
              {
                "name": "STRIMZI_DEFAULT_KAFKA_BRIDGE_IMAGE",
                "value": "quay.io/strimzi/kafka-bridge:0.32.0"
              },
              {
                "name": "STRIMZI_DEFAULT_KANIKO_EXECUTOR_IMAGE",
                "value": "quay.io/strimzi/kaniko-executor:0.47.0"
              },
              {
                "name": "STRIMZI_DEFAULT_MAVEN_BUILDER",
                "value": "quay.io/strimzi/maven-builder:0.47.0"
              },
              {
                "name": "STRIMZI_OPERATOR_NAMESPACE",
                "value": "${join(",", var.watched_namespaces_list)}"
              },
              {
                "name": "STRIMZI_FEATURE_GATES",
                "value": ""
              },
              {
                "name": "STRIMZI_LEADER_ELECTION_ENABLED",
                "value": "true"
              },
              {
                "name": "STRIMZI_LEADER_ELECTION_LEASE_NAME",
                "value": "strimzi-cluster-operator"
              },
              {
                "name": "STRIMZI_LEADER_ELECTION_LEASE_NAMESPACE",
                "valueFrom": {
                  "fieldRef": {
                    "fieldPath": "metadata.namespace"
                  }
                }
              },
              {
                "name": "STRIMZI_LEADER_ELECTION_IDENTITY",
                "valueFrom": {
                  "fieldRef": {
                    "fieldPath": "metadata.name"
                  }
                }
              }
            ]
          }
        ]'

      # Wait for rollout to complete
      kubectl rollout status deployment/strimzi-cluster-operator -n ${var.operator_namespace} --timeout=300s

      echo "Strimzi operator patched successfully"
    EOT

    environment = {
      KUBECONFIG = var.cluster_kubeconfig_path != "" ? var.cluster_kubeconfig_path : null
    }
  }
}