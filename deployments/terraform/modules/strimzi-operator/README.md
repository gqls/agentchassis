The file [060-Deployment-strimzi-cluster-operator.yaml](strimzi-0.47.0/060-Deployment-strimzi-cluster-operator.yaml)
was altered to add the namespaces that we want strimzi kafka to watch
s/b value: "kafka,personae,strimzi"
(not valueFrom: fieldRef: fieldPath: metadata.namespace)

all myproject namespaces in yamls have to be sed replaced or find/replaced to strimzi

added the clusterrolebinding added-clusterrolebinding-operator-watched.yaml in config dir

github of strimzi files is:
https://github.com/strimzi/strimzi-kafka-operator

--

# From the 030-strimzi-operator directory
ls -la ../../../../modules/strimzi-operator/strimzi-0.47.0/

# First, check if the namespace exists
kubectl get ns strimzi

# If not, create it
kubectl create ns strimzi

# Apply the Strimzi operator YAMLs manually
kubectl apply -f ~/projects/agent-chassis/deployments/terraform/modules/strimzi-operator/strimzi-0.47.0/ -n strimzi

--
4. Wait for Strimzi to be ready:

# Check if the operator is running
kubectl get pods -n strimzi

# Wait for it to be ready
kubectl wait --for=condition=available --timeout=300s deployment/strimzi-cluster-operator -n strimzi

# Verify CRDs were created
kubectl get crd | grep kafka

--
5. Once Strimzi is running, deploy the Kafka cluster:

# Go to Kafka cluster directory
cd ~/projects/agent-chassis/deployments/terraform/environments/development/uk_dev/040-kafka-cluster

# Apply the Kafka cluster
terraform apply -auto-approve

# Check if Kafka resources are created
kubectl get kafka -n kafka
kubectl get pods -n kafka


