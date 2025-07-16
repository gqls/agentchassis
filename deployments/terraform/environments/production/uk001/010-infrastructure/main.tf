# ~/projects/terraform/rackspace_generic/terraform/environments/production/uk001/010-infrastructure/main.tf
module "kubernetes_cluster" { # Local name for this instance of the module
  source = "../../../../modules/kubernetes_cluster_rackspace" # Path to the reusable module

  # Map variables from this root module (010-infrastructure) to the module's input variables
  cluster_name         = var.instance_cluster_name
  rackspace_region     = var.instance_rackspace_region
  preemption_webhook_url = var.instance_slack_webhook_url
  ondemand_node_flavor = var.instance_ondemand_node_flavor
  spot_node_flavor     = var.instance_spot_node_flavor
  ondemand_node_taints   = var.instance_ondemand_node_taints
  ondemand_node_count = var.instance_ondemand_node_count
  spot_min_nodes = var.instance_spot_min_nodes
  spot_max_nodes = var.instance_spot_max_nodes
  spot_max_price = var.instance_spot_max_price

  # Pass values for other variables defined in the module's variables.tf
  # If the module has defaults for these, you only need to pass them if you want to override.
  # kubernetes_version   = "1.31.1" # Or use var.instance_k8s_version
  # ondemand_node_count  = 1
  # spot_min_nodes       = 1
  # spot_max_nodes       = 2
  # ... etc. for cni, hacontrol_plane, spot_max_price, labels


  # Ensure ondemand_node_count is set, likely via terraform.tfvars
  # For example, if you have 'instance_ondemand_node_count' in your root variables.tf:
  # ondemand_node_count    = var.instance_ondemand_node_count
}