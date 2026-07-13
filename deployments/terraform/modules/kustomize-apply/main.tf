resource "null_resource" "apply_kustomization" {
  triggers = {
    # This trigger helps re-apply when kustomize files change.
    config_sha_trigger = sha1(join("", [
      for f in fileset(var.kustomize_path, "**/*.yaml") : filesha1("${var.kustomize_path}/${f}")
    ]))
    image_tag_trigger = var.image_tag
  }

  provisioner "local-exec" {
    # This command block is now smarter.
    command = <<-EOT
      set -e
      echo "Applying Kustomize overlay at ${var.kustomize_path}"
      kubectl apply -k ${var.kustomize_path}

      # This 'if' block makes the image update conditional
      if [ -n "${var.deployment_name}" ]; then
        echo "Setting image for deployment/${var.deployment_name} to :${var.image_tag}"
        kubectl set image deployment/${var.deployment_name} app=${var.image_repository}:${var.image_tag} -n ${var.namespace}

        echo "Waiting for rollout of deployment/${var.deployment_name}..."
        kubectl rollout status deployment/${var.deployment_name} -n ${var.namespace} --timeout=5m
      fi
    EOT
  }
}