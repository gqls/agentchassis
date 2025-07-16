data "spot_serverclasses" "available_flavors" {
  # This data source does not take a region argument.
  # It's here mostly for reference if you output it.
}

resource "spot_cloudspace" "cluster" {
  cloudspace_name      = var.cluster_name
  region               = var.rackspace_region
  hacontrol_plane      = var.hacontrol_plane
  preemption_webhook   = var.preemption_webhook_url # Using the module's input variable
  wait_until_ready     = true
  kubernetes_version   = var.kubernetes_version
  cni                  = var.cni
}

# +++ CREATES MULTIPLE, DYNAMIC SPOT POOLS +++
resource "spot_spotnodepool" "spot_pools" {
  for_each = var.spot_node_pools

  cloudspace_name = spot_cloudspace.cluster.cloudspace_name
  server_class    = each.value.flavor
  bid_price       = each.value.max_price
  autoscaling = {
    min_nodes = each.value.min_nodes
    max_nodes = each.value.max_nodes
  }
  labels     = each.value.labels
  depends_on = [spot_cloudspace.cluster]
}

# +++ CREATES MULTIPLE, DYNAMIC ON-DEMAND POOLS +++
resource "spot_ondemandnodepool" "ondemand_pools" {
  for_each = var.ondemand_node_pools

  cloudspace_name      = spot_cloudspace.cluster.cloudspace_name
  server_class         = each.value.flavor
  desired_server_count = each.value.node_count
  labels               = each.value.labels
  taints               = each.value.taints
  depends_on           = [spot_cloudspace.cluster]
}


data "spot_kubeconfig" "cluster_kubeconfig" {
  cloudspace_name = spot_cloudspace.cluster.cloudspace_name
  depends_on = [
    spot_cloudspace.cluster,
    spot_ondemandnodepool.ondemand_pools, # Depends on the collection of pools
    spot_spotnodepool.spot_pools          # Depends on the collection of pools
  ]
}