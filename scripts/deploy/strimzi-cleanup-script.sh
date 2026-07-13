#!/bin/bash
# cleanup-strimzi.sh

echo "Starting complete Strimzi cleanup..."

# Set your kubeconfig
export KUBECONFIG=/home/ant/.kube/config_production_uk001

# 1. Delete all Kafka resources first
echo "Deleting Kafka resources..."
kubectl delete kafka --all --all-namespaces --wait=false
kubectl delete kafkauser --all --all-namespaces --wait=false
kubectl delete kafkatopic --all --all-namespaces --wait=false
kubectl delete kafkanodepool --all --all-namespaces --wait=false
kubectl delete kafkaconnect --all --all-namespaces --wait=false
kubectl delete kafkaconnector --all --all-namespaces --wait=false
kubectl delete kafkamirrormaker2 --all --all-namespaces --wait=false
kubectl delete kafkabridge --all --all-namespaces --wait=false
kubectl delete kafkarebalance --all --all-namespaces --wait=false

# 2. Delete Strimzi operator deployment and related resources
echo "Deleting Strimzi operator..."
kubectl delete deployment strimzi-cluster-operator -n strimzi --ignore-not-found=true
kubectl delete serviceaccount strimzi-cluster-operator -n strimzi --ignore-not-found=true
kubectl delete configmap strimzi-cluster-operator -n strimzi --ignore-not-found=true

# 3. Delete all RoleBindings in watched namespaces
echo "Deleting RoleBindings..."
for namespace in strimzi kafka personae; do
  kubectl delete rolebinding -n $namespace -l app=strimzi --ignore-not-found=true
  kubectl delete rolebinding strimzi-cluster-operator -n $namespace --ignore-not-found=true
  kubectl delete rolebinding strimzi-entity-operator -n $namespace --ignore-not-found=true
  kubectl delete rolebinding strimzi-cluster-operator-watched -n $namespace --ignore-not-found=true
  kubectl delete rolebinding strimzi-cluster-operator-entity-operator-delegation -n $namespace --ignore-not-found=true
  kubectl delete rolebinding strimzi-cluster-operator-leader-election -n $namespace --ignore-not-found=true
  kubectl delete rolebinding strimzi-cluster-operator-kafka-namespace-permissions -n $namespace --ignore-not-found=true
  kubectl delete rolebinding strimzi-cluster-operator-personae-namespace-permissions -n $namespace --ignore-not-found=true
done

# 4. Delete cluster-wide resources
echo "Deleting ClusterRoles and ClusterRoleBindings..."
kubectl delete clusterrole -l app=strimzi --ignore-not-found=true
kubectl delete clusterrolebinding -l app=strimzi --ignore-not-found=true
kubectl delete clusterrole strimzi-cluster-operator-namespaced --ignore-not-found=true
kubectl delete clusterrole strimzi-cluster-operator-global --ignore-not-found=true
kubectl delete clusterrole strimzi-cluster-operator-watched --ignore-not-found=true
kubectl delete clusterrole strimzi-entity-operator --ignore-not-found=true
kubectl delete clusterrole strimzi-kafka-broker --ignore-not-found=true
kubectl delete clusterrole strimzi-kafka-client --ignore-not-found=true
kubectl delete clusterrole strimzi-cluster-operator-leader-election --ignore-not-found=true
kubectl delete clusterrolebinding strimzi-cluster-operator --ignore-not-found=true
kubectl delete clusterrolebinding strimzi-cluster-operator-kafka-broker-delegation --ignore-not-found=true
kubectl delete clusterrolebinding strimzi-cluster-operator-kafka-client-delegation --ignore-not-found=true

# 5. Delete CRDs
echo "Deleting CRDs..."
kubectl delete crd kafkas.kafka.strimzi.io --ignore-not-found=true
kubectl delete crd kafkausers.kafka.strimzi.io --ignore-not-found=true
kubectl delete crd kafkatopics.kafka.strimzi.io --ignore-not-found=true
kubectl delete crd kafkanodepools.kafka.strimzi.io --ignore-not-found=true
kubectl delete crd kafkaconnects.kafka.strimzi.io --ignore-not-found=true
kubectl delete crd kafkaconnectors.kafka.strimzi.io --ignore-not-found=true
kubectl delete crd kafkamirrormaker2s.kafka.strimzi.io --ignore-not-found=true
kubectl delete crd kafkabridges.kafka.strimzi.io --ignore-not-found=true
kubectl delete crd kafkarebalances.kafka.strimzi.io --ignore-not-found=true
kubectl delete crd strimzipodsets.core.strimzi.io --ignore-not-found=true

# 6. Delete any remaining pods
echo "Cleaning up remaining pods..."
kubectl delete pods -n kafka --all --force --grace-period=0 2>/dev/null || true
kubectl delete pods -n strimzi --all --force --grace-period=0 2>/dev/null || true

# 7. Delete PVCs if any
echo "Cleaning up PVCs..."
kubectl delete pvc -n kafka --all --ignore-not-found=true

# 8. Optional: Delete namespaces (comment out if you want to keep them)
# echo "Deleting namespaces..."
# kubectl delete namespace kafka --ignore-not-found=true
# kubectl delete namespace personae --ignore-not-found=true
# kubectl delete namespace strimzi --ignore-not-found=true

echo "Cleanup complete!"
echo ""
echo "Verifying cleanup..."
echo "Remaining Strimzi CRDs:"
kubectl get crd | grep strimzi || echo "None found"
echo ""
echo "Remaining pods in kafka namespace:"
kubectl get pods -n kafka 2>/dev/null || echo "Namespace doesn't exist"
echo ""
echo "Remaining pods in strimzi namespace:"
kubectl get pods -n strimzi 2>/dev/null || echo "Namespace doesn't exist"