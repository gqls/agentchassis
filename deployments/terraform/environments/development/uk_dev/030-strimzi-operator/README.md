# Check events for the ReplicaSet
kubectl describe replicaset strimzi-cluster-operator-98f497d8c -n strimzi

# Check all events in the namespace
kubectl get events -n strimzi --sort-by='.lastTimestamp'

# Verify ServiceAccount exists
kubectl get serviceaccount -n strimzi

# Check if there are any admission webhooks or policies blocking pod creation
kubectl get validatingwebhookconfigurations
kubectl get mutatingwebhookconfigurations
kubectl get podsecuritypolicies

Since the ReplicaSet shows 0 current pods, it's failing to create pods. The most common causes are:

ServiceAccount still missing or misconfigured
Resource quotas or limits
Pod Security Policies
Node issues

# Check if the namespace has any resource quotas
kubectl get resourcequota -n strimzi

# Check if there are any network policies
kubectl get networkpolicies -n strimzi

# Try creating a simple test pod to see if pod creation works at all
kubectl run test-pod --image=busybox -n strimzi --command -- sleep 3600

# If that works, delete it
kubectl delete pod test-pod -n strimzi

The fact that the ReplicaSet exists but has 0 current pods 
strongly suggests the ServiceAccount issue persists 
or there's another authorization/admission control problem.

--

cd ~/projects/agent-chassis/deployments/terraform/modules/strimzi-operator/strimzi-0.47.0/

# Apply ServiceAccount
kubectl apply -f 010-ServiceAccount-strimzi-cluster-operator.yaml -n strimzi

# Apply ClusterRoles (no namespace needed)
kubectl apply -f 020-ClusterRole-strimzi-cluster-operator-role.yaml
kubectl apply -f 021-ClusterRole-strimzi-cluster-operator-role.yaml
kubectl apply -f 022-ClusterRole-strimzi-cluster-operator-role.yaml
kubectl apply -f 023-ClusterRole-strimzi-cluster-operator-role.yaml
kubectl apply -f 030-ClusterRole-strimzi-kafka-broker.yaml
kubectl apply -f 031-ClusterRole-strimzi-entity-operator.yaml
kubectl apply -f 033-ClusterRole-strimzi-kafka-client.yaml

# Apply RoleBindings (with namespace)
kubectl apply -f 020-RoleBinding-strimzi-cluster-operator.yaml -n strimzi
kubectl apply -f 022-RoleBinding-strimzi-cluster-operator.yaml -n strimzi
kubectl apply -f 023-RoleBinding-strimzi-cluster-operator.yaml -n strimzi

# Apply ClusterRoleBindings (no namespace, but may need to patch the subject namespace)
kubectl apply -f 021-ClusterRoleBinding-strimzi-cluster-operator.yaml
kubectl apply -f 030-ClusterRoleBinding-strimzi-cluster-operator-kafka-broker-delegation.yaml
kubectl apply -f 031-RoleBinding-strimzi-cluster-operator-entity-operator-delegation.yaml -n strimzi
kubectl apply -f 033-ClusterRoleBinding-strimzi-cluster-operator-kafka-client-delegation.yaml

# ConfigMap
kubectl apply -f 050-ConfigMap-strimzi-cluster-operator.yaml -n strimzi

# Check pods again
kubectl get pods -n strimzi