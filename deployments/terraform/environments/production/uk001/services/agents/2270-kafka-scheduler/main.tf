# deployments/terraform/environments/production/uk001/services/agents/2270-kafka-scheduler/main.tf
terraform {
  backend "kubernetes" {
    secret_suffix = "tfstate-svc-kafka-scheduler"
    config_path   = "/home/ant/.kube/config_production_uk001"
  }
}

provider "kubernetes" {
  config_path    = "/home/ant/.kube/config_production_uk001"
}

module "kafka_scheduler_deployment" {
  source = "../../../../../../modules/kustomize-apply"

  # Path to the production overlay for this service
  kustomize_path   = "../../../../../deployments/kustomize/services/kafka-scheduler/overlays/production/uk_001"
  image_repository = ""
  namespace        = ""
  service_name     = ""
}
