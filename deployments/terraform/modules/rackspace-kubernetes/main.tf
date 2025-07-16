# Main resource to create the Spot cloudspace (Kubernetes cluster).
resource "spot_cloudspace" "cluster" {
  cloudspace_name    = var.cluster_name
  region             = var.rackspace_region
  kubernetes_version = var.kubernetes_version
  preemption_webhook = var.preemption_webhook_url
  wait_until_ready   = true
}

# Creates an on-demand node pool if the count is greater than 0.
resource "spot_ondemandnodepool" "ondemand_pool" {
  count = var.ondemand_node_count > 0 ? 1 : 0

  cloudspace_name      = spot_cloudspace.cluster.cloudspace_name
  server_class         = var.ondemand_node_flavor
  desired_server_count = var.ondemand_node_count
  depends_on           = [spot_cloudspace.cluster]
}

# Creates multiple spot node pools based on the input map.
resource "spot_spotnodepool" "spot_pools" {
  for_each = var.spot_node_pools

  cloudspace_name = spot_cloudspace.cluster.cloudspace_name
  server_class    = each.value.flavor
  bid_price       = each.value.max_price
  autoscaling = {
    min_nodes = each.value.min_nodes
    max_nodes = each.value.max_nodes
  }
  labels = {
    "role"       = "spot-instance"
    "pool-name"  = each.key
    "managed-by" = "terraform"
  }
  depends_on = [spot_cloudspace.cluster]
}

# Data source to retrieve the kubeconfig after the cluster is ready.
data "spot_kubeconfig" "cluster_kubeconfig" {
  cloudspace_name = spot_cloudspace.cluster.cloudspace_name
  depends_on = [
    spot_cloudspace.cluster,
    spot_spotnodepool.spot_pools
  ]
}

