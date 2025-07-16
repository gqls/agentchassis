# ~/projects/terraform/rackspace_generic/terraform/environments/production/sydney/040-kafka-cluster/main.tf

module "kafka_cluster_service" {
  source = "../../../../modules/kafka_cluster" # Path to your reusable module

  kafka_cr_namespace      = var.target_kafka_namespace
  kafka_cr_yaml_file_path = var.kafka_cluster_cr_yaml_path_uk001
  kubeconfig_path = var.kubeconfig_path
  kafka_cr_cluster_name   = var.kafka_cluster_name_uk001
}
