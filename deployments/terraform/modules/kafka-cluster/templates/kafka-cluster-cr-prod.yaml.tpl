apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: ${cluster_name}
  namespace: ${namespace}
  annotations:
    strimzi.io/node-pools: ${use_node_pools ? "enabled" : "disabled"}
    strimzi.io/kraft: enabled
spec:
  kafka:
    version: ${kafka_version}
    %{~ if !use_node_pools ~}
    replicas: ${replicas}
    %{~ endif ~}
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
      - name: tls
        port: 9093
        type: internal
        tls: true
    config:
      offsets.topic.replication.factor: ${replication_factor}
      transaction.state.log.replication.factor: ${replication_factor}
      transaction.state.log.min.isr: ${min_insync_replicas}
      default.replication.factor: ${replication_factor}
      min.insync.replicas: ${min_insync_replicas}
    %{~ if !use_node_pools ~}
    storage:
      type: persistent-claim
      size: ${storage_size}
      class: ${storage_class}
      deleteClaim: ${delete_claim}
    %{~ endif ~}
  entityOperator:
    topicOperator:
      watchedNamespace: ${namespace}
      reconciliationIntervalMs: 90000
    userOperator:
      watchedNamespace: ${namespace}
      reconciliationIntervalMs: 120000