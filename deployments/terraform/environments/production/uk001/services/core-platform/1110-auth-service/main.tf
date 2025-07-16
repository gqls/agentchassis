# Create the ConfigMap for the auth-service, populating it from the root configs file.
resource "kubernetes_config_map_v1" "auth_service_config" {
  metadata {
    name      = "auth-service-config"
    namespace = "personae-system"
  }
  data = {
    "auth-service.yaml" = file("../../../../../../../configs/auth-service.yaml")
  }
}

# Create the secret for the auth-service.
resource "kubernetes_secret_v1" "auth_service_secrets" {
  metadata {
    name      = "auth-service-secrets"
    namespace = "personae-system"
  }
  data = {
    "AUTH_DB_PASSWORD"    = var.db_password
    "JWT_SECRET_KEY" = var.jwt_secret
  }
}

# Deploy the application using the generic kustomize-apply module.
module "auth_service_app" {
  source = "../../../../../../modules/kustomize-apply"

  service_name     = "auth-service"
  namespace        = "personae-system"
  image_repository = "aqls/personae-auth-service" # Your image repo
  image_tag        = var.image_tag

  # Point to the production Kustomize overlay for this service.
  kustomize_path = "../../../../../../../deployments/kustomize/services/auth-service/overlays/production"

  # Trigger a redeploy if the config file changes.
  config_sha = filesha1("../../../../../../../configs/auth-service.yaml")

  depends_on = [
    kubernetes_config_map_v1.auth_service_config,
    kubernetes_secret_v1.auth_service_secrets
  ]
}
