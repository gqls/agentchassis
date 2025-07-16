resource "null_resource" "apply_kustomization" {
  triggers = {
    image_tag_trigger = var.image_tag
    config_sha_trigger = var.config_sha
  }

  provisioner "local-exec" {
    command = <<-EOT
      set -e
      echo "Applying Kustomize overlay at ${var.kustomize_path}"
      kubectl apply -k ${var.kustomize_path}

      echo "Setting image for deployment/${var.service_name} to ${var.image_repository}:${var.image_tag}"
      kubectl set image deployment/${var.service_name} ${var.service_name}=${var.image_repository}:${var.image_tag} -n ${var.namespace}

      echo "Waiting for rollout of deployment/${var.service_name}..."
      kubectl rollout status deployment/${var.service_name} -n ${var.namespace} --timeout=5m
    EOT
  }
}
