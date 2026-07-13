# error
kubernetes_manifest.kafka_cluster: Creating...
╷
│ Error: Cannot create resource that already exists
│
│   with kubernetes_manifest.kafka_cluster,
│   on main.tf line 1, in resource "kubernetes_manifest" "kafka_cluster":
│    1: resource "kubernetes_manifest" "kafka_cluster" {
│
│ resource "kafka/personae-kafka-cluster" already exists
╵
make: *** [makefile:225: deploy-040-kafka] Error 1


cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/040-kafka-cluster
terraform import kubernetes_manifest.kafka_cluster "apiVersion=kafka.strimzi.io/v1beta2,kind=Kafka,namespace=kafka,name=personae-kafka-cluster"

---

kubectl delete kafkanodepool combined-pool -n kafka
kubectl delete kafka personae-kafka-cluster -n kafka
