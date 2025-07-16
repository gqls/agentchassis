# This module applies the Kafka cluster custom resource YAML.
resource "null_resource" "apply_kafka_cluster_cr" {
  triggers = {
    yaml_file_sha1 = fileexists(var.kafka_cr_yaml_file_path) ? filesha1(var.kafka_cr_yaml_file_path) : ""
  }

  provisioner "local-exec" {
    command = "kubectl apply --namespace ${var.kafka_cr_namespace} -f ${var.kafka_cr_yaml_file_path}"
  }
}
