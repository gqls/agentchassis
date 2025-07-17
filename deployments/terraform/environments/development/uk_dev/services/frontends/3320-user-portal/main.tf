# This instance deploys the main user-facing React application.
module "user_frontend_app" {
  source = "../../../../../../modules/kustomize-apply"

  service_name     = "user-frontend"
  namespace        = "personae-system"
  image_repository = "aqls/personae-web-interface" # Your frontend image
  image_tag        = var.image_tag

  # Point to the production Kustomize overlay for the frontend.
  # This directory would contain the deployment.yaml, service.yaml, ingress.yaml, etc.
  kustomize_path = "../../../../../../../deployments/kustomize/frontends/user-frontend/overlays/production"
}
