when trying to deploy using a kubernetes_manifest for this kafka CRD it continually bugged out when setting config variables, so we are doing it using null_resource kubectl apply ...

this didn't work:
resource "kubernetes_manifest" "kafka_cluster" {
manifest = {
"apiVersion" = "kafka.strimzi.io/v1beta2"
"kind"       = "Kafka"
"metadata" = {
"name"      = var.kafka_cluster_name
"namespace" = var.kafka_cluster_namespace
}
"spec" = {
"kafka" = {
"version"  = var.kafka_version
"replicas" = var.kafka_replicas
"listeners" = [ # Minimal listeners
{ "name": "plain", "port": 9092, "type": "internal", "tls": false },
{ "name" = "tls", "port" = 9093, "type" = "internal", "tls"  = true }
]
"storage" = merge( # Assuming you kept the merge logic for class
{
"type" = "persistent-claim"
"size" = var.kafka_persistent_claim_size
},
var.kafka_persistent_claim_storage_class == null ? {} : { "class" = var.kafka_persistent_claim_storage_class
}
),
"config" = var.kafka_config
}
"entityOperator" = {
"topicOperator" = var.enable_topic_operator ? {} : null
"userOperator"  = var.enable_user_operator ? {} : null
}
}
}
computed_fields = [
"spec.kafka.config"
]
}

because of the "config" = va.kafka_config didn't read the map properly:
variable "kafka_config" {
description = "List of Kafka broker configuration overrides."
type        = map(string)
default = {
"log.message.format.version" = "4.0",
"log.retention.hours"        = "168"
}
}

I got:
terraform apply -auto-approve
kubernetes_manifest.kafka_cluster: Refreshing state...
╷
│ Error: Failed to update proposed state from prior state
│
│   with kubernetes_manifest.kafka_cluster,
│   on main.tf line 3, in resource "kubernetes_manifest" "kafka_cluster":
│    3: resource "kubernetes_manifest" "kafka_cluster" {
│
│ AttributeName("config"): can't use tftypes.Object["log.message.format.version":tftypes.String, "log.retention.hours":tftypes.String] as
│ tftypes.Map[tftypes.String]
╵
