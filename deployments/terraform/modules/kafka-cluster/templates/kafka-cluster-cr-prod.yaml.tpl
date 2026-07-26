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
        # bugs_open/040-kafka-dial — DOCUMENTED, NOT APPLIED. Read before enabling.
        #
        # Strimzi advertises each broker under its short headless name, verified
        # live 2026-07-26 from the broker's own /tmp/strimzi.properties:
        #   PLAIN-9092://personae-kafka-cluster-combined-pool-prod-0.\
        #     personae-kafka-cluster-kafka-brokers.kafka.svc:9092
        # That name has three dots. Pods run with ndots:5 and a three-domain
        # search path, so the resolver tries the search suffixes FIRST and only
        # reaches the working name on its fourth attempt — three NXDOMAIN round
        # trips per lookup, each doubled by the parallel AAAA query. Measured
        # fleet-wide over 24h: 384,392 NXDOMAIN of 525,152 responses (73%), with
        # A and AAAA at exactly 1:1, i.e. about 7.5 queries per useful answer.
        #
        # Overriding advertisedHost with the fully-qualified name (ending
        # .svc.cluster.local, five dots, therefore tried as absolute first)
        # collapses that to a single query.
        #
        # Two warnings, both load-bearing:
        #  * Do NOT instead lower ndots. It looks like the cheap fix and it is
        #    strictly worse: the name genuinely needs cluster.local appended, so
        #    at ndots:2 the resolver tries it absolute (NXDOMAIN) and THEN still
        #    walks all three search domains — four rounds instead of three.
        #  * This is a volume reduction, not a proven cure. Measured directly:
        #    300 short-name lookups and 300 FQDN lookups both completed in under
        #    a second, so with healthy DNS the search path costs no meaningful
        #    latency. It removes packets, and therefore exposure to loss; it has
        #    NOT been shown to fix the reported timeouts.
        #
        # Applying this rewrites advertised.listeners and rolls every broker.
        # configuration:
        #   brokers:
        #     - broker: 0
        #       advertisedHost: personae-kafka-cluster-combined-pool-prod-0.personae-kafka-cluster-kafka-brokers.kafka.svc.cluster.local
        #     - broker: 1
        #       advertisedHost: personae-kafka-cluster-combined-pool-prod-1.personae-kafka-cluster-kafka-brokers.kafka.svc.cluster.local
        #     - broker: 2
        #       advertisedHost: personae-kafka-cluster-combined-pool-prod-2.personae-kafka-cluster-kafka-brokers.kafka.svc.cluster.local
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