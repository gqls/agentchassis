resource "null_resource" "kind_cluster" {
  triggers = {
    cluster_name = var.kind_cluster_name
    config_path  = var.kind_config_path
    node_image   = var.kind_node_image
  }

  provisioner "local-exec" {
    when    = create
    command = <<-EOT
      set -e
      if ! kind get clusters | grep -q "^${self.triggers.cluster_name}$$"; then
        echo "Creating Kind cluster '${self.triggers.cluster_name}'..."
        kind create cluster --name "${self.triggers.cluster_name}" --image "${self.triggers.node_image}" ${var.kind_config_path != null ? "--config \"${var.kind_config_path}\"" : ""}
        echo "Waiting for Kind cluster control plane to be ready..."
        timeout 120s bash -c 'while ! kubectl --context="kind-${self.triggers.cluster_name}" cluster-info >/dev/null 2>&1; do sleep 1; done' || \
          (echo "Timeout waiting for Kind cluster. Check 'kind get logs ${self.triggers.cluster_name}'" && exit 1)
        echo "Kind cluster '${self.triggers.cluster_name}' is ready."
      else
        echo "Kind cluster '${self.triggers.cluster_name}' already exists. Skipping creation."
      fi
    EOT
  }

  provisioner "local-exec" {
    when    = destroy
    # Use self.triggers.cluster_name which is known at destroy time based on the state
    command = "kind delete cluster --name \"${self.triggers.cluster_name}\" || true"
  }
}

resource "null_resource" "label_kind_node" {
  depends_on = [null_resource.kind_cluster]

  provisioner "local-exec" {
    command = <<-EOT
      # Wait for the node to be ready
      kubectl wait --for=condition=ready node --all --timeout=60s

      # Label the node
      kubectl label nodes --all role=spot-instance --overwrite
    EOT
  }
}