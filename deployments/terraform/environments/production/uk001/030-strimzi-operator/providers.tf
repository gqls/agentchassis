# terraform/environments/production/uk001/030-strimzi-operator/providers.tf

provider "kubernetes" {
  config_path = var.kubeconfig_path # Provided by Makefile or root tfvars
}
# No Helm provider needed if this component only uses kubectl apply via null_resourceer "helm" { kubernetes {} }