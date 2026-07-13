apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaNodePool
metadata:
  name: ${pool_name}
  namespace: ${namespace}
  labels:
    strimzi.io/cluster: ${cluster_name}
spec:
  replicas: ${replicas}
  roles:
    - broker
    - controller
  storage:
    type: persistent-claim
    size: ${storage_size}
    class: ${storage_class}
    deleteClaim: ${delete_claim}