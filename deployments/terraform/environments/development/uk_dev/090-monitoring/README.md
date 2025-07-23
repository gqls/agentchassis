there is a password for grafana hardcoded

To apply:
cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/090-monitoring
terraform init -upgrade
terraform apply -auto-approve

# Port forward Grafana
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80

# Access at http://localhost:3000
# Username: admin
# Password: admin or YOUR_SECURE_DEV_GRAFANA_PASSWORD


---


Your Kafka infrastructure is now fully deployed with monitoring. Here's a summary of what you have running:
✅ Successfully Deployed:

Kind Kubernetes Cluster (kind-personae-dev)
Ingress Controller (NGINX)
Strimzi Kafka Operator
Kafka Cluster

Single node Kafka with ZooKeeper
Bootstrap server: personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092


Kafka Users

core-manager-user - for topic management
personae-app-anonymous - for applications


Kafka Topics (system topics created)
Monitoring Stack (Prometheus + Grafana)

Grafana accessible at http://localhost:3000 (when port-forwarded)
Username: admin, Password: admin



🔧 Quick Commands:
Check Kafka Status:

kubectl get kafka -n kafka
kubectl get kafkatopics -n kafka
kubectl get pods -n kafka

Test Kafka Connectivity:

# List topics
kubectl run -it --rm kafka-test --image=strimzi/kafka:latest-kafka-3.9.0 --restart=Never -- \
bin/kafka-topics.sh --bootstrap-server personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092 --list

# Produce a test message
kubectl run -it --rm kafka-producer --image=strimzi/kafka:latest-kafka-3.9.0 --restart=Never -- \
bin/kafka-console-producer.sh --bootstrap-server personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092 \
--topic system.events

Access Grafana:
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
# Then visit http://localhost:3000

View Kafka Metrics in Grafana:
Go to Dashboards → Browse
Look for Kafka-related dashboards (if kafka-exporter was deployed)
Or import Kafka dashboard ID: 7589 from Grafana.com


