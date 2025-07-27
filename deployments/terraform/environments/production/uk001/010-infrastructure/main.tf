module "kubernetes_cluster" {
  source = "../../../../modules/rackspace-kubernetes"

  # Basic cluster configuration
  cluster_name           = var.instance_cluster_name
  rackspace_region       = var.instance_rackspace_region
  preemption_webhook_url = var.instance_slack_webhook_url
  kubernetes_version     = var.instance_kubernetes_version

  # Node pool configurations (using the new structure)
  spot_node_pools = {
    "spot_worker_pool" = {
      min_nodes = var.instance_spot_min_nodes
      max_nodes = var.instance_spot_max_nodes
      flavor    = var.instance_spot_node_flavor
      max_price = var.instance_spot_max_price
      labels = {
        "role"       = "spot-instance"
        "app.type"   = "stateless"
        "managed-by" = "terraform"
      }
    }
  }

  ondemand_node_pools = {
    "default_pool" = {
      node_count = var.instance_ondemand_node_count
      flavor     = var.instance_ondemand_node_flavor
      labels = {
        "role"       = "general"
        "app.type"   = "stateful"
        "managed-by" = "terraform"
      }
      taints = var.instance_ondemand_node_taints
    }
  }
}