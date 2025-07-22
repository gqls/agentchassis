1. Check Kafka cluster status in another terminal:

# Check if the Kafka resource exists and its status
kubectl get kafka -n kafka personae-kafka-cluster

# Get detailed status
kubectl describe kafka -n kafka personae-kafka-cluster

# Watch the status in real-time
kubectl get kafka -n kafka personae-kafka-cluster -w

2. Check Kafka pods status:

# See all pods in kafka namespace
kubectl get pods -n kafka

# If pods are not ready, check their logs
kubectl logs -n kafka -l strimzi.io/cluster=personae-kafka-cluster --tail=50

# Check events in the namespace
kubectl get events -n kafka --sort-by='.lastTimestamp'

3. Check Strimzi operator logs for any errors:

# Check operator logs
kubectl logs -n strimzi deployment/strimzi-cluster-operator --tail=100 -f

# Or just recent errors
kubectl logs -n strimzi deployment/strimzi-cluster-operator --tail=100 | grep -i error

4. Common issues to look for:
   Storage issues:

# Check PVCs
kubectl get pvc -n kafka

Resource constraints:

# Check node resources
kubectl top nodes
kubectl describe nodes | grep -A 5 "Allocated resources"

Check the Kafka CR details:

# This will show the status conditions
kubectl get kafka personae-kafka-cluster -n kafka -o jsonpath='{.status.conditions}' | jq .

5. If it's taking too long, check the specific Kafka config being applied:

# Check which Kafka CR YAML is being used
cat deployments/terraform/modules/kafka-cluster/config/kafka-cluster-cr-dev.yaml

The most common issues are:

Storage class not available - Kind uses standard storage class by default
Insufficient resources - Kafka needs decent CPU/memory
Image pull issues - Check if the Kafka images can be pulled