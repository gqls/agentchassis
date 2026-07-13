# deployments/terraform/environments/production/uk001/services/agents/2220-remote-job-spawner/main.tf
terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-remote-job-spawner"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

provider "kubernetes" {
  config_path    = "/home/ant/.kube/config_production_uk001"
}

module "remote_job_spawner_deployment" {
  source = "../../../../../../modules/kustomize-apply"

  kustomize_path   = "../../../../../deployments/kustomize/services/remote-job-spawner/overlays/production/uk_001"
  image_repository = ""
  namespace        = ""
  service_name     = ""
}
