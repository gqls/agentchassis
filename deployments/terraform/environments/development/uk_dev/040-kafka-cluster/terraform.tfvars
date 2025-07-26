# Kafka cluster configuration for development
use_node_pools      = true
node_pool_name      = "combined-pool"
node_pool_replicas  = 1
node_pool_storage_size = "10Gi"
kafka_version       = "3.9.1"
replication_factor  = 1
min_insync_replicas = 1
storage_class       = "standard"
delete_claim        = true