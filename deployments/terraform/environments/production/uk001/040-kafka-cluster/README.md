# Get the latest logs from the Strimzi operator
kubectl logs -n strimzi deployment/strimzi-cluster-operator --tail=100 -f
kubectl get kafka personae-kafka-cluster -n kafka -o yaml
kubectl top nodes

When the operator logs are silent, the best place to find the error is in the status field of the custom resources themselves.

kubectl get kafka personae-kafka-cluster -n kafka -o yaml
kubectl get kafkanodepool combined-pool-prod -n kafka -o yaml
