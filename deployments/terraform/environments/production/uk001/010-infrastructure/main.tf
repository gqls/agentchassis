provider "spot" {
  token = var.rackspace_spot_token
}

module "kubernetes_cluster" {
  source = "../../../../modules/rackspace-kubernetes"

  cluster_name     = "personae-prod-uk001"
  rackspace_region = "uk-lon-1"
  preemption_webhook_url = var.slack_webhook_url

  # Define the mix of spot instance pools from your old .tfvars
  spot_node_pools = {
    "gp-large" = {
      min_nodes = 3
      max_nodes = 6
      flavor    = "gp.vs1.large-lon"
      max_price = 0.035
    },
    "mh-medium" = {
      min_nodes = 2
      max_nodes = 4
      flavor    = "mh.vs1.medium-lon"
      max_price = 0.030
    }
  }
}
