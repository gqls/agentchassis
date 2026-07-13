# Wait for CRDs and Kafka cluster to be ready
resource "terraform_data" "wait_for_prerequisites" {
  provisioner "local-exec" {
    command = <<-EOT
      set -e
      echo "Checking for Strimzi operator..."

      # First check if the Strimzi operator is running
      if ! kubectl --kubeconfig=${abspath(pathexpand(var.kubeconfig_path))} --context=${var.kube_context_name} \
        get deployment -n strimzi strimzi-cluster-operator >/dev/null 2>&1; then
        echo "ERROR: Strimzi operator not found. Please ensure 030-strimzi-operator has been deployed."
        exit 1
      fi

      echo "Waiting for Strimzi operator to be ready..."
      kubectl --kubeconfig=${abspath(pathexpand(var.kubeconfig_path))} --context=${var.kube_context_name} \
        wait --for=condition=available --timeout=300s deployment/strimzi-cluster-operator -n strimzi

      # Wait a bit for the operator to create CRDs
      echo "Giving operator time to install CRDs..."
      sleep 10

      # Now wait for CRDs with a retry loop
      echo "Waiting for Strimzi CRDs to be established..."
      for i in {1..30}; do
        if kubectl --kubeconfig=${abspath(pathexpand(var.kubeconfig_path))} --context=${var.kube_context_name} \
          get crd kafkausers.kafka.strimzi.io >/dev/null 2>&1; then
          echo "CRDs found, waiting for them to be established..."
          kubectl --kubeconfig=${abspath(pathexpand(var.kubeconfig_path))} --context=${var.kube_context_name} \
            wait --for condition=established --timeout=300s crd/kafkausers.kafka.strimzi.io
          kubectl --kubeconfig=${abspath(pathexpand(var.kubeconfig_path))} --context=${var.kube_context_name} \
            wait --for condition=established --timeout=300s crd/kafkas.kafka.strimzi.io
          break
        else
          echo "CRDs not found yet, waiting... ($i/30)"
          sleep 5
        fi
      done

      # Verify CRDs exist
      if ! kubectl --kubeconfig=${abspath(pathexpand(var.kubeconfig_path))} --context=${var.kube_context_name} \
        get crd kafkausers.kafka.strimzi.io >/dev/null 2>&1; then
        echo "ERROR: Strimzi CRDs were not created after waiting. Check Strimzi operator logs."
        exit 1
      fi

      echo "Waiting for Kafka cluster to be ready..."
      kubectl --kubeconfig=${abspath(pathexpand(var.kubeconfig_path))} --context=${var.kube_context_name} \
        wait --namespace kafka kafka/personae-kafka-cluster --for=condition=Ready --timeout=300s
    EOT
  }
}

# Apply core-manager Kafka user
resource "terraform_data" "core_manager_kafka_user" {
  depends_on = [terraform_data.wait_for_prerequisites]

  input = {
    kubeconfig_path = abspath(pathexpand(var.kubeconfig_path))
    kube_context_name = var.kube_context_name
    user_name = "core-manager-user"
    namespace = "kafka"
    manifest = {
      apiVersion = "kafka.strimzi.io/v1beta2"
      kind       = "KafkaUser"
      metadata = {
        name      = "core-manager-user"
        namespace = "kafka"
        labels = {
          "strimzi.io/cluster" = "personae-kafka-cluster"
        }
      }
      spec = {
        authorization = {
          type = "simple"
          acls = [
            {
              resource = {
                type        = "topic"
                name        = "*"
                patternType = "literal"
              }
              operations = [
                "Create",
                "Describe",
                "Alter"
              ]
              host = "*"
            }
          ]
        }
      }
    }
  }

  provisioner "local-exec" {
    command = <<-EOT
      cat <<'EOF' | kubectl --kubeconfig=${self.input.kubeconfig_path} --context=${self.input.kube_context_name} apply -f -
${yamlencode(self.input.manifest)}
EOF
    EOT
  }

  provisioner "local-exec" {
    when = destroy
    command = <<-EOT
      kubectl --kubeconfig=${self.input.kubeconfig_path} --context=${self.input.kube_context_name} \
        delete kafkauser ${self.input.user_name} -n ${self.input.namespace} --ignore-not-found=true
    EOT
  }
}

# Apply personae-app-anonymous user
resource "terraform_data" "personae_app_anonymous_user" {
  depends_on = [terraform_data.wait_for_prerequisites]

  input = {
    kubeconfig_path = abspath(pathexpand(var.kubeconfig_path))
    kube_context_name = var.kube_context_name
    user_name = "personae-app-anonymous"
    namespace = "kafka"
    manifest = {
      apiVersion = "kafka.strimzi.io/v1beta2"
      kind       = "KafkaUser"
      metadata = {
        name      = "personae-app-anonymous"
        namespace = "kafka"
        labels = {
          "strimzi.io/cluster" = "personae-kafka-cluster"
        }
      }
      spec = {
        authorization = {
          type = "simple"
          acls = [
            # For init containers & tools to describe specific topics
            {
              resource = {
                type        = "topic"
                name        = "personae-core-requests"
                patternType = "literal"
              }
              operations = ["Describe"]
              host = "*"
            },
            # For init containers & tools to list/describe all topics
            {
              resource = {
                type        = "topic"
                name        = "*"
                patternType = "literal"
              }
              operations = ["Describe"]
              host = "*"
            },
            # For applications to Read/Write their topics
            {
              resource = {
                type        = "topic"
                name        = "personae-"
                patternType = "prefix"
              }
              operations = ["Read", "Write", "Create", "Describe"]
              host = "*"
            },
            # For applications and tools to describe the cluster
            {
              resource = {
                type = "cluster"
              }
              operations = ["Describe"]
              host = "*"
            },
            # For personae-core-manager consumer group
            {
              resource = {
                type        = "group"
                name        = "personae-core-manager"
                patternType = "literal"
              }
              operations = ["Read", "Describe"]
              host = "*"
            },
            # General describe for any other group
            {
              resource = {
                type        = "group"
                name        = "*"
                patternType = "literal"
              }
              operations = ["Describe"]
              host = "*"
            }
          ]
        }
      }
    }
  }

  provisioner "local-exec" {
    command = <<-EOT
      cat <<'EOF' | kubectl --kubeconfig=${self.input.kubeconfig_path} --context=${self.input.kube_context_name} apply -f -
${yamlencode(self.input.manifest)}
EOF
    EOT
  }

  provisioner "local-exec" {
    when = destroy
    command = <<-EOT
      kubectl --kubeconfig=${self.input.kubeconfig_path} --context=${self.input.kube_context_name} \
        delete kafkauser ${self.input.user_name} -n ${self.input.namespace} --ignore-not-found=true
    EOT
  }
}