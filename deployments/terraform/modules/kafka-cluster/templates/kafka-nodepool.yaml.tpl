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
  # bugs_open/040-kafka-dial. This template had NO resources block, so the
  # rendered production brokers run with `resources: {}` — verified live
  # 2026-07-26 on personae-kafka-cluster-combined-pool-prod-0. Two consequences:
  #
  #  1. The brokers have no scheduling guarantee at all. Under node pressure
  #     nothing reserves CPU or memory for them, and they are ranked for
  #     eviction alongside best-effort workloads. Intermittent inability to
  #     accept connections is exactly the symptom that produces.
  #  2. Strimzi derives the JVM heap from the memory LIMIT. With no limit the
  #     container ships `KAFKA_HEAP_OPTS=-Xms128M` and no `-Xmx` at all
  #     (verified live), so max heap falls back to the JVM default of about a
  #     quarter of the node's RAM — roughly 15GB on these 60GB nodes. Measured
  #     RSS is 4.8GB while idle, so this is latent rather than active, but it
  #     is unbounded and GC behaviour is uncontrolled.
  #
  # The figures below are NOT invented: they are the ones already sitting in
  # this module's own config/kafka-nodepool-cr-prod.yaml, a hand-written CR
  # that the production environment does not reference (variables.tf points at
  # this template instead). Adopting them makes the unused file's intent real.
  #
  # NOT YET APPLIED to the live cluster — applying this rolls all three brokers
  # one at a time. See RUNBOOK_040_kafka_dial.md for the exact kubectl patch and
  # the pre-flight checks. The checked-in terraform state for 040-kafka-cluster
  # is empty (serial 1, zero resources), so `terraform apply` is NOT the safe
  # route here.
  resources:
    requests:
      memory: "2Gi"
      cpu: "1"
    limits:
      memory: "4Gi"
      cpu: "2"
  jvmOptions:
    # Set explicitly rather than left to Strimzi's limit-derived default, so
    # the heap does not silently change if the memory limit above is retuned.
    -Xms: "2G"
    -Xmx: "2G"