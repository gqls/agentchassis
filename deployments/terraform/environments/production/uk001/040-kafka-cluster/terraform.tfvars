# Kafka cluster configuration for production
use_node_pools      = true
node_pool_name      = "combined-pool-prod"
node_pool_replicas  = 3
node_pool_storage_size = "50Gi"
kafka_version       = "3.9.0"
replication_factor  = 3
min_insync_replicas = 2
storage_class       = "ssd-large"
delete_claim        = false

kubeconfig_path = "/home/ant/.kube/config_production_uk001"